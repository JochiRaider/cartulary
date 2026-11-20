package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func parseUsersListScope(rawQuery string) (listquery.Result, *httpapi.APIError) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		Search: true,
		ExactFilters: map[string]listquery.ExactFilter{
			"is_active":           {Allowed: []string{"true", "false"}},
			"is_deployment_admin": {Allowed: []string{"true", "false"}},
		},
	})
	if queryErr == nil {
		return result, nil
	}
	if queryErr.Kind == listquery.ErrorKindPagination {
		return listquery.Result{}, userPaginationError(queryErr.ReasonCode)
	}
	return listquery.Result{}, userListQueryError(queryErr.ReasonCode)
}

func userListFilterFromScope(scope map[string]string) authn.UserListFilter {
	filter := authn.UserListFilter{SearchTokens: strings.Fields(scope["search"])}
	if value := scope["is_active"]; value != "" {
		parsed := value == "true"
		filter.IsActive = &parsed
	}
	if value := scope["is_deployment_admin"]; value != "" {
		parsed := value == "true"
		filter.IsDeploymentAdmin = &parsed
	}
	return filter
}

func (s *Service) handleListUsers(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedDeploymentAdmin(w, r)
	if !ok {
		return
	}
	listScope, apiErr := parseUsersListScope(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(
		listScope.Values,
		"users.list",
		principal.User.ID.String(),
		listScope.Scope,
	)
	if reasonCode != "" {
		writeAPIError(w, r, userPaginationError(reasonCode))
		return
	}

	users, listErr := s.userAdminStore.ListUsers(r.Context(), userListFilterFromScope(binding.Scope))
	if listErr != nil {
		writeAPIError(w, r, internalAPIError(listErr))
		return
	}
	resources := make([]map[string]any, 0, len(users))
	for _, user := range users {
		resource, err := s.buildSafeUserResource(r.Context(), user)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		resources = append(resources, resource)
	}
	rows, nextCursor, err := pagination.PageResources(binding, cursor, resources)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken):
		writeAPIError(w, r, userPaginationError(pagination.ReasonInvalidCursorToken))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	var nextCursorToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextCursorToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
		"users": rows,
	}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextCursorToken != nil,
		NextCursor: nextCursorToken,
	})
}

