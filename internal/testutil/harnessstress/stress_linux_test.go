//go:build linux

package harnessstress

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

// fakeSlice records each repetition's environment and fails the listed
// repetitions, publishing one failed row result as a real slice would.
type fakeSlice struct {
	results string
	failing map[int]bool
	calls   []map[string]string
	fresh   [][]string
}

func (f *fakeSlice) run(_ context.Context, environment []string, fresh []string) (string, int) {
	env := map[string]string{}
	for _, entry := range environment {
		key, value, _ := strings.Cut(entry, "=")
		env[key] = value
	}
	f.calls = append(f.calls, env)
	f.fresh = append(f.fresh, fresh)
	if !f.failing[len(f.calls)] {
		return "pass", 0
	}
	rows := filepath.Join(f.results, env["CARTULARY_TEST_RUN_ID"], "rows")
	if err := os.MkdirAll(rows, 0o700); err != nil {
		panic(err)
	}
	data, _ := json.Marshal(map[string]string{"row_id": "module.example.integration.flaky", "terminal_state": "failed"})
	if err := os.WriteFile(filepath.Join(rows, "module.example.integration.flaky.json"), data, 0o600); err != nil {
		panic(err)
	}
	return "fail", 1
}

func environment(results, runID string, extra ...string) []string {
	return append([]string{"CARTULARY_TEST_RESULTS_DIR=" + results, "CARTULARY_TEST_RUN_ID=" + runID,
		"OWNER=module.example", "ROWS=module.example.integration.flaky", "LOAD=none", "CARTULARY_HARNESS_CACHE_MODE=normal"}, extra...)
}

func TestStressRepetitions(t *testing.T) {
	t.Run("every repetition runs after a failure and the summary lists each", func(t *testing.T) {
		results := t.TempDir()
		slice := &fakeSlice{results: results, failing: map[int]bool{2: true}}
		var stdout, stderr bytes.Buffer
		code := stress(t.Context(), t.TempDir(), "/nonexistent", environment(results, "stress-a", "REPEAT=3"), &stdout, &stderr, slice.run)
		if code != 1 || len(slice.calls) != 3 {
			t.Fatalf("exit %d after %d repetitions, want 1 after 3; stderr %q", code, len(slice.calls), stderr.String())
		}
		for i, env := range slice.calls {
			want := "stress-a-r0" + string(rune('1'+i))
			if env["CARTULARY_TEST_RUN_ID"] != want || env["CARTULARY_TEST_TARGET"] != "test-slice" || env["REPEAT"] != "" || env["LOAD"] != "" {
				t.Fatalf("repetition %d environment %v", i+1, env)
			}
			if !strings.Contains(env["CARTULARY_MAKE_INPUT_SOURCES"], "OWNER=cli") || !strings.Contains(env["CARTULARY_MAKE_INPUT_SOURCES"], "CARTULARY_HARNESS_CACHE_MODE=cli") {
				t.Fatalf("repetition %d sources %q", i+1, env["CARTULARY_MAKE_INPUT_SOURCES"])
			}
			if !slices.Equal(slice.fresh[i], []string{"module.example.integration.flaky"}) {
				t.Fatalf("repetition %d fresh rows %v", i+1, slice.fresh[i])
			}
		}
		var summary struct {
			Repetitions []repetition `json:"repetitions"`
			Passed      int          `json:"passed"`
			Failed      int          `json:"failed"`
		}
		data, err := os.ReadFile(filepath.Join(results, "stress-a", "test-stress", "stress-summary.json"))
		if err != nil || json.Unmarshal(data, &summary) != nil {
			t.Fatalf("summary: %v", err)
		}
		if summary.Passed != 2 || summary.Failed != 1 || len(summary.Repetitions) != 3 ||
			!slices.Equal(summary.Repetitions[1].FailedRows, []string{"module.example.integration.flaky"}) || summary.Repetitions[1].Status != "fail" {
			t.Fatalf("summary %+v", summary)
		}
		if lines := strings.Count(stdout.String(), "\n"); lines != 4 || !strings.Contains(stdout.String(), "[STRESS-SUMMARY]") {
			t.Fatalf("output %q", stdout.String())
		}
	})
	t.Run("all passing exits zero", func(t *testing.T) {
		results := t.TempDir()
		slice := &fakeSlice{results: results}
		if code := stress(t.Context(), t.TempDir(), "/nonexistent", environment(results, "stress-b", "REPEAT=2"), &bytes.Buffer{}, &bytes.Buffer{}, slice.run); code != 0 || len(slice.calls) != 2 {
			t.Fatalf("exit %d after %d repetitions", code, len(slice.calls))
		}
	})
	t.Run("usage errors exit two before any repetition", func(t *testing.T) {
		results := t.TempDir()
		for _, extra := range []string{"REPEAT=0", "REPEAT=51", "REPEAT=x", "LOAD=heavy"} {
			slice := &fakeSlice{results: results}
			if code := stress(t.Context(), t.TempDir(), "/nonexistent", environment(results, "stress-c", extra), &bytes.Buffer{}, &bytes.Buffer{}, slice.run); code != 2 || len(slice.calls) != 0 {
				t.Fatalf("%s: exit %d after %d repetitions", extra, code, len(slice.calls))
			}
		}
	})
	t.Run("an existing repetition run root is refused", func(t *testing.T) {
		results := t.TempDir()
		if err := os.Mkdir(filepath.Join(results, "stress-d-r01"), 0o700); err != nil {
			t.Fatal(err)
		}
		slice := &fakeSlice{results: results}
		if code := stress(t.Context(), t.TempDir(), "/nonexistent", environment(results, "stress-d"), &bytes.Buffer{}, &bytes.Buffer{}, slice.run); code != 2 || len(slice.calls) != 0 {
			t.Fatalf("exit %d after %d repetitions", code, len(slice.calls))
		}
	})
	t.Run("cancellation stops further repetitions", func(t *testing.T) {
		results := t.TempDir()
		ctx, cancel := context.WithCancel(t.Context())
		slice := &fakeSlice{results: results}
		run := func(ctx context.Context, environment []string, fresh []string) (string, int) {
			status, code := slice.run(ctx, environment, fresh)
			cancel()
			return status, code
		}
		if code := stress(ctx, t.TempDir(), "/nonexistent", environment(results, "stress-e", "REPEAT=5"), &bytes.Buffer{}, &bytes.Buffer{}, run); code != 130 || len(slice.calls) != 1 {
			t.Fatalf("exit %d after %d repetitions", code, len(slice.calls))
		}
	})
}

