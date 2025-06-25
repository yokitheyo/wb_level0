package app

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yokitheyo/wb_level0/internal/cache"
	"github.com/yokitheyo/wb_level0/internal/config"
	"github.com/yokitheyo/wb_level0/internal/database"
	"github.com/yokitheyo/wb_level0/internal/handlers"
	"github.com/yokitheyo/wb_level0/internal/kafka"
	"github.com/yokitheyo/wb_level0/internal/repository"
	"github.com/yokitheyo/wb_level0/internal/services"
	"go.uber.org/zap"
)

type App struct {
	config   *config.Config
	logger   *zap.Logger
	db       *database.Database
	server   *http.Server
	consumer *kafka.Consumer
	cancel   context.CancelFunc
}

func NewApp(configPath string, logger *zap.Logger) (*App, error) {
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		return nil, err
	}

	return &App{
		config: cfg,
		logger: logger,
		db:     db,
	}, nil
}

func (a *App) Start() error {
	orderRepo := repository.NewOrderRepository(a.db.GetPool(), a.logger)
	orderCache := cache.NewOrderCache(a.logger)
	orderService := services.NewOrderService(orderRepo, orderCache, a.logger)

	if err := orderService.RestoreCache(context.Background()); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	consumer := kafka.NewConsumer(
		a.config.Kafka.Brokers,
		a.config.Kafka.Topic,
		a.config.Kafka.GroupID,
		orderService,
		a.logger,
	)
	a.consumer = consumer
	consumer.Start(ctx)

	router := a.setupRouter(orderService)

	a.server = &http.Server{
		Addr:    ":" + a.config.Server.Port,
		Handler: router,
	}

	go func() {
		a.logger.Info("starting http server", zap.String("port", a.config.Server.Port))
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatal("failed to start http server", zap.Error(err))
		}
	}()

	return nil
}

func (a *App) setupRouter(orderService services.OrderService) *gin.Engine {
	router := gin.Default()

	router.LoadHTMLGlob("templates/*")
	router.Static("/static", "static")

	orderHandler := handlers.NewOrderHandler(orderService, a.logger)
	orderHandler.RegisterRoutes(router)

	return router
}

func (a *App) Stop() error {
	a.logger.Info("shutting down application")

	if a.cancel != nil {
		a.cancel()
	}

	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(ctx); err != nil {
			a.logger.Error("failed to shutdown server", zap.Error(err))
			return err
		}
	}

	if a.db != nil {
		a.db.Close()
	}

	a.logger.Info("application shutdown successfully")
	return nil
}

func (a *App) Run() error {
	if err := a.Start(); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	return a.Stop()
}
