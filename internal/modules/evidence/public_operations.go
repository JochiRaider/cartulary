package evidence

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.evidence", http.MethodPost, "/api/v1/evidence-records/{record_id}/attach-blob", "attachBlobToEvidenceRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.evidence", http.MethodPost, "/api/v1/evidence-records/{record_id}/download-handle", "issueEvidenceDownloadHandle", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.evidence", http.MethodPost, "/api/v1/evidence-records/{record_id}/preview-handle", "issueEvidencePreviewHandle", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.evidence", http.MethodGet, "/api/v1/evidence-handles/{handle_token}", "redeemEvidenceHandle", httpapi.PublicAuthenticationPublic, false, http.StatusOK),
		httpapi.NewPublicOperation("module.evidence", http.MethodPost, "/api/v1/object-blobs", "createObjectBlobSlot", httpapi.PublicAuthenticationSession, true, http.StatusCreated),
		httpapi.NewPublicOperation("module.evidence", http.MethodPut, "/api/v1/object-uploads/{upload_token}", "uploadObjectBlobContent", httpapi.PublicAuthenticationPublic, true, http.StatusNoContent),
	}
}
