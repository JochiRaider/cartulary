package incidentbundles

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.incidentbundles", http.MethodPost, "/api/v1/incident-bundles/export", "exportIncidentBundle", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.incidentbundles", http.MethodPost, "/api/v1/incident-bundles/import", "importIncidentBundle", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.incidentbundles", http.MethodGet, "/api/v1/incident-bundles/{bundle_id}", "getIncidentBundleDescriptor", httpapi.PublicAuthenticationSession, false, http.StatusOK),
	}
}
