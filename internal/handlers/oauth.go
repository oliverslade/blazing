package handlers

import (
	"net/http"
)

// initiates GitHub OAuth flow
func GitHubAuth(w http.ResponseWriter, r *http.Request) {
	// TODO: Redirect to GitHub OAuth with secure state parameter

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// handles GitHub OAuth callback
func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: Exchange code for token, create/update user, start session

	http.Error(w, "Not implemented", http.StatusNotImplemented)
}
