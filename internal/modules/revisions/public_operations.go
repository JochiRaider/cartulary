package revisions

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.revisions", http.MethodDelete, "/api/v1/records/{record_id}", "deleteRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.revisions", http.MethodGet, "/api/v1/records/{record_id}/history", "getRecordHistory", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.revisions", http.MethodPost, "/api/v1/records/{record_id}/restore", "restoreRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.revisions", http.MethodPost, "/api/v1/records/{record_id}/rollback", "rollbackRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
