package timelineadmission

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
)

type MutationAPIErrorContext struct {
	ClientTxnID                 string
	IllegalTransitionReasonCode string
	NoEffectiveChangeField      string
}

// ClassifyMutationAPIError maps Timeline mutation errors to stable API envelopes.
func ClassifyMutationAPIError(err error, context MutationAPIErrorContext) (*httpapi.APIError, bool) {
	if err == nil {
		return nil, false
	}

	var sameFieldConflict *timeline.SameFieldConflictError
	var rowConflict *timeline.RowVersionConflictError
	var transitionErr *timeline.IllegalTransitionError

	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		return httpapi.ClientTxnConflictError(context.ClientTxnID), true
	case errors.Is(err, timeline.ErrIncidentClosed):
		return incidentClosedError(), true
	case errors.Is(err, timeline.ErrRecordNotFound):
		return incidentNotFoundError(), true
	case errors.Is(err, timeline.ErrRecordDeleted):
		return recordDeletedUseRestoreError(), true
	case errors.As(err, &sameFieldConflict):
		return sameFieldConflictAPIError(sameFieldConflict), true
	case errors.As(err, &rowConflict):
		return rowVersionConflictError(rowConflict.Details()), true
	case errors.Is(err, timeline.ErrRowVersionConflict):
		return rowVersionConflictError(map[string]any{}), true
	case errors.As(err, &transitionErr):
		return illegalTransitionError(context.IllegalTransitionReasonCode, transitionErr), true
	case errors.Is(err, timeline.ErrIllegalTransition):
		return illegalTransitionError(context.IllegalTransitionReasonCode), true
	case errors.Is(err, timeline.ErrNoEffectiveChange):
		field := context.NoEffectiveChangeField
		if field == "" {
			field = "changes"
		}
		return invalidMutationPayload(field, "no_effective_change"), true
	default:
		return nil, false
	}
}

func recordDeletedUseRestoreError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_deleted_use_restore",
		Message: "record deleted use restore",
		Details: map[string]any{},
	}
}

func sameFieldConflictAPIError(err *timeline.SameFieldConflictError) *httpapi.APIError {
	conflict := any(nil)
	if err != nil {
		conflict = err.Conflict
	}
	return &httpapi.APIError{
		Status:   http.StatusConflict,
		Code:     "same_field_conflict",
		Message:  "same field conflict",
		Details:  map[string]any{},
		Conflict: conflict,
	}
}

func invalidViewQuery(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidViewQueryValidation(err *viewquery.ValidationError) *httpapi.APIError {
	if err == nil {
		return invalidViewQuery("", "")
	}
	details := map[string]any{}
	if err.Field != "" {
		details["field"] = err.Field
	}
	if err.FieldKey != "" {
		details["field_key"] = err.FieldKey
	}
	if err.FilterIndex != nil {
		details["filter_index"] = *err.FilterIndex
	}
	if err.ReasonCode != "" {
		details["reason_code"] = err.ReasonCode
	}
	if err.RequestedCount != nil {
		details["requested_count"] = *err.RequestedCount
	}
	if err.MaxCount != nil {
		details["max_count"] = *err.MaxCount
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	return invalidMutationPayloadWithDetails(field, reasonCode, nil)
}

func invalidMutationPayloadWithDetails(field string, reasonCode string, extra map[string]any) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	for key, value := range extra {
		details[key] = value
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func rowVersionConflictError(details ...map[string]any) *httpapi.APIError {
	payload := map[string]any{}
	if len(details) > 0 && details[0] != nil {
		payload = details[0]
	}
	return &httpapi.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: payload}
}

func illegalTransitionError(reasonCode string, sourceErr ...error) *httpapi.APIError {
	details := map[string]any{}
	var transitionErr *timeline.IllegalTransitionError
	for _, err := range sourceErr {
		if errors.As(err, &transitionErr) {
			break
		}
	}
	if transitionErr != nil {
		if transitionErr.ReasonCode != "" {
			reasonCode = transitionErr.ReasonCode
		}
		details["from_status"] = transitionErr.FromStatus
		details["to_status"] = transitionErr.ToStatus
		details["violated_guards"] = append([]string{}, transitionErr.ViolatedGuards...)
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: details,
	}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
