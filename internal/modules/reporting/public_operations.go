package reporting

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.reporting", http.MethodPost, "/api/v1/releases", "createRelease", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.reporting", http.MethodGet, "/api/v1/releases/{release_id}", "getRelease", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reporting", http.MethodPost, "/api/v1/releases/{release_id}/approve", "approveRelease", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reporting", http.MethodPost, "/api/v1/releases/{release_id}/invalidate", "invalidateRelease", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reporting", http.MethodPost, "/api/v1/releases/{release_id}/publish", "publishRelease", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reporting", http.MethodPost, "/api/v1/snapshots", "createSnapshot", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.reporting", http.MethodGet, "/api/v1/snapshots/{snapshot_id}", "getSnapshot", httpapi.PublicAuthenticationSession, false, http.StatusOK),
	}
}
