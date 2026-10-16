//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func canonicalFixture(t *testing.T) (*Plan, []map[string]any) {
	t.Helper()
	a, b := engineUnit("a", 10), engineUnit("b", 10)
	b.Kind = "readiness"
	p := enginePlan(a, b)
	p.facts.Capacities["cpu"] = 2
	events := []map[string]any{}
	emit := func(at int64, name, id, status string, extra map[string]any) {
		u := a
		if id == "b" {
			u = b
		}
		if id == "harness:run" {
			u = unit{Needs: []string{}, Services: []string{}, Claims: map[string]int64{}}
		}
		e := map[string]any{"schema_id": "cartulary.harness_unit_event.v4", "seq": len(events) + 1, "monotonic_ms": at, "event": name, "unit_id": id, "status": status, "needs": u.Needs, "resource_claims": u.Claims, "service_dependencies": u.Services}
		for k, v := range extra {
			e[k] = v
		}
		events = append(events, e)
	}
	emit(0, "run_started", "harness:run", "running", nil)
	for _, u := range []unit{a, b} {
		emit(0, "queued", u.ID, "pending", map[string]any{"unit_kind": u.Kind, "row_ids": []string{}})
		emit(0, "eligible", u.ID, "pending", nil)
		emit(0, "wait_started", u.ID, "pending", waitDetails{"rank", []string{}, []string{}}.fields())
	}
	emit(10, "cache_bypass", "a", "pending", map[string]any{"cache_reason": "policy_none"})
	emit(10, "wait_ended", "a", "pending", waitDetails{"rank", []string{}, []string{}}.fields())
	emit(10, "admitted", "a", "running", nil)
	emit(10, "started", "a", "running", nil)
	emit(10, "wait_ended", "b", "pending", waitDetails{"rank", []string{}, []string{}}.fields())
	emit(10, "wait_started", "b", "pending", waitDetails{"resources", []string{"cpu"}, []string{"a"}}.fields())
	emit(15, "fixture_acquired", "a", "running", map[string]any{"fixture_lease_id": "fixture"})
	emit(18, "wait_ended", "b", "pending", waitDetails{"resources", []string{"cpu"}, []string{"a"}}.fields())
	emit(18, "wait_started", "b", "pending", waitDetails{"rank", []string{}, []string{}}.fields())
	emit(20, "wait_ended", "b", "pending", waitDetails{"rank", []string{}, []string{}}.fields())
	emit(20, "cache_bypass", "b", "pending", map[string]any{"cache_reason": "policy_none"})
	emit(20, "admitted", "b", "running", nil)
	emit(20, "started", "b", "running", nil)
	emit(30, "completed", "a", "passed", nil)
	emit(40, "completed", "b", "passed", nil)
	emit(40, "cleanup_started", "harness:run", "running", nil)
	emit(45, "cleanup_completed", "harness:run", "passed", nil)
	emit(45, "run_completed", "harness:run", "passed", nil)
	return p, events
}
func TestRuntimeCanonicalAccounting(t *testing.T) {
	p, events := canonicalFixture(t)
	f := newCanonicalFacts(p)
	for _, e := range events {
		if err := f.consume(e); err != nil {
			t.Fatal(e, err)
		}
	}
	if err := f.validate(); err != nil {
		t.Fatal(err)
	}
	want := map[string]int64{"setup_ms": 20, "fixture_ms": 5, "execution_ms": 15, "collation_ms": 0, "wrapper_ms": 5, "unattributed_ms": 0, "resource_blocking_ms": 8, "process_count": 2}
	if got := f.timing([]string{"a", "b"}, 45, true); !reflect.DeepEqual(got, want) {
		t.Fatalf("timing %v want %v", got, want)
	}
	if path, duration := f.criticalPath([]string{"a", "b"}); !slices.Equal(path, []string{"b"}) || duration != 40 {
		t.Fatal(path, duration)
	}
	if got := intervalUnion([]timeInterval{{start: 10, end: 30}, {start: 20, end: 40}}); got != 30 {
		t.Fatal(got)
	}
	if w := f.waits["b"]; w.count != 3 || w.byReason["resources"] != 8 || w.byReason["rank"] != 12 {
		t.Fatal(w)
	}
	pressure := f.pressure()
	if pressure["peak_use"].(map[string]int64)["cpu"] != 2 || pressure["saturation_ms"].(map[string]int64)["cpu"] != 10 {
		t.Fatal(pressure)
	}
	for _, kind := range []string{"duplicate", "reversal", "missing_wait", "forged_identity", "missing_terminal", "old_schema", "wait_gap", "wait_overlap", "wait_mismatch", "foreign_holder"} {
		t.Run(kind, func(t *testing.T) {
			bytes, _ := json.Marshal(events)
			var altered []map[string]any
			_ = json.Unmarshal(bytes, &altered)
			switch kind {
			case "old_schema":
				altered[0]["schema_id"] = "cartulary.harness_unit_event.v3"
			case "wait_gap", "wait_overlap", "wait_mismatch", "foreign_holder":
				for _, e := range altered {
					if e["unit_id"] == "b" && e["event"] == "wait_started" && e["wait_reason"] == "resources" {
						switch kind {
						case "wait_gap":
							e["monotonic_ms"] = float64(11)
						case "wait_overlap":
							e["event"] = "wait_ended"
						case "wait_mismatch":
							e["blocking_resources"] = []any{"memory"}
						case "foreign_holder":
							e["blocking_unit_ids"] = []any{"foreign"}
						}
						break
					}
				}
			case "duplicate":
				altered[1]["seq"] = float64(1)
			case "reversal":
				altered[12]["monotonic_ms"] = float64(1)
			case "missing_wait":
				altered[3]["event"] = "cache_bypass"
			case "forged_identity":
				altered[1]["unit_id"] = "foreign"
			case "missing_terminal":
				altered = altered[:len(altered)-1]
			}
			facts := newCanonicalFacts(p)
			var err error
			for _, e := range altered {
				if err = facts.consume(e); err != nil {
					break
				}
			}
			if err == nil {
				err = facts.validate()
			}
			if err == nil {
				t.Fatal("malformed accounting accepted")
			}
		})
	}
}
func TestRuntimeCanonicalNodeParity(t *testing.T) {
	p, events := canonicalFixture(t)
	facts := newCanonicalFacts(p)
	file := filepath.Join(t.TempDir(), "events.ndjson")
	stream := []byte{}
	for _, e := range events {
		if err := facts.consume(e); err != nil {
			t.Fatal(err)
		}
		bytes, _ := json.Marshal(e)
		stream = append(stream, append(bytes, '\n')...)
	}
	if err := os.WriteFile(file, stream, 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root = filepath.Clean(filepath.Join(root, "../../.."))
	node := os.Getenv("NODE_BIN")
	if node == "" {
		t.Fatal("Make must select NODE_BIN")
	}
	graph, _ := json.Marshal(p.facts.Graph)
	capacities, _ := json.Marshal(p.facts.Capacities)
	script := `import {reduceCanonicalUnitIntervals,canonicalTimingAccounting,projectResourcePressure,actualCriticalPath} from './tools/harness/evidence-accounting/index.mjs';const f=await reduceCanonicalUnitIntervals(process.argv[1]);const g=JSON.parse(process.argv[2]);console.log(JSON.stringify({timing:canonicalTimingAccounting(f,g,45),pressure:projectResourcePressure(f,g,JSON.parse(process.argv[3])),path:actualCriticalPath(g,f.intervals)}));`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, node, "--input-type=module", "-e", script, file, string(graph), string(capacities))
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Node parity: %v %s", err, output)
	}
	var expected, actual any
	if err := json.Unmarshal(output, &expected); err != nil {
		t.Fatal(err)
	}
	path, _ := facts.criticalPath([]string{"a", "b"})
	bytes, _ := json.Marshal(map[string]any{"timing": facts.timing([]string{"a", "b"}, 45, true), "pressure": facts.pressure(), "path": path})
	_ = json.Unmarshal(bytes, &actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Go %s Node %s", bytes, output)
	}
}
func TestRuntimeRequiredPublication(t *testing.T) {
	t.Run("cache_cancellation_publishes_terminal_without_execution", func(t *testing.T) {
		p, err := decodePlan(planBytes(t, planFixture(t)))
		if err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		if err := os.Chmod(root, 0700); err != nil {
			t.Fatal(err)
		}
		publisher, err := newGraphPublication(p, root)
		if err != nil {
			t.Fatal(err)
		}
		defer publisher.Abort()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cache := &testCache{lookup: func(ctx context.Context, _ unit) (cacheResult, error) {
			cancel()
			return cacheResult{}, ctx.Err()
		}}
		result, err := executeGraph(ctx, p, enginePorts{Cache: cache, Event: publisher.Event, Terminal: publisher.Terminal, Execute: func(context.Context, executionRequest) unitResult {
			t.Error("cancelled lookup launched a producer")
			return unitResult{Status: "passed"}
		}})
		if err != nil {
			t.Fatal("cancelled lookup aborted canonical publication", err)
		}
		summary, err := publisher.Finish(result)
		if err != nil || summary["status"] != "cancelled" || summary["failure_class"] != "interrupted" {
			t.Fatal(summary, err)
		}
		for _, name := range []string{"run-summary.json", "unit-events.ndjson", "cleanup-results.json"} {
			if _, err := os.Stat(filepath.Join(root, name)); err != nil {
				t.Fatal(name, err)
			}
		}
	})
	p, err := decodePlan(planBytes(t, planFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	publisher, err := newGraphPublication(p, root)
	if err != nil {
		t.Fatal(err)
	}
	defer publisher.Abort()
	contextBytes, err := os.ReadFile(filepath.Join(root, "execution-context.json"))
	if err != nil {
		t.Fatal("context was not published before child execution", err)
	}
	var retained map[string]any
	if err := json.Unmarshal(contextBytes, &retained); err != nil {
		t.Fatal(err)
	}
	if retained["manifest_digest"] != p.context["manifest_digest"] || retained["graph"] == nil || retained["projections"] == nil || len(retained["source_inputs"].([]any)) == 0 {
		t.Fatal("incomplete retained execution context")
	}
	if strings.Contains(string(contextBytes), "/private/node") || retained["bindings"] != nil || retained["cache_inputs"] != nil {
		t.Fatal("private launch state was retained")
	}
	if err := writeEvidence(root, "execution-context.json", "cartulary.harness_execution_context.v1", retained, false); err == nil {
		t.Fatal("immutable context was replaced")
	}
	publisher.Forbidden = func() ([]string, error) { return []string{"private-sentinel"}, nil }
	prefixes := make(chan string, len(p.facts.Graph.Units))
	result, err := executeGraph(context.Background(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal, Execute: func(context.Context, executionRequest) unitResult {
		prefix, err := publisher.Prefix()
		if err != nil {
			return unitResult{Status: "failed", ExitCode: 11, Failure: failed("artifact", "artifact_error")}
		}
		prefixes <- prefix
		return unitResult{Status: "passed", Stdout: []byte("PASSWORD=private-sentinel\npostgres://user:password@host/db")}
	}})
	if err != nil {
		t.Fatal(err)
	}
	close(prefixes)
	count := 0
	for encoded := range prefixes {
		var prefix struct {
			File  string `json:"file"`
			RunID string `json:"run_id"`
			Bytes int64  `json:"byte_length"`
			Seq   int64  `json:"last_seq"`
		}
		if err := json.Unmarshal([]byte(encoded), &prefix); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(prefix.File)
		if err != nil || prefix.Bytes < 1 || int64(len(data)) < prefix.Bytes || data[prefix.Bytes-1] != '\n' || prefix.RunID != p.facts.Manifest["run_id"] {
			t.Fatal("issued prefix lacks flushed complete records", prefix, err)
		}
		lines := strings.Split(strings.TrimSuffix(string(data[:prefix.Bytes]), "\n"), "\n")
		for index, line := range lines {
			var record struct {
				Seq int64 `json:"seq"`
			}
			if err := json.Unmarshal([]byte(line), &record); err != nil || record.Seq != int64(index+1) {
				t.Fatal("issued prefix contains an incomplete or noncontiguous record", err)
			}
		}
		if int64(len(lines)) != prefix.Seq {
			t.Fatal("issued prefix sequence does not close", prefix.Seq, len(lines))
		}
		count++
	}
	if count != len(p.facts.Graph.Units) {
		t.Fatal("not every launched unit received a committed prefix", count)
	}
	if _, err := os.Lstat(filepath.Join(root, "unit-events.ndjson")); !os.IsNotExist(err) {
		t.Fatal("incomplete stream exposed")
	}
	summary, err := publisher.Finish(result)
	if err != nil {
		t.Fatal(err)
	}
	if summary["status"] != "pass" {
		t.Fatal(summary)
	}
	bytes, err := os.ReadFile(filepath.Join(root, "unit-logs/unit-a/stdout.log"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(bytes), "private-sentinel") || strings.Contains(string(bytes), ":password@") {
		t.Fatal("unredacted diagnostic")
	}
	if err := publisher.Event(map[string]any{}); err == nil {
		t.Fatal("late event accepted")
	}
	if _, err := publisher.Finish(result); err == nil {
		t.Fatal("terminal publication repeated")
	}
	if _, err := os.Lstat(filepath.Join(root, "run-summary.json")); err != nil {
		t.Fatal(err)
	}
}
