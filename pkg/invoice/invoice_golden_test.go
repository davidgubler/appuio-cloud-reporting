package invoice_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type InvoiceGoldenSuite struct {
	dbtest.Suite

	prom fakeQuerrier
}

func (s *InvoiceGoldenSuite) SetupTest() {
	s.prom = fakeQuerrier{
		queries: map[string]fakeQueryResults{},
	}
	t := s.T()
	_, err := s.DB().Exec("TRUNCATE queries, date_times, facts, tenants, categories, products, discounts RESTART IDENTITY;")
	require.NoError(t, err)
}

func TestInvoiceIntegration(t *testing.T) {
	suite.Run(t, new(InvoiceGoldenSuite))
}

const dayLayout = "2006-01-02"

func runReport(t *testing.T, tdb *sqlx.DB, prom report.PromQuerier, queryName string, from, until string) {
	start, err := time.Parse(dayLayout, from)
	require.NoError(t, err)
	end, err := time.Parse(dayLayout, until)
	require.NoError(t, err)
	_, err = report.RunRange(context.Background(), tdb, prom, queryName, start, end)
	require.NoError(t, err)
}
func generateInvoice(t *testing.T, tdb *sqlx.DB, year int, month time.Month) []invoice.Invoice {
	tx, err := tdb.Beginx()
	require.NoError(t, err)
	defer tx.Rollback()
	invRun, err := invoice.Generate(context.Background(), tx, year, month)
	require.NoError(t, err)
	return invRun
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
