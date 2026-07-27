package admission

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	"github.com/google/uuid"
)

type Service struct {
	facade         *timeline.Facade
	incidentAccess incidents.Access
	authStore      *authn.Store
	keys           authn.MasterKeys
	now            func() time.Time
}

type RouteOptions struct {
	Facade *timeline.Facade
}

func RegisterRoutes(options RouteOptions) httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, options.Facade)
		if err != nil {
			return err
		}
		return httpapi.BindOwnerRoutes(mux, deps, "module.timeline", map[string]http.HandlerFunc{
			"getTimelineTimeConversionProfile": service.handleGetTimeConversionProfile,
			"markTimelineRecordReviewed":       service.handleMarkReviewed,
			"putTimelineTimeConversionProfile": service.handlePutTimeConversionProfile,
		})
	}
}

func newService(deps httpapi.DependencySet, facade *timeline.Facade) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, fmt.Errorf("load auth master key: %w", err)
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	if facade == nil {
		return nil, errors.New("timeline route composition requires a façade")
	}
	return &Service{
		facade:         facade,
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		keys:           keys,
		now:            now,
	}, nil
}

func (s *Service) handleMarkReviewed(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireTimelineRole(r.Context(), recordID, principal.User.ID, "reviewer", "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeTimelineActionRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}

	result, err := s.facade.MarkReviewedRow(r.Context(), timeline.MarkReviewedCommand{
		Actor:    principal.User,
		RecordID: recordID,
		Request:  request,
		RequestHash: ActionRequestHash(
			request.BaseRowVersion,
			request.ClientTxnID,
			request.Reason,
			nil,
		),
		RequestID: httpapi.RequestIDFromContext(r.Context()),
		Now:       s.now(),
	})
	if apiErr, ok := ClassifyMutationAPIError(err, MutationAPIErrorContext{
		ClientTxnID:                 request.ClientTxnID,
		IllegalTransitionReasonCode: "mark_reviewed_not_allowed",
	}); ok {
		writeAPIError(w, r, apiErr)
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
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, result.Payload)
}

func (s *Service) handleGetTimeConversionProfile(w http.ResponseWriter, r *http.Request) {
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
	profile, err := s.facade.GetTimeConversionProfile(r.Context(), incidentID, s.now())
	if err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	if err := s.slideSessionIfNeeded(r.Context(), &principal, r.Method, r.URL.Path); err != nil {
		writeAPIError(w, r, internalAPIError(err))
		return
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildTimeConversionProfilePayload(profile))
}

func (s *Service) handlePutTimeConversionProfile(w http.ResponseWriter, r *http.Request) {
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: true})
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentRole(r.Context(), incidentID, principal.User.ID, "admin"); apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	request, apiErr := DecodeTimelineTimeConversionProfilePutRequest(r.Body)
	if apiErr != nil {
		writeAPIError(w, r, apiErr)
		return
	}
	profile, err := s.facade.PutTimeConversionProfile(r.Context(), principal.User, incidentID, request, s.now())
	if apiErr, ok := ClassifyMutationAPIError(err, MutationAPIErrorContext{}); ok {
		writeAPIError(w, r, apiErr)
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
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, buildTimeConversionProfilePayload(profile))
}

func buildTimeConversionProfilePayload(profile timeline.TimeConversionProfile) map[string]any {
	return map[string]any{
		"incident_id":          profile.IncidentID.String(),
		"enabled":              profile.Enabled,
		"local_offset_minutes": derefInt(profile.LocalOffsetMinutes),
		"local_label":          derefString(profile.LocalLabel),
		"profile_version":      profile.ProfileVersion,
		"updated_at":           formatTimestamp(profile.UpdatedAt),
		"updated_by_user_id":   formatUUIDPointer(profile.UpdatedByUserID),
	}
}

func derefInt(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatTimestamp(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) requireTimelineRole(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, roles ...string) (uuid.UUID, *httpapi.APIError) {
	incidentID, err := s.facade.RecordIncident(ctx, recordID)
	if errors.Is(err, timeline.ErrRecordNotFound) {
		return uuid.UUID{}, incidentNotFoundError()
	}
	if err != nil {
		return uuid.UUID{}, internalAPIError(err)
	}
	if _, apiErr := s.requireIncidentRole(ctx, incidentID, userID, roles...); apiErr != nil {
		return uuid.UUID{}, apiErr
	}
	return incidentID, nil
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
