package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/lopezator/migrator"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrations returns all registered migrations.
var Migrations = func() []interface{} {
	m, err := loadMigrations()
	if err != nil {
		panic(fmt.Errorf("failed to load migrations: %w", err))
	}
	return m
}()

// Migrate migrates the database to the newest migration.
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

// Pending returns all pending migrations.
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

func loadMigrations() (migrations []interface{}, err error) {
	migrations = make([]interface{}, 0)

	// the only possible error is bad pattern and can be safely ignored
	files, _ := fs.Glob(migrationFiles, "migrations/*")

	for _, file := range files {
		file := file
		migration, err := fs.ReadFile(migrationFiles, file)
		if err != nil {
			return nil, fmt.Errorf("error reading migration file: %w", err)
		}
		migrations = append(migrations, &migrator.Migration{
			Name: file,
			Func: func(tx *sql.Tx) error {
				if _, err := tx.Exec(string(migration)); err != nil {
					return err
				}
				return nil
			},
		})
	}

	return migrations, nil
}
