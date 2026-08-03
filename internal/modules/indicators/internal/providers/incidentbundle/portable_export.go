package incidentbundle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
)

func exportFiles(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]incidentportability.File, error) {
	indicators, err := loadPortableIndicators(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	observations, err := loadPortableObservations(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	intervals, err := loadPortableIntervals(ctx, exportContext)
	if err != nil {
		return nil, err
	}
	envelopes, err := loadPortableRecordEnvelopes(ctx, exportContext.Query, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("indicator portability envelope export query failed")
	}
	for _, row := range indicators {
		envelope, present := envelopes[row.RecordID]
		if !present || !indicatorEnvelopeEqual(row, envelope) {
			return nil, errors.New("indicator portability envelope is invalid")
		}
	}
	indicatorPayload, err := encodePortableIndicators(ctx, exportContext, indicators)
	if err != nil {
		return nil, err
	}
	observationPayload, err := encodePortableObservations(ctx, exportContext, observations)
	if err != nil {
		return nil, err
	}
	intervalPayload, err := encodePortableIntervals(ctx, exportContext, intervals)
	if err != nil {
		return nil, err
	}
	return []incidentportability.File{
		{Path: indicatorsBundlePath, Payload: indicatorPayload},
		{Path: observationsBundlePath, Payload: observationPayload},
		{Path: intervalsBundlePath, Payload: intervalPayload},
	}, nil
}

func loadPortableIndicators(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]portableIndicatorRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT record_id, incident_id, indicator_type, value_kind, display_value,
       normalized_value, dedupe_key, defanged_value, hash_algorithm, hash_value,
       stix_pattern, row_version, created_at, updated_at, created_by_user_id,
       updated_by_user_id, deleted_at, deleted_by_user_id
  FROM indicators
 WHERE incident_id = $1
 ORDER BY record_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("indicator portability export query failed")
	}
	defer rows.Close()
	result := make([]portableIndicatorRow, 0)
	for rows.Next() {
		var row portableIndicatorRow
		var normalized, defanged, hashAlgorithm, hashValue, stix pgtype.Text
		var deletedAt pgtype.Timestamptz
		var deletedBy pgtype.UUID
		if err := rows.Scan(
			&row.RecordID, &row.IncidentID, &row.IndicatorType, &row.ValueKind,
			&row.DisplayValue, &normalized, &row.DedupeKey, &defanged,
			&hashAlgorithm, &hashValue, &stix, &row.RowVersion, &row.CreatedAt,
			&row.UpdatedAt, &row.RuntimeCreatedByID, &row.RuntimeUpdatedByID,
			&deletedAt, &deletedBy,
		); err != nil {
			return nil, errors.New("indicator portability export scan failed")
		}
		row.NormalizedValue = textFromPG(normalized)
		row.DefangedValue = textFromPG(defanged)
		row.HashAlgorithm = textFromPG(hashAlgorithm)
		row.HashValue = textFromPG(hashValue)
		row.STIXPattern = textFromPG(stix)
		row.DeletedAt = timeFromPG(deletedAt)
		row.RuntimeDeletedByID = uuidFromPG(deletedBy)
		if err := validatePortableIndicatorForExport(row, exportContext.IncidentID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("indicator portability export iteration failed")
	}
	return result, nil
}

func loadPortableObservations(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]portableObservationRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT indicator_observation_id, incident_id, source_record_id, source_field_key,
       origin_kind, origin_locator, observed_text, parsed_indicator_type,
       normalized_candidate, resolution_status, resolved_indicator_record_id,
       row_version, created_by_user_id, created_at, resolved_by_user_id,
       resolved_at, resolution_method, deleted_at, deleted_by_user_id
  FROM indicator_observations
 WHERE incident_id = $1
 ORDER BY indicator_observation_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("indicator observation portability export query failed")
	}
	defer rows.Close()
	result := make([]portableObservationRow, 0)
	for rows.Next() {
		var row portableObservationRow
		var originText string
		var parsedType, normalizedCandidate, resolutionMethod pgtype.Text
		var resolvedIndicator, resolvedBy, deletedBy pgtype.UUID
		var resolvedAt, deletedAt pgtype.Timestamptz
		if err := rows.Scan(
			&row.ObservationID, &row.IncidentID, &row.SourceRecordID,
			&row.SourceFieldKey, &originText, &row.OriginLocator, &row.ObservedText,
			&parsedType, &normalizedCandidate, &row.ResolutionStatus,
			&resolvedIndicator, &row.RowVersion, &row.RuntimeCreatedByID,
			&row.CreatedAt, &resolvedBy, &resolvedAt, &resolutionMethod,
			&deletedAt, &deletedBy,
		); err != nil {
			return nil, errors.New("indicator observation portability export scan failed")
		}
		origin, err := indicatororigin.Parse(originText)
		if err != nil {
			return nil, errors.New("indicator observation portability export origin is invalid")
		}
		row.OriginKind = origin
		row.ParsedIndicatorType = textFromPG(parsedType)
		row.NormalizedCandidate = textFromPG(normalizedCandidate)
		row.ResolvedIndicatorID = uuidFromPG(resolvedIndicator)
		row.RuntimeResolvedByID = uuidFromPG(resolvedBy)
		row.ResolvedAt = timeFromPG(resolvedAt)
		row.ResolutionMethod = textFromPG(resolutionMethod)
		row.DeletedAt = timeFromPG(deletedAt)
		row.RuntimeDeletedByID = uuidFromPG(deletedBy)
		if err := validatePortableObservationForExport(row, exportContext.IncidentID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("indicator observation portability export iteration failed")
	}
	return result, nil
}

