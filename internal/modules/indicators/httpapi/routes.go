package httpapi

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	recordenvelope "github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

type ownerApplication interface {
	CreateIndicatorObservation(context.Context, uuid.UUID, indicators.IndicatorObservationCreateParams) (indicators.IndicatorObservationMutationResult, error)
	ResolveIndicatorObservation(context.Context, uuid.UUID, indicators.IndicatorObservationResolveParams) (indicators.IndicatorObservationMutationResult, error)
	DismissIndicatorObservation(context.Context, uuid.UUID, indicators.IndicatorObservationActionParams) (indicators.IndicatorObservationMutationResult, error)
	RestoreIndicatorObservation(context.Context, uuid.UUID, indicators.IndicatorObservationActionParams) (indicators.IndicatorObservationMutationResult, error)
	AppendIndicatorLifecycleInterval(context.Context, uuid.UUID, indicators.IndicatorLifecycleAppendParams) (indicators.IndicatorLifecycleMutationResult, error)
	GetIndicatorObservation(context.Context, uuid.UUID) (indicators.IndicatorObservationRecord, error)
	ListSourceRecordIndicatorObservations(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorObservationRecord, error)
	ListIndicatorObservations(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorObservationRecord, error)
	ListIndicatorLifecycleIntervals(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]indicators.IndicatorLifecycleIntervalRecord, error)
}

type recordEnvelopeReader interface {
	LoadEnvelope(context.Context, uuid.UUID) (recordenvelope.Envelope, error)
}

type incidentAdmission interface {
	Check(context.Context, uuid.UUID, uuid.UUID, admission.Requirement) (admission.Grant, error)
}

// sessionSlider is the session surface Indicator transport slides against.
// Central admission authenticates with the application's authenticator.
type sessionSlider interface {
	httpauth.SessionSlider
}

type service struct {
	owner       ownerApplication
	incidents   incidentAdmission
	records     recordEnvelopeReader
	sessions    sessionSlider
	cursorCodec *pagination.Codec
	now         func() time.Time
}

// RegisterRoutes binds the Indicator routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits the path identifier before a handler runs.
func RegisterRoutes(owner ownerApplication, recordEnvelopes recordEnvelopeReader, incidents incidentAdmission, sessions sessionSlider) platformhttpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, owner, recordEnvelopes, incidents, sessions)
		if err != nil {
			return err
		}
		return platformhttpapi.BindAdmittedOwnerRoutes(mux, deps, "module.indicators", service.operationHandlers())
	}
}

// operationHandlers maps each Indicator operation to its one handler.
func (s *service) operationHandlers() map[string]http.HandlerFunc {
	return map[string]http.HandlerFunc{
		"listSourceRecordIndicatorObservations": s.handleListSourceObservations,
		"createManualIndicatorObservation":      s.handleCreateObservation,
		"listIndicatorObservations":             s.handleListIndicatorObservations,
		"resolveIndicatorObservation":           s.handleResolveObservation,
		"dismissIndicatorObservation":           s.handleDismissObservation,
		"restoreIndicatorObservation":           s.handleRestoreObservation,
		"listIndicatorStateIntervals":           s.handleListLifecycle,
		"appendIndicatorStateInterval":          s.handleAppendLifecycle,
	}
}

func newService(deps platformhttpapi.DependencySet, owner ownerApplication, recordEnvelopes recordEnvelopeReader, incidents incidentAdmission, sessions sessionSlider) (*service, error) {
	if nilHTTPDependency(owner) {
		return nil, errors.New("indicator routes: owner is required")
	}
	if nilHTTPDependency(recordEnvelopes) {
		return nil, errors.New("indicator routes: RecordEnvelopes is required")
	}
	if nilHTTPDependency(incidents) {
		return nil, errors.New("indicator routes: IncidentAdmission is required")
	}
	if nilHTTPDependency(sessions) {
		return nil, errors.New("indicator routes: Sessions is required")
	}
	if deps.Now == nil {
		return nil, errors.New("indicator routes: Now is required")
	}
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	codec := deps.CursorCodec
	if codec == nil {
		key, err := authn.DerivePurposeKey(keys, "pagination-cursor-v2")
		if err != nil {
			return nil, err
		}
		codec = pagination.NewCodec(key)
	}
	return &service{
		owner: owner, incidents: incidents, records: recordEnvelopes,
		sessions: sessions, cursorCodec: codec, now: deps.Now,
	}, nil
}

func nilHTTPDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (s *service) handleListSourceObservations(w http.ResponseWriter, r *http.Request) {
	s.handleObservationList(w, r, true)
}

