package revisions

import (
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
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
	DiffSummary              historycontract.Summary
	ChangeSetID              uuid.UUID
	Reversible               bool
	AvailableRollbackActions []string
	HistoryEntryRef          *string
	RevisionNo               *int64

	sequenceNo     int
	hasTargetEntry bool
}

// HistoryPosition names an immutable logical event, independently of transport encoding.
type HistoryPosition struct {
	CommittedAt time.Time
	ChangeSetID uuid.UUID
	Kind        HistoryEventKind
	SequenceNo  int
	RevisionNo  int64
}
type HistoryEventKind string

const (
	HistoryMutation HistoryEventKind = "mutation"
	HistoryRevision HistoryEventKind = "revision"
)

var ErrInvalidHistoryQuery = errors.New("revisions: invalid history query")

func (position HistoryPosition) Validate() error {
	if position.CommittedAt.IsZero() || position.ChangeSetID == uuid.Nil {
		return ErrInvalidHistoryQuery
	}
	switch position.Kind {
	case HistoryMutation:
		if position.SequenceNo < 1 || position.SequenceNo > 1<<31-1 || position.RevisionNo != 0 {
			return ErrInvalidHistoryQuery
		}
	case HistoryRevision:
		if position.SequenceNo != 0 || position.RevisionNo < 1 {
			return ErrInvalidHistoryQuery
		}
	default:
		return ErrInvalidHistoryQuery
	}
	return nil
}
func (query HistoryQuery) validate() error {
	if query.RecordID == uuid.Nil || query.Limit < 1 || query.Limit > 500 {
		return ErrInvalidHistoryQuery
	}
	if query.After != nil {
		return query.After.Validate()
	}
	return nil
}
