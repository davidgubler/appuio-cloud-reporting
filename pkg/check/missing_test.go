package check_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/check"
	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type TestSuite struct {
	dbtest.Suite
}

func (s *TestSuite) TestMissingFields() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	m, err := check.Missing(context.Background(), tx)
	require.NoError(t, err)
	require.Len(t, m, 0)

	expectedMissing := s.requireMissingTestEntries(t, tx)

	m, err = check.Missing(context.Background(), tx)
	require.NoError(t, err)
	require.Equal(t, expectedMissing, m)
}

func (s *TestSuite) requireMissingTestEntries(t *testing.T, tdb *sqlx.Tx) []check.MissingField {
	var catEmptyTarget db.Category
	require.NoError(t,
		db.GetNamed(tdb, &catEmptyTarget,
			"INSERT INTO categories (source,target) VALUES (:source,:target) RETURNING *", db.Category{
				Source: "af-south-1:uroboros-research",
			}))

	var tenantEmptyTarget db.Tenant
	require.NoError(t,
		db.GetNamed(tdb, &tenantEmptyTarget,
			"INSERT INTO tenants (source,target) VALUES (:source,:target) RETURNING *", db.Tenant{
				Source: "tricell",
			}))

	var productEmptyTarget db.Product
	require.NoError(t,
		db.GetNamed(tdb, &productEmptyTarget,
			"INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) RETURNING *", db.Product{
				Source: "test_memory:us-rac-2",
				Amount: 3,
				Unit:   "X",
				During: db.InfiniteRange(),
			}))

	var productEmptyAmountAndUnit db.Product
	require.NoError(t,
		db.GetNamed(tdb, &productEmptyAmountAndUnit,
			"INSERT INTO products (source,target,amount,unit,during) VALUES (:source,:target,:amount,:unit,:during) RETURNING *", db.Product{
				Source: "test_storage:us-rac-2",
				Target: sql.NullString{Valid: true, String: "666"},
				During: db.InfiniteRange(),
			}))

	return []check.MissingField{
		{Table: "categories", MissingField: "target", ID: catEmptyTarget.Id, Source: catEmptyTarget.Source},
		{Table: "products", MissingField: "amount", ID: productEmptyAmountAndUnit.Id, Source: productEmptyAmountAndUnit.Source},
		{Table: "products", MissingField: "target", ID: productEmptyTarget.Id, Source: productEmptyTarget.Source},
		{Table: "products", MissingField: "unit", ID: productEmptyAmountAndUnit.Id, Source: productEmptyAmountAndUnit.Source},
		{Table: "tenants", MissingField: "target", ID: tenantEmptyTarget.Id, Source: tenantEmptyTarget.Source},
	}
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
