package entities

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store         *Store
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
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/cartulary.view.hosts.v1/rows", service.handleHostCreate)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/cartulary.view.identities.v1/rows", service.handleIdentityCreate)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/cartulary.view.indicators.v1/rows", service.handleIndicatorCreate)
		mux.HandleFunc("POST /api/v1/records/{survivor_record_id}/merge", service.handleMerge)
		mux.HandleFunc("POST /api/v1/entity-mentions/{entity_mention_id}/resolve", service.handleMentionAction)
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
	return &Service{
		store:         NewStore(deps.PostgresHandle()),
		incidentStore: incidents.NewStore(deps.PostgresHandle()),
		authStore:     authn.NewStore(deps.PostgresHandle()),
		hub:           deps.WSHub,
		keys:          keys,
		now:           now,
	}, nil
}

func (s *Service) handleHostCreate(w http.ResponseWriter, r *http.Request) {
	s.handleCreate(w, r, HostsViewSchemaID)
}

func (s *Service) handleIdentityCreate(w http.ResponseWriter, r *http.Request) {
	s.handleCreate(w, r, IdentitiesViewSchemaID)
}

func (s *Service) handleIndicatorCreate(w http.ResponseWriter, r *http.Request) {
	s.handleCreate(w, r, IndicatorsViewSchemaID)
}

