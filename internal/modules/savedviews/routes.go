package savedviews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type service struct {
	application    *savedViewApplication
	incidentAccess *admission.Checker
	authStore      *authn.Store
	cursorCodec    *pagination.Codec
	now            func() time.Time
}

// RegisterRoutes binds the saved-view routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits incident_id and saved_view_id before a handler runs.
// Every resolved response of the two reads, failures included, carries
// Cache-Control: no-store (REQ-01-144B).
func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.savedviews", map[string]http.HandlerFunc{
			"createIncidentSavedView": service.handleCreate,
			"deleteIncidentSavedView": service.handleDelete,
			"getIncidentSavedView":    service.handleGet,
			"listIncidentSavedViews":  service.handleList,
			"patchIncidentSavedView":  service.handlePatch,
		}, httpapi.NoStore("getIncidentSavedView", "listIncidentSavedViews"))
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		return httpapi.BindTestRoute(mux, deps, guard, http.MethodPost, "/api/v1/test/incidents/{incident_id}/saved-views/system", service.handleTestSystemCreate)
	}
}

func newService(deps httpapi.DependencySet) (*service, error) {
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	cursorCodec := deps.CursorCodec
	if cursorCodec == nil {
		keys, err := authn.LoadMasterKeys(deps.Env)
		if err != nil {
			return nil, err
		}
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	return &service{
		application:    newSavedViewApplication(newPostgresSavedViewRepository(deps.PostgresHandle())),
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		cursorCodec:    cursorCodec,
		now:            now,
	}, nil
}

func (s *service) handleTestSystemCreate(w http.ResponseWriter, r *http.Request) {
	incidentID, err := uuid.Parse(r.PathValue("incident_id"))
	if err != nil {
		writeAPIError(w, r, httpapi.NotFound("incident_not_found"))
		return
	}
	request, apiErr := decodeSystemFixtureRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.createSystemFixture(r.Context(), incidentID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, buildResource(record))
}

func (s *service) handleList(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the list query and cursor are route handling, after
	// the incident's visibility (REQ-01-704).
	binding, cursor, apiErr := resolveSavedViewListRead(s.cursorCodec, r.URL.RawQuery, principal.User.ID.String(), incidentID.String())
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	pageRequest, reasonCode := savedViewListPageRequest(binding, cursor)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	pageRequest.ViewSchemaID = binding.Scope["view_schema_id"]
	records, err := s.application.listVisible(r.Context(), incidentID, principal.User.ID, pageRequest)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	anchor := s.now().UTC()
	if pageRequest.AnchorUpdatedAt != nil {
		anchor = *pageRequest.AnchorUpdatedAt
	} else if len(records) > 0 {
		anchor = records[0].UpdatedAt.UTC()
	}
	rows, nextCursor, err := buildSavedViewListPage(binding, anchor, records)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
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
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"saved_views": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *service) handleGet(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, savedViewID, apiErr := admittedSavedViewRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-707 step 5: the singleton query is route handling, after the
	// incident's visibility (REQ-01-704).
	if apiErr := validateSavedViewDetailQuery(r.URL.RawQuery); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.getVisible(r.Context(), incidentID, savedViewID, principal.User.ID)
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildResource(record))
}

func (s *service) handleCreate(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.create(r.Context(), incidentID, principal.User.ID, request, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, buildResource(record))
}

func (s *service) handlePatch(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, savedViewID, apiErr := admittedSavedViewRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	current, err := s.application.visibleForPatch(r.Context(), incidentID, savedViewID, principal.User.ID)
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	request, apiErr := decodePatchRequest(r.Body, current.ViewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.application.patch(r.Context(), incidentID, savedViewID, principal.User.ID, membership.Role.String(), request, s.now())
	if err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildResource(record))
}

func (s *service) handleDelete(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, savedViewID, apiErr := admittedSavedViewRoute(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.application.delete(r.Context(), incidentID, savedViewID, principal.User.ID, membership.Role.String()); err != nil {
		writeAPIError(w, r, savedViewError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"saved_view_id": savedViewID,
		"deleted":       true,
	})
}

func savedViewListPageRequest(binding pagination.Binding, cursor *pagination.Cursor) (listPageRequest, string) {
	request := listPageRequest{Limit: binding.Limit + 1}
	if cursor == nil {
		return request, ""
	}
	if cursor.Mode != pagination.ModeKeyset {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor, err := time.Parse(time.RFC3339Nano, cursor.Position["anchor_updated_at"])
	if err != nil {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastUpdatedAt, err := time.Parse(time.RFC3339Nano, cursor.Position["last_updated_at"])
	if err != nil {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	lastID, err := uuid.Parse(cursor.Position["last_saved_view_id"])
	if err != nil || lastUpdatedAt.After(anchor) || len(cursor.Position) != 3 {
		return listPageRequest{}, pagination.ReasonInvalidCursorToken
	}
	anchor = anchor.UTC()
	request.AnchorUpdatedAt = &anchor
	request.After = &listPosition{UpdatedAt: lastUpdatedAt.UTC(), SavedViewID: lastID}
	return request, ""
}

func buildSavedViewListPage(binding pagination.Binding, anchor time.Time, records []savedViewRecord) ([]json.RawMessage, *pagination.Cursor, error) {
	hasMore := len(records) > binding.Limit
	pageRecords := records
	if hasMore {
		pageRecords = records[:binding.Limit]
	}
	resources := make([]map[string]any, 0, len(pageRecords))
	for _, record := range pageRecords {
		resources = append(resources, buildResource(record))
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
			"anchor_updated_at":  anchor.UTC().Format(time.RFC3339Nano),
			"last_updated_at":    last.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"last_saved_view_id": last.SavedViewID.String(),
		},
	}, nil
}

func savedViewError(err error) *httpapi.APIError {
	var versionConflict *savedViewVersionConflictError
	switch {
	case errors.Is(err, errSavedViewNotFound):
		return savedViewNotFoundError()
	case errors.Is(err, errSavedViewMutationDenied):
		return authorizationDeniedError()
	case errors.As(err, &versionConflict):
		return savedViewVersionConflictAPIError(versionConflict)
	case errors.Is(err, errSavedViewVersionConflict):
		return savedViewVersionConflictAPIError(nil)
	default:
		return internalAPIError(err)
	}
}

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, savedViewNotFoundError()
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

// admittedIncidentRoute returns the principal and incident_id that central
// admission established (REQ-01-707 steps 2 and 3); their absence is a
// binding error.
func admittedIncidentRoute(r *http.Request) (httpauth.Principal, uuid.UUID, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	incidentID, identified := httpapi.PathUUID(r, "incident_id")
	if !authenticated || !identified {
		return httpauth.Principal{}, uuid.Nil, internalAPIError(errors.New("saved-view route was not admitted centrally"))
	}
	return principal, incidentID, nil
}

// admittedSavedViewRoute also returns the admitted saved_view_id.
func admittedSavedViewRoute(r *http.Request) (httpauth.Principal, uuid.UUID, uuid.UUID, *httpapi.APIError) {
	principal, incidentID, apiErr := admittedIncidentRoute(r)
	if apiErr != nil {
		return httpauth.Principal{}, uuid.Nil, uuid.Nil, apiErr
	}
	savedViewID, identified := httpapi.PathUUID(r, "saved_view_id")
	if !identified {
		return httpauth.Principal{}, uuid.Nil, uuid.Nil, internalAPIError(errors.New("saved-view route was not admitted centrally"))
	}
	return principal, incidentID, savedViewID, nil
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
