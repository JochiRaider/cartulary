package viewschemas

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"net/http"
	"time"
)

type Service struct {
	authStore   *authn.Store
	keys        authn.MasterKeys
	cursorCodec *pagination.Codec
	now         func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}

		if err := httpapi.DeclarePublicOperations(deps, PublicOperations()...); err != nil {
			return err
		}
		httpapi.HandlePublicRoute(mux, "GET /api/v1/view-schemas", service.handleViewSchemasCollection)
		httpapi.HandlePublicRoute(mux, "GET /api/v1/view-schemas/{view_schema_id}", service.handleViewSchemaMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}
	return &Service{
		authStore:   authn.NewStore(deps.PostgresHandle()),
		keys:        keys,
		cursorCodec: cursorCodec,
		now:         now,
	}, nil
}

func (s *Service) handleViewSchemasCollection(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
		r.URL.Query(),
		"view-schemas.list",
		principal.User.ID.String(),
		nil,
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}

	resources := viewschema.ListPublicResources()
	rawRows := make([]json.RawMessage, 0, len(resources))
	for _, resource := range resources {
		payload, err := json.Marshal(resource)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rawRows = append(rawRows, json.RawMessage(payload))
	}
	rows, nextCursor, err := pagination.PageRawMessages(binding, cursor, rawRows)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken):
		writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var nextToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"view_schemas": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *Service) handleViewSchemaMember(w http.ResponseWriter, r *http.Request) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	resource, ok := viewschema.LookupPublicResource(r.PathValue("view_schema_id"))
	if !ok {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "view_schema_not_found", Details: map[string]any{}})
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
