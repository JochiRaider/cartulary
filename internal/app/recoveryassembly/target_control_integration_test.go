package recoveryassembly

import (
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/application"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTargetControlTransactionUsesOriginalLeaseSession_Integration(t *testing.T) {
	ctx := t.Context()
	target := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "control-original-session")
	admin, err := pgxpool.New(ctx, target.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close()
	dsn, err := target.DSNForPurpose(postgres.PurposeRecovery)
	if err != nil {
		t.Fatal(err)
	}
	pool, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	lease, err := AcquireTargetServingAdmission(ctx, pool, 5*time.Second, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		settle, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		_ = lease.Release(settle)
	}()
	control := lease.(*targetServingAdmission)
	identity := control.admission.SessionIdentity()
	before, err := recovery.ReadTargetControl(ctx, admin)
	if err != nil {
		t.Fatal(err)
	}
	if err := control.WithControlTransaction(ctx, func(work context.Context, tx recovery.TargetControlTransaction) error {
		var session string
		if err := tx.QueryRow(work, `SELECT pg_backend_pid()::text`).Scan(&session); err != nil {
			return err
		}
		if session != identity {
			return errors.New("control transaction escaped original lease session")
		}
		// Cross a monitor interval while the original connection is occupied.
		if _, err := tx.Exec(work, `SELECT pg_sleep(0.6)`); err != nil {
			return err
		}
		_, err := tx.Exec(work, `UPDATE public.recovery_target_state SET revision=revision+1`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	after, err := recovery.ReadTargetControl(ctx, admin)
	if err != nil || after.Revision != before.Revision+1 {
		t.Fatalf("control commit: revision=%d error=%v", after.Revision, err)
	}
	if err := lease.AssertHeld(); err != nil {
		t.Fatalf("control transaction disrupted lease monitoring: %v", err)
	}
	abort := errors.New("abort control transaction")
	if err := control.WithControlTransaction(ctx, func(work context.Context, tx recovery.TargetControlTransaction) error {
		if _, err := tx.Exec(work, `UPDATE public.recovery_target_state SET revision=revision+1`); err != nil {
			return err
		}
		return abort
	}); !errors.Is(err, abort) {
		t.Fatalf("control rollback error: %v", err)
	}
	rolledBack, err := recovery.ReadTargetControl(ctx, admin)
	if err != nil || rolledBack.Revision != after.Revision {
		t.Fatalf("failed control transaction committed: %+v %v", rolledBack, err)
	}
	pid, err := strconv.ParseInt(identity, 10, 64)
	if err != nil {
		t.Fatal(err)
	}
	err = control.WithControlTransaction(ctx, func(work context.Context, tx recovery.TargetControlTransaction) error {
		if _, err := tx.Exec(work, `UPDATE public.recovery_target_state SET revision=revision+1`); err != nil {
			return err
		}
		var terminated bool
		if err := admin.QueryRow(ctx, `SELECT pg_terminate_backend($1,5000)`, pid).Scan(&terminated); err != nil {
			return err
		}
		if !terminated {
			return errors.New("original lease session was not terminated")
		}
		return nil
	})
	if !errors.Is(err, application.ErrTargetServingLeaseLost) {
		t.Fatalf("lost original session did not close control publication: %v", err)
	}
	lost, err := recovery.ReadTargetControl(ctx, admin)
	if err != nil || lost.Revision != after.Revision {
		t.Fatalf("lost-session control write committed: %+v %v", lost, err)
	}
	called := false
	if err := control.WithControlTransaction(ctx, func(context.Context, recovery.TargetControlTransaction) error { called = true; return nil }); err == nil || called {
		t.Fatal("stale control capability was reused")
	}
}
