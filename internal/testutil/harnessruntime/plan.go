// Package harnessruntime owns native graph preparation and execution coordination.
package harnessruntime

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

const planSchema = "cartulary.harness_runtime_plan.v6"
const maximumPlanBytes = harnesssession.MaximumPlan

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
type rowBinding struct {
	ID     string `json:"row_id"`
	Runner string `json:"runner"`
}
type unit struct {
	PostgresIsolation string                      `json:"postgres_isolation"`
	Renderer          string                      `json:"renderer_requirement"`
	Physical          map[string]map[string]int64 `json:"physical_allocations"`
	ID                string                      `json:"unit_id"`
	Owner             string                      `json:"owner_id"`
	Kind              string                      `json:"kind"`
	Command           command                     `json:"command"`
	Needs             []string                    `json:"needs"`
	Claims            map[string]int64            `json:"resource_claims"`
	SharedLocks       []string                    `json:"shared_locks"`
	ExclusiveLocks    []string                    `json:"exclusive_locks"`
	Affinity          string                      `json:"affinity_key"`
	Fixture           string                      `json:"fixture_lease"`
	FixtureProfile    string                      `json:"fixture_profile_id"`
	SnapshotKey       string                      `json:"snapshot_key"`
	Services          []string                    `json:"service_dependencies"`
	CachePolicy       string                      `json:"cache_policy"`
	TimeoutMS         int64                       `json:"timeout_ms"`
	Evidence          []string                    `json:"current_run_evidence_outputs"`
	Rows              []rowBinding                `json:"row_bindings"`
	Artifacts         []artifact                  `json:"reusable_artifact_outputs"`
	FailurePolicy     failurePolicy               `json:"failure_policy"`
	EstimateMS        int64                       `json:"estimated_work_ms"`
	Digest            string                      `json:"semantic_digest"`
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
type inputEnvironment struct {
	Clear   []string          `json:"clear"`
	Forward map[string]string `json:"forward"`
}
type planWire struct {
	BrowserLaneLimit int                   `json:"browser_lane_limit"`
	InputEnvironment inputEnvironment      `json:"input_environment"`
	SourceInputs     []sourceInput         `json:"source_inputs"`
	CacheInputs      map[string][]int      `json:"cache_inputs"`
	Graph            workGraph             `json:"graph"`
	Manifest         map[string]any        `json:"manifest"`
	Projections      map[string]projection `json:"projections"`
	Capacities       map[string]int64      `json:"capacities"`
	AgingMS          int64                 `json:"aging_quantum_ms"`
	Bindings         map[string]string     `json:"bindings"`
	Cache            []cacheFact           `json:"cache"`
	HelperDigest     string                `json:"helper_digest"`
	Preparation      map[string]any        `json:"preparation"`
}

// Plan can only be constructed through strict bounded validation. Its facts are
// private so a consumer cannot modify validated graph closure after preparation.
type Plan struct {
	facts   planWire
	context map[string]any
}

func (p *Plan) GraphDigest() string  { return p.facts.Graph.Digest }
func (p *Plan) SourceDigest() string { return p.facts.Manifest["source_digest"].(string) }
func (p *Plan) UnitCount() int       { return len(p.facts.Graph.Units) }

// The compiler projects typed public input forwarding. Private per-unit
// arguments and topology bindings are applied afterward by their owning paths.
func (p *Plan) unitEnvironment(parent map[string]string) map[string]string {
	child := make(map[string]string, len(parent)+len(p.facts.InputEnvironment.Forward))
	maps.Copy(child, parent)
	for _, name := range p.facts.InputEnvironment.Clear {
		delete(child, name)
	}
	maps.Copy(child, p.facts.InputEnvironment.Forward)
	return child
}
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
	policy, err := harnessschema.Decode([]byte(document["instrumentation_policy_bytes"].(string)))
	if err != nil || harnessschema.Validate("cartulary.harness_instrumentation_policy.v3", policy) != nil {
		return nil, errPlan
	}
	policyDigest, err := semanticDigest(policy)
	if err != nil || policyDigest != document["manifest"].(map[string]any)["instrumentation"].(map[string]any)["policy_digest"] {
		return nil, errPlan
	}
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
	if declared, ok := facts.Manifest["declared_inputs"].(map[string]any)["BROWSER_LANES"]; ok {
		if declared != strconv.Itoa(facts.BrowserLaneLimit) {
			return nil, errPlan
		}
	}
	if !sortedUnique(facts.InputEnvironment.Clear) || !slices.Contains(facts.InputEnvironment.Clear, "CARTULARY_MAKE_INPUT_SOURCES") {
		return nil, errPlan
	}
	for name := range facts.InputEnvironment.Forward {
		if name == "CARTULARY_MAKE_INPUT_SOURCES" || !slices.Contains(facts.InputEnvironment.Clear, name) {
			return nil, errPlan
		}
	}
	if facts.Manifest["graph_digest"] != facts.Graph.Digest {
		return nil, errPlan
	}
	snapshot := facts.Manifest["capability_snapshot"].(map[string]any)
	if facts.Manifest["go_worker_budget"].(float64) > snapshot["cpu_tokens"].(float64) {
		return nil, errPlan
	}
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
	rowOwners := map[string]bool{}
	indegree := map[string]int{}
	successors := map[string][]string{}
	for i, u := range facts.Graph.Units {
		if i > 0 && facts.Graph.Units[i-1].ID >= u.ID {
			return nil, errPlan
		}
		if !sortedUnique(u.Needs) || !sortedUnique(u.Services) || !sortedUnique(u.SharedLocks) || !sortedUnique(u.ExclusiveLocks) || !sortedUnique(u.Evidence) || len(u.Claims) == 0 {
			return nil, errPlan
		}
		if (u.Renderer == "none") != (len(u.Physical["renderer"]) == 0) || (u.Fixture == "browser_stack") != (len(u.Physical["browser_stack"]) != 0) {
			return nil, errPlan
		}
		for _, service := range []string{"postgres", "object_store"} {
			if (u.Fixture != "none" && slices.Contains(u.Services, service)) != (len(u.Physical[service]) != 0) {
				return nil, errPlan
			}
		}
		combined := maps.Clone(u.Claims)
		for _, claims := range u.Physical {
			for name, amount := range claims {
				if name == "cpu" || name == "io" {
					return nil, errPlan
				}
				if name == "volume" {
					combined[name] = max(combined[name], amount)
				} else {
					combined[name] += amount
				}
			}
		}
		for name, amount := range combined {
			if amount > facts.Capacities[name] {
				return nil, errPlan
			}
		}
		outputs := []string{}
		for _, output := range u.Evidence {
			if strings.HasPrefix(output, "rows/") {
				outputs = append(outputs, output)
			}
		}
		bindings := []string{}
		for _, row := range u.Rows {
			if rowOwners[row.ID] {
				return nil, errPlan
			}
			rowOwners[row.ID] = true
			bindings = append(bindings, "rows/"+row.ID+".json")
		}
		if !slices.Equal(outputs, bindings) {
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
	if facts.Manifest["cache_mode"] == "off" && len(facts.CacheInputs) != 0 {
		return nil, errPlan
	}
	hash, err = semanticDigest(map[string]any{"schema_id": "cartulary.source_snapshot_digest.v2", "entries": document["source_inputs"]})
	if err != nil || hash != facts.Manifest["source_digest"] {
		return nil, errPlan
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
	manifestDigest, err := semanticDigest(document["manifest"])
	if err != nil {
		return nil, errPlan
	}
	context := map[string]any{"schema_id": "cartulary.harness_execution_context.v3", "manifest_digest": manifestDigest}
	for _, key := range []string{"graph", "projections", "source_inputs", "host_profile", "instrumentation_policy_bytes"} {
		context[key] = document[key]
	}
	return &Plan{facts: facts, context: context}, nil
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
