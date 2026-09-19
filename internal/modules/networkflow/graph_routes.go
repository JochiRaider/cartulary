package networkflow

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func (s *routeService) handleGraphQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.graphs.query"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	request, apiErr := decodeGraphQueryRequest(r, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	composition, failure := s.graphQueries.execute(r.Context(), admitted.identity(), request, httpapi.RequestIDFromContext(r.Context()))
	apiErr = semanticHTTPError(failure)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, graphQueryResultResource(composition))
}

func (s *routeService) handleGraphContributorsQuery(w http.ResponseWriter, r *http.Request) {
	routeID := "nf.graphs.contributors.query"
	admitted, apiErr := s.admitOperation(r, routeID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal := admitted.principal
	request, apiErr := decodeGraphContributorQueryRequest(r, s.limits)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	outcome, failure := s.graphQueries.contributors(r.Context(), admitted.identity(), request)
	apiErr = semanticHTTPError(failure)
	result := graphContributorsResource(outcome)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, httpapi.InternalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result)
}

func decodeGraphQueryRequest(r *http.Request, limits EffectiveLimits) (graphQueryRequest, *httpapi.APIError) {
	value, failure := decodeGraphQueryRequestValue(r.Body, limits)
	return value, semanticHTTPError(failure)
}

func decodeGraphContributorQueryRequest(r *http.Request, limits EffectiveLimits) (graphContributorQueryRequest, *httpapi.APIError) {
	value, failure := decodeGraphContributorQueryRequestValue(r.Body, limits)
	return value, semanticHTTPError(failure)
}
