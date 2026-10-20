package harnessruntime

import (
	"cmp"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
)

var errCanonical = errors.New("canonical graph accounting does not close")

type timeInterval struct{ start, end, queue int64 }
type canonicalWait struct {
	timeInterval
	details             waitDetails
	closed              bool
	segmentStart, count int64
	resources           []timeInterval
	holders             map[string]bool
	byReason            map[string]int64
}
type canonicalTerminal struct {
	event, status string
	at, seq       int64
	failure       *failure
}
type canonicalAdmission struct{ at, seq int64 }
type canonicalPhase struct {
	unit, phase, outcome string
	start, end           int64
	closed               bool
}
type canonicalFacts struct {
	plan                         *Plan
	byID                         map[string]unit
	registered, eligible, hits   map[string]bool
	starts, fixtures, eligibleAt map[string]int64
	waits                        map[string]canonicalWait
	admissions                   map[string]canonicalAdmission
	terminals                    map[string]canonicalTerminal
	phases                       map[string]canonicalPhase
	phaseCounts                  map[string]int
	cache                        map[string]int64
	seq, finalMS                 int64
	started, completed           bool
	cleanupStarted               *int64
}

func newCanonicalFacts(p *Plan) *canonicalFacts {
	f := &canonicalFacts{plan: p, byID: map[string]unit{}, registered: map[string]bool{}, eligible: map[string]bool{}, hits: map[string]bool{}, starts: map[string]int64{}, eligibleAt: map[string]int64{}, fixtures: map[string]int64{}, waits: map[string]canonicalWait{}, admissions: map[string]canonicalAdmission{}, terminals: map[string]canonicalTerminal{}, phases: map[string]canonicalPhase{}, phaseCounts: map[string]int{}, cache: map[string]int64{"hit": 0, "miss": 0, "bypass": 0}}
	for _, u := range p.facts.Graph.Units {
		f.byID[u.ID] = u
	}
	return f
}
func integerValue(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	}
	return -1
}
func stringList(v any) []string {
	if values, ok := v.([]string); ok {
		return slices.Clone(values)
	}
	out := []string{}
	if values, ok := v.([]any); ok {
		for _, value := range values {
			if s, ok := value.(string); ok {
				out = append(out, s)
			}
		}
	}
	return out
}
func (f *canonicalFacts) consume(e map[string]any) error {
	if _, err := documentBytes("cartulary.harness_unit_event.v4", e); err != nil {
		return err
	}
	seq, at := integerValue(e["seq"]), integerValue(e["monotonic_ms"])
	event, id, status := e["event"].(string), e["unit_id"].(string), e["status"].(string)
	if f.completed || seq != f.seq+1 || at < f.finalMS {
		return errCanonical
	}
	f.seq = seq
	f.finalMS = at
	if event == "run_started" {
		if f.started || seq != 1 || at != 0 || id != "harness:run" {
			return errCanonical
		}
		f.started = true
		return nil
	}
	if !f.started {
		return errCanonical
	}
	if id == "harness:run" {
		switch event {
		case "cleanup_started":
			if f.cleanupStarted != nil {
				return errCanonical
			}
			value := at
			f.cleanupStarted = &value
		case "cleanup_completed":
			if f.cleanupStarted == nil {
				return errCanonical
			}
		case "run_completed":
			f.completed = true
		default:
			return errCanonical
		}
		return nil
	}
	u, ok := f.byID[id]
	if !ok {
		return errCanonical
	}
	if event == "queued" {
		if f.registered[id] || e["unit_kind"] != u.Kind || !slices.Equal(stringList(e["needs"]), u.Needs) || !slices.Equal(stringList(e["row_ids"]), rowIDs(u)) {
			return errCanonical
		}
		bytes, _ := json.Marshal(e["resource_claims"])
		var claims map[string]int64
		if json.Unmarshal(bytes, &claims) != nil || !reflect.DeepEqual(claims, u.Claims) {
			return errCanonical
		}
		f.registered[id] = true
		return nil
	}
	if !f.registered[id] {
		return errCanonical
	}
	if _, done := f.terminals[id]; done {
		return errCanonical
	}
	switch event {
	case "eligible":
		if f.eligible[id] {
			return errCanonical
		}
		f.eligible[id] = true
		f.eligibleAt[id] = at
	case "wait_started":
		w, exists := f.waits[id]
		_, admitted := f.admissions[id]
		if !f.eligible[id] || admitted || exists && (!w.closed || w.end != at) || !exists && f.eligibleAt[id] != at {
			return errCanonical
		}
		details := waitDetails{e["wait_reason"].(string), stringList(e["blocking_resources"]), stringList(e["blocking_unit_ids"])}
		if !slices.IsSorted(details.resources) || !slices.IsSorted(details.units) {
			return errCanonical
		}
		blocked := details.reason == "resources" || details.reason == "exclusive_barrier"
		if blocked != (len(details.resources) > 0) || blocked != (len(details.units) > 0) {
			return errCanonical
		}
		for _, holder := range details.units {
			if _, ok := f.byID[holder]; !ok || holder == id {
				return errCanonical
			}
		}
		if exists && reflect.DeepEqual(details, w.details) {
			return errCanonical
		}
		if !exists {
			w = canonicalWait{timeInterval: timeInterval{start: at}, holders: map[string]bool{}, byReason: map[string]int64{"resources": 0, "exclusive_barrier": 0, "rank": 0, "scheduler_stop": 0}}
		}
		w.segmentStart, w.closed, w.details = at, false, details
		w.count++
		f.waits[id] = w
	case "wait_ended":
		w, ok := f.waits[id]
		if !ok || w.closed || w.details.reason != e["wait_reason"] || !slices.Equal(w.details.resources, stringList(e["blocking_resources"])) || !slices.Equal(w.details.units, stringList(e["blocking_unit_ids"])) {
			return errCanonical
		}
		w.end, w.closed = at, true
		w.byReason[w.details.reason] += at - w.segmentStart
		if w.details.reason == "resources" {
			for _, holder := range w.details.units {
				w.holders[holder] = true
			}
			if at > w.segmentStart {
				n := len(w.resources)
				if n > 0 && w.resources[n-1].end == w.segmentStart {
					w.resources[n-1].end = at
				} else {
					w.resources = append(w.resources, timeInterval{start: w.segmentStart, end: at})
				}
			}
		}
		f.waits[id] = w
	case "admitted":
		w := f.waits[id]
		if _, ok := f.admissions[id]; ok || !w.closed || w.end != at {
			return errCanonical
		}
		f.admissions[id] = canonicalAdmission{at, seq}
	case "started":
		if _, ok := f.admissions[id]; !ok {
			return errCanonical
		}
		if _, ok := f.starts[id]; ok {
			return errCanonical
		}
		f.starts[id] = at
	case "cache_hit":
		if f.hits[id] {
			return errCanonical
		}
		f.hits[id] = true
		f.cache["hit"]++
	case "cache_miss":
		f.cache["miss"]++
	case "cache_bypass":
		f.cache["bypass"]++
	case "fixture_acquired":
		f.fixtures[id] = at
	case "fixture_released":
	case "phase_started":
		phase := e["interval_id"].(string)
		if _, ok := f.phases[phase]; ok || f.phaseCounts[id] >= maximumPhasesPerUnit {
			return errCanonical
		}
		f.phaseCounts[id]++
		f.phases[phase] = canonicalPhase{unit: id, phase: e["phase"].(string), start: at}
	case "phase_finished":
		phase := e["interval_id"].(string)
		p, ok := f.phases[phase]
		if !ok || p.closed || p.unit != id || p.phase != e["phase"] {
			return errCanonical
		}
		p.end = at
		p.closed = true
		p.outcome = e["phase_outcome"].(string)
		f.phases[phase] = p
	case "completed", "failed", "skipped", "cancelled":
		want := map[string]string{"completed": "passed", "failed": "failed", "skipped": "skipped", "cancelled": "cancelled"}[event]
		if status != want {
			return errCanonical
		}
		w := f.waits[id]
		if !w.closed || !f.eligible[id] {
			return errCanonical
		}
		_, started := f.starts[id]
		if !started && at != w.end {
			return errCanonical
		}
		if !started && !f.hits[id] && event != "skipped" && event != "cancelled" {
			return errCanonical
		}
		var failure *failure
		if class, ok := e["failure_class"].(string); ok {
			reason, _ := e["failure_reason"].(string)
			failure = failed(class, reason)
			failure.Sequence = seq
		}
		f.terminals[id] = canonicalTerminal{event, status, at, seq, failure}
	default:
		return errCanonical
	}
	return nil
}
func (f *canonicalFacts) validate() error {
	if !f.started || !f.completed || f.cleanupStarted == nil || len(f.terminals) != len(f.byID) || len(f.registered) != len(f.byID) {
		return errCanonical
	}
	for _, w := range f.waits {
		if !w.closed {
			return errCanonical
		}
	}
	for _, p := range f.phases {
		if !p.closed {
			return errCanonical
		}
	}
	return nil
}
func (f *canonicalFacts) interval(id string) timeInterval {
	terminal := f.terminals[id]
	start, ok := f.starts[id]
	if !ok {
		start = terminal.at
	}
	return timeInterval{start: start, end: terminal.at, queue: start - f.waits[id].start}
}
func intervalUnion(values []timeInterval) int64 {
	ordered := slices.Clone(values)
	slices.SortFunc(ordered, func(a, b timeInterval) int { return cmp.Or(cmp.Compare(a.start, b.start), cmp.Compare(a.end, b.end)) })
	var total, start, end int64
	open := false
	for _, i := range ordered {
		if i.end < i.start {
			continue
		}
		if !open || i.start > end {
			if open {
				total += end - start
			}
			start, end, open = i.start, i.end, true
		} else {
			end = max(end, i.end)
		}
	}
	if open {
		total += end - start
	}
	return total
}
func (f *canonicalFacts) criticalPath(selected []string) ([]string, int64) {
	allowed := map[string]bool{}
	for _, id := range selected {
		allowed[id] = true
	}
	type path struct {
		ids      []string
		duration int64
	}
	memo := map[string]path{}
	less := func(a, b path) bool {
		return a.duration > b.duration || a.duration == b.duration && strings.Join(a.ids, "\x00") < strings.Join(b.ids, "\x00")
	}
	var to func(string) path
	to = func(id string) path {
		if v, ok := memo[id]; ok {
			return v
		}
		prior := path{ids: []string{}}
		for _, dependency := range f.byID[id].Needs {
			if !allowed[dependency] {
				continue
			}
			candidate := to(dependency)
			if less(candidate, prior) {
				prior = candidate
			}
		}
		i := f.interval(id)
		p := path{append(slices.Clone(prior.ids), id), prior.duration + i.end - i.start + i.queue}
		memo[id] = p
		return p
	}
	best := path{ids: []string{}}
	for _, id := range selected {
		candidate := to(id)
		if len(best.ids) == 0 || less(candidate, best) {
			best = candidate
		}
	}
	return best.ids, best.duration
}
func (f *canonicalFacts) timing(selected []string, duration int64, envelope bool) map[string]int64 {
	type bucketInterval struct {
		timeInterval
		bucket string
	}
	intervals := []bucketInterval{}
	waits := []timeInterval{}
	first := duration
	count := int64(0)
	for _, id := range selected {
		if w, ok := f.waits[id]; ok {
			waits = append(waits, w.resources...)
		}
		start, ok := f.starts[id]
		if !ok {
			continue
		}
		first = min(first, start)
		count++
		terminal := f.terminals[id]
		if terminal.event != "completed" && terminal.event != "failed" && terminal.event != "cancelled" {
			continue
		}
		execution := start
		if acquired, ok := f.fixtures[id]; ok {
			if acquired > start {
				intervals = append(intervals, bucketInterval{timeInterval{start: start, end: acquired}, "fixture"})
			}
			execution = acquired
		}
		bucket := "execution"
		switch f.byID[id].Kind {
		case "finalizer", "projection":
			bucket = "collation"
		case "artifact", "readiness":
			bucket = "setup"
		}
		if terminal.at > execution {
			intervals = append(intervals, bucketInterval{timeInterval{start: execution, end: terminal.at}, bucket})
		}
	}
	if envelope && count > 0 && first > 0 {
		intervals = append(intervals, bucketInterval{timeInterval{start: 0, end: first}, "setup"})
	}
	if envelope && f.cleanupStarted != nil && duration > *f.cleanupStarted {
		intervals = append(intervals, bucketInterval{timeInterval{start: *f.cleanupStarted, end: duration}, "wrapper"})
	}
	boundaries := []int64{}
	if envelope {
		boundaries = append(boundaries, 0, duration)
	}
	for _, i := range intervals {
		boundaries = append(boundaries, i.start, i.end)
	}
	slices.Sort(boundaries)
	boundaries = slices.Compact(boundaries)
	totals := map[string]int64{"setup_ms": 0, "fixture_ms": 0, "execution_ms": 0, "collation_ms": 0, "wrapper_ms": 0, "unattributed_ms": 0, "resource_blocking_ms": intervalUnion(waits), "process_count": count}
	precedence := []string{"collation", "fixture", "execution", "setup", "wrapper"}
	for n := 1; n < len(boundaries); n++ {
		start, end := boundaries[n-1], boundaries[n]
		active := map[string]bool{}
		for _, i := range intervals {
			if i.start < end && i.end > start {
				active[i.bucket] = true
			}
		}
		bucket := ""
		for _, b := range precedence {
			if active[b] {
				bucket = b
				break
			}
		}
		if bucket != "" {
			totals[bucket+"_ms"] += end - start
		} else if envelope {
			totals["unattributed_ms"] += end - start
		}
	}
	return totals
}
func (f *canonicalFacts) pressure() map[string]any {
	type boundary struct {
		id                 string
		at, seq, direction int64
	}
	boundaries := []boundary{}
	for id, a := range f.admissions {
		boundaries = append(boundaries, boundary{id, a.at, a.seq, 1})
		terminal := f.terminals[id]
		if terminal.event == "completed" || terminal.event == "failed" || terminal.event == "cancelled" {
			boundaries = append(boundaries, boundary{id, terminal.at, terminal.seq, -1})
		}
	}
	slices.SortFunc(boundaries, func(a, b boundary) int { return cmp.Compare(a.seq, b.seq) })
	active, peaks, saturated, starts := map[string]int64{}, map[string]int64{}, map[string]int64{}, map[string]int64{}
	volumes := map[string]int64{}
	for _, b := range boundaries {
		for resource, amount := range f.byID[b.id].Claims {
			before := active[resource]
			after := before + b.direction*amount
			if resource == "volume" {
				if b.direction > 0 {
					volumes[b.id] = amount
				} else {
					delete(volumes, b.id)
				}
				after = 0
				for _, claim := range volumes {
					after = max(after, claim)
				}
			}
			active[resource] = after
			peaks[resource] = max(peaks[resource], after)
			capacity := f.plan.facts.Capacities[resource]
			if before < capacity && after >= capacity {
				starts[resource] = b.at
			}
			if before >= capacity && after < capacity {
				if start, ok := starts[resource]; ok {
					saturated[resource] += b.at - start
					delete(starts, resource)
				}
			}
		}
	}
	for r, start := range starts {
		saturated[r] += f.finalMS - start
	}
	requested := map[string]int64{}
	for r := range f.plan.facts.Capacities {
		for _, u := range f.byID {
			requested[r] = max(requested[r], u.Claims[r])
		}
	}
	blocked, holders := []string{}, []string{}
	var count int64
	for id, w := range f.waits {
		count += w.count
		if len(w.holders) > 0 {
			blocked = append(blocked, id)
		}
		for holder := range w.holders {
			holders = append(holders, holder)
		}
	}
	slices.Sort(blocked)
	slices.Sort(holders)
	holders = slices.Compact(holders)
	return map[string]any{"requested_capacity": requested, "resolved_capacity": f.plan.facts.Capacities, "peak_use": peaks, "saturation_ms": saturated, "wait_events": count, "blocked_units": blocked, "resource_holders": holders}
}
