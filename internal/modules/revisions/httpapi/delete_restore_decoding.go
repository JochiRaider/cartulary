package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func decodeDeleteRestoreRequest(reader io.Reader) (revisions.DeleteRestoreRequest, *platformhttpapi.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return revisions.DeleteRestoreRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return revisions.DeleteRestoreRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request revisions.DeleteRestoreRequest
	if value, ok := raw["base_row_version"]; !ok {
		return revisions.DeleteRestoreRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return revisions.DeleteRestoreRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return revisions.DeleteRestoreRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return revisions.DeleteRestoreRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	reason, ok := normalizeOptionalReason(raw, "reason")
	if !ok {
		return revisions.DeleteRestoreRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	request.Reason = reason
	return request, nil
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *platformhttpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil || raw == nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func normalizeOptionalReason(raw map[string]json.RawMessage, field string) (*string, bool) {
	value, ok := raw[field]
	if !ok || string(value) == "null" {
		return nil, true
	}
	var input string
	if err := json.Unmarshal(value, &input); err != nil {
		return nil, false
	}
	normalized, ok := fieldnorm.NormalizeNote(input)
	if !ok {
		return nil, true
	}
	if len([]rune(normalized)) > 4096 {
		return nil, false
	}
	return &normalized, true
}

func invalidMutationPayload(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func forbiddenError(requiredRole string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: map[string]any{"required_role": requiredRole},
	}
}

func rowVersionConflictError(details map[string]any) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: details}
}

func clientTxnConflictError(clientTxnID string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "client_txn_conflict",
		Message: "client transaction conflict",
		Details: map[string]any{"client_txn_id": clientTxnID},
	}
}

func recordAlreadyDeletedError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "record_already_deleted", Details: map[string]any{}}
}

func recordDeleteBlockedError(details map[string]any) *platformhttpapi.APIError {
	if details == nil {
		details = map[string]any{}
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_delete_blocked",
		Message: "record delete blocked",
		Details: details,
	}
}

func recordRestoreBlockedError(details map[string]any) *platformhttpapi.APIError {
	if details == nil {
		details = map[string]any{}
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_restore_blocked",
		Message: "record restore blocked",
		Details: details,
	}
}

func recordNotDeletedError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "record_not_deleted", Details: map[string]any{}}
}

func recordLockedError(recordID string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "record_locked", Details: map[string]any{"record_id": recordID}}
}

func recordDeletedUseRestoreError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Details: map[string]any{}}
}

func rollbackTargetNotFoundError(target revisions.RollbackTarget) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusNotFound,
		Code:    "rollback_target_not_found",
		Message: "rollback target not found",
		Details: map[string]any{"target": target.Normalized()},
	}
}

func rollbackPreconditionFailedError(reasonCode string) *platformhttpapi.APIError {
	if reasonCode == "" {
		reasonCode = "target_not_reversible"
	}
	return &platformhttpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "rollback_precondition_failed",
		Message: "rollback precondition failed",
		Details: map[string]any{"reason_code": reasonCode},
	}
}
