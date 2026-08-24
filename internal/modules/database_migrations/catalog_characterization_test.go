package database_migrations_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

type migrationHistoryManifest struct {
	SchemaID                string                          `json:"schema_id"`
	MigrationRoot           string                          `json:"migration_root"`
	ImmutableThroughVersion int64                           `json:"immutable_through_version"`
	Entries                 []migrationHistoryManifestEntry `json:"entries"`
}

type migrationHistoryManifestEntry struct {
	Version  int64  `json:"version"`
	Filename string `json:"filename"`
	SHA256   string `json:"sha256"`
}

const (
	expectedCanonicalLineageID       = "cartulary.prod_ddl_rebaseline.v2"
	expectedCanonicalLineageBoundary = "prod_ddl_rebaseline_v2"
)

func TestCanonicalEmbeddedMigrationCatalogCharacterization(t *testing.T) {
	manifest := loadMigrationHistoryManifest(t)
	if manifest.SchemaID != "cartulary.migration_history_manifest.v1" || manifest.MigrationRoot != "db/migrations" {
		t.Fatalf("unexpected migration manifest identity: schema=%q root=%q", manifest.SchemaID, manifest.MigrationRoot)
	}
	if manifest.ImmutableThroughVersion != 29 {
		t.Fatalf("unexpected immutable boundary: %d", manifest.ImmutableThroughVersion)
	}
	if len(manifest.Entries) < int(manifest.ImmutableThroughVersion) {
		t.Fatalf("migration manifest entry count = %d, below immutable boundary %d", len(manifest.Entries), manifest.ImmutableThroughVersion)
	}

	first, err := dbmigrations.Source()
	if err != nil {
		t.Fatalf("build first canonical source: %v", err)
	}
	inspection, err := database_migrations.InspectSource(first)
	if err != nil {
		t.Fatalf("inspect canonical source: %v", err)
	}
	if len(inspection.Entries) != len(manifest.Entries) {
		t.Fatalf("embedded migration count = %d, want %d", len(inspection.Entries), len(manifest.Entries))
	}
	for index, manifestEntry := range manifest.Entries {
		wantVersion := int64(index + 1)
		if manifestEntry.Version != wantVersion {
			t.Fatalf("manifest entry %d version = %d, want %d", index, manifestEntry.Version, wantVersion)
		}
		if inspection.Entries[index].Filename != manifestEntry.Filename {
			t.Fatalf("embedded entry %d = %q, want %q", index, inspection.Entries[index].Filename, manifestEntry.Filename)
		}
		if got := inspection.Entries[index].SHA256; got != manifestEntry.SHA256 {
			t.Fatalf("embedded migration %q SHA-256 = %s, want %s", manifestEntry.Filename, got, manifestEntry.SHA256)
		}
	}

	second, err := dbmigrations.Source()
	if err != nil {
		t.Fatalf("build second canonical source: %v", err)
	}
	if first != second {
		t.Fatal("repeated canonical source construction did not return the same pointer")
	}
	wantMaxVersion := manifest.Entries[len(manifest.Entries)-1].Version
	if inspection.VersionCount != len(manifest.Entries) || inspection.MinVersion != manifest.Entries[0].Version ||
		inspection.MaxVersion != wantMaxVersion || len(inspection.Entries) != len(manifest.Entries) {
		t.Fatalf("unexpected canonical source ordering facts: %#v", inspection)
	}
	if !inspection.Entries[0].HasGooseUp || !inspection.Entries[0].HasGooseDown {
		t.Fatalf("canonical source marker facts are incomplete: %#v", inspection.Entries[0])
	}
	inspection.Entries[0].Filename = "mutated.sql"
	again, err := database_migrations.InspectSource(first)
	if err != nil {
		t.Fatalf("inspect canonical source after copy mutation: %v", err)
	}
	if again.Entries[0].Filename != manifest.Entries[0].Filename {
		t.Fatalf("inspection copy mutation reached canonical source: %#v", again.Entries[0])
	}

	const runnerIdentity = "cartulary-postgres-migrate/goose/v3.27.0"
	hash, err := database_migrations.SchemaHash(first, runnerIdentity)
	if err != nil {
		t.Fatalf("hash canonical source: %v", err)
	}
	const wantHash = "ba0bffb76a7193e9616f4fdeee9e16086da7f3c0fe67f5dc26e47ff5f1b750c5"
	if hash != wantHash {
		t.Fatalf("canonical source hash = %s, want %s", hash, wantHash)
	}
	if _, err := database_migrations.SchemaHash(nil, runnerIdentity); err == nil {
		t.Fatal("nil source unexpectedly produced a schema hash")
	}
	if _, err := database_migrations.InspectSource(nil); err == nil {
		t.Fatal("nil source unexpectedly produced an inspection")
	}
	if _, err := database_migrations.InspectSource(&database_migrations.Source{}); err == nil {
		t.Fatal("zero source unexpectedly produced an inspection")
	}
	if _, err := database_migrations.SchemaHash(first, " "); err == nil {
		t.Fatal("empty runner identity unexpectedly produced a schema hash")
	}
}

func loadMigrationHistoryManifest(t testing.TB) migrationHistoryManifest {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve source test path")
	}
	manifestBytes, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "..", "..", "..", "tools", "migration_history_manifest.json"))
	if err != nil {
		t.Fatalf("read migration history manifest: %v", err)
	}
	var manifest migrationHistoryManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode migration history manifest: %v", err)
	}
	return manifest
}

func canonicalRepositoryHead(t testing.TB) int64 {
	t.Helper()
	manifest := loadMigrationHistoryManifest(t)
	if len(manifest.Entries) == 0 {
		t.Fatal("canonical migration manifest is empty")
	}
	return manifest.Entries[len(manifest.Entries)-1].Version
}
