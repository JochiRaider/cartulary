package evidence_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase5_ObjectUploadAttachWorkbookProjection_I_5_01(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-upload-attach-projection")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-i-01-incident",
		"incident_key":  "phase5-i-01",
		"title":         "Phase 5 upload attach projection",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	timelineData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v1", map[string]any{
		"client_txn_id":    "txn-phase5-i-01-timeline",
		"timeline.summary": "Endpoint screenshot received",
	})
	timelineRecordID := phase4test.MustUUID(t, timelineData["row"].(map[string]any)["record_id"].(string))
	evidenceData := requirePhase5HTTPWorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":                 "txn-phase5-i-01-evidence",
		"evidence.title":                "Endpoint screenshot",
		"evidence.collector_party_text": "IR collector",
	})
	evidenceRecordID := phase4test.MustUUID(t, evidenceData["row"].(map[string]any)["record_id"].(string))
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', $4, $4, now(), now())
`, incidentID, timelineRecordID, evidenceRecordID, adminID); err != nil {
		t.Fatalf("insert timeline attached evidence link: %v", err)
	}

	payload := []byte("phase5 projection object")
	sum := sha256.Sum256(payload)
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-i-01-blob",
		"byte_size":         len(payload),
		"filename_hint":     " endpoint.txt ",
		"content_type_hint": "text/plain",
		"sha256_hex":        fmt.Sprintf("%x", sum[:]),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, createData["upload_target"].(map[string]any)["href"].(string), payload, "text/plain")

	attachBody := map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    "txn-phase5-i-01-attach",
	}
	attachResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	attachData := httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
	if attachData["object_blob_id"] != createData["object_blob_id"] {
		t.Fatalf("attach object_blob_id got %#v want %#v", attachData["object_blob_id"], createData["object_blob_id"])
	}
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)

	replayResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+evidenceRecordID.String()+"/attach-blob", attachBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	if replayData["change_set_id"] != attachData["change_set_id"] {
		t.Fatalf("attach replay changed change_set_id: replay=%#v first=%#v", replayData["change_set_id"], attachData["change_set_id"])
	}
	requireTimelineEvidenceProjection(t, harness, login, incidentID, timelineRecordID, 1, true)
	if got := countEvidenceRevisions(t, harness, evidenceRecordID); got != 2 {
		t.Fatalf("attach replay created duplicate record revision: got %d want 2", got)
	}
	if got := countEvidenceBlobLinks(t, harness, evidenceRecordID); got != 1 {
		t.Fatalf("evidence row has duplicate or missing blob link: got %d want 1", got)
	}
}

func TestPhase5_ExpiredSlotReplay_I_5_02(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-expired-slot-replay")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-expired-slot-incident",
		"incident_key":  "phase5-expired-slot",
		"title":         "Phase 5 expired slot replay",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	issuedAt := time.Now().UTC().Add(time.Minute).Truncate(time.Second)
	httptestx.SetClockFixed(t, harness.Server, issuedAt)
	createBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-expired-slot",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	requireCreateExpiry(t, createData, "target_expires_at", issuedAt.Add(60*time.Minute))
	requireCreateExpiry(t, createData, "pending_expires_at", issuedAt.Add(24*time.Hour))

	replayAt := issuedAt.Add(61 * time.Minute)
	extendSessionForClockJump(t, harness, adminID, replayAt.Add(30*time.Minute))
	httptestx.SetClockFixed(t, harness.Server, replayAt)
	replayResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", createBody, authOptions(login)...)
	replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
	for _, key := range []string{"object_blob_id", "target_expires_at", "pending_expires_at"} {
		if replayData[key] != createData[key] {
			t.Fatalf("expired replay refreshed %s: replay=%#v create=%#v", key, replayData, createData)
		}
	}
	replayedTarget := mustParseTime(t, replayData["target_expires_at"].(string))
	if !replayedTarget.Before(replayAt) {
		t.Fatalf("replayed target should remain expired: got %s replay_at %s", replayedTarget, replayAt)
	}

	freshBody := map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     "txn-phase5-expired-slot-fresh",
		"byte_size":         17,
		"filename_hint":     " expired.txt ",
		"content_type_hint": "text/plain",
	}
	freshResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", freshBody, authOptions(login)...)
	freshData := httptestx.RequireSuccessEnvelope(t, freshResp, http.StatusCreated)["data"].(map[string]any)
	if freshData["object_blob_id"] == createData["object_blob_id"] {
		t.Fatalf("fresh client_txn_id reused expired object_blob_id: %#v", freshData)
	}
	requireCreateExpiry(t, freshData, "target_expires_at", replayAt.Add(60*time.Minute))
	requireCreateExpiry(t, freshData, "pending_expires_at", replayAt.Add(24*time.Hour))
	if got := countObjectBlobs(t, harness, incidentID); got != 2 {
		t.Fatalf("fresh target should create exactly one additional blob slot: got %d want 2", got)
	}
}

func TestPhase5_HandleRedeemInvalidatesOnCurrentStateLoss_I_5_03(t *testing.T) {
	harness := phase4test.StartServer(t, "phase5-handle-redeem-invalidates")
	login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-phase5-i-03-incident",
		"incident_key":  "phase5-i-03",
		"title":         "Phase 5 handle invalidation",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	activeLogin := login
	activeAdminID := adminID
	activeIncidentID := incidentID

	t.Run("membership loss and wrong session hide the handle", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("membership body"), "membership.txt", "text/plain", "txn-phase5-i-03-membership-blob", "txn-phase5-i-03-membership-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, activeIncidentID, activeAdminID); err != nil {
			t.Fatalf("remove incident membership: %v", err)
		}
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
			http.StatusNotFound,
			"handle_not_found_or_revoked",
		)

		phase4test.SeedIncidentMembership(t, harness.DB, activeIncidentID, activeAdminID, "Bootstrap Admin", "admin", activeAdminID)
		otherUser := phase4test.SeedLocalUserFlags(t, harness.DB, "phase5-i-03-other@example.test", "Phase5 Other", "Phase5Other1!", false, false, true)
		phase4test.SeedIncidentMembership(t, harness.DB, activeIncidentID, otherUser.ID, "Phase5 Other", "admin", activeAdminID)
		otherLogin := loginLocalUserNoMFA(t, harness, "phase5-i-03-other@example.test", "Phase5Other1!")
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(otherLogin.SessionCookie)),
			http.StatusNotFound,
			"handle_not_found_or_revoked",
		)
	})

	scenarios := []struct {
		name       string
		reasonCode string
		mutate     func(recordID uuid.UUID, objectBlobID uuid.UUID)
	}{
		{name: "blob detach", reasonCode: "no_visible_blob", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateEvidenceBlobLink(t, harness, recordID, nil)
		}},
		{name: "blob replacement", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			replacementID := linkSeededBlob(t, harness, activeIncidentID, activeAdminID, recordID, "available", "available", "phase5/i-03/replacement-"+recordID.String())
			updateEvidenceBlobLink(t, harness, recordID, &replacementID)
		}},
		{name: "evidence delete", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateRecordDeletedState(t, harness, recordID, true)
		}},
		{name: "evidence restore", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateRecordDeletedState(t, harness, recordID, true)
			updateRecordDeletedState(t, harness, recordID, false)
		}},
		{name: "quarantine", reasonCode: "evidence_quarantined", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "quarantined")
		}},
		{name: "pending transition", reasonCode: "blob_pending", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "pending")
		}},
		{name: "failed transition", reasonCode: "blob_failed", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobState(t, harness, objectBlobID, "failed")
		}},
		{name: "object missing", reasonCode: "blob_missing", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobStorageKey(t, harness, objectBlobID, "phase5/i-03/missing-"+recordID.String())
		}},
		{name: "metadata mismatch", reasonCode: "evidence_inconsistent", mutate: func(recordID uuid.UUID, objectBlobID uuid.UUID) {
			updateBlobObservedMetadata(t, harness, objectBlobID, int64(999), "text/plain", "")
		}},
	}

	for _, scenario := range scenarios {
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			t.Run(endpoint+" "+scenario.name, func(t *testing.T) {
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
				attachData := attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("phase5 invalidation"), "invalidate.txt", "text/plain", "txn-"+recordID.String()+"-blob", "txn-"+recordID.String()+"-attach")
				objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
				handle := issueEvidenceHandle(t, harness, activeLogin, recordID, endpoint)
				scenario.mutate(recordID, objectBlobID)
				requireEvidenceAccessUnavailableReason(t,
					phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
					scenario.reasonCode,
				)
			})
		}
	}

	t.Run("logout revokes issuing session before lookup", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("logout body"), "logout.txt", "text/plain", "txn-phase5-i-03-logout-blob", "txn-phase5-i-03-logout-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		logout := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/logout", map[string]any{}, authOptions(activeLogin)...)
		httptestx.RequireSuccessEnvelope(t, logout, http.StatusOK)
		httptestx.RequireErrorEnvelope(t,
			phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(activeLogin.SessionCookie)),
			http.StatusUnauthorized,
			"session_required",
		)
	})
}

func TestPhase5_QuarantineBoundaryPreservesTwoStepAttach_I_5_04(t *testing.T) {
	t.Run("AC-405 object bytes stay outside structured state and loss fails closed", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-object-boundary")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-boundary-incident",
			"incident_key":  "phase5-i-04-boundary",
			"title":         "Phase 5 object boundary",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)

		marker := "phase5-ac405-marker-" + uuid.NewString() + "-payload"
		payload := []byte("prefix-" + marker + "-suffix")
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, payload, "boundary.txt", "text/plain", "txn-phase5-i-04-boundary-blob", "txn-phase5-i-04-boundary-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))

		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+preview["href"].(string), login)); got != string(payload) {
			t.Fatalf("preview before object loss got %q want %q", got, string(payload))
		}
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
			t.Fatalf("download before object loss got %q want %q", got, string(payload))
		}

		structured := structuredTableText(t, harness,
			"records", "evidence", "object_blobs", "route_idempotency",
			"change_sets", "change_set_mutations", "record_revisions", "evidence_access_handles",
		)
		if strings.Contains(structured, marker) {
			t.Fatalf("structured Postgres state contains inline evidence payload marker %q", marker)
		}

		storageKey := blobStorageKey(t, harness, objectBlobID)
		if err := harness.Server.Runtime.ObjectStore.DeleteObject(context.Background(), storageKey); err != nil {
			t.Fatalf("delete object bytes: %v", err)
		}
		if got := countEvidenceBlobLinks(t, harness, recordID); got != 1 {
			t.Fatalf("object loss changed committed evidence blob link count: got %d want 1", got)
		}
		requireEvidenceStates(t, harness, recordID, "available", "available")
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"blob_missing",
			)
		}
	})

	t.Run("failed unattached cleanup deletes bytes and retains metadata", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-cleanup")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-cleanup-incident",
			"incident_key":  "phase5-i-04-cleanup",
			"title":         "Phase 5 cleanup",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		now := time.Now().UTC().Truncate(time.Second)

		cleanupBlobID := uuid.New()
		cleanupKey := "phase5/i-04/cleanup/" + cleanupBlobID.String()
		cleanupPayload := "phase5 cleanup orphan bytes"
		if err := harness.Server.Runtime.ObjectStore.PutObject(context.Background(), cleanupKey, strings.NewReader(cleanupPayload), int64(len(cleanupPayload)), "text/plain"); err != nil {
			t.Fatalf("put cleanup candidate object: %v", err)
		}
		insertFailedCleanupBlob(t, harness, incidentID, adminID, cleanupBlobID, cleanupKey, now)

		expiredBlobID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, expiredBlobID, "phase5/i-04/expired/"+expiredBlobID.String(), now)

		result, err := evidence.NewStore(harness.Server.Runtime.Postgres).CleanupFailedUnattachedBlobBytes(context.Background(), harness.Server.Runtime.ObjectStore, now, 10)
		if err != nil {
			t.Fatalf("cleanup failed unattached blob bytes: %v", err)
		}
		if result.ExpiredPendingCount != 1 || result.CleanedBlobCount != 1 {
			t.Fatalf("cleanup result got expired=%d cleaned=%d want 1/1", result.ExpiredPendingCount, result.CleanedBlobCount)
		}
		if _, err := harness.Server.Runtime.ObjectStore.StatObject(context.Background(), cleanupKey); err == nil {
			t.Fatalf("cleanup candidate object bytes still exist at %s", cleanupKey)
		}
		requireCleanedFailedBlobMetadata(t, harness, cleanupBlobID)
		requireExpiredPendingFailed(t, harness, expiredBlobID)
		if got := countEvidenceRowsForBlob(t, harness, cleanupBlobID); got != 0 {
			t.Fatalf("cleanup created or retained evidence link rows: got %d want 0", got)
		}
	})

	t.Run("quarantine bridges evidence and blocks attach preview and download", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase5-i-04-quarantine")
		login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
			"client_txn_id": "txn-phase5-i-04-quarantine-incident",
			"incident_key":  "phase5-i-04-quarantine",
			"title":         "Phase 5 quarantine",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
		attachData := attachUploadedBlobWithMetadata(t, harness, login, incidentID, recordID, []byte("phase5 quarantine body"), "quarantine.txt", "text/plain", "txn-phase5-i-04-quarantine-blob", "txn-phase5-i-04-quarantine-attach")
		objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
		preview := issueEvidenceHandle(t, harness, login, recordID, "preview-handle")
		download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
		beforeRevisions := countEvidenceRevisions(t, harness, recordID)

		store := evidence.NewStore(harness.Server.Runtime.Postgres)
		if _, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "unsupported_trigger", "req-phase5-i-04-bad-trigger", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("unsupported quarantine trigger got %v want ErrIllegalBlobTransition", err)
		}
		result, err := store.QuarantineBlob(context.Background(), adminID, objectBlobID, "content_inspection_quarantine", "req-phase5-i-04-quarantine", time.Now().UTC())
		if err != nil {
			t.Fatalf("quarantine blob: %v", err)
		}
		if result.ChangedEvidenceRecord != 1 || result.ChangeSetID == uuid.Nil {
			t.Fatalf("quarantine result got changed=%d change_set=%s", result.ChangedEvidenceRecord, result.ChangeSetID)
		}
		requireEvidenceStates(t, harness, recordID, "quarantined", "quarantined")
		if got := countEvidenceRevisions(t, harness, recordID); got != beforeRevisions+1 {
			t.Fatalf("quarantine revision count got %d want %d", got, beforeRevisions+1)
		}
		requireChangeSetSource(t, harness, result.ChangeSetID, "evidence.blob.quarantine")

		for _, handle := range []map[string]any{preview, download} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, phase4test.WithCookies(login.SessionCookie)),
				"evidence_quarantined",
			)
		}
		for _, endpoint := range []string{"preview-handle", "download-handle"} {
			requireEvidenceAccessUnavailableReason(t,
				phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/"+endpoint, map[string]any{}, authOptions(login)...),
				"evidence_quarantined",
			)
		}
		secondRecordID := uuid.New()
		seedEvidenceRecord(t, harness, incidentID, adminID, secondRecordID)
		requireEvidenceAccessUnavailableReason(t,
			phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+secondRecordID.String()+"/attach-blob", map[string]any{
				"object_blob_id":   objectBlobID.String(),
				"base_row_version": 1,
				"client_txn_id":    "txn-phase5-i-04-quarantine-attach-blocked",
			}, authOptions(login)...),
			"evidence_quarantined",
		)

		pendingID := uuid.New()
		insertExpiredPendingBlob(t, harness, incidentID, adminID, pendingID, "phase5/i-04/pending-quarantine/"+pendingID.String(), time.Now().UTC().Add(2*time.Hour))
		if _, err := store.QuarantineBlob(context.Background(), adminID, pendingID, "admin_quarantine", "req-phase5-i-04-pending-quarantine", time.Now().UTC()); !errors.Is(err, evidence.ErrIllegalBlobTransition) {
			t.Fatalf("pending quarantine got %v want ErrIllegalBlobTransition", err)
		}
	})

	t.Run("active content uses observed media state for preview policy", func(t *testing.T) {
		for _, active := range []struct {
			name        string
			contentType string
			filename    string
		}{
			{name: "html", contentType: "text/html", filename: "pretend-image.png"},
			{name: "svg", contentType: "image/svg+xml", filename: "pretend-raster.png"},
		} {
			t.Run(active.name, func(t *testing.T) {
				harness := phase4test.StartServer(t, "phase5-i-04-active-"+active.name)
				login, adminID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
				incident := phase4test.CreateIncident(t, harness.Server, login, map[string]any{
					"client_txn_id": "txn-phase5-i-04-active-" + active.name + "-incident",
					"incident_key":  "phase5-i-04-active-" + active.name,
					"title":         "Phase 5 active content " + active.name,
				})
				incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
				recordID := uuid.New()
				seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
				payload := []byte("<script>window.__cartulary_phase5_active_content = true</script>")
				attachData := attachUploadedBlobWithHints(t, harness, login, incidentID, recordID, payload, active.filename, "image/png", active.contentType, "txn-phase5-i-04-active-"+active.name+"-blob", "txn-phase5-i-04-active-"+active.name+"-attach")
				objectBlobID := phase4test.MustUUID(t, attachData["object_blob_id"].(string))
				requireObservedContentType(t, harness, objectBlobID, active.contentType)

				requireEvidenceAccessUnavailableReason(t,
					phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/preview-handle", map[string]any{}, authOptions(login)...),
					"unsupported_preview",
				)
				download := issueEvidenceHandle(t, harness, login, recordID, "download-handle")
				if got := string(redeemHandle(t, harness.Server.HTTP.URL+download["href"].(string), login)); got != string(payload) {
					t.Fatalf("download active content got %q want %q", got, string(payload))
				}
			})
		}
	})
}

func requireCreateExpiry(t testing.TB, data map[string]any, field string, want time.Time) {
	t.Helper()
	gotRaw, ok := data[field].(string)
	if !ok {
		t.Fatalf("%s got %T in %#v", field, data[field], data)
	}
	got := mustParseTime(t, gotRaw)
	if !got.Equal(want.UTC()) {
		t.Fatalf("%s got %s want %s", field, got, want.UTC())
	}
	if field == "target_expires_at" {
		target, ok := data["upload_target"].(map[string]any)
		if !ok {
			t.Fatalf("upload_target got %T", data["upload_target"])
		}
		if target["expires_at"] != gotRaw {
			t.Fatalf("upload_target.expires_at got %#v want %q", target["expires_at"], gotRaw)
		}
	}
}

func mustParseTime(t testing.TB, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		t.Fatalf("parse time %q: %v", value, err)
	}
	return parsed.UTC()
}

func extendSessionForClockJump(t testing.TB, harness *phase4test.ServerHarness, userID any, expiresAt time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE user_sessions
   SET idle_expires_at = $2,
       absolute_expires_at = $2,
       session_expires_at = $2,
       updated_at = now()
 WHERE user_id = $1
   AND revoked_at IS NULL
`, userID, expiresAt.UTC()); err != nil {
		t.Fatalf("extend test session for clock jump: %v", err)
	}
}

