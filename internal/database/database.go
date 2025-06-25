package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yokitheyo/wb_level0/internal/config"
	"go.uber.org/zap"
)

type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewDatabase(cfg *config.DatabaseConfig, logger *zap.Logger) (*Database, error) {
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode,
	)

	logger.Info("connecting to database", zap.String("url", dbURL))

	pool, err := pgxpool.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("connected to database successfully")

	return &Database{
		Pool:   pool,
		logger: logger,
	}, nil
}

func (db *Database) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("database connection closed")
	}
}

func (db *Database) GetPool() *pgxpool.Pool {
	return db.Pool
}
