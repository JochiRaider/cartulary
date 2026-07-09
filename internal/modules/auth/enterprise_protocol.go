package auth

import (
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
)

type enterpriseOIDCVerifier = enterpriseauth.OIDCVerifier
type enterpriseSAMLVerifier = enterpriseauth.SAMLVerifier
type enterpriseBeginRedirectBuilder func(provider authn.EnterpriseAuthProviderRecord, publicOrigin string, transaction authn.EnterpriseAuthTransactionRecord, pkceVerifier string) (enterpriseauth.BeginRedirect, error)