func loadPortableIntervals(
	ctx context.Context,
	exportContext sourceport.ExportContext,
) ([]portableIntervalRow, error) {
	rows, err := exportContext.Query.Query(ctx, `
SELECT indicator_state_interval_id, incident_id, indicator_record_id,
       lifecycle_state, valid_from, valid_to, confidence, rationale,
       support_refs, assessor, assessed_at, row_version, created_by_user_id,
       created_at, deleted_at, deleted_by_user_id
  FROM indicator_state_intervals
 WHERE incident_id = $1
 ORDER BY indicator_state_interval_id
`, exportContext.IncidentID)
	if err != nil {
		return nil, errors.New("indicator interval portability export query failed")
	}
	defer rows.Close()
	result := make([]portableIntervalRow, 0)
	for rows.Next() {
		var row portableIntervalRow
		var validTo, deletedAt pgtype.Timestamptz
		var confidence pgtype.Int4
		var rationale, assessor pgtype.Text
		var supportJSON []byte
		var deletedBy pgtype.UUID
		if err := rows.Scan(
			&row.IntervalID, &row.IncidentID, &row.IndicatorRecordID,
			&row.LifecycleState, &row.ValidFrom, &validTo, &confidence, &rationale,
			&supportJSON, &assessor, &row.AssessedAt, &row.RowVersion,
			&row.RuntimeCreatedByID, &row.CreatedAt, &deletedAt, &deletedBy,
		); err != nil {
			return nil, errors.New("indicator interval portability export scan failed")
		}
		row.ValidTo = timeFromPG(validTo)
		if confidence.Valid {
			value := int(confidence.Int32)
			row.Confidence = &value
		}
		row.Rationale = textFromPG(rationale)
		row.Assessor = textFromPG(assessor)
		row.DeletedAt = timeFromPG(deletedAt)
		row.RuntimeDeletedByID = uuidFromPG(deletedBy)
		var supportStrings []string
		if err := json.Unmarshal(supportJSON, &supportStrings); err != nil {
			return nil, errors.New("indicator interval portability support references are invalid")
		}
		row.SupportRefs = make([]uuid.UUID, 0, len(supportStrings))
		for _, raw := range supportStrings {
			parsed, err := uuid.Parse(raw)
			if err != nil || parsed.String() != raw {
				return nil, errors.New("indicator interval portability support reference is invalid")
			}
			row.SupportRefs = append(row.SupportRefs, parsed)
		}
		if err := validatePortableIntervalForExport(row, exportContext.IncidentID); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("indicator interval portability export iteration failed")
	}
	return result, nil
}

