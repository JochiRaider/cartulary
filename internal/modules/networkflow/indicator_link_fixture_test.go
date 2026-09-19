package networkflow

import (
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"net/http"
)

// Exercise the same object framing and admission sequence as the HTTP adapter.
func decodeIndicatorLinkFixture(r *http.Request, limits EffectiveLimits) (indicatorLinkRequest, *httpapi.APIError) {
	raw, apiErr := decodeNetworkFlowObject(r.Body)
	if apiErr != nil {
		return indicatorLinkRequest{}, semanticHTTPError(completeLinkError(apiErr, indicatorLinkRequest{}, ""))
	}
	request, failure := decodeIndicatorLinkObject(raw, limits)
	return request, semanticHTTPError(failure)
}
