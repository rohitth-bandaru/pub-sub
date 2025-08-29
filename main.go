package main

import (
	"context"
	"os"
	"os/signal"
	"pub-sub/config"
	"pub-sub/logger"
	"pub-sub/pubsub"
	"pub-sub/server"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Validate configuration
	if err := cfg.ValidateConfig(); err != nil {
		panic("Configuration validation failed: " + err.Error())
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel, cfg.LogFormat)

	// Initialize pub-sub system
	pubSubSystem := pubsub.NewPubSub(cfg, log)

	// Create and start server
	srv := server.NewServer(cfg, log, pubSubSystem)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
		os.Exit(1)
	}
}