func requirePhase5HTTPWorkbookCreate(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/rows", body, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func requireTimelineEvidenceProjection(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, wantCount int, wantHasEvidence bool) {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.timeline.v1/query", map[string]any{}, authOptions(login)...)
	data := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
	for _, raw := range data["rows"].([]any) {
		row := raw.(map[string]any)
		if row["record_id"] != recordID.String() {
			continue
		}
		cells := row["cells"].(map[string]any)
		if got := int(cells["timeline.evidence_count"].(map[string]any)["value"].(float64)); got != wantCount {
			t.Fatalf("timeline.evidence_count got %d want %d in row %#v", got, wantCount, row)
		}
		if got := cells["timeline.has_evidence"].(map[string]any)["value"].(bool); got != wantHasEvidence {
			t.Fatalf("timeline.has_evidence got %v want %v in row %#v", got, wantHasEvidence, row)
		}
		return
	}
	t.Fatalf("timeline row %s not found in query %#v", recordID, data["rows"])
}

func countEvidenceRevisions(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence revisions: %v", err)
	}
	return count
}

func countEvidenceBlobLinks(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM evidence WHERE record_id = $1 AND object_blob_id IS NOT NULL`, recordID).Scan(&count); err != nil {
		t.Fatalf("count evidence blob links: %v", err)
	}
	return count
}

func loginLocalUserNoMFA(t testing.TB, harness *phase4test.ServerHarness, username string, password string) phase4test.LoginResult {
	t.Helper()
	resp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == authn.SessionCookieName {
			sessionCookie = cookie
		}
		if cookie.Name == authn.CSRFCookieName {
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("login did not set session and csrf cookies: %#v", resp.Cookies())
	}
	return phase4test.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func updateRecordDeletedState(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, deleted bool) {
	t.Helper()
	deletedAt := any(nil)
	if deleted {
		deletedAt = time.Now().UTC()
	}
	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = CASE WHEN $2::timestamptz IS NULL THEN NULL ELSE created_by_user_id END,
       row_version = row_version + 1,
       updated_at = now()
 WHERE record_id = $1
`, recordID, deletedAt); err != nil {
		t.Fatalf("update record deleted state: %v", err)
	}
}

