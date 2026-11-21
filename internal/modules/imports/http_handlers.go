package imports

import (
	"context"
	"errors"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

// importEditingRoles is the role an import-session route that is not a plain
// read requires. Session, unit and preview reads admit any member and declare
// no authorization failure.
const importEditingRoles = "editor|reviewer|admin"

// importOperationHandlers binds one handler per Imports operation. Each is
// admitted centrally (Core 01 REQ-01-707): the registry resolves the route,
// authenticates by the operation's declared security, including cookie CSRF
// for the six state-changing operations of Table 17.2-A, and admits its path
// parameters before the handler runs.
func (s *service) importOperationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"applyImportSession":                s.handleApply,
		"createImportSession":               s.handleCreateSession,
		"createImportUnitRegion":            s.handleRegion,
		"getCommittedImportSource":          s.handleGetCommittedSource,
		"getCommittedImportSourceContent":   s.handleGetCommittedSourceContent,
		"getImportSession":                  s.handleGetSession,
		"getImportUnit":                     s.handleGetUnit,
		"getImportUnitPreview":              s.handleGetUnitPreview,
		"inspectImportUnitExtensionSource":  s.handleSourceInspection,
		"listCommittedImportSources":        s.handleCommittedSources,
		"listImportUnits":                   s.handleListUnits,
		"previewImportUnitExtensionMapping": s.handleMappingPreview,
		"putImportUnitMapping":              s.handleMapping,
		"selectImportUnit":                  s.handleSelect,
		"skipImportUnit":                    s.handleSkip,
	}
}

func bindRoutes(mux *http.ServeMux, deps httpapi.DependencySet, service *service) error {
	return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.imports", service.importOperationHandlers())
}

// errNotAdmitted reports a handler reached without central admission, which
// is a binding error.
var errNotAdmitted = errors.New("imports route was not admitted centrally")

// admittedPrincipal returns the principal central admission established.
func admittedPrincipal(r *http.Request) (httpauth.Principal, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	if !authenticated {
		return httpauth.Principal{}, internalAPIError(errNotAdmitted)
	}
	return principal, nil
}

// admittedPathUUID returns a uuid path parameter central admission admitted.
func admittedPathUUID(r *http.Request, name string) (uuid.UUID, *httpapi.APIError) {
	value, identified := httpapi.PathUUID(r, name)
	if !identified {
		return uuid.UUID{}, internalAPIError(errNotAdmitted)
	}
	return value, nil
}

// importSessionRoute is the session, and for a unit route the unit, that an
// import-session operation addresses.
type importSessionRoute struct {
	SessionID uuid.UUID
	UnitID    uuid.UUID
}

// admitSessionRoute reads the admitted principal, import_session_id and, when
// unitParameter names one, the unit, then decides the session's visibility and
// the route's role before any query, unit or body classification
// (REQ-01-704). It writes the response and reports false on failure.
func (s *service) admitSessionRoute(w http.ResponseWriter, r *http.Request, unitParameter string, roles admission.RoleSet, requiredRole string) (httpauth.Principal, importSessionRoute, bool) {
	principal, apiErr := admittedPrincipal(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, importSessionRoute{}, false
	}
	var route importSessionRoute
	if route.SessionID, apiErr = admittedPathUUID(r, "import_session_id"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, importSessionRoute{}, false
	}
	if unitParameter != "" {
		if route.UnitID, apiErr = admittedPathUUID(r, unitParameter); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return httpauth.Principal{}, importSessionRoute{}, false
		}
	}
	if apiErr := s.admitImportSession(r.Context(), route.SessionID, principal.User.ID, roles, requiredRole); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, importSessionRoute{}, false
	}
	return principal, route, true
}

// handleCreateSession parses the upload envelope before visibility: its
// metadata incident_id addresses the incident whose visibility is decided
// (Core 01 Table 17.2-A).
func (s *service) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := admittedPrincipal(r)
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
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, importEditingRoles); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *service) handleGetSession(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "", admission.RolesMember, "member")
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.applicationGetSession(r.Context(), route.SessionID)
	if apiErr := importSessionLookupAPIError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Value)
}

func (s *service) handleListUnits(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "", admission.RolesMember, "member")
	if !ok {
		return
	}
	// Core 01 Table 17.2-A: no list-query members, so after the session's
	// visibility the raw query admits only singleton limit and cursor_token.
	paging, apiErr := decodeImportListQuery(r.URL.RawQuery)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
		paging,
		"imports.units.list",
		principal.User.ID.String(),
		map[string]string{"import_session_id": route.SessionID.String()},
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}
	result, err := s.applicationListUnits(r.Context(), route.SessionID)
	if apiErr := importSessionLookupAPIError(err); apiErr != nil {
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
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"import_units": rows}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *service) handleGetUnit(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "import_unit_id", admission.RolesMember, "member")
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if apiErr := importUnitLookupAPIError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Value)
}

func (s *service) handleGetUnitPreview(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "import_unit_id", admission.RolesMember, "member")
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if apiErr := importUnitLookupAPIError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	preview, err, apiErr := s.genericPreview(r.Context(), route.SessionID, route.UnitID, result.Value)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, preview)
}

func (s *service) handleMappingPreview(w http.ResponseWriter, r *http.Request) {
	s.serveExtensionRead(w, r, false)
}

func (s *service) handleSourceInspection(w http.ResponseWriter, r *http.Request) {
	s.serveExtensionRead(w, r, true)
}

