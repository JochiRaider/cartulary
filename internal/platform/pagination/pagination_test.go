package pagination

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"
)

func testCodec() *Codec {
	return NewCodec([]byte("01234567890123456789012345678901"))
}

func TestResolveRequestReusesCursorBoundLimitAndValidatesBindings(t *testing.T) {
	codec := testCodec()
	cursorToken, err := codec.Encode(Cursor{
		Mode:        ModeOffset,
		Route:       "incidents.list",
		ActorUserID: "user-1",
		Limit:       25,
		Scope:       map[string]string{"incident_id": "incident-1"},
		Position:    map[string]string{"offset": "25"},
	})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	binding, cursor, reason := codec.ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
	}, "incidents.list", "user-1", map[string]string{"incident_id": "incident-1"})
	if reason != "" {
		t.Fatalf("unexpected resolve reason: %s", reason)
	}
	if cursor == nil {
		t.Fatal("expected decoded cursor")
	}
	if binding.Limit != 25 || binding.Route != "incidents.list" || binding.ActorUserID != "user-1" {
		t.Fatalf("unexpected resolved binding: %#v", binding)
	}

	_, _, reason = codec.ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
		"limit":        []string{"50"},
	}, "incidents.list", "user-1", map[string]string{"incident_id": "incident-1"})
	if reason != ReasonInvalidCursorToken {
		t.Fatalf("expected invalid cursor for mismatched explicit limit, got %q", reason)
	}

	_, _, reason = codec.ResolveRequest(url.Values{
		"cursor_token": []string{cursorToken},
	}, "incidents.list", "user-2", map[string]string{"incident_id": "incident-1"})
	if reason != ReasonInvalidCursorToken {
		t.Fatalf("expected invalid cursor for mismatched actor, got %q", reason)
	}

	_, _, reason = codec.ResolveRequest(url.Values{
		"page": []string{"2"},
	}, "incidents.list", "user-1", nil)
	if reason != ReasonPaginationNotSupported {
		t.Fatalf("expected pagination_not_supported, got %q", reason)
	}

	_, _, reason = codec.ResolveRequest(url.Values{
		"limit": []string{"0"},
	}, "incidents.list", "user-1", nil)
	if reason != ReasonInvalidLimit {
		t.Fatalf("expected invalid_limit, got %q", reason)
	}
}

func TestResolveViewQueryReportsContractMismatch(t *testing.T) {
	codec := testCodec()
	cursorToken, err := codec.Encode(Cursor{
		Mode:        ModeOffset,
		Route:       "workbook.view-query",
		ActorUserID: "user-1",
		Limit:       2,
		Scope: map[string]string{
			"incident_id":    "incident-1",
			"view_schema_id": "cartulary.view.timeline.v1",
			"query_contract": `{"filters":[],"sort":[{"field_key":"record_id","direction":"asc"}]}`,
		},
		Position: map[string]string{"offset": "2"},
	})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	_, cursor, reason := codec.ResolveViewQuery(Query{CursorToken: &cursorToken}, "workbook.view-query", "user-1", map[string]string{
		"incident_id":    "incident-1",
		"view_schema_id": "cartulary.view.timeline.v1",
		"query_contract": `{"filters":[],"sort":[{"field_key":"record_id","direction":"asc"}]}`,
	})
	if reason != "" || cursor == nil {
		t.Fatalf("expected successful view query cursor resolve, reason=%q cursor=%#v", reason, cursor)
	}

	_, _, reason = codec.ResolveViewQuery(Query{CursorToken: &cursorToken}, "workbook.view-query", "user-1", map[string]string{
		"incident_id":    "incident-1",
		"view_schema_id": "cartulary.view.timeline.v1",
		"query_contract": `{"filters":[],"sort":[{"field_key":"timeline.summary","direction":"asc"}]}`,
	})
	if reason != ReasonCursorQueryMismatch {
		t.Fatalf("expected query mismatch for changed normalized query, got %q", reason)
	}

	changedLimit := 3
	_, _, reason = codec.ResolveViewQuery(Query{CursorToken: &cursorToken, Limit: &changedLimit}, "workbook.view-query", "user-1", map[string]string{
		"incident_id":    "incident-1",
		"view_schema_id": "cartulary.view.timeline.v1",
		"query_contract": `{"filters":[],"sort":[{"field_key":"record_id","direction":"asc"}]}`,
	})
	if reason != ReasonCursorQueryMismatch {
		t.Fatalf("expected query mismatch for changed limit, got %q", reason)
	}
}

func TestCodecRejectsTamperedCursor(t *testing.T) {
	codec := testCodec()
	token, err := codec.Encode(Cursor{
		Mode:        ModeOffset,
		Route:       "users.list",
		ActorUserID: "admin-1",
		Limit:       1,
		Position:    map[string]string{"offset": "1"},
	})
	if err != nil {
		t.Fatalf("encode cursor: %v", err)
	}

	payloadToken, signatureToken, ok := strings.Cut(token, ".")
	if !ok {
		t.Fatalf("unexpected token shape %q", token)
	}
	payload, err := base64.RawURLEncoding.DecodeString(payloadToken)
	if err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	decoded["actor_user_id"] = "admin-2"
	tamperedPayload, err := json.Marshal(decoded)
	if err != nil {
		t.Fatalf("marshal tampered payload: %v", err)
	}
	tamperedToken := base64.RawURLEncoding.EncodeToString(tamperedPayload) + "." + signatureToken
	if _, err := codec.Decode(tamperedToken); err == nil {
		t.Fatal("expected tampered payload to fail signature validation")
	}

	tamperedSignature := token[:len(token)-1] + "A"
	if _, err := codec.Decode(tamperedSignature); err == nil {
		t.Fatal("expected tampered signature to fail validation")
	}
}

func TestPageRawMessagesUsesSignedOffsetCursorWithoutRetainingRows(t *testing.T) {
	binding := Binding{Route: "users.list", ActorUserID: "admin-1", Limit: 2}
	rows := []json.RawMessage{
		json.RawMessage(`{"user_id":"1"}`),
		json.RawMessage(`{"user_id":"2"}`),
		json.RawMessage(`{"user_id":"3"}`),
	}
	pageOne, cursor, err := PageRawMessages(binding, nil, rows)
	if err != nil {
		t.Fatalf("page one: %v", err)
	}
	if len(pageOne) != 2 || cursor == nil || cursor.Position["offset"] != "2" {
		t.Fatalf("unexpected page one: rows=%d cursor=%#v", len(pageOne), cursor)
	}
	pageTwo, cursor, err := PageRawMessages(binding, cursor, rows)
	if err != nil {
		t.Fatalf("page two: %v", err)
	}
	if len(pageTwo) != 1 || cursor != nil {
		t.Fatalf("unexpected page two: rows=%d cursor=%#v", len(pageTwo), cursor)
	}
}
