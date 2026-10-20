//go:build linux

package harnesscollect

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"golang.org/x/sys/unix"
)

var errQuiescence = errors.New("observation fence drain failed")

type processEntry struct {
	proof                        proof
	ref                          string
	unit, allocation, invocation any
	attribution                  string
	conflicted, gone             bool
}
type collector struct {
	policy                                        policy
	identity                                      object
	epoch                                         int64
	adapter                                       *linuxAdapter
	store                                         *resourceStore
	containers                                    *containerAdapter
	checkpoints                                   *checkpointEvidence
	admissionRoot                                 string
	capacities                                    harnessadmission.Quantities
	owner                                         harnessadmission.Proof
	processes                                     []*processEntry
	byIdentity                                    map[string]*processEntry
	relationships                                 map[string]bool
	containerIdentities                           map[string]object
	containerOrder                                []any
	containerSignals                              map[string]bool
	containerAvailability                         string
	containerOmissions, overruns                  int
	lastContext, lastContainer                    int64
	omitted, discovery, failures, sweeps, segment int
	truncated                                     bool
	maximumSweep                                  float64
	latest                                        object
	output                                        io.Writer
}

func (c *collector) elapsed() int64 {
	value := (monotonic() - c.epoch) / int64(time.Millisecond)
	if value < 0 {
		return 0
	}
	return value
}
func (c *collector) send(message object) error { return json.NewEncoder(c.output).Encode(message) }
func (c *collector) publish(value object) error {
	c.latest = value
	return c.send(object{"type": "latest", "value": value})
}
func (c *collector) relationship(identity string, unit, allocation, invocation any, attribution string, elapsed int64) error {
	relation := object{"kind": "relationship", "identity_digest": identity, "unit_id": unit, "allocation_ref": allocation, "invocation_id": invocation, "provenance": attribution}
	// Execution v2 defines this digest over the producer's ordered JSON tuple,
	// not a sorted-map encoding. Preserve the published cross-language identity.
	key, _ := json.Marshal(struct {
		Kind       string `json:"kind"`
		Identity   string `json:"identity_digest"`
		Unit       any    `json:"unit_id"`
		Allocation any    `json:"allocation_ref"`
		Invocation any    `json:"invocation_id"`
		Provenance string `json:"provenance"`
	}{"relationship", identity, unit, allocation, invocation, attribution})
	if c.relationships[string(key)] {
		return nil
	}
	if len(c.relationships) >= c.policy.integer("maximum_execution_records") {
		c.omitted++
		c.truncated = true
		return nil
	}
	c.relationships[string(key)] = true
	relation["ref"] = "relationship:" + digest(key)[7:]
	relation["observed_elapsed_ms"] = elapsed
	return c.send(object{"type": "relationship", "record": relation})
}
func (c *collector) register(p proof, unit, allocation, invocation any, attribution string) error {
	existing := c.byIdentity[p.Identity]
	if existing == nil && len(c.processes) >= c.policy.integer("maximum_lifetime_processes") {
		c.omitted++
		c.truncated = true
		return nil
	}
	if err := c.relationship(p.Identity, unit, allocation, invocation, attribution, c.elapsed()); err != nil {
		return err
	}
	if existing != nil {
		if attribution == "registered" && existing.attribution == "registered" && (existing.unit != unit || existing.allocation != allocation || existing.invocation != invocation) {
			existing.conflicted = true
			existing.unit = nil
			existing.allocation = nil
			existing.invocation = nil
		} else if attribution == "registered" && !existing.conflicted {
			existing.unit = unit
			existing.allocation = allocation
			existing.invocation = invocation
			existing.attribution = attribution
		}
		return nil
	}
	entry := &processEntry{proof: p, ref: fmt.Sprintf("process:%d", len(c.processes)+1), unit: unit, allocation: allocation, invocation: invocation, attribution: attribution}
	c.processes = append(c.processes, entry)
	c.byIdentity[p.Identity] = entry
	return nil
}
func (c *collector) pauseStatus() error {
	var elapsed any
	if c.latest != nil {
		elapsed = c.latest["sampled_elapsed_ms"]
	}
	return c.publish(object{"availability": "paused_measurement", "sampled_elapsed_ms": elapsed, "rss_bytes": nil, "process_count": nil, "omitted_observations": c.omitted, "discovery_truncations": c.discovery, "failed_sweeps": c.failures})
}
func (c *collector) fenced(ctx context.Context, operation func(context.Context) error) (quiet bool, err error) {
	quiet, err = harnessadmission.Quiet(c.admissionRoot)
	if err != nil {
		return false, err
	}
	if quiet {
		c.segment++
		return true, c.pauseStatus()
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		return false, err
	}
	token := hex.EncodeToString(nonce[:])
	admissionCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	claims := harnessadmission.Quantities{"browser_stack": 0, "cpu": 0, "io": 0, "memory_mb": 0, "object_store": 0, "port_lane": 0, "postgres": 0, "process": 0, "service_stack": 0, "volume": 0}
	granted, err := harnessadmission.Transact(admissionCtx, harnessadmission.Request{Root: c.admissionRoot, Operation: "admit", Token: token, Owner: c.owner, Mode: "observation", Claims: claims, Capacities: c.capacities})
	// Release even a queued request and even when admission completed concurrently
	// with cancellation. No parent event loop participates in this transaction.
	defer func() {
		releaseCtx, stop := context.WithTimeout(context.Background(), c.policy.duration("stop_ms"))
		defer stop()
		_, releaseErr := harnessadmission.Transact(releaseCtx, harnessadmission.Request{Root: c.admissionRoot, Operation: "release", Token: token})
		if releaseErr != nil {
			err = errors.Join(err, errQuiescence, releaseErr)
		}
	}()
	if err != nil {
		return false, err
	}
	if !granted {
		return false, nil
	}
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	return false, operation(ctx)
}
func (c *collector) sweep(ctx context.Context) (bool, error) {
	if c.adapter == nil || c.truncated || c.store.truncated || c.store.failed {
		return false, nil
	}
	return c.fenced(ctx, c.collect)
}
func (c *collector) collect(ctx context.Context) error {
	start := time.Now()
	deadline := start.Add(c.policy.duration("sweep_budget_ms"))
	found, truncated, err := c.adapter.discover(c.policy.integer("maximum_discovery_entries"), deadline)
	if err != nil {
		return err
	}
	if truncated {
		c.discovery++
	}
	byPID := map[int]proof{}
	ancestors := map[int]*processEntry{}
	for _, p := range found {
		byPID[p.PID] = p
	}
	for _, entry := range c.processes {
		if !entry.gone {
			ancestors[entry.proof.PID] = entry
		}
	}
	for _, p := range found {
		if owned := c.byIdentity[p.Identity]; owned != nil && owned.attribution == "registered" {
			continue
		}
		ancestor := p
		seen := map[int]bool{}
		for !seen[ancestor.PID] {
			seen[ancestor.PID] = true
			parent := int(ancestor.Stat["parent"].(float64))
			observedParent, exists := byPID[parent]
			if !exists {
				break
			}
			owned := ancestors[parent]
			if owned != nil && observedParent.Identity == owned.proof.Identity {
				childStart, _ := strconv.ParseUint(p.Start, 10, 64)
				parentStart, _ := strconv.ParseUint(owned.proof.Start, 10, 64)
				if childStart >= parentStart {
					if err := c.register(p, owned.unit, owned.allocation, owned.invocation, "observed_descendant"); err != nil {
						return err
					}
					break
				}
			}
			ancestor = observedParent
		}
	}
	sampledAt := c.elapsed()
	live := 0
	rss := float64(0)
	completeRSS := true
	for _, entry := range c.processes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entry.gone {
			continue
		}
		if live >= c.policy.integer("maximum_live_processes") || !time.Now().Before(deadline) {
			c.omitted++
			completeRSS = false
			continue
		}
		a, metrics := c.adapter.sample(entry.proof)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if a == "process_gone" {
			entry.gone = true
		} else if a == "available" {
			live++
		}
		memory, _ := metrics["rss_bytes"].(object)
		if !entry.gone && memory["availability"] == "available" {
			rss += memory["value"].(float64)
		} else if !entry.gone {
			completeRSS = false
		}
		if err := c.store.append(object{"elapsed_ms": sampledAt, "segment": c.segment, "identity_digest": entry.proof.Identity, "scope": "process", "scope_ref": entry.ref, "availability": a, "metrics": metrics}); err != nil {
			return err
		}
	}
	if ctx.Err() == nil && sampledAt-c.lastContext >= int64(c.policy["context_cadence_ms"]) && time.Now().Before(deadline) {
		c.lastContext = sampledAt
		records, err := c.adapter.context()
		if err != nil {
			return err
		}
		for _, record := range records {
			if ctx.Err() != nil || !time.Now().Before(deadline) {
				c.omitted++
				break
			}
			record["elapsed_ms"] = sampledAt
			record["segment"] = c.segment
			record["availability"] = "available"
			if err := c.store.append(record); err != nil {
				return err
			}
		}
	}
	if ctx.Err() == nil && sampledAt-c.lastContainer >= int64(c.policy["context_cadence_ms"]) {
		c.lastContainer = sampledAt
		if !time.Now().Before(deadline) {
			c.containerOmissions++
		} else {
			batch, err := c.containers.sample(ctx)
			if err != nil {
				return err
			}
			c.containerOmissions += batch.omitted
			c.overruns += batch.overruns
			if len(c.containerIdentities) > 0 {
				c.containerAvailability = "available"
			} else {
				c.containerAvailability = batch.availability
			}
			for _, observation := range batch.observations {
				if ctx.Err() != nil {
					c.containerOmissions++
					continue
				}
				identity, ok := observation["identity_digest"].(string)
				if !ok {
					return os.ErrInvalid
				}
				ownership := observation["ownership"]
				if ownership != "owned" && ownership != "borrowed" {
					return os.ErrInvalid
				}
				metrics, ok := observation["metrics"].(object)
				if !ok {
					return os.ErrInvalid
				}
				for name := range metrics {
					if !c.containerSignals[name] {
						return os.ErrInvalid
					}
				}
				entry := c.containerIdentities[identity]
				if entry == nil && len(c.containerIdentities) >= c.policy.integer("maximum_lifetime_containers") {
					c.containerOmissions++
					continue
				}
				if entry != nil && entry["ownership"] != ownership {
					c.containerOmissions++
					continue
				}
				if entry == nil {
					entry = object{"container_ref": fmt.Sprintf("container:%d", len(c.containerIdentities)+1), "identity_digest": identity, "ownership": ownership, "allocation_ref": nil}
				}
				record := object{"elapsed_ms": c.elapsed(), "segment": c.segment, "scope": "container", "scope_ref": entry["container_ref"], "identity_digest": identity, "availability": observation["availability"], "metrics": metrics}
				if err := c.store.append(record); err != nil {
					return err
				}
				if c.containerIdentities[identity] == nil {
					c.containerIdentities[identity] = entry
					c.containerOrder = append(c.containerOrder, entry)
					if err := c.relationship(identity, nil, nil, nil, "registered", c.elapsed()); err != nil {
						return err
					}
				}
			}
		}
	}
	duration := float64(time.Since(start)) / float64(time.Millisecond)
	if duration > c.maximumSweep {
		c.maximumSweep = duration
	}
	c.sweeps++
	a := "not_observed"
	var memory, count any
	if live > 0 {
		a = "available"
		count = live
		if completeRSS {
			memory = rss
		}
	}
	return c.publish(object{"availability": a, "sampled_elapsed_ms": sampledAt, "rss_bytes": memory, "process_count": count, "omitted_observations": c.omitted, "discovery_truncations": c.discovery, "failed_sweeps": c.failures})
}
func (c *collector) finish(omitted int) error {
	start := time.Now()
	c.omitted += omitted
	c.truncated = c.truncated || omitted > 0
	status := "complete"
	if c.adapter == nil {
		status = "unsupported"
	} else if c.truncated {
		status = "truncated"
	} else if c.failures+c.omitted+c.discovery+c.containerOmissions+c.overruns > 0 {
		status = "partial"
	}
	var cpu unix.Rusage
	if err := unix.Getrusage(unix.RUSAGE_THREAD, &cpu); err != nil {
		return err
	}
	var bytes int64
	if c.adapter != nil {
		bytes = c.adapter.bytes
	}
	processes := []any{}
	for _, entry := range c.processes {
		processes = append(processes, object{"process_ref": entry.ref, "identity_digest": entry.proof.Identity})
	}
	facts := object{"schema_id": "cartulary.harness_resource_index.v3", "status": status, "clock": "graph_process_monotonic", "coverage": "observed_partial", "omitted_observations": c.omitted, "discovery_truncations": c.discovery, "failed_sweeps": c.failures, "sweeps": c.sweeps, "processes": processes,
		"containers": object{"availability": c.containerAvailability, "identities": c.containerOrder, "omitted_observations": c.containerOmissions, "deadline_overruns": c.overruns},
		"observer":   object{"qualification": "component_diagnostic_only", "collector_thread_cpu_user_us": cpu.Utime.Nano() / 1000, "collector_thread_cpu_system_us": cpu.Stime.Nano() / 1000, "collector_js_heap_peak_bytes": 0, "read_bytes": bytes, "write_bytes": c.store.bytes, "maximum_sweep_ms": c.maximumSweep, "shutdown_ms": 0, "outside_component_accounting": "startup_parent_helpers_readers_publication"}}
	for key, value := range c.identity {
		facts[key] = value
	}
	// Optional diagnostic publication cannot change the owning test outcome.
	_ = c.publishCheckpoints()
	return c.store.finish(facts, start)
}

