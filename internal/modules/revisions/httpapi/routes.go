package httpapi

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformhttpapi "github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type commandApplication interface {
	GetHistory(context.Context, revisions.HistoryQuery) (revisions.HistoryResult, error)
	RollbackRecord(context.Context, revisions.RollbackCommand) (revisions.RollbackResult, error)
	SoftDeleteRecord(context.Context, revisions.DeleteRestoreCommand) (revisions.DeleteRestoreResult, error)
	RestoreRecord(context.Context, revisions.DeleteRestoreCommand) (revisions.DeleteRestoreResult, error)
}

type Service struct {
	commands       commandApplication
	incidentAccess *admission.Checker
	records        revisions.RecordEnvelopeReader
	authStore      *authn.Store
	keys           authn.MasterKeys
	cursorCodec    *pagination.Codec
	now            func() time.Time
}

func RegisterRoutes(commands commandApplication, records revisions.RecordEnvelopeReader) platformhttpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps platformhttpapi.DependencySet) error {
		service, err := newService(deps, commands, records)
		if err != nil {
			return err
		}
		return platformhttpapi.BindOwnerRoutes(mux, deps, "module.revisions", map[string]http.HandlerFunc{
			"deleteRecord":     service.handleRecordDelete,
			"getRecordHistory": service.handleRecordHistory,
			"restoreRecord":    service.handleRecordRestore,
			"rollbackRecord":   service.handleRecordRollback,
		})
	}
}

func newService(deps platformhttpapi.DependencySet, commands commandApplication, records revisions.RecordEnvelopeReader) (*Service, error) {
	if commands == nil {
		return nil, errors.New("revisions routes: command service is required")
	}
	if records == nil {
		return nil, errors.New("revisions routes: record envelope reader is required")
	}
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
	return &Service{
		commands:       commands,
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		records:        records,
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		cursorCodec:    cursorCodec,
		now:            now,
	}, nil
}

func (s *Service) handleRecordHistory(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}

	envelope, err := s.records.LoadEnvelope(r.Context(), recordID)
	if errors.Is(err, revisions.ErrEnvelopeNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), envelope.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	history, err := s.commands.GetHistory(r.Context(), revisions.HistoryQuery{RecordID: recordID})
	if errors.Is(err, revisions.ErrRecordNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	record := history.Record

	binding, cursor, reasonCode := s.cursorCodec.ResolveRequest(
		r.URL.Query(),
		"records.history",
		principal.User.ID.String(),
		map[string]string{"record_id": recordID.String()},
	)
	if reasonCode != "" {
		writeAPIError(w, r, invalidPaginationRequest(reasonCode))
		return
	}

	rows, nextCursor, err := pagination.PageResources(binding, cursor, history.Resources)
	switch {
	case errors.Is(err, pagination.ErrInvalidCursorToken):
		writeAPIError(w, r, invalidPaginationRequest(pagination.ReasonInvalidCursorToken))
		return
	case err != nil:
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
	_ = platformhttpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
		"incident_id": record.IncidentID.String(),
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"deleted":     record.Deleted,
		"items":       rows,
	}, platformhttpapi.PagingMeta{
		Limit:      binding.Limit,
		HasMore:    nextToken != nil,
		NextCursor: nextToken,
	})
}

func (s *Service) handleRecordDelete(w http.ResponseWriter, r *http.Request) {
	s.handleDeleteRestore(w, r, true)
}

func (s *Service) handleRecordRestore(w http.ResponseWriter, r *http.Request) {
	s.handleDeleteRestore(w, r, false)
}

