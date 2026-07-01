package httpapi

import (
	"net/http"
	"net/url"
)

type APIError struct {
	Status    int
	Code      string
	Message   string
	Details   map[string]any
	Conflict  any
	Retryable bool
}

func WriteAPIError(w http.ResponseWriter, r *http.Request, apiErr *APIError) {
	if apiErr == nil {
		apiErr = InternalAPIError(nil)
	}
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func InternalAPIError(err error) *APIError {
	message := "internal_error"
	if err != nil {
		message = err.Error()
	}
	return &APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: message,
		Details: map[string]any{},
	}
}

func SessionRequiredError() *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
}

func CSRFVerificationFailedError() *APIError {
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    "csrf_verification_failed",
		Message: "csrf proof is required for cookie-authenticated state-changing requests",
		Details: map[string]any{},
	}
}

func AuthorizationDeniedCapabilityError(capability string) *APIError {
	details := map[string]any{}
	if capability != "" {
		details["required_capability"] = capability
	}
	return &APIError{
		Status:  http.StatusForbidden,
		Code:    "authorization_denied",
		Message: "authorization denied",
		Details: details,
	}
}

func ClientTxnConflictError(clientTxnID string) *APIError {
	return &APIError{
		Status:  http.StatusConflict,
		Code:    "client_txn_conflict",
		Message: "client transaction conflicts with an existing request",
		Details: map[string]any{
			"client_txn_id": clientTxnID,
		},
	}
}

func ValidateSingletonReadQuery(query url.Values) *APIError {
	for _, key := range []string{"limit", "cursor_token", "page", "offset", "page_size", "block_size"} {
		if _, ok := query[key]; ok {
			return &APIError{
				Status: http.StatusBadRequest,
				Code:   "invalid_pagination_request",
				Details: map[string]any{
					"reason_code": "pagination_not_supported",
				},
			}
		}
	}
	return nil
}
