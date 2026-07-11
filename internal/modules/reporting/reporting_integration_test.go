package reporting_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/phase2test"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestPhase11_I_11_REPORTING_01_SnapshotReplayAndReleaseProvenanceAreStable(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-provenance")

	adminLogin, adminUserID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-incident",
		"incident_key":  "IR-REPORTING-01",
		"title":         "Reporting Incident",
		"description":   strings.Repeat("external recipient narrative ", 12),
		"severity":      "high",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)
	fixture := seedReportingWorkbookFixture(t, harness.DB, incidentID, adminUserID)

	createSnapshot := func(clientTxnID string, watermark *string) *http.Response {
		t.Helper()
		body := map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": clientTxnID,
		}
		if watermark != nil {
			body["source_change_set_high_watermark"] = *watermark
		}
		return phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/snapshots",
			body,
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
	}
	firstSnapshot := createSnapshot("txn-reporting-snapshot", nil)
	firstSnapshotJob := httptestx.RequireSuccessEnvelope(t, firstSnapshot, http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, firstSnapshotJob, "snapshot")
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if snapshot["export_model_sha256"] == "" || snapshot["source_change_set_high_watermark"] == "" {
		t.Fatalf("snapshot must expose immutable export model and source boundary provenance: %#v", snapshot)
	}
	initialWatermark := snapshot["source_change_set_high_watermark"].(string)
	if !strings.HasPrefix(initialWatermark, reporting.SourceBoundaryTokenPrefix) {
		t.Fatalf("snapshot source boundary must use v3 token prefix, got %q", initialWatermark)
	}
	requireSnapshotBoundaryJSON(t, harness.DB, snapshotID, initialWatermark, incidentID)
	exportModel := requireSnapshotExportModel(t, harness.DB, snapshotID)
	requireExportModelCoverage(t, exportModel, fixture)

	phase2test.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replaySnapshot := createSnapshot("txn-reporting-snapshot", nil)
	replaySnapshotJob := httptestx.RequireSuccessEnvelope(t, replaySnapshot, http.StatusAccepted)["data"].(map[string]any)
	if replaySnapshotJob["job_id"] != firstSnapshotJob["job_id"] {
		t.Fatalf("snapshot replay must return original job: first=%#v replay=%#v", firstSnapshotJob, replaySnapshotJob)
	}
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("snapshot idempotent replay must not create fresh state after live incident changed, got %d rows", got)
	}
	replayedSnapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	if replayedSnapshot["source_change_set_high_watermark"] != initialWatermark {
		t.Fatalf("snapshot replay must preserve original resolved source boundary: before=%v after=%v", initialWatermark, replayedSnapshot["source_change_set_high_watermark"])
	}
	staleIncidentBoundary := createSnapshot("txn-reporting-snapshot-stale-incident", &initialWatermark)
	httptestx.RequireErrorEnvelope(t, staleIncidentBoundary, http.StatusConflict, "snapshot_source_boundary_conflict")
	afterIncidentSnapshotJob := httptestx.RequireSuccessEnvelope(t, createSnapshot("txn-reporting-snapshot-after-incident", nil), http.StatusAccepted)["data"].(map[string]any)
	afterIncidentSnapshotID := requireSucceededJobResourceID(t, harness, adminLogin, afterIncidentSnapshotJob, "snapshot")
	afterIncidentSnapshot := requireSnapshot(t, harness, adminLogin, afterIncidentSnapshotID)
	afterIncidentWatermark := afterIncidentSnapshot["source_change_set_high_watermark"].(string)
	if afterIncidentWatermark == initialWatermark {
		t.Fatalf("incident metadata mutation must change source boundary token: %q", afterIncidentWatermark)
	}

	createWorkbookNote(t, harness, adminLogin, incidentID, "txn-reporting-boundary-note")
	afterWorkbookSnapshotJob := httptestx.RequireSuccessEnvelope(t, createSnapshot("txn-reporting-snapshot-after-workbook", nil), http.StatusAccepted)["data"].(map[string]any)
	afterWorkbookSnapshotID := requireSucceededJobResourceID(t, harness, adminLogin, afterWorkbookSnapshotJob, "snapshot")
	afterWorkbookSnapshot := requireSnapshot(t, harness, adminLogin, afterWorkbookSnapshotID)
	afterWorkbookWatermark := afterWorkbookSnapshot["source_change_set_high_watermark"].(string)
	if afterWorkbookWatermark == afterIncidentWatermark {
		t.Fatalf("workbook change-set mutation must change source boundary token: before=%q after=%q", afterIncidentWatermark, afterWorkbookWatermark)
	}
	staleWorkbookBoundary := createSnapshot("txn-reporting-snapshot-stale-workbook", &afterIncidentWatermark)
	httptestx.RequireErrorEnvelope(t, staleWorkbookBoundary, http.StatusConflict, "snapshot_source_boundary_conflict")

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
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createReleaseJob := httptestx.RequireSuccessEnvelope(t, createReleaseResp, http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, createReleaseJob, "release")
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
	bundleManifestSHA, bundleManifest := requireReleaseBundle(t, harness.DB, releaseID)
	if release["output_sha256"] != bundleManifestSHA || bundleManifest["schema_id"] != reporting.RenderBundleManifestSchemaID {
		t.Fatalf("release output_sha256 must bind render bundle manifest: release=%#v bundle_sha=%q bundle=%#v", release, bundleManifestSHA, bundleManifest)
	}
	rendered, manifest := requireReleaseArtifacts(t, harness.DB, releaseID)
	if refs := release["recipient_partition_refs"].([]any); len(refs) != 1 || refs[0] != "party:"+fixture["party"] {
		t.Fatalf("external recipient partitions must bind the selected party, got %#v", refs)
	}
	if strings.Contains(rendered, "Working note body") {
		t.Fatalf("external output must not package working material: %s", rendered)
	}
	if manifest["profile_sha256"] != release["redaction_profile_sha256"] {
		t.Fatalf("manifest profile digest must match release provenance: manifest=%#v release=%#v", manifest["profile_sha256"], release["redaction_profile_sha256"])
	}
	entries := manifestEntriesByPath(t, manifest)
	noteEntry := entries["/notes/"+fixture["note"]]
	if noteEntry["action"] != reporting.ActionDrop || noteEntry["outcome"] != "dropped" {
		t.Fatalf("note working material must be dropped with stable manifest outcome, got %#v", noteEntry)
	}
	descriptionEntry := entries["/incident/description"]
	if descriptionEntry["action"] != reporting.ActionTruncate || descriptionEntry["outcome"] != "truncated" {
		t.Fatalf("source evidence must be truncated for external release, got %#v", descriptionEntry)
	}
	partyEntry := entries["/parties/"+fixture["party"]]
	if partyEntry["outcome"] == "dropped_disclosure_partition" {
		t.Fatalf("selected party partition must be retained, got %#v", partyEntry)
	}

	tokenizedReleaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-tokenized-review",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.TokenizedRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             reporting.ReleaseScopeInternalReview,
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	tokenizedReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, tokenizedReleaseJob, "release")
	tokenizedRelease := requireRelease(t, harness, adminLogin, tokenizedReleaseID)
	if tokenizedRelease["redaction_profile_id"] != reporting.TokenizedRedactionProfileID {
		t.Fatalf("tokenized review release must bind tokenized profile, got %#v", tokenizedRelease)
	}
	_, tokenizedRedactionManifest := requireReleaseArtifacts(t, harness.DB, tokenizedReleaseID)
	tokenizedBundleSHA, tokenizedBundleManifest := requireReleaseBundle(t, harness.DB, tokenizedReleaseID)
	if tokenizedRelease["output_sha256"] != tokenizedBundleSHA {
		t.Fatalf("tokenized release output must bind bundle manifest: release=%#v bundle_sha=%q", tokenizedRelease["output_sha256"], tokenizedBundleSHA)
	}
	tokenManifestSHA, revealMapSHA, tokenManifest, revealMap := requireReleaseTokenArtifacts(t, harness.DB, tokenizedReleaseID)
	if tokenizedRedactionManifest["token_manifest_sha256"] != tokenManifestSHA {
		t.Fatalf("redaction manifest must bind token manifest: manifest=%#v token_sha=%q", tokenizedRedactionManifest, tokenManifestSHA)
	}
	if tokenizedBundleManifest["token_manifest_sha256"] != tokenManifestSHA {
		t.Fatalf("bundle manifest must bind token manifest: bundle=%#v token_sha=%q", tokenizedBundleManifest, tokenManifestSHA)
	}
	requireBundleManifestFile(t, tokenizedBundleManifest, "token_manifest", "validation/token-manifest.json", tokenManifestSHA, true)
	requireBundleManifestFile(t, tokenizedBundleManifest, "sensitive_reveal_map", "internal/reveal-map.json", revealMapSHA, false)
	requireTokenManifestEntry(t, tokenManifest, "party:"+fixture["party"])
	if revealMap["schema_id"] != reporting.RedactionRevealMapSchemaID ||
		revealMap["sensitivity"] != "internal_sensitive" ||
		revealMap["token_manifest_sha256"] != tokenManifestSHA {
		t.Fatalf("reveal map must be internal and token-manifest-bound: %#v", revealMap)
	}

	liveNotesBeforePartitionedRelease := queryLiveWorkbookRowsJSON(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1")
	createPartitionedRelease := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-party",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"], "party:" + fixture["party"]},
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	partitionedJob := httptestx.RequireSuccessEnvelope(t, createPartitionedRelease, http.StatusAccepted)["data"].(map[string]any)
	partitionedReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, partitionedJob, "release")
	partitionedRelease := requireRelease(t, harness, adminLogin, partitionedReleaseID)
	partitionRefs := partitionedRelease["recipient_partition_refs"].([]any)
	if len(partitionRefs) != 1 || partitionRefs[0] != "party:"+fixture["party"] {
		t.Fatalf("recipient partitions must sort and coalesce, got %#v", partitionRefs)
	}
	supersededInitialRelease := requireRelease(t, harness, adminLogin, releaseID)
	if supersededInitialRelease["release_state"] != reporting.ReleaseStateInvalidated ||
		supersededInitialRelease["invalidation_reason"] != "superseded_by_new_render" {
		t.Fatalf("same recipient partition slot must supersede the prior render, got %#v", supersededInitialRelease)
	}
	_, partitionedManifest := requireReleaseArtifacts(t, harness.DB, partitionedReleaseID)
	partitionedEntries := manifestEntriesByPath(t, partitionedManifest)
	if partitionedEntries["/parties/"+fixture["party"]]["outcome"] == "dropped_disclosure_partition" {
		t.Fatalf("requested party partition must be retained in partitioned external release, got %#v", partitionedEntries["/parties/"+fixture["party"]])
	}
	manifestRefs := partitionedManifest["recipient_partition_refs"].([]any)
	if len(manifestRefs) != 1 || manifestRefs[0] != "party:"+fixture["party"] {
		t.Fatalf("redaction manifest must bind recipient partitions, got %#v", manifestRefs)
	}
	secondPartitionedRelease := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-party-supersede",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "1",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	secondPartitionedJob := httptestx.RequireSuccessEnvelope(t, secondPartitionedRelease, http.StatusAccepted)["data"].(map[string]any)
	secondPartitionedReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, secondPartitionedJob, "release")
	supersededPartitionedRelease := requireRelease(t, harness, adminLogin, partitionedReleaseID)
	if supersededPartitionedRelease["release_state"] != reporting.ReleaseStateInvalidated ||
		supersededPartitionedRelease["invalidation_reason"] != "superseded_by_new_render" {
		t.Fatalf("same recipient partition slot must be superseded by a new render, got %#v", supersededPartitionedRelease)
	}
	currentPartitionedRelease := requireRelease(t, harness, adminLogin, secondPartitionedReleaseID)
	if currentPartitionedRelease["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("new partitioned release must remain current pending candidate, got %#v", currentPartitionedRelease)
	}
	liveNotesAfterPartitionedRelease := queryLiveWorkbookRowsJSON(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1")
	if liveNotesAfterPartitionedRelease != liveNotesBeforePartitionedRelease {
		t.Fatalf("recipient redaction must not change live workbook query results: before=%s after=%s", liveNotesBeforePartitionedRelease, liveNotesAfterPartitionedRelease)
	}
}

