package database_migrations_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRecoveryTargetControlInitializationAndPrivileges_Integration(t *testing.T) {
	h := pgtest.Start(t)
	source, err := dbmigrations.Source()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("fresh initialization and purpose privileges", func(t *testing.T) {
		target := h.NewDatabaseT(t, "recovery-control-fresh")
		work := recoveryControlPurposeDB(t, target.DSN, postgres.PurposeMigration)
		guard := recoveryControlPurposeDB(t, target.DSN, postgres.PurposeMigration)
		if err := migrations.ApplyWithCryptoAdmission(t.Context(), work, guard, source, emptyCryptoDependencies); err != nil {
			t.Fatal(err)
		}
		assertInitialRecoveryControl(t, work, "initialized")
		runtime := recoveryControlPurposeDB(t, target.DSN, postgres.PurposeRuntime)
		assertInitialRecoveryControl(t, runtime, "initialized")
		for _, statement := range []string{
			`UPDATE public.recovery_target_state SET revision=revision+1`,
			`DELETE FROM public.recovery_target_state`,
			`TRUNCATE public.recovery_target_state`,
			`INSERT INTO public.recovery_target_state (schema_id,state,revision) VALUES ('cartulary.recovery_target_state.v1','initialized',1)`,
		} {
			_, err := runtime.ExecContext(t.Context(), statement)
			assertRecoveryControlSQLState(t, err, "42501")
		}
		recovery := recoveryControlPurposeDB(t, target.DSN, postgres.PurposeRecovery)
		assertInitialRecoveryControl(t, recovery, "initialized")
		if _, err := recovery.ExecContext(t.Context(), `UPDATE public.recovery_target_state SET revision=revision+1`); err != nil {
			t.Fatalf("Recovery cannot update its control record: %v", err)
		}
		for _, statement := range []string{
			`DELETE FROM public.recovery_target_state`,
			`TRUNCATE public.recovery_target_state`,
			`INSERT INTO public.recovery_target_state (schema_id,state,revision) VALUES ('cartulary.recovery_target_state.v1','initialized',1)`,
		} {
			_, err := recovery.ExecContext(t.Context(), statement)
			assertRecoveryControlSQLState(t, err, "42501")
		}
		for _, statement := range []string{
			`UPDATE public.recovery_target_state SET state='complete'`,
			`UPDATE public.recovery_target_state SET state='restoring'`,
			`UPDATE public.recovery_target_state SET state='resetting'`,
			`UPDATE public.recovery_target_state SET schema_id='retired'`,
			`UPDATE public.recovery_target_state SET revision=0`,
		} {
			_, err := recovery.ExecContext(t.Context(), statement)
			assertRecoveryControlSQLState(t, err, "23514")
		}
	})
	t.Run("retained initialized data becomes unproven", func(t *testing.T) {
		target := h.MigrationDatabaseThroughT(t, 34)
		work := target.SQL()
		if _, err := work.ExecContext(t.Context(), `INSERT INTO public.application_crypto_format (singleton,format_id) VALUES (true,$1)`, migrations.ApplicationCryptoFormat); err != nil {
			t.Fatal(err)
		}
		if _, err := work.ExecContext(t.Context(), `INSERT INTO public.extension_state_metadata
 (profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at)
 VALUES ('reference_pack','reference_pack.state_v1',1,1,'2026-01-01T00:00:00Z','2026-01-01T00:00:00Z')`); err != nil {
			t.Fatal(err)
		}
		var before, after string
		const retained = `SELECT to_jsonb(t)::text FROM public.extension_state_metadata t WHERE profile_id='reference_pack'`
		if err := work.QueryRowContext(t.Context(), retained).Scan(&before); err != nil {
			t.Fatal(err)
		}
		identity, err := target.RuntimeDatabase()
		if err != nil {
			t.Fatal(err)
		}
		guard := recoveryControlPurposeDB(t, identity.DSN, postgres.PurposeMigration)
		if err := migrations.ApplyWithCryptoAdmission(t.Context(), work, guard, source, func(context.Context) error {
			t.Error("retained state was incorrectly treated as fresh initialization")
			return errors.New("unexpected freshness admission")
		}); err != nil {
			t.Fatal(err)
		}
		assertInitialRecoveryControl(t, work, "unproven")
		if err := work.QueryRowContext(t.Context(), retained).Scan(&after); err != nil || after != before {
			t.Fatalf("increment changed retained profile metadata: equal=%t error=%v", before == after, err)
		}
	})
}

func recoveryControlPurposeDB(t *testing.T, dsn string, purpose postgres.Purpose) *sql.DB {
	t.Helper()
	db, err := pgtest.OpenPurposeDatabase(t.Context(), dsn, purpose)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertInitialRecoveryControl(t *testing.T, db *sql.DB, want string) {
	t.Helper()
	var schema, state string
	var revision int64
	var unbound bool
	if err := db.QueryRowContext(t.Context(), `SELECT schema_id,state,revision,
 ROW(target_generation_id,purpose,operation,operation_id,attempt_id,backup_set_id,
 recovery_state_catalog_sha256,codec_registry_sha256,database_sha256,object_store_sha256,
 reference_pack_storage_sha256,export_outputs_sha256,completion_evidence_sha256) IS NULL
 FROM public.recovery_target_state WHERE singleton`).Scan(&schema, &state, &revision, &unbound); err != nil {
		t.Fatal(err)
	}
	if schema != "cartulary.recovery_target_state.v1" || state != want || revision != 1 || !unbound {
		t.Fatalf("unexpected initial control state: schema=%s state=%s revision=%d unbound=%t", schema, state, revision, unbound)
	}
}

func assertRecoveryControlSQLState(t *testing.T, err error, want string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != want {
		t.Fatalf("expected SQLSTATE %s, got %v", want, err)
	}
}
