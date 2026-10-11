//go:build linux

package harnessruntime

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"os"
	"path/filepath"
	"sync"
)

type graphObservation struct {
	epoch, canonicalEpoch, last int64
	clock                       graphClock
	first                       map[string]any
	intervals                   []map[string]any
	manifest                    map[string]any
	runRoot                     string
	enabled                     bool
	execution                   *executionSession
	collector                   *collectorClient
	live                        *livePublisher
	wrapper                     *harnessdiag.GraphRegistration
	once                        sync.Once
	closeErr                    error
}

func readInstrumentationPolicy(root string) (map[string]any, error) {
	b, err := os.ReadFile(filepath.Join(root, "tools/harness_instrumentation_policy.json"))
	if err != nil {
		return nil, err
	}
	v, err := harnessschema.Decode(b)
	if err != nil {
		return nil, err
	}
	if err = harnessschema.Validate("cartulary.harness_instrumentation_policy.v3", v); err != nil {
		return nil, err
	}
	return v.(map[string]any), nil
}
func newGraphObservation(enabled bool) *graphObservation {
	d := &graphObservation{enabled: enabled, epoch: monotonicNS(), clock: newGraphClock(), intervals: []map[string]any{}}
	if enabled {
		d.first = d.clock.anchor()
	}
	return d
}
func (d *graphObservation) mark(phase string) {
	n := max(d.last, (monotonicNS()-d.epoch)/1e6)
	if d.enabled {
		d.intervals = append(d.intervals, map[string]any{"phase": phase, "start_ms": d.last, "end_ms": n})
	}
	d.last = n
}
func (d *graphObservation) start(binary, runRoot, admission string, runtime *harnessfixtures.Runtime, plan *Plan, policy map[string]any, environment map[string]string) {
	d.manifest = plan.facts.Manifest
	d.runRoot = runRoot
	if d.enabled && environment["CARTULARY_HARNESS_INVOCATION_ROOT"] != "" {
		d.wrapper = harnessdiag.RegisterGraph(environment["CARTULARY_HARNESS_INVOCATION_ROOT"], runRoot, plan.facts.Manifest)
	}
	d.execution = newExecutionSession(runtime, runRoot, plan.facts.Manifest, policy, func(proof, record map[string]any) {
		d.collector.register(proof, record["unit_id"], nil, record["invocation_id"])
	})
	if d.enabled {
		d.collector = newCollector(binary, runRoot, admission, runtime, plan, policy, d.epoch, d.execution.record)
	}
	d.live = newLivePublisher(runRoot, plan, d.collector.snapshot)
}
func (d *graphObservation) pause(ctx context.Context) error {
	d.live.pause()
	return d.collector.pause(ctx)
}
func (d *graphObservation) resume() { d.collector.resume(); d.live.resume() }
func (d *graphObservation) close() error {
	d.once.Do(func() {
		if d.execution != nil {
			d.execution.capture()
		}
		d.closeErr = d.collector.stop()
		d.live.stop()
		// Failure to drain is required cleanup failure. No index, envelope or
		// scan may assert finality while an ordinary producer could still write.
		if d.closeErr != nil {
			return
		}
		if d.execution != nil {
			var clock any
			if d.canonicalEpoch > 0 {
				anchors := []map[string]any{}
				if d.first != nil {
					anchors = append(anchors, d.first)
				}
				if last := d.clock.anchor(); last != nil {
					anchors = append(anchors, last)
				}
				clock = map[string]any{"clock_identity": d.clock.identity, "epoch_ms": float64(d.canonicalEpoch) / 1e6, "anchors": anchors}
			}
			_ = d.execution.publish(clock)
		}
		d.mark("collector_shutdown")
		if !d.enabled || d.manifest == nil {
			return
		}
		envelope := map[string]any{"schema_id": "cartulary.harness_invocation_envelope.v2", "run_id": d.manifest["run_id"], "source_digest": d.manifest["source_digest"], "graph_digest": d.manifest["graph_digest"], "invocation_digest": d.wrapper.Digest(), "boundary": "graph_main_entry_to_pre_scan", "clock": "graph_process_monotonic", "canonical_ref": "unit-events.ndjson", "intervals": d.intervals, "exclusions": []string{"make_and_preflight", "module_imports", "retained_secret_scan", "command_return", "receipt_publication"}}
		_ = writeEvidence(d.runRoot, "diagnostics/invocation-envelope.json", "cartulary.harness_invocation_envelope.v2", envelope, false)
	})
	return d.closeErr
}
func (d *graphObservation) scan(secrets []string) error {
	if err := d.close(); err != nil {
		return err
	}
	d.wrapper.ScanStarted()
	err := publishScan(d.runRoot, secrets)
	d.wrapper.ScanFinished(err == nil)
	return err
}
