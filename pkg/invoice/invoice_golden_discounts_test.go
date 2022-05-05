package invoice_test

import (
	"database/sql"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"

	"github.com/stretchr/testify/require"
)

func (s *InvoiceGoldenSuite) TestInvoiceGolden_Discounts() {
	t := s.T()
	tdb := s.DB()

	_, err := db.CreateProduct(tdb, db.Product{
		Source: "my-product",
		Amount: 1,
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)

	_, err = db.CreateDiscount(tdb, db.Discount{
		Source: "my-product",
		During: db.InfiniteRange(),
	})
	require.NoError(t, err)

	_, err = db.CreateDiscount(tdb, db.Discount{
		Source:   "my-product:*:my-tenant",
		Discount: 0.25,
		During:   db.InfiniteRange(),
	})
	require.NoError(t, err)

	_, err = db.CreateDiscount(tdb, db.Discount{
		Source:   "my-product:my-cluster:my-tenant",
		Discount: 0.5,
		During:   db.InfiniteRange(),
	})
	require.NoError(t, err)

	_, err = db.CreateDiscount(tdb, db.Discount{
		Source:   "my-product:my-cluster:my-tenant:secret-namespace",
		Discount: 1,
		During:   db.InfiniteRange(),
	})
	require.NoError(t, err)

	q, err := db.CreateQuery(tdb, db.Query{
		Name:        "test",
		Description: "test description",
		Query:       "test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	s.prom.queries[q.Query] = fakeQueryResults{
		"my-product:my-cluster:my-tenant:my-namespace":    fakeQuerySample{Value: 42},
		"my-product:my-cluster:my-tenant:other-namespace": fakeQuerySample{Value: 42},
		"my-product:other-cluster:my-tenant:my-namespace": fakeQuerySample{Value: 42},
		"my-product:my-cluster:other-tenant:my-namespace": fakeQuerySample{Value: 23},
	}
	require.NoError(t, err)

	sq, err := db.CreateQuery(tdb, db.Query{
		ParentID: sql.NullString{
			String: q.Id,
			Valid:  true,
		},
		Name:        "sub-test",
		Description: "A sub query of Test",
		Query:       "sub-test",
		Unit:        "tps",
		During:      db.InfiniteRange(),
	})
	s.prom.queries[sq.Query] = fakeQueryResults{
		"my-product:my-cluster:my-tenant:my-namespace":    fakeQuerySample{Value: 4},
		"my-product:my-cluster:my-tenant:other-namespace": fakeQuerySample{Value: 4},
		"my-product:other-cluster:my-tenant:my-namespace": fakeQuerySample{Value: 4},
		"my-product:my-cluster:other-tenant:my-namespace": fakeQuerySample{Value: 2},
	}
	require.NoError(t, err)

	runReport(t, tdb, s.prom, q.Name, "2022-02-25", "2022-03-10")

	invoiceEqualsGolden(t, "discounts",
		generateInvoice(t, tdb, 2022, time.March),
		*updateGolden)
}
