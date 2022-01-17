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

	sampleProduct  db.Product
	sampleDiscount db.Discount
	sampleQuery    db.Query
}

const promTestquery = `
	label_replace(
		label_replace(
			label_replace(
				vector(%d),
				"category", "my-cluster:my-namespace", "", ""
			),
			"product", "my-product:my-cluster:my-tenant:my-namespace", "", ""
		),
		"tenant", "my-tenant", "", ""
	)
`

func (s *ReportSuite) SetupSuite() {
	s.Suite.SetupSuite()

	t := s.T()
	tdb := s.DB()

	require.NoError(t,
		db.GetNamed(tdb, &s.sampleProduct,
			"INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) RETURNING *", db.Product{
				Source: "my-product:my-cluster",
				Amount: 1,
				Unit:   "tps",
				During: infiniteRange(),
			}))

	require.NoError(t,
		db.GetNamed(tdb, &s.sampleDiscount,
			"INSERT INTO discounts (source,discount,during) VALUES (:source,:discount,:during) RETURNING *", db.Discount{
				Source:   "my-product:my-cluster",
				Discount: 50,
				During:   infiniteRange(),
			}))

	require.NoError(t,
		db.GetNamed(tdb, &s.sampleQuery,
			"INSERT INTO queries (name,description,query,unit,during) VALUES (:name,:description,:query,:unit,:during) RETURNING *", db.Query{
				Name:   "test",
				Query:  fmt.Sprintf(promTestquery, 24),
				Unit:   "tps",
				During: infiniteRange(),
			}))
}

func (s *ReportSuite) TestReportCreateFact() {
	t := s.T()
	tdb := s.DB()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	ts := time.Now()
	require.NoError(t, report.Run(tdb, prom, query.Name, ts))
	fact := s.getFactForQueryIdAndProductSource(query, "my-product:my-cluster", ts)
	require.Equal(t, float64(24), fact.Quantity)

	_, err := tdb.Exec("UPDATE queries SET query = $1 WHERE id = $2", fmt.Sprintf(promTestquery, 77), query.Id)
	require.NoError(t, err)
	require.NoError(t, report.Run(tdb, prom, query.Name, ts))
	fact = s.getFactForQueryIdAndProductSource(query, "my-product:my-cluster", ts)
	require.Equal(t, float64(77), fact.Quantity)
}

func (s *ReportSuite) TestReportUpdateFact() {
	t := s.T()
	tdb := s.DB()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	ts := time.Now()
	require.NoError(t, report.Run(tdb, prom, query.Name, ts))

	_, err := tdb.Exec("UPDATE queries SET query = $1 WHERE id = $2", fmt.Sprintf(promTestquery, 77), query.Id)
	require.NoError(t, err)
	require.NoError(t, report.Run(tdb, prom, query.Name, ts))
	fact := s.getFactForQueryIdAndProductSource(query, "my-product:my-cluster", ts)
	require.Equal(t, float64(77), fact.Quantity)
}

func TestReport(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

func (s *ReportSuite) getFactForQueryIdAndProductSource(q db.Query, productSource string, ts time.Time) db.Fact {
	var fact db.Fact
	require.NoError(
		s.T(),
		sqlx.Get(
			s.DB(), &fact,
			"SELECT facts.* FROM facts INNER JOIN products ON (facts.product_id = products.id) WHERE facts.query_id = $1 AND products.source = $2",
			q.Id, productSource))
	return fact
}

func infiniteRange() pgtype.Tstzrange {
	return db.Timerange(db.MustTimestamp(pgtype.NegativeInfinity), db.MustTimestamp(pgtype.Infinity))
}