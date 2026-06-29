package timecontract

import "time"

var utcLayouts = []string{
	"2006-01-02T15:04Z",
	"2006-01-02T15:04:05Z",
	"2006-01-02 15:04Z",
	"2006-01-02 15:04:05Z",
}

var localWallLayouts = []string{
	"2006-01-02T15:04",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02 15:04:05",
}

var localOffsetLayouts = []string{
	"2006-01-02T15:04-07:00",
	"2006-01-02T15:04:05-07:00",
}

func ParseUTC(text *string) (time.Time, bool) {
	if text == nil || *text == "" {
		return time.Time{}, false
	}
	for _, layout := range utcLayouts {
		parsed, err := time.Parse(layout, *text)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

type LocalParseResult struct {
	UTC           time.Time
	OffsetSeconds int
	HasWireOffset bool
}

func ParseLocalWithFixedOffset(text *string, offsetMinutes int) (LocalParseResult, bool) {
	if text == nil || *text == "" {
		return LocalParseResult{}, false
	}
	location := time.FixedZone("", offsetMinutes*60)
	for _, layout := range localWallLayouts {
		parsed, err := time.ParseInLocation(layout, *text, location)
		if err == nil {
			return LocalParseResult{
				UTC:           parsed.UTC(),
				OffsetSeconds: offsetMinutes * 60,
			}, true
		}
	}
	if parsed, offsetSeconds, ok := ParseLocalOffset(text); ok {
		return LocalParseResult{
			UTC:           parsed,
			OffsetSeconds: offsetSeconds,
			HasWireOffset: true,
		}, true
	}
	return LocalParseResult{}, false
}

func ParseLocalOffset(text *string) (time.Time, int, bool) {
	if text == nil || *text == "" {
		return time.Time{}, 0, false
	}
	for _, layout := range localOffsetLayouts {
		parsed, err := time.Parse(layout, *text)
		if err == nil {
			_, offsetSeconds := parsed.Zone()
			return parsed.UTC(), offsetSeconds, true
		}
	}
	return time.Time{}, 0, false
}

func FormatUTC(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}

func FormatLocal(value time.Time, offsetMinutes int) string {
	location := time.FixedZone("", offsetMinutes*60)
	return value.UTC().In(location).Format("2006-01-02T15:04:05-07:00")
}
