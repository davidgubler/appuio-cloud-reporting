package db_test

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	_ "github.com/appuio/appuio-cloud-reporting/pkg/db/flag"
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