func structuredTableText(t testing.TB, harness *phase4test.ServerHarness, tables ...string) string {
	t.Helper()
	var builder strings.Builder
	for _, table := range tables {
		query := fmt.Sprintf(`SELECT COALESCE(string_agg(to_jsonb(t)::text, E'\n'), '') FROM %s AS t`, quoteIdent(table))
		var text string
		if err := harness.DB.QueryRowContext(context.Background(), query).Scan(&text); err != nil {
			t.Fatalf("dump structured table %s: %v", table, err)
		}
		builder.WriteString(table)
		builder.WriteByte('\n')
		builder.WriteString(text)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func insertFailedCleanupBlob(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint, observed_size, observed_content_type,
    observed_sha256_hex, target_expires_at, pending_expires_at, finalized_at,
    terminal_reason, failed_at, cleanup_due_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'failed',
    27, 'cleanup.txt', 'text/plain', 27, 'text/plain',
    '0000000000000000000000000000000000000000000000000000000000000000',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours', NULL,
    'pending_timeout', $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '3 hours', $5::timestamptz - interval '2 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert failed cleanup blob: %v", err)
	}
}

func insertExpiredPendingBlob(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID, actorID uuid.UUID, objectBlobID uuid.UUID, storageKey string, now time.Time) {
	t.Helper()
	if _, err := harness.DB.ExecContext(context.Background(), `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, filename_hint, content_type_hint,
    target_expires_at, pending_expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, 'pending',
    10, 'expired.txt', 'text/plain',
    $5::timestamptz - interval '2 hours', $5::timestamptz - interval '1 hour',
    $5::timestamptz - interval '25 hours', $5::timestamptz - interval '25 hours'
)
`, objectBlobID, incidentID, actorID, storageKey, now.UTC()); err != nil {
		t.Fatalf("insert expired pending blob: %v", err)
	}
}

func requireCleanedFailedBlobMetadata(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &cleaned); err != nil {
		t.Fatalf("load cleaned blob metadata: %v", err)
	}
	if uploadState != "failed" || !cleaned {
		t.Fatalf("cleaned failed blob metadata got state=%s cleaned=%v want failed/true", uploadState, cleaned)
	}
}

