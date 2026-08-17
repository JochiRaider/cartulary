package performancefixturelifecycle

import (
	"encoding/json"
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
			DurationMS: 5, Batch: &BuildBatchDiagnostic{Strategy: "owner_set_oriented", BatchCount: 1, ConfiguredBatchSize: 20000, ItemCount: 20000},
		}},
		SemanticValidationTime: 3 * time.Millisecond,
	}, 10*time.Millisecond)
	diagnosticRaw, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(diagnosticRaw), "snapshot-build.json") || diagnostics.DurationMS != 10 || diagnostics.SemanticValidationDurationMS != 3 {
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
}
