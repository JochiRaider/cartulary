package auth

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

var sixDigitCodePattern = regexp.MustCompile(`^[0-9]{6}$`)

type APIError = httpapi.APIError
type AuthSource = httpauth.AuthSource

const (
	AuthSourceCookie = httpauth.AuthSourceCookie
	AuthSourceBearer = httpauth.AuthSourceBearer
)

type LoginRequest struct {
	Username     string
	Password     string
	SecondFactor *SecondFactorAssertion
}

type SecondFactorAssertion struct {
	Kind string
	Code string
}

type PasswordChangeRequest struct {
	ClientTxnID     string
	CurrentPassword string
	NewPassword     string
	SecondFactor    *SecondFactorAssertion
}

type TOTPBeginRequest struct {
	ClientTxnID     string
	CurrentPassword *string
	SecondFactor    *SecondFactorAssertion
}

type TOTPCompleteRequest struct {
	ClientTxnID  string
	EnrollmentID string
	Code         string
}

func DecodeLoginRequest(reader io.Reader) (LoginRequest, *APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return LoginRequest{}, invalidAuthRequest("", "request body must be one JSON object")
	}

	request := LoginRequest{}
	allowed := map[string]struct{}{
		"username":      {},
		"password":      {},
		"second_factor": {},
	}

	for key := range raw {
		if _, ok := allowed[key]; ok {
			continue
		}
		return LoginRequest{}, invalidAuthRequest(key, "unknown or forbidden top-level member")
	}

	if value, ok := raw["username"]; !ok {
		return LoginRequest{}, invalidAuthRequest("username", "missing username")
	} else {
		var username string
		if err := json.Unmarshal(value, &username); err != nil {
			return LoginRequest{}, invalidAuthRequest("username", "username must be a non-null string")
		}
		normalized, _, ok := authn.NormalizeEmailAddress(username)
		if !ok {
			return LoginRequest{}, invalidAuthRequest("username", "username must satisfy email_address_v1")
		}
		request.Username = normalized
	}

	if value, ok := raw["password"]; !ok {
		return LoginRequest{}, invalidAuthRequest("password", "missing password")
	} else {
		var password string
		if err := json.Unmarshal(value, &password); err != nil || password == "" {
			return LoginRequest{}, invalidAuthRequest("password", "password must be a non-empty string")
		}
		request.Password = password
	}

	if value, ok := raw["second_factor"]; ok {
		var secondFactor map[string]json.RawMessage
		if err := json.Unmarshal(value, &secondFactor); err != nil {
			return LoginRequest{}, invalidAuthRequest("second_factor", "second_factor must be an object")
		}
		for key := range secondFactor {
			switch key {
			case "kind", "assertion":
			default:
				return LoginRequest{}, invalidAuthRequest("second_factor."+key, "unknown second_factor member")
			}
		}

		var kind string
		if value, ok := secondFactor["kind"]; !ok {
			return LoginRequest{}, invalidAuthRequest("second_factor.kind", "missing second_factor.kind")
		} else if err := json.Unmarshal(value, &kind); err != nil {
			return LoginRequest{}, invalidAuthRequest("second_factor.kind", "second_factor.kind must be a string")
		}
		if kind != "totp" {
			return LoginRequest{}, invalidAuthRequest("second_factor.kind", "unsupported second-factor kind")
		}

		value, ok := secondFactor["assertion"]
		if !ok {
			return LoginRequest{}, invalidAuthRequest("second_factor.assertion", "missing second_factor.assertion")
		}
		var assertion map[string]json.RawMessage
		if err := json.Unmarshal(value, &assertion); err != nil {
			return LoginRequest{}, invalidAuthRequest("second_factor.assertion", "second_factor.assertion must be an object")
		}
		for key := range assertion {
			if key != "code" {
				return LoginRequest{}, invalidAuthRequest("second_factor.assertion."+key, "unknown assertion member")
			}
		}

		value, ok = assertion["code"]
		if !ok {
			return LoginRequest{}, invalidAuthRequest("second_factor.assertion.code", "missing second_factor.assertion.code")
		}
		var code string
		if err := json.Unmarshal(value, &code); err != nil || !sixDigitCodePattern.MatchString(code) {
			return LoginRequest{}, invalidAuthRequest("second_factor.assertion.code", "totp code must be exactly six ASCII decimal digits")
		}

		request.SecondFactor = &SecondFactorAssertion{
			Kind: kind,
			Code: code,
		}
	}

	return request, nil
}

func ValidateSingletonReadQuery(query url.Values) *APIError {
	return httpapi.ValidateSingletonReadQuery(query)
}

func ShouldSlideIdleExpiry(method string, path string) bool {
	return httpauth.ShouldSlideIdleExpiry(method, path)
}

const SessionSlideWriteInterval = httpauth.SessionSlideWriteInterval

