package revisions

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

type Service struct {
	store          *Store
	incidentAccess incidents.Access
	authStore      *authn.Store
	keys           authn.MasterKeys
	publisher      *collaboration.RecordChangePublisher
	cursorCodec    *pagination.Codec
	now            func() time.Time
}

type RouteOptions struct {
	ImportedAttributionResolver ImportedAttributionResolver
}

type RouteOption func(*RouteOptions)

func WithImportedAttributionResolver(resolver ImportedAttributionResolver) RouteOption {
	return func(options *RouteOptions) {
		options.ImportedAttributionResolver = resolver
	}
}

func RegisterRoutes(options ...RouteOption) httpapi.RouteRegistrar {
	routeOptions := RouteOptions{}
	for _, option := range options {
		if option != nil {
			option(&routeOptions)
		}
	}
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, routeOptions)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /api/v1/records/{record_id}/history", service.handleRecordHistory)
		mux.HandleFunc("DELETE /api/v1/records/{record_id}", service.handleRecordDelete)
		mux.HandleFunc("POST /api/v1/records/{record_id}/restore", service.handleRecordRestore)
		mux.HandleFunc("POST /api/v1/records/{record_id}/rollback", service.handleRecordRollback)
		return nil
	}
}

func newService(deps httpapi.DependencySet, options RouteOptions) (*Service, error) {
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
		store: NewStoreWithOptions(deps.PostgresHandle(), StoreOptions{
			ImportedAttributionResolver: options.ImportedAttributionResolver,
		}),
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		publisher:      collaboration.NewRecordChangePublisher(deps.WSHub),
		cursorCodec:    cursorCodec,
		now:            now,
	}, nil
}

