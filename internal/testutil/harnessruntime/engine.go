package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

type failure struct {
	Class      string `json:"failure_class"`
	Reason     string `json:"failure_reason"`
	Lifecycle  string `json:"lifecycle_step,omitempty"`
	Sequence   int64  `json:"scheduler_event_sequence,omitempty"`
	ChildOrder int64  `json:"child_registry_order,omitempty"`
	Artifact   string `json:"artifact,omitempty"`
}

func (f *failure) Error() string           { return f.Class + "/" + f.Reason }
func failed(class, reason string) *failure { return &failure{Class: class, Reason: reason} }
func failureOf(err error) *failure {
	var f *failure
	if errors.As(err, &f) {
		copy := *f
		return &copy
	}
	var typed interface{ HarnessFailure() (string, string) }
	if errors.As(err, &typed) {
		class, reason := typed.HarnessFailure()
		return normalizedFailure(failed(class, reason))
	}
	if errors.Is(err, context.Canceled) {
		return failed("interrupted", "cancelled_or_interrupted")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return failed("timing", "timeout_failure")
	}
	return failed("harness", "fixture_error")
}

type unitResult struct {
	Status          string
	Failure         *failure
	ExitCode        int
	Signal          string
	Stdout, Stderr  []byte
	ArtifactRefs    []string
	MissingOutputs  []string
	RetainedFixture bool
	Cleanup         []*failure
	Cache           cacheResult
}
type cacheResult struct{ Outcome, Reason, Profile, OutputDigest string }
type graphCache interface {
	Lookup(context.Context, unit) (cacheResult, error)
	Store(context.Context, unit) error
}
type fixtureRequest struct {
	PostgresIsolation                   string
	Services                            []string
	UnitID, Owner, Capability, Affinity string
	Environment                         map[string]string
}
type fixtureLease struct {
	ID          string
	Environment map[string]string
	Release     func(context.Context, bool, bool) (bool, error)
}
type fixtureOwner interface {
	Acquire(context.Context, fixtureRequest) (*fixtureLease, error)
	Close(context.Context) error
	Unresolved() bool
}
type hostLease struct {
	Token   string
	Bind    func(int) error
	Release func(context.Context) error
}
type executionRequest struct {
	Unit        unit
	Environment map[string]string
	Host        *hostLease
}
type cleanupRecord struct {
	Cause     error `json:"-"`
	Operation string
	UnitID    string
	LeaseID   string
	Outcome   string
	Failure   *failure
}
type graphResult struct {
	ObservationFailed bool
	Status            string
	DurationMS        int64
	Failure           *failure
	States            map[string]string
	Units             map[string]unitResult
	Admissions        []string
	Events            []map[string]any
	Cleanup           []cleanupRecord
}
type enginePorts struct {
	Cleanup         func(cleanupRecord)
	RetainEvents    bool
	Now             func() time.Duration
	Execute         func(context.Context, executionRequest) unitResult
	Cache           graphCache
	Fixtures        fixtureOwner
	Admit           func(context.Context, unit) (*hostLease, error)
	Event           func(map[string]any) error
	Terminal        func(unit, *unitResult) error
	CleanupServices func(context.Context) error
	Finalize        func(context.Context, bool, []cleanupRecord) error
	Pause           func(context.Context) error
	Resume          func()
	Observe         func(engineSnapshot)
}
type engineSnapshot struct {
	Phase     string
	ElapsedMS int64
	Units     []unitSnapshot
}
type unitSnapshot struct {
	ID, Status, Activity, WaitReason, Cache string
	StartedMS                               *int64
	BlockingResources, BlockingUnits        []string
	Failure                                 *failure
}
type engineMessage struct {
	ID, Event, Activity, WaitReason string
	At                              time.Duration
	Extra                           map[string]any
	Result                          *unitResult
	Cleanup                         *cleanupRecord
	Ack                             chan error
}
type waitDetails struct {
	reason           string
	resources, units []string
}

