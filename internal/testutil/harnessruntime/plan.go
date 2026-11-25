// Package harnessruntime owns native graph preparation and execution coordination.
package harnessruntime

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessrows"
)

var errPlan = errors.New("invalid harness runtime plan")

// planDocument is the run-root file the planner's document is published as.
const planDocument = "plan.json"

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
	ID        string   `json:"row_id"`
	Runner    string   `json:"runner"`
	Selectors []string `json:"selectors,omitempty"`
}
type unit struct {
	PostgresIsolation string           `json:"postgres_isolation"`
	Renderer          string           `json:"renderer_requirement"`
	ID                string           `json:"unit_id"`
	Owner             string           `json:"owner_id"`
	Kind              string           `json:"kind"`
	Command           command          `json:"command"`
	Needs             []string         `json:"needs"`
	Claims            map[string]int64 `json:"resource_claims"`
	SharedLocks       []string         `json:"shared_locks"`
	ExclusiveLocks    []string         `json:"exclusive_locks"`
	Affinity          string           `json:"affinity_key"`
	Fixture           string           `json:"fixture_lease"`
	FixtureProfile    string           `json:"fixture_profile_id"`
	SnapshotKey       string           `json:"snapshot_key"`
	Services          []string         `json:"service_dependencies"`
	CachePolicy       string           `json:"cache_policy"`
	TimeoutMS         int64            `json:"timeout_ms"`
	Evidence          []string         `json:"current_run_evidence_outputs"`
	Rows              []rowBinding     `json:"row_bindings"`
	Artifacts         []artifact       `json:"reusable_artifact_outputs"`
	FailurePolicy     failurePolicy    `json:"failure_policy"`
	EstimateMS        int64            `json:"estimated_work_ms"`
	Digest            string           `json:"semantic_digest"`
	// Residents are the physical allocations the unit holds, by the keys
	// its plan names.
	Residents []harnessfixtures.Resident `json:"-"`
}

// goRows are the rows of a Go unit, which the runtime runs and settles
// itself (TH-HARNESS-REQ-872); nil for every other unit.
func (u unit) goRows() []harnessrows.Row {
	if len(u.Rows) == 0 || u.Command.Executable != "go" {
		return nil
	}
	rows := make([]harnessrows.Row, len(u.Rows))
	for i, binding := range u.Rows {
		if binding.Runner != "go" || len(binding.Selectors) == 0 {
			return nil
		}
		rows[i] = harnessrows.Row{ID: binding.ID, Selectors: binding.Selectors}
	}
	return rows
}

// quiet reports a unit that holds host_activity in exclusive mode: a quiet
// measurement session or another fixed-profile quiet operation.
func (u unit) quiet() bool { return slices.Contains(u.ExclusiveLocks, "host_activity") }

type workGraph struct {
	Schema string `json:"schema_id"`
	Units  []unit `json:"units"`
	Digest string `json:"graph_digest"`
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
}

// Plan is a planner document bound to the runtime's execution model. Its
// facts are private so a consumer cannot modify the graph after preparation.
type Plan struct {
	facts planWire
	// source is the planner's document, published as plan.json.
	source *harnessplan.Plan
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
