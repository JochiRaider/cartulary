package httpapi

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type observationCreateRequest struct {
	ClientTxnID               string     `json:"client_txn_id"`
	BaseRowVersion            int64      `json:"base_row_version"`
	SourceFieldKey            string     `json:"source_field_key"`
	SpanStartByte             int        `json:"span_start_byte"`
	SpanEndByte               int        `json:"span_end_byte"`
	ParsedIndicatorType       *string    `json:"parsed_indicator_type,omitempty"`
	ResolvedIndicatorRecordID *uuid.UUID `json:"resolved_indicator_record_id,omitempty"`
}

type observationResolveRequest struct {
	ClientTxnID               string    `json:"client_txn_id"`
	BaseRowVersion            int64     `json:"base_row_version"`
	ResolvedIndicatorRecordID uuid.UUID `json:"resolved_indicator_record_id"`
}

type observationActionRequest struct {
	ClientTxnID    string `json:"client_txn_id"`
	BaseRowVersion int64  `json:"base_row_version"`
}

type lifecycleAppendRequest struct {
	ClientTxnID    string      `json:"client_txn_id"`
	BaseRowVersion int64       `json:"base_row_version"`
	LifecycleState string      `json:"lifecycle_state"`
	ValidFrom      time.Time   `json:"valid_from"`
	ValidTo        *time.Time  `json:"valid_to"`
	Confidence     *int        `json:"confidence"`
	Rationale      *string     `json:"rationale"`
	SupportRefs    []uuid.UUID `json:"support_refs"`
	Assessor       *string     `json:"assessor"`
}

func decodeObservationCreate(reader io.Reader) (observationCreateRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeClosedObject(reader, []string{
		"client_txn_id", "base_row_version", "source_field_key", "span_start_byte", "span_end_byte",
		"parsed_indicator_type", "resolved_indicator_record_id",
	}, []string{"client_txn_id", "base_row_version", "source_field_key", "span_start_byte", "span_end_byte"})
	if apiErr != nil {
		return observationCreateRequest{}, apiErr
	}
	var request observationCreateRequest
	if !decodeRequired(raw, "client_txn_id", &request.ClientTxnID) || !validClientTxnID(request.ClientTxnID) {
		return observationCreateRequest{}, invalidMutationPayload("client_txn_id", "invalid_value")
	}
	if !decodeRequired(raw, "base_row_version", &request.BaseRowVersion) || request.BaseRowVersion < 1 {
		return observationCreateRequest{}, invalidMutationPayload("base_row_version", "invalid_value")
	}
	if !decodeRequired(raw, "source_field_key", &request.SourceFieldKey) || request.SourceFieldKey == "" || strings.TrimSpace(request.SourceFieldKey) != request.SourceFieldKey || strings.ContainsRune(request.SourceFieldKey, 0) {
		return observationCreateRequest{}, invalidMutationPayload("source_field_key", "invalid_value")
	}
	if !decodeRequired(raw, "span_start_byte", &request.SpanStartByte) || request.SpanStartByte < 0 {
		return observationCreateRequest{}, invalidMutationPayload("span_start_byte", "invalid_value")
	}
	if !decodeRequired(raw, "span_end_byte", &request.SpanEndByte) || request.SpanEndByte <= request.SpanStartByte {
		return observationCreateRequest{}, invalidMutationPayload("span_end_byte", "invalid_value")
	}
	if value, ok := raw["parsed_indicator_type"]; ok {
		if isJSONNull(value) || json.Unmarshal(value, &request.ParsedIndicatorType) != nil || request.ParsedIndicatorType == nil || !validIndicatorType(*request.ParsedIndicatorType) {
			return observationCreateRequest{}, invalidMutationPayload("parsed_indicator_type", "invalid_value")
		}
	}
	if value, ok := raw["resolved_indicator_record_id"]; ok {
		parsed, valid := decodeCanonicalUUID(value)
		if !valid {
			return observationCreateRequest{}, invalidMutationPayload("resolved_indicator_record_id", "invalid_value")
		}
		request.ResolvedIndicatorRecordID = &parsed
	}
	return request, nil
}

