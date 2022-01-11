package dbtest_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type TestSuite struct {
	dbtest.Suite
}

func (s *TestSuite) TestTxQuery() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	_, err := tx.Exec("SELECT 1")
	require.NoError(t, err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
