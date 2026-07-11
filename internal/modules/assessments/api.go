package assessments

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

const (
	AssessmentsViewSchemaID = "cartulary.view.assessments.v1"

	assessmentCreateRouteKey = "assessments.rows.create"
	maxSupportActions        = 64
)

type CreateRequest struct {
	ClientTxnID     string
	SubjectRef      *uuid.UUID
	SubjectType     string
	AssessmentState string
	ConfidenceScore *int
	Rationale       string
	Assessor        *uuid.UUID
	AssessedAt      *time.Time
	SupportRefs     []uuid.UUID
}

type ProjectionRecord struct {
	RecordID            uuid.UUID
	IncidentID          uuid.UUID
	RowVersion          int64
	SubjectRef          uuid.UUID
	SubjectType         string
	AssessmentState     string
	ConfidenceScore     *int
	ConfidenceBand      string
	Rationale           string
	Assessor            uuid.UUID
	AssessedAt          time.Time
	SupportingLinkCount int
	SupportRefs         []SupportRef
}

type SupportRef struct {
	LinkedRecordID uuid.UUID
	RecordType     string
}

type MutationResult struct {
	Payload     map[string]any
	StatusCode  int
	Replayed    bool
	RecordID    uuid.UUID
	ChangeSetID uuid.UUID
	RowVersion  int64
}

type CreateValidationError struct {
	Field      string
	ReasonCode string
}

func (e *CreateValidationError) Error() string {
	return "assessments: invalid create request"
}