func requireExpiredPendingFailed(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) {
	t.Helper()
	var uploadState string
	var terminalReason string
	var cleanupDue bool
	var cleaned bool
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT upload_state, terminal_reason, cleanup_due_at IS NOT NULL, cleaned_up_at IS NOT NULL
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&uploadState, &terminalReason, &cleanupDue, &cleaned); err != nil {
		t.Fatalf("load expired pending blob: %v", err)
	}
	if uploadState != "failed" || terminalReason != "pending_timeout" || !cleanupDue || cleaned {
		t.Fatalf("expired pending blob got state=%s reason=%s cleanup_due=%v cleaned=%v", uploadState, terminalReason, cleanupDue, cleaned)
	}
}

func countEvidenceRowsForBlob(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID) int {
	t.Helper()
	var count int
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT count(*)
  FROM evidence
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&count); err != nil {
		t.Fatalf("count evidence rows for blob: %v", err)
	}
	return count
}

func requireEvidenceStates(t testing.TB, harness *phase4test.ServerHarness, recordID uuid.UUID, wantLifecycle string, wantUpload string) {
	t.Helper()
	var lifecycle string
	var upload string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT e.lifecycle_state, COALESCE(b.upload_state, e.upload_state)
  FROM evidence e
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID).Scan(&lifecycle, &upload); err != nil {
		t.Fatalf("load evidence states: %v", err)
	}
	if lifecycle != wantLifecycle || upload != wantUpload {
		t.Fatalf("evidence states got lifecycle=%s upload=%s want %s/%s", lifecycle, upload, wantLifecycle, wantUpload)
	}
}

