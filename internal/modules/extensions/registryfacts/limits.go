package registryfacts

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const (
	ProfileLimit           = 256
	FragmentLimit          = 512
	OwnerFactLimit         = 65536
	OwnerFragmentByteLimit = 1 << 20
	RegistryByteLimit      = 64 << 20
)

// LimitError carries only a fixed quantity name and observed numeric bounds.
// Adapters retain their bootstrap or registered diagnostic ownership.
type LimitError struct {
	Quantity      string
	Limit, Actual int
}

func (e *LimitError) Error() string { return "extension_registry_limit_exceeded" }

func (e *LimitError) Finding() OwnerFinding {
	return OwnerFinding{ConditionID: RegistryLimitCondition, Path: "$", Details: map[string]any{"phase": "registry_generation", "limit": e.Limit, "actual": e.Actual}}
}

// CheckOwnerInputLimits counts declarations before normalization or diagnostic
// construction. Empty fragments are retained. Profile capacity counts distinct
// Core recognition identities; duplicate facts remain collision declarations.
func CheckOwnerInputLimits(fragments []map[string]any) error {
	if len(fragments) > FragmentLimit {
		return &LimitError{"owner_fragments", FragmentLimit, len(fragments)}
	}
	total := 0
	for _, fragment := range fragments {
		rows, ok := fragment["facts"].([]any)
		if !ok {
			return ErrInvalidFragment
		}
		total += len(rows)
	}
	if total > OwnerFactLimit {
		return &LimitError{"owner_facts", OwnerFactLimit, total}
	}
	profiles := map[string]bool{}
	for _, fragment := range fragments {
		if fragment["owner_id"] == "core00" {
			for _, raw := range fragment["facts"].([]any) {
				fact, ok := raw.(map[string]any)
				if !ok {
					continue
				}
				profile, ok := fact["profile_id"].(string)
				if ok && profilePattern.MatchString(profile) && fact["fact_kind"] == "recognized_profile" {
					profiles[profile] = true
				}
				if len(profiles) > ProfileLimit {
					return &LimitError{"profiles", ProfileLimit, len(profiles)}
				}
			}
		}
	}
	maximum := 0
	for _, fragment := range fragments {
		body, err := canonicaljson.MarshalExtension(fragment)
		if err != nil {
			return ErrInvalidFragment
		}
		if len(body) > maximum {
			maximum = len(body)
		}
	}
	if maximum > OwnerFragmentByteLimit {
		return &LimitError{"owner_fragment_bytes", OwnerFragmentByteLimit, maximum}
	}
	return nil
}

func CheckCanonicalLimit(value any, quantity string, limit int) error {
	body, err := canonicaljson.MarshalExtension(value)
	if err != nil {
		return ErrInvalidFragment
	}
	if len(body) > limit {
		return &LimitError{quantity, limit, len(body)}
	}
	return nil
}

// CheckInputSchemaLimits prevents a structurally valid schema projection from
// silently narrowing or widening the adopted owner-input capacities.
func CheckInputSchemaLimits(schemas []map[string]any) error {
	want := map[string]struct {
		field   string
		maximum int
	}{
		"cartulary.extension_owner_fragment.v3":       {"facts", OwnerFactLimit},
		"cartulary.extension_owner_input_registry.v2": {"owner_fragments", FragmentLimit},
		"cartulary.extension_profile_registry.v1":     {"profiles", ProfileLimit},
	}
	for _, schema := range schemas {
		id, _ := schema["$id"].(string)
		rule, ok := want[id]
		if !ok {
			continue
		}
		properties, _ := schema["properties"].(map[string]any)
		field, _ := properties[rule.field].(map[string]any)
		if field["type"] != "array" || !limitIntegerEquals(field["minItems"], 0) || !limitIntegerEquals(field["maxItems"], rule.maximum) {
			return fmt.Errorf("owner input schema limits invalid")
		}
		delete(want, id)
	}
	if len(want) != 0 {
		return fmt.Errorf("owner input schema limits unresolved")
	}
	return nil
}

func limitIntegerEquals(value any, want int) bool {
	switch value := value.(type) {
	case int:
		return value == want
	case json.Number:
		return string(value) == strconv.Itoa(want)
	default:
		return false
	}
}