func decodeObservationResolve(reader io.Reader) (observationResolveRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeClosedObject(reader,
		[]string{"client_txn_id", "base_row_version", "resolved_indicator_record_id"},
		[]string{"client_txn_id", "base_row_version", "resolved_indicator_record_id"},
	)
	if apiErr != nil {
		return observationResolveRequest{}, apiErr
	}
	var request observationResolveRequest
	if !decodeRequired(raw, "client_txn_id", &request.ClientTxnID) || !validClientTxnID(request.ClientTxnID) {
		return observationResolveRequest{}, invalidMutationPayload("client_txn_id", "invalid_value")
	}
	if !decodeRequired(raw, "base_row_version", &request.BaseRowVersion) || request.BaseRowVersion < 1 {
		return observationResolveRequest{}, invalidMutationPayload("base_row_version", "invalid_value")
	}
	parsed, valid := decodeCanonicalUUID(raw["resolved_indicator_record_id"])
	if !valid {
		return observationResolveRequest{}, invalidMutationPayload("resolved_indicator_record_id", "invalid_value")
	}
	request.ResolvedIndicatorRecordID = parsed
	return request, nil
}

func decodeObservationAction(reader io.Reader) (observationActionRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeClosedObject(reader, []string{"client_txn_id", "base_row_version"}, []string{"client_txn_id", "base_row_version"})
	if apiErr != nil {
		return observationActionRequest{}, apiErr
	}
	var request observationActionRequest
	if !decodeRequired(raw, "client_txn_id", &request.ClientTxnID) || !validClientTxnID(request.ClientTxnID) {
		return observationActionRequest{}, invalidMutationPayload("client_txn_id", "invalid_value")
	}
	if !decodeRequired(raw, "base_row_version", &request.BaseRowVersion) || request.BaseRowVersion < 1 {
		return observationActionRequest{}, invalidMutationPayload("base_row_version", "invalid_value")
	}
	return request, nil
}

func decodeLifecycleAppend(reader io.Reader) (lifecycleAppendRequest, *platformhttpapi.APIError) {
	fields := []string{"client_txn_id", "base_row_version", "lifecycle_state", "valid_from", "valid_to", "confidence", "rationale", "support_refs", "assessor"}
	raw, apiErr := decodeClosedObject(reader, fields, fields)
	if apiErr != nil {
		return lifecycleAppendRequest{}, apiErr
	}
	var request lifecycleAppendRequest
	if !decodeRequired(raw, "client_txn_id", &request.ClientTxnID) || !validClientTxnID(request.ClientTxnID) {
		return lifecycleAppendRequest{}, invalidMutationPayload("client_txn_id", "invalid_value")
	}
	if !decodeRequired(raw, "base_row_version", &request.BaseRowVersion) || request.BaseRowVersion < 1 {
		return lifecycleAppendRequest{}, invalidMutationPayload("base_row_version", "invalid_value")
	}
	if !decodeRequired(raw, "lifecycle_state", &request.LifecycleState) || !validLifecycleState(request.LifecycleState) {
		return lifecycleAppendRequest{}, invalidMutationPayload("lifecycle_state", "invalid_value")
	}
	validFrom, valid := decodeCanonicalTimestamp(raw["valid_from"])
	if !valid {
		return lifecycleAppendRequest{}, invalidMutationPayload("valid_from", "invalid_value")
	}
	request.ValidFrom = validFrom
	if !isJSONNull(raw["valid_to"]) {
		validTo, valid := decodeCanonicalTimestamp(raw["valid_to"])
		if !valid || validTo.Before(validFrom) {
			return lifecycleAppendRequest{}, invalidMutationPayload("valid_to", "invalid_value")
		}
		request.ValidTo = &validTo
	}
	if !isJSONNull(raw["confidence"]) {
		var confidence int
		if json.Unmarshal(raw["confidence"], &confidence) != nil || confidence < 0 || confidence > 100 {
			return lifecycleAppendRequest{}, invalidMutationPayload("confidence", "invalid_value")
		}
		request.Confidence = &confidence
	}
	for field, target := range map[string]**string{"rationale": &request.Rationale, "assessor": &request.Assessor} {
		if isJSONNull(raw[field]) {
			continue
		}
		var value string
		if json.Unmarshal(raw[field], &value) != nil || value == "" || strings.ContainsRune(value, 0) {
			return lifecycleAppendRequest{}, invalidMutationPayload(field, "invalid_value")
		}
		*target = &value
	}
	var supportText []string
	if json.Unmarshal(raw["support_refs"], &supportText) != nil || supportText == nil || len(supportText) > 64 {
		return lifecycleAppendRequest{}, invalidMutationPayload("support_refs", "invalid_value")
	}
	seen := map[uuid.UUID]struct{}{}
	for _, text := range supportText {
		parsed, valid := parseCanonicalUUID(text)
		if !valid {
			return lifecycleAppendRequest{}, invalidMutationPayload("support_refs", "invalid_value")
		}
		if _, duplicate := seen[parsed]; duplicate {
			return lifecycleAppendRequest{}, invalidMutationPayload("support_refs", "duplicate_value")
		}
		seen[parsed] = struct{}{}
		request.SupportRefs = append(request.SupportRefs, parsed)
	}
	return request, nil
}

