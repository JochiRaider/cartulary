package workbookroutetest

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/workbook"
)

// RequireMutationResult checks a native outcome without translating rejection into an error.
func RequireMutationResult(t testing.TB, outcome workbook.MutationOutcome, err error) workbook.MutationResult {
	t.Helper()
	if err != nil {
		t.Fatalf("execute mutation: %v", err)
	}
	if err := outcome.Validate(); err != nil {
		t.Fatalf("invalid mutation outcome: %v", err)
	}
	result, ok := outcome.Result()
	if !ok {
		failure, _ := outcome.Failure()
		t.Fatalf("expected successful mutation, got %#v", failure)
	}
	return result
}
