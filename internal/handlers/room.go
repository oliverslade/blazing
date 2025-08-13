package handlers

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handlers) Room(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	user, ok := GetUserFromContext(r)
	if !ok {
		slog.Error("User not found in context for room access")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_ = roomID
	_ = user

	// TODO: Verify membership and render chat room with message history
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handlers) CreateRoom(w http.ResponseWriter, r *http.Request) {
	user, ok := GetUserFromContext(r)
	if !ok {
		slog.Error("User not found in context for room creation")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_ = user

	// TODO: Create new room with authenticated user as creator and seed member
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handlers) WebSocket(w http.ResponseWriter, r *http.Request) {
	roomID := chi.URLParam(r, "roomID")
	user, ok := GetUserFromContext(r)
	if !ok {
		slog.Error("User not found in context for WebSocket")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	_ = roomID
	_ = user

	// TODO: Upgrade to WebSocket and handle real-time messaging for room members
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
