//go:build linux

package harnesscollect

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
)

func fixturePolicy() policy {
	return policy{"process_cadence_ms": 2000, "context_cadence_ms": 5000, "sweep_budget_ms": 50, "maximum_live_processes": 512, "maximum_lifetime_processes": 4096, "maximum_discovery_entries": 8192, "maximum_record_bytes": 65536, "maximum_pending_bytes": 1048576, "maximum_sample_bytes": 67108864, "maximum_metadata_bytes": 4194304, "flush_bytes": 262144, "flush_ms": 5000, "stop_ms": 2000, "maximum_execution_records": 4096, "container_batch_deadline_ms": 250, "maximum_live_containers": 32, "maximum_lifetime_containers": 256}
}
func fixtureCollector(t *testing.T) (*collector, *bytes.Buffer) {
	t.Helper()
	p := fixturePolicy()
	root := filepath.Join(t.TempDir(), "run")
	store, err := newStore(root, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.stream.Abort() })
	a, err := newAdapter()
	if err != nil {
		t.Fatal(err)
	}
	owner, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	d := "sha256:" + strings.Repeat("a", 64)
	out := &bytes.Buffer{}
	c := &collector{policy: p, identity: object{"run_id": "fixture", "source_digest": d, "graph_digest": d, "policy_digest": d}, epoch: monotonic(), adapter: a, store: store, containers: &containerAdapter{policy: p}, admissionRoot: filepath.Join(t.TempDir(), "admission"), capacities: harnessadmission.Quantities{"browser_stack": 0, "cpu": 2, "io": 2, "memory_mb": 100, "process": 10}, owner: owner, byIdentity: map[string]*processEntry{}, relationships: map[string]bool{}, containerIdentities: map[string]object{}, containerOrder: []any{}, containerSignals: map[string]bool{"container_memory_usage_bytes": true}, containerAvailability: "not_observed", lastContext: -1 << 50, lastContainer: -1 << 50, output: out}
	return c, out
}
func TestCollectorLocalCounters(t *testing.T) {
	// Independent Linux stat arithmetic: 13 user ticks, 7 system ticks, 5 pages.
	fields := make([]string, 22)
	for i := range fields {
		fields[i] = "0"
	}
	fields[0] = "S"
	fields[1] = "3"
	fields[11] = "13"
	fields[12] = "7"
	fields[17] = "2"
	fields[19] = "123"
	fields[21] = "5"
	stat, err := parseStat("42 (name ) spaces) " + strings.Join(fields, " "))
	if err != nil {
		t.Fatal(err)
	}
	if stat["user_ticks"] != float64(13) || stat["system_ticks"] != float64(7) || stat["rss_pages"] != float64(5) {
		t.Fatalf("wrong independent counters: %v", stat)
	}
	for _, s := range []string{"missing", "1 (short) S 0", strings.Replace("42 (name) "+strings.Join(fields, " "), " 123 ", " 9007199254740992 ", 1)} {
		if _, err := parseStat(s); err == nil {
			t.Fatal("accepted malformed stat")
		}
	}
	a, err := newAdapter()
	if err != nil {
		t.Fatal(err)
	}
	p, err := a.proof(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	status, metrics := a.sample(p)
	if status != "available" || metrics["rss_bytes"].(object)["availability"] != "available" {
		t.Fatalf("self observation unavailable: %s", status)
	}
	p.Identity = "sha256:" + strings.Repeat("0", 64)
	status, metrics = a.sample(p)
	if status != "process_gone" || len(metrics) != 0 {
		t.Fatal("accepted replaced process identity")
	}
	if observed(-1)["value"] != nil || missing("permission_denied")["value"] != nil {
		t.Fatal("invented zero")
	}
}
func TestCollectorRelationshipsAndBounds(t *testing.T) {
	t.Run("exhausted sweep does not start Docker work", func(t *testing.T) {
		c, _ := fixtureCollector(t)
		adapter, fake := containerFixtureAdapter(t)
		c.containers = adapter
		// A zero budget is already exhausted before discovery, without a sleep
		// or a timing race. The separate 250 ms request deadline cannot admit
		// new work after the cooperative sweep boundary.
		c.policy["sweep_budget_ms"] = 0
		if err := c.collect(context.Background()); err != nil {
			t.Fatal(err)
		}
		if len(fake.inspected) != 0 || len(fake.sampled) != 0 || c.containerOmissions != 1 || c.lastContainer < 0 {
			t.Fatal("exhausted container batch was launched, hidden, or left due")
		}
	})
	c, out := fixtureCollector(t)
	p, err := c.adapter.proof(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	p.Identity = "sha256:" + strings.Repeat("a", 64)
	for _, unit := range []string{"unit:a", "unit:b", "unit:b"} {
		if err := c.register(p, unit, nil, nil, "registered"); err != nil {
			t.Fatal(err)
		}
	}
	if !bytes.Contains(out.Bytes(), []byte("relationship:0c7b13fc79b18823e4f32252443dd47c63a083a29780649917698d0eeb9e397b")) {
		t.Fatal("ordered execution v2 identity changed")
	}
	if len(c.processes) != 1 || !c.processes[0].conflicted || c.processes[0].unit != nil || len(c.relationships) != 2 {
		t.Fatal("conflicting attribution escaped")
	}
	if bytes.Count(out.Bytes(), []byte(`"type":"relationship"`)) != 2 {
		t.Fatal("duplicate relationship")
	}
	c.policy["maximum_lifetime_processes"] = 1
	p.Identity = "sha256:" + strings.Repeat("b", 64)
	if err := c.register(p, nil, nil, nil, "registered"); err != nil {
		t.Fatal(err)
	}
	if c.omitted != 1 || !c.truncated {
		t.Fatal("lifetime bound not disclosed")
	}
	if err := c.finish(0); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(c.store.directory, "resource-index.json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := harnessdiag.Decode(data)
	if err != nil || harnessdiag.Validate("cartulary.harness_resource_index.v3", v) != nil {
		t.Fatal("invalid published index")
	}
	if v.(object)["status"] != "truncated" {
		t.Fatal("loss reported complete")
	}
}

type failingStream struct {
	writes, aborts         int
	writeError, closeError error
}

func (s *failingStream) Write([]byte) error { s.writes++; return s.writeError }
func (s *failingStream) Flush() error       { return nil }
func (s *failingStream) Close() error       { return s.closeError }
func (s *failingStream) Abort() error       { s.aborts++; return nil }
func TestCollectorStorageFailures(t *testing.T) {
	for _, diskFailure := range []bool{false, true} {
		c, _ := fixtureCollector(t)
		c.store.stream.Abort()
		stream := &failingStream{writeError: errors.New("disk full")}
		c.store.stream = stream
		c.store.hash = sha256.New()
		if !diskFailure {
			c.policy["maximum_sample_bytes"] = 1
		}
		record := object{"elapsed_ms": 0, "segment": 0, "identity_digest": "sha256:" + strings.Repeat("a", 64), "scope": "process", "scope_ref": "process:1", "availability": "available", "metrics": object{"rss_bytes": observed(12)}}
		err := c.store.append(record)
		if diskFailure != (err != nil) {
			t.Fatalf("disk failure %v: %v", diskFailure, err)
		}
		if c.store.samples != 0 || c.store.bytes != 0 {
			t.Fatal("unwritten bytes counted")
		}
		if err := c.finish(0); err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(filepath.Join(c.store.directory, "resource-index.json"))
		v, _ := harnessdiag.Decode(data)
		index := v.(object)
		if diskFailure {
			if index["status"] != "collector_failed" || index["sample_digest"] != nil || stream.aborts != 1 {
				t.Fatal("disk failure concealed")
			}
		} else if index["status"] != "truncated" || stream.writes != 0 {
			t.Fatal("bounds allowed a write")
		}
	}
	c, _ := fixtureCollector(t)
	c.store.stream.Abort()
	c.store.stream = &failingStream{closeError: errors.New("publication interrupted")}
	if err := c.finish(0); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(c.store.directory, "resource-index.json"))
	v, _ := harnessdiag.Decode(data)
	if v.(object)["sample_digest"] != nil {
		t.Fatal("interrupted publication claimed a digest")
	}
}
func TestCollectorQuietAndCancellation(t *testing.T) {
	for _, control := range []object{{"type": "restart"}, {"type": "stop", "request": "stop"}, {"type": "resume", "extra": true}, {"type": "pause", "request": float64(1), "omitted_registrations": nil}} {
		if validControl(control) {
			t.Fatal("accepted open control protocol")
		}
	}
	c, _ := fixtureCollector(t)
	request := harnessadmission.Request{Root: c.admissionRoot, Operation: "admit", Token: strings.Repeat("a", 32), Owner: c.owner, Mode: "exclusive", Claims: harnessadmission.Quantities{"browser_stack": 0, "cpu": 0, "io": 0, "memory_mb": 0, "process": 0}, Capacities: c.capacities}
	granted, err := harnessadmission.Transact(context.Background(), request)
	if err != nil || !granted {
		t.Fatal(err)
	}
	if quiet, err := c.sweep(context.Background()); err != nil || !quiet {
		t.Fatal("did not pause for exclusive request", err)
	}
	if c.sweeps != 0 || c.store.samples != 0 || c.latest["sampled_elapsed_ms"] != nil {
		t.Fatal("sampled during quiet interval")
	}
	if _, err := harnessadmission.Transact(context.Background(), harnessadmission.Request{Root: c.admissionRoot, Operation: "release", Token: request.Token}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = c.sweep(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatal("ignored cancellation", err)
	}
	if c.store.samples != 0 {
		t.Fatal("cancelled sweep wrote")
	}
	p, err := c.adapter.proof(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err = c.register(p, nil, nil, nil, "registered"); err != nil {
		t.Fatal(err)
	}
	if _, err = c.sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if c.sweeps != 1 || c.store.samples == 0 {
		t.Fatal("did not resume")
	}
}

func containerFixtureAdapter(t *testing.T) (*containerAdapter, *observationFake) {
	t.Helper()
	proof, fake := observationFixture()
	file := filepath.Join(t.TempDir(), "proof.json")
	raw, _ := json.Marshal(proof)
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	return &containerAdapter{file: file, runID: "run", leaseID: "runtime", policy: fixturePolicy(), client: fake}, fake
}

func TestContainerMalformedAndTimeout(t *testing.T) {
	for _, kind := range []string{"duplicate", "oversize", "timeout", "revoked", "foreign", "missing"} {
		t.Run(kind, func(t *testing.T) {
			c, fake := containerFixtureAdapter(t)
			c.policy["container_batch_deadline_ms"] = 20
			switch kind {
			case "duplicate":
				fake.stats = func(context.Context, string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(`{"id":"a","id":"b"}`)), nil
				}
			case "oversize":
				fake.stats = func(context.Context, string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(strings.Repeat(" ", 1048577))), nil
				}
			case "timeout":
				fake.stats = func(ctx context.Context, _ string) (io.ReadCloser, error) { <-ctx.Done(); return nil, ctx.Err() }
			case "revoked":
				if err := os.WriteFile(c.file, []byte("null"), 0600); err != nil {
					t.Fatal(err)
				}
			case "foreign":
				c.runID = "other"
			case "missing":
				if err := os.Remove(c.file); err != nil {
					t.Fatal(err)
				}
			}
			batch, err := c.sample(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if kind == "revoked" || kind == "missing" {
				if batch.omitted != 0 || batch.availability != "not_observed" {
					t.Fatal("unavailable proof", batch)
				}
			} else if batch.omitted != 1 {
				t.Fatal("failure concealed", batch)
			}
			if kind == "timeout" && batch.overruns != 1 {
				t.Fatal("deadline concealed")
			}
			if (kind == "revoked" || kind == "foreign" || kind == "missing") && len(fake.inspected) != 0 {
				t.Fatal("unproven container observed")
			}
		})
	}
}

func TestContainerCancellationDrain(t *testing.T) {
	c, fake := containerFixtureAdapter(t)
	started, cancelled, drain := make(chan struct{}), make(chan struct{}), make(chan struct{})
	fake.stats = func(ctx context.Context, _ string) (io.ReadCloser, error) {
		close(started)
		<-ctx.Done()
		close(cancelled)
		<-drain
		return nil, ctx.Err()
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan containerBatch, 1)
	go func() { batch, _ := c.sample(ctx); done <- batch }()
	<-started
	cancel()
	<-cancelled
	select {
	case <-done:
		t.Fatal("acknowledged before drain")
	case <-time.After(10 * time.Millisecond):
	}
	close(drain)
	select {
	case batch := <-done:
		if batch.omitted != 1 || batch.overruns != 0 {
			t.Fatal("cancellation misreported", batch)
		}
	case <-time.After(time.Second):
		t.Fatal("drain did not complete")
	}
	closed := 0
	c.closeClient = func() error { closed++; return nil }
	if c.close() != nil || c.close() != nil || closed != 1 {
		t.Fatal("client closure is not idempotent")
	}
}
