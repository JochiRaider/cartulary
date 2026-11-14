package timeline

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type InvalidCollectionActionError struct{ FieldKey string }

func (e *InvalidCollectionActionError) Error() string { return "timeline: invalid collection action" }

type SameFieldConflictError struct {
	Conflict SameFieldConflict
}

func (e *SameFieldConflictError) Error() string {
	return "timeline: same field conflict"
}

type OptionalConflictValue struct {
	Present bool
	Value   any
}

type SameFieldConflict struct {
	ConflictToken           string
	RecordID                uuid.UUID
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	ClientValue             any
	ServerValue             any
	BaseValue               OptionalConflictValue
	ServerUpdatedBy         uuid.UUID
	ServerUpdatedAt         time.Time
	SuggestedMergedValue    OptionalConflictValue
}

func (value SameFieldConflict) PublicValue() map[string]any {
	payload := map[string]any{
		"conflict_token":            value.ConflictToken,
		"record_id":                 value.RecordID.String(),
		"field_key":                 value.FieldKey,
		"conflict_resolution_class": value.ConflictResolutionClass,
		"base_row_version":          value.BaseRowVersion,
		"current_row_version":       value.CurrentRowVersion,
		"client_value":              value.ClientValue,
		"server_value":              value.ServerValue,
		"server_updated_by":         value.ServerUpdatedBy.String(),
		"server_updated_at":         valuecodec.Timestamp(value.ServerUpdatedAt),
	}
	if value.BaseValue.Present {
		payload["base_value"] = value.BaseValue.Value
	}
	if value.SuggestedMergedValue.Present {
		payload["suggested_merged_value"] = value.SuggestedMergedValue.Value
	}
	return payload
}

type patchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]patchChangedField
}

