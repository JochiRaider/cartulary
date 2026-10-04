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
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestImportListReadReplayAndJobSummary_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-import")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	bundle := referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
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
	var raw string
	if err := harness.DB.QueryRow(`SELECT e.container_ref FROM reference_pack_candidates c JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE c.pack_key='type_registry.host' AND c.pack_version='1'`).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	if _, err := reference_data.ParseStorageRef(raw); err != nil || filepath.IsAbs(raw) {
		t.Fatalf("invalid opaque container reference: %v", err)
	}
	if _, err := os.Stat(filepath.Join(harness.Server.Config.Roots.ReferencePackStorage.Path, filepath.FromSlash(raw))); err != nil {
		t.Fatal(err)
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_operations WHERE job_id=$1 AND terminal_at IS NOT NULL`, jobID) != 1 {
		t.Fatal("missing terminal operation")
	}
	summary := job["result_summary"].(map[string]any)
	if summary["code"] != reference_data.ResultReferencePackImported {
		t.Fatalf("import job summary = %#v", summary)
	}
	requireReferencePackProof(t, harness.DB, jobID, reference_data.ImportJobKind, reference_data.ImportOperation)
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
	requireReferencePackResource(t, resource, "type_registry.evidence", "base-2026-10-02.1", reference_data.ConditionVerifiedAvailable, true)

	read := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.host/1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	readResource := requireSuccessEnvelope(t, read, http.StatusOK)["data"].(map[string]any)
	requireReferencePackResource(t, readResource, "type_registry.host", "1", reference_data.ConditionVerifiedAvailable, false)
	for _, key := range []string{"pack_release_sequence", "source_profile_id", "source_profile_sha256", "source_version", "license_expression", "redistribution", "trust_repository_id"} {
		if value, present := readResource[key]; !present || value == nil {
			t.Fatalf("successful manifest projection missing %s", key)
		}
	}
	if readResource["trust_repository_id"] != "integration.repo" || resource["trust_repository_id"] != nil {
		t.Fatal("operator and built-in trust provenance collapsed")
	}

	paginatedSingleton := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/reference-packs/type_registry.host/1?limit=1", nil, httptestx.WithCookies(adminLogin.SessionCookie))
	body := httptestx.RequireErrorEnvelope(t, paginatedSingleton, http.StatusBadRequest, "invalid_pagination_request")
	if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "pagination_not_supported" {
		t.Fatalf("singleton pagination details = %#v", body)
	}
}

func TestActivationDisableReverifyAndRefreshLifecycle_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-lifecycle")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	importReferencePack(t, harness, adminLogin, "type_registry.host", "1", "txn-rp-asset-v1")
	importReferencePack(t, harness, adminLogin, "type_registry.host", "2", "txn-rp-asset-v2")
	importReferencePack(t, harness, adminLogin, "type_registry.host", "3", "txn-rp-asset-v3")
	removeStoredBundle(t, harness, "type_registry.host", "3")
	missingStorage := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/3/activate", "txn-rp-activate-v3-missing", "")
	requireReasonError(t, missingStorage, http.StatusConflict, "reference_pack_activation_rejected", "not_verified_available")

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
	requireActivationState(t, harness.DB, "type_registry.host", sql.NullString{String: "base-2026-10-02.1", Valid: true}, sql.NullString{})
	disableAgain := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/disable", "txn-rp-disable-v2-again", "")
	requireReasonError(t, disableAgain, http.StatusConflict, "reference_pack_operation_rejected", "not_disableable")

	reverify := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/reverify", "txn-rp-reverify-v2", "")
	reverifyJob := requireSuccessEnvelope(t, reverify, http.StatusAccepted)["data"].(map[string]any)
	job := requireJob(t, harness, adminLogin, reverifyJob["job_id"].(string))
	if job["status"] != "succeeded" || job["result_summary"].(map[string]any)["code"] != reference_data.ResultReferencePackReverified {
		t.Fatalf("reverify job = %#v", job)
	}
	requireReferencePackProof(t, harness.DB, reverifyJob["job_id"].(string), reference_data.ReverifyJobKind, reference_data.ReverifyOperation)
	reverified := readReferencePack(t, harness, adminLogin, "type_registry.host", "2")
	requireReferencePackResource(t, reverified, "type_registry.host", "2", reference_data.ConditionDisabled, false)
	requirePackMetadata(t, harness.DB, "type_registry.host", "2")

	refresh := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh",
		"pack_keys":     []string{"type_registry.host", "type_registry.host"},
	}, csrfOptions(adminLogin)...)
	refreshJob := requireSuccessEnvelope(t, refresh, http.StatusAccepted)["data"].(map[string]any)
	refreshDone := requireJob(t, harness, adminLogin, refreshJob["job_id"].(string))
	if refreshDone["status"] != "failed" || refreshDone["error_summary"].(map[string]any)["code"] != "reference_pack_verification_failed" {
		t.Fatalf("refresh job = %#v", refreshDone)
	}
	requireReferencePackResource(t, readReferencePack(t, harness, adminLogin, "type_registry.host", "2"), "type_registry.host", "2", reference_data.ConditionDisabled, false)
	requireReferencePackResource(t, readReferencePack(t, harness, adminLogin, "type_registry.host", "3"), "type_registry.host", "3", reference_data.ConditionMissing, false)
	reactivate := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/2/activate", "txn-rp-reactivate-v2", "private operator reason")
	requireSuccessEnvelope(t, reactivate, http.StatusOK)
	for _, kind := range []string{"activation", "rollback_activation"} {
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE pack_key='type_registry.host' AND pack_version='2' AND event_kind=$1`, kind); got != 1 {
			t.Fatalf("first activation/reactivation classification %s: %d", kind, got)
		}
	}
	if got := queryCount(t, harness.DB, `SELECT count(*) FROM deployment_admin_audit_events WHERE event_source='reference_pack' AND after_json::text LIKE '%private operator reason%'`); got != 0 {
		t.Fatal("operator reason leaked into audit")
	}
}

