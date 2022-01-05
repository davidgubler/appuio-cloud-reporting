package types

import (
	"database/sql"
	"time"
)

type Query struct {
	Id string

	Name        string
	Description string
	Query       string
	Unit        string

	After  sql.NullTime
	Before sql.NullTime
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
	Amount int64
	Unit   string

	After  sql.NullTime
	Before sql.NullTime
}

type Discount struct {
	Id string

	Source   string
	Discount int

	After  sql.NullTime
	Before sql.NullTime
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
