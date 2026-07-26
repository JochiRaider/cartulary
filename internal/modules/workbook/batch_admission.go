package workbook

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/tabularingest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	maxWorkbookBatchRows = 500
	maxWorkbookBatchCols = 64
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

type TimelineClipboardPasteRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	ClipboardText string
	Format        string
	StartFieldKey string
	Columns       []string
	Targets       []TimelineBatchTarget
}

type TimelineBatchTarget struct {
	Kind           string
	RecordID       uuid.UUID
	BaseRowVersion int64
}

func DecodeTimelineClipboardPasteRequest(reader io.Reader) (TimelineClipboardPasteRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return TimelineClipboardPasteRequest{}, apiErr
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
			return TimelineClipboardPasteRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request TimelineClipboardPasteRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != timeline.TimelineViewSchemaID {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["clipboard_text"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClipboardText); err != nil || request.ClipboardText == "" {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("clipboard_text", "invalid_value")
	}
	request.Format = "auto"
	if value, ok := raw["format"]; ok {
		if err := json.Unmarshal(value, &request.Format); err != nil {
			return TimelineClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
		}
	}
	switch request.Format {
	case "", "auto", tabularingest.SourceFormatTSV, tabularingest.SourceFormatCSV:
		if request.Format == "" {
			request.Format = "auto"
		}
	default:
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("format", "invalid_value")
	}
	if value, ok := raw["start_field_key"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.StartFieldKey); err != nil || request.StartFieldKey == "" {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "invalid_value")
	}
	if _, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, request.StartFieldKey); !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("start_field_key", "unsupported_field_key")
	}
	if value, ok := raw["columns"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("columns", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Columns); err != nil || len(request.Columns) == 0 || len(request.Columns) > maxWorkbookBatchCols {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("columns", "invalid_value")
	}
	for _, fieldKey := range request.Columns {
		if _, ok := viewschema.LookupField(timeline.TimelineViewSchemaID, fieldKey); !ok {
			return TimelineClipboardPasteRequest{}, invalidMutationPayload("columns", "unsupported_field_key")
		}
	}
	if value, ok := raw["targets"]; !ok {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("targets", "missing_required_field")
	} else if err := decodeTimelineBatchTargets(value, true, &request.Targets); err != nil {
		return TimelineClipboardPasteRequest{}, invalidMutationPayload("targets", err.Error())
	}
	return request, nil
}

func BuildTimelineClipboardPlan(request TimelineClipboardPasteRequest) (tabularingest.TabularRowPlanV1, error) {
	return tabularingest.BuildTabularRowPlanV1(tabularingest.MappingRequest{
		ViewSchemaID:         request.ViewSchemaID,
		ClientTxnID:          request.ClientTxnID,
		SourceKind:           "clipboard_paste",
		Text:                 request.ClipboardText,
		Format:               request.Format,
		StartFieldKey:        request.StartFieldKey,
		Columns:              request.Columns,
		ExactHeaderLabels:    timelineV2ExactHeaderLabels,
		ExactHeaderFieldKeys: timelineV2ExactHeaderFieldKeys,
		RequireTargets:       len(request.Targets),
	})
}

func TimelineClipboardPasteRequestHash(request TimelineClipboardPasteRequest) []byte {
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

type TimelineBulkMutationRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	Kind          string
	FieldKey      string
	Value         string
	TagName       string
	NormalizedTag string
	Targets       []TimelineBatchTarget
}

func DecodeTimelineBulkMutationRequest(reader io.Reader, pathViewSchemaID string) (TimelineBulkMutationRequest, *httpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return TimelineBulkMutationRequest{}, apiErr
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
			return TimelineBulkMutationRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request TimelineBulkMutationRequest
	if value, ok := raw["view_schema_id"]; !ok {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("view_schema_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ViewSchemaID); err != nil || request.ViewSchemaID != pathViewSchemaID {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("view_schema_id", "invalid_view_schema_id")
	}
	if request.ViewSchemaID != timeline.TimelineViewSchemaID {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("view_schema_id", "unsupported_view_schema")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["kind"]; !ok {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.Kind); err != nil {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("kind", "invalid_value")
	}
	switch request.Kind {
	case timeline.OwnerBatchOperationFillDownV1:
		if _, hasTag := raw["tag_name"]; hasTag {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("tag_name", "forbidden_field")
		}
		if value, ok := raw["field_key"]; !ok {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("field_key", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.FieldKey); err != nil || request.FieldKey == "" {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("field_key", "invalid_value")
		}
		field, ok := viewschema.LookupField(request.ViewSchemaID, request.FieldKey)
		if !ok || !field.Writable || field.ConflictResolutionClass == "collection_review" {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("field_key", "unsupported_field_key")
		}
		if value, ok := raw["value"]; !ok {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("value", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.Value); err != nil {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("value", "invalid_value")
		}
	case timeline.OwnerBatchOperationMultiRowTagAssignmentV1:
		if _, hasField := raw["field_key"]; hasField {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("field_key", "forbidden_field")
		}
		if _, hasValue := raw["value"]; hasValue {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("value", "forbidden_field")
		}
		if value, ok := raw["tag_name"]; !ok {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("tag_name", "missing_required_field")
		} else if err := json.Unmarshal(value, &request.TagName); err != nil {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("tag_name", "invalid_value")
		}
		label, normalized, ok := fieldnorm.NormalizeTagLabel(request.TagName)
		if !ok {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("tag_name", "invalid_value")
		}
		request.TagName = label
		request.NormalizedTag = normalized
	default:
		return TimelineBulkMutationRequest{}, invalidMutationPayload("kind", "invalid_value")
	}
	if value, ok := raw["targets"]; !ok {
		return TimelineBulkMutationRequest{}, invalidMutationPayload("targets", "missing_required_field")
	} else if err := decodeTimelineBatchTargets(value, false, &request.Targets); err != nil {
		if _, invalidVersion := err.(invalidBaseRowVersionError); invalidVersion {
			return TimelineBulkMutationRequest{}, invalidMutationPayload("base_row_version", err.Error())
		}
		return TimelineBulkMutationRequest{}, invalidMutationPayload("targets", err.Error())
	}
	return request, nil
}

func TimelineBulkMutationRequestHash(request TimelineBulkMutationRequest) []byte {
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
	return hashRequestPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"kind":           request.Kind,
		"field_key":      fieldKey,
		"value":          value,
		"tag_name":       request.TagName,
		"targets":        targets,
	})
}

func decodeTimelineBatchTargets(raw json.RawMessage, allowCreate bool, out *[]TimelineBatchTarget) error {
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil || len(items) == 0 || len(items) > maxWorkbookBatchRows {
		return invalidTargetValueError{}
	}
	targets := make([]TimelineBatchTarget, 0, len(items))
	for _, item := range items {
		if allowCreate {
			if !objectHasOnlyFields(item, "kind") && !objectHasOnlyFields(item, "kind", "record_id", "base_row_version") {
				return invalidTargetValueError{}
			}
		} else if !objectHasOnlyFields(item, "record_id", "base_row_version") {
			return invalidTargetValueError{}
		}
		target := TimelineBatchTarget{}
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
