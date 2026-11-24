package harnessplan

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

// ErrSelection reports a selection the planner cannot compile.
var ErrSelection = errors.New("invalid harness selection")

// ErrUnsupported reports work the Go planner does not compile yet
// (overhaul milestone 4: Playwright and measurement rows).
var ErrUnsupported = errors.New("not yet supported by the Go planner")

// Selection names what to compile: catalog rows and operations or private
// target unions. Preparation resolves public commands, owner slices and
// command inputs into this form.
type Selection struct {
	Rows    []string
	Targets []string
	// Forwarded are public command inputs passed to the selected operations'
	// commands (not to the operations they depend on).
	Forwarded map[string]string
	// BrowserMode is set when the command runs browser stages in a mode of
	// its own (snapshot_update).
	BrowserMode string
}

// Options are the per-invocation facts compilation depends on.
type Options struct {
	// CPUTokens bounds a unit's CPU claim (the snapshot's cpu_tokens).
	CPUTokens int64
	// GoVersion is the pinned toolchain version (tools/toolchain_pins.json).
	GoVersion string
	// GoShuffleSeed, when set, orders Go tests with that seed.
	GoShuffleSeed string
	// AttachedServices is set when units borrow a persistent local service
	// session: its shared allocations are held by that session, not the run.
	AttachedServices bool
	// BrowserLanes bounds how many ordinary browser groups run at once; zero
	// takes the topology's default.
	BrowserLanes int
	// BrowserMode is the selection's browser mode.
	BrowserMode string
}

// Compile turns a selection into a work graph.
func Compile(in *Inputs, selection Selection, options Options) (*Graph, error) {
	if options.CPUTokens < 1 || options.GoVersion == "" {
		return nil, fmt.Errorf("%w: compilation needs CPU tokens and a Go version", ErrSelection)
	}
	if options.BrowserLanes < 1 {
		options.BrowserLanes = int(in.Topology.BrowserLanePolicy.Default)
	}
	options.BrowserMode = selection.BrowserMode
	c := &compiler{in: in, options: options, units: map[string]*Unit{}, targets: map[string][]string{}, rows: map[string]*Row{}, shellRows: map[string][]*Row{},
		selected: map[string]bool{}, forwarded: selection.Forwarded, assigned: map[string]string{}, planned: map[string]plannedUnit{}, browser: map[string]*browserCohort{}}
	for i := range in.Catalog.Rows {
		row := &in.Catalog.Rows[i]
		c.rows[row.RowID] = row
		if row.Runner == RunnerShell {
			c.shellRows[row.Selector.OperationID] = append(c.shellRows[row.Selector.OperationID], row)
		}
	}
	if len(selection.Rows) == 0 && len(selection.Targets) == 0 {
		return nil, fmt.Errorf("%w: the selection is empty", ErrSelection)
	}
	for _, target := range selection.Targets {
		c.selected[target] = true
	}
	// Every Go and Vitest row the selection reaches is packed in one pass, so
	// a row reached twice (selected, required by a policy, named by a union)
	// is bound to exactly one unit.
	reached, err := c.reach(selection)
	if err != nil {
		return nil, err
	}
	if err := c.assign(reached); err != nil {
		return nil, err
	}
	for _, target := range dedupe(selection.Targets) {
		if _, err := c.target(target, nil); err != nil {
			return nil, err
		}
	}
	if len(selection.Rows) != 0 {
		if _, err := c.compileRows(selection.Rows, nil); err != nil {
			return nil, err
		}
	}
	if err := c.browserFinalizers(); err != nil {
		return nil, err
	}
	graph := &Graph{Units: c.units}
	if err := graph.validate(); err != nil {
		return nil, err
	}
	graph.finalize()
	return graph, nil
}

