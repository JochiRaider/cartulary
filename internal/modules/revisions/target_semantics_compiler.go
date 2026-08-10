package revisions

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type targetAdmission struct {
	sourceOwnerID  string
	dispatchClass  rollbackcontract.DispatchClass
	recordType     string
	history        HistoryFacet
	rowProvider    rollbackcontract.RowSourceProvider
	nonRowProvider rollbackcontract.NonRowTargetProvider
}

func compileTargetSemanticsCatalog(requirements []targetSemanticsRequirement, contributions []ProviderContribution) (*TargetSemanticsCatalog, error) {
	required := make(map[string]targetSemanticsRequirement, len(requirements))
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

	admissions := make(map[string][]targetAdmission)
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			admissions["record"] = append(admissions["record"], targetAdmission{
				sourceOwnerID: "record_source_owner",
				dispatchClass: rollbackcontract.DispatchRow,
				recordType:    record.RecordType,
				history:       NewDirectRecordHistoryFacet(HistorySingleEntry),
				rowProvider:   record.RowRollbackProvider,
			})
			for _, targetKind := range record.HistoryTargetKinds {
				admissions[targetKind] = append(admissions[targetKind], targetAdmission{
					sourceOwnerID: string(contribution.SourceOwnerModule),
					dispatchClass: rollbackcontract.DispatchRow,
					recordType:    record.RecordType,
					history:       NewDirectRecordHistoryFacet(HistorySingleEntry),
					rowProvider:   record.RowRollbackProvider,
				})
			}
		}
		for _, target := range contribution.NonRowTargets {
			admissions[target.TargetKind] = append(admissions[target.TargetKind], targetAdmission{
				sourceOwnerID:  string(contribution.SourceOwnerModule),
				dispatchClass:  rollbackcontract.DispatchNonRow,
				history:        target.HistoryFacet,
				nonRowProvider: target.RollbackProvider,
			})
		}
	}

	catalog := &TargetSemanticsCatalog{
		byTargetKind:                 make(map[string]compiledTargetSemantics, len(requirements)),
		defaultRowTargetByRecordType: make(map[string]string),
	}
	for targetKind, values := range admissions {
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
		if entry.dispatchClass != rollbackcontract.DispatchRow {
			continue
		}
		for _, recordType := range entry.admittedRecordTypes {
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

func normalizeTargetRequirement(requirement targetSemanticsRequirement) (targetSemanticsRequirement, error) {
	requirement.TargetKind = strings.TrimSpace(requirement.TargetKind)
	requirement.SourceOwnerID = strings.TrimSpace(requirement.SourceOwnerID)
	requirement.AdmittedRecordTypes = sortedUniqueStrings(requirement.AdmittedRecordTypes)
	requirement.HistoryRecordIDFields = sortedUniqueStrings(requirement.HistoryRecordIDFields)
	if requirement.TargetKind == "" || requirement.SourceOwnerID == "" ||
		(requirement.DispatchClass != rollbackcontract.DispatchRow && requirement.DispatchClass != rollbackcontract.DispatchNonRow) ||
		(requirement.Addressability != HistorySingleEntry && requirement.Addressability != HistoryNotIndividuallyAddressable) {
		return targetSemanticsRequirement{}, fmt.Errorf("%w: incomplete requirement for target %q", ErrInvalidTargetSemantics, requirement.TargetKind)
	}
	return requirement, nil
}

func compileTargetSemantics(requirement targetSemanticsRequirement, values []targetAdmission) (compiledTargetSemantics, error) {
	if len(values) == 0 {
		return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q", ErrMissingTargetSemantics, requirement.TargetKind)
	}
	compiled := compiledTargetSemantics{
		dispatchClass: requirement.DispatchClass,
		rowProviders:  map[string]rollbackcontract.RowSourceProvider{},
	}
	if requirement.DispatchClass == rollbackcontract.DispatchNonRow && len(values) > 1 {
		return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q", ErrDuplicateTargetSemantics, requirement.TargetKind)
	}
	for _, admission := range values {
		if admission.sourceOwnerID != requirement.SourceOwnerID || admission.dispatchClass != requirement.DispatchClass || admission.history.isZero() {
			return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q contribution", ErrInvalidTargetSemantics, requirement.TargetKind)
		}
		contract := admission.history.historyContract()
		if contract.addressability != requirement.Addressability ||
			!reflect.DeepEqual(sortedUniqueStrings(contract.recordIDFields), requirement.HistoryRecordIDFields) ||
			(contract.directTargetID != (requirement.DispatchClass == rollbackcontract.DispatchRow)) {
			return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q history contract", ErrInvalidTargetSemantics, requirement.TargetKind)
		}
		if compiled.history.isZero() {
			compiled.history = admission.history
		}
		switch requirement.DispatchClass {
		case rollbackcontract.DispatchRow:
			if admission.recordType == "" || nilRowSourceProvider(admission.rowProvider) {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q row provider", ErrInvalidTargetSemantics, requirement.TargetKind)
			}
			if _, duplicate := compiled.rowProviders[admission.recordType]; duplicate {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q record type %q", ErrDuplicateTargetSemantics, requirement.TargetKind, admission.recordType)
			}
			compiled.rowProviders[admission.recordType] = admission.rowProvider
		case rollbackcontract.DispatchNonRow:
			if nilNonRowTargetProvider(admission.nonRowProvider) {
				return compiledTargetSemantics{}, fmt.Errorf("%w: target kind %q non-row provider", ErrInvalidTargetSemantics, requirement.TargetKind)
			}
			compiled.nonRowProvider = admission.nonRowProvider
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
	compiled.admittedRecordTypes = admitted
	return compiled, nil
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
