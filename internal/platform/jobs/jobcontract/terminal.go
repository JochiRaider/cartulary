// Package jobcontract owns the typed Core job value contracts used by live
// Jobs execution and by Extensions' inactive reconciliation. It performs no IO.
package jobcontract

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const TerminalSuccessSchema = "cartulary.common_job_terminal_success.v2"
const ResourceRefSchema = "cartulary.common_job_resource_ref.v1"
const ResourceIDSchema = "cartulary.common_job_resource_ref_id.v1"

type ResourceRef struct {
	Kind  string `json:"kind"`
	ID    string `json:"id"`
	Route string `json:"route,omitempty"`
}

type Success struct {
	Code         string        `json:"code"`
	Message      string        `json:"message"`
	ResourceRefs []ResourceRef `json:"resource_refs,omitempty"`
}

type ResourceContract struct {
	Kind     string `json:"resource_ref_kind"`
	SchemaID string `json:"resource_id_schema_id"`
	MinRefs  int    `json:"min_refs"`
	MaxRefs  int    `json:"max_refs"`
}

// TerminalContract is the pure selected-owner policy for a successful Job.
// Cancellation has its own fixed outcome and is never a successful result.
type TerminalContract struct {
	SchemaID     string             `json:"terminal_result_schema_id"`
	SuccessCodes []string           `json:"terminal_success_codes"`
	Resources    []ResourceContract `json:"resource_ref_contracts"`
}

func ValidateTerminalContract(contract TerminalContract) error {
	if _, ok := ResolveTerminal(contract.SchemaID); !ok {
		return errors.New("unresolved terminal schema")
	}
	if len(contract.SuccessCodes) < 1 || len(contract.SuccessCodes) > 64 {
		return errors.New("invalid terminal success code set")
	}
	previous := ""
	for _, code := range contract.SuccessCodes {
		if len(code) < 1 || len(code) > 64 || code <= previous || code == "job_canceled" || code[0] < 'a' || code[0] > 'z' {
			return errors.New("invalid terminal success code set")
		}
		for _, c := range code {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
				return errors.New("invalid terminal success code set")
			}
		}
		previous = code
	}
	return ValidateResourceContracts(contract.Resources)
}

func ValidateTerminalSuccess(contract TerminalContract, summary Success) error {
	if err := ValidateCurrentSuccess(summary); err != nil {
		return err
	}
	return validateSelectedSuccess(contract, summary)
}

func validateSelectedSuccess(contract TerminalContract, summary Success) error {
	if err := ValidateTerminalContract(contract); err != nil {
		return err
	}
	if !slices.Contains(contract.SuccessCodes, summary.Code) {
		return errors.New("undeclared terminal success code")
	}
	limits := make(map[string]ResourceContract, len(contract.Resources))
	counts := make(map[string]int, len(contract.Resources))
	for _, resource := range contract.Resources {
		limits[resource.Kind] = resource
	}
	for _, ref := range summary.ResourceRefs {
		limit, ok := limits[ref.Kind]
		if !ok {
			return errors.New("undeclared resource kind")
		}
		counts[ref.Kind]++
		if counts[ref.Kind] > limit.MaxRefs {
			return errors.New("resource count exceeds declared maximum")
		}
	}
	for _, resource := range contract.Resources {
		if counts[resource.Kind] < resource.MinRefs {
			return errors.New("resource count below declared minimum")
		}
	}
	return nil
}

// Resolution returns actual typed validators. Unknown schema identifiers never
// select a default reader or a validator inferred from stored examples.
func ResolveTerminal(schema string) (func([]byte) (Success, error), bool) {
	if schema != TerminalSuccessSchema {
		return nil, false
	}
	return decodeSuccess, true
}

func decodeSuccess(raw []byte) (Success, error) {
	var result Success
	value, err := canonicaljson.DecodeStrict(raw)
	if err != nil {
		return result, err
	}
	object, ok := value.(map[string]any)
	if !ok || !closed(object, "code", "message", "resource_refs") {
		return result, errors.New("invalid common-job success object")
	}
	if _, ok := object["code"].(string); !ok {
		return result, errors.New("missing success code")
	}
	if _, ok := object["message"].(string); !ok {
		return result, errors.New("missing success message")
	}
	if refs, present := object["resource_refs"]; present {
		rows, ok := refs.([]any)
		if !ok || len(rows) > 1024 {
			return result, errors.New("invalid success references")
		}
		for _, raw := range rows {
			row, ok := raw.(map[string]any)
			if !ok || !closed(row, "kind", "id", "route") {
				return result, errors.New("invalid success reference")
			}
			for _, key := range []string{"kind", "id"} {
				if _, ok := row[key].(string); !ok {
					return result, errors.New("missing reference identity")
				}
			}
			if route, present := row["route"]; present {
				if text, ok := route.(string); !ok || text == "" {
					return result, errors.New("invalid reference route")
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return result, err
	}
	return result, validateSuccess(result, false)
}

func closed(object map[string]any, names ...string) bool {
	for key := range object {
		found := false
		for _, name := range names {
			if key == name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func ValidateResourceContracts(contracts []ResourceContract) error {
	if len(contracts) > 64 {
		return errors.New("too many resource contracts")
	}
	total, previous := 0, ""
	for _, c := range contracts {
		if c.SchemaID != ResourceIDSchema {
			return errors.New("unresolved resource identity schema")
		}
		if !knownResourceKind(c.Kind) || c.Kind <= previous || c.MinRefs < 0 || c.MinRefs > c.MaxRefs || c.MaxRefs < 1 || c.MaxRefs > 1024 {
			return errors.New("invalid resource contract")
		}
		total += c.MaxRefs
		previous = c.Kind
	}
	if total > 1024 {
		return errors.New("resource aggregate exceeds 1024")
	}
	return nil
}

func CanonicalSuccess(contract TerminalContract, summary Success) (Success, json.RawMessage, json.RawMessage, string, error) {
	decode, ok := ResolveTerminal(contract.SchemaID)
	if !ok {
		return Success{}, nil, nil, "", errors.New("unresolved terminal schema")
	}
	result, err := CanonicalCurrentSuccess(summary)
	if err != nil {
		return Success{}, nil, nil, "", err
	}
	if err := validateSelectedSuccess(contract, result); err != nil {
		return Success{}, nil, nil, "", err
	}
	encoded, err := canonicaljson.Marshal(result)
	if err != nil {
		return Success{}, nil, nil, "", err
	}
	if _, err = decode(encoded); err != nil {
		return Success{}, nil, nil, "", err
	}
	refs, err := canonicaljson.Marshal(result.ResourceRefs)
	if err != nil {
		return Success{}, nil, nil, "", err
	}
	digest := sha256.Sum256(encoded)
	return result, encoded, refs, fmt.Sprintf("%x", digest[:]), nil
}
