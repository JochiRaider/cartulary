package testsupport

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestTargetFixtureCopiesCallerEnvironment_Unit(t *testing.T) {
	env := map[string]string{"DATABASE_URL": "borrowed", "SECRET": "ephemeral"}
	fixture := NewTargetFixture(env, nil, nil)

	env["DATABASE_URL"] = "mutated"
	delete(env, "SECRET")
	if maps.Equal(fixture.Env, env) || fixture.Env["DATABASE_URL"] != "borrowed" || fixture.Env["SECRET"] != "ephemeral" {
		t.Fatalf("target fixture did not retain an independent environment copy: %#v", fixture.Env)
	}
	fixture.Cleanup()
	fixture.Cleanup()
	if len(fixture.Env) != 0 {
		t.Fatalf("idempotent cleanup retained copied environment values: %#v", fixture.Env)
	}
}

func TestEvidenceArtifactUsesPrivatePermissions_Unit(t *testing.T) {
	location := EvidenceLocation{
		ResultsRoot: t.TempDir(),
		RunID:       "run",
		Target:      "backend-process",
		Group:       "backup-restore",
	}
	path := WriteEvidenceArtifact(t, location, "proof.json", []byte("safe"))

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat evidence directory: %v", err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat evidence file: %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("evidence directory mode got %#o want 0700", got)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("evidence file mode got %#o want 0600", got)
	}
}

func TestEvidenceLocationRejectsNestedExecutionSegments_Unit(t *testing.T) {
	location := EvidenceLocation{
		ResultsRoot: t.TempDir(),
		RunID:       "nested/run",
		Target:      "backend-process",
		Group:       "backup-restore",
	}
	if _, err := location.Dir(); err == nil {
		t.Fatal("nested run ID was accepted")
	}
}

func TestCaptureParamsPreservesExplicitAnchorsAndDefaultsMissingAnchors_Unit(t *testing.T) {
	retainedUntil := time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC)
	explicit := retainedUntil.Add(time.Hour)
	params := CaptureParams(recovery.CaptureBackupSetParams{
		RetainedUntil:                         retainedUntil,
		PostgresRestoreAnchorRetainedUntil:    explicit,
		ObjectStoreRestoreAnchorRetainedUntil: time.Time{},
	})
	if !params.PostgresRestoreAnchorRetainedUntil.Equal(explicit) {
		t.Fatalf("explicit Postgres anchor changed: %s", params.PostgresRestoreAnchorRetainedUntil)
	}
	if !params.ObjectStoreRestoreAnchorRetainedUntil.Equal(retainedUntil) {
		t.Fatalf("object-store anchor got %s want %s", params.ObjectStoreRestoreAnchorRetainedUntil, retainedUntil)
	}
}

func TestSnapshotDigestIsStableAcrossTableAndRowOrder_Unit(t *testing.T) {
	rowA := json.RawMessage(`{"id":"a","count":1}`)
	rowB := json.RawMessage(`{"count":2,"id":"b"}`)
	first := recovery.PostgresSnapshotArtifact{Tables: []recovery.PostgresSnapshotTable{
		{TableName: "records", Rows: []json.RawMessage{rowB, rowA}},
		{TableName: "change_sets", Rows: []json.RawMessage{rowA}},
	}}
	second := recovery.PostgresSnapshotArtifact{Tables: []recovery.PostgresSnapshotTable{
		{TableName: "change_sets", Rows: []json.RawMessage{rowA}},
		{TableName: "records", Rows: []json.RawMessage{rowA, rowB}},
	}}

	firstDigest, firstCount, err := SnapshotDigest(first, nil)
	if err != nil {
		t.Fatalf("digest first snapshot: %v", err)
	}
	secondDigest, secondCount, err := SnapshotDigest(second, nil)
	if err != nil {
		t.Fatalf("digest second snapshot: %v", err)
	}
	if firstDigest != secondDigest || firstCount != 3 || secondCount != 3 {
		t.Fatalf("canonical digest mismatch: first=%s/%d second=%s/%d", firstDigest, firstCount, secondDigest, secondCount)
	}
}

func TestCaptureInputRequiresExplicitIdentityAndBorrowedDependencies_Unit(t *testing.T) {
	input := CaptureInput{
		Prefix:                  "fixture",
		OlderBackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		OlderConsistencyPointAt: time.Now().UTC().Add(-time.Hour),
		BackupSetID:             uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		ConsistencyPointAt:      time.Now().UTC().Add(-time.Minute),
	}
	if err := input.Validate(); err == nil {
		t.Fatal("capture input without borrowed dependencies was accepted")
	}
}