func (s *Service) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	principal, ok := admittedDeploymentAdmin(w, r)
	if !ok {
		return
	}
	request, apiErr := DecodeUserCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"auth_kind":           request.AuthKind,
		"email":               request.Email,
		"display_name":        request.DisplayName,
		"initial_password":    requestSecretFingerprint(s.keys, request.InitialPassword),
		"mfa_required":        request.MFARequired,
		"is_deployment_admin": request.IsDeploymentAdmin,
	})
	if s.replayCredentialMutation(w, r, authn.ActorOnlyRouteIdempotencyKey("users.create", principal.User.ID, request.ClientTxnID), requestHash) {
		return
	}
	workflow := beginPasswordWorkflow(w, r)
	if workflow == nil {
		return
	}
	defer workflow.Close()
	passwordHash, err := workflow.Hash(request.InitialPassword)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("initial_password", "invalid_password"))
		return
	}
	result, err := s.userAdminStore.CreateUser(
		r.Context(),
		principal.User,
		request.Email,
		request.DisplayName,
		passwordHash,
		request.MFARequired,
		request.IsDeploymentAdmin,
		request.ClientTxnID,
		requestHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	}
	if authn.IsUniqueViolation(err) {
		writeAPIError(w, r, userEmailConflictError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	payload, err := decodeStoredResponse(result.ResponseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, payload)
}

// handleGetUser reads one safe user resource. The singleton query check runs
// after the deployment-administration gate and before the user lookup (Core 01
// REQ-01-118, REQ-01-707 step 5).
func (s *Service) handleGetUser(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	user, err := s.userAdminStore.GetUserByID(r.Context(), userID)
	if errors.Is(err, authn.ErrNotFound) {
		writeAPIError(w, r, userNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	resource, err := s.buildSafeUserResource(r.Context(), user)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handlePatchUser(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := s.requireTargetUser(r.Context(), userID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeUserPatchRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	user, revokedSessions, err := s.userAdminStore.UpdateUser(
		r.Context(),
		principal.User,
		userID,
		request.BaseUserVersion,
		request.Email,
		request.DisplayName,
		request.IsActive,
		request.MFARequired,
		request.IsDeploymentAdmin,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrUserVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
		return
	case errors.Is(err, authn.ErrLastDeploymentAdmin):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "last_deployment_admin", Details: map[string]any{}})
		return
	case authn.IsUniqueViolation(err):
		writeAPIError(w, r, userEmailConflictError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.publishTargetRevocations(w, principal, revokedSessions)
	resource, err := s.buildSafeUserResource(r.Context(), user)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := s.requireTargetUser(r.Context(), userID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeAdminPasswordResetRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version": request.BaseUserVersion,
		"new_password":      requestSecretFingerprint(s.keys, request.NewPassword),
		"reason":            request.Reason,
	})
	if s.replayCredentialMutation(w, r, authn.RouteIdempotencyKey{RouteKey: "users.password.reset", ActorUserID: principal.User.ID, ScopeKey: userID.String(), ClientTxnID: request.ClientTxnID}, requestHash) {
		return
	}
	workflow := beginPasswordWorkflow(w, r)
	if workflow == nil {
		return
	}
	defer workflow.Close()
	passwordHash, err := workflow.Hash(request.NewPassword)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("new_password", "invalid_password"))
		return
	}
	result, err := s.userAdminStore.AdminResetPassword(
		r.Context(),
		principal.User,
		userID,
		request.BaseUserVersion,
		passwordHash,
		request.ClientTxnID,
		requestHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, authn.ErrUserVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.publishTargetRevocations(w, principal, result.RevokedSessionIDs)
	s.writeStoredResponse(w, r, result.ResponseJSON)
}

func (s *Service) handleResetUserTOTP(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := s.requireTargetUser(r.Context(), userID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeAdminTOTPResetRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"base_user_version": request.BaseUserVersion,
		"reason":            request.Reason,
	})
	result, err := s.userAdminStore.AdminResetTOTP(
		r.Context(),
		principal.User,
		userID,
		request.BaseUserVersion,
		request.ClientTxnID,
		requestHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, authn.ErrUserVersionConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.publishTargetRevocations(w, principal, result.RevokedSessionIDs)
	s.writeStoredResponse(w, r, result.ResponseJSON)
}

func (s *Service) handleRevokeAllUserSessions(w http.ResponseWriter, r *http.Request) {
	principal, userID, ok := admittedUserTarget(w, r)
	if !ok {
		return
	}
	if apiErr := s.requireTargetUser(r.Context(), userID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeAdminRevokeAllRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestHash := hashRequestPayload(map[string]any{
		"reason": request.Reason,
	})
	result, err := s.userAdminStore.AdminRevokeAllSessions(
		r.Context(),
		principal.User,
		userID,
		request.ClientTxnID,
		requestHash,
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	s.publishTargetRevocations(w, principal, result.RevokedSessionIDs)
	s.writeStoredResponse(w, r, result.ResponseJSON)
}

// publishTargetRevocations publishes each revoked session to its sockets and
// clears the caller's cookies when its own session was among them.
func (s *Service) publishTargetRevocations(w http.ResponseWriter, principal httpauth.Principal, sessionIDs []uuid.UUID) {
	for _, sessionID := range sessionIDs {
		s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
		if sessionID == principal.Session.ID {
			s.clearAuthCookies(w)
		}
	}
}

func (s *Service) writeStoredResponse(w http.ResponseWriter, r *http.Request, responseJSON []byte) {
	payload, err := decodeStoredResponse(responseJSON)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

// userEmailConflictError is the REQ-01-123 and REQ-01-124 rejection for an
// email that already belongs to another local user.
// requireTargetUser answers a user_id that names no user with 404
// user_not_found and empty details (Core 01 REQ-01-118). Callers run it after
// the deployment-administration check and before any idempotency, mutation,
// session or audit effect.
func (s *Service) requireTargetUser(ctx context.Context, userID uuid.UUID) *httpapi.APIError {
	if _, err := s.userAdminStore.GetUserByID(ctx, userID); errors.Is(err, authn.ErrNotFound) {
		return userNotFoundError()
	} else if err != nil {
		return internalAPIError(err)
	}
	return nil
}

func userNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}}
}

func userEmailConflictError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "user_email_conflict", Details: map[string]any{"field": "email"}}
}
