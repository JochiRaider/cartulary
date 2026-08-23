package incidentbundle

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/vocabulary"
)

var (
	indicatorPortableMembers = []string{
		"record_id", "incident_id", "indicator_type", "value_kind",
		"display_value", "normalized_value", "dedupe_key", "defanged_value",
		"hash_algorithm", "hash_value", "stix_pattern", "row_version",
		"created_at", "updated_at", "created_by_user_id", "updated_by_user_id",
		"deleted_at", "deleted_by_user_id",
	}
	observationPortableMembers = []string{
		"indicator_observation_id", "incident_id", "source_record_id",
		"source_field_key", "origin_kind", "origin_locator", "observed_text",
		"parsed_indicator_type", "normalized_candidate", "resolution_status",
		"resolved_indicator_record_id", "row_version", "created_by_user_id",
		"created_at", "resolved_by_user_id", "resolved_at", "resolution_method",
		"deleted_at", "deleted_by_user_id",
	}
	intervalPortableMembers = []string{
		"indicator_state_interval_id", "incident_id", "indicator_record_id",
		"lifecycle_state", "valid_from", "valid_to", "confidence", "rationale",
		"support_refs", "assessor", "assessed_at", "row_version",
		"created_by_user_id", "created_at", "deleted_at", "deleted_by_user_id",
	}
)

