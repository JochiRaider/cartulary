package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

const (
	unauthorizedCode     = "session_required"
	firstMessageTimeout  = 10 * time.Second
	writeTimeout         = 2 * time.Second
	defaultSocketBuffer  = 32
	invalidFirstMessage  = "invalid_first_message"
	heartbeatCloseReason = "heartbeat_timeout"
	sessionRevokedReason = "session_revoked"
	incidentClosedReason = "incident_closed"
)

type Service struct {
	incidentAccess incidents.Access
	authStore      *authn.Store
	hub            *platformws.Hub
	keys           authn.MasterKeys
	publicOrigin   string
	serviceVersion string
	now            func() time.Time
}

type Settings struct {
	PublicOrigin   string
	ServiceVersion string
}

func RegisterRoutes(settings ...Settings) httpapi.RouteRegistrar {
	resolved := Settings{}
	if len(settings) > 0 {
		resolved = settings[0]
	}
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		service, err := newService(deps, resolved)
		if err != nil {
			return err
		}
		mux.HandleFunc("GET /ws/v1/incidents/{incident_id}", service.handleIncidentSocket)
		return nil
	}
}

func newService(deps httpapi.DependencySet, settings Settings) (*Service, error) {
	keys, err := authn.LoadMasterKeys(deps.Env)
	if err != nil {
		return nil, err
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Service{
		incidentAccess: incidents.NewAccess(deps.PostgresHandle()),
		authStore:      authn.NewStore(deps.PostgresHandle()),
		hub:            deps.WSHub,
		keys:           keys,
		publicOrigin:   settings.PublicOrigin,
		serviceVersion: settings.ServiceVersion,
		now:            now,
	}, nil
}

func (s *Service) handleIncidentSocket(w http.ResponseWriter, r *http.Request) {
	ctx, finishLifecycle := s.startWebSocketLifecycle(r.Context(), "connect")
	lifecycleResult := "failed"
	lifecycleErrorCode := ""
	defer func() {
		finishLifecycle(lifecycleResult, lifecycleErrorCode)
	}()

	if platformws.RejectUntrustedBrowserOrigin(w, r, s.publicOrigin) {
		lifecycleResult = "rejected"
		return
	}
	incidentID, ok := pathUUID(w, r, "incident_id")
	if !ok {
		lifecycleResult = "rejected"
		return
	}
	principal, apiErr := httpauth.AuthenticateRequest(r, httpauth.Options{Store: s.authStore, Keys: s.keys, Now: s.now, StateChanging: false})
	if apiErr != nil {
		lifecycleResult, lifecycleErrorCode = webSocketLifecycleResultForAPIError(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}
	if _, apiErr := s.requireIncidentMembership(r.Context(), incidentID, principal.User.ID); apiErr != nil {
		lifecycleResult, lifecycleErrorCode = webSocketLifecycleResultForAPIError(apiErr)
		writeAPIError(w, r, apiErr)
		return
	}

	conn, err := platformws.Accept(w, r, s.publicOrigin)
	if err != nil {
		lifecycleResult = "failed"
		return
	}
	closed := false
	defer func() {
		if !closed {
			conn.CloseNow()
		}
	}()

	sessionRevocations, unregisterSession := s.hub.RegisterSession(principal.Session.ID)
	defer unregisterSession()
	incidentRevocations, unregisterIncident := s.hub.RegisterIncidentSession(incidentID, principal.Session.ID)
	defer unregisterIncident()
	incidentTerminals, unregisterIncidentTerminal := s.hub.RegisterIncidentTerminal(incidentID)
	defer unregisterIncidentTerminal()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connectionID := uuid.New()
	first, err := readFirstMessage(ctx, conn)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			lifecycleResult = "timeout"
		} else {
			lifecycleResult = "rejected"
		}
		_ = conn.Close(websocket.StatusPolicyViolation, invalidFirstMessage)
		closed = true
		return
	}

	handshake, err := s.establishSession(ctx, conn, incidentID, principal, connectionID, first)
	if err != nil {
		var terminalIncident terminalIncidentError
		if errors.As(err, &terminalIncident) {
			lifecycleResult = "canceled"
			closed = true
			return
		}
		lifecycleResult = "rejected"
		lifecycleErrorCode = "invalid_websocket_handshake"
		_ = writeThenClose(ctx, conn, platformws.EphemeralMessage(incidentID, "error", map[string]any{
			"code":      "invalid_websocket_handshake",
			"message":   err.Error(),
			"retryable": false,
		}, s.now()), websocket.StatusPolicyViolation, invalidFirstMessage)
		closed = true
		return
	}
	lifecycleResult = "success"
	untrackConnection := s.hub.TrackActiveConnection()
	defer untrackConnection()

	messages, unsubscribe := s.hub.SubscribeIncident(incidentID, defaultSocketBuffer)
	defer unsubscribe()
	defer s.removePresence(incidentID, connectionID)

	incoming := make(chan platformws.Message, defaultSocketBuffer)
	readErrors := make(chan error, 1)
	go readLoop(ctx, conn, incoming, readErrors)

	lastInbound := s.now()
	lastSent := s.now()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			lifecycleResult = "canceled"
			return
		case reasonCode := <-sessionRevocations:
			lifecycleResult = "canceled"
			if s.writeSessionRevoked(ctx, conn, incidentID, reasonCode) {
				closed = true
			}
			return
		case reasonCode := <-incidentRevocations:
			lifecycleResult = "canceled"
			if s.writeSessionRevoked(ctx, conn, incidentID, reasonCode) {
				closed = true
			}
			return
		case reasonCode := <-incidentTerminals:
			lifecycleResult = "canceled"
			if s.writeTerminalIncidentError(ctx, conn, incidentID, reasonCode) {
				closed = true
			}
			return
		case message := <-messages:
			if writeMessage(ctx, conn, message) != nil {
				lifecycleResult = "failed"
				return
			}
			lastSent = s.now()
		case message := <-incoming:
			lastInbound = s.now()
			if !s.handleClientMessage(ctx, conn, incidentID, connectionID, principal, handshake.ClientInstanceID, message) {
				lifecycleResult = "rejected"
				return
			}
		case <-readErrors:
			return
		case <-ticker.C:
			now := s.now()
			if !principal.Session.SessionExpiresAt.After(now) {
				lifecycleResult = "canceled"
				_ = s.authStore.RevokeSession(context.Background(), principal.Session.ID, "session_expired", now)
				s.hub.RevokeSession(principal.Session.ID, "session_expired")
				if s.writeSessionRevoked(ctx, conn, incidentID, "session_expired") {
					closed = true
				}
				return
			}
			if now.Sub(lastInbound) > platformws.HeartbeatTimeout {
				lifecycleResult = "timeout"
				_ = conn.Close(websocket.StatusPolicyViolation, heartbeatCloseReason)
				closed = true
				return
			}
			if now.Sub(lastSent) >= platformws.HeartbeatInterval {
				if writeMessage(ctx, conn, platformws.EphemeralMessage(incidentID, "ping", map[string]any{}, now)) != nil {
					lifecycleResult = "failed"
					return
				}
				lastSent = now
			}
		}
	}
}

