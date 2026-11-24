package harnessplan

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

// PlanSchema is the runtime plan document the planner writes.
const PlanSchema = "cartulary.harness_runtime_plan.v9"

// ErrPlan reports invocation facts that do not form a valid plan.
var ErrPlan = errors.New("invalid harness runtime plan")

const (
	manifestSchema       = "cartulary.harness_run_manifest.v3"
	sourceSnapshotSchema = "cartulary.source_snapshot_digest.v2"
	// makeInputSources is always cleared and never forwarded: a unit must not
	// inherit the enclosing Make invocation's input provenance.
	makeInputSources = "CARTULARY_MAKE_INPUT_SOURCES"
	// maximumArgumentBytes keeps every command argument well inside the
	// kernel's per-argument limit (128 KiB).
	maximumArgumentBytes = 32 << 10
)

// Plan is a compiled graph bound to one invocation: everything the runtime
// needs to execute and publish it. Its JSON form is the published plan, which
// omits the launch state (bindings and input environment).
type Plan struct {
	SchemaID string   `json:"schema_id"`
	Graph    *Graph   `json:"graph"`
	Manifest Manifest `json:"manifest"`
	// InputEnvironment is launch state: the runtime applies it to each unit's
	// environment, and the published plan never carries it.
	InputEnvironment           InputEnvironment      `json:"-"`
	BrowserLaneLimit           int                   `json:"browser_lane_limit"`
	HostProfile                HostProfile           `json:"host_profile"`
	InstrumentationPolicyBytes string                `json:"instrumentation_policy_bytes"`
	Projections                map[string]Projection `json:"projections"`
	Capacities                 map[string]int64      `json:"capacities"`
	AgingQuantumMS             int64                 `json:"aging_quantum_ms"`
	// Bindings are private launch state: the runtime binds them when it
	// starts a unit, and the published plan never carries them.
	Bindings     map[string]string `json:"-"`
	Cache        []CacheFact       `json:"cache"`
	SourceInputs []SourceInput     `json:"source_inputs"`
	CacheInputs  map[string][]int  `json:"cache_inputs"`
	// binaries maps each runtime binary ID to the environment name that
	// carries its path.
	binaries map[string]string
}

// LaunchEnvironment is a unit's command environment with its runtime
// binaries bound: each binary the unit consumes is set to its bound path and
// every other runtime binary is cleared, so a unit cannot reach a binary its
// dependencies did not produce.
func (p *Plan) LaunchEnvironment(u *Unit) map[string]string {
	environment := maps.Clone(u.Command.Environment)
	for id, name := range p.binaries {
		environment[name] = ""
		if slices.Contains(u.RuntimeBinaryIDs, id) {
			environment[name] = p.Bindings[name]
		}
	}
	return environment
}

// Manifest is the run manifest (cartulary.harness_run_manifest.v3).
type Manifest struct {
	SchemaID           string             `json:"schema_id"`
	RunID              string             `json:"run_id"`
	CommandID          string             `json:"command_id"`
	Target             string             `json:"target"`
	DeclaredInputs     map[string]string  `json:"declared_inputs"`
	SourceCommit       string             `json:"source_commit"`
	SourceState        string             `json:"source_state"`
	SourceDigest       string             `json:"source_digest"`
	ToolchainDigest    string             `json:"toolchain_digest"`
	SystemDigest       string             `json:"system_digest"`
	CapabilitySnapshot CapabilitySnapshot `json:"capability_snapshot"`
	GoWorkerBudget     int64              `json:"go_worker_budget"`
	GraphDigest        string             `json:"graph_digest"`
	CacheMode          string             `json:"cache_mode"`
	StartedAt          string             `json:"started_at"`
	Instrumentation    Instrumentation    `json:"instrumentation"`
}

// Instrumentation is the run's instrumentation mode and policy identity.
type Instrumentation struct {
	Mode         string `json:"mode"`
	PolicyDigest string `json:"policy_digest"`
}

// InputEnvironment is how a unit's environment differs from the runtime's:
// every cleared name is removed, then the forwarded values are set.
type InputEnvironment struct {
	Clear   []string          `json:"clear"`
	Forward map[string]string `json:"forward"`
}

// HostProfile identifies the host the plan was prepared on.
type HostProfile struct {
	Platform      string   `json:"platform"`
	Architecture  string   `json:"architecture"`
	KernelRelease string   `json:"kernel_release"`
	CPUModels     []string `json:"cpu_models"`
}

// Projection is the set of units a public target reports on.
type Projection struct {
	CommandID string   `json:"command_id"`
	UnitIDs   []string `json:"unit_ids"`
}

