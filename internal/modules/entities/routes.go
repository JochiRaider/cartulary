package entities

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type Service struct {
	mergeStore     *merge.Store
	mentionStore   *mentions.Store
	incidentAccess incidents.Access
	authStore      *authn.Store
	publisher      *collaboration.RecordChangePublisher
	keys           authn.MasterKeys
	now            func() time.Time
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
		mergeStore:     merge.NewStore(deps.PostgresHandle()),
		mentionStore:   mentions.NewStore(deps.PostgresHandle()),
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		publisher:      collaboration.NewRecordChangePublisher(deps.WSHub),
		keys:           keys,
		now:            now,
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

	incidentID, err := s.mergeStore.GetMergeRouteIncident(r.Context(), survivorRecordID)
	if errors.Is(err, merge.ErrMergeTargetNotFound) {
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

	request, apiErr := merge.DecodeMergeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.mergeStore.MergeEntity(r.Context(), principal.User, survivorRecordID, request, merge.MergeRequestHash(survivorRecordID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		preconditionErr *merge.MergePreconditionError
		rowConflictErr  *merge.MergeRowVersionConflictError
		recordLockedErr *merge.MergeRecordLockedError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, merge.ErrMergeTargetNotFound):
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

	access, err := s.mentionStore.GetMentionActionAccess(r.Context(), mentionID, principal.User.ID)
	switch {
	case errors.Is(err, mentions.ErrEntityMentionNotFound):
		writeAPIError(w, r, mentions.EntityMentionNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !slices.Contains([]string{"editor", "reviewer", "admin"}, access.Role) {
		writeAPIError(w, r, authorizationDeniedError(requiredRoleDescription("editor", "reviewer", "admin")))
		return
	}

	request, apiErr := mentions.DecodeMentionActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.mentionStore.ApplyMentionAction(r.Context(), principal.User, mentionID, request, mentions.MentionActionRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		rowConflict   *mentions.MentionRowVersionConflictError
		transitionErr *mentions.MentionTransitionError
		targetErr     *mentions.MentionTargetValidationError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, mentions.ErrEntityMentionNotFound):
		writeAPIError(w, r, mentions.EntityMentionNotFoundError())
		return
	case errors.Is(err, mentions.ErrResolvedRecordNotFound):
		writeAPIError(w, r, mentions.ResolvedRecordNotFoundError())
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, mentions.RowVersionConflictAPIError(rowConflict))
		return
	case errors.As(err, &transitionErr):
		writeAPIError(w, r, mentions.IllegalTransitionAPIError(transitionErr))
		return
	case errors.As(err, &targetErr):
		writeAPIError(w, r, invalidMutationPayload("resolved_record_id", "invalid_value"))
		return
	case errors.Is(err, mentions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, mentions.RecordDeletedUseRestoreError())
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

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
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

func (s *Service) publishRecordChange(result mentions.MentionActionResult, actorUserID uuid.UUID) {
	if result.SourceRecordID != uuid.Nil && result.ChangeSetID != uuid.Nil {
		s.publisher.Publish(collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         result.SourceRecordID,
			RowVersion:       result.SourceRecordRowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: result.ChangedFieldKeys,
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})
	}
	for _, invalidation := range result.EntityInvalidations {
		if invalidation.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
			continue
		}
		s.publisher.Publish(collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
			ViewSchemaID:     invalidation.ViewSchemaID,
		})
	}
}

func (s *Service) publishMergeChanges(result merge.MergeResult, actorUserID uuid.UUID) {
	if result.ChangeSetID == uuid.Nil {
		return
	}

	viewSchemaID := entitycontract.IdentitiesViewSchemaID
	survivorKeys := []string{"identity.identity_state", "identity.edited_at", "identity.aliases", "identity.reusable_identifiers", "identity.aad_object_id", "identity.sid", "identity.upn", "identity.email", "identity.sam_account_name"}
	loserKeys := []string{"identity.identity_state", "identity.edited_at"}
	if result.RecordType == "host" {
		viewSchemaID = entitycontract.HostsViewSchemaID
		survivorKeys = []string{"host.host_state", "host.edited_at", "host.aliases", "host.reusable_identifiers", "host.aad_device_id", "host.fqdn", "host.hostname"}
		loserKeys = []string{"host.host_state", "host.edited_at"}
	}
	slices.Sort(survivorKeys)
	slices.Sort(loserKeys)

	s.publisher.Publish(collaboration.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.SurvivorRecordID,
		RowVersion:       result.SurvivorRowVersion,
		ChangeSetID:      result.ChangeSetID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: append([]string(nil), survivorKeys...),
		ViewSchemaID:     viewSchemaID,
	})
	s.publisher.Publish(collaboration.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.LoserRecordID,
		RowVersion:       result.LoserRowVersion,
		ChangeSetID:      result.ChangeSetID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: append([]string(nil), loserKeys...),
		ViewSchemaID:     viewSchemaID,
		ChangeKind:       "remove",
	})

	for _, invalidation := range result.TimelineInvalidations {
		s.publisher.Publish(collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangeSetID:      result.ChangeSetID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
			ViewSchemaID:     "cartulary.view.timeline.v2",
		})
	}
}
