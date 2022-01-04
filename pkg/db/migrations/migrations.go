package migrations

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/lopezator/migrator"
)

//go:embed *.sql
var migrationFiles embed.FS

var Migrations = func() []interface{} {
	m, err := load()
	if err != nil {
		panic(fmt.Errorf("failed to load migrations: %w", err))
	}
	return m
}()

func load() (migrations []interface{}, err error) {
	migrations = make([]interface{}, 0)

	// the only possible error is bad pattern and can be safely ignored
	files, _ := fs.Glob(migrationFiles, "*")

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
