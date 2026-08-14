package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestPerformanceFixtureSnapshotBuilderCreatesFourIsolatedCleanableClones_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	objectStoreHarness := s3test.Start(t)
	env := envMap(os.Environ())
	templateDB := strings.TrimSpace(env[suiteservices.PGTemplateDBEnv])
	if templateDB == "" {
		t.Fatal("service-backed performance fixture integration requires the suite migrated template")
	}
	env[suiteservices.ActiveEnv] = "1"
	env[suiteservices.SuiteIDEnv] = "ac043-snapshot-integration"
	env[suiteservices.PGAdminDSNEnv] = postgresHarness.AdminDSN()
	env[suiteservices.PGTemplateDBEnv] = templateDB
	env[suiteservices.S3EndpointEnv] = objectStoreHarness.Endpoint
	env[suiteservices.S3AccessKeyEnv] = objectStoreHarness.AccessKey
	env[suiteservices.S3SecretKeyEnv] = objectStoreHarness.SecretKey
	env[suiteservices.S3SecureEnv] = "false"
	env["CARTULARY_TEST_RESULTS_DIR"] = t.TempDir()
	env["CARTULARY_TEST_RUN_ID"] = "performance-fixture-integration"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	args := performanceFixtureBuildArgs{
		FixtureProfileID:     "ac043_large_grid_snapshot_v1",
		SnapshotKey:          performanceFixtureTestKey,
		MigrationDigest:      strings.Repeat("a", 64),
		SourceContractDigest: strings.Repeat("b", 64),
		BuilderUnitID:        "fixture_snapshot:default:ac043_large_grid_snapshot_v1:" + performanceFixtureTestKey,
		ArtifactFile:         filepath.Join(t.TempDir(), "snapshot-build.json"),
	}
	ctx := context.Background()
	build, err := buildPerformanceFixture(ctx, env, args)
	if err != nil {
		t.Fatal(err)
	}
	if build.State != "sealed" || len(build.ContributionReceipts) != 6 || !build.Validation.ConnectionsClosed {
		t.Fatalf("builder did not produce one closed sealed contribution set: %#v", build)
	}
	t.Cleanup(func() {
		if err := cleanupPerformanceFixtureSuite(context.Background(), env); err != nil {
			t.Errorf("cleanup performance fixture suite: %v", err)
		}
	})

	rows := []struct{ row, predicate string }{
		{"module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6", "perf.typing_ack.v1"},
		{"module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13", "perf.timeline_blank_row_create.v1"},
		{"module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3", "perf.timeline_summary_selection_down.v1"},
		{"module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95", "perf.timeline_summary_focus_edit.v1"},
	}
	fixtures := make([]webE2EFixture, 0, len(rows))
	for index, binding := range rows {
		cloneEnv := cloneEnv(env)
		cloneEnv["CARTULARY_FIXTURE_PROFILE_ID"] = args.FixtureProfileID
		cloneEnv["CARTULARY_FIXTURE_SNAPSHOT_KEY"] = args.SnapshotKey
		cloneEnv["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"] = args.BuilderUnitID
		cloneEnv["CARTULARY_FIXTURE_ROW_ID"] = binding.row
		cloneEnv["CARTULARY_FIXTURE_PREDICATE_ID"] = binding.predicate
		cloneEnv["CARTULARY_FIXTURE_CLONE_LEASE_ID"] = "lease-" + binding.predicate
		cloneEnv["CARTULARY_FIXTURE_CLONE_ORDINAL"] = strconv.Itoa(index + 1)
		cloneEnv["CARTULARY_WEB_E2E_RUNTIME_ROOT"] = filepath.Join(t.TempDir(), "predicate-runtime")
		prepared, err := preparePerformanceWebE2EFixture(ctx, cloneEnv)
		if err != nil {
			t.Fatal(err)
		}
		fixtures = append(fixtures, prepared)
	}
	if len(fixtures) != 4 {
		t.Fatalf("got %d clones, want four", len(fixtures))
	}
	seenDatabases := map[string]struct{}{}
	seenBuckets := map[string]struct{}{}
	for _, prepared := range fixtures {
		seenDatabases[prepared.DatabaseName] = struct{}{}
		seenBuckets[prepared.Bucket] = struct{}{}
		pool, err := pgxpool.New(ctx, prepared.DSN)
		if err != nil {
			t.Fatal(err)
		}
		var timelineRows int
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM timeline_events`).Scan(&timelineRows); err != nil {
			pool.Close()
			t.Fatal(err)
		}
		pool.Close()
		if timelineRows != 20000 {
			t.Fatalf("clone %d has %d timeline rows", prepared.CloneOrdinal, timelineRows)
		}
	}
	if len(seenDatabases) != 4 || len(seenBuckets) != 4 {
		t.Fatalf("clone resources are not unique: databases=%d buckets=%d", len(seenDatabases), len(seenBuckets))
	}
	admin, err := sql.Open("pgx", postgresHarness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	templateName := performanceFixtureTemplateName(env[suiteservices.SuiteIDEnv], performanceFixtureTestKey)
	quotedTemplateName := pgx.Identifier{templateName}.Sanitize()
	if _, err := admin.ExecContext(ctx, `ALTER DATABASE `+quotedTemplateName+` WITH ALLOW_CONNECTIONS true`); err != nil {
		_ = admin.Close()
		t.Fatal(err)
	}
	if err := validatePerformanceFixtureTemplate(ctx, postgresHarness.AdminDSN(), templateName, performanceFixtureTemplateOwner(env[suiteservices.SuiteIDEnv], performanceFixtureTestKey)); err == nil {
		_ = admin.Close()
		t.Fatal("corrupt unsealed performance fixture template was accepted")
	}
	if _, err := admin.ExecContext(ctx, `ALTER DATABASE `+quotedTemplateName+` WITH ALLOW_CONNECTIONS false`); err != nil {
		_ = admin.Close()
		t.Fatal(err)
	}
	if err := admin.Close(); err != nil {
		t.Fatal(err)
	}
	firstPool, err := pgxpool.New(ctx, fixtures[0].DSN)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := firstPool.Exec(ctx, `CREATE TABLE harness_fixture_isolation_marker (id integer PRIMARY KEY)`); err != nil {
		firstPool.Close()
		t.Fatal(err)
	}
	firstPool.Close()
	for _, prepared := range fixtures[1:] {
		pool, err := pgxpool.New(ctx, prepared.DSN)
		if err != nil {
			t.Fatal(err)
		}
		var isolated bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass('public.harness_fixture_isolation_marker') IS NULL`).Scan(&isolated); err != nil {
			pool.Close()
			t.Fatal(err)
		}
		pool.Close()
		if !isolated {
			t.Fatalf("clone %d observed another clone's mutation", prepared.CloneOrdinal)
		}
	}
	deps := defaultDependencies()
	for index, prepared := range fixtures {
		metadata := webE2EMetadata{
			DatabaseName:      prepared.DatabaseName,
			Bucket:            prepared.Bucket,
			Target:            "browser-e2e-measurement",
			FixtureProfileID:  prepared.FixtureProfileID,
			SnapshotKey:       prepared.SnapshotKey,
			BuilderUnitID:     prepared.BuilderUnitID,
			RowID:             prepared.RowID,
			PredicateID:       prepared.PredicateID,
			CloneLeaseID:      prepared.CloneLeaseID,
			CloneOrdinal:      prepared.CloneOrdinal,
			RuntimeBundleRoot: prepared.RuntimeBundleRoot,
		}
		if err := cleanupPerformanceFixtureLease(ctx, deps, env, metadata); err != nil {
			t.Fatalf("cleanup clone %d: %v", index+1, err)
		}
		if _, err := os.Lstat(prepared.RuntimeBundleRoot); !os.IsNotExist(err) {
			t.Fatalf("clone %d retained its credential copy: %v", index+1, err)
		}
	}
	leaseDirectory := filepath.Join(
		env["CARTULARY_TEST_RESULTS_DIR"],
		env["CARTULARY_TEST_RUN_ID"],
		"performance-fixtures",
		performanceFixtureTestKey,
		"leases",
	)
	leaseFiles, err := os.ReadDir(leaseDirectory)
	if err != nil {
		t.Fatal(err)
	}
	if len(leaseFiles) != 4 {
		t.Fatalf("got %d immutable clone lease artifacts, want four", len(leaseFiles))
	}
	partialKey := strings.Repeat("c", 64)
	partialTemplate := performanceFixtureTemplateName(env[suiteservices.SuiteIDEnv], partialKey)
	if err := cloneDatabase(ctx, postgresHarness.AdminDSN(), templateDB, partialTemplate); err != nil {
		t.Fatal(err)
	}
	if err := markPerformanceFixtureTemplateOwned(ctx, postgresHarness.AdminDSN(), partialTemplate, performanceFixtureTemplateOwner(env[suiteservices.SuiteIDEnv], partialKey)); err != nil {
		t.Fatal(err)
	}
	if err := cleanupPerformanceFixtureSuite(ctx, env); err != nil {
		t.Fatal(err)
	}
	admin, err = sql.Open("pgx", postgresHarness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	for _, database := range []string{templateName, partialTemplate} {
		var exists bool
		if err := admin.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`, database).Scan(&exists); err != nil {
			_ = admin.Close()
			t.Fatal(err)
		}
		if exists {
			_ = admin.Close()
			t.Fatalf("suite cleanup retained owned performance fixture database %q", database)
		}
	}
	if err := admin.Close(); err != nil {
		t.Fatal(err)
	}
	for _, runtimeRoot := range []string{
		filepath.Join(os.TempDir(), "cartulary-performance-fixture-runtime", suiteservices.ShortHash(env[suiteservices.SuiteIDEnv], 16)),
		filepath.Join(os.TempDir(), "cartulary-performance-fixture-clones", suiteservices.ShortHash(env[suiteservices.SuiteIDEnv], 16)),
	} {
		if _, err := os.Lstat(runtimeRoot); !os.IsNotExist(err) {
			t.Fatalf("suite cleanup retained runtime root %s: %v", runtimeRoot, err)
		}
	}
}