func TestPhase11_I_11_REPORTING_02_ExternalReleaseApprovalPublishAndStateConflicts(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-lifecycle")

	adminLogin, adminUserID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	reviewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "report-reviewer@example.test", "Report Reviewer", "ReviewerPass1!", false, false, true)
	editorID := phase2test.SeedLocalUserFlags(t, harness.DB, "report-editor@example.test", "Report Editor", "EditorPass1!", false, false, true)
	reviewerSession, reviewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "report-reviewer@example.test", "ReviewerPass1!")
	reviewerLogin := phase2test.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	editorSession, editorCSRF := phase2test.LoginLocalUser(t, harness.Server, "report-editor@example.test", "EditorPass1!")
	editorLogin := phase2test.LoginResult{SessionCookie: editorSession, CSRFCookie: editorCSRF}

	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-lifecycle-incident",
		"incident_key":  "IR-REPORTING-02",
		"title":         "Reporting Lifecycle Incident",
		"description":   "approval lifecycle source text",
	})
	incidentID := incident["incident_id"].(string)
	fixture := seedReportingWorkbookFixture(t, harness.DB, incidentID, adminUserID)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-editor-membership",
		"user_id":       editorID,
		"role":          "editor",
	})

	snapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-lifecycle-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")

	releaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
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
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")

	editorApprove := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-editor-approve",
			"reason":        "editor cannot approve release",
		},
		phase2test.WithCookies(editorLogin.SessionCookie, editorLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, editorLogin.CSRFCookie.Value),
	)
	editorError := httptestx.RequireErrorEnvelope(t, editorApprove, http.StatusConflict, "release_approval_rejected")
	if editorError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "actor_lacks_approval_role" {
		t.Fatalf("editor approval must use release approval rejection reason, got %#v", editorError)
	}

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

	snapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-idempotency-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")
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
	replayJob := httptestx.RequireSuccessEnvelope(t, replayWithExplicitBoundary, http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != snapshotJob["job_id"] {
		t.Fatalf("explicit snapshot replay must return original job: first=%#v replay=%#v", snapshotJob, replayJob)
	}
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
		t.Fatalf("explicit original boundary replay must not create a new snapshot, got %d rows", got)
	}

	remoteAssetIncident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-idempotency-remote-asset-incident",
		"incident_key":  "IR-REPORTING-03-REMOTE",
		"title":         "Reporting Remote Asset Incident",
		"description":   "![remote asset](https://cdn.example.test/report.png)",
	})
	remoteAssetIncidentID := remoteAssetIncident["incident_id"].(string)
	remoteAssetSnapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   remoteAssetIncidentID,
			"client_txn_id": "txn-reporting-idempotency-remote-asset-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	remoteAssetSnapshotID := requireSucceededJobResourceID(t, harness, adminLogin, remoteAssetSnapshotJob, "snapshot")
	remoteAssetReleaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               remoteAssetSnapshotID,
			"client_txn_id":             "txn-reporting-idempotency-remote-asset-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	remoteAssetReleaseID := requireFailedReleaseJob(t, harness, adminLogin, remoteAssetReleaseJob, "template_render_failed")
	remoteAssetRelease := requireRelease(t, harness, adminLogin, remoteAssetReleaseID)
	if remoteAssetRelease["release_state"] != reporting.ReleaseStateRenderFailed ||
		remoteAssetRelease["render_failed_reason_code"] != "template_render_failed" ||
		remoteAssetRelease["output_sha256"] != nil ||
		remoteAssetRelease["redaction_manifest_sha256"] != nil ||
		remoteAssetRelease["output_media_type"] != nil {
		t.Fatalf("self-contained output validation must persist render-failed release, got %#v", remoteAssetRelease)
	}
	failedReleaseID := remoteAssetReleaseID
	unknownReleaseID := "00000000-0000-0000-0000-000000000999"
	for _, action := range []string{"approve", "publish", "invalidate"} {
		malformed := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+failedReleaseID+"/"+action,
			[]string{"not", "an", "object"},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, malformed, http.StatusBadRequest, "invalid_release_request")

		missing := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+unknownReleaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-missing-" + action},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, missing, http.StatusNotFound, "release_not_found")
	}
	for _, action := range []string{"approve", "publish", "invalidate"} {
		resp := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+failedReleaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-render-failed-" + action},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "release_state_conflict")
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "render_failed" {
			t.Fatalf("%s on render_failed release must return render_failed conflict, got %#v", action, body)
		}
	}

	releaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
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
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
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

