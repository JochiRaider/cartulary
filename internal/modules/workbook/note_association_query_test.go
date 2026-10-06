package workbook

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/google/uuid"
)

func TestNoteAssociationListRequestAdmission_Unit(t *testing.T) {
	actor := uuid.New()
	target := RecordTarget{RecordID: uuid.New(), IncidentID: uuid.New()}
	key := [32]byte([]byte("01234567890123456789012345678901"))
	codec := pagination.NewCodec(key)
	for _, test := range []struct{ name, raw, code, reason string }{
		{"bad encoding precedes duplicates", "kind=source&x=1&x=2&bad=%zz", "invalid_list_query", "malformed_query"},
		{"bad UTF8 name precedes duplicates", "kind=source&x=1&x=2&%ff=value", "invalid_list_query", "malformed_query"},
		{"bad UTF8 value", "kind=%ff", "invalid_list_query", "malformed_query"},
		{"invalid separator", "kind=source;limit=1", "invalid_list_query", "malformed_query"},
		{"duplicate unknown precedes alias", "x=1&x=2&page=1", "invalid_list_query", "duplicate_query_member"},
		{"duplicate decoded member", "kind=source&%6bind=evidence", "invalid_list_query", "duplicate_query_member"},
		{"duplicate limit", "kind=source&limit=1&limit=1", "invalid_list_query", "duplicate_query_member"},
		{"duplicate cursor", "kind=source&cursor_token=a&cursor_token=b", "invalid_list_query", "duplicate_query_member"},
		{"page precedes unknown", "unknown=1&page=1", "invalid_pagination_request", "invalid_limit"},
		{"offset alias", "kind=source&offset=0", "invalid_pagination_request", "invalid_limit"},
		{"page size alias", "kind=source&page_size=1", "invalid_pagination_request", "invalid_limit"},
		{"block size alias", "kind=source&block_size=1", "invalid_pagination_request", "invalid_limit"},
		{"unknown precedes kind", "unknown=1", "invalid_list_query", "unknown_query_member"},
		{"missing kind", "", "invalid_list_query", "invalid_filter_value"},
		{"empty kind", "kind=", "invalid_list_query", "invalid_filter_value"},
		{"kind precedes limit", "kind=future&limit=0", "invalid_list_query", "invalid_filter_value"},
		{"kind is case sensitive", "kind=Source", "invalid_list_query", "invalid_filter_value"},
		{"kind is not trimmed", "kind=source+", "invalid_list_query", "invalid_filter_value"},
		{"kind is scalar", "kind=source,evidence", "invalid_list_query", "invalid_filter_value"},
		{"empty limit", "kind=source&limit=", "invalid_pagination_request", "invalid_limit"},
		{"noninteger limit", "kind=source&limit=1.5", "invalid_pagination_request", "invalid_limit"},
		{"zero limit precedes cursor", "kind=source&limit=0&cursor_token=bad", "invalid_pagination_request", "invalid_limit"},
		{"large limit", "kind=source&limit=501", "invalid_pagination_request", "invalid_limit"},
		{"overflow limit", "kind=source&limit=999999999999999999999999", "invalid_pagination_request", "invalid_limit"},
		{"empty cursor", "kind=source&cursor_token=", "invalid_pagination_request", "invalid_cursor_token"},
		{"malformed cursor", "kind=source&cursor_token=bad", "invalid_pagination_request", "invalid_cursor_token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, failure := decodeNoteAssociationListRequest(test.raw, actor, target, codec)
			if failure == nil || failure.Status != 400 || failure.Code != test.code || !reflect.DeepEqual(failure.Details, map[string]any{"reason_code": test.reason}) {
				t.Fatalf("error = %#v, want 400 %s with only reason %s", failure, test.code, test.reason)
			}
		})
	}
	for _, test := range []struct {
		raw   string
		limit int
	}{
		{"kind=source", 100}, {"kind=evidence&limit=1", 1}, {"kind=related_note&limit=500", 500},
	} {
		t.Run(test.raw, func(t *testing.T) {
			got, failure := decodeNoteAssociationListRequest(test.raw, actor, target, codec)
			if failure != nil || got.query.Limit != test.limit || got.query.BeforeTime != nil || got.query.BeforeID != uuid.Nil {
				t.Fatalf("first page = %#v, error %#v", got, failure)
			}
		})
	}
	stamp := time.Date(2026, 10, 2, 10, 20, 30, 123, time.UTC)
	linkID := uuid.New()
	newCursor := func() pagination.Cursor {
		return pagination.Cursor{Mode: pagination.ModeKeyset, Route: "workbook.note-associations", ActorUserID: actor.String(), Limit: 1,
			Scope:    map[string]string{"incident": target.IncidentID.String(), "note": target.RecordID.String(), "kind": "source", "order": "created_at_desc_link_id_desc"},
			Position: map[string]string{"created_at": stamp.Format(time.RFC3339Nano), "link_id": linkID.String()}}
	}
	for _, test := range []struct {
		name   string
		mutate func(*pagination.Cursor)
		query  string
		reason string
	}{
		{"reuse omitted limit", func(*pagination.Cursor) {}, "kind=source", ""},
		{"matching explicit limit", func(*pagination.Cursor) {}, "kind=source&limit=1", ""},
		{"changed limit", func(*pagination.Cursor) {}, "kind=source&limit=2", "cursor_query_mismatch"},
		{"changed kind", func(*pagination.Cursor) {}, "kind=evidence", "cursor_query_mismatch"},
		{"actor precedes query mismatch", func(c *pagination.Cursor) { c.ActorUserID = uuid.NewString(); c.Scope["kind"] = "evidence" }, "kind=source", "invalid_cursor_token"},
		{"different route", func(c *pagination.Cursor) { c.Route = "other.list" }, "kind=source", "invalid_cursor_token"},
		{"different incident", func(c *pagination.Cursor) { c.Scope["incident"] = uuid.NewString() }, "kind=source", "cursor_query_mismatch"},
		{"different Note", func(c *pagination.Cursor) { c.Scope["note"] = uuid.NewString() }, "kind=source", "cursor_query_mismatch"},
		{"different order", func(c *pagination.Cursor) { c.Scope["order"] = "ascending" }, "kind=source", "cursor_query_mismatch"},
		{"extra scope", func(c *pagination.Cursor) { c.Scope["future"] = "value" }, "kind=source", "cursor_query_mismatch"},
		{"query precedes position", func(c *pagination.Cursor) { c.Scope["kind"] = "evidence"; c.Position["link_id"] = "bad" }, "kind=source", "cursor_query_mismatch"},
		{"wrong mode", func(c *pagination.Cursor) { c.Mode = pagination.ModeOffset }, "kind=source", "invalid_cursor_token"},
		{"bad time", func(c *pagination.Cursor) { c.Position["created_at"] = "bad" }, "kind=source", "invalid_cursor_token"},
		{"bad ID", func(c *pagination.Cursor) { c.Position["link_id"] = "bad" }, "kind=source", "invalid_cursor_token"},
		{"zero ID", func(c *pagination.Cursor) { c.Position["link_id"] = uuid.Nil.String() }, "kind=source", "invalid_cursor_token"},
		{"extra position", func(c *pagination.Cursor) { c.Position["future"] = "value" }, "kind=source", "invalid_cursor_token"},
		{"missing position", func(c *pagination.Cursor) { delete(c.Position, "created_at") }, "kind=source", "invalid_cursor_token"},
	} {
		t.Run(test.name, func(t *testing.T) {
			cursor := newCursor()
			test.mutate(&cursor)
			token, err := codec.Encode(cursor)
			if err != nil {
				t.Fatal(err)
			}
			got, failure := decodeNoteAssociationListRequest(test.query+"&cursor_token="+url.QueryEscape(token), actor, target, codec)
			if test.reason != "" {
				if failure == nil || failure.Status != 400 || failure.Code != "invalid_pagination_request" || !reflect.DeepEqual(failure.Details, map[string]any{"reason_code": test.reason}) {
					t.Fatalf("error = %#v, want only reason %s", failure, test.reason)
				}
				return
			}
			if failure != nil || got.query.Limit != 1 || got.query.BeforeTime == nil || !got.query.BeforeTime.Equal(stamp) || got.query.BeforeID != linkID {
				t.Fatalf("continuation = %#v, error %#v", got, failure)
			}
		})
	}
	// Construct authenticated but unsupported-version input independently of
	// Codec.Encode, which always issues the supported version.
	master, err := cryptography.AdmitKey(key[:])
	if err != nil {
		t.Fatal(err)
	}
	cursor := newCursor()
	cursor.Version = "pagination.cursor.future"
	payload, err := json.Marshal(cursor)
	if err != nil {
		t.Fatal(err)
	}
	sealed, err := master.Seal(pagination.CursorVersion, nil, payload, []byte("pc2."))
	if err != nil {
		t.Fatal(err)
	}
	for _, tamper := range []bool{false, true} {
		t.Run(map[bool]string{false: "unsupported version", true: "failed token authentication"}[tamper], func(t *testing.T) {
			input := append([]byte(nil), sealed...)
			if tamper {
				input[len(input)-1] ^= 1
			}
			_, failure := decodeNoteAssociationListRequest("kind=source&cursor_token=pc2."+base64.RawURLEncoding.EncodeToString(input), actor, target, codec)
			if failure == nil || failure.Code != "invalid_pagination_request" || !reflect.DeepEqual(failure.Details, map[string]any{"reason_code": "invalid_cursor_token"}) {
				t.Fatalf("unsafe token failure = %#v", failure)
			}
		})
	}
}