// watcher sleeps in the kernel; it does not poll during exclusive measurement.
func watchAdmission(root string) (<-chan struct{}, func(), error) {
	fd, err := unix.InotifyInit1(unix.IN_CLOEXEC | unix.IN_NONBLOCK)
	if err != nil {
		return nil, nil, err
	}
	if _, err = unix.InotifyAddWatch(fd, root, unix.IN_MOVED_TO|unix.IN_CREATE|unix.IN_DELETE_SELF|unix.IN_MOVE_SELF); err != nil {
		unix.Close(fd)
		return nil, nil, err
	}
	file := os.NewFile(uintptr(fd), "admission-observation")
	events := make(chan struct{}, 1)
	go func() {
		defer close(events)
		buffer := make([]byte, 4096)
		for {
			_, err := file.Read(buffer)
			if err != nil {
				return
			}
			select {
			case events <- struct{}{}:
			default:
			}
		}
	}()
	return events, func() { file.Close() }, nil
}

// Main is a private, closed control protocol. It is never a general command
// runner. The pinned collection thread owns all publication and acknowledgement.
func Main(input io.Reader, output io.Writer) int {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	parent := os.Getppid()
	if parent <= 1 || unix.Prctl(unix.PR_SET_PDEATHSIG, uintptr(unix.SIGTERM), 0, 0, 0) != nil || os.Getppid() != parent {
		return 1
	}
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 65536), 65536)
	fail := func() int { json.NewEncoder(output).Encode(object{"type": "failed"}); return 1 }
	if !scanner.Scan() {
		return fail()
	}
	raw, err := harnessdiag.Decode(scanner.Bytes())
	if err != nil {
		return fail()
	}
	config, ok := raw.(object)
	if !ok || !closedCollectorConfig(config, "policy", "runRoot", "capacities", "admissionRoot", "containerProofFile", "runtimeLeaseID", "containerSignals", "epoch", "identity") {
		return fail()
	}
	rawPolicy, ok := config["policy"].(object)
	if !ok || harnessdiag.Validate("cartulary.harness_instrumentation_policy.v3", rawPolicy) != nil {
		return fail()
	}
	p := policy{}
	for key, value := range rawPolicy {
		if n, ok := value.(float64); ok {
			p[key] = n
		}
	}
	root, ok := config["runRoot"].(string)
	if !ok {
		return fail()
	}
	admissionRoot, ok := config["admissionRoot"].(string)
	if !ok {
		return fail()
	}
	epochText, ok := config["epoch"].(string)
	if !ok {
		return fail()
	}
	epoch, err := strconv.ParseInt(epochText, 10, 64)
	if err != nil || epoch < 0 {
		return fail()
	}
	identity, ok := config["identity"].(object)
	if !ok {
		return fail()
	}
	runID, ok := identity["run_id"].(string)
	if !ok || runID == "" {
		return fail()
	}
	caps, ok := config["capacities"].(object)
	if !ok {
		return fail()
	}
	capacities := harnessadmission.Quantities{}
	// Project graph-local fixture dimensions onto the existing host admission
	// resource domain, exactly as the Node hostClaims boundary does.
	for _, key := range []string{"browser_stack", "cpu", "io", "memory_mb", "object_store", "port_lane", "postgres", "process", "service_stack", "volume"} {
		value, present := caps[key]
		if !present {
			capacities[key] = 0
			continue
		}
		n, ok := value.(float64)
		if !ok {
			return fail()
		}
		capacities[key] = n
	}
	store, err := newStore(root, p)
	if err != nil {
		return fail()
	}
	defer store.stream.Abort()
	adapter, _ := newAdapter()
	owner, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return fail()
	}
	containerFile, _ := config["containerProofFile"].(string)
	leaseID, _ := config["runtimeLeaseID"].(string)
	if containerFile != "" && leaseID == "" {
		return fail()
	}
	c := &collector{policy: p, identity: identity, epoch: epoch, adapter: adapter, store: store, containers: &containerAdapter{file: containerFile, runID: runID, leaseID: leaseID, policy: p}, admissionRoot: admissionRoot, capacities: capacities, owner: owner, byIdentity: map[string]*processEntry{}, relationships: map[string]bool{}, containerIdentities: map[string]object{}, containerOrder: []any{}, containerSignals: map[string]bool{}, containerAvailability: "not_observed", lastContext: -1 << 50, lastContainer: -1 << 50, output: output}
	defer c.containers.close()
	names, ok := config["containerSignals"].([]any)
	if !ok {
		return fail()
	}
	for _, name := range names {
		s, ok := name.(string)
		if !ok {
			return fail()
		}
		c.containerSignals[s] = true
	}
	events := make(chan object, 16)
	var desired atomic.Int32
	var cancelMutex sync.Mutex
	var activeCancel context.CancelFunc
	go func() {
		defer close(events)
		defer func() {
			desired.Store(2)
			cancelMutex.Lock()
			if activeCancel != nil {
				activeCancel()
			}
			cancelMutex.Unlock()
		}()
		for scanner.Scan() {
			raw, err := harnessdiag.Decode(scanner.Bytes())
			if err != nil {
				return
			}
			message, ok := raw.(object)
			if !ok {
				return
			}
			kind, _ := message["type"].(string)
			if !validControl(message) {
				return
			}
			if kind == "pause" || kind == "stop" {
				if kind == "stop" {
					desired.Store(2)
				} else if desired.Load() != 2 {
					desired.Store(1)
				}
				cancelMutex.Lock()
				if activeCancel != nil {
					activeCancel()
				}
				cancelMutex.Unlock()
			}
			if kind == "resume" && desired.Load() != 2 {
				desired.Store(0)
			}
			events <- message
			if kind == "stop" {
				return
			}
		}
	}()
	timer := time.NewTimer(p.duration("process_cadence_ms"))
	defer timer.Stop()
	flush := time.NewTimer(p.duration("flush_ms"))
	defer flush.Stop()
	var wake <-chan struct{}
	var stopWatching func()
	defer func() {
		if stopWatching != nil {
			stopWatching()
		}
	}()
	reset := func() {
		if desired.Load() == 0 && !c.truncated && !c.store.truncated && !c.store.failed {
			timer.Reset(p.duration("process_cadence_ms"))
		}
	}
	observe := func(operation func(context.Context) (bool, error)) bool {
		ctx, cancel := context.WithCancel(context.Background())
		cancelMutex.Lock()
		activeCancel = cancel
		if desired.Load() != 0 {
			cancel()
		}
		cancelMutex.Unlock()
		quiet, sweepErr := operation(ctx)
		cancelMutex.Lock()
		activeCancel = nil
		cancelMutex.Unlock()
		cancel()
		if errors.Is(sweepErr, errQuiescence) {
			return true
		}
		if sweepErr != nil && !errors.Is(sweepErr, context.Canceled) {
			c.failures++
			value := object{"availability": "not_observed", "sampled_elapsed_ms": nil, "rss_bytes": nil, "process_count": nil}
			for key, v := range c.latest {
				value[key] = v
			}
			value["failed_sweeps"] = c.failures
			if c.publish(value) != nil {
				return true
			}
		}
		if quiet && desired.Load() == 0 {
			timer.Stop()
			flush.Stop()
			wake, stopWatching, err = watchAdmission(c.admissionRoot)
			if err != nil {
				return true
			}
			// Recheck after installing the watch to close the publication race.
			quiet, err = harnessadmission.Quiet(c.admissionRoot)
			if err != nil {
				return true
			}
			if !quiet {
				stopWatching()
				stopWatching = nil
				wake = nil
				flush.Reset(p.duration("flush_ms"))
				reset()
			}
		}
		return false

	}
	for {
		select {
		case message, open := <-events:
			if !open {
				return fail()
			}
			kind, _ := message["type"].(string)
			switch kind {
			case "register":
				if c.adapter != nil {
					data, err := json.Marshal(message["proof"])
					if err != nil {
						return fail()
					}
					var pr proof
					if json.Unmarshal(data, &pr) != nil || pr.PID < 1 || len(pr.Identity) != 71 || pr.Start == "" {
						return fail()
					}
					for _, name := range []string{"unit_id", "allocation_ref", "invocation_id"} {
						if v := message[name]; v != nil {
							if _, ok := v.(string); !ok {
								return fail()
							}
						}
					}
					if err = c.register(pr, message["unit_id"], message["allocation_ref"], message["invocation_id"], "registered"); err != nil {
						return fail()
					}
				}
				if c.send(object{"type": "registered", "request": message["request"]}) != nil {
					return fail()
				}
			case "checkpoints":
				if c.checkpoints == nil {
					c.checkpoints = &checkpointEvidence{}
				}
				if desired.Load() == 0 {
					if observe(func(ctx context.Context) (bool, error) { return c.fenced(ctx, c.captureCheckpoints) }) {
						return fail()
					}
				} else {
					_ = c.captureCheckpointServices(context.Background(), "measurement_quiet")
				}
				if c.send(object{"type": "checkpoints", "request": message["request"]}) != nil {
					return fail()
				}
			case "pause", "stop":
				timer.Stop()
				flush.Stop()
				if stopWatching != nil {
					stopWatching()
					stopWatching = nil
					wake = nil
				}
				c.segment++
				if kind == "stop" {
					omitted, _ := message["omitted_registrations"].(float64)
					if omitted < 0 || omitted != float64(int(omitted)) {
						return fail()
					}
					if c.finish(int(omitted)) != nil {
						return fail()
					}
					if c.send(object{"type": "stopped"}) != nil {
						return 1
					}
					return 0
				}
				if c.send(object{"type": "paused", "request": message["request"]}) != nil {
					return fail()
				}
			case "resume":
				c.segment++
				flush.Reset(p.duration("flush_ms"))
				reset()
			default:
				return fail()
			}
		case <-timer.C:
			if desired.Load() != 0 {
				continue
			}
			if observe(c.sweep) {
				return fail()
			}
			if wake == nil {
				reset()
			}
		case _, open := <-wake:
			if !open {
				return fail()
			}
			quiet, err := harnessadmission.Quiet(c.admissionRoot)
			if err != nil {
				return fail()
			}
			if !quiet {
				stopWatching()
				stopWatching = nil
				wake = nil
				flush.Reset(p.duration("flush_ms"))
				reset()
			}
		case <-flush.C:
			if desired.Load() == 0 && wake == nil {
				if observe(func(ctx context.Context) (bool, error) {
					return c.fenced(ctx, func(context.Context) error { return c.store.flush() })
				}) {
					return fail()
				}
				if wake == nil {
					flush.Reset(p.duration("flush_ms"))
				}
			}
		}
	}
}

