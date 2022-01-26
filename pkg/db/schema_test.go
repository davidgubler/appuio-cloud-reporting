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

func (s *SchemaTestSuite) TestDiscounts_Discount_MinMaxConstraint() {
	t := s.T()

	tests := []struct {
		name     string
		discount float64
		errf     func(*testing.T, error)
	}{
		{"overMax", 1.3, requireCheckConstraintError},
		{"underMin", -7, requireCheckConstraintError},
		{"inRangeLow", 0, func(t *testing.T, e error) { require.NoError(t, e) }},
		{"inRange", 0.78, func(t *testing.T, e error) { require.NoError(t, e) }},
		{"inRangeHigh", 1, func(t *testing.T, e error) { require.NoError(t, e) }},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tx := s.Begin()
			defer tx.Rollback()

			stmt, err := tx.PrepareNamed("INSERT INTO discounts (source, discount) VALUES (:source, :discount)")
			require.NoError(t, err)
			defer stmt.Close()

			_, err = stmt.Exec(db.Discount{
				Source:   testCase.name,
				Discount: testCase.discount,
			})
			testCase.errf(t, err)
		})
	}
}

func TestSchema(t *testing.T) {
	suite.Run(t, new(SchemaTestSuite))
}

func requireExclusionValidationError(t *testing.T, err error) {
	t.Helper()

	pgErr := &pgconn.PgError{}
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23P01", pgErr.SQLState(), "error code should match exclusion violation error")
}

func requireCheckConstraintError(t *testing.T, err error) {
	t.Helper()

	pgErr := &pgconn.PgError{}
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "23514", pgErr.SQLState(), "error code should match check constraint error")
}
