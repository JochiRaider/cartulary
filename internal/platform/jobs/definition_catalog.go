package jobs

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

type ExtensionResourceRefContract struct {
	Kind    string
	MaxRefs int
}

// ExtensionPolicy contains only owner-specific policy. Generic job identity
// remains on Definition so future non-extension owners do not need fake
// extension metadata.
type ExtensionPolicy struct {
	OwnerProfileID string
	OperationKind  string
	ContractSHA256 string
	ProofRequired  bool
	MaxProofBytes  int
	ResourceRefs   []ExtensionResourceRefContract
}

// Definition is one immutable job-kind binding.
type Definition struct {
	JobKind        string
	ProgressUnitID string
	HandlerName    string
	Extension      *ExtensionPolicy
}

// Catalog is an immutable definition catalog shared by every Jobs service.
type Catalog struct {
	byKind map[string]Definition
}

func NewCatalog(definitions []Definition) (*Catalog, error) {
	indexed := make(map[string]Definition, len(definitions))
	for _, definition := range definitions {
		if definition.JobKind == "" || len(definition.JobKind) > 191 || !safeJobToken(definition.JobKind) ||
			!validProgressUnitID(definition.ProgressUnitID) || definition.HandlerName == "" || len(definition.HandlerName) > 191 ||
			!safeJobToken(definition.HandlerName) {
			return nil, fmt.Errorf("%w: incomplete job definition %q", ErrInvalidJobDefinition, definition.JobKind)
		}
		if _, duplicate := indexed[definition.JobKind]; duplicate {
			return nil, fmt.Errorf("%w: duplicate job definition %q", ErrInvalidJobDefinition, definition.JobKind)
		}
		clone, err := cloneDefinition(definition)
		if err != nil {
			return nil, err
		}
		indexed[definition.JobKind] = clone
	}
	if len(indexed) == 0 {
		return nil, fmt.Errorf("%w: empty job definition catalog", ErrInvalidJobDefinition)
	}
	return &Catalog{byKind: indexed}, nil
}

func cloneDefinition(definition Definition) (Definition, error) {
	clone := definition
	if definition.Extension == nil {
		return clone, nil
	}
	policy := *definition.Extension
	if policy.OwnerProfileID == "" || policy.OperationKind == "" ||
		!lowerHexSHA256(policy.ContractSHA256) || !policy.ProofRequired ||
		policy.MaxProofBytes < 1 || policy.MaxProofBytes > 1048576 {
		return Definition{}, fmt.Errorf("%w: incomplete extension policy %q", ErrInvalidJobDefinition, definition.JobKind)
	}
	policy.ResourceRefs = append([]ExtensionResourceRefContract(nil), definition.Extension.ResourceRefs...)
	totalRefs := 0
	previousKind := ""
	for _, resourceRef := range policy.ResourceRefs {
		if resourceRef.Kind == "" || resourceRef.Kind <= previousKind ||
			resourceRef.MaxRefs < 1 || resourceRef.MaxRefs > 1024 {
			return Definition{}, fmt.Errorf("%w: invalid extension resource contract %q", ErrInvalidJobDefinition, resourceRef.Kind)
		}
		totalRefs += resourceRef.MaxRefs
		if totalRefs > 1024 {
			return Definition{}, fmt.Errorf("%w: extension resource aggregate limit exceeded", ErrInvalidJobDefinition)
		}
		previousKind = resourceRef.Kind
	}
	clone.Extension = &policy
	return clone, nil
}

func (catalog *Catalog) definition(jobKind string) (Definition, bool) {
	if catalog == nil {
		return Definition{}, false
	}
	definition, present := catalog.byKind[jobKind]
	if !present {
		return Definition{}, false
	}
	clone, err := cloneDefinition(definition)
	if err != nil {
		return Definition{}, false
	}
	return clone, true
}

func (catalog *Catalog) handlerNames() []string {
	if catalog == nil {
		return nil
	}
	unique := map[string]struct{}{}
	for _, definition := range catalog.byKind {
		unique[definition.HandlerName] = struct{}{}
	}
	result := make([]string, 0, len(unique))
	for name := range unique {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func (catalog *Catalog) hasHandlerName(handlerName string) bool {
	if catalog == nil || handlerName == "" {
		return false
	}
	for _, definition := range catalog.byKind {
		if definition.HandlerName == handlerName {
			return true
		}
	}
	return false
}

// ValidateStorageCatalog is the startup compatibility gate. It reports only
// bounded kind/status tokens and never exposes identifiers or payload data.
func (m *Manager) ValidateStorageCatalog(ctx context.Context) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	rows, err := m.pool.Query(ctx, `
SELECT job_kind, progress_unit_id, handler_name, status
  FROM jobs
 WHERE status IN ('queued', 'running', 'cancel_requested')
 ORDER BY job_kind, status
`)
	if err != nil {
		return err
	}
	defer rows.Close()
	invalidCount := 0
	tokens := map[string]struct{}{}
	for rows.Next() {
		var jobKind, progressUnitID, handlerName *string
		var status string
		if err := rows.Scan(&jobKind, &progressUnitID, &handlerName, &status); err != nil {
			return err
		}
		valid := false
		if jobKind != nil && progressUnitID != nil && handlerName != nil {
			definition, present := m.catalog.definition(*jobKind)
			valid = present && definition.ProgressUnitID == *progressUnitID && definition.HandlerName == *handlerName
		}
		if valid {
			continue
		}
		invalidCount++
		if len(tokens) < 10 {
			tokens[safeJobDiagnosticToken(jobKind, status)] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if invalidCount == 0 {
		return nil
	}
	bounded := make([]string, 0, len(tokens))
	for token := range tokens {
		bounded = append(bounded, token)
	}
	sort.Strings(bounded)
	return fmt.Errorf("%w: invalid_mutable_count=%d kind_status_tokens=%s", ErrStorageIncompatible, invalidCount, strings.Join(bounded, ","))
}

func safeJobDiagnosticToken(jobKind *string, status string) string {
	kind := "missing"
	if jobKind != nil && safeJobToken(*jobKind) {
		kind = *jobKind
	}
	if len(kind) > 63 {
		kind = kind[:63]
	}
	if !safeJobToken(status) {
		status = "invalid"
	}
	if len(status) > 31 {
		status = status[:31]
	}
	return kind + ":" + status
}

func safeJobToken(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') ||
			(character >= '0' && character <= '9') ||
			character == '_' || character == '.' || character == '-') {
			return false
		}
	}
	return true
}

func validProgressUnitID(value string) bool {
	if len(value) == 0 || len(value) > 191 {
		return false
	}
	segments := strings.Split(value, ".")
	if len(segments) < 3 {
		return false
	}
	for _, segment := range segments[:len(segments)-1] {
		if len(segment) == 0 || len(segment) > 63 || segment[0] < 'a' || segment[0] > 'z' {
			return false
		}
		for _, character := range segment[1:] {
			if !((character >= 'a' && character <= 'z') ||
				(character >= '0' && character <= '9') || character == '_') {
				return false
			}
		}
	}
	version := segments[len(segments)-1]
	if len(version) < 2 || version[0] != 'v' || version[1] < '1' || version[1] > '9' {
		return false
	}
	for _, character := range version[2:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
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
