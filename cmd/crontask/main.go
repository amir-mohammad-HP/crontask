package main

import (
	"log"
	"os"

	"github.com/amir-mohammad-HP/crontask/internal/app"
	"github.com/amir-mohammad-HP/crontask/internal/config"
	"github.com/amir-mohammad-HP/crontask/pkg/logger"
)

// main is the entry point of the crontask application.
// It performs the following sequence:
//  1. Loads application configuration
//  2. Initializes the structured logger
//  3. Creates and runs the main application instance
//
// Exit codes:
//   - 0: Successful execution
//   - 1: Configuration loading failed or application runtime error
func main() {
	// Load configuration from environment variables and/or config files
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize structured logger with configured log level
	logger := logger.New(cfg.LogLevel)

	// Create application instance with dependencies
	app := app.New(cfg, logger)

	// Run the main application loop
	if err := app.Run(); err != nil {
		logger.Error("Application failed: %s", err)
		os.Exit(1)
	}
}
