package reportingprovider

import (
	"errors"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/timecontract"
)

// DeriveSortFact interprets Timeline source time under the Core snapshot
// contract. It never reads a projection, viewer timezone or mutable clock.
func DeriveSortFact(id string, createdAt time.Time, utcText, localText, dateText *string) (exportprovider.TimelineSortFact, error) {
	if createdAt.IsZero() || createdAt.Nanosecond()%1000 != 0 {
		return exportprovider.TimelineSortFact{}, errors.New("invalid source record creation time")
	}
	fact := exportprovider.TimelineSortFact{RecordID: id, RecordCreatedAt: createdAt.UTC().Format("2006-01-02T15:04:05.000000Z"), ActivityParseState: "missing", TimeBucket: "unresolved_time"}
	for i, text := range []*string{utcText, localText} {
		if text == nil || *text == "" {
			continue
		}
		var parsed time.Time
		var ok bool
		if i == 0 {
			parsed, ok = timecontract.ParseUTC(text)
		} else {
			parsed, _, ok = timecontract.ParseLocalOffset(text)
		}
		if ok && sourcePrecisionSupported(*text) && parsed.Nanosecond()%1000 == 0 {
			value := parsed.UTC().Format("2006-01-02T15:04:05.000000Z")
			fact.ActivitySortTS = &value
			fact.ActivityParseState, fact.TimeBucket = "parsed", "parsed_time"
			fact.ActivityPrecisionRank = 5
			if len(*text) > 16 && (*text)[16] == ':' {
				fact.ActivityPrecisionRank = 6
			}
			break
		}
		if fact.ActivityParseState == "missing" {
			fact.ActivityParseState = unresolvedState(*text, i == 1)
		}
	}
	if dateText != nil {
		day, err := time.Parse("2006-01-02", *dateText)
		if err == nil {
			value := day.Format("2006-01-02")
			fact.DateEnteredSortKey = &value
		} else {
			parsed, ok := timecontract.ParseUTC(dateText)
			if !ok {
				parsed, _, ok = timecontract.ParseLocalOffset(dateText)
			}
			if ok && sourcePrecisionSupported(*dateText) {
				value := parsed.UTC().Format("2006-01-02")
				fact.DateEnteredSortKey = &value
			}
		}
	}
	return fact, fact.Validate()
}

func sourcePrecisionSupported(text string) bool {
	if strings.Contains(text, ",") {
		return false
	}
	if dot := strings.IndexByte(text, '.'); dot >= 0 {
		end := dot + 1
		for end < len(text) && text[end] >= '0' && text[end] <= '9' {
			end++
		}
		return end-dot-1 >= 1 && end-dot-1 <= 6
	}
	return true
}

func unresolvedState(text string, local bool) string {
	if local && sourcePrecisionSupported(text) {
		for _, layout := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", "2006-01-02 15:04", "2006-01-02 15:04:05"} {
			if _, err := time.Parse(layout, text); err == nil {
				return "ambiguous"
			}
		}
	}
	for _, layout := range []string{"2006", "2006-01", "2006-01-02", "2006-01-02T15", "2006-01-02 15"} {
		if _, err := time.Parse(layout, text); err == nil {
			return "incomplete"
		}
	}
	return "unparseable"
}
