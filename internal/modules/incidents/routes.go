package incidents

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const incidentUnauthorizedCode = "session_required"

type Service struct {
	store       *Store
	authStore   *authn.Store
	hub         *platformws.Hub
	keys        authn.MasterKeys
	cursorCodec *pagination.Codec
	now         func() time.Time
}

type membershipTargetLookup interface {
	GetUserByID(context.Context, uuid.UUID) (authn.UserRecord, error)
	GetUserByNormalizedEmail(context.Context, string) (authn.UserRecord, error)
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}

		mux.HandleFunc("/api/v1/incidents", service.handleIncidentsCollection)
		mux.HandleFunc("/api/v1/incidents/", service.handleIncidentsMember)
		mux.HandleFunc("/api/v1/extensions", service.handleExtensionsCollection)
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
		store:       NewStore(deps.Postgres),
		authStore:   authn.NewStore(deps.Postgres),
		hub:         deps.WSHub,
		keys:        keys,
		cursorCodec: cursorCodec,
		now:         now,
	}, nil
}

func (s *Service) handleExtensionsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	profiles := httpapi.CurrentExtensionProfiles()
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildExtensionsResponseData(profiles))
}

func (s *Service) incidentListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (IncidentListPageRequest, string) {
	request := IncidentListPageRequest{Limit: binding.Limit + 1}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_incident_id"])
	if err != nil {
		return IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &IncidentListPosition{UpdatedAt: lastUpdatedAt.UTC(), ID: lastID}
	return request, ""
}

func buildIncidentListPage(binding pagination.Binding, anchor time.Time, records []IncidentRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, BuildIncidentResource(record))
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
		Scope:       nil,
		Position: map[string]string{
			"anchor_updated_at": anchor.UTC().Format(time.RFC3339Nano),
			"last_updated_at":   last.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"last_incident_id":  last.ID.String(),
		},
	}, nil
}