func (s *Service) handleRecordRollback(w http.ResponseWriter, r *http.Request) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	record, err := s.records.LoadEnvelope(r.Context(), recordID)
	if errors.Is(err, revisions.ErrEnvelopeNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), record.IncidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if !roleIn(membership.Role.String(), "reviewer", "admin") {
		writeAPIError(w, r, forbiddenError("reviewer|admin"))
		return
	}
	if apiErr := requireJSONContentType(r, invalidRollbackRequest); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeRollbackRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.commands.RollbackRecord(r.Context(), revisions.RollbackCommand{
		Actor:       revisions.NewActorID(principal.User.ID),
		RecordID:    recordID,
		Request:     request,
		RequestHash: revisions.RollbackRequestHash(request),
		RequestID:   platformhttpapi.RequestIDFromContext(r.Context()),
	})
	switch {
	case errors.Is(err, revisions.ErrRecordNotFound), errors.Is(err, revisions.ErrCommandTargetNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, revisions.ErrCommandIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, revisions.ErrCommandRoleDenied):
		writeAPIError(w, r, forbiddenError("reviewer|admin"))
		return
	case errors.Is(err, revisions.ErrRecordDeletedUseRestore):
		writeAPIError(w, r, recordDeletedUseRestoreError())
		return
	case errors.Is(err, revisions.ErrRowVersionConflict):
		var conflict *revisions.RowVersionConflictError
		if errors.As(err, &conflict) {
			writeAPIError(w, r, rowVersionConflictError(conflict.Details()))
			return
		}
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.Is(err, revisions.ErrRecordLocked):
		var locked *revisions.RecordLockedError
		recordIDText := recordID.String()
		if errors.As(err, &locked) {
			recordIDText = locked.RecordID.String()
		}
		writeAPIError(w, r, recordLockedError(recordIDText))
		return
	case errors.Is(err, revisions.ErrRollbackTargetNotFound):
		writeAPIError(w, r, rollbackTargetNotFoundError(request.Target))
		return
	case errors.Is(err, revisions.ErrRollbackPreconditionFailed):
		var precondition *revisions.RollbackPreconditionError
		if errors.As(err, &precondition) {
			writeAPIError(w, r, rollbackPreconditionFailedError(precondition.ReasonCode))
			return
		}
		writeAPIError(w, r, rollbackPreconditionFailedError("target_not_reversible"))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleDeleteRestore(w http.ResponseWriter, r *http.Request, deleting bool) {
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	record, err := s.records.LoadEnvelope(r.Context(), recordID)
	if errors.Is(err, revisions.ErrEnvelopeNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	membership, apiErr := s.requireIncidentMembership(r.Context(), record.IncidentID, principal.User.ID)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if deleting {
		if !roleIn(membership.Role.String(), "editor", "reviewer", "admin") {
			writeAPIError(w, r, forbiddenError("editor|reviewer|admin"))
			return
		}
	} else if !roleIn(membership.Role.String(), "reviewer", "admin") {
		writeAPIError(w, r, forbiddenError("reviewer|admin"))
		return
	}
	if apiErr := requireJSONContentType(r, invalidMutationPayload); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := decodeDeleteRestoreRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	command := revisions.DeleteRestoreCommand{
		Actor:       revisions.NewActorID(principal.User.ID),
		RecordID:    recordID,
		Request:     request,
		RequestHash: revisions.DeleteRestoreRequestHash(request),
		RequestID:   platformhttpapi.RequestIDFromContext(r.Context()),
	}
	var result revisions.DeleteRestoreResult
	if deleting {
		result, err = s.commands.SoftDeleteRecord(r.Context(), command)
	} else {
		result, err = s.commands.RestoreRecord(r.Context(), command)
	}
	switch {
	case errors.Is(err, revisions.ErrRecordNotFound), errors.Is(err, revisions.ErrCommandTargetNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, revisions.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, revisions.ErrCommandIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, revisions.ErrCommandRoleDenied):
		requiredRole := "reviewer|admin"
		if deleting {
			requiredRole = "editor|reviewer|admin"
		}
		writeAPIError(w, r, forbiddenError(requiredRole))
		return
	case errors.Is(err, revisions.ErrRowVersionConflict):
		var conflict *revisions.RowVersionConflictError
		if errors.As(err, &conflict) {
			writeAPIError(w, r, rowVersionConflictError(conflict.Details()))
			return
		}
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.Is(err, revisions.ErrRecordAlreadyDeleted):
		writeAPIError(w, r, recordAlreadyDeletedError())
		return
	case errors.Is(err, revisions.ErrRecordDeleteBlocked):
		var blocked *revisions.RecordDeleteBlockedError
		if errors.As(err, &blocked) {
			writeAPIError(w, r, recordDeleteBlockedError(blocked.Details()))
			return
		}
		writeAPIError(w, r, recordDeleteBlockedError(nil))
		return
	case errors.Is(err, revisions.ErrRecordRestoreBlocked):
		var blocked *revisions.RecordRestoreBlockedError
		if errors.As(err, &blocked) {
			writeAPIError(w, r, recordRestoreBlockedError(blocked.Details()))
			return
		}
		writeAPIError(w, r, recordRestoreBlockedError(nil))
		return
	case errors.Is(err, revisions.ErrRecordNotDeleted):
		writeAPIError(w, r, recordNotDeletedError())
		return
	case errors.Is(err, revisions.ErrRecordLocked):
		var locked *revisions.RecordLockedError
		recordIDText := recordID.String()
		if errors.As(err, &locked) {
			recordIDText = locked.RecordID.String()
		}
		writeAPIError(w, r, recordLockedError(recordIDText))
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = platformhttpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *platformhttpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, incidentNotFoundError()
	case err != nil:
		return admission.Grant{}, internalAPIError(err)
	default:
		return grant, nil
	}
}

func roleIn(role string, allowed ...string) bool {
	for _, candidate := range allowed {
		if role == candidate {
			return true
		}
	}
	return false
}

func requireJSONContentType(r *http.Request, failure func(string, string) *platformhttpapi.APIError) *platformhttpapi.APIError {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(r.Header.Get("Content-Type")))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return failure("", "invalid_content_type")
	}
	return nil
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *platformhttpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = platformhttpapi.WriteErrorWithConflict(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details, apiErr.Conflict)
}

func pathUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	raw := r.PathValue(key)
	value, err := uuid.Parse(raw)
	if err != nil {
		writeAPIError(w, r, incidentNotFoundError())
		return uuid.UUID{}, false
	}
	return value, true
}

func invalidPaginationRequest(reasonCode string) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func incidentClosedError() *platformhttpapi.APIError {
	return &platformhttpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func internalAPIError(err error) *platformhttpapi.APIError {
	return &platformhttpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
