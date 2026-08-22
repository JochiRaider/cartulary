package entities

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type service struct {
	mergeStore     *merge.Store
	mentionStore   *mentions.Store
	incidentAccess incidentAdmissionChecker
	authStore      *authn.Store
	keys           authn.MasterKeys
	now            func() time.Time
}

type incidentAdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

type RouteOptions struct {
	MergeStore   *merge.Store
	MentionStore *mentions.Store
}

func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.entities", map[string]http.HandlerFunc{
			"mergeEntityRecord":    service.handleMerge,
			"resolveEntityMention": service.handleMentionAction,
		})
	}
}

func newService(deps httpapi.DependencySet, options RouteOptions) (*service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if options.MentionStore == nil {
		return nil, errors.New("entities route composition requires a mention store")
	}
	if options.MergeStore == nil {
		return nil, errors.New("entities route composition requires a merge store")
	}
	return &service{
		mergeStore:     options.MergeStore,
		mentionStore:   options.MentionStore,
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		now:            now,
	}, nil
}

func (s *service) handleMerge(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	survivorRecordID, err := parsePathUUID(r, "survivor_record_id")
	if err != nil {
		writeAPIError(w, r, incidentNotFoundError())
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
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesReviewerAdmin, "reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	request, apiErr := merge.DecodeMergeRequest(r.Body)
	if apiErr != nil {
		if invalidAPIErrorField(apiErr, "loser_record_id", "invalid_value") {
			writeAPIError(w, r, incidentNotFoundError())
			return
		}
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
	case admission.IsDenied(err, admission.DenialIncidentClosed):
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

	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *service) handleMentionAction(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	mentionID, err := parsePathUUID(r, "entity_mention_id")
	if err != nil {
		writeAPIError(w, r, entityMentionNotFoundError())
		return
	}

	access, err := s.mentionStore.GetMentionActionAccess(r.Context(), mentionID, principal.User.ID)
	switch {
	case errors.Is(err, mentions.ErrEntityMentionNotFound):
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
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, mentions.ErrEntityMentionNotFound):
		writeAPIError(w, r, entityMentionNotFoundError())
		return
	case errors.Is(err, mentions.ErrResolvedRecordNotFound):
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
	case errors.Is(err, mentions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, recordDeletedUseRestoreError())
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

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, incidentNotFoundError()
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, authorizationDeniedError(requiredRole)
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

func parsePathUUID(r *http.Request, key string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(key))
}

func invalidAPIErrorField(apiErr *httpapi.APIError, field string, reasonCode string) bool {
	if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
		return false
	}
	return apiErr.Details["field"] == field && apiErr.Details["reason_code"] == reasonCode
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func entityMentionNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusNotFound,
		Code:    "entity_mention_not_found",
		Message: "entity mention not found",
		Details: map[string]any{},
	}
}

func resolvedRecordNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusNotFound,
		Code:    "resolved_record_not_found",
		Message: "resolved record not found",
		Details: map[string]any{},
	}
}

func mentionRowVersionConflictError(conflict *mentions.MentionRowVersionConflictError) *httpapi.APIError {
	details := map[string]any{}
	if conflict != nil {
		details["entity_mention_id"] = conflict.EntityMentionID.String()
		details["base_mention_row_version"] = conflict.BaseMentionRowVersion
		details["current_mention_row_version"] = conflict.CurrentMentionRowVersion
		details["source_record_id"] = conflict.SourceRecordID.String()
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "row_version_conflict",
		Message: "row version conflict",
		Details: details,
	}
}

func mentionIllegalTransitionError(err *mentions.MentionTransitionError) *httpapi.APIError {
	details := map[string]any{}
	if err != nil {
		details["from_status"] = err.FromStatus
		details["to_status"] = err.ToStatus
		details["violated_guards"] = append([]string(nil), err.ViolatedGuards...)
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: details,
	}
}

func recordDeletedUseRestoreError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_deleted_use_restore",
		Message: "record deleted use restore",
		Details: map[string]any{},
	}
}
