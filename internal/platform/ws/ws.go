package ws

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Accept upgrades a request into a placeholder WebSocket connection.
// TODO: replace this bootstrap shim with the real collaboration transport.
func Accept(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return websocket.Accept(w, r, nil)
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func ReadJSON(ctx context.Context, conn *websocket.Conn, payload any) error {
	return wsjson.Read(ctx, conn, payload)
}

func WriteJSON(ctx context.Context, conn *websocket.Conn, payload any) error {
	return wsjson.Write(ctx, conn, payload)
}

func RawPayload(payload any) json.RawMessage {
	if payload == nil {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil
	}

	return data
}
