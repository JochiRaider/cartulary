package httpapi

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type ownerApplication interface {
	CreateIndicatorObservation(context.Context, authn.UserRecord, indicators.IndicatorObservationCreateParams) (indicators.IndicatorObservationMutationResult, error)
	ResolveIndicatorObservation(context.Context, authn.UserRecord, indicators.IndicatorObservationResolveParams) (indicators.IndicatorObservationMutationResult, error)
	DismissIndicatorObservation(context.Context, authn.UserRecord, indicators.IndicatorObservationActionParams) (indicators.IndicatorObservationMutationResult, error)
	RestoreIndicatorObservation(context.Context, authn.UserRecord, indicators.IndicatorObservationActionParams) (indicators.IndicatorObservationMutationResult, error)
	AppendIndicatorLifecycleInterval(context.Context, authn.UserRecord, indicators.IndicatorLifecycleAppendParams) (indicators.IndicatorLifecycleMutationResult, error)
	GetIndicatorObservation(context.Context, uuid.UUID) (indicators.IndicatorObservationRecord, error)
	ListSourceRecordIndicatorObservations(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorObservationRecord, error)
	ListIndicatorObservations(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorObservationRecord, error)
	ListIndicatorLifecycleIntervals(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorLifecycleIntervalRecord, error)
}

type Service struct {
	owner       ownerApplication
	incidents   *admission.Checker
	records     *records.Store
	authStore   *authn.Store
	keys        authn.MasterKeys
	cursorCodec *pagination.Codec
	now         func() time.Time
}

func RegisterRoutes(owner ownerApplication) platformhttpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, owner)
		if err != nil {
			return err
		}
		return platformhttpapi.BindOwnerRoutes(mux, deps, "module.indicators", map[string]http.HandlerFunc{
			"listSourceRecordIndicatorObservations": service.handleListSourceObservations,
			"createManualIndicatorObservation":      service.handleCreateObservation,
			"listIndicatorObservations":             service.handleListIndicatorObservations,
			"resolveIndicatorObservation":           service.handleResolveObservation,
			"dismissIndicatorObservation":           service.handleDismissObservation,
			"restoreIndicatorObservation":           service.handleRestoreObservation,
			"listIndicatorStateIntervals":           service.handleListLifecycle,
			"appendIndicatorStateInterval":          service.handleAppendLifecycle,
		})
	}
}

func newService(deps platformhttpapi.DependencySet, owner ownerApplication) (*Service, error) {
	if owner == nil {
		return nil, errors.New("indicator routes: owner is required")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	codec := deps.CursorCodec
	if codec == nil {
		key := authn.DerivePurposeKey(keys, "pagination-cursor-v1")
		codec = pagination.NewCodec(key[:])
	}
	postgres := deps.PostgresHandle()
	return &Service{
		owner: owner, incidents: admission.NewChecker(postgres), records: records.NewStore(postgres),
		authStore: authn.NewStore(postgres), keys: keys, cursorCodec: codec, now: now,
	}, nil
}

func (s *Service) handleListSourceObservations(w http.ResponseWriter, r *http.Request) {
	s.handleObservationList(w, r, true)
}

func (s *Service) handleListIndicatorObservations(w http.ResponseWriter, r *http.Request) {
	s.handleObservationList(w, r, false)
}

func (s *Service) handleObservationList(w http.ResponseWriter, r *http.Request, bySource bool) {
	principal, ok := s.authenticate(w, r, false)
	if !ok {
		return
	}
	pathKey := "indicator_id"
	route := "indicators.observations.by_indicator"
	expectedType := "indicator"
	notFound := indicatorNotFoundError
	if bySource {
		pathKey = "source_record_id"
		route = "indicators.observations.by_source"
		expectedType = ""
		notFound = indicatorSourceNotFoundError
	}
	recordID, valid := pathUUID(r, pathKey)
	if !valid {
		writeAPIError(w, r, notFound())
		return
	}
	envelope, apiErr := s.visibleEnvelope(r.Context(), principal.User.ID, recordID, expectedType, false)
	if apiErr != nil {
		writeAPIError(w, r, notFound())
		return
	}
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(r.URL.Query(), route, principal.User.ID.String(), map[string]string{"record_id": recordID.String()})
	if reason != "" {
		writeAPIError(w, r, invalidPaginationError(reason))
		return
	}
	afterTime, afterID, valid := observationCursorPosition(cursor)
	if !valid {
		writeAPIError(w, r, invalidPaginationError(pagination.ReasonInvalidCursorToken))
		return
	}
	var rows []indicators.IndicatorObservationRecord
	var err error
	if bySource {
		rows, err = s.owner.ListSourceRecordIndicatorObservations(r.Context(), recordID, afterTime, afterID, binding.Limit+1)
	} else {
		rows, err = s.owner.ListIndicatorObservations(r.Context(), recordID, afterTime, afterID, binding.Limit+1)
	}
	if err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	page, next, err := observationPage(s.cursorCodec, binding, rows)
	if err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	if err := s.slide(r.Context(), &principal, r); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = envelope
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"observations": page}, platformhttpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}

