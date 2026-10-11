//go:build linux

package harnessruntime

import (
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"strings"
)

type graphClock struct{ identity string }

func newGraphClock() graphClock { return graphClock{"clock:" + uuid()} }
func monotonicNS() int64 {
	var t unix.Timespec
	if unix.ClockGettime(unix.CLOCK_MONOTONIC, &t) != nil {
		panic("monotonic clock unavailable")
	}
	return t.Nano()
}
func bootReading() (string, int64, error) {
	boot, err := os.ReadFile("/proc/sys/kernel/random/boot_id")
	if err != nil {
		return "", 0, err
	}
	ns, err := os.Readlink("/proc/self/ns/time")
	if err != nil {
		return "", 0, err
	}
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "", 0, err
	}
	fields := strings.Fields(string(data))
	if len(fields) == 0 {
		return "", 0, errPlan
	}
	parts := strings.Split(fields[0], ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return "", 0, errPlan
	}
	sec, e := strconv.ParseInt(parts[0], 10, 64)
	frac, f := strconv.ParseInt(parts[1], 10, 64)
	if e != nil || f != nil || sec < 0 || sec > (1<<53-1)/1000 || frac < 0 || frac > 99 {
		return "", 0, errPlan
	}
	return bytesDigest([]byte(strings.TrimSpace(string(boot)) + ":" + ns)), sec*1000 + frac*10, nil
}
func (c graphClock) anchor() map[string]any {
	before := float64(monotonicNS()) / 1e6
	id, boot, err := bootReading()
	after := float64(monotonicNS()) / 1e6
	if err != nil || after < before {
		return nil
	}
	return map[string]any{"clock_identity": c.identity, "boot_identity": id, "local_before_ms": before, "local_after_ms": after, "boot_ms": boot, "boot_resolution_ms": 10}
}
