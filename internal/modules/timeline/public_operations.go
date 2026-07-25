package timeline

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.timeline", http.MethodGet, "/api/v1/incidents/{incident_id}/timeline-time-conversion-profile", "getTimelineTimeConversionProfile", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.timeline", http.MethodPut, "/api/v1/incidents/{incident_id}/timeline-time-conversion-profile", "putTimelineTimeConversionProfile", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.timeline", http.MethodPost, "/api/v1/records/{record_id}/mark-reviewed", "markTimelineRecordReviewed", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
