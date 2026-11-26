package harnessplan

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/testutil/browserbinding"
	"github.com/JochiRaider/cartulary/internal/testutil/runcontext"
)

const (
	browserOrdinary      = "browser_ordinary"
	browserQuiet         = "browser_measurement_quiet"
	browserGroupCLI      = "tools/harness/browser/browser-catalog-group-cli.mjs"
	browserFinalizerCLI  = "tools/harness/browser/browser-target-finalizer-cli.mjs"
	measurementCLI       = "tools/harness/browser/browser-measurement-finalizer-cli.mjs"
	snapshotBuilderCLI   = "tools/harness/performance-fixture/snapshot-builder-cli.mjs"
	visualUpdateTarget   = "browser-e2e-visual-update"
	snapshotUpdateMode   = "snapshot_update"
	browserLifecycleUnit = "harness.browser"
	// BrowserLaneResource is the graph-local capacity of concurrently running
	// ordinary browser groups: the BROWSER_LANES limit. Only the graph
	// scheduler charges it; host admission has no such resource.
	BrowserLaneResource = "browser_lane"
)

// browserCohort is one planned browser group: rows of one partition with the
// same needs, run by one browser group invocation on one stack.
type browserCohort struct {
	id        string
	name      string
	partition *BrowserPartition
	rows      []*Row
	needs     []string // targets the rows wait for
	execution string   // the partition and needs the cohort shares a stack with
	work      int64    // startup, selected work and cleanup
	affinity  string
	stack     string // the browser stack the cohort holds alone
	estimate  int64
	fixture   *rowFixture // the populated snapshot a measurement cohort clones
}

// planBrowser places the reached Playwright rows into cohorts and lanes. A
// row's partition fixes what it can share a stack with; its stage reports it.
func (c *compiler) planBrowser(rows []*Row) error {
	if len(rows) == 0 {
		return nil
	}
	partitions := map[string]*BrowserPartition{}
	for i := range c.in.Browser.Partitions {
		partition := &c.in.Browser.Partitions[i]
		for _, id := range partition.SelectedRowIDs {
			partitions[id] = partition
		}
	}
	groups := map[string][]*Row{}
	needs := map[string][]string{}
	for _, row := range rows {
		partition, ok := partitions[row.RowID]
		if !ok {
			return fmt.Errorf("%w: Playwright row %s has no browser partition", ErrInput, row.RowID)
		}
		rowNeeds := c.browserNeeds(row, partition)
		key := partition.Name + "\x00" + strings.Join(rowNeeds, ",")
		groups[key] = append(groups[key], row)
		needs[key] = rowNeeds
	}
	var cohorts []*browserCohort
	for _, key := range slices.Sorted(maps.Keys(groups)) {
		partition := partitions[groups[key][0].RowID]
		execution := digest(map[string]any{"partition": partition.Name, "needs": needs[key]})[:16]
		for _, bin := range c.packBrowser(partition, groups[key]) {
			cohort := &browserCohort{partition: partition, rows: bin.rows, needs: needs[key], execution: execution, work: bin.work}
			rowIDs := make([]string, len(bin.rows))
			for i, row := range bin.rows {
				rowIDs[i] = row.RowID
			}
			slices.Sort(rowIDs)
			identity := map[string]any{"partition": partition.Name, "rows": rowIDs, "mode": c.options.BrowserMode}
			if partition.FixtureProfileID != "" {
				cohort.fixture = c.in.rowFixtures[rowIDs[0]]
				identity["snapshot_key"] = cohort.fixture.key
			}
			cohort.name = "cohort-" + digest(identity)
			cohort.id = "browser_group:" + partition.ExecutionStage + ":" + cohort.name
			cohorts = append(cohorts, cohort)
		}
	}
	c.stackBrowser(cohorts)
	for _, cohort := range cohorts {
		c.browser[cohort.id] = cohort
		for _, row := range cohort.rows {
			c.assigned[row.RowID] = cohort.id
		}
	}
	return nil
}

