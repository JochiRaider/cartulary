package revisions

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/contractrevisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type RollbackDispatchClass string

const (
	RollbackDispatchRow    RollbackDispatchClass = "row"
	RollbackDispatchNonRow RollbackDispatchClass = "non_row"
)

type HistoryAddressability string

const (
	HistorySingleEntry                HistoryAddressability = "single_history_entry"
	HistoryNotIndividuallyAddressable HistoryAddressability = "not_individually_addressable"
)

var (
	ErrDuplicateTargetSemantics  = errors.New("revisions: duplicate target semantics")
	ErrMissingTargetSemantics    = errors.New("revisions: missing target semantics")
	ErrUnexpectedTargetSemantics = errors.New("revisions: unexpected target semantics")
	ErrInvalidTargetSemantics    = errors.New("revisions: invalid target semantics")
)

const TargetSemanticsRegistrySchemaID = "cartulary.revisions_target_semantics_registry.v1"

type StoredMutation struct {
	TargetKind  string
	TargetID    string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

type HistoryTargetDescription struct {
	TargetKind            string
	TargetID              string
	LogicalItemIdentity   string
	HistoryRecordIDs      []uuid.UUID
	HistoryEntryRecordIDs []uuid.UUID
}

type historySemanticsContract struct {
	directTargetID bool
	recordIDFields []string
	addressability HistoryAddressability
}

type HistoryTargetSemantics interface {
	DescribeMutation(StoredMutation) (HistoryTargetDescription, error)
	historyContract() historySemanticsContract
}

type configuredHistoryTargetSemantics struct {
	contract historySemanticsContract
}

func NewDirectRecordHistoryTargetSemantics(addressability HistoryAddressability) HistoryTargetSemantics {
	return configuredHistoryTargetSemantics{contract: historySemanticsContract{directTargetID: true, addressability: addressability}}
}

func NewFieldHistoryTargetSemantics(recordIDFields []string, addressability HistoryAddressability) HistoryTargetSemantics {
	return configuredHistoryTargetSemantics{contract: historySemanticsContract{
		recordIDFields: append([]string(nil), recordIDFields...),
		addressability: addressability,
	}}
}

func (semantics configuredHistoryTargetSemantics) historyContract() historySemanticsContract {
	return historySemanticsContract{
		directTargetID: semantics.contract.directTargetID,
		recordIDFields: append([]string(nil), semantics.contract.recordIDFields...),
		addressability: semantics.contract.addressability,
	}
}

func (semantics configuredHistoryTargetSemantics) DescribeMutation(mutation StoredMutation) (HistoryTargetDescription, error) {
	contract := semantics.historyContract()
	ids := make([]uuid.UUID, 0)
	if contract.directTargetID {
		recordID, err := uuid.Parse(mutation.TargetID)
		if err != nil || recordID == uuid.Nil {
			return HistoryTargetDescription{}, fmt.Errorf("%w: target %q id %q", ErrInvalidTargetSemantics, mutation.TargetKind, mutation.TargetID)
		}
		ids = append(ids, recordID)
	} else {
		for _, value := range []map[string]any{mutation.BeforeValue, mutation.AfterValue} {
			for _, field := range contract.recordIDFields {
				text, ok := value[field].(string)
				if !ok || text == "" {
					continue
				}
				recordID, err := uuid.Parse(text)
				if err != nil || recordID == uuid.Nil {
					return HistoryTargetDescription{}, fmt.Errorf("%w: target %q field %q", ErrInvalidTargetSemantics, mutation.TargetKind, field)
				}
				ids = append(ids, recordID)
			}
		}
	}
	ids = canonicalRecordIDs(ids)
	if len(ids) == 0 {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target %q has no history association", ErrInvalidTargetSemantics, mutation.TargetKind)
	}
	entryIDs := append(make([]uuid.UUID, 0, len(ids)), ids...)
	if contract.addressability == HistoryNotIndividuallyAddressable {
		entryIDs = []uuid.UUID{}
	}
	return HistoryTargetDescription{
		TargetKind:            mutation.TargetKind,
		TargetID:              mutation.TargetID,
		LogicalItemIdentity:   mutation.TargetKind + ":" + mutation.TargetID,
		HistoryRecordIDs:      ids,
		HistoryEntryRecordIDs: entryIDs,
	}, nil
}

type TargetSemanticsRequirement struct {
	TargetKind            string
	SourceOwnerID         string
	DispatchClass         RollbackDispatchClass
	AdmittedRecordTypes   []string
	HistoryRecordIDFields []string
	Addressability        HistoryAddressability
}

type TargetSemanticsDescriptor struct {
	TargetKind          string
	SourceOwnerID       string
	DispatchClass       RollbackDispatchClass
	AdmittedRecordTypes []string
	Addressability      HistoryAddressability
}

type compiledTargetSemantics struct {
	descriptor     TargetSemanticsDescriptor
	history        HistoryTargetSemantics
	rowProviders   map[string]rollbackcontract.RowSourceProvider
	nonRowProvider rollbackcontract.NonRowTargetProvider
}

type TargetSemanticsCatalog struct {
	byTargetKind                 map[string]compiledTargetSemantics
	defaultRowTargetByRecordType map[string]string
}

type targetSemanticsRegistry struct {
	SchemaID        string                         `json:"schema_id"`
	RegistryVersion int                            `json:"registry_version"`
	Targets         []targetSemanticsRegistryEntry `json:"targets"`
}

type targetSemanticsRegistryEntry struct {
	TargetKind            string                `json:"target_kind"`
	SourceOwner           string                `json:"source_owner"`
	DispatchClass         RollbackDispatchClass `json:"dispatch_class"`
	AdmittedRecordTypes   []string              `json:"admitted_record_types"`
	HistoryRecordIDFields []string              `json:"history_record_id_fields"`
	Addressability        HistoryAddressability `json:"addressability"`
}

// ParseTargetSemanticsRequirements decodes the adopted machine projection
// used by application composition. Source-owner contributions must still
// close the resulting requirements through NewTargetSemanticsCatalog.
func ParseTargetSemanticsRequirements(data []byte) ([]TargetSemanticsRequirement, error) {
	var registry targetSemanticsRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("%w: decode registry: %v", ErrInvalidTargetSemantics, err)
	}
	if registry.SchemaID != TargetSemanticsRegistrySchemaID || registry.RegistryVersion != 1 || len(registry.Targets) == 0 {
		return nil, fmt.Errorf("%w: registry identity", ErrInvalidTargetSemantics)
	}
	requirements := make([]TargetSemanticsRequirement, 0, len(registry.Targets))
	for _, entry := range registry.Targets {
		requirements = append(requirements, TargetSemanticsRequirement{
			TargetKind:            entry.TargetKind,
			SourceOwnerID:         entry.SourceOwner,
			DispatchClass:         entry.DispatchClass,
			AdmittedRecordTypes:   append([]string(nil), entry.AdmittedRecordTypes...),
			HistoryRecordIDFields: append([]string(nil), entry.HistoryRecordIDFields...),
			Addressability:        entry.Addressability,
		})
	}
	return requirements, nil
}

