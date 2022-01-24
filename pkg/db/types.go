package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

type Query struct {
	Id string

	Name        string
	Description string
	Query       string
	Unit        string

	During pgtype.Tstzrange
}

type Tenant struct {
	Id string

	Source string
	Target sql.NullString
}

type Category struct {
	Id string

	Source string
	Target sql.NullString
}

type Product struct {
	Id string

	Source string
	Target sql.NullString
	Amount float64
	Unit   string

	During pgtype.Tstzrange
}

type Discount struct {
	Id string

	Source   string
	Discount int

	During pgtype.Tstzrange
}

type DateTime struct {
	Id string

	Timestamp time.Time

	Year  int
	Month int
	Day   int
	Hour  int
}

type Fact struct {
	Id string

	DateTimeId string `db:"date_time_id"`
	QueryId    string `db:"query_id"`
	TenantId   string `db:"tenant_id"`
	CategoryId string `db:"category_id"`
	ProductId  string `db:"product_id"`
	DiscountId string `db:"discount_id"`

	Quantity float64
}

// Timestamp creates a Postgres timestamp from the given value.
// Valid values are nil, pgtype.Infinity/pgtype.NegativeInfinity, and a time.Time object.
func Timestamp(from interface{}) (pgtype.Timestamptz, error) {
	ts := pgtype.Timestamptz{}
	err := ts.Set(from)
	return ts, err
}

// MustTimestamp creates a Postgres timestamp from the given value.
// Valid values are nil, pgtype.Infinity/pgtype.NegativeInfinity, and a time.Time object.
// Panics if given an unsupported type.
func MustTimestamp(from interface{}) pgtype.Timestamptz {
	ts, err := Timestamp(from)
	if err != nil {
		panic(fmt.Errorf("expected to create valid timestamp: %s", err))
	}
	return ts
}

// Timerange creates a Postgres timerange from two Postgres timestamps with [lower,upper) bounds.
func Timerange(lower, upper pgtype.Timestamptz) pgtype.Tstzrange {
	return pgtype.Tstzrange{
		Lower:     lower,
		LowerType: pgtype.Inclusive,
		Upper:     upper,
		UpperType: pgtype.Exclusive,
		Status:    pgtype.Present,
	}
}
