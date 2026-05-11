package revisions

import (
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

const (
	deleteRouteKey  = "records.delete"
	restoreRouteKey = "records.restore"
)

type DeleteRestoreRequest struct {
	BaseRowVersion int64
	ClientTxnID    string
	Reason         *string
}

func DecodeDeleteRestoreRequest(reader io.Reader) (DeleteRestoreRequest, *auth.APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return DeleteRestoreRequest{}, apiErr
	}
	allowed := map[string]struct{}{
		"base_row_version": {},
		"client_txn_id":    {},
		"reason":           {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return DeleteRestoreRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request DeleteRestoreRequest
	if value, ok := raw["base_row_version"]; !ok {
		return DeleteRestoreRequest{}, invalidMutationPayload("base_row_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseRowVersion); err != nil || request.BaseRowVersion < 1 {
		return DeleteRestoreRequest{}, invalidMutationPayload("base_row_version", "invalid_base_row_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return DeleteRestoreRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return DeleteRestoreRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	reason, ok := normalizeOptionalReason(raw, "reason")
	if !ok {
		return DeleteRestoreRequest{}, invalidMutationPayload("reason", "invalid_value")
	}
	request.Reason = reason
	return request, nil
}

func DeleteRestoreRequestHash(request DeleteRestoreRequest) []byte {
	payload := map[string]any{
		"base_row_version": request.BaseRowVersion,
		"client_txn_id":    request.ClientTxnID,
		"reason":           stringOrNil(request.Reason),
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	hash := make([]byte, len(sum))
	copy(hash, sum[:])
	return hash
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *auth.APIError) {
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

func stringOrNil(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func forbiddenError(requiredRole string) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: map[string]any{"required_role": requiredRole},
	}
}

func rowVersionConflictError(details map[string]any) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: details}
}

func clientTxnConflictError(clientTxnID string) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusConflict,
		Code:    "client_txn_conflict",
		Message: "client transaction conflict",
		Details: map[string]any{"client_txn_id": clientTxnID},
	}
}

func recordAlreadyDeletedError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "record_already_deleted", Details: map[string]any{}}
}

func recordNotDeletedError() *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "record_not_deleted", Details: map[string]any{}}
}

func recordLockedError(recordID string) *auth.APIError {
	return &auth.APIError{Status: http.StatusConflict, Code: "record_locked", Details: map[string]any{"record_id": recordID}}
}
