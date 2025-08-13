package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSessionManager(t *testing.T) {
	manager, err := NewManager("test-secret-key-that-is-long-enough-for-testing")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	testUser := &User{
		ID:        1,
		GitHubUID: 12345,
		Login:     "testuser",
		AvatarURL: "https://example.com/avatar.jpg",
	}

	t.Run("Set and Get", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		if err := manager.Set(w, testUser); err != nil {
			t.Fatalf("Failed to set session: %v", err)
		}

		cookies := w.Result().Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == cookieName {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("Session cookie not set")
		}

		req.AddCookie(sessionCookie)

		user, err := manager.Get(req)
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		if user.ID != testUser.ID {
			t.Errorf("Expected user ID %d, got %d", testUser.ID, user.ID)
		}

		if user.Login != testUser.Login {
			t.Errorf("Expected login %s, got %s", testUser.Login, user.Login)
		}
	})

	t.Run("Invalid signature", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: "eyJ0ZXN0IjoidGFtcGVyZWQifQ==.invalidsignature",
		})

		_, err := manager.Get(req)
		if err != ErrInvalidSession {
			t.Errorf("Expected ErrInvalidSession, got %v", err)
		}
	})

	t.Run("No session", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)

		_, err := manager.Get(req)
		if err != ErrNoSession {
			t.Errorf("Expected ErrNoSession, got %v", err)
		}
	})

	t.Run("Clear session", func(t *testing.T) {
		w := httptest.NewRecorder()

		manager.Clear(w)

		cookies := w.Result().Cookies()
		for _, c := range cookies {
			if c.Name == cookieName {
				if c.MaxAge != -1 {
					t.Error("Expected MaxAge -1 for cleared cookie")
				}
				if c.Value != "" {
					t.Error("Expected empty value for cleared cookie")
				}
			}
		}
	})
}

func TestGenerateState(t *testing.T) {
	state1, err := GenerateState()
	if err != nil {
		t.Fatalf("Failed to generate state: %v", err)
	}

	if len(state1) == 0 {
		t.Error("Generated state is empty")
	}

	state2, err := GenerateState()
	if err != nil {
		t.Fatalf("Failed to generate second state: %v", err)
	}

	if state1 == state2 {
		t.Error("Generated states should be unique")
	}
}

func TestNewManager_Validation(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		expectErr bool
	}{
		{"Empty secret", "", true},
		{"Short secret", "short", true},
		{"Valid secret", strings.Repeat("a", 32), false},
		{"Long secret", strings.Repeat("a", 64), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.secret == "" {
				t.Setenv("SESSION_SECRET", "")
			}

			_, err := NewManager(tt.secret)
			if (err != nil) != tt.expectErr {
				t.Errorf("NewManager() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
