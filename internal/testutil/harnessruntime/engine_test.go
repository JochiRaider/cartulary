package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
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
	return unit{PostgresIsolation: "none", Renderer: "none", ID: id, Owner: "harness.command_surface", Kind: "runner", Needs: needs, Claims: map[string]int64{"cpu": 1}, SharedLocks: []string{"host_activity"}, ExclusiveLocks: []string{}, Fixture: "none", Services: []string{}, CachePolicy: "none", Evidence: []string{}, Artifacts: []artifact{}, EstimateMS: estimate, TimeoutMS: 1000, Command: command{Executable: "true", Args: []string{}, Environment: map[string]string{}}, FailurePolicy: failurePolicy{BlockDescendants: true, ContinueIndependent: true, AggregateEffect: "required"}}
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

// endedBlockers names the resources and units a wait_ended attributes.
func endedBlockers(event map[string]any) (resources, units []string) {
	list, _ := objectList(event["blocked_resources"])
	for _, r := range list {
		resources = append(resources, r["resource"].(string))
	}
	list, _ = objectList(event["blocking_units"])
	for _, u := range list {
		units = append(units, u["unit_id"].(string))
	}
	return resources, units
}
func TestWaitSegmentAttribution(t *testing.T) {
	s := newWaitSegment(waitDetails{"resources", []string{"cpu"}, []string{"a"}}, 10)
	s.decide(waitDetails{"resources", []string{"cpu", "memory"}, []string{"a", "b"}}, 14)
	s.decide(waitDetails{"resources", []string{"memory"}, []string{"c"}}, 20)
	s.decide(s.current, 25)
	got, _ := json.Marshal(s.attribution())
	want := `{"blocked_resources":[{"blocked_ms":10,"resource":"cpu"},{"blocked_ms":11,"resource":"memory"}],"blocking_units":[{"overlap_ms":10,"unit_id":"a"},{"overlap_ms":6,"unit_id":"b"},{"overlap_ms":5,"unit_id":"c"}],"blocking_units_omitted":0,"wait_reason":"resources"}`
	if string(got) != want {
		t.Fatalf("attribution %s want %s", got, want)
	}
	// A blocker decided at the closing instant counts zero but is named; the
	// bound keeps the eight largest overlaps, ties by unit ID.
	ids := []string{"h0", "h1", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9"}
	s = newWaitSegment(waitDetails{"resources", []string{"cpu"}, ids[:9]}, 0)
	s.decide(waitDetails{"resources", []string{"cpu"}, ids}, 4)
	s.decide(s.current, 4)
	a := s.attribution()
	_, units := endedBlockers(a)
	if !slices.Equal(units, []string{"h0", "h1", "h2", "h3", "h4", "h5", "h6", "h7"}) || a["blocking_units_omitted"] != 2 {
		t.Fatal(a)
	}
}
func TestPoolAwareRanks(t *testing.T) {
	// postgres (capacity 2) bounds this graph: 450 ms of its work against
	// 75 ms per CPU token. gate is a short build that pgDep needs.
	pgLong, pgShort, side := engineUnit("pg-long", 300), engineUnit("pg-short", 100), engineUnit("side", 150)
	gate, pgDep := engineUnit("gate", 5), engineUnit("pg-dep", 50, "gate")
	for _, u := range []*unit{&pgLong, &pgShort, &pgDep} {
		u.Claims["postgres"] = 1
	}
	p := enginePlan(pgLong, pgShort, side, gate, pgDep)
	p.facts.Capacities["cpu"], p.facts.Capacities["postgres"] = 8, 2
	states := map[string]string{}
	for _, u := range p.facts.Graph.Units {
		states[u.ID] = "pending"
	}
	e := &graphEngine{plan: p, byID: map[string]unit{}, admittedAt: map[string]int64{}, result: graphResult{States: states}}
	for _, u := range p.facts.Graph.Units {
		e.byID[u.ID] = u
	}
	e.computeRanks(0)
	r := e.ranks
	// Bottleneck work outranks side work, longer pool work goes first, and
	// the gate inherits its dependent's pool rank instead of its own 5 ms.
	if !(r["pg-long"] > r["pg-short"] && r["pg-short"] > r["side"] && r["gate"] > r["side"] && r["gate"] > r["pg-dep"]) {
		t.Fatal("pool-aware order", r)
	}
	if r["pg-short"] != 100+450/2 || r["side"] != 150+605/8 {
		t.Fatal("pool bounds", r)
	}
	// The pool drains as its work runs, so ranks fall with it.
	e.result.States["pg-long"], e.admittedAt["pg-long"] = "running", 0
	e.computeRanks(150)
	if e.ranks["pg-short"] != 100+300/2 {
		t.Fatal("pool bound did not drain", e.ranks)
	}
}
func TestRuntimeScheduling(t *testing.T) {
	testConsumerTerminalTransitions(t)
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
	// TH-HARNESS-REQ-353: postgres_heavy is additive; two lanes admit two
	// heavy units and the third waits for a lane to be released.
	t.Run("postgres_heavy_lanes_bound_concurrent_heavy_units", func(t *testing.T) {
		a, b, c := engineUnit("a", 3), engineUnit("b", 2), engineUnit("c", 1)
		for _, u := range []unit{a, b, c} {
			u.Claims["postgres_heavy"] = 1
		}
		p := enginePlan(a, b, c)
		p.facts.Capacities["cpu"], p.facts.Capacities["postgres_heavy"] = 3, 2
		// The first two lane holders must overlap; a third may start only after
		// one of them has released its lane.
		var active, peak, arrived atomic.Int64
		both := make(chan struct{})
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Execute: func(ctx context.Context, _ executionRequest) unitResult {
			now := active.Add(1)
			defer active.Add(-1)
			for {
				prior := peak.Load()
				if now <= prior || peak.CompareAndSwap(prior, now) {
					break
				}
			}
			if order := arrived.Add(1); order <= 2 {
				if order == 2 {
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
		if r.Status != "passed" || peak.Load() != 2 {
			t.Fatal("heavy lanes did not bound concurrent heavy units", r.Status, peak.Load())
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
		if peaks := facts.pressure()["peak_use"].(map[string]int64); peaks["postgres_heavy"] != 2 {
			t.Fatal("canonical pressure diverged from heavy-lane admission", peaks)
		}
	})
	t.Run("wait_segments_follow_reservations_and_barriers", func(t *testing.T) {
		p := enginePlan(engineUnit("holder", 10), engineUnit("waiter", 1))
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		var reasons []string
		for _, e := range r.Events {
			if e["unit_id"] == "waiter" && e["event"] == "wait_started" {
				reasons = append(reasons, e["wait_reason"].(string))
			}
			if _, units := endedBlockers(e); e["unit_id"] == "waiter" && e["event"] == "wait_ended" && e["wait_reason"] == "resources" && !slices.Equal(units, []string{"holder"}) {
				t.Fatal(e)
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
	t.Run("changing_holders_stay_in_one_attributed_segment", func(t *testing.T) {
		first, second, waiter := engineUnit("first", 100), engineUnit("second", 90), engineUnit("waiter", 1)
		waiter.Claims["cpu"] = 2
		p := enginePlan(first, second, waiter)
		p.facts.Capacities["cpu"] = 2
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Execute: func(_ context.Context, request executionRequest) unitResult {
			time.Sleep(map[string]time.Duration{"first": 5 * time.Millisecond, "second": 20 * time.Millisecond}[request.Unit.ID])
			return unitResult{Status: "passed"}
		}})
		segments := 0
		for _, e := range r.Events {
			if e["unit_id"] != "waiter" || e["event"] != "wait_ended" || e["wait_reason"] != "resources" {
				continue
			}
			segments++
			resources, units := endedBlockers(e)
			slices.Sort(units)
			if !slices.Equal(resources, []string{"cpu"}) || !slices.Equal(units, []string{"first", "second"}) || e["blocking_units_omitted"] != 0 {
				t.Fatal("holder change was not attributed to the segment", e)
			}
		}
		if segments != 1 {
			t.Fatal("a holder change opened a new segment", segments)
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
	t.Run("look_ahead_reserves_for_imminent_higher_ranked_work", func(t *testing.T) {
		// cohort needs build and outranks long. While build runs, cohort is
		// expected at 10 ms; long would outlast that and take the CPU cohort
		// needs, so it waits until cohort has started.
		plan := func(buildMS int64) *Plan {
			build, cohort := engineUnit("build", buildMS), engineUnit("cohort", 600, "build")
			long := engineUnit("long", 500)
			cohort.Claims["cpu"] = 2
			p := enginePlan(build, cohort, long)
			p.facts.Capacities["cpu"] = 2
			return p
		}
		run := func(p *Plan) []string {
			return boundedGraph(t, p, enginePorts{Now: func() time.Duration { return 0 }, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }}).Admissions
		}
		if got := run(plan(10)); !slices.Equal(got, []string{"build", "cohort", "long"}) {
			t.Fatal("look-ahead did not hold long for cohort", got)
		}
		// Beyond the horizon nothing is reserved.
		if got := run(plan(lookAheadHorizonMS + 10)); !slices.Equal(got[:2], []string{"build", "long"}) {
			t.Fatal("look-ahead reserved beyond its horizon", got)
		}
	})
	t.Run("look_ahead_reservations_are_bounded_by_capacity", func(t *testing.T) {
		// Two imminent cohorts each need all the CPU; only the first can start
		// when build ends, so only it reserves, and only against long.
		build := engineUnit("build", 10)
		first, second := engineUnit("first", 700, "build"), engineUnit("second", 650, "build")
		first.Claims["cpu"], second.Claims["cpu"] = 2, 2
		long := engineUnit("long", 500)
		p := enginePlan(build, first, second, long)
		p.facts.Capacities["cpu"] = 2
		e := &graphEngine{plan: p, byID: map[string]unit{}, ranks: map[string]int64{}, running: map[string]bool{"build": true},
			eligible: map[string]bool{}, admittedAt: map[string]int64{"build": 0}, result: graphResult{States: map[string]string{"build": "running", "first": "pending", "second": "pending", "long": "pending"}}}
		for _, u := range p.facts.Graph.Units {
			e.byID[u.ID] = u
		}
		e.computeRanks(0)
		soon := e.imminent(0)
		if len(soon) != 2 || soon[0].unit.ID != "first" || soon[0].at != 10 {
			t.Fatal("imminent units", soon)
		}
		if lacking := e.lookAheadHold(long, e.ranks["long"], 0, soon); !slices.Equal(lacking, []string{"cpu"}) {
			t.Fatal("long was not held for first", lacking)
		}
		short := engineUnit("short", 10)
		if lacking := e.lookAheadHold(short, e.ranks["short"], 0, soon); lacking != nil {
			t.Fatal("work that ends by then was held", lacking)
		}
		if lacking := e.lookAheadHold(first, e.ranks["build"]+1, 0, soon); lacking != nil {
			t.Fatal("a unit outranking every imminent unit was held", lacking)
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
	quietUnit := func(id string, estimate int64, needs ...string) unit {
		u := engineUnit(id, estimate, needs...)
		u.SharedLocks, u.ExclusiveLocks = []string{}, []string{"host_activity"}
		return u
	}
	passing := func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }
	// TH-HARNESS-REQ-806: quiet work is not ready while ordinary work is ready,
	// whatever its rank; once released, it still bars later shared work.
	t.Run("quiet_work_waits_for_ordinary_work_then_bars_later_shared_work", func(t *testing.T) {
		r := boundedGraph(t, enginePlan(engineUnit("a", 10), quietUnit("q", 10000), engineUnit("c", 10, "a")), enginePorts{RetainEvents: true, Execute: passing})
		if !slices.Equal(r.Admissions, []string{"a", "q", "c"}) || r.Status != "passed" {
			t.Fatal(r.Admissions, r.Status)
		}
		eligible, barred := -1, false
		for i, e := range r.Events {
			if e["unit_id"] == "q" && e["event"] == "eligible" {
				eligible = i
			}
			if e["unit_id"] == "a" && e["event"] == "admitted" && eligible >= 0 {
				t.Fatal("quiet unit became ready while ordinary work was ready")
			}
			barred = barred || e["unit_id"] == "c" && e["wait_reason"] == "exclusive_barrier"
		}
		if eligible < 0 || !barred {
			t.Fatal("quiet release or barrier not attributed", eligible, barred)
		}
	})
	// Quiet units released together stay ready, so the summary that one of them
	// unblocks waits behind the others instead of splitting the quiet block.
	t.Run("quiet_units_released_together_run_as_one_block", func(t *testing.T) {
		p := enginePlan(engineUnit("a", 1), quietUnit("q1", 20), quietUnit("q2", 10), engineUnit("s1", 1, "q1"))
		r := boundedGraph(t, p, enginePorts{Execute: passing})
		if !slices.Equal(r.Admissions, []string{"a", "q1", "q2", "s1"}) || r.Status != "passed" {
			t.Fatal(r.Admissions, r.Status)
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
			if event["unit_id"] == "consumer" && event["event"] == "wait_ended" && event["wait_reason"] == "exclusive_barrier" {
				resources, units := endedBlockers(event)
				slices.Sort(units)
				if !found && (!slices.Equal(units, []string{"left-holder", "right-holder"}) || !slices.Equal(resources, []string{"lock_left", "lock_right"})) {
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
		var arrived atomic.Int64
		started := make(chan struct{})
		r := boundedGraph(t, p, enginePorts{Execute: func(ctx context.Context, r executionRequest) unitResult {
			n := active.Add(1)
			defer active.Add(-1)
			for prior := peak.Load(); n > prior && !peak.CompareAndSwap(prior, n); prior = peak.Load() {
			}
			if r.Unit.ID != "c" {
				if arrived.Add(1) == 2 {
					close(started)
				}
				select {
				case <-started:
				case <-ctx.Done():
					return unitResult{Status: "cancelled", Failure: failureOf(ctx.Err())}
				}
			}
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
		if _, err := lease.Release(context.Background(), false); err != nil || !released {
			t.Fatal(err)
		}
		if err := bridge.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		ref := "_shared/test-services/fixture/service-scope.json"
		f := failureOf(&harnessfixtures.Failure{Class: "artifact", Reason: "artifact_error", Refs: []string{ref, "/private/foreign"}})
		if f.Class != "artifact" || f.Reason != "artifact_error" || len(f.ServiceRefs) != 1 || f.ServiceRefs[0] != ref {
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
			return &fixtureLease{ID: "lease", Release: func(context.Context, bool) (bool, error) { return true, nil }}, nil
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
	t.Run("bounded_entry_stage_preserves_failure", func(t *testing.T) {
		private := errors.New("private path and credential must not be exposed")
		for _, stage := range []string{"nested_lifecycle", "diagnostic_mode", "accounting_membership", "accounting_required_scope", "accounting_executable_identity", "accounting_invocation_identity", "runtime_plan"} {
			for _, cause := range []error{errors.Join(failed("config", "configuration_error"), private), context.Canceled, context.DeadlineExceeded, failed("harness", "cleanup_error")} {
				wrapped := fmt.Errorf("private wrapper: %w", entryFailureAt(stage, cause))
				beforeClass, beforeReason := PublicFailure(cause)
				class, reason := PublicFailure(wrapped)
				if PublicFailureStage(wrapped) != stage || class != beforeClass || reason != beforeReason || !errors.Is(wrapped, cause) || failureExitCode(failureOf(wrapped), "") != failureExitCode(failureOf(cause), "") {
					t.Fatal("stage attribution changed the failure contract")
				}
			}
		}
		for _, err := range []error{nil, private, entryFailureAt(private.Error(), private)} {
			if PublicFailureStage(err) != "unavailable" {
				t.Fatal("unattributed or unknown stage escaped the closed vocabulary")
			}
		}
	})
	for _, productFirst := range []bool{false, true} {
		t.Run(fmt.Sprintf("concurrent_failures_product_first_%v", productFirst), func(t *testing.T) {
			p := enginePlan(engineUnit("infra", 1), engineUnit("product", 1))
			p.facts.Capacities["cpu"] = 2
			var arrived atomic.Int64
			both := make(chan struct{})
			firstDone := make(chan struct{})
			wantOrder := []string{"infra", "product"}
			if productFirst {
				wantOrder = []string{"product", "infra"}
			}
			var order []string
			r := boundedGraph(t, p, enginePorts{Event: func(event map[string]any) error {
				if event["event"] == "failed" {
					id := event["unit_id"].(string)
					order = append(order, id)
					if id == wantOrder[0] {
						close(firstDone)
					}
				}
				return nil
			}, Execute: func(ctx context.Context, r executionRequest) unitResult {
				if arrived.Add(1) == 2 {
					close(both)
				}
				select {
				case <-both:
				case <-ctx.Done():
					return unitResult{Status: "cancelled", Failure: failureOf(ctx.Err())}
				}
				if r.Unit.ID != wantOrder[0] {
					select {
					case <-firstDone:
					case <-ctx.Done():
						return unitResult{Status: "cancelled", Failure: failureOf(ctx.Err())}
					}
				}
				if r.Unit.ID == "product" {
					return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
				}
				return unitResult{Status: "failed", Failure: failed("infra", "service_start_error")}
			}})
			if r.Failure.Class != "product" || r.States["infra"] != "failed" || r.States["product"] != "failed" || !slices.Equal(order, wantOrder) {
				t.Fatal(r, order)
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

func testConsumerTerminalTransitions(t *testing.T) {
	t.Run("consumer_terminal_covers_cache_skip_failure_and_pass", func(t *testing.T) {
		units := []unit{engineUnit("cached", 100), engineUnit("failed", 90), engineUnit("skipped", 80, "failed"), engineUnit("passed", 10)}
		notifications := map[string]string{}
		launches := []string{}
		result := boundedGraph(t, enginePlan(units...), enginePorts{Cache: &testCache{hits: map[string]bool{"cached": true}}, ConsumerTerminal: func(id, status string) error {
			if _, ok := notifications[id]; ok {
				t.Errorf("duplicate terminal %s", id)
			}
			notifications[id] = status
			return nil
		}, Execute: func(_ context.Context, r executionRequest) unitResult {
			launches = append(launches, r.Unit.ID)
			if r.Unit.ID == "failed" {
				return unitResult{Status: "failed", Failure: failed("product", "test_assertion_failure")}
			}
			return unitResult{Status: "passed"}
		}})
		for _, id := range []string{"cached", "failed", "skipped", "passed"} {
			if notifications[id] != id {
				t.Fatal("incomplete terminal accounting", notifications, result.States)
			}
		}
		if !slices.Equal(launches, []string{"failed", "passed"}) {
			t.Fatal("invented runner", launches)
		}
	})
	for _, fault := range []string{"admission", "publication", "cancel_before", "cancel_running", "notification"} {
		t.Run("consumer_terminal_"+fault, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if fault == "cancel_before" {
				cancel()
			}
			notifications := map[string]string{}
			ports := enginePorts{ConsumerTerminal: func(id, status string) error {
				if _, ok := notifications[id]; ok {
					t.Error("duplicate terminal")
				}
				notifications[id] = status
				if fault == "notification" {
					return failed("harness", "fixture_error")
				}
				return nil
			}, Execute: func(context.Context, executionRequest) unitResult {
				if fault == "cancel_running" {
					cancel()
					return unitResult{Status: "cancelled", Failure: failed("interrupted", "cancelled_or_interrupted")}
				}
				return unitResult{Status: "passed"}
			}}
			if fault == "admission" {
				ports.Admit = func(context.Context, unit) (*hostLease, error) { return nil, failed("infra", "resource_conflict") }
			}
			if fault == "publication" {
				ports.Terminal = func(unit, *unitResult) error { return failed("artifact", "artifact_error") }
			}
			result, _ := executeGraph(ctx, enginePlan(engineUnit("first", 10), engineUnit("next", 1, "first")), ports)
			if len(notifications) != 2 {
				t.Fatal("missing terminal", notifications, result.States)
			}
			for id, status := range result.States {
				if notifications[id] != status {
					t.Fatal("terminal differs", id, status, notifications)
				}
			}
			if result.Status == "passed" {
				t.Fatal("fault hidden", fault)
			}
		})
	}
}
