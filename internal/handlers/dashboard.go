package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"blazing/internal/session"
)

type DashboardData struct {
	User *session.User
}

func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := h.app.Session.Get(r)
	if err != nil {
		if errors.Is(err, session.ErrNoSession) || errors.Is(err, session.ErrInvalidSession) {
			if err := h.loginTemplate.ExecuteTemplate(w, "login", nil); err != nil {
				slog.Error("Failed to render login template", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}
		slog.Error("Unexpected session error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	data := DashboardData{
		User: user,
	}

	if err := h.dashboardTemplate.ExecuteTemplate(w, "dashboard", data); err != nil {
		slog.Error("Failed to render dashboard template", "error", err, "user_id", user.ID)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
