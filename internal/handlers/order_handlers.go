package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yokitheyo/wb_level0/internal/services"
	"go.uber.org/zap"
)

type OrderHandler struct {
	service services.OrderService
	logger  *zap.Logger
}

func NewOrderHandler(service services.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	orderUID := c.Param("order_uid")
	if orderUID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order_uid is required"})
		return
	}

	order, err := h.service.GetOrderByID(c, orderUID)
	if err != nil {
		h.logger.Error("failed to get order", zap.Error(err), zap.String("order_uid", orderUID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get order"})
		return
	}

	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *OrderHandler) GetHomePage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Сервис заказов WB Level 0",
	})
}

func (h *OrderHandler) GetCacheStats(c *gin.Context) {
	stats := h.service.GetCacheStats()
	c.JSON(http.StatusOK, stats)
}

func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/", h.GetHomePage)
	router.GET("/order/:order_uid", h.GetOrderByID)
	router.GET("/cache/stats", h.GetCacheStats)
}
