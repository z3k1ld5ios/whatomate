// Package main is the entry point for the whatomate webhook notification service.
// It reads configuration from environment variables, sets up the HTTP server,
// and wires together the webhook handler and notifier.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shridarpatil/whatomate/internal/notifier"
	"github.com/shridarpatil/whatomate/internal/webhook"
)

const (
	defaultPort           = 8080
	defaultShutdownTimeout = 10 * time.Second
)

// config holds the runtime configuration for the service.
type config struct {
	Port            int
	NotifierURL     string
	NotifierTimeout time.Duration
	RetryAttempts   int
}

// configFromEnv reads configuration from environment variables, applying
// sensible defaults where values are not set.
func configFromEnv() (*config, error) {
	cfg := &config{
		Port:            defaultPort,
		NotifierTimeout: notifier.DefaultTimeout,
		RetryAttempts:   3,
	}

	if raw := os.Getenv("PORT"); raw != "" {
		p, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT %q: %w", raw, err)
		}
		cfg.Port = p
	}

	cfg.NotifierURL = os.Getenv("NOTIFIER_URL")
	if cfg.NotifierURL == "" {
		return nil, errors.New("NOTIFIER_URL environment variable is required")
	}

	if raw := os.Getenv("NOTIFIER_TIMEOUT"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid NOTIFIER_TIMEOUT %q: %w", raw, err)
		}
		cfg.NotifierTimeout = d
	}

	if raw := os.Getenv("RETRY_ATTEMPTS"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid RETRY_ATTEMPTS %q: %w", raw, err)
		}
		cfg.RetryAttempts = n
	}

	return cfg, nil
}

func main() {
	cfg, err := configFromEnv()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	n := notifier.New(
		notifier.WithTimeout(cfg.NotifierTimeout),
		notifier.WithRetryAttempts(cfg.RetryAttempts),
	)

	h := webhook.NewHandler(n, cfg.NotifierURL)

	mux := http.NewServeMux()
	mux.Handle("/webhook", h)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so we can listen for shutdown signals.
	go func() {
		log.Printf("whatomate listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("server stopped")
}
