package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"blazing/internal/db"
	"blazing/internal/session"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

func getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
		RedirectURL:  getRedirectURL(),
	}
}

func getRedirectURL() string {
	redirectURL := os.Getenv("GITHUB_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/github/callback"
	}
	return redirectURL
}

func (h *Handlers) GitHubAuth(w http.ResponseWriter, r *http.Request) {
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		slog.Error("GitHub OAuth not configured")
		http.Error(w, "GitHub OAuth not configured. Please set up a GitHub OAuth app and configure GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET environment variables.", http.StatusInternalServerError)
		return
	}

	state, err := session.GenerateState()
	if err != nil {
		slog.Error("Failed to generate OAuth state", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   os.Getenv("GO_ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	})

	oauthConfig := getOAuthConfig()
	url := oauthConfig.AuthCodeURL(state)

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *Handlers) GitHubCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !h.verifyState(r) {
		slog.Error("OAuth state verification failed")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	h.clearStateCookie(w)

	code := r.URL.Query().Get("code")
	if code == "" {
		slog.Error("No authorization code in callback")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	githubUser, err := h.getGitHubUser(ctx, code)
	if err != nil {
		slog.Error("Failed to get GitHub user", "error", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	user, err := h.createOrUpdateUser(ctx, githubUser)
	if err != nil {
		slog.Error("Failed to create/update user", "error", err, "github_uid", githubUser.ID)
		http.Error(w, "Failed to process user", http.StatusInternalServerError)
		return
	}

	sessionUser := &session.User{
		ID:        user.ID,
		GitHubUID: user.GithubUid,
		Login:     user.Login,
		AvatarURL: user.AvatarUrl.String,
	}

	if err := h.app.Session.Set(w, sessionUser); err != nil {
		slog.Error("Failed to set session", "error", err, "user_id", sessionUser.ID)
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	h.app.Session.Clear(w)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (h *Handlers) verifyState(r *http.Request) bool {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil {
		return false
	}
	return r.URL.Query().Get("state") == stateCookie.Value
}

func (h *Handlers) clearStateCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   os.Getenv("GO_ENV") == "production",
	})
}

func (h *Handlers) getGitHubUser(ctx context.Context, code string) (*GitHubUser, error) {
	oauthConfig := getOAuthConfig()
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var user GitHubUser
	if err := json.Unmarshal(body, &user); err != nil {
		slog.Error("Failed to parse GitHub user data", "error", err)
		return nil, fmt.Errorf("failed to parse user data: %w", err)
	}

	return &user, nil
}

func (h *Handlers) createOrUpdateUser(ctx context.Context, githubUser *GitHubUser) (*db.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := h.app.DB.GetUserByGitHubUID(ctx, githubUser.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			newUser, err := h.app.DB.CreateUser(ctx, db.CreateUserParams{
				GithubUid: githubUser.ID,
				Login:     githubUser.Login,
				AvatarUrl: sql.NullString{String: githubUser.AvatarURL, Valid: githubUser.AvatarURL != ""},
			})
			if err != nil {
				slog.Error("Failed to create new user", "error", err, "github_uid", githubUser.ID)
				return nil, err
			}
			return &newUser, nil
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	if user.Login != githubUser.Login || user.AvatarUrl.String != githubUser.AvatarURL {
		err = h.app.DB.UpdateUser(ctx, db.UpdateUserParams{
			Login:     githubUser.Login,
			AvatarUrl: sql.NullString{String: githubUser.AvatarURL, Valid: githubUser.AvatarURL != ""},
			ID:        user.ID,
		})
		if err != nil {
			slog.Error("Failed to update user", "error", err, "user_id", user.ID)
		}
	}

	return &user, nil
}