func prepareIndicatorImport(
	bundle sourceport.Bundle,
	importContext sourceport.ImportContext,
) (preparedIndicatorImport, error) {
	prepared := preparedIndicatorImport{binding: portableBinding{
		operationID: importContext.OperationID, incidentID: importContext.IncidentID,
		bundleVersion: importContext.BundleVersion, contractMajor: sourceport.ContractMajor,
	}}
	if bundle == nil || importContext.OperationID == "" || importContext.IncidentID == uuid.Nil ||
		importContext.ActorUserID == uuid.Nil ||
		importContext.BundleVersion != 2 {
		return preparedIndicatorImport{}, indicatorSourceFailure(representationInvariant)
	}

	var failures []indicatorFailureCandidate
	indicatorRows := decodeIndicatorRows(bundle, indicatorsBundlePath, &failures)
	observationRows := decodeIndicatorRows(bundle, observationsBundlePath, &failures)
	intervalRows := decodeIndicatorRows(bundle, intervalsBundlePath, &failures)
	if len(failures) > 0 {
		return preparedIndicatorImport{}, selectedIndicatorFailure(failures)
	}

	indicatorSeen := make(map[uuid.UUID]struct{}, len(indicatorRows))
	activeIdentitySeen := make(map[string]uuid.UUID, len(indicatorRows))
	indicatorInputOrder := make([]string, 0, len(indicatorRows))
	for _, raw := range indicatorRows {
		row, rowFailures := preparePortableIndicatorRow(raw, importContext)
		failures = append(failures, rowFailures...)
		if len(rowFailures) > 0 && row.RecordID == uuid.Nil {
			continue
		}
		identityText := row.RecordID.String()
		indicatorInputOrder = append(indicatorInputOrder, identityText)
		if _, duplicate := indicatorSeen[row.RecordID]; duplicate {
			failures = append(failures, indicatorFailure(representationInvariant, indicatorsBundlePath, identityText, raw))
		}
		indicatorSeen[row.RecordID] = struct{}{}
		if row.DeletedAt == nil && row.IndicatorType != "" && row.DedupeKey != "" {
			key := row.IndicatorType + "\x00" + row.DedupeKey
			if prior, duplicate := activeIdentitySeen[key]; duplicate {
				failures = append(failures,
					indicatorFailure(identityUniqueInvariant, indicatorsBundlePath, prior.String(), raw),
					indicatorFailure(identityUniqueInvariant, indicatorsBundlePath, identityText, raw),
				)
			} else {
				activeIdentitySeen[key] = row.RecordID
			}
		}
		prepared.indicators = append(prepared.indicators, row)
	}
	if !sort.StringsAreSorted(indicatorInputOrder) && len(indicatorInputOrder) > 0 {
		failures = append(failures, indicatorFailure(representationInvariant, indicatorsBundlePath, indicatorInputOrder[0], indicatorInputOrder))
	}

	observationSeen := make(map[uuid.UUID]struct{}, len(observationRows))
	observationInputOrder := make([]string, 0, len(observationRows))
	for _, raw := range observationRows {
		row, rowFailures := preparePortableObservationRow(raw, importContext)
		failures = append(failures, rowFailures...)
		if len(rowFailures) > 0 && row.ObservationID == uuid.Nil {
			continue
		}
		identityText := row.ObservationID.String()
		observationInputOrder = append(observationInputOrder, identityText)
		if _, duplicate := observationSeen[row.ObservationID]; duplicate {
			failures = append(failures, indicatorFailure(representationInvariant, observationsBundlePath, identityText, raw))
		}
		observationSeen[row.ObservationID] = struct{}{}
		prepared.observations = append(prepared.observations, row)
	}
	if !sort.StringsAreSorted(observationInputOrder) && len(observationInputOrder) > 0 {
		failures = append(failures, indicatorFailure(observationOrderedInvariant, observationsBundlePath, observationInputOrder[0], observationInputOrder))
	}

	intervalSeen := make(map[uuid.UUID]struct{}, len(intervalRows))
	intervalInputOrder := make([]string, 0, len(intervalRows))
	for _, raw := range intervalRows {
		row, rowFailures := preparePortableIntervalRow(raw, importContext)
		failures = append(failures, rowFailures...)
		if len(rowFailures) > 0 && row.IntervalID == uuid.Nil {
			continue
		}
		identityText := row.IntervalID.String()
		intervalInputOrder = append(intervalInputOrder, identityText)
		if _, duplicate := intervalSeen[row.IntervalID]; duplicate {
			failures = append(failures, indicatorFailure(representationInvariant, intervalsBundlePath, identityText, raw))
		}
		intervalSeen[row.IntervalID] = struct{}{}
		prepared.intervals = append(prepared.intervals, row)
	}
	if !sort.StringsAreSorted(intervalInputOrder) && len(intervalInputOrder) > 0 {
		failures = append(failures, indicatorFailure(intervalOrderedInvariant, intervalsBundlePath, intervalInputOrder[0], intervalInputOrder))
	}

	if len(failures) > 0 {
		return preparedIndicatorImport{}, selectedIndicatorFailure(failures)
	}
	sort.Slice(prepared.indicators, func(i, j int) bool {
		return prepared.indicators[i].RecordID.String() < prepared.indicators[j].RecordID.String()
	})
	sort.Slice(prepared.observations, func(i, j int) bool {
		return prepared.observations[i].ObservationID.String() < prepared.observations[j].ObservationID.String()
	})
	sort.Slice(prepared.intervals, func(i, j int) bool {
		return prepared.intervals[i].IntervalID.String() < prepared.intervals[j].IntervalID.String()
	})
	return prepared, nil
}

func decodeIndicatorRows(
	bundle sourceport.Bundle,
	path string,
	failures *[]indicatorFailureCandidate,
) []map[string]any {
	payload, ok := bundle.File(path)
	if !ok {
		*failures = append(*failures, indicatorFailure(representationInvariant, path, "", path))
		return nil
	}
	rows, err := incidentportability.DecodeStrictNDJSONObjects(payload, path)
	if err != nil {
		*failures = append(*failures, indicatorFailure(representationInvariant, path, "", payload))
		return nil
	}
	return rows
}

