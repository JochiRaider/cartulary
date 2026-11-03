package crossownertransaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func assertParticipantPrecedence(t *testing.T) {
	for _, test := range []struct {
		name     string
		count    int
		mutation string
		want     string
	}{
		{"empty", 0, "", ""}, {"nil", 0, "nil", "extension_validation_result_invalid"},
		{"semantic", 1, "", "cross_owner_transaction_validation_failed"},
		{"256", 256, "", "cross_owner_transaction_validation_failed"},
		{"257", 257, "", "extension_validation_result_invalid"}, {"4096", 4096, "", "extension_validation_result_invalid"},
		{"4097 malformed", 4097, "malformed", "extension_diagnostic_overflow"},
		{"5000 malformed", 5000, "malformed", "extension_diagnostic_overflow"},
		{"invocation", 4097, "invoke", "extension_admission_validation_failed"},
		{"panic", 0, "panic", "extension_admission_validation_failed"},
		{"duplicate bytes", 1, "duplicate", "extension_validation_result_invalid"},
		{"invalid unicode", 1, "unicode", "extension_validation_result_invalid"},
		{"unknown finding", 1, "private", "extension_validation_result_invalid"},
		{"duplicate finding", 2, "same", "extension_validation_result_invalid"},
	} {
		t.Run("precedence "+test.name, func(t *testing.T) {
			events := []string{}
			tx := &transactionFake{events: &events, commitOutcome: CommitProven}
			coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, nil)
			coordinator.validationPolicy = orderedFindingPolicy{}
			result := Valid()
			for i := 0; i < test.count; i++ {
				result.Findings = append(result.Findings, Finding{Path: fmt.Sprintf("$.value_%04d", i), ReasonCode: "fixture_invalid", Message: "Invalid value.", Details: []byte("{}")})
			}
			if test.count > 0 {
				result.Status = "invalid"
			}
			participant := &participantFake{id: "a", events: &events, validation: result}
			switch test.mutation {
			case "nil":
				participant.validation.Findings = nil
			case "malformed":
				participant.validation.Status = "private"
				participant.validation.Findings[0].Details = []byte("not-json")
			case "invoke":
				participant.validationErr = errors.New("private participant bytes")
			case "panic":
				participant.validationPanic = true
			case "duplicate":
				participant.validation.Findings[0].Details = []byte(`{"x":1,"x":2}`)
			case "unicode":
				participant.validation.Findings[0].Details = []byte(`{"x":"\ud800"}`)
			case "private":
				participant.validation.Findings[0].Message = "private participant bytes"
			case "same":
				participant.validation.Findings[1] = participant.validation.Findings[0]
			}
			_, err := coordinator.Execute(context.Background(), operation(participant))
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *ValidationError
			if !errors.Is(err, ErrValidation) || !errors.As(err, &failure) || failure.ReasonCode != test.want {
				t.Fatalf("classification = %v; want %s", err, test.want)
			}
			if test.want == "extension_diagnostic_overflow" {
				if len(failure.Findings) != 1 {
					t.Fatal("overflow omitted its single bounded finding")
				}
				var details map[string]any
				if err := json.Unmarshal(failure.Findings[0].Details, &details); err != nil || details["limit"] != float64(4096) || details["actual"] != float64(4097) || details["phase"] != "transaction" {
					t.Fatalf("invalid overflow details: %v", details)
				}
			}
			if strings.Contains(fmt.Sprint(failure), "private") || !tx.rolledBack || tx.commitCalls != 0 || contains(events, "write:a") {
				t.Fatalf("unsafe failure/transaction outcome: %v", err)
			}
		})
	}
}

type orderedFindingPolicy struct{ validationPolicyFixture }

func (orderedFindingPolicy) AcceptFinding(_ Descriptor, f Finding) bool {
	return strings.HasPrefix(f.Path, "$.value_") && f.ReasonCode == "fixture_invalid" && f.Message == "Invalid value." && string(f.Details) == "{}"
}
