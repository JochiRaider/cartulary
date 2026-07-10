package conflicts

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/riskrefs"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/historyquery"
	"github.com/JochiRaider/cartulary/internal/modules/workbook/collectionpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type PatchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]PatchChangedField
}

type PatchChangedField struct {
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

type PatchChange struct {
	FieldKey   string
	Value      any
	Collection *CollectionActionPayload
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

type SameFieldConflictParams struct {
	RouteKey          string
	RecordID          uuid.UUID
	ViewSchemaID      string
	BaseRowVersion    int64
	CurrentRowVersion int64
	RequestHash       []byte
	Window            PatchConflictWindow
	Change            PatchChange
	Changed           PatchChangedField
	CurrentRow        map[string]any
	Codec             ConflictTokenCodec
}

type RevisionWindowError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RevisionWindowError) Error() string {
	return "workbook conflict revision window is unavailable"
}

type textMergeHunk struct {
	start       int
	end         int
	replacement []string
}

func BuildPatchConflictWindow(recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, currentRowVersion int64, rows []historyquery.RevisionWindowRow) (PatchConflictWindow, error) {
	window := PatchConflictWindow{ChangedFields: make(map[string]PatchChangedField)}
	for _, row := range rows {
		if row.RowVersion == baseRowVersion {
			baseRow, ok := DecodeRevisionRow(row.AfterJSON)
			if !ok {
				return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
			}
			window.BaseRow = baseRow
			continue
		}
		beforeRow, beforeOK := DecodeRevisionRow(row.BeforeJSON)
		afterRow, afterOK := DecodeRevisionRow(row.AfterJSON)
		if !beforeOK || !afterOK {
			return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
		}
		for _, fieldKey := range ChangedRevisionWritableFieldKeys(viewSchemaID, beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = PatchChangedField{
				ServerUpdatedBy: row.ActorUserID,
				ServerUpdatedAt: row.CreatedAt.UTC(),
			}
		}
	}
	if window.BaseRow == nil {
		return PatchConflictWindow{}, &RevisionWindowError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	return window, nil
}

func DecodeRevisionRow(data []byte) (map[string]any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var row map[string]any
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, false
	}
	if _, ok := row["cells"].(map[string]any); !ok {
		return nil, false
	}
	return row, true
}

func ChangedRevisionWritableFieldKeys(viewSchemaID string, beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
		if !ok || !field.Writable || isReadOnlySystemField(fieldKey) {
			continue
		}
		if !reflect.DeepEqual(beforeCells[fieldKey], afterCell) {
			changed = append(changed, fieldKey)
		}
	}
	slices.Sort(changed)
	return changed
}

func OverlappingPatchChange(changes []PatchChange, changedFields map[string]PatchChangedField) (PatchChange, PatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, PatchChangedField{}, false
}