func ShouldPersistIdleExpirySlide(timing authn.SessionTiming, now time.Time) bool {
	return httpauth.ShouldPersistIdleExpirySlide(timing, now)
}

func AllowsBootstrapTokenRoute(path string) bool {
	switch path {
	case "/api/v1/auth/mfa/totp/begin", "/api/v1/auth/mfa/totp/complete":
		return true
	default:
		return false
	}
}

func BuildCredentialStateResource(user authn.UserRecord, pendingExpiresAt *time.Time) map[string]any {
	totpState := "not_enrolled"
	if pendingExpiresAt != nil {
		totpState = "pending"
	}
	if user.TOTPEnrolledAt != nil {
		totpState = "active"
	}

	return map[string]any{
		"user_id":        user.ID,
		"auth_kind":      "local",
		"mfa_required":   user.MFARequired,
		"recovery_model": "admin_assisted",
		"password": map[string]any{
			"changed_at": user.PasswordChangedAt,
		},
		"totp": map[string]any{
			"state":              totpState,
			"enrolled_at":        user.TOTPEnrolledAt,
			"pending_expires_at": pendingExpiresAt,
		},
	}
}

func BuildTOTPSetup(secretBase32 string, username string) map[string]any {
	label := "Cartulary:" + username
	return map[string]any{
		"secret_base32":  secretBase32,
		"otpauth_uri":    "otpauth://totp/" + url.PathEscape(label) + "?secret=" + url.QueryEscape(secretBase32) + "&issuer=Cartulary&algorithm=SHA1&digits=6&period=30",
		"algorithm":      "SHA1",
		"digits":         6,
		"period_seconds": 30,
	}
}

func ShouldRevokeSessionsOnTOTPComplete(replacesActive bool) bool {
	return replacesActive
}

func ValidateCSRF(method string, authSource AuthSource, cookieValue string, headerValue string) *APIError {
	return httpauth.ValidateCSRF(method, authSource, cookieValue, headerValue)
}

func BootstrapRejectedError(reasonCode string) *APIError {
	return httpauth.BootstrapRejectedError(reasonCode)
}

func TOTPSetupNotPendingError(reasonCode string) *APIError {
	return &APIError{
		Status:  http.StatusConflict,
		Code:    "totp_setup_not_pending",
		Message: "totp setup is not pending",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func ClientTxnConflictError(clientTxnID string) *APIError {
	return httpapi.ClientTxnConflictError(clientTxnID)
}

func DecodePasswordChangeRequest(reader io.Reader) (PasswordChangeRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return PasswordChangeRequest{}, apiErr
	}

	request := PasswordChangeRequest{}
	allowed := map[string]struct{}{
		"client_txn_id":    {},
		"current_password": {},
		"new_password":     {},
		"second_factor":    {},
	}
	for key := range raw {
		if _, ok := allowed[key]; ok {
			continue
		}
		return PasswordChangeRequest{}, invalidAuthRequest(key, "unknown top-level member")
	}

	if value, ok := raw["client_txn_id"]; !ok {
		return PasswordChangeRequest{}, invalidAuthRequest("client_txn_id", "missing client_txn_id")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return PasswordChangeRequest{}, invalidAuthRequest("client_txn_id", "client_txn_id must be a non-empty string")
	}

	if value, ok := raw["current_password"]; !ok {
		return PasswordChangeRequest{}, invalidAuthRequest("current_password", "missing current_password")
	} else if err := json.Unmarshal(value, &request.CurrentPassword); err != nil || request.CurrentPassword == "" {
		return PasswordChangeRequest{}, invalidAuthRequest("current_password", "current_password must be a non-empty string")
	}

	if value, ok := raw["new_password"]; !ok {
		return PasswordChangeRequest{}, invalidAuthRequest("new_password", "missing new_password")
	} else if err := json.Unmarshal(value, &request.NewPassword); err != nil {
		return PasswordChangeRequest{}, invalidAuthRequest("new_password", "new_password must be a string")
	}
	if _, err := authn.ValidatePasswordProvision(request.NewPassword); err != nil {
		return PasswordChangeRequest{}, invalidAuthRequest("new_password", err.Error())
	}

	secondFactor, apiErr := decodeOptionalSecondFactor(raw["second_factor"])
	if apiErr != nil {
		return PasswordChangeRequest{}, apiErr
	}
	request.SecondFactor = secondFactor
	return request, nil
}

func DecodeTOTPBeginRequest(reader io.Reader) (TOTPBeginRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return TOTPBeginRequest{}, apiErr
	}

	request := TOTPBeginRequest{}
	allowed := map[string]struct{}{
		"client_txn_id":    {},
		"current_password": {},
		"second_factor":    {},
	}
	for key := range raw {
		if _, ok := allowed[key]; ok {
			continue
		}
		return TOTPBeginRequest{}, invalidAuthRequest(key, "unknown top-level member")
	}

	if value, ok := raw["client_txn_id"]; !ok {
		return TOTPBeginRequest{}, invalidAuthRequest("client_txn_id", "missing client_txn_id")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return TOTPBeginRequest{}, invalidAuthRequest("client_txn_id", "client_txn_id must be a non-empty string")
	}

	if value, ok := raw["current_password"]; ok {
		var password string
		if err := json.Unmarshal(value, &password); err != nil || password == "" {
			return TOTPBeginRequest{}, invalidAuthRequest("current_password", "current_password must be a non-empty string")
		}
		request.CurrentPassword = &password
	}

	secondFactor, apiErr := decodeOptionalSecondFactor(raw["second_factor"])
	if apiErr != nil {
		return TOTPBeginRequest{}, apiErr
	}
	request.SecondFactor = secondFactor
	return request, nil
}

