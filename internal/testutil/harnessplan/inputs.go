// Package harnessplan builds the harness work plan from authored inputs.
//
// It replaces the JavaScript work-graph compiler (work-graph overhaul,
// Stage 1). The planner is pure: it reads the authored owner model, execution
// topology, scheduler resource registry and owner catalog, validates the
// invariants planning depends on, and turns a selection into a plan. Schema
// validation of authored inputs stays with the source checks; the planner
// decodes only the fields it uses.
package harnessplan

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/pgschema"
)

// Authored input locations, relative to the repository root.
const (
	ownerPath           = "tools/harness_work_graph_owner.json"
	topologyPath        = "tools/execution_topology_manifest.json"
	registryPath        = "tools/scheduler_resource_registry.json"
	catalogRegistryPath = "tools/test_catalog_owner.json"
	taskSurfacePath     = "tools/task_surface_owner.json"
	cacheRegistryPath   = "tools/harness_cache_registry.json"
	browserBatchPath    = "tools/browser_e2e_batch_manifest.json"
	fixtureBuilderPath  = "tools/performance_fixture_builder_policy.json"
	maxInputBytes       = 16 << 20
)

// Inputs is the validated authored model a plan is built from.
type Inputs struct {
	Owner       Owner
	Topology    Topology
	Registry    ResourceRegistry
	Catalog     Catalog
	TaskSurface TaskSurface
	Cache       CacheRegistry
	Browser     BrowserBatch
	Fixtures    FixtureBuilderPolicy
	// MigrationDigest identifies the migrated PostgreSQL template every
	// populated fixture snapshot is built from. It is the digest the harness
	// and services binaries embed and verify, not one read under root.
	MigrationDigest string
	// rowIndex finds a catalog row by ID; rowFixtures holds each populated
	// fixture row's resolved snapshot.
	rowIndex    map[string]int
	rowFixtures map[string]*rowFixture
}

// FixtureBuilderPolicy says how each active populated fixture profile's
// snapshot is built.
type FixtureBuilderPolicy struct {
	Builders []FixtureBuilder `json:"builders"`
}

type FixtureBuilder struct {
	FixtureProfileID    string   `json:"fixture_profile_id"`
	RuntimeProfileID    string   `json:"runtime_profile_id"`
	ResourceProfileID   string   `json:"resource_profile_id"`
	FixtureCapability   string   `json:"fixture_capability"`
	ServiceDependencies []string `json:"service_dependencies"`
	PostgresIsolation   string   `json:"postgres_isolation"`
}

// ErrInput reports an authored input that cannot be read or violates a
// planning invariant. Wrapped errors name the file and the violated rule.
var ErrInput = errors.New("invalid harness planning input")

// Load reads and validates every authored planning input under root.
func Load(root string) (*Inputs, error) {
	inputs := &Inputs{}
	if err := readDocument(root, ownerPath, "cartulary.harness_work_graph_owner.v12", &inputs.Owner); err != nil {
		return nil, err
	}
	if err := readDocument(root, topologyPath, "cartulary.execution_topology.v16", &inputs.Topology); err != nil {
		return nil, err
	}
	if err := readDocument(root, registryPath, "cartulary.scheduler_resource_registry.v9", &inputs.Registry); err != nil {
		return nil, err
	}
	if err := readDocument(root, taskSurfacePath, "cartulary.task_surface_owner.v4", &inputs.TaskSurface); err != nil {
		return nil, err
	}
	inputs.TaskSurface.projectBrowserLanes(inputs.Topology.BrowserLanePolicy)
	if err := readDocument(root, cacheRegistryPath, "cartulary.harness_cache_registry.v3", &inputs.Cache); err != nil {
		return nil, err
	}
	if err := readDocument(root, browserBatchPath, "cartulary.browser_e2e_batch_manifest.v16", &inputs.Browser); err != nil {
		return nil, err
	}
	if err := readDocument(root, fixtureBuilderPath, "cartulary.performance_fixture_builder_policy.v2", &inputs.Fixtures); err != nil {
		return nil, err
	}
	migrations, err := pgschema.Hash()
	if err != nil {
		return nil, fmt.Errorf("%w: migration digest: %v", ErrInput, err)
	}
	inputs.MigrationDigest = migrations
	catalog, err := loadCatalog(root)
	if err != nil {
		return nil, err
	}
	inputs.Catalog = catalog
	inputs.rowIndex = make(map[string]int, len(catalog.Rows))
	for i, row := range catalog.Rows {
		inputs.rowIndex[row.RowID] = i
	}
	if err := inputs.validate(); err != nil {
		return nil, err
	}
	return inputs, nil
}

