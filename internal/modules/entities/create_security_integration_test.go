package entities_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
)

func TestEntityCreateAuthAndCSRFFailBeforeMalformedBody_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-entity-create-auth-csrf-order")
	adminLogin, _ := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-entity_linking-entity-create-auth-order-incident",
		"incident_key":  "IR-AUTH-CSRF-ORDER",
		"title":         "Entity create auth csrf ordering",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	type entityCreateFailureCounts struct {
		Records        int
		ChangeSets     int
		MutationRows   int
		HostProjection int
	}
	counts := func() entityCreateFailureCounts {
		return entityCreateFailureCounts{
			Records:        appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID),
			ChangeSets:     appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows:   appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, incidentID),
			HostProjection: appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE incident_id = $1`, incidentID),
		}
	}
	before := counts()
	url := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + viewtest.HostsViewSchemaID + "/rows"

	unauthenticated := doEntitiesRawJSON(t, http.MethodPost, url, "{")
	appsupport.RequireErrorBody(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	appsupport.RequireErrorBody(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		withHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	appsupport.RequireErrorBody(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	for _, body := range []string{
		`null`,
		`{"client_txn_id":"first","client_txn_id":"second"}`,
		`{"client_txn_id":"txn-trailing","host.display_name":"Host"} {}`,
	} {
		response := doEntitiesRawJSON(
			t,
			http.MethodPost,
			url,
			body,
			withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, response, http.StatusBadRequest, "invalid_mutation_payload")
	}

	if after := counts(); after != before {
		t.Fatalf("auth/csrf failures must not mutate entity state: before=%#v after=%#v", before, after)
	}
	incidentwstest.ExpectNoSocketMessage(t, socket)
}

// entity-resolution / REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
