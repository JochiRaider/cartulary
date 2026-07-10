package hostidentity

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ClipboardPasteRequest struct {
	ViewSchemaID   string
	ClientTxnID    string
	ClipboardText  string
	Format         string
	StartFieldKey  string
	Columns        []string
	CreateOnlyRows int
}

func DecodeClipboardPasteRequest(reader io.Reader, pathViewSchemaID string) (ClipboardPasteRequest, *httpapi.APIError) {
	raw, apiErr := decodeClipboardPasteObject(reader)
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
			return ClipboardPasteRequest{}, invalidClipboardPastePayload(key, "unknown_field")
		}
	}

	var request ClipboardPasteRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != pathViewSchemaID {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("view_schema_id", "invalid_view_schema_id")
	}
	if _, ok := viewschema.Lookup(request.ViewSchemaID); !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("view_schema_id", "unknown_view_schema")
	}
	if request.ViewSchemaID != HostsViewSchemaID && request.ViewSchemaID != IdentitiesViewSchemaID {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("view_schema_id", "unsupported_view_schema")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["clipboard_text"]; !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("clipboard_text", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClipboardText); err != nil || request.ClipboardText == "" {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("clipboard_text", "invalid_value")
	}
	request.Format = "auto"
	if value, ok := raw["format"]; ok {
		if err := json.Unmarshal(value, &request.Format); err != nil {
			return ClipboardPasteRequest{}, invalidClipboardPastePayload("format", "invalid_value")
		}
	}
	switch request.Format {
	case "", "auto", "tsv", "csv":
		if request.Format == "" {
			request.Format = "auto"
		}
	default:
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("format", "invalid_value")
	}
	if value, ok := raw["start_field_key"]; !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("start_field_key", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.StartFieldKey); err != nil || request.StartFieldKey == "" {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("start_field_key", "invalid_value")
	}
	if field, ok := viewschema.LookupField(request.ViewSchemaID, request.StartFieldKey); !ok || !field.Writable {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("start_field_key", "unsupported_field_key")
	}
	if value, ok := raw["columns"]; !ok {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Columns); err != nil || len(request.Columns) == 0 || len(request.Columns) > tabularingest.MaxClipboardCols {
		return ClipboardPasteRequest{}, invalidClipboardPastePayload("columns", "invalid_value")
	}
	for _, fieldKey := range request.Columns {
		if _, ok := viewschema.LookupField(request.ViewSchemaID, fieldKey); !ok {
			return ClipboardPasteRequest{}, invalidClipboardPastePayload("columns", "unsupported_field_key")
		}
	}
	request.CreateOnlyRows = -1
	if value, ok := raw["targets"]; ok {
		var targets []map[string]json.RawMessage
		if err := json.Unmarshal(value, &targets); err != nil || len(targets) == 0 || len(targets) > tabularingest.MaxClipboardRows {
			return ClipboardPasteRequest{}, invalidClipboardPastePayload("targets", "invalid_value")
		}
		for _, target := range targets {
			if !clipboardPasteObjectHasOnlyFields(target, "kind") {
				return ClipboardPasteRequest{}, invalidClipboardPastePayload("targets", "invalid_value")
			}
			var kind string
			if err := json.Unmarshal(target["kind"], &kind); err != nil || kind != "create" {
				return ClipboardPasteRequest{}, invalidClipboardPastePayload("targets", "invalid_value")
			}
		}
		request.CreateOnlyRows = len(targets)
	}
	return request, nil
}

func BuildClipboardPastePlan(request ClipboardPasteRequest) (tabularingest.BatchPlan, error) {
	return tabularingest.BuildBatchPlan(request.mappingRequest())
}

func (request ClipboardPasteRequest) RequestHash() []byte {
	return EntityClipboardPasteRequestHash(request.ViewSchemaID, request.ClientTxnID, request.ClipboardText, request.Format, request.StartFieldKey, request.Columns)
}

func (request ClipboardPasteRequest) mappingRequest() tabularingest.MappingRequest {
	return tabularingest.MappingRequest{
		ViewSchemaID:   request.ViewSchemaID,
		ClientTxnID:    request.ClientTxnID,
		SourceKind:     "clipboard_paste",
		Text:           request.ClipboardText,
		Format:         request.Format,
		StartFieldKey:  request.StartFieldKey,
		Columns:        request.Columns,
		RequireTargets: request.CreateOnlyRows,
	}
}

func decodeClipboardPasteObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidClipboardPastePayload("", "request_not_object")
	}
	return raw, nil
}

func invalidClipboardPastePayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_mutation_payload", Message: "invalid mutation payload", Details: details}
}

func clipboardPasteObjectHasOnlyFields(object map[string]json.RawMessage, fields ...string) bool {
	allowed := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		allowed[field] = struct{}{}
		if _, ok := object[field]; !ok {
			return false
		}
	}
	for key := range object {
		if _, ok := allowed[key]; !ok {
			return false
		}
	}
	return true
}
