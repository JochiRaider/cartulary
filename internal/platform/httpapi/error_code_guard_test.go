package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestErrorCodeGuardHonoursDeclarations covers Core 01 REQ-01-702 and
// TH-HARNESS-REQ-858: a resolved operation may write only its declared codes;
// the guard reports any other code and, when enforcing, replaces it with 500
// internal_error. internal_error has no exemption: every operation declares
// it at 500, the status it is always emitted with, and an operation that does
// not is reported. Requests that resolve to no operation are outside the
// rule, which is how REQ-01-707's route-resolution codes are written; a
// resolved operation writing one is undeclared. It also
// covers the TH-HARNESS-REQ-862 emission hook: every error a resolved
// operation writes, declared or not, is observed with the operation's own
// code, never the enforcement substitute. An operation whose codes an owner
// registry governs (Network Flow) is outside both.
func TestErrorCodeGuardHonoursDeclarations(t *testing.T) {
	operation := ContractOperation{
		OwnerID: "module.example", OperationID: "exampleOperation", Availability: BaseOperationAvailability,
		ErrorCodes: map[int][]string{
			http.StatusNotFound:            {"incident_not_found"},
			http.StatusInternalServerError: {"internal_error"},
			http.StatusServiceUnavailable:  {"service_unavailable"},
		},
	}
	type violation struct {
		operationID string
		status      int
		code        string
	}
	var observed []violation
	write := func(t *testing.T, enforce, resolved bool, status int, code string) (int, string, []violation) {
		t.Helper()
		var reported []violation
		observed = nil
		guard := &ErrorCodeGuard{Enforce: enforce, Report: func(operationID string, status int, code string) {
			reported = append(reported, violation{operationID, status, code})
		}, Observe: func(operationID string, status int, code string) {
			observed = append(observed, violation{operationID, status, code})
		}}
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = WriteError(w, r, status, code, "message", nil)
		})
		var routed http.Handler = handler
		if resolved {
			routed = withOperationMetadata(operation, handler)
		}
		recorder := httptest.NewRecorder()
		withErrorCodeGuard(routed, guard).ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/api/v1/example", nil))
		var envelope ErrorEnvelope
		if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
			t.Fatalf("decode error envelope: %v", err)
		}
		return recorder.Code, envelope.Error.Code, reported
	}

	emitted := func(t *testing.T, want ...violation) {
		t.Helper()
		if len(observed) != len(want) || len(want) == 1 && observed[0] != want[0] {
			t.Fatalf("observed %v, want %v", observed, want)
		}
	}
	t.Run("declared code passes", func(t *testing.T) {
		status, code, reported := write(t, true, true, http.StatusNotFound, "incident_not_found")
		if status != http.StatusNotFound || code != "incident_not_found" || len(reported) != 0 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t, violation{"exampleOperation", http.StatusNotFound, "incident_not_found"})
	})
	t.Run("undeclared code is observed unchanged", func(t *testing.T) {
		status, code, reported := write(t, false, true, http.StatusConflict, "illegal_transition")
		if status != http.StatusConflict || code != "illegal_transition" {
			t.Fatalf("observe mode changed the response: %d %s", status, code)
		}
		if len(reported) != 1 || reported[0] != (violation{"exampleOperation", http.StatusConflict, "illegal_transition"}) {
			t.Fatalf("reported %v", reported)
		}
		emitted(t, violation{"exampleOperation", http.StatusConflict, "illegal_transition"})
	})
	t.Run("undeclared code is enforced as internal_error", func(t *testing.T) {
		status, code, reported := write(t, true, true, http.StatusNotFound, "user_not_found")
		if status != http.StatusInternalServerError || code != "internal_error" || len(reported) != 1 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t, violation{"exampleOperation", http.StatusNotFound, "user_not_found"})
	})
	t.Run("declared internal_error and service_unavailable pass", func(t *testing.T) {
		for _, tc := range []struct {
			status int
			code   string
		}{{http.StatusInternalServerError, "internal_error"}, {http.StatusServiceUnavailable, "service_unavailable"}} {
			status, code, reported := write(t, true, true, tc.status, tc.code)
			if status != tc.status || code != tc.code || len(reported) != 0 {
				t.Fatalf("%s: got %d %s reported=%v", tc.code, status, code, reported)
			}
			emitted(t, violation{"exampleOperation", tc.status, tc.code})
		}
	})
	t.Run("internal_error is observed and checked at its emitted status", func(t *testing.T) {
		status, code, reported := write(t, true, true, http.StatusBadRequest, "internal_error")
		if status != http.StatusInternalServerError || code != "internal_error" || len(reported) != 0 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t, violation{"exampleOperation", http.StatusInternalServerError, "internal_error"})
	})
	t.Run("internal_error is checked as a declared code", func(t *testing.T) {
		declared := operation
		operation.ErrorCodes = map[int][]string{http.StatusNotFound: {"incident_not_found"}}
		defer func() { operation = declared }()
		status, code, reported := write(t, true, true, http.StatusInternalServerError, "internal_error")
		if status != http.StatusInternalServerError || code != "internal_error" ||
			len(reported) != 1 || reported[0] != (violation{"exampleOperation", http.StatusInternalServerError, "internal_error"}) {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t, violation{"exampleOperation", http.StatusInternalServerError, "internal_error"})
	})
	t.Run("a declared code at another status is undeclared", func(t *testing.T) {
		status, code, reported := write(t, true, true, http.StatusInternalServerError, "service_unavailable")
		if status != http.StatusInternalServerError || code != "internal_error" ||
			len(reported) != 1 || reported[0] != (violation{"exampleOperation", http.StatusInternalServerError, "service_unavailable"}) {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
	})
	t.Run("a resolved operation may not write a route-resolution code", func(t *testing.T) {
		for _, tc := range []struct {
			status int
			code   string
		}{{http.StatusNotFound, "route_not_found"}, {http.StatusMethodNotAllowed, "method_not_allowed"}} {
			status, code, reported := write(t, true, true, tc.status, tc.code)
			if status != http.StatusInternalServerError || code != "internal_error" || len(reported) != 1 || reported[0] != (violation{"exampleOperation", tc.status, tc.code}) {
				t.Fatalf("%s: got %d %s reported=%v", tc.code, status, code, reported)
			}
		}
	})
	t.Run("route resolution writes its codes outside the rule", func(t *testing.T) {
		status, code, reported := write(t, true, false, http.StatusNotFound, "route_not_found")
		if status != http.StatusNotFound || code != "route_not_found" || len(reported) != 0 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t)
	})
	t.Run("an owner-registry operation is outside the rule and its coverage", func(t *testing.T) {
		// Network Flow's codes are governed by its own registry (NF NLSpec
		// Table 21-A), not by REQ-01-702 OpenAPI declarations.
		declared := operation
		operation = ContractOperation{OwnerID: "module.example", OperationID: "nf.example.get", Availability: "example_profile", ErrorCodeRegistry: "contracts/example/errors.json"}
		defer func() { operation = declared }()
		status, code, reported := write(t, true, true, http.StatusConflict, "transaction_conflict")
		if status != http.StatusConflict || code != "transaction_conflict" || len(reported) != 0 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t)
	})
	t.Run("an unresolved request is outside the rule", func(t *testing.T) {
		status, code, reported := write(t, true, false, http.StatusNotFound, "user_not_found")
		if status != http.StatusNotFound || code != "user_not_found" || len(reported) != 0 {
			t.Fatalf("got %d %s reported=%v", status, code, reported)
		}
		emitted(t)
	})
}
