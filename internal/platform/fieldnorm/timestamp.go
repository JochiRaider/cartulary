package fieldnorm

import "time"

func NormalizeTimestampInstant(raw string) (time.Time, bool) {
	if raw == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	if err != nil {
		return time.Time{}, false
	}
	return parsed.UTC(), true
}

func NormalizeTimestampInstantText(raw string) (string, bool) {
	parsed, ok := NormalizeTimestampInstant(raw)
	if !ok {
		return "", false
	}
	return parsed.UTC().Format(time.RFC3339Nano), true
}
