package incidentbundles_test

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestExportJobIdempotencyAndDescriptor_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-export")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
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
	divergent := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":         incidentID,
		"client_txn_id":       "txn-export-bundle",
		"reference_pack_mode": "embedded",
	})
	httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	unsupportedRequired := postExport(t, harness.Server, admin, map[string]any{
		"incident_id":           incidentID,
		"client_txn_id":         "txn-export-required-capability",
		"required_capabilities": []string{"snapshots"},
	})
	body := httptestx.RequireErrorEnvelope(t, unsupportedRequired, http.StatusBadRequest, "invalid_incident_bundle_request")
	if details := httptestx.RequireErrorDetails(t, body); details["reason_code"] != "invalid_required_capabilities" {
		t.Fatalf("required capability rejection reason mismatch: %#v", details)
	}

	terminal := waitJob(t, harness.Server, admin, job["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentbundles.ResultIncidentBundleExported {
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

	dequeueGate := &incidentBundleTestDequeueGate{}
	harness.Server.Runtime.JobRunner.ConfigureDequeueGate(dequeueGate)
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

	recoverIncidentBundleJobsThroughGate(t, harness.Server, dequeueGate)
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
	if summary["code"] != incidentbundles.ResultIncidentBundleImported {
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
	if got := countRows(
		t,
		targetHarness.DB,
		`SELECT count(*) FROM collaboration_event_intents WHERE incident_id = $1 AND event_family = 'record_changed'`,
		incidentID,
	); got != 0 {
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
	rc, _, err := targetHarness.Server.Runtime.ObjectStore.ReadObject(context.Background(), importedStorageKey, objectstore.ReadOptions{})
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
	if actions, _ := nonreversibleFirst["available_rollback_actions"].([]any); len(actions) != 0 || nonreversibleFirst["reversible"] != false {
		t.Fatalf("record-tag create history item should remain ordinarily non-reversible: %#v", nonreversibleFirst)
	}
	nonreversibleRollback := postRollback(t, targetHarness.Server, targetAdmin, seededState.HistoryHostRecordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-imported-nonreversible-rollback",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": nonreversibleRef},
	})
	nonreversibleError := httptestx.RequireErrorEnvelope(t, nonreversibleRollback, http.StatusConflict, "rollback_precondition_failed")
	if details := nonreversibleError["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "target_not_reversible" {
		t.Fatalf("nonreversible rollback reason mismatch: %#v", details)
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
	verified, err := incidentbundles.VerifyBundle(incidentbundles.VerificationInput{Bundle: v1Bundle})
	if err != nil {
		t.Fatalf("verify converted v1 bundle: %v", err)
	}
	if verified.Manifest.BundleVersion != incidentbundles.LegacyBundleVersion {
		t.Fatalf("converted bundle version = %d; want %d", verified.Manifest.BundleVersion, incidentbundles.LegacyBundleVersion)
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

func TestImportFinalPublicationRechecksSubmitterAvailability_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-finalize-source")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-finalize-source",
		"incident_key":  "BUNDLE-FINALIZE",
		"title":         "Incident bundle finalization",
	})
	incidentID := incident["incident_id"].(string)
	timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-finalize-row",
		"timeline.activity_synopsis_text": "Portable finalization event",
	})
	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-finalize-fixture")

	cases := []struct {
		name   string
		mutate func(testing.TB, *sql.DB, string)
	}{
		{
			name: "submitter demoted",
			mutate: func(t testing.TB, db *sql.DB, userID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE users SET is_deployment_admin = false WHERE id = $1`, userID); err != nil {
					t.Fatalf("demote import submitter: %v", err)
				}
			},
		},
		{
			name: "submitter inactive",
			mutate: func(t testing.TB, db *sql.DB, userID string) {
				t.Helper()
				if _, err := db.Exec(`UPDATE users SET is_active = false WHERE id = $1`, userID); err != nil {
					t.Fatalf("deactivate import submitter: %v", err)
				}
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-finalize-"+strings.ReplaceAll(tc.name, " ", "-"))
			targetAdmin, targetAdminID := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
			sequenceBefore := snapshotRecordRevisionSequence(t, targetHarness.DB)
			observerPassword := "ExtensionProfileImportObserverPass!"
			observerUser := flowtest.SeedLocalUserRecord(t, targetHarness.DB, "extension_profile-import-observer-"+strings.ReplaceAll(tc.name, " ", "-")+"@example.test", "Enterprise integration Import Observer", observerPassword, false, true, true)
			observerCookies, observerCSRF := flowtest.LoginLocalUser(t, targetHarness.Server.HTTP.URL, observerUser.Email, observerPassword, nil)
			observerLogin := flowtest.LoginResult{SessionCookie: observerCookies, CSRFCookie: observerCSRF}

			dequeueGate := &incidentBundleTestDequeueGate{}
			targetHarness.Server.Runtime.JobRunner.ConfigureDequeueGate(dequeueGate)
			resp := postImport(t, targetHarness.Server, targetAdmin, `{"client_txn_id":"txn-import-finalize-`+strings.ReplaceAll(tc.name, " ", "-")+`"}`, bundleBytes, "bundle.zip")
			job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)

			tc.mutate(t, targetHarness.DB, targetAdminID)
			recoverIncidentBundleJobsThroughGate(t, targetHarness.Server, dequeueGate)

			terminal := waitFailedJob(t, targetHarness.Server, observerLogin, job["job_id"].(string))
			requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", "initial_admin_unavailable")
			if countRows(t, targetHarness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID) != 0 {
				t.Fatalf("initial-admin-unavailable import must not make incident visible")
			}
			if sideEffects := snapshotImportFinalizationSideEffects(t, targetHarness.DB, incidentID, targetAdminID); sideEffects != (importFinalizationSideEffects{}) {
				t.Fatalf("initial-admin-unavailable import left finalization side effects: %#v", sideEffects)
			}
			if sequenceAfter := snapshotRecordRevisionSequence(t, targetHarness.DB); sequenceAfter != sequenceBefore {
				t.Fatalf("failed import changed record revision sequence: before=%#v after=%#v", sequenceBefore, sequenceAfter)
			}
			assertNoIncidentBundleStaging(t, targetHarness.Server)
		})
	}
}

type recordRevisionSequenceState struct {
	LastValue int64
	IsCalled  bool
}

func snapshotRecordRevisionSequence(t testing.TB, db *sql.DB) recordRevisionSequenceState {
	t.Helper()
	var state recordRevisionSequenceState
	if err := db.QueryRow(`
SELECT last_value, is_called
  FROM public.record_revisions_revision_id_seq
`).Scan(&state.LastValue, &state.IsCalled); err != nil {
		t.Fatalf("snapshot record revision sequence: %v", err)
	}
	return state
}

type incidentBundleTestDequeueGate struct {
	open atomic.Bool
}

func (gate *incidentBundleTestDequeueGate) AdmissionOpen() bool {
	return gate != nil && gate.open.Load()
}

func recoverIncidentBundleJobsThroughGate(t testing.TB, server *httptestx.Server, gate *incidentBundleTestDequeueGate) {
	t.Helper()
	if err := server.Runtime.JobRunner.RecoverHandler(context.Background(), incidentbundles.BundleWorkerKind); err != nil {
		t.Fatalf("queue Incident Bundle recovery behind dequeue gate: %v", err)
	}
	gate.open.Store(true)
	if err := server.Runtime.JobRunner.Activate(context.Background()); err != nil {
		t.Fatalf("activate Incident Bundle recovery through dequeue gate: %v", err)
	}
}

func TestSupersededTimelineReplacementSurvivesImport_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-supersede-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-supersede-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-supersede-source",
		"incident_key":  "BUNDLE-SUPERSEDE",
		"title":         "Incident bundle supersede",
	})
	incidentID := incident["incident_id"].(string)
	replacement := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-supersede-replacement",
		"timeline.activity_synopsis_text": "Replacement event",
	})
	replacementID := replacement["row"].(map[string]any)["record_id"].(string)
	superseded := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-superseded-event",
		"timeline.activity_synopsis_text": "Superseded event",
	})
	supersededID := superseded["row"].(map[string]any)["record_id"].(string)

	supersedeResp := httptestx.DoJSON(t, http.MethodPost, sourceHarness.Server.HTTP.URL+"/api/v1/records/"+supersededID+"/supersede", map[string]any{
		"base_row_version":      1,
		"client_txn_id":         "txn-incident-bundle-supersede-action",
		"reason":                "Replacement preserves portability lineage",
		"replacement_record_id": replacementID,
	}, httptestx.WithCookies(sourceAdmin.SessionCookie, sourceAdmin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRFCookie.Value))
	supersedeData := httptestx.RequireSuccessEnvelope(t, supersedeResp, http.StatusOK)["data"].(map[string]any)
	if supersedeData["replacement_record_id"] != replacementID {
		t.Fatalf("source supersede response missing replacement id: %#v", supersedeData)
	}
	if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(sourceHarness.DB), incidentID, replacementID, supersededID); got != 1 {
		t.Fatalf("source supersedes link count: got %d want 1", got)
	}
	sourceRow := requireTimelineQueryRow(t, sourceHarness.Server, sourceAdmin, incidentID, supersededID)
	if got := timelineCellValue(t, sourceRow, "timeline.replacement_record_id"); got != replacementID {
		t.Fatalf("source query did not surface replacement id: got %#v want %s", got, replacementID)
	}

	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-superseded-timeline")
	importBundleAndWait(t, targetHarness.Server, targetAdmin, bundleBytes, "txn-import-superseded-timeline")

	if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(targetHarness.DB), incidentID, replacementID, supersededID); got != 1 {
		t.Fatalf("target supersedes link count: got %d want 1", got)
	}
	targetRow := requireTimelineQueryRow(t, targetHarness.Server, targetAdmin, incidentID, supersededID)
	if got := timelineCellValue(t, targetRow, "timeline.replacement_record_id"); got != replacementID {
		t.Fatalf("target query did not surface imported replacement id: got %#v want %s", got, replacementID)
	}
}

func TestFailureFamiliesLeaveNoVisibleIncident_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-failure-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-failure-target")
	sourceAdmin, sourceAdminID := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-failure-source",
		"incident_key":  "BUNDLE-FAILURES",
		"title":         "Incident bundle failures",
	})
	incidentID := incident["incident_id"].(string)
	row := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-failure-row",
		"timeline.activity_synopsis_text": "Portable failure event",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	seedIncidentBundlePortableState(t, sourceHarness, incidentID, recordID, sourceAdminID)
	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-failure-fixture")
	blobPath := firstZipMemberWithPrefix(t, bundleBytes, "blobs/sha256/")

	t.Run("export missing blob", func(t *testing.T) {
		brokenHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-export-missing-blob")
		brokenAdmin, brokenAdminID := flowtest.ProvisionBootstrapAdmin(t, brokenHarness.Server.HTTP.URL)
		brokenIncident := scenariotest.CreateIncident(t, brokenHarness.Server, brokenAdmin, map[string]any{
			"client_txn_id": "txn-incident-bundle-export-missing-blob",
			"incident_key":  "BUNDLE-MISSING-BLOB",
			"title":         "Incident bundle missing blob",
		})
		brokenIncidentID := brokenIncident["incident_id"].(string)
		seedMissingIncidentBundleBlob(t, brokenHarness, brokenIncidentID, brokenAdminID)
		job := httptestx.RequireSuccessEnvelope(t, postExport(t, brokenHarness.Server, brokenAdmin, map[string]any{
			"incident_id":   brokenIncidentID,
			"client_txn_id": "txn-export-missing-blob",
		}), http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, brokenHarness.Server, brokenAdmin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "missing_required_blob")
		if countRows(t, brokenHarness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, job["job_id"].(string)) != 0 {
			t.Fatalf("failed missing-blob export must not publish a bundle descriptor")
		}
	})

	importCases := []struct {
		name       string
		txn        string
		bundle     []byte
		wantReason string
	}{
		{
			name:       "checksum mismatch",
			txn:        "txn-import-checksum-mismatch",
			bundle:     corruptZipMember(t, bundleBytes, "data/records.ndjson"),
			wantReason: "checksum_mismatch",
		},
		{
			name:       "missing blob",
			txn:        "txn-import-missing-blob",
			bundle:     removeZipMember(t, bundleBytes, blobPath),
			wantReason: "missing_required_blob",
		},
		{
			name:       "blob hash mismatch",
			txn:        "txn-import-blob-hash-mismatch",
			bundle:     replaceZipMemberAndChecksum(t, bundleBytes, blobPath, []byte("extension_profile wrong blob bytes\n")),
			wantReason: "blob_hash_mismatch",
		},
		{
			name: "malformed manifest",
			txn:  "txn-import-malformed-manifest",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["bundle_version"] = nil
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "malformed_manifest",
		},
		{
			name: "unsupported bundle version",
			txn:  "txn-import-unsupported-bundle-version",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["bundle_version"] = float64(3)
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "unsupported_bundle_version",
		},
		{
			name: "unsupported required capability",
			txn:  "txn-import-unsupported-required-capability",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["required_capabilities"] = []any{"snapshots"}
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "unsupported_required_capability",
		},
		{
			name:       "signature mismatch",
			txn:        "txn-import-signature-mismatch",
			bundle:     appendZipMember(t, bundleBytes, "integrity/signature.ed25519", []byte("not-a-supported-signature")),
			wantReason: "signature_mismatch",
		},
		{
			name: "unknown optional section",
			txn:  "txn-import-unknown-optional-section",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["optional_sections"] = []any{"unknown_section"}
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "malformed_manifest",
		},
	}
	for _, tc := range importCases {
		t.Run(tc.name, func(t *testing.T) {
			assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, tc.txn, tc.bundle, tc.wantReason)
		})
	}

	t.Run("records invariant failures are closed and atomic", func(t *testing.T) {
		recordRows := decodeNDJSONRows(t, zipMemberMap(t, bundleBytes)["data/records.ndjson"])
		if len(recordRows) == 0 {
			t.Fatal("records invariant fixture requires at least one envelope")
		}
		incidentMismatchRows := append([]map[string]any(nil), recordRows...)
		incidentMismatchRows[0] = mapsClone(incidentMismatchRows[0])
		incidentMismatchRows[0]["incident_id"] = uuid.NewString()
		envelopeRows := append([]map[string]any(nil), recordRows...)
		envelopeRows[0] = mapsClone(envelopeRows[0])
		envelopeRows[0]["hostile_member_name"] = "SELECT secret FROM records"
		cases := []struct {
			name        string
			txn         string
			bundle      []byte
			invariantID string
			hostile     string
		}{
			{
				name: "incident scope",
				txn:  "txn-import-records-incident-scope",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/records.ndjson",
					encodeNDJSONRows(t, incidentMismatchRows),
				),
				invariantID: "records.incident_scope",
			},
			{
				name: "envelope legal",
				txn:  "txn-import-records-envelope-legal",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/records.ndjson",
					encodeNDJSONRows(t, envelopeRows),
				),
				invariantID: "records.envelope_legal",
				hostile:     "SELECT secret FROM records",
			},
			{
				name: "subtype complete",
				txn:  "txn-import-records-subtype-complete",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/timeline_records.ndjson",
					nil,
				),
				invariantID: "records.subtype_complete",
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				terminal := assertImportFailureLeavesState(
					t,
					targetHarness,
					targetAdmin,
					incidentID,
					testCase.txn,
					testCase.bundle,
					"source_family_invalid",
				)
				errorSummary := terminal["error_summary"].(map[string]any)
				details := errorSummary["details"].(map[string]any)
				if len(details) != 3 ||
					details["reason_code"] != "source_family_invalid" ||
					details["source_family_id"] != "records" ||
					details["invariant_id"] != testCase.invariantID {
					t.Fatalf("records failure details are not closed: %#v", details)
				}
				encoded, err := json.Marshal(terminal)
				if err != nil {
					t.Fatalf("encode records failure result: %v", err)
				}
				if testCase.hostile != "" &&
					strings.Contains(string(encoded), testCase.hostile) {
					t.Fatalf("records failure exposed hostile source content: %s", encoded)
				}
			})
		}
	})

	t.Run("safe directory entries import", func(t *testing.T) {
		directoryHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-directory-import")
		directoryAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, directoryHarness.Server.HTTP.URL)
		terminal := importBundleAndWait(t, directoryHarness.Server, directoryAdmin, appendZipDirectoryMembers(t, bundleBytes, "data/", "integrity/", "ext/"), "txn-import-safe-directories")
		if terminal["status"] != "succeeded" {
			encoded, _ := json.MarshalIndent(terminal, "", "  ")
			t.Fatalf("directory-bearing import did not succeed: %s", encoded)
		}
		summary := terminal["result_summary"].(map[string]any)
		if summary["code"] != "incident_bundle_imported" {
			t.Fatalf("directory-bearing import summary mismatch: %#v", summary)
		}
		refs := summary["resource_refs"].([]any)
		if len(refs) != 1 || refs[0].(map[string]any)["id"] != incidentID {
			t.Fatalf("directory-bearing import refs mismatch: %#v", refs)
		}
		assertNoIncidentBundleStaging(t, directoryHarness.Server)
	})

	t.Run("archive extracted byte limit", func(t *testing.T) {
		limitHarness := startIsolatedIncidentBundleServerWithEnv(t, runtime, "extension_profile-incident-bundle-extracted-limit", map[string]string{
			"CARTULARY__LIMITS__INCIDENT_BUNDLES__MAX_EXTRACTED_BYTES": "1",
		})
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-extracted-limit", bundleBytes, "archive_extracted_bytes_exceeded")
	})
	t.Run("archive member count limit", func(t *testing.T) {
		limitHarness := startIsolatedIncidentBundleServerWithEnv(t, runtime, "extension_profile-incident-bundle-member-limit", map[string]string{
			"CARTULARY__LIMITS__ARCHIVES__MAX_MEMBERS": "20",
		})
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-member-limit", bundleBytes, "archive_member_count_exceeded")
	})

	importBundleAndWait(t, targetHarness.Server, targetAdmin, bundleBytes, "txn-import-duplicate-baseline")
	assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, "txn-import-duplicate-incident", bundleBytes, "duplicate_incident_id")
}

func TestNetworkFlowRetainedStateBlocksBundleExport_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-network-flow-block")
	admin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-network-flow-block",
		"incident_key":  "BUNDLE-NF-BLOCK",
		"title":         "Incident bundle Network Flow block",
	})
	incidentID := incident["incident_id"].(string)
	tableID := seedIncidentBundleNetworkFlowTable(t, harness.DB, incidentID, adminID)

	assertBlocked := func(clientTxnID string) {
		job := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, admin, map[string]any{
			"incident_id": incidentID, "client_txn_id": clientTxnID,
		}), http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, harness.Server, admin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "missing_required_file")
		if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, job["job_id"].(string)) != 0 {
			t.Fatal("blocked Network Flow export published a bundle descriptor")
		}
	}
	assertBlocked("txn-export-network-flow-active")
	if _, err := harness.DB.Exec(`
UPDATE network_flow_tables
   SET table_status = 'soft_deleted',
       deleted_at = now(),
       updated_at = now()
 WHERE network_flow_table_id = $1
`, tableID); err != nil {
		t.Fatalf("soft delete Network Flow table: %v", err)
	}
	assertBlocked("txn-export-network-flow-soft-deleted")
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

func seedIncidentBundleNetworkFlowTable(t testing.TB, db *sql.DB, incidentID, actorID string) string {
	t.Helper()
	sessionID := uuid.New()
	unitID := uuid.New()
	tableID := "nft_" + strings.Repeat("a", 32)
	digest := strings.Repeat("1", 64)
	if _, err := db.Exec(`
INSERT INTO import_sessions (
    import_session_id, incident_id, created_by_user_id, client_txn_id, assistant_profile,
    source_file_kind, original_filename, source_content_sha256, source_media_type, source_byte_size,
    parser_profile_id, parser_version, session_status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'network_flow_test', 'csv', 'flows.csv', $5, 'text/csv', 12,
    'network_flow.rfc4180_headered_csv.v1', 'test', 'ready_to_apply', now(), now()
)
`, sessionID, incidentID, actorID, "txn-import-"+unitID.String(), digest); err != nil {
		t.Fatalf("seed Network Flow import session: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO import_units (
    import_unit_id, import_session_id, unit_status, locator_kind, locator, source_rect_a1,
    header_row_ref, data_start_row_ref, inferred_row_count, inferred_column_count,
    warning_codes, mapping_fingerprint, approved_mapping_json, columns_json, source_rows_json,
    preview_rows_json, approved_target_kind, approved_extension_profile_id, discovery_sequence,
    created_at, updated_at
) VALUES (
    $1, $2, 'ready', 'csv', 'unit-1', 'A1:Z2', 1, 2, 1, 9,
    '{}', $3, '{}'::jsonb, '[]'::jsonb, '[]'::jsonb, '[]'::jsonb,
    'network_flow_table', 'network_flow_activity', 1, now(), now()
)
`, unitID, sessionID, digest); err != nil {
		t.Fatalf("seed Network Flow import unit: %v", err)
	}
	if _, err := db.Exec(`
INSERT INTO network_flow_tables (
    network_flow_table_id, incident_id, display_name, table_status,
    source_import_session_id, source_import_unit_id, source_content_sha256,
    source_filename_display, source_filename_digest, source_filename_digest_key_id,
    mapping_fingerprint, source_profile_id, parser_profile_id,
    row_count_accepted, row_count_rejected, created_by_user_id
) VALUES (
    $1, $2, 'Retained flows', 'active', $3, $4, $5,
    'flows.csv', $5, 'nf-test-key', $5,
    'network_flow.cisco_sna_netflow_csv.v1', 'network_flow.rfc4180_headered_csv.v1',
    1, 0, $6
)
`, tableID, incidentID, sessionID, unitID, digest, actorID); err != nil {
		t.Fatalf("seed Network Flow table: %v", err)
	}
	return tableID
}

func TestImportEnvelopeFailuresCreateNoDurableState_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartDefaultServer(t, "extension_profile-incident-bundle-envelope-failures")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	validFile := []byte("not a bundle but parser-valid bytes")
	cases := []struct {
		name           string
		build          func(testing.TB, *httptestx.Server, flowtest.LoginResult) *http.Request
		wantReason     string
		wantPart       string
		wantContentErr bool
	}{
		{
			name: "missing boundary",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", strings.NewReader("not multipart"))
				if err != nil {
					t.Fatalf("create missing-boundary request: %v", err)
				}
				req.Header.Set("Content-Type", "multipart/form-data")
				addImportAuth(req, login)
				return req
			},
			wantReason: "unsupported_upload_envelope",
		},
		{
			name: "duplicate metadata part",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-duplicate-metadata"}`),
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-duplicate-metadata"}`),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "duplicate_part",
			wantPart:   "metadata",
		},
		{
			name: "unexpected part",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-unexpected-part"}`),
					fileUploadPart("extra", "extra.txt", "text/plain", []byte("extra")),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "unexpected_part",
		},
		{
			name: "malformed metadata json",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":`),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "malformed_metadata_json",
			wantPart:   "metadata",
		},
		{
			name: "duplicate metadata json key",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-a","client_txn_id":"txn-b"}`),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "malformed_metadata_json",
			wantPart:   "metadata",
		},
		{
			name: "non-object metadata",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `[]`),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "request_not_object",
			wantPart:   "metadata",
		},
		{
			name: "invalid file content type",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-file-content-type"}`),
					fileUploadPart("file", "bundle.txt", "text/plain", validFile),
				})
			},
			wantReason:     "invalid_part_content_type",
			wantPart:       "file",
			wantContentErr: true,
		},
		{
			name: "forbidden import mode field",
			build: func(t testing.TB, server *httptestx.Server, login flowtest.LoginResult) *http.Request {
				return newImportEnvelopeRequest(t, server, login, []uploadPart{
					jsonUploadPart("metadata", "", `{"client_txn_id":"txn-envelope-forbidden-mode","clone_mode":"copy"}`),
					fileUploadPart("file", "bundle.zip", incidentbundles.MediaTypeZip, validFile),
				})
			},
			wantReason: "unknown_field",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := snapshotEnvelopeDurability(t, harness.DB)
			resp := httptestx.Do(t, http.DefaultClient, tc.build(t, harness.Server, admin))
			body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_incident_bundle_request")
			details := httptestx.RequireErrorDetails(t, body)
			if details["reason_code"] != tc.wantReason {
				t.Fatalf("reason mismatch: got %#v want %s", details, tc.wantReason)
			}
			if tc.wantPart != "" && details["part_name"] != tc.wantPart {
				t.Fatalf("part_name mismatch: got %#v want %s", details, tc.wantPart)
			}
			if tc.wantContentErr {
				if details["received_content_type"] != "text/plain" {
					t.Fatalf("received content type missing: %#v", details)
				}
				if len(stringArray(t, details["allowed_content_types"])) == 0 {
					t.Fatalf("allowed content types missing: %#v", details)
				}
			}
			after := snapshotEnvelopeDurability(t, harness.DB)
			if after != before {
				t.Fatalf("early envelope failure created durable rows: before=%#v after=%#v", before, after)
			}
			assertNoIncidentBundleStaging(t, harness.Server)
		})
	}
}

