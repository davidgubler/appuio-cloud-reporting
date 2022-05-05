package invoice_test

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/appuio/appuio-cloud-reporting/pkg/sourcekey"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
)

var (
	updateGolden = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

type fakeQuerySample struct {
	Value model.SampleValue
}
type fakeQueryResults map[string]fakeQuerySample
type fakeQuerrier struct {
	queries map[string]fakeQueryResults
}

func (q fakeQuerrier) Query(ctx context.Context, query string, ts time.Time) (model.Value, apiv1.Warnings, error) {
	var res model.Vector
	for k, s := range q.queries[query] {
		sk, err := sourcekey.Parse(k)
		if err != nil {
			return nil, nil, err
		}
		res = append(res, &model.Sample{
			Metric: map[model.LabelName]model.LabelValue{
				"product":  model.LabelValue(k),
				"category": model.LabelValue(fmt.Sprintf("%s:%s", sk.Zone, sk.Namespace)),
				"tenant":   model.LabelValue(sk.Tenant),
			},
			Value: s.Value,
		})
	}
	return res, nil, nil
}

func invoiceEqual(t *testing.T, expInv, inv invoice.Invoice) bool {
	sortInvoice(&inv)
	sortInvoice(&expInv)
	return assert.Equal(t, expInv, inv)
}

func sortInvoices(invSlice []invoice.Invoice) []invoice.Invoice {
	sort.Slice(invSlice, func(i, j int) bool {
		return invSlice[i].Tenant.Source < invSlice[j].Tenant.Source
	})
	for k := range invSlice {
		sortInvoice(&invSlice[k])
	}
	return invSlice
}

func sortInvoice(inv *invoice.Invoice) {
	sort.Slice(inv.Categories, func(i, j int) bool {
		// This is horrible, but I don't really have any ID or similar to sort on..
		iraw, _ := json.Marshal(inv.Categories[i])
		jraw, _ := json.Marshal(inv.Categories[j])
		return string(iraw) < string(jraw)
	})
	for catIter := range inv.Categories {
		sort.Slice(inv.Categories[catIter].Items, func(i, j int) bool {
			// This is horrible, but I don't really have any ID or similar to sort on..
			iraw, _ := json.Marshal(inv.Categories[catIter].Items[i])
			jraw, _ := json.Marshal(inv.Categories[catIter].Items[j])
			return string(iraw) < string(jraw)
		})
		for itemIter := range inv.Categories[catIter].Items {
			sort.Slice(inv.Categories[catIter].Items[itemIter].SubItems, func(i, j int) bool {
				return inv.Categories[catIter].Items[itemIter].SubItems[i].Description < inv.Categories[catIter].Items[itemIter].SubItems[j].Description
			})
		}
	}
}
