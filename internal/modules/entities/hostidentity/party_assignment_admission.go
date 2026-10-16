package hostidentity

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/google/uuid"
)

func isPartyAssignmentField(field string) bool { return links.IsPartyAssignmentField(field) }

func decodePartyAssignmentActions(raw json.RawMessage) ([]CollectionAction, bool) {
	var payload map[string]json.RawMessage
	if json.Unmarshal(raw, &payload) != nil || len(payload) != 2 {
		return nil, false
	}
	var kind string
	var actions []map[string]json.RawMessage
	if json.Unmarshal(payload["kind"], &kind) != nil || kind != "collection_actions_v1" || json.Unmarshal(payload["actions"], &actions) != nil || len(actions) < 1 || len(actions) > maxCollectionActions {
		return nil, false
	}
	result := make([]CollectionAction, 0, len(actions))
	for _, rawAction := range actions {
		if len(rawAction) != 2 {
			return nil, false
		}
		var op string
		if json.Unmarshal(rawAction["op"], &op) != nil {
			return nil, false
		}
		switch op {
		case "add_party_ref":
			var rawID string
			if json.Unmarshal(rawAction["party_id"], &rawID) != nil {
				return nil, false
			}
			id, err := uuid.Parse(rawID)
			if err != nil || id == uuid.Nil || id.String() != rawID {
				return nil, false
			}
			result = append(result, CollectionAction{Op: op, PartyID: id})
		case "remove_party_ref":
			var ref string
			if json.Unmarshal(rawAction["item_ref"], &ref) != nil {
				return nil, false
			}
			id, err := links.ParsePartyRefItemRef(ref)
			if err != nil || id == uuid.Nil {
				return nil, false
			}
			result = append(result, CollectionAction{Op: op, ItemRef: ref})
		default:
			return nil, false
		}
	}
	return result, true
}
