package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"blazing/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handlers.Dashboard)
	r.Get("/auth/github", handlers.GitHubAuth)
	r.Get("/auth/github/callback", handlers.GitHubCallback)
	r.Get("/rooms/{roomID}", handlers.Room)
	r.Post("/rooms", handlers.CreateRoom)
	r.Get("/ws/{roomID}", handlers.WebSocket)

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
