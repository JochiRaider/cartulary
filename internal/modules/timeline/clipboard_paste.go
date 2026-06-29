package timeline

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	clipboardPasteRouteKey = "timeline.clipboard_paste"
	maxClipboardPasteRows  = 500
	maxClipboardPasteCols  = 64
)

var timelineV2ExactHeaderFieldKeys = []string{
	"timeline.date_entered_text",
	"timeline.analyst_text",
	"timeline.mitre_stage_text",
	"timeline.device_object_text",
	"timeline.ip_address_text",
	"timeline.activity_utc_text",
	"timeline.activity_local_text",
	"timeline.raw_activity_text",
	"timeline.activity_synopsis_text",
	"timeline.data_source_text",
}

var timelineV2ExactHeaderLabels = []string{
	"Date Entered",
	"Analyst",
	"MITRE",
	"Device/Object",
	"IP Address",
	"Activity Date (UTC)",
	"Activity Date (Local Time)",
	"RAW Activity",
	"Activity Synopsis",
	"Data Source",
}

type ClipboardPasteRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	ClipboardText string
	Format        string
	StartFieldKey string
	Columns       []string
	Targets       []ClipboardPasteTarget
	SourceKind    string
	RouteKey      string
}

type ClipboardPasteTarget struct {
	Kind           string
	RecordID       uuid.UUID
	BaseRowVersion int64
}

type ClipboardPastePlan struct {
	ClientTxnID string
	Rows        []ClipboardPasteRowPlan
}

type ClipboardPasteRowPlan struct {
	RowOrdinal int
	Cells      []clipboardPasteCell
	Unknown    []ClipboardRawImportColumn
}

type clipboardPasteCell struct {
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

type ClipboardPasteResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	IncidentID  uuid.UUID
	ChangeSetID uuid.UUID
	ClientTxnID string
	Rows        []ClipboardPasteRowResult
}

type ClipboardPasteRowResult struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
	Row              map[string]any
}

func DecodeTimelineClipboardPasteRequest(reader io.Reader) (ClipboardPasteRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return ClipboardPasteRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"view_schema_id":  {},
		"client_txn_id":   {},
		"clipboard_text":  {},
		"format":          {},
		"start_field_key": {},
		"columns":         {},
		"targets":         {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ClipboardPasteRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request ClipboardPasteRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != TimelineViewSchemaID {
		return ClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["clipboard_text"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClipboardText); err != nil || request.ClipboardText == "" {
		return ClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "invalid_value")
	}
	request.Format = "auto"
	if value, ok := raw["format"]; ok {
		if err := json.Unmarshal(value, &request.Format); err != nil {
			return ClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
		}
	}
	switch request.Format {
	case "", "auto", "tsv", "csv":
		if request.Format == "" {
			request.Format = "auto"
		}
	default:
		return ClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
	}
	if value, ok := raw["start_field_key"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.StartFieldKey); err != nil || request.StartFieldKey == "" {
		return ClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "invalid_value")
	}
	if _, ok := viewschema.LookupField(TimelineViewSchemaID, request.StartFieldKey); !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "unsupported_field_key")
	}
	if value, ok := raw["columns"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Columns); err != nil || len(request.Columns) == 0 || len(request.Columns) > maxClipboardPasteCols {
		return ClipboardPasteRequest{}, invalidMutationPayload("columns", "invalid_value")
	}
	for _, fieldKey := range request.Columns {
		if _, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey); !ok {
			return ClipboardPasteRequest{}, invalidMutationPayload("columns", "unsupported_field_key")
		}
	}
	if value, ok := raw["targets"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("targets", "missing_required_field")
	} else if err := decodeClipboardPasteTargets(value, &request.Targets); err != nil {
		return ClipboardPasteRequest{}, invalidMutationPayload("targets", err.Error())
	}
	return request, nil
}

