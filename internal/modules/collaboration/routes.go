package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpauth"
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
	slowConsumerReason   = "slow_consumer"
	invalidMessageReason = "invalid_message"
)

type Service struct {
	incidentAccess incidents.Access
	authStore      *authn.Store
	hub            *Hub
	replay         *ReplayStore
	keys           authn.MasterKeys
	acceptSocket   AcceptSocket
	checkOrigin    CheckBrowserOrigin
	codec          Codec
	serviceVersion string
	now            func() time.Time
}

type Settings struct {
	AcceptSocket       AcceptSocket
	CheckBrowserOrigin CheckBrowserOrigin
	Hub                *Hub
	ServiceVersion     string
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
	if settings.AcceptSocket == nil {
		return nil, errors.New("collaboration WebSocket accept dependency is required")
	}
	if settings.CheckBrowserOrigin == nil {
		return nil, errors.New("collaboration WebSocket Origin dependency is required")
	}
	if settings.Hub == nil {
		return nil, errors.New("collaboration Hub dependency is required")
	}
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
		hub:            settings.Hub,
		replay:         NewReplayStore(deps.PostgresHandle(), now),
		keys:           keys,
		acceptSocket:   settings.AcceptSocket,
		checkOrigin:    settings.CheckBrowserOrigin,
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
	incidentRevocations, unregisterIncident := s.hub.RegisterIncidentSession(incidentID, principal.Session.ID)
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
		_ = s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "error", map[string]any{
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

	incoming := make(chan Message, defaultSocketBuffer)
	readErrors := make(chan error, 1)
	go s.readLoop(ctx, conn, incoming, readErrors)

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
		case message, ok := <-messages:
			if !ok {
				lifecycleResult = "dropped"
				lifecycleErrorCode = slowConsumerReason
				_ = s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "resume_ack", map[string]any{
					"connection_id":                connectionID.String(),
					"status":                       ResumeStatusResetNeeded,
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
				if s.writeTerminalIncidentError(ctx, conn, incidentID, IncidentTerminalClosed) {
					closed = true
				}
				return
			}
			if s.writeMessage(ctx, conn, message) != nil {
				lifecycleResult = "failed"
				return
			}
			lastSent = s.now()
		case message := <-incoming:
			lastInbound = s.now()
			if !s.handleClientMessage(ctx, conn, incidentID, connectionID, messages, principal, handshake.ClientInstanceID, message) {
				lifecycleResult = "rejected"
				return
			}
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
			if now.Sub(lastInbound) > HeartbeatTimeout {
				lifecycleResult = "timeout"
				_ = conn.Close(1008, heartbeatCloseReason)
				closed = true
				return
			}
			if now.Sub(lastSent) >= HeartbeatInterval {
				if s.writeMessage(ctx, conn, EphemeralMessage(incidentID, "ping", map[string]any{}, now)) != nil {
					lifecycleResult = "failed"
					return
				}
				lastSent = now
			}
		}
	}
}

type establishedSession struct {
	ClientInstanceID   string
	LiveAfterStreamSeq int64
}

