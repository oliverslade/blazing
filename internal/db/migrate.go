package db

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// finds and executes all pending migrations in the migrations directory.
func RunMigrations(db *sql.DB) error {
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	migrationFiles, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	for _, filename := range migrationFiles {
		if _, applied := appliedMigrations[filename]; !applied {
			log.Printf("Running migration: %s", filename)

			if err := runMigration(db, filename); err != nil {
				return fmt.Errorf("failed to run migration %s: %w", filename, err)
			}

			if err := recordMigration(db, filename); err != nil {
				return fmt.Errorf("failed to record migration %s: %w", filename, err)
			}

			log.Printf("Migration completed: %s", filename)
		}
	}

	if len(migrationFiles) > len(appliedMigrations) {
		log.Printf("All migrations completed (%d applied)", len(migrationFiles)-len(appliedMigrations))
	}

	return nil
}

func getAppliedMigrations(db *sql.DB) (map[string]bool, error) {
	applied := make(map[string]bool)

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='migrations'").Scan(&count)
	if err != nil {
		return applied, err
	}

	if count == 0 {
		return applied, nil
	}

	rows, err := db.Query("SELECT filename FROM migrations ORDER BY filename")
	if err != nil {
		return applied, err
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return applied, err
		}
		applied[filename] = true
	}

	return applied, rows.Err()
}

func getMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(os.DirFS("."), "migrations")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}

	// Sort files to ensure consistent order
	sort.Strings(files)

	return files, nil
}

func runMigration(db *sql.DB, filename string) error {
	migrationSQL, err := fs.ReadFile(os.DirFS("."), filepath.Join("migrations", filename))
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err = db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	return nil
}

func recordMigration(db *sql.DB, filename string) error {
	_, err := db.Exec("INSERT INTO migrations (filename) VALUES (?)", filename)
	return err
}