func decodeClipboardPasteTargets(raw json.RawMessage, out *[]ClipboardPasteTarget) error {
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 || len(items) > maxClipboardPasteRows {
		return fmt.Errorf("invalid_value")
	}
	targets := make([]ClipboardPasteTarget, 0, len(items))
	for _, item := range items {
		if !objectHasOnlyFields(item, "kind") && !objectHasOnlyFields(item, "kind", "record_id", "base_row_version") {
			return fmt.Errorf("invalid_value")
		}
		var target ClipboardPasteTarget
		if err := json.Unmarshal(item["kind"], &target.Kind); err != nil {
			return fmt.Errorf("invalid_value")
		}
		switch target.Kind {
		case "create":
			if len(item) != 1 {
				return fmt.Errorf("invalid_value")
			}
		case "record":
			var rawID string
			if err := json.Unmarshal(item["record_id"], &rawID); err != nil {
				return fmt.Errorf("invalid_value")
			}
			recordID, err := uuid.Parse(rawID)
			if err != nil || recordID == uuid.Nil {
				return fmt.Errorf("invalid_value")
			}
			target.RecordID = recordID
			if err := json.Unmarshal(item["base_row_version"], &target.BaseRowVersion); err != nil || target.BaseRowVersion < 1 {
				return fmt.Errorf("invalid_base_row_version")
			}
		default:
			return fmt.Errorf("invalid_value")
		}
		targets = append(targets, target)
	}
	*out = targets
	return nil
}

func BuildClipboardPastePlan(request ClipboardPasteRequest) (ClipboardPastePlan, error) {
	if plan, ok, err := buildExactTimelineV2HeaderPlan(request); ok || err != nil {
		return plan, err
	}
	batch, err := tabularingest.BuildBatchPlan(tabularingest.MappingRequest{
		ViewSchemaID:   request.ViewSchemaID,
		ClientTxnID:    request.ClientTxnID,
		SourceKind:     request.sourceKind(),
		Text:           request.ClipboardText,
		Format:         request.Format,
		StartFieldKey:  request.StartFieldKey,
		Columns:        request.Columns,
		RequireTargets: len(request.Targets),
	})
	if err != nil {
		return ClipboardPastePlan{}, err
	}
	plan := ClipboardPastePlan{ClientTxnID: request.ClientTxnID, Rows: make([]ClipboardPasteRowPlan, 0, len(batch.Rows))}
	for _, batchRow := range batch.Rows {
		rowPlan := ClipboardPasteRowPlan{RowOrdinal: batchRow.RowOrdinal}
		for _, cell := range batchRow.Cells {
			change, ok := clipboardValueToPatchChange(cell.FieldKey, cell.RawValue)
			if !ok {
				fieldKey := cell.FieldKey
				rowPlan.Unknown = append(rowPlan.Unknown, ClipboardRawImportColumn{
					SourceKind:          batch.SourceKind,
					PasteClientTxnID:    batch.ClientTxnID,
					SourceRowOrdinal:    batchRow.RowOrdinal,
					SourceColumnOrdinal: cell.SourceColumnOrdinal,
					SourceHeaderText:    fieldKey,
					RawValue:            cell.RawValue,
				})
				continue
			}
			rowPlan.Cells = append(rowPlan.Cells, clipboardPasteCell{FieldKey: cell.FieldKey, Value: cell.RawValue, Change: change})
		}
		for _, unknown := range batchRow.Unknown {
			rowPlan.Unknown = append(rowPlan.Unknown, ClipboardRawImportColumn{
				SourceKind:          unknown.SourceKind,
				PasteClientTxnID:    unknown.SourceClientTxnID,
				SourceRowOrdinal:    unknown.SourceRowOrdinal,
				SourceColumnOrdinal: unknown.SourceColumnOrdinal,
				SourceHeaderText:    unknown.SourceHeaderText,
				RawValue:            unknown.RawValue,
			})
		}
		plan.Rows = append(plan.Rows, rowPlan)
	}
	return plan, nil
}

