package incidentbundles_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentBundleV1TranslationIsLossless_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "incident-bundle-v1-translation-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "incident-bundle-v1-translation-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-v1-translation-source",
		"incident_key":  "BUNDLE-V1-TRANSLATION",
		"title":         "Incident bundle v1 translation",
	})
	incidentID := incident["incident_id"].(string)
	row := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-v1-translation-row",
		"timeline.activity_synopsis_text": "Lossless legacy translation",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	sourceMetadata := map[string]any{
		"source_kind":         "clipboard",
		"paste_client_txn_id": "txn-legacy-source",
		"mapping_fingerprint": strings.Repeat("a", 64),
	}
	sourceMetadataJSON, err := json.Marshal(sourceMetadata)
	if err != nil {
		t.Fatalf("encode legacy provenance metadata: %v", err)
	}
	sourceIdentity := sha256.Sum256(sourceMetadataJSON)
	if _, err := sourceHarness.DB.Exec(`
INSERT INTO timeline_source_provenance (
    record_id, source_identity_hash, source_row_ordinal,
    source_column_ordinal, source_kind, source_metadata,
    source_header_json, raw_value, cell_kind, created_at
) VALUES ($1, $2, 7, 3, 'clipboard', $3::jsonb, '["Legacy Header"]'::jsonb, 'legacy raw value', 'text', now())
`, recordID, sourceIdentity[:], sourceMetadataJSON); err != nil {
		t.Fatalf("seed legacy provenance: %v", err)
	}

	v2Bundle := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-v1-translation")
	v1Bundle := convertV2TimelineBundleToV1(t, v2Bundle)
	var convertedManifest incidentBundleManifestMirror
	if err := json.Unmarshal(zipMemberMap(t, v1Bundle)["manifest.json"], &convertedManifest); err != nil {
		t.Fatalf("decode converted v1 manifest: %v", err)
	}
	if convertedManifest.BundleVersion != incidentBundleLegacyVersion {
		t.Fatalf("converted bundle version = %d; want %d", convertedManifest.BundleVersion, incidentBundleLegacyVersion)
	}
	terminal := importBundleAndWait(t, targetHarness.Server, targetAdmin, v1Bundle, "txn-import-v1-translation")
	if terminal["status"] != "succeeded" {
		t.Fatalf("v1 import terminal state = %#v", terminal)
	}
	var synopsis string
	if err := targetHarness.DB.QueryRow(`
SELECT activity_synopsis_text
  FROM timeline_events
 WHERE record_id = $1
`, recordID).Scan(&synopsis); err != nil {
		t.Fatalf("query translated timeline row: %v", err)
	}
	if synopsis != "Lossless legacy translation" {
		t.Fatalf("translated timeline synopsis = %q", synopsis)
	}
	var identityHex string
	var rowOrdinal int
	var columnOrdinal int
	var sourceKind string
	var metadataJSON string
	var headerJSON string
	var rawValue string
	var cellKind string
	if err := targetHarness.DB.QueryRow(`
SELECT encode(source_identity_hash, 'hex'), source_row_ordinal,
       source_column_ordinal, source_kind, source_metadata::text,
       source_header_json::text, raw_value, cell_kind
  FROM timeline_source_provenance
 WHERE record_id = $1
`, recordID).Scan(
		&identityHex, &rowOrdinal, &columnOrdinal, &sourceKind,
		&metadataJSON, &headerJSON, &rawValue, &cellKind,
	); err != nil {
		t.Fatalf("query translated Timeline provenance: %v", err)
	}
	if identityHex != hex.EncodeToString(sourceIdentity[:]) ||
		rowOrdinal != 7 || columnOrdinal != 3 || sourceKind != "clipboard" ||
		rawValue != "legacy raw value" || cellKind != "text" {
		t.Fatalf("translated Timeline provenance changed: %s/%d/%d/%s/%s/%s", identityHex, rowOrdinal, columnOrdinal, sourceKind, rawValue, cellKind)
	}
	var gotMetadata any
	var wantMetadata any
	if err := json.Unmarshal([]byte(metadataJSON), &gotMetadata); err != nil {
		t.Fatalf("decode imported provenance metadata: %v", err)
	}
	if err := json.Unmarshal(sourceMetadataJSON, &wantMetadata); err != nil {
		t.Fatalf("decode source provenance metadata: %v", err)
	}
	if !reflect.DeepEqual(gotMetadata, wantMetadata) || headerJSON != `["Legacy Header"]` {
		t.Fatalf("translated Timeline provenance metadata changed: metadata=%s header=%s", metadataJSON, headerJSON)
	}
}