// browserNeeds are the targets a Playwright row waits for: its partition's
// stage needs and preparations, the frontend artifact its stage serves, the
// Playwright invocation policy's preparations, and its own.
func (c *compiler) browserNeeds(row *Row, partition *BrowserPartition) []string {
	targets := slices.Concat(partition.SchedulerNeeds, partition.PreparationOperations, row.PreparationOperations,
		c.in.Owner.InvocationPolicies[RunnerPlaywright].PreparationOperations)
	if artifact, ok := c.in.Topology.BrowserBatch.FrontendArtifacts[partition.FrontendArtifactID]; ok {
		targets = append(targets, artifact.ProducerOperation)
	}
	for _, binary := range c.in.Topology.RuntimeBinaries {
		if slices.Contains(row.RuntimeBinaryIDs, binary.ID) {
			targets = append(targets, binary.ProducerOperation)
		}
	}
	return dedupe(targets)
}

type browserBin struct {
	rows []*Row
	work int64
}

// packBrowser divides a partition group into cohorts. Rows outside
// test-owned state run alone, as does each populated fixture row, whose
// predicate gets its own snapshot clone and quiet session. Otherwise the work
// is split into as many balanced cohorts as fit the resource profile's
// timeout less allocation headroom, and ordinary cohorts are spread over the
// browser lanes while each cohort's work still covers one stack acquisition.
func (c *compiler) packBrowser(partition *BrowserPartition, rows []*Row) []browserBin {
	policy := c.in.Browser.InvocationPolicy
	overhead := policy.StartupMS + policy.CleanupMS
	lifecycle := c.in.Owner.LifecycleEstimatesMS
	envelope := c.timeout(partition.ResourceProfileID) - max(lifecycle["browser_allocation_acquire"], lifecycle["browser_reset"]) - lifecycle["browser_allocation_release"] - overhead
	// Cohorts are sized to the Playwright unit target, well inside the
	// timeout envelope, so accurate estimates leave timeout headroom.
	if target := c.in.Owner.InvocationPolicies[RunnerPlaywright].UnitTargetMS; target > 0 {
		envelope = min(envelope, target)
	}
	ordered := slices.Clone(rows)
	cost := func(row *Row) int64 { return c.in.rowEstimate(*row) }
	slices.SortFunc(ordered, func(a, b *Row) int { return cmp.Or(cmp.Compare(cost(b), cost(a)), strings.Compare(a.RowID, b.RowID)) })
	total := int64(0)
	for _, row := range ordered {
		total += cost(row)
	}
	count := max(1, int((total+envelope-1)/max(1, envelope)))
	switch {
	case partition.StateScope != "test_owned" || partition.FixtureProfileID != "":
		count = len(ordered)
	case partition.ResourceProfileID == browserOrdinary:
		useful := max(1, int(total/max(1, lifecycle["browser_allocation_acquire"])))
		count = max(count, min(c.options.BrowserLanes, len(ordered), useful))
	}
	for ; ; count++ {
		count = min(count, len(ordered))
		bins := make([]browserBin, count)
		for _, row := range ordered {
			least := 0
			for i := range bins {
				if bins[i].work < bins[least].work {
					least = i
				}
			}
			bins[least].rows = append(bins[least].rows, row)
			bins[least].work += cost(row)
		}
		fits := !slices.ContainsFunc(bins, func(bin browserBin) bool { return len(bin.rows) > 1 && bin.work > envelope })
		if fits || count == len(ordered) {
			for i := range bins {
				bins[i].work += overhead
			}
			return bins
		}
	}
}

// timeout is a resource profile's runner timeout, or the owner's default.
func (c *compiler) timeout(profileID string) int64 {
	index := slices.IndexFunc(c.in.Topology.ResourceProfiles, func(p ResourceProfile) bool { return p.ID == profileID })
	if index >= 0 && c.in.Topology.ResourceProfiles[index].RunnerTimeoutMS > 0 {
		return c.in.Topology.ResourceProfiles[index].RunnerTimeoutMS
	}
	return c.in.Owner.DefaultTimeoutMS
}

