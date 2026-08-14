package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const performanceFixtureTestKey = "85a9ceb4cc34f66356baa07b68bf7f3636844beef90aa51ad8b1751d4b046c72"

func TestPerformanceFixtureBuildArgumentsRequireCanonicalClosedIdentity(t *testing.T) {
	builder := "fixture_snapshot:default:ac043_large_grid_snapshot_v1:" + performanceFixtureTestKey
	args := []string{
		"--fixture-profile", "ac043_large_grid_snapshot_v1",
		"--snapshot-key", performanceFixtureTestKey,
		"--migration-digest", strings.Repeat("a", 64),
		"--source-contract-digest", strings.Repeat("b", 64),
		"--builder-unit-id", builder,
		"--artifact-file", filepath.Join(t.TempDir(), "snapshot-build.json"),
	}
	parsed, err := parsePerformanceFixtureBuildArgs(args)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.SnapshotKey != performanceFixtureTestKey || parsed.BuilderUnitID != builder {
		t.Fatalf("unexpected canonical builder identity: %#v", parsed)
	}
	for _, mutation := range []func([]string){
		func(values []string) { values[1] = "unknown_profile_v1" },
		func(values []string) { values[3] = strings.ToUpper(performanceFixtureTestKey) },
		func(values []string) { values[9] = builder + "-diverged" },
		func(values []string) { values[11] = "relative.json" },
	} {
		candidate := append([]string(nil), args...)
		mutation(candidate)
		if _, err := parsePerformanceFixtureBuildArgs(candidate); err == nil {
			t.Fatalf("invalid build identity was accepted: %#v", candidate)
		}
	}
}

func TestPerformanceFixtureOwnedNamesAreDeterministicAndBounded(t *testing.T) {
	suiteID := "suite-performance-fixture"
	name := performanceFixtureTemplateName(suiteID, performanceFixtureTestKey)
	if !safePostgresIdentifier(name) || len(name) > 63 {
		t.Fatalf("unsafe template name %q", name)
	}
	wantPrefix := "ct_pfs_" + suiteservices.ShortHash(suiteID, 8) + "_"
	if !strings.HasPrefix(name, wantPrefix) || !strings.HasSuffix(name, performanceFixtureTestKey[:12]) {
		t.Fatalf("template name is not bound to suite and snapshot: %q", name)
	}
	clone, err := newPerformanceFixtureCloneName()
	if err != nil {
		t.Fatal(err)
	}
	if !generatedWebE2EDatabaseName(clone) {
		t.Fatalf("profile clone does not satisfy the bounded janitor grammar: %q", clone)
	}
}

func TestPerformanceFixtureArtifactsAreImmutableAndRedacted(t *testing.T) {
	file := filepath.Join(t.TempDir(), "artifact.json")
	value := map[string]any{"schema_id": "example", "snapshot_key": performanceFixtureTestKey}
	if err := writeImmutableJSON(file, value); err != nil {
		t.Fatal(err)
	}
	if err := writeImmutableJSON(file, value); err == nil {
		t.Fatal("immutable artifact accepted a second write")
	}
	info, err := os.Lstat(file)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("immutable artifact mode is %v", info.Mode())
	}
	failed := failedPerformanceFixtureBuild(performanceFixtureBuildArgs{
		FixtureProfileID:     "ac043_large_grid_snapshot_v1",
		SnapshotKey:          performanceFixtureTestKey,
		MigrationDigest:      strings.Repeat("a", 64),
		SourceContractDigest: strings.Repeat("b", 64),
		BuilderUnitID:        "fixture_snapshot:default:ac043_large_grid_snapshot_v1:" + performanceFixtureTestKey,
	}, "contribution_invalid")
	raw := string(mustJSON(t, failed))
	for _, forbidden := range []string{"database_name", "bucket_name", "password", "email", "runtime_path", "user_id"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("failed build artifact contains forbidden field %q", forbidden)
		}
	}
}

