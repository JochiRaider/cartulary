package migrationevidence

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

func TestMigrationEvidenceSourceAuditReportsReachableFindings(t *testing.T) {
	validMigration := []byte("-- +goose Up\nSELECT 1;\n-- +goose Down\nSELECT 1;\n")
	inspection := database_migrations.SourceInspection{Entries: []database_migrations.SourceInspectionEntry{
		{Version: 1, Filename: "00001_valid.sql", SHA256: sha256Hex(validMigration), HasGooseUp: true, HasGooseDown: true},
		{Version: 2, Filename: "00002_workbook_phase9_source.sql", SHA256: sha256Hex(validMigration), HasGooseUp: true, HasGooseDown: true},
		{Version: 3, Filename: "00003_source_only.sql", SHA256: sha256Hex(validMigration), HasGooseUp: true, HasGooseDown: true},
	}}
	manifestPath := writeMigrationEvidenceManifest(t, manifestDocument{
		SchemaID:                "cartulary.migration_history_manifest.v1",
		MigrationRoot:           "db/migrations",
		ImmutableThroughVersion: 1,
		Entries: []manifestEntry{
			{Version: 1, Filename: "00001_valid.sql", SHA256: strings.Repeat("f", 64)},
			{Version: 2, Filename: "00002_manifest_name.sql", SHA256: sha256Hex(validMigration)},
			{Version: 4, Filename: "00004_manifest_only.sql", SHA256: strings.Repeat("0", 64)},
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
	assertMigrationEvidenceFinding(t, findings, "manifest_filename_mismatch")
	assertMigrationEvidenceFinding(t, findings, "future_phase_shaped_filename")
	assertMigrationEvidenceFinding(t, findings, "manifest_version_not_in_source")
	assertMigrationEvidenceFinding(t, findings, "source_version_not_in_manifest")
	for _, impossible := range []string{
		"source_filename_invalid",
		"source_duplicate_version",
		"source_marker_missing",
		"source_version_gap",
	} {
		assertMigrationEvidenceFindingAbsent(t, findings, impossible)
	}
}

func TestMigrationEvidenceInputDefaultsAndBindingNormalization(t *testing.T) {
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
	invalidPath := t.TempDir() + "/invalid-private-manifest.json"
	if err := os.WriteFile(invalidPath, []byte("{}"), 0o600); err != nil {
		t.Fatalf("write invalid manifest: %v", err)
	}
	tests := []struct {
		name   string
		path   string
		reason manifestFailureReason
	}{
		{name: "required", path: "  ", reason: manifestFailurePathRequired},
		{name: "unavailable", path: t.TempDir() + "/private-manifest-location.json", reason: manifestFailureUnavailable},
		{name: "invalid", path: invalidPath, reason: manifestFailureInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, _, err := loadManifest(test.path)
			if err == nil {
				t.Fatal("manifest unexpectedly loaded")
			}
			var failure manifestFailureError
			if !errors.As(err, &failure) || failure.reason != test.reason {
				t.Fatalf("unexpected typed manifest failure: %T %v", err, err)
			}
			if err.Error() != string(test.reason) {
				t.Fatalf("manifest failure = %q, want %q", err.Error(), test.reason)
			}
			if strings.Contains(err.Error(), test.path) || strings.Contains(err.Error(), "private-manifest") {
				t.Fatalf("manifest failure disclosed locator: %v", err)
			}
		})
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

func assertMigrationEvidenceFindingAbsent(t *testing.T, findings []Finding, reasonCode string) {
	t.Helper()
	for _, finding := range findings {
		if finding.ReasonCode == reasonCode {
			t.Fatalf("finding %q unexpectedly present in %#v", reasonCode, findings)
		}
	}
}
