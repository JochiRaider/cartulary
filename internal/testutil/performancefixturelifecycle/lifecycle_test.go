package performancefixturelifecycle

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestOwnedNamesAreDeterministicAndBounded(t *testing.T) {
	t.Parallel()
	suiteID := "suite-performance-fixture"
	name := templateName(suiteID, performanceFixtureTestKey)
	if !safePostgresIdentifier(name) || len(name) > 63 {
		t.Fatalf("unsafe template name %q", name)
	}
	if !strings.HasPrefix(name, "ct_pfs_") || !strings.HasSuffix(name, performanceFixtureTestKey[:12]) {
		t.Fatalf("template name is not bound to suite and snapshot: %q", name)
	}
	clone, err := newCloneName()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(clone, "ct_") || !strings.HasSuffix(clone, "_web_e2e") || !safePostgresIdentifier(clone) {
		t.Fatalf("profile clone does not satisfy the bounded lifecycle grammar: %q", clone)
	}
}

func TestArtifactsAreImmutableAndRedacted(t *testing.T) {
	t.Run("admission", testSnapshotBuildAdmission)
	t.Parallel()
	file := filepath.Join(t.TempDir(), "artifact.json")
	value := map[string]any{"schema_id": "example", "snapshot_key": performanceFixtureTestKey}
	if err := WriteImmutableJSON(file, value); err != nil {
		t.Fatal(err)
	}
	if err := WriteImmutableJSON(file, value); err == nil {
		t.Fatal("immutable artifact accepted a second write")
	}
	info, err := os.Lstat(file)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("immutable artifact mode is %v", info.Mode())
	}
	profile := lifecycleTestProfile(t)
	failed := FailedBuild(profile, BuildArgs{
		FixtureProfileID:     profile.FixtureProfileID,
		SnapshotKey:          performanceFixtureTestKey,
		MigrationDigest:      strings.Repeat("a", 64),
		SourceContractDigest: profile.SourceContractDigest,
		BuilderUnitID:        "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
	}, "contribution_invalid")
	raw, err := json.Marshal(failed)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"database_name", "bucket_name", "password", "email", "runtime_path", "user_id"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("failed build artifact contains forbidden field %q", forbidden)
		}
	}
	diagnostics := SuccessfulBuildDiagnostics(profile, BuildArgs{
		FixtureProfileID: profile.FixtureProfileID,
		SnapshotKey:      performanceFixtureTestKey,
		BuilderUnitID:    "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
	}, BuildArtifact{
		ContributionDiagnostics: []BuildContributionDiagnostic{{
			Ordinal: 1, ContributionID: "timeline.large_grid.v1", OwnerID: "module.timeline",
			State: "completed", DurationMS: 5, Batch: &BuildBatchDiagnostic{Strategy: "owner_set_oriented", BatchCount: 1, ConfiguredBatchSize: 20000, ItemCount: 20000},
		}},
		SemanticValidationTime: 3 * time.Millisecond,
	}, 10*time.Millisecond)
	diagnosticRaw, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diagnosticRaw), "snapshot-build.json") || diagnostics.SchemaID != "cartulary.performance_fixture_build_diagnostics.v2" || diagnostics.FailureStage != "none" || diagnostics.ConstructionCount != 1 || diagnostics.DurationMS != 10 || diagnostics.SemanticValidationDurationMS != 3 {
		t.Fatalf("invalid build diagnostics: %s", diagnosticRaw)
	}
	semanticRaw, err := json.Marshal(BuildArtifact{
		ContributionDiagnostics: diagnostics.Contributions,
		SemanticValidationTime:  3 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(semanticRaw), "duration_ms") || strings.Contains(string(semanticRaw), "contributions") {
		t.Fatalf("diagnostics entered semantic artifact JSON: %s", semanticRaw)
	}
	failedDiagnostics := FailedBuildDiagnostics(profile, BuildArgs{
		FixtureProfileID: profile.FixtureProfileID,
		SnapshotKey:      performanceFixtureTestKey,
		BuilderUnitID:    "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey,
	}, 12*time.Millisecond, withBuildFailureStage("finalization", errors.New("sensitive database path")))
	failedRaw, err := json.Marshal(failedDiagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if failedDiagnostics.FailureStage != "finalization" || failedDiagnostics.SemanticValidationDurationMS != 0 || len(failedDiagnostics.Contributions) != 0 {
		t.Fatalf("invalid failed diagnostics: %s", failedRaw)
	}
	if strings.Contains(string(failedRaw), "sensitive") || strings.Contains(string(failedRaw), "database path") {
		t.Fatalf("failed diagnostics leaked error text: %s", failedRaw)
	}
}

func testSnapshotBuildAdmission(t *testing.T) {
	profile := lifecycleTestProfile(t)
	env := map[string]string{"CARTULARY_TEST_RESULTS_DIR": t.TempDir(), "CARTULARY_TEST_RUN_ID": "build-admission", "CARTULARY_FIXTURE_SNAPSHOT_KEY": performanceFixtureTestKey, "CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID": "fixture_snapshot:default:" + profile.FixtureProfileID + ":" + performanceFixtureTestKey}
	build := FailedBuild(profile, BuildArgs{SnapshotKey: performanceFixtureTestKey, MigrationDigest: strings.Repeat("a", 64), SourceContractDigest: profile.SourceContractDigest, BuilderUnitID: env["CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID"]}, "")
	build.State = "sealed"
	build.Validation.ConnectionsClosed = true
	build.ObjectSnapshot = &ObjectSnapshotSummary{ManifestSHA256: "sha256:" + strings.Repeat("b", 64)}
	for _, c := range profile.Contributions {
		receipt := BuildReceipt{ContributionID: c.ContributionID, Version: c.Version, OwnerID: c.OwnerID}
		for _, count := range c.ExpectedReceiptCounts {
			receipt.Counts = append(receipt.Counts, ReceiptCount{CountID: count.CountID, Actual: count.Exact})
		}
		build.ContributionReceipts = append(build.ContributionReceipts, receipt)
	}
	for _, c := range profile.SemanticExpectations.Counts {
		build.Validation.Counts = append(build.Validation.Counts, ValidationCount{ExpectationID: c.ExpectationID, Actual: c.Exact, Passed: true})
	}
	for _, c := range profile.SemanticExpectations.Conditions {
		build.Validation.Conditions = append(build.Validation.Conditions, ValidationCondition{ExpectationID: c.ExpectationID, Actual: c.Required, Passed: true})
	}
	file, err := BuildArtifactPath(env, BuildArgs{SnapshotKey: performanceFixtureTestKey})
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteImmutableJSON(file, build); err != nil {
		t.Fatal(err)
	}
	if _, err := readBuildArtifact(env, profile); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*BuildArtifact){
		"retired":              func(b *BuildArtifact) { b.SchemaID = "cartulary.performance_fixture_snapshot.v2" },
		"missing object proof": func(b *BuildArtifact) { b.ObjectSnapshot = nil },
		"wrong object population": func(b *BuildArtifact) {
			b.ObjectSnapshot = &ObjectSnapshotSummary{ManifestSHA256: build.ObjectSnapshot.ManifestSHA256, ObjectCount: 1, TotalBytes: 13}
		},
		"failed build":         func(b *BuildArtifact) { b.State = "failed" },
		"wrong builder":        func(b *BuildArtifact) { b.BuilderUnitID += "wrong" },
		"missing contribution": func(b *BuildArtifact) { b.ContributionReceipts = nil },
		"missing validation":   func(b *BuildArtifact) { b.Validation.Counts = nil },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := build
			mutate(&candidate)
			raw, _ := json.Marshal(candidate)
			if err := os.WriteFile(file, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readBuildArtifact(env, profile); err == nil {
				t.Fatal("invalid build admitted")
			}
		})
	}
	raw, _ := json.Marshal(build)
	for name, invalid := range map[string][]byte{
		"duplicate": append([]byte(`{"schema_id":"ignored",`), raw[1:]...),
		"unknown":   append([]byte(`{"unexpected":true,`), raw[1:]...),
		"trailing":  append(append([]byte{}, raw...), []byte(` {}`)...),
	} {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(file, invalid, 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readBuildArtifact(env, profile); err == nil {
				t.Fatal("ambiguous build admitted")
			}
		})
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(t.TempDir(), "build.json")
	if err := os.WriteFile(external, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(external, file); err != nil {
		t.Fatal(err)
	}
	if _, err := readBuildArtifact(env, profile); err == nil {
		t.Fatal("symlink build admitted")
	}
}
