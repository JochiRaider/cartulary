package reportcomposition

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type Service struct {
	store          *Store
	incidentAccess incidents.Access
	authStore      *authn.Store
	keys           authn.MasterKeys
	cursorCodec    *pagination.Codec
	now            func() time.Time
	previewJobs    PreviewJobPort
}

type RouteOptions struct {
	PreviewJobs PreviewJobPort
}

func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		if err := httpapi.DeclarePublicOperations(deps, PublicOperations()...); err != nil {
			return err
		}
		httpapi.HandlePublicRoute(mux, "GET /api/v1/incidents/{incident_id}/report-compositions", service.handleCollection)
		httpapi.HandlePublicRoute(mux, "POST /api/v1/incidents/{incident_id}/report-compositions", service.handleCollection)
		httpapi.HandlePublicRoute(mux, "GET /api/v1/incidents/{incident_id}/report-compositions/{composition_id}", service.handleResource)
		httpapi.HandlePublicRoute(mux, "PATCH /api/v1/incidents/{incident_id}/report-compositions/{composition_id}", service.handleResource)
		httpapi.HandlePublicRoute(mux, "DELETE /api/v1/incidents/{incident_id}/report-compositions/{composition_id}", service.handleResource)
		httpapi.HandlePublicRoute(mux, "POST /api/v1/incidents/{incident_id}/report-compositions/{composition_id}/versions", service.handleVersions)
		httpapi.HandlePublicRoute(mux, "GET /api/v1/incidents/{incident_id}/report-compositions/{composition_id}/versions/{composition_version}", service.handleVersion)
		httpapi.HandlePublicRoute(mux, "POST /api/v1/incidents/{incident_id}/report-compositions/{composition_id}/validate", service.handleValidate)
		httpapi.HandlePublicRoute(mux, "POST /api/v1/incidents/{incident_id}/report-compositions/{composition_id}/preview", service.handlePreview)
		return nil
	}
}

func newService(deps httpapi.DependencySet, options RouteOptions) (*Service, error) {
	if options.PreviewJobs == nil {
		return nil, errors.New("report composition routes require Reporting preview job port")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
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
		store:          NewStore(deps.PostgresHandle()),
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		cursorCodec:    cursorCodec,
		now:            now,
		previewJobs:    options.PreviewJobs,
	}, nil
}

