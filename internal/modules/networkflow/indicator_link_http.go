package networkflow

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/google/uuid"
	"net/http"
)

func (s *routeService) handleIndicatorLinks(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.indicator_links.create"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, incidentID := admitted.principal, admitted.incidentID
	raw, failure := decodeNetworkFlowObject(r.Body)
	apiErr = semanticHTTPError(completeLinkError(failure, indicatorLinkRequest{}, ""))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, failure := decodeIndicatorLinkObject(raw, s.limits)
	apiErr = semanticHTTPError(failure)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	payload, status, apiErr := s.commitIndicatorLinkRoute(r.Context(), incidentID, principal.User, request, indicatorLinkRequestHash(request), httpapi.RequestIDFromContext(r.Context()))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, status, payload)
}
func (s *routeService) commitIndicatorLinkRoute(ctx context.Context, incident uuid.UUID, actor authn.UserRecord, request indicatorLinkRequest, hash []byte, requestID string) (map[string]any, int, *httpapi.APIError) {
	outcome, failure := s.links.execute(ctx, incident, actor, request, hash, requestID)
	if failure != nil {
		return nil, 0, semanticHTTPError(failure)
	}
	payload, status := indicatorLinkOutcomeResponse(outcome)
	return payload, status, nil
}
