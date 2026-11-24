package harnessplan

import (
	"encoding/json"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const (
	capabilitySnapshotSchema = "cartulary.harness_capability_snapshot.v1"
	capacityOverrideSchema   = "cartulary.harness_capacity_override.v1"
)

// Host is what capture detects about the machine. It is separate from the
// capacity rules so those can be tested without a host.
type Host struct {
	CPUs         int64
	MemoryBytes  int64
	CgroupMemory int64 // 0 when unbounded or unreadable
	ProcessLimit int64 // 0 when unbounded or unreadable
	WritableRoot bool
}

// CapacityOverride is an authored, bounded override of detected capacity.
type CapacityOverride struct {
	CPUTokens          *int64          `json:"cpu_tokens,omitempty"`
	MemoryBytes        *int64          `json:"memory_bytes,omitempty"`
	ProcessSlots       *int64          `json:"process_slots,omitempty"`
	IOTokens           *int64          `json:"io_tokens,omitempty"`
	PostgresLanes      *int64          `json:"postgres_lanes,omitempty"`
	PostgresHeavyLanes *int64          `json:"postgres_heavy_lanes,omitempty"`
	ObjectStoreLanes   *int64          `json:"object_store_lanes,omitempty"`
	PortLanes          *int64          `json:"port_lanes,omitempty"`
	WritableVolume     *bool           `json:"writable_volume,omitempty"`
	Services           map[string]bool `json:"services,omitempty"`
}

// ReadCapacityOverride reads and validates an override declaration.
func ReadCapacityOverride(path string) (*CapacityOverride, error) {
	data, err := readBounded(path)
	if err != nil {
		return nil, fmt.Errorf("%w: capacity override: %v", ErrInput, err)
	}
	document, err := harnessschema.Decode(data)
	if err != nil {
		return nil, fmt.Errorf("%w: capacity override: %v", ErrInput, err)
	}
	if err := harnessschema.Validate(capacityOverrideSchema, document); err != nil {
		return nil, fmt.Errorf("%w: capacity override: %v", ErrInput, err)
	}
	var override CapacityOverride
	if err := json.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("%w: capacity override: %v", ErrInput, err)
	}
	return &override, nil
}

// CapabilitySnapshot is the resolved capacity of one invocation's host.
type CapabilitySnapshot struct {
	SchemaID           string            `json:"schema_id"`
	CPUTokens          int64             `json:"cpu_tokens"`
	MemoryBytes        int64             `json:"memory_bytes"`
	ProcessSlots       int64             `json:"process_slots"`
	IOTokens           int64             `json:"io_tokens"`
	PostgresLanes      int64             `json:"postgres_lanes"`
	PostgresHeavyLanes int64             `json:"postgres_heavy_lanes"`
	ObjectStoreLanes   int64             `json:"object_store_lanes"`
	WritableVolume     bool              `json:"writable_volume"`
	PortLanes          int64             `json:"port_lanes"`
	Services           map[string]bool   `json:"services"`
	Sources            map[string]string `json:"sources"`
	SnapshotDigest     string            `json:"snapshot_digest"`
}

