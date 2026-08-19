package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type Application interface {
	ListVisibleIncidents(context.Context, uuid.UUID, incidents.IncidentListPageRequest) ([]incidents.IncidentRecord, error)
	GetVisibleIncident(context.Context, uuid.UUID, uuid.UUID) (incidents.IncidentRecord, error)
	CreateIncident(context.Context, authn.UserRecord, incidents.CreateIncidentRequest, string, time.Time) (incidents.CreateIncidentResult, error)
	UpdateIncident(context.Context, authn.UserRecord, uuid.UUID, incidents.IncidentPatchRequest, string, time.Time) (incidents.IncidentRecord, bool, error)
	ListMemberships(context.Context, uuid.UUID) ([]incidents.MembershipRecord, error)
	CreateMembership(context.Context, authn.UserRecord, uuid.UUID, authn.UserRecord, incidents.MembershipCreateRequest, string, time.Time) (incidents.MembershipCreateResult, error)
	UpdateMembership(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, incidents.MembershipPatchRequest, string, time.Time) (incidents.MembershipRecord, bool, error)
	ListAdministrativeAuditEvents(context.Context, administrativeaudit.ListFilter) ([]administrativeaudit.Record, error)
}

type AdmissionChecker interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

type TerminalMutationCoordinator interface {
	CoordinateIncidentLifecycle(context.Context, authn.UserRecord, uuid.UUID, string, incidents.IncidentLifecycleRequest, string, time.Time) (incidents.IncidentLifecycleResult, error)
	CoordinateMembershipDeletion(context.Context, authn.UserRecord, uuid.UUID, uuid.UUID, incidents.MembershipDeleteRequest, string) (incidents.MembershipDeleteResult, error)
}

type Dependencies struct {
	Application                 Application
	AdmissionChecker            AdmissionChecker
	TerminalMutationCoordinator TerminalMutationCoordinator
}

type service struct {
	application       Application
	terminalMutations TerminalMutationCoordinator
	incidentAdmission AdmissionChecker
	authStore         *authn.Store
	keys              authn.MasterKeys
	cursorCodec       *pagination.Codec
	now               func() time.Time
}

type membershipTargetLookup interface {
	GetUserByID(context.Context, uuid.UUID) (authn.UserRecord, error)
	GetUserByNormalizedEmail(context.Context, string) (authn.UserRecord, error)
}

func RegisterRoutes(dependencies Dependencies) platformhttpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, dependencies)
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

func newService(deps platformhttpapi.DependencySet, dependencies Dependencies) (*service, error) {
	if isNilRouteDependency(dependencies.Application) {
		return nil, errors.New("incidents: application is required for route registration")
	}
	if isNilRouteDependency(dependencies.AdmissionChecker) {
		return nil, errors.New("incidents: admission checker is required for route registration")
	}
	if isNilRouteDependency(dependencies.TerminalMutationCoordinator) {
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
	return &service{
		application:       dependencies.Application,
		terminalMutations: dependencies.TerminalMutationCoordinator,
		incidentAdmission: dependencies.AdmissionChecker,
		authStore:         authn.NewStore(deps.PostgresHandle()),
		keys:              keys,
		cursorCodec:       cursorCodec,
		now:               now,
	}, nil
}

func isNilRouteDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
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

func (s *service) incidentListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (incidents.IncidentListPageRequest, string) {
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

func (s *service) handleIncidentsCollection(w http.ResponseWriter, r *http.Request) {
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
		request, apiErr := decodeIncidentCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}

		result, err := s.application.CreateIncident(r.Context(), principal.User, request, platformhttpapi.RequestIDFromContext(r.Context()), s.now())
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

func (s *service) handleIncidentsMember(w http.ResponseWriter, r *http.Request) {
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
			request, apiErr := decodeIncidentPatchRequest(r.Body)
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

func (s *service) handleIncidentLifecycle(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, action string) {
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
	request, apiErr := decodeIncidentLifecycleRequest(r.Body)
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

func (s *service) handleMembershipsCollection(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID) {
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
		request, apiErr := decodeMembershipCreateRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		targetUser, apiErr := s.resolveMembershipTarget(r.Context(), request)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}

		result, err := s.application.CreateMembership(r.Context(), principal.User, incidentID, targetUser, request, platformhttpapi.RequestIDFromContext(r.Context()), s.now())
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

func (s *service) handleMembershipMember(w http.ResponseWriter, r *http.Request, incidentID uuid.UUID, userID uuid.UUID) {
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
		request, apiErr := decodeMembershipPatchRequest(r.Body)
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
		request, apiErr := decodeMembershipDeleteRequest(r.Body)
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

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *platformhttpapi.APIError) {
	return s.requireIncidentRole(ctx, incidentID, userID, admission.RolesMember, "")
}

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
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

func (s *service) resolveMembershipTarget(ctx context.Context, request incidents.MembershipCreateRequest) (authn.UserRecord, *platformhttpapi.APIError) {
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

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
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
