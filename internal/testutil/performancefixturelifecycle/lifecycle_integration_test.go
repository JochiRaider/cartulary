package performancefixturelifecycle

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	appfixture "github.com/JochiRaider/cartulary/internal/testutil/appsupport/performancefixture"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
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
	// Keep the admitted run and runtime capability together. Fixture names get
	// their own suite namespace; this test borrows the physical service owners.
	env[suiteservices.SuiteIDEnv] += "-snapshot-integration"
	if _, ok, err := suiteservices.ResolveSuiteRuntimeDir(env); err != nil || !ok {
		t.Fatalf("service-backed integration requires an admitted runtime: ok=%v err=%v", ok, err)
	}
	env[suiteservices.PGAdminDSNEnv] = postgresHarness.AdminDSN()
	env[suiteservices.PGTemplateDBEnv] = templateDB
	env[suiteservices.S3EndpointEnv] = objectStoreHarness.Endpoint
	env[suiteservices.S3AccessKeyEnv] = objectStoreHarness.AccessKey
	env[suiteservices.S3SecretKeyEnv] = objectStoreHarness.SecretKey
	env[suiteservices.S3SecureEnv] = "true"
	env["SSL_CERT_FILE"] = objectStoreHarness.RootCertificatePath()
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	profile := lifecycleTestProfile(t)
	profile.ObjectPopulation = performancefixtureprofile.ObjectPopulation{ObjectCount: 1, TotalBytes: 13}
	args := BuildArgs{
		FixtureProfileID:     profile.FixtureProfileID,
		SnapshotKey:          performanceFixtureTestKey,
		MigrationDigest:      strings.Repeat("a", 64),
		SourceContractDigest: profile.SourceContractDigest,
		BuilderUnitID:        "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
		ArtifactFile:         filepath.Join(t.TempDir(), "snapshot-build.json"),
	}
	t.Cleanup(func() {
		if err := CleanupSuite(context.Background(), env); err != nil {
			t.Errorf("cleanup performance fixture suite: %v", err)
		}
	})
	ctx := context.Background()
	build, err := Build(ctx, env, profile, args, func(db postgres.DB, objects objectstore.Store, actor authn.UserRecord, key string, profile performancefixtureprofile.Profile) (*fixture.Assembler, error) {
		// A generic binary contribution exercises the lifecycle independently of an
		// Evidence workload. The product fixture still goes through its real owners.
		if err := objects.PutObject(ctx, "snapshot/payload", strings.NewReader("fixture bytes"), 13, "text/plain"); err != nil {
			return nil, err
		}
		return appfixture.NewProduction(db, objects, actor, key, profile)
	})
	if err != nil {
		t.Fatal(err)
	}
	if build.State != "sealed" || len(build.ContributionReceipts) != 6 || !build.Validation.ConnectionsClosed || len(build.Validation.Counts) == 0 || len(build.Validation.Conditions) == 0 {
		t.Fatalf("builder did not produce one closed sealed contribution set: %#v", build)
	}

	if build.ObjectSnapshot == nil || build.ObjectSnapshot.ObjectCount != 1 || build.ObjectSnapshot.TotalBytes != 13 {
		t.Fatal("sealed object population is missing")
	}
	buildPath, err := BuildArtifactPath(env, args)
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteImmutableJSON(buildPath, build); err != nil {
		t.Fatal(err)
	}

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
		admitted, err := readBuildArtifact(cloneEnv, profile)
		if err != nil {
			t.Fatal(err)
		}
		prepared, err := prepareWithProfile(ctx, cloneEnv, profile, admitted)
		if err != nil {
			t.Fatal(err)
		}
		if !prepared.S3Secure || prepared.S3RootCertificatePath != objectStoreHarness.RootCertificatePath() {
			t.Fatal("performance clone lost its verified object-store trust binding")
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
		if prepared.ObjectSnapshot == nil || *prepared.ObjectSnapshot != *build.ObjectSnapshot {
			t.Fatal("clone object proof diverged")
		}
		objects := openCloneObjects(t, ctx, prepared)
		body, info, err := objects.ReadObject(ctx, "snapshot/payload", objectstore.ReadOptions{})
		if err != nil {
			t.Fatal(err)
		}
		raw, err := io.ReadAll(body)
		if closeErr := body.Close(); err != nil || closeErr != nil {
			t.Fatalf("read clone bytes: %v %v", err, closeErr)
		}
		if string(raw) != "fixture bytes" || info.ContentType != "text/plain" || info.Size != 13 {
			t.Fatal("clone did not restore exact bytes and metadata")
		}
		if prepared.CloneOrdinal == 1 {
			if err := objects.PutObject(ctx, "snapshot/payload", strings.NewReader("changed"), 7, "text/plain"); err != nil {
				t.Fatal(err)
			}
		}
		if err := objects.Close(); err != nil {
			t.Fatal(err)
		}
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
			ObjectSnapshot:    prepared.ObjectSnapshot,
		}
		ports := CleanupPorts{
			CleanupSessions: RevokeSessions,
			DetectLeaks:     func(context.Context, LeaseMetadata, map[string]string) error { return nil },
			CleanupDatabase: func(ctx context.Context, metadata LeaseMetadata, _ map[string]string) error {
				return dropDatabase(ctx, postgresHarness.AdminDSN(), metadata.DatabaseName)
			},
			CleanupBucket: func(ctx context.Context, metadata LeaseMetadata, _ map[string]string) error {
				return objectStoreHarness.CleanupBucket(ctx, metadata.Bucket)
			},
		}
		for attempt := 0; attempt < 2; attempt++ {
			if err := cleanupLeaseWithProfile(ctx, env, profile, metadata, ports); err != nil {
				t.Fatalf("cleanup clone %d attempt %d: %v", index+1, attempt+1, err)
			}
		}
		if _, err := os.Lstat(prepared.RuntimeBundleRoot); !os.IsNotExist(err) {
			t.Fatalf("clone %d retained its credential copy: %v", index+1, err)
		}
	}
	// A corrupt source must fail admission, reclaim acquired resources and retain
	// their actual cleanup outcomes without claiming browser execution occurred.
	templateRoot, err := templateRuntimeRoot(env, args.SnapshotKey)
	if err != nil {
		t.Fatal(err)
	}
	source, err := objectstore.NewFilesystemStore(filepath.Join(templateRoot, objectSnapshotDirectory))
	if err != nil {
		t.Fatal(err)
	}
	if err := source.PutObject(ctx, "snapshot/payload", strings.NewReader("corrupt bytes"), 13, "text/plain"); err != nil {
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}
	failedEnv := cloneEnvironment(env)
	failedEnv["CARTULARY_FIXTURE_PROFILE_ID"] = profile.FixtureProfileID
	failedEnv["CARTULARY_FIXTURE_SNAPSHOT_KEY"] = args.SnapshotKey
	failedEnv["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"] = args.BuilderUnitID
	failedEnv["CARTULARY_FIXTURE_ROW_ID"] = "harness.browser.integration.corrupt_object_snapshot"
	failedEnv["CARTULARY_FIXTURE_PREDICATE_ID"] = rows[0].predicate
	failedEnv["CARTULARY_FIXTURE_CLONE_LEASE_ID"] = "lease-corrupt-objects"
	failedEnv["CARTULARY_FIXTURE_CLONE_ORDINAL"] = "5"
	rejected, preparationErr := prepareWithProfile(ctx, failedEnv, profile, build)
	var failed *preparationFailure
	if preparationErr == nil || rejected.DatabaseName != "" || !errors.As(preparationErr, &failed) {
		t.Fatal("corrupt object population admitted a browser or lost cleanup evidence")
	}
	for _, result := range failed.cleanup {
		expected := "not_acquired"
		if result.ResourceClass == "database" || result.ResourceClass == "bucket" {
			expected = "complete"
		}
		if result.Outcome != expected {
			t.Fatalf("failed clone cleanup %s: %s", result.ResourceClass, result.Outcome)
		}
	}
	if err := failedLeaseWithProfile(failedEnv, profile, "clone_failed", preparationErr); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(env["CARTULARY_TEST_RESULTS_DIR"], env["CARTULARY_TEST_RUN_ID"], "performance-fixtures", args.SnapshotKey, "leases", failedEnv["CARTULARY_FIXTURE_ROW_ID"]+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var failureArtifact leaseArtifact
	if err := json.Unmarshal(raw, &failureArtifact); err != nil {
		t.Fatal(err)
	}
	if failureArtifact.ObjectSnapshot != nil || failureArtifact.CreationState != "failed" || failureArtifact.CleanupResults[0].Outcome != "complete" {
		t.Fatal("failed admission evidence is misleading")
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
	if len(leaseFiles) != 5 {
		t.Fatalf("got %d immutable clone lease artifacts, want five", len(leaseFiles))
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
		if profile.Status == "active" && profile.FixtureProfileID == "timeline_large_grid_snapshot_v2" {
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

func openCloneObjects(t *testing.T, ctx context.Context, prepared PreparedFixture) objectstore.Store {
	t.Helper()
	store, err := objectstore.Setup(ctx, objectstore.Settings{BindingKind: "managed_service", Endpoint: prepared.S3Endpoint, AccessKey: prepared.S3AccessKey, SecretKey: prepared.S3SecretKey, Secure: prepared.S3Secure, Bucket: prepared.Bucket, RootCertificatePath: prepared.S3RootCertificatePath}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatal(err)
	}
	return store
}
