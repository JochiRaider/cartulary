package harnessruntime

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/bootstrap"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// Explicit-only characterization, not an ordinary acquisition/reset hook. The
// broker owns the reset cluster; the declared active envelope covers one peer
// owned by this fixture. Neither cluster serves unrelated selections.
func TestResetCheckpointWorkPreservesRelationStorage(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 180*time.Second)
	defer cancel()
	h := pgtest.Start(t)
	database := h.PrepareIsolatedDatabaseT(t, "reset_characterization")
	db, err := sql.Open("pgx", database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	pool, err := postgres.Setup(ctx, postgres.Settings{BindingKind: "managed_service", DSN: database.Env()[suiteservices.PostgresDSNEnv], Purpose: postgres.PurposeRecovery, ExpectedRole: "cartulary_recovery"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	catalog, err := recoveryassembly.CurrentRecoveryStateCatalog()
	if err != nil {
		t.Fatal(err)
	}
	settings := bootstrap.Settings{ManifestPath: fixtures.Path("bootstrap-admin", "canonical.json")}
	plan := DatabaseResetPlan{Catalog: catalog, ResolveRebuild: recovery.PostgresDerivedRebuild, RestoreBootstrap: func(ctx context.Context, tx pgx.Tx) error { return bootstrap.PreflightTx(ctx, settings, tx) }}
	reset := func(selected DatabaseResetPlan) {
		t.Helper()
		result, err := ResetDatabase(ctx, pool.Pool(), selected)
		if err != nil || result.TransactionSettlement != "committed" || !result.BootstrapAdminRestored || !result.MigrationMetadataPreserved || len(result.SettlementFailures) != 0 {
			t.Fatalf("reset characterization failed: %v", err)
		}
	}
	reset(plan)
	var durable, vacuum bool
	if err := db.QueryRowContext(ctx, `SELECT current_setting('fsync')='on' AND current_setting('full_page_writes')='on' AND current_setting('synchronous_commit')='on' AND current_setting('data_checksums')='on', current_setting('autovacuum')='on'`).Scan(&durable, &vacuum); err != nil || !durable || !vacuum {
		t.Fatal("durability or ordinary vacuum disabled", err)
	}
	storage := func() string {
		t.Helper()
		var value string
		if err := db.QueryRowContext(ctx, `SELECT jsonb_agg(jsonb_build_array(c.oid,pg_relation_filenode(c.oid)) ORDER BY c.oid)::text FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname IN ('public','pg_toast') AND c.relkind IN ('r','i','t','S')`).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	identity := storage()
	type sample struct {
		files, bytes, dead, vacuums int64
		wal                         string
	}
	sampleDB := func() sample {
		t.Helper()
		var s sample
		if err := db.QueryRowContext(ctx, `SELECT (SELECT count(*) FROM pg_ls_dir('base/'||(SELECT oid::text FROM pg_database WHERE datname=current_database()))), pg_database_size(current_database()), (SELECT coalesce(sum(n_dead_tup),0)::bigint FROM pg_stat_user_tables), (SELECT coalesce(sum(autovacuum_count),0)::bigint FROM pg_stat_user_tables), pg_current_wal_insert_lsn()::text`).Scan(&s.files, &s.bytes, &s.dead, &s.vacuums, &s.wal); err != nil {
			t.Fatal(err)
		}
		return s
	}
	checkpoint := func() int64 {
		t.Helper()
		start := time.Now()
		if _, err := db.ExecContext(ctx, "CHECKPOINT"); err != nil {
			t.Fatal(err)
		}
		return time.Since(start).Milliseconds()
	}
	// Only these fixed experimental boundaries request checkpoints. The native
	// collector retains their numeric completions before the broker shuts down.
	settledMS := checkpoint()
	treatments := []map[string]any{}
	for _, treatment := range []struct {
		name      string
		resets    int
		populated bool
	}{
		{"zero", 0, false}, {"one", 1, false}, {"empty_64", 64, false}, {"populated_64", 64, true},
	} {
		before := sampleDB()
		start := time.Now()
		for range treatment.resets {
			if treatment.populated {
				seedTestRuntimeResetRows(t, db)
				if _, err := db.ExecContext(ctx, `INSERT INTO records (incident_id,record_type,created_by_user_id,updated_by_user_id) SELECT i.id,'timeline_event',u.id,u.id FROM incidents i CROSS JOIN users u CROSS JOIN generate_series(1,255) WHERE u.is_deployment_admin`); err != nil {
					t.Fatal(err)
				}
			}
			reset(plan)
		}
		executionMS := time.Since(start).Milliseconds()
		pending := sampleDB()
		checkpointMS := checkpoint()
		after := sampleDB()
		if storage() != identity {
			t.Fatal("ordinary reset replaced relation storage")
		}
		var walBytes int64
		if err := db.QueryRowContext(ctx, `SELECT pg_wal_lsn_diff($1::pg_lsn,$2::pg_lsn)::bigint`, pending.wal, before.wal).Scan(&walBytes); err != nil {
			t.Fatal(err)
		}
		inserted := 0
		if treatment.populated {
			inserted = treatment.resets * 256
		}
		treatments = append(treatments, map[string]any{"name": treatment.name, "resets": treatment.resets, "inserted_records": inserted, "execution_ms": executionMS, "checkpoint_ms": checkpointMS, "wal_bytes": walBytes, "files_before": before.files, "files_pending": pending.files, "files_settled": after.files, "database_bytes_before": before.bytes, "database_bytes_after": after.bytes, "estimated_dead_tuples_before": before.dead, "estimated_dead_tuples_after": after.dead, "autovacuum_count_before": before.vacuums, "autovacuum_count_after": after.vacuums})
	}

	// The peer has its own physical owner and never overwrites the reset
	// cluster's environment or attachment. Its drop must finish while a real
	// reset transaction remains open on the other cluster.
	peer, err := pgtest.StartOwned(ctx)
	if err != nil {
		t.Fatal(err)
	}
	peerClosed := false
	t.Cleanup(func() {
		if peerClosed {
			return
		}
		cleanup, done := context.WithTimeout(context.Background(), 15*time.Second)
		defer done()
		if err := peer.Close(cleanup); err != nil {
			t.Error(err)
		}
	})
	if peer.ProvisioningGeneration() == h.ProvisioningGeneration() {
		t.Fatal("exclusive physical generations were shared")
	}
	peerDB, err := peer.PrepareDatabase(ctx, "independent_drop")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if peerClosed {
			return
		}
		if err := peer.DropDatabase(context.Background(), peerDB.Name); err != nil {
			t.Error(err)
		}
	})
	var dropMS int64
	concurrent := plan
	concurrent.RestoreBootstrap = func(work context.Context, tx pgx.Tx) error {
		start := time.Now()
		if err := peer.DropDatabase(work, peerDB.Name); err != nil {
			return err
		}
		dropMS = time.Since(start).Milliseconds()
		return plan.RestoreBootstrap(work, tx)
	}
	crossStart := time.Now()
	reset(concurrent)
	if storage() != identity {
		t.Fatal("concurrent peer cleanup replaced reset storage")
	}
	cleanup, done := context.WithTimeout(context.Background(), 15*time.Second)
	err = peer.Close(cleanup)
	done()
	if err != nil {
		t.Fatal("independent physical cleanup failed", err)
	}
	peerClosed = true

	// Retain only bounded numeric/identity facts. This diagnostic is emitted
	// once by this exact fixture, not another measurement service or policy input.
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "../../.."))
	parent, runID := os.Getenv("CARTULARY_TEST_RESULTS_DIR"), os.Getenv("CARTULARY_TEST_RUN_ID")
	if parent == "" || runID == "" {
		t.Fatal("Make-owned retained identity required")
	}
	if !filepath.IsAbs(parent) {
		parent = filepath.Join(root, parent)
	}
	report := map[string]any{"schema_id": "cartulary.harness_reset_characterization.v1", "run_id": runID, "reset_generation": h.ProvisioningGeneration(), "peer_generation": peer.ProvisioningGeneration(), "durability_verified": durable, "autovacuum_enabled": vacuum, "relation_storage_unchanged": true, "settled_setup_checkpoint_ms": settledMS, "treatments": treatments, "independent_drop_ms": dropMS, "concurrent_reset_cycles": 1, "concurrent_phase_ms": time.Since(crossStart).Milliseconds(), "drop_completed_during_reset_transaction": true, "peer_cleanup_completed": true}
	encoded, err := json.Marshal(report)
	if err != nil || len(encoded) > 16384 {
		t.Fatal("characterization report encoding")
	}
	value, err := harnessschema.Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if err := harnessschema.Validate("cartulary.harness_reset_characterization.v1", value); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(parent, runID, "diagnostics", "reset-characterization.json")
	if err := harnesslocal.PublishNew(file, append(encoded, '\n')); err != nil {
		t.Fatal(fmt.Errorf("publish reset characterization: %w", err))
	}
}
