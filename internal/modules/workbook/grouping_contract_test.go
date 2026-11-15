package workbook

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestTimelineGroupingAndWorkbookPresentationOnly_Unit(t *testing.T) {
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
}

func quoteJSON(value string) string {
	payload, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(payload)
}
