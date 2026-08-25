package server

import (
	"context"
	"errors"
	"net/http"

	protocol "github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
)

type collaborationSocket struct {
	connection *platformws.Connection
}

type collaborationSocketTransport struct {
	publicOrigin string
}

func newCollaborationSocketTransport(publicOrigin string) collaborationSocketTransport {
	return collaborationSocketTransport{publicOrigin: publicOrigin}
}

func (transport collaborationSocketTransport) Accept(w http.ResponseWriter, r *http.Request) (protocol.Socket, error) {
	connection, err := platformws.Accept(w, r, transport.publicOrigin)
	if err != nil {
		return nil, err
	}
	connection.SetReadLimit(protocol.MaximumMessageBytes)
	return &collaborationSocket{connection: connection}, nil
}

func (transport collaborationSocketTransport) CheckBrowserOrigin(w http.ResponseWriter, r *http.Request) bool {
	return platformws.RejectUntrustedBrowserOrigin(w, r, transport.publicOrigin)
}

func (s *collaborationSocket) Read(ctx context.Context) (protocol.MessageKind, []byte, error) {
	kind, payload, err := s.connection.Read(ctx)
	if errors.Is(err, platformws.ErrMessageTooLarge) {
		return 0, nil, protocol.ErrMessageTooLarge
	}
	if err != nil {
		return 0, nil, err
	}
	switch kind {
	case platformws.MessageText:
		return protocol.MessageText, payload, nil
	case platformws.MessageBinary:
		return protocol.MessageBinary, payload, nil
	default:
		return 0, nil, errors.New("unsupported WebSocket transport message kind")
	}
}

func (s *collaborationSocket) Write(
	ctx context.Context,
	kind protocol.MessageKind,
	payload []byte,
) error {
	switch kind {
	case protocol.MessageText:
		return s.connection.Write(ctx, platformws.MessageText, payload)
	case protocol.MessageBinary:
		return s.connection.Write(ctx, platformws.MessageBinary, payload)
	default:
		return errors.New("unsupported Collaboration message kind")
	}
}

func (s *collaborationSocket) Close(code uint16, reason string) error {
	return s.connection.Close(code, reason)
}
