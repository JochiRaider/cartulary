package collaboration

import (
	"slices"

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

type RecordChangePublisher struct {
	hub *platformws.Hub
}

func NewRecordChangePublisher(hub *platformws.Hub) *RecordChangePublisher {
	return &RecordChangePublisher{hub: hub}
}

func (p *RecordChangePublisher) Publish(change RecordChange) {
	if p == nil || p.hub == nil || change.RecordID == uuid.Nil || change.ChangeSetID == uuid.Nil || change.ViewSchemaID == "" {
		return
	}
	changedKeys := append([]string(nil), change.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	patchCells := change.PatchCells
	if patchCells == nil && change.Row != nil && change.ChangeKind == "" {
		patchCells = platformws.BuildViewRowPatch(change.Row, changedKeys)
	}
	p.hub.PublishRecordChange(platformws.RecordChange{
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
}