func (s *Service) handleCollection(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := parsePathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	stateChanging := r.Method == http.MethodPost
	principal, apiErr := s.authenticate(r, stateChanging)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := validateCompositionListQuery(r); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		resources, err := s.store.ListResources(r.Context(), incidentID)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		binding, cursor, reason := s.cursorCodec.ResolveListRequest(r.URL.Query(), "report_compositions.list", principal.User.ID.String(), map[string]string{"incident_id": incidentID.String()})
		if reason != "" {
			writeAPIError(w, r, invalidPaginationRequest(reason))
			return
		}
		rows, nextCursor, err := pagination.PageResources(binding, cursor, resources)
		if errors.Is(err, pagination.ErrInvalidCursorToken) {
			writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
			return
		}
		if err != nil {
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
		if err := s.slideSessionIfNeeded(r, &principal); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"composition_resources": rows}, httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		})
	case http.MethodPost:
		if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeCreateDraftRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.store.CreateDraft(r.Context(), incidentID, principal.User.ID, request, s.now())
		s.writeMutationResult(w, r, &principal, request.ClientTxnID, result, err, "composition_not_found")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleResource(w http.ResponseWriter, r *http.Request) {
	incidentID, compositionID, ok := parseIncidentCompositionPath(w, r)
	if !ok {
		return
	}
	stateChanging := r.Method == http.MethodPatch || r.Method == http.MethodDelete
	principal, apiErr := s.authenticate(r, stateChanging)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		_, payload, err := s.store.GetResource(r.Context(), incidentID, compositionID)
		if err != nil {
			s.writeReadError(w, r, err, "composition_not_found")
			return
		}
		if err := s.slideSessionIfNeeded(r, &principal); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
	case http.MethodPatch:
		if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeUpdateDraftRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.store.UpdateDraft(r.Context(), incidentID, compositionID, principal.User.ID, request, s.now())
		s.writeMutationResult(w, r, &principal, request.ClientTxnID, result, err, "composition_not_found")
	case http.MethodDelete:
		if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := DecodeDraftVersionRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		result, err := s.store.RetireResource(r.Context(), incidentID, compositionID, principal.User.ID, request, s.now())
		s.writeMutationResult(w, r, &principal, request.ClientTxnID, result, err, "composition_not_found")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleVersions(w http.ResponseWriter, r *http.Request) {
	incidentID, compositionID, ok := parseIncidentCompositionPath(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, "editor", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeDraftVersionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.FreezeVersion(r.Context(), incidentID, compositionID, principal.User.ID, request, s.now())
	s.writeMutationResult(w, r, &principal, request.ClientTxnID, result, err, "composition_not_found")
}

func (s *Service) handleVersion(w http.ResponseWriter, r *http.Request) {
	incidentID, compositionID, ok := parseIncidentCompositionPath(w, r)
	if !ok {
		return
	}
	version, ok := parsePathCompositionVersion(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	_, payload, err := s.store.GetVersion(r.Context(), incidentID, compositionID, version)
	if err != nil {
		s.writeReadError(w, r, err, "composition_version_not_found")
		return
	}
	if err := s.slideSessionIfNeeded(r, &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, payload)
}

func (s *Service) handleValidate(w http.ResponseWriter, r *http.Request) {
	incidentID, compositionID, ok := parseIncidentCompositionPath(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeValidateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, _, err := s.store.GetResource(r.Context(), incidentID, compositionID)
	if err != nil {
		s.writeReadError(w, r, err, "composition_not_found")
		return
	}
	var summary map[string]any
	switch request.SourceKind {
	case SourceKindDraft:
		summary = validationSummaryForResource(record)
	case SourceKindVersion:
		version, _, err := s.store.GetVersion(r.Context(), incidentID, compositionID, *request.CompositionVersion)
		if err != nil {
			s.writeReadError(w, r, err, "composition_version_not_found")
			return
		}
		summary = validationSummaryForVersion(version)
	case SourceKindInline:
		summary = validateInlineComposition(record, *request.InlineComposition)
	default:
		writeAPIError(w, r, validationCodeError(http.StatusBadRequest, "invalid_request", "composition_source_invalid"))
		return
	}
	if err := s.slideSessionIfNeeded(r, &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, summary)
}

func (s *Service) handlePreview(w http.ResponseWriter, r *http.Request) {
	incidentID, compositionID, ok := parseIncidentCompositionPath(w, r)
	if !ok {
		return
	}
	principal, apiErr := s.authenticate(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodePreviewRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreatePreviewAttempt(r.Context(), incidentID, compositionID, principal.User.ID, request, s.previewJobs, s.now())
	if err == nil && !result.Replayed {
		if jobID, ok := result.Payload["render_attempt_id"].(string); ok {
			_ = s.previewJobs.DispatchPreviewJob(jobID)
		}
	}
	s.writePreviewResult(w, r, &principal, request.ClientTxnID, result, err)
}

func (s *Service) authenticate(r *http.Request, stateChanging bool) (httpauth.Principal, *httpapi.APIError) {
	return httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: stateChanging})
}

func (s *Service) requireIncidentMembership(r *http.Request, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(r.Context(), s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(r *http.Request, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(r.Context(), s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) slideSessionIfNeeded(r *http.Request, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(r.Context(), s.authStore, principal, r.Method, r.URL.Path, s.now)
}

func (s *Service) writeMutationResult(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, clientTxnID string, result MutationResult, err error, notFoundCode string) {
	if s.writeStoreError(w, r, err, clientTxnID, notFoundCode) {
		return
	}
	if err := s.slideSessionIfNeeded(r, principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) writePreviewResult(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, clientTxnID string, result PreviewResult, err error) {
	if s.writeStoreError(w, r, err, clientTxnID, "composition_not_found") {
		return
	}
	if err := s.slideSessionIfNeeded(r, principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) writeReadError(w http.ResponseWriter, r *http.Request, err error, validationCode string) {
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, validationCodeError(http.StatusNotFound, "not_found", validationCode))
		return
	}
	writeAPIError(w, r, internalAPIError(err))
}

func (s *Service) writeStoreError(w http.ResponseWriter, r *http.Request, err error, clientTxnID string, notFoundCode string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, authn.ErrClientTxnConflict) {
		writeAPIError(w, r, httpapi.ClientTxnConflictError(clientTxnID))
		return true
	}
	if errors.Is(err, ErrNotFound) {
		writeAPIError(w, r, validationCodeError(http.StatusNotFound, "not_found", notFoundCode))
		return true
	}
	if errors.Is(err, ErrDraftVersionConflict) {
		writeAPIError(w, r, validationCodeError(http.StatusConflict, "conflict", "composition_draft_version_conflict"))
		return true
	}
	if errors.Is(err, ErrResourceRetired) {
		writeAPIError(w, r, validationCodeError(http.StatusConflict, "conflict", "composition_resource_retired"))
		return true
	}
	if errors.Is(err, ErrVersionBound) {
		writeAPIError(w, r, validationCodeError(http.StatusConflict, "conflict", "composition_version_bound"))
		return true
	}
	var validationErr *validationStoreError
	if errors.As(err, &validationErr) {
		if validationErr.Summary != nil {
			writeAPIError(w, r, validationSummaryError(validationErr.Status, validationErr.Code, validationErr.Summary))
			return true
		}
		writeAPIError(w, r, validationCodeError(validationErr.Status, validationErr.Code, validationErr.ValidationCode))
		return true
	}
	writeAPIError(w, r, internalAPIError(err))
	return true
}

func parsePathUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, bool) {
	value, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		writeAPIError(w, r, schemaFieldError(name, "invalid_value"))
		return uuid.UUID{}, false
	}
	return value, true
}

func parseIncidentCompositionPath(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	incidentID, ok := parsePathUUID(w, r, "incident_id")
	if !ok {
		return uuid.UUID{}, uuid.UUID{}, false
	}
	compositionID, ok := parsePathUUID(w, r, "composition_id")
	if !ok {
		return uuid.UUID{}, uuid.UUID{}, false
	}
	return incidentID, compositionID, true
}

func parsePathCompositionVersion(w http.ResponseWriter, r *http.Request) (int64, bool) {
	version, ok := parseCompositionVersion(r.PathValue("composition_version"))
	if !ok {
		writeAPIError(w, r, schemaFieldError("composition_version", "invalid_value"))
		return 0, false
	}
	return version, true
}

func validateInlineComposition(record ResourceRecord, raw json.RawMessage) map[string]any {
	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return invalidInlineSummary(record, "composition_schema_invalid")
	}
	if decoded["incident_id"] != record.IncidentID.String() {
		return invalidInlineSummary(record, "composition_incident_mismatch")
	}
	if decoded["composition_id"] != record.CompositionID.String() {
		return invalidInlineSummary(record, "composition_id_mismatch")
	}
	if decoded["template_id"] != record.TemplateID || decoded["template_version"] != record.TemplateVersion {
		return invalidInlineSummary(record, "composition_template_mismatch")
	}
	sha, _ := decoded["composition_sha256"].(string)
	digest, err := digestFromCompositionBytes(raw)
	if err != nil || digest != sha {
		return invalidInlineSummary(record, "composition_digest_mismatch")
	}
	summary := map[string]any{
		"valid":               true,
		"stage":               nil,
		"issues":              []map[string]any{},
		"composition_id":      record.CompositionID.String(),
		"composition_version": decoded["composition_version"],
		"composition_sha256":  sha,
	}
	return summary
}

func invalidInlineSummary(record ResourceRecord, code string) map[string]any {
	stage := "schema_validation"
	if code == "composition_digest_mismatch" {
		stage = "canonical_digest_validation"
	}
	if code == "composition_incident_mismatch" || code == "composition_id_mismatch" || code == "composition_template_mismatch" {
		stage = "resource_binding_validation"
	}
	return map[string]any{
		"valid":               false,
		"stage":               stage,
		"issues":              []map[string]any{issue(code, map[string]any{"composition_id": record.CompositionID.String()})},
		"composition_id":      record.CompositionID.String(),
		"composition_version": nil,
		"composition_sha256":  nil,
	}
}

func validateCompositionListQuery(r *http.Request) *httpapi.APIError {
	for key := range r.URL.Query() {
		switch key {
		case "limit", "cursor_token", "page", "offset", "page_size", "block_size":
		default:
			return schemaFieldError(key, "unknown_query_member")
		}
	}
	return nil
}

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{"reason_code": reasonCode},
	}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	httpapi.WriteAPIError(w, r, apiErr)
}