func (s *Service) handleRecordHistory(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	record, err := s.store.GetHistoryRecord(r.Context(), recordID)
	if errors.Is(err, ErrRecordNotFound) {
		writeAPIError(w, r, incidentNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), record.IncidentID, principal.User.ID); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

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

	resources, err := s.store.ListRecordHistory(r.Context(), record)
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	rows, nextCursor, err := pagination.PageResources(binding, cursor, resources)
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
	_ = httpapi.WriteSuccessWithPaging(w, r, http.StatusOK, map[string]any{
		"incident_id": record.IncidentID.String(),
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"deleted":     record.Deleted,
		"items":       rows,
	}, httpapi.PagingMeta{
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
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeRollbackRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.GetHistoryRecord(r.Context(), recordID)
	if errors.Is(err, ErrRecordNotFound) {
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
	if !roleIn(membership.Role, "reviewer", "admin") {
		writeAPIError(w, r, forbiddenError("reviewer|admin"))
		return
	}

	result, err := s.store.RollbackRecord(r.Context(), principal.User, recordID, request, RollbackRequestHash(request), httpapi.RequestIDFromContext(r.Context()), s.now())
	switch {
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, ErrRecordDeletedUseRestore):
		writeAPIError(w, r, recordDeletedUseRestoreError())
		return
	case errors.Is(err, ErrRowVersionConflict):
		var conflict *RowVersionConflictError
		if errors.As(err, &conflict) {
			writeAPIError(w, r, rowVersionConflictError(conflict.Details()))
			return
		}
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.Is(err, ErrRecordLocked):
		var locked *RecordLockedError
		recordIDText := recordID.String()
		if errors.As(err, &locked) {
			recordIDText = locked.RecordID.String()
		}
		writeAPIError(w, r, recordLockedError(recordIDText))
		return
	case errors.Is(err, ErrRollbackTargetNotFound):
		writeAPIError(w, r, rollbackTargetNotFoundError(request.Target))
		return
	case errors.Is(err, ErrRollbackPreconditionFailed):
		var precondition *RollbackPreconditionError
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
	if !result.Replayed {
		s.publishRollbackChanges(result, principal.User.ID)
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleDeleteRestore(w http.ResponseWriter, r *http.Request, deleting bool) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeDeleteRestoreRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	record, err := s.store.GetHistoryRecord(r.Context(), recordID)
	if errors.Is(err, ErrRecordNotFound) {
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
		if !roleIn(membership.Role, "editor", "reviewer", "admin") {
			writeAPIError(w, r, forbiddenError("editor|reviewer|admin"))
			return
		}
	} else if !roleIn(membership.Role, "reviewer", "admin") {
		writeAPIError(w, r, forbiddenError("reviewer|admin"))
		return
	}

	requestHash := DeleteRestoreRequestHash(request)
	var result DeleteRestoreResult
	if deleting {
		result, err = s.store.SoftDeleteRecord(r.Context(), principal.User, recordID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	} else {
		result, err = s.store.RestoreRecord(r.Context(), principal.User, recordID, request, requestHash, httpapi.RequestIDFromContext(r.Context()), s.now())
	}
	switch {
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case errors.Is(err, authn.ErrClientTxnConflict):
		writeAPIError(w, r, clientTxnConflictError(request.ClientTxnID))
		return
	case errors.Is(err, incidents.ErrIncidentClosed):
		writeAPIError(w, r, incidentClosedError())
		return
	case errors.Is(err, ErrRowVersionConflict):
		var conflict *RowVersionConflictError
		if errors.As(err, &conflict) {
			writeAPIError(w, r, rowVersionConflictError(conflict.Details()))
			return
		}
		writeAPIError(w, r, rowVersionConflictError(map[string]any{}))
		return
	case errors.Is(err, ErrRecordAlreadyDeleted):
		writeAPIError(w, r, recordAlreadyDeletedError())
		return
	case errors.Is(err, ErrRecordDeleteBlocked):
		var blocked *RecordDeleteBlockedError
		if errors.As(err, &blocked) {
			writeAPIError(w, r, recordDeleteBlockedError(blocked.Details()))
			return
		}
		writeAPIError(w, r, recordDeleteBlockedError(nil))
		return
	case errors.Is(err, ErrRecordNotDeleted):
		writeAPIError(w, r, recordNotDeletedError())
		return
	case errors.Is(err, ErrRecordLocked):
		var locked *RecordLockedError
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
	if !result.Replayed {
		s.publishDeleteRestoreChange(result, principal.User.ID)
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) publishDeleteRestoreChange(result DeleteRestoreResult, actorUserID uuid.UUID) {
	if result.ViewSchemaID == "" {
		return
	}
	s.publisher.Publish(collaboration.RecordChange{
		IncidentID:       result.IncidentID,
		RecordID:         result.RecordID,
		RowVersion:       result.RowVersion,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: []string{},
		ViewSchemaID:     result.ViewSchemaID,
		ChangeKind:       result.ChangeKind,
	})
}

func (s *Service) publishRollbackChanges(result RollbackResult, actorUserID uuid.UUID) {
	for _, change := range result.Changes {
		if change.ViewSchemaID == "" {
			continue
		}
		s.publisher.Publish(collaboration.RecordChange{
			IncidentID:       result.IncidentID,
			RecordID:         change.RecordID,
			RowVersion:       change.RowVersion,
			ChangeSetID:      change.ChangeSetID,
			ClientTxnID:      result.ClientTxnID,
			ActorUserID:      actorUserID,
			ChangedFieldKeys: append([]string(nil), change.ChangedFieldKeys...),
			ViewSchemaID:     change.ViewSchemaID,
			ChangeKind:       "invalidate",
		})
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

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal, method string, path string) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, method, path, s.now)
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

func invalidPaginationRequest(reasonCode string) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_pagination_request",
		Message: "invalid pagination request",
		Details: map[string]any{
			"reason_code": reasonCode,
		},
	}
}

func incidentNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
}

func incidentClosedError() *httpapi.APIError {
	return &httpapi.APIError{Status: http.StatusConflict, Code: "incident_closed", Message: "incident closed", Details: map[string]any{}}
}

func internalAPIError(err error) *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
