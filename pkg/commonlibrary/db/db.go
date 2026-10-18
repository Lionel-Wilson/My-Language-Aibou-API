package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB) error {
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}

	migrationsDir, err := migrationsDir()
	if err != nil {
		return err
	}

	return goose.Up(db, migrationsDir)
}

func migrationsDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}

	dir := filepath.Join(filepath.Dir(exe), "migrations")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir, nil
	}

	if info, err := os.Stat("migrations"); err == nil && info.IsDir() {
		return "migrations", nil
	}

	return "", fmt.Errorf("migrations directory not found next to binary or working directory")
}
