package httpapi

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/reign-demo/flightboard/api/internal/store"
	"github.com/reign-demo/flightboard/api/internal/workorder"
)

func TestAddWorkOrderExpense(t *testing.T) {
	s := newTestServer(t)
	path := "/api/work-orders/9631810/expenses"
	ids := map[string]bool{"e1": true, "e2": true}
	for i, category := range []workorder.ExpenseCategory{
		workorder.ExpenseParts, workorder.ExpenseMileage, workorder.ExpenseTools, workorder.ExpenseOther,
	} {
		t.Run(string(category), func(t *testing.T) {
			body, err := json.Marshal(map[string]any{
				"category": category, "description": "Service supplies", "amountCents": 1000,
				"id": "caller-id", "status": "Approved", "createdAt": "caller-time", "reason": "caller reason",
			})
			if err != nil {
				t.Fatal(err)
			}
			before := time.Now()
			rec := do(t, s, "POST", path, string(body))
			if rec.Code != http.StatusCreated {
				t.Fatalf("code = %d body %s", rec.Code, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q", got)
			}
			created := decode[workorder.Expense](t, rec)
			if created.ID == "" || created.ID == "caller-id" || ids[created.ID] {
				t.Fatalf("id = %q, want a new server-assigned id", created.ID)
			}
			ids[created.ID] = true
			if created.Status != workorder.ExpensePending || created.Reason != "" {
				t.Errorf("server fields not assigned: %+v", created)
			}
			if created.CreatedAt.Before(before) || created.CreatedAt.After(time.Now()) || created.CreatedAt.Location() != time.UTC {
				t.Errorf("createdAt = %v, want current UTC time", created.CreatedAt)
			}
			if created.Category != category || created.Description != "Service supplies" || created.AmountCents != 1000 {
				t.Errorf("expense details changed: %+v", created)
			}
			rec = do(t, s, "GET", path, "")
			if rec.Code != http.StatusOK {
				t.Fatalf("list code = %d body %s", rec.Code, rec.Body)
			}
			list := decode[[]workorder.Expense](t, rec)
			if len(list) != i+3 || list[0] != created {
				t.Fatalf("expenses = %+v, want created expense first and %d entries", list, i+3)
			}
		})
	}
}

func TestAddWorkOrderExpenseErrors(t *testing.T) {
	const validBody = `{"category":"Parts","description":"Patch cable","amountCents":1000}`
	const categoryError = `invalid work order: unknown expense category ""`
	const descriptionError = "invalid work order: expense description is required"
	const amountError = "invalid work order: expense amount must be between 1 and 500000 cents"
	for _, tt := range []struct {
		name string
		id   string
		body string
		code int
		err  string
	}{
		{"unknown work order", "1", validBody, http.StatusNotFound, "not found"},
		{"invalid id", "abc", validBody, http.StatusBadRequest, "id must be a number"},
		{"malformed JSON", "9631810", `{not json`, http.StatusBadRequest, "invalid JSON body"},
		{"empty body", "9631810", "", http.StatusBadRequest, "invalid JSON body"},
		{"array body", "9631810", `[]`, http.StatusBadRequest, "invalid JSON body"},
		{"null body", "9631810", `null`, http.StatusBadRequest, categoryError},
		{"missing fields", "9631810", `{}`, http.StatusBadRequest, categoryError},
		{"missing category", "9631810", `{"description":"Cable","amountCents":1000}`, http.StatusBadRequest, categoryError},
		{"unknown category", "9631810", `{"category":"Meals","description":"Lunch","amountCents":1000}`, http.StatusBadRequest, `invalid work order: unknown expense category "Meals"`},
		{"missing description", "9631810", `{"category":"Parts","amountCents":1000}`, http.StatusBadRequest, descriptionError},
		{"empty description", "9631810", `{"category":"Parts","description":"","amountCents":1000}`, http.StatusBadRequest, descriptionError},
		{"blank description", "9631810", `{"category":"Parts","description":" \t","amountCents":1000}`, http.StatusBadRequest, descriptionError},
		{"long description", "9631810", `{"category":"Parts","description":"` + strings.Repeat("a", 201) + `","amountCents":1000}`, http.StatusBadRequest, "invalid work order: expense description must not exceed 200 characters"},
		{"missing amount", "9631810", `{"category":"Parts","description":"Cable"}`, http.StatusBadRequest, amountError},
		{"zero amount", "9631810", `{"category":"Parts","description":"Cable","amountCents":0}`, http.StatusBadRequest, amountError},
		{"negative amount", "9631810", `{"category":"Parts","description":"Cable","amountCents":-1}`, http.StatusBadRequest, amountError},
		{"over maximum amount", "9631810", `{"category":"Parts","description":"Cable","amountCents":500001}`, http.StatusBadRequest, amountError},
		{"fractional amount", "9631810", `{"category":"Parts","description":"Cable","amountCents":1.5}`, http.StatusBadRequest, "invalid JSON body"},
		{"string amount", "9631810", `{"category":"Parts","description":"Cable","amountCents":"1000"}`, http.StatusBadRequest, "invalid JSON body"},
		{"wrong category type", "9631810", `{"category":1,"description":"Cable","amountCents":1000}`, http.StatusBadRequest, "invalid JSON body"},
		{"wrong description type", "9631810", `{"category":"Parts","description":1,"amountCents":1000}`, http.StatusBadRequest, "invalid JSON body"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestServer(t)
			before := s.store.ListWorkOrders(store.Filter{})
			rec := do(t, s, "POST", "/api/work-orders/"+tt.id+"/expenses", tt.body)
			if rec.Code != tt.code {
				t.Fatalf("code = %d body %s", rec.Code, rec.Body)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q", got)
			}
			if got := decode[map[string]string](t, rec); !maps.Equal(got, map[string]string{"error": tt.err}) {
				t.Errorf("error body = %v, want error %q", got, tt.err)
			}
			for _, wo := range before {
				after, err := s.store.Expenses(wo.ID)
				if err != nil || !slices.Equal(after, wo.Expenses) {
					t.Errorf("failed request changed expenses for %d: %+v, err = %v", wo.ID, after, err)
				}
			}
		})
	}
}
