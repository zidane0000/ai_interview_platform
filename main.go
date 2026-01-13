// Entry point for the AI Interview Backend application
// Responsible for initializing configuration, database, router, and starting the server
package main

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/zidane0000/ai-interview-platform/api"
	"github.com/zidane0000/ai-interview-platform/config"
	"github.com/zidane0000/ai-interview-platform/data"
	"github.com/zidane0000/ai-interview-platform/utils"
)

//go:embed frontend/dist
var frontendFS embed.FS

// spaHandler serves the SPA (Single Page Application) with fallback to index.html
// This allows React Router to handle client-side routing
func spaHandler() http.Handler {
	// Get the frontend filesystem from the embedded FS
	frontendDist, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		utils.Errorf("Failed to create frontend filesystem: %v", err)
		// Return a simple error handler
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Frontend not available", http.StatusServiceUnavailable)
		})
	}

	fileServer := http.FileServer(http.FS(frontendDist))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Try to open the file
		_, err := frontendDist.Open(path)
		if err != nil {
			// File doesn't exist, serve index.html for SPA routing
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}

// gracefulShutdown handles graceful shutdown of the application
func gracefulShutdown(server *http.Server, timeout time.Duration) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)

	// Register the channel to receive specific signals
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	// Block until we receive a signal
	sig := <-quit
	utils.Errorf("Received signal: %v. Starting graceful shutdown...", sig)

	// Create a deadline to wait for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	// Attempt to gracefully shutdown the server
	if err := server.Shutdown(ctx); err != nil {
		utils.Errorf("Server forced to shutdown: %v", err)
		os.Exit(1) // Exit with error code 1
	}

	// Additional cleanup operations
	utils.Infof("Performing cleanup operations...")
	// Close database connections if available
	if data.GlobalStore != nil {
		if err := data.GlobalStore.Close(); err != nil {
			utils.Errorf("Error closing database connections: %v", err)
			os.Exit(2) // Exit with error code 2 for database cleanup failure
		}
	}

	utils.Infof("Graceful shutdown completed successfully")
}

func main() {
	// Load configuration
	utils.Infof("Loading configuration...")
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	// TODO: Initialize logging with proper configuration
	// TODO: Add structured logging with levels (debug, info, warn, error)
	// TODO: Add log rotation and file output options

	// Initialize hybrid store (auto-detects memory vs database backend)
	utils.Infof("Initializing data store...")
	err = data.InitGlobalStore()
	if err != nil {
		utils.Errorf("failed to initialize store: %v", err)
		os.Exit(1)
	}

	// Log the backend being used
	if data.GlobalStore.GetBackend() == data.BackendDatabase {
		utils.Infof("Using PostgreSQL database backend")
	} else {
		utils.Infof("Using in-memory store backend (set DATABASE_URL for database mode)")
	}
	// TODO: Add store health checks
	// if err := data.GlobalStore.Health(); err != nil {
	//     utils.Errorf("store health check failed: %v", err)
	// }

	// Set up router with injected config (includes API routes and frontend serving)
	frontendHandler := spaHandler()
	router := api.SetupRouter(cfg, frontendHandler)
	// TODO: Add HTTPS support with TLS configuration
	// TODO: Add health check endpoints
	// TODO: Add metrics and monitoring endpoints
	// TODO: Add API documentation serving (Swagger/OpenAPI)
	// Create HTTP server with security timeouts
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Errorf("Server failed to start: %v", err)
			os.Exit(1)
		}
	}()
	utils.Infof("Server successfully started on port %s", cfg.Port)
	utils.Infof("Frontend can now connect to: http://localhost:%s", cfg.Port)

	// Start graceful shutdown handler (this will block until shutdown signal)
	gracefulShutdown(server, cfg.ShutdownTimeout)
}
