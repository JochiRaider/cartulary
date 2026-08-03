package revisions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type rollbackResultCoordinator struct{}

func (rollbackResultCoordinator) payload(record rollbackRecordEnvelope, target RollbackTarget, plan rollbackPlan, applied rollbackApplyResult) map[string]any {
	rowVersion := rollbackPayloadRowVersion(record.RecordID, record.RowVersion, applied.Changes)
	return buildRollbackPayload(record.IncidentID, record.RecordID, rowVersion, target, plan.Target.ChangeSetID, applied.ChangeSetID, plan.Affected)
}

func (rollbackResultCoordinator) replayed(payload map[string]any, clientTxnID string) RollbackResult {
	return rollbackResultFromPayload(payload, clientTxnID)
}

func (rollbackResultCoordinator) committed(payload map[string]any, incidentID uuid.UUID, clientTxnID string, changes []RollbackRecordChange) RollbackResult {
	return RollbackResult{
		Payload:     payload,
		IncidentID:  incidentID,
		ClientTxnID: clientTxnID,
		Changes:     changes,
	}
}

func rollbackPayloadRowVersion(recordID uuid.UUID, fallback int64, changes []RollbackRecordChange) int64 {
	for _, change := range changes {
		if change.RecordID == recordID {
			return change.RowVersion
		}
	}
	return fallback
}

func buildRollbackPayload(incidentID uuid.UUID, recordID uuid.UUID, rowVersion int64, target RollbackTarget, targetChangeSetID uuid.UUID, rollbackChangeSetID uuid.UUID, affected []uuid.UUID) map[string]any {
	affectedText := make([]string, 0, len(affected))
	for _, id := range affected {
		affectedText = append(affectedText, id.String())
	}
	sort.Strings(affectedText)
	return map[string]any{
		"incident_id":            incidentID.String(),
		"record_id":              recordID.String(),
		"row_version":            rowVersion,
		"target":                 target.Normalized(),
		"target_change_set_id":   targetChangeSetID.String(),
		"rollback_change_set_id": rollbackChangeSetID.String(),
		"affected_record_ids":    affectedText,
	}
}

func decodeRollbackValue(data []byte) (map[string]any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("decode rollback value: %w", err)
	}
	return value, nil
}

func decodeStoredRollbackPayload(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode rollback idempotency payload: %w", err)
	}
	return payload, nil
}

func rollbackResultFromPayload(payload map[string]any, clientTxnID string) RollbackResult {
	result := RollbackResult{Payload: payload, ClientTxnID: clientTxnID, Replayed: true}
	if raw, ok := payload["incident_id"].(string); ok {
		result.IncidentID, _ = uuid.Parse(raw)
	}
	if raw, ok := payload["record_id"].(string); ok {
		recordID, _ := uuid.Parse(raw)
		changeSetID := uuid.UUID{}
		if changeSetRaw, ok := payload["rollback_change_set_id"].(string); ok {
			changeSetID, _ = uuid.Parse(changeSetRaw)
		}
		var rowVersion int64
		switch value := payload["row_version"].(type) {
		case float64:
			rowVersion = int64(value)
		case int64:
			rowVersion = value
		}
		result.Changes = []RollbackRecordChange{{RecordID: recordID, RowVersion: rowVersion, ChangeSetID: changeSetID}}
	}
	return result
}

func objectMap(parent map[string]any, key string) (map[string]any, bool) {
	if parent == nil {
		return nil, false
	}
	value, ok := parent[key].(map[string]any)
	return value, ok
}

func rollbackChangedFieldKeys(before map[string]any, after map[string]any) []string {
	beforeCells, _ := objectMap(before, "cells")
	afterCells, _ := objectMap(after, "cells")
	keys := make([]string, 0)
	seen := map[string]struct{}{}
	for key := range beforeCells {
		seen[key] = struct{}{}
	}
	for key := range afterCells {
		seen[key] = struct{}{}
	}
	for key := range seen {
		if !jsonishEqual(beforeCells[key], afterCells[key]) {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}

func jsonishEqual(left any, right any) bool {
	leftRaw, _ := json.Marshal(left)
	rightRaw, _ := json.Marshal(right)
	return bytes.Equal(leftRaw, rightRaw)
}
