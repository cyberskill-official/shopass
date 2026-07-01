package migrate

import (
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a new migrator instance.
func NewMigrator(db *sql.DB, sourceURL string) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"postgres", driver)
	if err != nil {
		return nil, err
	}

	return &Migrator{m: m}, nil
}

// Up runs all available migrations up.
func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err // ErrNoChange is idempotent, not an error
	}
	return nil
}

// Down runs n migrations down.
func (mg *Migrator) Down(n int) error {
	return mg.m.Steps(-n)
}

// Version returns the current migration version.
func (mg *Migrator) Version() (uint, bool, error) {
	return mg.m.Version()
}

// Close closes the underlying migrator and database connection.
func (mg *Migrator) Close() error {
	_, err := mg.m.Close()
	return err
}
