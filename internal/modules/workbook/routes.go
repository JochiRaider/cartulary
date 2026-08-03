package workbook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineadmission "github.com/JochiRaider/cartulary/internal/modules/timeline/admission"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	contributions  *WorkbookContributionCatalog
	mutations      workbookMutationPort
	recordTargets  workbookRecordTargetPort
	timelineOwner  workbookTimelineMutationPort
	entityOwner    workbookEntityMutationPort
	conflictTokens workbookConflictTokenPort
	incidentAccess incidents.Access
	startupStore   *workbookstartup.Store
	authStore      *authn.Store
	cursorCodec    *pagination.Codec
	keys           authn.MasterKeys
	now            func() time.Time
	serviceVersion string
}

type StartupStoreFactory func(httpapi.DependencySet) (*workbookstartup.Store, error)

type RouteDependencies struct {
	TimelineOwner       *timeline.Facade
	MutationStore       *Store
	EntityOwner         *hostidentity.Store
	ConflictTokens      conflicttokens.ConflictTokenCodec
	StartupStoreFactory StartupStoreFactory
}

func RegisterRoutes(routeDependencies RouteDependencies) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if routeDependencies.StartupStoreFactory == nil {
			return errors.New("workbook route composition requires a startup store factory")
		}
		startupStore, err := routeDependencies.StartupStoreFactory(deps)
		if err != nil {
			return fmt.Errorf("compose workbook startup store: %w", err)
		}
		if startupStore == nil {
			return errors.New("workbook startup store factory returned nil")
		}
		service, err := newService(
			deps,
			routeDependencies,
			startupStore,
		)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.workbook", map[string]http.HandlerFunc{
			"applyWorkbookBulkMutation":             service.handleBulkMutations,
			"createRecordLinkedNote":                service.handleLinkedNoteCreate,
			"createViewRow":                         service.handleCreate,
			"getCurrentUserWorkbookPreferences":     service.handleWorkbookPreferencesMe,
			"getIncidentDefaultWorkbookPreferences": service.handleWorkbookPreferencesDefault,
			"getIncidentWorkbookStartup":            service.handleWorkbookStartup,
			"patchRecord":                           service.handlePatch,
			"pasteWorkbookClipboard":                service.handleClipboardPaste,
			"putCurrentUserWorkbookPreferences":     service.handleWorkbookPreferencesMe,
			"putIncidentDefaultWorkbookPreferences": service.handleWorkbookPreferencesDefault,
			"queryWorkbookView":                     service.handleQuery,
			"resolveRecordSameFieldConflict":        service.handleConflictResolve,
			"supersedeRecord":                       service.handleSupersede,
		})
	}
}

func (s *Service) handleBulkMutations(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeTimelineBulkMutationRequest(r.Body, viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	targets := timelineOwnerBatchTargets(request.Targets)
	requestHash := TimelineBulkMutationRequestHash(request)
	var result timeline.ClipboardPasteResult
	var err error
	switch request.Kind {
	case timeline.OwnerBatchOperationFillDownV1:
		result, err = s.timelineOwner.ApplyFillDown(r.Context(), timeline.FillDownCommand{
			Actor:       principal.User,
			IncidentID:  incidentID,
			ClientTxnID: request.ClientTxnID,
			FieldKey:    request.FieldKey,
			Value:       request.Value,
			Targets:     targets,
			RequestHash: requestHash,
			RequestID:   httpapi.RequestIDFromContext(r.Context()),
			Now:         s.now(),
		})
	case timeline.OwnerBatchOperationMultiRowTagAssignmentV1:
		result, err = s.timelineOwner.ApplyMultiRowTagAssignment(r.Context(), timeline.MultiRowTagAssignmentCommand{
			Actor:         principal.User,
			IncidentID:    incidentID,
			ClientTxnID:   request.ClientTxnID,
			TagName:       request.TagName,
			NormalizedTag: request.NormalizedTag,
			Targets:       targets,
			RequestHash:   requestHash,
			RequestID:     httpapi.RequestIDFromContext(r.Context()),
			Now:           s.now(),
		})
	default:
		writeAPIError(w, r, invalidMutationPayload("kind", "invalid_value"))
		return
	}
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timelineadmission.MutationAPIErrorContext{ClientTxnID: request.ClientTxnID}):
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case isTimelineMentionMutationError(err):
		writeAPIError(w, r, invalidMutationPayload("value", "invalid_value"))
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
}