func TestPhase11_I_11_REPORTING_04_ExactShapesAndRouteScopedVisibility(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase11-reporting-shape-auth")

	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	reviewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "shape-reviewer@example.test", "Shape Reviewer", "ShapeReviewer1!", false, false, true)
	phase2test.SeedLocalUserFlags(t, harness.DB, "shape-outsider@example.test", "Shape Outsider", "ShapeOutsider1!", false, false, true)
	phase2test.SeedLocalUserFlags(t, harness.DB, "shape-deployment-admin@example.test", "Shape Deployment Admin", "ShapeDeployAdmin1!", false, true, true)
	reviewerSession, reviewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "shape-reviewer@example.test", "ShapeReviewer1!")
	reviewerLogin := phase2test.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	outsiderSession, outsiderCSRF := phase2test.LoginLocalUser(t, harness.Server, "shape-outsider@example.test", "ShapeOutsider1!")
	outsiderLogin := phase2test.LoginResult{SessionCookie: outsiderSession, CSRFCookie: outsiderCSRF}
	deploymentAdminSession, deploymentAdminCSRF := phase2test.LoginLocalUser(t, harness.Server, "shape-deployment-admin@example.test", "ShapeDeployAdmin1!")
	deploymentAdminLogin := phase2test.LoginResult{SessionCookie: deploymentAdminSession, CSRFCookie: deploymentAdminCSRF}

	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-shape-incident",
		"incident_key":  "IR-REPORTING-04",
		"title":         "Reporting Shape Incident",
		"description":   "resource shape and visibility source text",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-shape-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})
	_ = phase2test.CreateIncident(t, harness.Server, outsiderLogin, map[string]any{
		"client_txn_id": "txn-reporting-shape-other-incident",
		"incident_key":  "IR-REPORTING-04-OTHER",
		"title":         "Other Reporting Incident",
	})

	snapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")
	_ = requireSnapshot(t, harness, adminLogin, snapshotID)

	paginatedSnapshot := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID+"?limit=1",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, paginatedSnapshot, http.StatusBadRequest, "invalid_pagination_request")

	hiddenSnapshot := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		phase2test.WithCookies(outsiderLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, hiddenSnapshot, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminHiddenSnapshot := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenSnapshot, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminCreateSnapshot := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-deployment-admin-snapshot",
		},
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminCreateSnapshot, http.StatusNotFound, "incident_not_found")
	deploymentAdminReadJob := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/jobs/"+snapshotJob["job_id"].(string),
		nil,
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminReadJob, http.StatusNotFound, "job_not_found")
	deploymentAdminCancelJob := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/jobs/"+snapshotJob["job_id"].(string)+"/cancel",
		map[string]any{"client_txn_id": "txn-reporting-shape-deployment-admin-cancel"},
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminCancelJob, http.StatusNotFound, "job_not_found")

	hiddenReleaseCreate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-hidden-release-create",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(outsiderLogin.SessionCookie, outsiderLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, outsiderLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, hiddenReleaseCreate, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminHiddenReleaseCreate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-deployment-admin-release-create",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenReleaseCreate, http.StatusNotFound, "snapshot_not_found")

	releaseCreateBody := func(clientTxnID string, mutate func(map[string]any)) map[string]any {
		body := map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             clientTxnID,
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               reporting.OutputKindSlidev,
		}
		if mutate != nil {
			mutate(body)
		}
		return body
	}
	postReleaseCreate := func(actor phase2test.LoginResult, body map[string]any) *http.Response {
		t.Helper()
		return phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases",
			body,
			phase2test.WithCookies(actor.SessionCookie, actor.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
		)
	}
	hiddenSelectorCases := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "unsupported template", mutate: func(body map[string]any) { body["template_id"] = "cartulary.report.missing" }},
		{name: "unsupported redaction profile", mutate: func(body map[string]any) { body["redaction_profile_version"] = "missing" }},
		{name: "unsupported output kind", mutate: func(body map[string]any) { body["output_kind"] = "docx" }},
		{name: "unsupported release scope", mutate: func(body map[string]any) { body["release_scope"] = "public" }},
		{name: "recipient partitions with internal scope", mutate: func(body map[string]any) { body["recipient_partition_refs"] = []string{"party:hidden"} }},
	}
	for i, tc := range hiddenSelectorCases {
		body := releaseCreateBody(fmt.Sprintf("txn-reporting-shape-hidden-selector-%d", i), tc.mutate)
		resp := postReleaseCreate(outsiderLogin, body)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "snapshot_not_found")
	}
	visibleSelectorCases := []struct {
		name       string
		mutate     func(map[string]any)
		wantReason string
	}{
		{name: "unsupported template", mutate: func(body map[string]any) { body["template_id"] = "cartulary.report.missing" }, wantReason: "unsupported_template"},
		{name: "unsupported redaction profile", mutate: func(body map[string]any) { body["redaction_profile_version"] = "missing" }, wantReason: "unsupported_redaction_profile"},
		{name: "unsupported output kind", mutate: func(body map[string]any) { body["output_kind"] = "docx" }, wantReason: "unsupported_output_kind"},
		{name: "unsupported release scope", mutate: func(body map[string]any) { body["release_scope"] = "public" }, wantReason: "unsupported_release_scope"},
		{name: "recipient partitions with internal scope", mutate: func(body map[string]any) { body["recipient_partition_refs"] = []string{"party:visible"} }, wantReason: "recipient_partitions_not_allowed"},
	}
	for i, tc := range visibleSelectorCases {
		body := releaseCreateBody(fmt.Sprintf("txn-reporting-shape-visible-selector-%d", i), tc.mutate)
		resp := postReleaseCreate(adminLogin, body)
		envelope := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_release_request")
		details := httptestx.RequireErrorDetails(t, envelope)
		if details["reason_code"] != tc.wantReason {
			t.Fatalf("%s reason_code = %#v want %q", tc.name, details["reason_code"], tc.wantReason)
		}
	}

	unsupportedSnapshotJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-unsupported-derivation-snapshot",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	unsupportedSnapshotID := requireSucceededJobResourceID(t, harness, adminLogin, unsupportedSnapshotJob, "snapshot")
	if _, err := harness.DB.ExecContext(
		t.Context(),
		`UPDATE reporting_snapshots SET derivation_version = 'cartulary.reporting_derivation_profile.unsupported.v1' WHERE snapshot_id = $1`,
		unsupportedSnapshotID,
	); err != nil {
		t.Fatalf("mark unsupported derivation snapshot: %v", err)
	}
	unsupportedDerivationRelease := postReleaseCreate(adminLogin, releaseCreateBody("txn-reporting-shape-unsupported-derivation-release", func(body map[string]any) {
		body["snapshot_id"] = unsupportedSnapshotID
	}))
	unsupportedEnvelope := httptestx.RequireErrorEnvelope(t, unsupportedDerivationRelease, http.StatusBadRequest, "invalid_release_request")
	unsupportedDetails := httptestx.RequireErrorDetails(t, unsupportedEnvelope)
	if unsupportedDetails["reason_code"] != "unsupported_derivation_version" {
		t.Fatalf("unsupported snapshot derivation reason_code = %#v want unsupported_derivation_version", unsupportedDetails["reason_code"])
	}

	releaseJob := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "1",
			"output_kind":               reporting.OutputKindSlidev,
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_scope"] != reporting.ReleaseScopeInternalDraft || release["output_kind"] != reporting.OutputKindSlidev {
		t.Fatalf("release resource must expose resolved closed vocabularies, got %#v", release)
	}

	paginatedRelease := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"?limit=1",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, paginatedRelease, http.StatusBadRequest, "invalid_pagination_request")

	hiddenRelease := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		phase2test.WithCookies(outsiderLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, hiddenRelease, http.StatusNotFound, "release_not_found")
	deploymentAdminHiddenRelease := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		phase2test.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenRelease, http.StatusNotFound, "release_not_found")

	for _, action := range []string{"approve", "publish", "invalidate"} {
		paginatedAction := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action+"?limit=1",
			map[string]any{"client_txn_id": "txn-reporting-shape-paginated-" + action},
			phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, paginatedAction, http.StatusBadRequest, "invalid_pagination_request")

		hiddenAction := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-shape-hidden-" + action},
			phase2test.WithCookies(outsiderLogin.SessionCookie, outsiderLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, outsiderLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, hiddenAction, http.StatusNotFound, "release_not_found")

		deploymentAdminHiddenAction := phase2test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-shape-deployment-admin-" + action},
			phase2test.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
			phase2test.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenAction, http.StatusNotFound, "release_not_found")
	}

	reviewerPublish := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{"client_txn_id": "txn-reporting-shape-reviewer-publish"},
		phase2test.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerPublish, http.StatusForbidden, "authorization_denied")

	published := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{"client_txn_id": "txn-reporting-shape-admin-publish"},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusOK)["data"].(map[string]any)
	requireReleaseResourceShape(t, published)
	if got := requireRelease(t, harness, adminLogin, releaseID); !reflect.DeepEqual(published, got) {
		t.Fatalf("publish action response must match GET release resource: action=%#v get=%#v", published, got)
	}

	reviewerInvalidate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/invalidate",
		map[string]any{"client_txn_id": "txn-reporting-shape-reviewer-invalidate"},
		phase2test.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerInvalidate, http.StatusForbidden, "authorization_denied")

	invalidated := httptestx.RequireSuccessEnvelope(t, phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/invalidate",
		map[string]any{"client_txn_id": "txn-reporting-shape-admin-invalidate"},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusOK)["data"].(map[string]any)
	requireReleaseResourceShape(t, invalidated)
	if got := requireRelease(t, harness, adminLogin, releaseID); !reflect.DeepEqual(invalidated, got) {
		t.Fatalf("invalidate action response must match GET release resource: action=%#v get=%#v", invalidated, got)
	}
}

