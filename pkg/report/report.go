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
	var query db.Query
	err := sqlx.Get(dbx, &query, "SELECT * FROM queries WHERE name = $1 AND (during @> $2::timestamptz)", queryName, ts)
	if err != nil {
		return err
	}

	res, warn, err := prom.Query(context.TODO(), query.Query, ts)
	if err != nil {
		return err
	}

	fmt.Println(res, warn)

	return nil
}
