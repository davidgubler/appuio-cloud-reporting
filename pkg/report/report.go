package report

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/sourcekey"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PromQuerier interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, apiv1.Warnings, error)
}

// RunRange executes prometheus queries like Run() until the `until` timestamp is reached or an error occurred.
// Returns the number of reports run and a possible error.
func RunRange(db *sqlx.DB, prom PromQuerier, queryName string, from time.Time, until time.Time) (int, error) {
	n := 0
	for currentTime := from; until.After(currentTime); currentTime = currentTime.Add(time.Hour) {
		if err := inTransaction(db, func(tx *sqlx.Tx) error {
			return Run(tx, prom, queryName, currentTime)
		}); err != nil {
			return n, fmt.Errorf("error running report at %s: %w", currentTime.Format(time.RFC3339), err)
		}
		n++
	}

	return n, nil
}

// Run executes a prometheus query loaded from queries with using the `queryName` and the timestamp.
// The results of the query are saved in the facts table.
func Run(tx *sqlx.Tx, prom PromQuerier, queryName string, from time.Time) error {
	from = from.In(time.UTC)
	if !from.Truncate(time.Hour).Equal(from) {
		return fmt.Errorf("timestamp should only contain full hours based on UTC, got: %s", from.Format(time.RFC3339Nano))
	}

	var query db.Query
	if err := sqlx.Get(tx, &query, "SELECT * FROM queries WHERE name = $1 AND (during @> $2::timestamptz)", queryName, from); err != nil {
		return fmt.Errorf("failed to load query '%s' at '%s': %w", queryName, from.Format(time.RFC3339), err)
	}

	// The data in the database is from T to T+1h. Prometheus queries backwards from T to T-1h.
	res, _, err := prom.Query(context.TODO(), query.Query, from.Add(time.Hour))
	if err != nil {
		return fmt.Errorf("failed to query prometheus: %w", err)
	}

	samples, ok := res.(model.Vector)
	if !ok {
		return fmt.Errorf("expected prometheus query to return a model.Vector, got %T", res)
	}

	for _, sample := range samples {
		if err := processSample(tx, from, query, sample); err != nil {
			return fmt.Errorf("failed to process sample: %w", err)
		}
	}

	return nil
}

