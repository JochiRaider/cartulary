package collaboration

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
)

const (
	firstMessageTimeout  = 10 * time.Second
	defaultSocketBuffer  = 32
	heartbeatCloseReason = "heartbeat_timeout"
	slowConsumerReason   = "slow_consumer"
)

func (s *routeService) handleIncidentSocket(w http.ResponseWriter, r *http.Request) {
	ctx, finishLifecycle := s.startWebSocketLifecycle(r.Context(), "connect")
	lifecycleResult := "failed"
	lifecycleErrorCode := ""
	defer func() {
		finishLifecycle(lifecycleResult, lifecycleErrorCode)
	}()

	if s.checkOrigin(w, r) {
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

	conn, err := s.acceptSocket(w, r)
	if err != nil {
		lifecycleResult = "failed"
		return
	}
	closed := false
	defer func() {
		if !closed {
			_ = conn.Close(1001, "")
		}
	}()

	sessionRevocations, unregisterSession := s.hub.RegisterSession(principal.Session.ID)
	defer unregisterSession()
	incidentRevocations, unregisterIncident := s.hub.RegisterIncidentUser(incidentID, principal.User.ID)
	defer unregisterIncident()
	incidentTerminals, unregisterIncidentTerminal := s.hub.RegisterIncidentTerminal(incidentID)
	defer unregisterIncidentTerminal()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	connectionID := uuid.New()
	messages, unsubscribe := s.hub.SubscribeIncident(incidentID, defaultSocketBuffer)
	defer unsubscribe()
	first, err := s.readFirstMessage(ctx, conn)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			lifecycleResult = "timeout"
			_ = conn.Close(1008, invalidFirstMessage)
		} else {
			lifecycleResult = "rejected"
			s.closeForDecodeFailure(ctx, conn, incidentID, err, true)
		}
		closed = true
		return
	}

	handshake, err := s.establishSession(ctx, conn, incidentID, principal, connectionID, messages, first)
	if err != nil {
		var terminalIncident terminalIncidentError
		if errors.As(err, &terminalIncident) {
			lifecycleResult = "canceled"
			closed = true
			return
		}
		lifecycleResult = "rejected"
		lifecycleErrorCode = "invalid_websocket_handshake"
		_ = s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "error", map[string]any{
			"code":      "invalid_websocket_handshake",
			"message":   err.Error(),
			"retryable": false,
		}, s.now()), 1008, invalidFirstMessage)
		closed = true
		return
	}
	lifecycleResult = "success"
	untrackConnection := s.hub.TrackActiveConnection()
	defer untrackConnection()

	defer s.removePresence(incidentID, connectionID)

	incoming := make(chan protocol.Message, defaultSocketBuffer)
	readErrors := make(chan error, 1)
	go s.readLoop(ctx, conn, incoming, readErrors)

	lastInbound := s.now()
	lastPing := s.now()
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
		case message, ok := <-messages:
			if !ok {
				lifecycleResult = "dropped"
				lifecycleErrorCode = slowConsumerReason
				_ = s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "resume_ack", map[string]any{
					"connection_id":                connectionID.String(),
					"status":                       protocol.ResumeStatusResetNeeded,
					"resume_token":                 "",
					"server_high_water_stream_seq": handshake.LiveAfterStreamSeq,
				}, s.now()), 1013, slowConsumerReason)
				closed = true
				return
			}
			if message.StreamSeq != nil && *message.StreamSeq <= handshake.LiveAfterStreamSeq {
				continue
			}
			if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
				lifecycleResult = "canceled"
				if s.writeSessionRevoked(ctx, conn, incidentID, "incident_access_revoked") {
					closed = true
				}
				return
			}
			if terminal, terminalErr := s.incidentClosed(ctx, incidentID, principal.User.ID); terminalErr != nil || terminal {
				lifecycleResult = "canceled"
				if s.writeTerminalIncidentError(ctx, conn, incidentID, protocol.IncidentTerminalClosed) {
					closed = true
				}
				return
			}
			if s.writeMessage(ctx, conn, message) != nil {
				lifecycleResult = "failed"
				return
			}
		case message := <-incoming:
			now := s.now()
			if !principal.Session.SessionExpiresAt.After(now) {
				lifecycleResult = "canceled"
				_ = s.authStore.RevokeSession(context.Background(), principal.Session.ID, "session_expired", now)
				s.hub.RevokeSession(principal.Session.ID, "session_expired")
				closed = s.writeSessionRevoked(ctx, conn, incidentID, "session_expired")
				return
			}
			if now.Sub(lastInbound) >= protocol.HeartbeatTimeout {
				lifecycleResult = "timeout"
				_ = conn.Close(1008, heartbeatCloseReason)
				closed = true
				return
			}
			if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
				lifecycleResult = "canceled"
				closed = s.writeSessionRevoked(ctx, conn, incidentID, "incident_access_revoked")
				return
			}
			if terminal, terminalErr := s.incidentClosed(ctx, incidentID, principal.User.ID); terminalErr != nil || terminal {
				lifecycleResult = "canceled"
				closed = s.writeTerminalIncidentError(ctx, conn, incidentID, protocol.IncidentTerminalClosed)
				return
			}
			if !s.handleClientMessage(ctx, conn, incidentID, connectionID, messages, principal, message) {
				lifecycleResult = "rejected"
				return
			}
			lastInbound = now
		case err := <-readErrors:
			lifecycleResult = "rejected"
			s.closeForDecodeFailure(ctx, conn, incidentID, err, false)
			closed = true
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
			if now.Sub(lastInbound) >= protocol.HeartbeatTimeout {
				lifecycleResult = "timeout"
				_ = conn.Close(1008, heartbeatCloseReason)
				closed = true
				return
			}
			if heartbeatPingDue(now, lastInbound, lastPing) {
				if s.writeMessage(ctx, conn, protocol.EphemeralMessage(incidentID, "ping", map[string]any{}, now)) != nil {
					lifecycleResult = "failed"
					return
				}
				lastPing = now
			}
		}
	}
}

