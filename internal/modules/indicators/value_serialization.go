package indicators

import (
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

func referenceFromRecord(record indicatorRecord) IndicatorReference {
	return IndicatorReference{
		RecordID: record.RecordID, IncidentID: record.IncidentID,
		IndicatorType: record.IndicatorType, ValueKind: record.ValueKind,
		DisplayValue:    record.DisplayValue,
		NormalizedValue: cloneStringPointer(record.NormalizedValue),
		DedupeKey:       record.DedupeKey,
	}
}

func entityVersionID(prefix string, recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("%s:%s:%d", prefix, recordID.String(), rowVersion)
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func buildIndicatorObservationValue(record IndicatorObservationRecord) map[string]any {
	return map[string]any{
		"indicator_observation_id":     record.ObservationID.String(),
		"incident_id":                  record.IncidentID.String(),
		"source_record_id":             record.SourceRecordID.String(),
		"source_field_key":             record.SourceFieldKey,
		"origin_kind":                  record.OriginKind,
		"origin_locator":               record.OriginLocator,
		"observed_text":                record.ObservedText,
		"parsed_indicator_type":        derefString(record.ParsedIndicatorType),
		"normalized_candidate":         derefString(record.NormalizedCandidate),
		"resolution_status":            record.ResolutionStatus,
		"resolved_indicator_record_id": formatUUIDPointer(record.ResolvedIndicatorRecordID),
		"row_version":                  record.RowVersion,
		"created_by_user_id":           record.CreatedByUserID.String(),
		"created_at":                   formatTimestamp(record.CreatedAt),
		"resolved_by_user_id":          formatUUIDPointer(record.ResolvedByUserID),
		"resolved_at":                  formatTimestampPointer(record.ResolvedAt),
		"resolution_method":            derefString(record.ResolutionMethod),
		"deleted_at":                   formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":           formatUUIDPointer(record.DeletedByUserID),
	}
}

func buildIndicatorLifecycleValue(record IndicatorLifecycleIntervalRecord) map[string]any {
	return map[string]any{
		"indicator_state_interval_id": record.IntervalID.String(),
		"incident_id":                 record.IncidentID.String(),
		"indicator_record_id":         record.IndicatorRecordID.String(),
		"lifecycle_state":             record.LifecycleState,
		"valid_from":                  formatTimestamp(record.ValidFrom),
		"valid_to":                    formatTimestampPointer(record.ValidTo),
		"confidence":                  derefInt(record.Confidence),
		"rationale":                   derefString(record.Rationale),
		"support_refs":                indicatorUUIDStrings(record.SupportRefs),
		"assessor":                    derefString(record.Assessor),
		"assessed_at":                 formatTimestamp(record.AssessedAt),
		"row_version":                 record.RowVersion,
		"created_by_user_id":          record.CreatedByUserID.String(),
		"created_at":                  formatTimestamp(record.CreatedAt),
		"deleted_at":                  formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":          formatUUIDPointer(record.DeletedByUserID),
	}
}

func indicatorUUIDStrings(values []uuid.UUID) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.String()
	}
	return result
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func derefInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func stringPointer(value string) *string {
	cloned := value
	return &cloned
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func jsonEqual(left map[string]any, right map[string]any) bool {
	return reflect.DeepEqual(left, right)
}
