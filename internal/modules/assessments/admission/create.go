package admission

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type createFieldDecoder func(json.RawMessage, *assessments.CreateInput) *httpapi.APIError

var createFieldDecoders = map[string]createFieldDecoder{
	"client_txn_id":               decodeClientTxnID,
	"assessment.subject_ref":      decodeSubjectRef,
	"assessment.subject_type":     decodeSubjectType,
	"assessment.assessment_state": decodeAssessmentState,
	"assessment.confidence_score": decodeConfidenceScore,
	"assessment.rationale":        decodeRationale,
	"assessment.assessor":         decodeAssessor,
	"assessment.assessed_at":      decodeAssessedAt,
	"assessment.support_refs":     decodeSupportReferences,
}

// DecodeCreateRequest admits the public create body into Assessment-owned
// command input. All Assessment field and collection semantics stay on the
// owner side of the Workbook boundary.
func DecodeCreateRequest(reader io.Reader) (assessments.CreateInput, *httpapi.APIError) {
	raw, err := httpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return assessments.CreateInput{}, invalidMutationPayload("", "request_not_object")
	}
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, ok := createFieldDecoders[key]; !ok {
			return assessments.CreateInput{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var input assessments.CreateInput
	clientTxnID, ok := raw["client_txn_id"]
	if !ok {
		return assessments.CreateInput{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if apiErr := createFieldDecoders["client_txn_id"](clientTxnID, &input); apiErr != nil {
		return assessments.CreateInput{}, apiErr
	}
	for _, key := range keys {
		if key == "client_txn_id" {
			continue
		}
		apiErr := createFieldDecoders[key](raw[key], &input)
		if apiErr != nil {
			return assessments.CreateInput{}, apiErr
		}
	}
	return input, nil
}

func decodeClientTxnID(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	if json.Unmarshal(raw, &input.ClientTxnID) != nil || strings.TrimSpace(input.ClientTxnID) == "" {
		return invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	return nil
}

func decodeSubjectRef(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	parsed, valid := decodeCanonicalUUID(raw)
	if !valid {
		return invalidMutationPayload("assessment.subject_ref", "invalid_value")
	}
	input.SubjectRef = parsed
	return nil
}

func decodeSubjectType(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	normalized, valid := decodeNormalizedLine(raw)
	if !valid {
		return invalidMutationPayload("assessment.subject_type", "invalid_value")
	}
	input.SubjectType = normalized
	return nil
}

func decodeAssessmentState(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	normalized, valid := decodeNormalizedLine(raw)
	if !valid {
		return invalidMutationPayload("assessment.assessment_state", "invalid_value")
	}
	input.AssessmentState = normalized
	return nil
}

func decodeConfidenceScore(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	if string(raw) == "null" {
		return nil
	}
	var score int
	if json.Unmarshal(raw, &score) != nil {
		return invalidMutationPayload("assessment.confidence_score", "invalid_value")
	}
	input.ConfidenceScore = &score
	return nil
}

func decodeRationale(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	if string(raw) == "null" {
		return invalidMutationPayload("assessment.rationale", "invalid_value")
	}
	var rawText string
	if json.Unmarshal(raw, &rawText) != nil {
		return invalidMutationPayload("assessment.rationale", "invalid_value")
	}
	normalized, valid := fieldnorm.NormalizeNote(rawText)
	if !valid {
		return invalidMutationPayload("assessment.rationale", "invalid_value")
	}
	input.Rationale = normalized
	return nil
}

func decodeAssessor(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	parsed, valid := decodeCanonicalUUID(raw)
	if !valid {
		return invalidMutationPayload("assessment.assessor", "invalid_value")
	}
	input.Assessor = &parsed
	return nil
}

func decodeAssessedAt(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	var rawTime string
	if string(raw) == "null" || json.Unmarshal(raw, &rawTime) != nil {
		return invalidMutationPayload("assessment.assessed_at", "invalid_value")
	}
	assessedAt, valid := fieldnorm.NormalizeTimestampInstant(rawTime)
	if !valid {
		return invalidMutationPayload("assessment.assessed_at", "invalid_value")
	}
	input.AssessedAt = &assessedAt
	return nil
}

func decodeSupportReferences(raw json.RawMessage, input *assessments.CreateInput) *httpapi.APIError {
	refs, apiErr := decodeSupportRefs(raw)
	if apiErr != nil {
		return apiErr
	}
	input.SupportRefs = refs
	return nil
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
	if len(actions) > assessmentpolicy.MaxInitialSupportReferences {
		return nil, invalidMutationPayloadWithDetails(
			"assessment.support_refs.actions",
			"collection_action_count_exceeded",
			map[string]any{
				"field_key":       "assessment.support_refs",
				"requested_count": len(actions),
				"max_count":       assessmentpolicy.MaxInitialSupportReferences,
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
