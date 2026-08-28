package recoveryassembly

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTargetServingAdmissionCancelsOnLeaseLoss_Integration(t *testing.T) {
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "recovery-target-admission-loss")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
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
	if err := pool.QueryRow(context.Background(), `SELECT pg_terminate_backend($1)`, backendPID).Scan(&terminated); err != nil {
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
	if err := admission.Release(context.Background()); err != nil {
		t.Fatalf("release lost target serving admission: %v", err)
	}
}
