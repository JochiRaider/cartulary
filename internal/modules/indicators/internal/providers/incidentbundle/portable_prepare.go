package incidentbundle

import (
	"context"
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

func prepareIndicatorImport(ctx context.Context, evaluate identity.Evaluator, bundle incidentportability.MemberReader, input sourceport.ImportContext) (preparedIndicatorImport, error) {
	prepared := preparedIndicatorImport{binding: portableBinding{operationID: input.OperationID, incidentID: input.IncidentID, bundleVersion: input.BundleVersion, contractMajor: sourceport.ContractMajor}, evaluate: evaluate}
	if bundle == nil || input.OperationID == "" || input.IncidentID == uuid.Nil || input.ActorUserID == uuid.Nil || input.BundleVersion != 7 {
		return preparedIndicatorImport{}, indicatorSourceFailure(representationInvariant)
	}
	members, err := incidentportability.BindMembers(bundle, []string{indicatorsBundlePath, observationsBundlePath, intervalsBundlePath})
	if err != nil {
		return preparedIndicatorImport{}, indicatorSourceFailure(representationInvariant)
	}
	prepared.members = members
	var failures indicatorFailureAccumulator
	for _, path := range []string{indicatorsBundlePath, observationsBundlePath, intervalsBundlePath} {
		previous := ""
		err := incidentportability.WalkPath(ctx, members, path, func(raw map[string]any) error {
			var rowFailures []indicatorFailureCandidate
			var id uuid.UUID
			orderInvariant := representationInvariant
			switch path {
			case indicatorsBundlePath:
				row, f := preparePortableIndicatorRow(evaluate, raw, input)
				rowFailures = f
				id = row.RecordID
				prepared.indicatorCount++
			case observationsBundlePath:
				row, f := preparePortableObservationRow(evaluate, raw, input)
				rowFailures = f
				id = row.ObservationID
				prepared.observationCount++
				orderInvariant = observationOrderedInvariant
			case intervalsBundlePath:
				row, f := preparePortableIntervalRow(raw, input)
				rowFailures = f
				id = row.IntervalID
				orderInvariant = intervalOrderedInvariant
			}
			failures.candidates(rowFailures)
			if id != uuid.Nil {
				current := id.String()
				if current == previous {
					failures.add(representationInvariant)
				} else if current < previous {
					failures.add(orderInvariant)
				}
				previous = current
			}
			return nil
		})
		if err != nil {
			return preparedIndicatorImport{}, indicatorSourceFailure(representationInvariant)
		}
	}
	if err := failures.err(); err != nil {
		return preparedIndicatorImport{}, err
	}
	return prepared, nil
}

func walkIndicatorSource[T any](ctx context.Context, prepared preparedIndicatorImport, input sourceport.ImportContext, path string, parse func(map[string]any) (T, []indicatorFailureCandidate), visit func(T) error) error {
	if !prepared.binding.matches(input) || prepared.members == nil {
		return indicatorSourceFailure(representationInvariant)
	}
	return incidentportability.WalkPath(ctx, prepared.members, path, func(raw map[string]any) error {
		row, failures := parse(raw)
		if err := selectedIndicatorFailure(failures); err != nil {
			return err
		}
		return visit(row)
	})
}
func (p preparedIndicatorImport) indicators(ctx context.Context, input sourceport.ImportContext, visit func(portableIndicatorRow) error) error {
	return walkIndicatorSource(ctx, p, input, indicatorsBundlePath, func(raw map[string]any) (portableIndicatorRow, []indicatorFailureCandidate) {
		return preparePortableIndicatorRow(p.evaluate, raw, input)
	}, visit)
}
func (p preparedIndicatorImport) observations(ctx context.Context, input sourceport.ImportContext, visit func(portableObservationRow) error) error {
	return walkIndicatorSource(ctx, p, input, observationsBundlePath, func(raw map[string]any) (portableObservationRow, []indicatorFailureCandidate) {
		return preparePortableObservationRow(p.evaluate, raw, input)
	}, visit)
}
func (p preparedIndicatorImport) intervals(ctx context.Context, input sourceport.ImportContext, visit func(portableIntervalRow) error) error {
	return walkIndicatorSource(ctx, p, input, intervalsBundlePath, func(raw map[string]any) (portableIntervalRow, []indicatorFailureCandidate) {
		return preparePortableIntervalRow(raw, input)
	}, visit)
}

func preparePortableIndicatorRow(
	evaluate identity.Evaluator,
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
	if !ok || len(dedupeKey) > 32784 {
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
	canonical, err := identity.Canonicalize(evaluate, identity.Input{
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
	evaluate identity.Evaluator,
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
		canonicalType, canonicalCandidate, err := identity.NormalizeObservationCandidate(evaluate, parsedIndicatorType, normalizedCandidate, observedText)
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
