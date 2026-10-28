package reporting_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSnapshotReplayAndReleaseProvenanceAreStable_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-reporting-provenance", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-incident",
		"incident_key":  "IR-REPORTING-01",
		"title":         "Reporting Incident",
		"description":   strings.Repeat("external recipient narrative ", 12),
		"severity":      "high",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)
	addUnpublishedReportingColumns(t, harness.DB)
	fixture := seedReportingWorkbookFixture(t, harness.DB, incidentID, adminUserID)
	assignment := seedReportingAssignment(t, harness.DB, incidentID, adminUserID, fixture)
	resolveReportingFixtureMention(t, harness.DB, adminUserID, fixture)
	if _, err := harness.DB.Exec(`UPDATE timeline_events SET activity_utc_text='2026-05-02T01:02:03.000001Z', activity_local_text='2020-01-01T00:00+01:00', date_entered_text='2026-05-02T00:30+02:00' WHERE record_id=$1`, fixture["timeline"]); err != nil {
		t.Fatal(err)
	}

	createSnapshot := func(clientTxnID string, watermark *string) *http.Response {
		t.Helper()
		body := map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": clientTxnID,
		}
		if watermark != nil {
			body["source_change_set_high_watermark"] = *watermark
		}
		return httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/snapshots",
			body,
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
	}
	firstSnapshot := createSnapshot("txn-reporting-snapshot", nil)
	firstSnapshotJob := httptestx.RequireSuccessEnvelope(t, firstSnapshot, http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, firstSnapshotJob, "snapshot")
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	assertSnapshotPublicationCandidate(t, harness, uuid.MustParse(firstSnapshotJob["job_id"].(string)), uuid.MustParse(snapshotID))
	if snapshot["source_capture_sha256"] == "" || snapshot["source_change_set_high_watermark"] == "" {
		t.Fatalf("snapshot must expose immutable export model and source boundary provenance: %#v", snapshot)
	}
	initialWatermark := snapshot["source_change_set_high_watermark"].(string)
	if !strings.HasPrefix(initialWatermark, reporting.SourceBoundaryTokenPrefix) {
		t.Fatalf("snapshot source boundary must use v3 token prefix, got %q", initialWatermark)
	}
	requireSnapshotBoundaryJSON(t, harness.DB, snapshotID, initialWatermark, incidentID)
	exportModel := requireSnapshotExportModel(t, harness.DB, snapshotID)
	requireExportModelCoverage(t, exportModel, fixture)
	requireClosedReportingSourceFields(t, exportModel, fixture)
	requireCoreRecordSummaries(t, harness.DB, exportModel, incidentID)
	requireFrozenTimelineSort(t, harness.DB, exportModel, fixture["timeline"])
	requireFrozenDisclosureFacts(t, harness.DB, exportModel, assignment, snapshotID, fixture)
	requireSnapshotSourceIdentities(t, exportModel, snapshotID)
	binding, ok := exportModel["reference_packs"].(map[string]any)
	if !ok {
		t.Fatal("snapshot omitted Reference Pack binding")
	}
	if !strings.HasPrefix(binding["pack_set_id"].(string), "rpset_") || len(binding["pack_set_sha256"].(string)) != 64 || len(binding["provenance"].([]any)) != 3 {
		t.Fatalf("invalid Reference Pack binding: %#v", binding)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reference_pack_pins WHERE owner_kind='reporting_snapshot_job' AND owner_id=$1 AND pack_set_id=$2`, firstSnapshotJob["job_id"], binding["pack_set_id"]); got != 1 {
		t.Fatal("snapshot admission did not pin its exact set")
	}

	scenariotest.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replaySnapshot := createSnapshot("txn-reporting-snapshot", nil)
	replaySnapshotJob := httptestx.RequireSuccessEnvelope(t, replaySnapshot, http.StatusAccepted)["data"].(map[string]any)
	if replaySnapshotJob["job_id"] != firstSnapshotJob["job_id"] {
		t.Fatalf("snapshot replay must return original job: first=%#v replay=%#v", firstSnapshotJob, replaySnapshotJob)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM reporting_snapshots`); got != 1 {
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

	createReleaseResp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "5",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
	requireRenderAdmissionIdentities(t, harness.DB, release, bundleManifest)
	if release["output_sha256"] != bundleManifestSHA || bundleManifest["schema_id"] != reporting.RenderBundleManifestSchemaID {
		t.Fatalf("release output_sha256 must bind render bundle manifest: release=%#v bundle_sha=%q bundle=%#v", release, bundleManifestSHA, bundleManifest)
	}
	if !reflect.DeepEqual(bundleManifest["reference_packs"], binding) {
		t.Fatal("release changed snapshot Reference Pack provenance")
	}
	rendered, manifest := requireReleaseArtifacts(t, harness.DB, releaseID)
	if strings.Contains(rendered, "future_private") || strings.Contains(rendered, "unpublished-source-value") {
		t.Fatal("release exposed an unadmitted source column")
	}
	if refs := release["recipient_partition_refs"].([]any); len(refs) != 1 || refs[0] != "party:"+fixture["party"] {
		t.Fatalf("external recipient partitions must bind the selected party, got %#v", refs)
	}
	if strings.Contains(rendered, "Working note body") {
		t.Fatalf("external output must not package working material: %s", rendered)
	}
	if manifest["redaction_profile_sha256"] != release["redaction_profile_sha256"] {
		t.Fatalf("manifest profile digest must match release provenance: manifest=%#v release=%#v", manifest["redaction_profile_sha256"], release["redaction_profile_sha256"])
	}
	entries := manifestEntriesBySource(t, manifest, exportModel)
	for _, noteEntry := range requireManifestSourceEntries(t, entries, "/notes/"+fixture["note"]) {
		if noteEntry["action"] != reporting.ActionDrop || noteEntry["outcome"] != "dropped" {
			t.Fatalf("every note field must be dropped with stable manifest outcome, got %#v", noteEntry)
		}
	}
	descriptionEntry := entries["/incident/description"]
	if descriptionEntry["action"] != reporting.ActionTruncate || descriptionEntry["outcome"] != "truncated" {
		t.Fatalf("source evidence must be truncated for external release, got %#v", descriptionEntry)
	}
	for _, partyEntry := range requireManifestSourceEntries(t, entries, "/parties/"+fixture["party"]) {
		if partyEntry["outcome"] == "dropped_disclosure_partition" {
			t.Fatalf("selected party field must be retained, got %#v", partyEntry)
		}
	}

	tokenizedReleaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-tokenized-review",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.TokenizedRedactionProfileID,
			"redaction_profile_version": "5",
			"release_scope":             reporting.ReleaseScopeInternalReview,
			"output_kind":               reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	tokenizedReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, tokenizedReleaseJob, "release")
	tokenizedRelease := requireRelease(t, harness, adminLogin, tokenizedReleaseID)
	if tokenizedRelease["redaction_profile_id"] != reporting.TokenizedRedactionProfileID {
		t.Fatalf("tokenized review release must bind tokenized profile, got %#v", tokenizedRelease)
	}
	tokenizedBundleSHA, tokenizedBundleManifest := requireReleaseBundle(t, harness.DB, tokenizedReleaseID)
	if tokenizedRelease["output_sha256"] != tokenizedBundleSHA {
		t.Fatalf("tokenized release output must bind bundle manifest: release=%#v bundle_sha=%q", tokenizedRelease["output_sha256"], tokenizedBundleSHA)
	}
	tokenArtifact, revealArtifact := requireReleaseTokenArtifacts(t, harness.DB, tokenizedReleaseID)
	assertPrivateRevealReadiness(t, harness, tokenizedReleaseID, false)
	assertNativeReleaseReadiness(t, harness, tokenizedReleaseID)
	tokenManifestSHA := tokenArtifact.objectSHA
	tokenManifest, revealMap := tokenArtifact.object, revealArtifact.object
	if tokenManifest["redaction_manifest_sha256"] != tokenizedBundleManifest["redaction_manifest_sha256"] || tokenManifest["redaction_manifest_sha256"] != tokenizedRelease["redaction_manifest_sha256"] {
		t.Fatal("token manifest must bind the same redaction decisions as the bundle and release")
	}
	if tokenizedBundleManifest["token_manifest_sha256"] != tokenManifestSHA {
		t.Fatalf("bundle manifest must bind token manifest: bundle=%#v token_sha=%q", tokenizedBundleManifest, tokenManifestSHA)
	}
	requireBundleManifestFile(t, tokenizedBundleManifest, "token_manifest", "validation/token-manifest.json", tokenArtifact.fileSHA, true)
	for _, item := range tokenizedBundleManifest["files"].([]any) {
		file := item.(map[string]any)
		if file["role"] == "sensitive_reveal_map" || file["path"] == "internal/reveal-map.json" {
			t.Fatal("reveal map was exposed in the ordinary bundle manifest")
		}
	}
	requireTokenManifestEntry(t, tokenManifest, "subject:party:record:"+fixture["party"])
	if revealMap["schema_id"] != reporting.RedactionRevealMapSchemaID ||
		revealMap["token_manifest_sha256"] != tokenManifestSHA {
		t.Fatalf("reveal map must be internal and token-manifest-bound: %#v", revealMap)
	}

	liveNotesBeforePartitionedRelease := queryLiveWorkbookRowsJSON(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1")
	createPartitionedRelease := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-party",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "5",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"], "party:" + fixture["party"]},
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
	partitionedEntries := manifestEntriesBySource(t, partitionedManifest, exportModel)
	for _, entry := range requireManifestSourceEntries(t, partitionedEntries, "/parties/"+fixture["party"]) {
		if entry["outcome"] == "dropped" {
			t.Fatal("requested party field must be retained in partitioned external release")
		}
		refs, ok := entry["disclosure_partition_refs"].([]any)
		if !ok || len(refs) != 1 || refs[0] != partitionRefs[0] {
			t.Fatal("retained party field lost its exact admitted disclosure binding")
		}
	}
	secondPartitionedRelease := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-release-party-supersede",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "5",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
	// Loss of pinned content blocks new derivation without changing immutable
	// output or preventing replay of a previously committed admission.
	if _, err := harness.DB.Exec(`UPDATE reference_pack_objects SET available=false WHERE object_id IN (SELECT object_id FROM reference_pack_object_refs WHERE owner_kind='version' AND logical_path='notices/LICENSE.txt')`); err != nil {
		t.Fatal(err)
	}
	releaseRequest := map[string]any{
		"snapshot_id": snapshotID, "client_txn_id": "missing-reference-new-render",
		"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		"redaction_profile_id": reporting.ExternalRedactionProfileID, "redaction_profile_version": "5",
		"release_scope": "external_release", "output_kind": reporting.OutputKindSlidev, "recipient_partition_refs": []string{"party:" + fixture["party"]},
	}
	submit := func() *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/releases", releaseRequest, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
	}
	httptestx.RequireErrorEnvelope(t, submit(), http.StatusConflict, "required_reference_pack_unavailable")
	releaseRequest["client_txn_id"] = "txn-reporting-release-party-supersede"
	replay := httptestx.RequireSuccessEnvelope(t, submit(), http.StatusAccepted)["data"].(map[string]any)
	if replay["job_id"] != secondPartitionedJob["job_id"] {
		t.Fatal("lost references changed committed admission replay")
	}
}

