package ws

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/google/uuid"
)

const (
	HeartbeatInterval       = 15 * time.Second
	HeartbeatTimeout        = 45 * time.Second
	PresenceTTL             = 45 * time.Second
	ResumeWindow            = 5 * time.Minute
	MinimumReplayRetention  = 10000
	ResumeStatusReplayed    = "replayed"
	ResumeStatusResetNeeded = "reset_required"
	IncidentTerminalClosed  = "incident_closed"
)

// Accept upgrades a request into a WebSocket connection using the configured
// public application origin as the only allowed cross-origin browser source.
func Accept(w http.ResponseWriter, r *http.Request, publicOrigin string) (*websocket.Conn, error) {
	options := &websocket.AcceptOptions{}
	if publicOrigin != "" {
		parsed, err := url.Parse(publicOrigin)
		if err != nil {
			http.Error(w, "invalid websocket origin configuration", http.StatusInternalServerError)
			return nil, err
		}
		options.OriginPatterns = []string{parsed.Scheme + "://" + parsed.Host}
	}
	return websocket.Accept(w, r, options)
}

func RejectUntrustedBrowserOrigin(w http.ResponseWriter, r *http.Request, publicOrigin string) bool {
	if _, err := r.Cookie("cartulary_session"); err != nil {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	if origin != publicOrigin {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return true
	}
	return false
}

type Message struct {
	Type       string          `json:"type"`
	IncidentID string          `json:"incident_id,omitempty"`
	EventID    string          `json:"event_id,omitempty"`
	EmittedAt  string          `json:"emitted_at,omitempty"`
	StreamSeq  *int64          `json:"stream_seq,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

func ReadJSON(ctx context.Context, conn *websocket.Conn, payload any) error {
	return wsjson.Read(ctx, conn, payload)
}

func WriteJSON(ctx context.Context, conn *websocket.Conn, payload any) error {
	return wsjson.Write(ctx, conn, payload)
}

func WriteThenClose(ctx context.Context, conn *websocket.Conn, payload any, status websocket.StatusCode, reason string) error {
	if err := WriteJSON(ctx, conn, payload); err != nil {
		return err
	}
	return conn.Close(status, reason)
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
	mu                    sync.Mutex
	sessions              map[uuid.UUID]map[chan string]struct{}
	incidentSessions      map[incidentSessionKey]map[chan string]struct{}
	incidentTerminals     map[uuid.UUID]map[chan string]struct{}
	recordChanges         []RecordChange
	recordSubscribers     map[chan RecordChange]struct{}
	incidentStreams       map[uuid.UUID]map[chan Message]struct{}
	replay                map[uuid.UUID][]replayEntry
	highWater             map[uuid.UUID]int64
	resumeTokens          map[string]ResumeToken
	presences             map[uuid.UUID]map[string]PresenceRecord
	serviceVersion        string
	activeConnections     atomic.Int64
	activeGaugeRegistered bool
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
	ChangeKind       string
	PatchCells       map[string]any
	StreamSeq        int64
	EventID          uuid.UUID
	EmittedAt        time.Time
}

const (
	JobScopeKindIncident   = "incident"
	JobScopeKindDeployment = "deployment"

	JobStatusQueued          = "queued"
	JobStatusRunning         = "running"
	JobStatusCancelRequested = "cancel_requested"
	JobStatusSucceeded       = "succeeded"
	JobStatusFailed          = "failed"
	JobStatusCanceled        = "canceled"
)

type JobScope struct {
	Kind       string `json:"kind"`
	IncidentID string `json:"incident_id,omitempty"`
}

type JobProgress struct {
	Completed int64  `json:"completed"`
	Total     *int64 `json:"total"`
}

type JobProgressPayload struct {
	JobID         string      `json:"job_id"`
	Scope         JobScope    `json:"scope"`
	Status        string      `json:"status"`
	Progress      JobProgress `json:"progress"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Cancelable    *bool       `json:"cancelable,omitempty"`
	Message       string      `json:"message,omitempty"`
	ResultSummary any         `json:"result_summary,omitempty"`
	ErrorSummary  any         `json:"error_summary,omitempty"`
	RetainedUntil *time.Time  `json:"retained_until,omitempty"`
}

type ResumeToken struct {
	Token            string
	SessionID        uuid.UUID
	IncidentID       uuid.UUID
	ClientInstanceID string
	ExpiresAt        time.Time
}

type PresenceInput struct {
	SheetRef map[string]string `json:"sheet_ref"`
	RecordID *string           `json:"record_id,omitempty"`
	FieldKey *string           `json:"field_key,omitempty"`
	Mode     string            `json:"mode"`
}

type PresenceRecord struct {
	ConnectionID string            `json:"connection_id"`
	UserID       string            `json:"user_id"`
	DisplayName  string            `json:"display_name"`
	SheetRef     map[string]string `json:"sheet_ref"`
	RecordID     *string           `json:"record_id,omitempty"`
	FieldKey     *string           `json:"field_key,omitempty"`
	Mode         string            `json:"mode"`
	ObservedAt   string            `json:"observed_at"`
	ExpiresAt    string            `json:"expires_at"`
}

type replayEntry struct {
	Message  Message
	StoredAt time.Time
}

type incidentSessionKey struct {
	IncidentID uuid.UUID
	SessionID  uuid.UUID
}

func NewHub() *Hub {
	return &Hub{
		sessions:          make(map[uuid.UUID]map[chan string]struct{}),
		incidentSessions:  make(map[incidentSessionKey]map[chan string]struct{}),
		incidentTerminals: make(map[uuid.UUID]map[chan string]struct{}),
		recordSubscribers: make(map[chan RecordChange]struct{}),
		incidentStreams:   make(map[uuid.UUID]map[chan Message]struct{}),
		replay:            make(map[uuid.UUID][]replayEntry),
		highWater:         make(map[uuid.UUID]int64),
		resumeTokens:      make(map[string]ResumeToken),
		presences:         make(map[uuid.UUID]map[string]PresenceRecord),
	}
}

func (h *Hub) PublishRecordChange(change RecordChange) {
	if h == nil {
		return
	}
	finishTelemetry := h.startEventSend("record_changed")
	defer finishTelemetry("success", "")

	now := time.Now().UTC()
	h.mu.Lock()
	h.highWater[change.IncidentID]++
	change.StreamSeq = h.highWater[change.IncidentID]
	change.EventID = uuid.New()
	change.EmittedAt = now
	h.recordChanges = append(h.recordChanges, change)
	message := recordChangedMessage(change)
	h.replay[change.IncidentID] = append(h.replay[change.IncidentID], replayEntry{Message: message, StoredAt: now})
	h.pruneReplayLocked(change.IncidentID, now)
	recordSubscribers := make([]chan RecordChange, 0, len(h.recordSubscribers))
	for subscriber := range h.recordSubscribers {
		recordSubscribers = append(recordSubscribers, subscriber)
	}
	streamSubscribers := make([]chan Message, 0, len(h.incidentStreams[change.IncidentID]))
	for subscriber := range h.incidentStreams[change.IncidentID] {
		streamSubscribers = append(streamSubscribers, subscriber)
	}
	h.mu.Unlock()

	for _, subscriber := range recordSubscribers {
		select {
		case subscriber <- change:
		default:
		}
	}
	for _, subscriber := range streamSubscribers {
		select {
		case subscriber <- message:
		default:
		}
	}
}

func (h *Hub) PublishJobProgress(incidentID uuid.UUID, payload JobProgressPayload) error {
	if h == nil {
		return nil
	}
	if err := ValidateIncidentJobProgressPayload(incidentID, payload); err != nil {
		h.recordEventSend("job_progress", "rejected", "")
		return err
	}
	finishTelemetry := h.startEventSend("job_progress")
	defer finishTelemetry("success", "")

	now := time.Now().UTC()
	h.mu.Lock()
	message := h.nextReplayableMessageLocked(incidentID, "job_progress", payload, now)
	h.replay[incidentID] = append(h.replay[incidentID], replayEntry{Message: message, StoredAt: now})
	h.pruneReplayLocked(incidentID, now)
	streamSubscribers := make([]chan Message, 0, len(h.incidentStreams[incidentID]))
	for subscriber := range h.incidentStreams[incidentID] {
		streamSubscribers = append(streamSubscribers, subscriber)
	}
	h.mu.Unlock()

	for _, subscriber := range streamSubscribers {
		select {
		case subscriber <- message:
		default:
		}
	}
	return nil
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
	h.recordSubscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.recordSubscribers, ch)
		h.mu.Unlock()
		close(ch)
	}
}

