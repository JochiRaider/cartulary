package harnessruntime

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
)

// fakeLedger is a host ledger for one graph: physical claims are charged on
// grant and released with the lease; every change is signalled.
type fakeLedger struct {
	mu       sync.Mutex
	capacity map[string]int64
	used     map[string]int64
	changes  chan struct{}
	tried    []string
	deny     bool
	// refuse names units whose evaluation the ledger refuses.
	refuse map[string]bool
}

func newFakeLedger(capacity map[string]int64) *fakeLedger {
	return &fakeLedger{capacity: capacity, used: map[string]int64{}, changes: make(chan struct{}, 1)}
}

// charges are a unit's physical claims with its residents', as the host
// ledger charges them together.
func (l *fakeLedger) charges(u unit) map[string]int64 {
	charges := map[string]int64{}
	for resource, amount := range u.Claims {
		if l.Physical(resource) {
			charges[resource] += amount
		}
	}
	for _, resident := range u.Residents {
		for resource, amount := range resident.Claims {
			charges[resource] += amount
		}
	}
	return charges
}
func (l *fakeLedger) fits(u unit) []string {
	deficits := []string{}
	for resource, amount := range l.charges(u) {
		if l.used[resource]+amount > l.capacity[resource] {
			deficits = append(deficits, resource)
		}
	}
	slices.Sort(deficits)
	return deficits
}
func (l *fakeLedger) View(context.Context) (admissionView, error) { return l, nil }
func (l *fakeLedger) Evaluate(u unit) (verdict, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.deny {
		return verdict{deficits: []string{"cpu"}}, nil
	}
	if l.refuse[u.ID] {
		return verdict{}, errors.New("host admission unavailable")
	}
	deficits := l.fits(u)
	return verdict{admissible: len(deficits) == 0, deficits: deficits}, nil
}
func (l *fakeLedger) TryAdmit(_ context.Context, u unit) (*hostLease, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tried = append(l.tried, u.ID)
	if l.deny || len(l.fits(u)) != 0 {
		return nil, false, nil
	}
	charges := l.charges(u)
	for resource, amount := range charges {
		l.used[resource] += amount
	}
	return &hostLease{Token: u.ID, Release: func(context.Context) error {
		l.mu.Lock()
		for resource, amount := range charges {
			l.used[resource] -= amount
		}
		l.mu.Unlock()
		select {
		case l.changes <- struct{}{}:
		default:
		}
		return nil
	}}, true, nil
}
func (l *fakeLedger) Changed() <-chan struct{}      { return l.changes }
func (l *fakeLedger) Physical(resource string) bool { return resource != "browser_lane" }

