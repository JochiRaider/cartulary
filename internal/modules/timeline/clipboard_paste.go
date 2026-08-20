package timeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ownerBatchRowPlanV1 struct {
	RowOrdinal int
	Cells      []ownerBatchCellV1
	Unmapped   []ClipboardRawImportColumn
}

type ownerBatchCellV1 struct {
	FieldKey string
	Value    string
	Change   PatchChange
}

type ClipboardRawImportColumn struct {
	SourceKind          string `json:"source_kind"`
	PasteClientTxnID    string `json:"paste_client_txn_id,omitempty"`
	ImportSessionID     string `json:"import_session_id,omitempty"`
	ImportUnitID        string `json:"import_unit_id,omitempty"`
	MappingFingerprint  string `json:"mapping_fingerprint,omitempty"`
	SourceFileKind      string `json:"source_file_kind,omitempty"`
	SourceContentSHA256 string `json:"source_content_sha256,omitempty"`
	ParserProfileID     string `json:"parser_profile_id,omitempty"`
	ParserVersion       string `json:"parser_version,omitempty"`
	LocatorKind         string `json:"locator_kind,omitempty"`
	Locator             string `json:"locator,omitempty"`
	SourceRectA1        string `json:"source_rect_a1,omitempty"`
	SourceRowOrdinal    int    `json:"source_row_ordinal"`
	SourceColumnOrdinal int    `json:"source_column_ordinal"`
	SourceHeaderText    any    `json:"source_header_text"`
	RawValue            string `json:"raw_value"`
	CellKind            string `json:"cell_kind,omitempty"`
}