// stackBrowser gives every cohort its own browser stack, created for it and
// retired with it. Ordinary cohorts are not placed on lanes: each claims one
// browser lane, so the scheduler runs at most the lane limit of them at once
// and starts the next ready cohort, by rank, whenever one finishes
// (TH-HARNESS-REQ-805). A fresh stack costs one acquisition; reusing a stack
// between cohorts saved too little to keep a plan-time chain for it.
func (c *compiler) stackBrowser(cohorts []*browserCohort) {
	lifecycle := c.in.Owner.LifecycleEstimatesMS
	for _, cohort := range cohorts {
		cohort.affinity = safeIdentifier(cohort.execution + "-" + cohort.name)
		cohort.stack = "stack:" + digest(map[string]any{"affinity": cohort.affinity, "stage": cohort.partition.ExecutionStage,
			"runtime_profile": cohort.partition.RuntimeProfileID, "services": cohort.partition.ServiceDependencies})
		cohort.estimate = cohort.work + lifecycle["browser_allocation_acquire"] + lifecycle["browser_allocation_release"]
	}
}

var unsafeIdentifier = regexp.MustCompile(`[^A-Za-z0-9_.+-]+`)

func safeIdentifier(value string) string { return unsafeIdentifier.ReplaceAllString(value, "-") }

// browserTarget is the target a cohort reports under: its partition's, or
// the snapshot update command when updating snapshots.
func (c *compiler) browserTarget(partition *BrowserPartition) string {
	if c.options.BrowserMode == snapshotUpdateMode {
		return visualUpdateTarget
	}
	return partition.Target
}

