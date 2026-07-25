package savedviews

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.savedviews", http.MethodGet, "/api/v1/incidents/{incident_id}/saved-views", "listIncidentSavedViews", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.savedviews", http.MethodPost, "/api/v1/incidents/{incident_id}/saved-views", "createIncidentSavedView", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.savedviews", http.MethodDelete, "/api/v1/incidents/{incident_id}/saved-views/{saved_view_id}", "deleteIncidentSavedView", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.savedviews", http.MethodPatch, "/api/v1/incidents/{incident_id}/saved-views/{saved_view_id}", "patchIncidentSavedView", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
