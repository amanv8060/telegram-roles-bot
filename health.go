package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

// HealthChecker provides health check functionality
type HealthChecker struct {
	db *sql.DB
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *sql.DB) *HealthChecker {
	return &HealthChecker{db: db}
}

// Check performs a health check
func (hc *HealthChecker) Check(ctx context.Context) error {
	// Check database connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hc.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// StartHealthServer starts a simple HTTP health check server
func StartHealthServer(port string, healthChecker *HealthChecker) {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := healthChecker.Check(ctx); err != nil {
			Logger.WithError(err).Error("Health check failed")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "UNHEALTHY: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "HEALTHY")
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := healthChecker.Check(ctx); err != nil {
			Logger.WithError(err).Error("Readiness check failed")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "NOT READY: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "READY")
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	Logger.WithField("port", port).Info("Starting health check server")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		Logger.WithError(err).Error("Health check server failed")
	}
}
