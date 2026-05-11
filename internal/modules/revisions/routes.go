package revisions

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
	pagination    *pagination.Registry
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /api/v1/records/{record_id}/history", service.handleRecordHistory)
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
	paginator := deps.Pagination
	if paginator == nil {
		paginator = pagination.NewRegistry()
	}
	return &Service{
		store:         NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		keys:          keys,
		pagination:    paginator,
		now:           now,
	}, nil
}

func (s *Service) handleRecordHistory(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	record, err := s.store.GetHistoryRecord(r.Context(), recordID)
	if errors.Is(err, ErrRecordNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), record.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	binding, cursor, reasonCode := pagination.ResolveRequest(
		r.URL.Query(),
		"records.history",
		principal.User.ID.String(),
		map[string]string{"record_id": recordID.String()},
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}

	var (
		rows       []json.RawMessage
		nextCursor *pagination.Cursor
	)
	if cursor == nil {
		resources, err := s.store.ListRecordHistory(r.Context(), record)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rows, err = pagination.MarshalResources(resources)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rows, nextCursor = s.pagination.Start(binding, rows)
	} else {
		var err error
		rows, nextCursor, err = s.pagination.Continue(binding, *cursor)
		switch {
		case errors.Is(err, pagination.ErrCursorSnapshotExpired):
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonCursorSnapshotUnavailable))
			return
		case errors.Is(err, pagination.ErrInvalidCursorToken):
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}

	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var nextToken *string
	if nextCursor != nil {
		token, err := pagination.EncodeCursor(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
		"incident_id": record.IncidentID.String(),
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"deleted":     record.Deleted,
		"items":       rows,
	}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
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

func (s *Service) authenticateSessionRequest(r *http.Request) (auth.SessionPrincipal, *auth.APIError) {
	return auth.AuthenticateSessionRequest(r, auth.SessionAuthOptions{
		Store:         s.authStore,
		Keys:          s.keys,
		Now:           s.now,
		StateChanging: false,
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
	}
	now := s.now()
	if !auth.ShouldPersistIdleExpirySlide(sliding, now) {
		return nil
	}
	sliding = sliding.Slide(now)
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
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	value, err := uuid.Parse(raw)
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}

func invalidPaginationRequest(reasonCode string) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