func TestExternalReleaseApprovalPublishAndStateConflicts_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-reporting-lifecycle", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	reviewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "report-reviewer@example.test", "Report Reviewer", "ReviewerPass1!", false, false, true)
	editorID := flowtest.SeedLocalUserFlags(t, harness.DB, "report-editor@example.test", "Report Editor", "EditorPass1!", false, false, true)
	reviewerSession, reviewerCSRF := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "report-reviewer@example.test", "ReviewerPass1!", nil)
	reviewerLogin := flowtest.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	editorSession, editorCSRF := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "report-editor@example.test", "EditorPass1!", nil)
	editorLogin := flowtest.LoginResult{SessionCookie: editorSession, CSRFCookie: editorCSRF}

	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-lifecycle-incident",
		"incident_key":  "IR-REPORTING-02",
		"title":         "Reporting Lifecycle Incident",
		"description":   "approval lifecycle source text",
	})
	incidentID := incident["incident_id"].(string)
	fixture := seedReportingWorkbookFixture(t, harness.DB, incidentID, adminUserID)
	seedReportingAssignment(t, harness.DB, incidentID, adminUserID, fixture)
	assertUnresolvedReportingMentionDenied(t, harness, adminLogin, incidentID, fixture)
	resolveReportingFixtureMention(t, harness.DB, adminUserID, fixture)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-editor-membership",
		"user_id":       editorID,
		"role":          "editor",
	})

	snapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-lifecycle-snapshot",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")

	releaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-lifecycle-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.ExternalRedactionProfileID,
			"redaction_profile_version": "5",
			"release_scope":             "external_release",
			"output_kind":               reporting.OutputKindSlidev,
			"recipient_partition_refs":  []string{"party:" + fixture["party"]},
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")

	editorApprove := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-editor-approve",
			"reason":        "editor cannot approve release",
		},
		httptestx.WithCookies(editorLogin.SessionCookie, editorLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, editorLogin.CSRFCookie.Value),
	)
	editorError := httptestx.RequireErrorEnvelope(t, editorApprove, http.StatusConflict, "release_approval_rejected")
	if editorError["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "actor_lacks_approval_role" {
		t.Fatalf("editor approval must use release approval rejection reason, got %#v", editorError)
	}

	reviewerApprove := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-reviewer-approve",
			"reason":        "reviewed for release",
		},
		httptestx.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	reviewerBody := httptestx.RequireSuccessEnvelope(t, reviewerApprove, http.StatusOK)["data"].(map[string]any)
	if reviewerBody["release_state"] != reporting.ReleaseStatePendingApproval {
		t.Fatalf("reviewer approval alone must leave external release pending, got %#v", reviewerBody["release_state"])
	}

	adminApprove := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/approve",
		map[string]any{
			"client_txn_id": "txn-reporting-admin-approve",
			"reason":        "approved for publication",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	adminBody := httptestx.RequireSuccessEnvelope(t, adminApprove, http.StatusOK)["data"].(map[string]any)
	if adminBody["release_state"] != reporting.ReleaseStateApproved || adminBody["approved_at"] == nil {
		t.Fatalf("distinct reviewer and admin approvals must approve external release, got %#v", adminBody)
	}

	publish := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish",
			"reason":        "publish",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	publishBody := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	if publishBody["release_state"] != reporting.ReleaseStatePublished || publishBody["published_at"] == nil {
		t.Fatalf("approved release must publish synchronously, got %#v", publishBody)
	}

	republish := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-publish-again",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	errorBody := httptestx.RequireErrorEnvelope(t, republish, http.StatusConflict, "release_state_conflict")
	if errorBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "already_published" {
		t.Fatalf("republish must fail with stable already_published reason, got %#v", errorBody)
	}
}

