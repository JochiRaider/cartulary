package revisions_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseAliasCommittedOrder(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID, actorID uuid.UUID, base time.Time) {
	for _, step := range []time.Duration{0, -time.Minute} {
		t.Run(step.String(), func(t *testing.T) {
			prefix := "committed-order-" + step.String()
			now := base.Add(12 * time.Minute)
			httptestx.SetClockFixed(t, h.Server, now.Add(2*step))
			data := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.hosts.v1/rows", map[string]any{"client_txn_id": prefix + "-create", "host.display_name": prefix, "host.hostname": prefix}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusCreated)
			recordID := uuid.MustParse(data["row"].(map[string]any)["record_id"].(string))
			patch := func(version int, txn string, action map[string]any) {
				httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), map[string]any{"view_schema_id": "cartulary.view.hosts.v1", "base_row_version": version, "client_txn_id": txn, "changes": []map[string]any{{"field_key": "host.aliases", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{action}}}}}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusOK)
			}
			patch(1, prefix+"-add", map[string]any{"op": "add_alias", "alias_text": "Recurring Alias"})
			var aliasID uuid.UUID
			if err := h.DB.QueryRow(`SELECT entity_alias_id FROM entity_aliases WHERE record_id=$1`, recordID).Scan(&aliasID); err != nil {
				t.Fatal(err)
			}
			initial := expectedAliasHistory(aliasID, incidentID, recordID, actorID, "host", "Recurring Alias", now.Add(2*step))
			addID := stringScalar(t, h.DB, `SELECT change_set_id::text FROM change_sets WHERE incident_id=$1 AND client_txn_id=$2`, incidentID, prefix+"-add")
			httptestx.SetClockFixed(t, h.Server, now)
			patch(2, prefix+"-delete", map[string]any{"op": "remove_alias", "item_ref": "entity_alias:" + aliasID.String()})
			deleteID := stringScalar(t, h.DB, `SELECT change_set_id::text FROM change_sets WHERE incident_id=$1 AND client_txn_id=$2`, incidentID, prefix+"-delete")
			httptestx.SetClockFixed(t, h.Server, now.Add(step))
			restored := httptestx.RequireSuccessData(t, rollbackRecord(t, h, login, recordID, map[string]any{"base_row_version": 3, "client_txn_id": prefix + "-restore", "target": map[string]any{"kind": "change_set", "change_set_id": deleteID}}), http.StatusOK)
			var active bool
			if err := h.DB.QueryRow(`SELECT deleted_at IS NULL AND raw_text='Recurring Alias' AND created_at=$2 FROM entity_aliases WHERE entity_alias_id=$1`, aliasID, now.Add(2*step)).Scan(&active); err != nil || !active {
				t.Fatalf("restored complete Alias: %v %v", active, err)
			}
			// The restored state equals the old add's after-image, but later committed
			// mutations still make that add stale. Capture all affected durable state.
			read := func() string {
				var value string
				if err := h.DB.QueryRow(`SELECT jsonb_build_object('record',(SELECT to_jsonb(r) FROM records r WHERE record_id=$1),'alias',(SELECT to_jsonb(a) FROM entity_aliases a WHERE entity_alias_id=$2),'history',(SELECT jsonb_agg(to_jsonb(c) ORDER BY change_set_id) FROM change_sets c WHERE incident_id=$3),'mutations',(SELECT jsonb_agg(to_jsonb(m) ORDER BY m.change_set_id,m.sequence_no) FROM change_set_mutations m JOIN change_sets c USING(change_set_id) WHERE c.incident_id=$3),'revisions',(SELECT jsonb_agg(to_jsonb(v) ORDER BY revision_id) FROM record_revisions v WHERE record_id=$1),'projection',(SELECT to_jsonb(p) FROM host_grid_projection p WHERE record_id=$1),'allocator',(SELECT last_order FROM revision_history_orders WHERE incident_id=$3),'idempotency',(SELECT jsonb_agg(to_jsonb(i) ORDER BY client_txn_id) FROM route_idempotency i WHERE scope_key=$4))::text`, recordID, aliasID, incidentID, recordID.String()).Scan(&value); err != nil {
					t.Fatal(err)
				}
				return value
			}
			deletedAt := now.UTC().Format(time.RFC3339Nano)
			deleted := expectedAliasHistory(aliasID, incidentID, recordID, actorID, "host", "Recurring Alias", now.Add(2*step))
			deleted["deleted_at"] = deletedAt
			requireAliasHistory(t, h.DB, restored["rollback_change_set_id"].(string), "rollback", aliasID, deleted, initial)
			before := read()
			intents := collaborationsupport.CountIntents(t, h.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()})
			requireRollbackReasonCode(t, rollbackRecord(t, h, login, recordID, map[string]any{"base_row_version": 4, "client_txn_id": prefix + "-old-add", "target": map[string]any{"kind": "change_set", "change_set_id": addID}}), "stale_target")
			if read() != before || collaborationsupport.CountIntents(t, h.DB, collaborationsupport.IntentSelector{IncidentID: incidentID.String()}) != intents {
				t.Fatal("stale history rollback changed durable state or publication")
			}
		})
	}
}
