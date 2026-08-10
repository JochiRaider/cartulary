package revisions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrClientTxnConflict     = errors.New("revisions: client transaction conflict")
	ErrIdempotencyNotFound   = errors.New("revisions: idempotency result not found")
	ErrCommandIncidentClosed = errors.New("revisions: command incident closed")
	ErrCommandTargetNotFound = errors.New("revisions: command target not found")
	ErrCommandRoleDenied     = errors.New("revisions: command role denied")
	ErrEnvelopeNotFound      = errors.New("revisions: record envelope not found")
	ErrEnvelopeLockContended = errors.New("revisions: record envelope lock contended")
)

type ActorID uuid.UUID

func NewActorID(value uuid.UUID) ActorID { return ActorID(value) }

func (actor ActorID) UUID() uuid.UUID { return uuid.UUID(actor) }

type TransactionRunner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}

type CommandKind string

const (
	CommandSoftDelete CommandKind = "soft_delete"
	CommandRestore    CommandKind = "restore"
	CommandRollback   CommandKind = "rollback"
)

type CommandAuthorizer interface {
	AuthorizeCommandTx(context.Context, pgx.Tx, uuid.UUID, ActorID, CommandKind) error
}

type IdempotencyKey struct {
	RouteKey    string
	ActorID     ActorID
	ScopeKey    string
	ClientTxnID string
}

type IdempotencyRecord struct {
	RequestHash  []byte
	ResponseJSON []byte
}

type IdempotencyPort interface {
	Get(context.Context, IdempotencyKey) (IdempotencyRecord, error)
	PutSuccessTx(context.Context, pgx.Tx, IdempotencyKey, []byte, map[string]any) error
}

type RecordEnvelope struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	RowVersion      int64
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type EnvelopeLockError struct {
	RecordID uuid.UUID
}

func (err *EnvelopeLockError) Error() string { return ErrEnvelopeLockContended.Error() }

func (err *EnvelopeLockError) Unwrap() error { return ErrEnvelopeLockContended }

type RecordEnvelopePort interface {
	LoadEnvelope(context.Context, uuid.UUID) (RecordEnvelope, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (RecordEnvelope, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	SetDeleteStateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time, bool) (int64, error)
	LockDestructiveRecordsNowaitTx(context.Context, pgx.Tx, []uuid.UUID) error
}

// RecordEnvelopeTxReader is the narrow source-owner capability needed while
// appending Collaboration consequences inside an existing transaction.
type RecordEnvelopeTxReader interface {
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (RecordEnvelope, error)
}

type RecordEnvelopeReader interface {
	LoadEnvelope(context.Context, uuid.UUID) (RecordEnvelope, error)
}

type HistoryQuery struct {
	RecordID uuid.UUID
}

type HistoryResult struct {
	Record    RecordHistoryRecord
	Resources []map[string]any
}

type DeleteRestoreCommand struct {
	Actor       ActorID
	RecordID    uuid.UUID
	Request     DeleteRestoreRequest
	RequestHash []byte
	RequestID   string
	effectiveAt time.Time
}

type RollbackCommand struct {
	Actor       ActorID
	RecordID    uuid.UUID
	Request     RollbackRequest
	RequestHash []byte
	RequestID   string
	effectiveAt time.Time
}