func (s *Service) handleClipboardPaste(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("", "invalid_value"))
		return
	}
	if viewSchemaID != timeline.TimelineViewSchemaID {
		s.handleEntityClipboardPaste(w, r, principal, incidentID, viewSchemaID, body)
		return
	}
	request, apiErr := DecodeTimelineClipboardPasteRequest(bytes.NewReader(body))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	plan, err := BuildTimelineClipboardPlan(request)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("clipboard_text", "invalid_value"))
		return
	}
	result, err := s.timelineOwner.ApplyClipboardPaste(r.Context(), timeline.ClipboardPasteCommand{
		Actor:       principal.User,
		IncidentID:  incidentID,
		ClientTxnID: request.ClientTxnID,
		Plan:        plan,
		Targets:     timelineOwnerBatchTargets(request.Targets),
		RequestHash: TimelineClipboardPasteRequestHash(request),
		RequestID:   httpapi.RequestIDFromContext(r.Context()),
		Now:         s.now(),
	})
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timelineadmission.MutationAPIErrorContext{ClientTxnID: request.ClientTxnID}):
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case isTimelineMentionMutationError(err):
		writeAPIError(w, r, invalidMutationPayload("clipboard_text", "invalid_value"))
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
}

func (s *Service) handleEntityClipboardPaste(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, incidentID uuid.UUID, viewSchemaID string, body []byte) {
	if viewSchemaID != hostidentity.HostsViewSchemaID && viewSchemaID != hostidentity.IdentitiesViewSchemaID {
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}
	request, apiErr := hostidentity.DecodeClipboardPasteRequest(bytes.NewReader(body), viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	plan, err := hostidentity.BuildClipboardPastePlan(request)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("clipboard_text", "invalid_value"))
		return
	}
	result, err := s.entityOwner.ApplyClipboardPastePlan(
		r.Context(),
		principal.User,
		incidentID,
		request.ViewSchemaID,
		plan,
		request.RequestHash(),
		httpapi.RequestIDFromContext(r.Context()),
		s.now(),
	)
	var entityConflict *hostidentity.ExactMatchConflictError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, hostidentity.ErrInvalidCreateRequest):
		writeAPIError(w, r, invalidMutationPayload("payload", "at_least_one_value_required"))
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
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
}

func timelineOwnerBatchTargets(targets []TimelineBatchTarget) []timeline.OwnerBatchTargetV1 {
	converted := make([]timeline.OwnerBatchTargetV1, 0, len(targets))
	for _, target := range targets {
		converted = append(converted, timeline.OwnerBatchTargetV1{
			Kind:           target.Kind,
			RecordID:       target.RecordID,
			BaseRowVersion: target.BaseRowVersion,
		})
	}
	return converted
}

func newService(
	deps httpapi.DependencySet,
	routeDependencies RouteDependencies,
	startupStore *workbookstartup.Store,
) (*Service, error) {
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
	if routeDependencies.TimelineOwner == nil {
		return nil, errors.New("workbook route composition requires a Timeline owner")
	}
	if routeDependencies.MutationStore == nil {
		return nil, errors.New("workbook route composition requires a mutation store")
	}
	if routeDependencies.MutationStore.contributions == nil {
		return nil, errors.New("workbook route composition requires a contribution catalog")
	}
	if routeDependencies.MutationStore.recordTargets == nil {
		return nil, errors.New("workbook route composition requires a record target owner")
	}
	if routeDependencies.EntityOwner == nil {
		return nil, errors.New("workbook route composition requires an Entity owner")
	}
	if startupStore == nil {
		return nil, errors.New("workbook route composition requires a startup store")
	}
	return &Service{
		contributions:  routeDependencies.MutationStore.contributions,
		mutations:      routeDependencies.MutationStore,
		recordTargets:  routeDependencies.MutationStore.recordTargets,
		timelineOwner:  routeDependencies.TimelineOwner,
		entityOwner:    routeDependencies.EntityOwner,
		conflictTokens: routeDependencies.ConflictTokens,
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		startupStore:   startupStore,
		authStore:      authn.NewStore(deps.PostgresHandle()),
		cursorCodec:    cursorCodec,
		keys:           keys,
		now:            now,
		serviceVersion: deps.Telemetry.ServiceVersion,
	}, nil
}

