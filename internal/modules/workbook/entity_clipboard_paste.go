package workbook

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type entityClipboardPasteRequest struct {
	ViewSchemaID   string
	ClientTxnID    string
	ClipboardText  string
	Format         string
	StartFieldKey  string
	Columns        []string
	CreateOnlyRows int
}

func decodeEntityClipboardPasteRequest(reader io.Reader, pathViewSchemaID string) (entityClipboardPasteRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return entityClipboardPasteRequest{}, apiErr
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
			return entityClipboardPasteRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request entityClipboardPasteRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != pathViewSchemaID {
		return entityClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if _, ok := viewschema.Lookup(request.ViewSchemaID); !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}
	if request.ViewSchemaID != hostidentity.HostsViewSchemaID && request.ViewSchemaID != hostidentity.IdentitiesViewSchemaID {
		return entityClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "unsupported_view_schema")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return entityClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["clipboard_text"]; !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClipboardText); err != nil || request.ClipboardText == "" {
		return entityClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "invalid_value")
	}
	request.Format = "auto"
	if value, ok := raw["format"]; ok {
		if err := json.Unmarshal(value, &request.Format); err != nil {
			return entityClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
		}
	}
	switch request.Format {
	case "", "auto", "tsv", "csv":
		if request.Format == "" {
			request.Format = "auto"
		}
	default:
		return entityClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
	}
	if value, ok := raw["start_field_key"]; !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.StartFieldKey); err != nil || request.StartFieldKey == "" {
		return entityClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "invalid_value")
	}
	if field, ok := viewschema.LookupField(request.ViewSchemaID, request.StartFieldKey); !ok || !field.Writable {
		return entityClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "unsupported_field_key")
	}
	if value, ok := raw["columns"]; !ok {
		return entityClipboardPasteRequest{}, invalidMutationPayload("columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Columns); err != nil || len(request.Columns) == 0 || len(request.Columns) > tabularingest.MaxClipboardCols {
		return entityClipboardPasteRequest{}, invalidMutationPayload("columns", "invalid_value")
	}
	for _, fieldKey := range request.Columns {
		if _, ok := viewschema.LookupField(request.ViewSchemaID, fieldKey); !ok {
			return entityClipboardPasteRequest{}, invalidMutationPayload("columns", "unsupported_field_key")
		}
	}
	request.CreateOnlyRows = -1
	if value, ok := raw["targets"]; ok {
		var targets []map[string]json.RawMessage
		if err := json.Unmarshal(value, &targets); err != nil || len(targets) == 0 || len(targets) > tabularingest.MaxClipboardRows {
			return entityClipboardPasteRequest{}, invalidMutationPayload("targets", "invalid_value")
		}
		for _, target := range targets {
			if !objectHasOnlyFields(target, "kind") {
				return entityClipboardPasteRequest{}, invalidMutationPayload("targets", "invalid_value")
			}
			var kind string
			if err := json.Unmarshal(target["kind"], &kind); err != nil || kind != "create" {
				return entityClipboardPasteRequest{}, invalidMutationPayload("targets", "invalid_value")
			}
		}
		request.CreateOnlyRows = len(targets)
	}
	return request, nil
}

func (request entityClipboardPasteRequest) mappingRequest() tabularingest.MappingRequest {
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

func buildEntityClipboardPastePlan(request entityClipboardPasteRequest) (tabularingest.BatchPlan, error) {
	return tabularingest.BuildBatchPlan(request.mappingRequest())
}

func entityClipboardPasteRequestHash(request entityClipboardPasteRequest) []byte {
	return hostidentity.EntityClipboardPasteRequestHash(request.ViewSchemaID, request.ClientTxnID, request.ClipboardText, request.Format, request.StartFieldKey, request.Columns)
}
