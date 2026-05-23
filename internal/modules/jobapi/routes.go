package jobapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type Service struct {
	manager       *jobs.Manager
	authStore     *authn.Store
	incidentStore *incidents.Store
	hub           *platformws.Hub
	keys          authn.MasterKeys
	now           func() time.Time
}

type cancelRequest struct {
	ClientTxnID string
	Reason      *string
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/jobs/", service.handleJobsMember)
		return nil
	}
}

func newService(deps httpapi.DependencySet) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		manager:       deps.Jobs,
		authStore:     authn.NewStore(deps.Postgres),
		incidentStore: incidents.NewStore(deps.Postgres),
		hub:           deps.WSHub,
		keys:          keys,
		now:           now,
	}, nil
}

func (s *Service) handleJobsMember(w http.ResponseWriter, r *http.Request) {
	jobID, action, ok := parseJobPath(r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch {
	case r.Method == http.MethodGet && action == "":
		s.handleGetJob(w, r, jobID)
	case r.Method == http.MethodPost && action == "cancel":
		s.handleCancelJob(w, r, jobID)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Service) handleGetJob(w http.ResponseWriter, r *http.Request, jobID uuid.UUID) {
	if apiErr := auth.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	principal, apiErr := s.authenticateSessionRequest(r, false)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, err := s.manager.Get(r.Context(), jobID)
	if errors.Is(err, jobs.ErrNotFound) {
		writeAPIError(w, r, jobNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if apiErr := s.authorizeJob(r.Context(), resource, principal, false); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleCancelJob(w http.ResponseWriter, r *http.Request, jobID uuid.UUID) {
	principal, apiErr := s.authenticateSessionRequest(r, true)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, normalized, apiErr := decodeCancelRequest(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, err := s.manager.Get(r.Context(), jobID)
	if errors.Is(err, jobs.ErrNotFound) {
		writeAPIError(w, r, jobNotFoundError())
		return
	}
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if apiErr := s.authorizeJob(r.Context(), resource, principal, true); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.manager.Cancel(r.Context(), jobs.CancelParams{
		JobID:             jobID,
		ActorUserID:       principal.User.ID,
		ClientTxnID:       request.ClientTxnID,
		NormalizedRequest: normalized,
	})
	switch {
	case errors.Is(err, jobs.ErrClientTxnConflict):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": request.ClientTxnID}})
		return
	case errors.Is(err, jobs.ErrCancelRejected):
		writeAPIError(w, r, &auth.APIError{Status: http.StatusConflict, Code: "job_cancel_rejected", Details: map[string]any{"reason_code": result.ReasonCode}})
		return
	case errors.Is(err, jobs.ErrNotFound):
		writeAPIError(w, r, jobNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Resource)
}

func (s *Service) authorizeJob(ctx context.Context, resource jobs.Resource, principal auth.SessionPrincipal, cancel bool) *auth.APIError {
	submittedByCurrentUser := resource.SubmittedByUserID == principal.User.ID.String()
	switch resource.Scope.Kind {
	case jobs.ScopeKindDeployment:
		if submittedByCurrentUser || principal.User.IsDeploymentAdmin {
			return nil
		}
		return jobNotFoundError()
	case jobs.ScopeKindIncident:
		if resource.Scope.IncidentID == nil {
			return jobNotFoundError()
		}
		membership, err := s.incidentStore.GetIncidentMembershipForUser(ctx, *resource.Scope.IncidentID, principal.User.ID)
		if errors.Is(err, incidents.ErrMembershipNotFound) {
			return jobNotFoundError()
		}
		if err != nil {
			return internalAPIError(err)
		}
		if !cancel || submittedByCurrentUser || membership.Role == "admin" {
			return nil
		}
		return &auth.APIError{Status: http.StatusForbidden, Code: "authorization_denied", Details: map[string]any{"required_role": "submitted_by|admin"}}
	default:
		return jobNotFoundError()
	}
}

func (s *Service) authenticateSessionRequest(r *http.Request, stateChanging bool) (auth.SessionPrincipal, *auth.APIError) {
	return auth.AuthenticateSessionRequest(r, auth.SessionAuthOptions{
		Store:         s.authStore,
		Keys:          s.keys,
		Hub:           s.hub,
		Now:           s.now,
		StateChanging: stateChanging,
	})
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *auth.SessionPrincipal, method string, path string) error {
	if !auth.ShouldSlideIdleExpiry(method, path) {
		return nil
	}
	sliding := authn.SessionTiming{
		AuthenticatedAt:          principal.Session.AuthenticatedAt,
		LastQualifyingActivityAt: principal.Session.LastQualifyingActivityAt,
		IdleExpiresAt:            principal.Session.IdleExpiresAt,
		AbsoluteExpiresAt:        principal.Session.AbsoluteExpiresAt,
		SessionExpiresAt:         principal.Session.SessionExpiresAt,
	}
	persisted, err := s.authStore.SlideSession(ctx, principal.Session.ID, sliding)
	if err != nil {
		return err
	}
	principal.Session.LastQualifyingActivityAt = persisted.LastQualifyingActivityAt
	principal.Session.IdleExpiresAt = persisted.IdleExpiresAt
	principal.Session.SessionExpiresAt = persisted.SessionExpiresAt
	return nil
}

func parseJobPath(path string) (uuid.UUID, string, bool) {
	rest := strings.TrimPrefix(path, "/api/v1/jobs/")
	if rest == path || rest == "" {
		return uuid.UUID{}, "", false
	}
	parts := strings.Split(rest, "/")
	if len(parts) > 2 || parts[0] == "" {
		return uuid.UUID{}, "", false
	}
	jobID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.UUID{}, "", false
	}
	if len(parts) == 1 {
		return jobID, "", true
	}
	if parts[1] == "cancel" {
		return jobID, "cancel", true
	}
	return uuid.UUID{}, "", false
}

func decodeCancelRequest(r *http.Request) (cancelRequest, []byte, *auth.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&raw); err != nil {
		return cancelRequest{}, nil, invalidMutationPayload("", "request_not_object")
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return cancelRequest{}, nil, invalidMutationPayload("", "request_not_object")
	}
	for key := range raw {
		switch key {
		case "client_txn_id", "reason":
		default:
			return cancelRequest{}, nil, invalidMutationPayload(key, "unknown_field")
		}
	}
	var request cancelRequest
	if value, ok := raw["client_txn_id"]; !ok {
		return cancelRequest{}, nil, invalidMutationPayload("client_txn_id", "missing_required_field")
	} else if err := json.Unmarshal(value, &request.ClientTxnID); err != nil || strings.TrimSpace(request.ClientTxnID) == "" {
		return cancelRequest{}, nil, invalidMutationPayload("client_txn_id", "missing_required_field")
	}
	if value, ok := raw["reason"]; ok && !bytes.Equal(value, []byte("null")) {
		var reason string
		if err := json.Unmarshal(value, &reason); err != nil {
			return cancelRequest{}, nil, invalidMutationPayload("reason", "invalid_value")
		}
		request.Reason = &reason
	}
	normalizedMap := map[string]any{"client_txn_id": request.ClientTxnID}
	if request.Reason != nil {
		normalizedMap["reason"] = *request.Reason
	} else {
		normalizedMap["reason"] = nil
	}
	normalized, err := json.Marshal(normalizedMap)
	if err != nil {
		return cancelRequest{}, nil, internalAPIError(err)
	}
	return request, normalized, nil
}

func invalidMutationPayload(field string, reasonCode string) *auth.APIError {
	details := map[string]any{"reason_code": reasonCode}
	if field != "" {
		details["field"] = field
	}
	return &auth.APIError{Status: http.StatusBadRequest, Code: "invalid_mutation_payload", Details: details}
}

func jobNotFoundError() *auth.APIError {
	return &auth.APIError{Status: http.StatusNotFound, Code: "job_not_found", Details: map[string]any{}}
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *auth.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *auth.APIError {
	return &auth.APIError{
		Status:  http.StatusInternalServerError,
		Code:    "internal_error",
		Message: err.Error(),
		Details: map[string]any{},
	}
}
