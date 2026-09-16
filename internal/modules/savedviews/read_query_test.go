package savedviews

import (
	"maps"
	"net/url"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func TestSavedViewReadQueryValidation_Unit(t *testing.T) {
	codec := pagination.NewCodec(make([]byte, 32))
	cases := []struct{ query, code, reason string }{
		{"view_schema_id=", "invalid_list_query", "invalid_filter_value"},
		{"view_schema_id=unknown", "invalid_list_query", "invalid_filter_value"},
		{"view_schema_id=%20cartulary.view.timeline.v2", "invalid_list_query", "invalid_filter_value"},
		{"view_schema_id=null", "invalid_list_query", "invalid_filter_value"},
		{"view_schema_id=a,b", "invalid_list_query", "invalid_filter_value"},
		{"view_schema_id=[]", "invalid_list_query", "invalid_filter_value"},
		{"unknown=x&unknown=y", "invalid_list_query", "duplicate_query_member"},
		{"limit=1&limit=2", "invalid_list_query", "duplicate_query_member"},
		{"scope=private", "invalid_list_query", "unknown_query_member"},
		{"limit=", "invalid_pagination_request", "invalid_limit"},
		{"limit=0", "invalid_pagination_request", "invalid_limit"},
		{"limit=501", "invalid_pagination_request", "invalid_limit"},
		{"cursor_token=", "invalid_pagination_request", "invalid_cursor_token"},
		{"offset=0", "invalid_pagination_request", "invalid_limit"},
		{"view_schema_id=%ff", "invalid_saved_view_read_request", "malformed_query"},
		{"view_schema_id=%xy", "invalid_saved_view_read_request", "malformed_query"},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			_, _, err := resolveSavedViewListRead(codec, tc.query, "actor", "incident")
			if err == nil || err.Code != tc.code || err.Details["reason_code"] != tc.reason {
				t.Fatalf("error=%+v want %s/%s", err, tc.code, tc.reason)
			}
		})
	}
	for _, query := range []string{"limit=", "cursor_token=x", "page=1", "offset=0", "page_size=2", "block_size=2"} {
		err := validateSavedViewDetailQuery(query)
		if err == nil || err.Code != "invalid_pagination_request" || err.Details["reason_code"] != "pagination_not_supported" {
			t.Fatalf("singleton %s: %+v", query, err)
		}
	}
	if err := validateSavedViewDetailQuery("view_schema_id=x"); err == nil || err.Code != "invalid_saved_view_read_request" {
		t.Fatalf("singleton filter: %+v", err)
	}
	if err := validateSavedViewDetailQuery(""); err != nil {
		t.Fatal(err)
	}
}

func TestSavedViewReadCursorBindings_Unit(t *testing.T) {
	codec := pagination.NewCodec(make([]byte, 32))
	cursor := pagination.Cursor{Version: pagination.CursorVersion, Mode: pagination.ModeKeyset, Route: "incident.saved-views.list", ActorUserID: "actor", Limit: 2, Scope: map[string]string{"incident_id": "incident"}, Position: map[string]string{"anchor_updated_at": "2026-09-01T00:00:00Z", "last_updated_at": "2026-09-01T00:00:00Z", "last_saved_view_id": "00000000-0000-4000-8000-000000000001"}}
	token, err := codec.Encode(cursor)
	if err != nil {
		t.Fatal(err)
	}
	binding, decoded, apiErr := resolveSavedViewListRead(codec, "cursor_token="+url.QueryEscape(token), "actor", "incident")
	if apiErr != nil || binding.Limit != 2 || binding.Scope["view_schema_id"] != "" {
		t.Fatalf("legacy cursor: %+v %+v", binding, apiErr)
	}
	if _, reason := savedViewListPageRequest(binding, decoded); reason != "" {
		t.Fatal(reason)
	}
	for _, tc := range []struct{ query, actor, incident, reason string }{
		{"&view_schema_id=cartulary.view.timeline.v2", "actor", "incident", "cursor_query_mismatch"},
		{"", "other", "incident", "invalid_cursor_token"},
		{"", "actor", "other", "invalid_cursor_token"},
		{"&limit=3", "actor", "incident", "invalid_cursor_token"},
	} {
		_, _, err := resolveSavedViewListRead(codec, "cursor_token="+url.QueryEscape(token)+tc.query, tc.actor, tc.incident)
		if err == nil || err.Details["reason_code"] != tc.reason {
			t.Fatalf("binding: %+v", err)
		}
	}
	cursor.Scope["view_schema_id"] = "cartulary.view.timeline.v2"
	token, err = codec.Encode(cursor)
	if err != nil {
		t.Fatal(err)
	}
	_, _, apiErr = resolveSavedViewListRead(codec, "cursor_token="+url.QueryEscape(token)+"&view_schema_id=cartulary.view.timeline.v2", "actor", "incident")
	if apiErr != nil {
		t.Fatal(apiErr)
	}
	_, _, apiErr = resolveSavedViewListRead(codec, "cursor_token="+url.QueryEscape(token), "actor", "incident")
	if apiErr == nil || apiErr.Details["reason_code"] != "cursor_query_mismatch" {
		t.Fatalf("missing filter: %+v", apiErr)
	}
	for name, change := range map[string]func(*pagination.Cursor){
		"route":           func(c *pagination.Cursor) { c.Route = "other.route" },
		"mode":            func(c *pagination.Cursor) { c.Mode = "offset" },
		"anchor":          func(c *pagination.Cursor) { c.Position["anchor_updated_at"] = "invalid" },
		"future position": func(c *pagination.Cursor) { c.Position["last_updated_at"] = "2026-09-02T00:00:00Z" },
		"identity":        func(c *pagination.Cursor) { c.Position["last_saved_view_id"] = "invalid" },
		"extra position":  func(c *pagination.Cursor) { c.Position["offset"] = "1" },
	} {
		t.Run(name, func(t *testing.T) {
			invalid := cursor
			invalid.Position = maps.Clone(cursor.Position)
			change(&invalid)
			wire, err := codec.Encode(invalid)
			if err != nil {
				t.Fatal(err)
			}
			bound, decoded, apiErr := resolveSavedViewListRead(codec, "cursor_token="+url.QueryEscape(wire)+"&view_schema_id=cartulary.view.timeline.v2", "actor", "incident")
			if apiErr != nil {
				if apiErr.Details["reason_code"] != pagination.ReasonInvalidCursorToken {
					t.Fatalf("wrong binding error: %+v", apiErr)
				}
				return
			}
			if _, reason := savedViewListPageRequest(bound, decoded); reason != pagination.ReasonInvalidCursorToken {
				t.Fatalf("invalid position accepted: %s", reason)
			}
		})
	}
}