// serveExtensionRead answers the two read-only analytical-extension
// operations: source inspection and mapping preview.
func (s *service) serveExtensionRead(w http.ResponseWriter, r *http.Request, inspection bool) {
	principal, route, ok := s.admitSessionRoute(w, r, "import_unit_id", admission.RolesEditorReviewerAdmin, importEditingRoles)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if apiErr := importUnitLookupAPIError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	var resource extensionImportReadResource
	var apiErr *httpapi.APIError
	if inspection {
		var request sourceInspectionRequest
		request, apiErr = decodeSourceInspectionRequest(r.Body)
		if apiErr == nil {
			resource, apiErr = s.applicationInspectSource(r.Context(), principal.User.ID, scope.IncidentID, route, request)
		}
	} else {
		var request mappingPreviewRequest
		request, apiErr = decodeMappingPreviewRequest(r.Body)
		if apiErr == nil {
			resource, apiErr = s.applicationPrepareMappingPreview(r.Context(), principal.User.ID, scope.IncidentID, route, request)
		}
	}
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *service) handleMapping(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "import_unit_id", admission.RolesEditorReviewerAdmin, importEditingRoles)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	scope, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if apiErr := importUnitLookupAPIError(err); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeMappingRequest(r.Body, nil)
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
}

func (s *service) handleSelect(w http.ResponseWriter, r *http.Request) {
	s.serveUnitAction(w, r, "imports.units.select", false)
}

func (s *service) handleSkip(w http.ResponseWriter, r *http.Request) {
	s.serveUnitAction(w, r, "imports.units.skip", true)
}

func (s *service) serveUnitAction(w http.ResponseWriter, r *http.Request, routeKey string, allowReason bool) {
	principal, route, ok := s.admitSessionRoute(w, r, "import_unit_id", admission.RolesEditorReviewerAdmin, importEditingRoles)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, err := s.applicationGetUnit(r.Context(), route.SessionID, route.UnitID)
	if apiErr := importUnitLookupAPIError(err); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *service) handleApply(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "", admission.RolesEditorReviewerAdmin, importEditingRoles)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

// handleRegion addresses its base unit through base_unit_id.
func (s *service) handleRegion(w http.ResponseWriter, r *http.Request) {
	principal, route, ok := s.admitSessionRoute(w, r, "base_unit_id", admission.RolesEditorReviewerAdmin, importEditingRoles)
	if !ok {
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
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
	case errors.Is(err, errSessionNotFound):
		writeAPIError(w, r, httpapi.NotFound("import_session_not_found"))
	case errors.Is(err, errNotFound):
		writeAPIError(w, r, httpapi.NotFound("import_unit_not_found"))
	case errors.As(err, &stateConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_state_conflict", Details: map[string]any{"reason_code": stateConflict.ReasonCode}})
	case errors.As(err, &applyBlocked):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "import_apply_blocked", Details: map[string]any{"reason_code": applyBlocked.ReasonCode}})
	default:
		writeAPIError(w, r, internalAPIError(err))
	}
	return false
}

// admitImportSession decides the caller's visibility of the session's
// incident immediately after authentication, before any query, session, unit
// or body classification (REQ-01-704). A missing session and a session whose
// incident the caller cannot see give the same import_session_not_found. A
// caller who can see the incident but lacks the route's role receives the
// route's authorization failure.
func (s *service) admitImportSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) *httpapi.APIError {
	incidentID, err := s.applicationSessionIncident(ctx, sessionID)
	if errors.Is(err, errSessionNotFound) {
		return httpapi.NotFound("import_session_not_found")
	}
	if err != nil {
		return internalAPIError(err)
	}
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admission.LifecycleAny,
	})
	if admission.IsDenied(err, admission.DenialNotVisible) {
		return httpapi.NotFound("import_session_not_found")
	}
	_, apiErr := importAdmissionResult(grant, err, requiredRole)
	return apiErr
}

// regionAdmissionAPIError maps the transactional admission recheck of region
// creation to the route's declared codes. A caller who lost visibility of the
// incident after the route admitted the request receives the session's
// not-found response, and one who lost the editing role receives
// authorization_denied, exactly as admitImportSession would have answered
// (REQ-01-704). Any other error is left to the store error mapping.
func regionAdmissionAPIError(err error) *httpapi.APIError {
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return httpapi.NotFound("import_session_not_found")
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": importEditingRoles}}
	default:
		return nil
	}
}

// importSessionLookupAPIError classifies a session read after the visibility
// gate. A session read addresses no unit, so it never reports a missing unit.
func importSessionLookupAPIError(err error) *httpapi.APIError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errSessionNotFound):
		return httpapi.NotFound("import_session_not_found")
	default:
		return internalAPIError(err)
	}
}

// importUnitLookupAPIError classifies a unit read after the visibility gate.
// The session is read first, and a missing session is never reported as a
// missing unit.
func importUnitLookupAPIError(err error) *httpapi.APIError {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, errSessionNotFound):
		return httpapi.NotFound("import_session_not_found")
	case errors.Is(err, errNotFound):
		return httpapi.NotFound("import_unit_not_found")
	default:
		return internalAPIError(err)
	}
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
		return admission.Grant{}, httpapi.NotFound("incident_not_found")
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": requiredRole}}
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

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: "The import operation could not be completed.",
		Details: map[string]any{},
	}
}
