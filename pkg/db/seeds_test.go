package db_test

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type SeedsTestSuite struct {
	dbtest.Suite
}

func (s *SeedsTestSuite) TestSeedDefaultQueries() {
	t := s.T()
	d := s.DB()

	_, err := d.Exec("DELETE FROM queries")
	require.NoError(t, err)

	expQueryNum := 5

	count := "SELECT ((SELECT COUNT(*) FROM queries) = $1)"
	requireQueryTrue(t, d, count, 0)
	err = db.Seed(d.DB)
	require.NoError(t, err)
	requireQueryTrue(t, d, count, expQueryNum)
	err = db.Seed(d.DB)
	require.NoError(t, err)
	requireQueryTrue(t, d, count, expQueryNum)
}

func requireQueryTrue(t *testing.T, q sqlx.Queryer, query string, args ...interface{}) {
	t.Helper()

	var result bool
	err := sqlx.Get(q, &result, query, args...)
	require.NoError(t, err)
	require.True(t, result)
}

func TestSeeds(t *testing.T) {
	suite.Run(t, new(SeedsTestSuite))
}
