package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func OpenSQLite(dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = os.Getenv("DB_PATH")
		if dbPath == "" {
			if os.Getenv("GO_ENV") == "test" {
				dbPath = "file::memory:?cache=shared"
			} else {
				dbPath = "/data/chat.db"
			}
		}
	}

	if dbPath != "file::memory:?cache=shared" && !filepath.IsAbs(dbPath) {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	dsn := buildDSN(dbPath)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func buildDSN(dbPath string) string {
	if dbPath == "file::memory:?cache=shared" {
		return dbPath
	}
	return "file:" + dbPath + "?_journal=WAL&_busy_timeout=5000&_foreign_keys=ON"
}
