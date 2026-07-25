package extensiondiscovery

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.extensions", http.MethodGet, "/api/v1/extensions", "listDeploymentExtensions", httpapi.PublicAuthenticationSession, false, http.StatusOK),
	}
}
