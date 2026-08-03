package revisions

import (
	"time"

	"github.com/google/uuid"
)

type RecordHistoryRecord struct {
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
	RecordType  string
	RowVersion  int64
	Deleted     bool
	DeletedAt   *time.Time
	DeletedByID *uuid.UUID
}

type RecordHistoryItem struct {
	ActorUserID              uuid.UUID
	SourceActorID            *string
	CommittedAt              time.Time
	HistoryItemRef           string
	Operation                string
	DiffSummary              map[string]any
	ChangeSetID              uuid.UUID
	Reversible               bool
	AvailableRollbackActions []string
	HistoryEntryRef          *string
	RevisionNo               *int64

	createdAt      time.Time
	changeSetID    uuid.UUID
	sequenceNo     int
	syntheticRank  int
	targetKey      string
	hasTargetEntry bool
}

func (item RecordHistoryItem) Resource() map[string]any {
	actions := make([]string, 0, len(item.AvailableRollbackActions))
	actions = append(actions, item.AvailableRollbackActions...)
	resource := map[string]any{
		"actor_user_id":              item.ActorUserID.String(),
		"committed_at":               item.CommittedAt.UTC().Format(time.RFC3339Nano),
		"history_item_ref":           item.HistoryItemRef,
		"operation":                  item.Operation,
		"diff_summary":               item.DiffSummary,
		"change_set_id":              item.ChangeSetID.String(),
		"reversible":                 item.Reversible,
		"available_rollback_actions": actions,
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
	return resource
}
