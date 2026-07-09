package auth

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func invalidEnterpriseAuthRequest(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{"reason_code": reasonCode}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_enterprise_auth_request", Details: details}
}

func enterpriseTransactionRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "enterprise_auth_transaction_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func providerResponseRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_response_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func providerIdentityRejected(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "provider_identity_rejected", Details: map[string]any{"reason_code": reasonCode}}
}

func authBindingConflict(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "auth_binding_conflict", Details: map[string]any{"reason_code": reasonCode}}
}
