package exportprovider

import (
	"errors"
	"time"
)

// TimelineSortFact contains source-owned ordering facts, never presentation
// text. Reporting validates this contract but does not parse source timestamps.
type TimelineSortFact struct {
	RecordID              string  `json:"record_id"`
	ActivitySortTS        *string `json:"activity_sort_ts"`
	ActivityParseState    string  `json:"activity_parse_state"`
	ActivityPrecisionRank int     `json:"activity_precision_rank"`
	DateEnteredSortKey    *string `json:"date_entered_sort_key"`
	RecordCreatedAt       string  `json:"record_created_at"`
	TimeBucket            string  `json:"time_bucket"`
}

func (fact TimelineSortFact) Validate() error {
	invalid := errors.New("invalid or missing timeline sort facts")
	if fact.RecordID == "" || !canonicalSortTimestamp(fact.RecordCreatedAt) {
		return invalid
	}
	if fact.DateEnteredSortKey != nil {
		value, err := time.Parse("2006-01-02", *fact.DateEnteredSortKey)
		if err != nil || value.Format("2006-01-02") != *fact.DateEnteredSortKey {
			return invalid
		}
	}
	if fact.ActivityParseState == "parsed" {
		if fact.ActivitySortTS == nil || !canonicalSortTimestamp(*fact.ActivitySortTS) || fact.ActivityPrecisionRank < 1 || fact.ActivityPrecisionRank > 6 || fact.TimeBucket != "parsed_time" {
			return invalid
		}
		return nil
	}
	switch fact.ActivityParseState {
	case "missing", "incomplete", "ambiguous", "unparseable":
		if fact.ActivitySortTS == nil && fact.ActivityPrecisionRank == 0 && fact.TimeBucket == "unresolved_time" {
			return nil
		}
	}
	return invalid
}

func canonicalSortTimestamp(text string) bool {
	value, err := time.Parse(time.RFC3339Nano, text)
	return err == nil && !value.IsZero() && value.Nanosecond()%1000 == 0 && value.UTC().Format("2006-01-02T15:04:05.000000Z") == text
}

func (fact TimelineSortFact) Clone() TimelineSortFact {
	if fact.ActivitySortTS != nil {
		value := *fact.ActivitySortTS
		fact.ActivitySortTS = &value
	}
	if fact.DateEnteredSortKey != nil {
		value := *fact.DateEnteredSortKey
		fact.DateEnteredSortKey = &value
	}
	return fact
}
