package timeline

import "github.com/google/uuid"

func BuildActionPayload(record projectedRecord, changeSetID uuid.UUID, reason *string) map[string]any {
	payload := map[string]any{
		"record_id":             record.RecordID.String(),
		"incident_id":           record.IncidentID.String(),
		"row_version":           record.RowVersion,
		"capture_state":         record.CaptureState,
		"change_set_id":         changeSetID.String(),
		"reason":                derefString(reason),
		"replacement_record_id": formatUUIDPointer(record.ReplacementRecordID),
	}
	return payload
}

func BuildMutationPayload(record projectedRecord, changeSetID uuid.UUID) map[string]any {
	return map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            buildRow(record),
	}
}
