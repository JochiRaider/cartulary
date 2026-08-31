package imports

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type importRouteDescriptor struct {
	Kind          string
	Method        string
	OperationID   string
	StateChanging bool
	Collection    bool
}

var importRouteCatalog = [...]importRouteDescriptor{
	{Kind: "create_session", Method: http.MethodPost, OperationID: "createImportSession", StateChanging: true, Collection: true},
	{Kind: "session", Method: http.MethodGet, OperationID: "getImportSession"},
	{Kind: "units", Method: http.MethodGet, OperationID: "listImportUnits"},
	{Kind: "unit", Method: http.MethodGet, OperationID: "getImportUnit"},
	{Kind: "preview", Method: http.MethodGet, OperationID: "getImportUnitPreview"},
	{Kind: "mapping_preview", Method: http.MethodPost, OperationID: "previewImportUnitExtensionMapping"},
	{Kind: "mapping", Method: http.MethodPut, OperationID: "putImportUnitMapping", StateChanging: true},
	{Kind: "select", Method: http.MethodPost, OperationID: "selectImportUnit", StateChanging: true},
	{Kind: "skip", Method: http.MethodPost, OperationID: "skipImportUnit", StateChanging: true},
	{Kind: "regions", Method: http.MethodPost, OperationID: "createImportUnitRegion", StateChanging: true},
	{Kind: "apply", Method: http.MethodPost, OperationID: "applyImportSession", StateChanging: true},
}

func importRouteByKind(kind string) (importRouteDescriptor, bool) {
	for _, route := range importRouteCatalog {
		if route.Kind == kind {
			return route, true
		}
	}
	return importRouteDescriptor{}, false
}

func (s *service) importOperationHandlers() map[string]http.HandlerFunc {
	handlers := make(map[string]http.HandlerFunc, len(importRouteCatalog))
	for _, route := range importRouteCatalog {
		handler := s.handleImportSessionsMember
		if route.Collection {
			handler = s.handleImportSessionsCollection
		}
		handlers[route.OperationID] = handler
	}
	return handlers
}

func bindOwnerRoutes(
	mux *http.ServeMux,
	deps httpapi.DependencySet,
	service *service,
) error {
	return httpapi.BindOwnerRoutes(mux, deps, "module.imports", service.importOperationHandlers())
}

func (s *service) handleImportSessionsCollection(w http.ResponseWriter, r *http.Request) {
	route, ok := importRouteByKind("create_session")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != route.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: route.StateChanging})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: importSessionFileContentTypes})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := decodeCreateSessionMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err, apiErr := s.applicationCreateSession(r.Context(), principal.User.ID, envelope, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
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
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleImportSessionsMember(w http.ResponseWriter, r *http.Request) {
	route, ok := parseImportSessionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	descriptor, ok := importRouteByKind(route.Kind)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != descriptor.Method {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{
		Store:         s.authStore,
		Keys:          s.keys,
		Now:           s.now,
		StateChanging: descriptor.StateChanging,
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch route.Kind {
	case "session":
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.applicationGetSession(r.Context(), route.SessionID)
		if errors.Is(err, errNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), result.IncidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Value)
	case "units":
		binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
			r.URL.Query(),
			"imports.units.list",
			principal.User.ID.String(),
			map[string]string{"import_session_id": route.SessionID.String()},
		)
		if reasonCode != "" {
			writeAPIError(w, r, invalidPaginationRequest(reasonCode))
			return
		}
		result, err := s.applicationListUnits(r.Context(), route.SessionID)
		if errors.Is(err, errNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), result.IncidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		rows, nextCursor, err := pagination.PageResources(binding, cursor, result.Value)
		if errors.Is(err, pagination.ErrInvalidCursorToken) {
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
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
		var nextToken *string
		if nextCursor != nil {
			token, err := s.cursorCodec.Encode(*nextCursor)
			if err != nil {
				writeAPIError(w, r, internalAPIError(err))
				return
			}
			nextToken = &token
		}
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"import_units": rows}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})
	case "unit":
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, errNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), result.IncidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Value)
	case "preview":
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.applicationGetPreview(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, errNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), result.IncidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Value)
	case "mapping_preview":
		s.handleMappingPreview(w, r, principal, route)
	case "mapping":
		s.handleMapping(w, r, principal, route)
	case "select":
		s.handleSelect(w, r, principal, route)
	case "skip":
		s.handleSkip(w, r, principal, route)
	case "regions":
		s.handleRegion(w, r, principal, route)
	case "apply":
		s.handleApply(w, r, principal, route)
	default:
		http.NotFound(w, r)
	}
}