// browserUnit builds a planned browser group unit.
func (c *compiler) browserUnit(id string, stack []string) (string, error) {
	if _, ok := c.units[id]; ok {
		return id, nil
	}
	cohort := c.browser[id]
	partition := cohort.partition
	var needs []string
	for _, target := range cohort.needs {
		ends, err := c.target(target, stack)
		if err != nil {
			return "", err
		}
		needs = append(needs, ends...)
	}
	if cohort.fixture != nil {
		builder, err := c.fixtureBuilder(cohort.fixture, stack)
		if err != nil {
			return "", err
		}
		needs = append(needs, builder)
	}
	claims, err := c.claims(partition.ResourceProfileID, partition.ServiceDependencies)
	if err != nil {
		return "", err
	}
	reset := c.in.Topology.BrowserResetPolicy
	preparation := int64(0)
	for name, value := range reset {
		if strings.HasSuffix(name, "_timeout_ms") {
			preparation += value
		}
	}
	target := c.browserTarget(partition)
	rowIDs := make([]string, len(cohort.rows))
	for i, row := range cohort.rows {
		rowIDs[i] = row.RowID
	}
	slices.Sort(rowIDs)
	environment := browserbinding.Binding{
		Stage: partition.ExecutionStage, RuntimeProfile: partition.RuntimeProfileID, ResourceProfile: partition.ResourceProfileID,
		ServiceRequirement: partition.ServiceRequirement, SessionContract: partition.BrowserSessionGroup, StackKey: cohort.stack,
		PreparationLabel: cohort.name, PreparationTimeoutMS: strconv.FormatInt(preparation, 10),
		ResetDrainTimeoutMS: strconv.FormatInt(reset["backend_drain_timeout_ms"], 10),
	}.Environment()
	environment["BROWSER_E2E_BATCH_MANIFEST"] = browserBatchPath
	maps.Copy(environment, runcontext.Services(partition.ServiceDependencies))
	runcontext.Context{Target: target}.Apply(environment)
	residents, cluster := c.browserResidents(cohort)
	if cluster != "" {
		environment[clusterKeyEnvironment] = cluster
	}
	if fixture := cohort.fixture; fixture != nil {
		browserbinding.Binding{FixtureProfile: fixture.profile.FixtureProfileID, SnapshotKey: fixture.key, SnapshotBuilderUnit: fixture.builderID,
			FixtureRow: rowIDs[0], FixturePredicate: fixture.predicate}.Apply(environment)
	}
	if c.options.BrowserMode == snapshotUpdateMode {
		environment["CARTULARY_BROWSER_MAINTENANCE_MODE"] = snapshotUpdateMode
		environment["CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS"] = "1"
	}
	shared, exclusive := []string{"host_activity"}, []string(nil)
	renderer := "pooled"
	if partition.ResourceProfileID == browserQuiet {
		shared, exclusive, renderer = nil, []string{"host_activity"}, "dedicated"
	} else {
		claims[BrowserLaneResource] = 1
	}
	if partition.Kind == "visual" {
		shared = append(shared, "browser_snapshots")
	}
	var binaries []string
	for _, row := range cohort.rows {
		binaries = append(binaries, row.RuntimeBinaryIDs...)
	}
	unit := &Unit{
		ID: id, OwnerID: browserLifecycleUnit, Kind: "runner", Needs: dedupe(needs), Claims: claims,
		Residents: residents, SharedLocks: shared, ExclusiveLocks: exclusive,
		FixtureLease: partition.FixtureCapability, ServiceDependencies: sortedCopy(partition.ServiceDependencies),
		PostgresIsolation: partition.PostgresIsolation, CachePolicy: "none", TimeoutMS: c.timeout(partition.ResourceProfileID),
		EvidenceOutputs:  []string{target + "/browser-groups/" + safeIdentifier(cohort.name) + "/browser-group-result.json"},
		RuntimeBinaryIDs: sortedCopy(binaries), FailurePolicy: FailurePolicy{BlockDescendants: false},
		EstimatedWorkMS: cohort.estimate, RendererRequirement: renderer, AffinityKey: cohort.affinity,
		Command: Command{Executable: "node", Args: []string{browserGroupCLI, "--manifest", browserBatchPath, "--stage", partition.ExecutionStage,
			"--partition", partition.Name, "--group", cohort.name, "--rows", strings.Join(rowIDs, ",")}, Environment: environment},
	}
	if cohort.fixture != nil {
		unit.FixtureProfileID, unit.SnapshotKey = cohort.fixture.profile.FixtureProfileID, cohort.fixture.key
	}
	for _, row := range rowIDs {
		unit.RowBindings = append(unit.RowBindings, RowBinding{RowID: row, Runner: RunnerPlaywright})
		unit.EvidenceOutputs = append(unit.EvidenceOutputs, "rows/"+row+".json")
	}
	c.units[id] = unit
	return id, nil
}

// browserResidents are a browser group's stack, the shared object store, and
// its PostgreSQL cluster, which it returns: the cluster of its stack, shared
// by the cohorts of one affinity, or the cluster holding its populated
// fixture's snapshot.
func (c *compiler) browserResidents(cohort *browserCohort) ([]Resident, string) {
	physical := c.in.Topology.PhysicalResourceClaims
	residents := []Resident{{Key: cohort.stack, Scope: ScopeUnit, Claims: maps.Clone(physical["browser_stack"])}}
	shared := ScopeGraph
	if c.options.AttachedServices {
		shared = ScopeAttached
	}
	if slices.Contains(cohort.partition.ServiceDependencies, "object_store") {
		residents = append(residents, Resident{Key: "object_store:shared", Scope: shared, Claims: maps.Clone(physical["object_store"])})
	}
	if !slices.Contains(cohort.partition.ServiceDependencies, "postgres") {
		return residents, ""
	}
	if cohort.fixture != nil {
		return append(residents, Resident{Key: cohort.fixture.cluster, Scope: ScopeGraph, Claims: maps.Clone(physical["postgres"])}), cohort.fixture.cluster
	}
	cluster := "postgres:" + cohort.stack
	return append(residents, Resident{Key: cluster, Scope: ScopeUnit, Claims: maps.Clone(physical["postgres"])}), cluster
}

