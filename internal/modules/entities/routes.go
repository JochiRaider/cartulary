package entities

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/candidates"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mutationadmission"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type service struct {
	candidateReader candidates.PageReader
	cursorCodec     *pagination.Codec
	mergeStore      *merge.Store
	mentionStore    *mentions.Store
	incidentAccess  incidentAdmissionChecker
	authStore       *authn.Store
	now             func() time.Time
}

type incidentAdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

type RouteOptions struct {
	CandidateReader candidates.PageReader
	MergeStore      *merge.Store
	MentionStore    *mentions.Store
}

// RegisterRoutes binds the entity routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits the path identifier before a handler runs.
func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.entities", map[string]http.HandlerFunc{
			"listEntityCandidates": service.handleCandidates,
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
	if options.CandidateReader == nil {
		return nil, errors.New("entities route composition requires a candidate reader")
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	return &service{
		candidateReader: options.CandidateReader,
		cursorCodec:     cursorCodec,
		mergeStore:      options.MergeStore,
		mentionStore:    options.MentionStore,
		incidentAccess:  admission.NewChecker(deps.PostgresHandle()),
		authStore:       authn.NewStore(deps.PostgresHandle()),
		now:             now,
	}, nil
}

func (s *service) handleMerge(w http.ResponseWriter, r *http.Request) {
	principal, survivorRecordID, apiErr := admittedPath(r, "survivor_record_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	incidentID, err := s.mergeStore.GetMergeRouteIncident(r.Context(), survivorRecordID)
	if errors.Is(err, merge.ErrMergeTargetNotFound) {
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
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

	request, failure := merge.DecodeMergeRequest(r.Body)
	if failure != nil {
		if mutationAdmissionFailureField(failure, "loser_record_id", mutationadmission.ReasonInvalidValue) {
			writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
			return
		}
		writeAPIError(w, r, mutationAdmissionAPIError(failure))
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
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
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

	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *service) handleMentionAction(w http.ResponseWriter, r *http.Request) {
	principal, mentionID, apiErr := admittedPath(r, "entity_mention_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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

	request, failure := mentions.DecodeMentionActionRequest(r.Body)
	if failure != nil {
		writeAPIError(w, r, mutationAdmissionAPIError(failure))
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

	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, httpapi.NotFound("incident_not_found")
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, authorizationDeniedError(requiredRole)
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

// admittedPath returns the principal and the uuid path parameter central
// admission established (REQ-01-707 steps 2 and 3); their absence is a
// binding error.
func admittedPath(r *http.Request, name string) (httpauth.Principal, uuid.UUID, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	value, identified := httpapi.PathUUID(r, name)
	if !authenticated || !identified {
		return httpauth.Principal{}, uuid.UUID{}, httpapi.InternalAPIError(errors.New("entity route was not admitted centrally"))
	}
	return principal, value, nil
}

func mutationAdmissionFailureField(
	failure *mutationadmission.Failure,
	field string,
	reasonCode mutationadmission.ReasonCode,
) bool {
	if failure == nil || failure.ReasonCode() != reasonCode {
		return false
	}
	failureField, ok := failure.Field()
	return ok && failureField == field
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

// entityMentionNotFoundError is the mention action's not-found response,
// identical to the one central admission writes for a malformed
// entity_mention_id (REQ-01-707).
func entityMentionNotFoundError() *httpapi.APIError {
	return httpapi.NotFound("entity_mention_not_found")
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
		details["violated_guards"] = append([]string{}, err.ViolatedGuards...)
		details["reason_code"] = err.ReasonCode
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
