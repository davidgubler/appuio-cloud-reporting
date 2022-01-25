package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
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