func decodeClosedObject(reader io.Reader, allowed []string, required []string) (map[string]json.RawMessage, *platformhttpapi.APIError) {
	raw, err := platformhttpapi.DecodeStrictJSONObject(reader)
	if err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, field := range allowed {
		allowedSet[field] = struct{}{}
	}
	for field := range raw {
		if _, ok := allowedSet[field]; !ok {
			return nil, invalidMutationPayload(field, "unknown_field")
		}
	}
	for _, field := range required {
		if _, ok := raw[field]; !ok {
			return nil, invalidMutationPayload(field, "missing_required_field")
		}
	}
	return raw, nil
}

func decodeRequired(raw map[string]json.RawMessage, field string, target any) bool {
	value, ok := raw[field]
	return ok && !isJSONNull(value) && json.Unmarshal(value, target) == nil
}

func decodeCanonicalUUID(raw json.RawMessage) (uuid.UUID, bool) {
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return uuid.Nil, false
	}
	return parseCanonicalUUID(text)
}

func parseCanonicalUUID(text string) (uuid.UUID, bool) {
	parsed, err := uuid.Parse(text)
	return parsed, err == nil && parsed != uuid.Nil && parsed.String() == text
}

func decodeCanonicalTimestamp(raw json.RawMessage) (time.Time, bool) {
	var text string
	if json.Unmarshal(raw, &text) != nil || text == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, text)
	if err != nil {
		return time.Time{}, false
	}
	canonical := parsed.UTC().Format(time.RFC3339Nano)
	return parsed.UTC(), text == canonical
}

func validClientTxnID(value string) bool {
	return value != "" && strings.TrimSpace(value) != "" && !strings.ContainsRune(value, 0)
}

func validIndicatorType(value string) bool {
	switch value {
	case "ipv4_addr", "ipv6_addr", "domain_name", "url", "sha256", "email_addr", "registry_key", "process_name", "text":
		return true
	default:
		return false
	}
}

func validLifecycleState(value string) bool {
	switch value {
	case "active", "benign", "false_positive", "retired":
		return true
	default:
		return false
	}
}

func isJSONNull(value json.RawMessage) bool {
	return string(value) == "null"
}

func requestHash(value any) []byte {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return digest[:]
}

func invalidMutationPayload(field string, reason string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reason != "" {
		details["reason_code"] = reason
	}
	return &platformhttpapi.APIError{Status: 400, Code: "invalid_mutation_payload", Message: "invalid mutation payload", Details: details}
}

func createParams(incidentID uuid.UUID, sourceID uuid.UUID, request observationCreateRequest, requestID string) indicators.IndicatorObservationCreateParams {
	return indicators.IndicatorObservationCreateParams{
		IncidentID: incidentID, SourceRecordID: sourceID, BaseRowVersion: request.BaseRowVersion,
		SourceFieldKey: request.SourceFieldKey, SpanStartByte: request.SpanStartByte, SpanEndByte: request.SpanEndByte,
		ParsedIndicatorType: request.ParsedIndicatorType, ResolvedIndicatorRecordID: request.ResolvedIndicatorRecordID,
		ClientTxnID: request.ClientTxnID, RequestID: requestID, RequestHash: requestHash(request),
	}
}
