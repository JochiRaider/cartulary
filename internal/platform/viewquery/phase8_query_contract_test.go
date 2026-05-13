package viewquery

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func TestPhase8_StableFieldKeysOnly_U_8_06(t *testing.T) {
	for name, body := range map[string]string{
		"visible label sort":         `{"sort":[{"field_key":"Summary","direction":"asc"}]}`,
		"storage column sort":        `{"sort":[{"field_key":"timeline_events.summary","direction":"asc"}]}`,
		"client record id sort":      `{"sort":[{"field_key":"record_id","direction":"asc"}]}`,
		"unknown filter field":       `{"filters":[{"field_key":"Summary","op":"eq","arg":{"value":"x"}}]}`,
		"operator not allowed":       `{"filters":[{"field_key":"timeline.summary","op":"contains_any","arg":{"values":["x"]}}]}`,
		"duplicate filter field":     `{"filters":[{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"rough"}},{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"enriched"}}]}`,
		"duplicate sort field":       `{"sort":[{"field_key":"timeline.summary","direction":"asc"},{"field_key":"timeline.summary","direction":"desc"}]}`,
		"unknown grouping key":       `{"group_by":"timeline.summary"}`,
		"malformed group by null":    `{"group_by":null}`,
		"null inside set operand":    `{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["alpha",null]}}]}`,
		"empty set after coalescing": `{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["alpha","alpha"]}},{"field_key":"timeline.tags","op":"contains_all","arg":{"values":["alpha"]}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Decode(strings.NewReader(body), timelineViewSchemaID)
			if err == nil {
				t.Fatal("expected invalid view query")
			}
			if err.ReasonCode == "" {
				t.Fatalf("expected stable reason code, got %+v", err)
			}
		})
	}
}

func TestPhase8_TimelineGroupingWhitelist_U_8_07(t *testing.T) {
	for _, key := range []string{"timeline.capture_state", "timeline.has_evidence", "timeline.has_unresolved_mentions"} {
		t.Run(key, func(t *testing.T) {
			query, err := Decode(strings.NewReader(`{"group_by":`+quoteJSON(key)+`}`), timelineViewSchemaID)
			if err != nil {
				t.Fatalf("expected grouping key %s to be allowed: %+v", key, err)
			}
			if query.Meta.GroupBy == nil || *query.Meta.GroupBy != key {
				t.Fatalf("unexpected group_by: %#v", query.Meta.GroupBy)
			}
		})
	}

	query, err := Decode(strings.NewReader(`{"group_by":"timeline.summary"}`), timelineViewSchemaID)
	if err == nil {
		t.Fatalf("expected invalid grouping key, got %#v", query)
	}
	if err.ReasonCode != "group_by_not_allowed" || err.FieldKey != "timeline.summary" {
		t.Fatalf("unexpected grouping error: %+v", err)
	}
}

func TestPhase8_QueryNormalizationMeta_U_8_08(t *testing.T) {
	query, err := Decode(strings.NewReader(`{
  "sort": [{"field_key": "timeline.summary", "direction": "asc"}],
  "filters": [
    {"field_key": "timeline.tags", "op": "contains_any", "arg": {"values": ["beta", "alpha", "alpha"]}},
    {"field_key": "timeline.capture_state", "op": "prefix", "arg": {"value": "  Alpha  "}}
  ]
}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("decode canonical AC-184 query: %+v", err)
	}

	wantSort := []viewschema.SortEntry{
		{FieldKey: "timeline.summary", Direction: "asc"},
		{FieldKey: "timeline.sort_ts", Direction: "asc"},
		{FieldKey: "record_id", Direction: "asc"},
	}
	if !reflect.DeepEqual(query.Meta.Sort, wantSort) {
		t.Fatalf("unexpected effective sort:\ngot  %#v\nwant %#v", query.Meta.Sort, wantSort)
	}
	if query.Meta.GroupBy != nil {
		t.Fatalf("inactive grouping must be omitted, got %#v", query.Meta.GroupBy)
	}
	if got := query.Meta.Filters[0].Arg["value"]; got != "alpha" {
		t.Fatalf("prefix value must be comparison-normalized, got %#v", got)
	}
	values, ok := query.Meta.Filters[1].Arg["values"].([]any)
	if !ok || !reflect.DeepEqual(values, []any{"alpha", "beta"}) {
		t.Fatalf("set-like values must be unique and sorted, got %#v", query.Meta.Filters[1].Arg)
	}

	for name, body := range map[string]string{
		"sort ceiling":   oversizeSortBody(config.PublicSortLimit + 1),
		"filter ceiling": oversizeFilterBody(config.PublicFilterLimit + 1),
		"zero tokens":    `{"filters":[{"field_key":"note.full_text","op":"full_text","arg":{"query":" -- "}}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			viewSchemaID := timelineViewSchemaID
			if strings.Contains(body, "note.full_text") {
				viewSchemaID = "cartulary.view.notes.v1"
			}
			_, err := Decode(strings.NewReader(body), viewSchemaID)
			if err == nil {
				t.Fatal("expected invalid view query")
			}
			switch name {
			case "sort ceiling":
				if err.ReasonCode != "sort_count_exceeded" || *err.RequestedCount != config.PublicSortLimit+1 || *err.MaxCount != config.PublicSortLimit {
					t.Fatalf("unexpected sort ceiling error: %+v", err)
				}
			case "filter ceiling":
				if err.ReasonCode != "filter_count_exceeded" || *err.RequestedCount != config.PublicFilterLimit+1 || *err.MaxCount != config.PublicFilterLimit {
					t.Fatalf("unexpected filter ceiling error: %+v", err)
				}
			case "zero tokens":
				if err.ReasonCode != "empty_full_text_after_tokenization" {
					t.Fatalf("unexpected zero-token error: %+v", err)
				}
			}
		})
	}
}

func quoteJSON(value string) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func oversizeSortBody(count int) string {
	entries := make([]string, 0, count)
	for index := 0; index < count; index++ {
		entries = append(entries, `{"field_key":"timeline.summary","direction":"asc"}`)
	}
	return `{"sort":[` + strings.Join(entries, ",") + `]}`
}

func oversizeFilterBody(count int) string {
	entries := make([]string, 0, count)
	for index := 0; index < count; index++ {
		entries = append(entries, `{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["tag`+string(rune('a'+index%26))+`"]}}`)
	}
	return `{"filters":[` + strings.Join(entries, ",") + `]}`
}
