package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

type Query struct {
	Id       string
	ParentID sql.NullString `db:"parent_id"`

	Name        string
	Description string
	Query       string
	Unit        string

	During pgtype.Tstzrange
}

type Tenant struct {
	Id string

	// Source is the tenant string read from the 'appuio.io/organization' label.
	Source string
	Target sql.NullString
}

type Category struct {
	Id string

	// Source consists of the cluster id and namespace in the form of "zone:namespace".
	Source string
	Target sql.NullString
}

type Product struct {
	Id string

	// Source is a string consisting of "query:zone:tenant:namespace:class" and can contain wildcards.
	// See package `sourcekey` for more information.
	Source string
	Target sql.NullString
	Amount float64
	Unit   string

	During pgtype.Tstzrange
}

type Discount struct {
	Id string

	// Source is a string consisting of "query:zone:tenant:namespace:class" and can contain wildcards.
	// See package `sourcekey` for more information.
	Source   string
	Discount float64

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

// BuildDateTime builds a DateTime object from the given timestamp.
func BuildDateTime(ts time.Time) DateTime {
	return DateTime{
		Timestamp: ts,

		Year:  ts.Year(),
		Month: int(ts.Month()),
		Day:   ts.Day(),
		Hour:  ts.Hour(),
	}
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
