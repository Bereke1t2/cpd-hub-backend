package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending up migrations from dir against databaseURL.
// No-op (nil) when the schema is already current.
func RunMigrations(databaseURL, dir string) error {
	m, err := migrate.New("file://"+dir, databaseURL)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}

// RollbackOne reverts the most recent migration (used by the CLI and tests).
func RollbackOne(databaseURL, dir string) error {
	m, err := migrate.New("file://"+dir, databaseURL)
	if err != nil {
		return fmt.Errorf("init migrate: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("rollback: %w", err)
	}
	return nil
}