// readDocument decodes one authored JSON document strictly (no duplicate
// keys, valid Unicode), checks its schema identity and fills into.
func readDocument(root, relative, schemaID string, into any) error {
	data, err := readBounded(filepath.Join(root, relative))
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrInput, relative, err)
	}
	document, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return fmt.Errorf("%w: %s: %v", ErrInput, relative, err)
	}
	object, ok := document.(map[string]any)
	if !ok || object["schema_id"] != schemaID {
		return fmt.Errorf("%w: %s: schema_id must be %s", ErrInput, relative, schemaID)
	}
	if err := json.Unmarshal(data, into); err != nil {
		return fmt.Errorf("%w: %s: %v", ErrInput, relative, err)
	}
	return nil
}

func readBounded(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > maxInputBytes {
		return nil, errors.New("not a bounded regular file")
	}
	data := make([]byte, info.Size())
	if _, err := file.ReadAt(data, 0); err != nil && info.Size() > 0 {
		return nil, err
	}
	return data, nil
}

// Owner is the planning part of the harness work-graph owner model.
type Owner struct {
	TierOrder               []string                  `json:"tier_order"`
	AggregateTiers          map[string]string         `json:"aggregate_tiers"`
	AggregatePolicyInherits map[string][]string       `json:"aggregate_policy_inherits"`
	AggregatePolicyRoots    map[string][]string       `json:"aggregate_policy_roots"`
	TargetMembers           map[string][]TargetMember `json:"target_members"`
	// ExplicitSelections are the private selections that run their
	// explicit-only rows.
	ExplicitSelections  []string              `json:"explicit_selections"`
	PolicyUnits         map[string]PolicyUnit `json:"policy_units"`
	Operations          map[string]Operation  `json:"operations"`
	EvidenceEstimatesMS map[string]int64      `json:"evidence_estimates_ms"`
	// LifecycleEstimatesMS charges browser allocation, reset and finalization
	// work at their lifecycle boundaries.
	LifecycleEstimatesMS map[string]int64               `json:"lifecycle_estimates_ms"`
	DefaultTimeoutMS     int64                          `json:"default_timeout_ms"`
	GoUnitTargetMS       int64                          `json:"go_cohort_packing_bound_ms"`
	GoHeavyUnitTargetMS  int64                          `json:"postgres_heavy_packing_target_ms"`
	AgingQuantumMS       int64                          `json:"aging_quantum_ms"`
	RowExecutionContexts map[string]RowExecutionContext `json:"row_execution_contexts"`
	InvocationPolicies   map[string]InvocationPolicy    `json:"invocation_policies"`
}

// PolicyUnit is the scheduling policy of one operation that runs as a unit.
type PolicyUnit struct {
	Needs                   []string         `json:"needs"`
	CachePolicy             string           `json:"cache_policy"`
	EstimatedWorkMS         int64            `json:"estimated_work_ms"`
	ResourceProfileID       string           `json:"resource_profile_id"`
	ServiceDependencies     []string         `json:"service_dependencies"`
	PostgresIsolation       string           `json:"postgres_isolation"`
	TimeoutMS               int64            `json:"timeout_ms,omitempty"`
	RequiredRows            []RequiredRow    `json:"required_rows,omitempty"`
	FixtureLease            string           `json:"fixture_lease,omitempty"`
	ReusableArtifactOutputs []ArtifactOutput `json:"reusable_artifact_outputs,omitempty"`
	// EvidenceOutputs are the run-root-relative outputs the operation
	// publishes beyond its row and unit results.
	EvidenceOutputs []string `json:"current_run_evidence_outputs,omitempty"`
}

// RequiredRow is a catalog row an operation requires whatever the aggregate
// tier, with the reason it is required.
type RequiredRow struct {
	RowID     string `json:"row_id"`
	Rationale string `json:"rationale"`
}

