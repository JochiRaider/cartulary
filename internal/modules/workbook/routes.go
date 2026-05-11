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

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	store         *Store
	incidentStore *incidents.Store
	authStore     *authn.Store
	hub           *platformws.Hub
	pagination    *pagination.Registry
	keys          authn.MasterKeys
	now           func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/query", service.handleQuery)
		mux.HandleFunc("POST /api/v1/incidents/{incident_id}/views/{view_schema_id}/rows", service.handleCreate)
		mux.HandleFunc("PATCH /api/v1/records/{record_id}", service.handlePatch)
		mux.HandleFunc("POST /api/v1/records/{record_id}/conflicts/{conflict_token}/resolve", service.handleConflictResolve)
		return nil
	}
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
	paginator := deps.Pagination
	if paginator == nil {
		paginator = pagination.NewRegistry()
	}
	return &Service{
		store:         NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		authStore:     authn.NewStore(deps.Postgres),
		hub:           deps.WSHub,
		pagination:    paginator,
		keys:          keys,
		now:           now,
	}, nil
}

func (s *Service) handleQuery(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}

	principal, apiErr := s.authenticateSessionRequest(r, false)
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
	scope, scopeErr := workbookQueryScope(incidentID, viewSchemaID, query.Meta)
	if scopeErr != nil {
		writeAPIError(w, r, internalAPIError(scopeErr))
		return
	}
	binding, cursor, reasonCode := pagination.ResolveViewQuery(
		query.Pagination,
		"workbook.view-query",
		principal.User.ID.String(),
		scope,
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidViewQuery("", reasonCode))
		return
	}

	var (
		rows       []json.RawMessage
		nextCursor *pagination.Cursor
		err        error
	)
	if cursor == nil {
		resources, err := s.store.QueryRows(r.Context(), incidentID, viewSchemaID, query.Meta)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rows, err = pagination.MarshalResources(resources)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		rows, nextCursor = s.pagination.Start(binding, rows)
	} else {
		rows, nextCursor, err = s.pagination.Continue(binding, *cursor)
		switch {
		case errors.Is(err, pagination.ErrCursorSnapshotExpired):
			writeAPIError(w, r, invalidViewQuery("", pagination.ReasonCursorSnapshotUnavailable))
			return
		case errors.Is(err, pagination.ErrInvalidCursorToken):
			writeAPIError(w, r, invalidViewQuery("", pagination.ReasonInvalidCursorToken))
			return
		case err != nil:
			writeAPIError(w, r, internalAPIError(err))
			return
		}
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	var nextToken *string
	if nextCursor != nil {
		token, err := pagination.EncodeCursor(*nextCursor)
		if err != nil {
			writeAPIError(w, r, internalAPIError(err))
			return
		}
		nextToken = &token
	}
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
		"incident_id":    incidentID.String(),
		"view_schema_id": viewSchemaID,
		"rows":           rows,
	}, httpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *Service) handleCreate(w http.ResponseWriter, r *http.Request) {
	viewSchemaID := r.PathValue("view_schema_id")
	timing := newTimelineCreateTiming(r, viewSchemaID)
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
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
	request, apiErr := DecodeCreateRequest(viewSchemaID, r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.store.CreateWorkbookRow(r.Context(), principal.User, incidentID, request, CreateRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
}

func (s *Service) handlePatch(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, true)
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
	result, err := s.store.PatchWorkbookRow(r.Context(), principal.User, recordID, request, PatchRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	writeMutationResult(w, r, s, &principal, result, err, request.ClientTxnID)
}

func (s *Service) handleConflictResolve(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	token := r.PathValue("conflict_token")
	claims, valid := parseWorkbookConflictToken(token)
	if !valid {
		timelineClaims, timelineValid := timeline.ParseConflictToken(token)
		if !timelineValid || timelineClaims.RecordID != recordID.String() {
			writeAPIError(w, r, invalidMutationPayload("conflict_token", "invalid_value"))
			return
		}
		s.handleTimelineConflictResolve(w, r, recordID, token, timelineClaims)
		return
	}
	if claims.ViewSchemaID == timeline.TimelineViewSchemaID {
		timelineClaims, timelineValid := timeline.ParseConflictToken(token)
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
	principal, apiErr := s.authenticateSessionRequest(r, true)
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
	principal, apiErr := s.authenticateSessionRequest(r, true)
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
	result, err := s.store.timelineStore.ResolveConflict(r.Context(), principal.User, recordID, claims, request, timeline.TimelineConflictResolveRequestHash(claims, request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		entityConflict        *entities.ExactMatchConflictError
		mentionTransitionErr  *entities.MentionTransitionError
		mentionTargetErr      *entities.MentionTargetValidationError
		timelineTransitionErr *timeline.IllegalTransitionError
		rowConflict           *timeline.RowVersionConflictError
		sameFieldConflict     *timeline.SameFieldConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, timeline.ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}})
		return
	case errors.As(err, &sameFieldConflict):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "same_field_conflict", Message: "same field conflict", Details: map[string]any{}, Conflict: sameFieldConflict.Conflict})
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, rowVersionConflictError(rowConflict.Details()))
		return
	case errors.Is(err, timeline.ErrRowVersionConflict):
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.As(err, &timelineTransitionErr):
		details := map[string]any{
			"from_status":     timelineTransitionErr.FromStatus,
			"to_status":       timelineTransitionErr.ToStatus,
			"violated_guards": append([]string(nil), timelineTransitionErr.ViolatedGuards...),
		}
		if timelineTransitionErr.ReasonCode != "" {
			details["reason_code"] = timelineTransitionErr.ReasonCode
		}
		writeAPIError(w, r, &auth.APIError{
			Status:  http.StatusConflict,
			Code:    "illegal_transition",
			Message: "illegal transition",
			Details: details,
		})
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.Is(err, timeline.ErrNoEffectiveChange):
		writeAPIError(w, r, invalidMutationPayload("changes", "no_effective_change"))
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case errors.Is(err, entities.ErrEntityMentionNotFound), errors.Is(err, entities.ErrResolvedRecordNotFound), errors.As(err, &mentionTargetErr), errors.Is(err, entities.ErrInvalidMentionResolution):
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

