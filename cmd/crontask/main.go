// Declare this as the main package - Go's entry point for executables
package main

// Import necessary packages with descriptive grouping
import (
	"context"   // For cancellation and timeout handling
	"fmt"       // For formatted I/O (printing)
	"log"       // For structured logging
	"os"        // For OS interaction (signals, exit)
	"os/signal" // For handling OS signals
	"sync"      // For synchronization primitives (WaitGroup)
	"syscall"   // For system-specific signals
	"time"      // For time-based operations
)

// Define a custom type 'App' that represents our application
type App struct {
	shutdown chan struct{}  // Channel to signal shutdown (empty struct = signal only)
	wg       sync.WaitGroup // WaitGroup to track running goroutines
	logger   *log.Logger    // Custom logger for application messages
}

// Constructor function that creates a new App instance
func NewApp() *App {
	return &App{
		shutdown: make(chan struct{}),                              // Initialize shutdown channel
		logger:   log.New(os.Stdout, "[CRONTASK] ", log.LstdFlags), // Create logger with prefix
	}
}

// Main worker function that runs in a goroutine
func (a *App) mainLoop() {
	defer a.wg.Done() // Mark this goroutine as done when function exits
	a.logger.Println("Main loop started")

	// Create a ticker that sends current time every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop() // Ensure ticker is stopped when function exits

	// Infinite loop that runs until shutdown signal
	for {
		select { // Wait for multiple channel operations
		case <-a.shutdown: // Case 1: Shutdown signal received
			a.logger.Println("Main loop received shutdown signal")
			return // Exit function (triggers defer a.wg.Done())
		case t := <-ticker.C: // Case 2: Ticker fired (every 5 seconds)
			// Print current datetime in specific format
			fmt.Printf("Current datetime: %s\n", t.Format("2006-01-02 15:04:05"))

		}
	}
}

// Cleanup function that runs after shutdown
func (a *App) cleanup() {
	a.logger.Println("Starting cleanup process...")
	// clean ups ...
	a.logger.Println("Cleanup completed")
}

// Main application runner
func (a *App) Run() {

	// Start main loop in a goroutine (concurrent execution)
	a.wg.Add(1)     // Add 1 to WaitGroup counter
	go a.mainLoop() // Launch goroutine

	// Wait for shutdown signal (blocks until channel closes)
	<-a.shutdown

	a.wg.Wait()
	a.cleanup()
	a.logger.Println("Application shutdown complete")
}

// Method to initiate graceful shutdown
func (a *App) Stop() {
	close(a.shutdown) // Close channel to signal all listeners
}

// Main function - program entry point
func main() {
	// Create new App instance
	app := NewApp()

	// Create buffered channel for OS signals (buffer size 1)
	sigChan := make(chan os.Signal, 1)
	// Register which signals to capture and forward to sigChan
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	// Create a channel to know when app.Run() completes
	appDone := make(chan struct{})

	// Start application in background goroutine
	go func() {
		app.Run()
		close(appDone) // Signal that app.Run() completed
	}()

	// Wait for termination signal (blocks until signal received)
	sig := <-sigChan
	app.logger.Printf("Received signal: %v\n", sig)

	// Start graceful shutdown process
	app.logger.Println("Initiating graceful shutdown...")
	app.Stop() // Signal app to shutdown

	// Create timeout context: shutdown must complete within 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Wait for either shutdown completion or timeout
	select {
	case <-appDone: // App shutdown completed
		app.logger.Println("Application shutdown completed")
	case <-ctx.Done(): // Timeout exceeded
		app.logger.Println("Shutdown timeout exceeded, forcing exit")
		os.Exit(1)
	}
}