// ArtifactOutput is a reusable artifact an operation publishes.
type ArtifactOutput struct {
	ArtifactType     string `json:"artifact_type"`
	RelativePath     string `json:"relative_path"`
	DestinationClass string `json:"destination_class"`
	Mode             string `json:"mode"`
	ProducerIdentity string `json:"producer_identity"`
}

// Operation is an authored harness operation. Its execution definition is
// typed when command construction lands (overhaul milestone 2).
type Operation struct {
	Kind      string          `json:"kind"`
	ReportKey string          `json:"report_key"`
	Execution json.RawMessage `json:"execution"`
}

// TargetMember is one member of a private target selection: an operation or
// union name, an exact row, a runner catalog slice, or an operation ID.
type TargetMember struct {
	Name        string         `json:"-"`
	RowID       string         `json:"row_id,omitempty"`
	OperationID string         `json:"operation_id,omitempty"`
	Catalog     *CatalogMember `json:"catalog,omitempty"`
}

// CatalogMember selects catalog rows by runner and optional stages or suite.
type CatalogMember struct {
	Runner  string   `json:"runner"`
	Stages  []string `json:"stages,omitempty"`
	GoSuite string   `json:"go_suite,omitempty"`
}

// UnmarshalJSON accepts the two authored member forms: a bare name or an
// object naming exactly one of row, operation or catalog slice.
func (m *TargetMember) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err == nil {
		*m = TargetMember{Name: name}
		return nil
	}
	type plain TargetMember
	var member plain
	if err := json.Unmarshal(data, &member); err != nil {
		return err
	}
	forms := 0
	for _, set := range []bool{member.RowID != "", member.OperationID != "", member.Catalog != nil} {
		if set {
			forms++
		}
	}
	if forms != 1 {
		return errors.New("target member must name exactly one of row_id, operation_id or catalog")
	}
	*m = TargetMember(member)
	return nil
}

// RowExecutionContext fixes an explicit execution context for one row.
type RowExecutionContext struct {
	CPUTokens   int64  `json:"cpu_tokens"`
	CachePolicy string `json:"cache_policy"`
	GoRace      bool   `json:"go_race,omitempty"`
}

// InvocationPolicy is a runner's fixed per-invocation cost and claims.
type InvocationPolicy struct {
	StartupMS         int64            `json:"startup_ms"`
	CleanupMS         int64            `json:"cleanup_ms"`
	CoordinatorClaims map[string]int64 `json:"coordinator_claims,omitempty"`
	// UnitTargetMS sizes a runner's units (Vitest; Go units use the owner's
	// targets).
	UnitTargetMS int64 `json:"unit_target_ms,omitempty"`
	// PreparationOperations are operations every unit of the runner waits for.
	PreparationOperations []string `json:"preparation_operations,omitempty"`
}

// Topology is the planning part of the execution topology.
type Topology struct {
	ResourceProfiles        []ResourceProfile           `json:"resource_profiles"`
	RuntimeProfiles         []RuntimeProfile            `json:"runtime_profiles"`
	ServiceResourceMinimums map[string]map[string]int64 `json:"service_resource_minimums"`
	PhysicalResourceClaims  map[string]map[string]int64 `json:"physical_resource_claims"`
	RuntimeBinaries         []RuntimeBinary             `json:"runtime_binaries"`
	BrowserLanePolicy       BrowserLanePolicy           `json:"browser_lane_policy"`
	BrowserBatch            struct {
		FrontendArtifacts map[string]struct {
			ProducerOperation string `json:"producer_operation"`
		} `json:"frontend_artifacts"`
	} `json:"browser_e2e_batch"`
	BrowserResetPolicy map[string]int64 `json:"browser_reset_policy"`
}

// BrowserLanePolicy bounds the browser lanes a command may request.
type BrowserLanePolicy struct {
	Default int64 `json:"default"`
	Maximum int64 `json:"maximum"`
}

// ResourceProfile is a named set of active claims.
type ResourceProfile struct {
	ID              string           `json:"id"`
	ResourceClaims  map[string]int64 `json:"resource_claims"`
	RunnerTimeoutMS int64            `json:"runner_timeout_ms,omitempty"`
}

// RuntimeProfile is a named application runtime configuration.
type RuntimeProfile struct {
	ID                string   `json:"id"`
	ManagedServiceIDs []string `json:"managed_service_ids"`
	BrowserCapable    bool     `json:"browser_capable"`
}

