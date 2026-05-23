package reporting_test

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase11_I_11_REPORTING_01_SnapshotReplayAndReleaseProvenanceAreStable(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-provenance")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-incident",
		"incident_key":  "IR-REPORTING-01",
		"title":         "Reporting Incident",
		"description":   strings.Repeat("external recipient narrative ", 12),
		"severity":      "high",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)

	createSnapshot := func() *http.Response {
		t.Helper()
		return phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/snapshots",
			map[string]any{
				"incident_id":   incidentID,
				"client_txn_id": "txn-reporting-snapshot",
			},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
	}
	firstSnapshot := createSnapshot()
	httptestx.RequireSuccessEnvelope(t, firstSnapshot, http.StatusAccepted)
	snapshotID := requireOnlySnapshotID(t, harness.DB)
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if snapshot["export_model_sha256"] == "" || snapshot["source_change_set_high_watermark"] == "" {
		t.Fatalf("snapshot must expose immutable export model and source boundary provenance: %#v", snapshot)
	}
	initialWatermark := snapshot["source_change_set_high_watermark"]

	phase2test.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replaySnapshot := createSnapshot()
	httptestx.RequireSuccessEnvelope(t, replaySnapshot, http.StatusAccepted)
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("snapshot idempotent replay must not create fresh state after live incident changed, got %d rows", got)
	}
	replayedSnapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if replayedSnapshot["source_change_set_high_watermark"] != initialWatermark {
		t.Fatalf("snapshot replay must preserve original resolved source boundary: before=%v after=%v", initialWatermark, replayedSnapshot["source_change_set_high_watermark"])
	}

	createReleaseResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               "html",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, createReleaseResp, http.StatusAccepted)
	releaseID := requireOnlyReleaseID(t, harness.DB)
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("external release must start pending approval, got %#v", release["release_state"])
	}
	if release["redaction_profile_id"] != reporting.ExternalRedactionProfileID || release["redaction_profile_sha256"] == "" {
		t.Fatalf("release must bind immutable redaction profile identity, got %#v", release)
	}
	if release["output_sha256"] == "" || release["redaction_manifest_sha256"] == "" {
		t.Fatalf("release must expose output and manifest hashes, got %#v", release)
	}
	rendered, manifest := requireReleaseArtifacts(t, harness.DB, releaseID)
	if strings.Contains(rendered, "internal workbook context") {
		t.Fatalf("external output must not package working material: %s", rendered)
	}
	if manifest["profile_sha256"] != release["redaction_profile_sha256"] {
		t.Fatalf("manifest profile digest must match release provenance: manifest=%#v release=%#v", manifest["profile_sha256"], release["redaction_profile_sha256"])
	}
	entries := manifestEntriesByPath(t, manifest)
	internalEntry := entries["/incident/internal_note"]
	if internalEntry["action"] != reporting.ActionDrop || internalEntry["outcome"] != "dropped" {
		t.Fatalf("working material must be dropped with stable manifest outcome, got %#v", internalEntry)
	}
	descriptionEntry := entries["/incident/description"]
	if descriptionEntry["action"] != reporting.ActionTruncate || descriptionEntry["outcome"] != "truncated" {
		t.Fatalf("source evidence must be truncated for external release, got %#v", descriptionEntry)
	}
}

func TestPhase11_I_11_REPORTING_02_ExternalReleaseApprovalPublishAndStateConflicts(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-lifecycle")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	reviewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "report-reviewer@example.test", "Report Reviewer", "ReviewerPass1!", false, false, true)
	reviewerSession, reviewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "report-reviewer@example.test", "ReviewerPass1!")
	reviewerLogin := phase2test.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}

	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-lifecycle-incident",
		"incident_key":  "IR-REPORTING-02",
		"title":         "Reporting Lifecycle Incident",
		"description":   "approval lifecycle source text",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})

	httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-lifecycle-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)
	snapshotID := requireOnlySnapshotID(t, harness.DB)

	httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-lifecycle-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               "html",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)
	releaseID := requireOnlyReleaseID(t, harness.DB)

	reviewerApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-reviewer-approve",
			"reason":        "reviewed for release",
		},
		phase2test.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	reviewerBody := httptestx.RequireSuccessEnvelope(t, reviewerApprove, http.StatusOK)["data"].(map[string]any)
	if reviewerBody["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("reviewer approval alone must leave external release pending, got %#v", reviewerBody["release_state"])
	}

	adminApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-admin-approve",
			"reason":        "approved for publication",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	adminBody := httptestx.RequireSuccessEnvelope(t, adminApprove, http.StatusOK)["data"].(map[string]any)
	if adminBody["release_state"] != reporting.ReleaseStateApproved || adminBody["approved_at"] == nil {
		t.Fatalf("distinct reviewer and admin approvals must approve external release, got %#v", adminBody)
	}

	publish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish",
			"reason":        "publish",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	publishBody := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	if publishBody["release_state"] != reporting.ReleaseStatePublished || publishBody["published_at"] == nil {
		t.Fatalf("approved release must publish synchronously, got %#v", publishBody)
	}

	republish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish-again",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	errorBody := httptestx.RequireErrorEnvelope(t, republish, http.StatusConflict, "release_state_conflict")
	if errorBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "already_published" {
		t.Fatalf("republish must fail with stable already_published reason, got %#v", errorBody)
	}
}

