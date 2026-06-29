package incidents_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestSupportPhase2_MembershipCreateReplayReturnsOriginalAndDivergentConflict(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-membership-replay")

	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	targetUserID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase2-membership-replay@example.test", "Replay Target", "ReplayTarget1!", false, false, true)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-replay-incident",
		"incident_key":  "IR-MREPLAY",
		"title":         "Membership Replay",
	})
	incidentID := incident["incident_id"].(string)

	firstCreate := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         " phase2-membership-replay@example.test ",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	firstBody := httptestx.RequireSuccessEnvelope(t, firstCreate, http.StatusCreated)["data"].(map[string]any)
	if firstBody["user_id"] != targetUserID || firstBody["role"] != "viewer" {
		t.Fatalf("unexpected created membership payload: %#v", firstBody)
	}

	replay := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         "phase2-membership-replay@example.test",
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayBody := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(firstBody, replayBody) {
		t.Fatalf("expected replayed membership payload to match original: first=%#v replay=%#v", firstBody, replayBody)
	}

	phase2test.RequireErrorContract(t, "client_txn_conflict", http.StatusConflict)
	divergentReplay := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         "phase2-membership-replay@example.test",
			"role":          "reviewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergentReplay, http.StatusConflict, "client_txn_conflict")

	phase2test.RequireErrorContract(t, "membership_exists_use_patch", http.StatusConflict)
	existingMembership := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-exists",
			"user_id":       targetUserID,
			"role":          "reviewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, existingMembership, http.StatusConflict, "membership_exists_use_patch")

	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, targetUserID); got != 1 {
		t.Fatalf("membership create replay must not duplicate rows, got %d", got)
	}
	if got := phase2test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'incident.memberships.create'
   AND actor_user_id::text = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, adminID, incidentID, "txn-membership-replay"); got != 1 {
		t.Fatalf("membership create replay must not duplicate idempotency rows, got %d", got)
	}
}

func TestSupportPhase2_IncidentPatchWritesAuditBeforeAfter(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-incident-audit")

	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-incident-audit-create",
		"incident_key":  "IR-IAUDIT",
		"title":         "Incident Audit",
	})
	incidentID := incident["incident_id"].(string)

	patchResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	events := phase2test.LookupOwnerMutations(
		t,
		harness.DB,
		phase2test.MutationSelector{IncidentID: incidentID},
		phase2test.MutationOwnerIncidentResource,
	)
	event := phase2test.RequireOwnerMutationEvent(
		t,
		events,
		phase2test.MutationOwnerIncidentResource,
		"incident_updated",
		adminID,
		"",
	)
	if event.EventSource != "incidents" || event.ClientTxnID != "" || event.RequestID == "" {
		t.Fatalf("unexpected incident_updated attribution: %#v", event)
	}
	if event.Before["incident_version"] != float64(1) || event.After["incident_version"] != float64(2) {
		t.Fatalf("unexpected incident version audit payload: before=%#v after=%#v", event.Before, event.After)
	}
	if event.Before["tlp"] != nil || event.After["tlp"] != "TLP:AMBER" {
		t.Fatalf("unexpected incident tlp audit payload: before=%#v after=%#v", event.Before, event.After)
	}
	if event.Before["current_phase"] != nil || event.After["current_phase"] != "containment" {
		t.Fatalf("unexpected incident current_phase audit payload: before=%#v after=%#v", event.Before, event.After)
	}
}

func TestSupportPhase2_MembershipMutationsWriteAuditBeforeAfter(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-membership-audit")

	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	targetUserID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase2-membership-audit@example.test", "Audit Target", "AuditTarget1!", false, false, true)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-audit-incident",
		"incident_key":  "IR-MAUDIT",
		"title":         "Membership Audit",
	})
	incidentID := incident["incident_id"].(string)

	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-audit-create",
			"user_id":       targetUserID,
			"role":          "viewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)

	patchResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createBody["membership_version"],
			"role":                    "reviewer",
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)

	deleteResp := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": patchBody["membership_version"],
		},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireStatus(t, deleteResp, http.StatusNoContent)

	events := phase2test.LookupOwnerMutations(
		t,
		harness.DB,
		phase2test.MutationSelector{IncidentID: incidentID},
		phase2test.MutationOwnerIncidentMembership,
	)
	created := phase2test.RequireOwnerMutationEvent(
		t,
		events,
		phase2test.MutationOwnerIncidentMembership,
		"incident_membership_created",
		adminID,
		targetUserID,
	)
	if created.ClientTxnID != "txn-membership-audit-create" {
		t.Fatalf("unexpected incident_membership_created attribution: %#v", created)
	}
	if created.Before["user_id"] != nil || created.After["role"] != "viewer" || created.After["user_id"] != targetUserID {
		t.Fatalf("unexpected incident_membership_created payload: before=%#v after=%#v", created.Before, created.After)
	}

	updated := phase2test.RequireOwnerMutationEvent(
		t,
		events,
		phase2test.MutationOwnerIncidentMembership,
		"incident_membership_updated",
		adminID,
		targetUserID,
	)
	if updated.ClientTxnID != "" || updated.RequestID == "" {
		t.Fatalf("unexpected incident_membership_updated attribution: %#v", updated)
	}
	if updated.Before["user_id"] != targetUserID || updated.After["user_id"] != targetUserID || updated.Before["role"] != "viewer" || updated.After["role"] != "reviewer" {
		t.Fatalf("unexpected incident_membership_updated payload: before=%#v after=%#v", updated.Before, updated.After)
	}

	deleted := phase2test.RequireOwnerMutationEvent(
		t,
		events,
		phase2test.MutationOwnerIncidentMembership,
		"incident_membership_deleted",
		adminID,
		targetUserID,
	)
	if deleted.ClientTxnID != "" || deleted.RequestID == "" {
		t.Fatalf("unexpected incident_membership_deleted attribution: %#v", deleted)
	}
	if deleted.Before["user_id"] != targetUserID || deleted.Before["role"] != "reviewer" || deleted.After["deleted"] != true {
		t.Fatalf("unexpected incident_membership_deleted payload: before=%#v after=%#v", deleted.Before, deleted.After)
	}
}