func TestDescriptorPaginationAndCanonicalManifest_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-descriptor-canonical")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-descriptor-canonical",
		"incident_key":  "BUNDLE-DESCRIPTOR",
		"title":         "Incident bundle descriptor",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-descriptor-row",
		"timeline.activity_synopsis_text": "Canonical descriptor event",
	})
	job := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-descriptor-canonical",
		"optional_sections":     []string{"snapshots", "reference_packs", "snapshots"},
		"required_capabilities": []string{},
		"reference_pack_mode":   "embedded",
	}), http.StatusAccepted)["data"].(map[string]any)
	terminal := waitJob(t, harness.Server, admin, job["job_id"].(string))
	ref := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	descriptorRoute := ref["route"].(string)
	descriptorResp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+descriptorRoute, nil, httptestx.WithCookies(admin.SessionCookie))
	descriptor := httptestx.RequireSuccessEnvelope(t, descriptorResp, http.StatusOK)["data"].(map[string]any)
	wantTokens := []string{"reference_packs", "snapshots"}
	if got := stringArray(t, descriptor["optional_sections"]); !slices.Equal(got, wantTokens) {
		t.Fatalf("descriptor optional_sections not canonical: got %#v want %#v", got, wantTokens)
	}
	if got := stringArray(t, descriptor["required_capabilities"]); len(got) != 0 {
		t.Fatalf("descriptor required_capabilities must be empty until capabilities are implemented: got %#v", got)
	}
	if descriptor["reference_pack_mode"] != "embedded" || descriptor["history_mode"] != "full" || descriptor["blob_mode"] != "full" {
		t.Fatalf("descriptor modes mismatch: %#v", descriptor)
	}
	for _, suffix := range []string{"?limit=1", "?cursor_token=abc"} {
		rejected := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+descriptorRoute+suffix, nil, httptestx.WithCookies(admin.SessionCookie))
		body := httptestx.RequireErrorEnvelope(t, rejected, http.StatusBadRequest, "invalid_pagination_request")
		if details := httptestx.RequireErrorDetails(t, body); details["reason_code"] != "pagination_not_supported" {
			t.Fatalf("descriptor pagination reason mismatch for %s: %#v", suffix, details)
		}
	}

	bundleStorageRef := stringScalar(t, harness.DB, `SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id = $1`, ref["id"].(string))
	bundleBytes, err := os.ReadFile(exportedBundleTestPath(t, harness.Server, bundleStorageRef))
	if err != nil {
		t.Fatalf("read descriptor bundle: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(zipMemberBytes(t, bundleBytes, "manifest.json"), &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if got := stringArray(t, manifest["optional_sections"]); !slices.Equal(got, wantTokens) {
		t.Fatalf("manifest optional_sections not canonical: got %#v want %#v", got, wantTokens)
	}
	if got := stringArray(t, manifest["required_capabilities"]); len(got) != 0 {
		t.Fatalf("manifest required_capabilities must be empty until capabilities are implemented: got %#v", got)
	}
	if manifest["reference_pack_mode"] != "embedded" || manifest["history_mode"] != "full" || manifest["blob_mode"] != "full" {
		t.Fatalf("manifest modes mismatch: %#v", manifest)
	}
}
