package reference_data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func DecodeImportMetadata(envelope httpapi.UploadEnvelope) (ImportMetadataRequest, *httpapi.APIError) {
	if len(envelope.MetadataRaw) != 0 {
		if _, err := canonicaljson.DecodeStrict(envelope.MetadataRaw); err != nil {
			return ImportMetadataRequest{}, invalidReferencePackRequest("metadata", "malformed_metadata_json")
		}
	}
	allowed := map[string]struct{}{
		"client_txn_id":     {},
		"activation_policy": {},
	}
	for key := range envelope.Metadata {
		if _, ok := allowed[key]; !ok {
			return ImportMetadataRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(envelope.Metadata, "client_txn_id")
	if apiErr != nil {
		return ImportMetadataRequest{}, apiErr
	}
	activationPolicy := "staged_only"
	if value, ok := envelope.Metadata["activation_policy"]; ok {
		if bytesEqualJSONNull(value) {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "field_not_nullable")
		}
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "invalid_activation_policy")
		}
		if parsed != "staged_only" {
			return ImportMetadataRequest{}, invalidReferencePackRequest("activation_policy", "auto_activation_not_supported")
		}
		activationPolicy = parsed
	}
	normalized, err := json.Marshal(map[string]any{
		"activation_policy": activationPolicy,
		"file_sha256":       envelope.FileSHA256Hex,
	})
	if err != nil {
		return ImportMetadataRequest{}, internalAPIError(err)
	}
	return ImportMetadataRequest{
		ClientTxnID:      clientTxnID,
		ActivationPolicy: activationPolicy,
		Normalized:       normalized,
	}, nil
}

func DecodeActionRequest(reader io.Reader) (ActionRequest, *httpapi.APIError) {
	return decodeActionRequest(reader, false)
}

func DecodeRemovalRequest(reader io.Reader) (ActionRequest, *httpapi.APIError) {
	return decodeActionRequest(reader, true)
}

func decodeActionRequest(reader io.Reader, requireReason bool) (ActionRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"reason":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return ActionRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id")
	if apiErr != nil {
		return ActionRequest{}, apiErr
	}
	var reason *string
	if value, ok := raw["reason"]; requireReason && (!ok || bytesEqualJSONNull(value)) {
		code := "missing_required_field"
		if ok {
			code = "field_not_nullable"
		}
		return ActionRequest{}, invalidReferencePackRequest("reason", code)
	}
	if value, ok := raw["reason"]; ok && !bytesEqualJSONNull(value) {
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return ActionRequest{}, invalidReferencePackRequest("reason", "invalid_reason")
		}
		if normalized, ok := fieldnorm.NormalizeNote(parsed); ok {
			if utf8.RuneCountInString(normalized) > 4096 {
				return ActionRequest{}, invalidReferencePackRequest("reason", "reason_too_long")
			}
			reason = &normalized
		} else if strings.TrimSpace(parsed) != "" {
			return ActionRequest{}, invalidReferencePackRequest("reason", "invalid_reason")
		}
	}
	if requireReason && reason == nil {
		return ActionRequest{}, invalidReferencePackRequest("reason", "invalid_reason")
	}
	normalized, err := json.Marshal(map[string]any{
		"client_txn_id": clientTxnID,
		"reason":        optionalString(reason),
	})
	if err != nil {
		return ActionRequest{}, internalAPIError(err)
	}
	return ActionRequest{ClientTxnID: clientTxnID, Reason: reason, Normalized: normalized}, nil
}

func DecodeRefreshRequest(reader io.Reader) (RefreshRequest, *httpapi.APIError) {
	raw, apiErr := decodeJSONObject(reader)
	if apiErr != nil {
		return RefreshRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"pack_keys":     {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return RefreshRequest{}, invalidReferencePackRequest(key, "unknown_field")
		}
	}
	clientTxnID, apiErr := requiredStringField(raw, "client_txn_id")
	if apiErr != nil {
		return RefreshRequest{}, apiErr
	}
	request := RefreshRequest{ClientTxnID: clientTxnID}
	if value, ok := raw["pack_keys"]; ok {
		request.PackKeysProvided = true
		if bytesEqualJSONNull(value) {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "field_not_nullable")
		}
		var values []any
		if err := json.Unmarshal(value, &values); err != nil {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "invalid_pack_keys")
		}
		if len(values) == 0 {
			return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "empty_pack_keys")
		}
		seen := map[string]struct{}{}
		for _, value := range values {
			packKey, ok := value.(string)
			if !ok || strings.TrimSpace(packKey) == "" {
				return RefreshRequest{}, invalidReferencePackRequest("pack_keys", "invalid_pack_keys")
			}
			seen[packKey] = struct{}{}
		}
		request.PackKeys = make([]string, 0, len(seen))
		for packKey := range seen {
			request.PackKeys = append(request.PackKeys, packKey)
		}
		sort.Strings(request.PackKeys)
	}
	return request, nil
}

func decodeJSONObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	data, err := io.ReadAll(io.LimitReader(reader, MaxAdministrativeRequestBytes+1))
	if err != nil {
		return nil, invalidReferencePackRequest("request", "request_not_object")
	}
	if len(data) > MaxAdministrativeRequestBytes {
		return nil, invalidReferencePackRequest("request", "request_too_large")
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, invalidReferencePackRequest("request", "request_not_object")
	}
	if _, ok := value.(map[string]any); !ok {
		return nil, invalidReferencePackRequest("request", "request_not_object")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_reference_pack_request", Details: map[string]any{"reason_code": "request_not_object"}}
	}
	if raw == nil {
		return nil, &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_reference_pack_request", Details: map[string]any{"reason_code": "request_not_object"}}
	}
	return raw, nil
}

func requiredStringField(raw map[string]json.RawMessage, field string) (string, *httpapi.APIError) {
	value, ok := raw[field]
	if !ok {
		return "", invalidReferencePackRequest(field, "missing_required_field")
	}
	if bytesEqualJSONNull(value) {
		return "", invalidReferencePackRequest(field, "field_not_nullable")
	}
	var parsed string
	if err := json.Unmarshal(value, &parsed); err != nil || strings.TrimSpace(parsed) == "" {
		return "", invalidReferencePackRequest(field, "missing_required_field")
	}
	return parsed, nil
}

func invalidReferencePackRequest(field string, reasonCode string) *httpapi.APIError {
	// Never reflect hostile object member names into diagnostics.
	switch field {
	case "client_txn_id", "reason", "pack_keys", "activation_policy", "metadata", "request":
	default:
		field = "request"
	}
	return &httpapi.APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_reference_pack_request",
		Details: map[string]any{
			"field":       field,
			"reason_code": reasonCode,
		},
	}
}

func uploadEnvelopeAPIError(apiErr *httpapi.UploadEnvelopeError) *httpapi.APIError {
	if apiErr == nil {
		return nil
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_reference_pack_request",
		Message: fmt.Sprintf("invalid reference pack request: %s", apiErr.ReasonCode),
		Details: apiErr.Details(),
	}
}

func referencePackNotFound() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "reference_pack_not_found", Details: map[string]any{}}
}

func referencePackActivationRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "reference_pack_activation_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func clientTxnConflict(clientTxnID string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": clientTxnID}}
}

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func invalidListQuery(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_list_query",
		Message: "invalid list query",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "reference pack operation could not be completed",
		Details: map[string]any{},
	}
}

func coordinatorAPIError(err error) *httpapi.APIError {
	if rejection, ok := ImportContentRejection(err); ok && rejection.Summary != nil {
		return &httpapi.APIError{Status: http.StatusConflict, Code: "reference_pack_verification_failed", Details: map[string]any{"reason_code": rejection.Code, "check_id": rejection.CheckID, "validation_summary": rejection.Summary}}
	}
	var rejected *OperationRejection
	if errors.As(err, &rejected) {
		return &httpapi.APIError{Status: http.StatusConflict, Code: "reference_pack_operation_rejected", Details: map[string]any{"reason_code": rejected.Reason}}
	}
	var request *RequestRejection
	if errors.As(err, &request) {
		return invalidReferencePackRequest(request.Field, request.Reason)
	}
	var activation *ActivationRejection
	if errors.As(err, &activation) {
		return referencePackActivationRejected(activation.Reason)
	}
	return internalAPIError(err)
}

func bytesEqualJSONNull(value json.RawMessage) bool {
	return strings.TrimSpace(string(value)) == "null"
}
