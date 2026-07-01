package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func parseUsersListScope(rawQuery string) (listquery.Result, *APIError) {
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

func (s *Service) handleUsersCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
			writeAPIError(w, r, apiErr)
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

		users, listErr := s.store.ListUsers(r.Context(), userListFilterFromScope(binding.Scope))
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
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
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
	case http.MethodPost:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
			writeAPIError(w, r, apiErr)
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
		passwordHash, err := authn.HashPassword(request.InitialPassword)
		if err != nil {
			writeAPIError(w, r, invalidMutationPayload("initial_password", "invalid_password"))
			return
		}
		result, err := s.store.CreateUser(
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
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		}
		if authn.IsUniqueViolation(err) {
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_mutation_payload", Details: map[string]any{"field": "email"}})
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
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleUsersMember(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/users/")
	if trimmed == "" || trimmed == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	segments := strings.Split(trimmed, "/")
	userID, err := uuid.Parse(segments[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if match, ok := httpapi.MatchReservedExtensionFamilyIn(s.profiles, r.URL.Path); ok && !match.Claimed {
		writeAPIError(w, r, &APIError{
			Status: http.StatusNotFound,
			Code:   "extension_profile_not_claimed",
			Details: map[string]any{
				"profile_id":   match.ProfileID,
				"route_family": match.RouteFamily,
			},
		})
		return
	}

	if s.handleEnterpriseAuthBindings(w, r, userID, segments) {
		return
	}

	if len(segments) == 1 {
		switch r.Method {
		case http.MethodGet:
			if apiErr := ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			principal, apiErr := s.authenticateSessionRequest(r, false)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			user, err := s.store.GetUserByID(r.Context(), userID)
			if errors.Is(err, authn.ErrNotFound) {
				writeAPIError(w, r, &APIError{Status: http.StatusNotFound, Code: "user_not_found", Details: map[string]any{}})
				return
			}
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			resource, err := s.buildSafeUserResource(r.Context(), user)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
		case http.MethodPatch:
			principal, apiErr := s.authenticateSessionRequest(r, true)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			request, apiErr := DecodeUserPatchRequest(r.Body)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			user, revokedSessions, err := s.store.UpdateUser(
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
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
				return
			case errors.Is(err, authn.ErrLastDeploymentAdmin):
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "last_deployment_admin", Details: map[string]any{}})
				return
			case authn.IsUniqueViolation(err):
				writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "invalid_mutation_payload", Details: map[string]any{"field": "email"}})
				return
			case err != nil:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			for _, sessionID := range revokedSessions {
				s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
				if sessionID == principal.Session.ID {
					s.clearAuthCookies(w)
				}
			}
			resource, err := s.buildSafeUserResource(r.Context(), user)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(segments) < 2 || r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := RequireDeploymentAdmin(principal.User); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	switch strings.Join(segments[1:], "/") {
	case "password/reset":
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
		passwordHash, err := authn.HashPassword(request.NewPassword)
		if err != nil {
			writeAPIError(w, r, invalidMutationPayload("new_password", "invalid_password"))
			return
		}
		result, err := s.store.AdminResetPassword(
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
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, authn.ErrUserVersionConflict):
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	case "mfa/totp/reset":
		request, apiErr := DecodeAdminTOTPResetRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"base_user_version": request.BaseUserVersion,
			"reason":            request.Reason,
		})
		result, err := s.store.AdminResetTOTP(
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
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, authn.ErrUserVersionConflict):
			writeAPIError(w, r, &APIError{Status: http.StatusConflict, Code: "user_version_conflict", Details: map[string]any{}})
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	case "sessions/revoke-all":
		request, apiErr := DecodeAdminRevokeAllRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		requestHash := hashRequestPayload(map[string]any{
			"reason": request.Reason,
		})
		result, err := s.store.AdminRevokeAllSessions(
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
			writeAPIError(w, r, ClientTxnConflictError(request.ClientTxnID))
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		for _, sessionID := range result.RevokedSessionIDs {
			s.publishSessionRevocation(sessionID, sessionRevokedReasonCode)
			if sessionID == principal.Session.ID {
				s.clearAuthCookies(w)
			}
		}
		payload, err := decodeStoredResponse(result.ResponseJSON)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	default:
		http.NotFound(w, r)
	}
}