func TestBoundaryReplayDefaultsAndActionIdempotency_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-reporting-idempotency", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-idempotency-incident",
		"incident_key":  "IR-REPORTING-03",
		"title":         "Reporting Idempotency Incident",
		"description":   "recipient-visible source text",
	})
	incidentID := incident["incident_id"].(string)

	snapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-idempotency-snapshot",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")
	snapshot := requireSnapshot(t, harness, adminLogin, snapshotID)
	watermark := snapshot["source_change_set_high_watermark"].(string)

	scenariotest.PatchIncident(t, harness.Server, adminLogin, incidentID, map[string]any{
		"base_incident_version": 1,
		"current_phase":         "containment",
	})
	replayWithExplicitBoundary := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":                      incidentID,
			"client_txn_id":                    "txn-reporting-idempotency-snapshot",
			"source_change_set_high_watermark": watermark,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, replayWithExplicitBoundary, http.StatusConflict, "client_txn_conflict")
	replayJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{"incident_id": incidentID, "client_txn_id": "txn-reporting-idempotency-snapshot"},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusAccepted)["data"].(map[string]any)
	if replayJob["job_id"] != snapshotJob["job_id"] {
		t.Fatal("exact omitted-boundary replay allocated another Job")
	}
	httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{"incident_id": incidentID, "client_txn_id": "txn-reporting-stale-explicit-snapshot", "source_change_set_high_watermark": watermark},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusConflict, "snapshot_source_boundary_conflict")
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reporting_snapshots`); got != 1 {
		t.Fatal("snapshot replay or conflict published another capture", got)
	}

	unknownRecipientRequest := map[string]any{
		"snapshot_id": snapshotID, "client_txn_id": "txn-reporting-unknown-recipient",
		"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		"redaction_profile_id": reporting.ExternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
		"output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeExternal,
		"recipient_partition_refs": []string{"party:missing"},
	}
	unknownRecipientJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases", unknownRecipientRequest,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusAccepted)["data"].(map[string]any)
	unknownRecipientTerminal := requireJobStatus(t, harness, adminLogin, unknownRecipientJob["job_id"].(string), "failed")
	unknownRecipientSummary := unknownRecipientTerminal["error_summary"].(map[string]any)
	unknownRecipientDetails := unknownRecipientSummary["details"].(map[string]any)
	if unknownRecipientSummary["code"] != "release_render_failed" || unknownRecipientDetails["reason_code"] != "unknown_recipient_partition" || unknownRecipientDetails["release_id"] != nil {
		t.Fatal("recipient preparation failure lost its reason or fabricated a release", unknownRecipientSummary)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reporting_releases`); got != 0 {
		t.Fatal("unprepared recipient failure published a release", got)
	}
	unknownRecipientReplay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases", unknownRecipientRequest,
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusAccepted)["data"].(map[string]any)
	if unknownRecipientReplay["job_id"] != unknownRecipientJob["job_id"] {
		t.Fatal("failed preparation replay allocated another admission")
	}

	remoteAssetIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-idempotency-remote-asset-incident",
		"incident_key":  "IR-REPORTING-03-REMOTE",
		"title":         "Reporting Remote Asset Incident",
		"description":   "![remote asset](https://cdn.example.test/report.png)",
	})
	remoteAssetIncidentID := remoteAssetIncident["incident_id"].(string)
	remoteAssetSnapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   remoteAssetIncidentID,
			"client_txn_id": "txn-reporting-idempotency-remote-asset-snapshot",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	remoteAssetSnapshotID := requireSucceededJobResourceID(t, harness, adminLogin, remoteAssetSnapshotJob, "snapshot")
	remoteAssetReleaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               remoteAssetSnapshotID,
			"client_txn_id":             "txn-reporting-idempotency-remote-asset-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "5",
			"output_kind":               reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	remoteAssetReleaseID := requireFailedReleaseJob(t, harness, adminLogin, remoteAssetReleaseJob, "remote_asset_reference_invalid")
	assertFailedReleaseReadiness(t, harness, remoteAssetReleaseID)
	remoteAssetRelease := requireRelease(t, harness, adminLogin, remoteAssetReleaseID)
	if remoteAssetRelease["release_state"] != reporting.ReleaseStateRenderFailed ||
		remoteAssetRelease["render_failed_reason_code"] != "remote_asset_reference_invalid" ||
		remoteAssetRelease["output_sha256"] != nil ||
		remoteAssetRelease["redaction_manifest_sha256"] != nil ||
		remoteAssetRelease["output_media_type"] != nil {
		t.Fatalf("self-contained output validation must persist render-failed release, got %#v", remoteAssetRelease)
	}
	failedReleaseID := remoteAssetReleaseID
	unknownReleaseID := "00000000-0000-0000-0000-000000000999"
	for _, action := range []string{"approve", "publish", "invalidate"} {
		malformed := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+failedReleaseID+"/"+action,
			[]string{"not", "an", "object"},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, malformed, http.StatusBadRequest, "invalid_release_request")

		missing := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+unknownReleaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-missing-" + action},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, missing, http.StatusNotFound, "release_not_found")
	}
	for _, action := range []string{"approve", "publish", "invalidate"} {
		resp := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+failedReleaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-render-failed-" + action},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "release_state_conflict")
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "render_failed" {
			t.Fatalf("%s on render_failed release must return render_failed conflict, got %#v", action, body)
		}
	}

	releaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":      snapshotID,
			"client_txn_id":    "txn-reporting-idempotency-release",
			"template_id":      reporting.DefaultTemplateID,
			"template_version": reporting.DefaultTemplateVersion,
			"output_kind":      reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_scope"] != reporting.ReleaseScopeInternalDraft || release["release_state"] != reporting.ReleaseStateApproved {
		t.Fatalf("omitted release_scope must default to approved internal draft, got %#v", release)
	}
	if release["redaction_profile_id"] != reporting.InternalRedactionProfileID || release["redaction_profile_version"] != reporting.DefaultRedactionProfileVersion || release["redaction_profile_sha256"] == "" {
		t.Fatal("omitted profile did not retain its complete admission identity")
	}
	explicitReplay := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases", map[string]any{
			"snapshot_id": snapshotID, "client_txn_id": "txn-reporting-idempotency-release",
			"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
			"redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
			"output_kind": reporting.OutputKindSlidev,
		}, httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusAccepted)["data"].(map[string]any)
	if explicitReplay["job_id"] != releaseJob["job_id"] {
		t.Fatal("explicit default created another admission for an identical request")
	}

	publish := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        "",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	firstPublish := httptestx.RequireSuccessEnvelope(t, publish, http.StatusOK)["data"].(map[string]any)
	replayPublish := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{
			"client_txn_id": "txn-reporting-idempotency-publish",
			"reason":        nil,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	secondPublish := httptestx.RequireSuccessEnvelope(t, replayPublish, http.StatusOK)["data"].(map[string]any)
	if firstPublish["published_at"] != secondPublish["published_at"] || secondPublish["release_state"] != reporting.ReleaseStatePublished {
		t.Fatalf("publish replay with null/empty reason must return original published resource: first=%#v second=%#v", firstPublish, secondPublish)
	}
}

