package entities_test

import (
	"net/http"
	"strings"
	"testing"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestEntityCreateIdempotencyIsActorScoped(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-entity-create-actor-scope")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	editor := authflowtest.SeedLocalUserRecord(t, harness.DB, "entity_linking-entity-scope-editor@example.test", "Entity Scope Editor", "EntityScopeEditor1!", false, false, true)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-entity_linking-entity-scope-incident",
		"incident_key":  "IR-E-ACTOR-SCOPE",
		"title":         "Entity actor-scoped idempotency",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	incidentstoretest.SeedMembership(t, harness.DB, incidentID, editor.ID, editor.DisplayName, "editor", adminUserID)
	editorLogin := loginLocalUser(t, harness.Server, editor.Email, "EntityScopeEditor1!")

	cases := []struct {
		name       string
		routeKey   string
		viewSchema string
		payload    func(label string) map[string]any
	}{
		{
			name:       "hosts",
			routeKey:   "entities.hosts.rows.create",
			viewSchema: viewtest.HostsViewSchemaID,
			payload: func(label string) map[string]any {
				return map[string]any{
					"client_txn_id":     "txn-entity_linking-shared-host-create",
					"host.display_name": "Actor scoped host " + label,
					"host.hostname":     "ACTOR-SCOPE-" + label,
				}
			},
		},
		{
			name:       "identities",
			routeKey:   "entities.identities.rows.create",
			viewSchema: viewtest.IdentitiesViewSchemaID,
			payload: func(label string) map[string]any {
				return map[string]any{
					"client_txn_id":         "txn-entity_linking-shared-identity-create",
					"identity.display_name": "Actor Scoped " + label,
					"identity.email":        "actor-scope-" + label + "@example.test",
				}
			},
		},
		{
			name:       "indicators",
			routeKey:   "indicators.rows.create",
			viewSchema: viewtest.IndicatorsViewSchemaID,
			payload: func(label string) map[string]any {
				value := "198.51.100.10"
				if label == "editor" {
					value = "198.51.100.11"
				}
				return map[string]any{
					"client_txn_id":              "txn-entity_linking-shared-indicator-create",
					"indicator.indicator_type":   "ipv4_addr",
					"indicator.value_kind":       "atomic",
					"indicator.display_value":    value,
					"indicator.normalized_value": value,
					"indicator.defanged_value":   strings.ReplaceAll(value, ".", "[.]"),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adminPayload := tc.payload("admin")
			editorPayload := tc.payload("editor")
			adminCreate := createEntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, adminLogin, adminPayload, http.StatusCreated)
			editorCreate := createEntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, editorLogin, editorPayload, http.StatusCreated)
			adminRecordID := adminCreate["row"].(map[string]any)["record_id"].(string)
			editorRecordID := editorCreate["row"].(map[string]any)["record_id"].(string)
			if adminRecordID == editorRecordID {
				t.Fatalf("cross-actor %s create must not replay another actor's row, got %s", tc.name, adminRecordID)
			}

			adminReplay := createEntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, adminLogin, adminPayload, http.StatusOK)
			if adminReplay["change_set_id"] != adminCreate["change_set_id"] {
				t.Fatalf("admin %s replay returned wrong payload: got %#v want %#v", tc.name, adminReplay, adminCreate)
			}
			editorReplay := createEntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, editorLogin, editorPayload, http.StatusOK)
			if editorReplay["change_set_id"] != editorCreate["change_set_id"] {
				t.Fatalf("editor %s replay returned wrong payload: got %#v want %#v", tc.name, editorReplay, editorCreate)
			}

			clientTxnID := adminPayload["client_txn_id"].(string)
			scopeKey := incidentID.String() + ":" + tc.viewSchema
			if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text IN ($2, $3)
   AND scope_key = $4
   AND client_txn_id = $5
`, tc.routeKey, adminUserID.String(), editor.ID.String(), scopeKey, clientTxnID); got != 2 {
				t.Fatalf("expected two actor-scoped %s idempotency rows, got %d", tc.name, got)
			}
			if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(DISTINCT actor_user_id)
  FROM route_idempotency
 WHERE route_key = $1
   AND scope_key = $2
   AND client_txn_id = $3
   AND actor_user_id::text IN ($4, $5)
`, tc.routeKey, scopeKey, clientTxnID, adminUserID.String(), editor.ID.String()); got != 2 {
				t.Fatalf("expected both actors represented for %s idempotency, got %d", tc.name, got)
			}
		})
	}
}
