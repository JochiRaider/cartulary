package workbook

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.workbook", http.MethodGet, "/api/v1/incidents/{incident_id}/workbook-preferences/default", "getIncidentDefaultWorkbookPreferences", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPut, "/api/v1/incidents/{incident_id}/workbook-preferences/default", "putIncidentDefaultWorkbookPreferences", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodGet, "/api/v1/incidents/{incident_id}/workbook-preferences/me", "getCurrentUserWorkbookPreferences", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPut, "/api/v1/incidents/{incident_id}/workbook-preferences/me", "putCurrentUserWorkbookPreferences", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodGet, "/api/v1/incidents/{incident_id}/workbook-startup", "getIncidentWorkbookStartup", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations", "applyWorkbookBulkMutation", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste", "pasteWorkbookClipboard", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/incidents/{incident_id}/views/{view_schema_id}/query", "queryWorkbookView", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", "createViewRow", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.workbook", http.MethodPatch, "/api/v1/records/{record_id}", "patchRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/records/{record_id}/conflicts/{conflict_token}/resolve", "resolveRecordSameFieldConflict", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/records/{record_id}/linked-notes", "createRecordLinkedNote", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.workbook", http.MethodPost, "/api/v1/records/{record_id}/supersede", "supersedeRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
