package workbook_test

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func exerciseHostIdentityCollectionConflicts(t *testing.T) {
	h, login, actor, incident := ConflictFixture(t, "entity-collection-conflicts", "IR-ENTITY-COLLECTION")
	parties := map[string]string{}
	for _, name := range []string{"Base", "Saved", "Local", "Merged"} {
		row := requireWorkbookCreate(t, h, login, incident, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "collection-party-" + name, "party.display_name": name, "party.party_kind": "person"})["row"].(map[string]any)
		parties[name] = row["record_id"].(string)
	}
	foreign := appsupport.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "entity-collection-foreign", "incident_key": "IR-ENTITY-COLLECTION-FOREIGN", "title": "Foreign Party"})
	foreignID := appsupport.MustUUID(t, foreign["incident_id"].(string))
	foreignParty := requireWorkbookCreate(t, h, login, foreignID, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "collection-foreign-party", "party.display_name": "Foreign", "party.party_kind": "person"})["row"].(map[string]any)["record_id"].(string)
	deletedParty := requireWorkbookCreate(t, h, login, incident, "cartulary.view.parties.v1", map[string]any{"client_txn_id": "collection-deleted-party", "party.display_name": "Deleted", "party.party_kind": "person"})["row"].(map[string]any)["record_id"].(string)
	httptestx.RequireSuccessData(t, doWorkbookJSON(t, h, login, http.MethodDelete, uuid.Nil, "", appsupport.MustUUID(t, deletedParty), map[string]any{"client_txn_id": "collection-delete-party", "base_row_version": 1}), http.StatusOK)
	for _, family := range []string{"host", "identity"} {
		view := "cartulary.view.hosts.v1"
		identifier := family + ".hostname"
		if family == "identity" {
			view = "cartulary.view.identities.v1"
			identifier = family + ".email"
		}
		for _, suffix := range []string{"aliases", "subject_party_ids", "responsible_party_ids", "observer_party_ids"} {
			field := family + "." + suffix
			alias := suffix == "aliases"
			add := func(name string) map[string]any {
				if alias {
					return map[string]any{"op": "add_alias", "alias_text": name}
				}
				return map[string]any{"op": "add_party_ref", "party_id": parties[name]}
			}
			remove := func(ref string) map[string]any {
				op := "remove_party_ref"
				if alias {
					op = "remove_alias"
				}
				return map[string]any{"op": op, "item_ref": ref}
			}
			for _, mode := range []string{"use_unsaved", "merged_value", "disjoint"} {
				t.Run(field+"/"+mode, func(t *testing.T) {
					prefix := strings.ReplaceAll(field, ".", "-") + "-" + mode
					created := requireWorkbookCreate(t, h, login, incident, view, map[string]any{"client_txn_id": prefix + "-create", family + ".display_name": "Original", field: collectionActions(add("Base"))})["row"].(map[string]any)
					id := appsupport.MustUUID(t, created["record_id"].(string))
					initial := requireEntityConflictCollection(t, cellMapValue(t, created, field), alias, parties, []string{"Base"}, nil)
					patch := func(txn string, base int, actions map[string]any, extra ...map[string]any) map[string]any {
						changes := append([]map[string]any{{"field_key": field, "action_payload": actions}}, extra...)
						return map[string]any{"client_txn_id": prefix + txn, "view_schema_id": view, "base_row_version": base, "changes": changes}
					}
					scalar := func(key string, value any) map[string]any { return map[string]any{"field_key": key, "value": value} }
					if mode == "disjoint" {
						requireWorkbookPatch(t, h, login, id, map[string]any{"client_txn_id": prefix + "-saved", "view_schema_id": view, "base_row_version": 1, "changes": []map[string]any{scalar(family+".display_name", "Saved scalar")}})
						body := patch("-rebased", 1, collectionActions(remove(initial["Base"]), add("Local")), scalar(identifier, prefix+"@example.test"))
						rebased := requireWorkbookPatch(t, h, login, id, body)
						row := rebased["row"].(map[string]any)
						requireEntityConflictCollection(t, cellMapValue(t, row, field), alias, parties, []string{"Local"}, nil)
						requireCellValue(t, row, family+".display_name", "Saved scalar")
						requireCellValue(t, row, identifier, prefix+"@example.test")
						if row["row_version"] != float64(3) {
							t.Fatal("collection rebase did not advance protected version 2")
						}
						var coherent bool
						if err := h.DB.QueryRowContext(t.Context(), `SELECT before_version_id=$3 AND after_version_id=$4 AND (before_value #>> '{record,row_version}')::bigint=2 AND (after_value #>> '{record,row_version}')::bigint=3 FROM change_set_mutations WHERE change_set_id=$1 AND target_id=$2`, rebased["change_set_id"], id.String(), fmt.Sprintf("%s:%s:2", family, id), fmt.Sprintf("%s:%s:3", family, id)).Scan(&coherent); err != nil {
							t.Fatal(err)
						}
						if !coherent {
							t.Fatal("collection rebase history has the rejected base version")
						}
						before := snapshotWorkbookState(t, h, login, incident, view)
						if replay := requireWorkbookPatch(t, h, login, id, body); !reflect.DeepEqual(replay, rebased) {
							t.Fatal("collection rebase receipt changed")
						}
						requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
						return
					}
					saved := requireWorkbookPatch(t, h, login, id, patch("-saved", 1, collectionActions(add("Saved"))))
					savedRefs := requireEntityConflictCollection(t, cellMapValue(t, saved["row"].(map[string]any), field), alias, parties, []string{"Base", "Saved"}, initial)
					duplicate := add("Local")
					if alias {
						duplicate = add(" local ")
					}
					actions := collectionActions(remove(initial["Base"]), add("Base"), add("Local"), duplicate)
					body := patch("-local", 1, actions, scalar(family+".display_name", "Requested scalar"))
					before := snapshotWorkbookState(t, h, login, incident, view)
					conflictResponse := func(body map[string]any) map[string]any {
						response := doWorkbookJSON(t, h, login, http.MethodPatch, uuid.Nil, "", id, body)
						failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "same_field_conflict")
						requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
						return failure["error"].(map[string]any)["conflict"].(map[string]any)
					}
					conflict := conflictResponse(body)
					if conflict["field_key"] != field || conflict["record_id"] != id.String() || conflict["conflict_resolution_class"] != "collection_review" || conflict["base_row_version"] != float64(1) || conflict["current_row_version"] != float64(2) || conflict["server_updated_by"] != actor.String() {
						t.Fatalf("collection conflict metadata=%#v", conflict)
					}
					var attributed bool
					if err := h.DB.QueryRowContext(t.Context(), `SELECT created_at=$2::timestamptz AND actor_user_id=$3 FROM change_sets WHERE change_set_id=$1`, saved["change_set_id"], conflict["server_updated_at"], actor).Scan(&attributed); err != nil {
						t.Fatal(err)
					}
					if !attributed {
						t.Fatal("collection conflict attribution differs from saved change set")
					}
					requireEntityConflictCollection(t, conflict["base_value"], alias, parties, []string{"Base"}, initial)
					requireEntityConflictCollection(t, conflict["server_value"], alias, parties, []string{"Base", "Saved"}, savedRefs)
					clientRefs := requireEntityConflictCollection(t, conflict["client_value"], alias, parties, []string{"Base", "Local"}, nil)
					if alias && clientRefs["Base"] == initial["Base"] {
						t.Fatal("alias remove/add preview reused removed identity")
					}
					repeated := conflictResponse(body)
					if !reflect.DeepEqual(repeated["client_value"], conflict["client_value"]) {
						t.Fatal("collection preview identities are not stable for exact replay")
					}
					noEffect := conflictResponse(patch("-duplicate-only", 1, collectionActions(add("Base"))))
					requireEntityConflictCollection(t, noEffect["client_value"], alias, parties, []string{"Base"}, initial)
					unknown := "party_ref:" + parties["Merged"]
					if alias {
						unknown = "entity_alias:" + uuid.NewString()
					}
					invalid := patch("-invalid", 1, collectionActions(remove(unknown), add("Merged")))
					httptestx.RequireErrorEnvelope(t, doWorkbookJSON(t, h, login, http.MethodPatch, uuid.Nil, "", id, invalid), http.StatusBadRequest, "invalid_mutation_payload")
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
					if !alias {
						for i, target := range []string{uuid.NewString(), id.String(), foreignParty, deletedParty} {
							bad := patch(fmt.Sprintf("-invalid-target-%d", i), 1, collectionActions(map[string]any{"op": "add_party_ref", "party_id": target}))
							httptestx.RequireErrorEnvelope(t, doWorkbookJSON(t, h, login, http.MethodPatch, uuid.Nil, "", id, bad), http.StatusBadRequest, "invalid_mutation_payload")
							requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
						}
					}
					token, ok := conflict["conflict_token"].(string)
					if !ok || token == "" {
						t.Fatal("missing collection conflict token")
					}
					for _, rejection := range []struct {
						name, reason string
						actions      map[string]any
					}{
						{"no-effect", "no_effective_change", collectionActions(add("Base"))},
						{"preview-only-ref", "invalid_value", collectionActions(remove(clientRefs["Local"]))},
					} {
						body := map[string]any{"conflict_token": token, "resolution_kind": "merged_value", "client_txn_id": prefix + "-reject-" + rejection.name, "resolved_value": rejection.actions}
						failure := httptestx.RequireErrorEnvelope(t, ResolveConflictRaw(t, h, login, id, token, body), http.StatusBadRequest, "invalid_mutation_payload")
						if failure["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != rejection.reason {
							t.Fatal("collection resolution rejection changed its reason")
						}
						requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
					}
					keep := map[string]any{"conflict_token": token, "resolution_kind": "keep_saved", "client_txn_id": prefix + "-keep"}
					kept := ResolveConflict(t, h, login, id, token, keep)
					RequireNoChangeSetID(t, kept)
					requireEntityConflictCollection(t, cellMapValue(t, kept["row"].(map[string]any), field), alias, parties, []string{"Base", "Saved"}, savedRefs)
					requireWorkbookKeepSaved(t, before, snapshotWorkbookState(t, h, login, incident, view))
					before = snapshotWorkbookState(t, h, login, incident, view)
					if replay := ResolveConflict(t, h, login, id, token, keep); !reflect.DeepEqual(replay, kept) {
						t.Fatal("collection keep-saved receipt changed")
					}
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
					resolved := actions
					expected := []string{"Base", "Saved", "Local"}
					if mode == "merged_value" {
						resolved = collectionActions(remove(savedRefs["Saved"]), add("Merged"))
						expected = []string{"Base", "Merged"}
					}
					resolve := map[string]any{"conflict_token": token, "resolution_kind": mode, "client_txn_id": prefix + "-resolve", "resolved_value": resolved}
					used := ResolveConflict(t, h, login, id, token, resolve)
					usedRow := used["row"].(map[string]any)
					usedRefs := requireEntityConflictCollection(t, cellMapValue(t, usedRow, field), alias, parties, expected, nil)
					requireCellValue(t, usedRow, family+".display_name", "Original")
					if usedRow["row_version"] != float64(3) {
						t.Fatal("collection resolution version is not 3")
					}
					before = snapshotWorkbookState(t, h, login, incident, view)
					if replay := ResolveConflict(t, h, login, id, token, resolve); !reflect.DeepEqual(replay, used) {
						t.Fatal("collection resolution receipt changed")
					}
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
					stale := map[string]any{"conflict_token": token, "resolution_kind": "merged_value", "client_txn_id": prefix + "-stale", "resolved_value": collectionActions(add("Local"))}
					failure := httptestx.RequireErrorEnvelope(t, ResolveConflictRaw(t, h, login, id, token, stale), http.StatusConflict, "same_field_conflict")
					fresh := failure["error"].(map[string]any)["conflict"].(map[string]any)
					if fresh["conflict_token"] == token || fresh["base_row_version"] != float64(2) || fresh["current_row_version"] != float64(3) {
						t.Fatal("stale collection resolution did not refresh conflict versions/token")
					}
					requireEntityConflictCollection(t, fresh["base_value"], alias, parties, []string{"Base", "Saved"}, savedRefs)
					requireEntityConflictCollection(t, fresh["server_value"], alias, parties, expected, usedRefs)
					requireEntityConflictCollection(t, fresh["client_value"], alias, parties, []string{"Base", "Saved", "Local"}, savedRefs)
					requireWorkbookStateUnchanged(t, before, snapshotWorkbookState(t, h, login, incident, view))
				})
			}
		}
	}
}

