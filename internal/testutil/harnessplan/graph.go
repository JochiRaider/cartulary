package harnessplan

import (
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// GraphSchema is the work-graph document the planner writes.
const GraphSchema = "cartulary.harness_work_graph.v9"

// Graph is a compiled work graph: units keyed by ID, every need resolved
// inside the graph.
type Graph struct {
	Units map[string]*Unit
	// Digest identifies the graph: the semantic digest of its sorted units.
	Digest string
}

// Sorted returns the units in ID order.
func (g *Graph) Sorted() []*Unit {
	units := make([]*Unit, 0, len(g.Units))
	for _, unit := range g.Units {
		units = append(units, unit)
	}
	slices.SortFunc(units, func(a, b *Unit) int { return cmp.Compare(a.ID, b.ID) })
	return units
}

type graphDocument struct {
	SchemaID    string  `json:"schema_id"`
	Units       []*Unit `json:"units"`
	GraphDigest string  `json:"graph_digest"`
}

// MarshalJSON writes the graph v9 document.
func (g *Graph) MarshalJSON() ([]byte, error) {
	return json.Marshal(graphDocument{SchemaID: GraphSchema, Units: g.Sorted(), GraphDigest: g.Digest})
}

// validate checks the graph's closure: every need names a unit in the graph,
// the needs are acyclic, and no row is bound to two units.
func (g *Graph) validate() error {
	rows := map[string]string{}
	for id, unit := range g.Units {
		for _, need := range unit.Needs {
			if _, ok := g.Units[need]; !ok {
				return fmt.Errorf("%w: %s needs missing unit %s", ErrSelection, id, need)
			}
		}
		for _, binding := range unit.RowBindings {
			if other, ok := rows[binding.RowID]; ok && other != id {
				return fmt.Errorf("%w: row %s is bound to %s and %s", ErrSelection, binding.RowID, other, id)
			}
			rows[binding.RowID] = id
		}
	}
	state := map[string]int{}
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("%w: dependency cycle through %s", ErrSelection, id)
		case 2:
			return nil
		}
		state[id] = 1
		for _, need := range g.Units[id].Needs {
			if err := visit(need); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range g.Units {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

// finalize puts every unit in canonical form, adds the evidence each unit
// always publishes, and computes the unit and graph digests.
func (g *Graph) finalize() {
	for _, unit := range g.Units {
		unit.canonicalize()
		unit.SemanticDigest = semanticDigest(unit, "semantic_digest")
	}
	g.Digest = semanticDigest(map[string]any{"units": g.Sorted()}, "")
}

// Unit is one schedulable piece of work (graph v9 unit).
type Unit struct {
	ID                  string           `json:"unit_id"`
	OwnerID             string           `json:"owner_id"`
	Kind                string           `json:"kind"`
	Command             Command          `json:"command"`
	Needs               []string         `json:"needs"`
	Claims              map[string]int64 `json:"resource_claims"`
	Residents           []Resident       `json:"residents"`
	SharedLocks         []string         `json:"shared_locks"`
	ExclusiveLocks      []string         `json:"exclusive_locks"`
	FixtureLease        string           `json:"fixture_lease"`
	ServiceDependencies []string         `json:"service_dependencies"`
	PostgresIsolation   string           `json:"postgres_isolation"`
	CachePolicy         string           `json:"cache_policy"`
	TimeoutMS           int64            `json:"timeout_ms"`
	EvidenceOutputs     []string         `json:"current_run_evidence_outputs"`
	RowBindings         []RowBinding     `json:"row_bindings"`
	ArtifactOutputs     []ArtifactOutput `json:"reusable_artifact_outputs"`
	RuntimeBinaryIDs    []string         `json:"runtime_binary_ids"`
	FailurePolicy       FailurePolicy    `json:"failure_policy"`
	EstimatedWorkMS     int64            `json:"estimated_work_ms"`
	// RendererRequirement is a browser group's renderer: pooled or dedicated.
	RendererRequirement string `json:"renderer_requirement,omitempty"`
	// AffinityKey names the browser stack a group reuses across cohorts.
	AffinityKey    string `json:"affinity_key,omitempty"`
	SemanticDigest string `json:"semantic_digest"`
}

// unsafeResultCharacters are replaced to name a unit's result file.
var unsafeResultCharacters = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

// ResultPath is the run-root-relative path of the unit's result.
func (u *Unit) ResultPath() string {
	return "unit-results/" + unsafeResultCharacters.ReplaceAllString(u.ID, "-") + ".json"
}

// canonicalize sorts every set-valued field, replaces absent lists with empty
// ones and adds the unit result to the evidence outputs.
func (u *Unit) canonicalize() {
	sets := []*[]string{&u.Needs, &u.SharedLocks, &u.ExclusiveLocks, &u.ServiceDependencies, &u.RuntimeBinaryIDs}
	u.EvidenceOutputs = append(u.EvidenceOutputs, u.ResultPath())
	sets = append(sets, &u.EvidenceOutputs)
	for _, set := range sets {
		*set = sortedCopy(*set)
	}
	if u.Command.Args == nil {
		u.Command.Args = []string{}
	}
	if u.Command.Environment == nil {
		u.Command.Environment = map[string]string{}
	}
	if u.Claims == nil {
		u.Claims = map[string]int64{}
	}
	u.Residents = slices.Clone(u.Residents)
	if u.Residents == nil {
		u.Residents = []Resident{}
	}
	slices.SortFunc(u.Residents, func(a, b Resident) int { return cmp.Compare(a.Key, b.Key) })
	u.RowBindings = slices.Clone(u.RowBindings)
	if u.RowBindings == nil {
		u.RowBindings = []RowBinding{}
	}
	slices.SortFunc(u.RowBindings, func(a, b RowBinding) int { return cmp.Compare(a.RowID, b.RowID) })
	u.ArtifactOutputs = slices.Clone(u.ArtifactOutputs)
	if u.ArtifactOutputs == nil {
		u.ArtifactOutputs = []ArtifactOutput{}
	}
	slices.SortFunc(u.ArtifactOutputs, func(a, b ArtifactOutput) int { return cmp.Compare(a.RelativePath, b.RelativePath) })
}

// Command is what the runtime launches for a unit. Runtime binary paths are
// not part of it: the unit lists the binaries it consumes and the runtime
// binds their paths at launch.
type Command struct {
	Executable  string            `json:"executable"`
	Args        []string          `json:"args"`
	Environment map[string]string `json:"environment"`
}

// Resident is a physical allocation that outlives a single process: a
// PostgreSQL generation or the shared object store. Its key identifies the
// allocation wherever it appears; its scope fixes its lifetime.
type Resident struct {
	Key    string           `json:"key"`
	Scope  string           `json:"scope"`
	Claims map[string]int64 `json:"claims"`
}

// Resident scopes.
const (
	// ScopeUnit: created for one unit and retired with it.
	ScopeUnit = "unit"
	// ScopeGraph: shared by consumers in the graph and retired after the last.
	ScopeGraph = "graph"
	// ScopeAttached: held by a persistent service session the run borrows;
	// the run neither creates nor charges it.
	ScopeAttached = "attached"
)

// RowBinding binds a catalog row to the unit that produces its evidence.
type RowBinding struct {
	RowID  string `json:"row_id"`
	Runner string `json:"runner"`
}

// FailurePolicy says whether a unit's failure blocks the units that need it.
// Independent work always continues.
type FailurePolicy struct {
	BlockDescendants bool `json:"block_descendants"`
}

// canonicalSum is the SHA-256 of value's canonical JSON, as hex, with the
// named top-level field removed first. The planner's own types always
// marshal, so a failure is a programming error.
func canonicalSum(value any, omit string) string {
	sum := sha256.Sum256(canonicalBytes(value, omit))
	return hex.EncodeToString(sum[:])
}

// canonicalBytes is value's canonical JSON with the named top-level field
// removed.
func canonicalBytes(value any, omit string) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		panic(err)
	}
	if object, ok := document.(map[string]any); ok && omit != "" {
		delete(object, omit)
	}
	canonical, err := canonicaljson.Marshal(document)
	if err != nil {
		panic(err)
	}
	return canonical
}

// digest is the hex SHA-256 of value's canonical JSON, used inside IDs.
func digest(value any) string { return canonicalSum(value, "") }

// semanticDigest is the harness digest form ("sha256:<hex>") of value's
// canonical JSON without the omitted field.
func semanticDigest(value any, omit string) string { return "sha256:" + canonicalSum(value, omit) }
