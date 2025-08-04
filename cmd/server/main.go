package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"blazing/internal/db"
	"blazing/internal/handlers"
)

type App struct {
	DB *sql.DB
}

// TODO: Replace with proper DI when handlers need DB
var DB *sql.DB

func main() {
	database, err := db.OpenSQLite("")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer func() {
		if err := database.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	app := &App{DB: database}
	DB = app.DB

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Get("/", handlers.Dashboard)
	r.Get("/auth/github", handlers.GitHubAuth)
	r.Get("/auth/github/callback", handlers.GitHubCallback)
	r.Get("/rooms/{roomID}", handlers.Room)
	r.Post("/rooms", handlers.CreateRoom)
	r.Get("/ws/{roomID}", handlers.WebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
