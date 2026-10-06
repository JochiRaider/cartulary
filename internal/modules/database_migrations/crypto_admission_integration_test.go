package database_migrations_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestCryptoAdmissionFreshConcurrentAndReadOnly_Integration(t *testing.T) {
	h := pgtest.Start(t)
	target := h.NewDatabaseT(t, "crypto-fresh")
	db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer guard.Close()
	// A one-connection work pool remains available while initializers wait on
	// the separate admission pool; there is no shared-pool exhaustion deadlock.
	db.SetMaxOpenConns(1)
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies)
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	var count int
	var format string
	if err := db.QueryRowContext(ctx, `SELECT count(*),min(format_id) FROM public.application_crypto_format`).Scan(&count, &format); err != nil {
		t.Fatal(err)
	}
	if count != 1 || format != migrations.ApplicationCryptoFormat {
		t.Fatalf("identity %d %q", count, format)
	}
	reader, err := pgx.Connect(ctx, target.DSN)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close(ctx)
	if _, err := reader.Exec(ctx, "SET default_transaction_read_only = on"); err != nil {
		t.Fatal(err)
	}
	if err := migrations.RequireApplicationCryptoFormat(ctx, reader); err != nil {
		t.Fatalf("read-only admission: %v", err)
	}
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("borrowed database closed: %v", err)
	}
	// With identity removed, a fully applied database cannot be stamped on retry.
	if _, err := db.ExecContext(ctx, `DELETE FROM public.application_crypto_format`); err != nil {
		t.Fatal(err)
	}
	if err := migrations.RequireApplicationCryptoFormat(ctx, reader); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
		t.Fatalf("read-only missing identity: %v", err)
	}
	if err := migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
		t.Fatalf("missing identity = %v", err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM public.application_crypto_format`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("rejection wrote identity: %d %v", count, err)
	}
}

func TestCryptoAdmissionExternalStateRecheck_Integration(t *testing.T) {
	h := pgtest.Start(t)
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	for _, rejectOn := range []int{1, 2} {
		t.Run(fmt.Sprint(rejectOn), func(t *testing.T) {
			target := h.NewDatabaseT(t, "external-state")
			db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer guard.Close()
			calls := 0
			inspect := func(context.Context) error {
				calls++
				if calls == rejectOn {
					return errors.New("retained storage")
				}
				return nil
			}
			if err := migrations.ApplyWithCryptoAdmission(context.Background(), db, guard, source, inspect); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
				t.Fatalf("external rejection: %v", err)
			}
			if calls != rejectOn {
				t.Fatalf("inspections %d", calls)
			}
			var ledger bool
			if err := db.QueryRow(`SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&ledger); err != nil {
				t.Fatal(err)
			}
			if rejectOn == 1 && ledger {
				t.Fatal("retained storage triggered database mutation")
			}
			if rejectOn == 2 {
				var count int
				if err := db.QueryRow(`SELECT count(*) FROM public.application_crypto_format`).Scan(&count); err != nil || count != 0 {
					t.Fatalf("published identity: %d %v", count, err)
				}
			}
		})
	}
}

func TestCryptoAdmissionRejectsRetainedState_Integration(t *testing.T) {
	h := pgtest.Start(t)
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`CREATE TABLE public.crypto_partial (id integer)`,
		`CREATE TABLE public.goose_db_version (id integer)`,
		`CREATE TABLE public.application_crypto_format (singleton boolean,format_id text); INSERT INTO public.application_crypto_format VALUES (true,'old')`,
	} {
		target := h.NewDatabaseT(t, "crypto-retained")
		db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
		if err != nil {
			t.Fatal(err)
		}
		func() {
			defer db.Close()
			guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
			if err != nil {
				t.Fatal(err)
			}
			defer guard.Close()
			ctx := context.Background()
			if _, err := db.ExecContext(ctx, statement); err != nil {
				t.Fatal(err)
			}
			if err := migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
				t.Fatalf("retained state = %v", err)
			}
			var lineage bool
			if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.schema_migration_lineage') IS NOT NULL`).Scan(&lineage); err != nil || lineage {
				t.Fatalf("rejection mutated schema: %t %v", lineage, err)
			}
		}()
	}
}

