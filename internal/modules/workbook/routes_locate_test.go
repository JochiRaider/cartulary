package workbook

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLocatorValidationPrecedence_Unit(t *testing.T) {
	const id = `"record_id":"10000000-0000-4000-8000-000000000001"`
	tests := []struct{ name, body, suffix, reason string }{
		{"non object", `[]`, "", "malformed_locator_request"},
		{"duplicate nested", `{` + id + `,"filters":[{"arg":{"value":1,"value":2}}]}`, "", "malformed_locator_request"},
		{"trailing", `{` + id + `} {}`, "", "malformed_locator_request"},
		{"unknown before missing id", `{"secret":true}`, "", "unknown_locator_member"},
		{"cursor forbidden", `{` + id + `,"cursor_token":"secret"}`, "", "unknown_locator_member"},
		{"url forbidden", `{` + id + `}`, "?limit=100", "unknown_locator_member"},
		{"missing id", `{}`, "", "invalid_record_id"},
		{"null id", `{"record_id":null}`, "", "invalid_record_id"},
		{"invalid id", `{"record_id":"secret","sort":null}`, "", "invalid_record_id"},
		{"null sort", `{` + id + `,"sort":null}`, "", "invalid_sort_entry"},
		{"sort before null filters", `{` + id + `,"sort":[{}],"filters":null}`, "", "invalid_sort_entry"},
		{"null filters", `{` + id + `,"filters":null}`, "", "invalid_filter_operand"},
		{"null grouping", `{` + id + `,"group_by":null}`, "", "invalid_group_by"},
		{"defaults", `{` + id + `}`, "", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/locate"+tc.suffix, strings.NewReader(tc.body))
			_, query, err := decodeLocatorRequest(r, "cartulary.view.timeline.v2")
			if tc.reason == "" {
				if err != nil || len(query.Meta.Sort) == 0 {
					t.Fatalf("default admission: %#v %#v", query, err)
				}
				return
			}
			if err == nil || err.Code != "invalid_view_query" || err.Details["reason_code"] != tc.reason {
				t.Fatalf("want %s, got %#v", tc.reason, err)
			}
			if _, ok := err.Details["secret"]; ok {
				t.Fatal("submitted content disclosed")
			}
		})
	}
}
