package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/performancefixtureprofile"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

const performanceFixtureTestKey = "85a9ceb4cc34f66356baa07b68bf7f3636844beef90aa51ad8b1751d4b046c72"

func TestPerformanceFixtureBuildArgumentsRequireCanonicalClosedIdentity(t *testing.T) {
	profile := performanceFixtureTestProfile(t)
	builder := "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey
	args := []string{
		"--fixture-profile", profile.FixtureProfileID,
		"--snapshot-key", performanceFixtureTestKey,
		"--migration-digest", strings.Repeat("a", 64),
		"--source-contract-digest", profile.SourceContractDigest,
		"--builder-unit-id", builder,
		"--artifact-file", filepath.Join(t.TempDir(), "snapshot-build.json"),
	}
	parsed, _, err := parsePerformanceFixtureBuildArgs(args)
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
		if _, _, err := parsePerformanceFixtureBuildArgs(candidate); err == nil {
			t.Fatalf("invalid build identity was accepted: %#v", candidate)
		}
	}
}

func TestPerformanceFixtureCleanupIsActiveCompleteAndIdempotentlyOwned(t *testing.T) {
	deps := defaultTestDependencies(t)
	env := cloneEnv(deps.env)
	authorizeSuiteEnv(env)
	env[suiteservices.SuiteIDEnv] = "suite-performance-cleanup"
	env[suiteservices.TargetEnv] = "browser-e2e-measurement"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	profile := performanceFixtureTestProfile(t)
	runtimeRoot := filepath.Join(t.TempDir(), "private-runtime")
	bundle, err := fixture.GenerateRuntimeBundle(profile, performanceFixtureTestKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.WriteRuntimeBundle(profile, runtimeRoot, bundle); err != nil {
		t.Fatal(err)
	}
	metadata := webE2EMetadata{
		DatabaseName:      "ct_0123456789abcdef_web_e2e",
		Bucket:            "ct-0123456789abcdef-web-e2e",
		Target:            "browser-e2e-measurement",
		FixtureProfileID:  profile.FixtureProfileID,
		SnapshotKey:       performanceFixtureTestKey,
		BuilderUnitID:     "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
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
	var artifact leaseArtifactEvidence
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.CleanupState != "complete" || !artifact.cleanupComplete("credential_copy") || !artifact.cleanupComplete("database") || !artifact.cleanupComplete("bucket") || !artifact.cleanupComplete("session") || !artifact.cleanupComplete("process") {
		t.Fatalf("incomplete lease artifact: %#v", artifact)
	}
}

func TestPerformanceFixtureCleanupRetainsFailureAndContinuesIndependentCleanup(t *testing.T) {
	deps := defaultTestDependencies(t)
	env := cloneEnv(deps.env)
	authorizeSuiteEnv(env)
	env[suiteservices.SuiteIDEnv] = "suite-performance-cleanup-failure"
	env["CARTULARY_FIXTURE_PROCESS_CLEANUP_COMPLETE"] = "1"
	profile := performanceFixtureTestProfile(t)
	runtimeRoot := filepath.Join(t.TempDir(), "private-runtime")
	bundle, err := fixture.GenerateRuntimeBundle(profile, performanceFixtureTestKey)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.WriteRuntimeBundle(profile, runtimeRoot, bundle); err != nil {
		t.Fatal(err)
	}
	metadata := webE2EMetadata{
		DatabaseName:      "ct_fedcba9876543210_web_e2e",
		Bucket:            "ct-fedcba9876543210-web-e2e",
		FixtureProfileID:  profile.FixtureProfileID,
		SnapshotKey:       performanceFixtureTestKey,
		BuilderUnitID:     "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
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
	var artifact leaseArtifactEvidence
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	if artifact.CleanupState != "failed" || artifact.FailureCode != "database_cleanup_failed" || !artifact.cleanupComplete("credential_copy") || !artifact.cleanupComplete("bucket") {
		t.Fatalf("cleanup failure artifact lost causal state: %#v", artifact)
	}
}

func performanceFixtureTestProfile(t *testing.T) performancefixtureprofile.Profile {
	t.Helper()
	for _, profile := range performancefixtureprofile.Profiles() {
		if profile.Status == "active" {
			return profile
		}
	}
	t.Fatal("generated active performance fixture profile is missing")
	return performancefixtureprofile.Profile{}
}

type leaseArtifactEvidence struct {
	CleanupResults []struct {
		ResourceClass string `json:"resource_class"`
		Outcome       string `json:"outcome"`
	} `json:"cleanup_results"`
	CleanupState string `json:"cleanup_state"`
	FailureCode  string `json:"failure_code"`
}

func (artifact leaseArtifactEvidence) cleanupComplete(resourceClass string) bool {
	for _, result := range artifact.CleanupResults {
		if result.ResourceClass == resourceClass {
			return result.Outcome == "complete"
		}
	}
	return false
}
