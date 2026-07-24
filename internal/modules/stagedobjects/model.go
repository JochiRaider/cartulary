// Package stagedobjects owns the semantic lifecycle for immutable bytes that
// must become reachable only through a later atomic database publication.
package stagedobjects

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

const (
	DefaultMaximumBytes = 64 * 1024 * 1024
	StagingLifetime     = 24 * time.Hour
)

var (
	ErrScopeDenied      = errors.New("staged_object_scope_denied")
	ErrInvalidLifecycle = errors.New("staged_object_invalid_lifecycle")
	ErrDependency       = errors.New("staged_object_dependency_unavailable")
	ErrIntegrity        = errors.New("staged_object_integrity_failure")
	ErrCleanupTimeout   = errors.New("staged_object_cleanup_timeout")
)

type FailureClass string

const (
	FailureDependency FailureClass = "dependency"
	FailureIntegrity  FailureClass = "integrity"
	FailureRetryable  FailureClass = "retryable_unknown"
)

type Failure struct {
	Class    FailureClass
	SafeCode string
	Cause    error
}

func (e *Failure) Error() string {
	if e == nil {
		return ""
	}
	if e.SafeCode != "" {
		return fmt.Sprintf("staged object %s failure: %s", e.Class, e.SafeCode)
	}
	return fmt.Sprintf("staged object %s failure", e.Class)
}

func (e *Failure) Unwrap() error {
	if e == nil {
		return nil
	}
	switch e.Class {
	case FailureDependency:
		return ErrDependency
	case FailureIntegrity:
		return ErrIntegrity
	default:
		return e.Cause
	}
}

func NewFailure(class FailureClass, safeCode string, cause error) error {
	return &Failure{Class: class, SafeCode: safeCode, Cause: cause}
}

func FailureKind(err error) FailureClass {
	var failure *Failure
	if errors.As(err, &failure) {
		return failure.Class
	}
	if errors.Is(err, ErrIntegrity) {
		return FailureIntegrity
	}
	if errors.Is(err, ErrDependency) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return FailureDependency
	}
	return FailureRetryable
}

type Reference struct {
	StagingID string
}

func (r Reference) String() string {
	return r.StagingID
}

type Allocation struct {
	StagingID       string
	OperationID     string
	OwnerProfileID  string
	StorageIdentity string
	ByteSize        int64
	SHA256          string
	StagedAt        time.Time
	ExpiresAt       time.Time
}

type CleanupCandidate struct {
	StagingID          string
	StorageIdentity    string
	DeleteAttemptCount int32
}

type DeletionFailure struct {
	StagingID     string
	AttemptCount  int32
	SafeErrorCode string
	NextAttemptAt time.Time
}

type Repository interface {
	Allocate(context.Context, Allocation) error
	MarkReady(context.Context, string, time.Time) error
	Abandon(context.Context, string, time.Time) error
	PrepareCleanupBatch(context.Context, time.Time, time.Time, int) ([]CleanupCandidate, error)
	RecordDeletionSuccess(context.Context, string) error
	RecordDeletionFailure(context.Context, DeletionFailure) error
}

type ByteWriteOutcome string

const (
	ByteWriteSuccess       ByteWriteOutcome = "success"
	ByteWriteDependency    ByteWriteOutcome = "dependency"
	ByteWriteIndeterminate ByteWriteOutcome = "indeterminate"
	ByteWriteIntegrity     ByteWriteOutcome = "integrity"
)

type DeleteOutcome string

const (
	DeleteSuccess          DeleteOutcome = "success"
	DeleteAbsent           DeleteOutcome = "absent"
	DeleteRetryableUnknown DeleteOutcome = "retryable_unknown"
	DeleteDependency       DeleteOutcome = "dependency"
	DeleteIntegrity        DeleteOutcome = "integrity"
)

type ByteStore interface {
	Put(context.Context, string, []byte) (ByteWriteOutcome, error)
	Delete(context.Context, string) (DeleteOutcome, error)
}

type HealthState struct {
	Available  bool
	ReasonCode string
}

type Health struct {
	mu    sync.RWMutex
	state HealthState
}

func NewHealth() *Health {
	return &Health{state: HealthState{Available: true, ReasonCode: "ready"}}
}

func (h *Health) Available() {
	if h == nil {
		return
	}
	h.mu.Lock()
	h.state = HealthState{Available: true, ReasonCode: "ready"}
	h.mu.Unlock()
}

func (h *Health) Unavailable(err error) {
	if h == nil {
		return
	}
	reason := "cleanup_unavailable"
	var failure *Failure
	if errors.As(err, &failure) && failure.SafeCode != "" {
		reason = failure.SafeCode
	}
	h.mu.Lock()
	h.state = HealthState{Available: false, ReasonCode: reason}
	h.mu.Unlock()
}

func (h *Health) State() HealthState {
	if h == nil {
		return HealthState{Available: true, ReasonCode: "ready"}
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.state
}

type Transfer struct {
	operationID string
	profileID   string
	references  []string
}

func (t Transfer) OperationID() string {
	return t.operationID
}

func (t Transfer) ProfileID() string {
	return t.profileID
}

func (t Transfer) References() []string {
	return append([]string(nil), t.references...)
}

type Publication struct {
	StagingID    string
	ResourceKind string
	ResourceID   string
	ByteSize     int64
	SHA256       string
}

type PublicationCapability interface {
	OperationID() string
	PublishStagedObject(context.Context, Publication) error
}

func PublishTransferred(ctx context.Context, transfer Transfer, publications []Publication, capability PublicationCapability) error {
	if capability == nil ||
		transfer.operationID == "" ||
		transfer.operationID != capability.OperationID() ||
		len(publications) != len(transfer.references) {
		return ErrScopeDenied
	}
	expected := append([]string(nil), transfer.references...)
	actual := make([]string, len(publications))
	for index, publication := range publications {
		if publication.StagingID == "" ||
			publication.ResourceKind == "" ||
			publication.ResourceID == "" ||
			publication.ByteSize < 0 ||
			!validSHA256(publication.SHA256) {
			return ErrInvalidLifecycle
		}
		actual[index] = publication.StagingID
	}
	sort.Strings(actual)
	if !equalStrings(expected, actual) {
		return ErrScopeDenied
	}
	for _, publication := range publications {
		if err := capability.PublishStagedObject(ctx, publication); err != nil {
			return err
		}
	}
	return nil
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