func TestPerformanceFixtureCleanupIsActiveCompleteAndIdempotentlyOwned(t *testing.T) {
	deps := defaultTestDependencies(t)
	env := cloneEnv(deps.env)
	env[suiteservices.ActiveEnv] = "1"
	env[suiteservices.SuiteIDEnv] = "suite-performance-cleanup"
	env[suiteservices.TargetEnv] = "browser-e2e-measurement"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	runtimeRoot := filepath.Join(t.TempDir(), "private-runtime")
	bundle, err := fixture.GenerateRuntimeBundle(performanceFixtureTestKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.WriteRuntimeBundle(runtimeRoot, bundle); err != nil {
		t.Fatal(err)
	}
	metadata := webE2EMetadata{
		DatabaseName:      "ct_0123456789abcdef_web_e2e",
		Bucket:            "ct-0123456789abcdef-web-e2e",
		Target:            "browser-e2e-measurement",
		FixtureProfileID:  "ac043_large_grid_snapshot_v1",
		SnapshotKey:       performanceFixtureTestKey,
		BuilderUnitID:     "fixture_snapshot:default:ac043_large_grid_snapshot_v1:" + performanceFixtureTestKey,
		RowID:             "module.timeline.measurement.test_row",
		PredicateID:       "perf.typing_ack.v1",
		CloneLeaseID:      "lease-000001",
		CloneOrdinal:      1,
		RuntimeBundleRoot: runtimeRoot,
	}
	metadataFile := filepath.Join(t.TempDir(), "metadata.json")
	if err := writeWebE2EMetadata(metadataFile, metadata); err != nil {
		t.Fatal(err)
	}
	var sessionCleaned, databaseCleaned, bucketCleaned bool
	deps.cleanupWebE2ESessions = func(_ context.Context, _ map[string]string, database string) error {
		sessionCleaned = database == metadata.DatabaseName
		return nil
	}
	deps.detectWebE2ELeaks = func(_ context.Context, fixtures []webE2EMetadata, _ map[string]string) error {
		if len(fixtures) != 1 || fixtures[0].DatabaseName != metadata.DatabaseName {
			t.Fatalf("unexpected isolation check: %#v", fixtures)
		}
		return nil
	}
	deps.cleanupWebE2EDB = func(context.Context, webE2EMetadata, map[string]string) error {
		databaseCleaned = true
		return nil
	}
	deps.cleanupWebE2EBucket = func(context.Context, webE2EMetadata, map[string]string) error {
		bucketCleaned = true
		return nil
	}
	if status := run([]string{"cleanup-web-e2e", "--metadata-file", metadataFile}, env, deps.dependencies); status != 0 {
		t.Fatalf("active performance fixture cleanup status = %d", status)
	}
	if !sessionCleaned || !databaseCleaned || !bucketCleaned {
		t.Fatalf("incomplete active cleanup: session=%t database=%t bucket=%t", sessionCleaned, databaseCleaned, bucketCleaned)
	}
	if _, err := os.Lstat(runtimeRoot); !os.IsNotExist(err) {
		t.Fatalf("private runtime bundle root remains after cleanup: %v", err)
	}
	artifactFile := filepath.Join(deps.resultsDir, "wrapper-tests", "performance-fixtures", performanceFixtureTestKey, "leases", metadata.RowID+".json")
	raw, err := os.ReadFile(artifactFile)
	if err != nil {
		t.Fatal(err)
	}
	var artifact performanceFixtureLeaseArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.CleanupState != "complete" || !artifact.CredentialCopyCleanup || !artifact.DatabaseCleanup || !artifact.BucketCleanup || !artifact.SessionCleanup || !artifact.ProcessCleanup {
		t.Fatalf("incomplete lease artifact: %#v", artifact)
	}
}

func TestPerformanceFixtureCleanupRetainsFailureAndContinuesIndependentCleanup(t *testing.T) {
	deps := defaultTestDependencies(t)
	env := cloneEnv(deps.env)
	env[suiteservices.ActiveEnv] = "1"
	env[suiteservices.SuiteIDEnv] = "suite-performance-cleanup-failure"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	runtimeRoot := filepath.Join(t.TempDir(), "private-runtime")
	bundle, err := fixture.GenerateRuntimeBundle(performanceFixtureTestKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.WriteRuntimeBundle(runtimeRoot, bundle); err != nil {
		t.Fatal(err)
	}
	metadata := webE2EMetadata{
		DatabaseName:      "ct_fedcba9876543210_web_e2e",
		Bucket:            "ct-fedcba9876543210-web-e2e",
		FixtureProfileID:  "ac043_large_grid_snapshot_v1",
		SnapshotKey:       performanceFixtureTestKey,
		BuilderUnitID:     "fixture_snapshot:default:ac043_large_grid_snapshot_v1:" + performanceFixtureTestKey,
		RowID:             "module.timeline.measurement.failure_row",
		PredicateID:       "perf.timeline_summary_focus_edit.v1",
		CloneLeaseID:      "lease-failure",
		CloneOrdinal:      2,
		RuntimeBundleRoot: runtimeRoot,
	}
	bucketAttempted := false
	deps.cleanupWebE2EDB = func(context.Context, webE2EMetadata, map[string]string) error {
		return errors.New("injected database cleanup failure")
	}
	deps.cleanupWebE2EBucket = func(context.Context, webE2EMetadata, map[string]string) error {
		bucketAttempted = true
		return nil
	}
	if err := cleanupPerformanceFixtureLease(context.Background(), deps.dependencies, env, metadata); err == nil {
		t.Fatal("injected cleanup failure was accepted")
	}
	if !bucketAttempted {
		t.Fatal("database failure prevented independent bucket cleanup")
	}
	if _, err := os.Lstat(runtimeRoot); !os.IsNotExist(err) {
		t.Fatalf("database failure retained credential copy: %v", err)
	}
	artifactFile := filepath.Join(deps.resultsDir, "wrapper-tests", "performance-fixtures", performanceFixtureTestKey, "leases", metadata.RowID+".json")
	raw, err := os.ReadFile(artifactFile)
	if err != nil {
		t.Fatal(err)
	}
	var artifact performanceFixtureLeaseArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.CleanupState != "failed" || artifact.FailureCode != "database_cleanup_failed" || !artifact.CredentialCopyCleanup || !artifact.BucketCleanup {
		t.Fatalf("cleanup failure artifact lost causal state: %#v", artifact)
	}
}

func TestPerformanceFixtureRuntimeJanitorIsBoundedAndIdentityScoped(t *testing.T) {
	base := t.TempDir()
	stale := filepath.Join(base, strings.Repeat("a", 16))
	recent := filepath.Join(base, strings.Repeat("b", 16))
	invalid := filepath.Join(base, "not-owned")
	for _, directory := range []string{stale, recent, invalid} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, "credential"), []byte("private"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now().UTC()
	old := now.Add(-performanceFixtureRuntimeStaleAge - time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatal(err)
	}
	if err := cleanupStalePerformanceFixtureRuntimeRootsUnder(now, []string{base}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale owned runtime root remains: %v", err)
	}
	for _, directory := range []string{recent, invalid} {
		if _, err := os.Lstat(directory); err != nil {
			t.Fatalf("janitor removed non-stale or unowned root %s: %v", directory, err)
		}
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	file := filepath.Join(t.TempDir(), "value.json")
	if err := writeImmutableJSON(file, value); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}
