package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type Service struct {
	application       *incidents.Application
	terminalMutations incidents.TerminalMutationCoordinator
	incidentAdmission *admission.Checker
	authStore         *authn.Store
	keys              authn.MasterKeys
	cursorCodec       *pagination.Codec
	now               func() time.Time
}

type RouteOptions struct {
	Application       *incidents.Application
	TerminalMutations incidents.TerminalMutationCoordinator
}

type membershipTargetLookup interface {
	GetUserByID(context.Context, uuid.UUID) (authn.UserRecord, error)
	GetUserByNormalizedEmail(context.Context, string) (authn.UserRecord, error)
}

func RegisterRoutes(options ...RouteOptions) platformhttpapi.RouteRegistrar {
	routeOptions := RouteOptions{}
	if len(options) > 0 {
		routeOptions = options[0]
	}
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, routeOptions)
		if err != nil {
			return err
		}
		return platformhttpapi.BindOwnerRoutes(mux, deps, "module.incidents", map[string]http.HandlerFunc{
			"closeIncident":                     service.handleIncidentsMember,
			"createIncident":                    service.handleIncidentsCollection,
			"createIncidentMembership":          service.handleIncidentsMember,
			"deleteIncidentMembership":          service.handleIncidentsMember,
			"getIncident":                       service.handleIncidentsMember,
			"listIncidentMembershipAuditEvents": service.handleIncidentsMember,
			"listIncidentMemberships":           service.handleIncidentsMember,
			"listVisibleIncidents":              service.handleIncidentsCollection,
			"patchIncident":                     service.handleIncidentsMember,
			"patchIncidentMembership":           service.handleIncidentsMember,
			"reopenIncident":                    service.handleIncidentsMember,
		})
	}
}

func newService(deps platformhttpapi.DependencySet, options RouteOptions) (*Service, error) {
	if options.Application == nil {
		return nil, errors.New("incidents: application is required for route registration")
	}
	if options.TerminalMutations == nil {
		return nil, errors.New("incidents: terminal mutation coordinator is required for route registration")
	}
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
		application:       options.Application,
		terminalMutations: options.TerminalMutations,
		incidentAdmission: admission.NewChecker(deps.PostgresHandle()),
		authStore:         authn.NewStore(deps.PostgresHandle()),
		keys:              keys,
		cursorCodec:       cursorCodec,
		now:               now,
	}, nil
}

func parseIncidentListScope(rawQuery string) (listquery.Result, *platformhttpapi.APIError) {
	result, queryErr := listquery.Parse(rawQuery, listquery.Config{
		Search: true,
		ExactFilters: map[string]listquery.ExactFilter{
			"status": {Allowed: []string{"active", "closed"}},
		},
	})
	if queryErr == nil {
		return result, nil
	}
	if queryErr.Kind == listquery.ErrorKindPagination {
		return listquery.Result{}, invalidPaginationRequest(queryErr.ReasonCode)
	}
	return listquery.Result{}, invalidListQuery(queryErr.ReasonCode)
}