func DecodeCreateRequest(reader io.Reader) (CreateRequest, *httpapi.APIError) {
	schema, ok := viewschema.Lookup(AssessmentsViewSchemaID)
	if !ok {
		return CreateRequest{}, invalidMutationPayload("view_schema_id", "unknown_view_schema")
	}

	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return CreateRequest{}, apiErr
	}

	allowed := map[string]struct{}{"client_txn_id": {}}
	for fieldKey, field := range schema.Fields() {
		if field.Writable || field.CreateWritable {
			allowed[fieldKey] = struct{}{}
		}
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return CreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request CreateRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return CreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	for fieldKey, value := range raw {
		switch fieldKey {
		case "client_txn_id":
		case "assessment.subject_ref":
			parsed, ok := decodeUUID(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.SubjectRef = &parsed
		case "assessment.subject_type":
			parsed, ok := decodeNormalizedLine(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.SubjectType = parsed
		case "assessment.assessment_state":
			parsed, ok := decodeNormalizedLine(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.AssessmentState = parsed
		case "assessment.confidence_score":
			score, ok := decodeOptionalInt(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.ConfidenceScore = score
		case "assessment.rationale":
			parsed, ok := decodeNormalizedNote(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.Rationale = parsed
		case "assessment.assessor":
			parsed, ok := decodeUUID(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.Assessor = &parsed
		case "assessment.assessed_at":
			parsed, ok := decodeTimestamp(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.AssessedAt = &parsed
		case "assessment.support_refs":
			refs, ok := decodeSupportActionPayload(value)
			if !ok {
				return CreateRequest{}, invalidMutationPayload(fieldKey, "invalid_value")
			}
			request.SupportRefs = refs
		default:
			return CreateRequest{}, invalidMutationPayload(fieldKey, "unknown_field")
		}
	}

	return request, nil
}

func CreateRequestHash(request CreateRequest) []byte {
	payload := map[string]any{
		"view_schema_id": AssessmentsViewSchemaID,
		"client_txn_id":  request.ClientTxnID,
	}
	if request.SubjectRef != nil {
		payload["assessment.subject_ref"] = request.SubjectRef.String()
	}
	if request.SubjectType != "" {
		payload["assessment.subject_type"] = request.SubjectType
	}
	if request.AssessmentState != "" {
		payload["assessment.assessment_state"] = request.AssessmentState
	}
	if request.ConfidenceScore != nil {
		payload["assessment.confidence_score"] = *request.ConfidenceScore
	}
	if request.Rationale != "" {
		payload["assessment.rationale"] = request.Rationale
	}
	if request.Assessor != nil {
		payload["assessment.assessor"] = request.Assessor.String()
	}
	if request.AssessedAt != nil {
		payload["assessment.assessed_at"] = request.AssessedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(request.SupportRefs) > 0 {
		refs := make([]string, 0, len(request.SupportRefs))
		for _, ref := range request.SupportRefs {
			refs = append(refs, ref.String())
		}
		slices.Sort(refs)
		payload["assessment.support_refs"] = refs
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func BuildAssessmentRow(record ProjectionRecord) map[string]any {
	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells": map[string]any{
			"assessment.subject_ref":           map[string]any{"value": record.SubjectRef.String()},
			"assessment.subject_type":          map[string]any{"value": record.SubjectType},
			"assessment.assessment_state":      map[string]any{"value": record.AssessmentState},
			"assessment.confidence_band":       map[string]any{"value": record.ConfidenceBand},
			"assessment.confidence_score":      map[string]any{"value": derefInt(record.ConfidenceScore)},
			"assessment.rationale":             map[string]any{"value": record.Rationale},
			"assessment.assessor":              map[string]any{"value": record.Assessor.String()},
			"assessment.assessed_at":           map[string]any{"value": formatTimestamp(record.AssessedAt)},
			"assessment.support_refs":          map[string]any{"value": collectionValue(false, supportRefItems(record.SupportRefs))},
			"assessment.supporting_link_count": map[string]any{"value": record.SupportingLinkCount},
		},
	}
	return row
}

func BuildMutationPayload(changeSetID uuid.UUID, row map[string]any) map[string]any {
	return map[string]any{
		"view_schema_id": AssessmentsViewSchemaID,
		"change_set_id":  changeSetID.String(),
		"row":            row,
	}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func decodeUUID(value json.RawMessage) (uuid.UUID, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return uuid.UUID{}, false
	}
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	return parsed, err == nil
}

func decodeNormalizedLine(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	return fieldnorm.NormalizeLine(raw)
}

func decodeNormalizedNote(value json.RawMessage) (string, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return "", false
	}
	return fieldnorm.NormalizeNote(raw)
}

func decodeOptionalInt(value json.RawMessage) (*int, bool) {
	if string(value) == "null" {
		return nil, true
	}
	var parsed int
	if err := json.Unmarshal(value, &parsed); err != nil {
		return nil, false
	}
	return &parsed, true
}

func decodeTimestamp(value json.RawMessage) (time.Time, bool) {
	var raw string
	if err := json.Unmarshal(value, &raw); err != nil {
		return time.Time{}, false
	}
	return fieldnorm.NormalizeTimestampInstant(raw)
}

func decodeSupportActionPayload(value json.RawMessage) ([]uuid.UUID, bool) {
	var payload struct {
		Kind    string                       `json:"kind"`
		Actions []map[string]json.RawMessage `json:"actions"`
	}
	if err := json.Unmarshal(value, &payload); err != nil {
		return nil, false
	}
	if payload.Kind != "collection_actions_v1" || len(payload.Actions) == 0 || len(payload.Actions) > maxSupportActions {
		return nil, false
	}
	refs := make([]uuid.UUID, 0, len(payload.Actions))
	for _, rawAction := range payload.Actions {
		if len(rawAction) != 2 {
			return nil, false
		}
		var op string
		if err := json.Unmarshal(rawAction["op"], &op); err != nil || op != "add_record_ref" {
			return nil, false
		}
		rawID, ok := rawAction["linked_record_id"]
		if !ok {
			return nil, false
		}
		parsed, ok := decodeUUID(rawID)
		if !ok {
			return nil, false
		}
		refs = append(refs, parsed)
	}
	return refs, true
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func collectionValue(ordered bool, items []map[string]any) map[string]any {
	if items == nil {
		items = []map[string]any{}
	}
	return map[string]any{
		"kind":    "collection_value_v1",
		"ordered": ordered,
		"items":   items,
	}
}

func supportRefItems(refs []SupportRef) []map[string]any {
	if len(refs) == 0 {
		return nil
	}
	items := make([]map[string]any, 0, len(refs))
	for _, ref := range refs {
		items = append(items, map[string]any{
			"item_ref":         "record_ref:" + ref.LinkedRecordID.String(),
			"item_kind":        "record_ref",
			"display_text":     ref.RecordType + ":" + ref.LinkedRecordID.String(),
			"linked_record_id": ref.LinkedRecordID.String(),
		})
	}
	return items
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func derefInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}
