package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"blazing/internal/app"
	"blazing/internal/db"
	"blazing/internal/handlers"
)

func main() {
	slog.Info("Starting chat server")
	if err := run(); err != nil {
		slog.Error("Server failed to run", "error", err)
		log.Fatal(err)
	}
}

func run() error {
	if err := validateConfig(); err != nil {
		slog.Error("Configuration validation failed", "error", err)
		return fmt.Errorf("configuration error: %w", err)
	}

	database, err := db.OpenSQLite("")
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer database.Close()

	application, err := app.New(database, os.Getenv("SESSION_SECRET"))
	if err != nil {
		slog.Error("Failed to create application", "error", err)
		return fmt.Errorf("failed to create app: %w", err)
	}

	h, err := handlers.New(application)
	if err != nil {
		slog.Error("Failed to create handlers", "error", err)
		return fmt.Errorf("failed to create handlers: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(15 * time.Second))

	// Public routes
	r.Get("/", h.Dashboard)
	r.Get("/auth/github", h.GitHubAuth)
	r.Get("/auth/github/callback", h.GitHubCallback)
	r.Get("/logout", h.Logout)

	// Authenticated routes
	r.Route("/rooms", func(r chi.Router) {
		r.Use(h.RequireAuth)
		r.Get("/{roomID}", h.Room)
		r.Post("/", h.CreateRoom)
	})
	r.Route("/ws", func(r chi.Router) {
		r.Use(h.RequireAuth)
		r.Get("/{roomID}", h.WebSocket)
	})

	slog.Info("HTTP routes configured successfully")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	slog.Info("Preparing HTTP server", "port", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("HTTP server listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err, "port", port)
			log.Fatal("Server failed to start:", err)
		}
	}()

	slog.Info("Server started successfully, waiting for interrupt signal")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	slog.Info("Interrupt signal received, beginning graceful shutdown")
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	slog.Info("Starting graceful shutdown", "timeout", "30s")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Forced server shutdown", "error", err)
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server shutdown completed successfully")
	log.Println("Server exited cleanly")
	return nil
}

// SESSION_SECRET and GITHUB_CLIENT_ID/GITHUB_CLIENT_SECRET are always required
func validateConfig() error {
	sessionSecret := os.Getenv("SESSION_SECRET")

	slog.Info("Validating configuration", "session_secret_length", len(sessionSecret))

	if len(sessionSecret) < 32 {
		slog.Error("SESSION_SECRET too short",
			"length", len(sessionSecret),
			"required_min", 32)
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}
	slog.Info("SESSION_SECRET validated", "length", len(sessionSecret))

	if os.Getenv("GITHUB_CLIENT_ID") == "" {
		slog.Error("GITHUB_CLIENT_ID missing - create a GitHub OAuth app and set this environment variable")
		return fmt.Errorf("GITHUB_CLIENT_ID environment variable is required")
	}
	if os.Getenv("GITHUB_CLIENT_SECRET") == "" {
		slog.Error("GITHUB_CLIENT_SECRET missing - create a GitHub OAuth app and set this environment variable")
		return fmt.Errorf("GITHUB_CLIENT_SECRET environment variable is required")
	}
	slog.Info("GitHub OAuth credentials validated")

	return nil
}
