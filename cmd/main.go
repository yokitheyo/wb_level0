package main

import (
	"fmt"
	"log"
	"os"

	"github.com/yokitheyo/wb_level0/internal/app"
	"go.uber.org/zap"
)

func main() {
	dir, _ := os.Getwd()
	fmt.Println("Working dir:", dir)

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	application, err := app.NewApp("config/config.yaml", logger)
	if err != nil {
		logger.Fatal("failed to create application", zap.Error(err))
	}

	if err := application.Run(); err != nil {
		logger.Fatal("application failed", zap.Error(err))
	}
}