func (s *Service) handleWorkbookPreferencesDefault(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.GetDefaultPreferences(r.Context(), incidentID)
		if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
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
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		membership, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin")
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeDefaultPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := s.startupStore.ValidatePreferenceSheetRef(request.DefaultSheetRef, membership.Role, "default_sheet_ref"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.PutDefaultPreferences(r.Context(), incidentID, principal.User.ID, request.DefaultSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildDefaultPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleWorkbookPreferencesMe(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	switch r.Method {
	case http.MethodGet:
		if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.GetUserPreferences(r.Context(), incidentID, principal.User.ID)
		if errors.Is(err, workbookstartup.ErrPreferencesNotFound) {
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
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))

	case http.MethodPut:
		principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		if apiErr := s.startupStore.ValidatePreferenceSheetRef(request.HomeSheetRef, membership.Role, "home_sheet_ref"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		record, err := s.startupStore.PutUserPreferences(r.Context(), incidentID, principal.User.ID, request.HomeSheetRef, s.now())
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildUserPreferencesResource(record))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleWorkbookStartup(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	explicitSheetRef, apiErr := workbookstartup.ParseExplicitSheetRef(r.URL.Query())
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := s.startupStore.ValidateExplicitSheetRef(explicitSheetRef, membership.Role); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.startupStore.Resolve(r.Context(), incidentID, principal.User.ID, membership.Role, explicitSheetRef, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, workbookstartup.BuildStartupResource(record))
}

func (s *Service) handleQuery(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}

	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	query, apiErr := decodeViewQueryRequest(r, viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookQuery(r.Context(), viewSchemaID)
	r = r.WithContext(ctx)
	telemetryResult := "failed"
	telemetryErrorCode := ""
	telemetryRowCount := -1
	defer func() {
		finishTelemetry(telemetryResult, telemetryErrorCode, telemetryRowCount)
	}()
	scope, scopeErr := workbookQueryScope(incidentID, viewSchemaID, query.Meta)
	if scopeErr != nil {
		apiErr := internalAPIError(scopeErr)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	binding, cursor, reasonCode := s.cursorCodec.ResolveViewQuery(
		query.Pagination,
		"workbook.view-query",
		principal.User.ID.String(),
		scope,
	)
	if reasonCode != "" {
		apiErr := invalidViewQuery("", reasonCode)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}

	window := querypage.Window{Limit: binding.Limit}
	if cursor != nil {
		window.Position = cursor.Position
	}
	queryProvider, ok := s.contributions.QueryFor(viewSchemaID)
	if !ok {
		apiErr := internalAPIError(fmt.Errorf("workbook query surface %q is not registered", viewSchemaID))
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	page, err := queryProvider.QueryRowsPage(r.Context(), QueryCommand{
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Query:        query.Meta,
		Window:       window,
	})
	if err != nil {
		if errors.Is(err, querypage.ErrInvalidPosition) {
			apiErr := invalidViewQuery("", pagination.ReasonInvalidCursorToken)
			telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
			writeAPIError(w, r, apiErr)
			return
		}
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	rows, nextCursor, err := pageBoundedWorkbookResources(binding, query.Meta, page)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken), errors.Is(err, querypage.ErrInvalidPosition):
		apiErr := invalidViewQuery("", pagination.ReasonInvalidCursorToken)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	case err != nil:
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	var nextToken *string
	if nextCursor != nil {
		token, err := s.cursorCodec.Encode(*nextCursor)
		if err != nil {
			apiErr := internalAPIError(err)
			telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
			writeAPIError(w, r, apiErr)
			return
		}
		nextToken = &token
	}
	telemetryResult = "success"
	telemetryRowCount = len(rows)
	_ = httpapi.WriteSuccessWithMeta(w, r, http.StatusOK, map[string]any{
		"incident_id":    incidentID.String(),
		"view_schema_id": viewSchemaID,
		"rows":           rows,
	}, httpapi.EnvelopeMeta{
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Paging: &httpapi.PagingMeta{
			Limit:      binding.Limit,
			HasMore:    nextToken != nil,
			NextCursor: nextToken,
		},
		Query: query.Meta,
	})
}

func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.CreateFor(viewSchemaID)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}
	admission, apiErr := provider.DecodeCreate(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), viewSchemaID, "create")
	result, err := provider.Create(ctx, CreateCommand{
		Actor:        principal.User,
		IncidentID:   incidentID,
		ViewSchemaID: viewSchemaID,
		Admission:    admission,
		RequestID:    httpapi.RequestIDFromContext(ctx),
		Now:          s.now(),
	})
	telemetryResult, telemetryErrorCode := workbookMutationErrorTelemetry(err, admission.ClientTxnID)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeMutationResult(w, r, s, &principal, result, err, admission.ClientTxnID)
}

func (s *Service) handlePatch(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.Resolve(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	provider, registered := s.contributions.PatchFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	admission, apiErr := provider.DecodePatch(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), admission.ViewSchemaID, "patch")
	result, err := provider.Patch(ctx, PatchCommand{
		Actor:                   principal.User,
		RecordID:                recordID,
		AuthoritativeRecordType: target.RecordType,
		Admission:               admission,
		RequestID:               httpapi.RequestIDFromContext(ctx),
		Now:                     s.now(),
	})
	telemetryResult, telemetryErrorCode := workbookMutationErrorTelemetry(err, admission.ClientTxnID)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeMutationResult(w, r, s, &principal, result, err, admission.ClientTxnID)
}

func (s *Service) handleLinkedNoteCreate(w http.ResponseWriter, r *http.Request) {
	sourceRecordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeLinkedNoteCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, err := s.mutations.LinkedNoteSourceIncident(r.Context(), sourceRecordID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.mutations.CreateLinkedNote(r.Context(), principal.User, sourceRecordID, request, LinkedNoteCreateRequestHash(sourceRecordID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
}

func (s *Service) handleSupersede(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	target, err := s.recordTargets.Resolve(r.Context(), recordID)
	if isRecordTargetNotFound(err) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if target.Deleted {
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}})
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := timelineadmission.DecodeTimelineSupersedeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch target.RecordType {
	case "timeline_event":
		s.handleTimelineSupersede(w, r, &principal, recordID, request)
	case "decision":
		requestHash := timelineadmission.ActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID)
		s.handleDecisionSupersede(w, r, &principal, recordID, request, requestHash)
	default:
		writeAPIError(w, r, &httpapi.APIError{
			Status:  http.StatusConflict,
			Code:    "illegal_transition",
			Message: "illegal transition",
			Details: map[string]any{
				"reason_code":     "supersede_not_allowed",
				"violated_guards": []string{"unsupported_record_type"},
			},
		})
	}
}

func (s *Service) handleTimelineSupersede(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, recordID uuid.UUID, request timeline.SupersedeRequest) {
	result, err := s.timelineOwner.SupersedeRow(r.Context(), timeline.SupersedeCommand{
		Actor:    principal.User,
		RecordID: recordID,
		Request:  request,
		RequestHash: timelineadmission.ActionRequestHash(
			request.BaseRowVersion,
			request.ClientTxnID,
			&request.Reason,
			request.ReplacementRecordID,
		),
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Now:       s.now(),
	})
	if apiErr, ok := timelineadmission.ClassifyMutationAPIError(err, timelineadmission.MutationAPIErrorContext{
		ClientTxnID:                 request.ClientTxnID,
		IllegalTransitionReasonCode: "supersede_not_allowed",
	}); ok {
		writeAPIError(w, r, apiErr)
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleDecisionSupersede(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, recordID uuid.UUID, request timeline.SupersedeRequest, requestHash []byte) {
	result, err := s.mutations.SupersedeDecision(r.Context(), principal.User, recordID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		validationErr *MutationValidationError
		lifecycleErr  *LifecycleValidationError
		rowConflict   *RowVersionConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}})
		return
	case errors.As(err, &validationErr):
		writeAPIError(w, r, invalidMutationPayload(validationErr.Field, validationErr.ReasonCode))
		return
	case errors.As(err, &lifecycleErr):
		details := map[string]any{
			"from_status":     lifecycleErr.FromStatus,
			"to_status":       lifecycleErr.ToStatus,
			"violated_guards": append([]string(nil), lifecycleErr.ViolatedGuards...),
		}
		if lifecycleErr.ReasonCode != "" {
			details["reason_code"] = lifecycleErr.ReasonCode
		}
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: details})
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, rowVersionConflictError(rowConflict.Details()))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleConflictResolve(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	target, err := s.recordTargets.Resolve(r.Context(), recordID)
	switch {
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), target.IncidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	token := r.PathValue("conflict_token")
	claims, valid := s.conflictTokens.Parse(token)
	if !valid || claims.RecordID != recordID.String() {
		writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
		return
	}
	provider, registered := s.contributions.ConflictFor(target.RecordType)
	if !registered {
		writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
		return
	}
	admission, apiErr := provider.DecodeConflict(r.Body, token, claims)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := provider.ResolveConflict(r.Context(), ConflictCommand{
		Actor:                   principal.User,
		RecordID:                recordID,
		AuthoritativeRecordType: target.RecordType,
		Claims:                  claims,
		Admission:               admission,
		RequestID:               httpapi.RequestIDFromContext(r.Context()),
		Now:                     s.now(),
	})
	writeMutationResult(w, r, s, &principal, result, err, admission.ClientTxnID)
}

func writeMutationResult(w http.ResponseWriter, r *http.Request, s *Service, principal *httpauth.Principal, result MutationResult, err error, clientTxnID string) {
	var (
		publicErr     *publicMutationError
		validationErr *MutationValidationError
		lifecycleErr  *LifecycleValidationError
		rowConflict   *RowVersionConflictError
		sameConflict  *SameFieldConflictError
	)
	switch {
	case errors.As(err, &publicErr):
		writeAPIError(w, r, publicErr.apiError)
		return
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(clientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case isRecordTargetNotFound(err):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}})
		return
	case errors.As(err, &validationErr):
		writeAPIError(w, r, invalidMutationPayload(validationErr.Field, validationErr.ReasonCode))
		return
	case errors.As(err, &lifecycleErr):
		details := map[string]any{
			"from_status":     lifecycleErr.FromStatus,
			"to_status":       lifecycleErr.ToStatus,
			"violated_guards": append([]string(nil), lifecycleErr.ViolatedGuards...),
		}
		if lifecycleErr.ReasonCode != "" {
			details["reason_code"] = lifecycleErr.ReasonCode
		}
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: details})
		return
	case errors.As(err, &sameConflict):
		writeAPIError(w, r, sameFieldConflictError(sameConflict))
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, rowVersionConflictError(rowConflict.Details()))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func classifyTimelineMutationError(w http.ResponseWriter, r *http.Request, err error, context timelineadmission.MutationAPIErrorContext) bool {
	apiErr, ok := timelineadmission.ClassifyMutationAPIError(err, context)
	if !ok {
		return false
	}
	writeAPIError(w, r, apiErr)
	return true
}

