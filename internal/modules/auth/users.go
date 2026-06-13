package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type UserCreateDefaults struct {
	MFARequired       bool
	IsDeploymentAdmin bool
}

type UserCreateRequest struct {
	ClientTxnID       string
	AuthKind          string
	Email             string
	DisplayName       string
	InitialPassword   string
	MFARequired       bool
	IsDeploymentAdmin bool
}

type UserPatchRequest struct {
	BaseUserVersion   int64
	Email             *string
	DisplayName       *string
	IsActive          *bool
	MFARequired       *bool
	IsDeploymentAdmin *bool
}

type AdminPasswordResetRequest struct {
	BaseUserVersion int64
	ClientTxnID     string
	NewPassword     string
	Reason          *string
}

type AdminTOTPResetRequest struct {
	BaseUserVersion int64
	ClientTxnID     string
	Reason          *string
}

type AdminRevokeAllRequest struct {
	ClientTxnID string
	Reason      *string
}

func ApplyUserCreateDefaults(mfaRequired *bool, isDeploymentAdmin *bool) UserCreateDefaults {
	defaults := UserCreateDefaults{
		MFARequired:       true,
		IsDeploymentAdmin: false,
	}
	if mfaRequired != nil {
		defaults.MFARequired = *mfaRequired
	}
	if isDeploymentAdmin != nil {
		defaults.IsDeploymentAdmin = *isDeploymentAdmin
	}
	return defaults
}

func BuildSafeUserResource(user authn.UserRecord) map[string]any {
	return BuildSafeUserResourceWithEnterpriseBindings(user, nil)
}

func BuildSafeUserResourceWithEnterpriseBindings(user authn.UserRecord, bindings []authn.EnterpriseAuthBindingSummary) map[string]any {
	return authn.SafeUserResponseWithEnterpriseBindings(user, bindings)
}

func WouldLeaveNoActiveDeploymentAdmins(currentIsAdmin bool, currentIsActive bool, activeAdminCount int, nextIsAdmin bool, nextIsActive bool) bool {
	if !(currentIsAdmin && currentIsActive) {
		return false
	}
	if nextIsAdmin && nextIsActive {
		return false
	}
	return activeAdminCount <= 1
}

func RequireDeploymentAdmin(user authn.UserRecord) *APIError {
	if user.IsDeploymentAdmin {
		return nil
	}
	return &APIError{Status: http.StatusUnauthorized, Code: unauthorizedCode, Details: map[string]any{}}
}