type seededIncidentBundlePortableState struct {
	BlobBytes                []byte
	BlobSHA                  string
	ObjectBlobID             string
	EvidenceRecordID         string
	HistoryHostRecordID      string
	PartyRecordID            string
	FindingArtifactRecordID  string
	QueryArtifactRecordID    string
	KeywordArtifactRecordID  string
	HandoffArtifactRecordID  string
	SavedViewID              string
	ReversibleChangeSetID    string
	NonReversibleChangeSetID string
}

func seedIncidentBundlePortableState(t testing.TB, harness *appsupport.ServerHarness, incidentID string, timelineRecordID string, actorUserID string) seededIncidentBundlePortableState {
	t.Helper()
	ctx := context.Background()
	incidentUUID := uuid.MustParse(incidentID)
	timelineUUID := uuid.MustParse(timelineRecordID)
	actorUUID := uuid.MustParse(actorUserID)
	if _, err := harness.DB.Exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'ExtensionProfile Portability', 'extension_profile-portability', $3)
`, incidentID, timelineRecordID, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}

	historyHostID := uuid.New()
	insertHostRecord(t, harness.DB, incidentUUID, historyHostID, actorUUID, "portable host before", "portable-host")
	if _, err := harness.DB.Exec(`
UPDATE records
   SET created_at = '2026-05-25T16:59:00Z',
       updated_at = '2026-05-25T16:59:00Z'
 WHERE record_id = $1
