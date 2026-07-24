package jobs

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
)

type ExtensionResourceRefContract struct {
	Kind    string
	MaxRefs int
}

type ExtensionJobContract struct {
	OwnerProfileID string
	JobKind        string
	OperationKind  string
	WorkerKind     string
	ContractSHA256 string
	ProofRequired  bool
	MaxProofBytes  int
	ResourceRefs   []ExtensionResourceRefContract
}

func (m *Manager) ConfigureExtensionContracts(contracts []ExtensionJobContract) error {
	if m == nil {
		return ErrNotConfigured
	}
	indexed := make(map[string]ExtensionJobContract, len(contracts))
	for _, contract := range contracts {
		if contract.OwnerProfileID == "" || contract.JobKind == "" ||
			contract.OperationKind == "" || contract.WorkerKind == "" ||
			!lowerHexSHA256(contract.ContractSHA256) || !contract.ProofRequired ||
			contract.MaxProofBytes < 1 || contract.MaxProofBytes > 1048576 {
			return fmt.Errorf("%w: incomplete extension job contract %q", ErrInvalidJobDefinition, contract.JobKind)
		}
		if _, duplicate := indexed[contract.JobKind]; duplicate {
			return fmt.Errorf("%w: duplicate extension job contract %q", ErrInvalidJobDefinition, contract.JobKind)
		}
		totalRefs := 0
		previousKind := ""
		for _, resourceRef := range contract.ResourceRefs {
			if resourceRef.Kind == "" || resourceRef.Kind <= previousKind ||
				resourceRef.MaxRefs < 1 || resourceRef.MaxRefs > 1024 {
				return fmt.Errorf("%w: invalid extension resource contract %q", ErrInvalidJobDefinition, resourceRef.Kind)
			}
			totalRefs += resourceRef.MaxRefs
			if totalRefs > 1024 {
				return fmt.Errorf("%w: extension resource aggregate limit exceeded", ErrInvalidJobDefinition)
			}
			previousKind = resourceRef.Kind
		}
		clone := contract
		clone.ResourceRefs = append([]ExtensionResourceRefContract(nil), contract.ResourceRefs...)
		indexed[contract.JobKind] = clone
	}
	m.extensionContracts = indexed
	return nil
}

func lowerHexSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !((character >= '0' && character <= '9') || (character >= 'a' && character <= 'f')) {
			return false
		}
	}
	return true
}

func (m *Manager) ExtensionContract(jobKind string) (ExtensionJobContract, bool) {
	if m == nil {
		return ExtensionJobContract{}, false
	}
	contract, present := m.extensionContracts[jobKind]
	contract.ResourceRefs = append([]ExtensionResourceRefContract(nil), contract.ResourceRefs...)
	return contract, present
}

func CanonicalExtensionTerminalSuccess(contract ExtensionJobContract, summary *ResultSummary) (*ResultSummary, json.RawMessage, json.RawMessage, string, error) {
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
	limits := make(map[string]int, len(contract.ResourceRefs))
	for _, resourceContract := range contract.ResourceRefs {
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

func CanonicalJSONEqual(left, right []byte) bool {
	var leftValue any
	var rightValue any
	if json.Unmarshal(left, &leftValue) != nil || json.Unmarshal(right, &rightValue) != nil {
		return false
	}
	leftCanonical, leftErr := json.Marshal(leftValue)
	rightCanonical, rightErr := json.Marshal(rightValue)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftCanonical, rightCanonical)
}