func requireChangeSetSource(t testing.TB, harness *phase4test.ServerHarness, changeSetID uuid.UUID, wantSource string) {
	t.Helper()
	var source string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT source
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&source); err != nil {
		t.Fatalf("load change set source: %v", err)
	}
	if source != wantSource {
		t.Fatalf("change set source got %s want %s", source, wantSource)
	}
}

func requireObservedContentType(t testing.TB, harness *phase4test.ServerHarness, objectBlobID uuid.UUID, want string) {
	t.Helper()
	var got string
	if err := harness.DB.QueryRowContext(context.Background(), `
SELECT observed_content_type
  FROM object_blobs
 WHERE object_blob_id = $1
`, objectBlobID).Scan(&got); err != nil {
		t.Fatalf("load observed content type: %v", err)
	}
	if got != want {
		t.Fatalf("observed content type got %q want %q", got, want)
	}
}

func attachUploadedBlobWithHints(t *testing.T, harness *phase4test.ServerHarness, login phase4test.LoginResult, incidentID uuid.UUID, recordID uuid.UUID, payload []byte, filename string, hintContentType string, uploadContentType string, createTxn string, attachTxn string) map[string]any {
	t.Helper()
	createResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
		"incident_id":       incidentID.String(),
		"client_txn_id":     createTxn,
		"byte_size":         len(payload),
		"filename_hint":     filename,
		"content_type_hint": hintContentType,
		"sha256_hex":        fmt.Sprintf("%x", sha256Sum(payload)),
	}, authOptions(login)...)
	createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	putObject(t, createData["upload_target"].(map[string]any)["href"].(string), payload, uploadContentType)
	attachResp := phase4test.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob", map[string]any{
		"object_blob_id":   createData["object_blob_id"],
		"base_row_version": 1,
		"client_txn_id":    attachTxn,
	}, authOptions(login)...)
	return httptestx.RequireSuccessEnvelope(t, attachResp, http.StatusOK)["data"].(map[string]any)
}