func TestCryptoAdmissionCanceledEmptyRetry_Integration(t *testing.T) {
	h := pgtest.Start(t)
	target := h.NewDatabaseT(t, "crypto-empty-retry")
	db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer guard.Close()
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation = %v", err)
	}
	if err := migrations.ApplyWithCryptoAdmission(context.Background(), db, guard, source, emptyCryptoDependencies); err != nil {
		t.Fatalf("empty retry = %v", err)
	}
}

func TestCryptoAdmissionInterruptedInitializationRejectsRetry_Integration(t *testing.T) {
	h := pgtest.Start(t)
	target := h.NewDatabaseT(t, "crypto-interrupted")
	db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer guard.Close()
	source := testMigrationSource(t, `-- +goose Up
SELECT 1 / 0;
-- +goose Down
SELECT 1;
`)
	ctx := context.Background()
	if err := migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies); err == nil {
		t.Fatal("failed migration admitted")
	}
	var ledger, identity bool
	if err := db.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL,to_regclass('public.application_crypto_format') IS NOT NULL`).Scan(&ledger, &identity); err != nil {
		t.Fatal(err)
	}
	if !ledger || identity {
		t.Fatalf("interrupted state ledger=%t identity=%t", ledger, identity)
	}
	canonical, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrations.ApplyWithCryptoAdmission(ctx, db, guard, canonical, emptyCryptoDependencies); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
		t.Fatalf("interrupted retry = %v", err)
	}
	var unlocked bool
	if err := guard.QueryRowContext(ctx, `SELECT pg_try_advisory_lock(4097083627::bigint)`).Scan(&unlocked); err != nil || !unlocked {
		t.Fatalf("lock not released: %v", err)
	}
	if _, err := guard.ExecContext(ctx, `SELECT pg_advisory_unlock(4097083627::bigint)`); err != nil {
		t.Fatal(err)
	}
}

func TestCryptoAdmissionConnectionLossNeverPublishesIdentity_Integration(t *testing.T) {
	h := pgtest.Start(t)
	target := h.NewDatabaseT(t, "crypto-lost-admission")
	db, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	guard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer guard.Close()
	guard.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	var pid int
	if err := guard.QueryRowContext(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
		t.Fatal(err)
	}
	admin := openAdminDatabase(t, h.AdminDSN(), target.Name)
	holder, err := admin.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	if _, err := holder.ExecContext(ctx, `SELECT pg_advisory_lock(4097083626::bigint)`); err != nil {
		t.Fatal(err)
	}
	defer holder.ExecContext(context.Background(), `SELECT pg_advisory_unlock(4097083626::bigint)`)
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	go func() { result <- migrations.ApplyWithCryptoAdmission(ctx, db, guard, source, emptyCryptoDependencies) }()
	for {
		var present bool
		if err := admin.QueryRowContext(ctx, `SELECT to_regclass('public.goose_db_version') IS NOT NULL`).Scan(&present); err != nil {
			t.Fatal(err)
		}
		if present {
			break
		}
		select {
		case err := <-result:
			t.Fatalf("initialization ended before ledger: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if _, err := admin.ExecContext(ctx, `SELECT pg_terminate_backend($1)`, pid); err != nil {
		t.Fatal(err)
	}
	// The current attempt may finish schema work, but its lost admission session
	// cannot commit identity. Another initializer rejects before migration work.
	retryGuard, err := pgtest.OpenPurposeDatabase(target.DSN, postgres.PurposeMigration)
	if err != nil {
		t.Fatal(err)
	}
	defer retryGuard.Close()
	if err := migrations.ApplyWithCryptoAdmission(ctx, db, retryGuard, source, emptyCryptoDependencies); !errors.Is(err, migrations.ErrIncompatibleCryptoState) {
		t.Fatalf("lost-session retry = %v", err)
	}
	if _, err := holder.ExecContext(ctx, `SELECT pg_advisory_unlock(4097083626::bigint)`); err != nil {
		t.Fatal(err)
	}
	if err := <-result; err == nil {
		t.Fatal("lost session published success")
	}
	var count int
	if err := admin.QueryRowContext(ctx, `SELECT count(*) FROM public.application_crypto_format`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("lost admission session published format identity")
	}
}

func emptyCryptoDependencies(context.Context) error { return nil }