func (s *Service) handleMerge(w http.ResponseWriter, r *http.Request) {
	survivorRecordID, ok := pathUUID(w, r, "survivor_record_id")
	if !ok {
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	incidentID, err := s.store.GetMergeRouteIncident(r.Context(), survivorRecordID)
	if errors.Is(err, ErrMergeTargetNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := DecodeMergeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.store.MergeEntity(r.Context(), principal.User, survivorRecordID, request, MergeRequestHash(survivorRecordID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		preconditionErr *MergePreconditionError
		rowConflictErr  *MergeRowVersionConflictError
		recordLockedErr *MergeRecordLockedError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, ErrMergeTargetNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.As(err, &preconditionErr):
		writeAPIError(w, r, mergePreconditionFailedError(preconditionErr))
		return
	case errors.As(err, &rowConflictErr):
		writeAPIError(w, r, mergeRowVersionConflictError(rowConflictErr))
		return
	case errors.As(err, &recordLockedErr):
		writeAPIError(w, r, recordLockedError(recordLockedErr))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if !result.Replayed {
		s.publishMergeChanges(result, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleMentionAction(w http.ResponseWriter, r *http.Request) {
	mentionID, ok := pathUUID(w, r, "entity_mention_id")
	if !ok {
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	access, err := s.store.GetMentionActionAccess(r.Context(), mentionID, principal.User.ID)
	switch {
	case errors.Is(err, ErrEntityMentionNotFound):
		writeAPIError(w, r, entityMentionNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !slices.Contains([]string{"editor", "reviewer", "admin"}, access.Role) {
		writeAPIError(w, r, authorizationDeniedError(requiredRoleDescription("editor", "reviewer", "admin")))
		return
	}

	request, apiErr := DecodeMentionActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.store.ApplyMentionAction(r.Context(), principal.User, mentionID, request, MentionActionRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		rowConflict   *MentionRowVersionConflictError
		transitionErr *MentionTransitionError
		targetErr     *MentionTargetValidationError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, ErrEntityMentionNotFound):
		writeAPIError(w, r, entityMentionNotFoundError())
		return
	case errors.Is(err, ErrResolvedRecordNotFound):
		writeAPIError(w, r, resolvedRecordNotFoundError())
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, mentionRowVersionConflictError(rowConflict))
		return
	case errors.As(err, &transitionErr):
		writeAPIError(w, r, mentionIllegalTransitionError(transitionErr))
		return
	case errors.As(err, &targetErr):
		writeAPIError(w, r, invalidMutationPayload("resolved_record_id", "invalid_value"))
		return
	case errors.Is(err, ErrRecordDeletedUseRestore):
		writeAPIError(w, r, recordDeletedUseRestoreError())
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
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request, viewSchemaID string) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := DecodeCreateRequest(viewSchemaID, r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	var (
		result MutationResult
		err    error
	)
	switch viewSchemaID {
	case HostsViewSchemaID:
		result, err = s.store.CreateHostRow(r.Context(), principal.User, incidentID, request, CreateRequestHash(viewSchemaID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	case IdentitiesViewSchemaID:
		result, err = s.store.CreateIdentityRow(r.Context(), principal.User, incidentID, request, CreateRequestHash(viewSchemaID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	case IndicatorsViewSchemaID:
		result, err = s.store.CreateIndicatorRow(r.Context(), principal.User, incidentID, request, CreateRequestHash(viewSchemaID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	default:
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unknown_view_schema"))
		return
	}
	var conflict *ExactMatchConflictError
	var createValidationErr *IndicatorCreateValidationError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, ErrInvalidCreateRequest):
		writeAPIError(w, r, invalidMutationPayload("payload", "at_least_one_value_required"))
		return
	case errors.As(err, &createValidationErr):
		writeAPIError(w, r, invalidMutationPayload(createValidationErr.Field, createValidationErr.ReasonCode))
		return
	case errors.As(err, &conflict):
		writeAPIError(w, r, exactMatchConflictError(conflict.EntityType, conflict.IdentifierClass, conflict.CandidateRecords))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
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
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
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

func (s *Service) publishRecordChange(result MentionActionResult, actorUserID uuid.UUID) {
	if s.hub == nil {
		return
	}
	if result.SourceRecordID != uuid.Nil && result.ChangeSetID != uuid.Nil {
		changedKeys := append([]string(nil), result.ChangedFieldKeys...)
		slices.Sort(changedKeys)
		s.hub.PublishRecordChange(platformws.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         result.SourceRecordID,
			RowVersion:       result.SourceRecordRowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: changedKeys,
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})
	}
	for _, invalidation := range result.EntityInvalidations {
		if invalidation.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
			continue
		}
		changedKeys := append([]string(nil), invalidation.ChangedFieldKeys...)
		slices.Sort(changedKeys)
		s.hub.PublishRecordChange(platformws.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: changedKeys,
			ViewSchemaID:     invalidation.ViewSchemaID,
		})
	}
}

func (s *Service) publishMergeChanges(result MergeResult, actorUserID uuid.UUID) {
	if s.hub == nil || result.ChangeSetID == uuid.Nil {
		return
	}

	viewSchemaID := IdentitiesViewSchemaID
	survivorKeys := []string{"identity.identity_state", "identity.edited_at", "identity.aliases", "identity.reusable_identifiers", "identity.aad_object_id", "identity.sid", "identity.upn", "identity.email", "identity.sam_account_name"}
	loserKeys := []string{"identity.identity_state", "identity.edited_at"}
	if result.RecordType == "host" {
		viewSchemaID = HostsViewSchemaID
		survivorKeys = []string{"host.host_state", "host.edited_at", "host.aliases", "host.reusable_identifiers", "host.aad_device_id", "host.fqdn", "host.hostname"}
		loserKeys = []string{"host.host_state", "host.edited_at"}
	}
	slices.Sort(survivorKeys)
	slices.Sort(loserKeys)

	s.hub.PublishRecordChange(platformws.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.SurvivorRecordID,
		RowVersion:       result.SurvivorRowVersion,
		ChangeSetID:      result.ChangeSetID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: append([]string(nil), survivorKeys...),
		ViewSchemaID:     viewSchemaID,
	})
	s.hub.PublishRecordChange(platformws.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.LoserRecordID,
		RowVersion:       result.LoserRowVersion,
		ChangeSetID:      result.ChangeSetID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: append([]string(nil), loserKeys...),
		ViewSchemaID:     viewSchemaID,
	})

	for _, invalidation := range result.TimelineInvalidations {
		changedKeys := append([]string(nil), invalidation.ChangedFieldKeys...)
		slices.Sort(changedKeys)
		s.hub.PublishRecordChange(platformws.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      result.ChangeSetID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: changedKeys,
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})
	}
}