func decodeViewQueryRequest(r *http.Request, viewSchemaID string) (viewquery.Query, *httpapi.APIError) {
	query, err := viewquery.Decode(r.Body, viewSchemaID)
	if err != nil {
		return viewquery.Query{}, invalidViewQueryValidation(err)
	}
	return query, nil
}

func workbookQueryScope(incidentID uuid.UUID, viewSchemaID string, queryMeta viewschema.QueryMeta) (map[string]string, error) {
	payload, err := json.Marshal(queryMeta)
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"incident_id":    incidentID.String(),
		"view_schema_id": viewSchemaID,
		"query_contract": string(payload),
	}, nil
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfPersistenceDue(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	value, err := uuid.Parse(raw)
	if err != nil {
		http.NotFound(w, r)
		return uuid.UUID{}, false
	}
	return value, true
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func isRecordTargetNotFound(err error) bool {
	return errors.Is(err, records.ErrEnvelopeNotFound) || errors.Is(err, pgx.ErrNoRows)
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func entityMatchConflictError(entityType string, identifierClass string, candidateRecordIDs []uuid.UUID) *httpapi.APIError {
	details := map[string]any{
		"reason_code":      "merge_required",
		"entity_type":      entityType,
		"identifier_class": identifierClass,
	}
	if len(candidateRecordIDs) > 0 {
		ids := make([]string, 0, len(candidateRecordIDs))
		for _, recordID := range candidateRecordIDs {
			ids = append(ids, recordID.String())
		}
		details["candidate_record_ids"] = ids
	}
	return &httpapi.APIError{Status: http.StatusConflict, Code: "entity_match_conflict", Message: "entity match conflict", Details: details}
}

func isTimelineMentionMutationError(err error) bool {
	var mentionTargetErr *mentions.MentionTargetValidationError
	return errors.Is(err, mentions.ErrEntityMentionNotFound) ||
		errors.Is(err, mentions.ErrResolvedRecordNotFound) ||
		errors.As(err, &mentionTargetErr) ||
		errors.Is(err, mentions.ErrInvalidMentionResolution)
}

func invalidViewQuery(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidViewQueryValidation(err *viewquery.ValidationError) *httpapi.APIError {
	if err == nil {
		return invalidViewQuery("", "")
	}
	details := map[string]any{}
	if err.Field != "" {
		details["field"] = err.Field
	}
	if err.FieldKey != "" {
		details["field_key"] = err.FieldKey
	}
	if err.FilterIndex != nil {
		details["filter_index"] = *err.FilterIndex
	}
	if err.ReasonCode != "" {
		details["reason_code"] = err.ReasonCode
	}
	if err.RequestedCount != nil {
		details["requested_count"] = *err.RequestedCount
	}
	if err.MaxCount != nil {
		details["max_count"] = *err.MaxCount
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}
