package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultLimit     = 100
	MaxLimit         = 500
	MinimumRetention = 10 * time.Minute
)

const (
	ReasonCursorSnapshotUnavailable = "cursor_snapshot_unavailable"
	ReasonInvalidCursorToken        = "invalid_cursor_token"
	ReasonInvalidLimit              = "invalid_limit"
	ReasonPaginationNotSupported    = "pagination_not_supported"
	ReasonCursorQueryMismatch       = "cursor_query_mismatch"
)

var (
	ErrInvalidCursorToken    = errors.New("pagination: invalid cursor token")
	ErrCursorSnapshotExpired = errors.New("pagination: cursor snapshot unavailable")
)

type Query struct {
	CursorToken *string
	Limit       *int
}

type Binding struct {
	Route       string
	ActorUserID string
	Limit       int
	Scope       map[string]string
}

type Cursor struct {
	Route       string            `json:"route"`
	ActorUserID string            `json:"actor_user_id"`
	Limit       int               `json:"limit"`
	Scope       map[string]string `json:"scope,omitempty"`
	SnapshotID  string            `json:"snapshot_id"`
	Offset      int               `json:"offset"`
}

type Registry struct {
	mu        sync.Mutex
	now       func() time.Time
	retention time.Duration
	nextID    uint64
	snapshots map[string]snapshot
}

type snapshot struct {
	binding    Binding
	lastUsedAt time.Time
	rows       []json.RawMessage
}

func NewRegistry(options ...func(*Registry)) *Registry {
	registry := &Registry{
		now:       func() time.Time { return time.Now().UTC() },
		retention: MinimumRetention,
		snapshots: make(map[string]snapshot),
	}
	for _, option := range options {
		if option != nil {
			option(registry)
		}
	}
	if registry.now == nil {
		registry.now = func() time.Time { return time.Now().UTC() }
	}
	if registry.retention < MinimumRetention {
		registry.retention = MinimumRetention
	}
	return registry
}

func WithNow(now func() time.Time) func(*Registry) {
	return func(registry *Registry) {
		registry.now = now
	}
}

func WithRetention(retention time.Duration) func(*Registry) {
	return func(registry *Registry) {
		registry.retention = retention
	}
}

func ParseQuery(values url.Values) (Query, string) {
	for _, key := range []string{"page", "offset", "page_size", "block_size"} {
		if _, ok := values[key]; ok {
			return Query{}, ReasonPaginationNotSupported
		}
	}

	var query Query
	if value := values.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > MaxLimit {
			return Query{}, ReasonInvalidLimit
		}
		query.Limit = &parsed
	}
	if value := values.Get("cursor_token"); value != "" {
		query.CursorToken = &value
	}
	return query, ""
}

func (q Query) EffectiveLimit(cursor *Cursor) int {
	if q.Limit != nil {
		return *q.Limit
	}
	if cursor != nil {
		return cursor.Limit
	}
	return DefaultLimit
}

func ResolveRequest(values url.Values, route string, actorUserID string, scope map[string]string) (Binding, *Cursor, string) {
	query, reason := ParseQuery(values)
	if reason != "" {
		return Binding{}, nil, reason
	}

	if query.CursorToken == nil {
		return Binding{
			Route:       route,
			ActorUserID: actorUserID,
			Limit:       query.EffectiveLimit(nil),
			Scope:       cloneScope(scope),
		}, nil, ""
	}

	cursor, err := DecodeCursor(*query.CursorToken)
	if err != nil {
		return Binding{}, nil, ReasonInvalidCursorToken
	}

	binding := Binding{
		Route:       route,
		ActorUserID: actorUserID,
		Limit:       query.EffectiveLimit(&cursor),
		Scope:       cloneScope(scope),
	}
	if err := cursor.Validate(binding); err != nil {
		return Binding{}, nil, ReasonInvalidCursorToken
	}
	return binding, &cursor, ""
}

func ResolveViewQuery(query Query, route string, actorUserID string, scope map[string]string) (Binding, *Cursor, string) {
	if query.CursorToken == nil {
		return Binding{
			Route:       route,
			ActorUserID: actorUserID,
			Limit:       query.EffectiveLimit(nil),
			Scope:       cloneScope(scope),
		}, nil, ""
	}

	cursor, err := DecodeCursor(*query.CursorToken)
	if err != nil {
		return Binding{}, nil, ReasonInvalidCursorToken
	}

	binding := Binding{
		Route:       route,
		ActorUserID: actorUserID,
		Limit:       query.EffectiveLimit(&cursor),
		Scope:       cloneScope(scope),
	}
	if err := cursor.Validate(binding); err != nil {
		return Binding{}, nil, ReasonCursorQueryMismatch
	}
	return binding, &cursor, ""
}