// SourceInput is one executable source file in the run's source snapshot.
type SourceInput struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	Mode       string `json:"mode"`
	ByteDigest string `json:"byte_digest"`
}

// CacheFact is a unit's cache eligibility. An eligible unit names its cache
// profile, policy, input digest and dependency closure; an ineligible one
// names only its reason.
type CacheFact struct {
	UnitID                        string  `json:"unit_id"`
	Eligible                      bool    `json:"eligible"`
	Reason                        *string `json:"reason"`
	ProfileID                     *string `json:"profile_id"`
	Policy                        *string `json:"policy"`
	InputDigest                   *string `json:"input_digest"`
	DependencyStrategy            *string `json:"dependency_strategy"`
	VulnerabilityDatabaseRevision *string `json:"vulnerability_database_revision"`
	DependencyDigest              *string `json:"dependency_digest"`
}

// Invocation is what one invocation binds into its plan besides the graph.
type Invocation struct {
	RunID     string
	CommandID string
	Target    string
	// DeclaredInputs are the command's resolved public inputs.
	DeclaredInputs   map[string]string
	InputEnvironment InputEnvironment
	BrowserLaneLimit int
	Snapshot         CapabilitySnapshot
	GoWorkerBudget   int64
	SourceInputs     []SourceInput
	SourceCommit     string
	SourceState      string
	ToolchainDigest  string
	CacheMode        string
	// InstrumentationMode is "off" or "basic"; InstrumentationPolicy is the
	// policy document's bytes.
	InstrumentationMode   string
	InstrumentationPolicy []byte
	Host                  HostProfile
	// Bindings are the launch bindings: names the runtime resolves to paths
	// and modes when it starts a unit.
	Bindings    map[string]string
	Cache       []CacheFact
	CacheInputs map[string][]int
	StartedAt   time.Time
	// Projections are the targets a command or aggregate graph reports on
	// besides the selected command itself.
	Projections map[string]Projection
}

// Assemble binds a compiled graph to one invocation and validates the result.
func Assemble(in *Inputs, graph *Graph, invocation Invocation) (*Plan, error) {
	var policy any
	if err := json.Unmarshal(invocation.InstrumentationPolicy, &policy); err != nil {
		return nil, fmt.Errorf("%w: instrumentation policy: %v", ErrPlan, err)
	}
	units := graph.Sorted()
	ids := make([]string, len(units))
	for i, unit := range units {
		ids[i] = unit.ID
	}
	plan := &Plan{
		SchemaID: PlanSchema,
		Graph:    graph,
		Manifest: Manifest{
			SchemaID: manifestSchema, RunID: invocation.RunID, CommandID: invocation.CommandID, Target: invocation.Target,
			DeclaredInputs: maps.Clone(invocation.DeclaredInputs), SourceCommit: invocation.SourceCommit, SourceState: invocation.SourceState,
			SourceDigest:    semanticDigest(map[string]any{"schema_id": sourceSnapshotSchema, "entries": invocation.SourceInputs}, ""),
			ToolchainDigest: invocation.ToolchainDigest, SystemDigest: invocation.Snapshot.SnapshotDigest,
			CapabilitySnapshot: invocation.Snapshot, GoWorkerBudget: invocation.GoWorkerBudget, GraphDigest: graph.Digest,
			CacheMode: invocation.CacheMode, StartedAt: invocation.StartedAt.UTC().Format(time.RFC3339Nano),
			Instrumentation: Instrumentation{Mode: invocation.InstrumentationMode, PolicyDigest: semanticDigest(policy, "")},
		},
		InputEnvironment:           invocation.InputEnvironment,
		BrowserLaneLimit:           invocation.BrowserLaneLimit,
		HostProfile:                invocation.Host,
		InstrumentationPolicyBytes: string(invocation.InstrumentationPolicy),
		Projections:                map[string]Projection{},
		Capacities:                 invocation.Snapshot.Capacities(),
		AgingQuantumMS:             in.Owner.AgingQuantumMS,
		Bindings:                   maps.Clone(invocation.Bindings),
		Cache:                      slices.Clone(invocation.Cache),
		SourceInputs:               slices.Clone(invocation.SourceInputs),
		CacheInputs:                maps.Clone(invocation.CacheInputs),
		binaries:                   map[string]string{},
	}
	maps.Copy(plan.Projections, invocation.Projections)
	plan.Projections[invocation.Target] = Projection{CommandID: invocation.CommandID, UnitIDs: ids}
	for _, binary := range in.Topology.RuntimeBinaries {
		plan.binaries[binary.ID] = binary.ConsumerEnv
	}
	if plan.CacheInputs == nil {
		plan.CacheInputs = map[string][]int{}
	}
	if err := plan.validate(); err != nil {
		return nil, err
	}
	// The planner owns its document contract: a plan that does not satisfy
	// its schema is a planner error, never a later publication failure.
	data, err := json.Marshal(plan)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrPlan, err)
	}
	document, err := harnessschema.Decode(data)
	if err == nil {
		err = harnessschema.Validate(PlanSchema, document)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: the plan does not satisfy %s: %v", ErrPlan, PlanSchema, err)
	}
	return plan, nil
}

