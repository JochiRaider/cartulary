package workbook

import "github.com/google/uuid"

// MutationResult is Workbook's neutral, immutable-at-outcome-construction
// representation of an owner mutation result.
type MutationResult struct {
	Payload                 map[string]any
	StatusCode              int
	Replayed                bool
	IncidentID              uuid.UUID
	RecordID                uuid.UUID
	ChangeSetID             uuid.UUID
	ClientTxnID             string
	RowVersion              int64
	ViewSchemaID            string
	ChangedFieldKeys        []string
	AdditionalRecordChanges []MutationResult
}

func BuildMutationPayload(viewSchemaID string, changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": viewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}