func TestExactShapesAndRouteScopedVisibility_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-reporting-shape-auth", TestRouteMode: httptestx.TestRouteModeDisabled})

	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	reviewerID := flowtest.SeedLocalUserFlags(t, harness.DB, "shape-reviewer@example.test", "Shape Reviewer", "ShapeReviewer1!", false, false, true)
	flowtest.SeedLocalUserFlags(t, harness.DB, "shape-outsider@example.test", "Shape Outsider", "ShapeOutsider1!", false, false, true)
	flowtest.SeedLocalUserFlags(t, harness.DB, "shape-deployment-admin@example.test", "Shape Deployment Admin", "ShapeDeployAdmin1!", false, true, true)
	reviewerSession, reviewerCSRF := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "shape-reviewer@example.test", "ShapeReviewer1!", nil)
	reviewerLogin := flowtest.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	outsiderSession, outsiderCSRF := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "shape-outsider@example.test", "ShapeOutsider1!", nil)
	outsiderLogin := flowtest.LoginResult{SessionCookie: outsiderSession, CSRFCookie: outsiderCSRF}
	deploymentAdminSession, deploymentAdminCSRF := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, "shape-deployment-admin@example.test", "ShapeDeployAdmin1!", nil)
	deploymentAdminLogin := flowtest.LoginResult{SessionCookie: deploymentAdminSession, CSRFCookie: deploymentAdminCSRF}

	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-reporting-shape-incident",
		"incident_key":  "IR-REPORTING-04",
		"title":         "Reporting Shape Incident",
		"description":   "resource shape and visibility source text",
	})
	incidentID := incident["incident_id"].(string)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-reporting-shape-reviewer-membership",
		"user_id":       reviewerID,
		"role":          "reviewer",
	})
	_ = scenariotest.CreateIncident(t, harness.Server, outsiderLogin, map[string]any{
		"client_txn_id": "txn-reporting-shape-other-incident",
		"incident_key":  "IR-REPORTING-04-OTHER",
		"title":         "Other Reporting Incident",
	})

	snapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-snapshot",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")
	_ = requireSnapshot(t, harness, adminLogin, snapshotID)

	paginatedSnapshot := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID+"?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, paginatedSnapshot, http.StatusBadRequest, "invalid_pagination_request")

	hiddenSnapshot := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		httptestx.WithCookies(outsiderLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, hiddenSnapshot, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminHiddenSnapshot := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenSnapshot, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminCreateSnapshot := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-deployment-admin-snapshot",
		},
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminCreateSnapshot, http.StatusNotFound, "incident_not_found")
	deploymentAdminReadJob := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/jobs/"+snapshotJob["job_id"].(string),
		nil,
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminReadJob, http.StatusNotFound, "job_not_found")
	deploymentAdminCancelJob := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/jobs/"+snapshotJob["job_id"].(string)+"/cancel",
		map[string]any{"client_txn_id": "txn-reporting-shape-deployment-admin-cancel"},
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminCancelJob, http.StatusNotFound, "job_not_found")

	hiddenReleaseCreate := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-hidden-release-create",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "5",
			"output_kind":               reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(outsiderLogin.SessionCookie, outsiderLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, outsiderLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, hiddenReleaseCreate, http.StatusNotFound, "snapshot_not_found")
	deploymentAdminHiddenReleaseCreate := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-deployment-admin-release-create",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "5",
			"output_kind":               reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenReleaseCreate, http.StatusNotFound, "snapshot_not_found")

	releaseCreateBody := func(clientTxnID string, mutate func(map[string]any)) map[string]any {
		body := map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             clientTxnID,
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "5",
			"output_kind":               reporting.OutputKindSlidev,
		}
		if mutate != nil {
			mutate(body)
		}
		return body
	}
	postReleaseCreate := func(actor flowtest.LoginResult, body map[string]any) *http.Response {
		t.Helper()
		return httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases",
			body,
			httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
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

	unsupportedSnapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{
			"incident_id":   incidentID,
			"client_txn_id": "txn-reporting-shape-unsupported-derivation-snapshot",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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

	releaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id":               snapshotID,
			"client_txn_id":             "txn-reporting-shape-release",
			"template_id":               reporting.DefaultTemplateID,
			"template_version":          reporting.DefaultTemplateVersion,
			"redaction_profile_id":      reporting.InternalRedactionProfileID,
			"redaction_profile_version": "5",
			"output_kind":               reporting.OutputKindSlidev,
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["release_scope"] != reporting.ReleaseScopeInternalDraft || release["output_kind"] != reporting.OutputKindSlidev {
		t.Fatalf("release resource must expose resolved closed vocabularies, got %#v", release)
	}

	paginatedRelease := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"?limit=1",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, paginatedRelease, http.StatusBadRequest, "invalid_pagination_request")

	hiddenRelease := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		httptestx.WithCookies(outsiderLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, hiddenRelease, http.StatusNotFound, "release_not_found")
	deploymentAdminHiddenRelease := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenRelease, http.StatusNotFound, "release_not_found")

	for _, action := range []string{"approve", "publish", "invalidate"} {
		paginatedAction := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action+"?limit=1",
			map[string]any{"client_txn_id": "txn-reporting-shape-paginated-" + action},
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, paginatedAction, http.StatusBadRequest, "invalid_pagination_request")

		hiddenAction := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-shape-hidden-" + action},
			httptestx.WithCookies(outsiderLogin.SessionCookie, outsiderLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, outsiderLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, hiddenAction, http.StatusNotFound, "release_not_found")

		deploymentAdminHiddenAction := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/"+action,
			map[string]any{"client_txn_id": "txn-reporting-shape-deployment-admin-" + action},
			httptestx.WithCookies(deploymentAdminLogin.SessionCookie, deploymentAdminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, deploymentAdminLogin.CSRFCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, deploymentAdminHiddenAction, http.StatusNotFound, "release_not_found")
	}

	reviewerPublish := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{"client_txn_id": "txn-reporting-shape-reviewer-publish"},
		httptestx.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerPublish, http.StatusForbidden, "authorization_denied")

	published := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/publish",
		map[string]any{"client_txn_id": "txn-reporting-shape-admin-publish"},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusOK)["data"].(map[string]any)
	requireReleaseResourceShape(t, published)
	if got := requireRelease(t, harness, adminLogin, releaseID); !reflect.DeepEqual(published, got) {
		t.Fatalf("publish action response must match GET release resource: action=%#v get=%#v", published, got)
	}

	reviewerInvalidate := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/invalidate",
		map[string]any{"client_txn_id": "txn-reporting-shape-reviewer-invalidate"},
		httptestx.WithCookies(reviewerLogin.SessionCookie, reviewerLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, reviewerLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, reviewerInvalidate, http.StatusForbidden, "authorization_denied")

	invalidated := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID+"/invalidate",
		map[string]any{"client_txn_id": "txn-reporting-shape-admin-invalidate"},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
	ids["investigative_query"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), artifact AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, 'investigative_query', 'Excluded reporting query', 'Query working material', $2
      FROM rec
    RETURNING record_id
), query AS (
    INSERT INTO artifact_investigative_queries (
        record_id, incident_id, query_id, platform, purpose, query_text, created_by_user_id
    )
    SELECT record_id, $1, 'reporting-query-excluded', 'edr', 'coverage', 'excluded synthetic query', $2
      FROM artifact
    RETURNING record_id
)
SELECT record_id::text FROM query
`, incidentID, actorID)
	seedReportingArtifactProjection(t, db, ids["investigative_query"])
	ids["forensic_keyword"] = querySeedID(t, db, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'artifact', $2, $2)
    RETURNING record_id
), artifact AS (
    INSERT INTO artifacts (record_id, incident_id, artifact_type, title, body, created_by_user_id)
    SELECT record_id, $1, 'forensic_keyword', 'Excluded reporting keyword', 'Keyword working material', $2
      FROM rec
    RETURNING record_id
), keyword AS (
    INSERT INTO artifact_forensic_keywords (
        record_id, incident_id, keyword_id, pattern, reason, match_mode, case_sensitive
    )
    SELECT record_id, $1, 'reporting-keyword-excluded', 'excluded-pattern', 'coverage', 'literal', false
      FROM artifact
    RETURNING record_id
)
SELECT record_id::text FROM keyword
`, incidentID, actorID)
	seedReportingArtifactProjection(t, db, ids["forensic_keyword"])
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
VALUES ($1, 'host', 'timeline.activity_synopsis_text', 'manual_entry', 'summary:1', 'host-reporting-01', 'host-reporting-01', 'unresolved', 1, $2)
`, ids["timeline"], actorID); err != nil {
		t.Fatalf("seed reporting entity mention: %v", err)
	}
	return ids
}

// Positive external-lifecycle evidence uses an explicitly resolved source
// mention. Unresolved mentions remain in the ordinary seed and denial cases.
func resolveReportingFixtureMention(t testing.TB, db *sql.DB, actorID string, ids map[string]string) {
	t.Helper()
	if _, err := db.Exec(`UPDATE entity_mentions SET resolution_status='resolved',
