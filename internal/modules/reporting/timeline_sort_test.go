package reporting

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

func testSnapshotTimelineSort(t *testing.T) {
	text := func(s string) *string { return &s }
	fact := func(id string, ts *string, rank int) exportprovider.TimelineSortFact {
		state, bucket := "missing", "unresolved_time"
		if ts != nil {
			state, bucket = "parsed", "parsed_time"
		}
		return exportprovider.TimelineSortFact{RecordID: id, ActivitySortTS: ts, ActivityParseState: state, ActivityPrecisionRank: rank, RecordCreatedAt: "2026-05-01T00:00:00.000000Z", TimeBucket: bucket}
	}
	facts := []exportprovider.TimelineSortFact{
		fact("unresolved", nil, 0), fact("later", text("2026-05-03T00:00:00.000000Z"), 6),
		fact("minute", text("2026-05-02T00:00:00.000000Z"), 5), fact("second", text("2026-05-02T00:00:00.000000Z"), 6),
	}
	source := SnapshotSource{SourceFacts: exportprovider.EmptySourceFacts(), Timeline: facts}
	for _, f := range facts {
		source.Fields = append(source.Fields, ExportField{SourceFamily: "timeline_event", SourceRecordID: f.RecordID, SourceRecordType: "timeline_event", Path: "/opaque/" + f.RecordID, ContentClass: ContentClassSourceEvidence, Value: map[string]any{"capture_state": "rough", "value": "not a timestamp"}})
	}
	build := func(s SnapshotSource) (SnapshotModel, error) {
		m, _, err := buildStructuredSnapshotModel("incident", fixtureSnapshotID, time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC), "boundary", s)
		return m, err
	}
	model, err := build(source)
	if err != nil {
		t.Fatal(err)
	}
	ids := func(events []ReportingTimelineEvent) []string {
		out := []string{}
		for _, e := range events {
			out = append(out, e.RecordID)
		}
		return out
	}
	if !reflect.DeepEqual(ids(model.TimelineEvents), []string{"second", "minute", "later", "unresolved"}) {
		t.Fatal("wrong default tuple order", ids(model.TimelineEvents))
	}
	if model.TimelineEvents[0].ActivitySortKey != `[0,"2026-05-02T00:00:00.000000Z",-6,"9999-12-31","2026-05-01T00:00:00.000000Z","second"]` {
		t.Fatal("wrong serialized sort tuple")
	}
	for _, tc := range []struct {
		descending, unresolvedFirst bool
		want                        []string
	}{
		{true, false, []string{"later", "second", "minute", "unresolved"}},
		{false, true, []string{"unresolved", "second", "minute", "later"}},
		{true, true, []string{"unresolved", "later", "second", "minute"}},
	} {
		events := slices.Clone(model.TimelineEvents)
		slices.SortFunc(events, func(a, b ReportingTimelineEvent) int {
			return compareTimelineEvents(a, b, tc.descending, tc.unresolvedFirst)
		})
		if !reflect.DeepEqual(ids(events), tc.want) {
			t.Fatal("variant reversed a forbidden tie-break", ids(events))
		}
	}
	for name, change := range map[string]func(*SnapshotSource){
		"missing":              func(s *SnapshotSource) { s.Timeline = nil },
		"duplicate":            func(s *SnapshotSource) { s.Timeline = append(s.Timeline, s.Timeline[0]) },
		"orphan":               func(s *SnapshotSource) { s.Timeline[0].RecordID = "foreign" },
		"no creation time":     func(s *SnapshotSource) { s.Timeline[0].RecordCreatedAt = "" },
		"unresolved timestamp": func(s *SnapshotSource) { s.Timeline[0].ActivitySortTS = text("2026-05-01T00:00:00.000000Z") },
		"unknown state":        func(s *SnapshotSource) { s.Timeline[0].ActivityParseState = "unresolved" },
		"invalid precision":    func(s *SnapshotSource) { s.Timeline[1].ActivityPrecisionRank = 0 },
		"invalid date":         func(s *SnapshotSource) { s.Timeline[0].DateEnteredSortKey = text("2026-02-30") },
		"noncanonical time":    func(s *SnapshotSource) { s.Timeline[1].ActivitySortTS = text("2026-05-03T01:00:00+01:00") },
	} {
		t.Run(name, func(t *testing.T) {
			candidate := source
			candidate.Timeline = slices.Clone(facts)
			change(&candidate)
			if _, err := build(candidate); err == nil {
				t.Fatal("invalid source facts admitted")
			}
		})
	}
	for name, change := range map[string]func(*SnapshotModel){
		"altered key":         func(m *SnapshotModel) { m.TimelineEvents[0].ActivitySortKey = "other" },
		"missing frozen time": func(m *SnapshotModel) { m.TimelineEvents[0].RecordCreatedAt = "" },
		"hidden display text": func(m *SnapshotModel) { m.TimelineEvents[0].DisplayTimes.SourceTextDisplay = text("hidden source") },
		"wrong order":         func(m *SnapshotModel) { slices.Reverse(m.TimelineEvents) },
	} {
		t.Run(name, func(t *testing.T) {
			candidate, err := build(source)
			if err != nil {
				t.Fatal(err)
			}
			change(&candidate)
			if err := validateSnapshotModelIdentity(candidate); err == nil {
				t.Fatal("altered frozen ordering admitted")
			}
			if _, err := candidate.RedactionFields(); err == nil {
				t.Fatal("redaction accepted altered ordering")
			}
		})
	}
	// Required-null members are not optional on retained or portable input.
	encoded, err := canonicalJSON(model.TimelineEvents[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"activity_sort_ts", "date_entered_sort_key", "record_created_at", "display_times", "display_times.primary_display", "display_times.precision_label", "display_times.source_text_display"} {
		t.Run("omitted "+path, func(t *testing.T) {
			var object map[string]any
			if err := json.Unmarshal(encoded, &object); err != nil {
				t.Fatal(err)
			}
			if key, nested, ok := strings.Cut(path, "."); ok {
				delete(object[key].(map[string]any), nested)
			} else {
				delete(object, path)
			}
			raw, _ := json.Marshal(object)
			var decoded ReportingTimelineEvent
			if json.Unmarshal(raw, &decoded) == nil {
				t.Fatal("omitted member accepted")
			}
		})
	}
	for _, mutation := range []string{
		strings.Replace(string(encoded), `"record_id":`, `"timeline_event_id":`, 1),
		strings.Replace(string(encoded), `"display_times":{`, `"display_times":{"primary":null,`, 1),
		strings.Replace(string(encoded), `"display_times":{`, `"display_times":null,"retired_display_times":{`, 1),
		strings.Replace(string(encoded), `"activity_precision_rank":6`, `"activity_precision_rank":null`, 1),
		strings.Replace(string(encoded), `"record_id":`, `"record_id":"duplicate","record_id":`, 1),
	} {
		var decoded ReportingTimelineEvent
		if json.Unmarshal([]byte(mutation), &decoded) == nil {
			t.Fatal("retired, duplicate or null required member admitted")
		}
	}
	// Date, creation and identity tie-breaks never reverse with timestamp order.
	base := model.TimelineEvents[0]
	for _, pair := range [][2]ReportingTimelineEvent{
		{func() ReportingTimelineEvent { x := base; x.DateEnteredSortKey = text("2026-05-01"); return x }(), base},
		{base, func() ReportingTimelineEvent { x := base; x.RecordCreatedAt = "2026-05-02T00:00:00.000000Z"; return x }()},
		{base, func() ReportingTimelineEvent { x := base; x.RecordID = "zz-last"; return x }()},
	} {
		for _, descending := range []bool{false, true} {
			if compareTimelineEvents(pair[0], pair[1], descending, false) >= 0 {
				t.Fatal("secondary tie-break reversed")
			}
		}
	}
	// Sorting and materialization must not alias source-owned timestamp pointers.
	*facts[1].ActivitySortTS = "2020-01-01T00:00:00.000000Z"
	if *model.TimelineEvents[2].ActivitySortTS != "2026-05-03T00:00:00.000000Z" {
		t.Fatal("snapshot aliases mutable source facts")
	}
}
