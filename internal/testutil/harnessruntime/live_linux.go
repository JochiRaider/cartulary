//go:build linux

package harnessruntime

import (
	"maps"
	"slices"
	"sync"
	"time"
)

// Live publication reads coherent engine snapshots. It does not reconstruct
// scheduler state or own required outcomes. One timer supplies dirty updates
// (1s) and idle heartbeats (5s); pause cancels it and drains publication.
type livePublisher struct {
	mu                    sync.Mutex
	root                  string
	identity              map[string]any
	targets               []map[string]any
	view                  engineSnapshot
	epoch, received, last int64
	offset                *int64
	revision, sequence    int64
	timerGeneration       uint64
	paused                int
	dirty, stopped        bool
	timer                 *time.Timer
	resources             func() map[string]any
	publicationErr        error
}

func newLivePublisher(root string, plan *Plan, resources func() map[string]any) *livePublisher {
	p := &livePublisher{root: root, identity: map[string]any{}, targets: []map[string]any{}, epoch: monotonicNS(), resources: resources}
	p.received = p.epoch
	for _, k := range []string{"run_id", "command_id", "source_digest", "graph_digest"} {
		p.identity[k] = plan.facts.Manifest[k]
	}
	p.identity["public_target"] = plan.facts.Manifest["target"]
	for _, target := range slices.Sorted(maps.Keys(plan.facts.Projections)) {
		ids := slices.Clone(plan.facts.Projections[target].Units)
		slices.Sort(ids)
		p.targets = append(p.targets, map[string]any{"target": target, "unit_ids": ids})
	}
	p.view = engineSnapshot{Phase: "executing", Units: []unitSnapshot{}}
	for _, u := range plan.facts.Graph.Units {
		p.view.Units = append(p.view.Units, unitSnapshot{ID: u.ID, Status: "pending", Activity: "pending", Cache: "not_observed", BlockingResources: []string{}, BlockingUnits: []string{}})
	}
	p.publishLocked()
	p.scheduleLocked()
	return p
}
func (p *livePublisher) publishLocked() {
	p.last = monotonicNS()
	p.dirty = false
	v := maps.Clone(p.identity)
	v["schema_id"] = "cartulary.harness_live_snapshot.v3"
	v["revision"] = p.revision + 1
	v["sampled_at"] = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	v["sampled_elapsed_ms"] = p.view.ElapsedMS + max(0, (p.last-p.received)/1e6)
	v["last_transition_seq"] = p.sequence
	v["publication_mode"] = "active"
	if p.paused > 0 {
		v["publication_mode"] = "paused_measurement"
	}
	v["phase"] = p.view.Phase
	v["availability"] = "available"
	v["unavailable_reason"] = nil
	v["targets"] = p.targets
	units := []map[string]any{}
	for _, u := range p.view.Units {
		var start any
		if u.StartedMS != nil {
			start = *u.StartedMS
		}
		var class, reason any
		if u.Failure != nil {
			class = u.Failure.Class
			reason = u.Failure.Reason
		}
		cache := u.Cache
		if cache == "" {
			cache = "not_observed"
		}
		activity := u.Activity
		if activity == "" {
			activity = "pending"
		}
		units = append(units, map[string]any{"unit_id": u.ID, "status": u.Status, "activity": activity, "started_elapsed_ms": start, "wait_reason": nullableString(u.WaitReason), "blocking_resources": append([]string{}, u.BlockingResources...), "blocking_unit_ids": append([]string{}, u.BlockingUnits...), "cache_disposition": cache, "failure_class": class, "failure_reason": reason})
	}
	slices.SortFunc(units, func(a, b map[string]any) int {
		if a["unit_id"].(string) < b["unit_id"].(string) {
			return -1
		}
		if a["unit_id"] == b["unit_id"] {
			return 0
		}
		return 1
	})
	v["units"] = units
	if p.resources != nil {
		if r := p.resources(); r != nil {
			v["resources"] = r
		}
	}
	b, err := documentBytes("cartulary.harness_live_snapshot.v3", v)
	if err == nil && len(b)+1 > 16*1024*1024 {
		delete(v, "targets")
		delete(v, "units")
		v["availability"] = "unavailable"
		v["unavailable_reason"] = "snapshot_too_large"
	}
	if err == nil {
		err = writeEvidence(p.root, "diagnostics/live-status.json", "cartulary.harness_live_snapshot.v3", v, true)
	}
	p.publicationErr = err
	if err == nil {
		p.revision++
	}
}
func (p *livePublisher) scheduleLocked() {
	p.timerGeneration++
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	if p.stopped || p.paused > 0 {
		return
	}
	interval := 5 * time.Second
	if p.dirty {
		interval = time.Second
	}
	delay := max(0, interval-time.Duration(monotonicNS()-p.last))
	generation := p.timerGeneration
	p.timer = time.AfterFunc(delay, func() {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.stopped || p.paused > 0 || generation != p.timerGeneration {
			return
		}
		p.publishLocked()
		p.scheduleLocked()
	})
}
func (p *livePublisher) update(v engineSnapshot) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.received = monotonicNS()
	if p.offset == nil {
		n := max(0, (p.received-p.epoch)/1e6-v.ElapsedMS)
		p.offset = &n
	}
	v.ElapsedMS += *p.offset
	v.Units = slices.Clone(v.Units)
	for i := range v.Units {
		if v.Units[i].StartedMS != nil {
			n := *v.Units[i].StartedMS + *p.offset
			v.Units[i].StartedMS = &n
		}
	}
	p.view = v
	p.dirty = true
	p.scheduleLocked()
}
func (p *livePublisher) transition(seq int64) { p.mu.Lock(); defer p.mu.Unlock(); p.sequence = seq }
func (p *livePublisher) pause() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.paused++
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	if p.paused == 1 {
		p.publishLocked()
	}
}
func (p *livePublisher) resume() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped || p.paused == 0 {
		return
	}
	p.paused--
	if p.paused == 0 {
		p.publishLocked()
		p.scheduleLocked()
	}
}
func (p *livePublisher) stop() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.stopped = true
	if p.timer != nil {
		p.timer.Stop()
		p.timer = nil
	}
	if p.paused == 0 {
		p.publishLocked()
	}
}
