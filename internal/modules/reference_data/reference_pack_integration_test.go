package reference_data_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestPhase11_I_11_REFERENCE_PACK_01_ImportListReadReplayAndJobSummary(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-import")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	bundle := referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.process",
		PackKind:    "type_registry",
		PackVersion: "1",
	})
	metadata := `{"client_txn_id":"txn-reference-pack-import"}`
	first := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, bundle, "process-pack.zip", reference_data.MediaTypeZip)
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
	if len(refs) != 1 || refs[0].(map[string]any)["kind"] != "reference_pack_version" || refs[0].(map[string]any)["id"] != "/api/v1/reference-packs/type_registry.process/1" {
		t.Fatalf("import job refs = %#v", refs)
	}

	replay := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-reference-pack-import","activation_policy":"staged_only"}`, bundle, "renamed.zip", reference_data.MediaTypeZip)
	replayJob := requireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != jobID {
		t.Fatalf("exact replay returned different job: first=%q replay=%#v", jobID, replayJob)
	}
	divergent := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, metadata, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.process",
		PackKind:    "type_registry",
		PackVersion: "2",
	}), "process-pack.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	list := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs?limit=1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	listBody := requireSuccessEnvelope(t, list, http.StatusOK)
	paging := listBody["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["limit"] != float64(1) || paging["has_more"] != true || paging["next_cursor"] == nil {
		t.Fatalf("unexpected paging: %#v", paging)
	}
	versions := listBody["data"].(map[string]any)["pack_versions"].([]any)
	if len(versions) != 1 {
		t.Fatalf("expected one version, got %#v", versions)
	}
	resource := versions[0].(map[string]any)
	requireReferencePackResource(t, resource, "type_registry.evidence", "1", reference_data.ConditionVerifiedAvailable, true)

	read := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.process/1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	readResource := requireSuccessEnvelope(t, read, http.StatusOK)["data"].(map[string]any)
	requireReferencePackResource(t, readResource, "type_registry.process", "1", reference_data.ConditionVerifiedAvailable, false)

	paginatedSingleton := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.process/1?limit=1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, paginatedSingleton, http.StatusBadRequest, "invalid_pagination_request")
	if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "pagination_not_supported" {
		t.Fatalf("singleton pagination details = %#v", body)
	}
}

