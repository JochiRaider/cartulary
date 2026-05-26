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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase11_I_11_INCIDENT_BUNDLES_01_ExportJobIdempotencyAndDescriptor(t *testing.T) {
	withIncidentPortabilityClaimed(t)
	harness := phase2test.StartRuntime(t).StartServer(t, "phase11-incident-bundle-export")
	admin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-source",
		"incident_key":  "BUNDLE-EXPORT",
		"title":         "Incident bundle export",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateTimelineRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":    "txn-incident-bundle-row",
		"timeline.summary": "Portable event",
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

	terminal := waitJob(t, harness.Server, admin, job["job_id"].(string))
	summary := terminal["result_summary"].(map[string]any)
	if summary["code"] != incidentbundles.ResultIncidentBundleExported {
		t.Fatalf("unexpected export summary: %#v", summary)
	}
	refs := summary["resource_refs"].([]any)
	if len(refs) != 1 || refs[0].(map[string]any)["kind"] != "incident_bundle" {
		t.Fatalf("export summary must contain one incident_bundle ref: %#v", refs)
	}
	descriptorRoute := refs[0].(map[string]any)["route"].(string)
	resp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+descriptorRoute, nil, phase2test.WithCookies(admin.SessionCookie))
	descriptor := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	if descriptor["history_mode"] != "full" || descriptor["blob_mode"] != "full" || descriptor["manifest_sha256"] == "" {
		t.Fatalf("descriptor missing fixed modes or manifest hash: %#v", descriptor)
	}
}

