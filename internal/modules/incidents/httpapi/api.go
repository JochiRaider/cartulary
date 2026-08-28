package httpapi

import (
	"errors"
	"io"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func admitIncidentCreateJSON(reader io.Reader) (incidents.IncidentCreateAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitIncidentCreateJSON(reader)
	return admission, mapAdmissionError(admissionErr, invalidIncidentCreate)
}

func admitIncidentPatchJSON(reader io.Reader) (incidents.IncidentPatchAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitIncidentPatchJSON(reader)
	return admission, mapAdmissionError(admissionErr, invalidIncidentPatch)
}

func admitIncidentLifecycleJSON(
	action incidents.LifecycleAction,
	reader io.Reader,
) (incidents.IncidentLifecycleAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitIncidentLifecycleJSON(action, reader)
	return admission, mapAdmissionError(admissionErr, invalidIncidentLifecycleRequest)
}

func admitMembershipCreateJSON(reader io.Reader) (incidents.MembershipCreateAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitMembershipCreateJSON(reader)
	return admission, mapAdmissionError(admissionErr, invalidMutationPayload)
}

func admitMembershipPatchJSON(reader io.Reader) (incidents.MembershipPatchAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitMembershipPatchJSON(reader)
	return admission, mapAdmissionError(admissionErr, invalidMutationPayload)
}

func admitMembershipDeleteJSON(reader io.Reader) (incidents.MembershipDeleteAdmission, *platformhttpapi.APIError) {
	admission, admissionErr := incidents.AdmitMembershipDeleteJSON(reader)
	return admission, mapAdmissionError(admissionErr, invalidMutationPayload)
}

func mapAdmissionError(
	admissionErr *incidents.AdmissionError,
	invalid func(string, string) *platformhttpapi.APIError,
) *platformhttpapi.APIError {
	if admissionErr == nil {
		return nil
	}
	field, _ := admissionErr.Field()
	return invalid(field, admissionErr.ReasonCode())
}

func invalidIncidentCreate(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_incident_create",
		Message: "invalid incident create request", Details: details,
	}
}

func invalidIncidentPatch(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_incident_patch",
		Message: "invalid incident patch request", Details: details,
	}
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
		Status: http.StatusBadRequest, Code: "invalid_mutation_payload",
		Message: "invalid mutation payload", Details: details,
	}
}

func invalidIncidentLifecycleRequest(field string, reasonCode string) *platformhttpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &platformhttpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_incident_lifecycle_request",
		Message: "invalid incident lifecycle request", Details: details,
	}
}

func invalidPaginationRequest(reasonCode string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_pagination_request",
		Message: "invalid pagination request", Details: map[string]any{"reason_code": reasonCode},
	}
}

func invalidListQuery(reasonCode string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status: http.StatusBadRequest, Code: "invalid_list_query",
		Message: "invalid list query", Details: map[string]any{"reason_code": reasonCode},
	}
}

func incidentNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func userNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}}
}

func incidentKeyConflictError(conflict *incidents.IncidentKeyConflictError) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status: http.StatusConflict, Code: "incident_key_conflict",
		Details: map[string]any{"field": "incident_key", "incident_key_canonical": conflict.IncidentKeyCanonical()},
	}
}

func incidentVersionConflictError(conflict *incidents.IncidentVersionConflictError) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "incident_version_conflict", Details: conflict.Details()}
}

func incidentClosedError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{},
	}
}

func incidentIllegalTransitionError(action incidents.LifecycleAction) *platformhttpapi.APIError {
	var reasonCode string
	switch action {
	case incidents.LifecycleActionClose:
		reasonCode = "incident_already_closed"
	case incidents.LifecycleActionReopen:
		reasonCode = "incident_not_closed"
	default:
		return internalAPIError(errors.New("incidents http: invalid lifecycle action"))
	}
	return &platformhttpapi.APIError{
		Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition",
		Details: map[string]any{"reason_code": reasonCode},
	}
}

func membershipNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "membership_not_found", Details: map[string]any{}}
}

func membershipExistsUsePatchError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "membership_exists_use_patch", Details: map[string]any{}}
}

func membershipVersionConflictError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "membership_version_conflict", Details: map[string]any{}}
}

func lastIncidentAdminError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "last_incident_admin", Details: map[string]any{}}
}

func userInactiveError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "user_inactive", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *platformhttpapi.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &platformhttpapi.APIError{
		Status: http.StatusForbidden, Code: "authorization_denied",
		Message: "authorization denied", Details: details,
	}
}