// CurrentTargetSemanticsRequirements loads the generated projection owned by
// Revisions. Application composition supplies source-owner contributions but
// does not interpret or locate owner contract artifacts.
func CurrentTargetSemanticsRequirements() ([]TargetSemanticsRequirement, error) {
	artifact, ok := contractrevisions.Index["contracts/revisions/target-semantics-registry.v1.json"]
	if !ok {
		return nil, ErrMissingTargetSemantics
	}
	return ParseTargetSemanticsRequirements([]byte(artifact.JSON))
}

type targetCandidate struct {
	sourceOwnerID  string
	dispatchClass  RollbackDispatchClass
	recordType     string
	history        HistoryTargetSemantics
	rowProvider    rollbackcontract.RowSourceProvider
	nonRowProvider rollbackcontract.NonRowTargetProvider
}

func NewTargetSemanticsCatalog(requirements []TargetSemanticsRequirement, contributions []ProviderContribution) (*TargetSemanticsCatalog, error) {
	required := make(map[string]TargetSemanticsRequirement, len(requirements))
	for _, requirement := range requirements {
		normalized, err := normalizeTargetRequirement(requirement)
		if err != nil {
			return nil, err
		}
		if _, duplicate := required[normalized.TargetKind]; duplicate {
			return nil, fmt.Errorf("%w: target kind %q", ErrDuplicateTargetSemantics, normalized.TargetKind)
		}
		required[normalized.TargetKind] = normalized
	}

	candidates := make(map[string][]targetCandidate)
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			candidates["record"] = append(candidates["record"], targetCandidate{
				sourceOwnerID: "record_source_owner",
				dispatchClass: RollbackDispatchRow,
				recordType:    record.RecordType,
				history:       NewDirectRecordHistoryTargetSemantics(HistorySingleEntry),
				rowProvider:   record.RowRollbackProvider,
			})
			for _, targetKind := range record.HistoryTargetKinds {
				candidates[targetKind] = append(candidates[targetKind], targetCandidate{
					sourceOwnerID: string(contribution.SourceOwnerModule),
					dispatchClass: RollbackDispatchRow,
					recordType:    record.RecordType,
					history:       NewDirectRecordHistoryTargetSemantics(HistorySingleEntry),
					rowProvider:   record.RowRollbackProvider,
				})
			}
		}
		for _, target := range contribution.NonRowTargets {
			candidates[target.TargetKind] = append(candidates[target.TargetKind], targetCandidate{
				sourceOwnerID:  string(contribution.SourceOwnerModule),
				dispatchClass:  RollbackDispatchNonRow,
				history:        target.HistorySemantics,
				nonRowProvider: target.RollbackProvider,
			})
		}
	}

	catalog := &TargetSemanticsCatalog{
		byTargetKind:                 make(map[string]compiledTargetSemantics, len(requirements)),
		defaultRowTargetByRecordType: make(map[string]string),
	}
	for targetKind, values := range candidates {
		requirement, expected := required[targetKind]
		if !expected {
			return nil, fmt.Errorf("%w: target kind %q", ErrUnexpectedTargetSemantics, targetKind)
		}
		compiled, err := compileTargetSemantics(requirement, values)
		if err != nil {
			return nil, err
		}
		catalog.byTargetKind[targetKind] = compiled
	}
	missing := make([]string, 0)
	for targetKind := range required {
		if _, ok := catalog.byTargetKind[targetKind]; !ok {
			missing = append(missing, targetKind)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: target kinds %v", ErrMissingTargetSemantics, missing)
	}
	for targetKind, entry := range catalog.byTargetKind {
		if entry.descriptor.DispatchClass != RollbackDispatchRow {
			continue
		}
		for _, recordType := range entry.descriptor.AdmittedRecordTypes {
			current, exists := catalog.defaultRowTargetByRecordType[recordType]
			if !exists || current == "record" {
				catalog.defaultRowTargetByRecordType[recordType] = targetKind
				continue
			}
			if targetKind != "record" && current != targetKind {
				return nil, fmt.Errorf("%w: record type %q has row targets %q and %q", ErrDuplicateTargetSemantics, recordType, current, targetKind)
			}
		}
	}
	return catalog, nil
}

