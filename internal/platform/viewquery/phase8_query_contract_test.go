package viewquery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"gopkg.in/yaml.v3"
)

func TestPhase8_StableFieldKeysOnly_U_8_06(t *testing.T) {
	valid, err := Decode(strings.NewReader(`{
  "sort": [{"field_key":"timeline.activity_synopsis_text","direction":"asc"}],
  "filters": [{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"rough"}}],
  "group_by": "timeline.capture_state"
}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("stable field_key query should be accepted: %+v", err)
	}
	if valid.Meta.Filters[0].FieldKey != "timeline.capture_state" || valid.Meta.Sort[0].FieldKey != "timeline.activity_synopsis_text" || valid.Meta.GroupBy == nil || *valid.Meta.GroupBy != "timeline.capture_state" {
		t.Fatalf("stable keys were not preserved in canonical query: %#v", valid.Meta)
	}

	for name, testCase := range map[string]struct {
		body       string
		reasonCode string
		fieldKey   string
	}{
		"visible label sort":       {body: `{"sort":[{"field_key":"Summary","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "Summary"},
		"vendor column sort":       {body: `{"sort":[{"field_key":"rdg:timeline.activity_synopsis_text","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "rdg:timeline.activity_synopsis_text"},
		"projection table sort":    {body: `{"sort":[{"field_key":"timeline_grid_projection.summary","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "timeline_grid_projection.summary"},
		"storage table sort":       {body: `{"sort":[{"field_key":"timeline_events.summary","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "timeline_events.summary"},
		"client record id sort":    {body: `{"sort":[{"field_key":"record_id","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "record_id"},
		"client row version sort":  {body: `{"sort":[{"field_key":"row_version","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "row_version"},
		"non sortable field":       {body: `{"sort":[{"field_key":"timeline.raw_activity_text","direction":"asc"}]}`, reasonCode: "sort_field_not_allowed", fieldKey: "timeline.raw_activity_text"},
		"duplicate sort field":     {body: `{"sort":[{"field_key":"timeline.activity_synopsis_text","direction":"asc"},{"field_key":"timeline.activity_synopsis_text","direction":"desc"}]}`, reasonCode: "duplicate_sort_field", fieldKey: "timeline.activity_synopsis_text"},
		"unknown filter label":     {body: `{"filters":[{"field_key":"Summary","op":"eq","arg":{"value":"x"}}]}`, reasonCode: "unknown_filter_field", fieldKey: "Summary"},
		"filter record id":         {body: `{"filters":[{"field_key":"record_id","op":"eq","arg":{"value":"x"}}]}`, reasonCode: "unknown_filter_field", fieldKey: "record_id"},
		"filter row version":       {body: `{"filters":[{"field_key":"row_version","op":"eq","arg":{"value":1}}]}`, reasonCode: "unknown_filter_field", fieldKey: "row_version"},
		"operator not allowed":     {body: `{"filters":[{"field_key":"timeline.activity_synopsis_text","op":"contains_any","arg":{"values":["x"]}}]}`, reasonCode: "unknown_filter_field", fieldKey: "timeline.activity_synopsis_text"},
		"invalid operator":         {body: `{"filters":[{"field_key":"timeline.capture_state","op":"contains_any","arg":{"values":["rough"]}}]}`, reasonCode: "operator_not_allowed", fieldKey: "timeline.capture_state"},
		"missing arg":              {body: `{"filters":[{"field_key":"timeline.capture_state","op":"eq"}]}`, reasonCode: "invalid_filter_operand"},
		"malformed arg shape":      {body: `{"filters":[{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"rough","unexpected":true}}]}`, reasonCode: "invalid_filter_operand", fieldKey: "timeline.capture_state"},
		"scalar values malformed":  {body: `{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":"alpha"}}]}`, reasonCode: "invalid_filter_operand", fieldKey: "timeline.tags"},
		"null inside set operand":  {body: `{"filters":[{"field_key":"timeline.tags","op":"contains_any","arg":{"values":["alpha",null]}}]}`, reasonCode: "invalid_filter_operand", fieldKey: "timeline.tags"},
		"duplicate filter field":   {body: `{"filters":[{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"rough"}},{"field_key":"timeline.capture_state","op":"eq","arg":{"value":"enriched"}}]}`, reasonCode: "duplicate_filter_field", fieldKey: "timeline.capture_state"},
		"unknown grouping key":     {body: `{"group_by":"timeline.activity_synopsis_text"}`, reasonCode: "group_by_not_allowed", fieldKey: "timeline.activity_synopsis_text"},
		"group record id":          {body: `{"group_by":"record_id"}`, reasonCode: "group_by_not_allowed", fieldKey: "record_id"},
		"group row version":        {body: `{"group_by":"row_version"}`, reasonCode: "group_by_not_allowed", fieldKey: "row_version"},
		"malformed group by null":  {body: `{"group_by":null}`, reasonCode: "invalid_group_by"},
		"malformed group by array": {body: `{"group_by":["timeline.capture_state"]}`, reasonCode: "invalid_group_by"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := Decode(strings.NewReader(testCase.body), timelineViewSchemaID)
			if err == nil {
				t.Fatal("expected invalid view query")
			}
			if err.ReasonCode != testCase.reasonCode {
				t.Fatalf("unexpected reason code: got %+v want %s", err, testCase.reasonCode)
			}
			if testCase.fieldKey != "" && err.FieldKey != testCase.fieldKey {
				t.Fatalf("unexpected field key: got %+v want %s", err, testCase.fieldKey)
			}
		})
	}
}

func TestTimelineGroupingWhitelistRejectsNonContractKeys(t *testing.T) {
	wantWhitelist := []string{
		"timeline.date_entered_sort_day",
		"timeline.date_entered_sort_day",
		"timeline.capture_state",
		"timeline.has_evidence",
		"timeline.has_unresolved_mentions",
	}
	schema, ok := viewschema.Lookup(timelineViewSchemaID)
	if !ok {
		t.Fatal("timeline schema not registered")
	}
	if !reflect.DeepEqual(schema.GroupingFields(), wantWhitelist) {
		t.Fatalf("timeline grouping whitelist changed:\ngot  %#v\nwant %#v", schema.GroupingFields(), wantWhitelist)
	}

	for _, key := range wantWhitelist {
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

	for name, testCase := range map[string]struct {
		key        string
		reasonCode string
	}{
		"visible label":     {key: "Capture State", reasonCode: "group_by_not_allowed"},
		"summary":           {key: "timeline.activity_synopsis_text", reasonCode: "group_by_not_allowed"},
		"tags collection":   {key: "timeline.tags", reasonCode: "group_by_not_allowed"},
		"record id":         {key: "record_id", reasonCode: "group_by_not_allowed"},
		"row version":       {key: "row_version", reasonCode: "group_by_not_allowed"},
		"projection column": {key: "timeline_grid_projection.capture_state", reasonCode: "group_by_not_allowed"},
		"storage column":    {key: "timeline_events.capture_state", reasonCode: "group_by_not_allowed"},
		"event type":        {key: "timeline.event_type", reasonCode: "group_by_not_allowed"},
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			query, err := Decode(strings.NewReader(`{"group_by":`+quoteJSON(testCase.key)+`}`), timelineViewSchemaID)
			if err == nil {
				t.Fatalf("expected invalid grouping key, got %#v", query)
			}
			if err.ReasonCode != testCase.reasonCode || err.FieldKey != testCase.key {
				t.Fatalf("unexpected grouping error: %+v", err)
			}
		})
	}
}

func TestPhase8_QueryNormalizationMeta_U_8_08(t *testing.T) {
	query, err := Decode(strings.NewReader(`{
  "sort": [{"field_key": "timeline.activity_synopsis_text", "direction": "asc"}],
  "filters": [
    {"field_key": "timeline.tags", "op": "contains_any", "arg": {"values": ["beta", "alpha", "alpha"]}},
    {"field_key": "timeline.capture_state", "op": "prefix", "arg": {"value": "  Alpha  "}}
  ]
}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("decode canonical AC-184 query: %+v", err)
	}

	wantSort := []viewschema.SortEntry{
		{FieldKey: "timeline.activity_synopsis_text", Direction: "asc"},
		{FieldKey: "timeline.activity_sort_ts", Direction: "asc"},
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

	metaPayload, marshalErr := json.Marshal(query.Meta)
	if marshalErr != nil {
		t.Fatalf("marshal query meta: %v", marshalErr)
	}
	var meta map[string]any
	if err := json.Unmarshal(metaPayload, &meta); err != nil {
		t.Fatalf("decode query meta JSON: %v", err)
	}
	if _, ok := meta["sort"].([]any); !ok {
		t.Fatalf("meta.query.sort must always be present as an array, got %#v", meta)
	}
	if _, ok := meta["filters"].([]any); !ok {
		t.Fatalf("meta.query.filters must always be present as an array, got %#v", meta)
	}
	if _, ok := meta["group_by"]; ok {
		t.Fatalf("inactive meta.query.group_by must be omitted, got %#v", meta)
	}

	defaultQuery, err := Decode(strings.NewReader(`{}`), timelineViewSchemaID)
	if err != nil {
		t.Fatalf("decode default query: %+v", err)
	}
	if !reflect.DeepEqual(defaultQuery.Meta.Sort, []viewschema.SortEntry{
		{FieldKey: "timeline.activity_sort_ts", Direction: "asc"},
		{FieldKey: "record_id", Direction: "asc"},
	}) {
		t.Fatalf("omitted sort must use schema default plus server tie-breaker, got %#v", defaultQuery.Meta.Sort)
	}

	fullTextQuery, err := Decode(strings.NewReader(`{
  "filters": [{"field_key":"note.full_text","op":"full_text","arg":{"query":" Beta alpha beta -- ALPHA "}}]
}`), "cartulary.view.notes.v1")
	if err != nil {
		t.Fatalf("decode canonical full_text query: %+v", err)
	}
	if got := fullTextQuery.Meta.Filters[0].Arg["query"]; got != "alpha beta" {
		t.Fatalf("full_text query must tokenize, de-duplicate, fold, and sort tokens, got %#v", got)
	}

	persisted, validationErr := NormalizePersisted(json.RawMessage(`{
  "sort": [{"field_key": "timeline.activity_synopsis_text", "direction": "desc"}],
  "filters": [
    {"field_key": "timeline.tags", "op": "contains_any", "arg": {"values": ["beta", "alpha", "alpha"]}},
    {"field_key": "timeline.capture_state", "op": "prefix", "arg": {"value": "  Alpha  "}}
  ]
}`), timelineViewSchemaID)
	if validationErr != nil {
		t.Fatalf("normalize persisted saved-view query: %+v", validationErr)
	}
	var saved PersistedQuery
	if err := json.Unmarshal(persisted, &saved); err != nil {
		t.Fatalf("decode persisted saved-view query: %v", err)
	}
	if !reflect.DeepEqual(saved.Sort, []viewschema.SortEntry{{FieldKey: "timeline.activity_synopsis_text", Direction: "desc"}}) {
		t.Fatalf("saved-view query_json.sort must store user overrides only, got %#v", saved.Sort)
	}
	if saved.GroupBy != nil {
		t.Fatalf("inactive saved-view query_json.group_by must be omitted, got %#v", saved.GroupBy)
	}
	if len(saved.Filters) != 2 || saved.Filters[0].FieldKey != "timeline.capture_state" || saved.Filters[1].FieldKey != "timeline.tags" {
		t.Fatalf("saved-view query_json.filters must use canonical field order, got %#v", saved.Filters)
	}

	emptyPersisted, validationErr := NormalizePersisted(json.RawMessage(`{}`), timelineViewSchemaID)
	if validationErr != nil {
		t.Fatalf("normalize empty persisted saved-view query: %+v", validationErr)
	}
	var emptySaved PersistedQuery
	if err := json.Unmarshal(emptyPersisted, &emptySaved); err != nil {
		t.Fatalf("decode empty persisted saved-view query: %v", err)
	}
	if len(emptySaved.Sort) != 0 || len(emptySaved.Filters) != 0 || emptySaved.GroupBy != nil {
		t.Fatalf("omitted saved-view query members must normalize to empty user overrides with omitted grouping, got %#v", emptySaved)
	}

	requireWorkbookQueryOpenAPIHasCanonicalMeta(t)

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

func requireWorkbookQueryOpenAPIHasCanonicalMeta(t testing.TB) {
	t.Helper()
	document := readOpenAPIContract(t)
	schemas := objectAt(t, objectAt(t, document, "components"), "schemas")

	envelope := schemaAt(t, schemas, "WorkbookQueryEnvelope")
	metaRef := stringAt(t, objectAt(t, objectAt(t, envelope, "properties"), "meta"), "$ref")
	if metaRef != "#/components/schemas/WorkbookQueryMeta" {
		t.Fatalf("workbook query envelope must use query-specific meta schema, got %q", metaRef)
	}

	queryMeta := schemaAt(t, schemas, "WorkbookQueryMeta")
	queryRef := stringAt(t, objectAt(t, objectAt(t, queryMeta, "properties"), "query"), "$ref")
	if queryRef != "#/components/schemas/WorkbookQueryMetaQuery" {
		t.Fatalf("workbook query meta must expose canonical query object, got %q", queryRef)
	}
	requireStringSet(t, queryMeta["required"], []string{"request_id", "query"})

	query := schemaAt(t, schemas, "WorkbookQueryMetaQuery")
	queryProps := objectAt(t, query, "properties")
	requireStringSet(t, query["required"], []string{"sort", "filters"})
	groupBy := objectAt(t, queryProps, "group_by")
	if stringAt(t, groupBy, "type") != "string" {
		t.Fatalf("meta.query.group_by must be optional string-only and non-nullable, got %#v", groupBy)
	}
}

func quoteJSON(value string) string {
	payload, _ := json.Marshal(value)
	return string(payload)
}

func oversizeSortBody(count int) string {
	entries := make([]string, 0, count)
	for index := 0; index < count; index++ {
		entries = append(entries, `{"field_key":"timeline.activity_synopsis_text","direction":"asc"}`)
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

func readOpenAPIContract(t testing.TB) map[string]any {
	t.Helper()
	path := filepath.Join("..", "..", "..", "contracts", "openapi", "cartulary.openapi.yaml")
	payload, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read OpenAPI contract %s: %v", path, err)
	}
	var document map[string]any
	if err := yaml.Unmarshal(payload, &document); err != nil {
		t.Fatalf("decode OpenAPI contract %s: %v", path, err)
	}
	return document
}

func schemaAt(t testing.TB, schemas map[string]any, name string) map[string]any {
	t.Helper()
	schema, ok := schemas[name].(map[string]any)
	if !ok {
		t.Fatalf("schema %s missing or not an object: %#v", name, schemas[name])
	}
	return schema
}

func objectAt(t testing.TB, parent map[string]any, key string) map[string]any {
	t.Helper()
	value, ok := parent[key].(map[string]any)
	if !ok {
		t.Fatalf("%s missing or not an object: %#v", key, parent[key])
	}
	return value
}

func stringAt(t testing.TB, parent map[string]any, key string) string {
	t.Helper()
	value, ok := parent[key].(string)
	if !ok {
		t.Fatalf("%s missing or not a string: %#v", key, parent[key])
	}
	return value
}

func requireStringSet(t testing.TB, raw any, want []string) {
	t.Helper()
	items, ok := raw.([]any)
	if !ok {
		t.Fatalf("expected string array, got %#v", raw)
	}
	got := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("expected string array, got %#v", raw)
		}
		got = append(got, value)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected string set: got %#v want %#v", got, want)
	}
}
