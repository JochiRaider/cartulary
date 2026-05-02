package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store         *Store
	entityStore   *entities.Store
	incidentStore *incidents.Store
	authStore     *authn.Store
	hub           *platformws.Hub
	keys          authn.MasterKeys
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("POST /api/v1/records/{record_id}/mark-reviewed", service.handleMarkReviewed)
		mux.HandleFunc("POST /api/v1/records/{record_id}/supersede", service.handleSupersede)
		return nil
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/test/timeline/record-changes", service.handleRecordChangeSnapshot)
		mux.HandleFunc("GET /api/v1/test/timeline/records/{record_id}/substrate", service.handleRecordSubstrateSnapshot)
		mux.HandleFunc("/ws/v1/test/record-changes", service.handleRecordChangeSocket)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		store:         NewStore(deps.Postgres),
		entityStore:   entities.NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		hub:           deps.WSHub,
		keys:          keys,
		now:           now,
	}, nil
}

func (s *Service) handleMarkReviewed(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireTimelineRole(r.Context(), recordID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeTimelineActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.store.MarkReviewed(r.Context(), principal.User, recordID, request, TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, request.Reason, nil), httpapi.RequestIDFromContext(r.Context()), s.now())
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, ErrRowVersionConflict):
		writeAPIError(w, r, rowVersionConflictError())
		return
	case errors.Is(err, ErrIllegalTransition):
		writeAPIError(w, r, illegalTransitionError("mark_reviewed_not_allowed", err))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if !result.Replayed {
		s.publishRecordChange(result, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleSupersede(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireTimelineRole(r.Context(), recordID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeTimelineSupersedeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.store.Supersede(r.Context(), principal.User, recordID, request, TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID), httpapi.RequestIDFromContext(r.Context()), s.now())
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, ErrRowVersionConflict):
		writeAPIError(w, r, rowVersionConflictError())
		return
	case errors.Is(err, ErrIllegalTransition):
		writeAPIError(w, r, illegalTransitionError("supersede_not_allowed", err))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if !result.Replayed {
		s.publishRecordChange(result, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleRecordChangeSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	changes := s.hub.SnapshotRecordChanges()
	items := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		items = append(items, recordChangePayload(change))
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"record_changes": items})
}

func (s *Service) handleRecordSubstrateSnapshot(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}

	snapshot, err := s.store.SnapshotRecordSubstrate(r.Context(), recordID)
	switch {
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"record_id":             snapshot.RecordID.String(),
		"row_version":           snapshot.RowVersion,
		"capture_state":         snapshot.CaptureState,
		"replacement_record_id": formatUUIDPointer(snapshot.ReplacementRecordID),
		"record_revision_count": snapshot.RecordRevisionCount,
	})
}

func (s *Service) handleRecordChangeSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := platformws.Accept(w, r, "")
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	changes, unsubscribe := s.hub.SubscribeRecordChanges(16)
	defer unsubscribe()

	if err := platformws.WriteJSON(ctx, conn, platformws.Message{
		Type:    "connected",
		Payload: platformws.RawPayload(map[string]any{"boundary": "/ws/v1/test/record-changes"}),
	}); err != nil {
		return
	}

	go func() {
		defer cancel()
		for {
			var message platformws.Message
			if err := platformws.ReadJSON(ctx, conn, &message); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case change, ok := <-changes:
			if !ok {
				return
			}
			if err := platformws.WriteJSON(ctx, conn, platformws.Message{
				Type:    "record_changed",
				Payload: platformws.RawPayload(recordChangePayload(change)),
			}); err != nil {
				return
			}
		}
	}
}

func (s *Service) publishRecordChange(result MutationResult, actorUserID uuid.UUID) {
	if result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		return
	}
	changedKeys := append([]string(nil), result.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	s.hub.PublishRecordChange(platformws.RecordChange{
		IncidentID:       result.Row.IncidentID,
		RecordID:         result.RecordID,
		RowVersion:       result.RowVersion,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: changedKeys,
		ViewSchemaID:     TimelineViewSchemaID,
	})
}

func recordChangePayload(change platformws.RecordChange) map[string]any {
	return map[string]any{
		"record_id":          change.RecordID.String(),
		"row_version":        change.RowVersion,
		"change_set_id":      change.ChangeSetID.String(),
		"client_txn_id":      change.ClientTxnID,
		"actor_user_id":      change.ActorUserID.String(),
		"changed_field_keys": append([]string(nil), change.ChangedFieldKeys...),
		"affected_views": []map[string]any{
			{
				"view_schema_id": change.ViewSchemaID,
				"change_kind":    "invalidate",
			},
		},
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

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *auth.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	if !slices.Contains(roles, record.Role) {
		return incidents.MembershipRecord{}, authorizationDeniedError(requiredRoleDescription(roles...))
	}
	return record, nil
}

func (s *Service) requireTimelineRole(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, roles ...string) (uuid.UUID, *auth.APIError) {
	incidentID, err := s.store.GetRecordIncident(ctx, recordID)
	if errors.Is(err, ErrRecordNotFound) {
		return uuid.UUID{}, incidentNotFoundError()
	}
	if err != nil {
		return uuid.UUID{}, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentRole(ctx, incidentID, userID, roles...); apiErr != nil {
		return uuid.UUID{}, apiErr
	}
	return incidentID, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return s.authenticateSessionToken(r, auth.AuthSourceBearer, token, stateChanging)
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	return s.authenticateSessionToken(r, auth.AuthSourceCookie, cookie.Value, stateChanging)
}

func (s *Service) authenticateSessionToken(r *http.Request, authSource auth.AuthSource, sessionToken string, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	if stateChanging && authSource == auth.AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(s.keys, sessionToken) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := auth.ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return auth.SessionPrincipal{}, apiErr
		}
	}

	session, user, err := s.authStore.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(s.keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}
	if err != nil {
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}

	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.authStore.RevokeSession(context.Background(), session.ID, "session_expired", now)
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}

	return auth.SessionPrincipal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
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
