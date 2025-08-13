package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"blazing/internal/session"
)

func TestRequireAuth(t *testing.T) {
	testApp, h := setupTestApp(t)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r)
		if !ok {
			http.Error(w, "No user in context", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(user.Login))
	})

	authHandler := h.RequireAuth(testHandler)

	t.Run("rejects unauthenticated requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("allows authenticated requests", func(t *testing.T) {
		// Create a test user session
		testUser := &session.User{
			ID:        1,
			GitHubUID: 12345,
			Login:     "testuser",
			AvatarURL: "https://example.com/avatar.jpg",
		}

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		if err := testApp.Session.Set(w, testUser); err != nil {
			t.Fatalf("Failed to set test session: %v", err)
		}

		cookies := w.Result().Cookies()
		for _, c := range cookies {
			req.AddCookie(c)
		}

		w = httptest.NewRecorder()
		authHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if body != "testuser" {
			t.Errorf("Expected body 'testuser', got '%s'", body)
		}
	})

	t.Run("rejects invalid session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "blazing_session",
			Value: "invalid.session.data",
		})
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}

func TestRequireAuthWithRedirect(t *testing.T) {
	testApp, h := setupTestApp(t)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("protected content"))
	})

	authHandler := h.RequireAuthWithRedirect(testHandler)

	t.Run("redirects unauthenticated requests", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Errorf("Expected status %d, got %d", http.StatusTemporaryRedirect, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/" {
			t.Errorf("Expected redirect to /, got %s", location)
		}
	})

	t.Run("allows authenticated requests", func(t *testing.T) {
		testUser := &session.User{
			ID:        1,
			GitHubUID: 12345,
			Login:     "testuser",
			AvatarURL: "https://example.com/avatar.jpg",
		}

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		if err := testApp.Session.Set(w, testUser); err != nil {
			t.Fatalf("Failed to set test session: %v", err)
		}

		cookies := w.Result().Cookies()
		for _, c := range cookies {
			req.AddCookie(c)
		}

		w = httptest.NewRecorder()
		authHandler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}

		body := w.Body.String()
		if body != "protected content" {
			t.Errorf("Expected body 'protected content', got '%s'", body)
		}
	})
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("returns user when present", func(t *testing.T) {
		testUser := &session.User{
			ID:        1,
			GitHubUID: 12345,
			Login:     "testuser",
			AvatarURL: "https://example.com/avatar.jpg",
		}

		req := httptest.NewRequest("GET", "/", nil)

		ctx := context.WithValue(req.Context(), userContextKey{}, testUser)
		req = req.WithContext(ctx)

		user, ok := GetUserFromContext(req)
		if !ok {
			t.Error("Expected user to be found in context")
		}

		if user.Login != "testuser" {
			t.Errorf("Expected login 'testuser', got '%s'", user.Login)
		}
	})

	t.Run("returns false when user not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, ok := GetUserFromContext(req)
		if ok {
			t.Error("Expected user not to be found in context")
		}
	})
}