func DecodeUserCreateRequest(reader io.Reader) (UserCreateRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return UserCreateRequest{}, invalidMutationPayload("", "request_not_object")
	}

	allowed := map[string]struct{}{
		"client_txn_id":       {},
		"auth_kind":           {},
		"email":               {},
		"display_name":        {},
		"initial_password":    {},
		"mfa_required":        {},
		"is_deployment_admin": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return UserCreateRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	request := UserCreateRequest{}
	if value, ok := raw["client_txn_id"]; !ok {
		return UserCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return UserCreateRequest{}, invalidMutationPayload("client_txn_id", "missing_required_field")
	}

	if value, ok := raw["auth_kind"]; !ok {
		return UserCreateRequest{}, invalidMutationPayload("auth_kind", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.AuthKind); err != nil || request.AuthKind != "local" {
		return UserCreateRequest{}, invalidMutationPayload("auth_kind", "invalid_auth_kind")
	}

	if value, ok := raw["email"]; !ok {
		return UserCreateRequest{}, invalidMutationPayload("email", "missing_required_field")
	} else {
		var email string
		if err := json.Unmarshal(value, &email); err != nil {
			return UserCreateRequest{}, invalidMutationPayload("email", "field_not_nullable")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(email)
		if !ok {
			return UserCreateRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		request.Email = normalized
	}

	if value, ok := raw["display_name"]; !ok {
		return UserCreateRequest{}, invalidMutationPayload("display_name", "missing_required_field")
	} else {
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return UserCreateRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
		}
		normalized, ok := authn.NormalizeDisplayNameLine(displayName)
		if !ok {
			return UserCreateRequest{}, invalidMutationPayload("display_name", "invalid_display_name")
		}
		request.DisplayName = normalized
	}

	if value, ok := raw["initial_password"]; !ok {
		return UserCreateRequest{}, invalidMutationPayload("initial_password", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.InitialPassword); err != nil {
		return UserCreateRequest{}, invalidMutationPayload("initial_password", "field_not_nullable")
	}
	if _, err := authn.ValidatePasswordProvision(request.InitialPassword); err != nil {
		return UserCreateRequest{}, invalidMutationPayload("initial_password", "invalid_password")
	}

	var requestedMFARequired *bool
	if value, ok := raw["mfa_required"]; ok {
		var flag bool
		if string(value) == "null" {
			return UserCreateRequest{}, invalidMutationPayload("mfa_required", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &flag); err != nil {
			return UserCreateRequest{}, invalidMutationPayload("mfa_required", "field_not_nullable")
		}
		requestedMFARequired = &flag
	}

	var requestedIsDeploymentAdmin *bool
	if value, ok := raw["is_deployment_admin"]; ok {
		var flag bool
		if string(value) == "null" {
			return UserCreateRequest{}, invalidMutationPayload("is_deployment_admin", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &flag); err != nil {
			return UserCreateRequest{}, invalidMutationPayload("is_deployment_admin", "field_not_nullable")
		}
		requestedIsDeploymentAdmin = &flag
	}

	defaults := ApplyUserCreateDefaults(requestedMFARequired, requestedIsDeploymentAdmin)
	request.MFARequired = defaults.MFARequired
	request.IsDeploymentAdmin = defaults.IsDeploymentAdmin
	return request, nil
}

func DecodeUserPatchRequest(reader io.Reader) (UserPatchRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return UserPatchRequest{}, invalidMutationPayload("", "request_not_object")
	}

	allowed := map[string]struct{}{
		"base_user_version":   {},
		"email":               {},
		"display_name":        {},
		"is_active":           {},
		"mfa_required":        {},
		"is_deployment_admin": {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return UserPatchRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request UserPatchRequest
	if value, ok := raw["base_user_version"]; !ok {
		return UserPatchRequest{}, invalidMutationPayload("base_user_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseUserVersion); err != nil || request.BaseUserVersion < 1 {
		return UserPatchRequest{}, invalidMutationPayload("base_user_version", "invalid_base_user_version")
	}

	if value, ok := raw["email"]; ok {
		var email string
		if err := json.Unmarshal(value, &email); err != nil {
			return UserPatchRequest{}, invalidMutationPayload("email", "field_not_nullable")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(email)
		if !ok {
			return UserPatchRequest{}, invalidMutationPayload("email", "invalid_email")
		}
		request.Email = &normalized
	}
	if value, ok := raw["display_name"]; ok {
		var displayName string
		if err := json.Unmarshal(value, &displayName); err != nil {
			return UserPatchRequest{}, invalidMutationPayload("display_name", "field_not_nullable")
		}
		normalized, ok := authn.NormalizeDisplayNameLine(displayName)
		if !ok {
			return UserPatchRequest{}, invalidMutationPayload("display_name", "invalid_display_name")
		}
		request.DisplayName = &normalized
	}
	if value, ok := raw["is_active"]; ok {
		var flag bool
		if string(value) == "null" {
			return UserPatchRequest{}, invalidMutationPayload("is_active", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &flag); err != nil {
			return UserPatchRequest{}, invalidMutationPayload("is_active", "field_not_nullable")
		}
		request.IsActive = &flag
	}
	if value, ok := raw["mfa_required"]; ok {
		var flag bool
		if string(value) == "null" {
			return UserPatchRequest{}, invalidMutationPayload("mfa_required", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &flag); err != nil {
			return UserPatchRequest{}, invalidMutationPayload("mfa_required", "field_not_nullable")
		}
		request.MFARequired = &flag
	}
	if value, ok := raw["is_deployment_admin"]; ok {
		var flag bool
		if string(value) == "null" {
			return UserPatchRequest{}, invalidMutationPayload("is_deployment_admin", "field_not_nullable")
		}
		if err := json.Unmarshal(value, &flag); err != nil {
			return UserPatchRequest{}, invalidMutationPayload("is_deployment_admin", "field_not_nullable")
		}
		request.IsDeploymentAdmin = &flag
	}
	return request, nil
}

func DecodeAdminPasswordResetRequest(reader io.Reader) (AdminPasswordResetRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return AdminPasswordResetRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"base_user_version": {},
		"client_txn_id":     {},
		"new_password":      {},
		"reason":            {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return AdminPasswordResetRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request AdminPasswordResetRequest
	if value, ok := raw["base_user_version"]; !ok {
		return request, invalidMutationPayload("base_user_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseUserVersion); err != nil || request.BaseUserVersion < 1 {
		return request, invalidMutationPayload("base_user_version", "invalid_base_user_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["new_password"]; !ok {
		return request, invalidMutationPayload("new_password", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.NewPassword); err != nil {
		return request, invalidMutationPayload("new_password", "field_not_nullable")
	}
	if _, err := authn.ValidatePasswordProvision(request.NewPassword); err != nil {
		return request, invalidMutationPayload("new_password", "invalid_password")
	}
	if value, ok := raw["reason"]; ok {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return request, invalidMutationPayload("reason", "field_not_nullable")
		}
		request.Reason = authn.NormalizeReasonNote(&reason)
	}
	return request, nil
}

func DecodeAdminTOTPResetRequest(reader io.Reader) (AdminTOTPResetRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return AdminTOTPResetRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"base_user_version": {},
		"client_txn_id":     {},
		"reason":            {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return AdminTOTPResetRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request AdminTOTPResetRequest
	if value, ok := raw["base_user_version"]; !ok {
		return request, invalidMutationPayload("base_user_version", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.BaseUserVersion); err != nil || request.BaseUserVersion < 1 {
		return request, invalidMutationPayload("base_user_version", "invalid_base_user_version")
	}
	if value, ok := raw["client_txn_id"]; !ok {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return request, invalidMutationPayload("reason", "field_not_nullable")
		}
		request.Reason = authn.NormalizeReasonNote(&reason)
	}
	return request, nil
}

func DecodeAdminRevokeAllRequest(reader io.Reader) (AdminRevokeAllRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return AdminRevokeAllRequest{}, invalidMutationPayload("", "request_not_object")
	}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"reason":        {},
	}
	for key := range raw {
		if _, ok := allowed[key]; !ok {
			return AdminRevokeAllRequest{}, invalidMutationPayload(key, "unknown_field")
		}
	}

	var request AdminRevokeAllRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return request, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return request, invalidMutationPayload("reason", "field_not_nullable")
		}
		request.Reason = authn.NormalizeReasonNote(&reason)
	}
	return request, nil
}

func userPaginationError(reasonCode string) *APIError {
	return &APIError{
		Status: http.StatusBadRequest,
		Code:   "invalid_pagination_request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func invalidMutationPayload(field string, reasonCode string) *APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}