func TestPhase11_I_11_REFERENCE_PACK_02_ActivationDisableReverifyAndRefreshLifecycle(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-lifecycle")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	importReferencePack(t, harness, adminLogin, "type_registry.asset", "1", "txn-rp-asset-v1")
	importReferencePack(t, harness, adminLogin, "type_registry.asset", "2", "txn-rp-asset-v2")

	activateV1 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/1/activate", "txn-rp-activate-v1", "initial")
	v1 := requireSuccessEnvelope(t, activateV1, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, v1, "type_registry.asset", "1", reference_data.ConditionVerifiedAvailable, true)
	activateV1Replay := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/1/activate", "txn-rp-activate-v1", "initial")
	v1Replay := requireSuccessEnvelope(t, activateV1Replay, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	if v1Replay["active"] != true {
		t.Fatalf("activation replay must return original success before fresh state checks: %#v", v1Replay)
	}
	alreadyActive := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/1/activate", "txn-rp-activate-v1-again", "")
	requireReasonError(t, alreadyActive, http.StatusConflict, "reference_pack_activation_rejected", "already_active")

	activateV2 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/2/activate", "txn-rp-activate-v2", "")
	v2 := requireSuccessEnvelope(t, activateV2, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, v2, "type_registry.asset", "2", reference_data.ConditionVerifiedAvailable, true)
	if v2["previous_active_version"] != "1" {
		t.Fatalf("previous active must be retained on new activation: %#v", v2)
	}
	requireActivationState(t, harness.DB, "type_registry.asset", sql.NullString{String: "2", Valid: true}, sql.NullString{String: "1", Valid: true})
	requirePackMetadata(t, harness.DB, "type_registry.asset", "2")
	prior := readReferencePack(t, harness, adminLogin, "type_registry.asset", "1")
	requireReferencePackResource(t, prior, "type_registry.asset", "1", reference_data.ConditionVerifiedAvailable, false)

	disableV2 := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/2/disable", "txn-rp-disable-v2", "")
	disabled := requireSuccessEnvelope(t, disableV2, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
	requireReferencePackResource(t, disabled, "type_registry.asset", "2", reference_data.ConditionDisabled, false)
	requireActivationState(t, harness.DB, "type_registry.asset", sql.NullString{}, sql.NullString{String: "2", Valid: true})
	disableAgain := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/2/disable", "txn-rp-disable-v2-again", "")
	requireReasonError(t, disableAgain, http.StatusConflict, "reference_pack_state_conflict", "already_disabled")

	reverify := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/2/reverify", "txn-rp-reverify-v2", "")
	reverifyJob := requireSuccessEnvelope(t, reverify, http.StatusAccepted)["data"].(map[string]any)
	job := requireJob(t, harness, adminLogin, reverifyJob["job_id"].(string))
	if job["status"] != "succeeded" || job["result_summary"].(map[string]any)["code"] != reference_data.ResultReferencePackReverified {
		t.Fatalf("reverify job = %#v", job)
	}
	reverified := readReferencePack(t, harness, adminLogin, "type_registry.asset", "2")
	requireReferencePackResource(t, reverified, "type_registry.asset", "2", reference_data.ConditionVerifiedAvailable, false)
	requirePackMetadata(t, harness.DB, "type_registry.asset", "2")

	refresh := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh",
		"pack_keys":     []string{"type_registry.asset", "type_registry.asset"},
	}, csrfOptions(adminLogin)...)
	refreshJob := requireSuccessEnvelope(t, refresh, http.StatusAccepted)["data"].(map[string]any)
	refreshDone := requireJob(t, harness, adminLogin, refreshJob["job_id"].(string))
	if refreshDone["status"] != "succeeded" || refreshDone["result_summary"].(map[string]any)["code"] != reference_data.ResultReferencePacksRefreshed {
		t.Fatalf("refresh job = %#v", refreshDone)
	}
}

