package viewschemas

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("platform.viewschema", http.MethodGet, "/api/v1/view-schemas", "listViewSchemas", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("platform.viewschema", http.MethodGet, "/api/v1/view-schemas/{view_schema_id}", "getViewSchema", httpapi.PublicAuthenticationSession, false, http.StatusOK),
	}
}