func buildExactTimelineV2HeaderPlan(request ClipboardPasteRequest) (ClipboardPastePlan, bool, error) {
	rows, err := tabularingest.ParseTable(request.ClipboardText, request.Format)
	if err != nil || len(rows) == 0 || !isTimelineV2ExactHeader(rows[0]) {
		return ClipboardPastePlan{}, false, nil
	}
	dataRows := rows[1:]
	if len(dataRows) == 0 || len(dataRows) > maxClipboardPasteRows {
		return ClipboardPastePlan{}, true, fmt.Errorf("invalid tabular row count")
	}
	if len(request.Targets) != len(dataRows) {
		return ClipboardPastePlan{}, true, fmt.Errorf("target count must equal tabular row count")
	}
	plan := ClipboardPastePlan{ClientTxnID: request.ClientTxnID, Rows: make([]ClipboardPasteRowPlan, 0, len(dataRows))}
	for rowIndex, values := range dataRows {
		rowPlan := ClipboardPasteRowPlan{RowOrdinal: rowIndex + 1}
		for columnIndex, fieldKey := range timelineV2ExactHeaderFieldKeys {
			rawValue := ""
			if columnIndex < len(values) {
				rawValue = values[columnIndex]
			}
			change, ok := clipboardValueToPatchChange(fieldKey, rawValue)
			if !ok {
				return ClipboardPastePlan{}, true, fmt.Errorf("invalid value for %s", fieldKey)
			}
			rowPlan.Cells = append(rowPlan.Cells, clipboardPasteCell{FieldKey: fieldKey, Value: rawValue, Change: change})
		}
		for columnIndex := len(timelineV2ExactHeaderFieldKeys); columnIndex < len(values); columnIndex++ {
			rowPlan.Unknown = append(rowPlan.Unknown, ClipboardRawImportColumn{
				SourceKind:          request.sourceKind(),
				PasteClientTxnID:    request.ClientTxnID,
				SourceRowOrdinal:    rowIndex + 1,
				SourceColumnOrdinal: columnIndex + 1,
				SourceHeaderText:    nil,
				RawValue:            values[columnIndex],
			})
		}
		plan.Rows = append(plan.Rows, rowPlan)
	}
	return plan, true, nil
}

func isTimelineV2ExactHeader(row []string) bool {
	if len(row) != len(timelineV2ExactHeaderLabels) {
		return false
	}
	for index, label := range timelineV2ExactHeaderLabels {
		if row[index] != label {
			return false
		}
	}
	return true
}

func (request ClipboardPasteRequest) sourceKind() string {
	if strings.TrimSpace(request.SourceKind) != "" {
		return request.SourceKind
	}
	return "clipboard_paste"
}

func (request ClipboardPasteRequest) routeKey() string {
	if strings.TrimSpace(request.RouteKey) != "" {
		return request.RouteKey
	}
	return clipboardPasteRouteKey
}

func ParseClipboardTable(text string, format string) ([][]string, error) {
	return tabularingest.ParseTable(text, format)
}

func clipboardValueToPatchChange(fieldKey string, rawValue string) (PatchChange, bool) {
	rawJSON, _ := json.Marshal(rawValue)
	change := PatchChange{FieldKey: fieldKey}
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
		change.ActionPayload = payload
		change.CanonicalAny = canonicalCollectionActionPayload(payload)
		return change, true
	}
	if _, ok := directWritableFieldKeys[fieldKey]; !ok {
		return PatchChange{}, false
	}
	value, ok := normalizeFieldTextValue(fieldKey, rawJSON)
	if !ok {
		return PatchChange{}, false
	}
	change.TextValue = value
	change.CanonicalAny = canonicalChangeValue(change)
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

func TimelineClipboardPasteRequestHash(request ClipboardPasteRequest) []byte {
	targets := make([]map[string]any, 0, len(request.Targets))
	for _, target := range request.Targets {
		entry := map[string]any{"kind": target.Kind}
		if target.Kind == "record" {
			entry["record_id"] = target.RecordID.String()
			entry["base_row_version"] = target.BaseRowVersion
		}
		targets = append(targets, entry)
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id":  request.ViewSchemaID,
		"clipboard_text":  request.ClipboardText,
		"format":          request.Format,
		"start_field_key": request.StartFieldKey,
		"columns":         request.Columns,
		"targets":         targets,
	})
}