// Capture resolves the snapshot from host facts, the authored capacity
// policies and an optional override. An override may lower a detected
// capacity but never raise it past the policy bound.
func Capture(host Host, registry ResourceRegistry, override *CapacityOverride) (CapabilitySnapshot, error) {
	if override == nil {
		override = &CapacityOverride{}
	}
	p := registry.CapacityPolicies
	resolve := func(field string, value *int64, automatic, maximum int64) (int64, error) {
		resolved := automatic
		if value != nil {
			resolved = *value
		}
		if resolved < 1 || resolved > maximum {
			return 0, fmt.Errorf("%w: capacity override %s=%d exceeds the detected policy bound %d", ErrInput, field, resolved, maximum)
		}
		return resolved, nil
	}
	detectedCPU := max(p["cpu_tokens"].Minimum, host.CPUs*100/(100+p["cpu_tokens"].SafetyMarginPercent))
	memory := max(1, host.MemoryBytes)
	if host.CgroupMemory > 0 {
		memory = min(host.CgroupMemory, memory)
	}
	memory = max(p["memory_bytes"].Minimum, memory)

	s := CapabilitySnapshot{SchemaID: capabilitySnapshotSchema}
	var err error
	if s.CPUTokens, err = resolve("cpu_tokens", override.CPUTokens, detectedCPU, detectedCPU); err != nil {
		return s, err
	}
	if s.MemoryBytes, err = resolve("memory_bytes", override.MemoryBytes, memory, memory); err != nil {
		return s, err
	}
	cpu := s.CPUTokens
	fixed := func(field string) int64 {
		policy := p[field]
		return max(policy.Minimum, min(policy.Default, cpu*policy.MaximumCPUMultiplier))
	}
	processSlots := max(p["process_slots"].Minimum, cpu*p["process_slots"].FallbackCPUMultiplier)
	if host.ProcessLimit > 0 {
		processSlots = max(p["process_slots"].Minimum, host.ProcessLimit*(100-p["process_slots"].SafetyMarginPercent)/100)
	}
	if s.ProcessSlots, err = resolve("process_slots", override.ProcessSlots, processSlots, processSlots); err != nil {
		return s, err
	}
	io := max(p["io_tokens"].Minimum, cpu*p["io_tokens"].CPUMultiplier)
	if s.IOTokens, err = resolve("io_tokens", override.IOTokens, io, io); err != nil {
		return s, err
	}
	if s.PostgresLanes, err = resolve("postgres_lanes", override.PostgresLanes, fixed("postgres_lanes"), cpu*p["postgres_lanes"].MaximumCPUMultiplier); err != nil {
		return s, err
	}
	heavy := max(p["postgres_heavy_lanes"].Minimum, min(p["postgres_heavy_lanes"].Default, s.PostgresLanes))
	if s.PostgresHeavyLanes, err = resolve("postgres_heavy_lanes", override.PostgresHeavyLanes, heavy, s.PostgresLanes); err != nil {
		return s, err
	}
	if s.ObjectStoreLanes, err = resolve("object_store_lanes", override.ObjectStoreLanes, fixed("object_store_lanes"), cpu*p["object_store_lanes"].MaximumCPUMultiplier); err != nil {
		return s, err
	}
	if s.PortLanes, err = resolve("port_lanes", override.PortLanes, fixed("port_lanes"), cpu*p["port_lanes"].MaximumCPUMultiplier); err != nil {
		return s, err
	}
	s.WritableVolume = host.WritableRoot
	if override.WritableVolume != nil {
		s.WritableVolume = *override.WritableVolume
	}
	s.Services = map[string]bool{}
	for name, available := range override.Services {
		s.Services[name] = available
	}
	source := func(set bool, detected string) string {
		if set {
			return "override"
		}
		return detected
	}
	memorySource, processSource := "host", "fallback"
	if host.CgroupMemory > 0 {
		memorySource = "cgroup"
	}
	if host.ProcessLimit > 0 {
		processSource = "host"
	}
	s.Sources = map[string]string{
		"cpu_tokens":           source(override.CPUTokens != nil, "host"),
		"io_tokens":            source(override.IOTokens != nil, "host"),
		"memory_bytes":         source(override.MemoryBytes != nil, memorySource),
		"object_store_lanes":   source(override.ObjectStoreLanes != nil, "fallback"),
		"port_lanes":           source(override.PortLanes != nil, "host"),
		"postgres_heavy_lanes": source(override.PostgresHeavyLanes != nil, "fallback"),
		"postgres_lanes":       source(override.PostgresLanes != nil, "fallback"),
		"process_slots":        source(override.ProcessSlots != nil, processSource),
		"services":             source(override.Services != nil, "host"),
		"writable_volume":      source(override.WritableVolume != nil, "host"),
	}
	s.SnapshotDigest = semanticDigest(s, "snapshot_digest")
	return s, nil
}

// Capacities maps the snapshot onto the scheduler's resource names. This is
// the single definition of that mapping. A service the snapshot marks
// unavailable has no capacity.
func (s CapabilitySnapshot) Capacities() map[string]int64 {
	service := func(name string, lanes int64) int64 {
		if available, ok := s.Services[name]; ok && !available {
			return 0
		}
		return lanes
	}
	volume := int64(0)
	if s.WritableVolume {
		volume = 1
	}
	return map[string]int64{
		"browser_stack":  service("browser", s.PortLanes),
		"cpu":            s.CPUTokens,
		"io":             s.IOTokens,
		"memory_mb":      max(1, s.MemoryBytes/(1<<20)),
		"object_store":   service("object_store", s.ObjectStoreLanes),
		"port_lane":      s.PortLanes,
		"postgres":       service("postgres", s.PostgresLanes),
		"postgres_heavy": s.PostgresHeavyLanes,
		"process":        s.ProcessSlots,
		"service_stack":  service("service_stack", s.PortLanes),
		"volume":         volume,
	}
}