func TestStressLoadEndsWithTheCommand(t *testing.T) {
	directory := t.TempDir()
	script := filepath.Join(directory, "load")
	// The fake load generator starts a grandchild in its group, like a
	// generator's threads would, and records both process IDs.
	marker := filepath.Join(directory, "pids")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 300 &\necho $$ $! > "+marker+"\nwait\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	stop, err := startLoad(script, profiles["cpu"], filepath.Join(directory, "load-files"))
	if err != nil {
		t.Fatal(err)
	}
	var pids []int
	for deadline := time.Now().Add(10 * time.Second); time.Now().Before(deadline); time.Sleep(10 * time.Millisecond) {
		data, err := os.ReadFile(marker)
		if err == nil && len(strings.Fields(string(data))) == 2 {
			for _, field := range strings.Fields(string(data)) {
				var pid int
				if _, err := fmt.Sscan(field, &pid); err == nil {
					pids = append(pids, pid)
				}
			}
			break
		}
	}
	if len(pids) != 2 {
		t.Fatal("fake load generator did not start")
	}
	stop()
	// A killed orphan stays a zombie until init reaps it; it must not run.
	running := func(pid int) bool {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			return false
		}
		fields := strings.Fields(string(data[strings.LastIndexByte(string(data), ')')+1:]))
		return len(fields) > 0 && fields[0] != "Z" && fields[0] != "X"
	}
	for _, pid := range pids {
		for deadline := time.Now().Add(5 * time.Second); running(pid) && time.Now().Before(deadline); {
			time.Sleep(10 * time.Millisecond)
		}
		if running(pid) {
			t.Fatalf("load process %d survived the command", pid)
		}
	}
	if _, err := os.Stat(filepath.Join(directory, "load-files")); !os.IsNotExist(err) {
		t.Fatalf("load files remain: %v", err)
	}
}

func TestLoadProfiles(t *testing.T) {
	for name, profile := range profiles {
		if profile.name != name {
			t.Fatalf("profile %s names itself %s", name, profile.name)
		}
	}
	if profiles["mixed"].threadsPerCPU != profiles["cpu"].threadsPerCPU || profiles["mixed"].flushWriters != profiles["io"].flushWriters ||
		profiles["none"].threadsPerCPU != 0 || profiles["none"].flushWriters != 0 {
		t.Fatal("mixed must combine cpu and io; none must add no load")
	}
	if code := LoadMain([]string{"--profile", "io"}, &bytes.Buffer{}); code != 2 {
		t.Fatalf("an io profile without a directory exited %d", code)
	}
}