func preparePortableIndicatorRow(
	raw map[string]any,
	importContext sourceport.ImportContext,
) (portableIndicatorRow, []indicatorFailureCandidate) {
	identityText := portableStableIdentity(raw["record_id"])
	representationFailure := func() (portableIndicatorRow, []indicatorFailureCandidate) {
		return portableIndicatorRow{}, []indicatorFailureCandidate{
			indicatorFailure(representationInvariant, indicatorsBundlePath, identityText, raw),
		}
	}
	if !exactPortableMembers(raw, indicatorPortableMembers) {
		return representationFailure()
	}
	recordID, ok := canonicalPortableUUID(raw["record_id"])
	if !ok {
		return representationFailure()
	}
	incidentID, ok := canonicalPortableUUID(raw["incident_id"])
	if !ok || incidentID != importContext.IncidentID {
		return representationFailure()
	}
	indicatorType, ok := portableText(raw["indicator_type"], false)
	if !ok {
		return representationFailure()
	}
	if !vocabulary.IsIndicatorType(indicatorType) {
		return representationFailure()
	}
	valueKind, ok := portableText(raw["value_kind"], false)
	if !ok {
		return representationFailure()
	}
	if !vocabulary.IsValueKind(valueKind) {
		return representationFailure()
	}
	displayValue, ok := portableText(raw["display_value"], false)
	if !ok {
		return representationFailure()
	}
	normalizedValue, ok := nullablePortableText(raw["normalized_value"], false)
	if !ok {
		return representationFailure()
	}
	dedupeKey, ok := portableText(raw["dedupe_key"], false)
	if !ok || !portableDedupePattern.MatchString(dedupeKey) {
		return representationFailure()
	}
	defangedValue, ok := nullablePortableText(raw["defanged_value"], true)
	if !ok {
		return representationFailure()
	}
	hashAlgorithm, ok := nullablePortableText(raw["hash_algorithm"], false)
	if !ok {
		return representationFailure()
	}
	hashValue, ok := nullablePortableText(raw["hash_value"], false)
	if !ok {
		return representationFailure()
	}
	stixPattern, ok := nullablePortableText(raw["stix_pattern"], true)
	if !ok {
		return representationFailure()
	}
	rowVersion, ok := canonicalPortableInteger(raw["row_version"])
	if !ok || rowVersion < 1 {
		return representationFailure()
	}
	createdAt, ok := canonicalPortableTimestamp(raw["created_at"])
	if !ok {
		return representationFailure()
	}
	updatedAt, ok := canonicalPortableTimestamp(raw["updated_at"])
	if !ok || updatedAt.Before(createdAt) {
		return representationFailure()
	}
	createdBy, ok := admittedPortableActor(raw["created_by_user_id"], importContext)
	if !ok {
		return representationFailure()
	}
	updatedBy, ok := admittedPortableActor(raw["updated_by_user_id"], importContext)
	if !ok {
		return representationFailure()
	}
	deletedAt, ok := nullablePortableTimestamp(raw["deleted_at"])
	if !ok {
		return representationFailure()
	}
	deletedBy, ok := nullableAdmittedPortableActor(raw["deleted_by_user_id"], importContext)
	if !ok || (deletedAt == nil) != (deletedBy == nil) ||
		(deletedAt != nil && (deletedAt.Before(createdAt) || deletedAt.After(updatedAt))) {
		return representationFailure()
	}

	row := portableIndicatorRow{
		RecordID: recordID, IncidentID: incidentID, IndicatorType: indicatorType,
		ValueKind: valueKind, DisplayValue: displayValue, NormalizedValue: normalizedValue,
		DedupeKey: dedupeKey, DefangedValue: defangedValue, HashAlgorithm: hashAlgorithm,
		HashValue: hashValue, STIXPattern: stixPattern, RowVersion: rowVersion,
		CreatedAt: createdAt, UpdatedAt: updatedAt,
		PortableCreatedByID: createdBy, PortableUpdatedByID: updatedBy,
		RuntimeCreatedByID: importContext.ActorUserID, RuntimeUpdatedByID: importContext.ActorUserID,
		DeletedAt: deletedAt, PortableDeletedByID: deletedBy,
		RuntimeDeletedByID: runtimeActorPointer(deletedBy, importContext.ActorUserID),
	}
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: indicatorType, ValueKind: valueKind, DisplayValue: displayValue,
		NormalizedValue: normalizedValue, DefangedValue: defangedValue,
		HashAlgorithm: hashAlgorithm, HashValue: hashValue, STIXPattern: stixPattern,
	})
	if err != nil {
		return row, []indicatorFailureCandidate{
			indicatorFailure(representationInvariant, indicatorsBundlePath, identityText, raw),
		}
	}
	if canonical.IndicatorType != indicatorType || canonical.ValueKind != valueKind ||
		canonical.DisplayValue != displayValue ||
		!portableStringPointersEqual(canonical.NormalizedValue, normalizedValue) ||
		!portableStringPointersEqual(canonical.HashAlgorithm, hashAlgorithm) ||
		!portableStringPointersEqual(canonical.HashValue, hashValue) ||
		canonical.DedupeKey != dedupeKey {
		return row, []indicatorFailureCandidate{
			indicatorFailure(normalizationInvariant, indicatorsBundlePath, identityText, raw),
		}
	}
	return row, nil
}

