package jobs

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type ExtensionResourceRefContract struct {
	Kind    string
	MaxRefs int
}

type jobDefinitionCatalog struct {
	byKind map[string]ExtensionJobContract
}

func newJobDefinitionCatalog(contracts []ExtensionJobContract) (*jobDefinitionCatalog, error) {
	indexed := make(map[string]ExtensionJobContract, len(contracts))
	for _, contract := range contracts {
		if contract.OwnerProfileID == "" || contract.JobKind == "" ||
			!validProgressUnitID(contract.ProgressUnitID) ||
			contract.OperationKind == "" || contract.WorkerKind == "" ||
			!lowerHexSHA256(contract.ContractSHA256) || !contract.ProofRequired ||
			contract.MaxProofBytes < 1 || contract.MaxProofBytes > 1048576 {
			return nil, fmt.Errorf("%w: incomplete extension job contract %q", ErrInvalidJobDefinition, contract.JobKind)
		}
		if _, duplicate := indexed[contract.JobKind]; duplicate {
			return nil, fmt.Errorf("%w: duplicate extension job contract %q", ErrInvalidJobDefinition, contract.JobKind)
		}
		totalRefs := 0
		previousKind := ""
		for _, resourceRef := range contract.ResourceRefs {
			if resourceRef.Kind == "" || resourceRef.Kind <= previousKind ||
				resourceRef.MaxRefs < 1 || resourceRef.MaxRefs > 1024 {
				return nil, fmt.Errorf("%w: invalid extension resource contract %q", ErrInvalidJobDefinition, resourceRef.Kind)
			}
			totalRefs += resourceRef.MaxRefs
			if totalRefs > 1024 {
				return nil, fmt.Errorf("%w: extension resource aggregate limit exceeded", ErrInvalidJobDefinition)
			}
			previousKind = resourceRef.Kind
		}
		clone := contract
		clone.ResourceRefs = append([]ExtensionResourceRefContract(nil), contract.ResourceRefs...)
		indexed[contract.JobKind] = clone
	}
	if len(indexed) == 0 {
		return nil, fmt.Errorf("%w: empty job definition catalog", ErrInvalidJobDefinition)
	}
	return &jobDefinitionCatalog{byKind: indexed}, nil
}

func (catalog *jobDefinitionCatalog) contains(contracts []ExtensionJobContract) bool {
	if catalog == nil || len(catalog.byKind) != len(contracts) {
		return false
	}
	for _, contract := range contracts {
		stored, present := catalog.byKind[contract.JobKind]
		if !present || !reflect.DeepEqual(stored, contract) {
			return false
		}
	}
	return true
}

// ValidateStorageCatalog is the startup compatibility gate. It inspects only
// mutable jobs and reports bounded kind/status tokens; job identifiers,
// payloads, progress-unit identifiers, and handler diagnostics never enter the
// error surface.
func (m *Manager) ValidateStorageCatalog(ctx context.Context) error {
	if err := m.ensureConfigured(); err != nil {
		return err
	}
	if m.definitions == nil || len(m.definitions.byKind) == 0 {
		return fmt.Errorf("%w: definition catalog unavailable", ErrStorageIncompatible)
	}
	rows, err := m.pool.Query(ctx, `
SELECT job_kind, progress_unit_id, status
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
		var jobKind *string
		var progressUnitID *string
		var status string
		if err := rows.Scan(&jobKind, &progressUnitID, &status); err != nil {
			return err
		}
		valid := false
		if jobKind != nil && progressUnitID != nil {
			definition, present := m.definitions.byKind[*jobKind]
			valid = present && definition.ProgressUnitID == *progressUnitID
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

type ExtensionJobContract struct {
	OwnerProfileID string
	JobKind        string
	ProgressUnitID string
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
	var catalog *jobDefinitionCatalog
	var err error
	if m.transactions != nil {
		catalog, err = m.transactions.configureDefinitions(contracts)
	} else {
		catalog, err = newJobDefinitionCatalog(contracts)
	}
	if err != nil {
		return err
	}
	m.definitions = catalog
	return nil
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

func (m *Manager) ExtensionContract(jobKind string) (ExtensionJobContract, bool) {
	if m == nil {
		return ExtensionJobContract{}, false
	}
	if m.definitions == nil {
		return ExtensionJobContract{}, false
	}
	contract, present := m.definitions.byKind[jobKind]
	contract.ResourceRefs = append([]ExtensionResourceRefContract(nil), contract.ResourceRefs...)
	return contract, present
}
