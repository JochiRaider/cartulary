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
	cursorCodec *pagination.Codec
	now         func() time.Time
}

// RegisterRoutes binds the view-schema discovery routes for central admission
// (Core 01 REQ-01-707): the registry authenticates by each operation's
// declared security and admits view_schema_id by its declared syntax before a
// handler runs.
func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}

		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "platform.viewschema", map[string]http.HandlerFunc{
			"getViewSchema":   service.handleViewSchemaMember,
			"listViewSchemas": service.handleViewSchemasCollection,
		})
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
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	return &Service{
		authStore:   authn.NewStore(deps.PostgresHandle()),
		cursorCodec: cursorCodec,
		now:         now,
	}, nil
}

// admittedPrincipal returns the principal central admission established
// (REQ-01-707 step 2); its absence is a binding error.
func admittedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, ok := httpauth.AdmittedPrincipal(r)
	if !ok {
		return httpauth.Principal{}, httpapi.InternalAPIError(errors.New("view schema route was not admitted centrally"))
	}
	return principal, nil
}

func (s *Service) handleViewSchemasCollection(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
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
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	viewSchemaID, ok := httpapi.PathString(r, "view_schema_id")
	if !ok {
		writeAPIError(w, r, httpapi.InternalAPIError(errors.New("view schema route did not admit view_schema_id")))
		return
	}
	// REQ-01-707 step 5: the singleton-query check is route handling, after
	// authentication and path-parameter admission.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// Central admission checked only the identifier's syntax; whether a
	// well-formed identifier names a public view schema is route handling.
	resource, ok := viewschema.LookupPublicResource(viewSchemaID)
	if !ok {
		writeAPIError(w, r, httpapi.NotFound("view_schema_not_found"))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
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