func (s *service) handleMappingPreview(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), scope.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeMappingPreviewRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, apiErr := s.applicationPrepareMappingPreview(
		r.Context(),
		principal.User.ID,
		scope.IncidentID,
		route,
		request,
	)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *service) handleMapping(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetMappingContext(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), scope.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeMappingRequest(r.Body, scope.Value)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	unit, err, apiErr := s.applicationApproveMapping(
		r.Context(),
		principal.User.ID,
		scope.IncidentID,
		route,
		request,
	)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
}

func (s *service) handleSelect(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.select", false)
}

func (s *service) handleSkip(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.skip", true)
}

func (s *service) handleUnitAction(
	w http.ResponseWriter,
	r *http.Request,
	principal httpauth.Principal,
	route importSessionRoute,
	routeKey string,
	allowReason bool,
) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), scope.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeActionRequest(r.Body, allowReason)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.applicationExecuteUnitAction(r.Context(), principal.User.ID, route, request, routeKey)
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *service) handleApply(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetSession(r.Context(), route.SessionID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), scope.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeApplyRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.applicationStartApply(r.Context(), principal.User.ID, route.SessionID, request)
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleRegion(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetSession(r.Context(), route.SessionID)
	if errors.Is(err, errNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), scope.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeRegionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err, apiErr := s.applicationCreateOperatorRegion(r.Context(), principal.User.ID, route, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusCreated, result.Unit)
}

func writeImportStoreError(w http.ResponseWriter, r *http.Request, err error, clientTxnID string) bool {
	if err == nil {
		return true
	}
	var stateConflict *stateConflictError
	var applyBlocked *applyBlockedError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}})
	case errors.Is(err, errNotFound):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
	case errors.As(err, &stateConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_state_conflict", Details: map[string]any{"reason_code": stateConflict.ReasonCode}})
	case errors.As(err, &applyBlocked):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_apply_blocked", Details: map[string]any{"reason_code": applyBlocked.ReasonCode}})
	default:
		writeAPIError(w, r, internalAPIError(err))
	}
	return false
}

func (s *service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	return importAdmissionResult(grant, err, "member")
}

func (s *service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admission.LifecycleAny,
	})
	return importAdmissionResult(grant, err, requiredRole)
}

func importAdmissionResult(grant admission.Grant, err error, requiredRole string) (admission.Grant, *httpapi.APIError) {
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

type importSessionRoute struct {
	Kind      string
	SessionID uuid.UUID
	UnitID    uuid.UUID
}

func parseImportSessionPath(path string) (importSessionRoute, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/import-sessions/")
	if rest == path || rest == "" {
		return importSessionRoute{}, false
	}
	parts := strings.Split(rest, "/")
	sessionID, err := uuid.Parse(parts[0])
	if err != nil {
		return importSessionRoute{}, false
	}
	if len(parts) == 1 {
		return importSessionRoute{Kind: "session", SessionID: sessionID}, true
	}
	if len(parts) == 2 && parts[1] == "units" {
		return importSessionRoute{Kind: "units", SessionID: sessionID}, true
	}
	if len(parts) == 3 && parts[1] == "units" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "unit", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "preview" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "preview", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "mapping-preview" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "mapping_preview", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "mapping" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "mapping", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "select" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "select", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "skip" {
		unitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "skip", SessionID: sessionID, UnitID: unitID}, true
	}
	if len(parts) == 4 && parts[1] == "units" && parts[3] == "regions" {
		baseUnitID, err := uuid.Parse(parts[2])
		if err != nil {
			return importSessionRoute{}, false
		}
		return importSessionRoute{Kind: "regions", SessionID: sessionID, UnitID: baseUnitID}, true
	}
	if len(parts) == 2 && parts[1] == "apply" {
		return importSessionRoute{Kind: "apply", SessionID: sessionID}, true
	}
	return importSessionRoute{}, false
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "The import operation could not be completed.",
		Details: map[string]any{},
	}
}
