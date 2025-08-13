package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blazing/internal/session"
)

func TestDashboard(t *testing.T) {
	testApp, h := setupTestApp(t)

	t.Run("shows login page when not authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()

		h.Dashboard(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, "Sign in with GitHub") {
			t.Error("Expected login page content")
		}
	})

	t.Run("shows dashboard when authenticated", func(t *testing.T) {
		testUser := &session.User{
			ID:        1,
			GitHubUID: 12345,
			Login:     "testuser",
			AvatarURL: "https://example.com/avatar.jpg",
		}

		tempW := httptest.NewRecorder()

		if err := testApp.Session.Set(tempW, testUser); err != nil {
			t.Fatalf("Failed to set test session: %v", err)
		}

		cookies := tempW.Result().Cookies()
		req := httptest.NewRequest("GET", "/", nil)
		for _, c := range cookies {
			req.AddCookie(c)
		}

		user, err := testApp.Session.Get(req)
		if err != nil {
			t.Fatalf("Session should be valid but got error: %v", err)
		}
		if user.Login != "testuser" {
			t.Errorf("Expected user login 'testuser', got '%s'", user.Login)
		}

		w := httptest.NewRecorder()
		h.Dashboard(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()

		if !strings.Contains(body, "Welcome, testuser") {
			t.Errorf("Expected dashboard to show username in nav, got body: %s", body)
		}

		if !strings.Contains(body, "Your chats will appear here") {
			t.Errorf("Expected dashboard content, got: %s", body)
		}
	})
}
