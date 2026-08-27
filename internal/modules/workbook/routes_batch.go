package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

func (s *service) handleBulkMutations(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
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
	if _, apiErr := s.requireIncidentRole(
		r.Context(), incidentID, principal.User.ID,
		admission.RolesEditorReviewerAdmin, "editor|reviewer|admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.BulkFor(viewSchemaID)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}
	operation, failure, err := provider.DecodeBulk(r.Body)
	if apiErr := decodeMutationAPIError(operation.execute != nil, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := operation.Execute(r.Context(), BulkCommand{
		Actor: principal.User, IncidentID: incidentID, ViewSchemaID: viewSchemaID,
		RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}

func (s *service) handleClipboardPaste(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
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
	if _, apiErr := s.requireIncidentRole(
		r.Context(), incidentID, principal.User.ID,
		admission.RolesEditorReviewerAdmin, "editor|reviewer|admin",
	); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.ClipboardFor(viewSchemaID)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}
	operation, failure, err := provider.DecodeClipboard(r.Body)
	if apiErr := decodeMutationAPIError(operation.execute != nil, failure, err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, err := operation.Execute(r.Context(), ClipboardCommand{
		Actor: principal.User, IncidentID: incidentID, ViewSchemaID: viewSchemaID,
		RequestID: httpapi.RequestIDFromContext(r.Context()), Now: s.now(),
	})
	result, outcomeErr := resolveMutationOutcome(outcome, err)
	writeResolvedMutationResult(w, r, s, &principal, result, outcomeErr)
}
