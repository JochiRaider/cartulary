package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *service) handleLinkedNoteCreate(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{
		Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true,
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
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
		writeAPIError(w, r, incidentNotFoundError())
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
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{
		Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true,
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.ResolveRecordTarget(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	case target.LifecycleState == RecordLifecycleDeleted:
		writeAPIError(w, r, mutationFailureAPIError(RecordDeletedFailure()))
		return
	}
	if _, apiErr := s.requireIncidentRole(
		r.Context(), target.IncidentID, principal.User.ID,
		admission.RolesReviewerAdmin, "reviewer|admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.SupersedeFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, mutationFailureAPIError(IllegalTransitionFailure(
			"", "", "supersede_not_allowed", []string{"unsupported_record_type"},
		)))
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
