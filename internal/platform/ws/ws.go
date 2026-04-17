package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
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

type Hub struct {
	mu       sync.Mutex
	sessions map[uuid.UUID]map[*registeredConn]struct{}
}

type registeredConn struct {
	conn *websocket.Conn
	once sync.Once
}

func NewHub() *Hub {
	return &Hub{
		sessions: make(map[uuid.UUID]map[*registeredConn]struct{}),
	}
}

func (h *Hub) RegisterSession(sessionID uuid.UUID, conn *websocket.Conn) func() {
	if h == nil || conn == nil {
		return func() {}
	}

	current := &registeredConn{conn: conn}

	h.mu.Lock()
	sessionSet := h.sessions[sessionID]
	if sessionSet == nil {
		sessionSet = make(map[*registeredConn]struct{})
		h.sessions[sessionID] = sessionSet
	}
	sessionSet[current] = struct{}{}
	h.mu.Unlock()

	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		sessionSet := h.sessions[sessionID]
		if sessionSet == nil {
			return
		}
		delete(sessionSet, current)
		if len(sessionSet) == 0 {
			delete(h.sessions, sessionID)
		}
	}
}

func (h *Hub) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	if h == nil {
		return
	}

	h.mu.Lock()
	sessionSet := h.sessions[sessionID]
	if len(sessionSet) == 0 {
		h.mu.Unlock()
		return
	}

	connections := make([]*registeredConn, 0, len(sessionSet))
	for current := range sessionSet {
		connections = append(connections, current)
	}
	delete(h.sessions, sessionID)
	h.mu.Unlock()

	for _, current := range connections {
		current.revoke(reasonCode)
	}
}

func (c *registeredConn) revoke(reasonCode string) {
	if c == nil || c.conn == nil {
		return
	}

	c.once.Do(func() {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			_ = WriteJSON(ctx, c.conn, Message{
				Type:    "session_revoked",
				Payload: RawPayload(map[string]any{"reason_code": reasonCode}),
			})
			_ = c.conn.Close(websocket.StatusPolicyViolation, "session_revoked")
		}()
	})
}
