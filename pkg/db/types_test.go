package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/appuio/appuio-cloud-reporting/pkg/db/dbtest"
)

func TestTimerange(t *testing.T) {
	lower := db.MustTimestamp(time.Now())
	upper := db.MustTimestamp(pgtype.Infinity)
	subject := db.Timerange(lower, upper)

	assert.Equal(t, subject.Status, pgtype.Present, "new timetamp should be present")
	assert.Equal(t, subject.Lower, lower)
	assert.Equal(t, subject.Upper, upper)
	assert.Equal(t, subject.LowerType, pgtype.Inclusive, "lower bound should be inclusive")
	assert.Equal(t, subject.UpperType, pgtype.Exclusive, "upper bound should be exclusive")
}

func TestBuildDateTime(t *testing.T) {
	ts := time.Date(2033, time.March, 23, 17, 0, 0, 0, time.UTC)
	subject := db.BuildDateTime(ts)

	assert.True(t, subject.Timestamp.Equal(ts))
	assert.Equal(t, subject.Year, 2033)
	assert.Equal(t, subject.Month, 3)
	assert.Equal(t, subject.Day, 23)
	assert.Equal(t, subject.Hour, 17)
}

func TestTypes(t *testing.T) {
	suite.Run(t, new(TypesTestSuite))
}

type TypesTestSuite struct {
	dbtest.Suite
}

func (s *TypesTestSuite) TestTypes_Query() {
	t := s.T()
	d := s.DB()

	_, err := db.CreateQuery(d, db.Query{
		Name:   "test",
		Query:  "test",
		Unit:   "tps",
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)

	count := "SELECT ((SELECT COUNT(*) FROM queries WHERE name=$1) = 1)"
	requireQueryTrue(t, d, count, "test")
}
func (s *TypesTestSuite) TestTypes_Product() {
	t := s.T()
	d := s.DB()

	_, err := db.CreateProduct(d, db.Product{
		Source: "test",
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)

	count := "SELECT ((SELECT COUNT(*) FROM products WHERE source=$1) = 1)"
	requireQueryTrue(t, d, count, "test")
}
func (s *TypesTestSuite) TestTypes_Discount() {
	t := s.T()
	d := s.DB()

	_, err := db.CreateDiscount(d, db.Discount{
		Source: "test",
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)

	count := "SELECT ((SELECT COUNT(*) FROM discounts WHERE source=$1) = 1)"
	requireQueryTrue(t, d, count, "test")
}
