package queryengine

import (
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

// Engine is the private, fail-closed semantic-intent registry. Compiled plans
// remain separate private values and are matched to these intents by runtime
// provider validation.
type Engine struct {
	intents map[string]providercontract.SurfaceIntent
}

func New(intents []providercontract.SurfaceIntent) (*Engine, error) {
	if len(intents) == 0 {
		return nil, fmt.Errorf("projection query intents are empty")
	}
	engine := &Engine{intents: make(map[string]providercontract.SurfaceIntent, len(intents))}
	for _, intent := range intents {
		if strings.TrimSpace(intent.ViewSchemaID) == "" {
			return nil, fmt.Errorf("projection query intent has empty view_schema_id")
		}
		if len(intent.FieldKeys) == 0 {
			return nil, fmt.Errorf("projection query intent %q has no field keys", intent.ViewSchemaID)
		}
		if _, exists := engine.intents[intent.ViewSchemaID]; exists {
			return nil, fmt.Errorf("duplicate projection query intent %q", intent.ViewSchemaID)
		}
		seenFields := make(map[string]struct{}, len(intent.FieldKeys))
		for _, fieldKey := range intent.FieldKeys {
			if strings.TrimSpace(fieldKey) == "" {
				return nil, fmt.Errorf("projection query intent %q has empty field key", intent.ViewSchemaID)
			}
			if _, exists := seenFields[fieldKey]; exists {
				return nil, fmt.Errorf("projection query intent %q has duplicate field key %q", intent.ViewSchemaID, fieldKey)
			}
			seenFields[fieldKey] = struct{}{}
		}
		engine.intents[intent.ViewSchemaID] = intent.Clone()
	}
	return engine, nil
}

func (engine *Engine) Ready() bool {
	return engine != nil && len(engine.intents) > 0
}
