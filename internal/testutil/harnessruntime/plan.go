// Package harnessruntime owns native graph preparation and execution coordination.
package harnessruntime

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

const planSchema = "cartulary.harness_runtime_plan.v1"
const maximumPlanBytes = 32 * 1024 * 1024

var errPlan = errors.New("invalid private harness runtime plan")

type command struct {
	Executable  string            `json:"executable"`
	Args        []string          `json:"args"`
	Environment map[string]string `json:"environment"`
}
type artifact struct {
	Type        string `json:"artifact_type"`
	Path        string `json:"relative_path"`
	Destination string `json:"destination_class"`
	Mode        string `json:"mode"`
	Producer    string `json:"producer_identity"`
}
type failurePolicy struct {
	BlockDescendants    bool   `json:"block_descendants"`
	ContinueIndependent bool   `json:"continue_independent"`
	AggregateEffect     string `json:"aggregate_effect"`
}
type unit struct {
	ID             string           `json:"unit_id"`
	Owner          string           `json:"owner_id"`
	Kind           string           `json:"kind"`
	Command        command          `json:"command"`
	Needs          []string         `json:"needs"`
	Claims         map[string]int64 `json:"resource_claims"`
	SharedLocks    []string         `json:"shared_locks"`
	ExclusiveLocks []string         `json:"exclusive_locks"`
	Affinity       string           `json:"affinity_key"`
	Fixture        string           `json:"fixture_lease"`
	FixtureProfile string           `json:"fixture_profile_id"`
	SnapshotKey    string           `json:"snapshot_key"`
	Services       []string         `json:"service_dependencies"`
	CachePolicy    string           `json:"cache_policy"`
	TimeoutMS      int64            `json:"timeout_ms"`
	Evidence       []string         `json:"current_run_evidence_outputs"`
	Artifacts      []artifact       `json:"reusable_artifact_outputs"`
	FailurePolicy  failurePolicy    `json:"failure_policy"`
	EstimateMS     int64            `json:"estimated_work_ms"`
	Digest         string           `json:"semantic_digest"`
}
type workGraph struct {
	Schema   string `json:"schema_id"`
	Selector string `json:"selector"`
	Units    []unit `json:"units"`
	Digest   string `json:"graph_digest"`
}
type projection struct {
	CommandID string   `json:"command_id"`
	Units     []string `json:"unit_ids"`
}
type sourceInput struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	Mode   string `json:"mode"`
	Digest string `json:"byte_digest"`
}
type cacheFact struct {
	DependencyDigest *string `json:"dependency_digest"`
	UnitID           string  `json:"unit_id"`
	Eligible         bool    `json:"eligible"`
	Reason           *string `json:"reason"`
	Profile          *string `json:"profile_id"`
	Policy           *string `json:"policy"`
	InputDigest      *string `json:"input_digest"`
	Strategy         *string `json:"dependency_strategy"`
	Vulnerability    *string `json:"vulnerability_database_revision"`
}
type planWire struct {
	SourceInputs []sourceInput         `json:"source_inputs"`
	CacheInputs  map[string][]int      `json:"cache_inputs"`
	Graph        workGraph             `json:"graph"`
	Manifest     map[string]any        `json:"manifest"`
	Projections  map[string]projection `json:"projections"`
	Capacities   map[string]int64      `json:"capacities"`
	AgingMS      int64                 `json:"aging_quantum_ms"`
	Bindings     map[string]string     `json:"bindings"`
	Cache        []cacheFact           `json:"cache"`
	HelperDigest string                `json:"helper_digest"`
	Preparation  map[string]any        `json:"preparation"`
}

// Plan can only be constructed through strict bounded validation. Its facts are
// private so a consumer cannot modify validated graph closure after preparation.
type Plan struct{ facts planWire }

