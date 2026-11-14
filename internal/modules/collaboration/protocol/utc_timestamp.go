package protocol

import (
	"errors"
	"regexp"
	"time"
)

var utcTimestampPattern = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(?:\.[0-9]+)?(?:Z|[+-]00:00)$`)

// time.Parse accepts some non-RFC3339 spellings. Check wire grammar before
// delegating calendar and clock-range validation to the standard library.
func parseUTCInstant(value string) (time.Time, error) {
	if !utcTimestampPattern.MatchString(value) {
		return time.Time{}, errors.New("timestamp must be RFC3339 UTC")
	}
	instant, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, errors.New("timestamp must be RFC3339 UTC")
	}
	return instant, nil
}