type patchChangedField struct {
	FieldKey        string
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

func newRowVersionConflict(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) *RowVersionConflictError {
	return &RowVersionConflictError{
		RecordID:          recordID,
		BaseRowVersion:    baseRowVersion,
		CurrentRowVersion: currentRowVersion,
	}
}

func overlappingPatchChange(changes []PatchChange, changedFields map[string]patchChangedField) (PatchChange, patchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, patchChangedField{}, false
}

func (s *store) buildSameFieldConflictTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, current workbookprojection.DerivedRecord, baseRowVersion int64, requestHash []byte, window patchConflictWindow, change PatchChange, changed patchChangedField) (*SameFieldConflictError, error) {
	baseValue, ok := rowCellValue(window.BaseRow, change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	serverValue, ok := rowCellValue(buildRow(current), change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	if isTimelineAttachedEvidenceCollection(change.FieldKey) && change.ActionPayload != nil {
		targets := make([]uuid.UUID, 0, len(change.ActionPayload.Actions))
		for _, action := range change.ActionPayload.Actions {
			if action.Op != "add_record_ref" {
				continue
			}
			if action.LinkedRecordID == nil {
				return nil, &InvalidCollectionActionError{FieldKey: change.FieldKey}
			}
			targets = append(targets, *action.LinkedRecordID)
		}
		if err := s.evidenceStore.ValidateTimelineAttachmentsTx(ctx, tx, current.IncidentID, targets); err != nil {
			return nil, err
		}
	}
	if isTimelineMentionCollection(change.FieldKey) && change.ActionPayload != nil {
		for _, action := range change.ActionPayload.Actions {
			if action.Op != "add_resolved_ref" && action.Op != "resolve_item" {
				continue
			}
			if action.ResolvedRecord == nil {
				return nil, &InvalidCollectionActionError{FieldKey: change.FieldKey}
			}
			err := s.entityStore.InspectResolvedTargetTx(ctx, tx, current.IncidentID, collectionEntityType(change.FieldKey), *action.ResolvedRecord)
			if errors.Is(err, ErrResolvedRecordNotFound) {
				return nil, &InvalidCollectionActionError{FieldKey: change.FieldKey}
			}
			if err != nil {
				return nil, err
			}
		}
	}
	var mentionFacts []MentionSourceFact
	if isTimelineMentionCollection(change.FieldKey) && change.ActionPayload != nil {
		refs := make([]uuid.UUID, 0)
		for _, action := range change.ActionPayload.Actions {
			switch action.Op {
			case "resolve_item", "dismiss_item", "revert_to_unresolved":
				id, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return nil, &InvalidCollectionActionError{FieldKey: change.FieldKey}
				}
				refs = append(refs, id)
			}
		}
		if len(refs) > 0 {
			facts, err := s.mentionStore.LoadCollectionSourceFactsTx(ctx, tx, recordID, change.FieldKey)
			if err != nil {
				return nil, err
			}
			mentionFacts = facts
			known := make(map[uuid.UUID]bool, len(facts))
			for _, fact := range facts {
				known[fact.MentionID] = true
			}
			for _, ref := range refs {
				if !known[ref] {
					return nil, &InvalidCollectionActionError{FieldKey: change.FieldKey}
				}
			}
		}
	}
	var clientValue any
	var err error
	if isTimelineMentionCollection(change.FieldKey) && change.ActionPayload != nil {
		clientValue, err = s.applyMentionConflictActions(recordID, change.FieldKey, baseValue, change.ActionPayload, requestHash, mentionFacts)
		if err != nil && !errors.Is(err, errInvalidMentionConflictBase) {
			return nil, err
		}
	} else {
		clientValue, err = patchClientConflictValue(recordID, change, baseValue, requestHash)
	}
	if err != nil {
		var invalid *InvalidCollectionActionError
		if errors.As(err, &invalid) {
			return nil, err
		}
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}

	field, _ := viewschema.LookupField(TimelineViewSchemaID, change.FieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	token, err := s.conflictToken(recordID, change.FieldKey, baseRowVersion, current.RowVersion, requestHash)
	if err != nil {
		return nil, err
	}
	conflict := SameFieldConflict{
		ConflictToken:           token,
		RecordID:                recordID,
		FieldKey:                change.FieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       current.RowVersion,
		ClientValue:             clientValue,
		ServerValue:             serverValue,
		BaseValue:               OptionalConflictValue{Present: true, Value: baseValue},
		ServerUpdatedBy:         changed.ServerUpdatedBy,
		ServerUpdatedAt:         changed.ServerUpdatedAt.UTC(),
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict.SuggestedMergedValue = OptionalConflictValue{Present: true, Value: suggested}
		}
	}
	return &SameFieldConflictError{Conflict: conflict}, nil
}

func (s *store) conflictToken(recordID uuid.UUID, fieldKey string, baseRowVersion int64, currentRowVersion int64, requestHash []byte) (string, error) {
	field, _ := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	return s.conflictTokens.Issue(conflicttokens.ConflictTokenClaims{
		RouteKey:                conflictResolveRouteKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            TimelineViewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicttokens.RequestHashTokenValue(requestHash),
	})
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil, false
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func patchClientConflictValue(recordID uuid.UUID, change PatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.ActionPayload == nil {
		return change.CanonicalValue(), nil
	}
	return applyCollectionConflictActions(recordID, change.FieldKey, baseValue, change.ActionPayload, requestHash)
}

func applyCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload *CollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_tag":
			duplicate := false
			for _, item := range items {
				label, _ := item["display_text"].(string)
				normalized, valid := normalizeCollectionToken(fieldKey, label)
				if valid && normalized == action.NormalizedText {
					duplicate = true
					break
				}
			}
			if !duplicate {
				items = append(items, newClientCollectionItem(recordID, fieldKey, action, requestHash, index, false))
			}
		case "add_record_ref":
			item := newClientCollectionItem(recordID, fieldKey, action, requestHash, index, true)
			if findCollectionItem(items, item["item_ref"].(string)) == nil {
				items = append(items, item)
			}
		case "remove_record_ref", "remove_tag":
			if findCollectionItem(items, action.ItemRef) == nil {
				return nil, &InvalidCollectionActionError{FieldKey: fieldKey}
			}
			items = removeCollectionItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return valuecodec.Collection(ordered, items), nil
}

func cloneCollectionConflictValue(value any) (bool, []map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != "collection_value_v1" {
		return false, nil, false
	}
	ordered, ok := object["ordered"].(bool)
	if !ok {
		return false, nil, false
	}
	items := make([]map[string]any, 0)
	switch rawItems := object["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				return false, nil, false
			}
			items = append(items, cloneMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func newClientCollectionItem(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int, resolved bool) map[string]any {
	rawText := action.RawText
	displayText := action.RawText
	if isTimelineTagCollection(fieldKey) {
		rawText = ""
		displayText = action.RawText
	}
	item := map[string]any{
		"item_ref":     clientCollectionItemRef(fieldKey, action, requestHash, actionIndex),
		"display_text": displayText,
		"raw_text":     rawText,
	}
	if isTimelineTagCollection(fieldKey) {
		item["item_kind"] = "tag"
		tagID := clientCollectionLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		item["item_ref"] = linkRecordTagItemRef(recordID, tagID)
		item["tag_id"] = tagID.String()
		delete(item, "raw_text")
		return item
	}
	if isTimelineAttachedEvidenceCollection(fieldKey) {
		item["item_kind"] = "record_ref"
		if action.LinkedRecordID != nil {
			item["item_ref"] = linkRecordRefItemRef(*action.LinkedRecordID)
			item["linked_record_id"] = action.LinkedRecordID.String()
			item["display_text"] = action.LinkedRecordID.String()
		}
		return item
	}

	item["entity_type"] = collectionEntityType(fieldKey)
	if resolved {
		item["item_kind"] = "resolved_ref"
		if action.ResolvedRecord != nil {
			item["resolved_record_id"] = action.ResolvedRecord.String()
		}
		return item
	}
	item["item_kind"] = "unresolved_mention"
	return item
}

func clientCollectionLocalUUID(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	sum := valuecodec.CanonicalJSONSHA256(map[string]any{
		"request_hash": base64.RawURLEncoding.EncodeToString(requestHash),
		"record_id":    recordID.String(),
		"field_key":    fieldKey,
		"action_index": actionIndex,
		"op":           action.Op,
		"text":         action.NormalizedText,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, sum)
}

func clientCollectionItemRef(fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) string {
	sum := valuecodec.CanonicalJSONSHA256(map[string]any{
		"request_hash":     base64.RawURLEncoding.EncodeToString(requestHash),
		"field_key":        fieldKey,
		"action_index":     actionIndex,
		"op":               action.Op,
		"raw_text":         action.NormalizedText,
		"item_ref":         action.ItemRef,
		"linked_record_id": valuecodec.OptionalUUID(action.LinkedRecordID),
	})
	token := base64.RawURLEncoding.EncodeToString(sum)
	if len(token) > 18 {
		token = token[:18]
	}
	return "client:" + token
}

func collectionEntityType(fieldKey string) string {
	if policy, ok := timelineCollectionPolicy(fieldKey); ok && policy.ExpectedTargetType != "" {
		return policy.ExpectedTargetType
	}
	return "host"
}

func findCollectionItem(items []map[string]any, itemRef string) map[string]any {
	for _, item := range items {
		if item["item_ref"] == itemRef {
			return item
		}
	}
	return nil
}

func removeCollectionItem(items []map[string]any, itemRef string) []map[string]any {
	for index, item := range items {
		if item["item_ref"] == itemRef {
			return append(items[:index], items[index+1:]...)
		}
	}
	return items
}

func removeResolutionMetadata(item map[string]any, removeResolvedID bool) {
	if removeResolvedID {
		delete(item, "resolved_record_id")
	}
	delete(item, "resolution_method")
	delete(item, "auto_resolved")
	delete(item, "provenance")
	delete(item, "confidence")
	delete(item, "matched_alias_text")
}