func BuildSameFieldConflict(params SameFieldConflictParams) (map[string]any, error) {
	baseValue, ok := RowCellValue(params.Window.BaseRow, params.Change.FieldKey)
	if !ok {
		return nil, &RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	serverValue, ok := RowCellValue(params.CurrentRow, params.Change.FieldKey)
	if !ok {
		return nil, &RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	clientValue, err := PatchClientConflictValue(params.RecordID, params.Change, baseValue, params.RequestHash)
	if err != nil {
		return nil, &RevisionWindowError{RecordID: params.RecordID, BaseRowVersion: params.BaseRowVersion, CurrentRowVersion: params.CurrentRowVersion}
	}
	field, _ := viewschema.LookupField(params.ViewSchemaID, params.Change.FieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflict := map[string]any{
		"conflict_token":            conflictToken(params.RouteKey, params.RecordID, params.ViewSchemaID, params.Change.FieldKey, conflictClass, params.BaseRowVersion, params.CurrentRowVersion, params.RequestHash, params.Codec),
		"record_id":                 params.RecordID.String(),
		"field_key":                 params.Change.FieldKey,
		"conflict_resolution_class": conflictClass,
		"base_row_version":          params.BaseRowVersion,
		"current_row_version":       params.CurrentRowVersion,
		"client_value":              clientValue,
		"server_value":              serverValue,
		"server_updated_by":         params.Changed.ServerUpdatedBy.String(),
		"server_updated_at":         params.Changed.ServerUpdatedAt.UTC().Format(time.RFC3339Nano),
		"base_value":                baseValue,
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := SuggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
		}
	}
	return conflict, nil
}

func RowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, _ := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func PatchClientConflictValue(recordID uuid.UUID, change PatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.Collection == nil {
		return change.Value, nil
	}
	return ApplyCollectionConflictActions(recordID, change.FieldKey, baseValue, *change.Collection, requestHash)
}

func ApplyCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload CollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := CloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref", "add_party_ref", "add_tag", "add_risk_ref":
			items = upsertWorkbookCollectionConflictItem(items, NewClientCollectionItem(recordID, fieldKey, action, requestHash, index))
		case "remove_record_ref", "remove_party_ref", "remove_tag", "remove_risk_ref":
			items = removeWorkbookCollectionConflictItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	if !ordered {
		slices.SortFunc(items, func(left map[string]any, right map[string]any) int {
			return strings.Compare(workbookCollectionSortKey(left), workbookCollectionSortKey(right))
		})
	}
	return map[string]any{"kind": "collection_value_v1", "ordered": ordered, "items": items}, nil
}

func CloneCollectionConflictValue(value any) (bool, []map[string]any, bool) {
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
			items = append(items, cloneWorkbookMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneWorkbookMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func NewClientCollectionItem(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) map[string]any {
	switch action.Op {
	case "add_record_ref":
		linkedID := action.LinkedRecordID.String()
		targetType := expectedTargetType(fieldKey)
		if targetType == "" {
			targetType = "record"
		}
		return map[string]any{
			"item_ref":         links.RecordRefItemRef(*action.LinkedRecordID),
			"item_kind":        "record_ref",
			"display_text":     targetType + ":" + linkedID,
			"linked_record_id": linkedID,
		}
	case "add_party_ref":
		partyID := action.PartyID.String()
		return map[string]any{
			"item_ref":     links.PartyRefItemRef(*action.PartyID),
			"item_kind":    "party_ref",
			"display_text": "party:" + partyID,
			"party_id":     partyID,
		}
	case "add_tag":
		tagID := ConflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":     links.RecordTagItemRef(recordID, tagID),
			"item_kind":    "tag",
			"display_text": action.RawText,
			"tag_id":       tagID.String(),
		}
	case "add_risk_ref":
		riskRefID := ConflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":      riskrefs.RiskRefItemRef(riskRefID),
			"item_kind":     "risk_ref",
			"display_text":  action.RiskRefText,
			"risk_ref_id":   riskRefID.String(),
			"risk_ref_text": action.RiskRefText,
		}
	default:
		return map[string]any{}
	}
}

func ConflictLocalUUID(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	seed, _ := json.Marshal(map[string]any{
		"record_id":     recordID.String(),
		"field_key":     fieldKey,
		"request_hash":  base64.RawURLEncoding.EncodeToString(requestHash),
		"action_index":  actionIndex,
		"op":            action.Op,
		"risk_ref_text": action.NormalizedText,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, seed)
}

func SuggestedTextMergeValue(baseValue any, serverValue any, clientValue any) (string, bool) {
	base, ok := conflictTextForMerge(baseValue)
	if !ok {
		return "", false
	}
	server, ok := conflictTextForMerge(serverValue)
	if !ok {
		return "", false
	}
	client, ok := conflictTextForMerge(clientValue)
	if !ok {
		return "", false
	}
	baseLines := strings.Split(base, "\n")
	serverHunk := changedTextMergeHunk(baseLines, strings.Split(server, "\n"))
	clientHunk := changedTextMergeHunk(baseLines, strings.Split(client, "\n"))
	if textMergeHunksEqual(serverHunk, clientHunk) {
		return strings.Join(applyTextMergeHunks(baseLines, serverHunk), "\n"), true
	}
	if serverHunk.start == serverHunk.end && clientHunk.start == clientHunk.end && serverHunk.start == clientHunk.start {
		return "", false
	}
	if serverHunk.end <= clientHunk.start {
		return strings.Join(applyTextMergeHunks(baseLines, clientHunk, serverHunk), "\n"), true
	}
	if clientHunk.end <= serverHunk.start {
		return strings.Join(applyTextMergeHunks(baseLines, serverHunk, clientHunk), "\n"), true
	}
	return "", false
}

func conflictToken(routeKey string, recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, codec ConflictTokenCodec) string {
	return codec.Issue(ConflictTokenClaims{
		RouteKey:                routeKey,
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             RequestHashTokenValue(requestHash),
	})
}

func cloneWorkbookMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func upsertWorkbookCollectionConflictItem(items []map[string]any, item map[string]any) []map[string]any {
	itemRef, _ := item["item_ref"].(string)
	if itemRef == "" {
		return items
	}
	for index, existing := range items {
		if existing["item_ref"] == itemRef {
			items[index] = item
			return items
		}
		if item["item_kind"] == "risk_ref" && existing["item_kind"] == "risk_ref" && existing["risk_ref_text"] == item["risk_ref_text"] {
			items[index] = item
			return items
		}
	}
	return append(items, item)
}

func removeWorkbookCollectionConflictItem(items []map[string]any, itemRef string) []map[string]any {
	filtered := items[:0]
	for _, item := range items {
		if item["item_ref"] != itemRef {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func workbookCollectionSortKey(item map[string]any) string {
	for _, key := range []string{"item_kind", "display_text", "item_ref"} {
		if value, ok := item[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func conflictTextForMerge(value any) (string, bool) {
	if value == nil {
		return "", true
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text, true
}

func changedTextMergeHunk(baseLines []string, variantLines []string) textMergeHunk {
	prefix := 0
	for prefix < len(baseLines) && prefix < len(variantLines) && baseLines[prefix] == variantLines[prefix] {
		prefix++
	}
	baseSuffix := len(baseLines)
	variantSuffix := len(variantLines)
	for baseSuffix > prefix && variantSuffix > prefix && baseLines[baseSuffix-1] == variantLines[variantSuffix-1] {
		baseSuffix--
		variantSuffix--
	}
	replacement := append([]string(nil), variantLines[prefix:variantSuffix]...)
	return textMergeHunk{start: prefix, end: baseSuffix, replacement: replacement}
}

func textMergeHunksEqual(left textMergeHunk, right textMergeHunk) bool {
	if left.start != right.start || left.end != right.end || len(left.replacement) != len(right.replacement) {
		return false
	}
	for index := range left.replacement {
		if left.replacement[index] != right.replacement[index] {
			return false
		}
	}
	return true
}

func applyTextMergeHunks(baseLines []string, hunks ...textMergeHunk) []string {
	result := append([]string(nil), baseLines...)
	for _, hunk := range hunks {
		next := make([]string, 0, len(result)-hunk.end+hunk.start+len(hunk.replacement))
		next = append(next, result[:hunk.start]...)
		next = append(next, hunk.replacement...)
		next = append(next, result[hunk.end:]...)
		result = next
	}
	return result
}

func expectedTargetType(fieldKey string) string {
	policy, ok := collectionpolicy.Lookup(fieldKey)
	if !ok {
		return ""
	}
	return policy.ExpectedTargetType
}

func isReadOnlySystemField(fieldKey string) bool {
	switch fieldKey {
	case "record_id", "row_version", "version_id", "updated_at", "created_at", "created_by_user_id", "updated_by_user_id":
		return true
	default:
		return false
	}
}
