package workbook

import (
	"context"
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	return s.requireIncidentRole(ctx, incidentID, userID, admission.RolesMember, "")
}

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, incidentNotFoundError()
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Message: "authorization denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfPersistenceDue(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithOptions(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, httpapi.ErrorOptions{
		Conflict: apiErr.Conflict, Retryable: apiErr.Retryable,
	})
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	value, err := uuid.Parse(raw)
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}

func internalAPIError(err error) *httpapi.APIError {
	_ = err
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "internal_error",
		Details: map[string]any{},
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func isRecordTargetNotFound(err error) bool {
	return errors.Is(err, ErrRecordTargetNotFound)
}

func decodeMutationAPIError(operationPresent bool, failure *MutationFailure, err error) *httpapi.APIError {
	if err != nil || (failure != nil && operationPresent) {
		return internalAPIError(err)
	}
	if failure != nil {
		return mutationFailureAPIError(failure)
	}
	if !operationPresent {
		return internalAPIError(errors.New("workbook provider returned no operation or failure"))
	}
	return nil
}

func resolveMutationOutcome(outcome MutationOutcome, err error) (MutationResult, *httpapi.APIError) {
	if err != nil {
		return MutationResult{}, internalAPIError(err)
	}
	if validationErr := outcome.Validate(); validationErr != nil {
		return MutationResult{}, internalAPIError(validationErr)
	}
	if failure, rejected := outcome.Failure(); rejected {
		return MutationResult{}, mutationFailureAPIError(failure)
	}
	result, ok := outcome.Result()
	if !ok {
		return MutationResult{}, internalAPIError(errors.New("workbook provider returned no mutation result"))
	}
	return result, nil
}

func writeResolvedMutationResult(
	w http.ResponseWriter,
	r *http.Request,
	s *service,
	principal *httpauth.Principal,
	result MutationResult,
	apiErr *httpapi.APIError,
) {
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func entityMatchConflictError(entityType string, identifierClass string, candidateRecordIDs []uuid.UUID) *httpapi.APIError {
	details := map[string]any{
		"reason_code":      "merge_required",
		"entity_type":      entityType,
		"identifier_class": identifierClass,
	}
	if len(candidateRecordIDs) > 0 {
		ids := make([]string, 0, len(candidateRecordIDs))
		for _, recordID := range candidateRecordIDs {
			ids = append(ids, recordID.String())
		}
		details["candidate_record_ids"] = ids
	}
	return &httpapi.APIError{Status: http.StatusConflict, Code: "entity_match_conflict", Message: "entity match conflict", Details: details}
}
