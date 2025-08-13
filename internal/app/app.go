package app

import (
	"database/sql"
	"fmt"

	"blazing/internal/db"
	"blazing/internal/session"
)

type App struct {
	DB      *db.Queries
	Session *session.Manager
}

func New(database *sql.DB, sessionSecret string) (*App, error) {
	if database == nil {
		return nil, fmt.Errorf("database is required")
	}

	sessionManager, err := session.NewManager(sessionSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to create session manager: %w", err)
	}

	return &App{
		DB:      db.New(database),
		Session: sessionManager,
	}, nil
}