func processSample(tx *sqlx.Tx, ts time.Time, query db.Query, s *model.Sample) error {
	category, err := getMetricLabel(s.Metric, "category")
	if err != nil {
		return err
	}
	productLabel, err := getMetricLabel(s.Metric, "product")
	if err != nil {
		return err
	}

	skey, err := sourcekey.Parse(string(productLabel))
	if err != nil {
		return fmt.Errorf("failed to parse source key from product label: %w", err)
	}

	var upsertedTenant db.Tenant
	if upsertTenant(tx, &upsertedTenant, db.Tenant{Source: skey.Tenant}); err != nil {
		return err
	}

	var upsertedCategory db.Category
	if err := upsertCategory(tx, &upsertedCategory, db.Category{Source: string(category)}); err != nil {
		return err
	}

	sourceLookup := skey.LookupKeys()

	var product db.Product
	if err := getBySourceKeyAndTime(tx, &product, pgx.Identifier{"products"}, sourceLookup, ts); err != nil {
		return fmt.Errorf("failed to load product for '%s': %w", productLabel, err)
	}

	var discount db.Discount
	if err := getBySourceKeyAndTime(tx, &discount, pgx.Identifier{"discounts"}, sourceLookup, ts); err != nil {
		return fmt.Errorf("failed to load discount for '%s': %w", productLabel, err)
	}

	var upsertedDateTime db.DateTime
	err = upsertDateTime(tx, &upsertedDateTime, db.DateTime{
		Timestamp: ts,
		Year:      ts.Year(),
		Month:     int(ts.Month()),
		Day:       ts.Day(),
		Hour:      ts.Hour(),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert date_time '%s': %w", ts.Format(time.RFC3339), err)
	}

	var upsertedFact db.Fact
	err = upsertFact(tx, &upsertedFact, db.Fact{
		DateTimeId: upsertedDateTime.Id,
		TenantId:   upsertedTenant.Id,
		CategoryId: upsertedCategory.Id,
		QueryId:    query.Id,
		ProductId:  product.Id,
		DiscountId: discount.Id,
		Quantity:   float64(s.Value),
	})
	if err != nil {
		return fmt.Errorf("failed to upsert fact '%s': %w", ts.Format(time.RFC3339), err)
	}

	return nil
}

// getBySourceKeyAndTime gets the first record matching a key in keys while preserving the priority or order of the keys.
// The first key has the highest priority while the last key has the lowest priority.
// If keys are [a,b,c] and records [a,c] exist a is chosen.
func getBySourceKeyAndTime(q sqlx.Queryer, dest interface{}, table pgx.Identifier, keys []string, ts time.Time) error {
	const query = `WITH keys AS (
		-- add a priority to keep track of which key match we should choose
		-- first key -> prio 1, third key -> prio 3
		SELECT row_number() over () AS prio, unnest as key
		-- unpack the given array of strings into rows
		FROM unnest($1::text[])
	)
	SELECT {{table}}.*
		FROM {{table}}
		INNER JOIN keys ON (keys.key = {{table}}.source)
		WHERE during @> $2::timestamptz
		ORDER BY prio
		LIMIT 1`
	return sqlx.Get(q, dest, strings.ReplaceAll(query, "{{table}}", table.Sanitize()), keys, ts)
}

func upsertFact(tx *sqlx.Tx, dst *db.Fact, src db.Fact) error {
	err := db.GetNamed(tx, dst,
		`INSERT INTO facts
				(date_time_id,query_id,tenant_id,category_id,product_id,discount_id,quantity)
			VALUES
				(:date_time_id,:query_id,:tenant_id,:category_id,:product_id,:discount_id,:quantity)
			ON CONFLICT (date_time_id,query_id,tenant_id,category_id,product_id,discount_id)
				DO UPDATE SET quantity = :quantity
			RETURNING *`,
		src)
	if err != nil {
		return fmt.Errorf("failed to upsert fact %+v: %w", src, err)
	}
	return nil
}

func upsertCategory(tx *sqlx.Tx, dst *db.Category, src db.Category) error {
	err := db.GetNamed(tx, dst,
		`WITH
				existing AS (
					SELECT * FROM categories WHERE source = :source
				),
				inserted AS (
					INSERT INTO categories (source)
					SELECT :source WHERE NOT EXISTS (SELECT 1 FROM existing)
					RETURNING *
				)
			SELECT * FROM inserted UNION ALL SELECT * FROM existing`,
		src)
	if err != nil {
		return fmt.Errorf("failed to upsert category %+v: %w", src, err)
	}
	return nil
}

func upsertTenant(tx *sqlx.Tx, dst *db.Tenant, src db.Tenant) error {
	err := db.GetNamed(tx, dst,
		`WITH
				existing AS (
					SELECT * FROM tenants WHERE source = :source
				),
				inserted AS (
					INSERT INTO tenants (source)
					SELECT :source WHERE NOT EXISTS (SELECT 1 FROM existing)
					RETURNING *
				)
			SELECT * FROM inserted UNION ALL SELECT * FROM existing`,
		src)
	if err != nil {
		return fmt.Errorf("failed to upsert tenant %+v: %w", src, err)
	}
	return nil
}

func upsertDateTime(tx *sqlx.Tx, dst *db.DateTime, src db.DateTime) error {
	err := db.GetNamed(tx, dst,
		`WITH
		existing AS (
			SELECT * FROM date_times WHERE year = :year AND month = :month AND day = :day AND hour = :hour
		),
		inserted AS (
			INSERT INTO date_times (timestamp, year, month, day, hour)
			SELECT :timestamp, :year, :month, :day, :hour WHERE NOT EXISTS (SELECT 1 FROM existing)
			RETURNING *
		)
		SELECT * FROM inserted UNION ALL SELECT * FROM existing`,
		src)
	if err != nil {
		return fmt.Errorf("failed to upsert date_time %+v: %w", src, err)
	}
	return nil
}

func getMetricLabel(m model.Metric, name string) (model.LabelValue, error) {
	value, ok := m[model.LabelName(name)]
	if !ok {
		return "", fmt.Errorf("expected sample to contain label '%s'", name)
	}
	return value, nil
}

func inTransaction(db *sqlx.DB, cb func(tx *sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback()

	if err := cb(tx); err != nil {
		return err
	}
	return tx.Commit()
}
