package report

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type PromQuerier interface {
	Query(ctx context.Context, query string, ts time.Time) (model.Value, apiv1.Warnings, error)
}

func Run(db *sqlx.DB, prom PromQuerier, queryName string, ts time.Time) error {
	return nil
}
