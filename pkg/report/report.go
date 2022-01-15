package report

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/jmoiron/sqlx"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PromQuerier interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, apiv1.Warnings, error)
}

func Run(dbx *sqlx.DB, prom PromQuerier, queryName string, ts time.Time) error {
	tx, err := dbx.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var query db.Query
	if err := sqlx.Get(tx, &query, "SELECT * FROM queries WHERE name = $1 AND (during @> $2::timestamptz)", queryName, ts); err != nil {
		return err
	}

	res, _, err := prom.Query(context.TODO(), query.Query, ts)
	if err != nil {
		return err
	}

	samples, ok := res.(model.Vector)
	if !ok {
		return fmt.Errorf("expected prometheus query to return a model.Vector, got %T", res)
	}

	for _, sample := range samples {
		if err := processSample(tx, ts, query, sample); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func processSample(p interface {
	db.NamedPreparer
	sqlx.Queryer
}, ts time.Time, query db.Query, s *model.Sample) error {
	category, err := getMetricLabel(s.Metric, "category")
	if err != nil {
		return err
	}
	tenant, err := getMetricLabel(s.Metric, "tenant")
	if err != nil {
		return err
	}
	productLabel, err := getMetricLabel(s.Metric, "product")
	if err != nil {
		return err
	}

	var upsertedTenant db.Tenant
	err = db.GetNamed(p, &upsertedTenant,
		"INSERT INTO tenants (source) VALUES (:source) ON CONFLICT (source) DO UPDATE SET source = :source RETURNING *", db.Tenant{
			Source: string(tenant),
		})
	if err != nil {
		return fmt.Errorf("failed to upsert tenant '%s': %w", tenant, err)
	}

	var upsertedCategory db.Category
	err = db.GetNamed(p, &upsertedCategory,
		"INSERT INTO categories (source) VALUES (:source) ON CONFLICT (source) DO UPDATE SET source = :source RETURNING *", db.Category{
			Source: string(category),
		})
	if err != nil {
		return fmt.Errorf("failed to upsert category '%s': %w", tenant, err)
	}

	var product db.Product
	err = sqlx.Get(p, &product,
		"SELECT * FROM products WHERE starts_with($1,source) AND (during @> $2::timestamptz)",
		string(productLabel), ts,
	)
	if err != nil {
		return err
	}

	var discount db.Discount
	err = sqlx.Get(p, &discount,
		"SELECT * FROM discounts WHERE starts_with($1,source) AND (during @> $2::timestamptz)",
		string(productLabel), ts,
	)
	if err != nil {
		return err
	}

	var upsertedDateTime db.DateTime
	err = db.GetNamed(p, &upsertedDateTime,
		"INSERT INTO date_times (timestamp,year,month,day,hour) VALUES (:timestamp,:year,:month,:day,:hour) ON CONFLICT (year,month,day,hour) DO UPDATE SET timestamp = :timestamp RETURNING *", db.DateTime{
			Timestamp: ts,
			Year:      ts.Year(),
			Month:     int(ts.Month()),
			Day:       ts.Day(),
			Hour:      ts.Hour(),
		})
	if err != nil {
		return err
	}

	var upsertedFact db.Fact
	err = db.GetNamed(p, &upsertedFact,
		"INSERT INTO facts (date_time_id,query_id,tenant_id,category_id,product_id,discount_id,quantity) VALUES (:date_time_id,:query_id,:tenant_id,:category_id,:product_id,:discount_id,:quantity) ON CONFLICT (date_time_id,query_id,tenant_id,category_id,product_id,discount_id) DO UPDATE SET quantity = :quantity RETURNING *", db.Fact{
			DateTimeId: upsertedDateTime.Id,
			TenantId:   upsertedTenant.Id,
			CategoryId: upsertedCategory.Id,
			QueryId:    query.Id,
			ProductId:  product.Id,
			DiscountId: discount.Id,
			Quantity:   float64(s.Value),
		})
	if err != nil {
		return err
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