func TestPhase11_I_11_INCIDENT_BUNDLES_03_ExportJobAuthorizationReDerivesIncidentMembership(t *testing.T) {
	withIncidentPortabilityClaimed(t)
	harness := phase2test.StartRuntime(t).StartServer(t, "phase11-incident-bundle-export-auth")
	admin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-source",
		"incident_key":  "BUNDLE-EXPORT-AUTH",
		"title":         "Incident bundle export auth",
	})
	incidentID := incident["incident_id"].(string)
	phase2test.CreateTimelineRow(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id":    "txn-incident-bundle-export-auth-row",
		"timeline.summary": "Portable event for auth",
	})

	submitterPassword := "BundleSubmitterPassphrase11!"
	memberAdminPassword := "BundleMemberAdminPassphrase11!"
	memberOnlyPassword := "BundleMemberOnlyPassphrase11!"
	nonmemberAdminPassword := "BundleNonmemberAdminPassphrase11!"
	submitterUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-bundle-submitter@example.test", "Phase11 Bundle Submitter", submitterPassword, false, true, true)
	memberAdminUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-bundle-member-admin@example.test", "Phase11 Bundle Member Admin", memberAdminPassword, false, true, true)
	memberOnlyUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-bundle-member-only@example.test", "Phase11 Bundle Member Only", memberOnlyPassword, false, false, true)
	nonmemberAdminUser := phase2test.SeedLocalUserRecord(t, harness.DB, "phase11-bundle-nonmember-admin@example.test", "Phase11 Bundle Nonmember Admin", nonmemberAdminPassword, false, true, true)
	submitterCookies, submitterCSRF := phase2test.LoginLocalUser(t, harness.Server, submitterUser.Email, submitterPassword)
	memberAdminCookies, memberAdminCSRF := phase2test.LoginLocalUser(t, harness.Server, memberAdminUser.Email, memberAdminPassword)
	memberOnlyCookies, _ := phase2test.LoginLocalUser(t, harness.Server, memberOnlyUser.Email, memberOnlyPassword)
	nonmemberAdminCookies, nonmemberAdminCSRF := phase2test.LoginLocalUser(t, harness.Server, nonmemberAdminUser.Email, nonmemberAdminPassword)
	submitterLogin := phase2test.LoginResult{SessionCookie: submitterCookies, CSRFCookie: submitterCSRF}
	memberAdminLogin := phase2test.LoginResult{SessionCookie: memberAdminCookies, CSRFCookie: memberAdminCSRF}

	phase2test.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-submitter-membership",
		"user_id":       submitterUser.ID.String(),
		"role":          "viewer",
	})
	memberAdminMembership := phase2test.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-member-admin-membership",
		"user_id":       memberAdminUser.ID.String(),
		"role":          "viewer",
	})
	phase2test.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
		"client_txn_id": "txn-incident-bundle-export-auth-member-only-membership",
		"user_id":       memberOnlyUser.ID.String(),
		"role":          "admin",
	})

	started := make(chan struct{})
	release := make(chan struct{})
	var startedOnce sync.Once
	var releaseOnce sync.Once
	restoreHook := incidentbundles.SetIncidentBundleWorkerStartHookForTesting(func(jobKind string) {
		if jobKind != "export" {
			return
		}
		startedOnce.Do(func() { close(started) })
		<-release
	})
	t.Cleanup(restoreHook)
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }) })

	exportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, harness.Server, submitterLogin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-auth-blocked",
	}), http.StatusAccepted)["data"].(map[string]any)
	jobID := exportJob["job_id"].(string)
	select {
	case <-started:
	case <-time.After(5 * time.Second):
		t.Fatal("incident bundle export worker did not reach test hook")
	}

	submitterRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(submitterCookies))
	httptestx.RequireSuccessEnvelope(t, submitterRead, http.StatusOK)
	memberAdminRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(memberAdminCookies))
	httptestx.RequireSuccessEnvelope(t, memberAdminRead, http.StatusOK)
	memberOnlyRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(memberOnlyCookies))
	httptestx.RequireErrorEnvelope(t, memberOnlyRead, http.StatusNotFound, "job_not_found")
	nonmemberAdminRead := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(nonmemberAdminCookies))
	httptestx.RequireErrorEnvelope(t, nonmemberAdminRead, http.StatusNotFound, "job_not_found")
	nonmemberAdminCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-nonmember-cancel",
	}, phase2test.WithCookies(nonmemberAdminCookies, nonmemberAdminCSRF), phase2test.WithHeader(authn.CSRFHeaderName, nonmemberAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, nonmemberAdminCancel, http.StatusNotFound, "job_not_found")
	memberViewerCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-member-viewer-cancel",
	}, phase2test.WithCookies(memberAdminCookies, memberAdminCSRF), phase2test.WithHeader(authn.CSRFHeaderName, memberAdminCSRF.Value))
	httptestx.RequireErrorEnvelope(t, memberViewerCancel, http.StatusForbidden, "authorization_denied")

	phase2test.PatchMembership(t, harness.Server, admin, incidentID, memberAdminUser.ID.String(), map[string]any{
		"base_membership_version": memberAdminMembership["membership_version"],
		"role":                    "admin",
	})
	memberAdminCancel := phase2test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID+"/cancel", map[string]any{
		"client_txn_id": "txn-export-auth-member-admin-cancel",
	}, phase2test.WithCookies(memberAdminCookies, memberAdminCSRF), phase2test.WithHeader(authn.CSRFHeaderName, memberAdminCSRF.Value))
	httptestx.RequireSuccessEnvelope(t, memberAdminCancel, http.StatusOK)

	releaseOnce.Do(func() { close(release) })
	terminal := waitJobWithStatus(t, harness.Server, memberAdminLogin, jobID, "canceled")
	if terminal["status"] != "canceled" {
		t.Fatalf("export job must stop at canceled after authorized cancel: %#v", terminal)
	}
}

