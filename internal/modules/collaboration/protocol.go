package collaboration

import (
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
)

type hub struct {
	mu                    sync.Mutex
	sessions              map[uuid.UUID]map[chan string]struct{}
	incidentUsers         map[incidentUserKey]map[chan string]struct{}
	incidentTerminals     map[uuid.UUID]map[chan string]struct{}
	incidentStreams       map[uuid.UUID]map[chan protocol.Message]struct{}
	presences             map[uuid.UUID]map[string]protocol.PresenceRecord
	serviceVersion        string
	activeConnections     atomic.Int64
	activeGaugeRegistered bool
}

type incidentUserKey struct {
	IncidentID uuid.UUID
	UserID     uuid.UUID
}

func newHub() *hub {
	return &hub{
		sessions:          make(map[uuid.UUID]map[chan string]struct{}),
		incidentUsers:     make(map[incidentUserKey]map[chan string]struct{}),
		incidentTerminals: make(map[uuid.UUID]map[chan string]struct{}),
		incidentStreams:   make(map[uuid.UUID]map[chan protocol.Message]struct{}),
		presences:         make(map[uuid.UUID]map[string]protocol.PresenceRecord),
	}
}

// DeliverReplayable delivers an already sequenced durable event. Collaboration
// owns event identity, sequence, replay, and retry; the hub owns only ephemeral
// fan-out.
func (h *hub) DeliverReplayable(message protocol.Message) error {
	if h == nil {
		return errors.New("websocket hub is unavailable")
	}
	incidentID, err := uuid.Parse(message.IncidentID)
	if err != nil || incidentID == uuid.Nil || message.EventID == "" || message.StreamSeq == nil || *message.StreamSeq < 1 {
		return errors.New("invalid sequenced websocket message")
	}
	if !protocol.IsReplayableMessageType(message.Type) || !json.Valid(message.Payload) {
		return errors.New("invalid replayable websocket message")
	}
	finishTelemetry := h.startEventSend(message.Type)

	h.mu.Lock()
	if message.Type == "record_changed" {
		if _, err := protocol.RecordChangeFromSequencedMessage(message); err != nil {
			h.mu.Unlock()
			finishTelemetry("rejected", "")
			return err
		}
	}
	droppedSlowConsumer := false
	for subscriber := range h.incidentStreams[incidentID] {
		select {
		case subscriber <- message:
		default:
			delete(h.incidentStreams[incidentID], subscriber)
			close(subscriber)
			droppedSlowConsumer = true
		}
	}
	if len(h.incidentStreams[incidentID]) == 0 {
		delete(h.incidentStreams, incidentID)
	}
	h.mu.Unlock()
	if droppedSlowConsumer {
		finishTelemetry("dropped", "queue_full")
	} else {
		finishTelemetry("success", "")
	}
	return nil
}

func (h *hub) SubscribeIncident(incidentID uuid.UUID, buffer int) (<-chan protocol.Message, func()) {
	if h == nil {
		ch := make(chan protocol.Message)
		close(ch)
		return ch, func() {}
	}
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan protocol.Message, buffer)
	h.mu.Lock()
	subscribers := h.incidentStreams[incidentID]
	if subscribers == nil {
		subscribers = make(map[chan protocol.Message]struct{})
		h.incidentStreams[incidentID] = subscribers
	}
	subscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		if subscribers := h.incidentStreams[incidentID]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(h.incidentStreams, incidentID)
			}
		}
		h.mu.Unlock()
	}
}

func (h *hub) UpsertPresence(incidentID uuid.UUID, connectionID uuid.UUID, userID uuid.UUID, displayName string, input protocol.PresenceInput, now time.Time) protocol.PresenceRecord {
	record := protocol.PresenceRecord{
		ConnectionID: connectionID.String(),
		UserID:       userID.String(),
		DisplayName:  displayName,
		SheetRef:     cloneSheetRef(input.SheetRef),
		RecordID:     cloneStringPointer(input.RecordID),
		FieldKey:     cloneStringPointer(input.FieldKey),
		Mode:         input.Mode,
		ObservedAt:   now.UTC().Format(time.RFC3339Nano),
		ExpiresAt:    now.UTC().Add(protocol.PresenceTTL).Format(time.RFC3339Nano),
	}
	h.mu.Lock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	if incidentPresences == nil {
		incidentPresences = make(map[string]protocol.PresenceRecord)
		h.presences[incidentID] = incidentPresences
	}
	incidentPresences[record.ConnectionID] = record
	h.mu.Unlock()
	return record
}

func (h *hub) RemovePresence(incidentID uuid.UUID, connectionID uuid.UUID, now time.Time) (protocol.PresenceRecord, bool) {
	if h == nil {
		return protocol.PresenceRecord{}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	if incidentPresences == nil {
		return protocol.PresenceRecord{}, false
	}
	record, ok := incidentPresences[connectionID.String()]
	if ok {
		delete(incidentPresences, connectionID.String())
	}
	if len(incidentPresences) == 0 {
		delete(h.presences, incidentID)
	}
	return record, ok
}

func (h *hub) PresenceSnapshot(incidentID uuid.UUID, now time.Time) []protocol.PresenceRecord {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	presences := make([]protocol.PresenceRecord, 0, len(incidentPresences))
	for _, record := range incidentPresences {
		presences = append(presences, record)
	}
	sort.Slice(presences, func(i, j int) bool {
		return presences[i].ConnectionID < presences[j].ConnectionID
	})
	return presences
}

func (h *hub) BroadcastPresenceDelta(incidentID uuid.UUID, kind string, presence protocol.PresenceRecord, now time.Time, excluded ...<-chan protocol.Message) {
	if h == nil {
		return
	}
	finishTelemetry := h.startEventSend("presence_delta")
	defer finishTelemetry("success", "")
	message := protocol.EphemeralMessage(incidentID, "presence_delta", map[string]any{
		"delta_kind": kind,
		"presence":   presence,
	}, now)
	var excludedSubscriber <-chan protocol.Message
	if len(excluded) > 0 {
		excludedSubscriber = excluded[0]
	}
	h.broadcastIncidentExcept(incidentID, message, excludedSubscriber)
}

func (h *hub) RegisterIncidentUser(incidentID uuid.UUID, userID uuid.UUID) (<-chan string, func()) {
	if h == nil {
		return nil, func() {}
	}
	ch := make(chan string, 1)
	key := incidentUserKey{IncidentID: incidentID, UserID: userID}
	h.mu.Lock()
	subscribers := h.incidentUsers[key]
	if subscribers == nil {
		subscribers = make(map[chan string]struct{})
		h.incidentUsers[key] = subscribers
	}
	subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if subscribers := h.incidentUsers[key]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(h.incidentUsers, key)
			}
		}
		h.mu.Unlock()
	}
}

