package incidents_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/mutationtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSupportPhase2_MembershipCreateReplayReturnsOriginalAndDivergentConflict(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-membership-replay")

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	targetUserID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase2-membership-replay@example.test", "Replay Target", "ReplayTarget1!", false, false, true)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-replay-incident",
		"incident_key":  "IR-MREPLAY",
		"title":         "Membership Replay",
	})
	incidentID := incident["incident_id"].(string)

	firstCreate := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         " phase2-membership-replay@example.test ",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	firstBody := httptestx.RequireSuccessEnvelope(t, firstCreate, http.StatusCreated)["data"].(map[string]any)
	if firstBody["user_id"] != targetUserID || firstBody["role"] != "viewer" {
		t.Fatalf("unexpected created membership payload: %#v", firstBody)
	}

	replay := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         "phase2-membership-replay@example.test",
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayBody := httptestx.RequireSuccessEnvelope(t, replay, http.StatusOK)["data"].(map[string]any)
	if !reflect.DeepEqual(firstBody, replayBody) {
		t.Fatalf("expected replayed membership payload to match original: first=%#v replay=%#v", firstBody, replayBody)
	}

	contracttest.RequireErrorContract(t, "client_txn_conflict", http.StatusConflict)
	divergentReplay := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-replay",
			"email":         "phase2-membership-replay@example.test",
			"role":          "reviewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, divergentReplay, http.StatusConflict, "client_txn_conflict")

	contracttest.RequireErrorContract(t, "membership_exists_use_patch", http.StatusConflict)
	existingMembership := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-exists",
			"user_id":       targetUserID,
			"role":          "reviewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, existingMembership, http.StatusConflict, "membership_exists_use_patch")

	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id::text = $1 AND user_id::text = $2`, incidentID, targetUserID); got != 1 {
		t.Fatalf("membership create replay must not duplicate rows, got %d", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `
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
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-incident-audit")

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-incident-audit-create",
		"incident_key":  "IR-IAUDIT",
		"title":         "Incident Audit",
	})
	incidentID := incident["incident_id"].(string)

	patchResp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID,
		map[string]any{
			"base_incident_version": 1,
			"tlp":                   "TLP:AMBER",
			"current_phase":         "containment",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)

	events := mutationtest.LookupOwnerMutations(
		t, mutationtest.SQLDatabase(

			harness.DB),

		mutationtest.MutationSelector{IncidentID: incidentID},
		mutationtest.MutationOwnerIncidentResource)

	event := mutationtest.RequireOwnerMutationEvent(
		t,
		events,
		mutationtest.MutationOwnerIncidentResource,
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
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase2-membership-audit")

	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	targetUserID := flowtest.SeedLocalUserFlags(t, harness.DB, "phase2-membership-audit@example.test", "Audit Target", "AuditTarget1!", false, false, true)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-audit-incident",
		"incident_key":  "IR-MAUDIT",
		"title":         "Membership Audit",
	})
	incidentID := incident["incident_id"].(string)

	createResp := httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships",
		map[string]any{
			"client_txn_id": "txn-membership-audit-create",
			"user_id":       targetUserID,
			"role":          "viewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createBody := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)

	patchResp := httptestx.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": createBody["membership_version"],
			"role":                    "reviewer",
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	patchBody := httptestx.RequireSuccessEnvelope(t, patchResp, http.StatusOK)["data"].(map[string]any)

	deleteResp := httptestx.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/memberships/"+targetUserID,
		map[string]any{
			"base_membership_version": patchBody["membership_version"],
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireStatus(t, deleteResp, http.StatusNoContent)

	events := mutationtest.LookupOwnerMutations(
		t, mutationtest.SQLDatabase(

			harness.DB),

		mutationtest.MutationSelector{IncidentID: incidentID},
		mutationtest.MutationOwnerIncidentMembership)

	created := mutationtest.RequireOwnerMutationEvent(
		t,
		events,
		mutationtest.MutationOwnerIncidentMembership,
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

	updated := mutationtest.RequireOwnerMutationEvent(
		t,
		events,
		mutationtest.MutationOwnerIncidentMembership,
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

	deleted := mutationtest.RequireOwnerMutationEvent(
		t,
		events,
		mutationtest.MutationOwnerIncidentMembership,
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
