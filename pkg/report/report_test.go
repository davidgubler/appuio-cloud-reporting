package report_test

import (
	"context"
	"database/sql"
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

const defaultQueryReturnValue = 42
const defaultSubQueryReturnValue = 13
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
const promBarTestquery = `
	label_replace(
		label_replace(
			label_replace(
				vector(%d),
				"category", "my-cluster:my-namespace", "", ""
			),
			"product", "bar-product:my-cluster:my-tenant:my-namespace", "", ""
		),
		"tenant", "my-tenant", "", ""
	)
`

func (s *ReportSuite) SetupSuite() {
	s.Suite.SetupSuite()

	t := s.T()
	tdb := s.DB()

	s.sampleProduct = s.createProduct(tdb, "my-product:my-cluster")

	require.NoError(t,
		db.GetNamed(tdb, &s.sampleDiscount,
			"INSERT INTO discounts (source,discount,during) VALUES (:source,:discount,:during) RETURNING *", db.Discount{
				Source:   "my-product:my-cluster",
				Discount: 0.5,
				During:   infiniteRange(),
			}))

	require.NoError(t,
		db.GetNamed(tdb, &s.sampleQuery,
			"INSERT INTO queries (name,description,query,unit,during) VALUES (:name,:description,:query,:unit,:during) RETURNING *", db.Query{
				Name:   "test",
				Query:  fmt.Sprintf(promTestquery, defaultQueryReturnValue),
				Unit:   "tps",
				During: infiniteRange(),
			}))

}

func (s *ReportSuite) TestReport_ReturnsErrorIfTimestampContainsUnitsSmallerOneHour() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	baseTime := time.Date(2020, time.January, 23, 17, 0, 0, 0, time.UTC)
	for _, d := range []time.Duration{time.Minute, time.Second, time.Nanosecond} {
		require.Error(t, report.Run(context.Background(), tx, prom, query.Name, baseTime.Add(d)))
	}
}

func (s *ReportSuite) TestReport_RunRange() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery
	tdb := s.DB()

	const hoursToCalculate = 3

	defer tdb.Exec("DELETE FROM facts")

	base := time.Date(2020, time.January, 23, 17, 0, 0, 0, time.UTC)

	expectedProgress := []report.Progress{
		{base.Add(0 * time.Hour), 1},
		{base.Add(1 * time.Hour), 2},
		{base.Add(2 * time.Hour), 3},
	}

	progress := make([]report.Progress, 0)

	c, err := report.RunRange(context.Background(), tdb, prom, query.Name, base, base.Add(hoursToCalculate*time.Hour),
		report.WithProgressReporter(func(p report.Progress) { progress = append(progress, p) }),
	)
	require.NoError(t, err)
	require.Equal(t, hoursToCalculate, c)
	require.Equal(t, expectedProgress, progress)

	var factCount int
	require.NoError(t, sqlx.Get(tdb, &factCount, "SELECT COUNT(*) FROM facts"))
	require.Equal(t, hoursToCalculate, factCount)
}

func (s *ReportSuite) TestReport_RunReportCreatesFact() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now().Truncate(time.Hour)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts, report.WithPrometheusQueryTimeout(time.Second)))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster", ts)
	require.Equal(t, float64(defaultQueryReturnValue), fact.Quantity)
}