func normalizeTargetRequirement(requirement TargetSemanticsRequirement) (TargetSemanticsRequirement, error) {
	requirement.TargetKind = strings.TrimSpace(requirement.TargetKind)
	requirement.SourceOwnerID = strings.TrimSpace(requirement.SourceOwnerID)
	requirement.AdmittedRecordTypes = sortedUniqueStrings(requirement.AdmittedRecordTypes)
	requirement.HistoryRecordIDFields = sortedUniqueStrings(requirement.HistoryRecordIDFields)
	if requirement.TargetKind == "" || requirement.SourceOwnerID == "" ||
		(requirement.DispatchClass != RollbackDispatchRow && requirement.DispatchClass != RollbackDispatchNonRow) ||
		(requirement.Addressability != HistorySingleEntry && requirement.Addressability != HistoryNotIndividuallyAddressable) {
		return TargetSemanticsRequirement{}, fmt.Errorf("%w: incomplete requirement for target %q", ErrInvalidTargetSemantics, requirement.TargetKind)
	}
	return requirement, nil
}

func compileTargetSemantics(requirement TargetSemanticsRequirement, values []targetCandidate) (compiledTargetSemantics, error) {
	if len(values) == 0 {
		return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, requirement.TargetKind)
	}
	compiled := compiledTargetSemantics{
		descriptor: TargetSemanticsDescriptor{
			TargetKind:     requirement.TargetKind,
			SourceOwnerID:  requirement.SourceOwnerID,
			DispatchClass:  requirement.DispatchClass,
			Addressability: requirement.Addressability,
		},
		rowProviders: map[string]rollbackcontract.RowSourceProvider{},
	}
	if requirement.DispatchClass == RollbackDispatchNonRow && len(values) > 1 {
		return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q", ErrDuplicateTargetSemantics, requirement.TargetKind)
	}
	for _, candidate := range values {
		if candidate.sourceOwnerID != requirement.SourceOwnerID || candidate.dispatchClass != requirement.DispatchClass || nilHistoryTargetSemantics(candidate.history) {
			return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q contribution", ErrInvalidTargetSemantics, requirement.TargetKind)
		}
		contract := candidate.history.historyContract()
		if contract.addressability != requirement.Addressability ||
			!reflect.DeepEqual(sortedUniqueStrings(contract.recordIDFields), requirement.HistoryRecordIDFields) ||
			(contract.directTargetID != (requirement.DispatchClass == RollbackDispatchRow)) {
			return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q history contract", ErrInvalidTargetSemantics, requirement.TargetKind)
		}
		if compiled.history == nil {
			compiled.history = candidate.history
		}
		switch requirement.DispatchClass {
		case RollbackDispatchRow:
			if candidate.recordType == "" || nilRowSourceProvider(candidate.rowProvider) {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q row provider", ErrInvalidTargetSemantics, requirement.TargetKind)
			}
			if _, duplicate := compiled.rowProviders[candidate.recordType]; duplicate {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q record type %q", ErrDuplicateTargetSemantics, requirement.TargetKind, candidate.recordType)
			}
			compiled.rowProviders[candidate.recordType] = candidate.rowProvider
		case RollbackDispatchNonRow:
			if nilNonRowTargetProvider(candidate.nonRowProvider) {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q non-row provider", ErrInvalidTargetSemantics, requirement.TargetKind)
			}
			compiled.nonRowProvider = candidate.nonRowProvider
		}
	}
	admitted := make([]string, 0, len(compiled.rowProviders))
	for recordType := range compiled.rowProviders {
		admitted = append(admitted, recordType)
	}
	sort.Strings(admitted)
	if !reflect.DeepEqual(admitted, requirement.AdmittedRecordTypes) {
		return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q admits %v, want %v", ErrInvalidTargetSemantics, requirement.TargetKind, admitted, requirement.AdmittedRecordTypes)
	}
	compiled.descriptor.AdmittedRecordTypes = admitted
	return compiled, nil
}

