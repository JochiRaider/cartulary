package evidence_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

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
    incident_id, src_record_id, dst_record_id, link_type,
    provenance, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'supported_by', 'manual', $4, $4, now(), now())
`, incidentID, timelineRecordID, evidenceRecordID, adminID); err != nil {
		t.Fatalf("insert timeline evidence support link: %v", err)
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
	t.Skip("Phase 5 I-5-04 pending Sprint 4 quarantine boundary implementation")
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