func (s *Service) incidentListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (incidents.IncidentListPageRequest, string) {
	request := incidents.IncidentListPageRequest{
		Limit:        binding.Limit + 1,
		SearchTokens: strings.Fields(binding.Scope["search"]),
	}
	if status := binding.Scope["status"]; status != "" {
		request.Status = &status
	}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_incident_id"])
	if err != nil {
		return incidents.IncidentListPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &incidents.IncidentListPosition{UpdatedAt: lastUpdatedAt.UTC(), ID: lastID}
	return request, ""
}

func buildIncidentListPage(binding pagination.Binding, anchor time.Time, records []incidents.IncidentRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, incidents.BuildIncidentResource(record))
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
		Scope:       binding.Scope,
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
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		listScope, apiErr := parseIncidentListScope(r.URL.RawQuery)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		binding, cursor, reasonCode := s.cursorCodec.ResolveListRequest(
			listScope.Values,
			"incidents.list",
			principal.User.ID.String(),
			listScope.Scope,
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
		records, listErr := s.application.ListVisibleIncidents(r.Context(), principal.User.ID, pageRequest)
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
		_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"incidents": rows}, platformhttpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})

	case http.MethodPost:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeIncidentCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}

		result, err := s.application.CreateIncident(r.Context(), principal.User, request, incidents.IncidentCreateRequestHash(request), platformhttpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID))
			return
		case errors.Is(err, incidents.ErrIncidentKeyConflict):
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
		w.Header().Set("Location", incidentLocation(result.Incident.ID))
		status := http.StatusOK
		if result.Created {
			status = http.StatusCreated
		}
		_ = platformhttpapi.WriteSuccess(w, r, status, result.Payload)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func incidentLocation(incidentID uuid.UUID) string {
	return "/api/v1/incidents/" + incidentID.String()
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
			if apiErr := platformhttpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			record, err := s.application.GetVisibleIncident(r.Context(), incidentID, principal.User.ID)
			if errors.Is(err, incidents.ErrIncidentNotFound) {
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
			_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildIncidentResource(record))

		case http.MethodPatch:
			principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
			if apiErr != nil {
				writeAPIError(w, r, apiErr)
				return
			}
			membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesReviewerAdmin, "reviewer|admin")
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
			record, _, err := s.application.UpdateIncident(r.Context(), principal.User, incidentID, request, platformhttpapi.RequestIDFromContext(r.Context()), s.now())
			var versionConflict *incidents.IncidentVersionConflictError
			switch {
			case admission.IsDenied(err, admission.DenialNotVisible):
				writeAPIError(w, r, incidentNotFoundError())
				return
			case admission.IsDenied(err, admission.DenialInsufficientRole):
				writeAPIError(w, r, authorizationDeniedError("reviewer|admin"))
				return
			case admission.IsDenied(err, admission.DenialIncidentClosed):
				writeAPIError(w, r, incidentClosedError())
				return
			case errors.Is(err, incidents.ErrIncidentNotFound):
				writeAPIError(w, r, incidentNotFoundError())
				return
			case errors.As(err, &versionConflict):
				writeAPIError(w, r, incidentVersionConflictError(versionConflict))
				return
			case errors.Is(err, incidents.ErrIncidentClosed):
				writeAPIError(w, r, incidentClosedError())
				return
			case err != nil:
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildIncidentResource(record))

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	if len(segments) == 2 && (segments[1] == "close" || segments[1] == "reopen") {
		s.handleIncidentLifecycle(w, r, incidentID, segments[1])
		return
	}
	if len(segments) == 2 && segments[1] == "membership-audit-events" {
		s.handleMembershipAuditEvents(w, r, incidentID)
		return
	}

	switch strings.Join(segments[1:], "/") {
	case "memberships":
		s.handleMembershipsCollection(w, r, incidentID)
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

func (s *Service) handleIncidentLifecycle(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, action string) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeIncidentLifecycleRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.terminalMutations.CoordinateIncidentLifecycle(
		r.Context(),
		principal.User,
		incidentID,
		action,
		request,
		incidents.IncidentLifecycleRequestHash(action, request),
		platformhttpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	var versionConflict *incidents.IncidentVersionConflictError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case admission.IsDenied(err, admission.DenialNotVisible):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		writeAPIError(w, r, authorizationDeniedError("admin"))
		return
	case errors.Is(err, incidents.ErrIncidentNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.As(err, &versionConflict):
		writeAPIError(w, r, incidentVersionConflictError(versionConflict))
		return
	case errors.Is(err, incidents.ErrIncidentIllegalTransition):
		writeAPIError(w, r, incidentIllegalTransitionError(action))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleMembershipsCollection(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
	switch r.Method {
	case http.MethodGet:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
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

		records, listErr := s.application.ListMemberships(r.Context(), incidentID)
		if listErr != nil {
			writeAPIError(w, r, internalAPIError(listErr))
			return
		}
		memberships := make([]map[string]any, 0, len(records))
		for _, record := range records {
			memberships = append(memberships, incidents.BuildMembershipResource(record))
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
		_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"memberships": rows}, platformhttpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})

	case http.MethodPost:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
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

		requestHash := incidents.MembershipCreateRequestHash(request)
		result, err := s.application.CreateMembership(r.Context(), principal.User, incidentID, targetUser, request, requestHash, platformhttpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case errors.Is(err, authn.ErrClientTxnConflict):
			writeAPIError(w, r, platformhttpapi.ClientTxnConflictError(request.ClientTxnID))
			return
		case admission.IsDenied(err, admission.DenialNotVisible):
			writeAPIError(w, r, incidentNotFoundError())
			return
		case admission.IsDenied(err, admission.DenialInsufficientRole):
			writeAPIError(w, r, authorizationDeniedError("admin"))
			return
		case errors.Is(err, incidents.ErrMembershipExistsUsePatch):
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
		status := http.StatusOK
		if result.Created {
			status = http.StatusCreated
		}
		_ = platformhttpapi.WriteSuccess(w, r, status, result.Payload)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleMembershipMember(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, userID uuid.UUID) {
	switch r.Method {
	case http.MethodPatch:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeMembershipPatchRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, _, err := s.application.UpdateMembership(r.Context(), principal.User, incidentID, userID, request, platformhttpapi.RequestIDFromContext(r.Context()), s.now())
		switch {
		case admission.IsDenied(err, admission.DenialNotVisible):
			writeAPIError(w, r, incidentNotFoundError())
			return
		case admission.IsDenied(err, admission.DenialInsufficientRole):
			writeAPIError(w, r, authorizationDeniedError("admin"))
			return
		case errors.Is(err, incidents.ErrMembershipNotFound):
			writeAPIError(w, r, membershipNotFoundError())
			return
		case errors.Is(err, incidents.ErrMembershipVersionConflict):
			writeAPIError(w, r, membershipVersionConflictError())
			return
		case errors.Is(err, incidents.ErrLastIncidentAdmin):
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
		_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, incidents.BuildMembershipResource(record))

	case http.MethodDelete:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, admission.RolesAdmin, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeMembershipDeleteRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, err := s.terminalMutations.CoordinateMembershipDeletion(r.Context(), principal.User, incidentID, userID, request, platformhttpapi.RequestIDFromContext(r.Context())); err != nil {
			switch {
			case admission.IsDenied(err, admission.DenialNotVisible):
				writeAPIError(w, r, incidentNotFoundError())
				return
			case admission.IsDenied(err, admission.DenialInsufficientRole):
				writeAPIError(w, r, authorizationDeniedError("admin"))
				return
			case errors.Is(err, incidents.ErrMembershipNotFound):
				writeAPIError(w, r, membershipNotFoundError())
				return
			case errors.Is(err, incidents.ErrMembershipVersionConflict):
				writeAPIError(w, r, membershipVersionConflictError())
				return
			case errors.Is(err, incidents.ErrLastIncidentAdmin):
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
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *platformhttpapi.APIError) {
	return s.requireIncidentRole(ctx, incidentID, userID, admission.RolesMember, "")
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
	grant, err := s.incidentAdmission.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	return incidentAdmissionResult(grant, err, requiredRole)
}

func incidentAdmissionResult(grant admission.Grant, err error, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
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

func (s *Service) resolveMembershipTarget(ctx context.Context, request incidents.MembershipCreateRequest) (authn.UserRecord, *platformhttpapi.APIError) {
	return resolveMembershipTarget(ctx, s.authStore, request)
}

func resolveMembershipTarget(ctx context.Context, lookup membershipTargetLookup, request incidents.MembershipCreateRequest) (authn.UserRecord, *platformhttpapi.APIError) {
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

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *platformhttpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = platformhttpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