func (s *Service) handleTimelineCreate(w http.ResponseWriter, r *http.Request, principal auth.SessionPrincipal, incidentID uuid.UUID, timing *timelineCreateTiming) {
	request, apiErr := timeline.DecodeTimelineCreateRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	timing.mark("decode")
	requestHash := timeline.TimelineCreateRequestHash(request)
	timing.mark("hash")
	storeCtx := timeline.WithCreateTimingRecorder(r.Context(), timing)
	result, err := s.store.timelineStore.CreateRow(storeCtx, principal.User, incidentID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	timing.mark("store_create")
	var mutationErr *MutationValidationError
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
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

func (s *Service) handleTimelinePatch(w http.ResponseWriter, r *http.Request, principal auth.SessionPrincipal, recordID uuid.UUID, body []byte) {
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
	result, err := s.store.timelineStore.PatchRow(r.Context(), principal.User, recordID, request, timeline.TimelinePatchRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	var (
		entityConflict        *entities.ExactMatchConflictError
		mentionTransitionErr  *entities.MentionTransitionError
		mentionTargetErr      *entities.MentionTargetValidationError
		timelineTransitionErr *timeline.IllegalTransitionError
		rowConflict           *timeline.RowVersionConflictError
		sameFieldConflict     *timeline.SameFieldConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, timeline.ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.As(err, &sameFieldConflict):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "same_field_conflict", Message: "same field conflict", Details: map[string]any{}, Conflict: sameFieldConflict.Conflict})
		return
	case errors.As(err, &rowConflict):
		writeAPIError(w, r, rowVersionConflictError(rowConflict.Details()))
		return
	case errors.Is(err, timeline.ErrRowVersionConflict):
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.As(err, &timelineTransitionErr):
		details := map[string]any{
			"from_status":     timelineTransitionErr.FromStatus,
			"to_status":       timelineTransitionErr.ToStatus,
			"violated_guards": append([]string{}, timelineTransitionErr.ViolatedGuards...),
		}
		if timelineTransitionErr.ReasonCode != "" {
			details["reason_code"] = timelineTransitionErr.ReasonCode
		}
		writeAPIError(w, r, &auth.APIError{
			Status:  http.StatusConflict,
			Code:    "illegal_transition",
			Message: "illegal transition",
			Details: details,
		})
		return
	case errors.Is(err, timeline.ErrIllegalTransition):
		writeAPIError(w, r, &auth.APIError{
			Status:  http.StatusConflict,
			Code:    "illegal_transition",
			Message: "illegal transition",
			Details: map[string]any{"reason_code": "superseded_terminal"},
		})
		return
	case errors.As(err, &mentionTransitionErr):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: map[string]any{"from_status": mentionTransitionErr.FromStatus, "to_status": mentionTransitionErr.ToStatus, "violated_guards": append([]string(nil), mentionTransitionErr.ViolatedGuards...)}})
		return
	case errors.Is(err, timeline.ErrNoEffectiveChange):
		writeAPIError(w, r, invalidMutationPayload("changes", "no_effective_change"))
		return
	case errors.As(err, &entityConflict):
		writeAPIError(w, r, entityMatchConflictError(entityConflict.EntityType, entityConflict.IdentifierClass, entityConflict.CandidateRecords))
		return
	case errors.Is(err, entities.ErrEntityMentionNotFound), errors.Is(err, entities.ErrResolvedRecordNotFound), errors.As(err, &mentionTargetErr), errors.Is(err, entities.ErrInvalidMentionResolution):
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

func writeMutationResult(w http.ResponseWriter, r *http.Request, s *Service, principal *auth.SessionPrincipal, result MutationResult, err error, clientTxnID string) {
	var (
		validationErr *MutationValidationError
		lifecycleErr  *LifecycleValidationError
		rowConflict   *RowVersionConflictError
		sameConflict  *SameFieldConflictError
	)
	switch {
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, auth.ClientTxnConflictError(clientTxnID))
		return
	case errors.Is(err, pgx.ErrNoRows):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Message: "record deleted use restore", Details: map[string]any{}})
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
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "illegal_transition", Message: "illegal transition", Details: details})
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
	if s.hub == nil || result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		return
	}
	changedKeys := append([]string(nil), result.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	row, _ := result.Payload["row"].(map[string]any)
	patchCells := platformws.BuildViewRowPatch(row, changedKeys)
	s.hub.PublishRecordChange(platformws.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		RowVersion:       result.RowVersion,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: changedKeys,
		ViewSchemaID:     result.ViewSchemaID,
		PatchCells:       patchCells,
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

func decodeViewQueryRequest(r *http.Request, viewSchemaID string) (viewquery.Query, *auth.APIError) {
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

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *auth.APIError) {
	record, err := s.incidentStore.GetIncidentMembershipForUser(ctx, incidentID, userID)
	if errors.Is(err, incidents.ErrMembershipNotFound) {
		return incidents.MembershipRecord{}, incidentNotFoundError()
	}
	if err != nil {
		return incidents.MembershipRecord{}, internalAPIError(err)
	}
	return record, nil
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *auth.APIError) {
	record, apiErr := s.requireIncidentMembership(ctx, incidentID, userID)
	if apiErr != nil {
		return incidents.MembershipRecord{}, apiErr
	}
	if !slices.Contains(roles, record.Role) {
		return incidents.MembershipRecord{}, authorizationDeniedError(requiredRoleDescription(roles...))
	}
	return record, nil
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return s.authenticateSessionToken(r, auth.AuthSourceBearer, token, stateChanging)
	}

	cookie, err := r.Cookie(authn.SessionCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
		}
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	return s.authenticateSessionToken(r, auth.AuthSourceCookie, cookie.Value, stateChanging)
}

func (s *Service) authenticateSessionToken(r *http.Request, authSource auth.AuthSource, sessionToken string, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	if stateChanging && authSource == auth.AuthSourceCookie {
		csrfCookie, _ := r.Cookie(authn.CSRFCookieName)
		if csrfCookie == nil || csrfCookie.Value != authn.CSRFTokenForSessionToken(s.keys, sessionToken) {
			return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusForbidden, Code: "csrf_verification_failed", Details: map[string]any{}}
		}
		if apiErr := auth.ValidateCSRF(r.Method, authSource, csrfCookie.Value, r.Header.Get(authn.CSRFHeaderName)); apiErr != nil {
			return auth.SessionPrincipal{}, apiErr
		}
	}

	session, user, err := s.authStore.GetSessionByFingerprint(r.Context(), authn.FingerprintToken(s.keys, sessionToken))
	if errors.Is(err, authn.ErrNotFound) {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}
	if err != nil {
		return auth.SessionPrincipal{}, internalAPIError(err)
	}
	if !user.IsActive || session.RevokedAt != nil {
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}

	now := s.now()
	if !session.SessionExpiresAt.After(now) {
		_ = s.authStore.RevokeSession(context.Background(), session.ID, "session_expired", now)
		return auth.SessionPrincipal{}, &auth.APIError{Status: http.StatusUnauthorized, Code: "session_required", Details: map[string]any{}}
	}

	return auth.SessionPrincipal{
		AuthSource:   authSource,
		SessionToken: sessionToken,
		Session:      session,
		User:         user,
	}, nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if principal == nil || !auth.ShouldSlideIdleExpiry(method, path) {
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
	if !auth.ShouldPersistIdleExpirySlide(sliding, now) {
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

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
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

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}

func incidentNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func authorizationDeniedError(requiredRole string) *auth.APIError {
	details := map[string]any{}
	if requiredRole != "" {
		details["required_role"] = requiredRole
	}
	return &auth.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Message: "authorization denied", Details: details}
}

func entityMatchConflictError(entityType string, identifierClass string, candidateRecordIDs []uuid.UUID) *auth.APIError {
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
	return &auth.APIError{Status: http.StatusConflict, Code: "entity_match_conflict", Message: "entity match conflict", Details: details}
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

func invalidViewQuery(field string, reasonCode string) *auth.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}

func invalidViewQueryValidation(err *viewquery.ValidationError) *auth.APIError {
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
	return &auth.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_view_query",
		Message: "invalid view query",
		Details: details,
	}
}
