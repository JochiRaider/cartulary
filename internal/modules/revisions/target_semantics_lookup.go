package revisions

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func (catalog *TargetSemanticsCatalog) DescribeMutation(mutation StoredMutation) (HistoryTargetDescription, error) {
	if catalog == nil {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, mutation.TargetKind)
	}
	entry, ok := catalog.byTargetKind[mutation.TargetKind]
	if !ok {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, mutation.TargetKind)
	}
	if entry.historyValidator != nil {
		if err := entry.historyValidator.ValidateHistoryMutation(mutation); err != nil {
			return HistoryTargetDescription{}, fmt.Errorf("%w: target kind %q history value", ErrInvalidTargetSemantics, mutation.TargetKind)
		}
	}
	return entry.history.DescribeMutation(mutation)
}

func (catalog *TargetSemanticsCatalog) DescribeValues(targetKind string, targetID string, operationKind string, beforeValue any, afterValue any) (HistoryTargetDescription, error) {
	beforeObject, err := historyMutationObject(beforeValue)
	if err != nil {
		return HistoryTargetDescription{}, err
	}
	afterObject, err := historyMutationObject(afterValue)
	if err != nil {
		return HistoryTargetDescription{}, err
	}
	return catalog.DescribeMutation(StoredMutation{
		TargetKind:    targetKind,
		TargetID:      targetID,
		OperationKind: operationKind,
		BeforeValue:   beforeObject,
		AfterValue:    afterObject,
	})
}

func (catalog *TargetSemanticsCatalog) historyRecordReferences(mutation portableChangeSetMutation, incidentID uuid.UUID) ([]HistoryRecordReference, error) {
	entry, ok := catalog.targetEntry(mutation.TargetKind)
	if !ok {
		return nil, ErrMissingTargetSemantics
	}
	if entry.historyReferences == nil {
		return nil, ErrMissingTargetSemantics
	}
	// Non-object sides still fail full history admission. Valid object references
	// must be resolved before lower-priority shape or transition checks.
	before, _ := mutation.BeforeValue.(map[string]any)
	after, _ := mutation.AfterValue.(map[string]any)
	references, err := entry.historyReferences(StoredMutation{
		TargetKind: mutation.TargetKind, TargetID: mutation.TargetID, OperationKind: mutation.OperationKind,
		BeforeValue: before, AfterValue: after,
	}, incidentID)
	if err != nil {
		return nil, ErrInvalidTargetSemantics
	}
	return references, nil
}

func historyMutationObject(value any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	if object, ok := value.(map[string]any); ok {
		return object, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: mutation value is not JSON: %v", ErrInvalidTargetSemantics, err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil || object == nil {
		return nil, fmt.Errorf("%w: mutation value is not an object", ErrInvalidTargetSemantics)
	}
	return object, nil
}

func (catalog *TargetSemanticsCatalog) hasTargetKind(targetKind string) bool {
	if catalog == nil {
		return false
	}
	_, ok := catalog.byTargetKind[targetKind]
	return ok
}

func (catalog *TargetSemanticsCatalog) requiresGenericPortableTargetID(targetKind string) bool {
	if catalog == nil {
		return true
	}
	entry, ok := catalog.byTargetKind[targetKind]
	return !ok || entry.historyValidator == nil
}

func (catalog *TargetSemanticsCatalog) targetKinds() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, 0, len(catalog.byTargetKind))
	for targetKind := range catalog.byTargetKind {
		result = append(result, targetKind)
	}
	sort.Strings(result)
	return result
}

// rowTargetKinds resolves all admitted spellings for one physical row family.
func (catalog *TargetSemanticsCatalog) rowTargetKinds(recordType string) []string {
	var result []string
	for _, kind := range catalog.targetKinds() {
		if _, err := catalog.rowProvider(kind, recordType); err == nil {
			result = append(result, kind)
		}
	}
	return result
}

func (catalog *TargetSemanticsCatalog) dispatchClass(targetKind string) (rollbackcontract.DispatchClass, error) {
	entry, ok := catalog.targetEntry(targetKind)
	if !ok {
		return "", fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, targetKind)
	}
	return entry.dispatchClass, nil
}

func (catalog *TargetSemanticsCatalog) rowProvider(targetKind string, recordType string) (rollbackcontract.RowSourceProvider, error) {
	entry, ok := catalog.targetEntry(targetKind)
	if !ok || entry.dispatchClass != rollbackcontract.DispatchRow {
		return nil, fmt.Errorf("%w: row target kind %q", ErrInvalidTargetSemantics, targetKind)
	}
	provider, ok := entry.rowProviders[recordType]
	if !ok || nilRowSourceProvider(provider) {
		return nil, fmt.Errorf("%w: target kind %q record type %q", ErrInvalidTargetSemantics, targetKind, recordType)
	}
	return provider, nil
}

func (catalog *TargetSemanticsCatalog) nonRowProvider(targetKind string) (rollbackcontract.NonRowTargetProvider, error) {
	entry, ok := catalog.targetEntry(targetKind)
	if !ok || entry.dispatchClass != rollbackcontract.DispatchNonRow || nilNonRowTargetProvider(entry.nonRowProvider) {
		return nil, fmt.Errorf("%w: non-row target kind %q", ErrInvalidTargetSemantics, targetKind)
	}
	return entry.nonRowProvider, nil
}

func (catalog *TargetSemanticsCatalog) defaultRowTargetKind(recordType string) (string, error) {
	if catalog == nil {
		return "", fmt.Errorf("%w: record type %q", ErrMissingTargetSemantics, recordType)
	}
	targetKind, ok := catalog.defaultRowTargetByRecordType[recordType]
	if !ok {
		return "", fmt.Errorf("%w: record type %q", ErrMissingTargetSemantics, recordType)
	}
	return targetKind, nil
}

func (catalog *TargetSemanticsCatalog) targetEntry(targetKind string) (compiledTargetSemantics, bool) {
	if catalog == nil {
		return compiledTargetSemantics{}, false
	}
	entry, ok := catalog.byTargetKind[targetKind]
	return entry, ok
}