func (s *Service) establishSession(ctx context.Context, conn Socket, incidentID uuid.UUID, principal httpauth.Principal, connectionID uuid.UUID, ownMessages <-chan Message, message Message) (establishedSession, error) {
	switch message.Type {
	case "hello":
		var payload struct {
			ClientInstanceID string        `json:"client_instance_id"`
			Presence         PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" {
			return establishedSession{}, errors.New("client_instance_id is required")
		}
		if err := ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, IncidentTerminalClosed)
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
		if err := s.writeMessage(ctx, conn, EphemeralMessage(incidentID, "hello_ack", map[string]any{
			"connection_id":         connectionID.String(),
			"resume_token":          resumeToken,
			"server_time":           now.UTC().Format(time.RFC3339Nano),
			"heartbeat_interval_ms": int(HeartbeatInterval / time.Millisecond),
			"presence_ttl_ms":       int(PresenceTTL / time.Millisecond),
			"resume_window_ms":      int(ResumeWindow / time.Millisecond),
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if err := s.writeMessage(ctx, conn, PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
			return establishedSession{}, err
		}
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		return establishedSession{ClientInstanceID: payload.ClientInstanceID, LiveAfterStreamSeq: highWater}, nil

	case "resume":
		var payload struct {
			ClientInstanceID  string        `json:"client_instance_id"`
			ResumeToken       string        `json:"resume_token"`
			LastSeenStreamSeq int64         `json:"last_seen_stream_seq"`
			Presence          PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			return establishedSession{}, err
		}
		if strings.TrimSpace(payload.ClientInstanceID) == "" || strings.TrimSpace(payload.ResumeToken) == "" {
			return establishedSession{}, errors.New("client_instance_id and resume_token are required")
		}
		if err := ValidatePresenceInput(payload.Presence); err != nil {
			return establishedSession{}, err
		}
		if _, apiErr := s.requireIncidentMembership(ctx, incidentID, principal.User.ID); apiErr != nil {
			return establishedSession{}, errors.New("incident authorization no longer valid")
		}
		if closed, err := s.incidentClosed(ctx, incidentID, principal.User.ID); err != nil {
			return establishedSession{}, err
		} else if closed {
			_ = s.writeTerminalIncidentError(ctx, conn, incidentID, IncidentTerminalClosed)
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
		if err := s.writeMessage(ctx, conn, EphemeralMessage(incidentID, "resume_ack", map[string]any{
			"connection_id":                connectionID.String(),
			"status":                       replay.Status,
			"resume_token":                 resumeToken,
			"server_high_water_stream_seq": replay.HighWater,
		}, now)); err != nil {
			return establishedSession{}, err
		}
		if replay.Status == ResumeStatusReplayed {
			for _, replayed := range replay.Messages {
				if err := s.writeMessage(ctx, conn, replayed); err != nil {
					return establishedSession{}, err
				}
			}
		}
		if err := s.writeMessage(ctx, conn, PresenceSnapshotMessage(incidentID, s.hub.PresenceSnapshot(incidentID, now), now)); err != nil {
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

func (s *Service) handleClientMessage(ctx context.Context, conn Socket, incidentID uuid.UUID, connectionID uuid.UUID, ownMessages <-chan Message, principal httpauth.Principal, clientInstanceID string, message Message) bool {
	switch message.Type {
	case "pong":
		var payload map[string]any
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		return true
	case "presence_update":
		var payload struct {
			Presence PresenceInput `json:"presence"`
		}
		if err := decodePayloadObject(message.Payload, &payload); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		if err := ValidatePresenceInput(payload.Presence); err != nil {
			s.writeInvalidLaterMessage(ctx, conn, incidentID, err.Error())
			return false
		}
		now := s.now()
		presence := s.hub.UpsertPresence(incidentID, connectionID, principal.User.ID, principal.User.DisplayName, payload.Presence, now)
		s.hub.BroadcastPresenceDelta(incidentID, "upsert", presence, now, ownMessages)
		return true
	case "resume", "hello":
		_ = clientInstanceID
		s.writeInvalidLaterMessage(ctx, conn, incidentID, "session establishment message already processed")
		return false
	default:
		s.writeInvalidLaterMessage(ctx, conn, incidentID, "unknown websocket message type")
		return false
	}
}

func (s *Service) removePresence(incidentID uuid.UUID, connectionID uuid.UUID) {
	now := s.now()
	presence, ok := s.hub.RemovePresence(incidentID, connectionID, now)
	if ok {
		s.hub.BroadcastPresenceDelta(incidentID, "remove", presence, now)
	}
}

func (s *Service) writeSessionRevoked(ctx context.Context, conn Socket, incidentID uuid.UUID, reasonCode string) bool {
	return s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "session_revoked", map[string]any{
		"reason_code": reasonCode,
	}, s.now()), 1008, sessionRevokedReason) == nil
}

func (s *Service) writeTerminalIncidentError(ctx context.Context, conn Socket, incidentID uuid.UUID, reasonCode string) bool {
	return s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "error", map[string]any{
		"code":      reasonCode,
		"message":   "incident closed",
		"retryable": false,
	}, s.now()), 1008, incidentClosedReason) == nil
}

func (s *Service) incidentClosed(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (bool, error) {
	record, err := s.incidentAccess.GetVisibleIncident(ctx, incidentID, userID)
	if err != nil {
		return false, err
	}
	return record.Status == "closed", nil
}

func (s *Service) readFirstMessage(ctx context.Context, conn Socket) (Message, error) {
	readCtx, cancel := context.WithTimeout(ctx, firstMessageTimeout)
	defer cancel()
	return s.readMessage(readCtx, conn)
}

func (s *Service) readLoop(ctx context.Context, conn Socket, incoming chan<- Message, errors chan<- error) {
	for {
		message, err := s.readMessage(ctx, conn)
		if err != nil {
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

func (s *Service) readMessage(ctx context.Context, conn Socket) (Message, error) {
	kind, payload, err := conn.Read(ctx)
	if err != nil {
		return Message{}, err
	}
	return s.codec.Decode(kind, payload)
}

func (s *Service) writeMessage(ctx context.Context, conn Socket, message Message) error {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	encoded, err := s.codec.Encode(message)
	if err != nil {
		return err
	}
	return conn.Write(writeCtx, MessageText, encoded)
}

func (s *Service) writeThenClose(ctx context.Context, conn Socket, message Message, status uint16, reason string) error {
	if err := s.writeMessage(ctx, conn, message); err != nil {
		return err
	}
	return conn.Close(status, reason)
}

func (s *Service) writeInvalidLaterMessage(ctx context.Context, conn Socket, incidentID uuid.UUID, message string) {
	_ = s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "error", map[string]any{
		"code":      "invalid_websocket_message",
		"message":   message,
		"retryable": false,
	}, s.now()), 1008, invalidMessageReason)
}

func (s *Service) closeForDecodeFailure(
	ctx context.Context,
	conn Socket,
	incidentID uuid.UUID,
	err error,
	first bool,
) {
	if errors.Is(err, ErrMessageTooLarge) {
		_ = conn.Close(1009, "message_too_large")
		return
	}
	var failure *DecodeFailure
	if errors.As(err, &failure) {
		switch failure.Kind {
		case DecodeFailureBinaryMessage:
			_ = conn.Close(1003, "binary_message_unsupported")
		case DecodeFailureInvalidJSON:
			_ = conn.Close(1007, "invalid_json")
		case DecodeFailureDuplicateMember:
			if first {
				_ = s.writeThenClose(ctx, conn, EphemeralMessage(incidentID, "error", map[string]any{
					"code":      "invalid_websocket_handshake",
					"message":   failure.Error(),
					"retryable": false,
				}, s.now()), 1008, invalidFirstMessage)
			} else {
				s.writeInvalidLaterMessage(ctx, conn, incidentID, failure.Error())
			}
		}
		return
	}
	_ = conn.Close(1001, "")
}

func decodePayloadObject(payload json.RawMessage, target any) error {
	trimmed := strings.TrimSpace(string(payload))
	if len(trimmed) < 2 || trimmed[0] != '{' {
		return errors.New("websocket payload must be an object")
	}
	if err := json.Unmarshal(payload, target); err != nil {
		return err
	}
	return nil
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
