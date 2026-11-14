package evidence

import (
	"strings"
	"time"
)

func portableSourceClock(raw any) (time.Time, bool) {
	value, ok := raw.(string)
	clock, err := time.Parse(time.RFC3339Nano, value)
	_, offset := clock.Zone()
	valid := ok && err == nil && offset == 0 && clock.Year() >= 1 && clock.Year() <= 9999 && clock.Nanosecond()%1000 == 0 &&
		(value == clock.UTC().Format(time.RFC3339Nano) || value == strings.TrimSuffix(clock.UTC().Format(time.RFC3339Nano), "Z")+"+00:00")
	return clock.UTC(), valid
}
