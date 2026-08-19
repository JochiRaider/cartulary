package savedviews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type service struct {
	application    *savedViewApplication
	incidentAccess *admission.Checker
	authStore      *authn.Store
	keys           authn.MasterKeys
	cursorCodec    *pagination.Codec
	now            func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.savedviews", map[string]http.HandlerFunc{
			"createIncidentSavedView": service.handleCollection,
			"deleteIncidentSavedView": service.handleItem,
			"listIncidentSavedViews":  service.handleCollection,
			"patchIncidentSavedView":  service.handleItem,
		})
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("POST /api/v1/test/incidents/{incident_id}/saved-views/system", guard.Protect(service.handleTestSystemCreate))
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
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		cursorCodec = pagination.NewCodec(cursorKey[:])
	}
	return &service{
		application:    newSavedViewApplication(newPostgresSavedViewRepository(deps.PostgresHandle())),
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		cursorCodec:    cursorCodec,
		now:            now,
	}, nil
}

func (s *service) handleCollection(w http.ResponseWriter, r *http.Request) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.handleList(w, r, incidentID)
	case http.MethodPost:
		s.handleCreate(w, r, incidentID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *service) handleItem(w http.ResponseWriter, r *http.Request) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	savedViewID, err := uuid.Parse(r.PathValue("saved_view_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		s.handlePatch(w, r, incidentID, savedViewID)
	case http.MethodDelete:
		s.handleDelete(w, r, incidentID, savedViewID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *service) handleTestSystemCreate(w http.ResponseWriter, r *http.Request) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	request, apiErr := decodeSystemFixtureRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.createSystemFixture(r.Context(), incidentID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, buildResource(record))
}

func (s *service) handleList(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
		r.URL.Query(),
		"incident.saved-views.list",
		principal.User.ID.String(),
		map[string]string{"incident_id": incidentID.String()},
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	pageRequest, reasonCode := savedViewListPageRequest(binding, cursor)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	records, err := s.application.listVisible(r.Context(), incidentID, principal.User.ID, pageRequest)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	anchor := s.now().UTC()
	if pageRequest.AnchorUpdatedAt != nil {
		anchor = *pageRequest.AnchorUpdatedAt
	} else if len(records) > 0 {
		anchor = records[0].UpdatedAt.UTC()
	}
	rows, nextCursor, err := buildSavedViewListPage(binding, anchor, records)
	if err != nil {
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
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"saved_views": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *service) handleCreate(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.create(r.Context(), incidentID, principal.User.ID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, buildResource(record))
}

func (s *service) handlePatch(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, savedViewID uuid.UUID) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	current, err := s.application.visibleForPatch(r.Context(), incidentID, savedViewID, principal.User.ID)
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	request, apiErr := decodePatchRequest(r.Body, current.ViewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.patch(r.Context(), incidentID, savedViewID, principal.User.ID, membership.Role.String(), request, s.now())
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildResource(record))
}

func (s *service) handleDelete(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, savedViewID uuid.UUID) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.application.delete(r.Context(), incidentID, savedViewID, principal.User.ID, membership.Role.String()); err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"saved_view_id": savedViewID,
		"deleted":       true,
	})
}

func savedViewListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (listPageRequest, string) {
	request := listPageRequest{Limit: binding.Limit + 1}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_saved_view_id"])
	if err != nil {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &listPosition{UpdatedAt: lastUpdatedAt.UTC(), SavedViewID: lastID}
	return request, ""
}

func buildSavedViewListPage(binding pagination.Binding, anchor time.Time, records []savedViewRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, buildResource(record))
	}
	rows, err := pagination.MarshalResources(resources)
	if err != nil {
		return nil, nil, err
	}
	if !hasMore || len(pageRecords) == 0 {
		return rows, nil, nil
	}
	last := pageRecords[len(pageRecords)-1]
	return rows, &pagination.Cursor{
		Version:     pagination.CursorVersion,
		Mode:        pagination.ModeKeyset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       binding.Scope,
		Position: map[string]string{
			"anchor_updated_at":  anchor.UTC().Format(time.RFC3339Nano),
			"last_updated_at":    last.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"last_saved_view_id": last.SavedViewID.String(),
		},
	}, nil
}

func savedViewError(err error) *httpapi.APIError {
	var versionConflict *savedViewVersionConflictError
	switch {
	case errors.Is(err, errSavedViewNotFound):
		return savedViewNotFoundError()
	case errors.Is(err, errSavedViewMutationDenied):
		return authorizationDeniedError()
	case errors.As(err, &versionConflict):
		return savedViewVersionConflictAPIError(versionConflict)
	case errors.Is(err, errSavedViewVersionConflict):
		return savedViewVersionConflictAPIError(nil)
	default:
		return internalAPIError(err)
	}
}

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, savedViewNotFoundError()
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}