func encodePortableIndicators(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	rows []portableIndicatorRow,
) ([]byte, error) {
	created := portableActorMap(ctx, exportContext, "indicators", "created_by_user_id", indicatorIDs(rows, false))
	updated := portableActorMap(ctx, exportContext, "indicators", "updated_by_user_id", indicatorIDs(rows, false))
	deleted := portableActorMap(ctx, exportContext, "indicators", "deleted_by_user_id", indicatorIDs(rows, true))
	recordCreated := portableActorMap(ctx, exportContext, "records", "created_by_user_id", indicatorIDs(rows, false))
	recordUpdated := portableActorMap(ctx, exportContext, "records", "updated_by_user_id", indicatorIDs(rows, false))
	recordDeleted := portableActorMap(ctx, exportContext, "records", "deleted_by_user_id", indicatorIDs(rows, true))
	if created.err != nil || updated.err != nil || deleted.err != nil ||
		recordCreated.err != nil || recordUpdated.err != nil || recordDeleted.err != nil {
		return nil, errors.New("indicator portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		rowID := row.RecordID.String()
		createdBy, err := portableActorID(created.values[rowID], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		updatedBy, err := portableActorID(updated.values[rowID], row.RuntimeUpdatedByID)
		if err != nil {
			return nil, err
		}
		recordCreatedBy, err := portableActorID(recordCreated.values[rowID], row.RuntimeCreatedByID)
		if err != nil || recordCreatedBy != createdBy {
			return nil, errors.New("indicator portability envelope creation attribution differs")
		}
		recordUpdatedBy, err := portableActorID(recordUpdated.values[rowID], row.RuntimeUpdatedByID)
		if err != nil || recordUpdatedBy != updatedBy {
			return nil, errors.New("indicator portability envelope update attribution differs")
		}
		var deletedAt any
		var deletedBy any
		if row.DeletedAt != nil {
			deletedAt = formatPortableTimestamp(*row.DeletedAt)
			if row.RuntimeDeletedByID == nil {
				return nil, errors.New("indicator portability export deletion tuple is invalid")
			}
			deletedBy, err = portableActorID(deleted.values[rowID], *row.RuntimeDeletedByID)
			if err != nil {
				return nil, err
			}
			recordDeletedBy, recordErr := portableActorID(recordDeleted.values[rowID], *row.RuntimeDeletedByID)
			if recordErr != nil || recordDeletedBy != deletedBy {
				return nil, errors.New("indicator portability envelope deletion attribution differs")
			}
		}
		if err := appendPortableRow(&payload, map[string]any{
			"record_id": rowID, "incident_id": row.IncidentID.String(),
			"indicator_type": row.IndicatorType, "value_kind": row.ValueKind,
			"display_value": row.DisplayValue, "normalized_value": row.NormalizedValue,
			"dedupe_key": row.DedupeKey, "defanged_value": row.DefangedValue,
			"hash_algorithm": row.HashAlgorithm, "hash_value": row.HashValue,
			"stix_pattern": row.STIXPattern, "row_version": row.RowVersion,
			"created_at":         formatPortableTimestamp(row.CreatedAt),
			"updated_at":         formatPortableTimestamp(row.UpdatedAt),
			"created_by_user_id": createdBy, "updated_by_user_id": updatedBy,
			"deleted_at": deletedAt, "deleted_by_user_id": deletedBy,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortableObservations(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	rows []portableObservationRow,
) ([]byte, error) {
	created := portableActorMap(ctx, exportContext, "indicator_observations", "created_by_user_id", observationIDs(rows, "created"))
	resolved := portableActorMap(ctx, exportContext, "indicator_observations", "resolved_by_user_id", observationIDs(rows, "resolved"))
	deleted := portableActorMap(ctx, exportContext, "indicator_observations", "deleted_by_user_id", observationIDs(rows, "deleted"))
	if created.err != nil || resolved.err != nil || deleted.err != nil {
		return nil, errors.New("indicator observation portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		rowID := row.ObservationID.String()
		createdBy, err := portableActorID(created.values[rowID], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		var resolvedBy any
		if row.RuntimeResolvedByID != nil {
			resolvedBy, err = portableActorID(resolved.values[rowID], *row.RuntimeResolvedByID)
			if err != nil {
				return nil, err
			}
		}
		var deletedBy any
		if row.RuntimeDeletedByID != nil {
			deletedBy, err = portableActorID(deleted.values[rowID], *row.RuntimeDeletedByID)
			if err != nil {
				return nil, err
			}
		}
		if err := appendPortableRow(&payload, map[string]any{
			"indicator_observation_id": rowID, "incident_id": row.IncidentID.String(),
			"source_record_id": row.SourceRecordID.String(), "source_field_key": row.SourceFieldKey,
			"origin_kind": row.OriginKind.String(), "origin_locator": row.OriginLocator,
			"observed_text": row.ObservedText, "parsed_indicator_type": row.ParsedIndicatorType,
			"normalized_candidate": row.NormalizedCandidate, "resolution_status": row.ResolutionStatus,
			"resolved_indicator_record_id": nullableUUIDString(row.ResolvedIndicatorID),
			"row_version":                  row.RowVersion, "created_by_user_id": createdBy,
			"created_at": formatPortableTimestamp(row.CreatedAt), "resolved_by_user_id": resolvedBy,
			"resolved_at": nullableTimestampString(row.ResolvedAt), "resolution_method": row.ResolutionMethod,
			"deleted_at": nullableTimestampString(row.DeletedAt), "deleted_by_user_id": deletedBy,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

func encodePortableIntervals(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	rows []portableIntervalRow,
) ([]byte, error) {
	created := portableActorMap(ctx, exportContext, "indicator_state_intervals", "created_by_user_id", intervalIDs(rows, false))
	deleted := portableActorMap(ctx, exportContext, "indicator_state_intervals", "deleted_by_user_id", intervalIDs(rows, true))
	if created.err != nil || deleted.err != nil {
		return nil, errors.New("indicator interval portability attribution resolution failed")
	}
	var payload bytes.Buffer
	for _, row := range rows {
		rowID := row.IntervalID.String()
		createdBy, err := portableActorID(created.values[rowID], row.RuntimeCreatedByID)
		if err != nil {
			return nil, err
		}
		var deletedBy any
		if row.RuntimeDeletedByID != nil {
			deletedBy, err = portableActorID(deleted.values[rowID], *row.RuntimeDeletedByID)
			if err != nil {
				return nil, err
			}
		}
		if err := appendPortableRow(&payload, map[string]any{
			"indicator_state_interval_id": rowID, "incident_id": row.IncidentID.String(),
			"indicator_record_id": row.IndicatorRecordID.String(), "lifecycle_state": row.LifecycleState,
			"valid_from": formatPortableTimestamp(row.ValidFrom), "valid_to": nullableTimestampString(row.ValidTo),
			"confidence": row.Confidence, "rationale": row.Rationale,
			"support_refs": uuidStrings(row.SupportRefs), "assessor": row.Assessor,
			"assessed_at": formatPortableTimestamp(row.AssessedAt), "row_version": row.RowVersion,
			"created_by_user_id": createdBy, "created_at": formatPortableTimestamp(row.CreatedAt),
			"deleted_at": nullableTimestampString(row.DeletedAt), "deleted_by_user_id": deletedBy,
		}); err != nil {
			return nil, err
		}
	}
	return payload.Bytes(), nil
}

type actorMapResult struct {
	values map[string]string
	err    error
}

func portableActorMap(
	ctx context.Context,
	exportContext sourceport.ExportContext,
	table string,
	column string,
	rowIDs []string,
) actorMapResult {
	if exportContext.PortableAttributions == nil || len(rowIDs) == 0 {
		return actorMapResult{values: map[string]string{}}
	}
	values, err := exportContext.PortableAttributions.ResolvePortableSourceActors(
		ctx, exportContext.Query, exportContext.IncidentID, table, column, rowIDs,
	)
	return actorMapResult{values: values, err: err}
}

func portableActorID(sourceActorID string, runtimeActorID uuid.UUID) (string, error) {
	if sourceActorID == "" {
		if runtimeActorID == uuid.Nil {
			return "", errors.New("indicator portability export actor is invalid")
		}
		return runtimeActorID.String(), nil
	}
	parsed, err := uuid.Parse(sourceActorID)
	if err != nil || parsed.String() != sourceActorID {
		return "", errors.New("indicator portability export attribution is invalid")
	}
	return sourceActorID, nil
}

func appendPortableRow(payload *bytes.Buffer, row map[string]any) error {
	encoded, err := incidentportability.CanonicalJSONString(row)
	if err != nil {
		return errors.New("indicator portability export encoding failed")
	}
	payload.Write(encoded)
	return nil
}

func validatePortableIndicatorForExport(row portableIndicatorRow, incidentID uuid.UUID) error {
	if row.RecordID == uuid.Nil || row.IncidentID != incidentID || row.RowVersion < 1 ||
		row.RuntimeCreatedByID == uuid.Nil || row.RuntimeUpdatedByID == uuid.Nil ||
		row.UpdatedAt.Before(row.CreatedAt) ||
		(row.DeletedAt == nil) != (row.RuntimeDeletedByID == nil) {
		return errors.New("indicator portability export row is invalid")
	}
	canonical, err := identity.Canonicalize(identity.Input{
		IndicatorType: row.IndicatorType, ValueKind: row.ValueKind,
		DisplayValue: row.DisplayValue, NormalizedValue: row.NormalizedValue,
		DefangedValue: row.DefangedValue, HashAlgorithm: row.HashAlgorithm,
		HashValue: row.HashValue, STIXPattern: row.STIXPattern,
	})
	if err != nil || canonical.DisplayValue != row.DisplayValue ||
		!portableStringPointersEqual(canonical.NormalizedValue, row.NormalizedValue) ||
		!portableStringPointersEqual(canonical.HashAlgorithm, row.HashAlgorithm) ||
		!portableStringPointersEqual(canonical.HashValue, row.HashValue) ||
		canonical.DedupeKey != row.DedupeKey {
		return errors.New("indicator portability export identity is invalid")
	}
	if row.DeletedAt != nil && (row.DeletedAt.Before(row.CreatedAt) || row.DeletedAt.After(row.UpdatedAt)) {
		return errors.New("indicator portability export deletion tuple is invalid")
	}
	return nil
}

func validatePortableObservationForExport(row portableObservationRow, incidentID uuid.UUID) error {
	if row.ObservationID == uuid.Nil || row.IncidentID != incidentID || row.SourceRecordID == uuid.Nil ||
		row.RowVersion < 1 || row.RuntimeCreatedByID == uuid.Nil ||
		strings.TrimSpace(row.SourceFieldKey) == "" || strings.TrimSpace(row.OriginLocator) == "" ||
		strings.TrimSpace(row.ObservedText) == "" ||
		(row.DeletedAt == nil) != (row.RuntimeDeletedByID == nil) {
		return errors.New("indicator observation portability export row is invalid")
	}
	if row.ParsedIndicatorType == nil && row.NormalizedCandidate != nil {
		return errors.New("indicator observation portability export candidate is invalid")
	}
	if row.ParsedIndicatorType != nil {
		canonicalType, canonicalCandidate, err := identity.NormalizeObservationCandidate(
			row.ParsedIndicatorType, row.NormalizedCandidate, row.ObservedText,
		)
		if err != nil || !portableStringPointersEqual(canonicalType, row.ParsedIndicatorType) ||
			!portableStringPointersEqual(canonicalCandidate, row.NormalizedCandidate) {
			return errors.New("indicator observation portability export normalization is invalid")
		}
	}
	switch row.ResolutionStatus {
	case "unresolved":
		if row.ResolvedIndicatorID != nil || row.RuntimeResolvedByID != nil ||
			row.ResolvedAt != nil || row.ResolutionMethod != nil {
			return errors.New("indicator observation portability export resolution is invalid")
		}
	case "resolved":
		if row.ResolvedIndicatorID == nil || row.RuntimeResolvedByID == nil ||
			row.ResolvedAt == nil || row.ResolutionMethod == nil || *row.ResolutionMethod == "" {
			return errors.New("indicator observation portability export resolution is invalid")
		}
	case "dismissed":
		if row.ResolvedIndicatorID != nil || row.RuntimeResolvedByID == nil ||
			row.ResolvedAt == nil || row.ResolutionMethod == nil || *row.ResolutionMethod == "" {
			return errors.New("indicator observation portability export resolution is invalid")
		}
	default:
		return errors.New("indicator observation portability export resolution is invalid")
	}
	if (row.ResolvedAt != nil && row.ResolvedAt.Before(row.CreatedAt)) ||
		(row.DeletedAt != nil && (row.DeletedAt.Before(row.CreatedAt) ||
			(row.ResolvedAt != nil && row.DeletedAt.Before(*row.ResolvedAt)))) {
		return errors.New("indicator observation portability export chronology is invalid")
	}
	return nil
}

func validatePortableIntervalForExport(row portableIntervalRow, incidentID uuid.UUID) error {
	if row.IntervalID == uuid.Nil || row.IncidentID != incidentID || row.IndicatorRecordID == uuid.Nil ||
		row.RowVersion < 1 || row.RuntimeCreatedByID == uuid.Nil ||
		strings.TrimSpace(row.LifecycleState) == "" || strings.TrimSpace(row.LifecycleState) != row.LifecycleState ||
		(row.DeletedAt == nil) != (row.RuntimeDeletedByID == nil) {
		return errors.New("indicator interval portability export row is invalid")
	}
	if (row.ValidTo != nil && row.ValidTo.Before(row.ValidFrom)) ||
		(row.Confidence != nil && (*row.Confidence < 0 || *row.Confidence > 100)) ||
		(row.DeletedAt != nil && row.DeletedAt.Before(row.CreatedAt)) {
		return errors.New("indicator interval portability export semantics are invalid")
	}
	return nil
}

func indicatorIDs(rows []portableIndicatorRow, deletedOnly bool) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if !deletedOnly || row.DeletedAt != nil {
			result = append(result, row.RecordID.String())
		}
	}
	return result
}

func observationIDs(rows []portableObservationRow, kind string) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		include := kind == "created" || (kind == "resolved" && row.ResolvedAt != nil) ||
			(kind == "deleted" && row.DeletedAt != nil)
		if include {
			result = append(result, row.ObservationID.String())
		}
	}
	return result
}

func intervalIDs(rows []portableIntervalRow, deletedOnly bool) []string {
	result := make([]string, 0, len(rows))
	for _, row := range rows {
		if !deletedOnly || row.DeletedAt != nil {
			result = append(result, row.IntervalID.String())
		}
	}
	return result
}

func nullableUUIDString(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullableTimestampString(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatPortableTimestamp(*value)
}

func textFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func timeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

func uuidFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	result, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return nil
	}
	return &result
}
