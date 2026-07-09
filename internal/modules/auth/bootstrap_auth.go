package auth

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

type bootstrapCredentialStore interface {
	GetBootstrapTokenByFingerprint(context.Context, []byte) (authn.BootstrapTokenRecord, authn.UserRecord, error)
}

type bootstrapCredentialAuthenticator struct {
	store bootstrapCredentialStore
	keys  authn.MasterKeys
	now   func() time.Time
}

type bootstrapCredentialResult struct {
	Token     authn.BootstrapTokenRecord
	TokenText string
	User      authn.UserRecord
}

func (a bootstrapCredentialAuthenticator) Authenticate(r *http.Request, token string, allowBootstrap bool) (bootstrapCredentialResult, bool, *httpapi.APIError) {
	bootstrapToken, user, err := a.store.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(a.keys, token))
	if errors.Is(err, authn.ErrNotFound) {
		return bootstrapCredentialResult{}, false, nil
	}
	if err != nil {
		return bootstrapCredentialResult{}, true, internalAPIError(err)
	}
	if reason := authn.BootstrapReasonCode(bootstrapToken, a.now()); reason != "" {
		return bootstrapCredentialResult{}, true, httpauth.BootstrapRejectedError(reason)
	}
	if !allowBootstrap || !AllowsBootstrapTokenRoute(r.URL.Path) {
		return bootstrapCredentialResult{}, true, httpauth.BootstrapRejectedError("not_allowed_for_route")
	}
	if cookie, _ := r.Cookie(authn.SessionCookieName); cookie != nil {
		return bootstrapCredentialResult{}, true, invalidAuthRequest("authorization", "exactly one auth mode is allowed")
	}
	return bootstrapCredentialResult{
		Token:     bootstrapToken,
		TokenText: token,
		User:      user,
	}, true, nil
}
