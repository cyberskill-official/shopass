package migrate

import (
	"errors"

	"github.com/golang-migrate/migrate/v4"
)

type Migrator struct {
	m *migrate.Migrate
}

func New(m *migrate.Migrate) *Migrator {
	return &Migrator{m: m}
}

func (mg *Migrator) Up() error {
	if err := mg.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func (mg *Migrator) Down(n int) error {
	return mg.m.Steps(-n)
}

func (mg *Migrator) Version() (uint, bool, error) {
	return mg.m.Version()
}