func TestFailuresRemainInactiveAndNoNetworkIsNeeded_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-failures")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	importReferencePack(t, harness, adminLogin, "type_registry.host", "1", "txn-rp-prior-v1")
	activate := postAction(t, harness, adminLogin, "/api/v1/reference-packs/type_registry.host/1/activate", "txn-rp-prior-activate", "")
	requireSuccessEnvelope(t, activate, http.StatusOK)

	failures := []struct {
		name       string
		bundle     []byte
		wantReason string
	}{
		{name: "checksum", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-checksum", BadPayloadSHA: true}), wantReason: "target_length_mismatch"},
		{name: "signature", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-signature", Signed: true, BadSignature: true}), wantReason: "signature_threshold_not_met"},
		{name: "path", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-path", ExtraPath: "../escape.json"}), wantReason: "path_traversal"},
		{name: "active-content", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-content", PayloadPath: "payload/run.js"}), wantReason: "target_not_declared"},
		{name: "missing-payload", bundle: referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "bad-missing", OmitPayload: true}), wantReason: "unexpected_target"},
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

func TestAdmissionQueuesBeforeVerificationAndCancelPreventsCommit_Integration(t *testing.T) {
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

	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-async-admission")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-rp-queued-import"}`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "queued.zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	jobID := job["job_id"].(string)
	select {
	case <-workerStarted:
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not start")
	}
	running := requireJobNow(t, harness, adminLogin, jobID)
	if running["status"] != "running" {
		t.Fatalf("job must be running before verification starts, got %#v", running)
	}
	requirePackRowCount(t, harness.DB, "type_registry.host", "1", 1)
	staged := readReferencePack(t, harness, adminLogin, "type_registry.host", "1")
	signers, ok := staged["verified_signer_key_ids"].([]any)
	if staged["last_verified_at"] != nil || !ok || len(signers) != 0 || staged["health"] != "staged" {
		t.Fatalf("staged success metadata: %#v", staged)
	}
	for _, key := range []string{"pack_release_sequence", "source_profile_id", "source_profile_sha256", "source_version", "source_as_of", "license_expression", "redistribution", "trust_repository_id"} {
		if value, present := staged[key]; !present || value != nil {
			t.Fatalf("staged projection fabricated or omitted %s", key)
		}
	}

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
	requirePackRowCount(t, harness.DB, "type_registry.host", "1", 1)
	aborted := readReferencePack(t, harness, adminLogin, "type_registry.host", "1")
	if aborted["health"] != "failed" || aborted["last_failure_code"] != nil || aborted["last_verified_at"] != nil {
		t.Fatalf("abort fabricated verdict: %#v", aborted)
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM extension_job_cancellation_observations WHERE job_id = $1`, jobID) != 1 {
		t.Fatal("accepted Reference Pack cancellation must retain one observation")
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, jobID) != 0 {
		t.Fatal("canceled Reference Pack job must not publish a success proof")
	}
	if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_envelopes WHERE pack_key='type_registry.host' AND pack_version='1'`) != 0 {
		t.Fatal("cancellation published an envelope")
	}

}

func TestMinimumDisconnectedBundleSeededExactly_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	t.Run("claimed profile seeds the exact minimum", func(t *testing.T) {
		harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-minimum-disconnected")
		rows, err := harness.DB.Query(`SELECT v.pack_key,v.pack_version,v.manifest_bytes FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id) JOIN reference_pack_versions v USING(pack_key,pack_version) ORDER BY v.pack_key`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		var got []string
		for rows.Next() {
			var key, version string
			var manifest []byte
			if err := rows.Scan(&key, &version, &manifest); err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(manifest, []byte(`"pack_contract_version":"cartulary.reference_pack_contract.v1"`)) {
				t.Fatal("noncanonical Base contract")
			}
			got = append(got, key+"@"+version)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		requireStringSlicesEqual(t, got, []string{"type_registry.evidence@base-2026-10-02.1", "type_registry.host@base-2026-10-02.1", "type_registry.indicator@base-2026-10-02.1"}, "Base inventory")
		if queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_indexes`) < 11 {
			t.Fatal("Base registries are empty")
		}

	})

	t.Run("unclaimed profile establishes Base without operator routes", func(t *testing.T) {
		unclaimed := runtime.StartServer(t, appsupport.ServerOptions{
			Prefix: "extension_profile-reference-pack-unclaimed",
			Env: map[string]string{
				"CARTULARY__REFERENCE_PACK__CLAIMED": "false",
			},
			TestRouteMode: httptestx.TestRouteModeDisabled,
		})
		if got := queryCount(t, unclaimed.DB, `SELECT count(*) FROM reference_pack_versions`); got != 3 {
			t.Fatalf("unclaimed Reference Pack profile seeded %d pack rows", got)
		}
		response := httptestx.DoJSON(t, http.MethodGet, unclaimed.Server.HTTP.URL+"/api/v1/reference-packs", nil)
		httptestx.RequireErrorEnvelope(t, response, http.StatusNotFound, "extension_profile_not_claimed")
	})
}

