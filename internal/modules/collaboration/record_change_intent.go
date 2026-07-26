package collaboration

import (
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

func NewRecordChangeIntent(change RecordChange, mutationOrdinal int, createdAt time.Time) (EventIntent, error) {
	if change.RecordID == uuid.Nil || change.ChangeSetID == uuid.Nil || change.IncidentID == uuid.Nil ||
		change.ActorUserID == uuid.Nil || change.RowVersion < 1 || change.ViewSchemaID == "" {
		return EventIntent{}, fmt.Errorf("record_change_intent_v1 identity is incomplete")
	}
	changedKeys := append([]string(nil), change.ChangedFieldKeys...)
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