type compiler struct {
	in        *Inputs
	options   Options
	units     map[string]*Unit
	targets   map[string][]string // target name -> its end units
	rows      map[string]*Row
	shellRows map[string][]*Row // operation -> shell rows that report it
	selected  map[string]bool   // operations and unions named by the selection
	forwarded map[string]string // inputs forwarded to selected operations
	// assigned binds each reached Go and Vitest row to its planned unit;
	// planned holds each planned unit's contract key, contract and rows.
	assigned map[string]string
	planned  map[string]plannedUnit
	browser  map[string]*browserCohort
}

// plannedUnit is a runner unit whose rows are assigned but which is built
// only when something needs it.
type plannedUnit struct {
	contract contract
	rows     []*Row
}

// reach collects every row the selection reaches: its rows, union members,
// policies' required rows, and the rows reached through the targets that
// rows and policies wait for. It compiles nothing.
func (c *compiler) reach(selection Selection) ([]*Row, error) {
	var rows []*Row
	seenRows, seenTargets := map[string]bool{}, map[string]bool{}
	var walkTarget func(string) error
	addRow := func(id string) error {
		row, ok := c.rows[id]
		if !ok {
			return fmt.Errorf("%w: unknown row %s", ErrSelection, id)
		}
		if seenRows[id] {
			return nil
		}
		seenRows[id] = true
		if row.Runner == RunnerShell {
			return walkTarget(row.Selector.OperationID)
		}
		rows = append(rows, row)
		if row.Runner != RunnerGo && row.Runner != RunnerVitest {
			return nil
		}
		targets := append(slices.Clone(row.PreparationOperations), c.in.Owner.InvocationPolicies[row.Runner].PreparationOperations...)
		if len(row.ServiceDependencies) != 0 {
			targets = append(targets, "harness.test_service_images")
		}
		for _, binary := range c.in.Topology.RuntimeBinaries {
			if slices.Contains(row.RuntimeBinaryIDs, binary.ID) {
				targets = append(targets, binary.ProducerOperation)
			}
		}
		for _, target := range targets {
			if err := walkTarget(target); err != nil {
				return err
			}
		}
		return nil
	}
	walkPolicy := func(name string) error {
		policy, ok := c.in.Owner.PolicyUnits[name]
		if !ok {
			return nil
		}
		for _, need := range policy.Needs {
			if err := walkTarget(need); err != nil {
				return err
			}
		}
		for _, required := range policy.RequiredRows {
			if err := addRow(required.RowID); err != nil {
				return err
			}
		}
		return nil
	}
	walkTarget = func(name string) error {
		if seenTargets[name] {
			return nil
		}
		seenTargets[name] = true
		if members, ok := c.in.Owner.TargetMembers[name]; ok {
			for _, member := range members {
				var err error
				switch {
				case member.RowID != "":
					err = addRow(member.RowID)
				case member.Catalog != nil:
					for _, id := range c.in.catalogRows(*member.Catalog) {
						if err = addRow(id); err != nil {
							break
						}
					}
				case member.OperationID != "":
					err = walkTarget(member.OperationID)
				default:
					err = walkTarget(member.Name)
				}
				if err != nil {
					return err
				}
			}
			return walkPolicy(name)
		}
		if _, ok := c.in.Owner.Operations[name]; ok {
			for _, row := range c.shellRows[name] {
				seenRows[row.RowID] = true
			}
			return walkPolicy(name)
		}
		return fmt.Errorf("%w: unknown target %s", ErrSelection, name)
	}
	for _, target := range dedupe(selection.Targets) {
		if err := walkTarget(target); err != nil {
			return nil, err
		}
	}
	for _, id := range dedupe(selection.Rows) {
		if err := addRow(id); err != nil {
			return nil, err
		}
	}
	return rows, nil
}

