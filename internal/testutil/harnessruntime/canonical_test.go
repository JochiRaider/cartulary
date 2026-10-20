//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"
)

func canonicalFixture(t *testing.T) (*Plan, []map[string]any) {
	t.Helper()
	a, b := engineUnit("a", 10), engineUnit("b", 10)
	a.Claims["volume"], b.Claims["volume"] = 1, 1
	b.Kind = "readiness"
	p := enginePlan(a, b)
	p.facts.Capacities["cpu"] = 2
	p.facts.Capacities["volume"] = 1
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
	if pressure["peak_use"].(map[string]int64)["volume"] != 1 || pressure["saturation_ms"].(map[string]int64)["volume"] != 30 {
		t.Fatal("writable-root capability did not span all owners", pressure)
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
	for _, stage := range []string{"admission", "acquisition", "cancelled"} {
		t.Run("prelaunch unit is readable by dependent finalizer "+stage, func(t *testing.T) {
			fixture := planFixture(t)
			graph := fixture["graph"].(map[string]any)
			a := graph["units"].([]any)[0].(map[string]any)
			b := maps.Clone(a)
			b["unit_id"], b["needs"] = "unit:b", []any{"unit:a"}
			b["current_run_evidence_outputs"] = []any{"unit-results/unit-b.json"}
			a["failure_policy"] = map[string]any{"block_descendants": false, "continue_independent": true, "aggregate_effect": "required"}
			if stage == "acquisition" {
				a["fixture_lease"], a["postgres_isolation"] = "postgres_dedicated", "exclusive"
				a["service_dependencies"] = []any{"postgres"}
				a["physical_allocations"] = map[string]any{"postgres": map[string]any{"memory_mb": 1, "process": 1, "postgres": 1, "port_lane": 1, "volume": 1}}
			}
			graph["units"] = []any{a, b}
			fixture["projections"].(map[string]any)["test-slice"].(map[string]any)["unit_ids"] = []any{"unit:a", "unit:b"}
			cache := maps.Clone(fixture["cache"].([]any)[0].(map[string]any))
			cache["unit_id"] = "unit:b"
			fixture["cache"] = append(fixture["cache"].([]any), cache)
			rehashPlan(t, fixture)
			p, err := decodePlan(planBytes(t, fixture))
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
			observed := false
			primary := failed("infra", "service_readiness_timeout")
			if stage == "cancelled" {
				primary = failed("interrupted", "cancelled_or_interrupted")
			}
			result, err := executeGraph(t.Context(), p, enginePorts{Event: publisher.Event, Terminal: publisher.Terminal,
				Admit: func(_ context.Context, u unit) (*hostLease, error) {
					if u.ID == "unit:a" && stage != "acquisition" {
						return nil, primary
					}
					return nil, nil
				}, Fixtures: &testFixtures{acquire: func(context.Context, fixtureRequest) (*fixtureLease, error) { return nil, primary }},
				Execute: func(_ context.Context, req executionRequest) unitResult {
					if req.Unit.ID != "unit:b" {
						t.Error("prelaunch failure executed runner")
					}
					data, err := os.ReadFile(filepath.Join(root, "unit-results/unit-a.json"))
					var unit map[string]any
					if err != nil || json.Unmarshal(data, &unit) != nil || unit["failure_reason"] != primary.Reason || unit["failure_class"] != primary.Class {
						t.Error("dependent finalizer lacks primary unit evidence", err, unit)
					}
					observed = true
					return unitResult{Status: "passed"}
				}})
			if err != nil || !observed || result.States["unit:b"] != "passed" {
				t.Fatal("dependent finalizer did not complete", err, result.States)
			}
			if _, err := publisher.Finish(result); err != nil {
				t.Fatal(err)
			}
		})
	}
	t.Run("cleanup failures retain separate safe causes after terminal publication", func(t *testing.T) {
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
		private := &os.PathError{Op: "open", Path: "/private/credential-sentinel", Err: syscall.EACCES}
		imported := &harnesssession.CleanupDiagnostic{Schema: "cartulary.harness_cleanup_diagnostic.v3", Stage: "browser_namespace", Cause: "deadline"}
		result, err := executeGraph(context.Background(), p, enginePorts{
			Event: publisher.Event, Terminal: publisher.Terminal,
			Execute: func(context.Context, executionRequest) unitResult {
				return unitResult{Status: "failed", ExitCode: 10, Failure: failed("product", "test_assertion_failure")}
			},
			Fixtures:        &testFixtures{close: func(context.Context) error { return nil }},
			CleanupServices: func(context.Context) error { return harnesssession.AtCleanupStage("resource_release", private) },
			Finalize: func(context.Context, bool, []cleanupRecord) error {
				return harnesssession.AtCleanupStage("runtime_close", errors.Join(errors.New("credential-sentinel"), imported))
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if len(result.Cleanup) != 3 || !errors.Is(result.Cleanup[1].Cause, syscall.EACCES) {
			t.Fatal("native cause discarded")
		}
		memoryBytes, err := json.Marshal(result.Cleanup)
		if err != nil || strings.Contains(string(memoryBytes), "credential-sentinel") {
			t.Fatal("raw cause serialized", err)
		}
		summary, err := publisher.Finish(result)
		if err != nil || summary["failure_class"] != "product" {
			t.Fatal("cleanup replaced primary failure", err)
		}
		data, err := os.ReadFile(filepath.Join(root, "cleanup-results.json"))
		if err != nil || strings.Contains(string(data), "credential-sentinel") {
			t.Fatal("unsafe cleanup evidence", err)
		}
		value, err := harnessschema.Decode(data)
		if err != nil || harnessschema.Validate("cartulary.harness_cleanup_results.v2", value) != nil {
			t.Fatal("invalid cleanup publication", err)
		}
		records := value.(map[string]any)["results"].([]any)
		if records[0].(map[string]any)["diagnostic"] != nil {
			t.Fatal("successful cleanup has diagnostic")
		}
		for i, stage := range []string{"resource_release", "browser_namespace"} {
			d := records[i+1].(map[string]any)["diagnostic"].(map[string]any)
			if d["stage"] != stage {
				t.Fatal("lost distinct boundary", d)
			}
		}
		first := records[1].(map[string]any)["diagnostic"].(map[string]any)
		if first["errno"] != float64(syscall.EACCES) || first["cause"] != "io" {
			t.Fatal("lost safe errno/cause", first)
		}
	})

	for _, runner := range []string{"vitest", "playwright"} {
		t.Run(runner+" cohort interruption preserves settled outcomes", func(t *testing.T) {
			u := engineUnit("unit:a", 1)
			completed := rowBinding{ID: "package.view_contracts.frontend_unit.references", Runner: runner}
			unfinished := rowBinding{ID: "package.view_contracts.frontend_unit.surfaces", Runner: runner}
			u.Rows = []rowBinding{completed, unfinished}
			p := enginePlan(u)
			root := t.TempDir()
			publisher := &graphPublication{plan: p, root: root, facts: newCanonicalFacts(p)}
			publisher.facts.starts[u.ID] = 0
			items, err := semanticEvidence(unit{ID: u.ID, Digest: "sha256:" + strings.Repeat("a", 64), Evidence: []string{"rows/" + completed.ID + ".json", "unit-results/a.json"}, Rows: []rowBinding{completed}}, semanticRows{Status: "passed", Rows: []cachedRow{{ID: completed.ID, Runner: runner}}}, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(root, "rows", completed.ID+".json")
			writeFixtureFile(t, file, items[0].Bytes, 0600)
			if err := publisher.publishRows(u, canonicalTerminal{status: "cancelled", failure: failed("interrupted", "cancelled_or_interrupted")}); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(file)
			if err != nil || string(got) != string(items[0].Bytes) {
				t.Fatal("settled file result was replaced", err)
			}
			data, err := os.ReadFile(filepath.Join(root, "rows", unfinished.ID+".json"))
			var result map[string]any
			if err != nil || json.Unmarshal(data, &result) != nil || result["terminal_state"] != "cancelled" || result["runner"] != runner {
				t.Fatal(result, err)
			}
		})
	}
	t.Run("dependency rows publish before consumers without replacing root failure", func(t *testing.T) {
		fixture := planFixture(t)
		graph := fixture["graph"].(map[string]any)
		a := graph["units"].([]any)[0].(map[string]any)
		b := maps.Clone(a)
		rowID := "harness.command_surface.unit.scheduler_smoke"
		b["unit_id"], b["needs"] = "unit:b", []any{"unit:a"}
		b["row_bindings"] = []any{map[string]any{"row_id": rowID, "runner": "go"}}
		b["current_run_evidence_outputs"] = []any{"rows/" + rowID + ".json", "unit-results/unit-b.json"}
		graph["units"] = []any{a, b}
		fixture["projections"].(map[string]any)["test-slice"].(map[string]any)["unit_ids"] = []any{"unit:a", "unit:b"}
		cache := maps.Clone(fixture["cache"].([]any)[0].(map[string]any))
		cache["unit_id"] = "unit:b"
		fixture["cache"] = append(fixture["cache"].([]any), cache)
		rehashPlan(t, fixture)
		p, err := decodePlan(planBytes(t, fixture))
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
		observed := false
		result, err := executeGraph(context.Background(), p, enginePorts{Event: func(e map[string]any) error {
			if err := publisher.Event(e); err != nil {
				return err
			}
			if e["event"] == "skipped" {
				data, err := os.ReadFile(filepath.Join(root, "unit-results/unit-b.json"))
				var unit map[string]any
				if err != nil || json.Unmarshal(data, &unit) != nil || unit["status"] != "skipped" || unit["exit_code"] != nil {
					t.Fatal("dependency unit was not published before consumers", err, unit)
				}
				var row map[string]any
				data, err = os.ReadFile(filepath.Join(root, "rows", rowID+".json"))
				if err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(data, &row); err != nil {
					t.Fatal(err)
				}
				if row["terminal_state"] != "skipped_dependency" || row["runner"] != "go" || row["duration_ms"] != float64(0) || row["started_at"] != row["finished_at"] {
					t.Fatal(row)
				}
				observed = true
			}
			return nil
		}, Terminal: publisher.Terminal, Execute: func(_ context.Context, req executionRequest) unitResult {
			if req.Unit.ID != "unit:a" {
				t.Error("blocked row executed")
			}
			return unitResult{Status: "failed", ExitCode: 10, Failure: failed("product", "test_assertion_failure")}
		}})
		if err != nil {
			t.Fatal(err)
		}
		summary, err := publisher.Finish(result)
		if err != nil || !observed || summary["failure_class"] != "product" || summary["failure_reason"] != "test_assertion_failure" {
			t.Fatal(summary, err)
		}
	})
	t.Run("unavailable row publication preserves evidence and fails closed", func(t *testing.T) {
		for _, mode := range []string{"setup", "timeout", "passed_existing", "malformed", "foreign_runner", "symlink", "missing_pass"} {
			t.Run(mode, func(t *testing.T) {
				u := engineUnit("unit:a", 1)
				row := rowBinding{ID: "harness.command_surface.unit.scheduler_smoke", Runner: "go"}
				u.Rows = []rowBinding{row}
				p := enginePlan(u)
				root := t.TempDir()
				publisher := &graphPublication{plan: p, root: root, facts: newCanonicalFacts(p)}
				publisher.facts.starts[u.ID] = 0
				terminal := canonicalTerminal{status: "failed", failure: failed("infra", "service_start_error")}
				if mode == "timeout" {
					terminal.failure = failed("timing", "timeout_failure")
				}
				if mode == "missing_pass" {
					terminal.status, terminal.failure = "passed", nil
				}
				file := filepath.Join(root, "rows", row.ID+".json")
				var original []byte
				if mode == "passed_existing" || mode == "foreign_runner" {
					items, err := semanticEvidence(unit{ID: "unit:a", Digest: "sha256:" + strings.Repeat("a", 64), Evidence: []string{"rows/" + row.ID + ".json", "unit-results/a.json"}, Rows: u.Rows}, semanticRows{Status: "passed", Rows: []cachedRow{{ID: row.ID, Runner: "go"}}}, time.Now())
					if err != nil {
						t.Fatal(err)
					}
					original = items[0].Bytes
					if mode == "foreign_runner" {
						original = []byte(strings.Replace(string(original), `"runner":"go"`, `"runner":"vitest"`, 1))
					}
					writeFixtureFile(t, file, original, 0600)
				} else if mode == "malformed" {
					original = []byte("{")
					writeFixtureFile(t, file, original, 0600)
				} else if mode == "symlink" {
					if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink("/dev/null", file); err != nil {
						t.Fatal(err)
					}
				}
				err := publisher.publishRows(u, terminal)
				unsafe := mode == "malformed" || mode == "foreign_runner" || mode == "symlink" || mode == "missing_pass"
				if (err != nil) != unsafe {
					t.Fatalf("unexpected publication result: %v", err)
				}
				if original != nil {
					got, e := os.ReadFile(file)
					if e != nil || string(got) != string(original) {
						t.Fatal("existing row was rewritten", e)
					}
				}
				if mode == "setup" || mode == "timeout" {
					data, e := os.ReadFile(file)
					if e != nil {
						t.Fatal(e)
					}
					var v map[string]any
					if json.Unmarshal(data, &v) != nil || v["terminal_state"] != "infrastructure_failed" || v["failure_reason"] != terminal.failure.Reason {
						t.Fatal(v)
					}
				}
			})
		}
	})
	t.Run("cache_cancellation_publishes_terminal_without_execution", func(t *testing.T) {
		fixture := planFixture(t)
		rowID := "harness.command_surface.unit.scheduler_smoke"
		fixture["graph"].(map[string]any)["units"].([]any)[0].(map[string]any)["row_bindings"] = []any{map[string]any{"row_id": rowID, "runner": "go"}}
		fixture["graph"].(map[string]any)["units"].([]any)[0].(map[string]any)["current_run_evidence_outputs"] = []any{"rows/" + rowID + ".json", "unit-results/unit-a.json"}
		rehashPlan(t, fixture)
		p, err := decodePlan(planBytes(t, fixture))
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
		for _, name := range []string{"run-summary.json", "unit-events.ndjson", "cleanup-results.json", "rows/" + rowID + ".json"} {
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
	if strings.Contains(string(contextBytes), "/private/node") || retained["bindings"] != nil || retained["cache_inputs"] != nil || retained["input_environment"] != nil {
		t.Fatal("private launch state was retained")
	}
	if err := writeEvidence(root, "execution-context.json", "cartulary.harness_execution_context.v3", retained, false); err == nil {
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