func preparePortableObservationRow(
	raw map[string]any,
	importContext sourceport.ImportContext,
) (portableObservationRow, []indicatorFailureCandidate) {
	identityText := portableStableIdentity(raw["indicator_observation_id"])
	representationFailure := func() (portableObservationRow, []indicatorFailureCandidate) {
		return portableObservationRow{}, []indicatorFailureCandidate{
			indicatorFailure(representationInvariant, observationsBundlePath, identityText, raw),
		}
	}
	if !exactPortableMembers(raw, observationPortableMembers) {
		return representationFailure()
	}
	observationID, ok := canonicalPortableUUID(raw["indicator_observation_id"])
	if !ok {
		return representationFailure()
	}
	incidentID, ok := canonicalPortableUUID(raw["incident_id"])
	if !ok || incidentID != importContext.IncidentID {
		return representationFailure()
	}
	sourceRecordID, ok := canonicalPortableUUID(raw["source_record_id"])
	if !ok {
		return representationFailure()
	}
	sourceFieldKey, ok := portableText(raw["source_field_key"], true)
	if !ok {
		return representationFailure()
	}
	originText, ok := portableText(raw["origin_kind"], true)
	if !ok {
		return representationFailure()
	}
	originKind, originErr := indicatororigin.Parse(originText)
	originLocator, ok := portableText(raw["origin_locator"], true)
	if !ok {
		return representationFailure()
	}
	observedText, ok := portableText(raw["observed_text"], true)
	if !ok {
		return representationFailure()
	}
	parsedIndicatorType, ok := nullablePortableText(raw["parsed_indicator_type"], false)
	if !ok {
		return representationFailure()
	}
	if parsedIndicatorType != nil {
		if !vocabulary.IsIndicatorType(*parsedIndicatorType) {
			return representationFailure()
		}
	}
	normalizedCandidate, ok := nullablePortableText(raw["normalized_candidate"], false)
	if !ok {
		return representationFailure()
	}
	resolutionStatus, ok := portableText(raw["resolution_status"], true)
	if !ok || !vocabulary.IsObservationStatus(resolutionStatus) {
		return representationFailure()
	}
	resolvedIndicatorID, ok := nullablePortableUUID(raw["resolved_indicator_record_id"])
	if !ok {
		return representationFailure()
	}
	rowVersion, ok := canonicalPortableInteger(raw["row_version"])
	if !ok {
		return representationFailure()
	}
	createdBy, ok := admittedPortableActor(raw["created_by_user_id"], importContext)
	if !ok {
		return representationFailure()
	}
	createdAt, ok := canonicalPortableTimestamp(raw["created_at"])
	if !ok {
		return representationFailure()
	}
	resolvedBy, ok := nullableAdmittedPortableActor(raw["resolved_by_user_id"], importContext)
	if !ok {
		return representationFailure()
	}
	resolvedAt, ok := nullablePortableTimestamp(raw["resolved_at"])
	if !ok {
		return representationFailure()
	}
	resolutionMethod, ok := nullablePortableText(raw["resolution_method"], false)
	if !ok {
		return representationFailure()
	}
	deletedAt, ok := nullablePortableTimestamp(raw["deleted_at"])
	if !ok {
		return representationFailure()
	}
	deletedBy, ok := nullableAdmittedPortableActor(raw["deleted_by_user_id"], importContext)
	if !ok || (deletedAt == nil) != (deletedBy == nil) {
		return representationFailure()
	}

	row := portableObservationRow{
		ObservationID: observationID, IncidentID: incidentID, SourceRecordID: sourceRecordID,
		SourceFieldKey: sourceFieldKey, OriginKind: originKind, OriginLocator: originLocator,
		ObservedText: observedText, ParsedIndicatorType: parsedIndicatorType,
		NormalizedCandidate: normalizedCandidate, ResolutionStatus: resolutionStatus,
		ResolvedIndicatorID: resolvedIndicatorID, RowVersion: rowVersion,
		PortableCreatedByID: createdBy, RuntimeCreatedByID: importContext.ActorUserID,
		CreatedAt: createdAt, PortableResolvedByID: resolvedBy,
		RuntimeResolvedByID: runtimeActorPointer(resolvedBy, importContext.ActorUserID),
		ResolvedAt:          resolvedAt, ResolutionMethod: resolutionMethod, DeletedAt: deletedAt,
		PortableDeletedByID: deletedBy,
		RuntimeDeletedByID:  runtimeActorPointer(deletedBy, importContext.ActorUserID),
	}
	var failures []indicatorFailureCandidate
	if parsedIndicatorType == nil && normalizedCandidate != nil {
		failures = append(failures, indicatorFailure(observationCoherentInvariant, observationsBundlePath, identityText, raw))
	} else if parsedIndicatorType != nil {
		canonicalType, canonicalCandidate, err := identity.NormalizeObservationCandidate(parsedIndicatorType, normalizedCandidate, observedText)
		if err != nil || !portableStringPointersEqual(canonicalType, parsedIndicatorType) ||
			!portableStringPointersEqual(canonicalCandidate, normalizedCandidate) {
			failures = append(failures, indicatorFailure(normalizationInvariant, observationsBundlePath, identityText, raw))
		}
	}
	if rowVersion < 1 || (resolvedAt != nil && resolvedAt.Before(createdAt)) ||
		(deletedAt != nil && (deletedAt.Before(createdAt) || (resolvedAt != nil && deletedAt.Before(*resolvedAt)))) {
		failures = append(failures, indicatorFailure(observationOrderedInvariant, observationsBundlePath, identityText, raw))
	}
	coherent := strings.TrimSpace(sourceFieldKey) != "" && strings.TrimSpace(originLocator) != "" &&
		strings.TrimSpace(observedText) != "" && originErr == nil
	switch resolutionStatus {
	case "unresolved":
		coherent = coherent && resolvedIndicatorID == nil && resolvedBy == nil && resolvedAt == nil && resolutionMethod == nil
	case "resolved":
		coherent = coherent && resolvedIndicatorID != nil && resolvedBy != nil && resolvedAt != nil && resolutionMethod != nil
	case "dismissed":
		coherent = coherent && resolvedIndicatorID == nil && resolvedBy != nil && resolvedAt != nil && resolutionMethod != nil
	default:
		coherent = false
	}
	if !coherent {
		failures = append(failures, indicatorFailure(observationCoherentInvariant, observationsBundlePath, identityText, raw))
	}
	return row, failures
}