// fixtureBuilder builds the one unit that constructs and seals a populated
// fixture snapshot, shared by every measurement group cloning it. It holds
// the snapshot's cluster, which stays resident until the last clone, and
// shares host activity, so it finishes before any dependent quiet session.
func (c *compiler) fixtureBuilder(fixture *rowFixture, stack []string) (string, error) {
	id := fixture.builderID
	if _, ok := c.units[id]; ok {
		return id, nil
	}
	builder := fixture.builder
	var needs []string
	if len(builder.ServiceDependencies) != 0 {
		ends, err := c.target("harness.test_service_images", stack)
		if err != nil {
			return "", err
		}
		needs = ends
	}
	claims, err := c.claims(builder.ResourceProfileID, builder.ServiceDependencies)
	if err != nil {
		return "", err
	}
	profile := fixture.profile.FixtureProfileID
	c.units[id] = &Unit{
		ID: id, OwnerID: browserLifecycleUnit, Kind: "fixture_builder", Needs: needs, Claims: claims,
		Residents:   []Resident{{Key: fixture.cluster, Scope: ScopeGraph, Claims: maps.Clone(c.in.Topology.PhysicalResourceClaims["postgres"])}},
		SharedLocks: []string{"host_activity"}, FixtureLease: builder.FixtureCapability, ServiceDependencies: sortedCopy(builder.ServiceDependencies),
		PostgresIsolation: builder.PostgresIsolation, CachePolicy: "none", TimeoutMS: c.timeout(builder.ResourceProfileID),
		EvidenceOutputs: []string{"performance-fixtures/" + fixture.key + "/build-diagnostics.json", "performance-fixtures/" + fixture.key + "/snapshot-build.json"},
		FailurePolicy:   FailurePolicy{BlockDescendants: true}, EstimatedWorkMS: c.in.Owner.LifecycleEstimatesMS["performance_fixture_prepare"],
		FixtureProfileID: profile, SnapshotKey: fixture.key,
		Command: Command{Executable: "node", Args: []string{snapshotBuilderCLI, "--fixture-profile", profile, "--snapshot-key", fixture.key},
			Environment: browserbinding.Binding{RuntimeProfile: builder.RuntimeProfileID, FixtureProfile: profile, SnapshotKey: fixture.key, SnapshotBuilderUnit: id}.Apply(map[string]string{
				"CARTULARY_FIXTURE_BUILDER_RESOURCE_PROFILE_ID": builder.ResourceProfileID,
				"CARTULARY_FIXTURE_MIGRATION_DIGEST":            c.in.MigrationDigest,
				"CARTULARY_FIXTURE_SOURCE_CONTRACT_DIGEST":      fixture.profile.SourceContractDigest,
				clusterKeyEnvironment:                           fixture.cluster,
			})},
	}
	return id, nil
}

// measurementSummary builds the finalizer that summarizes one measurement
// group's predicate against its snapshot. It runs even when the group fails,
// since a browser group does not block its descendants.
func (c *compiler) measurementSummary(cohort *browserCohort, groupUnit, target string) (string, error) {
	fixture := cohort.fixture
	id := "browser_measurement_summary:" + cohort.partition.ExecutionStage + ":" + cohort.name
	claims, err := c.claims("standard", nil)
	if err != nil {
		return "", err
	}
	schema := fixture.profile.ArtifactPolicy.SummarySchemaID
	c.units[id] = &Unit{
		ID: id, OwnerID: browserLifecycleUnit, Kind: "finalizer", Needs: []string{groupUnit}, Claims: claims,
		SharedLocks: []string{"host_activity"}, FixtureLease: "none", PostgresIsolation: "none", CachePolicy: "none",
		TimeoutMS:       c.in.Owner.DefaultTimeoutMS,
		EvidenceOutputs: []string{target + "/browser-groups/" + safeIdentifier(cohort.name) + "/frontend-measurement-summary." + schema[strings.LastIndex(schema, ".")+1:] + ".json"},
		FailurePolicy:   FailurePolicy{BlockDescendants: true}, EstimatedWorkMS: c.in.Owner.LifecycleEstimatesMS["measurement_finalize"],
		FixtureProfileID: fixture.profile.FixtureProfileID, SnapshotKey: fixture.key,
		Command: Command{Executable: "node", Args: []string{measurementCLI, "--target", target, "--stage", cohort.partition.ExecutionStage,
			"--group", cohort.name, "--row", cohort.rows[0].RowID, "--predicate", fixture.predicate}},
	}
	return id, nil
}

