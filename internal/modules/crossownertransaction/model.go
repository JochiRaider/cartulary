// Package crossownertransaction owns Cartulary's bounded, deterministic final
// database-commit protocol. It contains no storage-adapter or profile-owner
// dependencies; application composition supplies both through narrow ports.
package crossownertransaction

import (
	"context"
	"errors"
	"fmt"
	"time"
)

const (
	ParticipantLimit               = 16_384
	ParticipantInputByteLimit      = 16 * 1024 * 1024
	AggregateInputByteLimit        = 64 * 1024 * 1024
	SerializationKeysPerOwner      = 1_024
	AggregateSerializationKeyLimit = 4_096
	ResultByteLimit                = 1 * 1024 * 1024
	ValidationFindingLimit         = 256
)

var (
	ErrUnavailable         = errors.New("cross_owner_transaction_unavailable")
	ErrParticipantSet      = errors.New("cross_owner_transaction_participant_set_invalid")
	ErrInput               = errors.New("cross_owner_transaction_input_invalid")
	ErrPrepare             = errors.New("cross_owner_transaction_prepare_invalid")
	ErrSerializationKeys   = errors.New("cross_owner_transaction_serialization_keys_invalid")
	ErrValidation          = errors.New("cross_owner_transaction_validation_failed")
	ErrWrite               = errors.New("cross_owner_transaction_write_invalid")
	ErrCanceled            = errors.New("cross_owner_transaction_canceled")
	ErrTimeout             = errors.New("cross_owner_transaction_timeout")
	ErrConflict            = errors.New("cross_owner_transaction_conflict")
	ErrCommitAbsent        = errors.New("cross_owner_transaction_commit_absent")
	ErrCommitIndeterminate = errors.New("cross_owner_transaction_commit_indeterminate")
)

const (
	ParticipantContextSchema = "cartulary.extension_transaction_participant_context.v1"
	PrepareResultSchema      = "cartulary.extension_transaction_participant_prepare_result.v1"
	ValidationResultSchema   = "cartulary.extension_transaction_participant_validation_result.v1"
	WriteResultSchema        = "cartulary.extension_transaction_participant_write_result.v1"
)

type Descriptor struct {
	ParticipantID         string
	OwnerProfileID        string
	ContractSHA256        string
	InputSchemaID         string
	PrepareAlgorithmID    string
	ValidationAlgorithmID string
	WriteAlgorithmID      string
	SerializationKeyKinds []string
	OwnedStateFamilyIDs   []string
}

type Operation struct {
	OperationID             string
	NormalizedRequestSHA256 string
	Participants            []Participant
	Finalizer               Finalizer
	Timeout                 time.Duration
}

type Input struct {
	SchemaID       string
	CanonicalBytes []byte
}

type SerializationKey struct {
	KeyKind string
	Key     string
}

type PrepareResult struct {
	SerializationKeys []SerializationKey
}

type Finding struct {
	Path       string
	ReasonCode string
	Message    string
	Details    []byte
}

type ValidationResult struct {
	Status   string
	Findings []Finding
}

func Valid() ValidationResult {
	return ValidationResult{Status: "valid", Findings: []Finding{}}
}

type WriteResult struct {
	Status string
	Value  any
}

func Written(value any) WriteResult {
	return WriteResult{Status: "written", Value: value}
}

type Invocation struct {
	SchemaID                string
	Phase                   string
	OperationID             string
	ParticipantID           string
	OwnerProfileID          string
	NormalizedRequestSHA256 string
	Input                   Input
	CancellationRequested   bool
	DeadlineMonotonicNS     int64
	ReadAccess              ReadCapability
	WriteAccess             WriteCapability
}

type Participant interface {
	ID() string
	BuildInput(context.Context, OperationContext) (Input, error)
	Prepare(context.Context, Invocation) (PrepareResult, error)
	Validate(context.Context, Invocation) (ValidationResult, error)
	Write(context.Context, Invocation) (WriteResult, error)
}

type OperationContext struct {
	OperationID             string
	NormalizedRequestSHA256 string
	DeadlineMonotonicNS     int64
}

// ReadCapability and WriteCapability intentionally expose only an admitted
// participant scope. Owners refine these marker ports to their own typed
// logical operations. Physical transaction objects never cross this boundary.
type ReadCapability interface {
	ParticipantScope() string
}

type WriteCapability interface {
	ReadCapability
}

type FinalizationCapability interface {
	FinalizationScope() string
}

type Finalizer interface {
	Publish(context.Context, FinalizationCapability, map[string]any) error
}

type Backend interface {
	Begin(context.Context, []Descriptor) (Transaction, error)
}

type Transaction interface {
	AcquireSerializationLock(context.Context, OrderedSerializationKey) error
	ReadCapability(string) (ReadCapability, error)
	WriteCapability(string) (WriteCapability, error)
	FinalizationCapability() (FinalizationCapability, error)
	Commit(context.Context) (CommitOutcome, error)
	Rollback(context.Context) (CommitOutcome, error)
}

type OrderedSerializationKey struct {
	ParticipantID string
	KeyKind       string
	Key           string
}

type CommitOutcome string

const (
	CommitProven  CommitOutcome = "committed"
	CommitAbsent  CommitOutcome = "absent"
	CommitUnknown CommitOutcome = "indeterminate"
)

type Result struct {
	ParticipantValues map[string]any
}

type ConflictError struct {
	ReasonCode string
	Cause      error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s: %s", ErrConflict, e.ReasonCode)
}

func (e *ConflictError) Unwrap() error { return e.Cause }

type FatalIntegrityError struct {
	Cause error
}

func (e *FatalIntegrityError) Error() string {
	return ErrCommitIndeterminate.Error()
}

func (e *FatalIntegrityError) Unwrap() error { return e.Cause }

func (e *FatalIntegrityError) FatalReasonCode() string {
	return "indeterminate_database_commit"
}

func IsFatalIntegrity(err error) bool {
	var fatal *FatalIntegrityError
	return errors.As(err, &fatal)
}