func buildClipboardOwnerRows(plan tabularingest.TabularRowPlanV1) ([]ownerBatchRowPlanV1, error) {
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("validate tabular row plan: %w", err)
	}
	if plan.ViewSchemaID != TimelineViewSchemaID {
		return nil, fmt.Errorf("tabular plan view mismatch %q", plan.ViewSchemaID)
	}
	rows := make([]ownerBatchRowPlanV1, 0, len(plan.Rows))
	for _, plannedRow := range plan.Rows {
		row := ownerBatchRowPlanV1{RowOrdinal: plannedRow.RowOrdinal}
		for _, cell := range plannedRow.Cells {
			change, ok := clipboardValueToPatchChange(cell.FieldKey, cell.RawValue)
			if !ok {
				row.Unmapped = append(row.Unmapped, ClipboardRawImportColumn{
					SourceKind:          plan.SourceKind,
					PasteClientTxnID:    plan.ClientTxnID,
					MappingFingerprint:  plan.MappingFingerprint,
					SourceRowOrdinal:    plannedRow.RowOrdinal,
					SourceColumnOrdinal: cell.SourceColumnOrdinal,
					SourceHeaderText:    sourceColumnHeader(plan.SourceColumns, cell.SourceColumnOrdinal, cell.FieldKey),
					RawValue:            cell.RawValue,
				})
				continue
			}
			row.Cells = append(row.Cells, ownerBatchCellV1{
				FieldKey: cell.FieldKey,
				Value:    cell.RawValue,
				Change:   change,
			})
		}
		for _, unmapped := range plannedRow.Unmapped {
			row.Unmapped = append(row.Unmapped, ClipboardRawImportColumn{
				SourceKind:          unmapped.SourceKind,
				PasteClientTxnID:    unmapped.SourceClientTxnID,
				MappingFingerprint:  plan.MappingFingerprint,
				SourceRowOrdinal:    unmapped.SourceRowOrdinal,
				SourceColumnOrdinal: unmapped.SourceColumnOrdinal,
				SourceHeaderText:    unmapped.SourceHeaderText,
				RawValue:            unmapped.RawValue,
			})
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func buildFillDownOwnerRows(fieldKey string, rawValue string, rowCount int) ([]ownerBatchRowPlanV1, error) {
	change, ok := fillDownValueToPatchChange(fieldKey, rawValue)
	if !ok {
		return nil, fmt.Errorf("invalid fill-down value for %s", fieldKey)
	}
	rows := make([]ownerBatchRowPlanV1, 0, rowCount)
	for index := 0; index < rowCount; index++ {
		rows = append(rows, ownerBatchRowPlanV1{
			RowOrdinal: index + 1,
			Cells: []ownerBatchCellV1{{
				FieldKey: fieldKey,
				Value:    rawValue,
				Change:   change,
			}},
		})
	}
	return rows, nil
}

func buildTagAssignmentOwnerRows(tagName string, normalizedTag string, rowCount int) ([]ownerBatchRowPlanV1, error) {
	if tagName == "" || normalizedTag == "" {
		return nil, fmt.Errorf("invalid tag assignment")
	}
	payload := &CollectionActionPayload{Actions: []CollectionAction{{
		Op:             "add_tag",
		RawText:        tagName,
		NormalizedText: normalizedTag,
	}}}
	change := PatchChange{
		FieldKey:      "timeline.tags",
		ActionPayload: payload,
		CanonicalAny:  payload.CanonicalValue(),
	}
	rows := make([]ownerBatchRowPlanV1, 0, rowCount)
	for index := 0; index < rowCount; index++ {
		rows = append(rows, ownerBatchRowPlanV1{
			RowOrdinal: index + 1,
			Cells: []ownerBatchCellV1{{
				FieldKey: "timeline.tags",
				Value:    tagName,
				Change:   change,
			}},
		})
	}
	return rows, nil
}

func sourceColumnHeader(columns []tabularingest.SourceColumnV1, ordinal int, fallback string) any {
	for _, column := range columns {
		if column.Ordinal == ordinal {
			if column.HeaderText != nil {
				return column.HeaderText
			}
			break
		}
	}
	return fallback
}

func clipboardValueToPatchChange(fieldKey string, rawValue string) (PatchChange, bool) {
	field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	if !ok || !field.Writable {
		return PatchChange{}, false
	}
	if field.ConflictResolutionClass == "collection_review" {
		action, ok := clipboardCollectionAction(fieldKey, rawValue)
		if !ok {
			return PatchChange{}, false
		}
		payload := &CollectionActionPayload{Actions: []CollectionAction{action}}
		change := PatchChange{FieldKey: fieldKey, ActionPayload: payload}
		change.CanonicalAny = payload.CanonicalValue()
		return change, true
	}
	return fillDownValueToPatchChange(fieldKey, rawValue)
}

func fillDownValueToPatchChange(fieldKey string, rawValue string) (PatchChange, bool) {
	field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	if !ok || !field.Writable || field.ConflictResolutionClass == "collection_review" {
		return PatchChange{}, false
	}
	if !mutationpolicy.IsDirectWritableField(fieldKey) {
		return PatchChange{}, false
	}
	rawJSON, _ := json.Marshal(rawValue)
	value, ok := normalizeFieldTextValue(fieldKey, rawJSON)
	if !ok {
		return PatchChange{}, false
	}
	change := PatchChange{FieldKey: fieldKey, TextValue: value}
	change.CanonicalAny = change.CanonicalValue()
	return change, true
}

func clipboardCollectionAction(fieldKey string, rawValue string) (CollectionAction, bool) {
	switch fieldKey {
	case "timeline.host_refs", "timeline.identity_refs":
		normalized, ok := normalizeCollectionToken(fieldKey, rawValue)
		if !ok {
			return CollectionAction{}, false
		}
		return CollectionAction{Op: "add_token", RawText: rawValue, NormalizedText: normalized}, true
	case "timeline.tags":
		normalized, ok := normalizeCollectionToken(fieldKey, rawValue)
		if !ok {
			return CollectionAction{}, false
		}
		return CollectionAction{Op: "add_tag", RawText: strings.TrimSpace(rawValue), NormalizedText: normalized}, true
	default:
		return CollectionAction{}, false
	}
}
