package timeline

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
)

func buildActionPayload(record workbookprojection.DerivedRecord, changeSetID uuid.UUID, reason *string) map[string]any {
	payload := map[string]any{
		"record_id":             record.RecordID.String(),
		"incident_id":           record.IncidentID.String(),
		"row_version":           record.RowVersion,
		"capture_state":         record.CaptureState,
		"change_set_id":         changeSetID.String(),
		"reason":                valuecodec.OptionalString(reason),
		"replacement_record_id": valuecodec.OptionalUUID(record.ReplacementRecordID),
	}
	return payload
}

func buildMutationPayload(record workbookprojection.DerivedRecord, changeSetID uuid.UUID) map[string]any {
	return map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            buildRow(record),
	}
}
