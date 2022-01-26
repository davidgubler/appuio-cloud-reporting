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

func (s *UtilTestSuite) TestSelectNamed() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	type testTable struct{ Q string }
	_, err := tx.Exec("CREATE TEMPORARY TABLE t (q text)")
	require.NoError(t, err)

	query := "INSERT INTO t (q) VALUES (:q) RETURNING *"
	expected := []testTable{{"ping"}, {"pong"}}

	res := make([]testTable, 0)
	require.NoError(t, db.SelectNamed(tx, &res, query, expected))
	require.Equal(t, expected, res)

	res = make([]testTable, 0)
	require.NoError(t, db.SelectNamedContext(context.Background(), tx, &res, query, expected))
	require.Equal(t, expected, res)

	// Type castings must be in the form of `CAST(:q AS <type>)`
	strRes := make([]string, 0)
	require.Error(t, db.SelectNamed(tx, &strRes, "SELECT :q::text", map[string]interface{}{"q": "test"}))
	require.NoError(t, db.SelectNamed(tx, &strRes, "SELECT CAST(:q AS text)", map[string]interface{}{"q": "test"}))
	require.Error(t, db.SelectNamedContext(context.Background(), tx, &strRes, "SELECT :q::text", map[string]interface{}{"q": "test"}))
	require.NoError(t, db.SelectNamedContext(context.Background(), tx, &strRes, "SELECT CAST(:q AS text)", map[string]interface{}{"q": "test"}))
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilTestSuite))
}
