package db_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type UtilTestSuite struct {
	dbtest.Suite
}

func (s *UtilTestSuite) TestGetNamed() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	query := "SELECT :q"
	expected := "ping"
	namedParam := map[string]interface{}{"q": expected}

	var res string
	require.NoError(t, db.GetNamed(tx, &res, query, namedParam))
	require.Equal(t, expected, res)

	require.NoError(t, db.GetNamedContext(context.Background(), tx, &res, query, namedParam))
	require.Equal(t, expected, res)

	require.Error(t, db.GetNamed(tx, &res, "invalid", namedParam))
	require.Error(t, db.GetNamedContext(context.Background(), tx, &res, "invalid", namedParam))
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}
