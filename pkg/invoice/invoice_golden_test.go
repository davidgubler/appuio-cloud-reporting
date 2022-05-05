package invoice_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvoiceGoldenSuite struct {
	dbtest.Suite

	prom fakeQuerrier
}

func (s *InvoiceGoldenSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.prom = fakeQuerrier{
		queries: map[string]fakeQueryResults{},
	}
}

func TestInvoiceIntegration(t *testing.T) {
	suite.Run(t, new(InvoiceGoldenSuite))
}

const dayLayout = "2006-01-02"

func (s *InvoiceGoldenSuite) TestInvoiceGolden_Simple() {
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
	s.prom.queries[q.Query] = fakeQueryResults{
		"my-product:my-cluster:my-tenant:my-namespace": fakeQuerySample{Value: 42},
		"my-product:my-cluster:other-tenant:my-namespace": fakeQuerySample{Value: 23},
	}
	require.NoError(t, err)
	sq, err := db.CreateQuery(tdb, db.Query{
		ParentID: sql.NullString{
			String: q.Id,
			Valid:  true,
		},
		Name:        "sub-test",
		Description: "A sub query of Test",
		Query:       "sub-test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	s.prom.queries[sq.Query] = fakeQueryResults{
		"my-product:my-cluster:my-tenant:my-namespace": fakeQuerySample{Value: 4},
		"my-product:my-cluster:other-tenant:my-namespace": fakeQuerySample{Value: 2},
	}
	sq2, err := db.CreateQuery(tdb, db.Query{
		ParentID: sql.NullString{
			String: q.Id,
			Valid:  true,
		},
		Name:        "sub-test2",
		Description: "An other sub query of Test",
		Query:       "sub-test2",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	s.prom.queries[sq2.Query] = fakeQueryResults{
		"my-product:my-cluster:my-tenant:my-namespace": fakeQuerySample{Value: 7},
		"my-product:my-cluster:other-tenant:my-namespace": fakeQuerySample{Value: 0},
	}

	require.NoError(t, err)

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	start, err := time.Parse(dayLayout, "2022-03-01")
	require.NoError(t, err)
	end, err := time.Parse(dayLayout, "2022-04-01")
	require.NoError(t, err)
	n, err := report.RunRange(context.Background(), tdb, s.prom, q.Name, start, end)
	require.NoError(t, err)
	require.Greater(t, n, 0)

	invRun, err := invoice.Generate(context.Background(), tx, start.Year(), start.Month())
	require.NoError(t, err)
	invoiceEqualsGolden(t, "simple", invRun, *updateGolden)
}

func invoiceEqualsGolden(t *testing.T, goldenFile string, actual []invoice.Invoice, update bool) {
	t.Run(goldenFile, func(t *testing.T) {
		actualJSON, err := json.MarshalIndent(sortInvoices(actual), "", "\t")
		require.NoErrorf(t, err, "Failed to marshal invoice to JSON")

		goldenPath := path.Join("testdata", fmt.Sprintf("%s.json", goldenFile))
		if update {
			os.WriteFile(goldenPath, actualJSON, 0644)
			require.NoErrorf(t, err, "failed to update goldenFile %s", goldenPath)
			return
		}

		f, err := os.OpenFile(goldenPath, os.O_RDONLY, 0644)
		defer f.Close()
		require.NoErrorf(t, err, "failed to open goldenFile %s", goldenPath)
		expected, err := io.ReadAll(f)
		require.NoErrorf(t, err, "failed to read goldenFile %s", goldenPath)

		assert.JSONEq(t, string(expected), string(actualJSON))
	})
}