func (p *Plan) GraphDigest() string  { return p.facts.Graph.Digest }
func (p *Plan) SourceDigest() string { return p.facts.Manifest["source_digest"].(string) }
func (p *Plan) UnitCount() int       { return len(p.facts.Graph.Units) }
func semanticDigest(value any) (string, error) {
	b, err := canonicaljson.Marshal(value)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(b)), nil
}
func bytesDigest(data []byte) string { return fmt.Sprintf("sha256:%x", sha256.Sum256(data)) }
func safeRelative(value string) bool {
	return value != "" && !strings.Contains(value, "\\") && !strings.HasPrefix(value, "/") && path.Clean(value) == value && value != "." && value != ".." && !strings.HasPrefix(value, "../")
}
func sortedUnique(values []string) bool {
	for i, v := range values {
		if i > 0 && values[i-1] >= v {
			return false
		}
	}
	return true
}
func decodePlan(data []byte) (*Plan, error) {
	if len(data) == 0 || len(data) > maximumPlanBytes {
		return nil, errPlan
	}
	value, err := harnessschema.Decode(data)
	if err != nil {
		return nil, errPlan
	}
	if harnessschema.Validate(planSchema, value) != nil {
		return nil, errPlan
	}
	document := value.(map[string]any)
	graph := document["graph"].(map[string]any)
	hash, err := semanticDigest(map[string]any{"units": graph["units"]})
	if err != nil || hash != graph["graph_digest"] {
		return nil, errPlan
	}
	for _, raw := range graph["units"].([]any) {
		fields := raw.(map[string]any)
		expected := fields["semantic_digest"]
		delete(fields, "semantic_digest")
		hash, err := semanticDigest(fields)
		fields["semantic_digest"] = expected
		if err != nil || hash != expected {
			return nil, errPlan
		}
	}
	var facts planWire
	if json.Unmarshal(data, &facts) != nil {
		return nil, errPlan
	}
	if facts.Manifest["graph_digest"] != facts.Graph.Digest {
		return nil, errPlan
	}
	snapshot := facts.Manifest["capability_snapshot"].(map[string]any)
	snapshotDigest := snapshot["snapshot_digest"]
	delete(snapshot, "snapshot_digest")
	hash, err = semanticDigest(snapshot)
	snapshot["snapshot_digest"] = snapshotDigest
	if err != nil || hash != snapshotDigest || facts.Manifest["system_digest"] != snapshotDigest {
		return nil, errPlan
	}
	if !validCapacities(snapshot, facts.Capacities) {
		return nil, errPlan
	}
	ids := make(map[string]unit, len(facts.Graph.Units))
	indegree := map[string]int{}
	successors := map[string][]string{}
	for i, u := range facts.Graph.Units {
		if i > 0 && facts.Graph.Units[i-1].ID >= u.ID {
			return nil, errPlan
		}
		if !sortedUnique(u.Needs) || !sortedUnique(u.Services) || !sortedUnique(u.SharedLocks) || !sortedUnique(u.ExclusiveLocks) || !sortedUnique(u.Evidence) || len(u.Claims) == 0 {
			return nil, errPlan
		}
		if slices.Contains(u.SharedLocks, "host_activity") == slices.Contains(u.ExclusiveLocks, "host_activity") {
			return nil, errPlan
		}
		for _, lock := range u.SharedLocks {
			if slices.Contains(u.ExclusiveLocks, lock) {
				return nil, errPlan
			}
		}
		for name, claim := range u.Claims {
			capacity, ok := facts.Capacities[name]
			if !ok || claim > capacity {
				return nil, errPlan
			}
		}
		for _, output := range u.Evidence {
			if !safeRelative(output) {
				return nil, errPlan
			}
		}
		for j, output := range u.Artifacts {
			if !safeRelative(output.Path) || output.Producer != u.ID || j > 0 && u.Artifacts[j-1].Path >= output.Path {
				return nil, errPlan
			}
			for _, other := range u.Artifacts[:j] {
				if output.Destination == other.Destination && (strings.HasPrefix(output.Path, other.Path+"/") || strings.HasPrefix(other.Path, output.Path+"/")) {
					return nil, errPlan
				}
			}
		}
		ids[u.ID] = u
		indegree[u.ID] = len(u.Needs)
	}
	for _, u := range facts.Graph.Units {
		for _, dep := range u.Needs {
			if _, ok := ids[dep]; !ok {
				return nil, errPlan
			}
			successors[dep] = append(successors[dep], u.ID)
		}
	}
	queue := []string{}
	for _, u := range facts.Graph.Units {
		if indegree[u.ID] == 0 {
			queue = append(queue, u.ID)
		}
	}
	for i := 0; i < len(queue); i++ {
		for _, next := range successors[queue[i]] {
			indegree[next]--
			if indegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}
	if len(queue) != len(ids) {
		return nil, errPlan
	}
	target := facts.Manifest["target"].(string)
	selected, ok := facts.Projections[target]
	if !ok || selected.CommandID != facts.Manifest["command_id"] || len(selected.Units) != len(ids) {
		return nil, errPlan
	}
	for name, p := range facts.Projections {
		if !safeRelative(name) || strings.Contains(name, "/") || !sortedUnique(p.Units) {
			return nil, errPlan
		}
		for _, id := range p.Units {
			u, ok := ids[id]
			if !ok {
				return nil, errPlan
			}
			for _, dep := range u.Needs {
				if !slices.Contains(p.Units, dep) {
					return nil, errPlan
				}
			}
		}
	}
	if facts.Manifest["cache_mode"] == "off" {
		if len(facts.SourceInputs) != 0 || len(facts.CacheInputs) != 0 {
			return nil, errPlan
		}
	} else {
		hash, err := semanticDigest(map[string]any{"schema_id": "cartulary.source_snapshot_digest.v2", "entries": document["source_inputs"]})
		if err != nil || hash != facts.Manifest["source_digest"] {
			return nil, errPlan
		}
	}
	for i, input := range facts.SourceInputs {
		if !safeRelative(input.Path) || i > 0 && facts.SourceInputs[i-1].Path >= input.Path {
			return nil, errPlan
		}
	}
	for _, indices := range facts.CacheInputs {
		previous := -1
		for _, index := range indices {
			if index <= previous || index >= len(facts.SourceInputs) {
				return nil, errPlan
			}
			previous = index
		}
	}
	if len(facts.Cache) != len(ids) {
		return nil, errPlan
	}
	for i, c := range facts.Cache {
		if c.UnitID != facts.Graph.Units[i].ID {
			return nil, errPlan
		}
		if c.Eligible {
			if c.Reason != nil || c.Profile == nil || c.Policy == nil || c.InputDigest == nil || c.Strategy == nil || c.DependencyDigest == nil {
				return nil, errPlan
			}
			if _, ok := facts.CacheInputs[*c.DependencyDigest]; !ok {
				return nil, errPlan
			}
		} else if c.Reason == nil || c.InputDigest != nil || c.Strategy != nil || c.DependencyDigest != nil {
			return nil, errPlan
		}
	}
	return &Plan{facts: facts}, nil
}
func validCapacities(snapshot map[string]any, actual map[string]int64) bool {
	n := func(k string) int64 { return int64(snapshot[k].(float64)) }
	services := snapshot["services"].(map[string]any)
	service := func(k string, n int64) int64 {
		if services[k] == false {
			return 0
		}
		return n
	}
	volume := int64(0)
	if snapshot["writable_volume"] == true {
		volume = 1
	}
	expected := map[string]int64{"cpu": n("cpu_tokens"), "memory_mb": max(1, n("memory_bytes")/1048576), "process": n("process_slots"), "io": n("io_tokens"), "postgres": service("postgres", n("postgres_lanes")), "object_store": service("object_store", n("object_store_lanes")), "port_lane": n("port_lanes"), "browser_stack": service("browser", n("port_lanes")), "service_stack": service("service_stack", n("port_lanes")), "volume": volume}
	if len(expected) != len(actual) {
		return false
	}
	for k, v := range expected {
		if got, ok := actual[k]; !ok || got != v {
			return false
		}
	}
	return true
}
