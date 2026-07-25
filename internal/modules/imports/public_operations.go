package imports

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.imports", http.MethodPost, "/api/v1/import-sessions", "createImportSession", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.imports", http.MethodGet, "/api/v1/import-sessions/{import_session_id}", "getImportSession", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodPost, "/api/v1/import-sessions/{import_session_id}/apply", "applyImportSession", httpapi.PublicAuthenticationSession, true, http.StatusAccepted),
		httpapi.NewPublicOperation("module.imports", http.MethodGet, "/api/v1/import-sessions/{import_session_id}/units", "listImportUnits", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodGet, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}", "getImportUnit", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodPut, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping", "putImportUnitMapping", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodPost, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/mapping-preview", "previewImportUnitExtensionMapping", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodGet, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/preview", "getImportUnitPreview", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodPost, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/select", "selectImportUnit", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.imports", http.MethodPost, "/api/v1/import-sessions/{import_session_id}/units/{import_unit_id}/skip", "skipImportUnit", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
