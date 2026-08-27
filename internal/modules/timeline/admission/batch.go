package admission

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mutationpolicy"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/valuecodec"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type ClipboardPasteRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	ClipboardText string
	Format        string
	StartFieldKey string
	Columns       []string
	Targets       []timeline.OwnerBatchTargetV1
}

func DecodeClipboardPasteRequest(reader io.Reader) (ClipboardPasteRequest, *httpapi.APIError) {
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
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != timeline.TimelineViewSchemaID {
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
	case "", "auto", tabularingest.SourceFormatTSV, tabularingest.SourceFormatCSV:
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
	if _, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, request.StartFieldKey); !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "unsupported_field_key")
	}
	if value, ok := raw["columns"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Columns); err != nil || len(request.Columns) == 0 || len(request.Columns) > tabularingest.MaxClipboardCols {
		return ClipboardPasteRequest{}, invalidMutationPayload("columns", "invalid_value")
	}
	for _, fieldKey := range request.Columns {
		if _, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, fieldKey); !ok {
			return ClipboardPasteRequest{}, invalidMutationPayload("columns", "unsupported_field_key")
		}
	}
	if value, ok := raw["targets"]; !ok {
		return ClipboardPasteRequest{}, invalidMutationPayload("targets", "missing_required_field")
	} else if err := decodeBatchTargets(value, true, &request.Targets); err != nil {
		return ClipboardPasteRequest{}, invalidMutationPayload("targets", err.Error())
	}
	return request, nil
}

func BuildClipboardPlan(request ClipboardPasteRequest) (tabularingest.TabularRowPlanV1, error) {
	exactHeaderLabels, exactHeaderFieldKeys, err := timelineV2ExactHeaders()
	if err != nil {
		return tabularingest.TabularRowPlanV1{}, err
	}
	return tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
		ViewSchemaID:         request.ViewSchemaID,
		ClientTxnID:          request.ClientTxnID,
		SourceKind:           "clipboard_paste",
		Text:                 request.ClipboardText,
		Format:               request.Format,
		StartFieldKey:        request.StartFieldKey,
		Columns:              request.Columns,
		ExactHeaderLabels:    exactHeaderLabels,
		ExactHeaderFieldKeys: exactHeaderFieldKeys,
		RequireTargets:       len(request.Targets),
	})
}

func timelineV2ExactHeaders() ([]string, []string, error) {
	resource, ok := viewschema.LookupPublicResource(timeline.TimelineViewSchemaID)
	if !ok {
		return nil, nil, fmt.Errorf("timeline public view schema is unavailable")
	}
	labels := make([]string, 0, len(resource.Fields))
	fieldKeys := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.DefaultHidden || !field.GridEditable {
			continue
		}
		labels = append(labels, field.Label)
		fieldKeys = append(fieldKeys, field.FieldKey)
	}
	return labels, fieldKeys, nil
}

func ClipboardPasteRequestHash(request ClipboardPasteRequest) []byte {
	targets := make([]map[string]any, 0, len(request.Targets))
	for _, target := range request.Targets {
		entry := map[string]any{"kind": target.Kind}
		if target.Kind == "record" {
			entry["record_id"] = target.RecordID.String()
			entry["base_row_version"] = target.BaseRowVersion
		}
		targets = append(targets, entry)
	}
	return valuecodec.CanonicalJSONSHA256(map[string]any{
		"view_schema_id":  request.ViewSchemaID,
		"clipboard_text":  request.ClipboardText,
		"format":          request.Format,
		"start_field_key": request.StartFieldKey,
		"columns":         request.Columns,
		"targets":         targets,
	})
}

type BulkMutationRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	Kind          string
	FieldKey      string
	Value         string
	TagName       string
	NormalizedTag string
	Targets       []timeline.OwnerBatchTargetV1
}

