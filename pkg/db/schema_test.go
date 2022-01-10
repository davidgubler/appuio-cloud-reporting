package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgconn"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

type SchemaTestSuite struct {
	dbtest.Suite
}

func (s *SchemaTestSuite) TestQueries_NameUnitDuring_NonOverlapping() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	stmt, err := tx.PrepareNamed("INSERT INTO queries (name, query, unit, during) VALUES (:name, :query, :unit, :during)")
	require.NoError(t, err)
	defer stmt.Close()

	base := db.Query{
		Name:  "test",
		Unit:  "test",
		Query: "test",
		During: db.Timerange(
			db.MustTimestamp(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			db.MustTimestamp(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
		),
	}
	_, err = stmt.Exec(base)
	require.NoError(t, err)

	nonOverlapping := base
	nonOverlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(nonOverlapping)
	require.NoError(t, err)

	overlapping := base
	overlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(overlapping)
	requireExclusionValidationError(t, err)
}

func (s *SchemaTestSuite) TestProducts_SourceDuring_NonOverlapping() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	stmt, err := tx.PrepareNamed("INSERT INTO products (source, unit, during) VALUES (:source, :unit, :during)")
	require.NoError(t, err)
	defer stmt.Close()

	base := db.Product{
		Source: "test",
		During: db.Timerange(
			db.MustTimestamp(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			db.MustTimestamp(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
		),
	}
	_, err = stmt.Exec(base)
	require.NoError(t, err)

	nonOverlapping := base
	nonOverlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(nonOverlapping)
	require.NoError(t, err)

	overlapping := base
	overlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(overlapping)
	requireExclusionValidationError(t, err)
}

func (s *SchemaTestSuite) TestDiscounts_SourceDuring_NonOverlapping() {
	t := s.T()
	tx := s.Begin()
	defer tx.Rollback()

	stmt, err := tx.PrepareNamed("INSERT INTO discounts (source, during) VALUES (:source, :during)")
	require.NoError(t, err)
	defer stmt.Close()

	base := db.Discount{
		Source: "test",
		During: db.Timerange(
			db.MustTimestamp(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)),
			db.MustTimestamp(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)),
		),
	}
	_, err = stmt.Exec(base)
	require.NoError(t, err)

	nonOverlapping := base
	nonOverlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(1991, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(1995, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(nonOverlapping)
	require.NoError(t, err)

	overlapping := base
	overlapping.During = db.Timerange(
		db.MustTimestamp(time.Date(2019, time.January, 1, 0, 0, 0, 0, time.UTC)),
		db.MustTimestamp(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)),
	)
	_, err = stmt.Exec(overlapping)
	requireExclusionValidationError(t, err)
}

func TestSchema(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func requireExclusionValidationError(t *testing.T, err error) {
	t.Helper()

	pgErr := &pgconn.PgError{}
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, pgErr.SQLState(), "23P01", "error code should match exclusion violation error")
}
