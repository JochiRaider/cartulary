package evidence_test

// Evidence access-handle lifecycle contracts.
import (
	"context"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"net/http"
	"testing"
)

func TestHandleRedeemInvalidatesOnCurrentStateLoss_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "evidence_lifecycle-handle-redeem-invalidates")
	login, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence_lifecycle-i-03-incident",
		"incident_key":  "evidence_lifecycle-i-03",
		"title":         "Evidence handle invalidation",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	activeLogin := login
	activeAdminID := adminID
	activeIncidentID := incidentID

	t.Run("membership loss and wrong session hide the handle", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("membership body"), "membership.txt", "text/plain", "txn-evidence_lifecycle-i-03-membership-blob", "txn-evidence_lifecycle-i-03-membership-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		if _, err := harness.DB.ExecContext(context.Background(), `
DELETE FROM incident_memberships
 WHERE incident_id = $1
   AND user_id = $2
`, activeIncidentID, activeAdminID); err != nil {
			t.Fatalf("remove incident membership: %v", err)
		}
		httptestx.RequireErrorEnvelope(t,
			appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(activeLogin.SessionCookie)),
			http.StatusNotFound,
			"handle_not_found_or_revoked",
		)

		incidentstoretest.SeedMembership(t, harness.DB, activeIncidentID, activeAdminID, "Bootstrap Admin", "admin", activeAdminID)
		otherUser := authflowtest.SeedLocalUserRecord(t, harness.DB, "evidence_lifecycle-i-03-other@example.test", "EvidenceLifecycle Other", "EvidenceLifecycleOther1!", false, false, true)
		incidentstoretest.SeedMembership(t, harness.DB, activeIncidentID, otherUser.ID, "EvidenceLifecycle Other", "admin", activeAdminID)
		otherLogin := loginLocalUserNoMFA(t, harness, "evidence_lifecycle-i-03-other@example.test", "EvidenceLifecycleOther1!")
		httptestx.RequireErrorEnvelope(t,
			appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(otherLogin.SessionCookie)),
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
			replacementID := linkSeededBlob(t, harness, activeIncidentID, activeAdminID, recordID, "available", "available", "evidence_lifecycle/i-03/replacement-"+recordID.String())
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
			deleteBlobObject(t, harness, blobStorageKey(t, harness, objectBlobID))
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
				attachData := attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("evidence_lifecycle invalidation"), "invalidate.txt", "text/plain", "txn-"+recordID.String()+"-blob", "txn-"+recordID.String()+"-attach")
				objectBlobID := appsupport.MustUUID(t, attachData["object_blob_id"].(string))
				handle := issueEvidenceHandle(t, harness, activeLogin, recordID, endpoint)
				scenario.mutate(recordID, objectBlobID)
				requireEvidenceAccessUnavailableReason(t,
					appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(activeLogin.SessionCookie)),
					scenario.reasonCode,
				)
			})
		}
	}

	t.Run("logout revokes issuing session before lookup", func(t *testing.T) {
		recordID := uuid.New()
		seedEvidenceRecord(t, harness, activeIncidentID, activeAdminID, recordID)
		attachUploadedBlobWithMetadata(t, harness, activeLogin, activeIncidentID, recordID, []byte("logout body"), "logout.txt", "text/plain", "txn-evidence_lifecycle-i-03-logout-blob", "txn-evidence_lifecycle-i-03-logout-attach")
		handle := issueEvidenceHandle(t, harness, activeLogin, recordID, "preview-handle")
		logout := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/logout", map[string]any{}, authOptions(activeLogin)...)
		httptestx.RequireSuccessEnvelope(t, logout, http.StatusOK)
		httptestx.RequireErrorEnvelope(t,
			appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+handle["href"].(string), nil, appsupport.WithCookies(activeLogin.SessionCookie)),
			http.StatusUnauthorized,
			"session_required",
		)
	})
}
