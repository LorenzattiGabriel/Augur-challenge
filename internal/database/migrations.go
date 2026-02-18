package database

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	if err := createMigrationsTable(db); err != nil {
		return err
	}

	files, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var upMigrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			upMigrations = append(upMigrations, file.Name())
		}
	}
	sort.Strings(upMigrations)

	for _, fileName := range upMigrations {
		migrationName := strings.TrimSuffix(fileName, ".up.sql")

		applied, err := isMigrationApplied(db, migrationName)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", fileName))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", fileName, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", fileName, err)
		}

		if err := recordMigration(db, migrationName); err != nil {
			return err
		}
	}

	return nil
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}
	return nil
}

func isMigrationApplied(db *sql.DB, version string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`
	err := db.QueryRow(query, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return exists, nil
}

func recordMigration(db *sql.DB, version string) error {
	query := `INSERT INTO schema_migrations (version) VALUES ($1)`
	_, err := db.Exec(query, version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	return nil
}

func RollbackMigration(db *sql.DB, version string) error {
	fileName := version + ".down.sql"

	content, err := fs.ReadFile(migrationsFS, filepath.Join("migrations", fileName))
	if err != nil {
		return fmt.Errorf("failed to read rollback migration %s: %w", fileName, err)
	}

	if _, err := db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute rollback migration %s: %w", fileName, err)
	}

	query := `DELETE FROM schema_migrations WHERE version = $1`
	if _, err := db.Exec(query, version); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	return nil
}
