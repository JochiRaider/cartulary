package reporting

import (
	"cmp"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
)

const unresolvedTimelineTimestamp = "9999-12-31T23:59:59.999999Z"

// The serialized key is a canonical JSON tuple for traceability. Comparisons
// use typed members: JSON's textual representation is not an ordering codec.
func timelineSortKey(fact exportprovider.TimelineSortFact) string {
	bucket, timestamp, date := timelineSortParts(fact)
	encoded, _ := canonicalJSON([]any{bucket, timestamp, -fact.ActivityPrecisionRank, date, fact.RecordCreatedAt, fact.RecordID})
	return string(encoded)
}

func timelineSortParts(fact exportprovider.TimelineSortFact) (int, string, string) {
	bucket, timestamp, date := 1, unresolvedTimelineTimestamp, "9999-12-31"
	if fact.TimeBucket == "parsed_time" {
		bucket = 0
	}
	if fact.ActivitySortTS != nil {
		timestamp = *fact.ActivitySortTS
	}
	if fact.DateEnteredSortKey != nil {
		date = *fact.DateEnteredSortKey
	}
	return bucket, timestamp, date
}

func (event ReportingTimelineEvent) sortFact() exportprovider.TimelineSortFact {
	bucket := "unresolved_time"
	if event.ActivityParseState == "parsed" {
		bucket = "parsed_time"
	}
	return exportprovider.TimelineSortFact{
		RecordID: event.RecordID, ActivitySortTS: event.ActivitySortTS,
		ActivityParseState: event.ActivityParseState, ActivityPrecisionRank: event.ActivityPrecisionRank,
		DateEnteredSortKey: event.DateEnteredSortKey, RecordCreatedAt: event.RecordCreatedAt, TimeBucket: bucket,
	}
}

func compareTimelineEvents(left, right ReportingTimelineEvent, descending, unresolvedFirst bool) int {
	a, b := left.sortFact(), right.sortFact()
	ab, ats, ad := timelineSortParts(a)
	bb, bts, bd := timelineSortParts(b)
	if order := cmp.Compare(ab, bb); order != 0 {
		if unresolvedFirst {
			return -order
		}
		return order
	}
	if order := cmp.Compare(ats, bts); order != 0 {
		if descending && a.ActivitySortTS != nil && b.ActivitySortTS != nil {
			return -order
		}
		return order
	}
	if order := cmp.Compare(b.ActivityPrecisionRank, a.ActivityPrecisionRank); order != 0 {
		return order
	}
	if order := cmp.Compare(ad, bd); order != 0 {
		return order
	}
	if order := cmp.Compare(a.RecordCreatedAt, b.RecordCreatedAt); order != 0 {
		return order
	}
	return cmp.Compare(a.RecordID, b.RecordID)
}

func validateSnapshotTimelineEvents(events []ReportingTimelineEvent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	for i, event := range events {
		fact := event.sortFact()
		if event.SchemaID != "cartulary.reporting_timeline_event.v1" || fact.Validate() != nil || event.ActivitySortKey != timelineSortKey(fact) {
			issues.add(newRenderValidationError("export_model_invalid", "timeline_sort_key_missing"))
		}
		// A snapshot has no template or redaction selection. Display values
		// belong to render derivation, not a source-text side channel.
		if event.DisplayTimes != (ReportingDisplayTimes{}) {
			issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value"))
		}
		if i > 0 && compareTimelineEvents(events[i-1], event, false, false) >= 0 {
			issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value"))
		}
	}
	return issues.finish()
}
