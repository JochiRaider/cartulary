package revisions

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrRollbackTargetNotFound      = errors.New("revisions: rollback target not found")
	ErrRollbackPreconditionFailed  = errors.New("revisions: rollback precondition failed")
	ErrUnsupportedRollbackMutation = errors.New("revisions: unsupported rollback mutation")
)

type RollbackPreconditionError struct {
	ReasonCode string
}

func (e *RollbackPreconditionError) Error() string {
	return ErrRollbackPreconditionFailed.Error()
}

func (e *RollbackPreconditionError) Unwrap() error {
	return ErrRollbackPreconditionFailed
}

type RollbackResult struct {
	Payload     map[string]any
	IncidentID  uuid.UUID
	ClientTxnID string
	Changes     []RollbackRecordChange
	Replayed    bool
}

type RollbackRecordChange struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type rollbackRecordEnvelope = RecordEnvelope

type rollbackMutationTarget struct {
	ChangeSetID   uuid.UUID
	CreatedAt     time.Time
	Source        string
	SequenceNo    int
	TargetKind    string
	TargetID      string
	OperationKind string
	BeforeValue   map[string]any
	AfterValue    map[string]any
}

type rollbackPlan struct {
	Target            rollbackMutationTarget
	Targets           []rollbackMutationTarget
	Companion         []rollbackMutationTarget
	Affected          []uuid.UUID
	Addressed         uuid.UUID
	RecordType        string
	WholeSet          bool
	RequiresChangeSet bool
	RestoreRevisionNo int64
	RestoreSnapshot   map[string]any
	DeferredErr       error
	ExpectedVersions  map[uuid.UUID]int64
	ApplyOrder        []rollbackPlanStep
}

type rollbackPlanStep struct {
	Order            int
	TargetIdentity   string
	ProviderID       string
	Target           rollbackMutationTarget
	ChangedFieldKeys []string
}

type rollbackApplyResult struct {
	ChangeSetID uuid.UUID
	Changes     []RollbackRecordChange
}

type rollbackProtectedSet struct {
	Affected    []uuid.UUID
	DeferredErr error
}
