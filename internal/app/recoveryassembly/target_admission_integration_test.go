package recoveryassembly

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTargetServingAdmissionCancelsOnLeaseLoss_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "recovery-target-admission-loss")
	adminPool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer adminPool.Close()
	recoveryDSN, err := testDB.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(context.Background(), postgres.Settings{
		BindingKind:  "managed_service",
		DSN:          recoveryDSN,
		Purpose:      postgres.PurposeRecovery,
		ExpectedRole: "cartulary_recovery",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	admission, err := AcquireTargetServingAdmission(
		context.Background(),
		pool,
		5*time.Second,
		20*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("acquire target serving admission: %v", err)
	}
	concreteAdmission := admission.(*targetServingAdmission)
	backendPID, err := strconv.ParseInt(concreteAdmission.admission.SessionIdentity(), 10, 64)
	if err != nil {
		t.Fatalf("parse serving lease backend identity: %v", err)
	}
	var terminated bool
	if err := adminPool.QueryRow(context.Background(), `SELECT pg_terminate_backend($1)`, backendPID).Scan(&terminated); err != nil {
		t.Fatalf("terminate serving lease backend: %v", err)
	}
	if !terminated {
		t.Fatal("serving lease backend was not terminated")
	}
	select {
	case <-admission.Context().Done():
	case <-time.After(time.Second):
		t.Fatal("target serving admission did not cancel after lease-session loss")
	}
	if !errors.Is(context.Cause(admission.Context()), application.ErrTargetServingLeaseLost) {
		t.Fatalf("lease-loss context cause = %v", context.Cause(admission.Context()))
	}
	if !errors.Is(admission.AssertHeld(), application.ErrTargetServingLeaseLost) {
		t.Fatalf("lease-loss assertion = %v", admission.AssertHeld())
	}
	settleCtx, cancelSettle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelSettle()
	for range 2 {
		if err := admission.Release(settleCtx); !errors.Is(err, application.ErrTargetServingLeaseLost) {
			t.Fatalf("release lost target serving admission: %v", err)
		}
	}
	if !errors.Is(context.Cause(admission.Context()), application.ErrTargetServingLeaseLost) {
		t.Fatalf("release replaced the primary lease-loss cause: %v", context.Cause(admission.Context()))
	}
	replacement, err := AcquireTargetServingAdmission(settleCtx, pool, 5*time.Second, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("acquire independent admission after lost session cleanup: %v", err)
	}
	if err := replacement.Release(settleCtx); err != nil {
		t.Fatalf("release independent admission: %v", err)
	}
}
