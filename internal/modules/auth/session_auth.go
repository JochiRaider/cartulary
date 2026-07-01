package auth

import (
	"net/http"

	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

type SessionAuthStore = httpauth.SessionStore
type SessionRevoker = httpauth.SessionRevoker
type SessionAuthOptions = httpauth.Options

func AuthenticateSessionRequest(r *http.Request, options SessionAuthOptions) (SessionPrincipal, *APIError) {
	return httpauth.AuthenticateRequest(r, options)
}
