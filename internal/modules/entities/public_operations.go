package entities

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func PublicOperations() []httpapi.PublicOperation {
	return []httpapi.PublicOperation{
		httpapi.NewPublicOperation("module.entities", http.MethodPost, "/api/v1/entity-mentions/{entity_mention_id}/resolve", "resolveEntityMention", httpapi.PublicAuthenticationSession, true, http.StatusOK),
		httpapi.NewPublicOperation("module.entities", http.MethodPost, "/api/v1/records/{survivor_record_id}/merge", "mergeEntityRecord", httpapi.PublicAuthenticationSession, true, http.StatusOK),
	}
}