func (s *ReportSuite) TestReport_RunReportCreatesSubFact() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	tdb := s.DB()
	s.createProduct(tdb, "bar-product:my-cluster")
	disc := db.Discount{}
	require.NoError(t,
		db.GetNamed(tdb, &disc,
			"INSERT INTO discounts (source,discount,during) VALUES (:source,:discount,:during) RETURNING *", db.Discount{
				Source:   "bar-product",
				Discount: 0,
				During:   infiniteRange(),
			}))
	query := db.Query{}
	require.NoError(t,
		db.GetNamed(tdb, &query,
			"INSERT INTO queries (name,description,query,unit,during) VALUES (:name,:description,:query,:unit,:during) RETURNING *", db.Query{
				Name:   "bar",
				Query:  fmt.Sprintf(promBarTestquery, defaultQueryReturnValue),
				Unit:   "tps",
				During: infiniteRange(),
			}))
	subquery := db.Query{}
	require.NoError(t,
		db.GetNamed(tdb, &subquery,
			"INSERT INTO queries (parent_id,name,description,query,unit,during) VALUES (:parent_id,:name,:description,:query,:unit,:during) RETURNING *", db.Query{
				ParentID: sql.NullString{
					String: query.Id,
					Valid:  true,
				},
				Name:   "sub-bar",
				Query:  fmt.Sprintf(promBarTestquery, defaultSubQueryReturnValue),
				Unit:   "tps",
				During: infiniteRange(),
			}))

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now().Truncate(time.Hour)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts, report.WithPrometheusQueryTimeout(time.Second)))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "bar-product:my-cluster", ts)
	require.Equal(t, float64(defaultQueryReturnValue), fact.Quantity)
	subfact := s.requireFactForQueryIdAndProductSource(tx, subquery, "bar-product:my-cluster", ts)
	require.Equal(t, float64(defaultSubQueryReturnValue), subfact.Quantity)
}

func (s *ReportSuite) TestReport_RunReportNonLockingUpsert() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	defer s.DB().Exec("DELETE FROM facts")

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	tx1, err := s.DB().BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx1.Rollback()

	tx2, err := s.DB().BeginTxx(ctx, nil)
	require.NoError(t, err)
	defer tx2.Rollback()

	baseTime := time.Date(2020, time.January, 23, 17, 0, 0, 0, time.UTC)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, baseTime))
	require.NoError(t, tx.Commit())

	require.NoError(t, report.Run(context.Background(), tx1, prom, query.Name, baseTime.Add(1*time.Hour)), "transaction should not be blocked on upsert")
	require.NoError(t, report.Run(context.Background(), tx2, prom, query.Name, baseTime.Add(2*time.Hour)), "transaction should not be blocked on upsert")

	require.NoError(t, tx2.Commit(), "transaction should not be blocked on commit")
	require.NoError(t, tx1.Commit(), "transaction should not be blocked on commit")
}

func (s *ReportSuite) TestReport_RerunReportUpdatesFactQuantity() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now().Truncate(time.Hour)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts))

	_, err = tx.Exec("UPDATE queries SET query = $1 WHERE id = $2", fmt.Sprintf(promTestquery, 77), query.Id)
	require.NoError(t, err)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster", ts)
	require.Equal(t, float64(77), fact.Quantity)
}

func (s *ReportSuite) TestReport_ProductSpecificityOfSource() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now().Truncate(time.Hour)
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts))
	s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster", ts)

	wildcardProduct := s.createProduct(tx, "my-product:*:my-tenant")
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "my-product:*:my-tenant", ts)
	require.Equal(t, wildcardProduct.Id, fact.ProductId)

	specificProduct := s.createProduct(tx, "my-product:my-cluster:my-tenant:my-namespace")
	require.NoError(t, report.Run(context.Background(), tx, prom, query.Name, ts))
	fact = s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster:my-tenant:my-namespace", ts)
	require.Equal(t, specificProduct.Id, fact.ProductId)
}

func TestReport(t *testing.T) {
	suite.Run(t, new(ReportSuite))
}

func (s *ReportSuite) createProduct(p db.NamedPreparer, source string) db.Product {
	var product db.Product
	require.NoError(s.T(),
		db.GetNamed(p, &product,
			"INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) RETURNING *", db.Product{
				Source: source,
				Amount: 1,
				Unit:   "tps",
				During: infiniteRange(),
			}))

	return product
}

func (s *ReportSuite) requireFactForQueryIdAndProductSource(dbq sqlx.Queryer, q db.Query, productSource string, ts time.Time) db.Fact {
	var fact db.Fact
	require.NoError(
		s.T(),
		sqlx.Get(
			dbq, &fact,
			"SELECT facts.* FROM facts INNER JOIN products ON (facts.product_id = products.id) WHERE facts.query_id = $1 AND products.source = $2",
			q.Id, productSource))
	return fact
}

func infiniteRange() pgtype.Tstzrange {
	return db.Timerange(db.MustTimestamp(pgtype.NegativeInfinity), db.MustTimestamp(pgtype.Infinity))
}