func TestPhase11_I_11_REPORTING_03_BoundaryReplayDefaultsAndActionIdempotency(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-idempotency")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-idempotency-incident",
		"incident_key":  "IR-REPORTING-03",
		"title":         "Reporting Idempotency Incident",
		"description":   "recipient-visible source text",
	})
	incidentID := incident["incident_id"].(string)

	httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-idempotency-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)
	snapshotID := requireOnlySnapshotID(t, harness.DB)
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	watermark := snapshot["source_change_set_high_watermark"].(string)

	phase2test.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replayWithExplicitBoundary := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":                      incidentID,
			"client_txn_id":                    "txn-reporting-idempotency-snapshot",
			"source_change_set_high_watermark": watermark,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, replayWithExplicitBoundary, http.StatusAccepted)
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("explicit original boundary replay must not create a new snapshot, got %d rows", got)
	}

	httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-idempotency-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               "markdown",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)
	releaseID := requireOnlyReleaseID(t, harness.DB)
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_scope"] != reporting.ReleaseScopeInternalDraft || release["release_state"] != reporting.ReleaseStateApproved {
		t.Fatalf("omitted release_scope must default to approved internal draft, got %#v", release)
	}

	publish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        "",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	firstPublish := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	replayPublish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        nil,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	secondPublish := httptestx.RequireSuccessEnvelope(t, replayPublish, http.StatusOK)["data"].(map[string]any)
	if firstPublish["published_at"] != secondPublish["published_at"] || secondPublish["release_state"] != reporting.ReleaseStatePublished {
		t.Fatalf("publish replay with null/empty reason must return original published resource: first=%#v second=%#v", firstPublish, secondPublish)
	}
}

func requireOnlySnapshotID(t testing.TB, db *sql.DB) string {
	t.Helper()
	var snapshotID string
	if err := db.QueryRow(`SELECT snapshot_id::text FROM reporting_snapshots ORDER BY created_at DESC, snapshot_id DESC LIMIT 1`).Scan(&snapshotID); err != nil {
		t.Fatalf("query snapshot id: %v", err)
	}
	return snapshotID
}

func requireOnlyReleaseID(t testing.TB, db *sql.DB) string {
	t.Helper()
	var releaseID string
	if err := db.QueryRow(`SELECT release_id::text FROM reporting_releases ORDER BY created_at DESC, release_id DESC LIMIT 1`).Scan(&releaseID); err != nil {
		t.Fatalf("query release id: %v", err)
	}
	return releaseID
}

func requireSnapshot(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, snapshotID string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		phase2test.WithCookies(actor.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireRelease(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, releaseID string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		phase2test.WithCookies(actor.SessionCookie),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireReleaseArtifacts(t testing.TB, db *sql.DB, releaseID string) (string, map[string]any) {
	t.Helper()
	var rendered string
	var manifestBytes []byte
	if err := db.QueryRow(`
SELECT rendered_output, redaction_manifest_json
  FROM reporting_releases
 WHERE release_id::text = $1
`, releaseID).Scan(&rendered, &manifestBytes); err != nil {
		t.Fatalf("query release artifacts: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode redaction manifest: %v", err)
	}
	return rendered, manifest
}

func manifestEntriesByPath(t testing.TB, manifest map[string]any) map[string]map[string]any {
	t.Helper()
	rows := manifest["entries"].([]any)
	entries := make(map[string]map[string]any, len(rows))
	for _, row := range rows {
		entry := row.(map[string]any)
		entries[entry["path"].(string)] = entry
	}
	return entries
}
