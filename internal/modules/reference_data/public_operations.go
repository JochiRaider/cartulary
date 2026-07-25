package reference_data

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.reference_data", http.MethodGet, "/api/v1/reference-packs", "listReferencePacks", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reference_data", http.MethodPost, "/api/v1/reference-packs/import", "importReferencePack", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.reference_data", http.MethodPost, "/api/v1/reference-packs/refresh", "refreshReferencePacks", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.reference_data", http.MethodGet, "/api/v1/reference-packs/{pack_key}/{pack_version}", "getReferencePackVersion", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.reference_data", http.MethodPost, "/api/v1/reference-packs/{pack_key}/{pack_version}/activate", "activateReferencePackVersion", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reference_data", http.MethodPost, "/api/v1/reference-packs/{pack_key}/{pack_version}/disable", "disableReferencePackVersion", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.reference_data", http.MethodPost, "/api/v1/reference-packs/{pack_key}/{pack_version}/reverify", "reverifyReferencePackVersion", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
	}
}