type establishedSession struct {
	ClientInstanceID   string
	LiveAfterStreamSeq int64
}

// Outgoing traffic is not proof of peer liveness and cannot postpone a ping.
func heartbeatPingDue(now, lastInbound, lastPing time.Time) bool {
	return now.Sub(lastInbound) >= protocol.HeartbeatInterval && now.Sub(lastPing) >= protocol.HeartbeatInterval
}

func (s *routeService) establishSession(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, principal httpauth.Principal, connectionID uuid.UUID, ownMessages <-chan protocol.Message, message protocol.Message) (establishedSession, error) {
	switch message.Type {
	case "hello":
		var payload struct {
			ClientInstanceID string                 `json:"client_instance_id"`
			Presence         protocol.PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" {
			return establishedSession{}, errors.New("client_instance_id is required")
		}
		if err := protocol.ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, protocol.IncidentTerminalClosed)
			return establishedSession{}, terminalIncidentError{}
		}
		now := s.now()
		resumeToken, _, err := s.replay.IssueResumeToken(ctx, principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		highWater, err := s.replay.CurrentHighWater(ctx, incidentID)
		if err != nil {
			return establishedSession{}, err
		}
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		if err := s.writeMessage(ctx, conn, protocol.EphemeralMessage(incidentID, "hello_ack", map[string]any{
			"connection_id":         connectionID.String(),
			"resume_token":          resumeToken,
			"server_time":           now.UTC().Format(time.RFC3339Nano),
			"heartbeat_interval_ms": int(protocol.HeartbeatInterval / time.Millisecond),
			"presence_ttl_ms":       int(protocol.PresenceTTL / time.Millisecond),
			"resume_window_ms":      int(protocol.ResumeWindow / time.Millisecond),
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if err := s.writeMessage(ctx, conn, protocol.PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
			return establishedSession{}, err
		}
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		return establishedSession{ClientInstanceID: payload.ClientInstanceID, LiveAfterStreamSeq: highWater}, nil

	case "resume":
		var payload struct {
			ClientInstanceID  string                 `json:"client_instance_id"`
			ResumeToken       string                 `json:"resume_token"`
			LastSeenStreamSeq int64                  `json:"last_seen_stream_seq"`
			Presence          protocol.PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" || strings.TrimSpace(payload.ResumeToken) == "" {
			return establishedSession{}, errors.New("client_instance_id and resume_token are required")
		}
		if err := protocol.ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
			return establishedSession{}, errors.New("incident authorization no longer valid")
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, protocol.IncidentTerminalClosed)
			return establishedSession{}, terminalIncidentError{}
		}
		now := s.now()
		replay, err := s.replay.ReplayMessages(ctx, principal.Session.ID, incidentID, payload.ClientInstanceID, payload.ResumeToken, payload.LastSeenStreamSeq, now)
		if err != nil {
			return establishedSession{}, err
		}
		resumeToken, _, err := s.replay.IssueResumeToken(ctx, principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		if err := s.writeMessage(ctx, conn, protocol.EphemeralMessage(incidentID, "resume_ack", map[string]any{
			"connection_id":                connectionID.String(),
			"status":                       replay.Status,
			"resume_token":                 resumeToken,
			"server_high_water_stream_seq": replay.HighWater,
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if replay.Status == protocol.ResumeStatusReplayed {
			for _, replayed := range replay.Messages {
				if err := s.writeMessage(ctx, conn, replayed); err != nil {
					return establishedSession{}, err
				}
			}
		}
		if err := s.writeMessage(ctx, conn, protocol.PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
			return establishedSession{}, err
		}
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		return establishedSession{ClientInstanceID: payload.ClientInstanceID, LiveAfterStreamSeq: replay.HighWater}, nil
	default:
		return establishedSession{}, errors.New("first message must be hello or resume")
	}
}

type terminalIncidentError struct{}

func (terminalIncidentError) Error() string {
	return "terminal incident websocket"
}

func (s *routeService) handleClientMessage(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, connectionID uuid.UUID, ownMessages <-chan protocol.Message, principal httpauth.Principal, message protocol.Message) bool {
	switch message.Type {
	case "pong":
		var payload map[string]any
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		now := s.now()
		if presence, renewed := s.hub.RenewPresence(incidentID, connectionID, now); renewed {
			s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		}
		return true
	case "presence_update":
		var payload struct {
			Presence protocol.PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		if err := protocol.ValidatePresenceInput(payload.Presence); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		now := s.now()
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		return true
	case "resume", "hello":
		s.writeInvalidLaterMessage(ctx, conn, incidentID, "session establishment message already processed")
		return false
	default:
		s.writeInvalidLaterMessage(ctx, conn, incidentID, "unknown websocket message type")
		return false
	}
}

func (s *routeService) removePresence(incidentID uuid.UUID, connectionID uuid.UUID) {
	now := s.now()
	presence, ok := s.hub.RemovePresence(incidentID, connectionID, now)
	if ok {
		s.hub.BroadcastPresenceDelta(incidentID, "remove", presence, now)
	}
}

func (s *routeService) incidentClosed(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (bool, error) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	if err != nil {
		return false, err
	}
	return grant.IncidentStatus == admission.IncidentStatusClosed, nil
}

func (s *routeService) requireIncidentMembership(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (admission.Grant, *httpapi.APIError) {
	grant, err := s.incidentAccess.Check(ctx, incidentID, userID, admission.Requirement{
		AllowedRoles: admission.RolesMember,
		Lifecycle:    admission.LifecycleAny,
	})
	switch {
	case admission.IsDenied(err, admission.DenialNotVisible):
		return admission.Grant{}, &httpapi.APIError{Status: http.StatusNotFound, Code: "incident_not_found", Details: map[string]any{}}
	case err != nil:
		return admission.Grant{}, httpapi.InternalAPIError(err)
	default:
		return grant, nil
	}
}
