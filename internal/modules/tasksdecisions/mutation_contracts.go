package tasksdecisions

import (
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/contracttasksdecisions"
)

const (
	TaskRequestsViewSchemaID = contracttasksdecisions.TaskRequestViewSchemaID
	DecisionsViewSchemaID    = contracttasksdecisions.DecisionViewSchemaID
)

type CreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]FieldValue
	Collections  map[string]CollectionActionPayload
}

type PatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []PatchChange
}

type PatchChange struct {
	FieldKey       string
	Value          *FieldValue
	Collection     *CollectionActionPayload
	CanonicalValue any
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	LinkedRecordID *uuid.UUID
	ItemRef        string
}

type CreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Request     CreateRequest
	RequestHash []byte
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type PatchCommand struct {
	ActorUserID      uuid.UUID
	RecordID         uuid.UUID
	Request          PatchRequest
	RequestHash      []byte
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type MutationResult struct {
	Row              map[string]any
	Created          bool
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type SupersedeRequest struct {
	BaseRowVersion      int64
	ClientTxnID         string
	Reason              string
	ReplacementRecordID *uuid.UUID
}

type SupersedeCommand struct {
	ActorUserID    uuid.UUID
	TargetRecordID uuid.UUID
	Request        SupersedeRequest
	RequestHash    []byte
	RequestID      string
	RouteKey       string
	Now            time.Time
}

type SupersedeMutationResult struct {
	Row                     map[string]any
	Replayed                bool
	IncidentID              uuid.UUID
	RecordID                uuid.UUID
	ChangeSetID             uuid.UUID
	ClientTxnID             string
	RowVersion              int64
	ViewSchemaID            string
	ChangedFieldKeys        []string
	AdditionalRecordChanges []SupersedeMutationResult
	Facts                   SupersedeFacts
}

type SupersedeFacts struct {
	TargetRecordID        uuid.UUID
	SupersedingRecordID   uuid.UUID
	TargetRowVersion      int64
	SupersedingRowVersion int64
	TargetStatus          string
	Reason                string
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "tasksdecisions: row version conflict"
}

type SameFieldConflictError struct {
	Conflict SameFieldConflict
}

func (e *SameFieldConflictError) Error() string {
	return "tasksdecisions: same field conflict"
}

type OptionalConflictValue struct {
	Present bool
	Value   any
}

type SameFieldConflict struct {
	ConflictToken           string
	RecordID                uuid.UUID
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	ClientValue             any
	ServerValue             any
	BaseValue               OptionalConflictValue
	ServerUpdatedBy         uuid.UUID
	ServerUpdatedAt         time.Time
	SuggestedMergedValue    OptionalConflictValue
}