func preparePortableIntervalRow(
	raw map[string]any,
	importContext sourceport.ImportContext,
) (portableIntervalRow, []indicatorFailureCandidate) {
	identityText := portableStableIdentity(raw["indicator_state_interval_id"])
	representationFailure := func() (portableIntervalRow, []indicatorFailureCandidate) {
		return portableIntervalRow{}, []indicatorFailureCandidate{
			indicatorFailure(representationInvariant, intervalsBundlePath, identityText, raw),
		}
	}
	if !exactPortableMembers(raw, intervalPortableMembers) {
		return representationFailure()
	}
	intervalID, ok := canonicalPortableUUID(raw["indicator_state_interval_id"])
	if !ok {
		return representationFailure()
	}
	incidentID, ok := canonicalPortableUUID(raw["incident_id"])
	if !ok || incidentID != importContext.IncidentID {
		return representationFailure()
	}
	indicatorRecordID, ok := canonicalPortableUUID(raw["indicator_record_id"])
	if !ok {
		return representationFailure()
	}
	lifecycleState, ok := portableText(raw["lifecycle_state"], true)
	if !ok || !vocabulary.IsLifecycleState(lifecycleState) {
		return representationFailure()
	}
	validFrom, ok := canonicalPortableTimestamp(raw["valid_from"])
	if !ok {
		return representationFailure()
	}
	validTo, ok := nullablePortableTimestamp(raw["valid_to"])
	if !ok {
		return representationFailure()
	}
	var confidence *int
	if raw["confidence"] != nil {
		value, ok := canonicalPortableInteger(raw["confidence"])
		if !ok {
			return representationFailure()
		}
		converted := int(value)
		if int64(converted) != value {
			return representationFailure()
		}
		confidence = &converted
	}
	rationale, ok := nullablePortableText(raw["rationale"], true)
	if !ok {
		return representationFailure()
	}
	supportValues, ok := raw["support_refs"].([]any)
	if !ok {
		return representationFailure()
	}
	supportRefs := make([]uuid.UUID, 0, len(supportValues))
	seenSupportRefs := make(map[uuid.UUID]struct{}, len(supportValues))
	for _, value := range supportValues {
		parsed, ok := canonicalPortableUUID(value)
		if !ok {
			return representationFailure()
		}
		if _, duplicate := seenSupportRefs[parsed]; duplicate {
			return representationFailure()
		}
		seenSupportRefs[parsed] = struct{}{}
		supportRefs = append(supportRefs, parsed)
	}
	assessor, ok := nullablePortableText(raw["assessor"], true)
	if !ok {
		return representationFailure()
	}
	assessedAt, ok := canonicalPortableTimestamp(raw["assessed_at"])
	if !ok {
		return representationFailure()
	}
	rowVersion, ok := canonicalPortableInteger(raw["row_version"])
	if !ok {
		return representationFailure()
	}
	createdBy, ok := admittedPortableActor(raw["created_by_user_id"], importContext)
	if !ok {
		return representationFailure()
	}
	createdAt, ok := canonicalPortableTimestamp(raw["created_at"])
	if !ok {
		return representationFailure()
	}
	deletedAt, ok := nullablePortableTimestamp(raw["deleted_at"])
	if !ok {
		return representationFailure()
	}
	deletedBy, ok := nullableAdmittedPortableActor(raw["deleted_by_user_id"], importContext)
	if !ok || (deletedAt == nil) != (deletedBy == nil) {
		return representationFailure()
	}

	row := portableIntervalRow{
		IntervalID: intervalID, IncidentID: incidentID, IndicatorRecordID: indicatorRecordID,
		LifecycleState: lifecycleState, ValidFrom: validFrom, ValidTo: validTo,
		Confidence: confidence, Rationale: rationale, SupportRefs: supportRefs,
		Assessor: assessor, AssessedAt: assessedAt, RowVersion: rowVersion,
		PortableCreatedByID: createdBy, RuntimeCreatedByID: importContext.ActorUserID,
		CreatedAt: createdAt, DeletedAt: deletedAt, PortableDeletedByID: deletedBy,
		RuntimeDeletedByID: runtimeActorPointer(deletedBy, importContext.ActorUserID),
	}
	var failures []indicatorFailureCandidate
	if rowVersion < 1 || (validTo != nil && validTo.Before(validFrom)) ||
		(deletedAt != nil && deletedAt.Before(createdAt)) {
		failures = append(failures, indicatorFailure(intervalOrderedInvariant, intervalsBundlePath, identityText, raw))
	}
	if confidence != nil && (*confidence < 0 || *confidence > 100) {
		failures = append(failures, indicatorFailure(intervalCoherentInvariant, intervalsBundlePath, identityText, raw))
	}
	return row, failures
}
