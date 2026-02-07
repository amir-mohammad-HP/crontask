package main

import (
	"log"
	"os"

	"github.com/amir-mohammad-HP/crontask/internal/app"
	"github.com/amir-mohammad-HP/crontask/internal/config"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := logger.New(cfg.LogLevel)

	// Create and run application
	app := app.New(cfg, logger)
	if err := app.Run(); err != nil {
		logger.Error("Application failed %s", err)
		os.Exit(1)
	}
}
