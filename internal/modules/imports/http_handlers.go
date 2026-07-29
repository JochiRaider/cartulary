package imports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

func (s *Service) handleImportSessionsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, envelopeErr := httpapi.ParseUploadEnvelope(r, httpapi.UploadEnvelopePolicy{FileContentTypes: ImportSessionFileContentTypes})
	if envelopeErr != nil {
		writeAPIError(w, r, uploadEnvelopeAPIError(envelopeErr))
		return
	}
	request, apiErr := DecodeCreateSessionMetadata(envelope)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), request.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	sourceFileKind := detectSourceFileKind(envelope)
	units, apiErr := s.discoverImportUnits(envelope, sourceFileKind)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	normalized, err := json.Marshal(map[string]any{
		"incident_id":           request.IncidentID.String(),
		"client_txn_id":         request.ClientTxnID,
		"assistant_profile":     request.AssistantProfile,
		"source_content_sha256": envelope.FileSHA256Hex,
	})
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := s.store.CreateAcceptedSession(r.Context(), CreateAcceptedSessionParams{
		ActorUserID:         principal.User.ID,
		Request:             request,
		SourceFileKind:      sourceFileKind,
		OriginalFilename:    envelope.FileName,
		SourceContentSHA256: envelope.FileSHA256Hex,
		SourceMediaType:     envelope.FileContentType,
		SourceByteSize:      int64(len(envelope.File)),
		SourceBytes:         envelope.File,
		Units:               units,
		NormalizedRequest:   normalized,
		Now:                 s.now(),
	})
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, clientTxnConflict(request.ClientTxnID))
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		if err := s.executeDiscoveryJob(r.Context(), uuid.MustParse(result.Job.JobID)); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func (s *Service) handleImportSessionsMember(w http.ResponseWriter, r *http.Request) {
	route, ok := parseImportSessionPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{
		Store:         s.authStore,
		Keys:          s.keys,
		Now:           s.now,
		StateChanging: route.Kind == "regions",
	})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch route.Kind {
	case "session":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		resource, incidentID, err := s.store.GetSession(r.Context(), route.SessionID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
	case "units":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
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
		units, incidentID, err := s.store.ListUnits(r.Context(), route.SessionID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		rows, nextCursor, err := pagination.PageResources(binding, cursor, units)
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
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		unit, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
	case "preview":
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		preview, incidentID, err := s.store.GetPreview(r.Context(), route.SessionID, route.UnitID)
		if errors.Is(err, ErrNotFound) {
			writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
			return
		}
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, preview)
	case "mapping_preview":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleMappingPreview(w, r, principal, route)
	case "mapping":
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleMapping(w, r, principal, route)
	case "select":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleSelect(w, r, principal, route)
	case "skip":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleSkip(w, r, principal, route)
	case "regions":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleRegion(w, r, principal, route)
	case "apply":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		s.handleApply(w, r, principal, route)
	default:
		http.NotFound(w, r)
	}
}

func (s *Service) handleMappingPreview(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeMappingPreviewRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	mapping := ApprovedMapping{
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	}
	if apiErr := s.validateApprovedMapping(mapping); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, ok := lookupApprovedImportTarget(mapping)
	if !ok {
		writeAPIError(w, r, invalidImportRequest("target_kind", "target_kind_not_importable"))
		return
	}
	facade := s.extensionImportFacades[extensionImportFacadeKey(target)]
	if facade == nil {
		writeAPIError(
			w,
			r,
			invalidImportRequest("target_kind", "owner_preview_contract_unavailable"),
		)
		return
	}
	sourceCapability, err := s.store.SourceCapabilityForUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result, err := facade.PrepareImportUnitMapping(r.Context(), ExtensionImportMappingRequest{
		IncidentID:           incidentID,
		ActorUserID:          principal.User.ID,
		TargetKind:           request.TargetKind,
		ExtensionProfileID:   request.ExtensionProfileID,
		ImportSessionID:      route.SessionID,
		ImportUnitID:         route.UnitID,
		SourceCapability:     sourceCapability,
		OwnerMappingSchemaID: request.OwnerMappingSchemaID,
		OwnerMapping:         append(json.RawMessage(nil), request.OwnerMapping...),
	})
	if err != nil {
		writeAPIError(w, r, extensionFacadeAPIError(target, facade, err))
		return
	}
	if err := facade.ValidateImportUnitMappingResult(result); err != nil {
		writeAPIError(
			w,
			r,
			invalidImportRequest("owner_result", "owner_preview_validation_failed"),
		)
		return
	}
	resource := ExtensionMappingPreviewResource{
		SchemaID:            ExtensionMappingPreviewResultSchemaID,
		ImportSessionID:     route.SessionID.String(),
		ImportUnitID:        route.UnitID.String(),
		TargetKind:          request.TargetKind,
		ExtensionProfileID:  request.ExtensionProfileID,
		OwnerResultSchemaID: result.OwnerResultSchemaID,
		OwnerResult:         result.OwnerResult,
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleMapping(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	columns, incidentID, err := s.store.GetUnitColumns(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeMappingRequest(r.Body, columns)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	materialized, apiErr := s.prepareApprovedMapping(r.Context(), principal.User.ID, incidentID, route, request)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request = materialized
	unit, _, err := s.store.SaveMapping(r.Context(), MappingParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, unit)
}

func (s *Service) handleSelect(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.select", false, s.store.SelectUnit)
}

func (s *Service) handleSkip(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	s.handleUnitAction(w, r, principal, route, "imports.units.skip", true, s.store.SkipUnit)
}

func (s *Service) handleUnitAction(
	w http.ResponseWriter,
	r *http.Request,
	principal httpauth.Principal,
	route importSessionRoute,
	routeKey string,
	allowReason bool,
	action func(context.Context, UnitActionParams) (UnitActionResult, error),
) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, incidentID, err := s.store.GetUnit(r.Context(), route.SessionID, route.UnitID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_unit_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeActionRequest(r.Body, allowReason)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := action(r.Context(), UnitActionParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		UnitID:            route.UnitID,
		RouteKey:          routeKey,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleApply(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, route importSessionRoute) {
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	session, incidentID, err := s.store.GetSession(r.Context(), route.SessionID)
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusNotFound, Code: "import_session_not_found", Details: map[string]any{}})
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeApplyRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_ = session
	result, err := s.store.StartApply(r.Context(), ApplyStartParams{
		ActorUserID:       principal.User.ID,
		SessionID:         route.SessionID,
		Request:           request,
		NormalizedRequest: request.Normalized,
		Now:               s.now(),
	})
	if !writeImportStoreError(w, r, err, request.ClientTxnID) {
		return
	}
	if !result.Replayed {
		if err := s.executeApplyJob(r.Context(), uuid.MustParse(result.Job.JobID)); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusAccepted, result.Job)
}

func writeImportStoreError(w http.ResponseWriter, r *http.Request, err error, clientTxnID string) bool {
	if err == nil {
		return true
	}
	var stateConflict *StateConflictError
	var applyBlocked *ApplyBlockedError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflict(clientTxnID))
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}})
	case errors.Is(err, ErrNotFound):
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

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
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
