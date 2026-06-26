package workbook

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/viewquery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/google/uuid"
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

	recordID := uuid.MustParse("00000000-0000-0000-0000-000000000807")
	groupBy := "host.host_state"
	row, err := buildGenericRow(genericSurface{
		viewSchemaID: "cartulary.view.hosts.v1",
		recordExpr:   "h.record_id",
		fields: []genericField{
			{key: "host.display_name", kind: fieldKindText},
			{key: "host.host_state", kind: fieldKindText},
			{key: "host.edited_at", kind: fieldKindTimestamp},
		},
	}, &groupBy, []any{
		recordID,
		int64(12),
		"Host A",
		"reviewed",
		time.Date(2026, 5, 16, 12, 30, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("build grouped workbook row: %v", err)
	}

	if !reflect.DeepEqual(sortedKeys(row), []string{"cells", "group_values", "record_id", "row_version"}) {
		t.Fatalf("grouped workbook response row must remain a full row resource, got keys %#v in %#v", sortedKeys(row), row)
	}
	if row["record_id"] != recordID.String() || row["row_version"] != int64(12) {
		t.Fatalf("grouped workbook row must keep top-level record identity and version, got %#v", row)
	}
	cells, ok := row["cells"].(map[string]any)
	if !ok || len(cells) != 3 {
		t.Fatalf("grouped workbook row must serialize field cells, got %#v", row["cells"])
	}
	groupValues, ok := row["group_values"].(map[string]any)
	if !ok || groupValues[groupBy] != "reviewed" {
		t.Fatalf("grouped workbook row must serialize group_values without a header row, got %#v", row["group_values"])
	}

	payload, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal grouped workbook row: %v", err)
	}
	encoded := string(payload)
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
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("grouped workbook row serialized presentation-only marker %s: %s", forbidden, encoded)
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

func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
