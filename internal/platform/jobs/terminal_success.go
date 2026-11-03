package jobs

import (
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

func CanonicalTerminalSuccess(definition Definition, summary *ResultSummary) (*ResultSummary, json.RawMessage, json.RawMessage, string, error) {
	if summary == nil {
		return nil, nil, nil, "", fmt.Errorf("%w: missing terminal success", ErrInvalidJobDefinition)
	}
	normalized, terminal, resources, digest, err := jobcontract.CanonicalSuccess(definition.Terminal, *summary)
	if err != nil {
		return nil, nil, nil, "", fmt.Errorf("%w: %v", ErrInvalidJobDefinition, err)
	}
	return &normalized, terminal, resources, digest, nil
}
