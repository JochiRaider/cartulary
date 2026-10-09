package suiteservices

import (
	"crypto/sha256"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// lifecycleClock uses the same Linux boot clock as the browser helper. Its
// hundredth-second resolution and suspend semantics differ from scheduler time.
// Clock unavailability is diagnostic incompleteness, not a lifecycle failure.
func lifecycleClock() (string, *string, *int64) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "unavailable", nil, nil
	}
	namespace, err := os.Readlink("/proc/self/ns/time")
	if err != nil {
		return "unavailable", nil, nil
	}
	raw, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unavailable", nil, nil
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return "unavailable", nil, nil
	}
	parts := strings.Split(fields[0], ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return "unavailable", nil, nil
	}
	seconds, err := strconv.ParseInt(parts[0], 10, 64)
	fraction, fractionErr := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || fractionErr != nil || seconds < 0 || seconds > (1<<53-1)/1000 || fraction < 0 || fraction > 99 {
		return "unavailable", nil, nil
	}
	milliseconds := seconds*1000 + fraction*10
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(strings.TrimSpace(string(boot))+":"+namespace)))
	return "linux_boottime_10ms", &digest, &milliseconds
}
