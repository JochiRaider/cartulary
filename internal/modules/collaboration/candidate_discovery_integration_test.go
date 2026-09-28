package collaboration_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// Discovery is a read: it must not publish a durable Collaboration intent.
// The physical intent table belongs to this module, not to Entities tests.
func TestEntityCandidateDiscoveryDoesNotPublish_Integration(t *testing.T) {
	h := appsupport.StartServer(t, "collaboration-candidate-discovery")
	login, _ := appsupport.ProvisionBootstrapAdmin(t, h.Server)
	incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{
		"client_txn_id": "collaboration-candidate-incident",
		"incident_key":  "IR-COLLAB-CANDIDATES",
		"title":         "Candidate discovery publication",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	count := func() int {
		return appsupport.QueryCount(t, h.DB,
			`SELECT count(*) FROM collaboration_event_intents WHERE incident_id=$1`, incidentID)
	}
	before := count()
	base := h.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/entity-candidates"
	for _, query := range []string{"?entity_type=host", "?entity_type=identity&search=absent"} {
		response := appsupport.DoJSON(t, http.MethodGet, base+query, nil,
			appsupport.WithCookies(login.SessionCookie))
		httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	}
	if after := count(); after != before {
		t.Fatalf("discovery published collaboration events: before=%d after=%d", before, after)
	}
}
