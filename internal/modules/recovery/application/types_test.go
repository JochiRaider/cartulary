package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestFailureKindsAreClosedUniqueAndRoundTrip_Unit(t *testing.T) {
	kinds := AllFailureKinds()
	if len(kinds) == 0 {
		t.Fatal("failure kind catalog is empty")
	}

	seen := make(map[FailureKind]struct{}, len(kinds))
	for _, kind := range kinds {
		if !kind.Valid() {
			t.Fatalf("catalog contains invalid failure kind %q", kind)
		}
		if _, duplicate := seen[kind]; duplicate {
			t.Fatalf("catalog contains duplicate failure kind %q", kind)
		}
		seen[kind] = struct{}{}

		cause := errors.New("typed recovery failure")
		failure := NewFailure(kind, cause)
		got, ok := FailureKindOf(failure)
		if !ok || got != kind {
			t.Fatalf("failure kind round trip got (%q, %t) want (%q, true)", got, ok, kind)
		}
		if !errors.Is(failure, cause) {
			t.Fatalf("typed failure %q did not preserve its cause", kind)
		}
	}

	kinds[0] = "mutated"
	if AllFailureKinds()[0] == "mutated" {
		t.Fatal("failure kind catalog escaped by reference")
	}
}

func TestEnsureFailurePreservesTypedFailureAndClassifiesDeadline_Unit(t *testing.T) {
	typed := NewFailure(FailureBackupObject, errors.New("object capture failed"))
	if got := EnsureFailure(FailureBackupPublication, typed); got != typed {
		t.Fatal("EnsureFailure replaced an existing typed failure")
	}

	deadline := EnsureFailure(FailureBackupPublication, context.DeadlineExceeded)
	kind, ok := FailureKindOf(deadline)
	if !ok || kind != FailureTimeoutElapsed {
		t.Fatalf("deadline failure got (%q, %t) want (%q, true)", kind, ok, FailureTimeoutElapsed)
	}
	if !errors.Is(deadline, context.DeadlineExceeded) {
		t.Fatal("deadline failure did not preserve context deadline cause")
	}
}

func TestTerminalEvidenceFailureOverridesPriorOperationFailureKind_Unit(t *testing.T) {
	service := Service{
		NewEvidenceRepository: func(PostgresPool) (RecoveryEvidenceRepository, error) {
			return failingCompletionRepository{}, nil
		},
		ProjectFailureEvidence: func(FailureKind) (string, string) {
			return "restore_failed", "invariant_check_failed"
		},
	}
	operationErr := NewFailure(FailureRestoreInvariantCheck, errors.New("restore invariant failed"))
	service.finishJournalAndAudit(
		context.Background(),
		nil,
		operationRequest{
			OperationID: uuid.MustParse("00000000-0000-0000-0000-000000004286"),
			Operation:   OperationRestoreLatest,
			StartedAt:   time.Date(2026, 7, 29, 5, 0, 0, 0, time.UTC),
		},
		Result{},
		&operationErr,
	)
	kind, ok := FailureKindOf(operationErr)
	if !ok || kind != FailureRestoreJournalWrite {
		t.Fatalf("terminal evidence failure kind got (%q, %t) want (%q, true)", kind, ok, FailureRestoreJournalWrite)
	}
}

type failingCompletionRepository struct{}

func (failingCompletionRepository) AppendAdmission(context.Context, RecoveryAdmissionRecord) error {
	return nil
}

func (failingCompletionRepository) AppendCompletion(context.Context, RecoveryCompletionRecord) error {
	return errors.New("injected terminal evidence failure")
}