func (w waitDetails) fields() map[string]any {
	return map[string]any{"wait_reason": w.reason, "blocking_resources": w.resources, "blocking_unit_ids": w.units}
}
func stoppedWait() waitDetails { return waitDetails{"scheduler_stop", []string{}, []string{}} }

// graphEngine is the sole writer of scheduling state and canonical event order.
// Workers own only their acquired lifecycle resources and communicate through
// the bounded mailbox. Required publication failure stops admission but still
// drains every worker and finalizer; optional observation never owns outcomes.
type graphEngine struct {
	plan                      *Plan
	ports                     enginePorts
	ctx                       context.Context
	cancel                    context.CancelFunc
	mailbox                   chan engineMessage
	byID                      map[string]unit
	ranks, eligibleAt, claims map[string]int64
	shared                    map[string]int
	exclusive                 map[string]bool
	running, eligible         map[string]bool
	waits                     map[string]waitDetails
	activities                map[string]string
	activityWait              map[string]string
	starts                    map[string]*int64
	result                    graphResult
	seq, lastMS               int64
	phaseSequence             atomic.Int64
	phase                     string
	err                       error
	stopAdmission             bool
	deferredCache             []unit
}

func executeGraph(ctx context.Context, plan *Plan, ports enginePorts) (graphResult, error) {
	if plan == nil || ports.Execute == nil {
		return graphResult{}, errPlan
	}
	if ports.Now == nil {
		epoch := time.Now()
		ports.Now = func() time.Duration { return time.Since(epoch) }
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	e := &graphEngine{plan: plan, ports: ports, ctx: ctx, cancel: cancel, mailbox: make(chan engineMessage, 256),
		byID: map[string]unit{}, ranks: map[string]int64{}, eligibleAt: map[string]int64{}, claims: map[string]int64{},
		shared: map[string]int{}, exclusive: map[string]bool{}, running: map[string]bool{}, eligible: map[string]bool{},
		waits: map[string]waitDetails{}, activities: map[string]string{}, activityWait: map[string]string{}, starts: map[string]*int64{}, phase: "executing",
		result: graphResult{States: map[string]string{}, Units: map[string]unitResult{}, Admissions: []string{}, Events: []map[string]any{}, Cleanup: []cleanupRecord{}}}
	for _, u := range plan.facts.Graph.Units {
		e.byID[u.ID] = u
		e.result.States[u.ID] = "pending"
	}
	e.computeRanks()
	e.emit(0, "run_started", "harness:run", "running", nil)
	for _, u := range plan.facts.Graph.Units {
		e.emit(e.now(), "queued", u.ID, "pending", nil)
	}
	for {
		e.settleDependencies()
		for ctx.Err() == nil && !e.stopAdmission && e.err == nil {
			ready := e.ready()
			barriers := readyExclusiveWaiters(ready)
			for _, u := range ready {
				e.openWait(u.ID, e.blocking(u, barriers), e.now())
			}
			chosen := e.choose(ready, barriers)
			if chosen == nil {
				break
			}

			e.admit(*chosen)
			e.settleDependencies()
		}
		if ctx.Err() != nil || e.stopAdmission {
			e.stopPending()
		}
		e.observe()
		if len(e.running) == 0 {
			pending := false
			for _, state := range e.result.States {
				pending = pending || state == "pending"
			}
			if !pending {
				break
			}
			e.requiredError(failed("harness", "scheduler_accounting_error"))
			continue
		}
		select {
		case message := <-e.mailbox:
			batch := []engineMessage{message}
			for len(batch) < cap(e.mailbox) {
				select {
				case m := <-e.mailbox:
					batch = append(batch, m)
				default:
					goto collected
				}
			}
		collected:
			// Lifecycle messages precede their own terminal message. Completions
			// already available at a boundary use terminal time, then stable ID.
			terminal := []engineMessage{}
			for _, m := range batch {
				if m.Result != nil {
					terminal = append(terminal, m)
				} else {
					e.receive(m)
				}
			}
			slices.SortStableFunc(terminal, func(a, b engineMessage) int {
				if a.At < b.At {
					return -1
				}
				if a.At > b.At {
					return 1
				}
				return strings.Compare(a.ID, b.ID)
			})
			for _, m := range terminal {
				e.complete(m)
			}
		case <-ctx.Done():
			e.stopPending()
			// Cancellation stays level-triggered; consume a worker message so
			// it cannot starve lifecycle acknowledgements during draining.
			e.receive(<-e.mailbox)
		}
	}
	e.finish()
	return e.result, e.err
}
func (e *graphEngine) now() int64 { return max(0, e.ports.Now().Milliseconds()) }
func (e *graphEngine) requiredError(err error) {
	if err != nil && e.err == nil {
		e.err = err
		e.cancel()
	}
}
func (e *graphEngine) emit(at int64, event, id, status string, extra map[string]any) {
	u := e.byID[id]
	needs := u.Needs
	if needs == nil {
		needs = []string{}
	}
	services := u.Services
	if services == nil {
		services = []string{}
	}
	claims := u.Claims
	if claims == nil {
		claims = map[string]int64{}
	}
	e.seq++
	e.lastMS = max(e.lastMS, at)
	r := map[string]any{"schema_id": "cartulary.harness_unit_event.v4", "seq": e.seq, "monotonic_ms": e.lastMS, "event": event, "unit_id": id, "status": status, "needs": needs, "resource_claims": claims, "service_dependencies": services}
	if event == "queued" {
		r["unit_kind"] = u.Kind
		r["row_ids"] = rowIDs(u)
	}
	for k, v := range extra {
		r[k] = v
	}
	data, err := json.Marshal(r)
	if err == nil {
		var normalized any
		normalized, err = harnessschema.Decode(data)
		if err == nil {
			err = harnessschema.Validate("cartulary.harness_unit_event.v4", normalized)
		}
	}
	if err != nil {
		e.requiredError(err)
		return
	}
	if e.ports.RetainEvents {
		e.result.Events = append(e.result.Events, r)
	}
	if e.ports.Event != nil && e.err == nil {
		e.requiredError(e.ports.Event(r))
	}
}
func rowIDs(u unit) []string {
	ids := []string{}
	for _, p := range u.Evidence {
		if strings.HasPrefix(p, "rows/") && strings.HasSuffix(p, ".json") {
			ids = append(ids, strings.TrimSuffix(strings.TrimPrefix(p, "rows/"), ".json"))
		}
	}
	slices.Sort(ids)
	return ids
}
func (e *graphEngine) computeRanks() {
	successors := map[string][]string{}
	for _, u := range e.byID {
		for _, n := range u.Needs {
			successors[n] = append(successors[n], u.ID)
		}
	}
	var rank func(string) int64
	rank = func(id string) int64 {
		if r, ok := e.ranks[id]; ok {
			return r
		}
		var downstream int64
		for _, n := range successors[id] {
			downstream = max(downstream, rank(n))
		}
		r := e.byID[id].EstimateMS + downstream
		e.ranks[id] = r
		return r
	}
	for _, u := range e.plan.facts.Graph.Units {
		rank(u.ID)
	}
}
func terminalState(s string) bool { return s != "pending" && s != "running" }
func (e *graphEngine) ready() []unit {
	ready := []unit{}
	for _, u := range e.plan.facts.Graph.Units {
		if e.result.States[u.ID] != "pending" {
			continue
		}
		ok := true
		for _, n := range u.Needs {
			s := e.result.States[n]
			if s != "passed" && !(terminalState(s) && !e.byID[n].FailurePolicy.BlockDescendants) {
				ok = false
				break
			}
		}
		if ok {
			ready = append(ready, u)
		}
	}
	return ready
}

// The ready set is immutable during one scheduling decision or snapshot. Index
// its exclusive waiters once, rather than rescanning every ready peer for every
// pending unit. Rebuild after each reservation transition to retain exact facts.
func readyExclusiveWaiters(ready []unit) map[string][]string {
	waiters := map[string][]string{}
	for _, u := range ready {
		for _, lock := range u.ExclusiveLocks {
			waiters[lock] = append(waiters[lock], u.ID)
		}
	}
	return waiters
}
func (e *graphEngine) blocking(u unit, barriers map[string][]string) waitDetails {
	w := waitDetails{"rank", []string{}, []string{}}
	locks := []string{}
	for r, n := range u.Claims {
		requested := e.claims[r] + n
		if r == "volume" {
			requested = max(e.claims[r], n)
		}
		if requested > e.plan.facts.Capacities[r] {
			w.resources = append(w.resources, r)
		}
	}
	for _, l := range u.SharedLocks {
		if e.exclusive[l] {
			locks = append(locks, l)
		}
	}
	for _, l := range u.ExclusiveLocks {
		if e.exclusive[l] || e.shared[l] > 0 {
			locks = append(locks, l)
		}
	}
	for id := range e.running {
		a := e.byID[id]
		blocked := false
		for _, r := range w.resources {
			blocked = blocked || a.Claims[r] > 0
		}
		for _, l := range locks {
			blocked = blocked || slices.Contains(a.SharedLocks, l) || slices.Contains(a.ExclusiveLocks, l)
		}
		if blocked {
			w.units = append(w.units, id)
		}
	}
	for _, l := range locks {
		w.resources = append(w.resources, "lock_"+lockToken(l))
	}
	slices.Sort(w.resources)
	slices.Sort(w.units)
	if len(w.resources) > 0 {
		w.reason = "resources"
		return w
	}
	for _, l := range u.SharedLocks {
		for _, other := range barriers[l] {
			if other != u.ID {
				w.resources = append(w.resources, "lock_"+lockToken(l))
				w.units = append(w.units, other)
			}
		}
	}
	slices.Sort(w.resources)
	slices.Sort(w.units)
	w.resources = slices.Compact(w.resources)
	w.units = slices.Compact(w.units)
	if len(w.units) > 0 {
		w.reason = "exclusive_barrier"
	}
	return w
}
func lockToken(s string) string {
	var b strings.Builder
	separator := false
	for _, c := range s {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' {
			b.WriteRune(c)
			separator = false
		} else if !separator {
			b.WriteByte('_')
			separator = true
		}
	}
	return b.String()
}
func (e *graphEngine) choose(ready []unit, barriers map[string][]string) *unit {
	ordered := slices.Clone(ready)
	now := e.now()
	quantum := max(1, e.plan.facts.AgingMS)
	slices.SortFunc(ordered, func(a, b unit) int {
		ar := e.ranks[a.ID] + max(0, now-e.eligibleAt[a.ID])/quantum*quantum
		br := e.ranks[b.ID] + max(0, now-e.eligibleAt[b.ID])/quantum*quantum
		if ar > br {
			return -1
		}
		if ar < br {
			return 1
		}
		return strings.Compare(a.ID, b.ID)
	})
	for _, u := range ordered {
		if e.blocking(u, barriers).reason == "rank" {
			return &u
		}
	}
	return nil
}
func (e *graphEngine) openWait(id string, w waitDetails, at int64) {
	if prior, ok := e.waits[id]; ok {
		if prior.reason == w.reason && slices.Equal(prior.resources, w.resources) && slices.Equal(prior.units, w.units) {
			return
		}
		e.closeWait(id, at)
	}
	if !e.eligible[id] {
		e.eligible[id] = true
		e.eligibleAt[id] = at
		e.emit(at, "eligible", id, "pending", nil)
	}
	e.waits[id] = w
	e.emit(at, "wait_started", id, "pending", w.fields())
}