func (h *Hub) SubscribeIncident(incidentID uuid.UUID, buffer int) (<-chan Message, func()) {
	if h == nil {
		ch := make(chan Message)
		close(ch)
		return ch, func() {}
	}
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan Message, buffer)
	h.mu.Lock()
	subscribers := h.incidentStreams[incidentID]
	if subscribers == nil {
		subscribers = make(map[chan Message]struct{})
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
		close(ch)
	}
}

func (h *Hub) IssueResumeToken(sessionID uuid.UUID, incidentID uuid.UUID, clientInstanceID string, sessionExpiresAt time.Time, now time.Time) (string, time.Time, error) {
	token, err := opaqueToken()
	if err != nil {
		return "", time.Time{}, err
	}
	expiresAt := now.Add(ResumeWindow)
	if sessionExpiresAt.Before(expiresAt) {
		expiresAt = sessionExpiresAt
	}
	record := ResumeToken{
		Token:            token,
		SessionID:        sessionID,
		IncidentID:       incidentID,
		ClientInstanceID: clientInstanceID,
		ExpiresAt:        expiresAt,
	}
	h.mu.Lock()
	h.resumeTokens[token] = record
	h.mu.Unlock()
	return token, expiresAt, nil
}

func (h *Hub) ReplayMessages(sessionID uuid.UUID, incidentID uuid.UUID, clientInstanceID string, token string, lastSeenStreamSeq int64, now time.Time) (string, []Message, int64) {
	if h == nil {
		return ResumeStatusResetNeeded, nil, 0
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pruneReplayLocked(incidentID, now)
	highWater := h.highWater[incidentID]
	record, ok := h.resumeTokens[token]
	if !ok || record.SessionID != sessionID || record.IncidentID != incidentID || record.ClientInstanceID != clientInstanceID || !record.ExpiresAt.After(now) {
		return ResumeStatusResetNeeded, nil, highWater
	}
	if lastSeenStreamSeq > highWater {
		return ResumeStatusResetNeeded, nil, highWater
	}
	entries := h.replay[incidentID]
	firstSeq := int64(0)
	for _, entry := range entries {
		if !IsReplayableMessageType(entry.Message.Type) || entry.Message.StreamSeq == nil {
			continue
		}
		firstSeq = *entry.Message.StreamSeq
		break
	}
	if firstSeq > 0 && lastSeenStreamSeq < firstSeq-1 {
		return ResumeStatusResetNeeded, nil, highWater
	}
	missed := make([]Message, 0)
	for _, entry := range entries {
		if !IsReplayableMessageType(entry.Message.Type) || entry.Message.StreamSeq == nil {
			continue
		}
		if *entry.Message.StreamSeq > lastSeenStreamSeq {
			missed = append(missed, entry.Message)
		}
	}
	return ResumeStatusReplayed, missed, highWater
}

func IsReplayableMessageType(messageType string) bool {
	switch messageType {
	case "record_changed", "job_progress":
		return true
	default:
		return false
	}
}

func (h *Hub) HighWater(incidentID uuid.UUID) int64 {
	if h == nil {
		return 0
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.highWater[incidentID]
}

func (h *Hub) UpsertPresence(incidentID uuid.UUID, connectionID uuid.UUID, userID uuid.UUID, displayName string, input PresenceInput, now time.Time) PresenceRecord {
	record := PresenceRecord{
		ConnectionID: connectionID.String(),
		UserID:       userID.String(),
		DisplayName:  displayName,
		SheetRef:     cloneSheetRef(input.SheetRef),
		RecordID:     cloneStringPointer(input.RecordID),
		FieldKey:     cloneStringPointer(input.FieldKey),
		Mode:         input.Mode,
		ObservedAt:   now.UTC().Format(time.RFC3339Nano),
		ExpiresAt:    now.UTC().Add(PresenceTTL).Format(time.RFC3339Nano),
	}
	h.mu.Lock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	if incidentPresences == nil {
		incidentPresences = make(map[string]PresenceRecord)
		h.presences[incidentID] = incidentPresences
	}
	incidentPresences[record.ConnectionID] = record
	h.mu.Unlock()
	return record
}

func (h *Hub) RemovePresence(incidentID uuid.UUID, connectionID uuid.UUID, now time.Time) (PresenceRecord, bool) {
	if h == nil {
		return PresenceRecord{}, false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	if incidentPresences == nil {
		return PresenceRecord{}, false
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

func (h *Hub) PresenceSnapshot(incidentID uuid.UUID, now time.Time) []PresenceRecord {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.prunePresenceLocked(incidentID, now)
	incidentPresences := h.presences[incidentID]
	presences := make([]PresenceRecord, 0, len(incidentPresences))
	for _, record := range incidentPresences {
		presences = append(presences, record)
	}
	sort.Slice(presences, func(i, j int) bool {
		return presences[i].ConnectionID < presences[j].ConnectionID
	})
	return presences
}

func (h *Hub) BroadcastPresenceDelta(incidentID uuid.UUID, kind string, presence PresenceRecord, now time.Time) {
	if h == nil {
		return
	}
	finishTelemetry := h.startEventSend("presence_delta")
	defer finishTelemetry("success", "")
	message := EphemeralMessage(incidentID, "presence_delta", map[string]any{
		"delta_kind": kind,
		"presence":   presence,
	}, now)
	h.broadcastIncident(incidentID, message)
}

func (h *Hub) RegisterIncidentSession(incidentID uuid.UUID, sessionID uuid.UUID) (<-chan string, func()) {
	if h == nil {
		return nil, func() {}
	}
	ch := make(chan string, 1)
	key := incidentSessionKey{IncidentID: incidentID, SessionID: sessionID}
	h.mu.Lock()
	subscribers := h.incidentSessions[key]
	if subscribers == nil {
		subscribers = make(map[chan string]struct{})
		h.incidentSessions[key] = subscribers
	}
	subscribers[ch] = struct{}{}
	h.mu.Unlock()
	return ch, func() {
		h.mu.Lock()
		if subscribers := h.incidentSessions[key]; subscribers != nil {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(h.incidentSessions, key)
			}
		}
		h.mu.Unlock()
	}
}

func (h *Hub) RegisterIncidentTerminal(incidentID uuid.UUID) (<-chan string, func()) {
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

func (h *Hub) TrackActiveConnection() func() {
	if h == nil {
		return func() {}
	}
	h.activeConnections.Add(1)
	return func() {
		h.activeConnections.Add(-1)
	}
}

func (h *Hub) ActiveConnections() int64 {
	if h == nil {
		return 0
	}
	count := h.activeConnections.Load()
	if count < 0 {
		return 0
	}
	return count
}

func (h *Hub) RevokeIncidentSession(incidentID uuid.UUID, sessionID uuid.UUID, reasonCode string) {
	if h == nil {
		return
	}
	key := incidentSessionKey{IncidentID: incidentID, SessionID: sessionID}
	h.mu.Lock()
	sessionSet := h.incidentSessions[key]
	subscribers := make([]chan string, 0, len(sessionSet))
	for current := range sessionSet {
		subscribers = append(subscribers, current)
	}
	delete(h.incidentSessions, key)
	h.mu.Unlock()
	for _, current := range subscribers {
		current <- reasonCode
	}
}

func (h *Hub) RevokeIncidentAccess(incidentID uuid.UUID, sessionID uuid.UUID) {
	h.RevokeIncidentSession(incidentID, sessionID, "incident_access_revoked")
}

func (h *Hub) TerminateIncident(incidentID uuid.UUID, reasonCode string) {
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

func (h *Hub) RegisterSession(sessionID uuid.UUID) (<-chan string, func()) {
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

func (h *Hub) broadcastIncident(incidentID uuid.UUID, message Message) {
	h.mu.Lock()
	subscribers := make([]chan Message, 0, len(h.incidentStreams[incidentID]))
	for subscriber := range h.incidentStreams[incidentID] {
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

func (h *Hub) nextReplayableMessageLocked(incidentID uuid.UUID, messageType string, payload any, now time.Time) Message {
	h.highWater[incidentID]++
	streamSeq := h.highWater[incidentID]
	return Message{
		Type:       messageType,
		IncidentID: incidentID.String(),
		EventID:    uuid.New().String(),
		EmittedAt:  now.UTC().Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    RawPayload(nonNilPayload(payload)),
	}
}

func (h *Hub) pruneReplayLocked(incidentID uuid.UUID, now time.Time) {
	entries := h.replay[incidentID]
	if len(entries) <= MinimumReplayRetention {
		return
	}
	cutoff := now.Add(-ResumeWindow)
	keepFrom := 0
	for keepFrom < len(entries)-MinimumReplayRetention && entries[keepFrom].StoredAt.Before(cutoff) {
		keepFrom++
	}
	if keepFrom > 0 {
		h.replay[incidentID] = append([]replayEntry(nil), entries[keepFrom:]...)
	}
}

func (h *Hub) prunePresenceLocked(incidentID uuid.UUID, now time.Time) {
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

func EphemeralMessage(incidentID uuid.UUID, messageType string, payload any, now time.Time) Message {
	return Message{
		Type:       messageType,
		IncidentID: incidentID.String(),
		EventID:    uuid.New().String(),
		EmittedAt:  now.UTC().Format(time.RFC3339Nano),
		Payload:    RawPayload(nonNilPayload(payload)),
	}
}

func recordChangedMessage(change RecordChange) Message {
	streamSeq := change.StreamSeq
	return Message{
		Type:       "record_changed",
		IncidentID: change.IncidentID.String(),
		EventID:    change.EventID.String(),
		EmittedAt:  change.EmittedAt.UTC().Format(time.RFC3339Nano),
		StreamSeq:  &streamSeq,
		Payload:    RawPayload(RecordChangePayload(change)),
	}
}

func RecordChangePayload(change RecordChange) map[string]any {
	changedKeys := append([]string(nil), change.ChangedFieldKeys...)
	slices.Sort(changedKeys)
	changedKeys = slices.Compact(changedKeys)
	changeKind := change.ChangeKind
	if changeKind == "" {
		changeKind = "invalidate"
	}
	affectedView := map[string]any{
		"view_schema_id": change.ViewSchemaID,
		"change_kind":    changeKind,
	}
	if change.PatchCells != nil {
		affectedView = map[string]any{
			"view_schema_id": change.ViewSchemaID,
			"change_kind":    "patch",
			"patch_cells":    change.PatchCells,
		}
	}
	return map[string]any{
		"record_id":          change.RecordID.String(),
		"row_version":        change.RowVersion,
		"change_set_id":      change.ChangeSetID.String(),
		"client_txn_id":      change.ClientTxnID,
		"actor_user_id":      change.ActorUserID.String(),
		"changed_field_keys": changedKeys,
		"affected_views":     []map[string]any{affectedView},
	}
}

func BuildViewRowPatch(row map[string]any, changedFieldKeys []string) map[string]any {
	if row == nil {
		return nil
	}
	recordID, ok := row["record_id"]
	if !ok {
		return nil
	}
	rowVersion, ok := row["row_version"]
	if !ok {
		return nil
	}
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil
	}

	changed := make(map[string]struct{}, len(changedFieldKeys))
	for _, fieldKey := range changedFieldKeys {
		changed[fieldKey] = struct{}{}
	}

	patchCells := make(map[string]any, len(changed))
	for fieldKey := range changed {
		if cell, ok := cells[fieldKey]; ok {
			patchCells[fieldKey] = cell
		}
	}
	if len(patchCells) == 0 {
		return nil
	}

	patch := map[string]any{
		"record_id":   recordID,
		"row_version": rowVersion,
		"cells":       patchCells,
	}
	if groupValues, ok := row["group_values"].(map[string]any); ok {
		patchGroupValues := make(map[string]any)
		for fieldKey := range changed {
			if value, ok := groupValues[fieldKey]; ok {
				patchGroupValues[fieldKey] = value
			}
		}
		if len(patchGroupValues) > 0 {
			patch["group_values"] = patchGroupValues
		}
	}
	return patch
}

func NewIncidentJobProgressPayload(jobID string, incidentID uuid.UUID, status string, progress JobProgress, updatedAt time.Time) JobProgressPayload {
	return JobProgressPayload{
		JobID: jobID,
		Scope: JobScope{
			Kind:       JobScopeKindIncident,
			IncidentID: incidentID.String(),
		},
		Status:    status,
		Progress:  progress,
		UpdatedAt: updatedAt.UTC(),
	}
}

func ValidateIncidentJobProgressPayload(incidentID uuid.UUID, payload JobProgressPayload) error {
	if strings.TrimSpace(payload.JobID) == "" {
		return fmt.Errorf("job_progress.job_id is required")
	}
	if payload.Scope.Kind != JobScopeKindIncident {
		return fmt.Errorf("job_progress.scope.kind must be incident")
	}
	scopeIncidentID, err := uuid.Parse(payload.Scope.IncidentID)
	if err != nil {
		return fmt.Errorf("job_progress.scope.incident_id is invalid")
	}
	if scopeIncidentID != incidentID {
		return fmt.Errorf("job_progress.scope.incident_id must match envelope incident_id")
	}
	switch payload.Status {
	case JobStatusQueued, JobStatusRunning, JobStatusCancelRequested, JobStatusSucceeded, JobStatusFailed, JobStatusCanceled:
	default:
		return fmt.Errorf("job_progress.status is invalid")
	}
	if payload.Progress.Completed < 0 {
		return fmt.Errorf("job_progress.progress.completed must be non-negative")
	}
	if payload.Progress.Total != nil {
		if *payload.Progress.Total <= 0 {
			return fmt.Errorf("job_progress.progress.total must be positive or null")
		}
		if payload.Progress.Completed > *payload.Progress.Total {
			return fmt.Errorf("job_progress.progress.completed must not exceed total")
		}
	}
	if payload.UpdatedAt.IsZero() {
		return fmt.Errorf("job_progress.updated_at is required")
	}
	return nil
}

func PresenceSnapshotMessage(incidentID uuid.UUID, presences []PresenceRecord, now time.Time) Message {
	if presences == nil {
		presences = []PresenceRecord{}
	}
	return EphemeralMessage(incidentID, "presence_snapshot", map[string]any{"presences": presences}, now)
}

func ValidatePresenceInput(input PresenceInput) error {
	if input.SheetRef == nil || input.SheetRef["kind"] == "" || input.SheetRef["id"] == "" {
		return fmt.Errorf("presence.sheet_ref is required")
	}
	switch input.SheetRef["kind"] {
	case "view_schema", "saved_view":
	default:
		return fmt.Errorf("presence.sheet_ref.kind is invalid")
	}
	switch input.Mode {
	case "viewing", "editing", "idle":
	default:
		return fmt.Errorf("presence.mode is invalid")
	}
	if input.FieldKey != nil && input.Mode != "editing" {
		return fmt.Errorf("presence.field_key requires editing mode")
	}
	return nil
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

func nonNilPayload(payload any) any {
	if payload == nil {
		return map[string]any{}
	}
	return payload
}

func opaqueToken() (string, error) {
	var data [32]byte
	if _, err := rand.Read(data[:]); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}
