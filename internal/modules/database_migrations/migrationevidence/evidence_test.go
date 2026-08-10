package migrationevidence

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

func TestMigrationEvidenceSourceAuditReportsManifestAndSourceFindings(t *testing.T) {
	validMigration := []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")
	missingDownMigration := []byte("-- +goose Up\nSELECT 2;\n")
	gapMigration := []byte("-- +goose Up\nSELECT 4;\n-- +goose Down\nSELECT 4;\n")
	inspection := database_migrations.SourceInspection{Entries: []database_migrations.SourceInspectionEntry{
		{Version: 1, Filename: "00001_valid.sql", SHA256: sha256Hex(validMigration), HasGooseUp: true, HasGooseDown: true},
		{Version: 2, Filename: "00002_workbook_phase9_missing_down.sql", SHA256: sha256Hex(missingDownMigration), HasGooseUp: true, HasGooseDown: false},
		{Version: 4, Filename: "00004_gap.sql", SHA256: sha256Hex(gapMigration), HasGooseUp: true, HasGooseDown: true},
	}}
	manifestPath := writeMigrationEvidenceManifest(t, manifestDocument{
		SchemaID:                "cartulary.migration_history_manifest.v1",
		MigrationRoot:           "db/migrations",
		ImmutableThroughVersion: 1,
		Entries: []manifestEntry{
			{Version: 1, Filename: "00001_valid.sql", SHA256: strings.Repeat("f", 64)},
			{Version: 2, Filename: "00002_workbook_phase9_missing_down.sql", SHA256: sha256Hex(missingDownMigration)},
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
	audit, sourceFindings := auditSource(inspection, manifest, manifestByVersion)
	if len(audit) != 3 {
		t.Fatalf("expected every embedded source file to be audited, got %d", len(audit))
	}
	for index := 1; index < len(audit); index++ {
		if audit[index-1].Version >= audit[index].Version {
			t.Fatalf("source audit order is not deterministic: %#v", audit)
		}
	}
	findings := append(manifestFindings, sourceFindings...)
	assertMigrationEvidenceFinding(t, findings, "manifest_hash_mismatch")
	assertMigrationEvidenceFinding(t, findings, "source_marker_missing")
	assertMigrationEvidenceFinding(t, findings, "future_phase_shaped_filename")
	assertMigrationEvidenceFinding(t, findings, "manifest_version_not_in_source")
	assertMigrationEvidenceFinding(t, findings, "source_version_not_in_manifest")
	assertMigrationEvidenceFinding(t, findings, "source_version_gap")
}

func TestMigrationEvidenceInputDefaultsAndBindingNormalization(t *testing.T) {
	if _, _, _, err := loadManifest("  "); err == nil || err.Error() != "migration evidence manifest path is required" {
		t.Fatalf("unexpected empty manifest error: %v", err)
	}

	got := normalizeDatabaseBinding(DatabaseBinding{
		BindingKind: "  managed_service\t",
		ServiceRef:  " postgres-primary \n",
	})
	if got.BindingKind != "managed_service" || got.ServiceRef != "postgres-primary" {
		t.Fatalf("binding was not normalized: %#v", got)
	}
	if safeDatabaseBinding(DatabaseBinding{BindingKind: "managed_service", ServiceRef: "/srv/private/postgres"}) {
		t.Fatal("path-shaped service reference was admitted")
	}
}

func TestMigrationEvidenceManifestFailuresDoNotDiscloseLocators(t *testing.T) {
	secretPath := t.TempDir() + "/private-manifest-location.json"
	_, _, _, err := loadManifest(secretPath)
	if err == nil {
		t.Fatal("missing manifest unexpectedly loaded")
	}
	var failure ManifestFailure
	if !errors.As(err, &failure) || failure.ReasonCode() != "migration evidence manifest unavailable" {
		t.Fatalf("unexpected typed manifest failure: %T %v", err, err)
	}
	if strings.Contains(err.Error(), secretPath) || strings.Contains(err.Error(), "private-manifest-location") {
		t.Fatalf("manifest failure disclosed locator: %v", err)
	}
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
