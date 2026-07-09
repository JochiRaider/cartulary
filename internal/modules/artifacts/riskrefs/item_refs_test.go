package riskrefs

import (
	"testing"

	"github.com/google/uuid"
)

func TestRiskRefItemRefIsCanonical(t *testing.T) {
	riskRefID := uuid.MustParse("10000000-0000-4000-8000-000000000010")
	if got := RiskRefItemRef(riskRefID); got != "risk_ref:10000000-0000-4000-8000-000000000010" {
		t.Fatalf("risk ref = %q", got)
	}
	parsed, err := ParseRiskRefItemRef(RiskRefItemRef(riskRefID))
	if err != nil {
		t.Fatalf("parse risk ref: %v", err)
	}
	if parsed != riskRefID {
		t.Fatalf("parsed risk ref = %s", parsed)
	}
}

func TestRiskRefItemRefParsingIsStrict(t *testing.T) {
	riskRefID := uuid.MustParse("10000000-0000-4000-8000-000000000010")
	for _, itemRef := range []string{
		"",
		"record_ref:" + riskRefID.String(),
		"risk_ref:" + riskRefID.String() + ":extra",
		"risk_ref:" + riskRefID.String() + " ",
	} {
		if _, err := ParseRiskRefItemRef(itemRef); err == nil {
			t.Fatalf("ParseRiskRefItemRef(%q) succeeded", itemRef)
		}
	}
}
