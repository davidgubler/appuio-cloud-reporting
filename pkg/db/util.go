package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
)

// NamedPreparer is an interface used by GetNamed.
type NamedPreparer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

// NamedPreparerContext is an interface used by GetNamedContext.
type NamedPreparerContext interface {
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
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

// GetNamedContext is like sqlx.GetContext but for named statements.
func GetNamedContext(ctx context.Context, p NamedPreparerContext, dest interface{}, query string, arg interface{}) error {
	st, err := p.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer st.Close()
	return st.GetContext(ctx, dest, arg)
}

// InfiniteRange returns an infinite PostgreSQL timerange [-Inf,Inf).
func InfiniteRange() pgtype.Tstzrange {
	return Timerange(MustTimestamp(pgtype.NegativeInfinity), MustTimestamp(pgtype.Infinity))
}
