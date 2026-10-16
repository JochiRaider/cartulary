package hostidentity

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// Called by the existing entity admission entrypoint and catalog row.
func testPartyAssignmentAdmission(t *testing.T) {
	const party = "10000000-0000-4000-8000-000000000001"
	for _, surface := range []struct{ family, view string }{{"host", "hosts"}, {"identity", "identities"}} {
		view := "cartulary.view." + surface.view + ".v1"
		for _, kind := range []string{"subject", "responsible", "observer"} {
			field := surface.family + "." + kind + "_party_ids"
			t.Run(field, func(t *testing.T) {
				request := func(payload string) string {
					return `{"view_schema_id":"` + view + `","client_txn_id":"party-admission","base_row_version":1,"changes":[{"field_key":"` + field + `","action_payload":` + payload + `}]}`
				}
				valid := `{"kind":"collection_actions_v1","actions":[{"op":"add_party_ref","party_id":"` + party + `"}]}`
				admitted, failure := DecodePatchRequest(strings.NewReader(request(valid)))
				if failure != nil {
					t.Fatal(failure)
				}
				other, failure := DecodePatchRequest(strings.NewReader(request(strings.ReplaceAll(valid, party, "10000000-0000-4000-8000-000000000002"))))
				if failure != nil {
					t.Fatal(failure)
				}
				if bytes.Equal(PatchRequestHash(admitted), PatchRequestHash(other)) {
					t.Fatal("party identity missing from idempotency preimage")
				}
				for _, payload := range []string{
					`null`, `[]`, `{"kind":"collection_actions_v1","actions":[]}`,
					`{"kind":"collection_actions_v1","actions":[],"extra":true}`,
					strings.Replace(valid, `"party_id":"`+party+`"`, `"party_id":null`, 1),
					strings.ReplaceAll(valid, party, "00000000-0000-0000-0000-000000000000"),
					strings.ReplaceAll(valid, party, "10000000000040008000000000000001"),
					strings.Replace(valid, `"party_id":`, `"extra":true,"party_id":`, 1),
					strings.Replace(valid, `"op":"add_party_ref"`, `"op":"add_alias"`, 1),
					`{"kind":"collection_actions_v1","actions":[{"op":"remove_party_ref","item_ref":"entity_alias:` + party + `"}]}`,
					`{"kind":"collection_actions_v1","actions":[{"op":"add_party_ref","party_id":"` + party + `","party_id":"` + party + `"}]}`,
				} {
					if _, failure := DecodePatchRequest(strings.NewReader(request(payload))); failure == nil {
						t.Fatalf("accepted invalid payload %s", payload)
					}
				}
				scalar := strings.Replace(request(valid), `"action_payload":`+valid, `"value":"`+party+`"`, 1)
				if _, failure := DecodePatchRequest(strings.NewReader(scalar)); failure == nil {
					t.Fatal("accepted scalar replacement")
				}
				create := map[string]any{"client_txn_id": "assignment-create", surface.family + ".display_name": "Entity", field: json.RawMessage(valid)}
				raw, _ := json.Marshal(create)
				created, failure := DecodeCreateRequest(view, bytes.NewReader(raw))
				if failure != nil || len(created.PartyAdds[field]) != 1 {
					t.Fatal("assignment create rejected", failure)
				}
				create[field] = json.RawMessage(strings.ReplaceAll(valid, party, "10000000-0000-4000-8000-000000000002"))
				raw, _ = json.Marshal(create)
				different, failure := DecodeCreateRequest(view, bytes.NewReader(raw))
				if failure != nil || bytes.Equal(CreateRequestHash(view, created), CreateRequestHash(view, different)) {
					t.Fatal("create hash omits party identity", failure)
				}
				delete(create, surface.family+".display_name")
				raw, _ = json.Marshal(create)
				if _, failure := DecodeCreateRequest(view, bytes.NewReader(raw)); failure == nil {
					t.Fatal("assignment alone satisfied entity seed")
				}
				create[surface.family+".display_name"] = "Entity"
				create[field] = json.RawMessage(`{"kind":"collection_actions_v1","actions":[{"op":"remove_party_ref","item_ref":"party_ref:` + party + `"}]}`)
				raw, _ = json.Marshal(create)
				if _, failure := DecodeCreateRequest(view, bytes.NewReader(raw)); failure == nil {
					t.Fatal("create accepted removal")
				}
			})
		}
	}
}