func TestRuntimeLedgerAdmission(t *testing.T) {
	t.Run("waiting work holds nothing and admissible work backfills", func(t *testing.T) {
		a, b, c := engineUnit("a", 30), engineUnit("b", 20), engineUnit("c", 10)
		a.Claims["postgres"], b.Claims["postgres"] = 1, 1
		p := enginePlan(a, b, c)
		ledger := newFakeLedger(map[string]int64{"cpu": 3, "postgres": 1})
		// The graph's own capacities would admit all three; only the ledger
		// bounds them.
		p.facts.Capacities["cpu"], p.facts.Capacities["postgres"] = 3, 3
		release := make(chan struct{})
		var cRan atomic.Bool
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Admission: ledger, Execute: func(ctx context.Context, req executionRequest) unitResult {
			switch req.Unit.ID {
			case "a":
				select {
				case <-release:
				case <-ctx.Done():
				}
			case "c":
				cRan.Store(true)
				close(release)
			}
			return unitResult{Status: "passed"}
		}})
		if r.Status != "passed" || !slices.Equal(r.Admissions, []string{"a", "c", "b"}) || !cRan.Load() {
			t.Fatal("ledger admission order", r.Admissions, r.Status)
		}
		var waitedOn []string
		for _, event := range r.Events {
			if event["unit_id"] == "b" && event["event"] == "wait_started" && event["wait_reason"] == "resources" {
				waitedOn = event["blocking_resources"].([]string)
			}
		}
		if !slices.Equal(waitedOn, []string{"postgres"}) {
			t.Fatal("denied unit's wait does not name its deficit", waitedOn)
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
	})
	t.Run("graph-local browser lanes stay with the graph", func(t *testing.T) {
		units := []unit{engineUnit("a", 3), engineUnit("b", 2), engineUnit("c", 1)}
		for i := range units {
			units[i].Claims["browser_lane"] = 1
		}
		p := enginePlan(units...)
		p.facts.Capacities["cpu"], p.facts.Capacities["browser_lane"] = 3, 1
		ledger := newFakeLedger(map[string]int64{"cpu": 3})
		var active, peak atomic.Int64
		r := boundedGraph(t, p, enginePorts{Admission: ledger, Execute: func(context.Context, executionRequest) unitResult {
			now := active.Add(1)
			defer active.Add(-1)
			for prior := peak.Load(); now > prior && !peak.CompareAndSwap(prior, now); prior = peak.Load() {
			}
			return unitResult{Status: "passed"}
		}})
		if r.Status != "passed" || peak.Load() != 1 {
			t.Fatal("browser lanes exceeded", peak.Load(), r.Status)
		}
	})
	t.Run("a reusable result completes without host admission", func(t *testing.T) {
		a := engineUnit("a", 1)
		a.CachePolicy = "content_addressed"
		p := enginePlan(a)
		ledger := newFakeLedger(map[string]int64{"cpu": 1})
		r := boundedGraph(t, p, enginePorts{Admission: ledger, Cache: hitCache{}, Execute: func(context.Context, executionRequest) unitResult {
			t.Error("a cache hit executed")
			return unitResult{Status: "passed"}
		}})
		if r.Status != "passed" || len(ledger.tried) != 0 {
			t.Fatal("cache hit reached host admission", ledger.tried, r.Status)
		}
	})
	t.Run("with nothing running the top-ranked unit waits in host admission", func(t *testing.T) {
		a, b := engineUnit("a", 2), engineUnit("b", 1)
		p := enginePlan(a, b)
		ledger := newFakeLedger(map[string]int64{"cpu": 1})
		ledger.deny = true
		var waited []string
		var mu sync.Mutex
		r := boundedGraph(t, p, enginePorts{Admission: ledger, Admit: func(_ context.Context, u unit) (*hostLease, error) {
			mu.Lock()
			waited = append(waited, u.ID)
			mu.Unlock()
			return &hostLease{Release: func(context.Context) error { return nil }}, nil
		}, Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		if r.Status != "passed" || !slices.Equal(waited, []string{"a", "b"}) {
			t.Fatal("blocking admission fallback", waited, r.Status)
		}
	})
}

// hitCache reuses every result.
type hitCache struct{}

func (hitCache) Lookup(context.Context, unit) (cacheResult, error) {
	return cacheResult{Outcome: "hit", Reason: "record_valid", Profile: "test_rows", OutputDigest: "sha256:" + strings.Repeat("0", 64)}, nil
}
func (hitCache) Store(context.Context, unit) error { return nil }

func TestRuntimeRefusedEvaluation(t *testing.T) {
	// The ledger refuses to evaluate "refused": it goes to its own blocking
	// admission, which decides it, instead of failing the whole graph.
	refused, other := engineUnit("refused", 20), engineUnit("other", 10)
	p := enginePlan(refused, other)
	ledger := newFakeLedger(map[string]int64{"cpu": 2})
	ledger.refuse = map[string]bool{"refused": true}
	var blocking []string
	var mu sync.Mutex
	r := boundedGraph(t, p, enginePorts{RetainEvents: true, Admission: ledger,
		Admit: func(_ context.Context, u unit) (*hostLease, error) {
			mu.Lock()
			blocking = append(blocking, u.ID)
			mu.Unlock()
			return &hostLease{Token: strings.Repeat("a", 32), Release: func(context.Context) error { return nil }}, nil
		},
		Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
	if r.Status != "passed" || !slices.Equal(blocking, []string{"refused"}) || slices.Contains(ledger.tried, "refused") {
		t.Fatal("refused evaluation", r.Status, blocking, ledger.tried)
	}
}

func TestRuntimeResidentHeldWaits(t *testing.T) {
	generation := func(id string) []harnessfixtures.Resident {
		return []harnessfixtures.Resident{{Key: "postgres:unit:" + id, Claims: map[string]int64{"postgres_generation": 1}}}
	}
	waitOf := func(r graphResult, id string) (string, []string, []string) {
		for _, event := range r.Events {
			if event["unit_id"] == id && event["event"] == "wait_started" && event["wait_reason"] != "rank" {
				return event["wait_reason"].(string), event["blocking_resources"].([]string), event["blocking_unit_ids"].([]string)
			}
		}
		return "", nil, nil
	}
	t.Run("a running unit holding the resource through its resident is the holder", func(t *testing.T) {
		holder, waiter := engineUnit("holder", 20), engineUnit("waiter", 10)
		holder.Residents, waiter.Residents = generation("holder"), generation("waiter")
		p := enginePlan(holder, waiter)
		ledger := newFakeLedger(map[string]int64{"cpu": 2, "postgres_generation": 1})
		release := make(chan struct{})
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Admission: ledger, Execute: func(ctx context.Context, req executionRequest) unitResult {
			if req.Unit.ID == "holder" {
				select {
				case <-release:
				case <-time.After(200 * time.Millisecond):
				}
			}
			return unitResult{Status: "passed"}
		}})
		close(release)
		reason, resources, units := waitOf(r, "waiter")
		if r.Status != "passed" || reason != "resources" || !slices.Equal(resources, []string{"postgres_generation"}) || !slices.Equal(units, []string{"holder"}) {
			t.Fatal(r.Status, reason, resources, units)
		}
	})
	t.Run("capacity no running unit holds sends one unit to host admission and the rest wait by rank", func(t *testing.T) {
		first, second := engineUnit("first", 20), engineUnit("second", 10)
		first.Residents, second.Residents = generation("first"), generation("second")
		p := enginePlan(first, second)
		ledger := newFakeLedger(map[string]int64{"cpu": 2, "postgres_generation": 1})
		ledger.used["postgres_generation"] = 1
		var waited []string
		var mu sync.Mutex
		r := boundedGraph(t, p, enginePorts{RetainEvents: true, Admission: ledger,
			Admit: func(ctx context.Context, u unit) (*hostLease, error) {
				mu.Lock()
				waited = append(waited, u.ID)
				mu.Unlock()
				return &hostLease{Token: strings.Repeat("a", 32), Release: func(context.Context) error { return nil }}, nil
			},
			Execute: func(context.Context, executionRequest) unitResult { return unitResult{Status: "passed"} }})
		// While first waits and then runs it holds the generation through
		// its resident, so second waits on it by name.
		reason, resources, units := waitOf(r, "second")
		if r.Status != "passed" || !slices.Equal(waited, []string{"first", "second"}) || reason != "resources" || !slices.Equal(resources, []string{"postgres_generation"}) || !slices.Equal(units, []string{"first"}) {
			t.Fatal(r.Status, waited, reason, resources, units)
		}
	})
}
