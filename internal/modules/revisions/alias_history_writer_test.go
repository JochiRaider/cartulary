package revisions_test

import (
	"database/sql"
	"encoding/json"
	"maps"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func expectedAliasHistory(id, incidentID, recordID, actorID uuid.UUID, entityType, text string, createdAt time.Time) map[string]any {
	return map[string]any{"entity_alias_id": id.String(), "incident_id": incidentID.String(), "record_id": recordID.String(), "entity_type": entityType,
		"raw_text": text, "normalized_text": text, "classification": "suggestion_only", "created_by_user_id": actorID.String(), "created_at": createdAt.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano), "deleted_at": nil}
}

func requireAliasHistory(t testing.TB, db *sql.DB, changeSetID, operation string, id uuid.UUID, before, after map[string]any) {
	t.Helper()
	encode := func(value map[string]any) any {
		if value == nil {
			return nil
		}
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return payload
	}
	var count int
	var exact bool
	if err := db.QueryRow(`SELECT count(*),coalesce(bool_and(operation_kind=$3 AND coalesce(before_value,'null'::jsonb)=coalesce($4::jsonb,'null'::jsonb) AND after_value=$5::jsonb),false)
FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='entity_alias' AND target_id=$2`, changeSetID, "entity_alias:"+id.String(), operation, encode(before), encode(after)).Scan(&count, &exact); err != nil {
		t.Fatal(err)
	}
	if count != 1 || !exact {
		var gotBefore, gotAfter sql.NullString
		err := db.QueryRow(`SELECT before_value::text,after_value::text FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='entity_alias' AND target_id=$2`, changeSetID, "entity_alias:"+id.String()).Scan(&gotBefore, &gotAfter)
		t.Fatalf("complete physical Alias history: count=%d exact=%v operation=%s before=%s after=%s wantBefore=%v wantAfter=%s diagnosticError=%v", count, exact, operation, gotBefore.String, gotAfter.String, encode(before), encode(after), err)
	}
	var sourceExact bool
	if err := db.QueryRow(`SELECT (to_jsonb(a)-'created_at'-'deleted_at')=($2::jsonb-'created_at'-'deleted_at') AND created_at=($2::jsonb->>'created_at')::timestamptz AND deleted_at IS NOT DISTINCT FROM ($2::jsonb->>'deleted_at')::timestamptz FROM entity_aliases a WHERE entity_alias_id=$1`, id, encode(after)).Scan(&sourceExact); err != nil || !sourceExact {
		t.Fatalf("Alias source facts differ: exact=%v error=%v", sourceExact, err)
	}
}

func exerciseAliasWriterLifecycle(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incidentID, actorID uuid.UUID, clockBase time.Time) {
	for _, kind := range []string{"host", "identity"} {
		t.Run(kind, func(t *testing.T) {
			schema := "cartulary.view.hosts.v1"
			if kind == "identity" {
				schema = "cartulary.view.identities.v1"
			}
			createAt := clockBase.Add(10*time.Minute + 789*time.Nanosecond)
			httptestx.SetClockFixed(t, h.Server, createAt)
			body := map[string]any{"client_txn_id": "alias-history-create-" + kind, kind + ".display_name": "Alias History " + kind, kind + ".aliases": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_alias", "alias_text": "  Initial   Alias  "}}}}
			if kind == "host" {
				body["host.hostname"] = "alias-history-host"
			} else {
				body["identity.upn"] = "alias-history@example.test"
			}
			data := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+schema+"/rows", body, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusCreated)
			recordID := uuid.MustParse(data["row"].(map[string]any)["record_id"].(string))
			lookup := func(text string) uuid.UUID {
				var id uuid.UUID
				if err := h.DB.QueryRow(`SELECT entity_alias_id FROM entity_aliases WHERE record_id=$1 AND raw_text=$2`, recordID, text).Scan(&id); err != nil {
					t.Fatal(err)
				}
				return id
			}
			changeSet := func(txn string) string {
				return stringScalar(t, h.DB, `SELECT change_set_id::text FROM change_sets WHERE incident_id=$1 AND client_txn_id=$2`, incidentID, txn)
			}
			initialID := lookup("Initial   Alias")
			initial := expectedAliasHistory(initialID, incidentID, recordID, actorID, kind, "Initial   Alias", createAt)
			requireAliasHistory(t, h.DB, changeSet("alias-history-create-"+kind), "create", initialID, nil, initial)
			patch := func(version int, txn string, action map[string]any) {
				httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, h.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), map[string]any{"view_schema_id": schema, "base_row_version": version, "client_txn_id": txn, "changes": []map[string]any{{"field_key": kind + ".aliases", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{action}}}}}, httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusOK)
			}
			addAt := createAt.Add(time.Minute)
			httptestx.SetClockFixed(t, h.Server, addAt)
			patch(1, "alias-history-add-"+kind, map[string]any{"op": "add_alias", "alias_text": "Added Alias"})
			addedID := lookup("Added Alias")
			requireAliasHistory(t, h.DB, changeSet("alias-history-add-"+kind), "create", addedID, nil, expectedAliasHistory(addedID, incidentID, recordID, actorID, kind, "Added Alias", addAt))
			deleteAt := addAt.Add(time.Minute)
			httptestx.SetClockFixed(t, h.Server, deleteAt)
			patch(2, "alias-history-delete-"+kind, map[string]any{"op": "remove_alias", "item_ref": "entity_alias:" + initialID.String()})
			deleted := maps.Clone(initial)
			deleted["deleted_at"] = deleteAt.UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)
			deleteChangeSet := changeSet("alias-history-delete-" + kind)
			requireAliasHistory(t, h.DB, deleteChangeSet, "delete", initialID, initial, deleted)
			httptestx.SetClockFixed(t, h.Server, deleteAt.Add(time.Minute))
			requireNonRowCurrentness(t, h.DB, incidentID, recordID, initialID, "entity_aliases", "entity_alias_id", map[string]map[string]any{
				"text":     {"raw_text": "Changed Alias", "normalized_text": "Changed Alias"},
				"creation": {"created_at": createAt.Add(-time.Second).UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)},
				"deletion": {"deleted_at": deleteAt.Add(time.Second).UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano)},
			}, func() *http.Response {
				return rollbackRecord(t, h, login, recordID, map[string]any{"base_row_version": 3, "client_txn_id": "alias-history-restore-" + kind, "target": map[string]any{"kind": "change_set", "change_set_id": deleteChangeSet}})
			})
			rollback := httptestx.RequireSuccessData(t, rollbackRecord(t, h, login, recordID, map[string]any{"base_row_version": 3, "client_txn_id": "alias-history-restore-" + kind, "target": map[string]any{"kind": "change_set", "change_set_id": deleteChangeSet}}), http.StatusOK)
			requireAliasHistory(t, h.DB, rollback["rollback_change_set_id"].(string), "rollback", initialID, deleted, initial)
		})
	}
}