// browserFinalizers add one target finalizer for every stage that reports
// selected rows; it waits for every browser group holding those rows.
func (c *compiler) browserFinalizers() error {
	for _, stage := range c.in.Browser.Stages {
		var groups, groupTargets, groupUnits, needs []string
		quiet := true
		visual, measured := false, false
		for _, membership := range stage.Groups {
			for _, row := range membership.SelectedRowIDs {
				id, ok := c.assigned[row]
				if _, planned := c.browser[id]; !ok || !planned {
					continue
				}
				if _, built := c.units[id]; !built {
					continue
				}
				cohort := c.browser[id]
				if slices.Contains(needs, id) {
					continue
				}
				needs = append(needs, id)
				groups = append(groups, cohort.name)
				groupTargets = append(groupTargets, cohort.name+"="+c.browserTarget(cohort.partition))
				groupUnits = append(groupUnits, cohort.name+"="+id)
				quiet = quiet && cohort.partition.ResourceProfileID == browserQuiet
				visual = visual || cohort.partition.Kind == "visual"
				if cohort.fixture != nil {
					summary, err := c.measurementSummary(cohort, id, c.browserTarget(cohort.partition))
					if err != nil {
						return err
					}
					needs = append(needs, summary)
					measured = true
				}
			}
		}
		if len(needs) == 0 {
			continue
		}
		target := stage.Target
		if c.options.BrowserMode == snapshotUpdateMode {
			target = visualUpdateTarget
		}
		slices.Sort(groups)
		slices.Sort(groupTargets)
		slices.Sort(groupUnits)
		args := []string{browserFinalizerCLI, "--target", target}
		for _, group := range groups {
			args = append(args, "--groups", group)
		}
		for _, entry := range groupTargets {
			args = append(args, "--group-targets", entry)
		}
		children := ""
		if target == stage.Target {
			children = strings.Join(stage.SummaryChildren, ",")
		}
		args = append(args, "--children", children)
		for _, entry := range groupUnits {
			args = append(args, "--group-units", entry)
		}
		claims, err := c.claims("standard", nil)
		if err != nil {
			return err
		}
		shared, exclusive := []string{"host_activity"}, []string(nil)
		if quiet {
			shared, exclusive = nil, []string{"host_activity"}
		}
		evidence := []string{target + "/browser-target-result.json"}
		if visual {
			evidence = append(evidence, target+"/frontend-visual-reconciliation.json")
		}
		if measured {
			evidence = append(evidence, target+"/frontend-measurement-aggregate.json")
		}
		id := "browser_target_summary:" + safeIdentifier(target)
		c.units[id] = &Unit{
			ID: id, OwnerID: browserLifecycleUnit, Kind: "finalizer", Needs: dedupe(needs), Claims: claims,
			SharedLocks: shared, ExclusiveLocks: exclusive, FixtureLease: "none", PostgresIsolation: "none", CachePolicy: "none",
			TimeoutMS: c.in.Owner.DefaultTimeoutMS, EvidenceOutputs: evidence, FailurePolicy: FailurePolicy{BlockDescendants: true},
			EstimatedWorkMS: c.in.Owner.LifecycleEstimatesMS["browser_target_finalize"],
			Command:         Command{Executable: "node", Args: args, Environment: map[string]string{}},
		}
	}
	return nil
}
