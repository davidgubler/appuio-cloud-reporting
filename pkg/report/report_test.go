package report_test

import (
	"context"
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

	s.sampleProduct = s.createProduct(tdb, "my-product:my-cluster")

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
				Query:  fmt.Sprintf(promTestquery, defaultQueryReturnValue),
				Unit:   "tps",
				During: infiniteRange(),
			}))
}

func (s *ReportSuite) TestReport_RunReportCreatesFact() {
	t := s.T()
	prom := s.PrometheusAPIClient()
	query := s.sampleQuery

	tx, err := s.DB().Beginx()
	require.NoError(t, err)
	defer tx.Rollback()

	ts := time.Now()
	require.NoError(t, report.Run(tx, prom, query.Name, ts))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster", ts)
	require.Equal(t, float64(defaultQueryReturnValue), fact.Quantity)
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
	require.NoError(t, report.Run(tx, prom, query.Name, baseTime))
	require.NoError(t, tx.Commit())

	require.NoError(t, report.Run(tx1, prom, query.Name, baseTime.Add(1*time.Hour)), "transaction should not be blocked on upsert")
	require.NoError(t, report.Run(tx2, prom, query.Name, baseTime.Add(2*time.Hour)), "transaction should not be blocked on upsert")

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

	ts := time.Now()
	require.NoError(t, report.Run(tx, prom, query.Name, ts))

	_, err = tx.Exec("UPDATE queries SET query = $1 WHERE id = $2", fmt.Sprintf(promTestquery, 77), query.Id)
	require.NoError(t, err)
	require.NoError(t, report.Run(tx, prom, query.Name, ts))
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

	ts := time.Now()
	require.NoError(t, report.Run(tx, prom, query.Name, ts))
	s.requireFactForQueryIdAndProductSource(tx, query, "my-product:my-cluster", ts)

	wildcardProduct := s.createProduct(tx, "my-product:*:my-tenant")
	require.NoError(t, report.Run(tx, prom, query.Name, ts))
	fact := s.requireFactForQueryIdAndProductSource(tx, query, "my-product:*:my-tenant", ts)
	require.Equal(t, wildcardProduct.Id, fact.ProductId)

	specificProduct := s.createProduct(tx, "my-product:my-cluster:my-tenant:my-namespace")
	require.NoError(t, report.Run(tx, prom, query.Name, ts))
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
