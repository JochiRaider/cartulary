package jobapi

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.jobapi", http.MethodGet, "/api/v1/jobs/{job_id}", "getJob", httpapi.PublicAuthenticationSession, false, http.StatusOK),
		httpapi.NewPublicOperation("module.jobapi", http.MethodPost, "/api/v1/jobs/{job_id}/cancel", "cancelJob", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