func TestPhase11_I_11_REFERENCE_PACK_03_FailuresRemainInactiveAndNoNetworkIsNeeded(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-failures")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	importReferencePack(t, harness, adminLogin, "type_registry.asset", "1", "txn-rp-prior-v1")
	activate := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.asset/1/activate", "txn-rp-prior-activate", "")
	requireSuccessEnvelope(t, activate, http.StatusOK)

	failures := []struct {
		name       string
		bundle     []byte
		wantReason string
	}{
		{name: "checksum", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.asset", PackKind: "type_registry", PackVersion: "bad-checksum", BadPayloadSHA: true}), wantReason: "checksum_mismatch"},
		{name: "signature", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.asset", PackKind: "type_registry", PackVersion: "bad-signature", Signed: true, BadSignature: true}), wantReason: "signature_mismatch"},
		{name: "path", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.asset", PackKind: "type_registry", PackVersion: "bad-path", ExtraPath: "../escape.json"}), wantReason: "path_traversal"},
		{name: "active-content", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.asset", PackKind: "type_registry", PackVersion: "bad-content", PayloadPath: "payload/run.js"}), wantReason: "disallowed_content"},
		{name: "missing-payload", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.asset", PackKind: "type_registry", PackVersion: "bad-missing", OmitPayload: true}), wantReason: "payload_missing"},
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
			prior := readReferencePack(t, harness, adminLogin, "type_registry.asset", "1")
			requireReferencePackResource(t, prior, "type_registry.asset", "1", reference_data.ConditionVerifiedAvailable, true)
		})
	}
}

func TestPhase11_I_11_REFERENCE_PACK_04_AdmissionQueuesBeforeVerificationAndCancelPreventsCommit(t *testing.T) {
	releaseWorker := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(releaseWorker)
		}
	}()
	workerStarted := make(chan struct{})
	restoreHook := reference_data.SetReferencePackWorkerStartHookForTesting(func(jobKind string) {
		if jobKind != "import" {
			return
		}
		close(workerStarted)
		<-releaseWorker
	})
	defer restoreHook()

	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-async-admission")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-rp-queued-import"}`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.queued",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "queued.zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	jobID := job["job_id"].(string)
	<-workerStarted
	queued := requireJobNow(t, harness, adminLogin, jobID)
	if queued["status"] != "queued" {
		t.Fatalf("job should be accepted before verification starts, got %#v", queued)
	}
	requirePackRowCount(t, harness.DB, "type_registry.queued", "1", 0)

	cancel := cancelJob(t, harness, adminLogin, jobID, "txn-rp-cancel-queued-import")
	cancelBody := requireSuccessEnvelope(t, cancel, http.StatusOK)["data"].(map[string]any)
	if cancelBody["status"] != "cancel_requested" {
		t.Fatalf("cancel response = %#v", cancelBody)
	}
	close(releaseWorker)
	released = true
	done := requireJob(t, harness, adminLogin, jobID)
	if done["status"] != "canceled" {
		t.Fatalf("job should cancel before durable pack commit: %#v", done)
	}
	requirePackRowCount(t, harness.DB, "type_registry.queued", "1", 0)
}

func TestPhase11_I_11_REFERENCE_PACK_05_MinimumDisconnectedBundleSeededExactly(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-minimum-disconnected")

	rows, err := harness.DB.Query(`
SELECT rp.pack_key, rp.version, rp.pack_kind, rp.pack_contract_version, rp.verification_method, rpas.active_version
  FROM reference_packs rp
  JOIN reference_pack_activation_state rpas ON rpas.pack_key = rp.pack_key AND rpas.active_version = rp.version
 ORDER BY rp.pack_key ASC
`)
	if err != nil {
		t.Fatalf("query seeded packs: %v", err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var packKey, version, packKind, contractVersion, verificationMethod, activeVersion string
		if err := rows.Scan(&packKey, &version, &packKind, &contractVersion, &verificationMethod, &activeVersion); err != nil {
			t.Fatalf("scan seeded pack: %v", err)
		}
		if packKind != "type_registry" || contractVersion != reference_data.PackContractVersionV1 || verificationMethod != "manifest_sha256_v1" || activeVersion != "1" {
			t.Fatalf("seeded pack metadata mismatch for %s: kind=%s contract=%s method=%s active=%s", packKey, packKind, contractVersion, verificationMethod, activeVersion)
		}
		got = append(got, packKey+"@"+version)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("seed rows: %v", err)
	}
	want := []string{"type_registry.evidence@1", "type_registry.host@1", "type_registry.indicator@1"}
	if len(got) != len(want) {
		t.Fatalf("seeded pack count = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("seeded packs = %#v, want %#v", got, want)
		}
	}
}

func TestPhase11_I_11_REFERENCE_PACK_06_RefreshOmittedSelectorReplayUsesAdmittedSet(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-refresh-replay")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	first := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh-omitted",
	}, csrfOptions(adminLogin)...)
	firstJob := requireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	firstJobID := firstJob["job_id"].(string)
	requireJob(t, harness, adminLogin, firstJobID)

	importReferencePack(t, harness, adminLogin, "type_registry.after_refresh", "1", "txn-rp-after-refresh-import")

	replay := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh-omitted",
	}, csrfOptions(adminLogin)...)
	replayJob := requireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != firstJobID {
		t.Fatalf("omitted refresh replay should return original job after visibility changes: first=%s replay=%#v", firstJobID, replayJob)
	}
}

func TestPhase11_I_11_REFERENCE_PACK_07_UploadEnvelopeFailureCreatesNoDurableStateAndAdminIsRequired(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-envelope-and-authz")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	beforeJobs := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_job_payloads`)
	badEnvelope := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.bad_envelope",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "bad-envelope.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, badEnvelope, http.StatusBadRequest, "invalid_reference_pack_request")
	requirePackRowCount(t, harness.DB, "type_registry.bad_envelope", "1", 0)
	afterJobs := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_job_payloads`)
	if afterJobs != beforeJobs {
		t.Fatalf("upload envelope failure created job payloads: before=%d after=%d", beforeJobs, afterJobs)
	}

	createUser := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/users", map[string]any{
		"client_txn_id":    "txn-rp-create-non-admin",
		"auth_kind":        "local",
		"email":            "rp-non-admin@example.test",
		"display_name":     "Reference Pack Non Admin",
		"initial_password": "ReferencePackPass123!",
		"mfa_required":     false,
	}, csrfOptions(adminLogin)...)
	httptestx.RequireSuccessEnvelope(t, createUser, http.StatusCreated)
	sessionCookie, csrfCookie := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "rp-non-admin@example.test", "ReferencePackPass123!", nil)
	nonAdminLogin := flowtest.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
	sessionCheck := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/auth/session", nil, httptestx.WithCookies(nonAdminLogin.SessionCookie))
	httptestx.RequireSuccessEnvelope(t, sessionCheck, http.StatusOK)
	denied := postReferencePackUpload(t, harness.Server.HTTP.URL, nonAdminLogin, `{"client_txn_id":"txn-rp-denied"}`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.denied",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "denied.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
	requirePackRowCount(t, harness.DB, "type_registry.denied", "1", 0)
}

func TestPhase11_I_11_REFERENCE_PACK_08_OptionalPackStatesDegradeOnlyOptionalSurfacesAndPreserveCoreWorkflows(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-optional-degradation")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	baselineViewSchemas := viewSchemaIDs(t, harness, adminLogin)
	cases := []struct {
		name    string
		packKey string
		arrange func(t *testing.T, packKey string)
	}{
		{
			name:    "absent",
			packKey: "framework.attack",
			arrange: func(t *testing.T, packKey string) {
				requirePackRowCount(t, harness.DB, packKey, "1", 0)
			},
		},
		{
			name:    "disabled",
			packKey: "framework.d3fend",
			arrange: func(t *testing.T, packKey string) {
				importReferencePackWithKind(t, harness, adminLogin, packKey, "framework", "1", "txn-rp-degrade-disabled-import")
				resp := postAction(t, harness, adminLogin, "/api/v1/reference-packs/"+packKey+"/1/disable", "txn-rp-degrade-disabled-disable", "")
				resource := requireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)["pack_version"].(map[string]any)
				requireReferencePackResource(t, resource, packKey, "1", reference_data.ConditionDisabled, false)
			},
		},
		{
			name:    "failed",
			packKey: "enrichment.tor",
			arrange: func(t *testing.T, packKey string) {
				importReferencePackWithKind(t, harness, adminLogin, packKey, "enrichment", "1", "txn-rp-degrade-failed-import")
				overwriteStoredBundle(t, harness.DB, packKey, "1", referencePackBundle(t, bundleOptions{
					PackKey:       packKey,
					PackKind:      "enrichment",
					PackVersion:   "1",
					BadPayloadSHA: true,
				}))
				requireReverifyFailure(t, harness, adminLogin, packKey, "1", "txn-rp-degrade-failed-reverify", "checksum_mismatch")
				resource := readReferencePack(t, harness, adminLogin, packKey, "1")
				requireReferencePackResource(t, resource, packKey, "1", reference_data.ConditionFailed, false)
			},
		},
		{
			name:    "missing",
			packKey: "enrichment.cisa_kev",
			arrange: func(t *testing.T, packKey string) {
				importReferencePackWithKind(t, harness, adminLogin, packKey, "enrichment", "1", "txn-rp-degrade-missing-import")
				activate := postAction(t, harness, adminLogin, "/api/v1/reference-packs/"+packKey+"/1/activate", "txn-rp-degrade-missing-activate", "")
				requireSuccessEnvelope(t, activate, http.StatusOK)
				removeStoredBundle(t, harness.DB, packKey, "1")
				requireReverifyFailure(t, harness, adminLogin, packKey, "1", "txn-rp-degrade-missing-reverify", "payload_missing")
				resource := readReferencePack(t, harness, adminLogin, packKey, "1")
				requireReferencePackResource(t, resource, packKey, "1", reference_data.ConditionMissing, false)
				requireActivationState(t, harness.DB, packKey, sql.NullString{}, sql.NullString{String: "1", Valid: true})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.arrange(t, tc.packKey)
			exerciseCoreWorkflowDuringOptionalPackDegradation(t, harness, adminLogin, adminID, tc.name)
			gotViewSchemas := viewSchemaIDs(t, harness, adminLogin)
			requireStringSlicesEqual(t, gotViewSchemas, baselineViewSchemas, "view schema inventory changed while optional pack was "+tc.name)
		})
	}
}

func TestPhase11_I_11_REFERENCE_PACK_09_JobsRequireDeploymentAdminAtPollAndCancelTime(t *testing.T) {
	releaseWorker := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(releaseWorker)
		}
	}()
	workerStarted := make(chan struct{})
	restoreHook := reference_data.SetReferencePackWorkerStartHookForTesting(func(jobKind string) {
		if jobKind != "import" {
			return
		}
		close(workerStarted)
		<-releaseWorker
	})
	defer restoreHook()

	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reference-pack-job-authz")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-rp-job-auth-import"}`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.job_auth",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "job-auth.zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	jobID := job["job_id"].(string)
	<-workerStarted
	queued := requireJobNow(t, harness, adminLogin, jobID)
	if queued["status"] != "queued" {
		t.Fatalf("job should be queued before auth mutation, got %#v", queued)
	}

	setDeploymentAdmin(t, harness.DB, adminID, false)
	afterDemotionRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(adminLogin.SessionCookie))
	httptestx.RequireErrorEnvelope(t, afterDemotionRead, http.StatusNotFound, "job_not_found")
	afterDemotionCancel := cancelJob(t, harness, adminLogin, jobID, "txn-rp-job-auth-cancel-after-demotion")
	httptestx.RequireErrorEnvelope(t, afterDemotionCancel, http.StatusNotFound, "job_not_found")

	setDeploymentAdmin(t, harness.DB, adminID, true)
	close(releaseWorker)
	released = true
	done := requireJob(t, harness, adminLogin, jobID)
	if done["status"] != "succeeded" {
		t.Fatalf("job should complete after admin restoration: %#v", done)
	}
}