`, historyHostID); err != nil {
		t.Fatalf("normalize portable host envelope time: %v", err)
	}

	identityID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'identity', $3, $3)
`, identityID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed identity envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO identities (
    record_id, incident_id, display_name, upn, email, sam_account_name,
    entity_origin, identity_state, created_by_user_id, updated_by_user_id
)
VALUES (
    $1, $2, 'Portable Identity', 'portable.identity@example.test',
    'portable.identity@example.test', 'portable.identity', 'entity_import',
    'canonical', $3, $3
)
`, identityID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed identity row: %v", err)
	}
	assessmentID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'assessment', $3, $3)
`, assessmentID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed assessment envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO assessments (
    record_id, incident_id, subject_record_id, subject_type,
    assessment_state, confidence_score, assessor_user_id, rationale
)
VALUES ($1, $2, $3, 'host', 'suspected', 70, $4, 'Portable assessment')
`, assessmentID, incidentUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed assessment row: %v", err)
	}

	indicatorID := uuid.New()
	indicatorTimestamp := time.Now().UTC().Truncate(time.Microsecond)
	if _, err := harness.DB.Exec(`
INSERT INTO records (
    record_id, incident_id, record_type, created_at, created_by_user_id,
    updated_at, updated_by_user_id
)
VALUES ($1, $2, 'indicator', $3, $4, $3, $4)
`, indicatorID, incidentUUID, indicatorTimestamp, actorUUID); err != nil {
		t.Fatalf("seed indicator envelope: %v", err)
	}
	indicatortest.SeedSubtype(t, harness.DB, incidentUUID, indicatorID, "domain_name", "atomic", "portable.example.test")
	if _, err := harness.DB.Exec(`
