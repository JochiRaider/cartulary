// Package extensiondiscovery owns the HTTP transport, authentication, session,
// query, and serialization behavior for the Extensions discovery singleton.
package extensiondiscovery

import (
	"context"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

type service struct {
	authStore *authn.Store
	keys      authn.MasterKeys
	epoch     httpapi.ExtensionEpochProvider
	now       func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		discovery, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/extensions", discovery.handleCollection)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &service{
		authStore: authn.NewStore(deps.PostgresHandle()),
		keys:      keys,
		epoch:     deps.ExtensionEpoch,
		now:       now,
	}, nil
}

func (s *service) handleCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := slideSessionIfNeeded(r.Context(), s.authStore, &principal, r.Method, r.URL.Path, s.now); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	profiles := httpapi.ExtensionProfilesFromEpoch(s.epoch)
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildResponseData(profiles))
}

func slideSessionIfNeeded(
	ctx context.Context,
	store *authn.Store,
	principal *httpauth.Principal,
	method string,
	path string,
	now func() time.Time,
) error {
	return httpauth.SlideSessionIfNeeded(ctx, store, principal, method, path, now)
}

func buildResource(profile httpapi.ExtensionProfile) map[string]any {
	var contractMajor any
	if profile.ContractMajor != nil {
		contractMajor = *profile.ContractMajor
	}
	return map[string]any{
		"profile_id":     profile.ProfileID,
		"claimable":      profile.Claimable,
		"claimed":        profile.Claimed,
		"contract_major": contractMajor,
		"route_families": presentStrings(profile.RouteFamilies),
		"workspace_keys": presentStrings(profile.WorkspaceKeys),
		"capabilities":   presentStrings(profile.Capabilities),
	}
}

func buildResponseData(profiles []httpapi.ExtensionProfile) map[string]any {
	items := make([]map[string]any, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, buildResource(profile))
	}
	return map[string]any{"extensions": items}
}

func presentStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
