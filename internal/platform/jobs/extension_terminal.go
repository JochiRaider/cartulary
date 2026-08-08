package jobs

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

func CanonicalExtensionTerminalSuccess(definition Definition, summary *ResultSummary) (*ResultSummary, json.RawMessage, json.RawMessage, string, error) {
	if summary == nil || summary.Code == "" || summary.Message == "" {
		return nil, nil, nil, "", fmt.Errorf("%w: extension terminal success is incomplete", ErrInvalidJobDefinition)
	}
	normalized := &ResultSummary{
		Code:         summary.Code,
		Message:      summary.Message,
		ResourceRefs: append([]ResourceRef{}, summary.ResourceRefs...),
	}
	sort.Slice(normalized.ResourceRefs, func(left int, right int) bool {
		if normalized.ResourceRefs[left].Kind != normalized.ResourceRefs[right].Kind {
			return normalized.ResourceRefs[left].Kind < normalized.ResourceRefs[right].Kind
		}
		if normalized.ResourceRefs[left].ID != normalized.ResourceRefs[right].ID {
			return normalized.ResourceRefs[left].ID < normalized.ResourceRefs[right].ID
		}
		return normalized.ResourceRefs[left].Route < normalized.ResourceRefs[right].Route
	})
	if definition.Extension == nil {
		return nil, nil, nil, "", fmt.Errorf("%w: extension policy is missing", ErrInvalidJobDefinition)
	}
	limits := make(map[string]int, len(definition.Extension.ResourceRefs))
	for _, resourceContract := range definition.Extension.ResourceRefs {
		if resourceContract.Kind == "" || resourceContract.MaxRefs < 1 {
			return nil, nil, nil, "", fmt.Errorf("%w: invalid resource reference contract", ErrInvalidJobDefinition)
		}
		limits[resourceContract.Kind] = resourceContract.MaxRefs
	}
	counts := map[string]int{}
	previous := ResourceRef{}
	for index, resourceRef := range normalized.ResourceRefs {
		if resourceRef.Kind == "" || resourceRef.ID == "" ||
			(resourceRef.Route != "" && resourceRef.Route[0] != '/') {
			return nil, nil, nil, "", fmt.Errorf("%w: invalid extension resource reference", ErrInvalidJobDefinition)
		}
		limit, admitted := limits[resourceRef.Kind]
		if !admitted {
			return nil, nil, nil, "", fmt.Errorf("%w: undeclared extension resource reference %q", ErrInvalidJobDefinition, resourceRef.Kind)
		}
		counts[resourceRef.Kind]++
		if counts[resourceRef.Kind] > limit {
			return nil, nil, nil, "", fmt.Errorf("%w: extension resource reference limit exceeded", ErrInvalidJobDefinition)
		}
		if index > 0 && previous.Kind == resourceRef.Kind && previous.ID == resourceRef.ID {
			return nil, nil, nil, "", fmt.Errorf("%w: duplicate extension resource reference", ErrInvalidJobDefinition)
		}
		previous = resourceRef
	}
	terminalJSON, err := json.Marshal(normalized)
	if err != nil {
		return nil, nil, nil, "", err
	}
	resourceRefsJSON, err := json.Marshal(normalized.ResourceRefs)
	if err != nil {
		return nil, nil, nil, "", err
	}
	digest := sha256.Sum256(terminalJSON)
	return normalized, terminalJSON, resourceRefsJSON, fmt.Sprintf("%x", digest[:]), nil
}