INSERT INTO indicator_observations (
    incident_id, source_record_id, source_field_key, origin_kind, origin_locator,
    observed_text, parsed_indicator_type, normalized_candidate, resolution_status,
    resolved_indicator_record_id, row_version, created_by_user_id, resolved_by_user_id,
    resolved_at, resolution_method, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, 'timeline.activity_synopsis_text', 'extraction', 'extension_profile', 'portable.example.test', 'domain_name', 'portable.example.test', 'resolved', $3, 2, $4, $4, now(), 'fixture', now(), $4)
`, incidentUUID, timelineUUID, indicatorID, actorUUID); err != nil {
		t.Fatalf("seed indicator observation: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO indicator_state_intervals (
    incident_id, indicator_record_id, lifecycle_state, valid_from, support_refs,
    row_version, created_by_user_id, deleted_at, deleted_by_user_id
)
VALUES ($1, $2, 'active', now(), '[]'::jsonb, 2, $3, now(), $3)
`, incidentUUID, indicatorID, actorUUID); err != nil {
		t.Fatalf("seed indicator lifecycle tombstone: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal,
    created_by_user_id, resolved_record_id, resolved_by_user_id, resolved_at, resolution_method
)
VALUES ($1, 'host', 'timeline.activity_synopsis_text', 'manual', 'extension_profile', 'portable host', 'portable host', 'resolved', 1, 1, $2, $3, $2, now(), 'fixture')
`, timelineUUID, actorUUID, historyHostID); err != nil {
		t.Fatalf("seed entity mention: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key, provenance,
    owner_user_id, created_by_user_id, decided_at
)
VALUES ($1, $2, $3, 'observed_on_host', 'timeline.host_refs', 'manual', $4, $4, now())
`, incidentUUID, timelineUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed record link: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO timeline_time_conversion_profiles (
    incident_id, enabled, local_offset_minutes, local_label, updated_by_user_id
)
VALUES ($1, true, -300, 'America/New_York', $2)
`, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed timeline time conversion profile: %v", err)
	}
	partyRecordID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'party', $3, $3)
`, partyRecordID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed party envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name,
    role_title, primary_email, timezone_name, external_ref, notes
)
VALUES ($1, $2, 'Portable Party', 'person', 'Cartulary IR', 'Incident lead', 'portable-party@example.test', 'America/New_York', 'party-1', 'portable party note')
`, partyRecordID, incidentUUID); err != nil {
		t.Fatalf("seed party row: %v", err)
	}
	taskRequestID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'task_request', $3, $3)
`, taskRequestID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed task-request envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority,
    task_kind, workstream, requester_party_id
)
VALUES (
    $1, $2, 'Portable task request', 'open', $3, 'high',
    'request', 'portable', $4
)
`, taskRequestID, incidentUUID, actorUUID, partyRecordID); err != nil {
		t.Fatalf("seed task-request row: %v", err)
	}
	decisionID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'decision', $3, $3)
`, decisionID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed decision envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id,
    decision_type, decided_at, rationale
)
VALUES (
    $1, $2, 'Portable decision', 'approved', $3,
    'containment', '2026-05-25T17:02:00Z', 'Portable rationale'
)
`, decisionID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed decision row: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO entity_preserved_identifiers (
    incident_id, record_id, entity_type, identifier_type, raw_value,
    normalized_value, classification, created_by_user_id
)
VALUES ($1, $2, 'host', 'hostname', 'Portable-Host', 'portable-host', 'exact_match_reuse', $3)
`, incidentUUID, historyHostID, actorUUID); err != nil {
		t.Fatalf("seed entity preserved identifier: %v", err)
	}
	findingArtifactID := uuid.New()
	queryArtifactID := uuid.New()
	keywordArtifactID := uuid.New()
	handoffArtifactID := uuid.New()
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, findingArtifactID, actorUUID, "finding", "Portable finding")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_findings (
    record_id, incident_id, kind, statement, state, confidence_score, owner_user_id
)
VALUES ($1, $2, 'finding', 'Portable finding statement', 'open', 87, $3)
`, findingArtifactID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed artifact finding: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, queryArtifactID, actorUUID, "investigative_query", "Portable investigative query")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_investigative_queries (
    record_id, incident_id, query_id, platform, purpose, query_text, created_by_user_id
)
VALUES ($1, $2, 'portable-query-1', 'kusto', 'Find portable events', 'SecurityEvent | take 10', $3)
`, queryArtifactID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed artifact investigative query: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, keywordArtifactID, actorUUID, "forensic_keyword", "Portable forensic keyword")
	if _, err := harness.DB.Exec(`
INSERT INTO artifact_forensic_keywords (
    record_id, incident_id, keyword_id, pattern, reason, match_mode, case_sensitive
)
VALUES ($1, $2, 'portable-keyword-1', 'PortableKeyword', 'Portable keyword reason', 'literal', true)
`, keywordArtifactID, incidentUUID); err != nil {
		t.Fatalf("seed artifact forensic keyword: %v", err)
	}
	seedPortableArtifactRecord(t, harness.DB, incidentUUID, handoffArtifactID, actorUUID, "handoff", "Portable handoff")
	if _, err := harness.DB.Exec(`
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text, created_by_user_id
)
VALUES ($1, $2, 'Portable Risk', 'portable risk', $3)
`, incidentUUID, handoffArtifactID, actorUUID); err != nil {
		t.Fatalf("seed handoff risk ref: %v", err)
	}
	savedViewID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name,
    query_json, layout_json, owner_user_id
)
VALUES (
    $1,
    $2,
    $4,
    'private',
    'Portable saved view',
    '{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["portable"]}}]}'::jsonb,
    '{}'::jsonb,
    $3
)
`, savedViewID, incidentUUID, actorUUID, timeline.TimelineViewSchemaID); err != nil {
		t.Fatalf("seed saved view: %v", err)
	}

	reversibleChangeSetID := uuid.New()
	seedPortableRollbackHostPatch(t, harness.DB, incidentUUID, historyHostID, actorUUID, reversibleChangeSetID, time.Date(2026, 5, 25, 17, 0, 0, 0, time.UTC), "portable host before", "portable host after")
	nonReversibleChangeSetID := uuid.New()
	seedPortableRecordTagCreateHistory(t, harness.DB, incidentUUID, historyHostID, actorUUID, nonReversibleChangeSetID, time.Date(2026, 5, 25, 17, 1, 0, 0, time.UTC))

	blobBytes := []byte("extension_profile incident bundle blob\n")
	sum := sha256.Sum256(blobBytes)
	blobSHA := hex.EncodeToString(sum[:])
	sourceStorageKey := "extension_profile/source/" + incidentID + "/" + blobSHA
	if err := harness.Server.Runtime.ObjectStore.PutObject(ctx, sourceStorageKey, bytes.NewReader(blobBytes), int64(len(blobBytes)), "text/plain"); err != nil {
		t.Fatalf("seed source object bytes: %v", err)
	}

	var objectBlobID string
	if err := harness.DB.QueryRow(`
