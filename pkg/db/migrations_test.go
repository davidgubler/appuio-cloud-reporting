package db_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type MigrationTestSuite struct {
	dbtest.Suite
}

func (s *MigrationTestSuite) TestMigrations_DatabaseShouldBeFullyMigrated() {
	t := s.T()
	pending, err := db.Pending(s.DB().DB)
	require.NoError(t, err)
	require.Lenf(t, pending, 0, "the test database should be migrated to the newest version before running tests")
}

func TestMigrations(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}