func DecodeBulkMutationRequest(reader io.Reader, pathViewSchemaID string) (BulkMutationRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader, invalidMutationPayload)
	if apiErr != nil {
		return BulkMutationRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"view_schema_id": {},
		"client_txn_id":  {},
		"kind":           {},
		"field_key":      {},
		"value":          {},
		"tag_name":       {},
		"targets":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return BulkMutationRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request BulkMutationRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return BulkMutationRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != pathViewSchemaID {
		return BulkMutationRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if request.ViewSchemaID != timeline.TimelineViewSchemaID {
		return BulkMutationRequest{}, invalidMutationPayload("view_schema_id", "unsupported_view_schema")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return BulkMutationRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return BulkMutationRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["kind"]; !ok {
		return BulkMutationRequest{}, invalidMutationPayload("kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Kind); err != nil {
		return BulkMutationRequest{}, invalidMutationPayload("kind", "invalid_value")
	}
	switch request.Kind {
	case timeline.OwnerBatchOperationFillDownV1:
		if _, hasTag := raw["tag_name"]; hasTag {
			return BulkMutationRequest{}, invalidMutationPayload("tag_name", "forbidden_field")
		}
		if value, ok := raw["field_key"]; !ok {
			return BulkMutationRequest{}, invalidMutationPayload("field_key", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.FieldKey); err != nil || request.FieldKey == "" {
			return BulkMutationRequest{}, invalidMutationPayload("field_key", "invalid_value")
		}
		field, ok := viewschema.LookupField(request.ViewSchemaID, request.FieldKey)
		if !ok || !field.Writable || field.ConflictResolutionClass == "collection_review" {
			return BulkMutationRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		if value, ok := raw["value"]; !ok {
			return BulkMutationRequest{}, invalidMutationPayload("value", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.Value); err != nil {
			return BulkMutationRequest{}, invalidMutationPayload("value", "invalid_value")
		}
	case timeline.OwnerBatchOperationMultiRowTagAssignmentV1:
		if _, hasField := raw["field_key"]; hasField {
			return BulkMutationRequest{}, invalidMutationPayload("field_key", "forbidden_field")
		}
		if _, hasValue := raw["value"]; hasValue {
			return BulkMutationRequest{}, invalidMutationPayload("value", "forbidden_field")
		}
		if value, ok := raw["tag_name"]; !ok {
			return BulkMutationRequest{}, invalidMutationPayload("tag_name", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.TagName); err != nil {
			return BulkMutationRequest{}, invalidMutationPayload("tag_name", "invalid_value")
		}
		label, normalized, ok := fieldnorm.NormalizeTagLabel(request.TagName)
		if !ok {
			return BulkMutationRequest{}, invalidMutationPayload("tag_name", "invalid_value")
		}
		request.TagName = label
		request.NormalizedTag = normalized
	default:
		return BulkMutationRequest{}, invalidMutationPayload("kind", "invalid_value")
	}
	if value, ok := raw["targets"]; !ok {
		return BulkMutationRequest{}, invalidMutationPayload("targets", "missing_required_field")
	} else if err := decodeBatchTargets(value, false, &request.Targets); err != nil {
		if _, invalidVersion := err.(invalidBaseRowVersionError); invalidVersion {
			return BulkMutationRequest{}, invalidMutationPayload("base_row_version", err.Error())
		}
		return BulkMutationRequest{}, invalidMutationPayload("targets", err.Error())
	}
	return request, nil
}

func BulkMutationRequestHash(request BulkMutationRequest) []byte {
	targets := make([]map[string]any, 0, len(request.Targets))
	for _, target := range request.Targets {
		targets = append(targets, map[string]any{
			"record_id":        target.RecordID.String(),
			"base_row_version": target.BaseRowVersion,
		})
	}
	fieldKey := request.FieldKey
	value := request.Value
	if request.Kind == timeline.OwnerBatchOperationMultiRowTagAssignmentV1 {
		fieldKey = "timeline.tags"
		value = request.TagName
	}
	return valuecodec.CanonicalJSONSHA256(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"kind":           request.Kind,
		"field_key":      fieldKey,
		"value":          value,
		"tag_name":       request.TagName,
		"targets":        targets,
	})
}

func decodeBatchTargets(raw json.RawMessage, allowCreate bool, out *[]timeline.OwnerBatchTargetV1) error {
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 || len(items) > mutationpolicy.MaxOwnerBatchTargets {
		return invalidTargetValueError{}
	}
	targets := make([]timeline.OwnerBatchTargetV1, 0, len(items))
	for _, item := range items {
		if allowCreate {
			if !objectHasOnlyFields(item, "kind") && !objectHasOnlyFields(item, "kind", "record_id", "base_row_version") {
				return invalidTargetValueError{}
			}
		} else if !objectHasOnlyFields(item, "record_id", "base_row_version") {
			return invalidTargetValueError{}
		}
		target := timeline.OwnerBatchTargetV1{}
		if rawKind, ok := item["kind"]; ok {
			if err := json.Unmarshal(rawKind, &target.Kind); err != nil {
				return invalidTargetValueError{}
			}
		} else if !allowCreate {
			target.Kind = "record"
		}
		switch target.Kind {
		case "create":
			if !allowCreate || len(item) != 1 {
				return invalidTargetValueError{}
			}
		case "record":
			var rawID string
			if err := json.Unmarshal(item["record_id"], &rawID); err != nil {
				return invalidTargetValueError{}
			}
			recordID, err := uuid.Parse(rawID)
			if err != nil || recordID == uuid.Nil {
				return invalidTargetValueError{}
			}
			target.RecordID = recordID
			if err := json.Unmarshal(item["base_row_version"], &target.BaseRowVersion); err != nil || target.BaseRowVersion < 1 {
				return invalidBaseRowVersionError{}
			}
		default:
			return invalidTargetValueError{}
		}
		targets = append(targets, target)
	}
	*out = targets
	return nil
}

type invalidTargetValueError struct{}

func (invalidTargetValueError) Error() string { return "invalid_value" }

type invalidBaseRowVersionError struct{}

func (invalidBaseRowVersionError) Error() string { return "invalid_base_row_version" }
