package reference_data_test

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase11_I_11_REFERENCE_PACK_01_ImportListReadReplayAndJobSummary(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-import")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	bundle := referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	})
	metadata := `{"client_txn_id":"txn-reference-pack-import"}`
	first := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, bundle, "host-pack.zip", reference_data.MediaTypeZip)
	firstJob := requireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	jobID := firstJob["job_id"].(string)
	job := requireJob(t, harness, adminLogin, jobID)
	if job["status"] != "succeeded" {
		t.Fatalf("import job status = %#v", job)
	}
	summary := job["result_summary"].(map[string]any)
	if summary["code"] != reference_data.ResultReferencePackImported {
		t.Fatalf("import job summary = %#v", summary)
	}
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 || refs[0].(map[string]any)["kind"] != "reference_pack_version" || refs[0].(map[string]any)["id"] != "/api/v1/reference-packs/type_registry.host/1" {
		t.Fatalf("import job refs = %#v", refs)
	}

	replay := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-reference-pack-import","activation_policy":"staged_only"}`, bundle, "renamed.zip", reference_data.MediaTypeZip)
	replayJob := requireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != jobID {
		t.Fatalf("exact replay returned different job: first=%q replay=%#v", jobID, replayJob)
	}
	divergent := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "2",
	}), "host-pack.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	list := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs?limit=1", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	listBody := requireSuccessEnvelope(t, list, http.StatusOK)
	paging := listBody["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["limit"] != float64(1) || paging["has_more"] != false || paging["next_cursor"] != nil {
		t.Fatalf("unexpected paging: %#v", paging)
	}
	versions := listBody["data"].(map[string]any)["pack_versions"].([]any)
	if len(versions) != 1 {
		t.Fatalf("expected one version, got %#v", versions)
	}
	resource := versions[0].(map[string]any)
	requireReferencePackResource(t, resource, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, false)

	read := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.host/1", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	readResource := requireSuccessEnvelope(t, read, http.StatusOK)["data"].(map[string]any)
	requireReferencePackResource(t, readResource, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, false)

	paginatedSingleton := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.host/1?limit=1", nil, phase2test.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, paginatedSingleton, http.StatusBadRequest, "invalid_pagination_request")
	if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "pagination_not_supported" {
		t.Fatalf("singleton pagination details = %#v", body)
	}
}

func TestPhase11_I_11_REFERENCE_PACK_02_ActivationDisableReverifyAndRefreshLifecycle(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-lifecycle")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	importReferencePack(t, harness, adminLogin, "type_registry.host", "1", "txn-rp-host-v1")
	importReferencePack(t, harness, adminLogin, "type_registry.host", "2", "txn-rp-host-v2")

	activateV1 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/1/activate", "txn-rp-activate-v1", "initial")
	v1 := requireSuccessEnvelope(t, activateV1, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, v1, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, true)
	activateV1Replay := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/1/activate", "txn-rp-activate-v1", "initial")
	v1Replay := requireSuccessEnvelope(t, activateV1Replay, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	if v1Replay["active"] != true {
		t.Fatalf("activation replay must return original success before fresh state checks: %#v", v1Replay)
	}
	alreadyActive := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/1/activate", "txn-rp-activate-v1-again", "")
	requireReasonError(t, alreadyActive, http.StatusConflict, "reference_pack_activation_rejected", "already_active")

	activateV2 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/activate", "txn-rp-activate-v2", "")
	v2 := requireSuccessEnvelope(t, activateV2, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, v2, "type_registry.host", "2", reference_data.ConditionVerifiedAvailable, true)
	if v2["previous_active_version"] != "1" {
		t.Fatalf("previous active must be retained on new activation: %#v", v2)
	}
	requireActivationState(t, harness.DB, "type_registry.host", sql.NullString{String: "2", Valid: true}, sql.NullString{String: "1", Valid: true})
	requirePackMetadata(t, harness.DB, "type_registry.host", "2")
	prior := readReferencePack(t, harness, adminLogin, "type_registry.host", "1")
	requireReferencePackResource(t, prior, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, false)

	disableV2 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/disable", "txn-rp-disable-v2", "")
	disabled := requireSuccessEnvelope(t, disableV2, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, disabled, "type_registry.host", "2", reference_data.ConditionDisabled, false)
	requireActivationState(t, harness.DB, "type_registry.host", sql.NullString{}, sql.NullString{String: "2", Valid: true})
	disableAgain := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/disable", "txn-rp-disable-v2-again", "")
	requireReasonError(t, disableAgain, http.StatusConflict, "reference_pack_state_conflict", "already_disabled")

	reverify := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/reverify", "txn-rp-reverify-v2", "")
	reverifyJob := requireSuccessEnvelope(t, reverify, http.StatusAccepted)["data"].(map[string]any)
	job := requireJob(t, harness, adminLogin, reverifyJob["job_id"].(string))
	if job["status"] != "succeeded" || job["result_summary"].(map[string]any)["code"] != reference_data.ResultReferencePackReverified {
		t.Fatalf("reverify job = %#v", job)
	}
	reverified := readReferencePack(t, harness, adminLogin, "type_registry.host", "2")
	requireReferencePackResource(t, reverified, "type_registry.host", "2", reference_data.ConditionVerifiedAvailable, false)
	requirePackMetadata(t, harness.DB, "type_registry.host", "2")

	refresh := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh",
		"pack_keys":     []string{"type_registry.host", "type_registry.host"},
	}, csrfOptions(adminLogin)...)
	refreshJob := requireSuccessEnvelope(t, refresh, http.StatusAccepted)["data"].(map[string]any)
	refreshDone := requireJob(t, harness, adminLogin, refreshJob["job_id"].(string))
	if refreshDone["status"] != "succeeded" || refreshDone["result_summary"].(map[string]any)["code"] != reference_data.ResultReferencePacksRefreshed {
		t.Fatalf("refresh job = %#v", refreshDone)
	}
}

