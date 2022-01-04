package migrations

import (
	"database/sql"
	"fmt"

	"github.com/lopezator/migrator"
)

func Migrate(db *sql.DB) error {
	m, err := migrator.New(
		migrator.Migrations(Migrations...),
	)
	if err != nil {
		return fmt.Errorf("error while loading migrations: %w", err)
	}

	if err := m.Migrate(db); err != nil {
		return fmt.Errorf("error while migrating: %w", err)
	}

	return nil
}
