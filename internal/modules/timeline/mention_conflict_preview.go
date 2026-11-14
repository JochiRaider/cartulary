package timeline

import (
	"errors"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/google/uuid"
)

var errInvalidMentionConflictBase = errors.New("timeline: invalid Mention conflict base")

func (s *store) applyMentionConflictActions(recordID uuid.UUID, field string, base any, payload *CollectionActionPayload, requestHash []byte, facts []MentionSourceFact) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(base)
	if !ok {
		return nil, errInvalidMentionConflictBase
	}
	states := make(map[string]string, len(items))
	for _, item := range items {
		ref, ok := item["item_ref"].(string)
		if !ok {
			return nil, errInvalidMentionConflictBase
		}
		switch item["item_kind"] {
		case "unresolved_mention":
			states[ref] = "unresolved"
		case "resolved_ref":
			states[ref] = "resolved"
		default:
			return nil, errInvalidMentionConflictBase
		}
	}

	byRef := make(map[string]MentionSourceFact, len(facts))
	for _, fact := range facts {
		byRef["entity_mention:"+fact.MentionID.String()] = fact
	}
	supplemented := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "resolve_item", "dismiss_item", "revert_to_unresolved":
			if _, present := states[action.ItemRef]; present {
				continue
			}
			fact, present := byRef[action.ItemRef]
			if !present || fact.ResolutionStatus != "dismissed" {
				return nil, &InvalidCollectionActionError{FieldKey: field}
			}
			items = append(items, map[string]any{
				"entity_mention_id": fact.MentionID.String(), "item_ref": action.ItemRef,
				"item_kind": "unresolved_mention", "entity_type": fact.EntityType,
				"raw_text": fact.RawText, "display_text": fact.RawText, "mention_row_version": fact.RowVersion,
			})
			states[action.ItemRef] = "dismissed"
			supplemented = true
		}
	}
	if supplemented {
		for _, item := range items {
			if _, present := byRef[item["item_ref"].(string)]; !present {
				return nil, errInvalidMentionConflictBase
			}
		}
		slices.SortStableFunc(items, func(a, b map[string]any) int {
			left, right := byRef[a["item_ref"].(string)], byRef[b["item_ref"].(string)]
			if left.Ordinal < right.Ordinal {
				return -1
			}
			if left.Ordinal > right.Ordinal {
				return 1
			}
			return strings.Compare(left.MentionID.String(), right.MentionID.String())
		})
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_token", "add_resolved_ref":
			resolved := action.Op == "add_resolved_ref"
			item := newClientCollectionItem(recordID, field, action, requestHash, index, resolved)
			items = append(items, item)
			state := "unresolved"
			if resolved {
				state = "resolved"
			}
			states[item["item_ref"].(string)] = state
		case "resolve_item", "dismiss_item", "revert_to_unresolved":
			item := findCollectionItem(items, action.ItemRef)
			if item == nil {
				return nil, &InvalidCollectionActionError{FieldKey: field}
			}
			to, err := s.mentionStore.TransitionStatus(field, states[action.ItemRef], action.Op)
			if err != nil {
				return nil, err
			}
			states[action.ItemRef] = to
			if to == "resolved" {
				item["item_kind"] = "resolved_ref"
				if action.ResolvedRecord != nil {
					item["resolved_record_id"] = action.ResolvedRecord.String()
				}
				removeResolutionMetadata(item, false)
			} else {
				item["item_kind"] = "unresolved_mention"
				removeResolutionMetadata(item, true)
			}
		default:
			return nil, &InvalidCollectionActionError{FieldKey: field}
		}
	}
	visible := items[:0]
	for _, item := range items {
		if states[item["item_ref"].(string)] != "dismissed" {
			visible = append(visible, item)
		}
	}
	return valuecodec.Collection(ordered, visible), nil
}
