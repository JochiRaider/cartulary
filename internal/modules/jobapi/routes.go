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

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

type Service struct {
	jobs           jobReadCanceler
	authStore      *authn.Store
	incidentAccess *admission.Checker
	now            func() time.Time
}

type jobReadCanceler interface {
	Get(context.Context, uuid.UUID) (jobs.Resource, error)
	Cancel(context.Context, jobs.CancelParams) (jobs.CancelResult, error)
}

type cancelRequest struct {
	ClientTxnID string
	Reason      *string
}

// RegisterRoutes binds the job routes for central admission (Core 01
// REQ-01-707): the registry authenticates by each operation's declared
// security and admits job_id before a handler runs.
func RegisterRoutes(jobService jobReadCanceler) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, jobService)
		if err != nil {
			return err
		}
		return httpapi.BindAdmittedOwnerRoutes(mux, deps, "module.jobapi", map[string]http.HandlerFunc{
			"cancelJob": service.handleCancelJob,
			"getJob":    service.handleGetJob,
		})
	}
}

func newService(deps httpapi.DependencySet, jobService jobReadCanceler) (*Service, error) {
	if jobService == nil {
		return nil, fmt.Errorf("job API requires read/cancel capability")
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		jobs:           jobService,
		authStore:      authn.NewStore(deps.PostgresHandle()),
		incidentAccess: admission.NewChecker(deps.PostgresHandle()),
		now:            now,
	}, nil
}

// admitted returns the principal and job_id that central admission
// established; their absence is a binding error.
func admitted(r *http.Request) (httpauth.Principal, uuid.UUID, *httpapi.APIError) {
	principal, authenticated := httpauth.AdmittedPrincipal(r)
	jobID, identified := httpapi.PathUUID(r, "job_id")
	if !authenticated || !identified {
		return httpauth.Principal{}, uuid.UUID{}, internalAPIError(errors.New("job route was not admitted centrally"))
	}
	return principal, jobID, nil
}

func (s *Service) handleGetJob(w http.ResponseWriter, r *http.Request) {
	principal, jobID, apiErr := admitted(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	resource, err := s.jobs.Get(r.Context(), jobID)
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
	// REQ-01-707 step 5: the singleton-query check is route handling, after
	// the job's visibility.
	if apiErr := httpapi.ValidateSingletonReadQuery(r.URL.Query()); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, resource)
}

func (s *Service) handleCancelJob(w http.ResponseWriter, r *http.Request) {
	principal, jobID, apiErr := admitted(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	// REQ-01-704: the job's visibility, including the cancel relationship, is
	// decided before the request body is classified.
	resource, err := s.jobs.Get(r.Context(), jobID)
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
	request, normalized, apiErr := decodeCancelRequest(r)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	result, err := s.jobs.Cancel(r.Context(), jobs.CancelParams{
		JobID:             jobID,
		ActorUserID:       principal.User.ID,
		ClientTxnID:       request.ClientTxnID,
		NormalizedRequest: normalized,
	})
	switch {
	case errors.Is(err, jobs.ErrClientTxnConflict):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "client_txn_conflict", Details: map[string]any{"client_txn_id": request.ClientTxnID}})
		return
	case errors.Is(err, jobs.ErrCancelRejected):
		writeAPIError(w, r, &httpapi.APIError{Status: http.StatusConflict, Code: "job_cancel_rejected", Details: map[string]any{"reason_code": result.ReasonCode}})
		return
	case errors.Is(err, jobs.ErrNotFound):
		writeAPIError(w, r, jobNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Resource)
}

func (s *Service) authorizeJob(ctx context.Context, resource jobs.Resource, principal httpauth.Principal, cancel bool) *httpapi.APIError {
	submittedByCurrentUser := resource.SubmittedByUserID != nil && *resource.SubmittedByUserID == principal.User.ID.String()
	switch resource.Scope.Kind {
	case jobs.ScopeKindDeployment:
		switch resource.AuthPolicy {
		case jobs.AuthPolicyDeploymentAdmin:
			if principal.User.IsDeploymentAdmin {
				return nil
			}
			return jobNotFoundError()
		case "", jobs.AuthPolicySubmitterOrDeploymentAdmin:
			if submittedByCurrentUser || principal.User.IsDeploymentAdmin {
				return nil
			}
			return jobNotFoundError()
		default:
			return jobNotFoundError()
		}
	case jobs.ScopeKindIncident:
		if resource.Scope.IncidentID == nil {
			return jobNotFoundError()
		}
		if resource.AuthPolicy == jobs.AuthPolicyDeploymentAdminIncidentMembership && !principal.User.IsDeploymentAdmin {
			return jobNotFoundError()
		}
		grant, err := s.incidentAccess.Check(ctx, *resource.Scope.IncidentID, principal.User.ID, admission.Requirement{
			AllowedRoles: admission.RolesMember,
			Lifecycle:    admission.LifecycleAny,
		})
		if admission.IsDenied(err, admission.DenialNotVisible) {
			return jobNotFoundError()
		}
		if err != nil {
			return internalAPIError(err)
		}
		switch resource.AuthPolicy {
		case "", jobs.AuthPolicyIncidentMembership:
		case jobs.AuthPolicyDeploymentAdminIncidentMembership:
		default:
			return jobNotFoundError()
		}
		// Core 01 §3.3.9.1: a member who neither submitted the job nor
		// administers its incident cannot cancel it, and learns only that no
		// such job exists.
		if !cancel || submittedByCurrentUser || grant.Role == admission.RoleAdmin {
			return nil
		}
		return jobNotFoundError()
	default:
		return jobNotFoundError()
	}
}

func (s *Service) slideSessionIfNeeded(ctx context.Context, principal *httpauth.Principal) error {
	return httpauth.SlideSessionIfNeeded(ctx, s.authStore, principal, s.now)
}

func decodeCancelRequest(r *http.Request) (cancelRequest, []byte, *httpapi.APIError) {
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

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{"reason_code": reasonCode}
	if field != "" {
		details["field"] = field
	}
	return &httpapi.APIError{Status: http.StatusBadRequest, Code: "invalid_mutation_payload", Details: details}
}

func jobNotFoundError() *httpapi.APIError {
	return httpapi.NotFound("job_not_found")
}

func writeAPIError(w http.ResponseWriter, r *http.Request, apiErr *httpapi.APIError) {
	message := apiErr.Message
	if message == "" {
		message = apiErr.Code
	}
	_ = httpapi.WriteError(w, r, apiErr.Status, apiErr.Code, message, apiErr.Details)
}

func internalAPIError(err error) *httpapi.APIError {
	return httpapi.InternalAPIError(err)
}
