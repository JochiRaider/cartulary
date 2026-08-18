package incidentbundles_test

import (
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"testing"

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
		manifest["bundle_version"] = float64(1)
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