func (s *Service) handleCreateObservation(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authenticate(w, r, true)
	if !ok {
		return
	}
	sourceID, valid := pathUUID(r, "source_record_id")
	if !valid {
		writeAPIError(w, r, indicatorSourceNotFoundError())
		return
	}
	envelope, apiErr := s.visibleEnvelope(r.Context(), principal.User.ID, sourceID, "", true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := requireJSON(r); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeObservationCreate(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.owner.CreateIndicatorObservation(r.Context(), principal.User, createParams(envelope.IncidentID, sourceID, request, platformhttpapi.RequestIDFromContext(r.Context())))
	if err != nil {
		writeAPIError(w, r, mutationError(err, request.ClientTxnID))
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	s.writeMutation(w, r, &principal, status, result)
}

func (s *Service) handleResolveObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "resolve")
}

func (s *Service) handleDismissObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "dismiss")
}

func (s *Service) handleRestoreObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "restore")
}

func (s *Service) handleObservationAction(w http.ResponseWriter, r *http.Request, action string) {
	principal, ok := s.authenticate(w, r, true)
	if !ok {
		return
	}
	observationID, valid := pathUUID(r, "observation_id")
	if !valid {
		writeAPIError(w, r, observationNotFoundError())
		return
	}
	observation, err := s.owner.GetIndicatorObservation(r.Context(), observationID)
	if err != nil {
		writeAPIError(w, r, observationNotFoundError())
		return
	}
	if _, apiErr := requireIncidentAdmission(r.Context(), s.incidents, observation.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := requireJSON(r); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	requestID := platformhttpapi.RequestIDFromContext(r.Context())
	var result indicators.IndicatorObservationMutationResult
	var clientTxnID string
	if action == "resolve" {
		request, apiErr := decodeObservationResolve(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		clientTxnID = request.ClientTxnID
		result, err = s.owner.ResolveIndicatorObservation(r.Context(), principal.User, indicators.IndicatorObservationResolveParams{
			ObservationID: observationID, ResolvedIndicatorRecordID: request.ResolvedIndicatorRecordID,
			BaseRowVersion: request.BaseRowVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID, RequestHash: requestHash(request),
		})
	} else {
		request, apiErr := decodeObservationAction(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		clientTxnID = request.ClientTxnID
		params := indicators.IndicatorObservationActionParams{ObservationID: observationID, BaseRowVersion: request.BaseRowVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID, RequestHash: requestHash(request)}
		if action == "dismiss" {
			result, err = s.owner.DismissIndicatorObservation(r.Context(), principal.User, params)
		} else {
			result, err = s.owner.RestoreIndicatorObservation(r.Context(), principal.User, params)
		}
	}
	if err != nil {
		writeAPIError(w, r, mutationError(err, clientTxnID))
		return
	}
	s.writeMutation(w, r, &principal, http.StatusOK, result)
}

func (s *Service) handleListLifecycle(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authenticate(w, r, false)
	if !ok {
		return
	}
	indicatorID, valid := pathUUID(r, "indicator_id")
	if !valid {
		writeAPIError(w, r, indicatorNotFoundError())
		return
	}
	if _, apiErr := s.visibleEnvelope(r.Context(), principal.User.ID, indicatorID, "indicator", false); apiErr != nil {
		writeAPIError(w, r, indicatorNotFoundError())
		return
	}
	binding, cursor, reason := s.cursorCodec.ResolveListRequest(r.URL.Query(), "indicators.lifecycle.by_indicator", principal.User.ID.String(), map[string]string{"record_id": indicatorID.String()})
	if reason != "" {
		writeAPIError(w, r, invalidPaginationError(reason))
		return
	}
	afterTime, afterID, valid := lifecycleCursorPosition(cursor)
	if !valid {
		writeAPIError(w, r, invalidPaginationError(pagination.ReasonInvalidCursorToken))
		return
	}
	rows, err := s.owner.ListIndicatorLifecycleIntervals(r.Context(), indicatorID, afterTime, afterID, binding.Limit+1)
	if err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	page, next, err := lifecyclePage(s.cursorCodec, binding, rows)
	if err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	if err := s.slide(r.Context(), &principal, r); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"intervals": page}, platformhttpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}

func (s *Service) handleAppendLifecycle(w http.ResponseWriter, r *http.Request) {
	principal, ok := s.authenticate(w, r, true)
	if !ok {
		return
	}
	indicatorID, valid := pathUUID(r, "indicator_id")
	if !valid {
		writeAPIError(w, r, indicatorNotFoundError())
		return
	}
	envelope, apiErr := s.visibleEnvelope(r.Context(), principal.User.ID, indicatorID, "indicator", true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if apiErr := requireJSON(r); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeLifecycleAppend(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.owner.AppendIndicatorLifecycleInterval(r.Context(), principal.User, indicators.IndicatorLifecycleAppendParams{
		IncidentID: envelope.IncidentID, IndicatorRecordID: indicatorID, BaseRowVersion: request.BaseRowVersion,
		LifecycleState: request.LifecycleState, ValidFrom: request.ValidFrom, ValidTo: request.ValidTo,
		Confidence: request.Confidence, Rationale: request.Rationale, SupportRefs: request.SupportRefs, Assessor: request.Assessor,
		ClientTxnID: request.ClientTxnID, RequestID: platformhttpapi.RequestIDFromContext(r.Context()), RequestHash: requestHash(request),
	})
	if err != nil {
		writeAPIError(w, r, mutationError(err, request.ClientTxnID))
		return
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	s.writeMutation(w, r, &principal, status, result)
}

func (s *Service) authenticate(w http.ResponseWriter, r *http.Request, changing bool) (httpauth.Principal, bool) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: changing})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return httpauth.Principal{}, false
	}
	return principal, true
}

func (s *Service) visibleEnvelope(ctx context.Context, actorID uuid.UUID, recordID uuid.UUID, expectedType string, mutation bool) (records.Envelope, *platformhttpapi.APIError) {
	envelope, err := s.records.LoadEnvelope(ctx, recordID)
	if err != nil || envelope.DeletedAt != nil || (expectedType != "" && envelope.RecordType != expectedType) {
		if expectedType == "indicator" {
			return records.Envelope{}, indicatorNotFoundError()
		}
		return records.Envelope{}, indicatorSourceNotFoundError()
	}
	if mutation {
		if _, apiErr := requireIncidentAdmission(ctx, s.incidents, envelope.IncidentID, actorID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin"); apiErr != nil {
			return records.Envelope{}, apiErr
		}
	} else if _, apiErr := requireIncidentAdmission(ctx, s.incidents, envelope.IncidentID, actorID, admission.RolesMember, "member"); apiErr != nil {
		return records.Envelope{}, apiErr
	}
	return envelope, nil
}

func (s *Service) writeMutation(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, status int, value any) {
	if err := s.slide(r.Context(), principal, r); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, status, value)
}