resolved_record_id=$2, resolved_by_user_id=$3, resolved_at=created_at,
resolution_method='manual', row_version=row_version+1 WHERE source_record_id=$1`, ids["timeline"], ids["host"], actorID); err != nil {
		t.Fatal(err)
	}
}

func assertUnresolvedReportingMentionDenied(t *testing.T, harness *appsupport.ServerHarness, actor flowtest.LoginResult, incidentID string, ids map[string]string) {
	t.Helper()
	submit := func(route string, body map[string]any) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+route, body,
			httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value)), http.StatusAccepted)["data"].(map[string]any)
	}
	snapshot := submit("/api/v1/snapshots", map[string]any{"incident_id": incidentID, "client_txn_id": "unresolved-disclosure-snapshot"})
	snapshotID := requireSucceededJobResourceID(t, harness, actor, snapshot, "snapshot")
	job := submit("/api/v1/releases", map[string]any{
		"snapshot_id": snapshotID, "client_txn_id": "unresolved-disclosure-release",
		"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
		"redaction_profile_id": reporting.ExternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
		"release_scope": "external_release", "output_kind": reporting.OutputKindSlidev,
		"recipient_partition_refs": []string{"party:" + ids["party"]},
	})
	releaseID := requireFailedReleaseJob(t, harness, actor, job, "disclosure_partition_unresolved")
	assertFailedReleaseReadiness(t, harness, releaseID)
	release := requireRelease(t, harness, actor, releaseID)
	if release["output_sha256"] != nil || release["redaction_manifest_sha256"] != nil {
		t.Fatal("unresolved disclosure published successful output identities")
	}
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
    investigative_query_query_id,
    investigative_query_platform,
    investigative_query_purpose,
    investigative_query_query_text,
    investigative_query_created_by_user_id,
    investigative_query_created_at,
    investigative_query_created_day,
    forensic_keyword_keyword_id,
    forensic_keyword_pattern,
    forensic_keyword_reason,
    forensic_keyword_match_mode,
    forensic_keyword_case_sensitive,
    forensic_keyword_created_at,
    forensic_keyword_created_day,
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
    q.query_id,
    q.platform,
    q.purpose,
    q.query_text,
    q.created_by_user_id,
    q.created_at,
    q.created_at::date,
    k.keyword_id,
    k.pattern,
    k.reason,
    k.match_mode,
    k.case_sensitive,
    k.created_at,
    k.created_at::date,
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
  LEFT JOIN artifact_investigative_queries q
    ON q.incident_id = a.incident_id
   AND q.record_id = a.record_id
  LEFT JOIN artifact_forensic_keywords k
    ON k.incident_id = a.incident_id
   AND k.record_id = a.record_id
 WHERE a.record_id::text = $1
`, recordID); err != nil {
		t.Fatalf("seed reporting artifact projection: %v", err)
	}
}

