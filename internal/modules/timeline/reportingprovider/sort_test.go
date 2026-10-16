package reportingprovider

import (
	"testing"
	"time"
)

func TestTimelineReportingSortFacts(t *testing.T) {
	created := time.Date(2026, 5, 1, 1, 2, 3, 123456000, time.FixedZone("source", 3600))
	for _, tc := range []struct {
		name, utc, local, date, state, timestamp, day string
		rank                                          int
	}{
		{name: "missing", state: "missing"},
		{name: "minute", utc: "2026-05-02T01:02Z", state: "parsed", timestamp: "2026-05-02T01:02:00.000000Z", rank: 5},
		{name: "seconds", utc: "2026-05-02 01:02:03Z", state: "parsed", timestamp: "2026-05-02T01:02:03.000000Z", rank: 6},
		{name: "fraction", utc: "2026-05-02T01:02:03.000001Z", state: "parsed", timestamp: "2026-05-02T01:02:03.000001Z", rank: 6},
		{name: "offset day rollover", local: "2026-05-02T00:30+02:00", date: "2026-05-02T00:30+02:00", state: "parsed", timestamp: "2026-05-01T22:30:00.000000Z", day: "2026-05-01", rank: 5},
		{name: "UTC precedence", utc: "2026-05-02T01:02Z", local: "2020-01-01T00:00+01:00", state: "parsed", timestamp: "2026-05-02T01:02:00.000000Z", rank: 5},
		{name: "local fallback", utc: "unresolved source", local: "2026-05-02T01:02:03-02:00", state: "parsed", timestamp: "2026-05-02T03:02:03.000000Z", rank: 6},
		{name: "no timezone inference", local: "2026-05-02T01:02", state: "ambiguous"},
		{name: "date is incomplete activity", utc: "2026-05-02", date: "2026-05-02", state: "incomplete", day: "2026-05-02"},
		{name: "year is incomplete", utc: "2026", state: "incomplete"},
		{name: "month is incomplete", utc: "2026-05", state: "incomplete"},
		{name: "hour is incomplete", utc: "2026-05-02T01", state: "incomplete"},
		{name: "first unresolved input", utc: "unrecognized", local: "2026-05-02T01:02", state: "unparseable"},
		{name: "invalid calendar", utc: "2026-02-30T01:02Z", date: "2026-02-30", state: "unparseable"},
		{name: "no trimming", utc: " 2026-05-02T01:02Z", state: "unparseable"},
		{name: "no fractional truncation", utc: "2026-05-02T01:02:03.1234567Z", state: "unparseable"},
		{name: "no excessive zero precision", utc: "2026-05-02T01:02:03.1234560Z", state: "unparseable"},
		{name: "no decimal comma", utc: "2026-05-02T01:02:03,1Z", state: "unparseable"},
		{name: "no leap second", utc: "2026-05-02T01:02:60Z", state: "unparseable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fact, err := DeriveSortFact("source-id", created, &tc.utc, &tc.local, &tc.date)
			if err != nil {
				t.Fatal(err)
			}
			if fact.RecordID != "source-id" || fact.RecordCreatedAt != "2026-05-01T00:02:03.123456Z" || fact.ActivityParseState != tc.state || fact.ActivityPrecisionRank != tc.rank {
				t.Fatalf("facts = %+v", fact)
			}
			if tc.timestamp == "" {
				if fact.ActivitySortTS != nil || fact.TimeBucket != "unresolved_time" {
					t.Fatal("fabricated time", fact)
				}
			} else if fact.ActivitySortTS == nil || *fact.ActivitySortTS != tc.timestamp || fact.TimeBucket != "parsed_time" {
				t.Fatal("lost exact source time", fact)
			}
			if tc.day == "" {
				if fact.DateEnteredSortKey != nil {
					t.Fatal("fabricated day")
				}
			} else if fact.DateEnteredSortKey == nil || *fact.DateEnteredSortKey != tc.day {
				t.Fatal("lost day key")
			}
		})
	}
	for _, created := range []time.Time{{}, created.Add(time.Nanosecond)} {
		if _, err := DeriveSortFact("id", created, nil, nil, nil); err == nil {
			t.Fatal("invalid immutable creation time admitted")
		}
	}
	if _, err := DeriveSortFact("", created, nil, nil, nil); err == nil {
		t.Fatal("missing record identity admitted")
	}
}
