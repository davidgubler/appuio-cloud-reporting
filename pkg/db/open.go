package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
)

const driver = "pgx"

// Open opens a postgres database with the pgx driver.
func Open(dataSourceName string) (*sql.DB, error) {
	return sql.Open(driver, dataSourceName)
}

// Openx opens a postgres database with the pgx driver and wraps it in an sqlx.DB.
func Openx(dataSourceName string) (*sqlx.DB, error) {
	db, err := Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, driver), nil
}

// NewDBx wraps the given database in an sqlx.DB.
func NewDBx(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, driver)
}