func (s *Service) handleIncidentsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
			r.URL.Query(),
			"incidents.list",
			principal.User.ID.String(),
			nil,
		)
		if reasonCode != "" {
			writeAPIError(w, r, invalidPaginationRequest(reasonCode))
			return
		}

		pageRequest, reasonCode := s.incidentListPageRequest(binding, cursor)
		if reasonCode != "" {
			writeAPIError(w, r, invalidPaginationRequest(reasonCode))
			return
		}
		records, listErr := s.store.ListVisibleIncidents(r.Context(), principal.User.ID, pageRequest)
		if listErr != nil {
			writeAPIError(w, r, internalAPIError(listErr))
			return
		}
		anchor := s.now().UTC()
		if pageRequest.AnchorUpdatedAt != nil {
			anchor = *pageRequest.AnchorUpdatedAt
		} else if len(records) > 0 {
			anchor = records[0].UpdatedAt.UTC()
		}
		rows, nextCursor, err := buildIncidentListPage(binding, anchor, records)
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
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"incidents": rows}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})

	case http.MethodPost:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeIncidentCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}

		result, err := s.store.CreateIncident(r.Context(), principal.User, request, IncidentCreateRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, ErrIncidentKeyConflict):
			writeAPIError(w, r, incidentKeyConflictError(request.IncidentKey))
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}

		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		w.Header().Set("Location", result.Location)
		_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleIncidentsMember(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/incidents/")
	if trimmed == "" || trimmed == r.URL.Path {
		http.NotFound(w, r)
		return
	}
	segments := strings.Split(trimmed, "/")
	incidentID, err := uuid.Parse(segments[0])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if len(segments) == 1 {
		switch r.Method {
		case http.MethodGet:
			if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			principal, apiErr := s.authenticateSessionRequest(r, false)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			record, err := s.store.GetVisibleIncident(r.Context(), incidentID, principal.User.ID)
			if errors.Is(err, ErrIncidentNotFound) {
				writeAPIError(w, r, incidentNotFoundError())
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
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildIncidentResource(record))

		case http.MethodPatch:
			principal, apiErr := s.authenticateSessionRequest(r, true)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "reviewer", "admin")
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			_ = membership
			request, apiErr := DecodeIncidentPatchRequest(r.Body)
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			record, _, err := s.store.UpdateIncident(r.Context(), principal.User, incidentID, request, httpapi.RequestIDFromContext(r.Context()), s.now())
			var versionConflict *IncidentVersionConflictError
			switch {
			case errors.Is(err, ErrIncidentNotFound):
				writeAPIError(w, r, incidentNotFoundError())
				return
			case errors.As(err, &versionConflict):
				writeAPIError(w, r, incidentVersionConflictError(versionConflict))
				return
			case err != nil:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildIncidentResource(record))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	switch strings.Join(segments[1:], "/") {
	case "memberships":
		s.handleMembershipsCollection(w, r, incidentID)
		return
	case "workbook-preferences/default":
		s.handleIncidentWorkbookPreferencesDefault(w, r, incidentID)
		return
	case "workbook-preferences/me":
		s.handleIncidentWorkbookPreferencesMe(w, r, incidentID)
		return
	}

	if len(segments) == 3 && segments[1] == "memberships" {
		userID, err := uuid.Parse(segments[2])
		if err != nil {
			http.NotFound(w, r)
			return
		}
		s.handleMembershipMember(w, r, incidentID, userID)
		return
	}

	http.NotFound(w, r)
}

func (s *Service) handleMembershipsCollection(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
			r.URL.Query(),
			"incident.memberships.list",
			principal.User.ID.String(),
			map[string]string{"incident_id": incidentID.String()},
		)
		if reasonCode != "" {
			writeAPIError(w, r, invalidPaginationRequest(reasonCode))
			return
		}

		records, listErr := s.store.ListMemberships(r.Context(), incidentID)
		if listErr != nil {
			writeAPIError(w, r, internalAPIError(listErr))
			return
		}
		memberships := make([]map[string]any, 0, len(records))
		for _, record := range records {
			memberships = append(memberships, BuildMembershipResource(record))
		}
		rows, nextCursor, err := pagination.PageResources(binding, cursor, memberships)
		switch {
		case errors.Is(err, pagination.ErrInvalidCursorToken):
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
			return
		case err != nil:
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
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"memberships": rows}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})

	case http.MethodPost:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeMembershipCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		targetUser, apiErr := s.resolveMembershipTarget(r.Context(), request)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}

		requestHash := hashRequestPayload(map[string]any{
			"client_txn_id": request.ClientTxnID,
			"user_id":       request.UserID,
			"email":         request.Email,
			"role":          request.Role,
		})
		result, err := s.store.CreateMembership(r.Context(), principal.User, incidentID, targetUser, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, ErrMembershipExistsUsePatch):
			writeAPIError(w, r, membershipExistsUsePatchError())
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

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleMembershipMember(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, userID uuid.UUID) {
	switch r.Method {
	case http.MethodPatch:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeMembershipPatchRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, _, err := s.store.UpdateMembership(r.Context(), principal.User, incidentID, userID, request, httpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case errors.Is(err, ErrMembershipNotFound):
			writeAPIError(w, r, membershipNotFoundError())
			return
		case errors.Is(err, ErrMembershipVersionConflict):
			writeAPIError(w, r, membershipVersionConflictError())
			return
		case errors.Is(err, ErrLastIncidentAdmin):
			writeAPIError(w, r, lastIncidentAdminError())
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildMembershipResource(record))

	case http.MethodDelete:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeMembershipDeleteRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.store.DeleteMembership(r.Context(), principal.User, incidentID, userID, request, httpapi.RequestIDFromContext(r.Context())); err != nil {
			switch {
			case errors.Is(err, ErrMembershipNotFound):
				writeAPIError(w, r, membershipNotFoundError())
				return
			case errors.Is(err, ErrMembershipVersionConflict):
				writeAPIError(w, r, membershipVersionConflictError())
				return
			case errors.Is(err, ErrLastIncidentAdmin):
				writeAPIError(w, r, lastIncidentAdminError())
				return
			default:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		s.revokeIncidentAccess(r.Context(), incidentID, userID)
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleIncidentWorkbookPreferencesDefault(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	switch r.Method {
	case http.MethodGet:
		if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.store.GetIncidentWorkbookPreferences(r.Context(), incidentID)
		if errors.Is(err, ErrIncidentNotFound) {
			writeAPIError(w, r, incidentNotFoundError())
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
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildDefaultWorkbookPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeDefaultWorkbookPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.store.PutIncidentWorkbookPreferences(r.Context(), incidentID, principal.User.ID, request.DefaultSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildDefaultWorkbookPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleIncidentWorkbookPreferencesMe(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	switch r.Method {
	case http.MethodGet:
		if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		principal, apiErr := s.authenticateSessionRequest(r, false)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.store.GetUserWorkbookPreferences(r.Context(), incidentID, principal.User.ID)
		if errors.Is(err, ErrIncidentNotFound) {
			writeAPIError(w, r, incidentNotFoundError())
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
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildUserWorkbookPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := s.authenticateSessionRequest(r, true)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeUserWorkbookPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.store.PutUserWorkbookPreferences(r.Context(), incidentID, principal.User.ID, request.HomeSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, BuildUserWorkbookPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, *auth.APIError) {
	record, err := s.store.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, ErrMembershipNotFound) {
		return MembershipRecord{}, IncidentAccessError(nil, false)
	}
	if err != nil {
		return MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (MembershipRecord, *auth.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return MembershipRecord{}, apiErr
	}
	if apiErr := IncidentAccessError(&record, false, roles...); apiErr != nil {
		return MembershipRecord{}, apiErr
	}
	return record, nil
}

func (s *Service) revokeIncidentAccess(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) {
	if s.hub == nil {
		return
	}
	sessions, err := s.authStore.ListActiveSessionsForUser(ctx, userID)
	if err != nil {
		return
	}
	for _, session := range sessions {
		s.hub.RevokeIncidentAccess(incidentID, session.ID)
	}
}

func (s *Service) resolveMembershipTarget(ctx context.Context, request MembershipCreateRequest) (authn.UserRecord, *auth.APIError) {
	return resolveMembershipTarget(ctx, s.authStore, request)
}

func resolveMembershipTarget(ctx context.Context, lookup membershipTargetLookup, request MembershipCreateRequest) (authn.UserRecord, *auth.APIError) {
	var (
		user authn.UserRecord
		err  error
	)
	if request.UserID != nil {
		user, err = lookup.GetUserByID(ctx, *request.UserID)
	} else {
		user, err = lookup.GetUserByNormalizedEmail(ctx, *request.Email)
	}
	if errors.Is(err, authn.ErrNotFound) {
		return authn.UserRecord{}, userNotFoundError()
	}
	if err != nil {
		return authn.UserRecord{}, internalAPIError(err)
	}
	if !user.IsActive {
		return authn.UserRecord{}, userInactiveError()
	}
	return user, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
		}
		if principal, apiErr := s.authenticateSessionToken(r, auth.AuthSourceBearer, token, stateChanging); apiErr == nil {
			return principal, nil
		} else if apiErr.Code != incidentUnauthorizedCode {
			return auth.SessionPrincipal{}, apiErr
		}

		bootstrapToken, _, err := s.authStore.GetBootstrapTokenByFingerprint(r.Context(), authn.FingerprintToken(s.keys, token))
		if err == nil {
			if reason := authn.BootstrapReasonCode(bootstrapToken, s.now()); reason != "" {
				return auth.SessionPrincipal{}, auth.BootstrapRejectedError(reason)
			}
			return auth.SessionPrincipal{}, auth.BootstrapRejectedError("not_allowed_for_route")
		}
		if err != nil && !errors.Is(err, authn.ErrNotFound) {
			return auth.SessionPrincipal{}, internalAPIError(err)
		}
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
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
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
	}
	if err != nil {
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
	}

	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.authStore.RevokeSession(context.Background(), session.ID, "session_expired", now)
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: incidentUnauthorizedCode, Details: map[string]any{}}
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

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
