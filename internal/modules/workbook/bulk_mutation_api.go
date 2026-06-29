package workbook

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const workbookBulkMutationRouteKey = "workbook.bulk_mutations"

type BulkMutationRequest struct {
	ViewSchemaID  string
	ClientTxnID   string
	Kind          string
	FieldKey      string
	Value         string
	TagName       string
	NormalizedTag string
	Targets       []BulkMutationTarget
}

type BulkMutationTarget struct {
	RecordID       uuid.UUID
	BaseRowVersion int64
}

func DecodeBulkMutationRequest(reader io.Reader, pathViewSchemaID string) (BulkMutationRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
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
	case "fill_down_v1":
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
	case "multi_row_tag_assignment_v1":
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
		request.FieldKey = "timeline.tags"
		request.Value = label
	default:
		return BulkMutationRequest{}, invalidMutationPayload("kind", "invalid_value")
	}
	targets, apiErr := decodeBulkMutationTargets(raw["targets"])
	if apiErr != nil {
		return BulkMutationRequest{}, apiErr
	}
	request.Targets = targets
	return request, nil
}

func decodeBulkMutationTargets(value json.RawMessage) ([]BulkMutationTarget, *auth.APIError) {
	if len(value) == 0 {
		return nil, invalidMutationPayload("targets", "missing_required_field")
	}
	var rawTargets []map[string]json.RawMessage
	if err := json.Unmarshal(value, &rawTargets); err != nil || len(rawTargets) == 0 || len(rawTargets) > 500 {
		return nil, invalidMutationPayload("targets", "invalid_value")
	}
	targets := make([]BulkMutationTarget, 0, len(rawTargets))
	for _, raw := range rawTargets {
		if !objectHasOnlyFields(raw, "record_id", "base_row_version") {
			return nil, invalidMutationPayload("targets", "invalid_value")
		}
		var rawID string
		if err := json.Unmarshal(raw["record_id"], &rawID); err != nil {
			return nil, invalidMutationPayload("targets", "invalid_value")
		}
		recordID, err := uuid.Parse(rawID)
		if err != nil || recordID == uuid.Nil {
			return nil, invalidMutationPayload("targets", "invalid_value")
		}
		var baseRowVersion int64
		if err := json.Unmarshal(raw["base_row_version"], &baseRowVersion); err != nil || baseRowVersion < 1 {
			return nil, invalidMutationPayload("base_row_version", "invalid_base_row_version")
		}
		targets = append(targets, BulkMutationTarget{RecordID: recordID, BaseRowVersion: baseRowVersion})
	}
	return targets, nil
}

func timelineBulkClipboardRequest(request BulkMutationRequest) timeline.ClipboardPasteRequest {
	lines := make([]string, 0, len(request.Targets))
	targets := make([]timeline.ClipboardPasteTarget, 0, len(request.Targets))
	for _, target := range request.Targets {
		lines = append(lines, request.Value)
		targets = append(targets, timeline.ClipboardPasteTarget{
			Kind:           "record",
			RecordID:       target.RecordID,
			BaseRowVersion: target.BaseRowVersion,
		})
	}
	return timeline.ClipboardPasteRequest{
		ViewSchemaID:  request.ViewSchemaID,
		ClientTxnID:   request.ClientTxnID,
		ClipboardText: strings.Join(lines, "\n"),
		Format:        "tsv",
		StartFieldKey: request.FieldKey,
		Columns:       []string{request.FieldKey},
		Targets:       targets,
		SourceKind:    "bulk_edit",
		RouteKey:      workbookBulkMutationRouteKey,
	}
}

func BulkMutationRequestHash(request BulkMutationRequest) []byte {
	targets := make([]map[string]any, 0, len(request.Targets))
	for _, target := range request.Targets {
		targets = append(targets, map[string]any{
			"record_id":        target.RecordID.String(),
			"base_row_version": target.BaseRowVersion,
		})
	}
	return hashRequestPayload(map[string]any{
		"view_schema_id": request.ViewSchemaID,
		"kind":           request.Kind,
		"field_key":      request.FieldKey,
		"value":          request.Value,
		"tag_name":       request.TagName,
		"targets":        targets,
	})
}