func TestRefreshOmittedSelectorReplayUsesAdmittedSet_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-refresh-replay")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	first := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh-omitted",
	}, csrfOptions(adminLogin)...)
	firstJob := requireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	firstJobID := firstJob["job_id"].(string)
	requireJob(t, harness, adminLogin, firstJobID)

	importReferencePack(t, harness, adminLogin, "type_registry.host", "1", "txn-rp-after-refresh-import")

	replay := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/refresh", map[string]any{
		"client_txn_id": "txn-rp-refresh-omitted",
	}, csrfOptions(adminLogin)...)
	replayJob := requireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != firstJobID {
		t.Fatalf("omitted refresh replay should return original job after visibility changes: first=%s replay=%#v", firstJobID, replayJob)
	}
}

func TestUploadEnvelopeFailureCreatesNoDurableStateAndAdminIsRequired_Integration(t *testing.T) {
	t.Run("streamed container limit", testOversizedUploadAdmission)
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-envelope-and-authz")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	beforeJobs := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_operations`)
	badEnvelope := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "bad-envelope.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, badEnvelope, http.StatusBadRequest, "invalid_reference_pack_request")
	requirePackRowCount(t, harness.DB, "type_registry.host", "1", 0)
	afterJobs := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_operations`)
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
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "denied.zip", reference_data.MediaTypeZip)
	httptestx.RequireErrorEnvelope(t, denied, http.StatusForbidden, "authorization_denied")
	requirePackRowCount(t, harness.DB, "type_registry.host", "1", 0)
}

func TestOptionalPackStatesDegradeOnlyOptionalSurfacesAndPreserveCoreWorkflows_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-optional-degradation")
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
				path := storedBundlePath(t, harness, packKey, "1")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				data[0] ^= 1
				overwriteStoredBundle(t, harness, packKey, "1", data)
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
				removeStoredBundle(t, harness, packKey, "1")
				requireReverifyFailure(t, harness, adminLogin, packKey, "1", "txn-rp-degrade-missing-reverify", "payload_missing")
				resource := readReferencePack(t, harness, adminLogin, packKey, "1")
				requireReferencePackResource(t, resource, packKey, "1", reference_data.ConditionMissing, false)
				requireActivationState(t, harness.DB, packKey, sql.NullString{}, sql.NullString{String: "1", Valid: true})
			},
		},
		{
			name:    "symlinked",
			packKey: "enrichment.windows_sids",
			arrange: func(t *testing.T, packKey string) {
				importReferencePackWithKind(t, harness, adminLogin, packKey, "enrichment", "1", "txn-rp-degrade-symlink-import")
				path := storedBundlePath(t, harness, packKey, "1")
				if err := os.Remove(path); err != nil {
					t.Fatalf("remove stored bundle before symlink: %v", err)
				}
				outside := filepath.Join(t.TempDir(), "outside.bundle")
				if err := os.WriteFile(outside, []byte("must not be read"), 0o600); err != nil {
					t.Fatalf("write symlink target: %v", err)
				}
				if err := os.Symlink(outside, path); err != nil {
					t.Fatalf("replace stored bundle with symlink: %v", err)
				}
				resp := postAction(t, harness, adminLogin, "/api/v1/reference-packs/"+packKey+"/1/reverify", "txn-rp-degrade-symlink-reverify", "")
				job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
				done := requireJob(t, harness, adminLogin, job["job_id"].(string))
				if done["status"] != "failed" || done["error_summary"].(map[string]any)["code"] != "reference_pack_verification_failed" || done["error_summary"].(map[string]any)["details"].(map[string]any)["reason_code"] != "payload_missing" {
					t.Fatalf("confined read outcome: %#v", done)
				}
				resource := readReferencePack(t, harness, adminLogin, packKey, "1")
				requireReferencePackResource(t, resource, packKey, "1", reference_data.ConditionMissing, false)
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

func TestJobsRequireDeploymentAdminAtPollAndCancelTime_Integration(t *testing.T) {
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

	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "extension_profile-reference-pack-job-authz")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	resp := postReferencePackUpload(t, harness.Server.HTTP.URL, adminLogin, `{"client_txn_id":"txn-rp-job-auth-import"}`, referencePackBundle(t, bundleOptions{
		PackKey:     "type_registry.host",
		PackKind:    "type_registry",
		PackVersion: "1",
	}), "job-auth.zip", reference_data.MediaTypeZip)
	job := requireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	jobID := job["job_id"].(string)
	select {
	case <-workerStarted:
	case <-time.After(10 * time.Second):
		t.Fatal("worker did not start")
	}
	running := requireJobNow(t, harness, adminLogin, jobID)
	if running["status"] != "running" {
		t.Fatalf("job must be running before auth mutation, got %#v", running)
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

func importReferencePack(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string, clientTxnID string) {
	t.Helper()
	importReferencePackWithKind(t, harness, login, packKey, "type_registry", packVersion, clientTxnID)
}

func importReferencePackWithKind(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, packKey string, packKind string, packVersion string, clientTxnID string) {
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

func cancelJob(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, jobID string, clientTxnID string) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": clientTxnID,
	}, csrfOptions(login)...)
}

func postAction(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, path string, clientTxnID string, reason string) *http.Response {
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

func readReferencePack(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string) map[string]any {
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
		"verification_method", "last_verified_at", "verified_signer_key_ids", "previous_active_version",
		"health", "administratively_disabled", "removed", "missing_reason",
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
	if err := db.QueryRow(`SELECT (SELECT m.pack_version FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key=$1)`, packKey).Scan(&active); err != nil {
		t.Fatal(err)
	}
	if active != wantActive {
		t.Fatalf("active=%#v, want %#v", active, wantActive)
	}
	if wantPrevious.Valid && wantActive.Valid {
		var previous sql.NullString
		if err := db.QueryRow(`SELECT convert_from(e.canonical_attestation,'UTF8')::jsonb->>'prior_active_version' FROM reference_pack_events e JOIN reference_pack_operations o USING(operation_id) WHERE e.pack_key=$1 AND e.pack_version=$2 AND e.event_kind IN ('activation','rollback_activation') ORDER BY o.admitted_at DESC LIMIT 1`, packKey, wantActive.String).Scan(&previous); err != nil {
			t.Fatal(err)
		}
		if previous != wantPrevious {
			t.Fatal("missing previous set provenance")
		}
	}

}

func requirePackRowCount(t testing.TB, db *sql.DB, packKey string, packVersion string, want int) {
	t.Helper()
	got := queryCount(t, db, `SELECT count(*) FROM reference_pack_candidates WHERE pack_key = $1 AND pack_version = $2`, packKey, packVersion)
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

func requireReferencePackProof(t testing.TB, db *sql.DB, jobID string, jobKind string, operationKind string) {
	t.Helper()
	var ownerProfileID string
	var actualOperationKind string
	var finalCommitID string
	if err := db.QueryRow(`
SELECT owner_profile_id, operation_kind, final_commit_id
  FROM extension_job_commit_proofs
 WHERE job_id::text = $1
`, jobID).Scan(&ownerProfileID, &actualOperationKind, &finalCommitID); err != nil {
		t.Fatalf("read Reference Pack proof for job %s: %v", jobID, err)
	}
	var operationID string
	if err := db.QueryRow(`SELECT operation_id::text FROM reference_pack_operations WHERE job_id=$1`, jobID).Scan(&operationID); err != nil {
		t.Fatal(err)
	}
	if ownerProfileID != reference_data.ProfileID || actualOperationKind != operationKind || finalCommitID != reference_data.ProfileID+":"+operationID {
		t.Fatalf(
			"unexpected Reference Pack proof: owner=%q operation=%q commit=%q",
			ownerProfileID,
			actualOperationKind,
			finalCommitID,
		)
	}
	var admittedProfileID string
	var actualJobKind string
	var workerKind string
	if err := db.QueryRow(`
SELECT extension_owner_profile_id, job_kind, handler_name
  FROM jobs
 WHERE job_id::text = $1
`, jobID).Scan(&admittedProfileID, &actualJobKind, &workerKind); err != nil {
		t.Fatalf("read Reference Pack job binding for job %s: %v", jobID, err)
	}
	if admittedProfileID != reference_data.ProfileID || actualJobKind != jobKind || workerKind != reference_data.LifecycleWorkerKind {
		t.Fatalf(
			"unexpected Reference Pack job binding: owner=%q job=%q worker=%q",
			admittedProfileID,
			actualJobKind,
			workerKind,
		)
	}
}

func requirePackMetadata(t testing.TB, db *sql.DB, packKey string, packVersion string) {
	t.Helper()
	var manifestSHA, payloadSHA string
	var envelope []byte
	if err := db.QueryRow(`SELECT v.manifest_sha256,v.payload_sha256,e.canonical_envelope FROM reference_pack_candidates c JOIN reference_pack_versions v USING(pack_key,pack_version) JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE c.pack_key=$1 AND c.pack_version=$2`, packKey, packVersion).Scan(&manifestSHA, &payloadSHA, &envelope); err != nil {
		t.Fatal(err)
	}
	if len(manifestSHA) != 64 || len(payloadSHA) != 64 || !bytes.Contains(envelope, []byte(`"trust_proposal":{`)) {
		t.Fatal("incomplete successful envelope")
	}
	if queryCount(t, db, `SELECT count(*) FROM reference_pack_events WHERE pack_key=$1 AND pack_version=$2`, packKey, packVersion) == 0 {
		t.Fatal("missing operation attestation")
	}

}

func exerciseCoreWorkflowDuringOptionalPackDegradation(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, adminID string, suffix string) {
	t.Helper()
	incident := scenariotest.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-rp-degrade-" + suffix + "-incident",
		"incident_key":  "IR-RP-DEGRADE-" + strings.ToUpper(suffix),
		"title":         "Reference Pack degradation " + suffix,
	})
	incidentID := incident["incident_id"].(string)

	hostResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+entitycontract.HostsViewSchemaID+"/rows", map[string]any{
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

	evidenceResp := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+evidence.ViewSchemaID+"/rows", map[string]any{
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
	putObject(t, harness.Server.HTTP.URL, blobData["upload_target"].(map[string]any), payload, login)
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

func queryViewRow(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, viewSchemaID string, recordID string) map[string]any {
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

func viewSchemaIDs(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult) []string {
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

func overwriteStoredBundle(t testing.TB, harness *appsupport.ServerHarness, packKey string, packVersion string, bundle []byte) {
	t.Helper()
	path := storedBundlePath(t, harness, packKey, packVersion)
	if err := os.WriteFile(path, bundle, 0o600); err != nil {
		t.Fatalf("overwrite stored bundle: %v", err)
	}
}

func removeStoredBundle(t testing.TB, harness *appsupport.ServerHarness, packKey string, packVersion string) {
	t.Helper()
	path := storedBundlePath(t, harness, packKey, packVersion)
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove stored bundle: %v", err)
	}
}

func storedBundlePath(t testing.TB, harness *appsupport.ServerHarness, packKey string, packVersion string) string {
	t.Helper()
	var rawReference string
	if err := harness.DB.QueryRow(`SELECT o.storage_ref FROM reference_pack_object_refs r JOIN reference_pack_objects o USING(object_id) WHERE r.owner_kind='version' AND r.owner_id=$1 AND r.logical_path='payload/entries.ndjson'`, fmt.Sprintf("rpver_%x", sha256.Sum256([]byte(packKey+"\x00"+packVersion)))).Scan(&rawReference); err != nil {
		t.Fatalf("query stored bundle reference: %v", err)
	}
	reference, err := reference_data.ParseStorageRef(rawReference)
	if err != nil {
		t.Fatalf("parse stored bundle reference: %v", err)
	}
	return filepath.Join(
		harness.Server.Config.Roots.ReferencePackStorage.Path,
		filepath.FromSlash(reference.String()),
	)
}

func requireReverifyFailure(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, packKey string, packVersion string, clientTxnID string, reasonCode string) {
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

func putObject(t testing.TB, baseURL string, target map[string]any, payload []byte, login flowtest.LoginResult) {
	t.Helper()
	href := target["href"].(string)
	if strings.HasPrefix(href, "/") {
		href = baseURL + href
	}
	req, err := http.NewRequest(target["method"].(string), href, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create object upload request: %v", err)
	}
	for name, value := range target["headers"].(map[string]any) {
		req.Header.Set(name, value.(string))
	}
	for _, option := range csrfOptions(login) {
		option(req)
	}
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
