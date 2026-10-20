package harnessruntime

import (
	"context"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func engineUnit(id string, estimate int64, needs ...string) unit {
	if needs == nil {
		needs = []string{}
	}
	return unit{PostgresIsolation: "none", Renderer: "none", Physical: map[string]map[string]int64{}, ID: id, Owner: "harness.command_surface", Kind: "runner", Needs: needs, Claims: map[string]int64{"cpu": 1}, SharedLocks: []string{"host_activity"}, ExclusiveLocks: []string{}, Fixture: "none", Services: []string{}, CachePolicy: "none", Evidence: []string{}, Artifacts: []artifact{}, EstimateMS: estimate, TimeoutMS: 1000, Command: command{Executable: "true", Args: []string{}, Environment: map[string]string{}}, FailurePolicy: failurePolicy{BlockDescendants: true, ContinueIndependent: true, AggregateEffect: "required"}}
}
func enginePlan(units ...unit) *Plan {
	slices.SortFunc(units, func(a, b unit) int { return strings.Compare(a.ID, b.ID) })
	return &Plan{facts: planWire{Graph: workGraph{Units: units}, Capacities: map[string]int64{"cpu": 1}, AgingMS: 1000}}
}
func boundedGraph(t *testing.T, p *Plan, ports enginePorts) graphResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	r, err := executeGraph(ctx, p, ports)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Err() != nil {
		t.Fatal("engine did not drain")
	}
	return r
}
func TestRuntimeScheduling(t *testing.T) {
	t.Run("writable_root_is_shared_while_cpu_and_settlement_remain_accounted", func(t *testing.T) {
		a, b, c := engineUnit("a", 3), engineUnit("b", 2), engineUnit("c", 1, "a", "b")
		for _, u := range []unit{a, b, c} {
			u.Claims["volume"] = 1
		}
		p := enginePlan(a, b, c)
		p.facts.Capacities["cpu"], p.facts.Capacities["volume"] = 2, 1
		var arrived atomic.Int64
		both := make(chan struct{})
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Execute: func(ctx context.Context, req executionRequest) unitResult {
			if req.Unit.ID != "c" {
				if arrived.Add(1) == 2 {
					close(both)
				}
				select {
				case <-both:
				case <-ctx.Done():
					return unitResult{Status: "cancelled", Failure: failureOf(ctx.Err())}
				}
			}
			return unitResult{Status: "passed"}
		}})
		if r.Status != "passed" || !slices.Equal(r.Admissions, []string{"a", "b", "c"}) {
			t.Fatal("volume serialized or stranded admitted work", r.Admissions, r.Status)
		}
		facts := newCanonicalFacts(p)
		for _, event := range r.Events {
			if err := facts.consume(event); err != nil {
				t.Fatal(err)
			}
		}
		if err := facts.validate(); err != nil {
			t.Fatal(err)
		}
		peaks := facts.pressure()["peak_use"].(map[string]int64)
		if peaks["cpu"] != 2 || peaks["volume"] != 1 {
			t.Fatal("capability and additive reservations diverged", peaks)
		}
	})
	t.Run("wait_segments_follow_reservations_and_barriers", func(t *testing.T) {
		p := enginePlan(engineUnit("holder", 10), engineUnit("waiter", 1))
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		var reasons []string
		for _, e := range r.Events {
			if e["unit_id"] == "waiter" && e["event"] == "wait_started" {
				reasons = append(reasons, e["wait_reason"].(string))
				if e["wait_reason"] == "resources" && !slices.Equal(stringList(e["blocking_unit_ids"]), []string{"holder"}) {
					t.Fatal(e)
				}
			}
		}
		if !slices.Equal(reasons, []string{"rank", "resources", "rank"}) {
			t.Fatal(reasons)
		}
		facts := newCanonicalFacts(p)
		for _, e := range r.Events {
			if err := facts.consume(e); err != nil {
				t.Fatal(e, err)
			}
		}
		if err := facts.validate(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("rapid_dispatch_does_not_manufacture_age", func(t *testing.T) {
		p := enginePlan(engineUnit("a1", 3000), engineUnit("a2", 2000, "a1"), engineUnit("a3", 1000, "a2"), engineUnit("waiter", 1))
		r := boundedGraph(t, p, enginePorts{Now: func() time.Duration { return 0 }, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		if !slices.Equal(r.Admissions, []string{"a1", "a2", "a3", "waiter"}) {
			t.Fatal(r.Admissions)
		}
	})
	t.Run("elapsed_eligible_age_overtakes_new_critical_work", func(t *testing.T) {
		var clock atomic.Int64
		p := enginePlan(engineUnit("a1", 3000), engineUnit("a2", 2000, "a1"), engineUnit("a3", 1000, "a2"), engineUnit("waiter", 1))
		r := boundedGraph(t, p, enginePorts{Now: func() time.Duration { return time.Duration(clock.Load()) * time.Millisecond }, Execute: func(_ context.Context, req executionRequest) unitResult {
			if req.Unit.ID == "a1" {
				clock.Store(3000)
			}
			return unitResult{Status: "passed"}
		}})
		if !slices.Equal(r.Admissions, []string{"a1", "waiter", "a2", "a3"}) {
			t.Fatal(r.Admissions)
		}
	})
	t.Run("age_starts_at_eligibility_and_uses_whole_quanta", func(t *testing.T) {
		for _, elapsed := range []int64{999, 1000} {
			var clock atomic.Int64
			r := boundedGraph(t, enginePlan(engineUnit("seed", 10000), engineUnit("fresh", 500, "seed"), engineUnit("waiter", 1)), enginePorts{Now: func() time.Duration { return time.Duration(clock.Load()) * time.Millisecond }, Execute: func(_ context.Context, req executionRequest) unitResult {
				if req.Unit.ID == "seed" {
					clock.Store(elapsed)
				}
				return unitResult{Status: "passed"}
			}})
			want := []string{"seed", "fresh", "waiter"}
			if elapsed == 1000 {
				want = []string{"seed", "waiter", "fresh"}
			}
			if !slices.Equal(r.Admissions, want) {
				t.Fatal(elapsed, r.Admissions)
			}
		}
	})
	t.Run("retained_fixture_does_not_override_rank", func(t *testing.T) {
		a, b, c := engineUnit("seed", 100), engineUnit("warm", 1), engineUnit("critical", 20)
		a.Affinity, b.Affinity = "shared", "shared"
		r := boundedGraph(t, enginePlan(a, b, c), enginePorts{Now: func() time.Duration { return 0 }, Execute: func(_ context.Context, req executionRequest) unitResult {
			return unitResult{Status: "passed", RetainedFixture: req.Unit.ID == "seed"}
		}})
		if !slices.Equal(r.Admissions, []string{"seed", "critical", "warm"}) {
			t.Fatal(r.Admissions)
		}
	})
	t.Run("fitting_backfill_preserves_claims", func(t *testing.T) {
		a, b, c := engineUnit("holder", 10000), engineUnit("large", 100), engineUnit("fitting", 1)
		b.Claims["cpu"] = 2
		p := enginePlan(a, b, c)
		p.facts.Capacities["cpu"] = 2
		backfilled := make(chan struct{})
		r := boundedGraph(t, p, enginePorts{Now: func() time.Duration { return 0 }, Execute: func(ctx context.Context, req executionRequest) unitResult {
			if req.Unit.ID == "holder" {
				select {
				case <-backfilled:
				case <-ctx.Done():
				}
			}
			if req.Unit.ID == "fitting" {
				close(backfilled)
			}
			return unitResult{Status: "passed"}
		}})
		if !slices.Equal(r.Admissions, []string{"holder", "fitting", "large"}) {
			t.Fatal(r.Admissions)
		}
	})
	t.Run("critical_path_and_ASCII", func(t *testing.T) {
		p := enginePlan(engineUnit("a", 10), engineUnit("b", 10), engineUnit("c", 2), engineUnit("d", 30, "c"))
		r := boundedGraph(t, p, enginePorts{Now: func() time.Duration { return 0 }, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		if want := []string{"c", "d", "a", "b"}; !slices.Equal(r.Admissions, want) {
			t.Fatalf("admission %v want %v", r.Admissions, want)
		}
		if r.Status != "passed" {
			t.Fatalf("%+v", r.Failure)
		}
	})
	t.Run("exclusive_waiter_precedes_shared_backfill", func(t *testing.T) {
		a, b := engineUnit("a", 10000), engineUnit("b", 1)
		b.SharedLocks = []string{}
		b.ExclusiveLocks = []string{"host_activity"}
		r := boundedGraph(t, enginePlan(a, b), enginePorts{RetainEvents: true, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		found := false
		for _, e := range r.Events {
			if e["unit_id"] == "a" && e["wait_reason"] == "exclusive_barrier" {
				found = true
			}
		}
		if !found {
			t.Fatal("exclusive barrier not attributed")
		}
		if !slices.Equal(r.Admissions, []string{"b", "a"}) {
			t.Fatal(r.Admissions)
		}
	})
	t.Run("exclusive_barriers_name_only_overlapping_ready_owners", func(t *testing.T) {
		consumer := engineUnit("consumer", 1000)
		consumer.SharedLocks = append(consumer.SharedLocks, "left", "right")
		left, right, unrelated := engineUnit("left-holder", 3), engineUnit("right-holder", 2), engineUnit("unrelated", 1)
		left.ExclusiveLocks, right.ExclusiveLocks, unrelated.ExclusiveLocks = []string{"left"}, []string{"right"}, []string{"other"}
		p := enginePlan(consumer, left, right, unrelated)
		p.facts.Capacities["cpu"] = 2
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Now: func() time.Duration { return 0 }, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		found := false
		for _, event := range r.Events {
			if event["unit_id"] == "consumer" && event["event"] == "wait_started" && event["wait_reason"] == "exclusive_barrier" {
				if !found && (!slices.Equal(stringList(event["blocking_unit_ids"]), []string{"left-holder", "right-holder"}) || !slices.Equal(stringList(event["blocking_resources"]), []string{"lock_left", "lock_right"})) {
					t.Fatal(event)
				}
				found = true
			}
		}
		if !found || r.Status != "passed" {
			t.Fatal(found, r.Status)
		}
	})
	t.Run("parallel_capacity_and_dependency", func(t *testing.T) {
		p := enginePlan(engineUnit("a", 1), engineUnit("b", 1), engineUnit("c", 1, "a", "b"))
		p.facts.Capacities["cpu"] = 2
		var active, peak atomic.Int64
		started := make(chan struct{}, 2)
		r := boundedGraph(t, p, enginePorts{Execute: func(_ context.Context, r executionRequest) unitResult {
			n := active.Add(1)
			for prior := peak.Load(); n > prior && !peak.CompareAndSwap(prior, n); prior = peak.Load() {
			}
			if r.Unit.ID != "c" {
				started <- struct{}{}
				for len(started) < 2 {
					time.Sleep(time.Millisecond)
				}
			}
			active.Add(-1)
			return unitResult{Status: "passed"}
		}})
		if peak.Load() != 2 || r.Admissions[2] != "c" {
			t.Fatalf("peak %d, admission %v", peak.Load(), r.Admissions)
		}
	})
	t.Run("failure_propagation_and_nonblocking_finalizer", func(t *testing.T) {
		a := engineUnit("a", 20)
		a.FailurePolicy.BlockDescendants = false
		p := enginePlan(a, engineUnit("b", 1, "a"), engineUnit("c", 2), engineUnit("d", 1, "c"))
		r := boundedGraph(t, p, enginePorts{Execute: func(_ context.Context, r executionRequest) unitResult {
			if r.Unit.ID == "a" || r.Unit.ID == "c" {
				return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
			}
			return unitResult{Status: "passed"}
		}})
		if r.States["b"] != "passed" || r.States["d"] != "skipped" || r.Failure.Class != "product" {
			t.Fatalf("%+v", r)
		}
	})
}

type testCache struct {
	lookup func(context.Context, unit) (cacheResult, error)
	mu     sync.Mutex
	stored []string
	hits   map[string]bool
}

func (c *testCache) Lookup(ctx context.Context, u unit) (cacheResult, error) {
	if c.lookup != nil {
		return c.lookup(ctx, u)
	}
	if c.hits[u.ID] {
		return cacheResult{Outcome: "hit", Reason: "record_valid", OutputDigest: "sha256:" + strings.Repeat("a", 64)}, nil
	}
	return cacheResult{Outcome: "miss", Reason: "record_missing"}, nil
}
func (c *testCache) Store(_ context.Context, u unit) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stored = append(c.stored, u.ID)
	return nil
}

type testFixtures struct {
	acquire    func(context.Context, fixtureRequest) (*fixtureLease, error)
	close      func(context.Context) error
	unresolved bool
}

func (f *testFixtures) Acquire(ctx context.Context, r fixtureRequest) (*fixtureLease, error) {
	return f.acquire(ctx, r)
}
func (f *testFixtures) Close(ctx context.Context) error {
	if f.close != nil {
		return f.close(ctx)
	}
	return nil
}
func (f *testFixtures) Unresolved() bool { return f.unresolved }
func TestRuntimeLifecycle(t *testing.T) {
	t.Run("native_fixture_bridge_preserves_owner_facts_and_failure", func(t *testing.T) {
		acquired := false
		released := false
		environment := map[string]string{
			"CARTULARY_BROWSER_RUNTIME_PROFILE_ID": "network_flow_claimed", "CARTULARY_BROWSER_STAGE": "measurement",
			"CARTULARY_FIXTURE_PROFILE_ID": "import-snapshot", "CARTULARY_FIXTURE_SNAPSHOT_KEY": "snapshot-key",
			"CARTULARY_FIXTURE_SNAPSHOT_BUILDER_UNIT_ID": "default-builder", "CARTULARY_FIXTURE_ROW_ID": "row", "CARTULARY_FIXTURE_PREDICATE_ID": "predicate",
		}
		assertIdentity := func(r harnessfixtures.Request) {
			t.Helper()
			if r.RuntimeProfile != "network_flow_claimed" || r.BrowserStage != "measurement" || r.FixtureProfile != "import-snapshot" || r.SnapshotKey != "snapshot-key" || r.BuilderUnit != "default-builder" || r.RowID != "row" || r.PredicateID != "predicate" {
				t.Fatal("snapshot owner facts lost between graph, admission and acquisition", r)
			}
		}
		u := engineUnit("unit", 1)
		u.PostgresIsolation, u.Fixture, u.Services, u.Command.Environment = "exclusive", "browser_stack", []string{"object_store", "postgres"}, environment
		assertIdentity(workRequirement(u).Request)
		broker := harnessfixtures.New(harnessfixtures.Options{Providers: map[string]harnessfixtures.Provider{"browser_stack": &harnessfixtures.ProviderFuncs{AcquireFunc: func(_ context.Context, r harnessfixtures.Request) (*harnessfixtures.Allocation, error) {
			assertIdentity(r)
			if r.UnitID != "unit" || r.Affinity != "affinity" {
				t.Fatal(r)
			}
			acquired = true
			return &harnessfixtures.Allocation{Ownership: "owned", ResourceIDs: []string{"browser:one"}, Environment: map[string]string{"PORT": "1234"}, Release: func(_ context.Context, healthy bool) error {
				released = true
				if healthy {
					t.Fatal("failed work retained health")
				}
				return nil
			}}, nil
		}}}})
		bridge := nativeFixtures{production: &harnessfixtures.Production{Broker: broker}}
		lease, err := bridge.Acquire(context.Background(), fixtureRequest{PostgresIsolation: "exclusive", Services: []string{"object_store", "postgres"}, UnitID: "unit", Capability: "browser_stack", Affinity: "affinity", Environment: environment})
		if err != nil || !acquired || lease.Environment["PORT"] != "1234" || lease.ID == "" {
			t.Fatal(lease, err)
		}
		if _, err := lease.Release(context.Background(), false, false); err != nil || !released {
			t.Fatal(err)
		}
		if err := bridge.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		f := failureOf(&harnessfixtures.Failure{Class: "artifact", Reason: "artifact_error"})
		if f.Class != "artifact" || f.Reason != "artifact_error" {
			t.Fatal(f)
		}
	})

	t.Run("optional_observation_cannot_mutate_or_fail_execution", func(t *testing.T) {
		calls := 0
		r := boundedGraph(t, enginePlan(engineUnit("a", 1)), enginePorts{Observe: func(s engineSnapshot) { calls++; panic("failed optional sink") }, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		if r.Status != "passed" || !r.ObservationFailed || calls != 1 {
			t.Fatal(r, calls)
		}
	})
	t.Run("exclusive_observation_pause_drains_before_execution", func(t *testing.T) {
		u := engineUnit("a", 1)
		u.SharedLocks = nil
		u.ExclusiveLocks = []string{"host_activity"}
		paused, resumed := false, false
		r := boundedGraph(t, enginePlan(u), enginePorts{Pause: func(context.Context) error { paused = true; return nil }, Resume: func() { resumed = true }, Execute: func(context.Context, executionRequest) unitResult {
			if !paused || resumed {
				t.Error("exclusive execution outside quiet window")
			}
			return unitResult{Status: "passed"}
		}})
		if !resumed || r.Status != "passed" {
			t.Fatal(r, resumed)
		}
	})
	t.Run("cache_hit_never_admits_or_launches", func(t *testing.T) {
		cache := &testCache{hits: map[string]bool{"a": true}}
		r := boundedGraph(t, enginePlan(engineUnit("a", 1)), enginePorts{Cache: cache, Execute: func(context.Context, executionRequest) unitResult {
			t.Error("cache hit launched")
			return unitResult{Status: "passed"}
		}, Admit: func(context.Context, unit) (*hostLease, error) { t.Error("cache hit acquired host"); return nil, nil }})
		if r.Status != "passed" || len(r.Admissions) != 0 {
			t.Fatal(r)
		}
	})
	t.Run("warm_cache_waits_for_physical_cleanup", func(t *testing.T) {
		u := engineUnit("a", 1)
		u.Fixture = "browser_stack"
		u.Affinity = "browser"
		cache := &testCache{}
		closed := false
		fixtures := &testFixtures{acquire: func(context.Context, fixtureRequest) (*fixtureLease, error) {
			return &fixtureLease{ID: "lease", Release: func(context.Context, bool, bool) (bool, error) { return true, nil }}, nil
		}, close: func(context.Context) error {
			if len(cache.stored) != 0 {
				t.Error("cache published before physical cleanup")
			}
			closed = true
			return nil
		}}
		r := boundedGraph(t, enginePlan(u), enginePorts{Fixtures: fixtures, Cache: cache, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		if !closed || r.Status != "passed" || !slices.Equal(cache.stored, []string{"a"}) {
			t.Fatalf("close %v, stored %v, status %v", closed, cache.stored, r.Status)
		}
	})
	t.Run("cleanup_preserves_product_failure", func(t *testing.T) {
		r := boundedGraph(t, enginePlan(engineUnit("a", 1)), enginePorts{Execute: func(context.Context, executionRequest) unitResult {
			return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
		}, CleanupServices: func(context.Context) error { return failed("harness", "cleanup_error") }})
		if r.Failure.Class != "product" || len(r.Cleanup) != 1 || r.Cleanup[0].Failure.Reason != "cleanup_error" {
			t.Fatal(r)
		}
	})
	t.Run("cancellation_drains_before_independent_finalizer", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		drained := false
		r, err := executeGraph(ctx, enginePlan(engineUnit("a", 5), engineUnit("b", 1)), enginePorts{Execute: func(ctx context.Context, _ executionRequest) unitResult {
			cancel()
			<-ctx.Done()
			drained = true
			return unitResult{Status: "cancelled", Failure: failed("interrupted", "cancelled_or_interrupted")}
		}, Finalize: func(ctx context.Context, _ bool, _ []cleanupRecord) error {
			if ctx.Err() != nil || !drained {
				t.Error("finalizer received cancelled context or preceded drain")
			}
			return nil
		}})
		if err != nil || r.States["b"] != "cancelled" || !drained {
			t.Fatalf("%+v %v", r, err)
		}
	})
	t.Run("required_writer_failure_drains_host", func(t *testing.T) {
		var released atomic.Int64
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err := executeGraph(ctx, enginePlan(engineUnit("a", 1)), enginePorts{Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }, Admit: func(context.Context, unit) (*hostLease, error) {
			return &hostLease{Release: func(context.Context) error { released.Add(1); return nil }}, nil
		}, Event: func(e map[string]any) error {
			if e["event"] == "phase_finished" && e["phase"] == "host_admission" {
				return failed("artifact", "artifact_error")
			}
			return nil
		}})
		if err == nil || released.Load() != 1 || ctx.Err() != nil {
			t.Fatalf("release=%d error=%v context=%v", released.Load(), err, ctx.Err())
		}
	})
}
func TestRuntimeFailurePrecedence(t *testing.T) {
	for _, productFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("concurrent_failures_product_first_%v", productFirst), func(t *testing.T) {
			p := enginePlan(engineUnit("infra", 1), engineUnit("product", 1))
			p.facts.Capacities["cpu"] = 2
			arrived := make(chan struct{}, 2)
			firstDone := make(chan struct{})
			r := boundedGraph(t, p, enginePorts{Execute: func(_ context.Context, r executionRequest) unitResult {
				arrived <- struct{}{}
				for len(arrived) < 2 {
					time.Sleep(time.Millisecond)
				}
				first := (r.Unit.ID == "product") == productFirst
				if !first {
					<-firstDone
				} else {
					close(firstDone)
				}
				if r.Unit.ID == "product" {
					return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
				}
				return unitResult{Status: "failed", Failure: failed("infra", "service_start_error")}
			}})
			if r.Failure.Class != "product" || r.States["infra"] != "failed" || r.States["product"] != "failed" {
				t.Fatal(r)
			}
		})
	}

	product := failed("product", "test_assertion_failure")
	cleanup := failed("config", "configuration_error")
	cleanup.Lifecycle = "cleanup_finalizers"
	for _, ordered := range [][]*failure{{cleanup, product}, {product, cleanup}} {
		if got := primaryFailure(ordered); got.Class != "product" {
			t.Fatal(got)
		}
	}
	a, b := failed("infra", "service_start_error"), failed("infra", "preflight_error")
	a.Sequence = 2
	b.Sequence = 1
	if got := primaryFailure([]*failure{a, b}); got.Reason != "preflight_error" {
		t.Fatal(got)
	}
	if failureExitCode(failed("interrupted", "cancelled_or_interrupted"), "SIGTERM") != 143 || failureExitCode(failed("artifact", "artifact_error"), "") != 11 {
		t.Fatal("exit mapping")
	}
}
