package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

const (
	writeTimeout         = 2 * time.Second
	invalidFirstMessage  = "invalid_first_message"
	sessionRevokedReason = "session_revoked"
	incidentClosedReason = "incident_closed"
	invalidMessageReason = "invalid_message"
)

func (s *routeService) writeSessionRevoked(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, reasonCode string) bool {
	return s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "session_revoked", map[string]any{
		"reason_code": reasonCode,
	}, s.now()), 1008, sessionRevokedReason) == nil
}

func (s *routeService) writeTerminalIncidentError(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, reasonCode string) bool {
	return s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "error", map[string]any{
		"code":      reasonCode,
		"message":   "incident closed",
		"retryable": false,
	}, s.now()), 1008, incidentClosedReason) == nil
}

func (s *routeService) readFirstMessage(ctx context.Context, conn protocol.Socket) (protocol.Message, error) {
	readCtx, cancel := context.WithTimeout(ctx, firstMessageTimeout)
	defer cancel()
	return s.readMessage(readCtx, conn)
}

func (s *routeService) readLoop(ctx context.Context, conn protocol.Socket, incoming chan<- protocol.Message, errors chan<- error) {
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

func (s *routeService) readMessage(ctx context.Context, conn protocol.Socket) (protocol.Message, error) {
	kind, payload, err := conn.Read(ctx)
	if err != nil {
		return protocol.Message{}, err
	}
	return s.codec.Decode(kind, payload)
}

func (s *routeService) writeMessage(ctx context.Context, conn protocol.Socket, message protocol.Message) error {
	writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
	defer cancel()
	encoded, err := s.codec.Encode(message)
	if err != nil {
		return err
	}
	return conn.Write(writeCtx, protocol.MessageText, encoded)
}

func (s *routeService) writeThenClose(ctx context.Context, conn protocol.Socket, message protocol.Message, status uint16, reason string) error {
	if err := s.writeMessage(ctx, conn, message); err != nil {
		return err
	}
	return conn.Close(status, reason)
}

func (s *routeService) writeInvalidLaterMessage(ctx context.Context, conn protocol.Socket, incidentID uuid.UUID, message string) {
	_ = s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "error", map[string]any{
		"code":      "invalid_websocket_message",
		"message":   message,
		"retryable": false,
	}, s.now()), 1008, invalidMessageReason)
}

func (s *routeService) closeForDecodeFailure(
	ctx context.Context,
	conn protocol.Socket,
	incidentID uuid.UUID,
	err error,
	first bool,
) {
	if errors.Is(err, protocol.ErrMessageTooLarge) {
		_ = conn.Close(1009, "message_too_large")
		return
	}
	var failure *protocol.DecodeFailure
	if errors.As(err, &failure) {
		switch failure.Kind {
		case protocol.DecodeFailureBinaryMessage:
			_ = conn.Close(1003, "binary_message_unsupported")
		case protocol.DecodeFailureInvalidJSON:
			_ = conn.Close(1007, "invalid_json")
		case protocol.DecodeFailureDuplicateMember:
			if first {
				_ = s.writeThenClose(ctx, conn, protocol.EphemeralMessage(incidentID, "error", map[string]any{
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
