package invoice_test

import (
	"encoding/json"
	"flag"
	"os"
	"sort"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/invoice"
	"github.com/stretchr/testify/assert"
)

var (
	updateGolden = flag.Bool("update", false, "update the golden files of this test")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
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
	}
}
