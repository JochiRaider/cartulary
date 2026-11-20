package reportcomposition

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type Service struct {
	store          *Store
	incidentAccess *admission.Checker
	authStore      *authn.Store
	cursorCodec    *pagination.Codec
	now            func() time.Time
	previewJobs    PreviewJobPort
}

type RouteOptions struct {
	PreviewJobs PreviewJobPort
}

// RegisterRoutes binds the composition routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits incident_id, composition_id and composition_version
// before a handler runs. A malformed value receives the parameter's declared
// not-found response, which carries its validation_code.
func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.reportcomposition", map[string]http.HandlerFunc{
			"createReportComposition":        service.handleCreate,
			"freezeReportCompositionVersion": service.handleFreeze,
			"getReportComposition":           service.handleGet,
			"getReportCompositionVersion":    service.handleVersion,
			"listReportCompositions":         service.handleList,
			"previewReportComposition":       service.handlePreview,
			"retireReportComposition":        service.handleRetire,
			"updateReportCompositionDraft":   service.handleUpdate,
			"validateReportComposition":      service.handleValidate,
		})
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
		cursorKey, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		cursorCodec = pagination.NewCodec(cursorKey)
	}
	return &Service{
		store:          NewStore(deps.PostgresHandle()),
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		cursorCodec:    cursorCodec,
		now:            now,
		previewJobs:    options.PreviewJobs,
	}, nil
}

// admittedIncident returns the principal and incident_id central admission
// established (REQ-01-707 steps 2 and 3); their absence is a binding error.
func admittedIncident(r *http.Request) (httpauth.Principal, uuid.UUID, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	incidentID, identified := httpapi.PathUUID(r, "incident_id")
	if !authenticated || !identified {
		return httpauth.Principal{}, uuid.UUID{}, httpapi.InternalAPIError(errors.New("report composition route was not admitted centrally"))
	}
	return principal, incidentID, nil
}

// admittedComposition additionally returns the admitted composition_id.
func admittedComposition(r *http.Request) (httpauth.Principal, uuid.UUID, uuid.UUID, *httpapi.APIError) {
	principal, incidentID, apiErr := admittedIncident(r)
	if apiErr != nil {
		return httpauth.Principal{}, uuid.UUID{}, uuid.UUID{}, apiErr
	}
	compositionID, ok := httpapi.PathUUID(r, "composition_id")
	if !ok {
		return httpauth.Principal{}, uuid.UUID{}, uuid.UUID{}, httpapi.InternalAPIError(errors.New("report composition route did not admit composition_id"))
	}
	return principal, incidentID, compositionID, nil
}

func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncident(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-704 and REQ-01-707 step 5: the query is route handling, decided
	// after the incident's visibility.
	if apiErr := validateCompositionListQuery(r); apiErr != nil {
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
}

func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, apiErr := admittedIncident(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
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
}

func (s *Service) handleGet(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-704 and REQ-01-707 step 5: the singleton-query check is route
	// handling, decided after the incident's visibility (REQ-RC-031a).
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
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
}

func (s *Service) handleUpdate(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
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
}

func (s *Service) handleRetire(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
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
}

func (s *Service) handleFreeze(w http.ResponseWriter, r *http.Request) {
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r, incidentID, principal.User.ID, admission.RolesEditorAdmin, "editor|admin"); apiErr != nil {
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
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	versionText, ok := httpapi.PathString(r, "composition_version")
	if !ok {
		writeAPIError(w, r, httpapi.InternalAPIError(errors.New("report composition route did not admit composition_version")))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r, incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-704 and REQ-01-707 step 5: the singleton-query check is route
	// handling, decided after the incident's visibility (REQ-RC-031a).
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// Central admission checked the declared syntax; a version beyond the
	// representable range names no version, so it is the missing version.
	version, ok := parseCompositionVersion(versionText)
	if !ok {
		writeAPIError(w, r, validationCodeError(http.StatusNotFound, "not_found", "composition_version_not_found"))
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
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
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
		summary = validationSummaryForVersion(record, version)
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
	principal, incidentID, compositionID, apiErr := admittedComposition(r)
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

func (s *Service) requireIncidentMembership(r *http.Request, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(r.Context(), incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	return reportCompositionAdmissionResult(grant, err, "member")
}

func (s *Service) requireIncidentRole(r *http.Request, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(r.Context(), incidentID, userID, admission.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admission.LifecycleAny,
	})
	return reportCompositionAdmissionResult(grant, err, requiredRole)
}

func reportCompositionAdmissionResult(grant admission.Grant, err error, requiredRole string) (admission.Grant, *httpapi.APIError) {
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, httpapi.NotFound("incident_not_found")
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admission.Grant{}, httpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}

func (s *Service) slideSessionIfNeeded(r *http.Request, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(r.Context(), s.authStore, principal, s.now)
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
	if errors.Is(err, ErrPreviewInputRejected) {
		writeAPIError(w, r, validationCodeError(http.StatusBadRequest, "invalid_request", "composition_preview_input_invalid"))
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

func validateInlineComposition(record ResourceRecord, raw json.RawMessage) map[string]any {
	source, err := decodeBoundRenderSource(raw, func(source RenderSource) string {
		return compositionResourceBinding(record, source)
	})
	if err != nil {
		var sourceErr *RenderSourceError
		if errors.As(err, &sourceErr) {
			return invalidInlineSummary(record, sourceErr.ReasonCode)
		}
		return invalidInlineSummary(record, "composition_schema_invalid")
	}
	summary := map[string]any{
		"validation_scope":    "local_authoring",
		"valid":               true,
		"stage":               nil,
		"issues":              []map[string]any{},
		"composition_id":      record.CompositionID.String(),
		"composition_version": *source.CompositionVersion,
		"composition_sha256":  *source.CompositionSHA256,
	}
	return summary
}

func invalidInlineSummary(record ResourceRecord, code string) map[string]any {
	stage := compositionStages[compositionFindingStage(code)]
	return map[string]any{
		"validation_scope":    "local_authoring",
		"valid":               false,
		"stage":               stage,
		"issues":              []map[string]any{issue(code, map[string]any{"composition_id": record.CompositionID.String()}).public()},
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