func EncodeCursor(cursor Cursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(token string) (Cursor, error) {
	var cursor Cursor
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, err
	}
	if err := json.Unmarshal(data, &cursor); err != nil {
		return Cursor{}, err
	}
	if cursor.Route == "" || cursor.ActorUserID == "" || cursor.Limit < 1 || cursor.Limit > MaxLimit || cursor.SnapshotID == "" || cursor.Offset < 1 {
		return Cursor{}, ErrInvalidCursorToken
	}
	return cursor, nil
}

func (c Cursor) Validate(binding Binding) error {
	if c.Route != binding.Route || c.ActorUserID != binding.ActorUserID || c.Limit != binding.Limit {
		return ErrInvalidCursorToken
	}
	if !equalScope(c.Scope, binding.Scope) {
		return ErrInvalidCursorToken
	}
	return nil
}

func (r *Registry) Start(binding Binding, rows []json.RawMessage) ([]json.RawMessage, *Cursor) {
	if binding.Limit < 1 {
		return nil, nil
	}

	page := sliceRows(rows, 0, binding.Limit)
	if len(rows) <= binding.Limit {
		return page, nil
	}

	now := r.now().UTC()
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneExpiredLocked(now)
	r.nextID++
	snapshotID := encodeSnapshotID(r.nextID)
	r.snapshots[snapshotID] = snapshot{
		binding:    cloneBinding(binding),
		lastUsedAt: now,
		rows:       cloneRows(rows),
	}

	return page, &Cursor{
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       cloneScope(binding.Scope),
		SnapshotID:  snapshotID,
		Offset:      binding.Limit,
	}
}

func (r *Registry) Continue(binding Binding, cursor Cursor) ([]json.RawMessage, *Cursor, error) {
	now := r.now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneExpiredLocked(now)
	snapshot, ok := r.snapshots[cursor.SnapshotID]
	if !ok {
		return nil, nil, ErrCursorSnapshotExpired
	}
	if !equalBinding(snapshot.binding, binding) {
		delete(r.snapshots, cursor.SnapshotID)
		return nil, nil, ErrCursorSnapshotExpired
	}
	if cursor.Offset >= len(snapshot.rows)+1 {
		return nil, nil, ErrInvalidCursorToken
	}
	if cursor.Offset > len(snapshot.rows) {
		return nil, nil, ErrInvalidCursorToken
	}

	page := sliceRows(snapshot.rows, cursor.Offset, binding.Limit)
	snapshot.lastUsedAt = now
	r.snapshots[cursor.SnapshotID] = snapshot

	nextOffset := cursor.Offset + len(page)
	if nextOffset >= len(snapshot.rows) {
		return page, nil, nil
	}
	return page, &Cursor{
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       cloneScope(binding.Scope),
		SnapshotID:  cursor.SnapshotID,
		Offset:      nextOffset,
	}, nil
}

func MarshalResources(resources []map[string]any) ([]json.RawMessage, error) {
	rows := make([]json.RawMessage, 0, len(resources))
	for _, resource := range resources {
		payload, err := json.Marshal(resource)
		if err != nil {
			return nil, err
		}
		rows = append(rows, json.RawMessage(payload))
	}
	return rows, nil
}

func (r *Registry) pruneExpiredLocked(now time.Time) {
	for snapshotID, snapshot := range r.snapshots {
		if now.Sub(snapshot.lastUsedAt) > r.retention {
			delete(r.snapshots, snapshotID)
		}
	}
}

func encodeSnapshotID(id uint64) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.FormatUint(id, 10)))
}

func sliceRows(rows []json.RawMessage, start int, limit int) []json.RawMessage {
	if start < 0 || start >= len(rows) || limit < 1 {
		return []json.RawMessage{}
	}
	end := start + limit
	if end > len(rows) {
		end = len(rows)
	}
	page := make([]json.RawMessage, 0, end-start)
	for _, row := range rows[start:end] {
		page = append(page, append(json.RawMessage(nil), row...))
	}
	return page
}

func cloneRows(rows []json.RawMessage) []json.RawMessage {
	cloned := make([]json.RawMessage, 0, len(rows))
	for _, row := range rows {
		cloned = append(cloned, append(json.RawMessage(nil), row...))
	}
	return cloned
}

func cloneBinding(binding Binding) Binding {
	return Binding{
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       cloneScope(binding.Scope),
	}
}

func cloneScope(scope map[string]string) map[string]string {
	if len(scope) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(scope))
	for key, value := range scope {
		cloned[key] = value
	}
	return cloned
}

func equalBinding(left Binding, right Binding) bool {
	return left.Route == right.Route &&
		left.ActorUserID == right.ActorUserID &&
		left.Limit == right.Limit &&
		equalScope(left.Scope, right.Scope)
}

func equalScope(left map[string]string, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	if len(left) == 0 {
		return true
	}
	keys := make([]string, 0, len(left))
	for key := range left {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		value, ok := right[key]
		if !ok || left[key] != value {
			return false
		}
	}
	return true
}
