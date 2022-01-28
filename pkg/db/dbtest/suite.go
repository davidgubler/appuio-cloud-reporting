package dbtest

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
)

var DatabaseURL = urlFromEnv()

// Suite holds a database test suite. Each Suite holds its own clone of
// the database given by the `ACR_DB_URL` environment variable.
// The database is cloned before the suite starts and dropped in the suite teardown.
// Suites can be run in parallel.
type Suite struct {
	suite.Suite

	maintenanceDB *sqlx.DB

	tmpDB     *sqlx.DB
	tmpDBName string
}

func (ts *Suite) DB() *sqlx.DB {
	return ts.tmpDB
}

func (ts *Suite) Begin() *sqlx.Tx {
	txx, err := ts.DB().Beginx()
	require.NoError(ts.T(), err)
	return txx
}

func (ts *Suite) SetupSuite() {
	u, err := url.Parse(DatabaseURL)
	require.NoError(ts.T(), err)
	dbName := strings.TrimPrefix(u.Path, "/")
	tmpDbName := dbName + "-tmp-" + uuid.NewString()
	ts.tmpDBName = tmpDbName

	// Connect to a neutral database
	mdb, err := openMaintenance(DatabaseURL)
	require.NoError(ts.T(), err)
	ts.maintenanceDB = mdb

	require.NoError(ts.T(),
		cloneDB(ts.maintenanceDB, pgx.Identifier{tmpDbName}, pgx.Identifier{dbName}),
	)

	// Connect to the temporary database
	tmpURL := new(url.URL)
	*tmpURL = *u
	tmpURL.Path = "/" + tmpDbName
	ts.T().Logf("Using database name: %s", tmpDbName)
	dbx, err := db.Openx(tmpURL.String())
	require.NoError(ts.T(), err)
	ts.tmpDB = dbx
}

func (ts *Suite) TearDownSuite() {
	t := ts.T()
	require.NoError(t, ts.tmpDB.Close())
	require.NoError(t, dropDB(ts.maintenanceDB, pgx.Identifier{ts.tmpDBName}))
	require.NoError(t, ts.maintenanceDB.Close())
}

func cloneDB(maint *sqlx.DB, dst, src pgx.Identifier) error {
	_, err := maint.Exec(
		fmt.Sprintf(`CREATE DATABASE %s TEMPLATE %s`,
			dst.Sanitize(),
			src.Sanitize(),
		),
	)
	if err != nil {
		return fmt.Errorf("error cloning database `%s` to `%s`: %w", src.Sanitize(), dst.Sanitize(), err)
	}
	return nil
}

func dropDB(maint *sqlx.DB, db pgx.Identifier) error {
	_, err := maint.Exec(
		fmt.Sprintf(`DROP DATABASE %s WITH (FORCE)`, db.Sanitize()),
	)
	if err != nil {
		return fmt.Errorf("error dropping database `%s`: %w", db.Sanitize(), err)
	}
	return nil
}

func openMaintenance(dbURL string) (*sqlx.DB, error) {
	maintURL, err := url.Parse(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing url: %w", err)
	}
	maintURL.Path = "/postgres"
	mdb, err := db.Openx(maintURL.String())
	if err != nil {
		return nil, fmt.Errorf("error connecting to maintenance (`%s`) database: %w", maintURL.Path, err)
	}
	return mdb, nil
}

func urlFromEnv() string {
	if u, exists := os.LookupEnv("ACR_DB_URL"); exists {
		return u
	}
	return "postgres://postgres@localhost/reporting-db?sslmode=disable"
}