type establishedSession struct {
	ClientInstanceID string
}

func (s *Service) establishSession(ctx context.Context, conn *websocket.Conn, incidentID uuid.UUID, principal httpauth.Principal, connectionID uuid.UUID, message platformws.Message) (establishedSession, error) {
	switch message.Type {
	case "hello":
		var payload struct {
			ClientInstanceID string                   `json:"client_instance_id"`
			Presence         platformws.PresenceInput `json:"presence"`
		}
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" {
			return establishedSession{}, errors.New("client_instance_id is required")
		}
		if err := platformws.ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, platformws.IncidentTerminalClosed)
			return establishedSession{}, terminalIncidentError{}
		}
		now := s.now()
		resumeToken, _, err := s.hub.IssueResumeToken(principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		if err := writeMessage(ctx, conn, platformws.EphemeralMessage(incidentID, "hello_ack", map[string]any{
			"connection_id":         connectionID.String(),
			"resume_token":          resumeToken,
			"server_time":           now.UTC().Format(time.RFC3339Nano),
			"heartbeat_interval_ms": int(platformws.HeartbeatInterval / time.Millisecond),
			"presence_ttl_ms":       int(platformws.PresenceTTL / time.Millisecond),
			"resume_window_ms":      int(platformws.ResumeWindow / time.Millisecond),
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if err := writeMessage(ctx, conn, platformws.PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
			return establishedSession{}, err
		}
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now)
		return establishedSession{ClientInstanceID: payload.ClientInstanceID}, nil

	case "resume":
		var payload struct {
			ClientInstanceID  string                   `json:"client_instance_id"`
			ResumeToken       string                   `json:"resume_token"`
			LastSeenStreamSeq int64                    `json:"last_seen_stream_seq"`
			Presence          platformws.PresenceInput `json:"presence"`
		}
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" || strings.TrimSpace(payload.ResumeToken) == "" {
			return establishedSession{}, errors.New("client_instance_id and resume_token are required")
		}
		if err := platformws.ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
			return establishedSession{}, errors.New("incident authorization no longer valid")
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, platformws.IncidentTerminalClosed)
			return establishedSession{}, terminalIncidentError{}
		}
		now := s.now()
		status, missed, highWater := s.hub.ReplayMessages(principal.Session.ID, incidentID, payload.ClientInstanceID, payload.ResumeToken, payload.LastSeenStreamSeq, now)
		resumeToken, _, err := s.hub.IssueResumeToken(principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		if err := writeMessage(ctx, conn, platformws.EphemeralMessage(incidentID, "resume_ack", map[string]any{
			"connection_id":                connectionID.String(),
			"status":                       status,
			"resume_token":                 resumeToken,
			"server_high_water_stream_seq": highWater,
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if status == platformws.ResumeStatusReplayed {
			for _, replayed := range missed {
				if err := writeMessage(ctx, conn, replayed); err != nil {
					return establishedSession{}, err
				}
			}
		}
		if err := writeMessage(ctx, conn, platformws.PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
			return establishedSession{}, err
		}
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now)
		return establishedSession{ClientInstanceID: payload.ClientInstanceID}, nil
	default:
		return establishedSession{}, errors.New("first message must be hello or resume")
	}
}

type terminalIncidentError struct{}

func (terminalIncidentError) Error() string {
	return "terminal incident websocket"
}

func (s *Service) handleClientMessage(ctx context.Context, conn *websocket.Conn, incidentID uuid.UUID, connectionID uuid.UUID, principal httpauth.Principal, clientInstanceID string, message platformws.Message) bool {
	switch message.Type {
	case "pong":
		return true
	case "presence_update":
		var payload struct {
			Presence platformws.PresenceInput `json:"presence"`
		}
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return false
		}
		if err := platformws.ValidatePresenceInput(payload.Presence); err != nil {
			return false
		}
		now := s.now()
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now)
		return true
	case "resume", "hello":
		_ = clientInstanceID
		_ = writeThenClose(ctx, conn, platformws.EphemeralMessage(incidentID, "error", map[string]any{
			"code":      "invalid_websocket_message",
			"message":   "session establishment message already processed",
			"retryable": false,
		}, s.now()), websocket.StatusPolicyViolation, "invalid_message")
		return false
	default:
		return true
	}
}