func seedReportingWorkbookFixture(t testing.TB, db *sql.DB, incidentID string, actorID string) map[string]string {
	t.Helper()
	ids := map[string]string{}
	ids["timeline"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'timeline_event', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, raw_activity_text, data_source_text, capture_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'Initial timeline summary', 'Timeline details', 'Timeline source text', 'rough', $2, $2
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["host"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'host', $2, $2)
    RETURNING record_id
), host AS (
    INSERT INTO hosts (record_id, incident_id, display_name, hostname, host_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'host-reporting-01', 'host-reporting-01.example.test', 'canonical', $2, $2
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO host_grid_projection (record_id, incident_id, row_version, display_name, hostname, host_state, edited_at)
    SELECT record_id, $1, 1, 'host-reporting-01', 'host-reporting-01.example.test', 'canonical', now()
      FROM host
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["identity"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'identity', $2, $2)
    RETURNING record_id
), identity AS (
    INSERT INTO identities (record_id, incident_id, display_name, upn, email, identity_state, created_by_user_id, updated_by_user_id)
    SELECT record_id, $1, 'Report Identity', 'report.identity@example.test', 'report.identity@example.test', 'canonical', $2, $2
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO identity_grid_projection (record_id, incident_id, row_version, display_name, upn, email, identity_state, edited_at)
    SELECT record_id, $1, 1, 'Report Identity', 'report.identity@example.test', 'report.identity@example.test', 'canonical', now()
      FROM identity
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["party"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'party', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO parties (record_id, incident_id, display_name, party_kind, organization_name, primary_email)
    SELECT record_id, $1, 'Recipient Organization', 'organization', 'Recipient Org', 'recipient@example.test'
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["evidence"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'evidence', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, requested_at, blob_hash, storage_ref)
    SELECT record_id, $1, 'Evidence metadata only', 'received', 'complete', now(), 'raw-blob-hash-must-not-export', 'object://evidence/reporting'
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	ids["task"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'task_request', $2, $2)
    RETURNING record_id
), task AS (
    INSERT INTO task_requests (record_id, incident_id, title, status, owner_user_id, priority, task_kind, workstream)
    SELECT record_id, $1, 'Contain reporting host', 'open', $2, 'high', 'remediation', 'containment'
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO task_request_grid_projection (
        record_id, incident_id, row_version, title, status, owner_user_id,
        priority, task_kind, workstream, linked_record_count, updated_at, no_owner
    )
    SELECT record_id, $1, 1, 'Contain reporting host', 'open', $2,
           'high', 'remediation', 'containment', 0, now(), false
      FROM task
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["decision"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'decision', $2, $2)
    RETURNING record_id
), decision AS (
    INSERT INTO decisions (record_id, incident_id, summary, status, owner_user_id, decision_type, rationale)
    SELECT record_id, $1, 'Escalate reporting case', 'approved', $2, 'scope', 'Evidence supports escalation.'
      FROM rec
    RETURNING record_id
), projection AS (
    INSERT INTO decision_grid_projection (
        record_id, incident_id, row_version, summary, status, owner_user_id,
        decision_type, rationale, affected_record_count, updated_at, is_superseded
    )
    SELECT record_id, $1, 1, 'Escalate reporting case', 'approved', $2,
           'scope', 'Evidence supports escalation.', 0, now(), false
      FROM decision
    RETURNING record_id
)
SELECT record_id::text FROM projection
`, incidentID, actorID)
	ids["note"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, 'note', 'Reporting analyst note', 'Working note body', $2
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID)
	seedReportingArtifactProjection(t, db, ids["note"])
	ids["artifact"] = ids["note"]
	ids["finding"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), artifact AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, 'finding', 'Reporting finding', 'Finding working notes', $2
      FROM rec
    RETURNING record_id
), finding AS (
    INSERT INTO artifact_findings (record_id, incident_id, kind, statement, state, confidence_score, owner_user_id)
    SELECT record_id, $1, 'finding', 'Evidence supports reporting escalation.', 'open', 85, $2
      FROM artifact
    RETURNING record_id
)
SELECT record_id::text FROM finding
`, incidentID, actorID)
	seedReportingArtifactProjection(t, db, ids["finding"])
	for _, seed := range []struct {
		key          string
		artifactType string
		title        string
		body         string
	}{
		{key: "comm_log", artifactType: "comm_log", title: "Reporting communication", body: "Coordination communication body"},
		{key: "handoff", artifactType: "handoff", title: "Reporting handoff", body: "Handoff risk body"},
		{key: "status_review", artifactType: "status_review", title: "Reporting status review", body: "Status review body"},
		{key: "lesson", artifactType: "lesson", title: "Reporting lesson", body: "Lesson body"},
	} {
		ids[seed.key] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), inserted AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, $3, $4, $5, $2
      FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM inserted
`, incidentID, actorID, seed.artifactType, seed.title, seed.body)
		seedReportingArtifactProjection(t, db, ids[seed.key])
	}
	if _, err := db.Exec(`
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, 'supported_by', 'manual', $4, $4)
`, incidentID, ids["task"], ids["evidence"], actorID); err != nil {
		t.Fatalf("seed reporting support link: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO record_links (incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id)
VALUES ($1, $2, $3, 'supported_by', 'manual', $4, $4)
`, incidentID, ids["finding"], ids["evidence"], actorID); err != nil {
		t.Fatalf("seed reporting finding support link: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'reporting-tag', 'reporting-tag', $3)
`, incidentID, ids["host"], actorID); err != nil {
		t.Fatalf("seed reporting tag: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, ordinal, created_by_user_id
)
VALUES ($1, 'host', 'timeline.activity_synopsis_text', 'manual', 'summary:1', 'host-reporting-01', 'host-reporting-01', 'unresolved', 1, $2)
`, ids["timeline"], actorID); err != nil {
		t.Fatalf("seed reporting entity mention: %v", err)
	}
	return ids
}

