package httpapi

import (
	"net/http"
	"slices"
	"testing"
)

func TestListWorkOrdersBuyerFilter(t *testing.T) {
	allIDs := make([]int, 25)
	for i := range allIDs {
		allIDs[i] = 9631825 - i
	}
	buyerIDs := []int{9631825, 9631821, 9631815, 9631810, 9631809, 9631801}
	s := newTestServer(t)
	for _, tt := range []struct {
		name  string
		query string
		want  []int
	}{
		{name: "missing buyer", want: allIDs},
		{name: "empty buyer", query: "?buyer=", want: allIDs},
		{name: "matching buyer", query: "?buyer=Northwind%20Retail", want: buyerIDs},
		{name: "lowercase buyer", query: "?buyer=northwind%20retail", want: buyerIDs},
		{name: "uppercase buyer", query: "?buyer=NORTHWIND%20RETAIL", want: buyerIDs},
		{name: "mixed case buyer", query: "?buyer=NoRtHwInD%20ReTaIl", want: buyerIDs},
		{name: "unknown buyer", query: "?buyer=Unknown", want: []int{}},
		{name: "partial buyer does not match", query: "?buyer=Northwind", want: []int{}},
		{
			name: "with status", query: "?buyer=Northwind%20Retail&status=Assigned",
			want: []int{9631815, 9631810},
		},
		{
			name: "with provider", query: "?buyer=Northwind%20Retail&providerId=p-101",
			want: []int{9631810},
		},
		{
			name: "with type", query: "?buyer=Northwind%20Retail&typeOfWork=Point%20of%20Sale",
			want: []int{9631825, 9631815, 9631801},
		},
		{
			name: "with all filters", query: "?buyer=northwind%20retail&status=Assigned&providerId=p-101&typeOfWork=networking",
			want: []int{9631810},
		},
		{
			name: "nonmatching buyer with other filters", query: "?buyer=Bluepine%20Clinics&status=Assigned&providerId=p-101&typeOfWork=Networking",
			want: []int{},
		},
		{
			name: "nonmatching status", query: "?buyer=Northwind%20Retail&status=Paid",
			want: []int{},
		},
		{
			name: "nonmatching provider", query: "?buyer=Northwind%20Retail&providerId=p-103",
			want: []int{},
		},
		{
			name: "nonmatching type", query: "?buyer=Northwind%20Retail&typeOfWork=Server",
			want: []int{},
		},
		{
			name: "empty buyer preserves other filters", query: "?buyer=&status=Assigned&providerId=p-101&typeOfWork=Networking",
			want: []int{9631810},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, s, "GET", "/api/work-orders"+tt.query, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("code = %d body %s", rec.Code, rec.Body)
			}
			list := decode[[]woJSON](t, rec)
			got := make([]int, len(list))
			for i, wo := range list {
				got[i] = wo.ID
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ids = %v, want %v", got, tt.want)
			}
			if len(tt.want) == 0 && rec.Body.String() != "[]\n" {
				t.Errorf("body = %q, want %q", rec.Body.String(), "[]\n")
			}
		})
	}
}