// Refresh after each reservation change, including batched completions. The
// state transition and both segment boundaries share one canonical instant.
func (e *graphEngine) refreshWaits(at int64) {
	ready := e.ready()
	barriers := readyExclusiveWaiters(ready)
	for _, u := range ready {
		if e.eligible[u.ID] {
			e.openWait(u.ID, e.blocking(u, barriers), at)
		}
	}
}
func (e *graphEngine) closeWait(id string, at int64) {
	w, ok := e.waits[id]
	if !ok {
		e.requiredError(failed("harness", "scheduler_accounting_error"))
		return
	}
	delete(e.waits, id)
	e.emit(at, "wait_ended", id, "pending", w.fields())
}
func (e *graphEngine) reserve(u unit, direction int64) {
	for k, v := range u.Claims {
		if k == "volume" {
			e.claims[k] = 0
			for id := range e.running {
				e.claims[k] = max(e.claims[k], e.byID[id].Claims[k])
			}
			continue
		}
		e.claims[k] += direction * v
	}
	for _, l := range u.SharedLocks {
		e.shared[l] += int(direction)
	}
	for _, l := range u.ExclusiveLocks {
		if direction > 0 {
			e.exclusive[l] = true
		} else {
			delete(e.exclusive, l)
		}
	}
}
func (e *graphEngine) directPhase(u unit, name string, operation func() error) error {
	details := map[string]any{"interval_id": fmt.Sprintf("phase:%d", e.phaseSequence.Add(1)), "phase": name}
	e.emit(e.now(), "phase_started", u.ID, e.result.States[u.ID], details)
	err := operation()
	outcome := "completed"
	if err != nil {
		outcome = "failed"
		if e.ctx.Err() != nil {
			outcome = "cancelled"
		}
	}
	details["phase_outcome"] = outcome
	e.emit(e.now(), "phase_finished", u.ID, e.result.States[u.ID], details)
	return err
}
func (e *graphEngine) admit(u unit) {
	c := cacheResult{Outcome: "bypass", Reason: "cache_unconfigured"}
	if e.ports.Cache != nil {
		if err := e.directPhase(u, "cache_lookup", func() error { var err error; c, err = e.ports.Cache.Lookup(e.ctx, u); return err }); err != nil {
			// The unit is still pending. Let ordinary cancellation close its wait
			// and publish its terminal event instead of aborting required evidence.
			if e.ctx.Err() != nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
				return
			}
			e.requiredError(err)
			return
		}
	}
	extra := map[string]any{"cache_reason": c.Reason}
	if c.Profile != "" {
		extra["cache_profile_id"] = c.Profile
	}
	if c.OutputDigest != "" {
		extra["output_digest"] = c.OutputDigest
	}
	status := "pending"
	if c.Outcome == "hit" {
		status = "passed"
	}
	e.emit(e.now(), "cache_"+c.Outcome, u.ID, status, extra)
	at := e.now()
	e.closeWait(u.ID, at)
	if e.err != nil {
		return
	}
	if c.Outcome == "hit" {
		e.result.States[u.ID] = "passed"
		e.result.Units[u.ID] = unitResult{Status: "passed", Cache: c}
		e.emit(at, "completed", u.ID, "passed", map[string]any{"output_digest": c.OutputDigest})
		return
	}
	e.result.States[u.ID] = "running"
	e.result.Units[u.ID] = unitResult{Cache: c}
	e.activities[u.ID] = "admitted"
	e.running[u.ID] = true
	e.reserve(u, 1)
	e.result.Admissions = append(e.result.Admissions, u.ID)
	e.emit(at, "admitted", u.ID, "running", nil)
	e.emit(at, "started", u.ID, "running", nil)
	e.refreshWaits(at)
	go e.work(u, c)
}
func (e *graphEngine) settleDependencies() {
	for changed := true; changed; {
		changed = false
		for _, u := range e.plan.facts.Graph.Units {
			if e.result.States[u.ID] != "pending" {
				continue
			}
			blocked := false
			for _, n := range u.Needs {
				s := e.result.States[n]
				blocked = blocked || terminalState(s) && s != "passed" && e.byID[n].FailurePolicy.BlockDescendants
			}
			if !blocked {
				continue
			}
			at := e.now()
			e.openWait(u.ID, stoppedWait(), at)
			e.closeWait(u.ID, at)
			e.result.States[u.ID] = "skipped"
			e.result.Units[u.ID] = unitResult{Status: "skipped"}
			e.emit(at, "skipped", u.ID, "skipped", map[string]any{"failure_reason": "dependency_failure"})
			changed = true
		}
	}
}
func (e *graphEngine) stopPending() {
	e.phase = "draining"
	for _, u := range e.plan.facts.Graph.Units {
		if e.result.States[u.ID] != "pending" {
			continue
		}
		at := e.now()
		e.openWait(u.ID, stoppedWait(), at)
		e.closeWait(u.ID, at)
		f := failed("interrupted", "cancelled_or_interrupted")
		e.result.States[u.ID] = "cancelled"
		e.result.Units[u.ID] = unitResult{Status: "cancelled", Failure: f}
		e.emit(at, "cancelled", u.ID, "cancelled", failureFields(f))
	}
}
func failureFields(f *failure) map[string]any {
	if f == nil {
		return nil
	}
	return map[string]any{"failure_class": f.Class, "failure_reason": f.Reason}
}
func (e *graphEngine) receive(m engineMessage) {
	if m.Result != nil {
		e.complete(m)
		return
	}
	if m.Cleanup != nil {
		e.recordCleanup(*m.Cleanup)
	}
	if m.Activity != "" {
		e.activities[m.ID] = m.Activity
		e.activityWait[m.ID] = m.WaitReason
		if m.Activity == "executing" {
			ms := m.At.Milliseconds()
			e.starts[m.ID] = &ms
		}
	}
	if m.Event != "" {
		e.emit(m.At.Milliseconds(), m.Event, m.ID, e.result.States[m.ID], m.Extra)
	}
	if m.Ack != nil {
		m.Ack <- e.err
	}
	e.observe()
}
func (e *graphEngine) complete(m engineMessage) {
	if !e.running[m.ID] {
		e.requiredError(failed("harness", "scheduler_accounting_error"))
		return
	}
	u := e.byID[m.ID]
	delete(e.running, m.ID)
	e.reserve(u, -1)
	r := *m.Result
	e.result.States[m.ID] = r.Status
	e.result.Units[m.ID] = r
	if r.RetainedFixture {
		e.deferredCache = append(e.deferredCache, u)
	}
	event := "completed"
	if r.Status != "passed" {
		event = r.Status
		e.phase = "draining"
		if !u.FailurePolicy.ContinueIndependent {
			e.stopAdmission = true
		}
	}
	e.emit(m.At.Milliseconds(), event, m.ID, r.Status, failureFields(r.Failure))
	e.refreshWaits(e.lastMS)
	e.observe()
}
func (e *graphEngine) observe() {
	if e.ports.Observe == nil {
		return
	}
	s := engineSnapshot{Phase: e.phase, ElapsedMS: e.now(), Units: []unitSnapshot{}}
	barriers := readyExclusiveWaiters(e.ready())
	for _, u := range e.plan.facts.Graph.Units {
		status := e.result.States[u.ID]
		r := e.result.Units[u.ID]
		v := unitSnapshot{ID: u.ID, Status: status, Activity: e.activities[u.ID], WaitReason: e.activityWait[u.ID], Cache: r.Cache.Outcome, StartedMS: e.starts[u.ID], Failure: r.Failure, BlockingResources: []string{}, BlockingUnits: []string{}}
		if v.Failure != nil {
			f := *v.Failure
			v.Failure = &f
		}
		if v.StartedMS != nil {
			ms := *v.StartedMS
			v.StartedMS = &ms
		}
		if terminalState(status) {
			v.Activity = "terminal"
			v.WaitReason = ""
		} else if status == "pending" {
			for _, n := range u.Needs {
				if !terminalState(e.result.States[n]) {
					v.BlockingUnits = append(v.BlockingUnits, n)
				}
			}
			if len(v.BlockingUnits) > 0 {
				v.WaitReason = "dependencies"
			} else if _, ok := e.waits[u.ID]; ok {
				w := e.blocking(u, barriers)
				v.WaitReason = w.reason
				v.BlockingResources = w.resources
				v.BlockingUnits = w.units
			}
			if v.WaitReason != "" {
				v.Activity = "waiting"
			} else {
				v.Activity = "pending"
			}
		}
		s.Units = append(s.Units, v)
	}
	// Observation adapters return no required result. Their own availability and
	// bounds are reported separately from canonical execution.
	if !optionalObservation(func() { e.ports.Observe(s) }) {
		e.result.ObservationFailed = true
		e.ports.Observe = nil
	}
}
func (e *graphEngine) finish() {
	e.phase = "finalizing"
	e.observe()
	e.emit(e.now(), "cleanup_started", "harness:run", "running", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	unresolved := false
	finalizationFailures := []*failure{}
	attempt := func(operation string, fn func() error) {
		err := fn()
		r := cleanupRecord{Operation: operation, Outcome: "completed", Cause: err}
		if err != nil {
			r.Outcome = "failed"
			r.Failure = cleanupFailureOf(err)
		}
		e.recordCleanup(r)
	}
	if e.ports.Fixtures != nil {
		attempt("provider_close", func() error { return e.ports.Fixtures.Close(ctx) })
		unresolved = e.ports.Fixtures.Unresolved()
	}
	if e.ports.CleanupServices != nil {
		if unresolved {
			e.recordCleanup(cleanupRecord{Operation: "services_close", Outcome: "blocked"})
		} else {
			attempt("services_close", func() error { return e.ports.CleanupServices(ctx) })
		}
	}
	cleanupFailed := unresolved
	for _, r := range e.result.Cleanup {
		cleanupFailed = cleanupFailed || r.Outcome != "completed"
	}
	if !cleanupFailed && e.err == nil && e.ports.Cache != nil {
		for _, u := range e.deferredCache {
			if err := e.ports.Cache.Store(ctx, u); err != nil {
				f := failed("artifact", "artifact_error")
				f.Lifecycle = "cleanup_finalizers"
				finalizationFailures = append(finalizationFailures, f)
			}
		}
	}
	if e.ports.Finalize != nil {
		attempt("runtime_close", func() error { return e.ports.Finalize(ctx, unresolved, e.result.Cleanup) })
	}
	failures := []*failure{}
	for _, u := range e.plan.facts.Graph.Units {
		r := e.result.Units[u.ID]
		if r.Failure != nil {
			failures = append(failures, r.Failure)
		}
	}
	cleanupFailures := finalizationFailures
	for _, r := range e.result.Cleanup {
		if r.Failure != nil {
			cleanupFailures = append(cleanupFailures, r.Failure)
		}
	}
	cleanupPrimary := primaryFailure(cleanupFailures)
	status := "passed"
	if cleanupPrimary != nil || unresolved {
		status = "failed"
	}
	e.emit(e.now(), "cleanup_completed", "harness:run", status, failureFields(cleanupPrimary))
	failures = append(failures, cleanupFailures...)
	if e.err != nil {
		failures = append(failures, failureOf(e.err))
	}
	if unresolved && cleanupPrimary == nil {
		failures = append(failures, failed("harness", "cleanup_error"))
	}
	e.result.Failure = primaryFailure(failures)
	e.result.Status = "passed"
	if e.result.Failure != nil {
		e.result.Status = "failed"
	}
	status = e.result.Status
	for _, s := range e.result.States {
		if s == "cancelled" {
			status = "cancelled"
			break
		}
	}
	e.emit(e.now(), "run_completed", "harness:run", status, failureFields(e.result.Failure))
	e.result.DurationMS = e.lastMS
}

func fixtureAffinity(u unit) string {
	if u.Affinity != "" {
		return u.Affinity
	}
	return u.Owner
}

// Recover only at an explicitly optional observation boundary; required owners
// and lifecycle failures retain their ordinary failure paths.
func optionalObservation(call func()) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	call()
	return true
}

func (e *graphEngine) recordCleanup(r cleanupRecord) {
	e.result.Cleanup = append(e.result.Cleanup, r)
	if e.ports.Cleanup != nil {
		e.ports.Cleanup(r)
	}
}
