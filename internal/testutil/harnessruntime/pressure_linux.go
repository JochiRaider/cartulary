//go:build linux

package harnessruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// TH-HARNESS-REQ-870: every graph keeps the last hour of host pressure in
// memory, sampled once a second, and a failed unit retains the window it
// ran in beside its logs. The ring is diagnostic only: it never feeds
// scheduling, admission, outcomes or required evidence.
const (
	pressureSchema   = "cartulary.harness_host_pressure_window.v1"
	pressureInterval = time.Second
	pressureCapacity = 3600
	// pressureLeadIn precedes a unit's admission, so the window shows the
	// load the unit started into.
	pressureLeadIn = 30 * time.Second
)

// pressureSample is the share of one interval during which some or all
// non-idle tasks stalled on a resource, from the kernel's cumulative PSI
// totals. A resource the kernel does not report is null.
type pressureSample struct {
	AtMS       int64    `json:"at_ms"`
	CPUSome    *float64 `json:"cpu_some_percent"`
	IOSome     *float64 `json:"io_some_percent"`
	IOFull     *float64 `json:"io_full_percent"`
	MemorySome *float64 `json:"memory_some_percent"`
	MemoryFull *float64 `json:"memory_full_percent"`
}

type pressureTotals map[string]int64 // "<resource>_<some|full>" → cumulative µs

type pressureRing struct {
	mu      sync.Mutex
	samples []pressureSample
	next    int
	full    bool
	read    func(string) ([]byte, error)
	now     func() time.Duration
}

func newPressureRing(now func() time.Duration) *pressureRing {
	return &pressureRing{samples: make([]pressureSample, pressureCapacity), read: os.ReadFile, now: now}
}

// run samples until ctx ends. A host without PSI yields null samples, which
// a window reports as unavailable.
func (r *pressureRing) run(ctx context.Context) {
	ticker := time.NewTicker(pressureInterval)
	defer ticker.Stop()
	previous, at := r.totals(), r.now()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		current, now := r.totals(), r.now()
		r.add(sampleBetween(previous, current, now-at, now))
		previous, at = current, now
	}
}

func (r *pressureRing) totals() pressureTotals {
	totals := pressureTotals{}
	for _, resource := range []string{"cpu", "io", "memory"} {
		data, err := r.read("/proc/pressure/" + resource)
		if err != nil {
			continue
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			fields := strings.Fields(string(line))
			if len(fields) == 0 || fields[0] != "some" && fields[0] != "full" {
				continue
			}
			for _, field := range fields[1:] {
				if value, ok := strings.CutPrefix(field, "total="); ok {
					if n, err := strconv.ParseInt(value, 10, 64); err == nil && n >= 0 {
						totals[resource+"_"+fields[0]] = n
					}
				}
			}
		}
	}
	return totals
}

func sampleBetween(previous, current pressureTotals, elapsed, at time.Duration) pressureSample {
	share := func(key string) *float64 {
		before, ok1 := previous[key]
		after, ok2 := current[key]
		if !ok1 || !ok2 || after < before || elapsed <= 0 {
			return nil
		}
		// Kernel totals are microseconds; round to a tenth of a percent.
		percent := float64(int64(float64(after-before)/float64(elapsed.Microseconds())*1000+0.5)) / 10
		percent = min(percent, 100)
		return &percent
	}
	return pressureSample{AtMS: at.Milliseconds(), CPUSome: share("cpu_some"), IOSome: share("io_some"), IOFull: share("io_full"),
		MemorySome: share("memory_some"), MemoryFull: share("memory_full")}
}

func (r *pressureRing) add(sample pressureSample) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.samples[r.next] = sample
	r.next = (r.next + 1) % len(r.samples)
	r.full = r.full || r.next == 0
}

// window returns the retained samples taken in [from, to], oldest first, and
// whether earlier samples of that interval had already left the ring.
func (r *pressureRing) window(from, to time.Duration) ([]pressureSample, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ordered := r.samples[:r.next]
	if r.full {
		ordered = append(append([]pressureSample{}, r.samples[r.next:]...), r.samples[:r.next]...)
	}
	out := []pressureSample{}
	for _, sample := range ordered {
		if sample.AtMS >= from.Milliseconds() && sample.AtMS <= to.Milliseconds() {
			out = append(out, sample)
		}
	}
	truncated := r.full && len(ordered) > 0 && ordered[0].AtMS > from.Milliseconds()
	return out, truncated
}

// writeWindow retains a failed unit's pressure window in its log directory.
// It reports failure to its caller, which leaves the unit's outcome alone.
func (r *pressureRing) writeWindow(logRoot, unitID string, admitted time.Duration) error {
	to := r.now()
	from := max(0, admitted-pressureLeadIn)
	samples, truncated := r.window(from, to)
	available := false
	for _, sample := range samples {
		available = available || sample.CPUSome != nil || sample.IOSome != nil || sample.MemorySome != nil
	}
	document := map[string]any{
		"schema_id": pressureSchema, "unit_id": unitID, "clock": "graph_monotonic",
		"interval_ms": pressureInterval.Milliseconds(), "window_start_ms": from.Milliseconds(), "window_end_ms": to.Milliseconds(),
		"admitted_ms": admitted.Milliseconds(), "availability": map[bool]string{true: "available", false: "unavailable"}[available],
		"truncated": truncated, "samples": samples,
	}
	data, err := json.Marshal(document)
	if err != nil {
		return err
	}
	value, err := harnessschema.Decode(data)
	if err == nil {
		err = harnessschema.Validate(pressureSchema, value)
	}
	if err != nil {
		return err
	}
	return harnesslocal.ReplaceFile(filepath.Join(logRoot, "host-pressure.json"), append(data, '\n'))
}