// Expected values come from the request vocabulary and fixture identities. Only
// new alias IDs are server-created; validate their shape and cross-request use
// separately from exact base/server identity assertions.
func requireEntityConflictCollection(t testing.TB, value any, alias bool, parties map[string]string, names []string, known map[string]string) map[string]string {
	t.Helper()
	collection, ok := value.(map[string]any)
	if !ok || len(collection) != 3 || collection["kind"] != "collection_value_v1" || collection["ordered"] != false {
		t.Fatalf("invalid entity collection %#v", value)
	}
	items, ok := collection["items"].([]any)
	if !ok || len(items) != len(names) {
		t.Fatalf("entity collection count/type: %#v", collection)
	}
	want := map[string]bool{}
	for _, name := range names {
		want[name] = true
	}
	result := map[string]string{}
	last := ""
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok || len(item) != 4 {
			t.Fatal("invalid typed collection item")
		}
		ref, ok := item["item_ref"].(string)
		if !ok || ref <= last {
			t.Fatal("collection is not in strict item_ref order")
		}
		last = ref
		name, ok := item["display_text"].(string)
		if !ok || !want[name] || result[name] != "" {
			t.Fatalf("unexpected collection member %#v", item)
		}
		if alias {
			if item["item_kind"] != "alias" || item["alias_text"] != name || !strings.HasPrefix(ref, "entity_alias:") {
				t.Fatal("alias kind or text changed")
			}
			id, err := uuid.Parse(strings.TrimPrefix(ref, "entity_alias:"))
			if err != nil || id == uuid.Nil || ref != "entity_alias:"+id.String() {
				t.Fatal("invalid alias identity")
			}
		} else if item["item_kind"] != "party_ref" || item["party_id"] != parties[name] || ref != "party_ref:"+parties[name] {
			t.Fatal("Party reference kind or identity changed")
		}
		if expected := known[name]; expected != "" && ref != expected {
			t.Fatal("retained collection identity changed")
		}
		result[name] = ref
	}
	return result
}