func (s *Service) removePresence(incidentID uuid.UUID, connectionID uuid.UUID) {
	now := s.now()
	presence, ok := s.hub.RemovePresence(incidentID, connectionID, now)
	if ok {
		s.hub.BroadcastPresenceDelta(incidentID, "remove", presence, now)
	}
}

func (s *Service) writeSessionRevoked(ctx context.Context, conn *websocket.Conn, incidentID uuid.UUID, reasonCode string) bool {
	return writeThenClose(ctx, conn, platformws.EphemeralMessage(incidentID, "session_revoked", map[string]any{
		"reason_code": reasonCode,
	}, s.now()), websocket.StatusPolicyViolation, sessionRevokedReason) == nil
}

func (s *Service) writeTerminalIncidentError(ctx context.Context, conn *websocket.Conn, incidentID uuid.UUID, reasonCode string) bool {
	return writeThenClose(ctx, conn, platformws.EphemeralMessage(incidentID, "error", map[string]any{
		"code":      reasonCode,
		"message":   "incident closed",
		"retryable": false,
	}, s.now()), websocket.StatusPolicyViolation, incidentClosedReason) == nil
}

func (s *Service) incidentClosed(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (bool, error) {
	record, err := s.incidentAccess.GetVisibleIncident(ctx, incidentID, userID)
	if err != nil {
		return false, err
	}
	return record.Status == "closed", nil
}

func readFirstMessage(ctx context.Context, conn *websocket.Conn) (platformws.Message, error) {
	readCtx, cancel := context.WithTimeout(ctx, firstMessageTimeout)
	defer cancel()
	var message platformws.Message
	err := platformws.ReadJSON(readCtx, conn, &message)
	return message, err
}

func readLoop(ctx context.Context, conn *websocket.Conn, incoming chan<- platformws.Message, errors chan<- error) {
	for {
		var message platformws.Message
		if err := platformws.ReadJSON(ctx, conn, &message); err != nil {
			select {
			case errors <- err:
			default:
			}
			return
		}
		select {
		case incoming <- message:
		case <-ctx.Done():
			return
		}
	}
}

func writeMessage(ctx context.Context, conn *websocket.Conn, message platformws.Message) error {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	return platformws.WriteJSON(writeCtx, conn, message)
}

func writeThenClose(ctx context.Context, conn *websocket.Conn, message platformws.Message, status websocket.StatusCode, reason string) error {
	if err := writeMessage(ctx, conn, message); err != nil {
		return err
	}
	return conn.Close(status, reason)
}

func (s *Service) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (incidents.MembershipRecord, *httpapi.APIError) {
	return incidents.RequireIncidentMembership(ctx, s.incidentAccess, incidentID, userID)
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
