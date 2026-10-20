//go:build linux

package harnesscollect

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/leaseproof"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckpointLogsAreBoundedAndNumeric(t *testing.T) {
	stamp := time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC)
	line := "2026-10-19T00:00:01Z 2026-10-19 00:00:01.000 UTC [123] LOG:  checkpoint complete: wrote 10 buffers (1.0%), wrote 0 SLRU buffers; 0 WAL file(s) added, 0 removed, 0 recycled; write=0.002 s, sync=0.003 s, total=0.005 s; sync files=1, longest=0.003 s, average=0.003 s; distance=2 kB, estimate=3 kB; lsn=0/10, redo lsn=0/F\n"
	t.Run("owned_service_observation_and_cancellation", func(t *testing.T) {
		c, _ := fixtureCollector(t)
		adapter, fake := containerFixtureAdapter(t)
		c.containers = adapter
		original := fake.inspect
		fake.inspect = func(id string, call int) (dockercontainer.InspectResponse, error) {
			v, err := original(id, call)
			v.State.StartedAt = stamp.Format(time.RFC3339Nano)
			return v, err
		}
		calls := 0
		adapter.client = &checkpointFake{observationFake: fake, logs: func(ctx context.Context, id string, options dockerclient.ContainerLogsOptions) (dockerclient.ContainerLogsResult, error) {
			calls++
			if id != "exact-container" || !options.Timestamps || !options.ShowStderr || options.ShowStdout || options.Follow || options.Details {
				t.Fatal("log request expanded ownership or logging")
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			header := make([]byte, 8)
			header[0] = 2
			binary.BigEndian.PutUint32(header[4:], uint32(len(line)))
			return io.NopCloser(bytes.NewReader(append(header, []byte(line)...))), nil
		}}
		if err := c.captureCheckpoints(context.Background()); err != nil || calls != 1 || len(c.checkpoints.records) != 1 {
			t.Fatal("owned log observation missing", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		cancelled, _ := fixtureCollector(t)
		cancelled.containers = adapter
		if err := cancelled.captureCheckpoints(ctx); err != nil || cancelled.checkpoints.unavailable != 1 {
			t.Fatal("optional cancellation changed outcome", err)
		}
		if err := c.publishCheckpoints(); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(filepath.Join(c.store.directory, "postgres-checkpoints.json"))
		if err != nil {
			t.Fatal(err)
		}
		value, err := harnessdiag.Decode(data)
		if err != nil || harnessdiag.Validate(checkpointSchema, value) != nil {
			t.Fatal("invalid numeric checkpoint evidence", err)
		}
		if strings.Contains(string(data), "exact-container") || strings.Contains(string(data), "LOG:") {
			t.Fatal("private log leaked")
		}
		if err := os.WriteFile(adapter.file, []byte("null"), 0600); err != nil {
			t.Fatal(err)
		}
		before := calls
		if err := c.captureCheckpoints(context.Background()); err != nil || calls != before {
			t.Fatal("revoked proof reached Docker")
		}
	})

	t.Run("multiple_generations_exact_retirement_and_clocks", func(t *testing.T) {
		c, _ := fixtureCollector(t)
		adapter, fake := containerFixtureAdapter(t)
		c.containers = adapter
		first, _ := observationFixture()
		second, _ := observationFixture()
		second.Lease.LeaseID = "lease-two"
		second.Lease.SuiteID = "suite-two"
		second.Lease.PostgresProvisioning.Generation = "2"
		second.Lease.Resources[0].ContainerID = "second-container"
		second.Lease.Resources[0].Labels[leaseproof.LabelSuiteID] = "suite-two"
		roster := observationRoster(first)
		roster.Services = append(roster.Services, second)
		publish := func() {
			t.Helper()
			b, _ := json.Marshal(roster)
			if err := os.WriteFile(adapter.file, b, 0600); err != nil {
				t.Fatal(err)
			}
		}
		fake.inspect = func(id string, _ int) (dockercontainer.InspectResponse, error) {
			proof := first
			started := stamp
			if id == "second-container" {
				proof = second
				started = stamp.Add(-time.Second)
			}
			return dockercontainer.InspectResponse{ID: id, State: &dockercontainer.State{Running: true, StartedAt: started.Format(time.RFC3339Nano)}, Config: &dockercontainer.Config{Labels: proof.Lease.Resources[0].Labels}}, nil
		}
		fake.stats = func(_ context.Context, id string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader(fmt.Sprintf(`{"id":%q,"memory_stats":{"usage":1}}`, id))), nil
		}
		calls := []string{}
		adapter.client = &checkpointFake{observationFake: fake, logs: func(_ context.Context, id string, _ dockerclient.ContainerLogsOptions) (dockerclient.ContainerLogsResult, error) {
			calls = append(calls, id)
			h := make([]byte, 8)
			h[0] = 2
			binary.BigEndian.PutUint32(h[4:], uint32(len(line)))
			return io.NopCloser(bytes.NewReader(append(h, []byte(line)...))), nil
		}}
		publish()
		batch, err := adapter.sample(t.Context())
		if err != nil || len(batch.observations) != 2 {
			t.Fatal("live roster lost sibling", err)
		}
		if err := c.captureCheckpoints(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := c.captureCheckpoints(t.Context()); err != nil || len(calls) != 1 {
			t.Fatal("duplicate retirement reread logs", err)
		}
		roster.Services = []containerObservationProof{second}
		roster.RetiringService = second.Lease.LeaseID
		publish()
		if err := c.captureCheckpoints(t.Context()); err != nil {
			t.Fatal(err)
		}
		if len(calls) != 2 || calls[0] != "exact-container" || calls[1] != "second-container" || len(c.checkpoints.records) != 2 || c.checkpoints.records[0].ElapsedMS != 1000 || c.checkpoints.records[1].ElapsedMS != 2000 {
			t.Fatal("generation/clock attribution lost")
		}
		if c.checkpoints.serviceResults[0].Generation != "1" || c.checkpoints.serviceResults[1].Generation != "2" || c.checkpoints.serviceResults[0].Identity == nil || *c.checkpoints.serviceResults[0].Identity == *c.checkpoints.serviceResults[1].Identity {
			t.Fatal("physical generation proof lost")
		}
		if err := c.publishCheckpoints(); err != nil {
			t.Fatal(err)
		}
	})

	frame := func(text string) []byte {
		header := make([]byte, 8)
		header[0] = 2
		binary.BigEndian.PutUint32(header[4:], uint32(len(text)))
		return append(header, []byte(text)...)
	}
	data := append(frame("raw secret SQL that must be discarded\n"+line[:100]), frame(line[100:])...)
	records, unknown, truncated := readCheckpointFrames(bytes.NewReader(data), stamp, 4096)
	if unknown != "" || truncated || len(records) != 1 || records[0].SyncMS != 3 {
		t.Fatalf("framing: count=%d cause=%v truncated=%v", len(records), unknown, truncated)
	}
	records, unknown, truncated = readCheckpointFrames(bytes.NewReader(frame(line+line)), stamp, 1)
	if len(records) != 1 || unknown != "" || !truncated {
		t.Fatal("record limit lost")
	}
	_, unknown, _ = readCheckpointFrames(bytes.NewReader(frame(strings.Replace(line, "sync=0.003", "sync=garbage", 1))), stamp, 4096)
	if unknown == "" {
		t.Fatal("unknown format accepted")
	}
	header := make([]byte, 8)
	header[0] = 2
	binary.BigEndian.PutUint32(header[4:], 1<<30)
	if _, unknown, _ := readCheckpointFrames(bytes.NewReader(header), stamp, 4096); unknown == "" {
		t.Fatal("oversized frame accepted")
	}
	_, unknown, truncated = readCheckpointFrames(io.LimitReader(strings.NewReader("\x02\x00\x00\x00\x00\x00"), 6), stamp, 4096)
	if unknown == "" || truncated {
		t.Fatal("incomplete frame misclassified")
	}
	large := bytes.Repeat(frame(strings.Repeat("x", 4096)+"\n"), 600)
	if _, _, truncated := readCheckpointFrames(bytes.NewReader(large), stamp, 4096); !truncated {
		t.Fatal("raw byte bound ignored")
	}
}

type checkpointFake struct {
	*observationFake
	logs func(context.Context, string, dockerclient.ContainerLogsOptions) (dockerclient.ContainerLogsResult, error)
}

func (f *checkpointFake) ContainerLogs(ctx context.Context, id string, options dockerclient.ContainerLogsOptions) (dockerclient.ContainerLogsResult, error) {
	return f.logs(ctx, id, options)
}
