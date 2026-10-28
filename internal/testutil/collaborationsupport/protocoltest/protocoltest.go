package protocoltest

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func RawPayload(t testing.TB, payload any) json.RawMessage {
	t.Helper()
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal websocket fixture: %v", err)
	}
	return data
}

func NewIncidentJobProgressPayload(
	jobID string,
	incidentID uuid.UUID,
	status string,
	progress protocol.JobProgress,
	updatedAt time.Time,
) protocol.JobProgressPayload {
	return protocol.JobProgressPayload{
		JobID: jobID,
		Scope: protocol.JobScope{
			Kind:       protocol.JobScopeKindIncident,
			IncidentID: incidentID.String(),
		},
		Status: status, Progress: progress, UpdatedAt: updatedAt.UTC(),
	}
}

// RecordChangedEvent is an observation of the received payload, never a fixture builder.
type RecordChangedEvent struct {
	Payload          json.RawMessage
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	ActorUserID      uuid.UUID
	ChangedFieldKeys []string
	AffectedViews    []RecordChangedView
	StreamSeq        int64
	EventID          string
	EmittedAt        time.Time
}

type RecordChangedView struct {
	ViewSchemaID string
	ChangeKind   string
	PatchCells   map[string]any
}

func RecordChangeFromSequencedMessage(message protocol.Message) (RecordChangedEvent, error) {
	if err := protocol.ValidateSequencedReplayableMessage(message); err != nil {
		return RecordChangedEvent{}, err
	}
	if message.Type != "record_changed" {
		return RecordChangedEvent{}, fmt.Errorf("message type %q is not record_changed", message.Type)
	}
	var payload struct {
		RecordID         string           `json:"record_id"`
		RowVersion       int64            `json:"row_version"`
		ChangeSetID      string           `json:"change_set_id"`
		ClientTxnID      string           `json:"client_txn_id"`
		ActorUserID      string           `json:"actor_user_id"`
		ChangedFieldKeys []string         `json:"changed_field_keys"`
		AffectedViews    []map[string]any `json:"affected_views"`
	}
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		return RecordChangedEvent{}, err
	}
	if payload.ChangedFieldKeys == nil {
		return RecordChangedEvent{}, fmt.Errorf("record_changed changed_field_keys must be a present array")
	}
	incidentID, _ := uuid.Parse(message.IncidentID)
	recordID, _ := uuid.Parse(payload.RecordID)
	changeSetID, _ := uuid.Parse(payload.ChangeSetID)
	actorUserID, _ := uuid.Parse(payload.ActorUserID)
	emittedAt, _ := time.Parse(time.RFC3339Nano, message.EmittedAt)
	affectedViews := make([]RecordChangedView, 0, len(payload.AffectedViews))
	for _, rawView := range payload.AffectedViews {
		viewSchemaID, _ := rawView["view_schema_id"].(string)
		changeKind, _ := rawView["change_kind"].(string)
		patchCells, _ := rawView["patch_cells"].(map[string]any)
		affectedViews = append(affectedViews, RecordChangedView{
			ViewSchemaID: viewSchemaID, ChangeKind: changeKind, PatchCells: patchCells,
		})
	}
	return RecordChangedEvent{
		Payload:    append(json.RawMessage(nil), message.Payload...),
		IncidentID: incidentID, RecordID: recordID, RowVersion: payload.RowVersion,
		ChangeSetID: changeSetID, ClientTxnID: payload.ClientTxnID, ActorUserID: actorUserID,
		ChangedFieldKeys: payload.ChangedFieldKeys, AffectedViews: affectedViews,
		StreamSeq: *message.StreamSeq, EventID: message.EventID, EmittedAt: emittedAt,
	}, nil
}

func SequencedMessage(
	t testing.TB,
	family string,
	incidentID uuid.UUID,
	eventID string,
	streamSeq int64,
	emittedAt time.Time,
	payload any,
) protocol.Message {
	t.Helper()
	return protocol.Message{
		Type: family, IncidentID: incidentID.String(), EventID: eventID,
		EmittedAt: emittedAt.UTC().Format(time.RFC3339Nano), StreamSeq: &streamSeq,
		Payload: RawPayload(t, payload),
	}
}