func (s *service) handleListIndicatorObservations(w http.ResponseWriter, r *http.Request) {
	s.handleObservationList(w, r, false)
}

func (s *service) handleObservationList(w http.ResponseWriter, r *http.Request, bySource bool) {
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
	principal, recordID, apiErr := admitted(r, pathKey)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, apiErr := s.visibleReadEnvelope(r.Context(), principal.User.ID, recordID, expectedType, notFound)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	if err := s.slide(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = envelope
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"observations": page}, platformhttpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}

func (s *service) handleCreateObservation(w http.ResponseWriter, r *http.Request) {
	principal, sourceID, apiErr := admitted(r, "source_record_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, apiErr := s.admitMutationTarget(r.Context(), principal.User.ID, sourceID, indicatorSourceNotFoundError)
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
	result, err := s.owner.CreateIndicatorObservation(r.Context(), principal.User.ID, createParams(envelope.IncidentID, sourceID, request, platformhttpapi.RequestIDFromContext(r.Context())))
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

func (s *service) handleResolveObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "resolve")
}

func (s *service) handleDismissObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "dismiss")
}

func (s *service) handleRestoreObservation(w http.ResponseWriter, r *http.Request) {
	s.handleObservationAction(w, r, "restore")
}

func (s *service) handleObservationAction(w http.ResponseWriter, r *http.Request, action string) {
	principal, observationID, apiErr := admitted(r, "observation_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// Only the typed missing observation is not-found; any other lookup
	// failure is internal_error (REQ-01-706).
	observation, err := s.owner.GetIndicatorObservation(r.Context(), observationID)
	if errors.Is(err, indicators.ErrIndicatorObservationNotFound) {
		writeAPIError(w, r, observationNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	if _, apiErr := requireIncidentAdmission(r.Context(), s.incidents, observation.IncidentID, principal.User.ID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin", observationNotFoundError); apiErr != nil {
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
		result, err = s.owner.ResolveIndicatorObservation(r.Context(), principal.User.ID, indicators.IndicatorObservationResolveParams{
			ObservationID: observationID, ResolvedIndicatorRecordID: request.ResolvedIndicatorRecordID,
			BaseRowVersion: request.BaseRowVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID,
		})
	} else {
		request, apiErr := decodeObservationAction(r.Body)
		if apiErr != nil {
			writeAPIError(w, r, apiErr)
			return
		}
		clientTxnID = request.ClientTxnID
		params := indicators.IndicatorObservationActionParams{ObservationID: observationID, BaseRowVersion: request.BaseRowVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID}
		if action == "dismiss" {
			result, err = s.owner.DismissIndicatorObservation(r.Context(), principal.User.ID, params)
		} else {
			result, err = s.owner.RestoreIndicatorObservation(r.Context(), principal.User.ID, params)
		}
	}
	if err != nil {
		writeAPIError(w, r, mutationError(err, clientTxnID))
		return
	}
	s.writeMutation(w, r, &principal, http.StatusOK, result)
}

func (s *service) handleListLifecycle(w http.ResponseWriter, r *http.Request) {
	principal, indicatorID, apiErr := admitted(r, "indicator_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.visibleReadEnvelope(r.Context(), principal.User.ID, indicatorID, "indicator", indicatorNotFoundError); apiErr != nil {
		writeAPIError(w, r, apiErr)
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
	if err := s.slide(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{"intervals": page}, platformhttpapi.PagingMeta{Limit: binding.Limit, HasMore: next != nil, NextCursor: next})
}

func (s *service) handleAppendLifecycle(w http.ResponseWriter, r *http.Request) {
	principal, indicatorID, apiErr := admitted(r, "indicator_id")
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	envelope, apiErr := s.admitMutationTarget(r.Context(), principal.User.ID, indicatorID, indicatorNotFoundError)
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
	result, err := s.owner.AppendIndicatorLifecycleInterval(r.Context(), principal.User.ID, indicators.IndicatorLifecycleAppendParams{
		IncidentID: envelope.IncidentID, IndicatorRecordID: indicatorID, BaseRowVersion: request.BaseRowVersion,
		LifecycleState: request.LifecycleState, ValidFrom: request.ValidFrom, ValidTo: request.ValidTo,
		Confidence: request.Confidence, Rationale: request.Rationale, SupportRefs: request.SupportRefs, Assessor: request.Assessor,
		ClientTxnID: request.ClientTxnID, RequestID: platformhttpapi.RequestIDFromContext(r.Context()),
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

// admitted returns the principal and the uuid path parameter central
// admission established (REQ-01-707 steps 2 and 3); their absence is a
// binding error.
func admitted(r *http.Request, name string) (httpauth.Principal, uuid.UUID, *platformhttpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	value, identified := platformhttpapi.PathUUID(r, name)
	if !authenticated || !identified {
		return httpauth.Principal{}, uuid.UUID{}, internalError()
	}
	return principal, value, nil
}

// visibleReadEnvelope admits the record a read addresses by path. A missing,
// soft-deleted or wrongly typed record, and a record in an incident the caller
// cannot see, all receive the route's not-found response (REQ-01-704). Any
// other envelope read failure is internal_error (REQ-01-706).
func (s *service) visibleReadEnvelope(ctx context.Context, actorID uuid.UUID, recordID uuid.UUID, expectedType string, notFound func() *platformhttpapi.APIError) (recordenvelope.Envelope, *platformhttpapi.APIError) {
	envelope, err := s.records.LoadEnvelope(ctx, recordID)
	if err != nil && !errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return recordenvelope.Envelope{}, internalError()
	}
	if records.ClassifyTargetPresence(err == nil, envelope.DeletedAt) != nil || (expectedType != "" && envelope.RecordType != expectedType) {
		return recordenvelope.Envelope{}, notFound()
	}
	if _, apiErr := requireIncidentAdmission(ctx, s.incidents, envelope.IncidentID, actorID, admission.RolesMember, "member", notFound); apiErr != nil {
		return recordenvelope.Envelope{}, apiErr
	}
	return envelope, nil
}

// admitMutationTarget admits the record a child mutation addresses by path
// (REQ-01-704, REQ-01-705): a missing record is the route's not-found
// response, a caller who cannot see the record's incident receives that same
// response, and a visible caller without the route's role is denied. The owner
// then decides an exact committed replay first and, under the record lock, the
// deleted state, the record type, the incident lifecycle and the base version,
// in that order. Any other envelope read failure is internal_error
// (REQ-01-706).
func (s *service) admitMutationTarget(ctx context.Context, actorID uuid.UUID, recordID uuid.UUID, notFound func() *platformhttpapi.APIError) (recordenvelope.Envelope, *platformhttpapi.APIError) {
	envelope, err := s.records.LoadEnvelope(ctx, recordID)
	if errors.Is(err, recordenvelope.ErrEnvelopeNotFound) {
		return recordenvelope.Envelope{}, notFound()
	}
	if err != nil {
		return recordenvelope.Envelope{}, internalError()
	}
	if _, apiErr := requireIncidentAdmission(ctx, s.incidents, envelope.IncidentID, actorID, admission.RolesEditorReviewerAdmin, "editor|reviewer|admin", notFound); apiErr != nil {
		return recordenvelope.Envelope{}, apiErr
	}
	return envelope, nil
}

func (s *service) writeMutation(w http.ResponseWriter, r *http.Request, principal *httpauth.Principal, status int, value any) {
	if err := s.slide(r.Context(), principal); err != nil {
		writeAPIError(w, r, internalError())
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, status, value)
}

func (s *service) slide(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.sessions, principal, s.now)
}

func requireJSON(r *http.Request) *platformhttpapi.APIError {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(r.Header.Get("Content-Type")))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return invalidMutationPayload("", "invalid_content_type")
	}
	return nil
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *platformhttpapi.APIError) {
	platformhttpapi.WriteAPIError(w, r, apiErr)
}

func mutationError(err error, clientTxnID string) *platformhttpapi.APIError {
	var transition *indicators.IllegalTransitionError
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
	case errors.Is(err, records.ErrTargetDeleted):
		return recordDeletedUseRestoreError()
	case errors.Is(err, indicators.ErrRowVersionConflict):
		return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "row_version_conflict", Details: map[string]any{}}
	case errors.As(err, &transition):
		// Core 01 §3.3.6.2: observation resolution is matrix-only.
		return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "illegal_transition", Details: map[string]any{
			"from_status": transition.FromStatus, "to_status": transition.ToStatus,
			"violated_guards": []string{}, "reason_code": "transition_not_allowed",
		}}
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

func recordDeletedUseRestoreError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "record_deleted_use_restore", Details: map[string]any{}}
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

// requireIncidentAdmission decides the caller's incident visibility and role.
// A caller who cannot see the incident receives the route's own not-found
// response, identical to the response for a missing resource (REQ-01-704).
func requireIncidentAdmission(ctx context.Context, checker incidentAdmission, incidentID uuid.UUID, userID uuid.UUID, roles admission.RoleSet, requiredRole string, notFound func() *platformhttpapi.APIError) (admission.Grant, *platformhttpapi.APIError) {
	grant, err := checker.Check(ctx, incidentID, userID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, notFound()
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
