package workbook

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestPhase8_TimelineGroupingAndWorkbookPresentationOnly_U_8_07(t *testing.T) {
	wantWhitelist := []string{
		"timeline.date_entered_sort_day",
		"timeline.activity_time_pair_state",
		"timeline.capture_state",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
	}
	schema, ok := viewschema.Lookup("cartulary.view.timeline.v2")
	if !ok {
		t.Fatal("timeline schema not registered")
	}
	if !reflect.DeepEqual(schema.GroupingFields(), wantWhitelist) {
		t.Fatalf("timeline grouping whitelist changed:\ngot  %#v\nwant %#v", schema.GroupingFields(), wantWhitelist)
	}
	for _, key := range wantWhitelist {
		query, err := viewquery.Decode(strings.NewReader(`{"group_by":`+quoteJSON(key)+`}`), "cartulary.view.timeline.v2")
		if err != nil {
			t.Fatalf("expected grouping key %s to be allowed: %+v", key, err)
		}
		if query.Meta.GroupBy == nil || *query.Meta.GroupBy != key {
			t.Fatalf("unexpected group_by: %#v", query.Meta.GroupBy)
		}
	}
	for name, key := range map[string]string{
		"visible label":     "Capture State",
		"summary":           "timeline.activity_synopsis_text",
		"tags collection":   "timeline.tags",
		"record id":         "record_id",
		"row version":       "row_version",
		"projection column": "timeline_grid_projection.capture_state",
		"storage column":    "timeline_events.capture_state",
		"event type":        "timeline.event_type",
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			query, err := viewquery.Decode(strings.NewReader(`{"group_by":`+quoteJSON(key)+`}`), "cartulary.view.timeline.v2")
			if err == nil {
				t.Fatalf("expected invalid grouping key, got %#v", query)
			}
			if err.ReasonCode != "group_by_not_allowed" || err.FieldKey != key {
				t.Fatalf("unexpected grouping error: %+v", err)
			}
		})
	}

	encoded, err := json.Marshal(map[string]any{
		"record_id":   "00000000-0000-0000-0000-000000000807",
		"row_version": int64(12),
		"cells": map[string]any{
			"host.display_name": map[string]any{"value": "Host A"},
			"host.host_state":   map[string]any{"value": "reviewed"},
		},
		"group_values": map[string]any{"host.host_state": "reviewed"},
	})
	if err != nil {
		t.Fatalf("marshal grouped workbook row fixture: %v", err)
	}
	for _, forbidden := range []string{
		`"row_kind"`,
		`"is_group_header"`,
		`"subtotal"`,
		`"writable"`,
		`"group_header_id"`,
		`"mutation_target"`,
		`"paste_target"`,
		`"target_record_id"`,
	} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("grouped workbook row serialized presentation-only marker %s: %s", forbidden, string(encoded))
		}
	}
}

func quoteJSON(value string) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(payload)
}