// target compiles an operation or a private union and returns its end units:
// the units in its closure that no other unit of the closure needs.
func (c *compiler) target(name string, stack []string) ([]string, error) {
	if ends, ok := c.targets[name]; ok {
		return ends, nil
	}
	if slices.Contains(stack, name) {
		return nil, fmt.Errorf("%w: dependency cycle %s", ErrSelection, strings.Join(append(stack, name), " -> "))
	}
	stack = append(slices.Clone(stack), name)
	var ends []string
	var err error
	if members, ok := c.in.Owner.TargetMembers[name]; ok {
		ends, err = c.union(name, members, stack)
	} else if _, ok := c.in.Owner.Operations[name]; ok {
		var id string
		id, err = c.operation(name, stack)
		ends = []string{id}
	} else {
		err = fmt.Errorf("%w: unknown target %s", ErrSelection, name)
	}
	if err != nil {
		return nil, err
	}
	c.targets[name] = ends
	return ends, nil
}

// union expands a private target's members and attaches the union's own
// policy needs to the member units that have no needs inside the union.
func (c *compiler) union(name string, members []TargetMember, stack []string) ([]string, error) {
	var rows, ids []string
	for _, member := range members {
		switch {
		case member.RowID != "":
			rows = append(rows, member.RowID)
		case member.Catalog != nil:
			selected := c.in.catalogRows(*member.Catalog)
			if len(selected) == 0 {
				return nil, fmt.Errorf("%w: target %s selects no catalog rows", ErrSelection, name)
			}
			rows = append(rows, selected...)
		default:
			operation := member.Name
			if member.OperationID != "" {
				operation = member.OperationID
			}
			ends, err := c.target(operation, stack)
			if err != nil {
				return nil, err
			}
			ids = append(ids, ends...)
		}
	}
	if len(rows) != 0 {
		units, err := c.compileRows(dedupe(rows), stack)
		if err != nil {
			return nil, err
		}
		ids = append(ids, units...)
	}
	ids = dedupe(ids)
	if policy, ok := c.in.Owner.PolicyUnits[name]; ok {
		needs, err := c.needs(policy, stack)
		if err != nil {
			return nil, err
		}
		for _, id := range ids {
			if unit := c.units[id]; len(unit.Needs) == 0 {
				unit.Needs = needs
			}
		}
	}
	return ends(c.units, ids), nil
}

// catalogRows selects rows by runner and, for Go, by suite.
func (in *Inputs) catalogRows(member CatalogMember) []string {
	var rows []string
	for _, row := range in.Catalog.Rows {
		if row.Runner != member.Runner {
			continue
		}
		if len(member.Stages) != 0 && !slices.Contains(member.Stages, row.Selector.Stage) {
			continue
		}
		if member.GoSuite != "" && goSuite(row) != member.GoSuite {
			continue
		}
		rows = append(rows, row.RowID)
	}
	return rows
}

// goSuite routes a Go row to a backend suite. The catalog encodes the suite
// in the family name's last segment; integration support rows count as
// integration.
func goSuite(row Row) string {
	if row.EvidenceClass == "unit" && row.RuntimeProfileID == "none" {
		return "unit"
	}
	switch family := row.FamilyID[strings.LastIndex(row.FamilyID, ".")+1:]; family {
	case "engine", "fixtures", "support_unit", "unit":
		return "unit"
	case "storage", "store":
		return "store"
	case "process":
		return "process"
	default:
		return "integration"
	}
}

// needs compiles a policy's needs and required rows and returns the end
// units a dependent waits for.
func (c *compiler) needs(policy PolicyUnit, stack []string) ([]string, error) {
	var ids []string
	for _, need := range policy.Needs {
		ends, err := c.target(need, stack)
		if err != nil {
			return nil, err
		}
		ids = append(ids, ends...)
	}
	if len(policy.RequiredRows) != 0 {
		rows := make([]string, len(policy.RequiredRows))
		for i, required := range policy.RequiredRows {
			rows[i] = required.RowID
		}
		units, err := c.compileRows(rows, stack)
		if err != nil {
			return nil, err
		}
		ids = append(ids, units...)
	}
	return dedupe(ids), nil
}

