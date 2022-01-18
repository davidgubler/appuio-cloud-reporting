package db

import (
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
)

type NamedPreparer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

// GetNamed is like sqlx.Get but for named statements.
func GetNamed(p NamedPreparer, dest interface{}, query string, arg interface{}) error {
	st, err := p.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer st.Close()
	return st.Get(dest, arg)
}

// InfiniteRange returns an infinite PostgreSQL timerange [-Inf,Inf).
func InfiniteRange() pgtype.Tstzrange {
	return Timerange(MustTimestamp(pgtype.NegativeInfinity), MustTimestamp(pgtype.Infinity))
}
