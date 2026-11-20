package tasksdecisions

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionports"
	"github.com/JochiRaider/cartulary/internal/platform/collectionvalue"
)

// errInvalidCollectionConflictBase reports a reconstructed base value that
// cannot carry the requested collection actions. Like an unavailable window it
// fails closed as row_version_conflict; an action the admission contract
// excludes is an invariant failure instead.
var errInvalidCollectionConflictBase = errors.New("invalid base collection value")

// adaptRevisionWindowError classifies only the typed unavailable-window
// condition, which fails closed as row_version_conflict. Every other failure,
// including a database, field-catalog or conflict-token error, stays
// unclassified and surfaces as internal_error (REQ-01-706).
func adaptRevisionWindowError(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64, err error) error {
	if err == nil {
		return nil
	}
	var windowErr *conflicttokens.RevisionWindowError
	if errors.As(err, &windowErr) {
		return &RowVersionConflictError{RecordID: windowErr.RecordID, BaseRowVersion: windowErr.BaseRowVersion, CurrentRowVersion: windowErr.CurrentRowVersion}
	}
	return fmt.Errorf("record %s revision window %d..%d: %w", recordID, baseRowVersion, currentRowVersion, err)
}

type sameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            conflicttokens.PatchConflictWindow
	Change            PatchChange
	Changed           conflicttokens.PatchChangedField
	CurrentRow        map[string]any
	FieldDescriptors  conflicttokens.FieldDescriptorSet
	Codec             conflicttokens.ConflictTokenCodec
}

func overlappingPatchChange(changes []PatchChange, changedFields map[string]conflicttokens.PatchChangedField) (PatchChange, conflicttokens.PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, conflicttokens.PatchChangedField{}, false
}

func buildSameFieldConflict(params sameFieldConflictParams) (SameFieldConflict, error) {
	baseValue, ok := rowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := rowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	// Stored conflict facts may predate REQ-01-067 label order; re-serve them.
	baseValue, serverValue = collectionvalue.Canonicalize(baseValue), collectionvalue.Canonicalize(serverValue)
	clientValue, err := patchClientConflictValue(params.Change, baseValue)
	if errors.Is(err, errInvalidCollectionConflictBase) {
		return SameFieldConflict{}, &conflicttokens.RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	if err != nil {
		return SameFieldConflict{}, err
	}
	conflictClass := params.FieldDescriptors.ConflictResolutionClass(params.Change.FieldKey)
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflictToken, err := conflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec)
	if err != nil {
		return SameFieldConflict{}, err
	}
	conflict := SameFieldConflict{
		ConflictToken: conflictToken, RecordID: params.RecordID, FieldKey: params.Change.FieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion,
		ClientValue: clientValue, ServerValue: serverValue,
		BaseValue:       OptionalConflictValue{Present: true, Value: baseValue},
		ServerUpdatedBy: params.Changed.ServerUpdatedBy,
		ServerUpdatedAt: params.Changed.ServerUpdatedAt.UTC(),
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := conflicttokens.SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict.SuggestedMergedValue = OptionalConflictValue{Present: true, Value: suggested}
		}
	}
	return conflict, nil
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, _ := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func patchClientConflictValue(change PatchChange, baseValue any) (any, error) {
	if change.Collection == nil {
		return change.CanonicalValue, nil
	}
	return applyCollectionConflictActions(change.FieldKey, baseValue, *change.Collection)
}

func applyCollectionConflictActions(fieldKey string, baseValue any, payload CollectionActionPayload) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("%w for %s", errInvalidCollectionConflictBase, fieldKey)
	}
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return nil, fmt.Errorf("missing linked record")
			}
			items = upsertCollectionConflictItem(items, newRecordRefConflictItem(*action.LinkedRecordID))
		case "remove_record_ref":
			items = removeCollectionConflictItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	return collectionvalue.New(ordered, items), nil
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

// newRecordRefConflictItem carries a provisional label that
// labelConflictTx replaces with the REQ-01-333a label.
func newRecordRefConflictItem(linkedRecordID uuid.UUID) map[string]any {
	return map[string]any{
		"item_ref":         links.RecordRefItemRef(linkedRecordID),
		"item_kind":        "record_ref",
		"display_text":     collectionvalue.UnavailableRecordLabel,
		"linked_record_id": linkedRecordID.String(),
	}
}

// labelConflictTx labels the record_ref items of every conflict value with
// REQ-01-333a labels read in one lookup inside the conflict transaction.
// Records not live in the incident read UnavailableRecordLabel.
func labelConflictTx(ctx context.Context, tx pgx.Tx, labels taskdecisionprojection.MutationRows, incidentID uuid.UUID, conflict *SameFieldConflict) error {
	ids := collectionvalue.LinkedRecordIDs(conflict.BaseValue.Value, conflict.ServerValue, conflict.ClientValue)
	if len(ids) == 0 {
		return nil
	}
	byID, err := labels.RecordSummaryLabelsTx(ctx, tx, incidentID, ids)
	if err != nil {
		return err
	}
	conflict.BaseValue.Value = collectionvalue.WithLinkedRecordLabels(conflict.BaseValue.Value, byID)
	conflict.ServerValue = collectionvalue.WithLinkedRecordLabels(conflict.ServerValue, byID)
	conflict.ClientValue = collectionvalue.WithLinkedRecordLabels(conflict.ClientValue, byID)
	return nil
}

func upsertCollectionConflictItem(items []map[string]any, item map[string]any) []map[string]any {
	ref, _ := item["item_ref"].(string)
	for index, existing := range items {
		if existingRef, _ := existing["item_ref"].(string); existingRef == ref {
			items[index] = item
			return items
		}
	}
	return append(items, item)
}

func removeCollectionConflictItem(items []map[string]any, itemRef string) []map[string]any {
	result := items[:0]
	for _, item := range items {
		if existingRef, _ := item["item_ref"].(string); existingRef != itemRef {
			result = append(result, item)
		}
	}
	return result
}

func cloneMap(input map[string]any) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}

func conflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec conflicttokens.ConflictTokenCodec) (string, error) {
	return codec.Issue(conflicttokens.ConflictTokenClaims{
		RouteKey:                routeKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             conflicttokens.RequestHashTokenValue(requestHash),
	})
}
