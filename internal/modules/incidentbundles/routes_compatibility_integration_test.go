package incidentbundles_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentBundleRetiredVersionIsRejectedWithoutEffects_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "incident-bundle-retired-version-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "incident-bundle-retired-version-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-retired-version-source",
		"incident_key":  "BUNDLE-RETIRED-VERSION",
		"title":         "Incident bundle retired version rejection",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-retired-version-row",
		"timeline.activity_synopsis_text": "Retired version must remain inert",
	})

	bundle := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-retired-version")
	retiredBundle := replaceZipMember(t, bundle, "manifest.json", func(original []byte) []byte {
		var manifest map[string]any
		if err := json.Unmarshal(original, &manifest); err != nil {
			t.Fatalf("decode current manifest: %v", err)
		}
		manifest["bundle_version"] = float64(2)
		payload, err := json.Marshal(manifest)
		if err != nil {
			t.Fatalf("encode retired-version manifest: %v", err)
		}
		return append(payload, '\n')
	})
	beforeDurability := snapshotEnvelopeDurability(t, targetHarness.DB)
	terminal := assertImportFailureLeavesState(
		t, targetHarness, targetAdmin, incidentID,
		"txn-import-retired-version", retiredBundle, "unsupported_bundle_version",
	)
	errorSummary := terminal["error_summary"].(map[string]any)
	if errorSummary["retryable"] != false {
		t.Fatalf("retired version failure must be non-retryable: %#v", errorSummary)
	}
	replay := postImport(
		t, targetHarness.Server, targetAdmin,
		`{"client_txn_id":"txn-import-retired-version"}`,
		retiredBundle, "retired-version-replay.zip",
	)
	replayedJob := httptestx.RequireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != terminal["job_id"] {
		t.Fatalf("retired version replay returned a different job: first=%v replay=%v", terminal["job_id"], replayedJob["job_id"])
	}
	afterDurability := snapshotEnvelopeDurability(t, targetHarness.DB)
	if afterDurability.Jobs != beforeDurability.Jobs+1 ||
		afterDurability.Payloads != beforeDurability.Payloads+1 ||
		afterDurability.Idempotency != beforeDurability.Idempotency+1 {
		t.Fatalf("retired version durable admission mismatch: before=%#v after=%#v", beforeDurability, afterDurability)
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

func TestLegacyBundleLayoutConversionAndHistoricalReplay_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	source := runtime.StartDefaultServer(t, "frozen-layout-legacy-source")
	target := startIsolatedIncidentBundleServer(t, runtime, "frozen-layout-legacy-target")
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, source.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, target.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, source.Server, admin, map[string]any{
		"client_txn_id": "frozen-legacy-source", "incident_key": "FROZEN-LEGACY", "title": "Legacy layout compatibility fixture",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, source.Server, admin, incidentID, map[string]any{
		"client_txn_id": "frozen-legacy-row", "timeline.activity_synopsis_text": "Legacy identifying data",
	})
	current, layoutErr := viewschema.DefaultLayout("cartulary.view.timeline.v2")
	if layoutErr != nil {
		t.Fatal(layoutErr)
	}
	var layout map[string]any
	if err := json.Unmarshal(current, &layout); err != nil {
		t.Fatal(err)
	}
	layout["column_widths"] = []map[string]any{{"field_key": "timeline.activity_synopsis_text", "width_px": 240}}
	layout["layout_schema_id"] = "cartulary.layout.v1"
	delete(layout, "frozen_through_field_key")
	legacyLayout, _ := json.Marshal(layout)
	savedID := uuid.New().String()
	if _, err := source.DB.Exec(`INSERT INTO saved_views (saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json, owner_user_id) VALUES ($1, $2, 'cartulary.view.timeline.v2', 'private', 'Legacy fixture', '{"filters":[],"sort":[]}'::jsonb, $3::jsonb, $4)`, savedID, incidentID, legacyLayout, adminID); err != nil {
		t.Fatal(err)
	}
	beforeStored := stringScalar(t, source.DB, `SELECT layout_json::text FROM saved_views WHERE saved_view_id=$1`, savedID)
	exported := exportBundleBytes(t, source, admin, incidentID, "frozen-historical-export")
	if stringScalar(t, source.DB, `SELECT layout_json::text FROM saved_views WHERE saved_view_id=$1`, savedID) != beforeStored {
		t.Fatal("export rewrote legacy stored layout")
	}
	rows := decodeNDJSONRows(t, zipMemberBytes(t, exported, "data/saved_views.ndjson"))
	if len(rows) != 1 {
		t.Fatalf("saved rows=%d", len(rows))
	}
	exportedLayout := rows[0]["layout_json"].(map[string]any)
	if exportedLayout["layout_schema_id"] != "cartulary.layout.v2" || exportedLayout["frozen_through_field_key"] != nil {
		t.Fatalf("exported layout=%#v", exportedLayout)
	}
	// Construct an original v3 fixture from a layout that has never had freezing.
	// This is test fixture construction, never a downgrade of authored freezing.
	exportedLayout["layout_schema_id"] = "cartulary.layout.v1"
	delete(exportedLayout, "frozen_through_field_key")
	legacy := replaceStructuredBundleMember(t, exported, "data/saved_views.ndjson", encodeNDJSONRows(t, rows))
	// A v4 archive carrying v1 rows must fail after integrity admission, atomically.
	assertImportFailureLeavesState(t, target, targetAdmin, incidentID, "frozen-version-mismatch", legacy, "source_family_invalid")
	legacy = replaceZipMember(t, legacy, "manifest.json", func(raw []byte) []byte {
		var manifest map[string]any
		if err := json.Unmarshal(raw, &manifest); err != nil {
			t.Fatal(err)
		}
		manifest["bundle_version"] = 3
		encoded, err := json.Marshal(manifest)
		if err != nil {
			t.Fatal(err)
		}
		return append(encoded, '\n')
	})
	originalDigest := hashHexBytes(legacy)
	// Integrity failures take precedence over invalid row/layout conversion.
	tampered := replaceZipMember(t, legacy, "data/saved_views.ndjson", func([]byte) []byte { return []byte("invalid layout bytes\n") })
	assertImportFailureLeavesState(t, target, targetAdmin, incidentID, "frozen-integrity-first", tampered, "checksum_mismatch")
	assertImportFailureLeavesState(t, target, targetAdmin, incidentID, "frozen-signature-first", appendZipMember(t, legacy, "integrity/signature.ed25519", []byte("invalid-signature")), "signature_mismatch")
	terminal := importBundleAndWait(t, target.Server, targetAdmin, legacy, "frozen-legacy-import")
	if terminal["result_summary"].(map[string]any)["code"] != incidentBundleImportedCode {
		t.Fatalf("legacy import=%#v", terminal)
	}
	if hashHexBytes(legacy) != originalDigest {
		t.Fatal("import changed original archive bytes")
	}
	expected, errLayout := viewschema.NormalizeLayout(legacyLayout, "cartulary.view.timeline.v2")
	if errLayout != nil {
		t.Fatal(errLayout)
	}
	stored := stringScalar(t, target.DB, `SELECT layout_json::text FROM saved_views WHERE saved_view_id=$1`, savedID)
	normalized, errLayout := viewschema.NormalizeLayout([]byte(stored), "cartulary.view.timeline.v2")
	if errLayout != nil || !bytes.Equal(expected, normalized) {
		t.Fatalf("prepared current layout mismatch: %s err=%v", stored, errLayout)
	}
	if !bytes.Contains([]byte(stored), []byte("cartulary.layout.v2")) {
		t.Fatal("import did not persist current layout")
	}
	beforeEffects := snapshotEnvelopeDurability(t, target.DB)
	replayedImport := importBundleAndWait(t, target.Server, targetAdmin, legacy, "frozen-legacy-import")
	if replayedImport["job_id"] != terminal["job_id"] || snapshotEnvelopeDurability(t, target.DB) != beforeEffects {
		t.Fatal("legacy import replay created durable work")
	}
	// Seed a completed historical artifact into this disposable fixture. Its
	// existing request/receipt identity remains exactly as admitted before upgrade.
	var storageRef string
	if err := source.DB.QueryRow(`SELECT bundle_storage_ref FROM incident_bundle_exports WHERE incident_id=$1`, incidentID).Scan(&storageRef); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(exportedBundleTestPath(t, source.Server, storageRef), legacy, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := source.DB.Exec(`UPDATE incident_bundle_exports SET bundle_sha256=$2, bundle_byte_size=$3, manifest_sha256=$4 WHERE incident_id=$1`, incidentID, originalDigest, len(legacy), hashHexBytes(zipMemberBytes(t, legacy, "manifest.json"))); err != nil {
		t.Fatal(err)
	}
	receiptSQL := `SELECT encode(request_hash,'hex') || response_json::text FROM route_idempotency WHERE route_key='incident_bundles.export' AND client_txn_id='frozen-historical-export'`
	receipt := stringScalar(t, source.DB, receiptSQL)
	jobs := countRows(t, source.DB, `SELECT count(*) FROM jobs`)
	replayed := exportBundleBytes(t, source, admin, incidentID, "frozen-historical-export")
	if !bytes.Equal(replayed, legacy) || stringScalar(t, source.DB, receiptSQL) != receipt || countRows(t, source.DB, `SELECT count(*) FROM jobs`) != jobs {
		t.Fatal("historical export replay rebuilt artifact, receipt, hash or job")
	}
}