func validControl(message object) bool {
	keys := map[string][]string{
		"register":    {"type", "request", "proof", "unit_id", "allocation_ref", "invocation_id"},
		"pause":       {"type", "request", "omitted_registrations"},
		"checkpoints": {"type", "request", "omitted_registrations"},
		"stop":        {"type", "request", "omitted_registrations"}, "resume": {"type"},
	}
	kind, _ := message["type"].(string)
	expected, ok := keys[kind]
	if !ok || len(message) != len(expected) {
		return false
	}
	for _, key := range expected {
		if _, ok := message[key]; !ok {
			return false
		}
	}
	if kind == "register" || kind == "pause" || kind == "checkpoints" {
		n, ok := message["request"].(float64)
		if !ok || n < 1 || n != float64(int64(n)) {
			return false
		}
	}
	if kind == "stop" && message["request"] != "stop" {
		return false
	}
	if kind == "stop" || kind == "pause" || kind == "checkpoints" {
		n, ok := message["omitted_registrations"].(float64)
		if !ok || n < 0 || n != float64(int64(n)) {
			return false
		}
	}
	return true
}

func closedCollectorConfig(value object, keys ...string) bool {
	if len(value) != len(keys) {
		return false
	}
	for _, key := range keys {
		if _, exists := value[key]; !exists {
			return false
		}
	}
	return true
}
