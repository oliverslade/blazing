package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Room(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	_ = roomID

	// TODO: Verify membership and render chat room with message history

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func CreateRoom(w http.ResponseWriter, r *http.Request) {
	// TODO: Create new room with authenticated user as creator and seed member

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func WebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	_ = roomID

	// TODO: Upgrade to WebSocket and handle real-time messaging for room members

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