func requireSnapshotExportModel(t testing.TB, db *sql.DB, snapshotID string) map[string]any {
	t.Helper()
	var capture []byte
	var digest string
	if err := db.QueryRow(`SELECT p.request_json->'source_capture',s.source_capture_sha256 FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id::text=$1`, snapshotID).Scan(&capture, &digest); err != nil {
		t.Fatal(err)
	}
	projection, err := reporting.MaterializeSnapshotCapture(context.Background(), capture, digest)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
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

func createWorkbookNote(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, clientTxnID string) {
	t.Helper()
	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.notes.v1/rows",
		map[string]any{
			"client_txn_id": clientTxnID,
			"note.title":    "Boundary note",
			"note.body":     "Boundary note body",
		},
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)
}

func queryLiveWorkbookRowsJSON(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, viewSchemaID string) string {
	t.Helper()
	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/query",
		map[string]any{"sort": []map[string]any{{"field_key": "note.title", "direction": "asc"}}},
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
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
	if model["schema_id"] != reporting.SnapshotModelSchemaID || model["derivation_version"] != reporting.DerivationVersion {
		t.Fatalf("snapshot export model must use current identity, got %#v", model)
	}
	for _, key := range []string{"sections", "records", "relationships", "timeline_events", "subjects", "diagrams", "assets", "support_index", "source_index", "validation_summary"} {
		if _, ok := model[key]; !ok {
			t.Fatalf("snapshot export model missing member %s: %#v", key, model)
		}
	}
	if _, ok := model["fields"]; ok {
		t.Fatalf("snapshot export model must not persist legacy top-level fields[]")
	}
	byPath := exportModelFieldsByPath(t, model)
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
		if byPath[path+"/record_id"] == nil {
			t.Fatalf("snapshot export model missing workbook path %s; paths=%#v", path, byPath)
		}
	}
	for _, path := range []string{
		"/investigative_queries/" + ids["investigative_query"],
		"/forensic_keywords/" + ids["forensic_keyword"],
	} {
		if byPath[path+"/record_id"] != nil {
			t.Fatalf("reporting export admitted explicitly excluded artifact path %s", path)
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
		if byPath[path+"/record_id"]["source_family"] != family {
			t.Fatalf("path %s source_family = %#v want %q", path, byPath[path+"/record_id"]["source_family"], family)
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
		if byPath[path+"/record_id"]["content_class"] != class {
			t.Fatalf("path %s content_class = %#v want %q", path, byPath[path+"/record_id"]["content_class"], class)
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
	supportID := ""
	for _, value := range model["support_index"].([]any) {
		ref := value.(map[string]any)
		if ref["support_kind"] == "evidence_item" && ref["support_target_ref"] == "evidence:"+ids["evidence"] {
			if supportID != "" {
				t.Fatal("Evidence support was duplicated in the index")
			}
			supportID = ref["support_ref_id"].(string)
		}
	}
	if !strings.HasPrefix(supportID, "sup_") || len(supportID) != 68 {
		t.Fatal("Evidence support has no registered snapshot identity")
	}
	supportRefs := byPath["/task_requests/"+ids["task"]+"/record_id"]["support_refs"].([]any)
	if len(supportRefs) != 1 || supportRefs[0] != supportID {
		t.Fatalf("task export field must carry deterministic support refs, got %#v", supportRefs)
	}
	findingSupportRefs := byPath["/findings/"+ids["finding"]+"/record_id"]["support_refs"].([]any)
	if len(findingSupportRefs) != 1 || findingSupportRefs[0] != supportID {
		t.Fatalf("curated finding export field must carry deterministic support refs, got %#v", findingSupportRefs)
	}
	partyPartitions := byPath["/parties/"+ids["party"]+"/record_id"]["disclosure_partition_refs"].([]any)
	if len(partyPartitions) != 1 || partyPartitions[0] != "party:"+ids["party"] {
		t.Fatalf("party export field must carry disclosure partition refs, got %#v", partyPartitions)
	}
}

// Fixture labels keep source-obligation assertions readable. They are authored
// here, not retained selectors or a compatibility reader. Public references
// resolve exclusively through source_index and explicit source ownership.
func exportModelFieldsByPath(t testing.TB, model map[string]any) map[string]map[string]any {
	t.Helper()
	byPath := map[string]map[string]any{}
	for _, sectionItem := range model["sections"].([]any) {
		section := sectionItem.(map[string]any)
		collectExportModelBlockFields(section["blocks"].([]any), byPath)
	}
	aliases := fixtureSourceAliases(t, model, byPath)
	result := map[string]map[string]any{}
	for source, field := range byPath {
		alias, exists := aliases[source]
		if !exists || result[alias] != nil {
			t.Fatal("fixture field has no unique source-index binding")
		}
		result[alias] = field
	}
	return result
}

func fixtureSourceAliases(t testing.TB, model map[string]any, fields map[string]map[string]any) map[string]string {
	t.Helper()
	roots := map[string]string{
		"record_envelope": "record_envelopes", "timeline_event": "timeline", "host": "hosts", "identity": "identities", "party": "parties", "evidence": "evidence", "task_request": "task_requests", "decision": "decisions", "note": "notes", "finding_hypothesis": "findings", "comm_log": "comm_log", "handoff": "handoffs", "status_review": "status_reviews", "lesson": "lessons", "record_link": "relationships", "record_tag": "tags", "entity_mention": "entity_mentions",
	}
	aliases := map[string]string{}
	incidentRefs := map[string]string{}
	for _, raw := range model["source_index"].([]any) {
		ref := raw.(map[string]any)
		id := ref["source_ref_id"].(string)
		if ref["source_path"] == nil {
			continue
		}
		field, exists := fields[id]
		if !exists {
			t.Fatal("indexed field absent from fixture sections")
		}
		family := ref["source_family"].(string)
		field["source_family"] = family
		if family == "incident_metadata" {
			incidentRefs[ref["source_path"].(string)] = id
			continue
		}
		root, exists := roots[family]
		if !exists {
			t.Fatalf("fixture has no declared label for source family %q", family)
		}
		aliases[id] = "/" + root + "/" + ref["source_record_id"].(string) + "/" + field["field_key"].(string)
	}
	// This fixture authors these five incident scalars in canonical section order.
	// The immutable model does not retain their old provider selectors.
	paths := make([]string, 0, len(incidentRefs))
	for path := range incidentRefs {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	labels := []string{"current_phase", "description", "severity", "status", "title"}
	if len(paths) != len(labels) {
		t.Fatal("fixture incident scalar inventory changed")
	}
	for i, path := range paths {
		aliases[incidentRefs[path]] = "/incident/" + labels[i]
	}
	return aliases
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
				"field_key":                 field["field_key"],
				"path":                      path,
				"source_family":             "incident_metadata",
				"content_class":             block["content_class"],
				"value":                     field["value"],
				"support_refs":              field["support_refs"],
				"disclosure_partition_refs": field["disclosure_partition_refs"],
			}
		}
		collectExportModelBlockFields(block["children"].([]any), byPath)
	}
}

func requireSucceededJobResourceID(t testing.TB, harness *appsupport.ServerHarness, actor flowtest.LoginResult, job map[string]any, wantKind string) string {
	t.Helper()
	jobID := job["job_id"].(string)
	finalJob := requireJobStatus(t, harness, actor, jobID, "succeeded")
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM extension_job_commit_proofs WHERE job_id::text = $1`, jobID); got != 1 {
		t.Fatalf("successful Reporting job %s proof rows = %d want 1", jobID, got)
	}
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

func requireFailedReleaseJob(t testing.TB, harness *appsupport.ServerHarness, actor flowtest.LoginResult, job map[string]any, wantReason string) string {
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

func requireJobStatus(t testing.TB, harness *appsupport.ServerHarness, actor flowtest.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID,
			nil,
			httptestx.WithCookies(actor.SessionCookie),
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

func requireSnapshot(t testing.TB, harness *appsupport.ServerHarness, actor flowtest.LoginResult, snapshotID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/snapshots/"+snapshotID,
		nil,
		httptestx.WithCookies(actor.SessionCookie),
	)
	resource := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	requireSnapshotResourceShape(t, resource)
	return resource
}

func requireRelease(t testing.TB, harness *appsupport.ServerHarness, actor flowtest.LoginResult, releaseID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/releases/"+releaseID,
		nil,
		httptestx.WithCookies(actor.SessionCookie),
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
		"source_capture_sha256",
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

type retainedTokenArtifact struct {
	fileSHA, objectSHA string
	object             map[string]any
}

func requireReleaseTokenArtifacts(t testing.TB, db *sql.DB, releaseID string) (retainedTokenArtifact, retainedTokenArtifact) {
	t.Helper()
	rows, err := db.Query(`
SELECT role, bundle_path, file_sha256, inline_bytes
  FROM reporting_render_bundle_files
 WHERE release_id::text = $1
   AND role = 'token_manifest'
`, releaseID)
	if err != nil {
		t.Fatalf("query token artifacts: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var token, reveal retainedTokenArtifact
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
		fileSum := sha256.Sum256(inlineBytes)
		if hex.EncodeToString(fileSum[:]) != fileSHA {
			t.Fatal("retained file checksum mismatch")
		}
		objectSum := sha256.Sum256(append([]byte(decoded["schema_id"].(string)+"\n"), inlineBytes...))
		artifact := retainedTokenArtifact{fileSHA, hex.EncodeToString(objectSum[:]), decoded}
		if artifact.fileSHA == artifact.objectSHA {
			t.Fatal("file and object hash domains conflated")
		}
		switch role {
		case "token_manifest":
			if path != "validation/token-manifest.json" || decoded["schema_id"] != reporting.RedactionTokenManifestSchemaID {
				t.Fatalf("unexpected token manifest artifact: path=%q sha=%q decoded=%#v", path, fileSHA, decoded)
			}
			token = artifact
		default:
			t.Fatalf("unexpected token artifact role %q", role)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate token artifacts: %v", err)
	}
	var revealBytes []byte
	if err := db.QueryRow(`SELECT reveal_map_sha256,canonical_bytes FROM reporting_release_reveal_maps WHERE release_id::text=$1`, releaseID).Scan(&reveal.objectSHA, &revealBytes); err != nil {
		t.Fatal("private reveal map missing", err)
	}
	if err := json.Unmarshal(revealBytes, &reveal.object); err != nil {
		t.Fatal(err)
	}
	fileSum := sha256.Sum256(revealBytes)
	reveal.fileSHA = hex.EncodeToString(fileSum[:])
	objectSum := sha256.Sum256(append([]byte(reporting.RedactionRevealMapSchemaID+"\n"), revealBytes...))
	if reveal.objectSHA != hex.EncodeToString(objectSum[:]) {
		t.Fatal("private reveal map digest mismatch")
	}
	var sensitiveMembers int
	if err := db.QueryRow(`SELECT count(*) FROM reporting_render_bundle_files WHERE release_id::text=$1 AND (role='sensitive_reveal_map' OR bundle_path LIKE 'internal/%')`, releaseID).Scan(&sensitiveMembers); err != nil || sensitiveMembers != 0 {
		t.Fatal("sensitive bytes remained in ordinary bundle storage", err)
	}
	if token.fileSHA == "" || reveal.fileSHA == "" || token.object == nil || reveal.object == nil {
		t.Fatal("release token artifacts incomplete")
	}
	return token, reveal
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
		if _, ok := entry["canonical_display_value"]; ok {
			t.Fatalf("token manifest must not expose original values: %#v", entry)
		}
		return
	}
	t.Fatalf("token manifest missing stable subject %q: %#v", stableSubjectRef, tokenManifest)
}

func requireManifestSourceEntries(t testing.TB, entries map[string]map[string]any, source string) []map[string]any {
	t.Helper()
	selected := []map[string]any{}
	for path, entry := range entries {
		if strings.HasPrefix(path, source+"/") {
			selected = append(selected, entry)
		}
	}
	if len(selected) == 0 {
		t.Fatalf("manifest has no scalar fields for source %s", source)
	}
	return selected
}

func manifestEntriesBySource(t testing.TB, manifest, model map[string]any) map[string]map[string]any {
	t.Helper()
	byPath := map[string]map[string]any{}
	for _, row := range manifest["entries"].([]any) {
		entry := row.(map[string]any)
		path, ok := entry["export_model_path"].(string)
		if !ok || !strings.HasPrefix(path, "$.sections[") || byPath[path] != nil {
			t.Fatal("manifest is not keyed by exact model address")
		}
		if _, legacy := entry["path"]; legacy {
			t.Fatal("legacy source path emitted")
		}
		byPath[path] = entry
	}
	keyed := func(parent, member, id string) string {
		return parent + "." + member + "['" + strings.ReplaceAll(strings.ReplaceAll(id, "\\", "\\\\"), "'", "\\'") + "']"
	}
	entries := map[string]map[string]any{}
	var visit func(string, string, []any)
	visit = func(parent, member string, blocks []any) {
		for _, raw := range blocks {
			block := raw.(map[string]any)
			path := keyed(parent, member, block["block_id"].(string))
			for _, rawField := range block["fields"].([]any) {
				field := rawField.(map[string]any)
				fieldPath := keyed(path, "fields", field["field_key"].(string))
				entry, exists := byPath[fieldPath]
				if !exists {
					t.Fatal("source field has no model-addressed manifest entry")
				}
				for _, ref := range field["source_refs"].([]any) {
					entries[ref.(string)] = entry
				}
				delete(byPath, fieldPath)
			}
			visit(path, "children", block["children"].([]any))
		}
	}
	for _, raw := range model["sections"].([]any) {
		section := raw.(map[string]any)
		visit(keyed("$", "sections", section["section_id"].(string)), "blocks", section["blocks"].([]any))
	}
	if len(byPath) != 0 {
		t.Fatal("manifest contains an unbound field address")
	}
	fields := map[string]map[string]any{}
	for _, raw := range model["sections"].([]any) {
		collectExportModelBlockFields(raw.(map[string]any)["blocks"].([]any), fields)
	}
	result := map[string]map[string]any{}
	for source, alias := range fixtureSourceAliases(t, model, fields) {
		entry, exists := entries[source]
		if !exists || result[alias] != nil {
			t.Fatal("manifest has no unique fixture source")
		}
		result[alias] = entry
	}
	return result
}

func requireSnapshotSourceIdentities(t testing.TB, value any, snapshotID string) {
	t.Helper()
	count := 0
	var visit func(any)
	visit = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "source_snapshot_id" {
					count++
					if v != snapshotID {
						t.Fatalf("source reference escaped frozen snapshot: %v", v)
					}
				}
				visit(v)
			}
		case []any:
			for _, v := range x {
				visit(v)
			}
		}
	}
	visit(value)
	if count == 0 {
		t.Fatal("snapshot fixture did not exercise source references")
	}
}

func requireRenderAdmissionIdentities(t testing.TB, db *sql.DB, release, bundle map[string]any) {
	t.Helper()
	at, err := time.Parse(time.RFC3339Nano, release["render_admitted_at"].(string))
	if err != nil {
		t.Fatal(err)
	}
	stamp := at.UTC().Format("2006-01-02T15:04:05.000000Z")
	if bundle["release_id"] != release["release_id"] || bundle["snapshot_id"] != release["snapshot_id"] || bundle["preview_attempt_id"] != nil || bundle["bundle_created_at"] != stamp || bundle["render_admitted_at"] != stamp || bundle["admitted_export_model_sha256"] != release["export_model_sha256"] || bundle["export_model_sha256"] == release["export_model_sha256"] {
		t.Fatal("render bundle lost admitted identity or time")
	}
	// Independent standard-library serialization of this ASCII-only literal tuple.
	tuple := map[string]any{"schema_id": "cartulary.reporting_export_model_id.v1", "release_id": release["release_id"], "snapshot_id": release["snapshot_id"], "derivation_version": reporting.DerivationVersion, "render_admitted_at": stamp}
	raw, _ := json.Marshal(tuple)
	sum := sha256.Sum256(append([]byte("cartulary.reporting_export_model_id.v1\n"), raw...))
	expected := "expm_" + hex.EncodeToString(sum[:])
	if bundle["export_model_id"] != expected {
		t.Fatal("render model used a different identity tuple")
	}
	rows, err := db.Query(`SELECT bundle_path,inline_bytes FROM reporting_render_bundle_files WHERE release_id=$1 AND role IN ('toolchain_snapshot','validation_summary','redaction_manifest','deck_model') ORDER BY bundle_path`, release["release_id"])
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var path string
		var raw []byte
		if err := rows.Scan(&path, &raw); err != nil {
			t.Fatal(err)
		}
		var artifact map[string]any
		if err := json.Unmarshal(raw, &artifact); err != nil {
			t.Fatal(err)
		}
		count++
		if path == "intermediate/deck-model.json" {
			if artifact["release_id"] != release["release_id"] || artifact["snapshot_id"] != release["snapshot_id"] || artifact["render_admitted_at"] != stamp || !strings.HasPrefix(artifact["deck_id"].(string), "deck_") {
				t.Fatal("deck lost admitted identity")
			}
		} else if artifact["created_at"] != stamp {
			t.Fatalf("%s has a different generated time", path)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Fatalf("generated artifact coverage=%d", count)
	}
}

func requirePreparedRenderProfile(t testing.TB, harness *appsupport.ServerHarness, jobID uuid.UUID) reporting.RedactionProfile {
	t.Helper()
	var raw []byte
	if err := harness.DB.QueryRow(`SELECT request_json FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var admission struct {
		SnapshotID string                     `json:"snapshot_id"`
		Profile    reporting.RedactionProfile `json:"authored_redaction_profile"`
		Recipients []string                   `json:"recipient_partition_refs"`
	}
	if err := json.Unmarshal(raw, &admission); err != nil {
		t.Fatal(err)
	}
	var capture []byte
	var digest string
	if err := harness.DB.QueryRow(`SELECT p.request_json->'source_capture',s.source_capture_sha256 FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id::text=$1`, admission.SnapshotID).Scan(&capture, &digest); err != nil {
		t.Fatal(err)
	}
	model, err := reporting.MaterializeSnapshotCapture(context.Background(), capture, digest)
	if err != nil {
		t.Fatal(err)
	}
	profile, _, err := reporting.ResolveRedactionProfile(admission.Profile.ProfileID, admission.Profile.Version, admission.Recipients, model.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	return profile
}
