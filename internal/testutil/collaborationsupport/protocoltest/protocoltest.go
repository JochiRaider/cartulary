package protocoltest

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

func RawPayload(payload any) json.RawMessage {
	if payload == nil {
		return nil
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return nil
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

func CanonicalExtensionResourceChangePayload(
	payload protocol.ExtensionResourceChangePayload,
) protocol.ExtensionResourceChangePayload {
	payload.ExtensionProfileID = strings.TrimSpace(payload.ExtensionProfileID)
	payload.ResourceKind = strings.TrimSpace(payload.ResourceKind)
	payload.ResourceID = strings.TrimSpace(payload.ResourceID)
	payload.ChangeKind = strings.TrimSpace(payload.ChangeKind)
	payload.ReasonCode = strings.TrimSpace(payload.ReasonCode)
	refs := append([]protocol.ExtensionWorkspaceRef(nil), payload.WorkspaceRefs...)
	for index := range refs {
		refs[index].Kind = strings.TrimSpace(refs[index].Kind)
		refs[index].ExtensionProfileID = strings.TrimSpace(refs[index].ExtensionProfileID)
		refs[index].WorkspaceKey = strings.TrimSpace(refs[index].WorkspaceKey)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].WorkspaceKey < refs[j].WorkspaceKey })
	if len(refs) > 0 {
		payload.WorkspaceRefs = refs
	}
	return payload
}

type RecordChangedEvent struct {
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
		IncidentID: incidentID, RecordID: recordID, RowVersion: payload.RowVersion,
		ChangeSetID: changeSetID, ClientTxnID: payload.ClientTxnID, ActorUserID: actorUserID,
		ChangedFieldKeys: payload.ChangedFieldKeys, AffectedViews: affectedViews,
		StreamSeq: *message.StreamSeq, EventID: message.EventID, EmittedAt: emittedAt,
	}, nil
}

func RecordChangePayload(change RecordChangedEvent) map[string]any {
	changedKeys := append([]string(nil), change.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	views := append([]RecordChangedView(nil), change.AffectedViews...)
	slices.SortFunc(views, func(left RecordChangedView, right RecordChangedView) int {
		return strings.Compare(left.ViewSchemaID, right.ViewSchemaID)
	})
	affectedViews := make([]map[string]any, 0, len(views))
	for _, affected := range views {
		view := map[string]any{"view_schema_id": affected.ViewSchemaID, "change_kind": affected.ChangeKind}
		if affected.PatchCells != nil {
			view["change_kind"] = "patch"
			view["patch_cells"] = affected.PatchCells
		}
		affectedViews = append(affectedViews, view)
	}
	return map[string]any{
		"record_id": change.RecordID.String(), "row_version": change.RowVersion,
		"change_set_id": change.ChangeSetID.String(), "client_txn_id": change.ClientTxnID,
		"actor_user_id": change.ActorUserID.String(), "changed_field_keys": changedKeys,
		"affected_views": affectedViews,
	}
}

func SequencedMessage(
	family string,
	incidentID uuid.UUID,
	eventID string,
	streamSeq int64,
	emittedAt time.Time,
	payload any,
) protocol.Message {
	return protocol.Message{
		Type: family, IncidentID: incidentID.String(), EventID: eventID,
		EmittedAt: emittedAt.UTC().Format(time.RFC3339Nano), StreamSeq: &streamSeq,
		Payload: RawPayload(payload),
	}
}
