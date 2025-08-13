package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	cookieName = "blazing_session"
	cookieAge  = 7 * 24 * time.Hour
)

var (
	ErrInvalidSession = errors.New("invalid session")
	ErrNoSession      = errors.New("no session found")
)

type User struct {
	ID        int64  `json:"id"`
	GitHubUID int64  `json:"github_uid"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}

type Manager struct {
	key []byte
}

func NewManager(secretKey string) (*Manager, error) {
	if secretKey == "" {
		secretKey = os.Getenv("SESSION_SECRET")
		if secretKey == "" {
			return nil, errors.New("SESSION_SECRET environment variable is required")
		}
	}

	if len(secretKey) < 32 {
		return nil, errors.New("session secret must be at least 32 characters")
	}

	return &Manager{
		key: []byte(secretKey),
	}, nil
}

func (m *Manager) sign(data string) string {
	h := hmac.New(sha256.New, m.key)
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func (m *Manager) verify(data, signature string) bool {
	expected := m.sign(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (m *Manager) Set(w http.ResponseWriter, user *User) error {
	data, err := json.Marshal(user)
	if err != nil {
		slog.Error("Failed to marshal user session data", "error", err, "user_id", user.ID)
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(data)
	signature := m.sign(encoded)
	value := fmt.Sprintf("%s.%s", encoded, signature)

	isSecure := os.Getenv("GO_ENV") == "production"
	maxAge := int(cookieAge.Seconds())

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

func (m *Manager) Get(r *http.Request) (*User, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrNoSession
		}
		return nil, err
	}

	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidSession
	}

	data, signature := parts[0], parts[1]

	if !m.verify(data, signature) {
		return nil, ErrInvalidSession
	}

	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return nil, ErrInvalidSession
	}

	var user User
	if err := json.Unmarshal(decoded, &user); err != nil {
		slog.Error("Failed to unmarshal session user data", "error", err)
		return nil, ErrInvalidSession
	}

	return &user, nil
}

func (m *Manager) Clear(w http.ResponseWriter) {
	isSecure := os.Getenv("GO_ENV") == "production"
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
