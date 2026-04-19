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
	mu            sync.Mutex
	sessions      map[uuid.UUID]map[*registeredConn]struct{}
	recordChanges []RecordChange
	subscribers   map[chan RecordChange]struct{}
}

type RecordChange struct {
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	ActorUserID      uuid.UUID
	ChangedFieldKeys []string
	ViewSchemaID     string
}

type registeredConn struct {
	conn *websocket.Conn
	once sync.Once
}

func NewHub() *Hub {
	return &Hub{
		sessions:    make(map[uuid.UUID]map[*registeredConn]struct{}),
		subscribers: make(map[chan RecordChange]struct{}),
	}
}

func (h *Hub) PublishRecordChange(change RecordChange) {
	if h == nil {
		return
	}

	h.mu.Lock()
	h.recordChanges = append(h.recordChanges, change)
	subscribers := make([]chan RecordChange, 0, len(h.subscribers))
	for subscriber := range h.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	h.mu.Unlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- change:
		default:
		}
	}
}

func (h *Hub) SnapshotRecordChanges() []RecordChange {
	if h == nil {
		return nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	snapshot := make([]RecordChange, len(h.recordChanges))
	copy(snapshot, h.recordChanges)
	return snapshot
}

func (h *Hub) SubscribeRecordChanges(buffer int) (<-chan RecordChange, func()) {
	if h == nil {
		ch := make(chan RecordChange)
		close(ch)
		return ch, func() {}
	}
	if buffer < 1 {
		buffer = 1
	}

	ch := make(chan RecordChange, buffer)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		h.mu.Unlock()
		close(ch)
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
