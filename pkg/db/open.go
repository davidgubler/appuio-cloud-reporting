package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // postgres driver
)

func Open(dataSourceName string) (*sql.DB, error) {
	return sql.Open("pgx", dataSourceName)
}