func (s *Service) slide(ctx context.Context, principal *httpauth.Principal, r *http.Request) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, r.Method, r.URL.Path, s.now)
}

func requireJSON(r *http.Request) *platformhttpapi.APIError {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(r.Header.Get("Content-Type")))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return invalidMutationPayload("", "invalid_content_type")
	}
	return nil
}

func pathUUID(r *http.Request, key string) (uuid.UUID, bool) {
	return parseCanonicalUUID(r.PathValue(key))
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *platformhttpapi.APIError) {
	platformhttpapi.WriteAPIError(w, r, apiErr)
}

func mutationError(err error, clientTxnID string) *platformhttpapi.APIError {
	switch {
	case errors.Is(err, indicators.ErrInvalidCreateRequest), errors.Is(err, indicators.ErrSourceTextUnavailable):
		return invalidMutationPayload("", "invalid_value")
	case errors.Is(err, indicators.ErrIndicatorSourceNotFound):
		return indicatorSourceNotFoundError()
	case errors.Is(err, indicators.ErrResolvedIndicatorNotFound):
		return resolvedIndicatorNotFoundError()
	case errors.Is(err, indicators.ErrIndicatorObservationNotFound):
		return observationNotFoundError()
	case errors.Is(err, indicators.ErrIndicatorNotFound):
		return indicatorNotFoundError()
	case errors.Is(err, indicators.ErrRowVersionConflict):
		return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: map[string]any{}}
	case errors.Is(err, indicators.ErrIllegalTransition):
		return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Details: map[string]any{}}
	case errors.Is(err, authn.ErrClientTxnConflict):
		return platformhttpapi.ClientTxnConflictError(clientTxnID)
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Details: map[string]any{}}
	default:
		return internalError()
	}
}

func indicatorSourceNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "indicator_source_record_not_found", Details: map[string]any{}}
}

func indicatorNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "indicator_not_found", Details: map[string]any{}}
}

func observationNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "indicator_observation_not_found", Details: map[string]any{}}
}

func resolvedIndicatorNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "resolved_indicator_not_found", Details: map[string]any{}}
}

func requireIncidentAdmission(ctx context.Context, checker *admission.Checker, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string) (admission.Grant, *platformhttpapi.APIError) {
	grant, err := checker.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return admission.Grant{}, &platformhttpapi.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": requiredRole}}
	case err != nil:
		return admission.Grant{}, platformhttpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}

func invalidPaginationError(reason string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_pagination_request", Details: map[string]any{"reason_code": reason}}
}

func internalError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusInternalServerError, Code: "internal_error", Message: "internal_error", Details: map[string]any{}}
}

func observationCursorPosition(cursor *pagination.Cursor) (*time.Time, *uuid.UUID, bool) {
	return keysetPosition(cursor, "created_at", "observation_id")
}

func lifecycleCursorPosition(cursor *pagination.Cursor) (*time.Time, *uuid.UUID, bool) {
	return keysetPosition(cursor, "valid_from", "interval_id")
}

func keysetPosition(cursor *pagination.Cursor, timeKey string, idKey string) (*time.Time, *uuid.UUID, bool) {
	if cursor == nil {
		return nil, nil, true
	}
	if cursor.Mode != pagination.ModeKeyset || len(cursor.Position) != 2 {
		return nil, nil, false
	}
	instant, err := time.Parse(time.RFC3339Nano, cursor.Position[timeKey])
	identifier, valid := parseCanonicalUUID(cursor.Position[idKey])
	if err != nil || !valid {
		return nil, nil, false
	}
	instant = instant.UTC()
	return &instant, &identifier, true
}

func observationPage(codec *pagination.Codec, binding pagination.Binding, rows []indicators.IndicatorObservationRecord) ([]indicators.IndicatorObservationRecord, *string, error) {
	if len(rows) <= binding.Limit {
		return rows, nil, nil
	}
	page := rows[:binding.Limit]
	last := page[len(page)-1]
	next, err := encodeKeyset(codec, binding, map[string]string{"created_at": last.CreatedAt.UTC().Format(time.RFC3339Nano), "observation_id": last.ObservationID.String()})
	return page, next, err
}

func lifecyclePage(codec *pagination.Codec, binding pagination.Binding, rows []indicators.IndicatorLifecycleIntervalRecord) ([]indicators.IndicatorLifecycleIntervalRecord, *string, error) {
	if len(rows) <= binding.Limit {
		return rows, nil, nil
	}
	page := rows[:binding.Limit]
	last := page[len(page)-1]
	next, err := encodeKeyset(codec, binding, map[string]string{"valid_from": last.ValidFrom.UTC().Format(time.RFC3339Nano), "interval_id": last.IntervalID.String()})
	return page, next, err
}

func encodeKeyset(codec *pagination.Codec, binding pagination.Binding, position map[string]string) (*string, error) {
	token, err := codec.Encode(pagination.Cursor{Mode: pagination.ModeKeyset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: binding.Limit, Scope: binding.Scope, Position: position})
	if err != nil {
		return nil, fmt.Errorf("encode Indicator cursor: %w", err)
	}
	return &token, nil
}
