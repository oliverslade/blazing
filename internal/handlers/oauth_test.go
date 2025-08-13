package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"blazing/internal/app"
	"blazing/internal/db"
	"blazing/internal/session"
)

func setupTestApp(t *testing.T) (*app.App, *Handlers) {
	database, err := db.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	testApp, err := app.New(database, "test-secret-key-that-is-long-enough-for-testing")
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}

	h, err := New(testApp)
	if err != nil {
		t.Fatalf("Failed to create handlers: %v", err)
	}

	return testApp, h
}

func TestGitHubAuth(t *testing.T) {
	originalClientID := os.Getenv("GITHUB_CLIENT_ID")
	originalClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	originalGoEnv := os.Getenv("GO_ENV")

	t.Cleanup(func() {
		os.Setenv("GITHUB_CLIENT_ID", originalClientID)
		os.Setenv("GITHUB_CLIENT_SECRET", originalClientSecret)
		os.Setenv("GO_ENV", originalGoEnv)
	})

	os.Setenv("GITHUB_CLIENT_ID", "test-client-id")
	os.Setenv("GITHUB_CLIENT_SECRET", "test-client-secret")
	os.Setenv("GO_ENV", "production")

	_, h := setupTestApp(t)

	req := httptest.NewRequest("GET", "/auth/github", nil)
	w := httptest.NewRecorder()

	h.GitHubAuth(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("Expected Location header, got none")
	}

	u, err := url.Parse(location)
	if err != nil {
		t.Fatalf("Failed to parse redirect URL: %v", err)
	}

	if u.Host != "github.com" {
		t.Errorf("Expected redirect to github.com, got %s", u.Host)
	}

	state := u.Query().Get("state")
	if state == "" {
		t.Error("Expected state parameter in redirect URL")
	}

	cookies := w.Result().Cookies()
	var stateCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "oauth_state" {
			stateCookie = c
			break
		}
	}

	if stateCookie == nil {
		t.Fatal("Expected oauth_state cookie to be set")
	}

	if stateCookie.Value != state {
		t.Error("State cookie value doesn't match URL state parameter")
	}
}

func TestGitHubAuthMissingConfig(t *testing.T) {
	originalClientID := os.Getenv("GITHUB_CLIENT_ID")
	originalClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	t.Cleanup(func() {
		os.Setenv("GITHUB_CLIENT_ID", originalClientID)
		os.Setenv("GITHUB_CLIENT_SECRET", originalClientSecret)
	})

	os.Setenv("GITHUB_CLIENT_ID", "")
	os.Setenv("GITHUB_CLIENT_SECRET", "")

	_, h := setupTestApp(t)

	req := httptest.NewRequest("GET", "/auth/github", nil)
	w := httptest.NewRecorder()

	h.GitHubAuth(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "GitHub OAuth not configured") {
		t.Errorf("Expected error message about GitHub OAuth configuration, got: %s", body)
	}
}

func TestGitHubCallback_InvalidState(t *testing.T) {
	_, h := setupTestApp(t)

	tests := []struct {
		name        string
		cookieState string
		urlState    string
	}{
		{"No cookie", "", "some-state"},
		{"No URL state", "some-state", ""},
		{"State mismatch", "state1", "state2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/auth/github/callback?state="+tt.urlState, nil)
			if tt.cookieState != "" {
				req.AddCookie(&http.Cookie{
					Name:  "oauth_state",
					Value: tt.cookieState,
				})
			}

			w := httptest.NewRecorder()
			h.GitHubCallback(w, req)

			if w.Code != http.StatusTemporaryRedirect {
				t.Errorf("Expected redirect status %d, got %d", http.StatusTemporaryRedirect, w.Code)
			}

			if location := w.Header().Get("Location"); location != "/" {
				t.Errorf("Expected redirect to /, got %s", location)
			}
		})
	}
}

func TestGitHubCallback_NoCode(t *testing.T) {
	_, h := setupTestApp(t)

	req := httptest.NewRequest("GET", "/auth/github/callback?state=valid-state", nil)
	req.AddCookie(&http.Cookie{
		Name:  "oauth_state",
		Value: "valid-state",
	})

	w := httptest.NewRecorder()
	h.GitHubCallback(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	if location := w.Header().Get("Location"); location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}
}

func TestLogout(t *testing.T) {
	testApp, h := setupTestApp(t)

	testUser := &session.User{
		ID:        1,
		GitHubUID: 12345,
		Login:     "testuser",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	req := httptest.NewRequest("GET", "/logout", nil)
	w := httptest.NewRecorder()

	if err := testApp.Session.Set(w, testUser); err != nil {
		t.Fatalf("Failed to set test session: %v", err)
	}

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		req.AddCookie(c)
	}

	w = httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status %d, got %d", http.StatusTemporaryRedirect, w.Code)
	}

	if location := w.Header().Get("Location"); location != "/" {
		t.Errorf("Expected redirect to /, got %s", location)
	}

	cookies = w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "blazing_session" && c.MaxAge != -1 {
			t.Error("Expected session cookie to be cleared (MaxAge = -1)")
		}
	}
}

func TestCreateOrUpdateUser(t *testing.T) {
	testApp, h := setupTestApp(t)

	githubUser := &GitHubUser{
		ID:        12345,
		Login:     "testuser",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	ctx := context.Background()
	user, err := h.createOrUpdateUser(ctx, githubUser)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.GithubUid != githubUser.ID {
		t.Errorf("Expected GitHub UID %d, got %d", githubUser.ID, user.GithubUid)
	}

	if user.Login != githubUser.Login {
		t.Errorf("Expected login %s, got %s", githubUser.Login, user.Login)
	}

	user2, err := h.createOrUpdateUser(ctx, githubUser)
	if err != nil {
		t.Fatalf("Failed to get existing user: %v", err)
	}

	if user.ID != user2.ID {
		t.Error("Expected same user ID for existing user")
	}

	githubUser.Login = "updateduser"
	githubUser.AvatarURL = "https://example.com/new-avatar.jpg"

	user3, err := h.createOrUpdateUser(ctx, githubUser)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	if user3.ID != user.ID {
		t.Error("Expected same user ID after update")
	}

	updatedUser, err := testApp.DB.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to fetch updated user: %v", err)
	}

	if updatedUser.Login != "updateduser" {
		t.Errorf("Expected updated login 'updateduser', got %s", updatedUser.Login)
	}
}
