// Package main is the entry point for the Telegram Role Bot application.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"didactic-spork/internal/bot"
	"didactic-spork/internal/config"
	"didactic-spork/internal/database"
	"didactic-spork/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel, cfg.Env == "production")
	log.Info("Starting Telegram Role Bot")

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create bot service
	botService, err := bot.New(cfg, db, log)
	if err != nil {
		return fmt.Errorf("failed to create bot service: %w", err)
	}

	// Set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Info("Shutdown signal received")
		cancel()
	}()

	// Start bot
	if err := botService.Start(ctx); err != nil {
		return fmt.Errorf("bot service error: %w", err)
	}

	log.Info("Bot stopped gracefully")
	return nil
}