func (h *hub) RegisterIncidentTerminal(incidentID uuid.UUID) (<-chan string, func()) {
	if h == nil {
		return nil, func() {}
	}
	ch := make(chan string, 1)
	h.mu.Lock()
	subscribers := h.incidentTerminals[incidentID]
	if subscribers == nil {
		subscribers = make(map[chan string]struct{})
		h.incidentTerminals[incidentID] = subscribers
	}
	subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if subscribers := h.incidentTerminals[incidentID]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(h.incidentTerminals, incidentID)
			}
		}
		h.mu.Unlock()
	}
}

func (h *hub) TrackActiveConnection() func() {
	if h == nil {
		return func() {}
	}
	h.activeConnections.Add(1)
	return func() {
		h.activeConnections.Add(-1)
	}
}

func (h *hub) ActiveConnections() int64 {
	if h == nil {
		return 0
	}
	count := h.activeConnections.Load()
	if count < 0 {
		return 0
	}
	return count
}

func (h *hub) RevokeIncidentUser(incidentID uuid.UUID, userID uuid.UUID, reasonCode string) {
	if h == nil {
		return
	}
	key := incidentUserKey{IncidentID: incidentID, UserID: userID}
	h.mu.Lock()
	userSet := h.incidentUsers[key]
	subscribers := make([]chan string, 0, len(userSet))
	for current := range userSet {
		subscribers = append(subscribers, current)
	}
	delete(h.incidentUsers, key)
	h.mu.Unlock()
	for _, current := range subscribers {
		current <- reasonCode
	}
}

func (h *hub) RevokeIncidentAccess(incidentID uuid.UUID, userID uuid.UUID) {
	h.RevokeIncidentUser(incidentID, userID, "incident_access_revoked")
}

func (h *hub) TerminateIncident(incidentID uuid.UUID, reasonCode string) {
	if h == nil {
		return
	}
	h.mu.Lock()
	terminalSet := h.incidentTerminals[incidentID]
	subscribers := make([]chan string, 0, len(terminalSet))
	for current := range terminalSet {
		subscribers = append(subscribers, current)
	}
	delete(h.incidentTerminals, incidentID)
	h.mu.Unlock()
	for _, current := range subscribers {
		current <- reasonCode
	}
}

func (h *hub) RegisterSession(sessionID uuid.UUID) (<-chan string, func()) {
	if h == nil {
		return nil, func() {}
	}

	current := make(chan string, 1)

	h.mu.Lock()
	sessionSet := h.sessions[sessionID]
	if sessionSet == nil {
		sessionSet = make(map[chan string]struct{})
		h.sessions[sessionID] = sessionSet
	}
	sessionSet[current] = struct{}{}
	h.mu.Unlock()

	return current, func() {
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

func (h *hub) RevokeSession(sessionID uuid.UUID, reasonCode string) {
	if h == nil {
		return
	}

	h.mu.Lock()
	sessionSet := h.sessions[sessionID]
	if len(sessionSet) == 0 {
		h.mu.Unlock()
		return
	}

	subscribers := make([]chan string, 0, len(sessionSet))
	for current := range sessionSet {
		subscribers = append(subscribers, current)
	}
	delete(h.sessions, sessionID)
	h.mu.Unlock()

	for _, current := range subscribers {
		current <- reasonCode
	}
}

func (h *hub) broadcastIncidentExcept(incidentID uuid.UUID, message protocol.Message, excluded <-chan protocol.Message) {
	h.mu.Lock()
	subscribers := make([]chan protocol.Message, 0, len(h.incidentStreams[incidentID]))
	for subscriber := range h.incidentStreams[incidentID] {
		if excluded != nil && (<-chan protocol.Message)(subscriber) == excluded {
			continue
		}
		subscribers = append(subscribers, subscriber)
	}
	h.mu.Unlock()
	for _, subscriber := range subscribers {
		select {
		case subscriber <- message:
		default:
		}
	}
}

func (h *hub) prunePresenceLocked(incidentID uuid.UUID, now time.Time) {
	incidentPresences := h.presences[incidentID]
	if len(incidentPresences) == 0 {
		return
	}
	for connectionID, record := range incidentPresences {
		expiresAt, err := time.Parse(time.RFC3339Nano, record.ExpiresAt)
		if err != nil || !expiresAt.After(now) {
			delete(incidentPresences, connectionID)
		}
	}
	if len(incidentPresences) == 0 {
		delete(h.presences, incidentID)
	}
}

func cloneSheetRef(input map[string]string) map[string]string {
	if input == nil {
		return map[string]string{}
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
