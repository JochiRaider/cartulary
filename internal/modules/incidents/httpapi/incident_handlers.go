package httpapi

import (
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
)

func (s *service) handleListVisibleIncidents(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	listScope, apiErr := parseIncidentListScope(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(
		listScope.Values,
		"incidents.list",
		principal.User.ID.String(),
		listScope.Scope,
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}

	pageRequest, reasonCode := s.incidentListPageRequest(binding, cursor)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	records, listErr := s.application.ListVisibleIncidents(r.Context(), principal.User.ID, pageRequest)
	if listErr != nil {
		writeAPIError(w, r, internalAPIError(listErr))
		return
	}
	anchor := s.now().UTC()
	if pageRequest.AnchorUpdatedAt != nil {
		anchor = *pageRequest.AnchorUpdatedAt
	} else if len(records) > 0 {
		anchor = records[0].UpdatedAt.UTC()
	}
	rows, nextCursor, err := buildIncidentListPage(binding, anchor, records)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var nextToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextToken = &token
	}
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"incidents": rows}, platformhttpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *service) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := admitIncidentCreateJSON(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.application.CreateIncident(r.Context(), principal.User, request, platformhttpapi.RequestIDFromContext(r.Context()))
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID()))
		return
	case errors.Is(err, incidents.ErrIncidentKeyConflict):
		var conflict *incidents.IncidentKeyConflictError
		if !errors.As(err, &conflict) {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		writeAPIError(w, r, incidentKeyConflictError(conflict))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.Header().Set("Location", incidentLocation(result.Incident.ID))
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	_ = platformhttpapi.WriteSuccess(w, r, status, result.Payload)
}

func (s *service) handleGetIncident(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.GetVisibleIncident(r.Context(), incidentID, principal.User.ID)
	if errors.Is(err, incidents.ErrIncidentNotFound) {
		writeAPIError(w, r, platformhttpapi.NotFound("incident_not_found"))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	// REQ-01-707 step 5: the singleton-query check is route handling, after the
	// incident's visibility (REQ-01-704).
	if apiErr := platformhttpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildIncidentResource(record))
}

func (s *service) handlePatchIncident(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, apiErr := admittedPathUUID(r, "incident_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesReviewerAdmin, "reviewer|admin")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_ = membership
	request, apiErr := admitIncidentPatchJSON(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, _, err := s.application.UpdateIncident(r.Context(), principal.User, incidentID, request, platformhttpapi.RequestIDFromContext(r.Context()))
	var versionConflict *incidents.IncidentVersionConflictError
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		writeAPIError(w, r, platformhttpapi.NotFound("incident_not_found"))
		return
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		writeAPIError(w, r, authorizationDeniedError("reviewer|admin"))
		return
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, incidents.ErrIncidentNotFound):
		writeAPIError(w, r, platformhttpapi.NotFound("incident_not_found"))
		return
	case errors.As(err, &versionConflict):
		writeAPIError(w, r, incidentVersionConflictError(versionConflict))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildIncidentResource(record))
}

func incidentLocation(incidentID uuid.UUID) string {
	return "/api/v1/incidents/" + incidentID.String()
}
