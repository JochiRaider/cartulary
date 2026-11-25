//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRuntimeHostPressure(t *testing.T) {
	t.Run("window", testPressureWindow)
	t.Run("publication", testPressurePublication)
}

func testPressureWindow(t *testing.T) {
	t.Run("shares come from total deltas and unreported resources are null", func(t *testing.T) {
		totals := map[string]string{
			"/proc/pressure/cpu": "some avg10=0.00 avg60=0.00 avg300=0.00 total=1000000\nfull avg10=0.00 avg60=0.00 avg300=0.00 total=0\n",
			"/proc/pressure/io":  "some avg10=0.00 avg60=0.00 avg300=0.00 total=2000000\nfull avg10=0.00 avg60=0.00 avg300=0.00 total=500000\n",
		}
		ring := newPressureRing(func() time.Duration { return 0 })
		ring.read = func(path string) ([]byte, error) {
			if text, ok := totals[path]; ok {
				return []byte(text), nil
			}
			return nil, errors.New("absent")
		}
		before := ring.totals()
		totals["/proc/pressure/cpu"] = "some total=1250000\nfull total=0\n"
		totals["/proc/pressure/io"] = "some total=2100000\nfull total=550000\n"
		sample := sampleBetween(before, ring.totals(), time.Second, 3*time.Second)
		if sample.AtMS != 3000 || *sample.CPUSome != 25 || *sample.IOSome != 10 || *sample.IOFull != 5 || sample.MemorySome != nil || sample.MemoryFull != nil {
			t.Fatalf("sample %+v", sample)
		}
	})
	t.Run("a window holds its interval and marks samples that left the ring", func(t *testing.T) {
		ring := newPressureRing(func() time.Duration { return 0 })
		ring.samples = make([]pressureSample, 4)
		for second := int64(1); second <= 6; second++ {
			ring.add(pressureSample{AtMS: second * 1000})
		}
		samples, truncated := ring.window(2*time.Second, 5*time.Second)
		if len(samples) != 3 || samples[0].AtMS != 3000 || samples[2].AtMS != 5000 || !truncated {
			t.Fatalf("window %+v truncated=%v", samples, truncated)
		}
		if samples, truncated := ring.window(4*time.Second, 6*time.Second); len(samples) != 3 || truncated {
			t.Fatalf("window %+v truncated=%v", samples, truncated)
		}
	})
	t.Run("a failed unit's window validates and lands beside its logs", func(t *testing.T) {
		now := 90 * time.Second
		ring := newPressureRing(func() time.Duration { return now })
		share := 12.5
		for second := int64(1); second <= 90; second++ {
			ring.add(pressureSample{AtMS: second * 1000, CPUSome: &share, IOSome: &share, IOFull: &share, MemorySome: &share, MemoryFull: &share})
		}
		logs := t.TempDir()
		if err := os.Chmod(logs, 0o700); err != nil {
			t.Fatal(err)
		}
		if err := ring.writeWindow(logs, "go:unit", 60*time.Second); err != nil {
			t.Fatal(err)
		}
		var window struct {
			Start        int64            `json:"window_start_ms"`
			End          int64            `json:"window_end_ms"`
			Availability string           `json:"availability"`
			Samples      []pressureSample `json:"samples"`
		}
		data, err := os.ReadFile(filepath.Join(logs, "host-pressure.json"))
		if err == nil {
			err = json.Unmarshal(data, &window)
		}
		if err != nil || window.Start != 30000 || window.End != 90000 || window.Availability != "available" || len(window.Samples) != 61 {
			t.Fatalf("window %+v: %v", window, err)
		}
	})
}

func testPressurePublication(t *testing.T) {
	for _, status := range []string{"failed", "passed"} {
		t.Run(status, func(t *testing.T) {
			p := planFixture(t)
			root := t.TempDir()
			if err := os.Chmod(root, 0o700); err != nil {
				t.Fatal(err)
			}
			publisher, err := newGraphPublication(p, root)
			if err != nil {
				t.Fatal(err)
			}
			defer publisher.Abort()
			publisher.Pressure = newPressureRing(func() time.Duration { return time.Second })
			result, err := executeGraph(t.Context(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal,
				Fixtures: &testFixtures{close: func(context.Context) error { return nil }},
				Execute: func(context.Context, executionRequest) unitResult {
					if status == "failed" {
						return unitResult{Status: "failed", ExitCode: 10, Failure: failed("product", "test_assertion_failure")}
					}
					return unitResult{Status: "passed"}
				}})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := publisher.Finish(result); err != nil {
				t.Fatal(err)
			}
			_, err = os.Stat(filepath.Join(root, "unit-logs", "unit-a", "host-pressure.json"))
			if retained := err == nil; retained != (status == "failed") {
				t.Fatalf("%s unit retained a window: %v", status, retained)
			}
		})
	}
}
