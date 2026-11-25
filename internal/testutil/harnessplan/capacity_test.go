package harnessplan

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

func TestCapabilityCapture(t *testing.T) {
	inputs, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	registry := inputs.Registry
	host := Host{CPUs: 24, MemoryBytes: 32 << 30, ProcessLimit: 127192, WritableRoot: true}
	value := func(n int64) *int64 { return &n }

	t.Run("detected capacities follow the authored policies", func(t *testing.T) {
		snapshot, err := Capture(host, registry, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := snapshot.Capacities()
		want := map[string]int64{
			"cpu": 19, "io": 38, "memory_mb": 32768, "process": 95394,
			"postgres": 8, "postgres_generation": 8, "postgres_heavy": 4, "object_store": 8, "port_lane": 32,
			"browser_stack": 32, "service_stack": 32, "volume": 1,
		}
		for name, expected := range want {
			if got[name] != expected {
				t.Errorf("%s = %d, want %d", name, got[name], expected)
			}
		}
		if len(got) != len(want) {
			t.Fatalf("capacities = %v", got)
		}
		data, err := json.Marshal(snapshot)
		if err != nil {
			t.Fatal(err)
		}
		document, err := harnessschema.Decode(data)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate(capabilitySnapshotSchema, document); err != nil {
			t.Fatalf("snapshot does not satisfy its schema: %v", err)
		}
		again, err := Capture(host, registry, nil)
		if err != nil || again.SnapshotDigest != snapshot.SnapshotDigest {
			t.Fatal("snapshot digest is not deterministic")
		}
	})

	t.Run("a small host bounds lane counts by CPU", func(t *testing.T) {
		snapshot, err := Capture(Host{CPUs: 4, MemoryBytes: 8 << 30, WritableRoot: true}, registry, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := snapshot.Capacities()
		// 4 CPUs leave 3 tokens; lanes follow them, and the fallback process
		// bound is four slots per token.
		if got["cpu"] != 3 || got["postgres"] != 3 || got["postgres_heavy"] != 3 || got["object_store"] != 3 || got["process"] != 12 {
			t.Fatalf("capacities = %v", got)
		}
	})

	t.Run("cgroup memory bounds host memory", func(t *testing.T) {
		snapshot, err := Capture(Host{CPUs: 8, MemoryBytes: 32 << 30, CgroupMemory: 4 << 30}, registry, nil)
		if err != nil {
			t.Fatal(err)
		}
		if snapshot.MemoryBytes != 4<<30 || snapshot.Sources["memory_bytes"] != "cgroup" {
			t.Fatalf("memory = %d from %s", snapshot.MemoryBytes, snapshot.Sources["memory_bytes"])
		}
	})

	t.Run("an override lowers but never raises capacity", func(t *testing.T) {
		snapshot, err := Capture(host, registry, &CapacityOverride{PostgresLanes: value(8), PostgresHeavyLanes: value(2)})
		if err != nil {
			t.Fatal(err)
		}
		if got := snapshot.Capacities(); got["postgres"] != 8 || got["postgres_heavy"] != 2 || snapshot.Sources["postgres_lanes"] != "override" {
			t.Fatalf("capacities = %v, sources = %v", got, snapshot.Sources)
		}
		for name, override := range map[string]*CapacityOverride{
			"cpu above detected":       {CPUTokens: value(20)},
			"heavy above postgres":     {PostgresLanes: value(4), PostgresHeavyLanes: value(5)},
			"postgres above cpu bound": {CPUTokens: value(4), PostgresLanes: value(5)},
		} {
			if _, err := Capture(host, registry, override); !errors.Is(err, ErrInput) {
				t.Errorf("%s: Capture = %v, want ErrInput", name, err)
			}
		}
	})

	t.Run("an unavailable service has no capacity", func(t *testing.T) {
		snapshot, err := Capture(host, registry, &CapacityOverride{Services: map[string]bool{"postgres": false, "browser": false}})
		if err != nil {
			t.Fatal(err)
		}
		if got := snapshot.Capacities(); got["postgres"] != 0 || got["browser_stack"] != 0 || got["object_store"] != 8 {
			t.Fatalf("capacities = %v", got)
		}
	})
}
