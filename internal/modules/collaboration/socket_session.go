package collaboration

import (
	"context"
	"errors"
	"net/http"
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
	first, err := s.readFirstMessage(ctx, conn, incidentID)
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

	// Cover failed acknowledgement/replay writes as well as the live loop.
	defer s.removePresence(incidentID, connectionID)
	authority := &socketAuthority{
		sessionRevocations: sessionRevocations, incidentRevocations: incidentRevocations,
		incidentTerminals: incidentTerminals, sessionExpiresAt: principal.Session.SessionExpiresAt, now: s.now,
	}
	var handshake establishedSession
	// The upgrade may precede the first application frame by seconds. Resolve
	// current session state again without sliding its idle expiry.
	fresh, authenticationError := httpauth.AuthenticateRequest(r.WithContext(ctx), httpauth.Options{Store: s.authStore, Keys: s.keys, Hub: s.hub, Now: s.now})
	if authenticationError != nil {
		err = authority.check()
		if err == nil {
			err = &socketTermination{reasonCode: "session_revoked"}
		}
	} else {
		principal = fresh
		authority.sessionExpiresAt = fresh.Session.SessionExpiresAt
		if err = authority.check(); err == nil {
			handshake, err = s.establishSession(ctx, establishmentSocket{Socket: conn, authority: authority}, incidentID, principal, connectionID, messages, first)
		}
	}
	if err != nil {
		var terminal *socketTermination
		if errors.As(err, &terminal) {
			lifecycleResult = "canceled"
			if terminal.incidentClosed {
				closed = s.writeTerminalIncidentError(ctx, conn, incidentID, terminal.reasonCode)
			} else {
				closed = s.writeSessionRevoked(ctx, conn, incidentID, terminal.reasonCode)
			}
			return
		}
		lifecycleResult = "rejected"
		lifecycleErrorCode = "invalid_websocket_handshake"
		_ = s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "error", map[string]any{
			"code":      "invalid_websocket_handshake",
			"message":   "unable to establish collaboration session",
			"retryable": false,
		}, s.now()), 1008, invalidFirstMessage)
		closed = true
		return
	}
	lifecycleResult = "success"
	untrackConnection := s.hub.TrackActiveConnection()
	defer untrackConnection()

	incoming := make(chan protocol.Message, defaultSocketBuffer)
	readErrors := make(chan error, 1)
	go s.readLoop(ctx, conn, incidentID, incoming, readErrors)

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
	payload, err := protocol.DecodeClientPayload(message)
	if err != nil {
		return establishedSession{}, err
	}
	if err := s.workspaces.validate(payload.Presence); err != nil {
		return establishedSession{}, err
	}
	if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
		return establishedSession{}, &socketTermination{reasonCode: "incident_access_revoked"}
	}
	if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
		return establishedSession{}, err
	} else if closed {
		return establishedSession{}, &socketTermination{reasonCode: protocol.IncidentTerminalClosed, incidentClosed: true}
	}
	switch message.Type {
	case "hello":
		now := s.now()
		resumeToken, _, err := s.replay.IssueResumeToken(ctx, principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		highWater, err := s.replay.CurrentHighWater(ctx, incidentID)
		if err != nil {
			return establishedSession{}, err
		}
		ack := protocol.EphemeralMessage(incidentID, "hello_ack", map[string]any{
			"connection_id":         connectionID.String(),
			"resume_token":          resumeToken,
			"server_time":           now.UTC().Format(time.RFC3339Nano),
			"heartbeat_interval_ms": int(protocol.HeartbeatInterval / time.Millisecond),
			"presence_ttl_ms":       int(protocol.PresenceTTL / time.Millisecond),
			"resume_window_ms":      int(protocol.ResumeWindow / time.Millisecond),
		}, now)
		if err := s.publishSessionPresence(ctx, conn, incidentID, connectionID, ownMessages, principal, payload.Presence, ack, nil); err != nil {
			return establishedSession{}, err
		}
		return establishedSession{ClientInstanceID: payload.ClientInstanceID, LiveAfterStreamSeq: highWater}, nil

	case "resume":
		now := s.now()
		// Issue the replacement before retaining a snapshot connection. This does
		// not require the pool to provide a second connection during preparation.
		resumeToken, _, err := s.replay.IssueResumeToken(ctx, principal.Session.ID, incidentID, payload.ClientInstanceID, principal.Session.SessionExpiresAt, now)
		if err != nil {
			return establishedSession{}, err
		}
		replayCtx, cancelReplay := context.WithTimeout(ctx, protocol.HeartbeatTimeout)
		defer cancelReplay()
		replay, err := s.replay.PrepareReplay(replayCtx, principal.Session.ID, incidentID, payload.ClientInstanceID, payload.ResumeToken, payload.LastSeenStreamSeq, now)
		if err != nil {
			return establishedSession{}, err
		}
		defer replay.Close()
		ack := protocol.EphemeralMessage(incidentID, "resume_ack", map[string]any{
			"connection_id":                connectionID.String(),
			"status":                       replay.Status(),
			"resume_token":                 resumeToken,
			"server_high_water_stream_seq": replay.HighWater(),
		}, now)
		if err := s.publishSessionPresence(replayCtx, conn, incidentID, connectionID, ownMessages, principal, payload.Presence, ack, func() error {
			return replay.Visit(replayCtx, func(replayed protocol.Message) error { return s.writeMessage(replayCtx, conn, replayed) })
		}); err != nil {
			return establishedSession{}, err
		}
		return establishedSession{ClientInstanceID: payload.ClientInstanceID, LiveAfterStreamSeq: replay.HighWater()}, nil
	default:
		return establishedSession{}, errors.New("first message must be hello or resume")
	}
}

// Presence belongs to the entire establishment attempt. Failed acknowledgement,
// replay or snapshot writes must not leave a presence awaiting TTL expiry.
func (s *routeService) publishSessionPresence(ctx context.Context, conn protocol.Socket, incidentID, connectionID uuid.UUID, ownMessages <-chan protocol.Message, principal httpauth.Principal, input protocol.PresenceInput, ack protocol.Message, replay func() error) (err error) {
	now := s.now()
	presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, input, now)
	defer func() {
		if err != nil {
			s.removePresence(incidentID, connectionID)
		}
	}()
	if err = s.writeMessage(ctx, conn, ack); err != nil {
		return err
	}
	if replay != nil {
		if err = replay(); err != nil {
			return err
		}
	}
	now = s.now()
	if err = s.writeMessage(ctx, conn, protocol.PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
		return err
	}
	s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
	return nil
}

func (s *routeService) handleClientMessage(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, connectionID uuid.UUID, ownMessages <-chan protocol.Message, principal httpauth.Principal, message protocol.Message) bool {
	payload, err := protocol.DecodeClientPayload(message)
	if err != nil {
		s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
		return false
	}
	switch message.Type {
	case "pong":
		now := s.now()
		if presence, renewed := s.hub.RenewPresence(incidentID, connectionID, now); renewed {
			s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		}
		return true
	case "presence_update":
		if err := s.workspaces.validate(payload.Presence); err != nil {
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
