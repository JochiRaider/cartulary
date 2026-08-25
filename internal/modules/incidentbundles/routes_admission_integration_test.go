package incidentbundles_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestExportJobIdempotencyAndDescriptor_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	database := runtime.PrepareServerDatabase(t, "extension_profile-incident-bundle-export")
	harness := runtime.StartServerWithDatabase(t, "extension_profile-incident-bundle-export", database)
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-source",
		"incident_key":  "BUNDLE-EXPORT",
		"title":         "Incident bundle export",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-row",
		"timeline.activity_synopsis_text": "Portable event",
	})

	first := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-bundle",
		"optional_sections":     []string{"snapshots", "reference_packs", "snapshots"},
		"reference_pack_mode":   "refs_only",
		"required_capabilities": []string{},
	})
	job := httptestx.RequireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	replay := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-bundle",
		"optional_sections":     []string{"reference_packs", "snapshots"},
		"reference_pack_mode":   "refs_only",
		"required_capabilities": []string{},
	})
	replayedJob := httptestx.RequireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != job["job_id"] {
		t.Fatalf("exact replay returned different job: first=%v replay=%v", job["job_id"], replayedJob["job_id"])
	}
	var storedStatus int
	var storedResponse []byte
	if err := harness.DB.QueryRow(`
SELECT status_code, response_json
  FROM route_idempotency
 WHERE route_key = 'incident_bundles.export'
   AND actor_user_id = $1
   AND scope_key = $2
   AND client_txn_id = 'txn-export-bundle'
`, adminID, incidentID).Scan(&storedStatus, &storedResponse); err != nil {
		t.Fatalf("read shared Auth idempotency record: %v", err)
	}
	var storedJob map[string]any
	if err := json.Unmarshal(storedResponse, &storedJob); err != nil || storedStatus != http.StatusAccepted || storedJob["job_id"] != job["job_id"] {
		t.Fatalf("stored idempotency response changed: status=%d response=%s err=%v", storedStatus, storedResponse, err)
	}
	divergent := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":         incidentID,
		"client_txn_id":       "txn-export-bundle",
		"reference_pack_mode": "embedded",
	})
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	if countRows(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.export' AND actor_user_id = $1 AND scope_key = $2 AND client_txn_id = 'txn-export-bundle'`, adminID, incidentID) != 1 {
		t.Fatal("conflicting replay changed the shared idempotency row count")
	}
	for index, capabilities := range [][]string{{"snapshots"}, {"unknown/hostile capability"}} {
		beforeJobs := countRows(t, harness.DB, `SELECT count(*) FROM jobs`)
		beforePayloads := countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_job_payloads`)
		beforeIdempotency := countRows(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.export'`)
		unsupportedRequired := postExport(t, harness.Server, admin, map[string]any{
			"incident_id":           incidentID,
			"client_txn_id":         "txn-export-required-capability-" + fmt.Sprint(index),
			"required_capabilities": capabilities,
		})
		body := httptestx.RequireErrorEnvelope(t, unsupportedRequired, http.StatusConflict, "extension_capability_not_supported")
		details := httptestx.RequireErrorDetails(t, body)
		if len(details) != 1 || details["profile_id"] != incidentPortabilityProfileID || strings.Contains(string(jsonRaw(t, details)), capabilities[0]) {
			t.Fatalf("required capability details were not exact and value-free: %#v", details)
		}
		if countRows(t, harness.DB, `SELECT count(*) FROM jobs`) != beforeJobs ||
			countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_job_payloads`) != beforePayloads ||
			countRows(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.export'`) != beforeIdempotency {
			t.Fatal("capability rejection created a job, payload, or idempotency row")
		}
	}
	rejectedReplay := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-required-capability-0",
		"required_capabilities": []string{"different-value-same-terminal-result"},
	})
	replayBody := httptestx.RequireErrorEnvelope(t, rejectedReplay, http.StatusConflict, "extension_capability_not_supported")
	if details := httptestx.RequireErrorDetails(t, replayBody); len(details) != 1 || details["profile_id"] != incidentPortabilityProfileID {
		t.Fatalf("replayed capability rejection details mismatch: %#v", details)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.export' AND client_txn_id = 'txn-export-required-capability-0'`) != 0 {
		t.Fatal("replayed capability rejection created idempotency state")
	}

	terminal := waitJob(t, harness.Server, admin, job["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentBundleExportedCode {
		t.Fatalf("unexpected export summary: %#v", summary)
	}
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 || refs[0].(map[string]any)["kind"] != "incident_bundle" {
		t.Fatalf("export summary must contain one incident_bundle ref: %#v", refs)
	}
	requireIncidentPortabilityProof(
		t,
		harness.DB,
		job["job_id"].(string),
		"incident_portability.export",
	)
	descriptorRoute := refs[0].(map[string]any)["route"].(string)
	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+descriptorRoute, nil, httptestx.WithCookies(admin.SessionCookie))
	descriptor := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if descriptor["history_mode"] != "full" || descriptor["blob_mode"] != "full" || descriptor["manifest_sha256"] == "" {
		t.Fatalf("descriptor missing fixed modes or manifest hash: %#v", descriptor)
	}

	t.Run("queued export recovers through the named runner", func(t *testing.T) {
		testQueuedExportRecoveryThroughNamedRunner(
			t,
			harness,
			admin,
			uuid.MustParse(adminID),
			uuid.MustParse(incidentID),
			func() *appsupport.ServerHarness {
				return runtime.StartServerWithDatabaseAndObjectStore(
					t,
					"incident-bundle-named-runner-recovery",
					database,
					harness.ObjectStore,
				)
			},
		)
	})
}

