//go:build linux

package harnessruntime

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func observationRoot() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(f), "../../.."))
}
func observationFixture(t *testing.T) (*Plan, *harnessfixtures.Runtime, string, map[string]any) {
	t.Helper()
	root := observationRoot()
	policy, err := readInstrumentationPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	p := planFixture(t)
	manifest := p["manifest"].(map[string]any)
	manifest["instrumentation"].(map[string]any)["mode"] = "basic"
	digest, _ := semanticDigest(policy)
	manifest["instrumentation"].(map[string]any)["policy_digest"] = digest
	plan, err := decodePlan(planBytes(t, p))
	if err != nil {
		t.Fatal(err)
	}
	run := t.TempDir()
	if err := os.Chmod(run, 0700); err != nil {
		t.Fatal(err)
	}
	r, err := harnessfixtures.NewRuntime(root, run, "fixture", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := r.Close(); err != nil {
			t.Error(err)
		}
	})
	return plan, r, run, policy
}
func TestRuntimeObservations(t *testing.T) {
	t.Run("clock_bracket", func(t *testing.T) {
		c := newGraphClock()
		before := float64(monotonicNS()) / 1e6
		a := c.anchor()
		after := float64(monotonicNS()) / 1e6
		if a == nil || a["clock_identity"] != c.identity || a["local_before_ms"].(float64) < before || a["local_after_ms"].(float64) > after || a["boot_resolution_ms"] != 10 {
			t.Fatal("invalid producer bracket", a)
		}
		if a["local_before_ms"].(float64) > a["local_after_ms"].(float64) {
			t.Fatal("reversed anchor")
		}
	})
	t.Run("execution_shutdown_and_closure", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		s := newExecutionSession(r, run, p.facts.Manifest, policy, nil)
		identity := map[string]any{"schema_id": "cartulary.harness_launch_context.v1", "run_id": "fixture", "invocation_id": "00000000-0000-4000-8000-000000000001", "parent_invocation_id": "00000000-0000-4000-8000-000000000002", "unit_id": "unit:a", "command_id": nil, "row_ids": []string{}, "producer": "unit"}
		o := s.launch(identity)
		o.Spawned(os.Getpid())
		o.Settled("passed")
		if err := harnesslocal.PublishNew(filepath.Join(s.root, "slot-0"), []byte(`{"kind":"launch","kind":"activity"}`)); err != nil {
			t.Fatal(err)
		}
		s.capture()
		if b, _ := os.ReadFile(filepath.Join(s.root, "open")); string(b) != "0\n" {
			t.Fatal("producer channel not closed")
		}
		if err := s.publish(nil); err != nil {
			t.Fatal(err)
		}
		file := filepath.Join(run, "diagnostics/execution-index.json")
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var index map[string]any
		if json.Unmarshal(b, &index) != nil {
			t.Fatal("invalid index")
		}
		if index["status"] != "partial" || len(index["omitted_refs"].([]any)) != 1 || len(index["records"].([]any)) != 1 {
			t.Fatal("loss or reference closure misrepresented", index)
		}
		s.record(map[string]any{"kind": "late"})
		if err := s.publish(nil); err != nil {
			t.Fatal(err)
		}
		after, _ := os.ReadFile(file)
		if !bytes.Equal(b, after) {
			t.Fatal("publication changed after shutdown")
		}
	})
	t.Run("execution_bounds_and_incomplete", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		policy["maximum_execution_records"] = float64(1)
		s := newExecutionSession(r, run, p.facts.Manifest, policy, nil)
		for _, ref := range []string{"allocation:a", "allocation:b"} {
			s.record(map[string]any{"kind": "allocation", "ref": ref, "allocation_ref": ref, "unit_id": "unit:a", "capability": "none", "ownership": nil, "outcome": "incomplete"})
		}
		if err := s.publish(nil); err != nil {
			t.Fatal(err)
		}
		b, _ := os.ReadFile(filepath.Join(run, "diagnostics/execution-index.json"))
		var v map[string]any
		_ = json.Unmarshal(b, &v)
		if v["status"] != "truncated" || v["omitted_records"] != nil || len(v["records"].([]any)) != 1 {
			t.Fatal(v)
		}
	})
	t.Run("scan_independent_expectations", func(t *testing.T) {
		cases := []struct {
			name, text string
			want       bool
		}{{"plain", `{"cpu_tokens":"2","message":"<redacted>"}`, true}, {"redacted", `{"secret":"<redacted>"} PASSWORD=<redacted> postgres://u:<redacted>@h`, true}, {"token_substring", `{"tokens_token":"leak"}`, false}, {"json", `{"my_password":"leak"}`, false}, {"assignment", "TOKEN=leak", false}, {"url", "postgresql://user:leak@host", false}, {"registered", strings.Repeat("x", 65530) + "private-sentinel-12345", false}, {"unicode", "abc 中文\n", true}}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				a := t.TempDir()
				for _, p := range []string{a} {
					_ = os.Chmod(p, 0700)
					if err := os.WriteFile(filepath.Join(p, "data.json"), []byte(c.text), 0600); err != nil {
						t.Fatal(err)
					}
				}
				ours, err := scanRetained(a, []string{"private-sentinel-12345"})
				if (err == nil) != c.want {
					t.Fatalf("native scan: %v", err)
				}
				if c.want && (ours["policy_digest"] != "sha256:eaa74efae69ea73a6e0f18823a94ccc70a19b242e2cb2880e446611c3084ee11" || ours["scanned_files"] != 1 || ours["status"] != "pass") {
					t.Fatal("scan result differs from frozen independent expectations", ours)
				}
			})
		}
	})
	t.Run("scan_rejects_unsafe_entries", func(t *testing.T) {
		for _, kind := range []string{"symlink", "hardlink", "fifo", "public_file", "forbidden_name"} {
			t.Run(kind, func(t *testing.T) {
				root := t.TempDir()
				_ = os.Chmod(root, 0700)
				file := filepath.Join(root, "unsafe")
				switch kind {
				case "symlink":
					if err := os.Symlink("/dev/null", file); err != nil {
						t.Fatal(err)
					}
				case "hardlink":
					source := filepath.Join(t.TempDir(), "source")
					_ = os.WriteFile(source, []byte("safe"), 0600)
					if err := os.Link(source, file); err != nil {
						t.Fatal(err)
					}
				case "fifo":
					if err := syscall.Mkfifo(file, 0600); err != nil {
						t.Fatal(err)
					}
				case "public_file":
					_ = os.WriteFile(file, []byte("safe"), 0644)
				case "forbidden_name":
					_ = os.WriteFile(filepath.Join(root, "suite-lease.json"), []byte("safe"), 0600)
				}
				if _, err := scanRetained(root, nil); err == nil {
					t.Fatal("unsafe entry accepted")
				}
			})
		}
	})
	t.Run("live_quietness_and_terminal", func(t *testing.T) {
		p, _, run, _ := observationFixture(t)
		live := newLivePublisher(run, p, nil)
		defer live.stop()
		file := filepath.Join(run, "diagnostics/live-status.json")
		live.pause()
		before, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		var v map[string]any
		_ = json.Unmarshal(before, &v)
		if v["publication_mode"] != "paused_measurement" {
			t.Fatal(v)
		}
		live.update(engineSnapshot{Phase: "executing", ElapsedMS: 10, Units: []unitSnapshot{{ID: "unit:a", Status: "running", Activity: "executing", BlockingResources: []string{}, BlockingUnits: []string{}}}})
		time.Sleep(1100 * time.Millisecond)
		after, _ := os.ReadFile(file)
		if !bytes.Equal(before, after) {
			t.Fatal("publication during exclusive window")
		}
		live.resume()
		live.stop()
		after, err = os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := harnessschema.Decode(after)
		if err != nil || harnessschema.Validate("cartulary.harness_live_snapshot.v2", raw) != nil {
			t.Fatal("invalid live snapshot", err)
		}
		if live.publicationErr != nil {
			t.Fatal(live.publicationErr)
		}
	})
	t.Run("collector_forced_death_drains", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		c := newCollector(filepath.Join(observationRoot(), ".cache/cartulary/toolbin/cartulary-harness"), run, filepath.Join(t.TempDir(), "admission"), r, p, policy, monotonicNS(), nil)
		if c.cmd == nil {
			t.Fatal("collector not started")
		}
		if err := c.cmd.Process.Kill(); err != nil {
			t.Fatal(err)
		}
		if err := c.stop(); err != nil {
			t.Fatal("dead optional collector prevented required cleanup", err)
		}
		select {
		case <-c.done:
		default:
			t.Fatal("collector not reaped")
		}
		if c.snapshot()["availability"] != "collector_failed" {
			t.Fatal("loss reported available")
		}
		before, _ := os.ReadDir(filepath.Join(run, "diagnostics"))
		c.register(map[string]any{}, nil, nil, nil)
		c.resume()
		after, _ := os.ReadDir(filepath.Join(run, "diagnostics"))
		if len(before) != len(after) {
			t.Fatal("late publication")
		}
	})
}
