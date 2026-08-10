package rollbackcontract

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrTargetNotFound      = errors.New("rollback provider: target not found")
	ErrStaleTarget         = errors.New("rollback provider: stale target")
	ErrTargetNotReversible = errors.New("rollback provider: target not reversible")
)

type TargetReference struct {
	TargetKind string
	TargetID   string
}

type DispatchClass string

const (
	DispatchRow    DispatchClass = "row"
	DispatchNonRow DispatchClass = "non_row"
)

type SiblingTarget struct {
	TargetReference
	SequenceNo    int
	DispatchClass DispatchClass
}

type NonRowTarget struct {
	IncidentID    uuid.UUID
	ChangeSetID   uuid.UUID
	SequenceNo    int
	TargetKind    string
	TargetID      string
	OperationKind string
	BeforeValue   map[string]any
	AfterValue    map[string]any
}

type DescribeRequest struct {
	Target            NonRowTarget
	AddressedRecordID uuid.UUID
	SiblingTargets    []SiblingTarget
}

type TargetDescriptor struct {
	AffectedRecordIDs          []uuid.UUID
	RequiresWholeChangeSet     bool
	SingleEntryCompanions      []TargetReference
	RequiresWholeChangeSetWith []TargetReference
}

type ApplyInverseRequest struct {
	Target      NonRowTarget
	ActorUserID uuid.UUID
	Now         time.Time
}

type ApplyInverseResult struct {
	AffectedRecordIDs []uuid.UUID
	BeforeValue       map[string]any
	AfterValue        map[string]any
	ChangedFieldKeys  map[uuid.UUID][]string
}

type NonRowTargetProvider interface {
	DescribeTx(context.Context, pgx.Tx, DescribeRequest) (TargetDescriptor, error)
	ApplyInverseTx(context.Context, pgx.Tx, ApplyInverseRequest) (ApplyInverseResult, error)
}

type RestoreRequest struct {
	RecordID       uuid.UUID
	ActorUserID    uuid.UUID
	Now            time.Time
	NextRowVersion int64
	RetainedValue  map[string]any
}

type RowSourceProvider interface {
	ValidateRollbackValue(value map[string]any) error
	RestoreTx(ctx context.Context, tx pgx.Tx, request RestoreRequest) error
}