INSERT INTO object_blobs (
    incident_id,
    created_by_user_id,
    storage_key,
    upload_state,
    byte_size,
    filename_hint,
    content_type_hint,
    expected_sha256_hex,
    observed_size,
    observed_content_type,
    observed_sha256_hex,
    target_expires_at,
    pending_expires_at,
    finalized_at
)
VALUES ($1, $2, $3, 'available', $4, 'extension_profile.txt', 'text/plain', $5, $4, 'text/plain', $5, now() + interval '1 hour', now() + interval '1 hour', now())
RETURNING object_blob_id
`, incidentID, actorUserID, sourceStorageKey, len(blobBytes), blobSHA).Scan(&objectBlobID); err != nil {
		t.Fatalf("seed object blob row: %v", err)
	}

	var evidenceRecordID string
	if err := harness.DB.QueryRow(`
INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, 'evidence', $2, $2)
RETURNING record_id
`, incidentID, actorUserID).Scan(&evidenceRecordID); err != nil {
		t.Fatalf("seed evidence record envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO evidence (
    record_id,
    incident_id,
    title,
    lifecycle_state,
    requested_at,
    received_at,
    storage_ref,
    blob_hash,
    upload_state,
    object_blob_id
)
VALUES ($1, $2, 'Portable evidence', 'received', now(), now(), $3, $4, 'available', $5)
`, evidenceRecordID, incidentID, sourceStorageKey, blobSHA, objectBlobID); err != nil {
		t.Fatalf("seed evidence row: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO evidence_custody_events (
    incident_id,
    evidence_record_id,
    custody_event_type,
    actor_user_id,
    location_text,
    note
)
VALUES ($1, $2, 'made_available', $3, 'source deployment', 'seeded portable custody event')
`, incidentID, evidenceRecordID, actorUserID); err != nil {
		t.Fatalf("seed evidence custody event: %v", err)
	}
	return seededIncidentBundlePortableState{
		BlobBytes:                blobBytes,
		BlobSHA:                  blobSHA,
		ObjectBlobID:             objectBlobID,
		EvidenceRecordID:         evidenceRecordID,
		HistoryHostRecordID:      historyHostID.String(),
		PartyRecordID:            partyRecordID.String(),
		FindingArtifactRecordID:  findingArtifactID.String(),
		QueryArtifactRecordID:    queryArtifactID.String(),
		KeywordArtifactRecordID:  keywordArtifactID.String(),
		HandoffArtifactRecordID:  handoffArtifactID.String(),
		SavedViewID:              savedViewID.String(),
		ReversibleChangeSetID:    reversibleChangeSetID.String(),
		NonReversibleChangeSetID: nonReversibleChangeSetID.String(),
	}
}

func seedPortableArtifactRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, artifactType string, title string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'artifact', $3, $3)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed artifact envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, title, body, timestamp_utc, created_by_user_id
)
VALUES ($1, $2, $3, $4, 'portable artifact body', $5, $6)
`, recordID, incidentID, artifactType, title, time.Date(2026, 5, 25, 16, 30, 0, 0, time.UTC), actorID); err != nil {
		t.Fatalf("seed artifact row: %v", err)
	}
}

func insertHostRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, displayName string, hostname string) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'host', $3, $3)
`, recordID, incidentID, actorID); err != nil {
		t.Fatalf("seed host record envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (
    record_id, incident_id, display_name, hostname, host_state,
    row_version, created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, $3, $4, 'canonical', 1, $5, $5)
`, recordID, incidentID, displayName, hostname, actorID); err != nil {
		t.Fatalf("seed host row: %v", err)
	}
}

func seedPortableRollbackHostPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time, beforeName string, afterName string) {
	t.Helper()
	beforeRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 1}
	afterRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 2}
	beforeSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": beforeName, "hostname": "portable-host", "host_state": "canonical", "row_version": 1}
	afterSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": afterName, "hostname": "portable-host", "host_state": "canonical", "row_version": 2}
	beforeValue := map[string]any{"record": beforeRecord, "source": beforeSource}
	afterValue := map[string]any{"record": afterRecord, "source": afterSource}
	if _, err := db.ExecContext(context.Background(), `
UPDATE records
   SET row_version = 2,
       updated_at = $4,
       updated_by_user_id = $3
 WHERE record_id = $1
   AND incident_id = $2
`, recordID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("advance portable rollback host envelope: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
UPDATE hosts
   SET display_name = $3,
       row_version = 2,
       updated_at = $4,
       updated_by_user_id = $5
 WHERE record_id = $1
   AND incident_id = $2
`, recordID, incidentID, afterName, createdAt, actorID); err != nil {
		t.Fatalf("advance portable rollback host source: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at)
VALUES ($1, $2, $3, 'workbook.records.patch', 'portable rollback seed', 'txn-portable-host-patch', 'req-portable-host-patch', $4)
`, changeSetID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable rollback change set: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_version_id, after_version_id, before_value, after_value
)
VALUES ($1, 1, 'host', $2, 'field_update', $3, $4, $5, $6)
`, changeSetID, recordID.String(), "host:"+recordID.String()+":1", "host:"+recordID.String()+":2", jsonRaw(t, beforeValue), jsonRaw(t, afterValue)); err != nil {
		t.Fatalf("seed portable rollback mutation: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES ($1, $2, 2, $3, $4, $5)
`, changeSetID, recordID, jsonRaw(t, beforeValue), jsonRaw(t, afterValue), createdAt); err != nil {
		t.Fatalf("seed portable rollback revision: %v", err)
	}
}

func seedPortableRecordTagCreateHistory(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time) {
	t.Helper()
	recordTagID := uuid.New()
	afterValue := map[string]any{
		"record_id": recordID.String(),
		"tag_name":  "ExtensionProfile History",
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'ExtensionProfile History', 'extension_profile-history', $4, $5, $5)
`, recordTagID, incidentID, recordID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable history record tag: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_sets (change_set_id, incident_id, actor_user_id, source, reason, client_txn_id, request_id, created_at)
VALUES ($1, $2, $3, 'records.tags.create', 'portable tag seed', 'txn-portable-tag-create', 'req-portable-tag-create', $4)
`, changeSetID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("seed portable record-tag change set: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO change_set_mutations (change_set_id, sequence_no, target_kind, target_id, operation_kind, before_value, after_value)
VALUES ($1, 1, 'record_tag', $2, 'create', NULL, $3)
`, changeSetID, recordTagID.String(), jsonRaw(t, afterValue)); err != nil {
		t.Fatalf("seed portable record-tag mutation: %v", err)
	}
}

func postExport(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/export", body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func postImport(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, metadata string, file []byte, filename string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadataPart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="metadata"`,
		"Content-Type":        "application/json; charset=utf-8",
	}))
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(metadataPart, metadata); err != nil {
		t.Fatalf("write metadata: %v", err)
	}
	filePart, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
		"Content-Disposition": `form-data; name="file"; filename="` + filename + `"`,
		"Content-Type":        incidentbundles.MediaTypeZip,
	}))
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write(file); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatalf("create import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	return httptestx.Do(t, http.DefaultClient, req)
}

func waitJob(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		switch status {
		case "succeeded":
			return job
		case "failed", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached terminal non-success state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not finish", jobID)
	return nil
}

func waitJobWithStatus(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		if status == wantStatus {
			return job
		}
		switch status {
		case "succeeded", "failed", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached terminal status %q while waiting for %q: %s", status, wantStatus, encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not reach status %q", jobID, wantStatus)
	return nil
}

func waitFailedJob(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
		job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
		status := job["status"].(string)
		switch status {
		case "failed":
			return job
		case "succeeded", "canceled":
			encoded, _ := json.MarshalIndent(job, "", "  ")
			t.Fatalf("job reached unexpected terminal state: %s", encoded)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("job %s did not fail", jobID)
	return nil
}

func corruptZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip for corruption: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	found := false
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		if member.Name == memberPath {
			data = append(data, []byte("corrupt\n")...)
			found = true
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create corrupt member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write corrupt member %s: %v", member.Name, err)
		}
	}
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close corrupt zip: %v", err)
	}
	return buf.Bytes()
}

func zipMemberBytes(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, member := range reader.File {
		if member.Name != memberPath {
			continue
		}
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", memberPath, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", memberPath, err)
		}
		return data
	}
	t.Fatalf("zip member %s not found", memberPath)
	return nil
}

func firstZipMemberWithPrefix(t testing.TB, bundle []byte, prefix string) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	for _, member := range reader.File {
		if strings.HasPrefix(member.Name, prefix) {
			return member.Name
		}
	}
	t.Fatalf("zip member with prefix %s not found", prefix)
	return ""
}

func removeZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	return rewriteZipMembers(t, bundle, func(path string, data []byte) ([]byte, bool) {
		if path == memberPath {
			return nil, false
		}
		return data, true
	})
}

func replaceZipMember(t testing.TB, bundle []byte, memberPath string, replace func([]byte) []byte) []byte {
	t.Helper()
	found := false
	rewritten := rewriteZipMembers(t, bundle, func(path string, data []byte) ([]byte, bool) {
		if path != memberPath {
			return data, true
		}
		found = true
		return replace(data), true
	})
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	return rewritten
}

func appendZipMember(t testing.TB, bundle []byte, memberPath string, payload []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	w, err := writer.Create(memberPath)
	if err != nil {
		t.Fatalf("create appended member %s: %v", memberPath, err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatalf("write appended member %s: %v", memberPath, err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func appendZipDirectoryMembers(t testing.TB, bundle []byte, directories ...string) []byte {
	t.Helper()
	result := bundle
	for _, directory := range directories {
		result = appendZipMember(t, result, directory, nil)
	}
	return result
}

func replaceZipMemberAndChecksum(t testing.TB, bundle []byte, memberPath string, replacement []byte) []byte {
	t.Helper()
	replacementSHA := hashHexBytes(replacement)
	rewritten := replaceZipMember(t, bundle, memberPath, func([]byte) []byte {
		return replacement
	})
	return replaceZipMember(t, rewritten, "integrity/checksums.sha256", func(original []byte) []byte {
		lines := strings.Split(strings.TrimSpace(string(original)), "\n")
		for idx, line := range lines {
			if strings.HasSuffix(line, "  "+memberPath) {
				lines[idx] = replacementSHA + "  " + memberPath
			}
		}
		return []byte(strings.Join(lines, "\n") + "\n")
	})
}

func replaceStructuredBundleMember(
	t testing.TB,
	bundle []byte,
	memberPath string,
	replacement []byte,
) []byte {
	t.Helper()
	files := zipMemberMap(t, bundle)
	files[memberPath] = append([]byte(nil), replacement...)
	var manifest incidentbundles.BundleManifest
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatalf("decode manifest for structured replacement: %v", err)
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		if path == "manifest.json" || strings.HasPrefix(path, "integrity/") {
			continue
		}
		paths = append(paths, path)
	}
	slices.Sort(paths)
	manifest.Files = make([]incidentbundles.ManifestFile, 0, len(paths))
	checksumLines := make([]string, 0, len(paths))
	for _, path := range paths {
		digest := hashHexBytes(files[path])
		manifest.Files = append(manifest.Files, incidentbundles.ManifestFile{
			Path:      path,
			SHA256:    "sha256:" + digest,
			SizeBytes: int64(len(files[path])),
			Required:  !strings.HasPrefix(path, "ext/"),
		})
		checksumLines = append(checksumLines, digest+"  "+path)
	}
	sourceBoundary, err := json.Marshal(manifest.Files)
	if err != nil {
		t.Fatalf("encode structured replacement boundary: %v", err)
	}
	manifest.SourceChangeSetHighWatermark = "cartulary.source_boundary.v1:" +
		hashHexBytes(sourceBoundary)
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode structured replacement manifest: %v", err)
	}
	files["manifest.json"] = append(manifestBytes, '\n')
	files["integrity/checksums.sha256"] = []byte(
		strings.Join(checksumLines, "\n") + "\n",
	)
	return writeZipMemberMap(t, files)
}

func encodeNDJSONRows(t testing.TB, rows []map[string]any) []byte {
	t.Helper()
	var payload bytes.Buffer
	for _, row := range rows {
		encoded, err := json.Marshal(row)
		if err != nil {
			t.Fatalf("encode NDJSON row: %v", err)
		}
		payload.Write(encoded)
		payload.WriteByte('\n')
	}
	return payload.Bytes()
}

func mapsClone(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func convertV2TimelineBundleToV1(t testing.TB, bundle []byte) []byte {
	t.Helper()
	files := zipMemberMap(t, bundle)
	var manifest incidentbundles.BundleManifest
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatalf("decode v2 manifest for v1 conversion: %v", err)
	}
	recordEnvelopes := ndjsonRowsByIdentity(t, files["data/records.ndjson"], "record_id")
	provenanceRows := decodeNDJSONRows(t, files["data/timeline_source_provenance.ndjson"])
	provenanceByRecord := map[string][]map[string]any{}
	for _, provenance := range provenanceRows {
		recordID, _ := provenance["record_id"].(string)
		provenanceByRecord[recordID] = append(provenanceByRecord[recordID], provenance)
	}
	legacyRows := decodeNDJSONRows(t, files["data/timeline_records.ndjson"])
	var legacyPayload bytes.Buffer
	for _, legacyRow := range legacyRows {
		recordID, _ := legacyRow["record_id"].(string)
		envelope := recordEnvelopes[recordID]
		for target, source := range map[string]string{
			"row_version":        "row_version",
			"recorded_at":        "created_at",
			"edited_at":          "updated_at",
			"created_by_user_id": "created_by_user_id",
			"updated_by_user_id": "updated_by_user_id",
		} {
			legacyRow[target] = envelope[source]
		}
		importColumns := make([]map[string]any, 0, len(provenanceByRecord[recordID]))
		for _, provenance := range provenanceByRecord[recordID] {
			column := map[string]any{
				"source_kind":           provenance["source_kind"],
				"source_row_ordinal":    provenance["source_row_ordinal"],
				"source_column_ordinal": provenance["source_column_ordinal"],
				"source_header_text":    provenance["source_header"],
				"raw_value":             provenance["raw_value"],
			}
			if cellKind, ok := provenance["cell_kind"]; ok && cellKind != nil && cellKind != "" {
				column["cell_kind"] = cellKind
			}
			if metadata, ok := provenance["source_metadata"].(map[string]any); ok {
				for key, value := range metadata {
					if key != "source_kind" {
						column[key] = value
					}
				}
			}
			importColumns = append(importColumns, column)
		}
		legacyRow["raw_capture"] = map[string]any{"import_columns": importColumns}
		encoded, err := json.Marshal(legacyRow)
		if err != nil {
			t.Fatalf("encode v1 Timeline row: %v", err)
		}
		legacyPayload.Write(encoded)
		legacyPayload.WriteByte('\n')
	}
	files["data/timeline_time_conversion_profiles.ndjson"] = files["data/timeline_time_profiles.ndjson"]
	files["data/timeline_events.ndjson"] = legacyPayload.Bytes()
	delete(files, "data/timeline_time_profiles.ndjson")
	delete(files, "data/timeline_records.ndjson")
	delete(files, "data/timeline_source_provenance.ndjson")
	delete(files, "manifest.json")
	delete(files, "integrity/checksums.sha256")

	paths := make([]string, 0, len(files))
	for path := range files {
		if !strings.HasPrefix(path, "integrity/") {
			paths = append(paths, path)
		}
	}
	slices.Sort(paths)
	manifest.Files = make([]incidentbundles.ManifestFile, 0, len(paths))
	for _, path := range paths {
		manifest.Files = append(manifest.Files, incidentbundles.ManifestFile{
			Path:      path,
			SHA256:    "sha256:" + hashHexBytes(files[path]),
			SizeBytes: int64(len(files[path])),
			Required:  !strings.HasPrefix(path, "ext/"),
		})
	}
	sourceBoundaryBytes, err := json.Marshal(manifest.Files)
	if err != nil {
		t.Fatalf("encode v1 source boundary: %v", err)
	}
	manifest.BundleVersion = incidentbundles.LegacyBundleVersion
	manifest.SourceChangeSetHighWatermark = "cartulary.source_boundary.v1:" + hashHexBytes(sourceBoundaryBytes)
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode v1 manifest: %v", err)
	}
	files["manifest.json"] = append(manifestBytes, '\n')
	checksumLines := make([]string, 0, len(paths))
	for _, path := range paths {
		checksumLines = append(checksumLines, hashHexBytes(files[path])+"  "+path)
	}
	files["integrity/checksums.sha256"] = []byte(strings.Join(checksumLines, "\n") + "\n")
	return writeZipMemberMap(t, files)
}

func zipMemberMap(t testing.TB, bundle []byte) map[string][]byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip member map: %v", err)
	}
	files := make(map[string][]byte, len(reader.File))
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open zip member %s: %v", member.Name, err)
		}
		payload, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read zip member %s: %v", member.Name, err)
		}
		files[member.Name] = payload
	}
	return files
}