func querySeedID(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var id string
	if err := db.QueryRow(query, args...).Scan(&id); err != nil {
		t.Fatalf("seed query failed: %v", err)
	}
	return id
}

func seedReportingArtifactProjection(t testing.TB, db *sql.DB, recordID string) {
	t.Helper()
	if _, err := db.Exec(`
INSERT INTO artifact_grid_projection (
    record_id,
    incident_id,
    row_version,
    artifact_type,
    title,
    body,
    timestamp_utc,
    updated_at,
    created_at,
    created_by_user_id,
    comm_id,
    comm_type,
    audience,
    channel_or_meeting,
    summary,
    next_report_at,
    privilege_tag,
    handoff_id,
    outgoing_owner_user_id,
    incoming_owner_user_id,
    current_state_summary,
    next_checks,
    acknowledged_at,
    status_review_id,
    review_owner_user_id,
    active_risks_summary,
    lesson_id,
    owner_user_id,
    closure_state,
    finding_statement,
    finding_kind,
    finding_state,
    finding_owner_user_id,
    finding_confidence_score,
    finding_closed_at,
    finding_updated_at,
    finding_confidence_band,
    timestamp_day,
    next_report_day,
    ack_state,
    linked_record_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.artifact_type,
    a.title,
    a.body,
    a.timestamp_utc,
    a.updated_at,
    a.created_at,
    a.created_by_user_id,
    a.comm_id,
    a.comm_type,
    a.audience,
    a.channel_or_meeting,
    a.summary,
    a.next_report_at,
    a.privilege_tag,
    a.handoff_id,
    a.outgoing_owner_user_id,
    a.incoming_owner_user_id,
    a.current_state_summary,
    a.next_checks,
    a.acknowledged_at,
    a.status_review_id,
    a.review_owner_user_id,
    a.active_risks_summary,
    a.lesson_id,
    a.owner_user_id,
    a.closure_state,
    f.statement,
    f.kind,
    f.state,
    f.owner_user_id,
    f.confidence_score,
    f.closed_at,
    GREATEST(a.updated_at, f.updated_at),
    cartulary_confidence_band(f.confidence_score),
    a.timestamp_utc::date,
    a.next_report_at::date,
    CASE WHEN a.acknowledged_at IS NULL THEN 'pending' ELSE 'acknowledged' END,
    0
  FROM artifacts a
  JOIN records r
    ON r.incident_id = a.incident_id
   AND r.record_id = a.record_id
   AND r.deleted_at IS NULL
  LEFT JOIN artifact_findings f
    ON f.incident_id = a.incident_id
   AND f.record_id = a.record_id
 WHERE a.record_id::text = $1
`, recordID); err != nil {
		t.Fatalf("seed reporting artifact projection: %v", err)
	}
}