// validate checks every invariant the runtime relies on before it launches
// work.
func (p *Plan) validate() error {
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%w: %s", ErrPlan, fmt.Sprintf(format, args...))
	}
	m := p.Manifest
	if m.GraphDigest != p.Graph.Digest || m.SystemDigest != m.CapabilitySnapshot.SnapshotDigest {
		return fail("manifest digests do not match the graph and snapshot")
	}
	if !slices.Contains([]string{"normal", "cold", "off"}, m.CacheMode) || !slices.Contains([]string{"off", "basic"}, m.Instrumentation.Mode) {
		return fail("unknown cache or instrumentation mode")
	}
	if m.GoWorkerBudget < 1 || m.GoWorkerBudget > m.CapabilitySnapshot.CPUTokens {
		return fail("Go worker budget %d is outside 1..%d", m.GoWorkerBudget, m.CapabilitySnapshot.CPUTokens)
	}
	if p.BrowserLaneLimit < 1 || p.BrowserLaneLimit > 6 {
		return fail("browser lane limit %d is outside 1..6", p.BrowserLaneLimit)
	}
	if declared, ok := m.DeclaredInputs["BROWSER_LANES"]; ok && declared != strconv.Itoa(p.BrowserLaneLimit) {
		return fail("declared BROWSER_LANES %s disagrees with the lane limit %d", declared, p.BrowserLaneLimit)
	}
	cleared := p.InputEnvironment.Clear
	if !strictlySorted(cleared) || !slices.Contains(cleared, makeInputSources) {
		return fail("input environment must clear a sorted set including %s", makeInputSources)
	}
	for name := range p.InputEnvironment.Forward {
		if name == makeInputSources || !slices.Contains(cleared, name) {
			return fail("input environment forwards %s without clearing it", name)
		}
	}
	if !maps.Equal(p.Capacities, m.CapabilitySnapshot.Capacities()) {
		return fail("capacities are not the snapshot's")
	}
	for i, input := range p.SourceInputs {
		if !safeRelative(input.Path) || i > 0 && p.SourceInputs[i-1].Path >= input.Path {
			return fail("source inputs must be sorted, unique, relative paths")
		}
	}
	if err := p.validateUnits(); err != nil {
		return err
	}
	if err := p.validateProjections(); err != nil {
		return err
	}
	return p.validateCache()
}

func (p *Plan) validateUnits() error {
	fail := func(unit, format string, args ...any) error {
		return fmt.Errorf("%w: unit %s: %s", ErrPlan, unit, fmt.Sprintf(format, args...))
	}
	residents := map[string]Resident{}
	for _, u := range p.Graph.Sorted() {
		if len(u.Claims) == 0 {
			return fail(u.ID, "declares no claims")
		}
		for _, argument := range u.Command.Args {
			if len(argument) > maximumArgumentBytes {
				return fail(u.ID, "a command argument exceeds %d bytes", maximumArgumentBytes)
			}
		}
		for _, id := range u.RuntimeBinaryIDs {
			if p.Bindings[p.binaries[id]] == "" {
				return fail(u.ID, "runtime binary %s has no launch binding", id)
			}
		}
		if slices.Contains(u.SharedLocks, "host_activity") == slices.Contains(u.ExclusiveLocks, "host_activity") {
			return fail(u.ID, "must hold host_activity in exactly one mode")
		}
		footprint := maps.Clone(u.Claims)
		for _, resident := range u.Residents {
			if prior, ok := residents[resident.Key]; ok && (prior.Scope != resident.Scope || !maps.Equal(prior.Claims, resident.Claims)) {
				return fail(u.ID, "resident %s disagrees with another unit's", resident.Key)
			}
			residents[resident.Key] = resident
			if resident.Scope == ScopeAttached {
				continue
			}
			for name, amount := range resident.Claims {
				switch name {
				case "cpu", "io":
					return fail(u.ID, "resident %s claims active resource %s", resident.Key, name)
				case "volume":
					footprint[name] = max(footprint[name], amount)
				default:
					footprint[name] += amount
				}
			}
		}
		for name, amount := range footprint {
			if capacity, ok := p.Capacities[name]; !ok || amount > capacity {
				return fail(u.ID, "%s footprint %d exceeds capacity %d", name, amount, capacity)
			}
		}
		rows := []string{}
		for _, binding := range u.RowBindings {
			rows = append(rows, "rows/"+binding.RowID+".json")
		}
		declared := []string{}
		for _, output := range u.EvidenceOutputs {
			if !safeRelative(output) {
				return fail(u.ID, "evidence output %q is not a safe relative path", output)
			}
			if strings.HasPrefix(output, "rows/") {
				declared = append(declared, output)
			}
		}
		slices.Sort(rows)
		if !slices.Equal(declared, rows) || !slices.Contains(u.EvidenceOutputs, u.ResultPath()) {
			return fail(u.ID, "evidence outputs do not match its rows and result")
		}
		for i, output := range u.ArtifactOutputs {
			if !safeRelative(output.RelativePath) || output.ProducerIdentity != u.ID {
				return fail(u.ID, "artifact output %q is unsafe or names another producer", output.RelativePath)
			}
			for _, other := range u.ArtifactOutputs[:i] {
				if other.DestinationClass == output.DestinationClass && (strings.HasPrefix(output.RelativePath, other.RelativePath+"/") || strings.HasPrefix(other.RelativePath, output.RelativePath+"/")) {
					return fail(u.ID, "artifact outputs %q and %q overlap", other.RelativePath, output.RelativePath)
				}
			}
		}
	}
	return nil
}