// RuntimeBinary is a built binary a row may consume, and its producer.
type RuntimeBinary struct {
	ID                 string `json:"id"`
	OutputMakeVariable string `json:"output_make_variable"`
	ConsumerEnv        string `json:"consumer_env"`
	DefaultOutputPath  string `json:"default_output_path"`
	ProducerOperation  string `json:"producer_operation"`
}

// ResourceRegistry holds the authored capacity policies.
type ResourceRegistry struct {
	CapacityPolicies map[string]CapacityPolicy `json:"capacity_policies"`
}

// CapacityPolicy is one resource's authored capacity rule.
type CapacityPolicy struct {
	Source                string `json:"source"`
	Minimum               int64  `json:"minimum"`
	Default               int64  `json:"default"`
	SafetyMarginPercent   int64  `json:"safety_margin_percent"`
	CPUMultiplier         int64  `json:"cpu_multiplier"`
	MaximumCPUMultiplier  int64  `json:"maximum_cpu_multiplier"`
	FallbackCPUMultiplier int64  `json:"fallback_cpu_multiplier"`
}

// CacheRegistry binds operations and runners to cache profiles.
type CacheRegistry struct {
	Profiles []CacheProfile `json:"profiles"`
}

// CacheProfile is one cache policy: what it keys on and which work it covers.
type CacheProfile struct {
	ProfileID          string      `json:"profile_id"`
	Policy             string      `json:"policy"`
	DependencyStrategy string      `json:"dependency_strategy"`
	InputRoots         []string    `json:"input_roots"`
	UnitInputs         *UnitInputs `json:"unit_inputs"`
	// RequiresVulnerabilityDatabaseRevision makes reuse depend on a proven
	// vulnerability-database revision.
	RequiresVulnerabilityDatabaseRevision bool     `json:"requires_vulnerability_database_revision"`
	OperationIDs                          []string `json:"operation_ids"`
	RunnerTypes                           []string `json:"runner_types"`
}

// UnitInputs are the inputs every unit-aware closure contains: paths with
// everything beneath them, and path prefixes.
type UnitInputs struct {
	Paths    []string `json:"paths"`
	Prefixes []string `json:"prefixes"`
}

// BrowserBatch is the generated browser batch manifest: compatibility
// partitions of Playwright rows and the stages that report them.
type BrowserBatch struct {
	InvocationPolicy struct {
		StartupMS int64 `json:"startup_ms"`
		CleanupMS int64 `json:"cleanup_ms"`
	} `json:"invocation_policy"`
	Partitions []BrowserPartition `json:"partitions"`
	Stages     []BrowserStage     `json:"stages"`
}

// BrowserPartition is a set of Playwright rows that can share a browser
// stack: one execution stage, frontend artifact, profiles and services.
type BrowserPartition struct {
	Name                  string   `json:"name"`
	Target                string   `json:"target"`
	Kind                  string   `json:"kind"`
	ExecutionStage        string   `json:"execution_stage"`
	FrontendArtifactID    string   `json:"frontend_artifact_id"`
	SchedulerNeeds        []string `json:"scheduler_needs"`
	StateScope            string   `json:"state_scope"`
	RuntimeProfileID      string   `json:"runtime_profile_id"`
	ResourceProfileID     string   `json:"resource_profile_id"`
	FixtureCapability     string   `json:"fixture_capability"`
	ServiceRequirement    string   `json:"service_requirement"`
	ServiceDependencies   []string `json:"service_dependencies"`
	PostgresIsolation     string   `json:"postgres_isolation"`
	PreparationOperations []string `json:"preparation_operations"`
	FixtureProfileID      string   `json:"fixture_profile_id"`
	BrowserSessionGroup   string   `json:"browser_session_group"`
	SelectedRowIDs        []string `json:"selected_row_ids"`
}

// BrowserStage reports a set of partitions' rows under one target.
type BrowserStage struct {
	Name            string   `json:"name"`
	Target          string   `json:"target"`
	SummaryChildren []string `json:"summary_children"`
	Groups          []struct {
		Partition      string   `json:"partition"`
		SelectedRowIDs []string `json:"selected_row_ids"`
	} `json:"groups"`
}
