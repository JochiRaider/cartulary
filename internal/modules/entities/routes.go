package entities

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/google/uuid"
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

func (s *Service) handleMerge(w http.ResponseWriter, r *http.Request) {
	survivorRecordID, ok := pathUUID(w, r, "survivor_record_id")
	if !ok {
		return
	}

	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
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
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
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

	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
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
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
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

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	record, err := s.incidentStore.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, incidentNotFoundError()
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	if !slices.Contains(roles, record.Role) {
		return incidents.MembershipRecord{}, authorizationDeniedError(requiredRoleDescription(roles...))
	}
	return record, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	if principal == nil || !httpauth.ShouldSlideIdleExpiry(method, path) {
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

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
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

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
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
