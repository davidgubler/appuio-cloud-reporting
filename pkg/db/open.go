package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v4/stdlib" // postgres driver
	"github.com/jmoiron/sqlx"
)

const driver = "pgx"

func Open(dataSourceName string) (*sql.DB, error) {
	return sql.Open(driver, dataSourceName)
}

func Openx(dataSourceName string) (*sqlx.DB, error) {
	db, err := Open(dataSourceName)
	if err != nil {
		return nil, err
	}
	return sqlx.NewDb(db, driver), nil
}

func NewDBx(db *sql.DB) *sqlx.DB {
	return sqlx.NewDb(db, driver)
}