// operation compiles one operation unit. An operation runs under its own
// policy unit; there is no default policy. Shell rows that report the
// operation bind to its unit, and the unit then publishes their evidence.
func (c *compiler) operation(id string, stack []string) (string, error) {
	unitID := "operation:" + id
	if _, ok := c.units[unitID]; ok {
		return unitID, nil
	}
	operation := c.in.Owner.Operations[id]
	policy, ok := c.in.Owner.PolicyUnits[id]
	if !ok {
		return "", fmt.Errorf("%w: operation %s has no policy unit", ErrSelection, id)
	}
	needs, err := c.needs(policy, stack)
	if err != nil {
		return "", err
	}
	claims, err := c.claims(policy.ResourceProfileID, policy.ServiceDependencies)
	if err != nil {
		return "", err
	}
	timeout := policy.TimeoutMS
	if timeout == 0 {
		timeout = c.in.Owner.DefaultTimeoutMS
	}
	lease := policy.FixtureLease
	if lease == "" {
		lease = "none"
	}
	unit := &Unit{
		ID: unitID, OwnerID: "harness.command_surface", Kind: operation.Kind, Needs: needs, Claims: claims,
		Residents: c.residents(unitID, lease, policy.ServiceDependencies, policy.PostgresIsolation), FixtureLease: lease,
		ServiceDependencies: slices.Clone(policy.ServiceDependencies), PostgresIsolation: policy.PostgresIsolation,
		CachePolicy: policy.CachePolicy, TimeoutMS: timeout, ArtifactOutputs: policy.ReusableArtifactOutputs,
		EvidenceOutputs: slices.Clone(policy.EvidenceOutputs), FailurePolicy: FailurePolicy{BlockDescendants: true},
		EstimatedWorkMS: policy.EstimatedWorkMS,
		Command: Command{Executable: "node", Args: []string{"tools/harness/execution/operation-runner-cli.mjs", id}, Environment: map[string]string{
			"CARTULARY_TEST_TARGET":                  operation.ReportKey,
			"CARTULARY_OPERATION_DIGEST":             "sha256:" + digest(operation.Execution),
			"CARTULARY_HARNESS_SERVICE_DEPENDENCIES": strings.Join(policy.ServiceDependencies, ","),
		}},
	}
	if c.selected[id] {
		for name, value := range c.forwarded {
			unit.Command.Environment[name] = value
		}
	}
	if rows := c.shellRows[id]; len(rows) != 0 {
		ids := make([]string, len(rows))
		for i, row := range rows {
			if row.ResourceProfileID != policy.ResourceProfileID {
				return "", fmt.Errorf("%w: shell row %s resource profile %s disagrees with operation %s policy %s", ErrSelection, row.RowID, row.ResourceProfileID, id, policy.ResourceProfileID)
			}
			ids[i] = row.RowID
			unit.RowBindings = append(unit.RowBindings, RowBinding{RowID: row.RowID, Runner: RunnerShell})
			unit.EvidenceOutputs = append(unit.EvidenceOutputs, "rows/"+row.RowID+".json")
		}
		if len(rows) == 1 {
			unit.OwnerID = rows[0].OwnerID
		}
		unit.Command.Args = []string{"tools/harness/execution/runners/row-runner-cli.mjs", "--row-ids", strings.Join(ids, ",")}
		unit.Command.Environment["CARTULARY_TEST_ROWS"] = strings.Join(ids, ",")
	}
	c.lock(unit, policy.ResourceProfileID)
	c.units[unitID] = unit
	return unitID, nil
}