func writeZipMemberMap(t testing.TB, files map[string][]byte) []byte {
	t.Helper()
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, path := range paths {
		member, err := writer.Create(path)
		if err != nil {
			t.Fatalf("create zip member %s: %v", path, err)
		}
		if _, err := member.Write(files[path]); err != nil {
			t.Fatalf("write zip member %s: %v", path, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close v1 zip: %v", err)
	}
	return output.Bytes()
}

func decodeNDJSONRows(t testing.TB, payload []byte) []map[string]any {
	t.Helper()
	lines := bytes.Split(bytes.TrimSpace(payload), []byte("\n"))
	rows := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(line, &row); err != nil {
			t.Fatalf("decode NDJSON row: %v", err)
		}
		rows = append(rows, row)
	}
	return rows
}

func ndjsonRowsByIdentity(t testing.TB, payload []byte, identity string) map[string]map[string]any {
	t.Helper()
	result := map[string]map[string]any{}
	for _, row := range decodeNDJSONRows(t, payload) {
		key, _ := row[identity].(string)
		if key == "" {
			t.Fatalf("NDJSON row has no %s identity: %#v", identity, row)
		}
		result[key] = row
	}
	return result
}

func rewriteZipMembers(t testing.TB, bundle []byte, transform func(path string, data []byte) ([]byte, bool)) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		nextData, keep := transform(member.Name, data)
		if !keep {
			continue
		}
		w, err := writer.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(nextData); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close rewritten zip: %v", err)
	}
	return buf.Bytes()
}

func postRollback(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, recordID string, body map[string]any) *http.Response {
	t.Helper()
	return httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/records/"+recordID+"/rollback", body,
		httptestx.WithCookies(login.SessionCookie, login.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func getRecordHistoryItems(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, recordID string) []any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/records/"+recordID+"/history", nil, httptestx.WithCookies(login.SessionCookie))
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("history items missing: %#v", data)
	}
	return items
}

func requireHistoryItemForChangeSet(t testing.TB, items []any, changeSetID string) map[string]any {
	t.Helper()
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if item["change_set_id"] == changeSetID {
			return item
		}
	}
	t.Fatalf("history item for change_set_id=%s not found in %#v", changeSetID, items)
	return nil
}

func textprotoMIMEHeader(values map[string]string) textproto.MIMEHeader {
	header := textproto.MIMEHeader{}
	for key, value := range values {
		header.Set(key, value)
	}
	return header
}

func countRows(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

type importFinalizationSideEffects struct {
	MembershipRows           int
	DefaultPreferenceRows    int
	UserPreferenceRows       int
	MembershipAuditRows      int
	MembershipProjectionRows int
}

func (s importFinalizationSideEffects) equal(other importFinalizationSideEffects) bool {
	return s == other
}

func snapshotImportFinalizationSideEffects(t testing.TB, db *sql.DB, incidentID string, userID string) importFinalizationSideEffects {
	t.Helper()
	return importFinalizationSideEffects{
		MembershipRows: countRows(t, db, `
SELECT count(*)
  FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
   AND role = 'admin'
   AND added_by_user_id = $2
   AND updated_by_user_id = $2
   AND membership_version = 1
`, incidentID, userID),
		DefaultPreferenceRows: countRows(t, db, `
SELECT count(*)
  FROM incident_workbook_preferences
 WHERE incident_id = $1
   AND default_sheet_ref IS NULL
   AND updated_by_user_id = $2
`, incidentID, userID),
		UserPreferenceRows: countRows(t, db, `
SELECT count(*)
  FROM user_workbook_preferences
 WHERE incident_id = $1
   AND user_id = $2
   AND home_sheet_ref IS NULL
`, incidentID, userID),
		MembershipAuditRows: countRows(t, db, `
SELECT count(*)
  FROM deployment_admin_audit_events
 WHERE incident_id = $1
   AND actor_user_id = $2
   AND target_user_id = $2
   AND event_source = 'incidents'
   AND event_kind = 'incident_membership_created'
`, incidentID, userID),
		MembershipProjectionRows: countRows(t, db, `
SELECT count(*)
  FROM administrative_audit_projections
 WHERE scope_kind = 'incident'
   AND scope_id = $1
   AND actor_user_id = $2
   AND action_code = 'membership_created'
`, incidentID, userID),
	}
}

func startIsolatedIncidentBundleServer(t testing.TB, runtime *appsupport.Runtime, prefix string) *appsupport.ServerHarness {
	t.Helper()
	return startIsolatedIncidentBundleServerWithEnv(t, runtime, prefix, nil)
}

func startIsolatedIncidentBundleServerWithEnv(t testing.TB, runtime *appsupport.Runtime, prefix string, extraEnv map[string]string) *appsupport.ServerHarness {
	t.Helper()
	testDB := runtime.Postgres.PrepareIsolatedDatabaseT(t, prefix)
	bucket, err := runtime.S3.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare isolated target bucket: %v", err)
	}
	env := testDB.Env()
	for key, value := range runtime.S3.Env(bucket) {
		env[key] = value
	}
	for key, value := range extraEnv {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open isolated target db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &appsupport.ServerHarness{Server: server, DB: db}
}

func compareSourceTargetCount(t testing.TB, source *sql.DB, target *sql.DB, query string, incidentID string, label string) {
	t.Helper()
	sourceCount := countRows(t, source, query, incidentID)
	targetCount := countRows(t, target, query, incidentID)
	if targetCount != sourceCount {
		t.Fatalf("imported %s mismatch: target=%d source=%d", label, targetCount, sourceCount)
	}
}

func stringScalar(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRow(query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func exportedBundleTestPath(t testing.TB, server *httptestx.Server, rawReference string) string {
	t.Helper()
	reference, err := incidentbundles.ParseBundleStorageRef(rawReference)
	if err != nil {
		t.Fatalf("database export storage reference %q is invalid: %v", rawReference, err)
	}
	if filepath.IsAbs(rawReference) || strings.Contains(rawReference, server.Config.Roots.ExportOutputs.Path) {
		t.Fatalf("database export storage reference disclosed a host root: %q", rawReference)
	}
	return filepath.Join(server.Config.Roots.ExportOutputs.Path, filepath.FromSlash(reference.String()))
}

func stagedBundleTestPath(t testing.TB, server *httptestx.Server, rawReference string) string {
	t.Helper()
	reference, err := incidentbundles.ParseBundleStagingRef(rawReference)
	if err != nil {
		t.Fatalf("database staging reference %q is invalid: %v", rawReference, err)
	}
	if filepath.IsAbs(rawReference) || strings.Contains(rawReference, server.Config.Roots.TemporaryWork.Path) {
		t.Fatalf("database staging reference disclosed a host root: %q", rawReference)
	}
	return filepath.Join(server.Config.Roots.TemporaryWork.Path, filepath.FromSlash(reference.String()))
}

func jsonRaw(t testing.TB, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture json: %v", err)
	}
	return payload
}

func exportBundleBytes(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, clientTxnID string) []byte {
	t.Helper()
	job := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, login, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": clientTxnID,
	}), http.StatusAccepted)["data"].(map[string]any)
	terminal := waitJob(t, harness.Server, login, job["job_id"].(string))
	ref := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	storageRef := stringScalar(t, harness.DB, `SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id = $1`, ref["id"].(string))
	bundleBytes, err := os.ReadFile(exportedBundleTestPath(t, harness.Server, storageRef))
	if err != nil {
		t.Fatalf("read exported bundle: %v", err)
	}
	return bundleBytes
}

func importBundleAndWait(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, bundle []byte, clientTxnID string) map[string]any {
	t.Helper()
	resp := postImport(t, server, login, `{"client_txn_id":"`+clientTxnID+`"}`, bundle, "bundle.zip")
	job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	return waitJob(t, server, login, job["job_id"].(string))
}

func requireIncidentPortabilityProof(
	t testing.TB,
	db *sql.DB,
	jobID string,
	operationKind string,
) {
	t.Helper()
	var ownerProfileID string
	var actualOperationKind string
	var finalCommitID string
	var terminalCode string
	if err := db.QueryRow(`
SELECT owner_profile_id, operation_kind, final_commit_id,
       terminal_result->>'code'
  FROM extension_job_commit_proofs
 WHERE job_id::text = $1
`, jobID).Scan(
		&ownerProfileID,
		&actualOperationKind,
		&finalCommitID,
		&terminalCode,
	); err != nil {
		t.Fatalf("read Incident Portability proof for job %s: %v", jobID, err)
	}
	if ownerProfileID != incidentbundles.ProfileID ||
		actualOperationKind != operationKind ||
		finalCommitID == "" ||
		terminalCode == "" {
		t.Fatalf(
			"unexpected Incident Portability proof: owner=%q operation=%q commit=%q code=%q",
			ownerProfileID,
			actualOperationKind,
			finalCommitID,
			terminalCode,
		)
	}
}

func requireFailedJobReason(t testing.TB, job map[string]any, wantCode string, wantReason string) {
	t.Helper()
	errorSummary := job["error_summary"].(map[string]any)
	if errorSummary["code"] != wantCode {
		t.Fatalf("failed job code mismatch: got %#v want %s", errorSummary, wantCode)
	}
	details := errorSummary["details"].(map[string]any)
	if details["reason_code"] != wantReason {
		t.Fatalf("failed job reason mismatch: got %#v want %s", details, wantReason)
	}
}