func testQueuedExportRecoveryThroughNamedRunner(
	t *testing.T,
	first *appsupport.ServerHarness,
	admin flowtest.LoginResult,
	adminID uuid.UUID,
	incidentID uuid.UUID,
	restart func() *appsupport.ServerHarness,
) {
	ctx := context.Background()

	// Stop the first process before admitting the job so no in-process Notify
	// hint can participate. The second process must discover the durable queued
	// job during the named runner's activation scan.
	first.Server.Close()
	normalizedRequest, err := json.Marshal(map[string]any{
		"client_txn_id":         "txn-incident-bundle-recovery",
		"incident_id":           incidentID.String(),
		"optional_sections":     []string{},
		"reference_pack_mode":   "refs_only",
		"required_capabilities": []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	idempotencyKey := jobs.NewRouteIdempotencyKey(
		"incident_bundles.export",
		adminID,
		incidentID.String(),
		"txn-incident-bundle-recovery",
	)
	scope := jobs.Scope{Kind: jobs.ScopeKindIncident, IncidentID: &incidentID}
	admission, err := jobs.NewExtensionJobAdmission(
		incidentPortabilityProfileID,
		idempotencyKey,
		scope,
		normalizedRequest,
	)
	if err != nil {
		t.Fatalf("construct Incident Bundle recovery admission: %v", err)
	}
	total := 1
	queued, err := first.Jobs.Create(ctx, jobs.EnqueueParams{
		JobKind:           incidentBundleExportJobKind,
		Scope:             scope,
		SubmittedByUserID: adminID,
		AuthPolicy:        jobs.AuthPolicyDeploymentAdminIncidentMembership,
		Cancelable:        true,
		Progress:          jobs.Progress{Completed: 0, Total: &total},
		Extension:         admission,
	})
	if err != nil {
		t.Fatalf("create queued Incident Bundle recovery job: %v", err)
	}
	jobID := uuid.MustParse(queued.JobID)
	tx, err := first.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
INSERT INTO incident_bundle_job_payloads (
    job_id, job_kind, actor_user_id, incident_id, request_json, created_at, updated_at
)
VALUES ($1, 'export', $2, $3, $4, $5, $5)
`, jobID, adminID, incidentID, normalizedRequest, queued.SubmittedAt); err != nil {
		t.Fatalf("insert queued Incident Bundle recovery payload: %v", err)
	}
	requestDigest := sha256.Sum256(normalizedRequest)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey:    idempotencyKey.RouteKey,
		ActorUserID: idempotencyKey.ActorUserID,
		ScopeKey:    idempotencyKey.ScopeKey,
		ClientTxnID: idempotencyKey.ClientTxnID,
	}, nil, requestDigest[:], http.StatusAccepted, queued); err != nil {
		t.Fatalf("insert queued Incident Bundle recovery idempotency: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	second := restart()
	terminal := waitJob(t, second.Server, admin, queued.JobID)
	if summary := terminal["result_summary"].(map[string]any); summary["code"] != incidentBundleExportedCode {
		t.Fatalf("recovered export summary = %#v", summary)
	}
	if countRows(t, second.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, jobID) != 1 {
		t.Fatal("named recovery produced missing or duplicate Incident Bundle descriptors")
	}
	if countRows(t, second.DB, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, jobID) != 1 {
		t.Fatal("named recovery produced missing or duplicate terminal commit proofs")
	}
}

func TestExportJobAuthorizationReDerivesIncidentMembership_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-export-auth")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-source",
		"incident_key":  "BUNDLE-EXPORT-AUTH",
		"title":         "Incident bundle export auth",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-export-auth-row",
		"timeline.activity_synopsis_text": "Portable event for auth",
	})

	submitterPassword := "BundleSubmitterPassphrase11!"
	memberAdminPassword := "BundleMemberAdminPassphrase11!"
	memberOnlyPassword := "BundleMemberOnlyPassphrase11!"
	nonmemberAdminPassword := "BundleNonmemberAdminPassphrase11!"
	submitterUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-bundle-submitter@example.test", "ExtensionProfile Bundle Submitter", submitterPassword, false, true, true)
	memberAdminUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-bundle-member-admin@example.test", "ExtensionProfile Bundle Member Admin", memberAdminPassword, false, true, true)
	memberOnlyUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-bundle-member-only@example.test", "ExtensionProfile Bundle Member Only", memberOnlyPassword, false, false, true)
	nonmemberAdminUser := flowtest.SeedLocalUserRecord(t, harness.DB, "extension_profile-bundle-nonmember-admin@example.test", "ExtensionProfile Bundle Nonmember Admin", nonmemberAdminPassword, false, true, true)
	submitterCookies, submitterCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, submitterUser.Email, submitterPassword, nil)
	memberAdminCookies, memberAdminCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, memberAdminUser.Email, memberAdminPassword, nil)
	memberOnlyCookies, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, memberOnlyUser.Email, memberOnlyPassword, nil)
	nonmemberAdminCookies, nonmemberAdminCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, nonmemberAdminUser.Email, nonmemberAdminPassword, nil)
	submitterLogin := flowtest.LoginResult{SessionCookie: submitterCookies, CSRFCookie: submitterCSRF}
	memberAdminLogin := flowtest.LoginResult{SessionCookie: memberAdminCookies, CSRFCookie: memberAdminCSRF}
	nonmemberAdminLogin := flowtest.LoginResult{SessionCookie: nonmemberAdminCookies, CSRFCookie: nonmemberAdminCSRF}

	scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-submitter-membership",
		"user_id":       submitterUser.ID.String(),
		"role":          "viewer",
	})
	memberAdminMembership := scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-member-admin-membership",
		"user_id":       memberAdminUser.ID.String(),
		"role":          "viewer",
	})
	scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-member-only-membership",
		"user_id":       memberOnlyUser.ID.String(),
		"role":          "admin",
	})
	capabilityProbe := postExport(t, harness.Server, nonmemberAdminLogin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-capability-nonmember",
		"required_capabilities": []string{"membership-oracle-probe"},
	})
	capabilityProbeBody := httptestx.RequireErrorEnvelope(t, capabilityProbe, http.StatusNotFound, "incident_bundle_not_found")
	if strings.Contains(string(jsonRaw(t, capabilityProbeBody)), "membership-oracle-probe") {
		t.Fatal("membership-first capability rejection echoed the submitted value")
	}

	exportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, submitterLogin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-auth-blocked",
	}), http.StatusAccepted)["data"].(map[string]any)
	jobID := exportJob["job_id"].(string)

	submitterRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(submitterCookies))
	httptestx.RequireSuccessEnvelope(t, submitterRead, http.StatusOK)
	memberAdminRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(memberAdminCookies))
	httptestx.RequireSuccessEnvelope(t, memberAdminRead, http.StatusOK)
	memberOnlyRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(memberOnlyCookies))
	httptestx.RequireErrorEnvelope(t, memberOnlyRead, http.StatusNotFound, "job_not_found")
	nonmemberAdminRead := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(nonmemberAdminCookies))
	httptestx.RequireErrorEnvelope(t, nonmemberAdminRead, http.StatusNotFound, "job_not_found")
	nonmemberAdminCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-nonmember-cancel",
	}, httptestx.WithCookies(nonmemberAdminCookies, nonmemberAdminCSRF), httptestx.WithHeader(authn.CSRFHeaderName, nonmemberAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, nonmemberAdminCancel, http.StatusNotFound, "job_not_found")
	memberViewerCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-member-viewer-cancel",
	}, httptestx.WithCookies(memberAdminCookies, memberAdminCSRF), httptestx.WithHeader(authn.CSRFHeaderName, memberAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, memberViewerCancel, http.StatusForbidden, "authorization_denied")

	scenariotest.PatchMembership(t, harness.Server, admin, incidentID, memberAdminUser.ID.String(), map[string]any{
		"base_membership_version": memberAdminMembership["membership_version"],
		"role":                    "admin",
	})
	memberAdminCancel := httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-member-admin-cancel",
	}, httptestx.WithCookies(memberAdminCookies, memberAdminCSRF), httptestx.WithHeader(authn.CSRFHeaderName, memberAdminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, memberAdminCancel, http.StatusOK)

	terminal := waitJobWithStatus(t, harness.Server, memberAdminLogin, jobID, "canceled")
	if terminal["status"] != "canceled" {
		t.Fatalf("export job must stop at canceled after authorized cancel: %#v", terminal)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM extension_job_cancellation_observations WHERE job_id = $1`, jobID) != 1 {
		t.Fatal("accepted Incident Portability cancellation must retain one observation")
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM extension_job_commit_proofs WHERE job_id = $1`, jobID) != 0 {
		t.Fatal("canceled Incident Portability job must not publish a success proof")
	}
}

func TestImportEnvelopeIdempotencyAndImportedIncidentOpen_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-target")
	sourceAdmin, sourceAdminID := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, targetAdminID := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-import-source",
		"incident_key":  "BUNDLE-IMPORT",
		"title":         "Incident bundle import",
	})
	incidentID := incident["incident_id"].(string)
	row := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-import-row",
		"timeline.activity_synopsis_text": "Imported portable event",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	seededState := seedIncidentBundlePortableState(t, sourceHarness, incidentID, recordID, sourceAdminID)
	sourceViewerID := flowtest.SeedLocalUserFlags(t, sourceHarness.DB, "extension_profile-import-source-viewer@example.test", "Enterprise integration Import Source Viewer", "ExtensionProfileImportViewer1!", false, false, true)
	scenariotest.CreateMembership(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-source-viewer",
		"user_id":       sourceViewerID,
		"role":          "viewer",
	})
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incidents`) != 0 {
		t.Fatalf("target deployment must start empty before incident bundle import")
	}
	exportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-for-import",
	}), http.StatusAccepted)["data"].(map[string]any)
	exportTerminal := waitJob(t, sourceHarness.Server, sourceAdmin, exportJob["job_id"].(string))
	exportRef := exportTerminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	bundleID := exportRef["id"].(string)
	var bundleStorageRef string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id = $1`, bundleID).Scan(&bundleStorageRef); err != nil {
		t.Fatalf("query exported bundle reference: %v", err)
	}
	bundleBytes, err := os.ReadFile(exportedBundleTestPath(t, sourceHarness.Server, bundleStorageRef))
	if err != nil {
		t.Fatalf("read exported bundle: %v", err)
	}
	secondExportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-for-import-determinism",
	}), http.StatusAccepted)["data"].(map[string]any)
	secondExportTerminal := waitJob(t, sourceHarness.Server, sourceAdmin, secondExportJob["job_id"].(string))
	secondExportRef := secondExportTerminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	var secondBundleStorageRef string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id = $1`, secondExportRef["id"].(string)).Scan(&secondBundleStorageRef); err != nil {
		t.Fatalf("query second exported bundle reference: %v", err)
	}
	secondBundleBytes, err := os.ReadFile(exportedBundleTestPath(t, sourceHarness.Server, secondBundleStorageRef))
	if err != nil {
		t.Fatalf("read second exported bundle: %v", err)
	}
	for _, memberPath := range []string{
		"integrity/checksums.sha256",
		"data/record_tags.ndjson",
		"data/timeline_time_profiles.ndjson",
		"data/parties.ndjson",
		"data/entity_preserved_identifiers.ndjson",
		"data/artifact_findings.ndjson",
		"data/artifact_investigative_queries.ndjson",
		"data/artifact_forensic_keywords.ndjson",
		"data/handoff_risk_refs.ndjson",
		"data/change_sets.ndjson",
		"data/record_revisions.ndjson",
	} {
		if !bytes.Equal(zipMemberBytes(t, bundleBytes, memberPath), zipMemberBytes(t, secondBundleBytes, memberPath)) {
			t.Fatalf("source-state export member %s changed across identical DB exports", memberPath)
		}
	}

	corruptBytes := corruptZipMember(t, bundleBytes, "data/records.ndjson")
	failed := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-bundle-corrupt"}`, corruptBytes, "bundle-corrupt.zip")
	failedJob := httptestx.RequireSuccessEnvelope(t, failed, http.StatusAccepted)["data"].(map[string]any)
	failedTerminal := waitFailedJob(t, targetHarness.Server, targetAdmin, failedJob["job_id"].(string))
	errorSummary := failedTerminal["error_summary"].(map[string]any)
	if errorSummary["code"] != "incident_bundle_import_rejected" {
		t.Fatalf("corrupt import must fail closed with incident_bundle_import_rejected: %#v", errorSummary)
	}
	details := errorSummary["details"].(map[string]any)
	if details["reason_code"] != "checksum_mismatch" {
		t.Fatalf("corrupt import reason mismatch: %#v", details)
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID) != 0 {
		t.Fatalf("failed import must not make incident visible")
	}
	var failedStagingRef string
	if err := targetHarness.DB.QueryRow(`SELECT bundle_staging_ref FROM incident_bundle_job_payloads WHERE job_id = $1`, failedJob["job_id"].(string)).Scan(&failedStagingRef); err != nil {
		t.Fatalf("query failed import staging reference: %v", err)
	}
	failedStagingPath := stagedBundleTestPath(t, targetHarness.Server, failedStagingRef)
	if _, err := os.Stat(failedStagingPath); !os.IsNotExist(err) {
		t.Fatalf("failed import staging reference must be cleaned up, stat err=%v ref=%s", err, failedStagingRef)
	}

	first := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-bundle"}`, bundleBytes, "bundle.zip")
	importJob := httptestx.RequireSuccessEnvelope(t, first, http.StatusAccepted)["data"].(map[string]any)
	replay := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-bundle"}`, bundleBytes, "bundle-renamed.zip")
	replayedJob := httptestx.RequireSuccessEnvelope(t, replay, http.StatusAccepted)["data"].(map[string]any)
	if replayedJob["job_id"] != importJob["job_id"] {
		t.Fatalf("exact import replay returned different job: first=%v replay=%v", importJob["job_id"], replayedJob["job_id"])
	}
	divergent := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-bundle"}`, append(bundleBytes, 0), "bundle.zip")
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")

	terminal := waitJob(t, targetHarness.Server, targetAdmin, importJob["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentBundleImportedCode {
		t.Fatalf("unexpected import summary: %#v", summary)
	}
	ref := summary["resource_refs"].([]any)[0].(map[string]any)
	if ref["kind"] != "incident" || ref["id"] != incidentID {
		t.Fatalf("import summary must reference imported incident: %#v", ref)
	}
	requireIncidentPortabilityProof(
		t,
		targetHarness.DB,
		importJob["job_id"].(string),
		"incident_portability.import",
	)
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID, "change_set count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_revisions rr JOIN records r ON r.record_id = rr.record_id WHERE r.incident_id = $1`, incidentID, "revision count")
	if got := collaborationsupport.CountIntents(t, targetHarness.DB, collaborationsupport.IntentSelector{
		IncidentID: incidentID, EventFamily: "record_changed",
	}); got != 0 {
		t.Fatalf("historical revision import emitted %d live record_changed intents, want 0", got)
	}
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_links WHERE incident_id = $1`, incidentID, "record-link count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_tags WHERE incident_id = $1`, incidentID, "record-tag attachment count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM evidence_custody_events WHERE incident_id = $1`, incidentID, "evidence custody count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM entity_mentions em JOIN records r ON r.record_id = em.source_record_id WHERE r.incident_id = $1`, incidentID, "entity mention count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id = $1`, incidentID, "timeline time conversion profile count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM parties WHERE incident_id = $1`, incidentID, "party count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM identities WHERE incident_id = $1`, incidentID, "identity count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM assessments WHERE incident_id = $1`, incidentID, "assessment count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM task_requests WHERE incident_id = $1`, incidentID, "task-request count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM decisions WHERE incident_id = $1`, incidentID, "decision count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM entity_preserved_identifiers WHERE incident_id = $1`, incidentID, "entity preserved identifier count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM artifact_findings WHERE incident_id = $1`, incidentID, "artifact finding count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM artifact_investigative_queries WHERE incident_id = $1`, incidentID, "artifact investigative query count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM artifact_forensic_keywords WHERE incident_id = $1`, incidentID, "artifact forensic keyword count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM handoff_risk_refs WHERE incident_id = $1`, incidentID, "handoff risk ref count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM saved_views WHERE incident_id = $1`, incidentID, "saved view count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM indicator_observations WHERE incident_id = $1`, incidentID, "indicator observation count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM indicator_state_intervals WHERE incident_id = $1`, incidentID, "indicator lifecycle interval count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM records WHERE incident_id = $1 AND row_version = 2`, incidentID, "row_version=2 record count")
	if got := stringScalar(t, targetHarness.DB, `SELECT local_label FROM timeline_time_conversion_profiles WHERE incident_id = $1`, incidentID); got != "America/New_York" {
		t.Fatalf("imported time conversion profile changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT display_name FROM parties WHERE record_id = $1`, seededState.PartyRecordID); got != "Portable Party" {
		t.Fatalf("imported party display name changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT normalized_value FROM entity_preserved_identifiers WHERE record_id = $1`, seededState.HistoryHostRecordID); got != "portable-host" {
		t.Fatalf("imported preserved identifier changed: got %q", got)
	}
	if countRows(t, targetHarness.DB, `SELECT COUNT(*) FROM indicator_observations WHERE incident_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL`, uuid.MustParse(incidentID)) != 1 {
		t.Fatal("import did not retain the Indicator observation tombstone")
	}
	if countRows(t, targetHarness.DB, `SELECT COUNT(*) FROM indicator_state_intervals WHERE incident_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL`, uuid.MustParse(incidentID)) != 1 {
		t.Fatal("import did not retain the Indicator lifecycle tombstone")
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT finding_statement FROM artifact_grid_projection WHERE record_id = $1`, seededState.FindingArtifactRecordID); got != "Portable finding statement" {
		t.Fatalf("imported finding projection changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT investigative_query_query_text FROM artifact_grid_projection WHERE record_id = $1`, seededState.QueryArtifactRecordID); got != "SecurityEvent | take 10" {
		t.Fatalf("imported investigative query projection changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT forensic_keyword_pattern FROM artifact_grid_projection WHERE record_id = $1`, seededState.KeywordArtifactRecordID); got != "PortableKeyword" {
		t.Fatalf("imported forensic keyword projection changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT risk_ref_text FROM handoff_risk_refs WHERE handoff_record_id = $1`, seededState.HandoffArtifactRecordID); got != "Portable Risk" {
		t.Fatalf("imported handoff risk ref changed: got %q", got)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT display_name FROM saved_views WHERE saved_view_id = $1 AND owner_user_id = $2`, seededState.SavedViewID, targetAdminID); got != "Portable saved view" {
		t.Fatalf("imported saved view owner/display changed: got %q", got)
	}
	openResp := httptestx.DoJSON(t, http.MethodGet, targetHarness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup", nil, httptestx.WithCookies(targetAdmin.SessionCookie))
	httptestx.RequireSuccessEnvelope(t, openResp, http.StatusOK)
	var projectionCount int
	if err := targetHarness.DB.QueryRow(`SELECT count(*) FROM timeline_grid_projection WHERE record_id = $1`, recordID).Scan(&projectionCount); err != nil {
		t.Fatalf("query imported projection: %v", err)
	}
	if projectionCount != 1 {
		t.Fatalf("imported timeline projection missing, count=%d", projectionCount)
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1`, incidentID) != 1 {
		t.Fatalf("import should create only local importer membership")
	}
	finalization := snapshotImportFinalizationSideEffects(t, targetHarness.DB, incidentID, targetAdminID)
	if finalization.MembershipRows != 1 || finalization.DefaultPreferenceRows != 1 || finalization.UserPreferenceRows != 1 || finalization.MembershipAuditRows != 1 || finalization.MembershipProjectionRows != 1 {
		t.Fatalf("import finalization side effects mismatch: %#v", finalization)
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, incidentID, sourceViewerID) != 0 {
		t.Fatalf("import must not recreate source deployment-local membership")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_bundle_imported_actors WHERE incident_id = $1 AND source_actor_id = $2 AND local_user_id IS NULL`, incidentID, sourceAdminID) == 0 {
		t.Fatalf("import must preserve historical attribution as inert actor descriptors")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id = $1 AND source_table = 'change_sets' AND source_column = 'actor_user_id' AND source_actor_id = $2 AND local_user_id = $3`, incidentID, sourceAdminID, targetAdminID) == 0 {
		t.Fatalf("import must retain source actor attribution sidecars")
	}
	reexportedBundle := exportBundleBytes(
		t,
		targetHarness,
		targetAdmin,
		incidentID,
		"txn-reexport-imported-record-attribution",
	)
	if !bytes.Equal(
		zipMemberBytes(t, bundleBytes, "data/records.ndjson"),
		zipMemberBytes(t, reexportedBundle, "data/records.ndjson"),
	) {
		t.Fatal("re-export changed portable Records source attribution or envelope fields")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM record_tags WHERE incident_id = $1 AND record_id = $2 AND normalized_tag_name = 'extension_profile-portability'`, incidentID, recordID) != 1 {
		t.Fatalf("import must preserve record tag attachments")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM evidence_custody_events WHERE incident_id = $1 AND evidence_record_id = $2`, incidentID, seededState.EvidenceRecordID) != 1 {
		t.Fatalf("import must preserve evidence custody events")
	}
	var importedStorageKey, importedSHA string
	if err := targetHarness.DB.QueryRow(`SELECT storage_key, observed_sha256_hex FROM object_blobs WHERE incident_id = $1 AND object_blob_id = $2`, incidentID, seededState.ObjectBlobID).Scan(&importedStorageKey, &importedSHA); err != nil {
		t.Fatalf("query imported object blob: %v", err)
	}
	if importedSHA != seededState.BlobSHA {
		t.Fatalf("imported blob hash mismatch: got %s want %s", importedSHA, seededState.BlobSHA)
	}
	wantStorageKey := "incidents/" + incidentID + "/object-blobs/" + seededState.ObjectBlobID
	if importedStorageKey != wantStorageKey {
		t.Fatalf("imported blob must use target-owned storage key, got %s want %s", importedStorageKey, wantStorageKey)
	}
	rc, _, err := targetHarness.ObjectStore.ReadObject(context.Background(), importedStorageKey, objectstore.ReadOptions{})
	if err != nil {
		t.Fatalf("read imported blob: %v", err)
	}
	importedBytes, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil {
		t.Fatalf("read imported blob bytes: %v", err)
	}
	if !bytes.Equal(importedBytes, seededState.BlobBytes) {
		t.Fatalf("imported blob bytes changed: got %q want %q", importedBytes, seededState.BlobBytes)
	}
	historyResp := httptestx.DoJSON(t, http.MethodGet, targetHarness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/history", nil, httptestx.WithCookies(targetAdmin.SessionCookie))
	historyData := httptestx.RequireSuccessEnvelope(t, historyResp, http.StatusOK)["data"].(map[string]any)
	historyItems := historyData["items"].([]any)
	foundSourceActor := false
	for _, raw := range historyItems {
		item := raw.(map[string]any)
		if item["source_actor_id"] == sourceAdminID {
			foundSourceActor = true
			break
		}
	}
	if !foundSourceActor {
		t.Fatalf("imported history must expose source_actor_id=%s, got %#v", sourceAdminID, historyItems)
	}
	firstImportedHistory := getRecordHistoryItems(t, targetHarness.Server, targetAdmin, seededState.HistoryHostRecordID)
	secondImportedHistory := getRecordHistoryItems(t, targetHarness.Server, targetAdmin, seededState.HistoryHostRecordID)
	reversibleFirst := requireHistoryItemForChangeSet(t, firstImportedHistory, seededState.ReversibleChangeSetID)
	reversibleSecond := requireHistoryItemForChangeSet(t, secondImportedHistory, seededState.ReversibleChangeSetID)
	nonreversibleFirst := requireHistoryItemForChangeSet(t, firstImportedHistory, seededState.NonReversibleChangeSetID)
	nonreversibleSecond := requireHistoryItemForChangeSet(t, secondImportedHistory, seededState.NonReversibleChangeSetID)
	reversibleRef, _ := reversibleFirst["history_entry_ref"].(string)
	reversibleSecondRef, _ := reversibleSecond["history_entry_ref"].(string)
	nonreversibleRef, _ := nonreversibleFirst["history_entry_ref"].(string)
	nonreversibleSecondRef, _ := nonreversibleSecond["history_entry_ref"].(string)
	if reversibleRef == "" || reversibleRef != reversibleSecondRef {
		t.Fatalf("imported reversible history selector must be stable across reads: first=%#v second=%#v", reversibleFirst, reversibleSecond)
	}
	if nonreversibleRef == "" || nonreversibleRef != nonreversibleSecondRef {
		t.Fatalf("imported nonreversible history selector must be stable across reads: first=%#v second=%#v", nonreversibleFirst, nonreversibleSecond)
	}
	if actions, _ := nonreversibleFirst["available_rollback_actions"].([]any); len(actions) != 2 || nonreversibleFirst["reversible"] != true {
		t.Fatalf("canonical imported record-tag create history should remain reversible: %#v", nonreversibleFirst)
	}
	reversibleRollback := postRollback(t, targetHarness.Server, targetAdmin, seededState.HistoryHostRecordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-imported-reversible-rollback",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": reversibleRef},
	})
	if reversibleRollback.StatusCode != http.StatusOK {
		t.Fatalf("imported reversible rollback failed: status=%d body=%#v", reversibleRollback.StatusCode, httptestx.ReadJSONBody(t, reversibleRollback))
	}
	rollbackData := httptestx.RequireSuccessEnvelope(t, reversibleRollback, http.StatusOK)["data"].(map[string]any)
	if rollbackData["rollback_change_set_id"] == "" || rollbackData["row_version"] != float64(3) {
		t.Fatalf("imported reversible rollback returned unexpected payload: %#v", rollbackData)
	}
	if got := stringScalar(t, targetHarness.DB, `SELECT display_name FROM hosts WHERE record_id = $1`, seededState.HistoryHostRecordID); got != "portable host before" {
		t.Fatalf("imported rollback did not restore host display_name: got %q", got)
	}

	replayAfterTerminal := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-bundle"}`, bundleBytes, "bundle-replay.zip")
	replayedTerminalJob := httptestx.RequireSuccessEnvelope(t, replayAfterTerminal, http.StatusAccepted)["data"].(map[string]any)
	if replayedTerminalJob["job_id"] != importJob["job_id"] {
		t.Fatalf("terminal import replay returned different job: first=%v replay=%v", importJob["job_id"], replayedTerminalJob["job_id"])
	}
	if afterReplay := snapshotImportFinalizationSideEffects(t, targetHarness.DB, incidentID, targetAdminID); !afterReplay.equal(finalization) {
		t.Fatalf("terminal import replay duplicated finalization side effects: before=%#v after=%#v", finalization, afterReplay)
	}
}
