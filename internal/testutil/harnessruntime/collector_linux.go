//go:build linux

package harnessruntime

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

type collectorClient struct {
	mu                                sync.Mutex
	cmd                               *exec.Cmd
	input                             io.WriteCloser
	queue                             chan []byte
	done, readDone, writeDone         chan struct{}
	acks                              map[string]chan struct{}
	pending                           map[string]int
	bytes, dropped, sequence          int
	failed, terminal, stopped, reaped bool
	latest                            map[string]any
	epoch                             int64
	policy, manifest                  map[string]any
	proofPath, lease                  string
	relationship                      func(map[string]any)
	stopOnce                          sync.Once
	stopErr                           error
}

func newCollector(binary, runRoot, admissionRoot string, runtime *harnessfixtures.Runtime, plan *Plan, policy map[string]any, epoch int64, relationship func(map[string]any)) *collectorClient {
	c := &collectorClient{queue: make(chan []byte, 64), done: make(chan struct{}), readDone: make(chan struct{}), writeDone: make(chan struct{}), acks: map[string]chan struct{}{}, pending: map[string]int{}, epoch: epoch, policy: policy, manifest: plan.facts.Manifest, lease: runtime.LeaseID, relationship: relationship, proofPath: filepath.Join(runtime.Root, "container-observation", "proof.json")}
	fail := func() *collectorClient {
		c.failed = true
		close(c.done)
		close(c.readDone)
		close(c.writeDone)
		return c
	}
	dir, err := harnesslocal.PrivateDirectory(filepath.Dir(c.proofPath))
	if err != nil {
		return fail()
	}
	dir.Close()
	cmd := exec.Command(binary, "collect")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	cmd.Stderr = io.Discard
	cmd.WaitDelay = 2 * time.Second
	closeFDs, err := harnessadmission.InheritWorkspace(cmd)
	if err != nil {
		return fail()
	}
	defer closeFDs()
	in, err := cmd.StdinPipe()
	if err != nil {
		return fail()
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		in.Close()
		return fail()
	}
	if err = cmd.Start(); err != nil {
		in.Close()
		out.Close()
		return fail()
	}
	c.cmd = cmd
	c.input = in
	go c.read(out)
	go func() {
		defer close(c.writeDone)
		for {
			select {
			case b := <-c.queue:
				if _, err := in.Write(b); err != nil {
					c.markFailed()
					return
				}
			case <-c.done:
				return
			}
		}
	}()
	go func() {
		var info unix.Siginfo
		var err error
		for {
			err = unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if err != unix.EINTR {
				break
			}
		}
		c.mu.Lock()
		if err == nil {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		_ = cmd.Wait()
		c.reaped = true
		in.Close()
		c.mu.Unlock()
		<-c.readDone
		c.mu.Lock()
		if err != nil || !c.terminal {
			c.failed = true
		}
		c.wakeLocked()
		c.mu.Unlock()
		close(c.done)
	}()
	instrumentation := plan.facts.Manifest["instrumentation"].(map[string]any)
	identity := map[string]any{"run_id": c.manifest["run_id"], "source_digest": c.manifest["source_digest"], "graph_digest": c.manifest["graph_digest"], "policy_digest": instrumentation["policy_digest"]}
	config := map[string]any{"policy": policy, "runRoot": runRoot, "capacities": plan.facts.Capacities, "admissionRoot": admissionRoot, "containerProofFile": c.proofPath, "runtimeLeaseID": c.lease, "containerSignals": runtimeContainerSignals(), "epoch": strconv.FormatInt(epoch, 10), "identity": identity}
	c.enqueue(config, false)
	if p, err := harnesscollect.ProcessIdentity(os.Getpid()); err == nil {
		c.register(p, nil, nil, nil)
	}
	return c
}
func (c *collectorClient) wakeLocked() {
	for id, ch := range c.acks {
		close(ch)
		delete(c.acks, id)
	}
}
func (c *collectorClient) markFailed() { c.mu.Lock(); c.failed = true; c.wakeLocked(); c.mu.Unlock() }
func (c *collectorClient) read(out io.ReadCloser) {
	defer close(c.readDone)
	defer out.Close()
	scanner := bufio.NewScanner(out)
	scanner.Buffer(make([]byte, 65536), 1024*1024)
	for scanner.Scan() {
		raw, err := harnessschema.Decode(scanner.Bytes())
		v, ok := raw.(map[string]any)
		if err != nil || !ok {
			c.markFailed()
			return
		}
		kind, _ := v["type"].(string)
		if kind == "relationship" {
			r, ok := v["record"].(map[string]any)
			if !ok || harnessschema.Validate(executionRecordSchema, r) != nil {
				c.markFailed()
				return
			}
			if c.relationship != nil {
				optionalObservation(func() { c.relationship(r) })
			}
			continue
		}
		c.mu.Lock()
		switch kind {
		case "registered":
			id := controlID(v["request"])
			c.bytes -= c.pending[id]
			delete(c.pending, id)
		case "latest":
			if value, ok := v["value"].(map[string]any); ok && !c.stopped {
				c.latest = value
			}
		case "paused":
			id := controlID(v["request"])
			if ch := c.acks[id]; ch != nil {
				close(ch)
				delete(c.acks, id)
			}
		case "stopped", "failed":
			c.terminal = true
			if kind == "failed" {
				c.failed = true
			}
			c.wakeLocked()
		default:
			c.failed = true
			c.wakeLocked()
		}
		c.mu.Unlock()
	}
	if scanner.Err() != nil {
		c.markFailed()
	}
}
func controlID(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if n, ok := v.(float64); ok {
		return strconv.FormatFloat(n, 'f', 0, 64)
	}
	return ""
}
func (c *collectorClient) enqueue(v map[string]any, registration bool) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failed || c.reaped || registration && c.stopped {
		return false
	}
	if registration {
		c.sequence++
		v["request"] = c.sequence
	}
	b, err := json.Marshal(v)
	if err != nil {
		c.dropped++
		return false
	}
	b = append(b, '\n')
	if registration && c.bytes+len(b) > int(c.policy["maximum_pending_bytes"].(float64)) {
		c.dropped++
		return false
	}
	select {
	case c.queue <- b:
		if registration {
			id := strconv.Itoa(c.sequence)
			c.pending[id] = len(b)
			c.bytes += len(b)
		}
		return true
	default:
		c.dropped++
		if !registration {
			c.failed = true
			c.wakeLocked()
		}
		return false
	}
}
func (c *collectorClient) register(proof map[string]any, unit, allocation, invocation any) {
	if c == nil {
		return
	}
	c.enqueue(map[string]any{"type": "register", "proof": proof, "unit_id": unit, "allocation_ref": allocation, "invocation_id": invocation}, true)
}
func (c *collectorClient) containers(proof map[string]any) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failed || c.stopped {
		return
	}
	var value any
	if proof != nil {
		v := maps.Clone(proof)
		v["version"] = 1
		v["run_id"] = c.manifest["run_id"]
		v["runtime_lease_id"] = c.lease
		value = v
	}
	b, err := json.Marshal(value)
	if err != nil || len(b) > 1024*1024 {
		c.dropped++
		return
	}
	if harnesslocal.ReplaceFile(c.proofPath, b) != nil {
		c.dropped++
	}
}
func (c *collectorClient) terminate() error {
	if c.cmd == nil {
		return nil
	}
	c.mu.Lock()
	if !c.reaped {
		_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
	}
	c.input.Close()
	c.mu.Unlock()
	timer := time.NewTimer(time.Duration(c.policy["stop_ms"].(float64)) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-c.done:
		<-c.writeDone
		return nil
	case <-timer.C:
		return failed("harness", "cleanup_error")
	}
}
func (c *collectorClient) request(ctx context.Context, kind string) error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	if c.failed {
		c.mu.Unlock()
		return c.terminate()
	}
	c.sequence++
	id := strconv.Itoa(c.sequence)
	var request any = c.sequence
	if kind == "stop" {
		id = "stop"
		request = id
	}
	ack := make(chan struct{})
	c.acks[id] = ack
	omitted := c.dropped
	c.mu.Unlock()
	if !c.enqueue(map[string]any{"type": kind, "request": request, "omitted_registrations": omitted}, false) {
		return c.terminate()
	}
	timer := time.NewTimer(time.Duration(c.policy["stop_ms"].(float64)) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ack:
	case <-ctx.Done():
		c.markFailed()
	case <-timer.C:
		c.markFailed()
	}
	c.mu.Lock()
	failed := c.failed
	c.mu.Unlock()
	if failed || kind == "stop" {
		return c.terminate()
	}
	return nil
}
func (c *collectorClient) pause(ctx context.Context) error {
	if c == nil {
		return nil
	}
	return c.request(ctx, "pause")
}
func (c *collectorClient) resume() {
	if c == nil {
		return
	}
	c.mu.Lock()
	stopped := c.stopped
	c.mu.Unlock()
	if !stopped {
		c.enqueue(map[string]any{"type": "resume"}, false)
	}
}
func (c *collectorClient) stop() error {
	if c == nil {
		return nil
	}
	c.stopOnce.Do(func() {
		c.mu.Lock()
		c.stopped = true
		c.mu.Unlock()
		c.stopErr = c.request(context.Background(), "stop")
	})
	return c.stopErr
}
func (c *collectorClient) snapshot() map[string]any {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	v := map[string]any{"coverage": "observed_partial", "omitted_observations": nil, "discovery_truncations": nil, "failed_sweeps": nil, "schema_id": "cartulary.harness_resource_live.v1", "clock": "graph_process_monotonic", "availability": "not_observed", "sampled_elapsed_ms": nil, "rss_bytes": nil, "process_count": nil}
	if c.failed {
		v["availability"] = "collector_failed"
		return v
	}
	maps.Copy(v, c.latest)
	var age any
	stale := false
	if sample, ok := v["sampled_elapsed_ms"].(float64); ok {
		n := max(0, float64((monotonicNS()-c.epoch)/1e6)-sample)
		age = n
		stale = n > c.policy["stale_ms"].(float64) && v["availability"] != "paused_measurement"
	}
	v["age_ms"] = age
	v["stale"] = stale
	return v
}