func TestPhase11_I_11_INCIDENT_BUNDLES_02_ImportEnvelopeIdempotencyAndImportedIncidentOpen(t *testing.T) {
	withIncidentPortabilityClaimed(t)
	runtime := phase2test.StartRuntime(t)
	sourceHarness := runtime.StartServer(t, "phase11-incident-bundle-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "phase11-incident-bundle-target")
	sourceAdmin, sourceAdminID := phase2test.ProvisionBootstrapAdmin(t, sourceHarness.Server)
	targetAdmin, targetAdminID := phase2test.ProvisionBootstrapAdmin(t, targetHarness.Server)
	incident := phase2test.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-import-source",
		"incident_key":  "BUNDLE-IMPORT",
		"title":         "Incident bundle import",
	})
	incidentID := incident["incident_id"].(string)
	row := phase2test.CreateTimelineRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":    "txn-incident-bundle-import-row",
		"timeline.summary": "Imported portable event",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	seededState := seedIncidentBundlePortableState(t, sourceHarness, incidentID, recordID, sourceAdminID)
	sourceViewerID := phase2test.SeedLocalUserFlags(t, sourceHarness.DB, "phase11-import-source-viewer@example.test", "Phase 11 Import Source Viewer", "Phase11ImportViewer1!", false, false, true)
	phase2test.CreateMembership(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
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
	var bundlePath string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_storage_path FROM incident_bundle_exports WHERE bundle_id = $1`, bundleID).Scan(&bundlePath); err != nil {
		t.Fatalf("query exported bundle path: %v", err)
	}
	bundleBytes, err := os.ReadFile(bundlePath)
	if err != nil {
		t.Fatalf("read exported bundle: %v", err)
	}
	secondExportJob := httptestx.RequireSuccessEnvelope(t, postExport(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"incident_id":   incidentID,
		"client_txn_id": "txn-export-for-import-determinism",
	}), http.StatusAccepted)["data"].(map[string]any)
	secondExportTerminal := waitJob(t, sourceHarness.Server, sourceAdmin, secondExportJob["job_id"].(string))
	secondExportRef := secondExportTerminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	var secondBundlePath string
	if err := sourceHarness.DB.QueryRow(`SELECT bundle_storage_path FROM incident_bundle_exports WHERE bundle_id = $1`, secondExportRef["id"].(string)).Scan(&secondBundlePath); err != nil {
		t.Fatalf("query second exported bundle path: %v", err)
	}
	secondBundleBytes, err := os.ReadFile(secondBundlePath)
	if err != nil {
		t.Fatalf("read second exported bundle: %v", err)
	}
	for _, memberPath := range []string{
		"integrity/checksums.sha256",
		"data/record_tags.ndjson",
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
	var failedStagingPath string
	if err := targetHarness.DB.QueryRow(`SELECT bundle_staging_path FROM incident_bundle_job_payloads WHERE job_id = $1`, failedJob["job_id"].(string)).Scan(&failedStagingPath); err != nil {
		t.Fatalf("query failed import staging path: %v", err)
	}
	if _, err := os.Stat(failedStagingPath); !os.IsNotExist(err) {
		t.Fatalf("failed import staging path must be cleaned up, stat err=%v path=%s", err, failedStagingPath)
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
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID, "change_set count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_revisions rr JOIN records r ON r.record_id = rr.record_id WHERE r.incident_id = $1`, incidentID, "revision count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_links WHERE incident_id = $1`, incidentID, "record-link count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM record_tags WHERE incident_id = $1`, incidentID, "record-tag attachment count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM evidence_custody_events WHERE incident_id = $1`, incidentID, "evidence custody count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM entity_mentions em JOIN records r ON r.record_id = em.source_record_id WHERE r.incident_id = $1`, incidentID, "entity mention count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM indicator_observations WHERE incident_id = $1`, incidentID, "indicator observation count")
	compareSourceTargetCount(t, sourceHarness.DB, targetHarness.DB, `SELECT count(*) FROM records WHERE incident_id = $1 AND row_version = 2`, incidentID, "row_version=2 record count")
	openResp := phase2test.DoJSON(t, http.MethodGet, targetHarness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/workbook-startup", nil, phase2test.WithCookies(targetAdmin.SessionCookie))
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
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_memberships WHERE incident_id = $1 AND user_id = $2`, incidentID, sourceViewerID) != 0 {
		t.Fatalf("import must not recreate source deployment-local membership")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_bundle_imported_actors WHERE incident_id = $1 AND source_actor_id = $2 AND local_user_id IS NULL`, incidentID, sourceAdminID) == 0 {
		t.Fatalf("import must preserve historical attribution as inert actor descriptors")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id = $1 AND source_table = 'change_sets' AND source_column = 'actor_user_id' AND source_actor_id = $2 AND local_user_id = $3`, incidentID, sourceAdminID, targetAdminID) == 0 {
		t.Fatalf("import must retain source actor attribution sidecars")
	}
	if countRows(t, targetHarness.DB, `SELECT count(*) FROM record_tags WHERE incident_id = $1 AND record_id = $2 AND normalized_tag_name = 'phase11-portability'`, incidentID, recordID) != 1 {
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
	wantStoragePrefix := "incident-bundles/imported/" + incidentID + "/sha256/"
	if !strings.HasPrefix(importedStorageKey, wantStoragePrefix) {
		t.Fatalf("imported blob must use target-owned storage key, got %s want prefix %s", importedStorageKey, wantStoragePrefix)
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
	historyResp := phase2test.DoJSON(t, http.MethodGet, targetHarness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/history", nil, phase2test.WithCookies(targetAdmin.SessionCookie))
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
}

type seededIncidentBundlePortableState struct {
	BlobBytes                []byte
	BlobSHA                  string
	ObjectBlobID             string
	EvidenceRecordID         string
	HistoryHostRecordID      string
	ReversibleChangeSetID    string
	NonReversibleChangeSetID string
}

func seedIncidentBundlePortableState(t testing.TB, harness *phase2test.ServerHarness, incidentID string, timelineRecordID string, actorUserID string) seededIncidentBundlePortableState {
	t.Helper()
	ctx := context.Background()
	incidentUUID := uuid.MustParse(incidentID)
	timelineUUID := uuid.MustParse(timelineRecordID)
	actorUUID := uuid.MustParse(actorUserID)
	if _, err := harness.DB.Exec(`
INSERT INTO record_tags (incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, 'Phase11 Portability', 'phase11-portability', $3)
`, incidentID, timelineRecordID, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}

	historyHostID := uuid.New()
	insertHostRecord(t, harness.DB, incidentUUID, historyHostID, actorUUID, "portable host before", "portable-host")

	indicatorID := uuid.New()
	if _, err := harness.DB.Exec(`
INSERT INTO records (record_id, incident_id, record_type, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'indicator', $3, $3)
`, indicatorID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed indicator envelope: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO indicators (
    record_id, incident_id, indicator_type, value_kind, display_value, normalized_value,
    dedupe_key, row_version, created_by_user_id, updated_by_user_id
)
VALUES ($1, $2, 'domain', 'atomic', 'portable.example.test', 'portable.example.test', 'domain:portable.example.test', 1, $3, $3)
`, indicatorID, incidentUUID, actorUUID); err != nil {
		t.Fatalf("seed indicator row: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO indicator_observations (
    incident_id, source_record_id, source_field_key, origin_kind, origin_locator,
    observed_text, parsed_indicator_type, normalized_candidate, resolution_status,
    resolved_indicator_record_id, row_version, created_by_user_id, resolved_by_user_id,
    resolved_at, resolution_method
)
VALUES ($1, $2, 'timeline.summary', 'auto_extract', 'phase11', 'portable.example.test', 'domain', 'portable.example.test', 'resolved', $3, 1, $4, $4, now(), 'fixture')
`, incidentUUID, timelineUUID, indicatorID, actorUUID); err != nil {
		t.Fatalf("seed indicator observation: %v", err)
	}
	if _, err := harness.DB.Exec(`
INSERT INTO entity_mentions (
    source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal,
    created_by_user_id, resolved_record_id, resolved_by_user_id, resolved_at, resolution_method
)
VALUES ($1, 'host', 'timeline.summary', 'manual', 'phase11', 'portable host', 'portable host', 'resolved', 1, 1, $2, $3, $2, now(), 'fixture')
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

	reversibleChangeSetID := uuid.New()
	seedPortableRollbackHostPatch(t, harness.DB, incidentUUID, historyHostID, actorUUID, reversibleChangeSetID, time.Date(2026, 5, 25, 17, 0, 0, 0, time.UTC), "portable host before", "portable host after")
	nonReversibleChangeSetID := uuid.New()
	seedPortableRecordTagCreateHistory(t, harness.DB, incidentUUID, historyHostID, actorUUID, nonReversibleChangeSetID, time.Date(2026, 5, 25, 17, 1, 0, 0, time.UTC))

	blobBytes := []byte("phase11 incident bundle blob\n")
	sum := sha256.Sum256(blobBytes)
	blobSHA := hex.EncodeToString(sum[:])
	sourceStorageKey := "phase11/source/" + incidentID + "/" + blobSHA
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
VALUES ($1, $2, $3, 'available', $4, 'phase11.txt', 'text/plain', $5, $4, 'text/plain', $5, now() + interval '1 hour', now() + interval '1 hour', now())
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
		ReversibleChangeSetID:    reversibleChangeSetID.String(),
		NonReversibleChangeSetID: nonReversibleChangeSetID.String(),
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
		"tag_name":  "Phase11 History",
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id, created_at, updated_at)
VALUES ($1, $2, $3, 'Phase11 History', 'phase11-history', $4, $5, $5)
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

func postExport(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, body map[string]any) *http.Response {
	t.Helper()
	return phase2test.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/incident-bundles/export", body,
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func postImport(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, metadata string, file []byte, filename string) *http.Response {
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

func waitJob(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(login.SessionCookie))
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

func waitJobWithStatus(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, jobID string, wantStatus string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(login.SessionCookie))
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

func waitFailedJob(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, phase2test.WithCookies(login.SessionCookie))
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

func postRollback(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, recordID string, body map[string]any) *http.Response {
	t.Helper()
	return phase2test.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/records/"+recordID+"/rollback", body,
		phase2test.WithCookies(login.SessionCookie, login.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value),
	)
}

func getRecordHistoryItems(t testing.TB, server *httptestx.Server, login phase2test.LoginResult, recordID string) []any {
	t.Helper()
	resp := phase2test.DoJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/records/"+recordID+"/history", nil, phase2test.WithCookies(login.SessionCookie))
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

func startIsolatedIncidentBundleServer(t testing.TB, runtime *phase2test.RuntimeHarness, prefix string) *phase2test.ServerHarness {
	t.Helper()
	testDB := runtime.Postgres.PrepareDatabaseT(t, prefix)
	bucket, err := runtime.S3.BootstrapBucket(context.Background(), prefix)
	if err != nil {
		t.Fatalf("prepare isolated target bucket: %v", err)
	}
	env := testDB.Env()
	for key, value := range runtime.S3.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")
	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open isolated target db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &phase2test.ServerHarness{Server: server, DB: db}
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

func jsonRaw(t testing.TB, value any) []byte {
	t.Helper()
	payload, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal fixture json: %v", err)
	}
	return payload
}

func withIncidentPortabilityClaimed(t testing.TB) {
	t.Helper()
	restore := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{ProfileID: "enterprise_authentication", Claimed: false, RouteFamilies: []string{"/api/v1/auth/oidc", "/api/v1/auth/providers", "/api/v1/auth/saml", "/api/v1/users/{user_id}/auth-bindings"}},
		{ProfileID: "import", Claimed: true, RouteFamilies: []string{"/api/v1/import-sessions"}},
		{ProfileID: incidentbundles.ProfileID, Claimed: true, RouteFamilies: []string{"/api/v1/incident-bundles"}},
		{ProfileID: "reference_pack", Claimed: true, RouteFamilies: []string{"/api/v1/reference-packs"}},
		{ProfileID: "snapshot_reporting", Claimed: true, RouteFamilies: []string{"/api/v1/releases", "/api/v1/snapshots"}},
	})
	t.Cleanup(restore)
}
