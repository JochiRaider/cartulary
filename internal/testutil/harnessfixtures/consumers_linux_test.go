//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"path/filepath"
	"testing"
	"time"
)

func testConsumerAccounting(t *testing.T) {
	t.Run("cancelled_before_service_admission", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{})
		admissionRoot := filepath.Join(root, "admission")
		capacities := map[string]int64{"cpu": 2, "io": 2, "memory_mb": 2048, "process": 32, "object_store": 4, "port_lane": 8}
		p.ConfigureAdmission(admissionRoot, nil, capacities)
		blocker, err := harnessadmission.Acquire(t.Context(), admissionRoot, quantities(map[string]int64{"cpu": 2}), quantities(capacities), nil, false)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = blocker.Release(context.Background()) })
		work := WorkRequirement{Request: Request{UnitID: "cancelled", Capability: "object_store_namespace", PostgresIsolation: "none", Services: []string{"object_store"}}, Renderer: "none", Physical: map[string]map[string]int64{"object_store": {"memory_mb": 128, "process": 2, "object_store": 1, "port_lane": 1}}, Active: map[string]int64{"cpu": 1, "memory_mb": 64, "process": 1}}
		if err := p.RegisterConsumers([]WorkRequirement{work}); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = p.Close(context.Background()) })
		starts := 0
		p.Suite.Acquire = func(context.Context) (*Suite, error) {
			starts++
			return nil, errors.New("unadmitted service started")
		}
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		ctx = harnessadmission.WithObservation(ctx, work.Request.UnitID, func() float64 { return 0 }, func(o harnessadmission.Observation) {
			if o.Decision != nil && !o.Decision.Granted {
				cancel()
			}
		})
		host, finish, err := p.Admit(ctx, work)
		if !errors.Is(err, context.Canceled) || host != nil || finish != nil {
			t.Fatal("pending admission did not cancel", err)
		}
		if err := p.ConsumerTerminal(work.Request.UnitID, "cancelled"); err != nil {
			t.Fatal(err)
		}
		// The unrelated active claim remains held. No physical service exists,
		// so retiring this pending proof must not acquire drain entitlement.
		if err := p.retireUnused(t.Context()); err != nil {
			t.Fatal("unstarted service retirement failed", err)
		}
		if err := p.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
		resources, err := runtime.Resources(32)
		if err != nil || len(resources) != 0 || starts != 0 {
			t.Fatal("unstarted service retained resources or started", len(resources), starts, err)
		}
	})
	t.Run("exact_consumer_keys_and_terminal_dispositions", func(t *testing.T) {
		runtime, root, run := runtimeFixture(t)
		p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{})
		p.ConfigureAdmission(filepath.Join(root, "admission"), nil, map[string]int64{"cpu": 2, "io": 2, "memory_mb": 2048, "process": 32, "port_lane": 8})
		base := WorkRequirement{Request: Request{UnitID: "passed", Capability: "browser_stack", PostgresIsolation: "exclusive", Services: []string{"postgres", "object_store"}, RuntimeProfile: "default", BrowserStage: "functional", Affinity: "lane"}, Renderer: "pooled"}
		works := []WorkRequirement{}
		for _, id := range []string{"passed", "failed", "skipped", "cached", "cancelled"} {
			w := base
			w.Request.UnitID = id
			works = append(works, w)
		}
		if err := p.RegisterConsumers(works); err != nil {
			t.Fatal(err)
		}
		if err := p.RegisterConsumers(works); err == nil {
			t.Fatal("duplicate roster accepted")
		}
		// Roster input does not survive as a mutable alias.
		works[0].Request.Services[0] = "mutated"
		if p.stackConsumers(sharedKey(works[1].Request), "") == 5 {
			t.Fatal("roster keys followed mutated input")
		}
		base.Request.Services = []string{"postgres", "object_store"}
		key := sharedKey(base.Request)
		if p.stackConsumers(key, "") != 5 || p.stackConsumers(key, "passed") != 4 {
			t.Fatal("selected consumer count")
		}
		permuted := base
		permuted.Request.Services = []string{"object_store", "postgres"}
		if sharedKey(base.Request) != sharedKey(permuted.Request) {
			t.Fatal("service ordering changed compatibility")
		}
		for _, mutate := range []func(*WorkRequirement){
			func(w *WorkRequirement) { w.Request.RuntimeProfile = "other" },
			func(w *WorkRequirement) { w.Request.BrowserStage = "visual" },
			func(w *WorkRequirement) { w.Request.Affinity = "other" },
			func(w *WorkRequirement) { w.Request.PostgresIsolation = "shared" },
			func(w *WorkRequirement) { w.Request.FixtureProfile = "snapshot" },
			func(w *WorkRequirement) { w.Request.SnapshotKey = "snapshot" },
			func(w *WorkRequirement) { w.Request.BuilderUnit = "builder" },
			func(w *WorkRequirement) { w.Request.Services = []string{"postgres"} },
		} {
			w := base
			mutate(&w)
			if p.stackConsumers(sharedKey(w.Request), "") != 0 {
				t.Fatal("incompatible stack consumer", w)
			}
		}
		for _, key := range []rendererConsumerKey{{"other", "pooled", false}, {"default", "dedicated", false}, {"default", "pooled", true}} {
			if p.rendererConsumers(key, "") != 0 {
				t.Fatal("incompatible renderer consumer")
			}
		}
		if err := p.ConsumerTerminal("unknown", "passed"); err == nil {
			t.Fatal("unknown accepted")
		}
		if err := p.ConsumerTerminal("passed", "invalid"); err == nil {
			t.Fatal("invalid accepted")
		}
		for _, id := range []string{"passed", "failed", "skipped", "cached", "cancelled"} {
			if err := p.ConsumerTerminal(id, id); err != nil {
				t.Fatal(err)
			}
			if err := p.ConsumerTerminal(id, id); err != nil {
				t.Fatal("idempotent terminal", err)
			}
			opposite := "passed"
			if id == opposite {
				opposite = "failed"
			}
			if err := p.ConsumerTerminal(id, opposite); err == nil {
				t.Fatal("contradictory terminal accepted")
			}
		}
		if p.stackConsumers(key, "") != 0 {
			t.Fatal("terminal consumers retained")
		}
		if err := p.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	for _, scenario := range []string{"settled", "cleanup_failure", "pressure_race"} {
		cleanupFails := scenario == "cleanup_failure"
		t.Run("skipped_last_consumer_"+scenario, func(t *testing.T) {
			runtime, root, run := runtimeFixture(t)
			var lifecycle *AllocationLifecycle
			p := NewProduction(ProductionOptions{Runtime: runtime, Root: root, RunRoot: run}, Options{Observe: func(o Observation) {
				if o.Lifecycle != nil {
					copy := *o.Lifecycle
					lifecycle = &copy
				}
			}})
			p.ConfigureAdmission(filepath.Join(root, "admission"), nil, map[string]int64{"cpu": 2, "io": 2, "memory_mb": 2048, "process": 32, "port_lane": 8})
			work := WorkRequirement{Request: Request{UnitID: "first", Capability: "browser_stack", PostgresIsolation: "exclusive", Services: []string{"postgres"}, RuntimeProfile: "default", Affinity: "lane"}, Renderer: "none", Physical: map[string]map[string]int64{"browser_stack": {"memory_mb": 128, "port_lane": 2}}, Active: map[string]int64{"cpu": 1, "memory_mb": 64, "process": 1}}
			next := work
			next.Request.UnitID = "never-started"
			if err := p.RegisterConsumers([]WorkRequirement{work, next}); err != nil {
				t.Fatal(err)
			}
			entered, finishCleanup := make(chan struct{}), make(chan struct{})
			p.Broker.options.Providers["browser_stack"] = &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
				return ownedAllocation(func(ctx context.Context, _ bool) error {
					close(entered)
					select {
					case <-finishCleanup:
					case <-ctx.Done():
						return ctx.Err()
					}
					if cleanupFails {
						return errors.New("injected physical settlement failure")
					}
					return p.Options.residency.close(ctx, stackKey(work.Request))
				}), nil
			}}
			host, finish, err := p.Admit(t.Context(), work)
			if err != nil {
				t.Fatal(err)
			}
			if err := p.ConsumerTerminal("first", "passed"); err == nil {
				t.Fatal("active consumer terminal accepted")
			}
			lease, err := p.Acquire(t.Context(), work.Request)
			if err != nil {
				t.Fatal(err)
			}
			if retained, err := lease.Release(t.Context(), true); err != nil || !retained {
				t.Fatal(retained, err)
			}
			if err := errors.Join(finish(t.Context()), host.Release(t.Context())); err != nil {
				t.Fatal(err)
			}
			if err := p.ConsumerTerminal("first", "passed"); err != nil {
				t.Fatal(err)
			}
			pressureDone := make(chan error, 1)
			if scenario == "pressure_race" {
				go func() {
					p.retireMu.Lock()
					defer p.retireMu.Unlock()
					_, err := p.retireIdleStacks(t.Context(), Retirement{"capacity_pressure", "pending"}, 1, func(Request) (bool, error) { return true, p.ConsumerTerminal("never-started", "skipped") })
					pressureDone <- err
				}()
			} else if err := p.ConsumerTerminal("never-started", "skipped"); err != nil {
				t.Fatal(err)
			}
			select {
			case <-entered:
			case <-time.After(2 * time.Second):
				t.Fatal("unused stack did not promptly enter cleanup")
			}
			// Duplicate terminal remains nonblocking even while physical cleanup waits.
			if err := p.ConsumerTerminal("never-started", "skipped"); err != nil {
				t.Fatal(err)
			}
			close(finishCleanup)
			if scenario == "pressure_race" {
				if err := <-pressureDone; err != nil {
					t.Fatal(err)
				}
			}
			joined := p.joinConsumerCleanup()
			if lifecycle == nil || lifecycle.Reason != "terminal_borrower_release" || lifecycle.Trigger == nil || *lifecycle.Trigger != "never-started" {
				t.Fatal("retirement trigger lost", lifecycle)
			}
			if cleanupFails {
				if joined == nil || !p.Broker.Unresolved() {
					t.Fatal("cleanup failure hidden", joined)
				}
				record := p.Options.residency.allocations[stackKey(work.Request)]
				if record == nil {
					t.Fatal("uncertain claims released")
				}
				granted, err := harnessadmission.Transact(t.Context(), harnessadmission.Request{Root: record.Root, Operation: "resident_granted", Token: record.Allocation.Token, Reference: record.Allocation.Reference})
				if err != nil || !granted {
					t.Fatal("uncertain ownership not charged", granted, err)
				}
				if _, err := p.Broker.Acquire(t.Context(), work.Request); err == nil {
					t.Fatal("quarantined allocation reused")
				}
				_ = p.Broker.Close(t.Context())
				// This fixture owns no physical service. Settle only its synthetic
				// resident after asserting that the failed owner retained authority.
				if err := p.Options.residency.close(t.Context(), stackKey(work.Request)); err != nil {
					t.Fatal(err)
				}
				return
			}
			if joined != nil {
				t.Fatal(joined)
			}
			if err := p.Close(t.Context()); err != nil {
				t.Fatal(err)
			}
			resources, err := runtime.Resources(4096)
			if err != nil || len(resources) != 0 {
				t.Fatal("unsettled ownership", resources, err)
			}
		})
	}
	t.Run("stack_release_set_excludes_shared_and_measurement_generations", func(t *testing.T) {
		for _, kind := range []string{"ordinary", "measurement", "shared"} {
			p := NewProduction(ProductionOptions{}, Options{})
			request := Request{UnitID: "unit", Capability: "browser_stack", PostgresIsolation: "exclusive", Affinity: "lane"}
			if kind == "measurement" {
				request.FixtureProfile = "profile"
				request.BuilderUnit = "builder"
				request.SnapshotKey = "snapshot"
			}
			if kind == "shared" {
				request.PostgresIsolation = "shared"
			}
			stack, cluster, unrelated := &residentRecord{}, &residentRecord{}, &residentRecord{}
			p.Options.residency.allocations[stackKey(request)] = stack
			p.Options.residency.allocations[clusterKey(request)] = cluster
			p.Options.residency.allocations["object_store:shared"] = unrelated
			records := p.stackRetirementRecords(request)
			expected := 1
			if kind == "ordinary" {
				expected = 2
			}
			if len(records) != expected || records[0] != stack || expected == 2 && records[1] != cluster {
				t.Fatal("incorrect physical release set", kind, records)
			}
		}
	})
	t.Run("empty_and_invalid_rosters", func(t *testing.T) {
		for _, works := range [][]WorkRequirement{nil, {{Request: Request{UnitID: "", Capability: "none", PostgresIsolation: "none"}, Renderer: "none"}}} {
			p := NewProduction(ProductionOptions{}, Options{})
			err := p.RegisterConsumers(works)
			if len(works) == 0 {
				if err != nil {
					t.Fatal(err)
				}
				if err := p.joinConsumerCleanup(); err != nil {
					t.Fatal(err)
				}
			} else if err == nil {
				t.Fatal("invalid roster accepted")
			}
		}
	})
	t.Run("missing_terminal_is_required_failure", func(t *testing.T) {
		p := NewProduction(ProductionOptions{}, Options{})
		w := WorkRequirement{Request: Request{UnitID: "pending", Capability: "none", PostgresIsolation: "none"}, Renderer: "none"}
		if err := p.RegisterConsumers([]WorkRequirement{w}); err != nil {
			t.Fatal(err)
		}
		if err := p.joinConsumerCleanup(); err == nil {
			t.Fatal("missing terminal hidden")
		}
	})
}
