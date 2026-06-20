package pagination

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"slices"
	"strconv"
)

const (
	DefaultLimit = 100
	MaxLimit     = 500
)

const (
	CursorVersion = "pagination.cursor.v1"

	ModeOffset = "offset"
	ModeKeyset = "keyset"
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
	Version     string            `json:"version"`
	Mode        string            `json:"mode"`
	Route       string            `json:"route"`
	ActorUserID string            `json:"actor_user_id"`
	Limit       int               `json:"limit"`
	Scope       map[string]string `json:"scope,omitempty"`
	Position    map[string]string `json:"position,omitempty"`
}

type Codec struct {
	aead cipher.AEAD
}

func NewCodec(key []byte) *Codec {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Sprintf("pagination: invalid cursor key: %v", err))
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Sprintf("pagination: initialize cursor AEAD: %v", err))
	}
	return &Codec{aead: aead}
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

func (c *Codec) ResolveRequest(values url.Values, route string, actorUserID string, scope map[string]string) (Binding, *Cursor, string) {
	query, reason := ParseQuery(values)
	if reason != "" {
		return Binding{}, nil, reason
	}
	return c.ResolveQuery(query, route, actorUserID, scope, ReasonInvalidCursorToken)
}

func (c *Codec) ResolveListRequest(values url.Values, route string, actorUserID string, scope map[string]string) (Binding, *Cursor, string) {
	query, reason := ParseQuery(values)
	if reason != "" {
		return Binding{}, nil, reason
	}
	return c.ResolveQuery(query, route, actorUserID, scope, ReasonCursorQueryMismatch)
}

func (c *Codec) ResolveViewQuery(query Query, route string, actorUserID string, scope map[string]string) (Binding, *Cursor, string) {
	return c.ResolveQuery(query, route, actorUserID, scope, ReasonCursorQueryMismatch)
}

func (c *Codec) ResolveQuery(query Query, route string, actorUserID string, scope map[string]string, mismatchReason string) (Binding, *Cursor, string) {
	if query.CursorToken == nil {
		return Binding{
			Route:       route,
			ActorUserID: actorUserID,
			Limit:       query.EffectiveLimit(nil),
			Scope:       cloneScope(scope),
		}, nil, ""
	}

	cursor, err := c.Decode(*query.CursorToken)
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
		return Binding{}, nil, mismatchReason
	}
	return binding, &cursor, ""
}

func (c *Codec) Encode(cursor Cursor) (string, error) {
	cursor.Version = CursorVersion
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nonce, nonce, payload, []byte(CursorVersion))
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (c *Codec) Decode(token string) (Cursor, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Cursor{}, ErrInvalidCursorToken
	}
	nonceSize := c.aead.NonceSize()
	if len(sealed) <= nonceSize {
		return Cursor{}, ErrInvalidCursorToken
	}
	nonce := sealed[:nonceSize]
	ciphertext := sealed[nonceSize:]
	payload, err := c.aead.Open(nil, nonce, ciphertext, []byte(CursorVersion))
	if err != nil {
		return Cursor{}, ErrInvalidCursorToken
	}

	var cursor Cursor
	if err := json.Unmarshal(payload, &cursor); err != nil {
		return Cursor{}, ErrInvalidCursorToken
	}
	if cursor.Version != CursorVersion || cursor.Mode == "" || cursor.Route == "" || cursor.ActorUserID == "" || cursor.Limit < 1 || cursor.Limit > MaxLimit {
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

func NewOffsetCursor(binding Binding, offset int) *Cursor {
	if offset < 1 {
		return nil
	}
	return &Cursor{
		Version:     CursorVersion,
		Mode:        ModeOffset,
		Route:       binding.Route,
		ActorUserID: binding.ActorUserID,
		Limit:       binding.Limit,
		Scope:       cloneScope(binding.Scope),
		Position:    map[string]string{"offset": strconv.Itoa(offset)},
	}
}

func Offset(cursor *Cursor) (int, error) {
	if cursor == nil {
		return 0, nil
	}
	if cursor.Mode != ModeOffset {
		return 0, ErrInvalidCursorToken
	}
	value, ok := cursor.Position["offset"]
	if !ok {
		return 0, ErrInvalidCursorToken
	}
	offset, err := strconv.Atoi(value)
	if err != nil || offset < 1 {
		return 0, ErrInvalidCursorToken
	}
	return offset, nil
}

func PageRawMessages(binding Binding, cursor *Cursor, rows []json.RawMessage) ([]json.RawMessage, *Cursor, error) {
	offset, err := Offset(cursor)
	if err != nil {
		return nil, nil, err
	}
	page := SliceRows(rows, offset, binding.Limit)
	nextOffset := offset + len(page)
	if nextOffset >= len(rows) {
		return page, nil, nil
	}
	return page, NewOffsetCursor(binding, nextOffset), nil
}

func PageResources(binding Binding, cursor *Cursor, resources []map[string]any) ([]json.RawMessage, *Cursor, error) {
	rows, err := MarshalResources(resources)
	if err != nil {
		return nil, nil, err
	}
	return PageRawMessages(binding, cursor, rows)
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

func SliceRows(rows []json.RawMessage, start int, limit int) []json.RawMessage {
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
