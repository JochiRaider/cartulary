//go:build linux

package harnessruntime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func observationRoot() string {
	_, f, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(f), "../../.."))
}
func observationFixture(t *testing.T) (*Plan, *harnesssession.Runtime, string, map[string]any) {
	t.Helper()
	root := observationRoot()
	policy, err := readInstrumentationPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	source := fixturePlan(t)
	digest, _ := semanticDigest(policy)
	source.Manifest.Instrumentation = harnessplan.Instrumentation{Mode: "basic", PolicyDigest: digest}
	plan := adoptFixture(t, source)
	run := t.TempDir()
	if err := os.Chmod(run, 0700); err != nil {
		t.Fatal(err)
	}
	r, err := harnesssession.NewRuntime(root, run, "fixture", t.TempDir())
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
	t.Run("admission_failure_is_optional_and_discloses_lost_coverage", func(t *testing.T) {
		s := &executionSession{enabled: true, policy: map[string]any{}}
		phase := harnessadmission.Observation{UnitID: "u", Clock: "graph-monotonic", Kind: "phase", Phase: "admission_wait", StartMS: 1, EndMS: 2, DurationMS: 1, Count: 1}
		// A malformed diagnostic policy must not panic through the product owner.
		s.admission(phase)
		if s.hostAdmission.document()["status"] != "partial" || s.hostAdmission.omissions != 1 {
			t.Fatal(s.hostAdmission.document())
		}
	})
	t.Run("admission_aggregates_durations_without_claiming_continuous_observation", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		s := newExecutionSession(r, run, p.facts.Manifest, policy, nil)
		phase := harnessadmission.Observation{UnitID: "u", Clock: "graph-monotonic", Kind: "phase", Phase: "admission_wait", StartMS: 1, EndMS: 3, DurationMS: 2, Count: 1}
		s.admission(phase)
		phase.StartMS = 10
		phase.EndMS = 12
		s.admission(phase)
		doc := s.hostAdmission.document()
		if len(s.hostAdmission.records) != 1 || doc["status"] != "complete" {
			t.Fatal(doc)
		}
		got := s.hostAdmission.records[0]
		if got.StartMS != 1 || got.EndMS != 12 || got.DurationMS != 4 || got.Count != 2 {
			t.Fatal(got)
		}
		policy["maximum_host_admission_records"] = float64(1)
		phase.Phase = "logical_slot"
		s.admission(phase)
		if s.hostAdmission.document()["status"] != "truncated" || s.hostAdmission.omissions != 1 {
			t.Fatal(s.hostAdmission.document())
		}
		// Independent verbose limits cannot remove admission observations.
		policy["maximum_execution_records"] = float64(0)
		if len(s.hostAdmission.records) != 1 {
			t.Fatal("admission section coupled to verbose records")
		}
		s.enabled = false
		phase.UnitID = "disabled"
		s.admission(phase)
		if s.hostAdmission.omissions != 1 {
			t.Fatal("off diagnostics recorded a unit")
		}
	})
	t.Run("admission_decisions_coalesce_by_identity_not_retry_frequency", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		s := newExecutionSession(r, run, p.facts.Manifest, policy, nil)
		request := harnessadmission.ObservationReference("request", "waiting")
		holder := harnessadmission.ObservationReference("request", "holder")
		decision := func(at float64, position int, processDeficit float64, holderProcess float64, granted bool) harnessadmission.Observation {
			deficits := harnessadmission.Quantities{}
			if processDeficit > 0 {
				deficits["process"] = processDeficit
			}
			return harnessadmission.Observation{UnitID: "u", RequestRef: &request, Clock: "graph-monotonic", Kind: "decision", StartMS: at, EndMS: at, Count: 1, Outcome: "committed", Decision: &harnessadmission.Decision{
				RequestRef: request, Sequence: 7, QueuePosition: position, Mode: "shared", Granted: granted,
				RequestedClaims:    harnessadmission.Quantities{"cpu": 1, "process": 1, "postgres": 0},
				Capacities:         harnessadmission.Quantities{"cpu": 19, "process": 76, "volume": 0},
				Deficits:           deficits,
				PersistentDeficits: harnessadmission.Quantities{"postgres": 0},
				Ancestors:          []harnessadmission.ObservedHolder{},
				Holders:            []harnessadmission.ObservedHolder{{Reference: holder, Claims: harnessadmission.Quantities{"cpu": 1, "process": holderProcess, "postgres": 0}}},
				Residents:          []harnessadmission.ObservedHolder{},
				Prospective:        []harnessadmission.ObservedHolder{},
				PriorityBlockers:   []string{},
				ExclusiveBlockers:  []string{},
			}}
		}
		// Retries that change only queue position, deficit amount and holder
		// claims are one decision; the record carries the last observation.
		s.admission(decision(10, 3, 9, 8, false))
		s.admission(decision(20, 2, 4, 3, false))
		s.admission(decision(30, 1, 1, 1, false))
		if len(s.hostAdmission.records) != 1 {
			t.Fatalf("identity-preserving retries produced %d records", len(s.hostAdmission.records))
		}
		got := s.hostAdmission.records[0]
		if got.StartMS != 10 || got.EndMS != 30 || got.Count != 3 || got.Decision.QueuePosition != 1 || got.Decision.Deficits["process"] != 1 {
			t.Fatal(got)
		}
		// Zero quantities are omitted; zero capacities are facts and remain.
		if _, ok := got.Decision.RequestedClaims["postgres"]; ok {
			t.Fatal("zero requested claim retained", got.Decision.RequestedClaims)
		}
		if _, ok := got.Decision.Holders[0].Claims["postgres"]; ok {
			t.Fatal("zero holder claim retained", got.Decision.Holders[0].Claims)
		}
		if len(got.Decision.PersistentDeficits) != 0 {
			t.Fatal("zero persistent deficit retained", got.Decision.PersistentDeficits)
		}
		if v, ok := got.Decision.Capacities["volume"]; !ok || v != 0 {
			t.Fatal("zero capacity dropped", got.Decision.Capacities)
		}
		// A grant changes identity and starts the successor record.
		s.admission(decision(40, 1, 0, 1, true))
		if len(s.hostAdmission.records) != 2 || s.hostAdmission.records[1].StartMS != 40 || !s.hostAdmission.records[1].Decision.Granted {
			t.Fatal(s.hostAdmission.records)
		}
		// A different deficit resource set is a new decision, even mid-wait.
		other := harnessadmission.ObservationReference("request", "other")
		blocked := decision(50, 2, 9, 8, false)
		blocked.RequestRef, blocked.Decision.RequestRef = &other, other
		s.admission(blocked)
		blocked = decision(60, 2, 0, 8, false)
		blocked.RequestRef, blocked.Decision.RequestRef = &other, other
		blocked.Decision.Deficits = harnessadmission.Quantities{"postgres": 1}
		s.admission(blocked)
		if len(s.hostAdmission.records) != 4 {
			t.Fatalf("deficit set change coalesced: %d records", len(s.hostAdmission.records))
		}
		if doc := s.hostAdmission.document(); doc["status"] != "complete" || doc["omitted_records"] != 0 {
			t.Fatal(doc)
		}
	})
	testAllocationObservationBounds(t)
	t.Run("admission_preserves_typed_owner_failure", func(t *testing.T) {
		for _, tc := range []struct {
			err           error
			class, reason string
		}{
			{&harnessfixtures.Failure{Class: "config", Reason: "configuration_error", Cause: errors.New("private credentials")}, "config", "configuration_error"},
			{&harnessfixtures.Failure{Class: "infra", Reason: "service_start_error"}, "infra", "service_start_error"},
			{context.DeadlineExceeded, "infra", "resource_conflict"},
			{harnessadmission.ErrConflict, "infra", "resource_conflict"},
			{harnessadmission.ErrRelease, "harness", "cleanup_error"},
		} {
			got := failureOf(classifyAdmissionFailure(t.Context(), tc.err))
			if got.Class != tc.class || got.Reason != tc.reason || strings.Contains(got.Error(), "private") {
				t.Fatal("admission classification lost", got)
			}
		}
	})

	t.Run("container_roster_retirement_preserves_other_services", func(t *testing.T) {
		c := &collectorClient{proofPath: filepath.Join(t.TempDir(), "proof.json"), manifest: map[string]any{"run_id": "run"}, lease: "runtime", queue: make(chan []byte, 1), acks: map[string]chan struct{}{}, pending: map[string]int{}, policy: map[string]any{"stop_ms": float64(1000)}}
		if err := os.Chmod(filepath.Dir(c.proofPath), 0700); err != nil {
			t.Fatal(err)
		}
		proof := func(id string) map[string]any {
			return map[string]any{"lease": map[string]any{"lease_id": id}, "ownership": "owned"}
		}
		read := func() map[string]any {
			t.Helper()
			b, err := os.ReadFile(c.proofPath)
			if err != nil {
				t.Fatal(err)
			}
			var v map[string]any
			if json.Unmarshal(b, &v) != nil {
				t.Fatal("bad roster")
			}
			return v
		}
		c.containers(t.Context(), proof("one"), false)
		c.containers(t.Context(), proof("two"), false)
		if len(read()["services"].([]any)) != 2 {
			t.Fatal("proof overwritten")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		observed := make(chan map[string]any, 1)
		settled := make(chan struct{})
		t.Cleanup(func() {
			cancel()
			select {
			case <-settled:
			case <-time.After(time.Second):
				t.Error("collector roster fixture worker did not settle")
			}
		})
		go func() {
			defer close(settled)
			var b []byte
			select {
			case b = <-c.queue:
			case <-ctx.Done():
				return
			}
			var request map[string]any
			_ = json.Unmarshal(b, &request)
			raw, _ := os.ReadFile(c.proofPath)
			var roster map[string]any
			_ = json.Unmarshal(raw, &roster)
			observed <- roster
			c.mu.Lock()
			if ack := c.acks[controlID(request["request"])]; ack != nil {
				close(ack)
				delete(c.acks, controlID(request["request"]))
			} else {
				t.Error("collector roster fixture lost acknowledgement")
			}
			c.mu.Unlock()
		}()
		c.containers(ctx, proof("one"), true)
		var before map[string]any
		select {
		case before = <-observed:
		case <-ctx.Done():
			t.Fatal("collector roster fixture did not receive retirement", ctx.Err())
		}
		if before["retiring_service"] != "one" || len(before["services"].([]any)) != 2 {
			t.Fatal("retirement lacks exact live proof")
		}
		after := read()
		services := after["services"].([]any)
		if after["retiring_service"] != "" || len(services) != 1 || services[0].(map[string]any)["lease"].(map[string]any)["lease_id"] != "two" {
			t.Fatal("sibling proof revoked")
		}
	})

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
		if err != nil || harnessschema.Validate("cartulary.harness_live_snapshot.v3", raw) != nil {
			t.Fatal("invalid live snapshot", err)
		}
		if live.publicationErr != nil {
			t.Fatal(live.publicationErr)
		}
	})
	t.Run("collector_closed_pipes_leave_no_live_descendant", func(t *testing.T) {
		for _, mode := range []string{"leader_exits", "stop_live_leader"} {
			t.Run(mode, func(t *testing.T) {
				p, r, run, policy := observationFixture(t)
				directory := t.TempDir()
				binary, pidFile := filepath.Join(directory, "collector"), filepath.Join(directory, "child")
				// Read the initial frame so the real client has completed launch.
				// The descendant closes inherited pipes; the live leader ignores
				// protocol requests and can only be stopped by its process owner.
				ending := "exit 0"
				if mode == "stop_live_leader" {
					ending = "wait"
				}
				script := "#!/bin/sh\nread -r frame\nsleep 60 </dev/null >/dev/null 2>&1 &\necho $! > " + strconv.Quote(pidFile) + "\n" + ending + "\n"
				if err := os.WriteFile(binary, []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
				c := newCollector(binary, run, filepath.Join(directory, "admission"), r, p, policy, monotonicNS(), nil)
				t.Cleanup(func() {
					if err := c.stop(); err != nil {
						t.Errorf("settle collector fixture: %v", err)
					}
				})
				deadline := time.NewTimer(5 * time.Second)
				defer deadline.Stop()
				tick := time.NewTicker(10 * time.Millisecond)
				defer tick.Stop()
				var data []byte
				for {
					var err error
					data, err = os.ReadFile(pidFile)
					if err == nil && len(strings.TrimSpace(string(data))) > 0 {
						break
					}
					if err != nil && !os.IsNotExist(err) {
						t.Fatal(err)
					}
					select {
					case <-deadline.C:
						t.Fatal("collector descendant did not start")
					case <-tick.C:
					}
				}
				pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
				if err != nil || pid < 2 {
					t.Fatal("invalid collector descendant", err)
				}
				if mode == "leader_exits" {
					select {
					case <-c.done:
					case <-deadline.C:
						t.Fatal("collector failed to settle")
					}
				} else {
					select {
					case <-c.done:
						t.Fatal("live collector exited before stop")
					default:
					}
				}
				if err := c.stop(); err != nil {
					t.Fatal(err)
				}
				select {
				case <-c.done:
				default:
					t.Fatal("collector stop returned before settlement")
				}
				stat, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
				if err == nil {
					state := strings.Fields(string(stat)[strings.LastIndex(string(stat), ") ")+2:])[0]
					if state != "Z" && state != "X" {
						t.Fatal("collector reported cleanup while descendant was live")
					}
				} else if !os.IsNotExist(err) {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("collector_forced_death_drains", func(t *testing.T) {
		p, r, run, policy := observationFixture(t)
		c := newCollector(filepath.Join(observationRoot(), ".cache/cartulary/toolbin/cartulary-harness"), run, filepath.Join(t.TempDir(), "admission"), r, p, policy, monotonicNS(), nil)
		t.Cleanup(func() {
			if err := c.stop(); err != nil {
				t.Errorf("settle collector fixture: %v", err)
			}
		})
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