func (p *Plan) validateProjections() error {
	selected, ok := p.Projections[p.Manifest.Target]
	if !ok || selected.CommandID != p.Manifest.CommandID || len(selected.UnitIDs) != len(p.Graph.Units) {
		return fmt.Errorf("%w: the target's projection must cover the whole graph", ErrPlan)
	}
	for name, projection := range p.Projections {
		if !safeRelative(name) || strings.Contains(name, "/") || !strictlySorted(projection.UnitIDs) {
			return fmt.Errorf("%w: projection %s is malformed", ErrPlan, name)
		}
		for _, id := range projection.UnitIDs {
			unit, ok := p.Graph.Units[id]
			if !ok {
				return fmt.Errorf("%w: projection %s names unknown unit %s", ErrPlan, name, id)
			}
			for _, need := range unit.Needs {
				if !slices.Contains(projection.UnitIDs, need) {
					return fmt.Errorf("%w: projection %s omits %s, which %s needs", ErrPlan, name, need, id)
				}
			}
		}
	}
	return nil
}

func (p *Plan) validateCache() error {
	units := p.Graph.Sorted()
	if len(p.Cache) != len(units) {
		return fmt.Errorf("%w: cache facts must cover every unit", ErrPlan)
	}
	if p.Manifest.CacheMode == "off" && len(p.CacheInputs) != 0 {
		return fmt.Errorf("%w: cache mode off carries cache inputs", ErrPlan)
	}
	for digest, indices := range p.CacheInputs {
		previous := -1
		for _, index := range indices {
			if index <= previous || index >= len(p.SourceInputs) {
				return fmt.Errorf("%w: cache inputs %s index outside the source snapshot", ErrPlan, digest)
			}
			previous = index
		}
	}
	for i, fact := range p.Cache {
		if fact.UnitID != units[i].ID {
			return fmt.Errorf("%w: cache facts are not in unit order", ErrPlan)
		}
		if fact.Eligible {
			if fact.Reason != nil || fact.ProfileID == nil || fact.Policy == nil || fact.InputDigest == nil || fact.DependencyStrategy == nil || fact.DependencyDigest == nil {
				return fmt.Errorf("%w: eligible cache fact for %s is incomplete", ErrPlan, fact.UnitID)
			}
			if _, ok := p.CacheInputs[*fact.DependencyDigest]; !ok {
				return fmt.Errorf("%w: cache fact for %s names unknown dependencies", ErrPlan, fact.UnitID)
			}
		} else if fact.Reason == nil || fact.InputDigest != nil || fact.DependencyStrategy != nil || fact.DependencyDigest != nil {
			return fmt.Errorf("%w: ineligible cache fact for %s must give only its reason", ErrPlan, fact.UnitID)
		}
	}
	return nil
}

func safeRelative(value string) bool {
	return value != "" && !strings.Contains(value, "\\") && !strings.HasPrefix(value, "/") && path.Clean(value) == value &&
		value != "." && value != ".." && !strings.HasPrefix(value, "../")
}

func strictlySorted(values []string) bool {
	for i := 1; i < len(values); i++ {
		if values[i-1] >= values[i] {
			return false
		}
	}
	return true
}
