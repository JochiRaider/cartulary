package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *service) handleLinkedNoteCreate(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, apiErr := admittedPathUUID(r, "record_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(
		r.Context(), target.IncidentID, principal.User.ID,
		admission.RolesEditorReviewerAdmin, "editor|reviewer|admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.LinkedNoteFor(target.RecordType)
	if !registered {
		// REQ-01-705: a soft-deleted record is classified before its type. No
		// linked-note receipt can exist for an unsupported type, so there is no
		// replay to return first.
		if target.LifecycleState == RecordLifecycleDeleted {
			writeAPIError(w, r, mutationFailureAPIError(RecordDeletedFailure()))
			return
		}
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	}
	operation, failure, err := provider.DecodeLinkedNote(r.Body)
	if apiErr := decodeMutationAPIError(operation.execute != nil, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := operation.Execute(r.Context(), LinkedNoteCommand{
		Actor: principal.User, Target: target,
		RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}

func (s *service) handleSupersede(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, apiErr := admittedPathUUID(r, "record_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	// REQ-01-704: visibility and role precede the deleted-state classification,
	// so a caller who cannot see the incident cannot tell a deleted record from
	// a missing one.
	if _, apiErr := s.requireIncidentRole(
		r.Context(), target.IncidentID, principal.User.ID,
		admission.RolesReviewerAdmin, "reviewer|admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if target.LifecycleState == RecordLifecycleDeleted {
		writeAPIError(w, r, mutationFailureAPIError(RecordDeletedFailure()))
		return
	}
	provider, registered := s.contributions.SupersedeFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, mutationFailureAPIError(UnsupportedLifecycleActionFailure("superseded")))
		return
	}
	operation, failure, err := provider.DecodeSupersede(r.Body)
	if apiErr := decodeMutationAPIError(operation.execute != nil, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := operation.Execute(r.Context(), SupersedeCommand{
		Actor: principal.User, Target: target,
		RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}