func requireTimelineQueryRow(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, incidentID string, recordID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query", map[string]any{}, httptestx.WithCookies(login.SessionCookie))
	rows := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)["rows"].([]any)
	for _, raw := range rows {
		row := raw.(map[string]any)
		if row["record_id"] == recordID {
			return row
		}
	}
	t.Fatalf("timeline query row %s not found in %#v", recordID, rows)
	return nil
}

func timelineCellValue(t testing.TB, row map[string]any, fieldKey string) any {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell := cells[fieldKey].(map[string]any)
	return cell["value"]
}

func assertImportFailureLeavesState(t testing.TB, harness *appsupport.ServerHarness, login flowtest.LoginResult, incidentID string, clientTxnID string, bundle []byte, wantReason string) map[string]any {
	t.Helper()
	before := snapshotImportFailureState(t, harness, incidentID)
	resp := postImport(t, harness.Server, login, `{"client_txn_id":"`+clientTxnID+`"}`, bundle, "bundle.zip")
	job := httptestx.RequireSuccessEnvelope(t, resp, http.StatusAccepted)["data"].(map[string]any)
	terminal := waitFailedJob(t, harness.Server, login, job["job_id"].(string))
	requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", wantReason)
	if summary, ok := terminal["result_summary"].(map[string]any); ok {
		if refs, ok := summary["resource_refs"].([]any); ok && len(refs) != 0 {
			t.Fatalf("failed import must not expose imported resource refs: %#v", summary)
		}
	}
	var stagingRef string
	if err := harness.DB.QueryRow(`SELECT bundle_staging_ref FROM incident_bundle_job_payloads WHERE job_id = $1`, job["job_id"].(string)).Scan(&stagingRef); err != nil {
		t.Fatalf("query failed import staging reference: %v", err)
	}
	if _, err := os.Stat(stagedBundleTestPath(t, harness.Server, stagingRef)); !os.IsNotExist(err) {
		t.Fatalf("failed import staging reference must be cleaned up, stat err=%v ref=%s", err, stagingRef)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_job_payloads WHERE job_id = $1 AND (imported_incident_id IS NOT NULL OR manifest_sha256 IS NOT NULL)`, job["job_id"].(string)) != 0 {
		t.Fatalf("failed import must not persist imported incident id or manifest sha")
	}
	var requestJSON string
	if err := harness.DB.QueryRow(`SELECT request_json::text FROM incident_bundle_job_payloads WHERE job_id = $1`, job["job_id"].(string)).Scan(&requestJSON); err != nil {
		t.Fatalf("query failed import request json: %v", err)
	}
	var normalized map[string]any
	if err := json.Unmarshal([]byte(requestJSON), &normalized); err != nil {
		t.Fatalf("decode failed import request json: %v", err)
	}
	fileSHA, ok := normalized["file_sha256"].(string)
	if len(normalized) != 1 || !ok || fileSHA == "" || strings.Contains(requestJSON, "manifest.json") {
		t.Fatalf("failed import request payload must retain only upload hash, got %s", requestJSON)
	}
	assertNoIncidentBundleStaging(t, harness.Server)
	after := snapshotImportFailureState(t, harness, incidentID)
	if !before.equal(after) {
		t.Fatalf("failed import left partial state: before=%#v after=%#v", before, after)
	}
	return terminal
}

type importFailureState struct {
	IncidentRows            int
	MembershipRows          int
	ProjectionRows          int
	ImportedActorRows       int
	ImportedAttributionRows int
	ExportRows              int
	ImportedObjectKeys      []string
}

func (s importFailureState) equal(other importFailureState) bool {
	return s.IncidentRows == other.IncidentRows &&
		s.MembershipRows == other.MembershipRows &&
		s.ProjectionRows == other.ProjectionRows &&
		s.ImportedActorRows == other.ImportedActorRows &&
		s.ImportedAttributionRows == other.ImportedAttributionRows &&
		s.ExportRows == other.ExportRows &&
		slices.Equal(s.ImportedObjectKeys, other.ImportedObjectKeys)
}

func snapshotImportFailureState(t testing.TB, harness *appsupport.ServerHarness, incidentID string) importFailureState {
	t.Helper()
	return importFailureState{
		IncidentRows:            countRows(t, harness.DB, `SELECT count(*) FROM incidents WHERE id = $1`, incidentID),
		MembershipRows:          countRows(t, harness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1`, incidentID),
		ProjectionRows:          countRows(t, harness.DB, `SELECT count(*) FROM timeline_grid_projection WHERE incident_id = $1`, incidentID),
		ImportedActorRows:       countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_actors WHERE incident_id = $1`, incidentID),
		ImportedAttributionRows: countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id = $1`, incidentID),
		ExportRows:              countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE incident_id = $1`, incidentID),
		ImportedObjectKeys:      objectKeysWithPrefix(t, harness.Server.Runtime.ObjectStore, "incidents/"+incidentID+"/object-blobs/"),
	}
}

func objectKeysWithPrefix(t testing.TB, store objectstore.Store, prefix string) []string {
	t.Helper()
	objects, err := store.ListObjects(context.Background(), prefix)
	if err != nil {
		t.Fatalf("list object store prefix %s: %v", prefix, err)
	}
	keys := make([]string, 0, len(objects))
	for _, object := range objects {
		keys = append(keys, object.Key)
	}
	slices.Sort(keys)
	return keys
}

func seedMissingIncidentBundleBlob(t testing.TB, harness *appsupport.ServerHarness, incidentID string, actorUserID string) {
	t.Helper()
	missingBytes := []byte("incident-bundle missing blob fixture")
	sha := hashHexBytes(missingBytes)
	if _, err := harness.DB.Exec(`
INSERT INTO object_blobs (
    incident_id,
    created_by_user_id,
    storage_key,
    upload_state,
    byte_size,
    filename_hint,
    content_type_hint,
    expected_sha256_hex,
    observed_size,
    observed_content_type,
    observed_sha256_hex,
    target_expires_at,
    pending_expires_at,
    finalized_at
)
VALUES ($1, $2, $3, 'available', $4, 'missing.txt', 'text/plain', $5, $4, 'text/plain', $5, now() + interval '1 hour', now() + interval '1 hour', now())
`, incidentID, actorUserID, "extension_profile/missing/"+incidentID+"/"+sha, len(missingBytes), sha); err != nil {
		t.Fatalf("seed missing object blob row: %v", err)
	}
}

type envelopeDurability struct {
	Jobs        int
	Payloads    int
	Idempotency int
}

func snapshotEnvelopeDurability(t testing.TB, db *sql.DB) envelopeDurability {
	t.Helper()
	return envelopeDurability{
		Jobs:        countRows(t, db, `SELECT count(*) FROM jobs`),
		Payloads:    countRows(t, db, `SELECT count(*) FROM incident_bundle_job_payloads`),
		Idempotency: countRows(t, db, `SELECT count(*) FROM route_idempotency WHERE route_key = 'incident_bundles.import'`),
	}
}

func assertNoIncidentBundleStaging(t testing.TB, server *httptestx.Server) {
	t.Helper()
	stagingDir := filepath.Join(server.Config.Roots.TemporaryWork.Path, "incident-bundles", "imports")
	entries, err := os.ReadDir(stagingDir)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		t.Fatalf("read staging dir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("incident bundle staging dir must remain empty, found %d entries in %s", len(entries), stagingDir)
	}
}

type uploadPart struct {
	Name        string
	Filename    string
	ContentType string
	Body        []byte
}

func jsonUploadPart(name string, filename string, body string) uploadPart {
	return uploadPart{Name: name, Filename: filename, ContentType: "application/json; charset=utf-8", Body: []byte(body)}
}

func fileUploadPart(name string, filename string, contentType string, body []byte) uploadPart {
	return uploadPart{Name: name, Filename: filename, ContentType: contentType, Body: body}
}

func newImportEnvelopeRequest(t testing.TB, server *httptestx.Server, login flowtest.LoginResult, parts []uploadPart) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		disposition := `form-data; name="` + part.Name + `"`
		if part.Filename != "" {
			disposition += `; filename="` + part.Filename + `"`
		}
		w, err := writer.CreatePart(textprotoMIMEHeader(map[string]string{
			"Content-Disposition": disposition,
			"Content-Type":        part.ContentType,
		}))
		if err != nil {
			t.Fatalf("create multipart part %s: %v", part.Name, err)
		}
		if _, err := w.Write(part.Body); err != nil {
			t.Fatalf("write multipart part %s: %v", part.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatalf("create import request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	addImportAuth(req, login)
	return req
}

func addImportAuth(req *http.Request, login flowtest.LoginResult) {
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
}

func stringArray(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected JSON array, got %T %#v", raw, raw)
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("expected string array item, got %T %#v", item, item)
		}
		result = append(result, value)
	}
	return result
}

func hashHexBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