func requireSnapshotExportModel(t testing.TB, db *sql.DB, snapshotID string) map[string]any {
	t.Helper()
	var raw []byte
	if err := db.QueryRow(`
SELECT export_model_json
  FROM reporting_snapshots
 WHERE snapshot_id::text = $1
`, snapshotID).Scan(&raw); err != nil {
		t.Fatalf("query snapshot export model: %v", err)
	}
	var model map[string]any
	if err := json.Unmarshal(raw, &model); err != nil {
		t.Fatalf("decode snapshot export model: %v", err)
	}
	for _, forbidden := range []string{"raw-blob-hash-must-not-export", "blob_hash", "storage_ref", "object_blob_id", "reporting_job_payloads", "client_txn_id"} {
		if strings.Contains(string(raw), forbidden) {
			t.Fatalf("snapshot export model must exclude forbidden source material %q: %s", forbidden, string(raw))
		}
	}
	return model
}

func requireSnapshotBoundaryJSON(t testing.TB, db *sql.DB, snapshotID string, token string, incidentID string) {
	t.Helper()
	if !strings.HasPrefix(token, reporting.SourceBoundaryTokenPrefix) {
		t.Fatalf("source boundary token must use current prefix, got %q", token)
	}
	var raw []byte
	if err := db.QueryRow(`
SELECT source_boundary_json
  FROM reporting_snapshots
 WHERE snapshot_id::text = $1
`, snapshotID).Scan(&raw); err != nil {
		t.Fatalf("query snapshot source boundary json: %v", err)
	}
	var boundary map[string]any
	if err := json.Unmarshal(raw, &boundary); err != nil {
		t.Fatalf("decode source boundary json: %v", err)
	}
	if boundary["incident_id"] != incidentID {
		t.Fatalf("source boundary json incident_id = %#v want %s: %s", boundary["incident_id"], incidentID, string(raw))
	}
	if boundary["incident_version"].(float64) < 1 {
		t.Fatalf("source boundary json must persist incident_version: %s", string(raw))
	}
	if _, ok := boundary["latest_change_set_id"]; !ok {
		t.Fatalf("source boundary json must include latest_change_set_id key: %s", string(raw))
	}
	if _, ok := boundary["latest_change_set_created_at"]; !ok {
		t.Fatalf("source boundary json must include latest_change_set_created_at key: %s", string(raw))
	}
}

func createWorkbookNote(t testing.TB, harness *phase2test.ServerHarness, login phase2test.LoginResult, incidentID string, clientTxnID string) {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.notes.v1/rows",
		map[string]any{
			"client_txn_id": clientTxnID,
			"note.title":    "Boundary note",
			"note.body":     "Boundary note body",
		},
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)
}

func queryLiveWorkbookRowsJSON(t testing.TB, harness *phase2test.ServerHarness, login phase2test.LoginResult, incidentID string, viewSchemaID string) string {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query",
		map[string]any{"sort": []map[string]any{{"field_key": "note.title", "direction": "asc"}}},
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	envelope := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	rows := envelope["data"].(map[string]any)["rows"]
	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatalf("encode live workbook rows: %v", err)
	}
	return string(encoded)
}

func requireExportModelCoverage(t testing.TB, model map[string]any, ids map[string]string) {
	t.Helper()
	if model["schema_id"] != reporting.ExportModelSchemaID || model["derivation_version"] != reporting.DerivationVersion {
		t.Fatalf("snapshot export model must use current v1 identity, got %#v", model)
	}
	for _, key := range []string{"sections", "records", "relationships", "timeline_events", "subjects", "diagrams", "assets", "support_index", "validation_summary"} {
		if _, ok := model[key]; !ok {
			t.Fatalf("snapshot export model missing v1 member %s: %#v", key, model)
		}
	}
	if _, ok := model["fields"]; ok {
		t.Fatalf("snapshot export model must not persist legacy top-level fields[]")
	}
	byPath := exportModelFieldsByPath(model)
	wantPaths := []string{
		"/timeline/" + ids["timeline"],
		"/hosts/" + ids["host"],
		"/identities/" + ids["identity"],
		"/parties/" + ids["party"],
		"/evidence/" + ids["evidence"],
		"/task_requests/" + ids["task"],
		"/decisions/" + ids["decision"],
		"/notes/" + ids["note"],
		"/findings/" + ids["finding"],
		"/comm_log/" + ids["comm_log"],
		"/handoffs/" + ids["handoff"],
		"/status_reviews/" + ids["status_review"],
		"/lessons/" + ids["lesson"],
	}
	for _, path := range wantPaths {
		if byPath[path] == nil {
			t.Fatalf("snapshot export model missing workbook path %s; paths=%#v", path, byPath)
		}
	}
	wantFamilies := map[string]string{
		"/timeline/" + ids["timeline"]:            "timeline_event",
		"/hosts/" + ids["host"]:                   "host",
		"/identities/" + ids["identity"]:          "identity",
		"/parties/" + ids["party"]:                "party",
		"/evidence/" + ids["evidence"]:            "evidence",
		"/task_requests/" + ids["task"]:           "task_request",
		"/decisions/" + ids["decision"]:           "decision",
		"/notes/" + ids["note"]:                   "note",
		"/findings/" + ids["finding"]:             "finding_hypothesis",
		"/comm_log/" + ids["comm_log"]:            "comm_log",
		"/handoffs/" + ids["handoff"]:             "handoff",
		"/status_reviews/" + ids["status_review"]: "status_review",
		"/lessons/" + ids["lesson"]:               "lesson",
	}
	for path, family := range wantFamilies {
		if byPath[path]["source_family"] != family {
			t.Fatalf("path %s source_family = %#v want %q", path, byPath[path]["source_family"], family)
		}
	}
	wantClasses := map[string]string{
		"/hosts/" + ids["host"]:                   reporting.ContentClassDerivedAnalytic,
		"/task_requests/" + ids["task"]:           reporting.ContentClassWorkingMaterial,
		"/decisions/" + ids["decision"]:           reporting.ContentClassWorkingMaterial,
		"/notes/" + ids["note"]:                   reporting.ContentClassWorkingMaterial,
		"/findings/" + ids["finding"]:             reporting.ContentClassCuratedNarrative,
		"/comm_log/" + ids["comm_log"]:            reporting.ContentClassWorkingMaterial,
		"/handoffs/" + ids["handoff"]:             reporting.ContentClassWorkingMaterial,
		"/status_reviews/" + ids["status_review"]: reporting.ContentClassWorkingMaterial,
		"/lessons/" + ids["lesson"]:               reporting.ContentClassWorkingMaterial,
	}
	for path, class := range wantClasses {
		if byPath[path]["content_class"] != class {
			t.Fatalf("path %s content_class = %#v want %q", path, byPath[path]["content_class"], class)
		}
	}
	if byPath["/relationships/"+ids["task"]] == nil {
		for path := range byPath {
			if strings.HasPrefix(path, "/relationships/") {
				goto relationshipPresent
			}
		}
		t.Fatalf("snapshot export model missing active relationship metadata; paths=%#v", byPath)
	}
relationshipPresent:
	tagPresent := false
	for path := range byPath {
		if strings.HasPrefix(path, "/tags/") {
			tagPresent = true
			break
		}
	}
	if !tagPresent {
		t.Fatalf("snapshot export model missing active tag metadata; paths=%#v", byPath)
	}
	mentionPresent := false
	for path := range byPath {
		if strings.HasPrefix(path, "/entity_mentions/") {
			mentionPresent = true
			break
		}
	}
	if !mentionPresent {
		t.Fatalf("snapshot export model missing entity mention metadata; paths=%#v", byPath)
	}
	supportRefs := byPath["/task_requests/"+ids["task"]]["support_refs"].([]any)
	if len(supportRefs) != 1 || supportRefs[0] != "/record_envelopes/"+ids["evidence"] {
		t.Fatalf("task export field must carry deterministic support refs, got %#v", supportRefs)
	}
	findingSupportRefs := byPath["/findings/"+ids["finding"]]["support_refs"].([]any)
	if len(findingSupportRefs) != 1 || findingSupportRefs[0] != "/record_envelopes/"+ids["evidence"] {
		t.Fatalf("curated finding export field must carry deterministic support refs, got %#v", findingSupportRefs)
	}
	partyPartitions := byPath["/parties/"+ids["party"]]["disclosure_partition_refs"].([]any)
	if len(partyPartitions) != 1 || partyPartitions[0] != "party:"+ids["party"] {
		t.Fatalf("party export field must carry disclosure partition refs, got %#v", partyPartitions)
	}
}

