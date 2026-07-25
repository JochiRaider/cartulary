package reportcomposition

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodGet, "/api/v1/incidents/{incident_id}/report-compositions", "listReportCompositions", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodPost, "/api/v1/incidents/{incident_id}/report-compositions", "createReportComposition", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodDelete, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}", "retireReportComposition", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodGet, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}", "getReportComposition", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodPatch, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}", "updateReportCompositionDraft", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodPost, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}/preview", "previewReportComposition", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodPost, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}/validate", "validateReportComposition", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodPost, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}/versions", "freezeReportCompositionVersion", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.reportcomposition", http.MethodGet, "/api/v1/incidents/{incident_id}/report-compositions/{composition_id}/versions/{composition_version}", "getReportCompositionVersion", httpapi.PublicAuthenticationSession, false, http.StatusOK),
	}
}
