package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"blazing/internal/session"
)

type userContextKey struct{}

func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.app.Session.Get(r)
		if err != nil {
			if errors.Is(err, session.ErrNoSession) || errors.Is(err, session.ErrInvalidSession) {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			slog.Error("Session error in auth middleware", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserFromContext(r *http.Request) (*session.User, bool) {
	user, ok := r.Context().Value(userContextKey{}).(*session.User)
	return user, ok
}

func (h *Handlers) RequireAuthWithRedirect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := h.app.Session.Get(r)
		if err != nil {
			if errors.Is(err, session.ErrNoSession) || errors.Is(err, session.ErrInvalidSession) {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			slog.Error("Session error in auth middleware", "error", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
