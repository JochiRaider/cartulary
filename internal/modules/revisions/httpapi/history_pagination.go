package httpapi

import (
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
	"strconv"
	"time"
)

const historyPositionVersion = "history.position.v1"

// Disposable History positions have their own version inside the shared protected
// envelope. Legacy anchors restart through the ordinary invalid-pagination path.
func historyPosition(cursor *pagination.Cursor) (*revisions.HistoryPosition, error) {
	if cursor == nil {
		return nil, nil
	}
	value := cursor.Position
	if cursor.Mode != pagination.ModeKeyset || len(value) != 6 || value["version"] != historyPositionVersion {
		return nil, pagination.ErrInvalidCursorToken
	}
	committed, err := time.Parse(time.RFC3339Nano, value["committed_at"])
	if err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	changeSet, err := uuid.Parse(value["change_set_id"])
	if err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	sequence, err := strconv.Atoi(value["sequence_no"])
	if err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	revision, err := strconv.ParseInt(value["revision_no"], 10, 64)
	if err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	position := &revisions.HistoryPosition{CommittedAt: committed, ChangeSetID: changeSet, Kind: revisions.HistoryEventKind(value["kind"]), SequenceNo: sequence, RevisionNo: revision}
	if err := position.Validate(); err != nil {
		return nil, pagination.ErrInvalidCursorToken
	}
	return position, nil
}

func historyCursor(binding pagination.Binding, position *revisions.HistoryPosition) *pagination.Cursor {
	if position == nil {
		return nil
	}
	return &pagination.Cursor{
		Version: pagination.CursorVersion, Mode: pagination.ModeKeyset,
		Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope,
		Position: map[string]string{
			"version":       historyPositionVersion,
			"committed_at":  position.CommittedAt.UTC().Format(time.RFC3339Nano),
			"change_set_id": position.ChangeSetID.String(), "kind": string(position.Kind),
			"sequence_no": strconv.Itoa(position.SequenceNo), "revision_no": strconv.FormatInt(position.RevisionNo, 10),
		},
	}
}

// HTTP owns serialization; the application returns typed selected events.
func historyResources(items []revisions.RecordHistoryItem) []map[string]any {
	resources := make([]map[string]any, 0, len(items))
	for _, item := range items {
		actions := append([]string{}, item.AvailableRollbackActions...)
		resource := map[string]any{
			"actor_user_id": item.ActorUserID.String(), "committed_at": item.CommittedAt.UTC().Format(time.RFC3339Nano),
			"history_item_ref": item.HistoryItemRef, "operation": item.Operation, "diff_summary": item.DiffSummary,
			"change_set_id": item.ChangeSetID.String(), "reversible": item.Reversible, "available_rollback_actions": actions,
		}
		if item.HistoryEntryRef != nil {
			resource["history_entry_ref"] = *item.HistoryEntryRef
		}
		if item.SourceActorID != nil {
			resource["source_actor_id"] = *item.SourceActorID
		}
		if item.RevisionNo != nil {
			resource["revision_no"] = *item.RevisionNo
		}
		resources = append(resources, resource)
	}
	return resources
}
