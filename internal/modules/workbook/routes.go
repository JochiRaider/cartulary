package workbook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	store          *Store
	incidentAccess incidents.Access
	startupStore   *workbookstartup.Store
	authStore      *authn.Store
	publisher      *collaboration.RecordChangePublisher
	cursorCodec    *pagination.Codec
	keys           authn.MasterKeys
	now            func() time.Time
	serviceVersion string
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query", service.handleQuery)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste", service.handleClipboardPaste)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/bulk-mutations", service.handleBulkMutations)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", service.handleCreate)
		mux.HandleFunc("GET /api/v1/incidents/{incident_id}/workbook-preferences/default", service.handleWorkbookPreferencesDefault)
		mux.HandleFunc("PUT /api/v1/incidents/{incident_id}/workbook-preferences/default", service.handleWorkbookPreferencesDefault)
		mux.HandleFunc("GET /api/v1/incidents/{incident_id}/workbook-preferences/me", service.handleWorkbookPreferencesMe)
		mux.HandleFunc("PUT /api/v1/incidents/{incident_id}/workbook-preferences/me", service.handleWorkbookPreferencesMe)
		mux.HandleFunc("GET /api/v1/incidents/{incident_id}/workbook-startup", service.handleWorkbookStartup)
		mux.HandleFunc("PATCH /api/v1/records/{record_id}", service.handlePatch)
		mux.HandleFunc("POST /api/v1/records/{record_id}/linked-notes", service.handleLinkedNoteCreate)
		mux.HandleFunc("POST /api/v1/records/{record_id}/supersede", service.handleSupersede)
		mux.HandleFunc("POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve", service.handleConflictResolve)
		return nil
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
	request, apiErr := timeline.DecodeBulkMutationRequest(r.Body, viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.timelineStore.ApplyBulkMutation(r.Context(), timeline.BulkMutationCommand{
		Actor:      principal.User,
		IncidentID: incidentID,
		Request:    request,
		RequestID:  httpapi.RequestIDFromContext(r.Context()),
		Now:        s.now(),
	})
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timeline.MutationAPIErrorContext{ClientTxnID: request.ClientTxnID}):
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
	if result.ChangeSetID != uuid.Nil && !result.Replayed {
		for _, row := range result.Rows {
			s.publishRecordChange(MutationResult{
				Payload:          map[string]any{"row": row.Row},
				IncidentID:       incidentID,
				RecordID:         row.RecordID,
				ChangeSetID:      result.ChangeSetID,
				ClientTxnID:      result.ClientTxnID,
				RowVersion:       row.RowVersion,
				ViewSchemaID:     timeline.TimelineViewSchemaID,
				ChangedFieldKeys: row.ChangedFieldKeys,
			}, principal.User.ID)
		}
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
	request, apiErr := timeline.DecodeTimelineClipboardPasteRequest(bytes.NewReader(body))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, err := timeline.BuildClipboardPastePlan(request); err != nil {
		writeAPIError(w, r, invalidMutationPayload("clipboard_text", "invalid_value"))
		return
	}
	result, err := s.store.timelineStore.ApplyClipboardPaste(r.Context(), timeline.ClipboardPasteCommand{
		Actor:      principal.User,
		IncidentID: incidentID,
		Request:    request,
		RequestID:  httpapi.RequestIDFromContext(r.Context()),
		Now:        s.now(),
	})
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timeline.MutationAPIErrorContext{ClientTxnID: request.ClientTxnID}):
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
	if result.ChangeSetID != uuid.Nil && !result.Replayed {
		for _, row := range result.Rows {
			s.publishRecordChange(MutationResult{
				Payload:          map[string]any{"row": row.Row},
				IncidentID:       incidentID,
				RecordID:         row.RecordID,
				ChangeSetID:      result.ChangeSetID,
				ClientTxnID:      result.ClientTxnID,
				RowVersion:       row.RowVersion,
				ViewSchemaID:     timeline.TimelineViewSchemaID,
				ChangedFieldKeys: row.ChangedFieldKeys,
			}, principal.User.ID)
		}
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
	request, apiErr := decodeEntityClipboardPasteRequest(bytes.NewReader(body), viewSchemaID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	plan, err := buildEntityClipboardPastePlan(request)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("clipboard_text", "invalid_value"))
		return
	}
	result, err := s.store.entityStore.ApplyClipboardPastePlan(
		r.Context(),
		principal.User,
		incidentID,
		request.ViewSchemaID,
		plan,
		entityClipboardPasteRequestHash(request),
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
	if result.ChangeSetID != uuid.Nil && !result.Replayed {
		for _, row := range result.Rows {
			s.publishRecordChange(MutationResult{
				Payload:          map[string]any{"row": row.Row},
				IncidentID:       incidentID,
				RecordID:         row.RecordID,
				ChangeSetID:      result.ChangeSetID,
				ClientTxnID:      result.ClientTxnID,
				RowVersion:       row.RowVersion,
				ViewSchemaID:     viewSchemaID,
				ChangedFieldKeys: row.ChangedFieldKeys,
			}, principal.User.ID)
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
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
	timelineStore := timeline.FacadeFromDependencies(deps)
	conflictTokens := revisions.NewConflictTokenCodec(keys)
	timelineStore.SetConflictTokenCodec(conflictTokens)
	store := newStoreWithTimelineFacade(deps.PostgresHandle(), timelineStore)
	store.SetConflictTokenCodec(conflictTokens)
	return &Service{
		store:          store,
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		startupStore:   workbookstartup.NewStore(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		publisher:      collaboration.NewRecordChangePublisher(deps.WSHub),
		cursorCodec:    cursorCodec,
		keys:           keys,
		now:            now,
		serviceVersion: deps.Config.Telemetry.Resource.ServiceVersion,
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
		if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeDefaultPreferencesPutRequest(r.Body)
		if apiErr != nil {
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
		if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		request, apiErr := workbookstartup.DecodeUserPreferencesPutRequest(r.Body)
		if apiErr != nil {
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
	record, err := s.startupStore.Resolve(r.Context(), incidentID, principal.User.ID, membership.Role, explicitSheetRef, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
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

	resources, err := s.store.QueryRows(r.Context(), incidentID, viewSchemaID, query.Meta)
	if err != nil {
		apiErr := internalAPIError(err)
		telemetryResult, telemetryErrorCode = workbookAPIErrorTelemetry(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	rows, nextCursor, err := pageWorkbookResources(binding, cursor, query.Meta, resources)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken):
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
	timing := newTimelineCreateTiming(r, viewSchemaID)
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	timing.mark("auth")
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "editor", "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	timing.mark("role_check")
	if viewSchemaID == timeline.TimelineViewSchemaID {
		s.handleTimelineCreate(w, r, principal, incidentID, timing)
		return
	}
	if viewSchemaID == hostidentity.HostsViewSchemaID || viewSchemaID == hostidentity.IdentitiesViewSchemaID {
		s.handleEntityCreate(w, r, principal, incidentID, viewSchemaID)
		return
	}
	if viewSchemaID == indicators.ViewSchemaID {
		s.handleIndicatorCreate(w, r, principal, incidentID)
		return
	}
	request, apiErr := DecodeCreateRequest(viewSchemaID, r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), request.ViewSchemaID, "create")
	result, err := s.store.CreateWorkbookRow(ctx, principal.User, incidentID, request, CreateRequestHash(request), httpapi.RequestIDFromContext(ctx), s.now())
	telemetryResult, telemetryErrorCode := workbookMutationErrorTelemetry(err, request.ClientTxnID)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
}

func (s *Service) handleEntityCreate(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, incidentID uuid.UUID, viewSchemaID string) {
	request, apiErr := hostidentity.DecodeCreateRequest(viewSchemaID, r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	var (
		result hostidentity.MutationResult
		err    error
	)
	requestHash := hostidentity.CreateRequestHash(viewSchemaID, request)
	switch viewSchemaID {
	case hostidentity.HostsViewSchemaID:
		result, err = s.store.entityStore.CreateHostRow(r.Context(), principal.User, incidentID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	case hostidentity.IdentitiesViewSchemaID:
		result, err = s.store.entityStore.CreateIdentityRow(r.Context(), principal.User, incidentID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	default:
		writeAPIError(w, r, invalidMutationPayload("view_schema_id", "unsupported_view_schema"))
		return
	}

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

	row, _ := result.Payload["row"].(map[string]any)
	if !result.Replayed {
		s.publishRecordChange(MutationResult{
			Payload:          result.Payload,
			IncidentID:       incidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      request.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     viewSchemaID,
			ChangedFieldKeys: changedFieldKeys(nil, row),
		}, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleIndicatorCreate(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, incidentID uuid.UUID) {
	request, apiErr := indicators.DecodeCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	requestHash := indicators.CreateRequestHash(request)
	result, err := s.store.indicatorStore.CreateIndicatorRow(r.Context(), principal.User, incidentID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())

	var createValidationErr *indicators.IndicatorCreateValidationError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, indicators.ErrInvalidCreateRequest):
		writeAPIError(w, r, invalidMutationPayload("payload", "at_least_one_value_required"))
		return
	case errors.As(err, &createValidationErr):
		writeAPIError(w, r, invalidMutationPayload(createValidationErr.Field, createValidationErr.ReasonCode))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	row, _ := result.Payload["row"].(map[string]any)
	if !result.Replayed {
		s.publishRecordChange(MutationResult{
			Payload:          result.Payload,
			IncidentID:       incidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      request.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     indicators.ViewSchemaID,
			ChangedFieldKeys: changedFieldKeys(nil, row),
		}, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeAPIError(w, r, invalidMutationPayload("", "invalid_value"))
		return
	}
	viewSchemaID := patchViewSchemaID(body)
	if viewSchemaID == timeline.TimelineViewSchemaID {
		s.handleTimelinePatch(w, r, principal, recordID, body)
		return
	}
	request, apiErr := DecodePatchRequest(bytes.NewReader(body))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, err := s.store.RecordIncident(r.Context(), recordID, request.ViewSchemaID)
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
	ctx, finishTelemetry := s.startWorkbookMutation(r.Context(), request.ViewSchemaID, "patch")
	result, err := s.store.PatchWorkbookRow(ctx, principal.User, recordID, request, PatchRequestHash(request), httpapi.RequestIDFromContext(ctx), s.now())
	telemetryResult, telemetryErrorCode := workbookMutationErrorTelemetry(err, request.ClientTxnID)
	finishTelemetry(telemetryResult, telemetryErrorCode)
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
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
	incidentID, err := s.store.LinkedNoteSourceIncident(r.Context(), sourceRecordID)
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
	result, err := s.store.CreateLinkedNote(r.Context(), principal.User, sourceRecordID, request, LinkedNoteCreateRequestHash(sourceRecordID, request), httpapi.RequestIDFromContext(r.Context()), s.now())
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
	target, err := s.store.RecordRouteTarget(r.Context(), recordID)
	if errors.Is(err, pgx.ErrNoRows) {
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
	request, apiErr := timeline.DecodeTimelineSupersedeRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	switch target.RecordType {
	case "timeline_event":
		s.handleTimelineSupersede(w, r, &principal, recordID, request)
	case "decision":
		requestHash := timeline.TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID)
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
	result, err := s.store.timelineStore.SupersedeRow(r.Context(), timeline.SupersedeCommand{
		Actor:     principal.User,
		RecordID:  recordID,
		Request:   request,
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Now:       s.now(),
	})
	if apiErr, ok := timeline.ClassifyMutationAPIError(err, timeline.MutationAPIErrorContext{
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
	if !result.Replayed {
		s.publishRecordChange(MutationResult{
			IncidentID:       result.IncidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     timeline.TimelineViewSchemaID,
			ChangedFieldKeys: result.ChangedFieldKeys,
		}, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleDecisionSupersede(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, recordID uuid.UUID, request timeline.SupersedeRequest, requestHash []byte) {
	result, err := s.store.SupersedeDecision(r.Context(), principal.User, recordID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
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
	case errors.Is(err, pgx.ErrNoRows):
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
	if !result.Replayed {
		for _, change := range result.AdditionalRecordChanges {
			s.publishRecordChange(change, principal.User.ID)
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleConflictResolve(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	token := r.PathValue("conflict_token")
	claims, valid := s.store.parseWorkbookConflictToken(token)
	if !valid {
		timelineClaims, timelineValid := s.store.timelineStore.ParseConflictToken(token)
		if !timelineValid || timelineClaims.RecordID != recordID.String() {
			writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
			return
		}
		s.handleTimelineConflictResolve(w, r, recordID, token, timelineClaims)
		return
	}
	if claims.ViewSchemaID == timeline.TimelineViewSchemaID {
		timelineClaims, timelineValid := s.store.timelineStore.ParseConflictToken(token)
		if !timelineValid || timelineClaims.RecordID != recordID.String() {
			writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
			return
		}
		s.handleTimelineConflictResolve(w, r, recordID, token, timelineClaims)
		return
	}
	if claims.RecordID != recordID.String() {
		writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeConflictResolveRequest(r.Body, token, claims)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, err := s.store.RecordIncident(r.Context(), recordID, claims.ViewSchemaID)
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
	result, err := s.store.ResolveWorkbookConflict(r.Context(), principal.User, recordID, claims, request, ConflictResolveRequestHash(claims, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
}

func (s *Service) handleTimelineConflictResolve(w http.ResponseWriter, r *http.Request, recordID uuid.UUID, token string, claims timeline.TimelineConflictTokenClaims) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := timeline.DecodeTimelineConflictResolveRequest(r.Body, token, claims)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	incidentID, err := s.store.RecordIncident(r.Context(), recordID, timeline.TimelineViewSchemaID)
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
	result, err := s.store.timelineStore.ResolveConflict(r.Context(), timeline.ConflictResolveCommand{
		Actor:     principal.User,
		RecordID:  recordID,
		Claims:    claims,
		Request:   request,
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Now:       s.now(),
	})
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timeline.MutationAPIErrorContext{ClientTxnID: request.ClientTxnID}):
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case isTimelineMentionMutationError(err):
		writeAPIError(w, r, invalidMutationPayload("action_payload", "invalid_value"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if result.ChangeSetID != uuid.Nil && !result.Replayed {
		s.publishRecordChange(MutationResult{
			IncidentID:       incidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     timeline.TimelineViewSchemaID,
			ChangedFieldKeys: result.ChangedFieldKeys,
		}, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleTimelineCreate(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, incidentID uuid.UUID, timing *timelineCreateTiming) {
	request, apiErr := timeline.DecodeTimelineCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	timing.mark("decode")
	timing.mark("hash")
	storeCtx := timeline.WithCreateTimingRecorder(r.Context(), timing)
	result, err := s.store.timelineStore.CreateRow(storeCtx, timeline.CreateRowCommand{
		Actor:      principal.User,
		IncidentID: incidentID,
		Request:    request,
		RequestID:  httpapi.RequestIDFromContext(r.Context()),
		Now:        s.now(),
	})
	timing.mark("store_create")
	var mutationErr *MutationValidationError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.As(err, &mutationErr):
		writeAPIError(w, r, invalidMutationPayload(mutationErr.Field, mutationErr.ReasonCode))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if !result.Replayed {
		s.publishRecordChange(MutationResult{
			IncidentID:       incidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     timeline.TimelineViewSchemaID,
			ChangedFieldKeys: result.ChangedFieldKeys,
		}, principal.User.ID)
	}
	timing.mark("websocket_publish")
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	timing.mark("session_slide")
	timing.mark("response_prep")
	timing.write(w)
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func (s *Service) handleTimelinePatch(w http.ResponseWriter, r *http.Request, principal httpauth.Principal, recordID uuid.UUID, body []byte) {
	incidentID, err := s.store.RecordIncident(r.Context(), recordID, timeline.TimelineViewSchemaID)
	if errors.Is(err, pgx.ErrNoRows) {
		writeAPIError(w, r, incidentNotFoundError())
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
	request, apiErr := timeline.DecodeTimelinePatchRequest(bytes.NewReader(body))
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.timelineStore.PatchRow(r.Context(), timeline.PatchRowCommand{
		Actor:     principal.User,
		RecordID:  recordID,
		Request:   request,
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Now:       s.now(),
	})
	var (
		entityConflict       *hostidentity.ExactMatchConflictError
		mentionTransitionErr *mentions.MentionTransitionError
	)
	switch {
	case classifyTimelineMutationError(w, r, err, timeline.MutationAPIErrorContext{
		ClientTxnID:                 request.ClientTxnID,
		IllegalTransitionReasonCode: "superseded_terminal",
		NoEffectiveChangeField:      "changes",
	}):
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case isTimelineMentionMutationError(err):
		writeAPIError(w, r, invalidMutationPayload("action_payload", "invalid_value"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	result.IncidentID = incidentID
	if !result.Replayed {
		s.publishRecordChange(MutationResult{
			IncidentID:       incidentID,
			RecordID:         result.RecordID,
			ChangeSetID:      result.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			RowVersion:       result.RowVersion,
			ViewSchemaID:     timeline.TimelineViewSchemaID,
			ChangedFieldKeys: result.ChangedFieldKeys,
		}, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func writeMutationResult(w http.ResponseWriter, r *http.Request, s *Service, principal *httpauth.Principal, result MutationResult, err error, clientTxnID string) {
	var (
		validationErr *MutationValidationError
		lifecycleErr  *LifecycleValidationError
		rowConflict   *RowVersionConflictError
		sameConflict  *SameFieldConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, httpapi.ClientTxnConflictError(clientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, pgx.ErrNoRows):
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
	if !result.Replayed {
		s.publishRecordChange(result, principal.User.ID)
	}
	if err := s.slideSessionIfNeeded(r.Context(), principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, result.StatusCode, result.Payload)
}

func classifyTimelineMutationError(w http.ResponseWriter, r *http.Request, err error, context timeline.MutationAPIErrorContext) bool {
	apiErr, ok := timeline.ClassifyMutationAPIError(err, context)
	if !ok {
		return false
	}
	writeAPIError(w, r, apiErr)
	return true
}

func patchViewSchemaID(body []byte) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return ""
	}
	var viewSchemaID string
	_ = json.Unmarshal(raw["view_schema_id"], &viewSchemaID)
	return viewSchemaID
}

func (s *Service) publishRecordChange(result MutationResult, actorUserID uuid.UUID) {
	if result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		return
	}
	row, _ := result.Payload["row"].(map[string]any)
	s.publisher.Publish(collaboration.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		RowVersion:       result.RowVersion,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: result.ChangedFieldKeys,
		ViewSchemaID:     result.ViewSchemaID,
		Row:              row,
	})
}

type timelineCreateTiming struct {
	enabled bool
	last    time.Time
	parts   []string
}

func newTimelineCreateTiming(r *http.Request, viewSchemaID string) *timelineCreateTiming {
	if viewSchemaID != timeline.TimelineViewSchemaID || r.Header.Get("X-Cartulary-Timing-Debug") != "1" {
		return nil
	}
	return &timelineCreateTiming{
		enabled: true,
		last:    time.Now(),
	}
}

func (t *timelineCreateTiming) MarkTimelineCreateTiming(name string) {
	t.mark(name)
}

func (t *timelineCreateTiming) AddTimelineCreateTiming(name string, duration time.Duration) {
	if t == nil || !t.enabled {
		return
	}
	t.parts = append(t.parts, fmt.Sprintf("%s;dur=%.3f", name, float64(duration.Microseconds())/1000))
	t.last = time.Now()
}

func (t *timelineCreateTiming) mark(name string) {
	if t == nil || !t.enabled {
		return
	}
	now := time.Now()
	if t.last.IsZero() {
		t.last = now
		return
	}
	t.parts = append(t.parts, fmt.Sprintf("%s;dur=%.3f", name, float64(now.Sub(t.last).Microseconds())/1000))
	t.last = now
}

func (t *timelineCreateTiming) write(w http.ResponseWriter) {
	if t == nil || !t.enabled || len(t.parts) == 0 {
		return
	}
	w.Header().Set("Server-Timing", strings.Join(t.parts, ", "))
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
	record, err := s.incidentAccess.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, incidentNotFoundError()
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	if !slices.Contains(roles, record.Role) {
		return incidents.MembershipRecord{}, authorizationDeniedError(requiredRoleDescription(roles...))
	}
	return record, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	if principal == nil || !httpauth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	now := s.now()
	if !httpauth.ShouldPersistIdleExpirySlide(sliding, now) {
		return nil
	}
	sliding = sliding.Slide(now)
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
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

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *httpapi.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &httpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Message: "authorization denied", Details: details}
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

func requiredRoleDescription(roles ...string) string {
	if len(roles) == 0 {
		return ""
	}
	if len(roles) == 1 {
		return roles[0]
	}
	return strings.Join(roles, "|")
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
