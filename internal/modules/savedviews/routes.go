package savedviews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type Service struct {
	store         *Store
	incidentStore *incidents.Store
	authStore     *authn.Store
	keys          authn.MasterKeys
	cursorCodec   *pagination.Codec
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /api/v1/incidents/{incident_id}/saved-views", service.handleCollection)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/saved-views", service.handleCollection)
		mux.HandleFunc("PATCH /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}", service.handleItem)
		mux.HandleFunc("DELETE /api/v1/incidents/{incident_id}/saved-views/{saved_view_id}", service.handleItem)
		return nil
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
		store:         NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		keys:          keys,
		cursorCodec:   cursorCodec,
		now:           now,
	}, nil
}

func (s *Service) handleCollection(w http.ResponseWriter, r *http.Request) {
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

func (s *Service) handleItem(w http.ResponseWriter, r *http.Request) {
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

func (s *Service) handleTestSystemCreate(w http.ResponseWriter, r *http.Request) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	request, apiErr := DecodeSystemFixtureRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.CreateSystemFixture(r.Context(), incidentID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, BuildResource(record))
}

func (s *Service) handleList(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	principal, apiErr := s.authenticateSessionRequest(r, false)
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
	records, err := s.store.ListVisible(r.Context(), incidentID, principal.User.ID, pageRequest)
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

func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.Create(r.Context(), principal.User, incidentID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, BuildResource(record))
}

func (s *Service) handlePatch(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, savedViewID uuid.UUID) {
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	current, err := s.store.GetVisibleForUpdate(r.Context(), incidentID, savedViewID, principal.User.ID)
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	request, apiErr := DecodePatchRequest(r.Body, current.ViewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.Patch(r.Context(), principal.User, membership.Role, incidentID, savedViewID, request, s.now())
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildResource(record))
}

func (s *Service) handleDelete(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, savedViewID uuid.UUID) {
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.store.Delete(r.Context(), principal.User, membership.Role, incidentID, savedViewID); err != nil {
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

func savedViewListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (ListPageRequest, string) {
	request := ListPageRequest{Limit: binding.Limit + 1}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return ListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return ListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return ListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_saved_view_id"])
	if err != nil {
		return ListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &ListPosition{UpdatedAt: lastUpdatedAt.UTC(), SavedViewID: lastID}
	return request, ""
}

func buildSavedViewListPage(binding pagination.Binding, anchor time.Time, records []Record) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, BuildResource(record))
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

func savedViewError(err error) *auth.APIError {
	var versionConflict *SavedViewVersionConflictError
	switch {
	case errors.Is(err, ErrSavedViewNotFound):
		return savedViewNotFoundError()
	case errors.Is(err, ErrSavedViewMutationDenied):
		return authorizationDeniedError()
	case errors.As(err, &versionConflict):
		return savedViewVersionConflictError(versionConflict)
	case errors.Is(err, ErrSavedViewVersionConflict):
		return savedViewVersionConflictError(nil)
	default:
		return internalAPIError(err)
	}
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *auth.APIError) {
	record, err := s.incidentStore.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, incidentNotFoundError()
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	return auth.AuthenticateSessionRequest(r, auth.SessionAuthOptions{
		Store:         s.authStore,
		Keys:          s.keys,
		Now:           s.now,
		StateChanging: stateChanging,
	})
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if principal == nil || !auth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}.Slide(s.now())
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}
