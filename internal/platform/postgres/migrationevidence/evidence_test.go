package migrationevidence

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMigrationEvidenceSourceAuditReportsManifestAndSourceFindings(t *testing.T) {
	validMigration := []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")
	missingDownMigration := []byte("-- +goose Up\nSELECT 2;\n")
	gapMigration := []byte("-- +goose Up\nSELECT 4;\n-- +goose Down\nSELECT 4;\n")
	sourceFS := fstest.MapFS{
		"00001_valid.sql": &fstest.MapFile{Data: validMigration},
		"00002_workbook_interaction9_missing_down.sql": &fstest.MapFile{Data: missingDownMigration},
		"00004_gap.sql": &fstest.MapFile{Data: gapMigration},
	}
	manifestPath := writeMigrationEvidenceManifest(t, manifestDocument{
		SchemaID:                "cartulary.migration_history_manifest.v1",
		MigrationRoot:           "db/migrations",
		ImmutableThroughVersion: 1,
		Entries: []manifestEntry{
			{Version: 1, Filename: "00001_valid.sql", SHA256: "not-the-source-hash"},
			{Version: 2, Filename: "00002_workbook_interaction9_missing_down.sql", SHA256: sha256Hex(missingDownMigration)},
			{Version: 3, Filename: "00003_missing.sql", SHA256: strings.Repeat("0", 64)},
		},
	})

	manifest, summary, manifestFindings, err := loadManifest(manifestPath)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if summary.SHA256 == "" || summary.ExpectedVersionCount != 3 {
		t.Fatalf("unexpected manifest summary: %#v", summary)
	}
	manifestByVersion := map[int64]manifestEntry{}
	for _, entry := range manifest.Entries {
		manifestByVersion[entry.Version] = entry
	}
	audit, sourceFindings, err := auditSource(sourceFS, manifest, manifestByVersion)
	if err != nil {
		t.Fatalf("audit source: %v", err)
	}
	if len(audit) != 3 {
		t.Fatalf("expected every embedded source file to be audited, got %d", len(audit))
	}
	findings := append(manifestFindings, sourceFindings...)
	assertMigrationEvidenceFinding(t, findings, "manifest_hash_mismatch")
	assertMigrationEvidenceFinding(t, findings, "source_marker_missing")
	assertMigrationEvidenceFinding(t, findings, "future_phase_shaped_filename")
	assertMigrationEvidenceFinding(t, findings, "manifest_version_not_in_source")
	assertMigrationEvidenceFinding(t, findings, "source_version_not_in_manifest")
	assertMigrationEvidenceFinding(t, findings, "source_version_gap")
}

func writeMigrationEvidenceManifest(t *testing.T, manifest manifestDocument) string {
	t.Helper()
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	path := t.TempDir() + "/migration_history_manifest.json"
	if err := os.WriteFile(path, body, 0o600); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	return path
}

func assertMigrationEvidenceFinding(t *testing.T, findings []Finding, reasonCode string) {
	t.Helper()
	for _, finding := range findings {
		if finding.ReasonCode == reasonCode {
			return
		}
	}
	t.Fatalf("finding %q not present in %#v", reasonCode, findings)
}
