package migrations

import (
	"database/sql"
	"fmt"

	"github.com/lopezator/migrator"
)

func Migrate(db *sql.DB) error {
	m, err := newMigrator()
	if err != nil {
		return err
	}

	if err := m.Migrate(db); err != nil {
		return fmt.Errorf("error while migrating: %w", err)
	}
	return nil
}

func Pending(db *sql.DB) ([]*migrator.Migration, error) {
	m, err := newMigrator()
	if err != nil {
		return nil, err
	}

	pending, err := m.Pending(db)
	if err != nil {
		return nil, fmt.Errorf("error while querying for pending migrations: %w", err)
	}

	pm := make([]*migrator.Migration, 0, len(pending))
	for _, pp := range pending {
		pm = append(pm, pp.(*migrator.Migration))
	}
	return pm, nil
}

func newMigrator() (*migrator.Migrator, error) {
	m, err := migrator.New(
		migrator.Migrations(Migrations...),
	)
	if err != nil {
		return m, fmt.Errorf("error while loading migrations: %w", err)
	}

	return m, nil
}
