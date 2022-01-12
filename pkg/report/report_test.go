package report_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/report"
	"github.com/appuio/appuio-cloud-reporting/pkg/testsuite"
)

type ReportSuite struct {
	testsuite.Suite
}

const promTestquery = `
	label_replace(
		label_replace(
			label_replace(
				vector(42),
				"category", "my-cluster:my-namespace", "", ""
			),
			"product", "my-product:my-cluster:my-tenant:my-namespace", "", ""
		),
		"tenant", "my-tenant", "", ""
	)
`

func (s *ReportSuite) TestReport() {
	t := s.T()
	tdb := s.DB()
	prom := s.PrometheusAPIClient()

	var product db.Product
	require.NoError(t,
		getNamed(tdb, &product,
			"INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) RETURNING *", db.Product{
				Source: "my-product:my-cluster",
				Amount: 1,
				Unit:   "tps",
				During: infiniteRange(),
			}))

	var discount db.Discount
	require.NoError(t,
		getNamed(tdb, &discount,
			"INSERT INTO discounts (source,discount,during) VALUES (:source,:discount,:during) RETURNING *", db.Discount{
				Source:   "my-product:my-cluster",
				Discount: 50,
				During:   infiniteRange(),
			}))

	var query db.Query
	require.NoError(t,
		getNamed(tdb, &query,
			"INSERT INTO queries (name,description,query,unit,during) VALUES (:name,:description,:query,:unit,:during) RETURNING *", db.Query{
				Name:   "test",
				Query:  promTestquery,
				Unit:   "tps",
				During: infiniteRange(),
			}))

	t.Log("Created", product, discount, query)

	require.NoError(t, report.Run(tdb, prom, query.Name, time.Now()))
}

type namedPreparer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
}

func getNamed(p namedPreparer, dest interface{}, query string, arg interface{}) error {
	st, err := p.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer st.Close()
	return st.Get(dest, arg)
}

func TestReport(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

func infiniteRange() pgtype.Tstzrange {
	return db.Timerange(db.MustTimestamp(pgtype.NegativeInfinity), db.MustTimestamp(pgtype.Infinity))
}
