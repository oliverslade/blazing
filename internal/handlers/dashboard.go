package handlers

import (
	"net/http"
)

// Dashboard handles the main view
func Dashboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Show GitHub sign-in button or user's room list based on auth status

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