func DecodeTOTPCompleteRequest(reader io.Reader) (TOTPCompleteRequest, *APIError) {
	raw, apiErr := decodeObject(reader)
	if apiErr != nil {
		return TOTPCompleteRequest{}, apiErr
	}

	request := TOTPCompleteRequest{}
	allowed := map[string]struct{}{
		"client_txn_id": {},
		"enrollment_id": {},
		"code":          {},
	}
	for key := range raw {
		if _, ok := allowed[key]; ok {
			continue
		}
		return TOTPCompleteRequest{}, invalidAuthRequest(key, "unknown top-level member")
	}

	if value, ok := raw["client_txn_id"]; !ok {
		return TOTPCompleteRequest{}, invalidAuthRequest("client_txn_id", "missing client_txn_id")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return TOTPCompleteRequest{}, invalidAuthRequest("client_txn_id", "client_txn_id must be a non-empty string")
	}
	if value, ok := raw["enrollment_id"]; !ok {
		return TOTPCompleteRequest{}, invalidAuthRequest("enrollment_id", "missing enrollment_id")
	} else if err := json.Unmarshal(value, &request.EnrollmentID); err != nil || strings.TrimSpace(request.EnrollmentID) == "" {
		return TOTPCompleteRequest{}, invalidAuthRequest("enrollment_id", "enrollment_id must be a non-empty string")
	}
	if value, ok := raw["code"]; !ok {
		return TOTPCompleteRequest{}, invalidAuthRequest("code", "missing code")
	} else if err := json.Unmarshal(value, &request.Code); err != nil || !sixDigitCodePattern.MatchString(request.Code) {
		return TOTPCompleteRequest{}, invalidAuthRequest("code", "code must be exactly six ASCII decimal digits")
	}

	return request, nil
}

func invalidAuthRequest(field string, message string) *APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	return &APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_auth_request",
		Message: message,
		Details: details,
	}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidAuthRequest("", "request body must be one JSON object")
	}
	return raw, nil
}

func decodeOptionalSecondFactor(value json.RawMessage) (*SecondFactorAssertion, *APIError) {
	if len(value) == 0 {
		return nil, nil
	}

	var secondFactor map[string]json.RawMessage
	if err := json.Unmarshal(value, &secondFactor); err != nil {
		return nil, invalidAuthRequest("second_factor", "second_factor must be an object")
	}
	for key := range secondFactor {
		switch key {
		case "kind", "assertion":
		default:
			return nil, invalidAuthRequest("second_factor."+key, "unknown second_factor member")
		}
	}

	var kind string
	if value, ok := secondFactor["kind"]; !ok {
		return nil, invalidAuthRequest("second_factor.kind", "missing second_factor.kind")
	} else if err := json.Unmarshal(value, &kind); err != nil {
		return nil, invalidAuthRequest("second_factor.kind", "second_factor.kind must be a string")
	}
	if kind != "totp" {
		return nil, invalidAuthRequest("second_factor.kind", "unsupported second-factor kind")
	}

	value, ok := secondFactor["assertion"]
	if !ok {
		return nil, invalidAuthRequest("second_factor.assertion", "missing second_factor.assertion")
	}
	var assertion map[string]json.RawMessage
	if err := json.Unmarshal(value, &assertion); err != nil {
		return nil, invalidAuthRequest("second_factor.assertion", "second_factor.assertion must be an object")
	}
	for key := range assertion {
		if key != "code" {
			return nil, invalidAuthRequest("second_factor.assertion."+key, "unknown assertion member")
		}
	}

	value, ok = assertion["code"]
	if !ok {
		return nil, invalidAuthRequest("second_factor.assertion.code", "missing second_factor.assertion.code")
	}
	var code string
	if err := json.Unmarshal(value, &code); err != nil || !sixDigitCodePattern.MatchString(code) {
		return nil, invalidAuthRequest("second_factor.assertion.code", "totp code must be exactly six ASCII decimal digits")
	}

	return &SecondFactorAssertion{Kind: kind, Code: code}, nil
}