func TestPhase11_I_11_REFERENCE_PACK_03_FailuresRemainInactiveAndNoNetworkIsNeeded(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-failures")
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	importReferencePack(t, harness, adminLogin, "type_registry.host", "1", "txn-rp-prior-v1")
	activate := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/1/activate", "txn-rp-prior-activate", "")
	requireSuccessEnvelope(t, activate, http.StatusOK)

	failures := []struct {
		name       string
		bundle     []byte
		wantReason string
	}{
		{name: "checksum", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-checksum", BadPayloadSHA: true}), wantReason: "checksum_mismatch"},
		{name: "signature", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-signature", Signed: true, BadSignature: true}), wantReason: "signature_mismatch"},
		{name: "path", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-path", ExtraPath: "../escape.json"}), wantReason: "path_traversal"},
		{name: "active-content", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-content", PayloadPath: "payload/run.js"}), wantReason: "disallowed_content"},
		{name: "missing-payload", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-missing", OmitPayload: true}), wantReason: "payload_missing"},
	}
	for _, tc := range failures {
		t.Run(tc.name, func(t *testing.T) {
			resp := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-rp-failure-`+tc.name+`"}`, tc.bundle, tc.name+".zip", reference_data.MediaTypeZip)
			job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
			done := requireJob(t, harness, adminLogin, job["job_id"].(string))
			if done["status"] != "failed" {
				t.Fatalf("failure job status = %#v", done)
			}
			errorSummary := done["error_summary"].(map[string]any)
			if errorSummary["code"] != "reference_pack_verification_failed" || errorSummary["details"].(map[string]any)["reason_code"] != tc.wantReason {
				t.Fatalf("failure summary = %#v", errorSummary)
			}
			prior := readReferencePack(t, harness, adminLogin, "type_registry.host", "1")
			requireReferencePackResource(t, prior, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, true)
		})
	}
}

func importReferencePack(t testing.TB, harness *phase2test.ServerHarness, login phase2test.LoginResult, packKey string, packVersion string, clientTxnID string) {
	t.Helper()
	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, login, `{"client_txn_id":"`+clientTxnID+`"}`, referencePackBundle(t, bundleOptions{
		PackKey:     packKey,
		PackKind:    "type_registry",
		PackVersion: packVersion,
	}), packKey+"-"+packVersion+".zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	done := requireJob(t, harness, login, job["job_id"].(string))
	if done["status"] != "succeeded" {
		t.Fatalf("import job failed: %#v", done)
	}
}

func postAction(t testing.TB, harness *phase2test.ServerHarness, login phase2test.LoginResult, path string, clientTxnID string, reason string) *http.Response {
	t.Helper()
	body := map[string]any{"client_txn_id": clientTxnID}
	if reason != "" {
		body["reason"] = reason
	}
	return phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+path, body, csrfOptions(login)...)
}

func csrfOptions(login phase2test.LoginResult) []func(*http.Request) {
	return []func(*http.Request){
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	}
}

func readReferencePack(t testing.TB, harness *phase2test.ServerHarness, login phase2test.LoginResult, packKey string, packVersion string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/"+packKey+"/"+packVersion, nil, phase2test.WithCookies(login.SessionCookie))
	return requireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireReferencePackResource(t testing.TB, resource map[string]any, packKey string, packVersion string, state string, active bool) {
	t.Helper()
	if resource["pack_key"] != packKey || resource["pack_version"] != packVersion || resource["pack_version_state"] != state || resource["active"] != active {
		t.Fatalf("unexpected reference pack resource: %#v", resource)
	}
	for _, key := range []string{
		"pack_kind", "source_identifier", "manifest_sha256", "payload_sha256", "pack_contract_version",
		"verification_method", "verification_result", "signer_key_id", "previous_active_version",
		"imported_by_user_id", "imported_at", "activated_by_user_id", "activated_at",
	} {
		if _, ok := resource[key]; !ok {
			t.Fatalf("resource missing %s: %#v", key, resource)
		}
	}
}

func requireReasonError(t testing.TB, resp *http.Response, status int, code string, reason string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, status, code)
	details := body["error"].(map[string]any)["details"].(map[string]any)
	if details["reason_code"] != reason {
		t.Fatalf("error reason = %#v, want %s", details, reason)
	}
}

func requireSuccessEnvelope(t testing.TB, resp *http.Response, status int) map[string]any {
	t.Helper()
	if resp.StatusCode != status {
		body := httptestx.ReadJSONBody(t, resp)
		t.Fatalf("unexpected status: got %d want %d body=%#v", resp.StatusCode, status, body)
	}
	return httptestx.RequireSuccessEnvelope(t, resp, status)
}

func requireActivationState(t testing.TB, db *sql.DB, packKey string, wantActive sql.NullString, wantPrevious sql.NullString) {
	t.Helper()
	var active sql.NullString
	var previous sql.NullString
	if err := db.QueryRow(`SELECT active_version, previous_active_version FROM reference_pack_activation_state WHERE pack_key = $1`, packKey).Scan(&active, &previous); err != nil {
		t.Fatalf("query activation state: %v", err)
	}
	if active.Valid != wantActive.Valid || active.String != wantActive.String || previous.Valid != wantPrevious.Valid || previous.String != wantPrevious.String {
		t.Fatalf("activation state = active=%#v previous=%#v, want active=%#v previous=%#v", active, previous, wantActive, wantPrevious)
	}
}

func requirePackMetadata(t testing.TB, db *sql.DB, packKey string, packVersion string) {
	t.Helper()
	var packKind, manifestSHA, payloadSHA, contractVersion, verificationMethod, verificationResult string
	var source sql.NullString
	var signer sql.NullString
	if err := db.QueryRow(`
SELECT pack_kind, source_identifier, manifest_sha256, payload_sha256, pack_contract_version,
       verification_method, verification_result, signer_key_id
  FROM reference_packs
 WHERE pack_key = $1 AND version = $2
`, packKey, packVersion).Scan(&packKind, &source, &manifestSHA, &payloadSHA, &contractVersion, &verificationMethod, &verificationResult, &signer); err != nil {
		t.Fatalf("query pack metadata: %v", err)
	}
	if packKind == "" || !source.Valid || manifestSHA == "" || payloadSHA == "" || contractVersion != reference_data.PackContractVersionV1 || verificationMethod == "" || verificationResult != "passed" {
		t.Fatalf("incomplete pack metadata: kind=%q source=%#v manifest=%q payload=%q contract=%q method=%q result=%q signer=%#v", packKind, source, manifestSHA, payloadSHA, contractVersion, verificationMethod, verificationResult, signer)
	}
	var attestations int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reference_pack_attestations WHERE pack_key = $1 AND pack_version = $2`, packKey, packVersion).Scan(&attestations); err != nil {
		t.Fatalf("query attestations: %v", err)
	}
	if attestations == 0 {
		t.Fatalf("expected attestation rows for %s/%s", packKey, packVersion)
	}
}