func exportModelFieldsByPath(model map[string]any) map[string]map[string]any {
	if fields, ok := model["fields"].([]any); ok {
		byPath := make(map[string]map[string]any, len(fields))
		for _, item := range fields {
			field := item.(map[string]any)
			byPath[field["path"].(string)] = field
		}
		return byPath
	}
	byPath := map[string]map[string]any{}
	for _, sectionItem := range model["sections"].([]any) {
		section := sectionItem.(map[string]any)
		collectExportModelBlockFields(section["blocks"].([]any), byPath)
	}
	return byPath
}

func collectExportModelBlockFields(blocks []any, byPath map[string]map[string]any) {
	for _, blockItem := range blocks {
		block := blockItem.(map[string]any)
		for _, fieldItem := range block["fields"].([]any) {
			field := fieldItem.(map[string]any)
			sourceRefs := field["source_refs"].([]any)
			if len(sourceRefs) == 0 {
				continue
			}
			path := sourceRefs[0].(string)
			byPath[path] = map[string]any{
				"path":                      path,
				"source_family":             sourceFamilyForExportPath(path),
				"content_class":             contentClassForExportPath(path),
				"value":                     field["value"],
				"support_refs":              field["support_refs"],
				"disclosure_partition_refs": field["disclosure_partition_refs"],
			}
		}
		collectExportModelBlockFields(block["children"].([]any), byPath)
	}
}

func sourceFamilyForExportPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	switch parts[0] {
	case "timeline":
		return "timeline_event"
	case "hosts":
		return "host"
	case "identities":
		return "identity"
	case "parties":
		return "party"
	case "evidence":
		return "evidence"
	case "task_requests":
		return "task_request"
	case "decisions":
		return "decision"
	case "notes":
		return "note"
	case "findings":
		return "finding_hypothesis"
	case "comm_log":
		return "comm_log"
	case "handoffs":
		return "handoff"
	case "status_reviews":
		return "status_review"
	case "lessons":
		return "lesson"
	case "relationships":
		return "record_link"
	case "tags":
		return "record_tag"
	case "entity_mentions":
		return "entity_mention"
	case "record_envelopes":
		return "record_envelope"
	default:
		return parts[0]
	}
}

func contentClassForExportPath(path string) string {
	switch sourceFamilyForExportPath(path) {
	case "host", "identity", "record_link", "record_tag", "record_envelope":
		return reporting.ContentClassDerivedAnalytic
	case "timeline_event", "party", "evidence", "entity_mention":
		return reporting.ContentClassSourceEvidence
	case "finding_hypothesis":
		return reporting.ContentClassCuratedNarrative
	default:
		return reporting.ContentClassWorkingMaterial
	}
}

func requireSucceededJobResourceID(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, job map[string]any, wantKind string) string {
	t.Helper()
	finalJob := requireJobStatus(t, harness, actor, job["job_id"].(string), "succeeded")
	summary := finalJob["result_summary"].(map[string]any)
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 {
		t.Fatalf("job must emit exactly one resource ref, got %#v", refs)
	}
	ref := refs[0].(map[string]any)
	if ref["kind"] != wantKind || ref["id"] == "" {
		t.Fatalf("unexpected job resource ref: got %#v want kind %q", ref, wantKind)
	}
	return ref["id"].(string)
}

func requireFailedReleaseJob(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, job map[string]any, wantReason string) string {
	t.Helper()
	finalJob := requireJobStatus(t, harness, actor, job["job_id"].(string), "failed")
	summary := finalJob["error_summary"].(map[string]any)
	if summary["code"] != "release_render_failed" {
		t.Fatalf("release render failure job code = %#v want release_render_failed: %#v", summary["code"], summary)
	}
	details := summary["details"].(map[string]any)
	if details["reason_code"] != wantReason || details["release_id"] == "" {
		t.Fatalf("release render failure details = %#v want reason %q and release id", details, wantReason)
	}
	return details["release_id"].(string)
}

func requireJobStatus(t testing.TB, harness *phase2test.ServerHarness, actor phase2test.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID,
			nil,
			phase2test.WithCookies(actor.SessionCookie),
		)
		last = httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		switch last["status"] {
		case "queued", "running", "cancel_requested":
			time.Sleep(25 * time.Millisecond)
			continue
		default:
			if last["status"] != wantStatus {
				t.Fatalf("job %s terminal status = %#v want %q: %#v", jobID, last["status"], wantStatus, last)
			}
			return last
		}
	}
	t.Fatalf("timed out waiting for job %s status %q after %#v", jobID, wantStatus, last)
	return nil
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
	resource := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	requireSnapshotResourceShape(t, resource)
	return resource
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
	resource := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	requireReleaseResourceShape(t, resource)
	return resource
}

func requireSnapshotResourceShape(t testing.TB, resource map[string]any) {
	t.Helper()
	requireExactResourceKeys(t, "snapshot resource", resource, []string{
		"snapshot_id",
		"incident_id",
		"created_by_user_id",
		"created_at",
		"snapshot_at",
		"source_change_set_high_watermark",
		"derivation_version",
		"export_model_sha256",
	})
}

func requireReleaseResourceShape(t testing.TB, resource map[string]any) {
	t.Helper()
	requireExactResourceKeys(t, "release resource", resource, []string{
		"release_id",
		"incident_id",
		"snapshot_id",
		"snapshot_at",
		"source_change_set_high_watermark",
		"derivation_version",
		"export_model_sha256",
		"template_id",
		"template_version",
		"redaction_profile_id",
		"redaction_profile_version",
		"redaction_profile_sha256",
		"output_kind",
		"output_media_type",
		"release_scope",
		"recipient_partition_refs",
		"output_options",
		"graph_projection_refs",
		"composition_id",
		"composition_version",
		"composition_sha256",
		"render_admitted_at",
		"output_sha256",
		"redaction_manifest_sha256",
		"release_state",
		"render_failed_reason_code",
		"created_by_user_id",
		"created_at",
		"approved_at",
		"invalidated_at",
		"published_at",
		"invalidation_reason",
	})
	requireNoSensitiveReleaseArtifactExposure(t, resource)
}

