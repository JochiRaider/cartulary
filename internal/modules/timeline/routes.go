package timeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/google/uuid"
)

type Service struct {
	facade         *Facade
	incidentAccess incidents.Access
	authStore      *authn.Store
	hub            *platformws.Hub
	publisher      *collaboration.RecordChangePublisher
	keys           authn.MasterKeys
	now            func() time.Time
}

func RegisterRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /api/v1/incidents/{incident_id}/timeline-time-conversion-profile", service.handleGetTimeConversionProfile)
		mux.HandleFunc("PUT /api/v1/incidents/{incident_id}/timeline-time-conversion-profile", service.handlePutTimeConversionProfile)
		mux.HandleFunc("POST /api/v1/records/{record_id}/mark-reviewed", service.handleMarkReviewed)
		return nil
	}
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		if !httpapi.TestRoutesEnabled(deps.Env) {
			return nil
		}
		guard, err := httpapi.NewTestRouteGuard(deps.Env)
		if err != nil {
			return err
		}
		service, err := newService(deps)
		if err != nil {
			return err
		}
		mux.HandleFunc("/api/v1/test/timeline/record-changes", guard.Protect(service.handleRecordChangeSnapshot))
		mux.HandleFunc("GET /api/v1/test/timeline/records/{record_id}/substrate", guard.Protect(service.handleRecordSubstrateSnapshot))
		mux.HandleFunc("/ws/v1/test/record-changes", guard.Protect(service.handleRecordChangeSocket))
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
	facade := FacadeFromDependencies(deps)
	facade.SetConflictTokenCodec(conflicttokens.NewConflictTokenCodec(keys))
	return &Service{
		facade:         facade,
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		hub:            deps.WSHub,
		publisher:      collaboration.NewRecordChangePublisher(deps.WSHub),
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

	result, err := s.facade.MarkReviewedRow(r.Context(), MarkReviewedCommand{
		Actor:     principal.User,
		RecordID:  recordID,
		Request:   request,
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

	if !result.Replayed {
		s.publishRecordChange(result, principal.User.ID)
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

func buildTimeConversionProfilePayload(profile TimeConversionProfile) map[string]any {
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

func (s *Service) handleRecordChangeSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	changes := s.hub.SnapshotRecordChanges()
	items := make([]map[string]any, 0, len(changes))
	for _, change := range changes {
		items = append(items, recordChangePayload(change))
	}
	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{"record_changes": items})
}

func (s *Service) handleRecordSubstrateSnapshot(w http.ResponseWriter, r *http.Request) {
	recordID, ok := pathUUID(w, r, "record_id")
	if !ok {
		return
	}

	snapshot, err := s.facade.SnapshotRecordSubstrate(r.Context(), recordID)
	switch {
	case errors.Is(err, ErrRecordNotFound):
		writeAPIError(w, r, incidentNotFoundError())
		return
	case err != nil:
		writeAPIError(w, r, internalAPIError(err))
		return
	}

	_ = httpapi.WriteSuccess(w, r, http.StatusOK, map[string]any{
		"record_id":             snapshot.RecordID.String(),
		"row_version":           snapshot.RowVersion,
		"capture_state":         snapshot.CaptureState,
		"replacement_record_id": formatUUIDPointer(snapshot.ReplacementRecordID),
		"record_revision_count": snapshot.RecordRevisionCount,
	})
}

func (s *Service) handleRecordChangeSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := platformws.Accept(w, r, "")
	if err != nil {
		return
	}
	defer conn.CloseNow()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	changes, unsubscribe := s.hub.SubscribeRecordChanges(16)
	defer unsubscribe()

	if err := platformws.WriteJSON(ctx, conn, platformws.Message{
		Type:    "connected",
		Payload: platformws.RawPayload(map[string]any{"boundary": "/ws/v1/test/record-changes"}),
	}); err != nil {
		return
	}

	go func() {
		defer cancel()
		for {
			var message platformws.Message
			if err := platformws.ReadJSON(ctx, conn, &message); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case change, ok := <-changes:
			if !ok {
				return
			}
			if err := platformws.WriteJSON(ctx, conn, platformws.Message{
				Type:    "record_changed",
				Payload: platformws.RawPayload(recordChangePayload(change)),
			}); err != nil {
				return
			}
		}
	}
}

func (s *Service) publishRecordChange(result MutationResult, actorUserID uuid.UUID) {
	if result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		return
	}
	s.publisher.Publish(collaboration.RecordChange{
		IncidentID:       result.Row.IncidentID,
		RecordID:         result.RecordID,
		RowVersion:       result.RowVersion,
		ChangeSetID:      result.ChangeSetID,
		ClientTxnID:      result.ClientTxnID,
		ActorUserID:      actorUserID,
		ChangedFieldKeys: result.ChangedFieldKeys,
		ViewSchemaID:     TimelineViewSchemaID,
		Row:              buildRow(result.Row),
	})
}

func recordChangePayload(change platformws.RecordChange) map[string]any {
	return map[string]any{
		"record_id":          change.RecordID.String(),
		"row_version":        change.RowVersion,
		"change_set_id":      change.ChangeSetID.String(),
		"client_txn_id":      change.ClientTxnID,
		"actor_user_id":      change.ActorUserID.String(),
		"changed_field_keys": append([]string(nil), change.ChangedFieldKeys...),
		"affected_views": []map[string]any{
			{
				"view_schema_id": change.ViewSchemaID,
				"change_kind":    "invalidate",
			},
		},
	}
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
}

func (s *Service) requireIncidentRole(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentRole(ctx, s.incidentAccess, incidentID, userID, roles...)
}

func (s *Service) requireTimelineRole(ctx context.Context, recordID uuid.UUID, userID uuid.UUID, roles ...string) (uuid.UUID, *httpapi.APIError) {
	incidentID, err := s.facade.RecordIncident(ctx, recordID)
	if errors.Is(err, ErrRecordNotFound) {
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
