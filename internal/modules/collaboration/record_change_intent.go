package collaboration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"

	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type RecordChange struct {
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	ActorUserID      uuid.UUID
	ChangedFieldKeys []string
	ViewSchemaID     string
	ChangeKind       string
	Row              map[string]any
	PatchCells       map[string]any
}

// ChangedCellKeys returns the canonical public cell keys whose JSON values
// differ between two revision snapshots. Source-owner producers use this one
// implementation when deriving deterministic record-change intents.
func ChangedCellKeys(beforeRow map[string]any, afterRow map[string]any) ([]string, error) {
	beforeCells, err := canonicalCells(beforeRow)
	if err != nil {
		return nil, fmt.Errorf("read before revision cells: %w", err)
	}
	afterCells, err := canonicalCells(afterRow)
	if err != nil {
		return nil, fmt.Errorf("read after revision cells: %w", err)
	}

	candidates := make(map[string]struct{}, len(beforeCells)+len(afterCells))
	for key := range beforeCells {
		candidates[key] = struct{}{}
	}
	for key := range afterCells {
		candidates[key] = struct{}{}
	}
	changed := make([]string, 0, len(candidates))
	for key := range candidates {
		beforeValue, beforeOK := beforeCells[key]
		afterValue, afterOK := afterCells[key]
		if beforeOK != afterOK || !bytes.Equal(beforeValue, afterValue) {
			changed = append(changed, key)
		}
	}
	slices.Sort(changed)
	return changed, nil
}

func canonicalCells(row map[string]any) (map[string]json.RawMessage, error) {
	if row == nil {
		return map[string]json.RawMessage{}, nil
	}
	value, ok := row["cells"]
	if !ok || value == nil {
		return map[string]json.RawMessage{}, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var cells map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &cells); err != nil {
		return nil, err
	}
	if cells == nil {
		cells = map[string]json.RawMessage{}
	}
	return cells, nil
}

func NewRecordChangeIntent(change RecordChange, mutationOrdinal int, createdAt time.Time) (EventIntent, error) {
	if change.RecordID == uuid.Nil || change.ChangeSetID == uuid.Nil || change.IncidentID == uuid.Nil ||
		change.ActorUserID == uuid.Nil || change.RowVersion < 1 || change.ViewSchemaID == "" {
		return EventIntent{}, fmt.Errorf("record_change_intent_v1 identity is incomplete")
	}
	changedKeys := append(make([]string, 0, len(change.ChangedFieldKeys)), change.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	patchCells := change.PatchCells
	if patchCells == nil && change.Row != nil && change.ChangeKind == "" {
		patchCells = platformws.BuildViewRowPatch(change.Row, changedKeys)
	}
	payload := platformws.RecordChangePayload(platformws.RecordChange{
		IncidentID:       change.IncidentID,
		RecordID:         change.RecordID,
		RowVersion:       change.RowVersion,
		ChangeSetID:      change.ChangeSetID,
		ClientTxnID:      change.ClientTxnID,
		ActorUserID:      change.ActorUserID,
		ChangedFieldKeys: changedKeys,
		ViewSchemaID:     change.ViewSchemaID,
		ChangeKind:       change.ChangeKind,
		PatchCells:       patchCells,
	})
	intentKey := fmt.Sprintf(
		"record_changed:%s:%s:%d",
		change.ChangeSetID,
		change.RecordID,
		change.RowVersion,
	)
	intent, err := NewEventIntent(
		intentKey,
		change.IncidentID,
		EventFamilyRecordChanged,
		payload,
		change.ChangeSetID.String()+":"+change.RecordID.String(),
		mutationOrdinal,
		createdAt,
	)
	if err != nil {
		return EventIntent{}, err
	}
	intent.SourceChangeSetID = &change.ChangeSetID
	intent.SourceRecordID = &change.RecordID
	intent.SourceRowVersion = &change.RowVersion
	return intent, nil
}