// compileRows compiles catalog rows and returns the IDs of their units: a
// shell row's operation unit, or the planned unit a runner row is assigned to.
func (c *compiler) compileRows(ids []string, stack []string) ([]string, error) {
	var units []string
	for _, id := range dedupe(ids) {
		row, ok := c.rows[id]
		if !ok {
			return nil, fmt.Errorf("%w: unknown row %s", ErrSelection, id)
		}
		var unit string
		var err error
		planned, ok := c.assigned[id]
		switch {
		case row.Runner == RunnerShell:
			unit, err = c.operation(row.Selector.OperationID, stack)
		case ok && c.browser[planned] != nil:
			unit, err = c.browserUnit(planned, stack)
		case ok:
			unit, err = c.runnerUnit(planned, stack)
		default:
			err = fmt.Errorf("%w: row %s was compiled without being reached", ErrSelection, id)
		}
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return dedupe(units), nil
}

// assign packs the reached Go and Vitest rows into planned units and the
// reached Playwright rows into browser cohorts.
func (c *compiler) assign(rows []*Row) error {
	var runnerRows, browserRows []*Row
	for _, row := range rows {
		switch {
		case row.Runner == RunnerGo && (row.FixtureProfileID != "" || row.EvidenceClass == "measurement"):
			return fmt.Errorf("%w: measurement row %s", ErrUnsupported, row.RowID)
		case row.Runner == RunnerGo, row.Runner == RunnerVitest:
			runnerRows = append(runnerRows, row)
		case row.Runner == RunnerPlaywright:
			browserRows = append(browserRows, row)
		default:
			return fmt.Errorf("%w: %s row %s", ErrUnsupported, row.Runner, row.RowID)
		}
	}
	if err := c.planBrowser(browserRows); err != nil {
		return err
	}
	groups := map[string][]*Row{}
	contracts := map[string]contract{}
	for _, row := range runnerRows {
		contract, err := c.contract(row)
		if err != nil {
			return err
		}
		key := digest(contract)
		groups[key] = append(groups[key], row)
		contracts[key] = contract
	}
	for _, key := range slices.Sorted(maps.Keys(groups)) {
		contract := contracts[key]
		for _, bin := range c.split(c.items(groups[key], contract), contract) {
			rowIDs := make([]string, len(bin))
			for i, row := range bin {
				rowIDs[i] = row.RowID
			}
			slices.Sort(rowIDs)
			id := contract.Runner + ":" + key[:16] + ":" + digest(rowIDs)[:12]
			c.planned[id] = plannedUnit{contract: contract, rows: bin}
			for _, row := range rowIDs {
				c.assigned[row] = id
			}
		}
	}
	return nil
}

// contract is everything that must be equal for rows to share a runner
// invocation: the runner and what it runs (a Go package or a Vitest project),
// profiles, fixture, services, binaries, preparations, cache policy and CPU.
type contract struct {
	Runner                string
	Package               string
	Project               string
	RuntimeProfileID      string
	ResourceProfileID     string
	FixtureCapability     string
	PostgresIsolation     string
	ServiceDependencies   []string
	RuntimeBinaryIDs      []string
	PreparationOperations []string
	CachePolicy           string
	CPUTokens             int64
	Race                  bool
	// Alone is set for rows that need a process of their own: process
	// isolation, a managed process, or an explicit execution context.
	Alone string
}

// item is an indivisible piece of a runner invocation: a Go row, or every
// selected row of one Vitest file (the runner runs a file once).
type item struct {
	rows []*Row
	work int64
}

func (c *compiler) contract(row *Row) (contract, error) {
	result := contract{
		Runner: row.Runner, RuntimeProfileID: row.RuntimeProfileID, ResourceProfileID: row.ResourceProfileID,
		FixtureCapability: row.FixtureCapability, PostgresIsolation: row.PostgresIsolation,
		ServiceDependencies: sortedCopy(row.ServiceDependencies), RuntimeBinaryIDs: sortedCopy(row.RuntimeBinaryIDs),
		PreparationOperations: sortedCopy(append(slices.Clone(row.PreparationOperations), c.in.Owner.InvocationPolicies[row.Runner].PreparationOperations...)),
		CachePolicy:           rowCachePolicy(*row),
	}
	switch row.Runner {
	case RunnerGo:
		result.Package = row.Selector.Package
	case RunnerVitest:
		result.Project = row.Selector.ProjectID
	}
	claims, err := c.claims(row.ResourceProfileID, row.ServiceDependencies)
	if err != nil {
		return result, err
	}
	result.CPUTokens = min(c.options.CPUTokens, max(1, claims["cpu"]))
	if context, ok := c.in.Owner.RowExecutionContexts[row.RowID]; ok {
		result.CPUTokens, result.CachePolicy, result.Race, result.Alone = context.CPUTokens, context.CachePolicy, context.GoRace, row.RowID
	}
	if row.ProcessIsolation == "exclusive" || row.FixtureCapability == "managed_process" {
		result.Alone = row.RowID
	}
	return result, nil
}

// rowCachePolicy: only service-free, fixture-free Go and Vitest rows outside
// the measurement, security, visual and release classes are cacheable.
func rowCachePolicy(row Row) string {
	if (row.Runner == RunnerGo || row.Runner == RunnerVitest) && row.FixtureCapability == "none" && len(row.ServiceDependencies) == 0 &&
		!slices.Contains([]string{"measurement", "security", "visual", "release"}, row.EvidenceClass) {
		return "content_addressed"
	}
	return "none"
}

// items divides a contract group into indivisible items: one per Go row,
// one per Vitest file.
func (c *compiler) items(rows []*Row, contract contract) []item {
	byKey := map[string]*item{}
	var keys []string
	for _, row := range rows {
		key := row.RowID
		if contract.Runner == RunnerVitest {
			key = row.Selector.File
		}
		if byKey[key] == nil {
			byKey[key] = &item{}
			keys = append(keys, key)
		}
		byKey[key].rows = append(byKey[key].rows, row)
		byKey[key].work += c.estimate(*row)
	}
	items := make([]item, len(keys))
	for i, key := range keys {
		items[i] = *byKey[key]
	}
	return items
}

// split divides a contract group into balanced units: as many as the work
// needs at the runner's target size, filled longest-first into the
// least-loaded unit.
func (c *compiler) split(items []item, contract contract) [][]*Row {
	target := c.unitTarget(contract)
	slices.SortFunc(items, func(a, b item) int {
		return cmp.Or(cmp.Compare(b.work, a.work), strings.Compare(a.rows[0].RowID, b.rows[0].RowID))
	})
	total := int64(0)
	for _, item := range items {
		total += item.work
	}
	count := max(1, int((total+target-1)/target))
	if contract.Alone != "" {
		count = len(items)
	}
	count = min(count, len(items))
	bins := make([][]*Row, count)
	loads := make([]int64, count)
	for _, item := range items {
		least := 0
		for i := range loads {
			if loads[i] < loads[least] {
				least = i
			}
		}
		bins[least] = append(bins[least], item.rows...)
		loads[least] += item.work
	}
	return bins
}

// unitTarget is a unit's size target: the Vitest invocation policy's, or for
// Go the owner's, smaller for units on the heavy PostgreSQL lane.
func (c *compiler) unitTarget(contract contract) int64 {
	if contract.Runner == RunnerVitest {
		return c.in.Owner.InvocationPolicies[RunnerVitest].UnitTargetMS
	}
	claims, _ := c.claims(contract.ResourceProfileID, nil)
	if claims["postgres_heavy"] > 0 {
		return c.in.Owner.GoHeavyUnitTargetMS
	}
	return c.in.Owner.GoUnitTargetMS
}

func (c *compiler) estimate(row Row) int64 {
	if row.EstimatedWorkMS > 0 {
		return row.EstimatedWorkMS
	}
	return c.in.Owner.EvidenceEstimatesMS[row.EvidenceClass]
}

// runnerUnit builds a planned Go or Vitest unit. Its ID is the digest of
// its contract and row set.
func (c *compiler) runnerUnit(id string, stack []string) (string, error) {
	if _, ok := c.units[id]; ok {
		return id, nil
	}
	contract, rows := c.planned[id].contract, c.planned[id].rows
	rowIDs := make([]string, len(rows))
	owners := map[string]bool{}
	work := int64(0)
	for i, row := range rows {
		rowIDs[i] = row.RowID
		owners[row.OwnerID] = true
		work += c.estimate(*row)
	}
	slices.Sort(rowIDs)
	var needs []string
	targets := slices.Clone(contract.PreparationOperations)
	if len(contract.ServiceDependencies) != 0 {
		targets = append(targets, "harness.test_service_images")
	}
	for _, binary := range c.in.Topology.RuntimeBinaries {
		if slices.Contains(contract.RuntimeBinaryIDs, binary.ID) {
			targets = append(targets, binary.ProducerOperation)
		}
	}
	for _, target := range dedupe(targets) {
		ends, err := c.target(target, stack)
		if err != nil {
			return "", err
		}
		needs = append(needs, ends...)
	}
	claims, err := c.claims(contract.ResourceProfileID, contract.ServiceDependencies)
	if err != nil {
		return "", err
	}
	claims["cpu"] = contract.CPUTokens
	policy := c.in.Owner.InvocationPolicies[contract.Runner]
	for resource, amount := range policy.CoordinatorClaims {
		claims[resource] += amount
	}
	claims["cpu"] = min(claims["cpu"], c.options.CPUTokens)
	environment := fixtureEnvironment(contract.FixtureCapability, contract.ServiceDependencies)
	switch contract.Runner {
	case RunnerGo:
		maps.Copy(environment, c.goEnvironment(contract.Race))
		environment["CARTULARY_UNIT_CPU_TOKENS"] = strconv.FormatInt(contract.CPUTokens, 10)
		environment["GOMAXPROCS"] = strconv.FormatInt(contract.CPUTokens, 10)
		environment["CARTULARY_GO_SHUFFLE_SEED"] = c.options.GoShuffleSeed
	case RunnerVitest:
		environment["CARTULARY_TEST_TARGET"] = "frontend-unit"
		environment["VITEST_MAX_WORKERS"] = "1"
	}
	owner := "harness.command_surface"
	if len(owners) == 1 {
		owner = rows[0].OwnerID
	}
	unit := &Unit{
		ID: id, OwnerID: owner, Kind: "runner", Needs: dedupe(needs), Claims: claims,
		Residents: c.residents(id, contract.FixtureCapability, contract.ServiceDependencies, contract.PostgresIsolation), FixtureLease: contract.FixtureCapability,
		ServiceDependencies: contract.ServiceDependencies, PostgresIsolation: contract.PostgresIsolation, CachePolicy: contract.CachePolicy,
		TimeoutMS: c.in.Owner.DefaultTimeoutMS, RuntimeBinaryIDs: contract.RuntimeBinaryIDs,
		FailurePolicy: FailurePolicy{BlockDescendants: true}, EstimatedWorkMS: policy.StartupMS + work + policy.CleanupMS,
		Command: Command{Executable: "node", Args: []string{"tools/harness/execution/runners/row-runner-cli.mjs", "--row-ids", strings.Join(rowIDs, ",")}, Environment: environment},
	}
	for _, row := range rowIDs {
		unit.RowBindings = append(unit.RowBindings, RowBinding{RowID: row, Runner: contract.Runner})
		unit.EvidenceOutputs = append(unit.EvidenceOutputs, "rows/"+row+".json")
	}
	c.lock(unit, contract.ResourceProfileID)
	c.units[id] = unit
	return id, nil
}

// goEnvironment is the fixed Go build context every Go unit runs under.
func (c *compiler) goEnvironment(race bool) map[string]string {
	environment := map[string]string{
		"GOOS": "linux", "GOARCH": "amd64", "GOAMD64": "v1", "CGO_ENABLED": "0", "GOFLAGS": "", "GOENV": "off",
		"GOEXPERIMENT": "", "GOTOOLCHAIN": "go" + c.options.GoVersion, "GOWORK": "off", "GO111MODULE": "on", "CARTULARY_GO_RACE": "0",
	}
	if race {
		environment["CGO_ENABLED"], environment["CARTULARY_GO_RACE"] = "1", "1"
	}
	return environment
}

func fixtureEnvironment(capability string, services []string) map[string]string {
	environment := map[string]string{
		"CARTULARY_FIXTURE_CAPABILITY":           capability,
		"CARTULARY_HARNESS_SERVICE_DEPENDENCIES": strings.Join(services, ","),
	}
	switch capability {
	case "postgres_transaction":
		environment["CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT"] = "transaction"
	case "postgres_dedicated", "postgres_migration":
		environment["CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT"] = "template_clone"
	}
	return environment
}

// claims is a resource profile's claims raised to each service's minimums.
func (c *compiler) claims(profileID string, services []string) (map[string]int64, error) {
	index := slices.IndexFunc(c.in.Topology.ResourceProfiles, func(p ResourceProfile) bool { return p.ID == profileID })
	if index < 0 {
		return nil, fmt.Errorf("%w: unknown resource profile %s", ErrSelection, profileID)
	}
	claims := maps.Clone(c.in.Topology.ResourceProfiles[index].ResourceClaims)
	for _, service := range services {
		for resource, minimum := range c.in.Topology.ServiceResourceMinimums[service] {
			claims[resource] = max(claims[resource], minimum)
		}
	}
	return claims, nil
}

// residents names the physical allocations a unit's fixture lease creates
// or borrows: its own PostgreSQL generation, or the shared generation, and
// the shared object store. Shared allocations of an attached service session
// belong to that session. A unit without a fixture lease manages or attaches
// to its services itself and has no residents.
func (c *compiler) residents(unitID, lease string, services []string, isolation string) []Resident {
	var residents []Resident
	if lease == "none" {
		return nil
	}
	shared := ScopeGraph
	if c.options.AttachedServices {
		shared = ScopeAttached
	}
	physical := c.in.Topology.PhysicalResourceClaims
	if slices.Contains(services, "object_store") {
		residents = append(residents, Resident{Key: "object_store:shared", Scope: shared, Claims: maps.Clone(physical["object_store"])})
	}
	switch isolation {
	case "exclusive":
		residents = append(residents, Resident{Key: "postgres:unit:" + unitID, Scope: ScopeUnit, Claims: maps.Clone(physical["postgres"])})
	case "shared":
		residents = append(residents, Resident{Key: "postgres:shared", Scope: shared, Claims: maps.Clone(physical["postgres"])})
	}
	return residents
}

// lock takes the shared host-activity lock for any unit with claims; the
// quiet-capacity profile takes it exclusively.
func (c *compiler) lock(unit *Unit, profileID string) {
	switch {
	case profileID == "credential_capacity_quiet":
		unit.ExclusiveLocks = []string{"host_activity"}
	case len(unit.Claims) != 0:
		unit.SharedLocks = []string{"host_activity"}
	}
}

// ends returns the units among ids that no other unit among ids needs.
func ends(units map[string]*Unit, ids []string) []string {
	needed := map[string]bool{}
	for _, id := range ids {
		for _, need := range units[id].Needs {
			needed[need] = true
		}
	}
	var result []string
	for _, id := range ids {
		if !needed[id] {
			result = append(result, id)
		}
	}
	return result
}

func dedupe(values []string) []string {
	result := slices.Clone(values)
	slices.Sort(result)
	return slices.Compact(result)
}

func sortedCopy(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return dedupe(values)
}
