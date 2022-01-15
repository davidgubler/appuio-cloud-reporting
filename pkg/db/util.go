package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type NamedPreparer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

func GetNamed(p NamedPreparer, dest interface{}, query string, arg interface{}) error {
	st, err := p.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer st.Close()
	return st.Get(dest, arg)
}
