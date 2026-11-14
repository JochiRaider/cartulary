//go:build linux

package harnesscollect

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"

	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

type observationFake struct {
	mu        sync.Mutex
	inspected []string
	sampled   []string
	inspect   func(string, int) (dockercontainer.InspectResponse, error)
	stats     func(context.Context, string) (io.ReadCloser, error)
}

func (f *observationFake) ContainerInspect(_ context.Context, id string, options dockerclient.ContainerInspectOptions) (dockerclient.ContainerInspectResult, error) {
	f.mu.Lock()
	f.inspected = append(f.inspected, id)
	call := len(f.inspected)
	f.mu.Unlock()
	if options.Size {
		return dockerclient.ContainerInspectResult{}, errors.New("expensive inspection")
	}
	value, err := f.inspect(id, call)
	return dockerclient.ContainerInspectResult{Container: value}, err
}
func (f *observationFake) ContainerStats(ctx context.Context, id string, options dockerclient.ContainerStatsOptions) (dockerclient.ContainerStatsResult, error) {
	f.mu.Lock()
	f.sampled = append(f.sampled, id)
	f.mu.Unlock()
	if options.Stream || options.IncludePreviousSample {
		return dockerclient.ContainerStatsResult{}, errors.New("streaming or previous sample requested")
	}
	body, err := f.stats(ctx, id)
	return dockerclient.ContainerStatsResult{Body: body}, err
}
func observationFixture() (containerObservationProof, *observationFake) {
	resource := leaseproof.LeaseResource{Kind: "container", Service: "postgres", ContainerID: "exact-container", Labels: map[string]string{
		leaseproof.LabelManaged: leaseproof.ManagedValue, leaseproof.LabelRunID: "run", leaseproof.LabelSuiteID: "suite", leaseproof.LabelService: "postgres"}}
	proof := containerObservationProof{Ownership: "owned", Lease: leaseproof.Lease{SchemaID: "cartulary.test_services.lease.v3", LeaseID: "lease", PostgresProvisioning: &leaseproof.PostgresProvisioning{Generation: "1"}, SuiteID: "suite", RunID: "run", OwnershipMode: "owned", CleanupState: "not_started", Resources: []leaseproof.LeaseResource{resource}}}
	fake := &observationFake{inspect: func(id string, _ int) (dockercontainer.InspectResponse, error) {
		return dockercontainer.InspectResponse{ID: id, State: &dockercontainer.State{Running: true, StartedAt: "start:1"}, Config: &dockercontainer.Config{Labels: resource.Labels}}, nil
	},
		stats: func(context.Context, string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(`{"id":"exact-container","cpu_stats":{"cpu_usage":{"total_usage":2500000}},"memory_stats":{"usage":2048,"limit":8192},"pids_stats":{"current":3},"blkio_stats":{"io_service_bytes_recursive":[{"op":"Read","value":4096},{"op":"Write","value":1024}]}}`)), nil
		}}
	return proof, fake
}
func TestContainerObserverExactProofAndCounterSemantics(t *testing.T) {
	for _, operation := range []string{"Read", "Write"} {
		t.Run(operation+"_aggregation_bounds", func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				values []uint64
				valid  bool
			}{
				{"exact_boundary", []uint64{9007199254740990, 1}, true},
				{"one_over", []uint64{9007199254740991, 1}, false},
				{"native_wrap", nil, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					values := tc.values
					if tc.name == "native_wrap" {
						values = make([]uint64, 2049)
						for i := range values {
							values[i] = 9007199254740991
						}
					}
					entries := make([]map[string]any, len(values))
					for i, value := range values {
						entries[i] = map[string]any{"op": operation, "value": value}
					}
					data, err := json.Marshal(map[string]any{"blkio_stats": map[string]any{"io_service_bytes_recursive": entries}})
					if err != nil {
						t.Fatal(err)
					}
					proof, fake := observationFixture()
					fake.stats = func(context.Context, string) (io.ReadCloser, error) {
						return io.NopCloser(strings.NewReader(string(data))), nil
					}
					result, err := observeContainer(t.Context(), fake, proof, proof.Lease.Resources[0])
					if (err == nil) != tc.valid {
						t.Fatalf("aggregate admission: valid=%t err=%v", tc.valid, err)
					}
					if tc.valid {
						metric := result.Metrics["container_block_"+strings.ToLower(operation)+"_bytes"]
						if metric.Availability != "available" || metric.Value == nil || *metric.Value != 9007199254740991 {
							t.Fatal("exact aggregate boundary lost", metric)
						}
					}
				})
			}
		})
	}
	proof, fake := observationFixture()
	result, err := observeContainer(context.Background(), fake, proof, proof.Lease.Resources[0])
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]float64{"container_cpu_us": 2500, "container_memory_usage_bytes": 2048, "container_memory_limit_bytes": 8192, "container_pids": 3, "container_block_read_bytes": 4096, "container_block_write_bytes": 1024} {
		if value := result.Metrics[key]; value.Value == nil || *value.Value != want {
			t.Fatalf("%s: %+v", key, value)
		}
	}
	if len(fake.inspected) != 2 || len(fake.sampled) != 1 || fake.sampled[0] != "exact-container" || result.Ownership != "owned" || !strings.HasPrefix(result.Identity, "sha256:") {
		t.Fatal("incorrect scope", result)
	}
	bytes, _ := json.Marshal(result)
	if strings.Contains(string(bytes), "exact-container") || strings.Contains(string(bytes), "start:1") {
		t.Fatal("raw identity escaped")
	}
	fake.stats = func(context.Context, string) (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader(`{"cpu_stats":{},"memory_stats":{}}`)), nil
	}
	missing, err := observeContainer(context.Background(), fake, proof, proof.Lease.Resources[0])
	if err != nil {
		t.Fatal(err)
	}
	if missing.Metrics["container_cpu_us"].Value != nil || missing.Metrics["container_memory_usage_bytes"].Value != nil {
		t.Fatal("missing data became zero")
	}
}
func TestContainerObserverRejectsForeignReplacedRestartedAndMalformed(t *testing.T) {
	for _, mode := range []string{"foreign", "replacement", "restart", "missing", "duplicate", "oversize"} {
		t.Run(mode, func(t *testing.T) {
			proof, fake := observationFixture()
			original := fake.inspect
			fake.inspect = func(id string, call int) (dockercontainer.InspectResponse, error) {
				v, e := original(id, call)
				switch mode {
				case "foreign":
					v.Config = &dockercontainer.Config{Labels: map[string]string{}}
				case "replacement":
					v.ID = "other"
				case "restart":
					if call == 2 {
						v.State.StartedAt = "start:2"
					}
				case "missing":
					return v, errors.New("gone")
				}
				return v, e
			}
			if mode == "duplicate" {
				fake.stats = func(context.Context, string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(`{"cpu_stats":{},"cpu_stats":{}}`)), nil
				}
			}
			if mode == "oversize" {
				fake.stats = func(context.Context, string) (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(strings.Repeat(" ", 1048577))), nil
				}
			}
			if _, err := observeContainer(context.Background(), fake, proof, proof.Lease.Resources[0]); err == nil {
				t.Fatal("invalid observation accepted")
			}
			if (mode == "foreign" || mode == "replacement") && len(fake.sampled) != 0 {
				t.Fatal("unproven container sampled")
			}
		})
	}
}
func TestContainerObserverDeadlineCancellationAndBatchCap(t *testing.T) {
	proof, fake := observationFixture()
	fake.stats = func(ctx context.Context, _ string) (io.ReadCloser, error) { <-ctx.Done(); return nil, ctx.Err() }
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	result := collectContainerBatch(ctx, fake, []containerObservationProof{proof})
	if result.Overruns != 1 || result.Omitted != 1 {
		t.Fatal(result)
	}
	proof, fake = observationFixture()
	resource := proof.Lease.Resources[0]
	proof.Lease.Resources = make([]leaseproof.LeaseResource, 40)
	for i := range proof.Lease.Resources {
		proof.Lease.Resources[i] = resource
	}
	result = collectContainerBatch(context.Background(), fake, []containerObservationProof{proof})
	if len(result.Observations) != 32 || result.Omitted != 8 || len(fake.sampled) != 32 {
		t.Fatal("batch cap", result)
	}
	proof, fake = observationFixture()
	ctx, cancel = context.WithCancel(context.Background())
	cancel()
	fake.stats = func(ctx context.Context, _ string) (io.ReadCloser, error) { return nil, ctx.Err() }
	result = collectContainerBatch(ctx, fake, []containerObservationProof{proof})
	if result.Omitted != 1 || result.Overruns != 0 {
		t.Fatal("cancellation", result)
	}
}
func TestContainerObserverPrivateOwnershipAndBorrowerProof(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "proof.json")
	proof, _ := observationFixture()
	write := func(value any) {
		t.Helper()
		raw, _ := json.Marshal(observationRoster(value.(containerObservationProof)))
		if err := os.WriteFile(file, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write(proof)
	if _, err := readContainerObservationProofs(file, "run", "runtime"); err != nil {
		t.Fatal(err)
	}
	if _, err := readContainerObservationProofs(file, "foreign", "runtime"); err == nil {
		t.Fatal("foreign run accepted")
	}
	proof.Lease.RunID = "foreign"
	write(proof)
	if _, err := readContainerObservationProofs(file, "run", "runtime"); err == nil {
		t.Fatal("foreign lease accepted")
	}
	proof, _ = observationFixture()
	proof.Ownership = "borrowed"
	proof.BorrowerFile = filepath.Join(root, "borrower.json")
	proof.BorrowerID = strings.Repeat("b", 32)
	proof.Lease.SuiteID = strings.Repeat("a", 24)
	for i := range proof.Lease.Resources {
		proof.Lease.Resources[i].Labels[leaseproof.LabelSuiteID] = proof.Lease.SuiteID
	}
	identity, err := ProcessIdentity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	borrower := map[string]any{"schema_id": "cartulary.test_services.local_session_borrower.v2", "session_id": proof.Lease.SuiteID, "borrower_id": proof.BorrowerID, "owner_uid": os.Getuid(), "run_id": "run", "pid": os.Getpid(), "process_start_proof": "linux-proc-start:" + identity["start"].(string), "created_at": "2026-10-09T00:00:00Z"}
	writeBorrower := func(value map[string]any) {
		t.Helper()
		raw, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(proof.BorrowerFile, raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	writeBorrower(borrower)
	write(proof)
	if got, err := readContainerObservationProofs(file, "run", "runtime"); err != nil || got.Services[0].Ownership != "borrowed" {
		t.Fatal("borrower", err)
	}
	for _, mutate := range []func(map[string]any){
		func(v map[string]any) { delete(v, "owner_uid") },
		func(v map[string]any) { v["owner_uid"] = os.Getuid() + 1 },
		func(v map[string]any) { v["schema_id"] = "cartulary.test_services.local_session_borrower.v1" },
		func(v map[string]any) { v["process_start_proof"] = "linux-proc-start:0" },
		func(v map[string]any) { v["pid"] = 0 },
		func(v map[string]any) { v["run_id"] = "foreign" },
		func(v map[string]any) { v["extra"] = true },
	} {
		changed := make(map[string]any, len(borrower))
		for k, v := range borrower {
			changed[k] = v
		}
		mutate(changed)
		writeBorrower(changed)
		if _, err := readContainerObservationProofs(file, "run", "runtime"); err == nil {
			t.Fatal("invalid borrower authority accepted", changed)
		}
	}
	writeBorrower(borrower)
	if err := os.Remove(proof.BorrowerFile); err != nil {
		t.Fatal(err)
	}
	if _, err := readContainerObservationProofs(file, "run", "runtime"); err == nil {
		t.Fatal("detached borrower accepted")
	}
	raw, _ := os.ReadFile(file)
	raw = append([]byte(`{"run_id":"foreign",`), raw[1:]...)
	if err := os.WriteFile(file, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readContainerObservationProofs(file, "run", "runtime"); err == nil {
		t.Fatal("duplicate proof accepted")
	}
	write(proof)
	if err := os.Chmod(file, 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := readContainerObservationProofs(file, "run", "runtime"); err == nil {
		t.Fatal("public proof accepted")
	}
}

func observationRoster(proof containerObservationProof) containerObservationRoster {
	return containerObservationRoster{Version: 2, RunID: "run", RuntimeLeaseID: "runtime", Services: []containerObservationProof{proof}, RetiringService: proof.Lease.LeaseID}
}