func importReferencePack(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string, clientTxnID string) {
	t.Helper()
	importReferencePackWithKind(t, harness, login, packKey, "type_registry", packVersion, clientTxnID)
}

func importReferencePackWithKind(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, packKey string, packKind string, packVersion string, clientTxnID string) {
	t.Helper()
	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, login, `{"client_txn_id":"`+clientTxnID+`"}`, referencePackBundle(t, bundleOptions{
		PackKey:     packKey,
		PackKind:    packKind,
		PackVersion: packVersion,
	}), packKey+"-"+packVersion+".zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	done := requireJob(t, harness, login, job["job_id"].(string))
	if done["status"] != "succeeded" {
		t.Fatalf("import job failed: %#v", done)
	}
}

func cancelJob(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, jobID string, clientTxnID string) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": clientTxnID,
	}, csrfOptions(login)...)
}

func postAction(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, path string, clientTxnID string, reason string) *http.Response {
	t.Helper()
	body := map[string]any{"client_txn_id": clientTxnID}
	if reason != "" {
		body["reason"] = reason
	}
	return httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+path, body, csrfOptions(login)...)
}

func csrfOptions(login flowtest.LoginResult) []func(*http.Request) {
	return []func(*http.Request){
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	}
}

func readReferencePack(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/"+packKey+"/"+packVersion, nil, httptestx.WithCookies(login.SessionCookie))
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

func requirePackRowCount(t testing.TB, db *sql.DB, packKey string, packVersion string, want int) {
	t.Helper()
	got := queryCount(t, db, `SELECT count(*) FROM reference_packs WHERE pack_key = $1 AND version = $2`, packKey, packVersion)
	if got != want {
		t.Fatalf("reference pack %s/%s row count = %d, want %d", packKey, packVersion, got, want)
	}
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return got
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

func exerciseCoreWorkflowDuringOptionalPackDegradation(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, adminID string, suffix string) {
	t.Helper()
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-rp-degrade-" + suffix + "-incident",
		"incident_key":  "IR-RP-DEGRADE-" + strings.ToUpper(suffix),
		"title":         "Reference Pack degradation " + suffix,
	})
	incidentID := incident["incident_id"].(string)

	hostResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+hostidentity.HostsViewSchemaID+"/rows", map[string]any{
		"client_txn_id":     "txn-rp-degrade-" + suffix + "-host",
		"host.display_name": "Reference Pack degradation host " + suffix,
		"host.hostname":     "rp-" + suffix + "-host",
	}, csrfOptions(login)...)
	hostData := requireSuccessEnvelope(t, hostResp, http.StatusCreated)["data"].(map[string]any)
	hostID := hostData["row"].(map[string]any)["record_id"].(string)

	timelineData := timelineroutetest.CreateRow(t, harness.Server, login, incidentID, map[string]any{
		"client_txn_id":                   "txn-rp-degrade-" + suffix + "-timeline",
		"timeline.activity_synopsis_text": "Reference Pack degradation timeline " + suffix,
		"timeline.host_refs": collectionActions(
			addResolvedRefAction("rp-"+suffix+"-host", hostID),
		),
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineID := timelineRow["record_id"].(string)
	timelineVersion := int(timelineRow["row_version"].(float64))
	requireResolvedCollectionItem(t, queryViewRow(t, harness, login, incidentID, timeline.TimelineViewSchemaID, timelineID), "timeline.host_refs", hostID)

	patchResp := httptestx.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+timelineID, map[string]any{
		"view_schema_id":   timeline.TimelineViewSchemaID,
		"base_row_version": timelineVersion,
		"client_txn_id":    "txn-rp-degrade-" + suffix + "-timeline-edit",
		"changes": []map[string]any{{
			"field_key": "timeline.raw_activity_text",
			"value":     "Core edit while optional Reference Pack is " + suffix,
		}},
	}, csrfOptions(login)...)
	requireSuccessEnvelope(t, patchResp, http.StatusOK)

	evidenceResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+workbook.EvidenceViewSchemaID+"/rows", map[string]any{
		"client_txn_id":  "txn-rp-degrade-" + suffix + "-evidence",
		"evidence.title": "Reference Pack degradation evidence " + suffix,
	}, csrfOptions(login)...)
	evidenceData := requireSuccessEnvelope(t, evidenceResp, http.StatusCreated)["data"].(map[string]any)
	evidenceRow := evidenceData["row"].(map[string]any)
	evidenceID := evidenceRow["record_id"].(string)
	evidenceVersion := int(evidenceRow["row_version"].(float64))

	payload := []byte("reference pack degradation evidence " + suffix)
	sum := sha256.Sum256(payload)
	blobResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID,
		"client_txn_id":     "txn-rp-degrade-" + suffix + "-blob",
		"byte_size":         len(payload),
		"filename_hint":     "rp-" + suffix + ".txt",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}, csrfOptions(login)...)
	blobData := requireSuccessEnvelope(t, blobResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, harness.Server.HTTP.URL, blobData["upload_target"].(map[string]any)["href"].(string), payload, "text/plain")
	attachResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceID+"/attach-blob", map[string]any{
		"object_blob_id":   blobData["object_blob_id"],
		"base_row_version": evidenceVersion,
		"client_txn_id":    "txn-rp-degrade-" + suffix + "-attach",
	}, csrfOptions(login)...)
	attachData := requireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	if attachData["object_blob_id"] != blobData["object_blob_id"] {
		t.Fatalf("attached blob mismatch: attach=%#v blob=%#v", attachData, blobData)
	}

	if got := queryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE actor_user_id::text = $1 AND incident_id::text = $2`, adminID, incidentID); got < 4 {
		t.Fatalf("expected core workflow mutations to commit during optional pack %s state, got %d change sets", suffix, got)
	}
}

func collectionActions(actions ...map[string]any) map[string]any {
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func addResolvedRefAction(rawText string, resolvedRecordID string) map[string]any {
	return map[string]any{"op": "add_resolved_ref", "raw_text": rawText, "resolved_record_id": resolvedRecordID}
}

func requireResolvedCollectionItem(t testing.TB, row map[string]any, fieldKey string, resolvedRecordID string) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	value := cell["value"].(map[string]any)
	rawItems := value["items"].([]any)
	if len(rawItems) != 1 {
		t.Fatalf("expected one %s item, got %#v", fieldKey, rawItems)
	}
	item := rawItems[0].(map[string]any)
	if item["item_kind"] != "resolved_ref" || item["resolved_record_id"] != resolvedRecordID {
		t.Fatalf("unexpected resolved collection item for %s: %#v", fieldKey, item)
	}
}

func queryViewRow(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, incidentID string, viewSchemaID string, recordID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query", map[string]any{}, httptestx.WithCookies(login.SessionCookie))
	body := requireSuccessEnvelope(t, resp, http.StatusOK)
	rows := body["data"].(map[string]any)["rows"].([]any)
	for _, rawRow := range rows {
		row := rawRow.(map[string]any)
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("expected row %s in %s rows %#v", recordID, viewSchemaID, rows)
	return nil
}

func viewSchemaIDs(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult) []string {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/view-schemas?limit=100", nil, httptestx.WithCookies(login.SessionCookie))
	body := requireSuccessEnvelope(t, resp, http.StatusOK)
	rawSchemas := body["data"].(map[string]any)["view_schemas"].([]any)
	ids := make([]string, 0, len(rawSchemas))
	for _, rawSchema := range rawSchemas {
		ids = append(ids, rawSchema.(map[string]any)["view_schema_id"].(string))
	}
	sort.Strings(ids)
	return ids
}

func requireStringSlicesEqual(t testing.TB, got []string, want []string, message string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %#v want %#v", message, got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("%s: got %#v want %#v", message, got, want)
		}
	}
}

func overwriteStoredBundle(t testing.TB, db *sql.DB, packKey string, packVersion string, bundle []byte) {
	t.Helper()
	path := storedBundlePath(t, db, packKey, packVersion)
	if err := os.WriteFile(path, bundle, 0o600); err != nil {
		t.Fatalf("overwrite stored bundle: %v", err)
	}
}

func removeStoredBundle(t testing.TB, db *sql.DB, packKey string, packVersion string) {
	t.Helper()
	path := storedBundlePath(t, db, packKey, packVersion)
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove stored bundle: %v", err)
	}
}

func storedBundlePath(t testing.TB, db *sql.DB, packKey string, packVersion string) string {
	t.Helper()
	var path string
	if err := db.QueryRow(`SELECT bundle_storage_path FROM reference_packs WHERE pack_key = $1 AND version = $2`, packKey, packVersion).Scan(&path); err != nil {
		t.Fatalf("query stored bundle path: %v", err)
	}
	return path
}

func requireReverifyFailure(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string, clientTxnID string, reasonCode string) {
	t.Helper()
	resp := postAction(t, harness, login, "/api/v1/reference-packs/"+packKey+"/"+packVersion+"/reverify", clientTxnID, "")
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	done := requireJob(t, harness, login, job["job_id"].(string))
	if done["status"] != "failed" {
		t.Fatalf("expected failed reverify job, got %#v", done)
	}
	summary := done["error_summary"].(map[string]any)
	if summary["code"] != "reference_pack_verification_failed" || summary["details"].(map[string]any)["reason_code"] != reasonCode {
		t.Fatalf("unexpected reverify error summary: %#v", summary)
	}
}

func setDeploymentAdmin(t testing.TB, db *sql.DB, userID string, isAdmin bool) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `UPDATE users SET is_deployment_admin = $2, updated_at = now() WHERE id::text = $1`, userID, isAdmin); err != nil {
		t.Fatalf("set deployment admin flag: %v", err)
	}
}

func putObject(t testing.TB, baseURL string, href string, payload []byte, contentType string) {
	t.Helper()
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(http.MethodPut, href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("upload object: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload object status %d: %s", resp.StatusCode, string(data))
	}
}
