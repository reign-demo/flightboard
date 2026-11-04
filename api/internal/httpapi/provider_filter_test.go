package httpapi

import (
	"net/http"
	"slices"
	"testing"

	"github.com/reign-demo/flightboard/api/internal/workorder"
)

func TestListProvidersStateFilter(t *testing.T) {
	allIDs := []string{"p-101", "p-102", "p-103", "p-104", "p-105", "p-106", "p-107", "p-108"}
	s := newTestServer(t)
	for _, tt := range []struct {
		name  string
		query string
		want  []string
	}{
		{name: "missing state", want: allIDs},
		{name: "empty state", query: "?state=", want: allIDs},
		{name: "uppercase state", query: "?state=MN", want: []string{"p-101", "p-104"}},
		{name: "lowercase state", query: "?state=mn", want: []string{"p-101", "p-104"}},
		{name: "mixed case state", query: "?state=mN", want: []string{"p-101", "p-104"}},
		{name: "other state", query: "?state=IL", want: []string{"p-102"}},
		{name: "unknown state", query: "?state=ZZ", want: []string{}},
		{name: "partial state", query: "?state=M", want: []string{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(t, s, "GET", "/api/providers"+tt.query, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("code = %d body %s", rec.Code, rec.Body)
			}
			list := decode[[]workorder.Provider](t, rec)
			if list == nil {
				t.Fatal("providers must be an array, not null")
			}
			got := make([]string, len(list))
			for i, p := range list {
				got[i] = p.ID
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("ids = %v, want %v", got, tt.want)
			}
		})
	}
}