func requireNoSensitiveReleaseArtifactExposure(t testing.TB, resource map[string]any) {
	t.Helper()
	for _, key := range []string{
		"redaction_manifest_json",
		"render_bundle_manifest_json",
		"redaction_profile_view",
		"redaction_profile_view_json",
		"token_manifest",
		"token_manifest_json",
		"reveal_map",
		"reveal_map_json",
	} {
		if _, ok := resource[key]; ok {
			t.Fatalf("release resource must not expose sensitive or internal artifact %q: %#v", key, resource)
		}
	}
}

func requireExactResourceKeys(t testing.TB, label string, resource map[string]any, want []string) {
	t.Helper()
	wantSet := make(map[string]struct{}, len(want))
	for _, key := range want {
		wantSet[key] = struct{}{}
	}
	var missing []string
	for _, key := range want {
		if _, ok := resource[key]; !ok {
			missing = append(missing, key)
		}
	}
	var extra []string
	for key := range resource {
		if _, ok := wantSet[key]; !ok {
			extra = append(extra, key)
		}
	}
	if len(missing) > 0 || len(extra) > 0 {
		sort.Strings(missing)
		sort.Strings(extra)
		t.Fatalf("%s exact shape mismatch: missing=%v extra=%v resource=%#v", label, missing, extra, resource)
	}
}

func requireReleaseArtifacts(t testing.TB, db *sql.DB, releaseID string) (string, map[string]any) {
	t.Helper()
	var rendered []byte
	var manifestBytes []byte
	if err := db.QueryRow(`
SELECT f.inline_bytes, r.redaction_manifest_json
  FROM reporting_releases r
  JOIN reporting_render_bundles b ON b.release_id = r.release_id
  JOIN reporting_render_bundle_files f
    ON f.release_id = b.release_id
   AND f.bundle_path = b.primary_bundle_path
 WHERE r.release_id::text = $1
`, releaseID).Scan(&rendered, &manifestBytes); err != nil {
		t.Fatalf("query release artifacts: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode redaction manifest: %v", err)
	}
	return string(rendered), manifest
}

func requireReleaseBundle(t testing.TB, db *sql.DB, releaseID string) (string, map[string]any) {
	t.Helper()
	var bundleSHA string
	var manifestBytes []byte
	var primaryPath string
	if err := db.QueryRow(`
SELECT bundle_manifest_sha256, bundle_manifest_json, primary_bundle_path
  FROM reporting_render_bundles
 WHERE release_id::text = $1
`, releaseID).Scan(&bundleSHA, &manifestBytes, &primaryPath); err != nil {
		t.Fatalf("query release render bundle: %v", err)
	}
	if primaryPath == "" {
		t.Fatalf("release render bundle must record a primary path")
	}
	var manifest map[string]any
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		t.Fatalf("decode render bundle manifest: %v", err)
	}
	return bundleSHA, manifest
}

func requireReleaseTokenArtifacts(t testing.TB, db *sql.DB, releaseID string) (string, string, map[string]any, map[string]any) {
	t.Helper()
	rows, err := db.Query(`
SELECT role, bundle_path, file_sha256, inline_bytes
  FROM reporting_render_bundle_files
 WHERE release_id::text = $1
   AND role IN ('token_manifest', 'sensitive_reveal_map')
 ORDER BY role
`, releaseID)
	if err != nil {
		t.Fatalf("query token artifacts: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var tokenSHA string
	var revealSHA string
	var tokenManifest map[string]any
	var revealMap map[string]any
	for rows.Next() {
		var role string
		var path string
		var fileSHA string
		var inlineBytes []byte
		if err := rows.Scan(&role, &path, &fileSHA, &inlineBytes); err != nil {
			t.Fatalf("scan token artifact: %v", err)
		}
		var decoded map[string]any
		if err := json.Unmarshal(inlineBytes, &decoded); err != nil {
			t.Fatalf("decode %s artifact at %s: %v", role, path, err)
		}
		switch role {
		case "token_manifest":
			if path != "validation/token-manifest.json" || decoded["schema_id"] != reporting.RedactionTokenManifestSchemaID {
				t.Fatalf("unexpected token manifest artifact: path=%q sha=%q decoded=%#v", path, fileSHA, decoded)
			}
			tokenSHA = fileSHA
			tokenManifest = decoded
		case "sensitive_reveal_map":
			if path != "internal/reveal-map.json" || decoded["schema_id"] != reporting.RedactionRevealMapSchemaID {
				t.Fatalf("unexpected reveal map artifact: path=%q sha=%q decoded=%#v", path, fileSHA, decoded)
			}
			revealSHA = fileSHA
			revealMap = decoded
		default:
			t.Fatalf("unexpected token artifact role %q", role)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate token artifacts: %v", err)
	}
	if tokenSHA == "" || revealSHA == "" || tokenManifest == nil || revealMap == nil {
		t.Fatalf("release token artifacts incomplete: token_sha=%q reveal_sha=%q token=%#v reveal=%#v", tokenSHA, revealSHA, tokenManifest, revealMap)
	}
	return tokenSHA, revealSHA, tokenManifest, revealMap
}

func requireBundleManifestFile(t testing.TB, manifest map[string]any, role string, path string, sha string, requiredForRelease bool) {
	t.Helper()
	files, ok := manifest["files"].([]any)
	if !ok {
		t.Fatalf("bundle manifest files must be an array: %#v", manifest)
	}
	for _, item := range files {
		file, ok := item.(map[string]any)
		if !ok || file["role"] != role {
			continue
		}
		if file["path"] != path || file["sha256"] != sha || file["required_for_release"] != requiredForRelease {
			t.Fatalf("bundle manifest file mismatch for role %q: got %#v want path=%q sha=%q required=%v", role, file, path, sha, requiredForRelease)
		}
		return
	}
	t.Fatalf("bundle manifest missing role %q: %#v", role, manifest)
}

func requireTokenManifestEntry(t testing.TB, tokenManifest map[string]any, stableSubjectRef string) {
	t.Helper()
	if tokenManifest["schema_id"] != reporting.RedactionTokenManifestSchemaID {
		t.Fatalf("unexpected token manifest schema: %#v", tokenManifest)
	}
	entries, ok := tokenManifest["entries"].([]any)
	if !ok || len(entries) == 0 {
		t.Fatalf("token manifest must contain entries: %#v", tokenManifest)
	}
	for _, item := range entries {
		entry, ok := item.(map[string]any)
		if !ok || entry["stable_subject_ref"] != stableSubjectRef {
			continue
		}
		displayToken, _ := entry["display_token"].(string)
		if !strings.HasPrefix(displayToken, "SUBJECT-") || entry["token_id"] == "" {
			t.Fatalf("token manifest entry must expose stable non-reversible token fields: %#v", entry)
		}
		if _, ok := entry["original_value"]; ok {
			t.Fatalf("token manifest must not expose original values: %#v", entry)
		}
		return
	}
	t.Fatalf("token manifest missing stable subject %q: %#v", stableSubjectRef, tokenManifest)
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
