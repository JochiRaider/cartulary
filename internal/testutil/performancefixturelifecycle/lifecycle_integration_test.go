package performancefixturelifecycle

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

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	appfixture "github.com/JochiRaider/cartulary/internal/testutil/appsupport/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const performanceFixtureTestKey = "85a9ceb4cc34f66356baa07b68bf7f3636844beef90aa51ad8b1751d4b046c72"

func TestPerformanceFixtureSnapshotBuilderCreatesFourIsolatedCleanableClones_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	objectStoreHarness := s3test.Start(t)
	env := testEnvironment(os.Environ())
	templateDB := strings.TrimSpace(env[suiteservices.PGTemplateDBEnv])
	if templateDB == "" {
		t.Fatal("service-backed performance fixture integration requires the suite migrated template")
	}
	env[suiteservices.CallModeEnv] = "owned"
	env[suiteservices.SuiteIDEnv] = "ac043-snapshot-integration"
	if env[suiteservices.SuiteRuntimeRootEnv] == "" {
		env[suiteservices.SuiteRuntimeRootEnv] = t.TempDir()
		env[suiteservices.SuiteRuntimeLeaseIDEnv] = "00000000-0000-4000-8000-000000000001"
		env[suiteservices.SuiteRuntimeRunIDEnv] = "performance-fixture-integration"
	}
	env[suiteservices.PGAdminDSNEnv] = postgresHarness.AdminDSN()
	env[suiteservices.PGTemplateDBEnv] = templateDB
	env[suiteservices.S3EndpointEnv] = objectStoreHarness.Endpoint
	env[suiteservices.S3AccessKeyEnv] = objectStoreHarness.AccessKey
	env[suiteservices.S3SecretKeyEnv] = objectStoreHarness.SecretKey
	env[suiteservices.S3SecureEnv] = "false"
	env["CARTULARY_TEST_RESULTS_DIR"] = t.TempDir()
	env["CARTULARY_TEST_RUN_ID"] = "performance-fixture-integration"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	profile := lifecycleTestProfile(t)
	args := BuildArgs{
		FixtureProfileID:     profile.FixtureProfileID,
		SnapshotKey:          performanceFixtureTestKey,
		MigrationDigest:      strings.Repeat("a", 64),
		SourceContractDigest: profile.SourceContractDigest,
		BuilderUnitID:        "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
		ArtifactFile:         filepath.Join(t.TempDir(), "snapshot-build.json"),
	}
	ctx := context.Background()
	build, err := Build(ctx, env, profile, args, appfixture.NewProduction)
	if err != nil {
		t.Fatal(err)
	}
	if build.State != "sealed" || len(build.ContributionReceipts) != 6 || !build.Validation.ConnectionsClosed || len(build.Validation.Counts) == 0 || len(build.Validation.Conditions) == 0 {
		t.Fatalf("builder did not produce one closed sealed contribution set: %#v", build)
	}
	t.Cleanup(func() {
		if err := CleanupSuite(context.Background(), env); err != nil {
			t.Errorf("cleanup performance fixture suite: %v", err)
		}
	})

	rows := []struct{ row, predicate string }{
		{"module.timeline.measurement.committed_timeline_summary_typing_acknowledgment_b615aabfe6", "perf.typing_ack.v1"},
		{"module.timeline.measurement.timeline_blank_row_creation_satisfies_the_paint_afddd2ce13", "perf.timeline_blank_row_create.v1"},
		{"module.timeline.measurement.timeline_summary_arrow_down_selection_satisfies_961a4ec1d3", "perf.timeline_summary_selection_down.v1"},
		{"module.timeline.measurement.timeline_summary_enter_focus_satisfies_the_paint_d03cf54e95", "perf.timeline_summary_focus_edit.v1"},
	}
	fixtures := make([]PreparedFixture, 0, len(rows))
	for index, binding := range rows {
		cloneEnv := cloneEnvironment(env)
		cloneEnv["CARTULARY_FIXTURE_PROFILE_ID"] = args.FixtureProfileID
		cloneEnv["CARTULARY_FIXTURE_SNAPSHOT_KEY"] = args.SnapshotKey
		cloneEnv["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"] = args.BuilderUnitID
		cloneEnv["CARTULARY_FIXTURE_ROW_ID"] = binding.row
		cloneEnv["CARTULARY_FIXTURE_PREDICATE_ID"] = binding.predicate
		cloneEnv["CARTULARY_FIXTURE_CLONE_LEASE_ID"] = "lease-" + binding.predicate
		cloneEnv["CARTULARY_FIXTURE_CLONE_ORDINAL"] = strconv.Itoa(index + 1)
		cloneEnv["CARTULARY_WEB_E2E_RUNTIME_ROOT"] = filepath.Join(t.TempDir(), "predicate-runtime")
		prepared, err := prepareWithProfile(ctx, cloneEnv, profile)
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
		if timelineRows != semanticCount(profile, "timeline_rows") {
			t.Fatalf("clone %d has %d timeline rows", prepared.CloneOrdinal, timelineRows)
		}
	}
	if len(seenDatabases) != 4 || len(seenBuckets) != 4 {
		t.Fatalf("clone resources are not unique: databases=%d buckets=%d", len(seenDatabases), len(seenBuckets))
	}
	templateDBName := templateName(env[suiteservices.SuiteIDEnv], performanceFixtureTestKey)
	quotedTemplateName := pgx.Identifier{templateDBName}.Sanitize()
	if err := withSharedCatalogMutation(ctx, postgresHarness.AdminDSN(), templateDBName, func(admin *sql.DB) error {
		_, err := admin.ExecContext(ctx, `ALTER DATABASE `+quotedTemplateName+` WITH ALLOW_CONNECTIONS true`)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := validateTemplate(ctx, postgresHarness.AdminDSN(), templateDBName, templateOwner(env[suiteservices.SuiteIDEnv], performanceFixtureTestKey)); err == nil {
		t.Fatal("corrupt unsealed performance fixture template was accepted")
	}
	if err := withSharedCatalogMutation(ctx, postgresHarness.AdminDSN(), templateDBName, func(admin *sql.DB) error {
		_, err := admin.ExecContext(ctx, `ALTER DATABASE `+quotedTemplateName+` WITH ALLOW_CONNECTIONS false`)
		return err
	}); err != nil {
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
	for index, prepared := range fixtures {
		metadata := LeaseMetadata{
			DatabaseName:      prepared.DatabaseName,
			Bucket:            prepared.Bucket,
			FixtureProfileID:  prepared.FixtureProfileID,
			SnapshotKey:       prepared.SnapshotKey,
			BuilderUnitID:     prepared.BuilderUnitID,
			RowID:             prepared.RowID,
			PredicateID:       prepared.PredicateID,
			CloneLeaseID:      prepared.CloneLeaseID,
			CloneOrdinal:      prepared.CloneOrdinal,
			RuntimeBundleRoot: prepared.RuntimeBundleRoot,
		}
		if err := cleanupLeaseWithProfile(ctx, env, profile, metadata, CleanupPorts{
			CleanupSessions: RevokeSessions,
			DetectLeaks:     func(context.Context, LeaseMetadata, map[string]string) error { return nil },
			CleanupDatabase: func(ctx context.Context, metadata LeaseMetadata, _ map[string]string) error {
				return dropDatabase(ctx, postgresHarness.AdminDSN(), metadata.DatabaseName)
			},
			CleanupBucket: func(ctx context.Context, metadata LeaseMetadata, _ map[string]string) error {
				return objectStoreHarness.CleanupBucket(ctx, metadata.Bucket)
			},
		}); err != nil {
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
	partialTemplate := templateName(env[suiteservices.SuiteIDEnv], partialKey)
	if err := cloneDatabase(ctx, postgresHarness.AdminDSN(), templateDB, partialTemplate); err != nil {
		t.Fatal(err)
	}
	if err := markTemplateOwned(ctx, postgresHarness.AdminDSN(), partialTemplate, templateOwner(env[suiteservices.SuiteIDEnv], partialKey)); err != nil {
		t.Fatal(err)
	}
	if err := CleanupSuite(ctx, env); err != nil {
		t.Fatal(err)
	}
	admin, err := sql.Open("pgx", postgresHarness.AdminDSN())
	if err != nil {
		t.Fatal(err)
	}
	for _, database := range []string{templateDBName, partialTemplate} {
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
	privateRoot, ok, err := suiteservices.ResolveSuiteRuntimeDir(env)
	if err != nil || !ok {
		t.Fatalf("resolve private suite runtime: ok=%v err=%v", ok, err)
	}
	for _, runtimeRoot := range []string{
		filepath.Join(privateRoot, "performance-fixtures", "templates", suiteservices.ShortHash(env[suiteservices.SuiteIDEnv], 16)),
		filepath.Join(privateRoot, "performance-fixtures", "clones", suiteservices.ShortHash(env[suiteservices.SuiteIDEnv], 16)),
	} {
		if _, err := os.Lstat(runtimeRoot); !os.IsNotExist(err) {
			t.Fatalf("suite cleanup retained runtime root %s: %v", runtimeRoot, err)
		}
	}
}

func testEnvironment(values []string) map[string]string {
	result := make(map[string]string, len(values))
	for _, value := range values {
		key, item, ok := strings.Cut(value, "=")
		if ok {
			result[key] = item
		}
	}
	return result
}

func cloneEnvironment(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func lifecycleTestProfile(t *testing.T) performancefixtureprofile.Profile {
	t.Helper()
	for _, profile := range performancefixtureprofile.Profiles() {
		if profile.Status == "active" {
			profile.FixtureProfileID = "harness_lifecycle_small_v1"
			profile.FixtureVersion = "cartulary.perf.lifecycle.small.v1"
			profile.Seed = 20260817
			profile.SourceContractDigest = strings.Repeat("b", 64)
			receiptCounts := map[string]int{
				"accounts": 2, "sessions": 0,
				"incidents": 1, "memberships": 2, "workspaces": 1,
				"hosts": 2, "identities": 2, "timeline_rows": 40,
				"links": 2, "mentions": 2, "tags": 2, "projection_sets": 3,
			}
			for contributionIndex := range profile.Contributions {
				for countIndex := range profile.Contributions[contributionIndex].ExpectedReceiptCounts {
					count := &profile.Contributions[contributionIndex].ExpectedReceiptCounts[countIndex]
					exact, ok := receiptCounts[count.CountID]
					if !ok {
						t.Fatalf("small lifecycle profile has no bounded value for receipt count %s", count.CountID)
					}
					count.Exact = exact
				}
			}
			semanticCounts := map[string]int{
				"active_sessions": 0, "background_analysts": 2,
				"host_rows": 2, "identity_rows": 2, "timeline_rows": 40,
				"link_assignments": 2, "mention_assignments": 2, "tag_assignments": 2,
			}
			for countIndex := range profile.SemanticExpectations.Counts {
				count := &profile.SemanticExpectations.Counts[countIndex]
				exact, ok := semanticCounts[count.ExpectationID]
				if !ok {
					t.Fatalf("small lifecycle profile has no bounded value for semantic count %s", count.ExpectationID)
				}
				count.Exact = exact
			}
			if len(profile.RuntimeCredentialSets) != 1 {
				t.Fatalf("small lifecycle profile requires one credential set, got %d", len(profile.RuntimeCredentialSets))
			}
			profile.RuntimeCredentialSets[0].AccountCount = 2
			return profile
		}
	}
	t.Fatal("generated active performance fixture profile is missing")
	return performancefixtureprofile.Profile{}
}

func semanticCount(profile performancefixtureprofile.Profile, expectationID string) int {
	for _, expectation := range profile.SemanticExpectations.Counts {
		if expectation.ExpectationID == expectationID {
			return expectation.Exact
		}
	}
	return -1
}
