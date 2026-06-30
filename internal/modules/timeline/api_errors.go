package timeline

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type MutationAPIErrorContext struct {
	ClientTxnID                 string
	IllegalTransitionReasonCode string
	NoEffectiveChangeField      string
}

// ClassifyMutationAPIError maps Timeline mutation errors to stable API envelopes.
func ClassifyMutationAPIError(err error, context MutationAPIErrorContext) (*auth.APIError, bool) {
	if err == nil {
		return nil, false
	}

	var sameFieldConflict *SameFieldConflictError
	var rowConflict *RowVersionConflictError
	var transitionErr *IllegalTransitionError

	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		return auth.ClientTxnConflictError(context.ClientTxnID), true
	case errors.Is(err, incidents.ErrIncidentClosed):
		return incidentClosedError(), true
	case errors.Is(err, ErrRecordNotFound):
		return incidentNotFoundError(), true
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		return recordDeletedUseRestoreError(), true
	case errors.As(err, &sameFieldConflict):
		return sameFieldConflictAPIError(sameFieldConflict), true
	case errors.As(err, &rowConflict):
		return rowVersionConflictError(rowConflict.Details()), true
	case errors.Is(err, ErrRowVersionConflict):
		return rowVersionConflictError(map[string]any{}), true
	case errors.As(err, &transitionErr):
		return illegalTransitionError(context.IllegalTransitionReasonCode, transitionErr), true
	case errors.Is(err, ErrIllegalTransition):
		return illegalTransitionError(context.IllegalTransitionReasonCode), true
	case errors.Is(err, ErrNoEffectiveChange):
		field := context.NoEffectiveChangeField
		if field == "" {
			field = "changes"
		}
		return invalidMutationPayload(field, "no_effective_change"), true
	default:
		return nil, false
	}
}

func recordDeletedUseRestoreError() *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusConflict,
		Code:    "record_deleted_use_restore",
		Message: "record deleted use restore",
		Details: map[string]any{},
	}
}

func sameFieldConflictAPIError(err *SameFieldConflictError) *auth.APIError {
	conflict := any(nil)
	if err != nil {
		conflict = err.Conflict
	}
	return &auth.APIError{
		Status:   http.StatusConflict,
		Code:     "same_field_conflict",
		Message:  "same field conflict",
		Details:  map[string]any{},
		Conflict: conflict,
	}
}
