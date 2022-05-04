package invoice_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type InvoiceIntegrationSuite struct {
	dbtest.Suite
}

func TestInvoiceIntegration(t *testing.T) {
	suite.Run(t, new(InvoiceIntegrationSuite))
}

func (s *InvoiceIntegrationSuite) TestInvoiceIntegration_Simple() {
	t := s.T()
	tdb := s.DB()

	_, err := db.CreateProduct(tdb, db.Product{
		Source: "my-product",
		Amount: 1,
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)
	_, err = db.CreateDiscount(tdb, db.Discount{
		Source: "my-product",
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)
	q, err := db.CreateQuery(tdb, db.Query{
		Name:        "test",
		Description: "test description",
		Query:       "test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	require.NoError(t, err)
	sq, err := db.CreateQuery(tdb, db.Query{
		ParentID: sql.NullString{
			String: q.Id,
			Valid:  true,
		},
		Name:        "sub-test",
		Description: "A sub query of Test",
		Query:       "test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	sq2, err := db.CreateQuery(tdb, db.Query{
		ParentID: sql.NullString{
			String: q.Id,
			Valid:  true,
		},
		Name:        "sub-test2",
		Description: "An other sub query of Test",
		Query:       "test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})

	require.NoError(t, err)

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now().Truncate(time.Hour)
	require.NoError(t, report.Run(context.Background(), tx, fakeQuerrier{}, q.Name, ts))
	require.NoError(t, report.Run(context.Background(), tx, fakeQuerrier{}, sq.Name, ts))
	require.NoError(t, report.Run(context.Background(), tx, fakeQuerrier{}, sq2.Name, ts))

	invRun, err := invoice.Generate(context.Background(), tx, ts.Year(), ts.Month())
	require.NoError(t, err)
	require.Len(t, invRun, 1)
}

type fakeQuerrier struct{}

func (q fakeQuerrier) Query(ctx context.Context, query string, ts time.Time) (model.Value, apiv1.Warnings, error) {
	return model.Vector{
		{
			Metric: map[model.LabelName]model.LabelValue{
				"product":  "my-product:my-cluster:my-tenant:my-namespace",
				"category": "my-cluster:my-namespace",
				"tenant":   "my-tenant",
			},
			Value:     1337,
			Timestamp: 0,
		},
	}, apiv1.Warnings{}, nil
}
