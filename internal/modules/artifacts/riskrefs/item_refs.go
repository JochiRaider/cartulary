package riskrefs

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func RiskRefItemRef(riskRefID uuid.UUID) string {
	return "risk_ref:" + riskRefID.String()
}

func ParseRiskRefItemRef(itemRef string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, "risk_ref:") {
		return uuid.UUID{}, fmt.Errorf("invalid risk ref item ref")
	}
	value := strings.TrimPrefix(itemRef, "risk_ref:")
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return uuid.UUID{}, fmt.Errorf("invalid risk ref item ref")
	}
	return parsed, nil
}
