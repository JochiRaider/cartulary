package admission

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const maxSupportActions = 64

var createFields = map[string]struct{}{
	"client_txn_id":               {},
	"assessment.subject_ref":      {},
	"assessment.subject_type":     {},
	"assessment.assessment_state": {},
	"assessment.confidence_score": {},
	"assessment.rationale":        {},
	"assessment.assessor":         {},
	"assessment.assessed_at":      {},
	"assessment.support_refs":     {},
}

// DecodeCreateRequest admits the public create body into Assessment-owned
// command input. All Assessment field and collection semantics stay on the
// owner side of the Workbook boundary.
func DecodeCreateRequest(reader io.Reader) (assessments.CreateInput, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return assessments.CreateInput{}, invalidMutationPayload("", "request_not_object")
	}
	for key := range raw {
		if _, ok := createFields[key]; !ok {
			return assessments.CreateInput{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var input assessments.CreateInput
	if value, ok := raw["client_txn_id"]; !ok {
		return assessments.CreateInput{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if json.Unmarshal(value, &input.ClientTxnID) != nil || strings.TrimSpace(input.ClientTxnID) == "" {
		return assessments.CreateInput{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["assessment.subject_ref"]; ok {
		parsed, valid := decodeCanonicalUUID(value)
		if !valid {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.subject_ref", "invalid_value")
		}
		input.SubjectRef = parsed
	}
	if value, ok := raw["assessment.subject_type"]; ok {
		normalized, valid := decodeNormalizedLine(value)
		if !valid {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.subject_type", "invalid_value")
		}
		input.SubjectType = normalized
	}
	if value, ok := raw["assessment.assessment_state"]; ok {
		normalized, valid := decodeNormalizedLine(value)
		if !valid {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.assessment_state", "invalid_value")
		}
		input.AssessmentState = normalized
	}
	if value, ok := raw["assessment.confidence_score"]; ok && string(value) != "null" {
		var score int
		if json.Unmarshal(value, &score) != nil {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.confidence_score", "invalid_value")
		}
		input.ConfidenceScore = &score
	}
	if value, ok := raw["assessment.rationale"]; ok {
		if string(value) == "null" {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.rationale", "invalid_value")
		} else {
			var rawText string
			if json.Unmarshal(value, &rawText) != nil {
				return assessments.CreateInput{}, invalidMutationPayload("assessment.rationale", "invalid_value")
			}
			normalized, valid := fieldnorm.NormalizeNote(rawText)
			if !valid {
				return assessments.CreateInput{}, invalidMutationPayload("assessment.rationale", "invalid_value")
			}
			input.Rationale = normalized
		}
	}
	if value, ok := raw["assessment.assessor"]; ok {
		parsed, valid := decodeCanonicalUUID(value)
		if !valid {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.assessor", "invalid_value")
		}
		input.Assessor = &parsed
	}
	if value, ok := raw["assessment.assessed_at"]; ok {
		var rawTime string
		if string(value) == "null" || json.Unmarshal(value, &rawTime) != nil {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.assessed_at", "invalid_value")
		}
		assessedAt, valid := fieldnorm.NormalizeTimestampInstant(rawTime)
		if !valid {
			return assessments.CreateInput{}, invalidMutationPayload("assessment.assessed_at", "invalid_value")
		}
		input.AssessedAt = &assessedAt
	}
	if value, ok := raw["assessment.support_refs"]; ok {
		refs, apiErr := decodeSupportRefs(value)
		if apiErr != nil {
			return assessments.CreateInput{}, apiErr
		}
		input.SupportRefs = refs
	}
	return input, nil
}

// CreateRequestHash returns the established canonical Assessment replay
// identity. Collection order is intentionally normalized for replay.
func CreateRequestHash(input assessments.CreateInput) []byte {
	payload := map[string]any{
		"view_schema_id": assessments.AssessmentsViewSchemaID,
		"client_txn_id":  input.ClientTxnID,
	}
	if input.SubjectRef != uuid.Nil {
		payload["assessment.subject_ref"] = input.SubjectRef.String()
	}
	if input.SubjectType != "" {
		payload["assessment.subject_type"] = input.SubjectType
	}
	if input.AssessmentState != "" {
		payload["assessment.assessment_state"] = input.AssessmentState
	}
	if input.ConfidenceScore != nil {
		payload["assessment.confidence_score"] = *input.ConfidenceScore
	}
	if input.Rationale != "" {
		payload["assessment.rationale"] = input.Rationale
	}
	if input.Assessor != nil {
		payload["assessment.assessor"] = input.Assessor.String()
	}
	if input.AssessedAt != nil {
		payload["assessment.assessed_at"] = input.AssessedAt.UTC().Format(time.RFC3339Nano)
	}
	if len(input.SupportRefs) > 0 {
		refs := make([]string, 0, len(input.SupportRefs))
		for _, ref := range input.SupportRefs {
			refs = append(refs, ref.String())
		}
		slices.Sort(refs)
		payload["assessment.support_refs"] = refs
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return append([]byte(nil), sum[:]...)
}

func decodeNormalizedLine(raw json.RawMessage) (string, bool) {
	if string(raw) == "null" {
		return "", false
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return "", false
	}
	return fieldnorm.NormalizeLine(value)
}

func decodeCanonicalUUID(raw json.RawMessage) (uuid.UUID, bool) {
	if string(raw) == "null" {
		return uuid.Nil, false
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return uuid.Nil, false
	}
	parsed, err := uuid.Parse(value)
	return parsed, err == nil && parsed.String() == value
}

func decodeSupportRefs(raw json.RawMessage) ([]uuid.UUID, *httpapi.APIError) {
	var object map[string]json.RawMessage
	if json.Unmarshal(raw, &object) != nil || !hasExactFields(object, "kind", "actions") {
		return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
	}
	var kind string
	if json.Unmarshal(object["kind"], &kind) != nil || kind != "collection_actions_v1" {
		return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
	}
	var actions []json.RawMessage
	if json.Unmarshal(object["actions"], &actions) != nil {
		return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
	}
	if len(actions) == 0 {
		return nil, invalidMutationPayloadWithDetails(
			"assessment.support_refs.actions",
			"empty_collection_actions",
			map[string]any{"field_key": "assessment.support_refs"},
		)
	}
	if len(actions) > maxSupportActions {
		return nil, invalidMutationPayloadWithDetails(
			"assessment.support_refs.actions",
			"collection_action_count_exceeded",
			map[string]any{
				"field_key":       "assessment.support_refs",
				"requested_count": len(actions),
				"max_count":       maxSupportActions,
			},
		)
	}
	refs := make([]uuid.UUID, 0, len(actions))
	for _, rawAction := range actions {
		var action map[string]json.RawMessage
		if json.Unmarshal(rawAction, &action) != nil || !hasExactFields(action, "op", "linked_record_id") {
			return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
		}
		var operation string
		if json.Unmarshal(action["op"], &operation) != nil || operation != "add_record_ref" {
			return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
		}
		ref, valid := decodeCanonicalUUID(action["linked_record_id"])
		if !valid {
			return nil, invalidMutationPayload("assessment.support_refs", "invalid_value")
		}
		refs = append(refs, ref)
	}
	return refs, nil
}

func hasExactFields(object map[string]json.RawMessage, fields ...string) bool {
	if len(object) != len(fields) {
		return false
	}
	for _, field := range fields {
		if _, ok := object[field]; !ok {
			return false
		}
	}
	return true
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
	details := make(map[string]any, len(extra)+2)
	for key, value := range extra {
		details[key] = value
	}
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
