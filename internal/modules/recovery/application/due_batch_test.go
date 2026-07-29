package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestExecuteDueVerificationBatchUsesIndependentAttemptsAndReturnsFirstFailure(t *testing.T) {
	firstBackup := dueBackupSet(
		"00000000-0000-0000-0000-000000001101",
		time.Date(2026, 7, 29, 12, 0, 0, 0, time.UTC),
	)
	secondBackup := dueBackupSet(
		"00000000-0000-0000-0000-000000001102",
		time.Date(2026, 7, 29, 12, 1, 0, 0, time.UTC),
	)
	attemptIDs := []uuid.UUID{
		uuid.MustParse("00000000-0000-0000-0000-000000001111"),
		uuid.MustParse("00000000-0000-0000-0000-000000001112"),
	}
	nextID := 0
	firstFailure := NewFailure(FailureVerificationWorkbookProbe, errors.New("probe rejected restored rows"))
	var attempted []uuid.UUID
	var remaining []time.Duration

	outcome, err := executeDueVerificationBatch(
		context.Background(),
		[]recovery.BackupSet{firstBackup, secondBackup},
		100*time.Millisecond,
		func() uuid.UUID {
			id := attemptIDs[nextID]
			nextID++
			return id
		},
		func(ctx context.Context, backupSet recovery.BackupSet, attemptID uuid.UUID) (Result, error, bool) {
			attempted = append(attempted, backupSet.BackupSetID)
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("attempt context has no deadline")
			}
			remaining = append(remaining, time.Until(deadline))
			if len(attempted) == 1 {
				time.Sleep(25 * time.Millisecond)
			}
			refBackupSetID := backupSet.BackupSetID
			result := ResultForStoredBackupSet(backupSet)
			result.ArtifactRefs = append(result.ArtifactRefs, ArtifactRefFor(
				"restore_verification",
				recovery.RestoreVerificationArtifactSchemaID,
				"restore_verification:"+attemptID.String(),
				&refBackupSetID,
			))
			if backupSet.BackupSetID == firstBackup.BackupSetID {
				return result, firstFailure, false
			}
			return result, nil, false
		},
	)
	if !errors.Is(err, firstFailure) {
		t.Fatalf("batch error got %v want first due-order failure %v", err, firstFailure)
	}
	if len(attempted) != 2 || attempted[0] != firstBackup.BackupSetID || attempted[1] != secondBackup.BackupSetID {
		t.Fatalf("attempt order got %v", attempted)
	}
	if len(remaining) != 2 || remaining[0] < 70*time.Millisecond || remaining[1] < 70*time.Millisecond {
		t.Fatalf("attempt deadlines were not independently allocated: %v", remaining)
	}
	if outcome.BackupSetID == nil || *outcome.BackupSetID != firstBackup.BackupSetID {
		t.Fatalf("result backup_set_id got %v want first selected backup %s", outcome.BackupSetID, firstBackup.BackupSetID)
	}
	if len(outcome.ArtifactRefs) != 2 ||
		outcome.ArtifactRefs[0].RefID != "restore_verification:"+attemptIDs[0].String() ||
		outcome.ArtifactRefs[1].RefID != "restore_verification:"+attemptIDs[1].String() {
		t.Fatalf("attempt artifact refs got %#v", outcome.ArtifactRefs)
	}

	t.Run("unsafe second attempt stops and preserves the earlier failure", func(t *testing.T) {
		thirdBackup := dueBackupSet(
			"00000000-0000-0000-0000-000000001103",
			time.Date(2026, 7, 29, 12, 2, 0, 0, time.UTC),
		)
		attemptCount := 0
		got, err := executeDueVerificationBatch(
			context.Background(),
			[]recovery.BackupSet{firstBackup, secondBackup, thirdBackup},
			time.Second,
			uuid.New,
			func(_ context.Context, backupSet recovery.BackupSet, _ uuid.UUID) (Result, error, bool) {
				attemptCount++
				switch attemptCount {
				case 1:
					return ResultForStoredBackupSet(backupSet), firstFailure, false
				case 2:
					return ResultForStoredBackupSet(backupSet), NewFailure(
						FailureVerificationInvariantCheck,
						errors.New("target reset failed"),
					), true
				default:
					t.Fatal("batch attempted a backup after unsafe reset failure")
					return Result{}, nil, false
				}
			},
		)
		if attemptCount != 2 {
			t.Fatalf("unsafe batch attempted %d backups want 2", attemptCount)
		}
		if !errors.Is(err, firstFailure) {
			t.Fatalf("unsafe batch error got %v want earlier due-order failure %v", err, firstFailure)
		}
		if got.BackupSetID == nil || *got.BackupSetID != firstBackup.BackupSetID {
			t.Fatalf("unsafe batch result lost first selected identity: %#v", got)
		}
	})
}

func TestExecuteDueVerificationBatchStopsOnAttemptTimeout(t *testing.T) {
	due := []recovery.BackupSet{
		dueBackupSet("00000000-0000-0000-0000-000000001121", time.Now().UTC()),
		dueBackupSet("00000000-0000-0000-0000-000000001122", time.Now().UTC().Add(time.Second)),
	}
	attempts := 0
	outcome, err := executeDueVerificationBatch(
		context.Background(),
		due,
		10*time.Millisecond,
		uuid.New,
		func(ctx context.Context, backupSet recovery.BackupSet, _ uuid.UUID) (Result, error, bool) {
			attempts++
			<-ctx.Done()
			return ResultForStoredBackupSet(backupSet), ctx.Err(), false
		},
	)
	if attempts != 1 {
		t.Fatalf("timed-out batch attempted %d backups want 1", attempts)
	}
	kind, ok := FailureKindOf(err)
	if !ok || kind != FailureTimeoutElapsed {
		t.Fatalf("timed-out batch failure got (%q, %t) err=%v", kind, ok, err)
	}
	if outcome.BackupSetID == nil || *outcome.BackupSetID != due[0].BackupSetID {
		t.Fatalf("timed-out batch lost first selected backup identity: %#v", outcome)
	}
}

func TestDueAttemptMustStopOnlyForUnsafeContinuation(t *testing.T) {
	indeterminate := NewFailure(FailureVerificationObject, &recovery.RestoreStageError{
		Stage: recovery.RestoreStepObjectStoreRestore,
		Cause: errors.New("object write outcome unknown"),
	})
	tests := []struct {
		name string
		err  error
		stop bool
	}{
		{"determinate workbook failure", NewFailure(FailureVerificationWorkbookProbe, errors.New("probe failed")), false},
		{"determinate projection failure", NewFailure(FailureVerificationProjectionRebuild, &recovery.RestoreStageError{Stage: recovery.RestoreStepProjectionRebuild, Cause: errors.New("rebuild failed")}), false},
		{"indeterminate mutation", indeterminate, true},
		{"lease loss", NewFailure(FailureTargetServingTraffic, errors.New("lease lost")), true},
		{"reset failure", NewFailure(FailureVerificationInvariantCheck, errors.New("reset failed")), false},
		{"timeout", NewFailure(FailureTimeoutElapsed, context.DeadlineExceeded), true},
		{"cancellation", context.Canceled, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := dueAttemptMustStop(test.err); got != test.stop {
				t.Fatalf("dueAttemptMustStop(%v) got %t want %t", test.err, got, test.stop)
			}
		})
	}
}

func dueBackupSet(id string, consistencyPointAt time.Time) recovery.BackupSet {
	return recovery.BackupSet{
		BackupSetID:        uuid.MustParse(id),
		ConsistencyPointAt: consistencyPointAt.UTC(),
	}
}