func (catalog *TargetSemanticsCatalog) DescribeMutation(mutation StoredMutation) (HistoryTargetDescription, error) {
	if catalog == nil {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, mutation.TargetKind)
	}
	entry, ok := catalog.byTargetKind[mutation.TargetKind]
	if !ok {
		return HistoryTargetDescription{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, mutation.TargetKind)
	}
	return entry.history.DescribeMutation(mutation)
}

func (catalog *TargetSemanticsCatalog) DescribeValues(targetKind string, targetID string, beforeValue any, afterValue any) (HistoryTargetDescription, error) {
	beforeObject, err := historyMutationObject(beforeValue)
	if err != nil {
		return HistoryTargetDescription{}, err
	}
	afterObject, err := historyMutationObject(afterValue)
	if err != nil {
		return HistoryTargetDescription{}, err
	}
	return catalog.DescribeMutation(StoredMutation{
		TargetKind:  targetKind,
		TargetID:    targetID,
		BeforeValue: beforeObject,
		AfterValue:  afterObject,
	})
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

func (catalog *TargetSemanticsCatalog) Descriptors() []TargetSemanticsDescriptor {
	if catalog == nil {
		return nil
	}
	result := make([]TargetSemanticsDescriptor, 0, len(catalog.byTargetKind))
	for _, entry := range catalog.byTargetKind {
		descriptor := entry.descriptor
		descriptor.AdmittedRecordTypes = append([]string(nil), descriptor.AdmittedRecordTypes...)
		result = append(result, descriptor)
	}
	sort.Slice(result, func(left, right int) bool { return result[left].TargetKind < result[right].TargetKind })
	return result
}

func (catalog *TargetSemanticsCatalog) TargetKinds() []string {
	descriptors := catalog.Descriptors()
	result := make([]string, 0, len(descriptors))
	for _, descriptor := range descriptors {
		result = append(result, descriptor.TargetKind)
	}
	return result
}

func (catalog *TargetSemanticsCatalog) dispatchClass(targetKind string) (RollbackDispatchClass, error) {
	entry, ok := catalog.targetEntry(targetKind)
	if !ok {
		return "", fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, targetKind)
	}
	return entry.descriptor.DispatchClass, nil
}

func (catalog *TargetSemanticsCatalog) rowProvider(targetKind string, recordType string) (rollbackcontract.RowSourceProvider, error) {
	entry, ok := catalog.targetEntry(targetKind)
	if !ok || entry.descriptor.DispatchClass != RollbackDispatchRow {
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
	if !ok || entry.descriptor.DispatchClass != RollbackDispatchNonRow || nilNonRowTargetProvider(entry.nonRowProvider) {
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

func nilRowSourceProvider(provider rollbackcontract.RowSourceProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func nilNonRowTargetProvider(provider rollbackcontract.NonRowTargetProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

func sortedUniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, duplicate := seen[value]; duplicate {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
