package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *service) handleCreate(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.CreateFor(viewSchemaID)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}
	mutationAdmission, failure, err := provider.DecodeCreate(r.Body)
	if apiErr := decodeMutationAPIError(mutationAdmission, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), viewSchemaID, "create")
	outcome, err := provider.Create(ctx, CreateCommand{
		Actor:        principal.User,
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Admission:    mutationAdmission,
		RequestID:    httpapi.RequestIDFromContext(ctx),
		Now:          s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	telemetryResult, telemetryErrorCode := workbookAPIErrorTelemetry(outcomeErr)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}

func (s *service) handlePatch(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
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
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.PatchFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	mutationAdmission, failure, err := provider.DecodePatch(r.Body)
	if apiErr := decodeMutationAPIError(mutationAdmission, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), mutationAdmission.AdmittedViewSchemaID(), "patch")
	outcome, err := provider.Patch(ctx, PatchCommand{
		Actor:                   principal.User,
		RecordID:                recordID,
		AuthoritativeRecordType: target.RecordType,
		Admission:               mutationAdmission,
		RequestID:               httpapi.RequestIDFromContext(ctx),
		Now:                     s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	telemetryResult, telemetryErrorCode := workbookAPIErrorTelemetry(outcomeErr)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}

func (s *service) handleConflictResolve(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
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
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	token := r.PathValue("conflict_token")
	claims, valid := s.conflictTokens.DecodeConflictToken(token)
	if !valid || claims.RecordID != recordID {
		writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
		return
	}
	provider, registered := s.contributions.ConflictFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
		return
	}
	mutationAdmission, failure, err := provider.DecodeConflict(r.Body, token, claims)
	if apiErr := decodeMutationAPIError(mutationAdmission, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := provider.ResolveConflict(r.Context(), ConflictCommand{
		Actor:                   principal.User,
		RecordID:                recordID,
		AuthoritativeRecordType: target.RecordType,
		Claims:                  claims,
		Admission:               mutationAdmission,
		RequestID:               httpapi.RequestIDFromContext(r.Context()),
		Now:                     s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}
