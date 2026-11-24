package harnessplan

import (
	"cmp"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

const (
	browserOrdinary      = "browser_ordinary"
	browserQuiet         = "browser_measurement_quiet"
	browserGroupCLI      = "tools/harness/browser/browser-catalog-group-cli.mjs"
	browserFinalizerCLI  = "tools/harness/browser/browser-target-finalizer-cli.mjs"
	visualUpdateTarget   = "browser-e2e-visual-update"
	snapshotUpdateMode   = "snapshot_update"
	browserLanePrefix    = "browser-"
	browserLifecycleUnit = "harness.browser"
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
	lane      string
	previous  string // the unit that runs before it on its lane
	affinity  string
	stack     string // the browser stack resident the cohort holds
	estimate  int64
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
		if partition.FixtureProfileID != "" {
			return fmt.Errorf("%w: measurement browser row %s", ErrUnsupported, row.RowID)
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
			cohort.name = "cohort-" + digest(map[string]any{"partition": partition.Name, "rows": rowIDs, "mode": c.options.BrowserMode})
			cohort.id = "browser_group:" + partition.ExecutionStage + ":" + cohort.name
			cohorts = append(cohorts, cohort)
		}
	}
	c.laneBrowser(cohorts)
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
// test-owned state run alone. Otherwise the work is split into as many
// balanced cohorts as fit the resource profile's timeout less allocation
// headroom, and ordinary cohorts are spread over the browser lanes while
// each cohort's work still covers one stack acquisition.
func (c *compiler) packBrowser(partition *BrowserPartition, rows []*Row) []browserBin {
	policy := c.in.Browser.InvocationPolicy
	overhead := policy.StartupMS + policy.CleanupMS
	lifecycle := c.in.Owner.LifecycleEstimatesMS
	envelope := c.browserTimeout(partition) - max(lifecycle["browser_allocation_acquire"], lifecycle["browser_reset"]) - lifecycle["browser_allocation_release"] - overhead
	ordered := slices.Clone(rows)
	cost := func(row *Row) int64 {
		if row.EstimatedWorkMS > 0 {
			return row.EstimatedWorkMS
		}
		return c.in.Owner.EvidenceEstimatesMS[row.EvidenceClass] * int64(len(row.Selector.Titles))
	}
	slices.SortFunc(ordered, func(a, b *Row) int { return cmp.Or(cmp.Compare(cost(b), cost(a)), strings.Compare(a.RowID, b.RowID)) })
	total := int64(0)
	for _, row := range ordered {
		total += cost(row)
	}
	count := max(1, int((total+envelope-1)/max(1, envelope)))
	switch {
	case partition.StateScope != "test_owned":
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

func (c *compiler) browserTimeout(partition *BrowserPartition) int64 {
	index := slices.IndexFunc(c.in.Topology.ResourceProfiles, func(p ResourceProfile) bool { return p.ID == partition.ResourceProfileID })
	if index >= 0 && c.in.Topology.ResourceProfiles[index].RunnerTimeoutMS > 0 {
		return c.in.Topology.ResourceProfiles[index].RunnerTimeoutMS
	}
	return c.in.Owner.DefaultTimeoutMS
}

// laneBrowser spreads ordinary cohorts over the browser lanes, longest first
// into the least-loaded lane; a cohort waits for the one before it on its
// lane, so at most the lane limit of ordinary browser groups run at once.
// Consecutive cohorts of one execution on a lane reuse its stack.
func (c *compiler) laneBrowser(cohorts []*browserCohort) {
	var ordinary []*browserCohort
	for _, cohort := range cohorts {
		if cohort.partition.ResourceProfileID == browserOrdinary {
			ordinary = append(ordinary, cohort)
		}
	}
	slices.SortFunc(ordinary, func(a, b *browserCohort) int {
		return cmp.Or(cmp.Compare(b.work, a.work), strings.Compare(a.name, b.name))
	})
	lanes := min(c.options.BrowserLanes, len(ordinary))
	loads, last := make([]int64, lanes), make([]string, lanes)
	for _, cohort := range ordinary {
		least := 0
		for i := range loads {
			if loads[i] < loads[least] {
				least = i
			}
		}
		cohort.lane = browserLanePrefix + strconv.Itoa(least+1)
		cohort.previous = last[least]
		loads[least] += cohort.work
		last[least] = cohort.id
	}
	lifecycle := c.in.Owner.LifecycleEstimatesMS
	finalUse := map[string]*browserCohort{}
	order := slices.Concat(ordinary, slices.DeleteFunc(slices.Clone(cohorts), func(cohort *browserCohort) bool {
		return cohort.partition.ResourceProfileID == browserOrdinary
	}))
	for _, cohort := range order {
		place := cohort.lane
		if place == "" {
			place = cohort.name
		}
		cohort.affinity = safeIdentifier(cohort.execution + "-" + place)
		cohort.stack = "stack:" + digest(map[string]any{"affinity": cohort.affinity, "stage": cohort.partition.ExecutionStage,
			"runtime_profile": cohort.partition.RuntimeProfileID, "services": cohort.partition.ServiceDependencies})
		cohort.estimate = cohort.work + lifecycle["browser_allocation_acquire"]
		if _, reused := finalUse[cohort.affinity]; reused {
			cohort.estimate = cohort.work + lifecycle["browser_reset"]
		}
		finalUse[cohort.affinity] = cohort
	}
	for _, cohort := range finalUse {
		cohort.estimate += lifecycle["browser_allocation_release"]
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
	if cohort.previous != "" {
		if _, err := c.browserUnit(cohort.previous, stack); err != nil {
			return "", err
		}
		needs = append(needs, cohort.previous)
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
	environment := map[string]string{
		"BROWSER_E2E_BATCH_MANIFEST":               browserBatchPath,
		"CARTULARY_BROWSER_STAGE":                  partition.ExecutionStage,
		"CARTULARY_BROWSER_RUNTIME_PROFILE_ID":     partition.RuntimeProfileID,
		"CARTULARY_BROWSER_RESOURCE_PROFILE_ID":    partition.ResourceProfileID,
		"CARTULARY_BROWSER_PREPARATION_LABEL":      cohort.name,
		"CARTULARY_BROWSER_RESET_DRAIN_TIMEOUT_MS": strconv.FormatInt(reset["backend_drain_timeout_ms"], 10),
		"CARTULARY_BROWSER_PREPARATION_TIMEOUT_MS": strconv.FormatInt(preparation, 10),
		"CARTULARY_BROWSER_SERVICE_REQUIREMENT":    partition.ServiceRequirement,
		"CARTULARY_HARNESS_SERVICE_DEPENDENCIES":   strings.Join(partition.ServiceDependencies, ","),
		"CARTULARY_BROWSER_SESSION_CONTRACT":       partition.BrowserSessionGroup,
		"CARTULARY_BROWSER_STACK_KEY":              cohort.stack,
		"CARTULARY_TEST_TARGET":                    target,
	}
	if cohort.lane != "" {
		environment["CARTULARY_BROWSER_LANE_ID"] = cohort.lane
	}
	if c.options.BrowserMode == snapshotUpdateMode {
		environment["CARTULARY_BROWSER_MAINTENANCE_MODE"] = snapshotUpdateMode
		environment["CARTULARY_PLAYWRIGHT_UPDATE_SNAPSHOTS"] = "1"
	}
	shared, exclusive := []string{"host_activity"}, []string{"browser_session:" + cohort.affinity}
	renderer := "pooled"
	if partition.ResourceProfileID == browserQuiet {
		shared, exclusive, renderer = nil, append(exclusive, "host_activity"), "dedicated"
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
		Residents: c.browserResidents(cohort), SharedLocks: shared, ExclusiveLocks: exclusive,
		FixtureLease: partition.FixtureCapability, ServiceDependencies: sortedCopy(partition.ServiceDependencies),
		PostgresIsolation: partition.PostgresIsolation, CachePolicy: "none", TimeoutMS: c.browserTimeout(partition),
		EvidenceOutputs:  []string{target + "/browser-groups/" + safeIdentifier(cohort.name) + "/browser-group-result.json"},
		RuntimeBinaryIDs: sortedCopy(binaries), FailurePolicy: FailurePolicy{BlockDescendants: false},
		EstimatedWorkMS: cohort.estimate, RendererRequirement: renderer, AffinityKey: cohort.affinity,
		Command: Command{Executable: "node", Args: []string{browserGroupCLI, "--manifest", browserBatchPath, "--stage", partition.ExecutionStage,
			"--partition", partition.Name, "--group", cohort.name, "--rows", strings.Join(rowIDs, ",")}, Environment: environment},
	}
	for _, row := range rowIDs {
		unit.RowBindings = append(unit.RowBindings, RowBinding{RowID: row, Runner: RunnerPlaywright})
		unit.EvidenceOutputs = append(unit.EvidenceOutputs, "rows/"+row+".json")
	}
	c.units[id] = unit
	return id, nil
}

// browserResidents are a browser group's stack and the PostgreSQL cluster it
// owns, shared by the cohorts of one affinity, and the shared object store.
func (c *compiler) browserResidents(cohort *browserCohort) []Resident {
	physical := c.in.Topology.PhysicalResourceClaims
	stack := cohort.stack
	residents := []Resident{{Key: stack, Scope: ScopeGraph, Claims: maps.Clone(physical["browser_stack"])}}
	shared := ScopeGraph
	if c.options.AttachedServices {
		shared = ScopeAttached
	}
	if slices.Contains(cohort.partition.ServiceDependencies, "object_store") {
		residents = append(residents, Resident{Key: "object_store:shared", Scope: shared, Claims: maps.Clone(physical["object_store"])})
	}
	if slices.Contains(cohort.partition.ServiceDependencies, "postgres") {
		residents = append(residents, Resident{Key: "postgres:" + stack, Scope: ScopeGraph, Claims: maps.Clone(physical["postgres"])})
	}
	return residents
}

// browserFinalizers add one target finalizer for every stage that reports
// selected rows; it waits for every browser group holding those rows.
func (c *compiler) browserFinalizers() error {
	for _, stage := range c.in.Browser.Stages {
		var groups, groupTargets, groupUnits, needs []string
		quiet := true
		visual := false
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
		id := "browser_target_summary:" + safeIdentifier(target)
		c.units[id] = &Unit{
			ID: id, OwnerID: browserLifecycleUnit, Kind: "finalizer", Needs: dedupe(needs), Claims: claims,
			SharedLocks: shared, ExclusiveLocks: exclusive, FixtureLease: "none", PostgresIsolation: "none", CachePolicy: "none",
			TimeoutMS: c.in.Owner.DefaultTimeoutMS, EvidenceOutputs: evidence, FailurePolicy: FailurePolicy{BlockDescendants: true},
			EstimatedWorkMS: c.in.Owner.LifecycleEstimatesMS["browser_target_finalize"],
			Command:         Command{Executable: "node", Args: args, Environment: map[string]string{"CARTULARY_DEFER_OBSERVABILITY_FINALIZE": "1"}},
		}
	}
	return nil
}
