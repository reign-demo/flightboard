package httpapi

import (
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/reign-demo/flightboard/api/internal/workorder"
)

func TestAddWorkOrderExpense(t *testing.T) {
	s := newTestServer(t)
	path := "/api/work-orders/9631810/expenses"
	for i, category := range []workorder.ExpenseCategory{
		workorder.ExpenseParts, workorder.ExpenseMileage, workorder.ExpenseTools, workorder.ExpenseOther,
	} {
		body := fmt.Sprintf(`{"category":%q,"description":"Service expense","amountCents":1,
			"id":"caller-id","status":"Approved","createdAt":"not a timestamp","reason":"caller reason"}`, category)
		before := time.Now()
		rec := do(t, s, "POST", path, body)
		if rec.Code != http.StatusCreated {
			t.Fatalf("category %s: code = %d body %s", category, rec.Code, rec.Body)
		}
		if got := rec.Header().Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		created := decode[workorder.Expense](t, rec)
		if created.ID != fmt.Sprintf("e%d", i+3) || created.Status != workorder.ExpensePending || created.Reason != "" {
			t.Fatalf("server-owned fields = %+v", created)
		}
		if created.CreatedAt.Before(before) || created.CreatedAt.After(time.Now()) || created.CreatedAt.Location() != time.UTC {
			t.Errorf("createdAt = %v, want current UTC time", created.CreatedAt)
		}
		if created.Category != category || created.Description != "Service expense" || created.AmountCents != 1 {
			t.Errorf("expense details = %+v", created)
		}
		rec = do(t, s, "GET", path, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("list code = %d body %s", rec.Code, rec.Body)
		}
		listed := decode[[]workorder.Expense](t, rec)
		if len(listed) != i+3 || listed[0] != created {
			t.Fatalf("expenses = %+v, want created expense first and %d entries", listed, i+3)
		}
	}
}

func TestAddWorkOrderExpenseErrors(t *testing.T) {
	valid := `{"category":"Parts","description":"Cable","amountCents":100}`
	for _, tt := range []struct {
		name string
		id   string
		body string
		code int
		err  string
	}{
		{name: "unknown work order", id: "1", body: valid, code: http.StatusNotFound, err: "not found"},
		{name: "invalid id", id: "abc", body: valid, code: http.StatusBadRequest, err: "id must be a number"},
		{name: "malformed JSON", body: `{not json`, err: "invalid JSON body"},
		{name: "empty body", err: "invalid JSON body"},
		{name: "wrong body type", body: `[]`, err: "invalid JSON body"},
		{name: "missing category", body: `{"description":"Cable","amountCents":100}`, err: `invalid work order: unknown expense category ""`},
		{name: "unknown category", body: `{"category":"Travel","description":"Cable","amountCents":100}`, err: `invalid work order: unknown expense category "Travel"`},
		{name: "missing description", body: `{"category":"Parts","amountCents":100}`, err: "invalid work order: expense description is required"},
		{name: "empty description", body: `{"category":"Parts","description":"","amountCents":100}`, err: "invalid work order: expense description is required"},
		{name: "blank description", body: `{"category":"Parts","description":" \t\n","amountCents":100}`, err: "invalid work order: expense description is required"},
		{name: "long description", body: `{"category":"Parts","description":"` + strings.Repeat("a", 201) + `","amountCents":100}`, err: "invalid work order: expense description must not exceed 200 characters"},
		{name: "missing amount", body: `{"category":"Parts","description":"Cable"}`, err: "invalid work order: expense amount must be between 1 and 500000 cents"},
		{name: "zero amount", body: `{"category":"Parts","description":"Cable","amountCents":0}`, err: "invalid work order: expense amount must be between 1 and 500000 cents"},
		{name: "negative amount", body: `{"category":"Parts","description":"Cable","amountCents":-1}`, err: "invalid work order: expense amount must be between 1 and 500000 cents"},
		{name: "excessive amount", body: `{"category":"Parts","description":"Cable","amountCents":500001}`, err: "invalid work order: expense amount must be between 1 and 500000 cents"},
		{name: "fractional amount", body: `{"category":"Parts","description":"Cable","amountCents":1.5}`, err: "invalid JSON body"},
		{name: "string amount", body: `{"category":"Parts","description":"Cable","amountCents":"100"}`, err: "invalid JSON body"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			before, err := s.store.Expenses(9631810)
			if err != nil {
				t.Fatal(err)
			}
			id := tt.id
			if id == "" {
				id = "9631810"
			}
			code := tt.code
			if code == 0 {
				code = http.StatusBadRequest
			}
			rec := do(t, s, "POST", "/api/work-orders/"+id+"/expenses", tt.body)
			if rec.Code != code {
				t.Fatalf("code = %d body %s, want %d", rec.Code, rec.Body, code)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q", got)
			}
			if got := decode[map[string]string](t, rec); !maps.Equal(got, map[string]string{"error": tt.err}) {
				t.Errorf("error body = %v, want error %q", got, tt.err)
			}
			after, err := s.store.Expenses(9631810)
			if err != nil || !slices.Equal(before, after) {
				t.Errorf("failed request changed expenses: %+v, err = %v", after, err)
			}
		})
	}
}
