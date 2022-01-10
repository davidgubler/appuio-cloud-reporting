package dbtest

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	dbflag "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
)

type Suite struct {
	suite.Suite
	db *sqlx.DB
}

func (ts *Suite) Begin() *sqlx.Tx {
	txx, err := ts.db.Beginx()
	require.NoError(ts.T(), err)
	return txx
}

func (ts *Suite) SetupSuite() {
	dbx, err := db.Openx(dbflag.DatabaseURL)
	require.NoError(ts.T(), err)
	ts.db = dbx
}

func (ts *Suite) TearDownSuite() {
	t := ts.T()
	require.NoError(t, ts.db.Close())
}
