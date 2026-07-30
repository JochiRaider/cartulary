package records

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestDestructiveLockContract_Integration(t *testing.T) {
	ctx := context.Background()
	testDB := pgtest.Start(t).PrepareGroupDatabaseT(t, "records-destructive-lock-contract", "lock-ordering")
	lockHolder, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect lock holder: %v", err)
	}
	t.Cleanup(func() { _ = lockHolder.Close(context.Background()) })
	contender, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect contender: %v", err)
	}
	t.Cleanup(func() { _ = contender.Close(context.Background()) })

	actorID, incidentID := seedEnvelopeOwnerContext(t, lockHolder, "destructive-lock")
	lowerRecordID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	higherRecordID := uuid.MustParse("f0000000-0000-4000-8000-000000000001")
	now := time.Date(2026, time.July, 29, 17, 0, 0, 0, time.UTC)
	for _, recordID := range []uuid.UUID{lowerRecordID, higherRecordID} {
		if _, err := lockHolder.Exec(ctx, `
INSERT INTO records (
    record_id, incident_id, record_type,
    created_by_user_id, created_at, updated_by_user_id, updated_at, row_version
) VALUES ($1, $2, 'timeline_event', $3, $4, $3, $4, 1)
`, recordID, incidentID, actorID, now); err != nil {
			t.Fatalf("seed destructive-lock envelope %s: %v", recordID, err)
		}
	}

	holderTx, err := lockHolder.Begin(ctx)
	if err != nil {
		t.Fatalf("begin lock holder transaction: %v", err)
	}
	defer func() { _ = holderTx.Rollback(ctx) }()
	if err := LockDestructiveOperationRecordsNowaitTx(ctx, holderTx, []uuid.UUID{
		higherRecordID,
		lowerRecordID,
	}); err != nil {
		t.Fatalf("lock ordered records: %v", err)
	}

	contenderTx, err := contender.Begin(ctx)
	if err != nil {
		t.Fatalf("begin contender transaction: %v", err)
	}
	defer func() { _ = contenderTx.Rollback(ctx) }()
	err = LockDestructiveOperationRecordsNowaitTx(ctx, contenderTx, []uuid.UUID{
		higherRecordID,
		lowerRecordID,
		lowerRecordID,
	})
	var lockedErr *DestructiveOperationRecordLockedError
	if !errors.As(err, &lockedErr) ||
		!errors.Is(err, ErrDestructiveOperationRecordLocked) ||
		lockedErr.RecordID != lowerRecordID {
		t.Fatalf("lock contention error = %#v, %v", lockedErr, err)
	}
	if err := contenderTx.Rollback(ctx); err != nil {
		t.Fatalf("roll back contender transaction: %v", err)
	}

	checkTx, err := contender.Begin(ctx)
	if err != nil {
		t.Fatalf("begin missing-record transaction: %v", err)
	}
	if err := LockDestructiveOperationRecordsNowaitTx(ctx, checkTx, nil); err != nil {
		_ = checkTx.Rollback(ctx)
		t.Fatalf("empty destructive lock: %v", err)
	}
	if err := LockDestructiveOperationRecordsNowaitTx(ctx, checkTx, []uuid.UUID{uuid.New()}); !errors.Is(err, ErrEnvelopeNotFound) {
		_ = checkTx.Rollback(ctx)
		t.Fatalf("missing destructive-lock envelope error = %v; want %v", err, ErrEnvelopeNotFound)
	}
	_ = checkTx.Rollback(ctx)
}
