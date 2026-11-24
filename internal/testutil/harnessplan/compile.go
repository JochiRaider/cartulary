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
// (overhaul milestones 3 and 4: Vitest, Playwright, measurement rows).
var ErrUnsupported = errors.New("not yet supported by the Go planner")

// Selection names what to compile: exact rows, an owner slice (optionally
// narrowed to rows), or an operation or private target union.
type Selection struct {
	Rows   []string
	Owner  string
	Target string
}

// Options are the per-invocation facts compilation depends on.
type Options struct {
	// CPUTokens bounds a unit's CPU claim (the snapshot's cpu_tokens).
	CPUTokens int64
	// GoVersion is the pinned toolchain version (tools/toolchain_pins.json).
	GoVersion string
	// GoShuffleSeed, when set, orders Go tests with that seed.
	GoShuffleSeed string
}

// Compile turns a selection into a work graph.
func Compile(in *Inputs, selection Selection, options Options) (*Graph, error) {
	if options.CPUTokens < 1 || options.GoVersion == "" {
		return nil, fmt.Errorf("%w: compilation needs CPU tokens and a Go version", ErrSelection)
	}
	c := &compiler{in: in, options: options, units: map[string]*Unit{}, targets: map[string][]string{}, rows: map[string]*Row{}, shellRows: map[string][]*Row{}}
	for i := range in.Catalog.Rows {
		row := &in.Catalog.Rows[i]
		c.rows[row.RowID] = row
		if row.Runner == RunnerShell {
			c.shellRows[row.Selector.OperationID] = append(c.shellRows[row.Selector.OperationID], row)
		}
	}
	var err error
	switch {
	case selection.Target != "" && selection.Owner == "" && len(selection.Rows) == 0:
		_, err = c.target(selection.Target, nil)
	case selection.Owner != "":
		var rows []string
		if rows, err = c.ownerRows(selection.Owner, selection.Rows); err == nil {
			_, err = c.compileRows(rows, nil)
		}
	case len(selection.Rows) != 0 && selection.Target == "":
		_, err = c.compileRows(selection.Rows, nil)
	default:
		err = fmt.Errorf("%w: select exactly one of rows, owner or target", ErrSelection)
	}
	if err != nil {
		return nil, err
	}
	graph := &Graph{Units: c.units}
	if err := graph.validate(); err != nil {
		return nil, err
	}
	return graph, nil
}

type compiler struct {
	in        *Inputs
	options   Options
	units     map[string]*Unit
	targets   map[string][]string // target name -> its end units
	rows      map[string]*Row
	shellRows map[string][]*Row // operation -> shell rows that report it
}

func (c *compiler) ownerRows(owner string, narrow []string) ([]string, error) {
	var rows []string
	for _, row := range c.in.Catalog.Rows {
		if row.OwnerID == owner && (len(narrow) == 0 || slices.Contains(narrow, row.RowID)) {
			rows = append(rows, row.RowID)
		}
	}
	if len(rows) == 0 || len(narrow) != 0 && len(rows) != len(narrow) {
		return nil, fmt.Errorf("%w: owner %s has no such rows", ErrSelection, owner)
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
			selected := c.catalogRows(*member.Catalog)
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
func (c *compiler) catalogRows(member CatalogMember) []string {
	var rows []string
	for _, row := range c.in.Catalog.Rows {
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
		Residents: c.residents(unitID, policy.ServiceDependencies, policy.PostgresIsolation), FixtureLease: lease,
		ServiceDependencies: slices.Clone(policy.ServiceDependencies), PostgresIsolation: policy.PostgresIsolation,
		CachePolicy: policy.CachePolicy, TimeoutMS: timeout, ArtifactOutputs: policy.ReusableArtifactOutputs,
		EstimatedWorkMS: policy.EstimatedWorkMS,
		Command: Command{Executable: "node", Args: []string{"tools/harness/execution/operation-runner-cli.mjs", id}, Environment: map[string]string{
			"CARTULARY_TEST_TARGET":                  operation.ReportKey,
			"CARTULARY_OPERATION_DIGEST":             "sha256:" + digest(operation.Execution),
			"CARTULARY_HARNESS_SERVICE_DEPENDENCIES": strings.Join(policy.ServiceDependencies, ","),
		}},
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

// compileRows compiles catalog rows and returns the IDs of their units.
func (c *compiler) compileRows(ids []string, stack []string) ([]string, error) {
	var goRows []*Row
	var units []string
	for _, id := range dedupe(ids) {
		row, ok := c.rows[id]
		if !ok {
			return nil, fmt.Errorf("%w: unknown row %s", ErrSelection, id)
		}
		switch row.Runner {
		case RunnerGo:
			if row.FixtureProfileID != "" || row.EvidenceClass == "measurement" {
				return nil, fmt.Errorf("%w: measurement row %s", ErrUnsupported, id)
			}
			goRows = append(goRows, row)
		case RunnerShell:
			unit, err := c.operation(row.Selector.OperationID, stack)
			if err != nil {
				return nil, err
			}
			units = append(units, unit)
		default:
			return nil, fmt.Errorf("%w: %s row %s", ErrUnsupported, row.Runner, id)
		}
	}
	goUnits, err := c.goUnits(goRows, stack)
	if err != nil {
		return nil, err
	}
	return dedupe(append(units, goUnits...)), nil
}

// goContract is everything that must be equal for Go rows to share a unit.
type goContract struct {
	Package               string
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

func (c *compiler) goUnits(rows []*Row, stack []string) ([]string, error) {
	groups := map[string][]*Row{}
	contracts := map[string]goContract{}
	for _, row := range rows {
		contract, err := c.goContract(row)
		if err != nil {
			return nil, err
		}
		key := digest(contract)
		groups[key] = append(groups[key], row)
		contracts[key] = contract
	}
	keys := slices.Sorted(maps.Keys(groups))
	var ids []string
	for _, key := range keys {
		contract := contracts[key]
		for _, bin := range c.split(groups[key], contract) {
			id, err := c.goUnit(key, contract, bin, stack)
			if err != nil {
				return nil, err
			}
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (c *compiler) goContract(row *Row) (goContract, error) {
	contract := goContract{
		Package: row.Selector.Package, RuntimeProfileID: row.RuntimeProfileID, ResourceProfileID: row.ResourceProfileID,
		FixtureCapability: row.FixtureCapability, PostgresIsolation: row.PostgresIsolation,
		ServiceDependencies: sortedCopy(row.ServiceDependencies), RuntimeBinaryIDs: sortedCopy(row.RuntimeBinaryIDs),
		PreparationOperations: sortedCopy(row.PreparationOperations), CachePolicy: rowCachePolicy(*row),
	}
	claims, err := c.claims(row.ResourceProfileID, row.ServiceDependencies)
	if err != nil {
		return contract, err
	}
	contract.CPUTokens = min(c.options.CPUTokens, max(1, claims["cpu"]))
	if context, ok := c.in.Owner.RowExecutionContexts[row.RowID]; ok {
		contract.CPUTokens, contract.CachePolicy, contract.Race, contract.Alone = context.CPUTokens, context.CachePolicy, context.GoRace, row.RowID
	}
	if row.ProcessIsolation == "exclusive" || row.FixtureCapability == "managed_process" {
		contract.Alone = row.RowID
	}
	return contract, nil
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

// split divides one contract group into balanced units: as many as the work
// needs at the target size, filled longest-first into the least-loaded unit.
func (c *compiler) split(rows []*Row, contract goContract) [][]*Row {
	target := c.in.Owner.GoUnitTargetMS
	if c.heavy(contract) {
		target = c.in.Owner.GoHeavyUnitTargetMS
	}
	ordered := slices.Clone(rows)
	slices.SortFunc(ordered, func(a, b *Row) int {
		return cmp.Or(cmp.Compare(c.estimate(*b), c.estimate(*a)), strings.Compare(a.RowID, b.RowID))
	})
	total := int64(0)
	for _, row := range ordered {
		total += c.estimate(*row)
	}
	count := max(1, int((total+target-1)/target))
	if contract.Alone != "" {
		count = len(ordered)
	}
	count = min(count, len(ordered))
	bins := make([][]*Row, count)
	loads := make([]int64, count)
	for _, row := range ordered {
		least := 0
		for i := range loads {
			if loads[i] < loads[least] {
				least = i
			}
		}
		bins[least] = append(bins[least], row)
		loads[least] += c.estimate(*row)
	}
	return bins
}

func (c *compiler) heavy(contract goContract) bool {
	claims, _ := c.claims(contract.ResourceProfileID, nil)
	return claims["postgres_heavy"] > 0
}

func (c *compiler) estimate(row Row) int64 {
	if row.EstimatedWorkMS > 0 {
		return row.EstimatedWorkMS
	}
	return c.in.Owner.EvidenceEstimatesMS[row.EvidenceClass]
}

func (c *compiler) goUnit(key string, contract goContract, rows []*Row, stack []string) (string, error) {
	rowIDs := make([]string, len(rows))
	work := int64(0)
	for i, row := range rows {
		rowIDs[i] = row.RowID
		work += c.estimate(*row)
	}
	slices.Sort(rowIDs)
	id := "go:" + key[:16] + ":" + digest(rowIDs)[:12]
	if _, ok := c.units[id]; ok {
		return id, nil
	}
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
	policy := c.in.Owner.InvocationPolicies[RunnerGo]
	environment := c.goEnvironment(contract.Race)
	maps.Copy(environment, fixtureEnvironment(contract.FixtureCapability, contract.ServiceDependencies))
	environment["CARTULARY_UNIT_CPU_TOKENS"] = strconv.FormatInt(contract.CPUTokens, 10)
	environment["GOMAXPROCS"] = strconv.FormatInt(contract.CPUTokens, 10)
	environment["CARTULARY_GO_SHUFFLE_SEED"] = c.options.GoShuffleSeed
	unit := &Unit{
		ID: id, OwnerID: "harness.backend", Kind: "runner", Needs: dedupe(needs), Claims: claims,
		Residents: c.residents(id, contract.ServiceDependencies, contract.PostgresIsolation), FixtureLease: contract.FixtureCapability,
		ServiceDependencies: contract.ServiceDependencies, PostgresIsolation: contract.PostgresIsolation, CachePolicy: contract.CachePolicy,
		TimeoutMS: c.in.Owner.DefaultTimeoutMS, RuntimeBinaryIDs: contract.RuntimeBinaryIDs,
		EstimatedWorkMS: policy.StartupMS + work + policy.CleanupMS,
		Command:         Command{Executable: "node", Args: []string{"tools/harness/execution/runners/row-runner-cli.mjs", "--unit-rows"}, Environment: environment},
	}
	for _, row := range rowIDs {
		unit.RowBindings = append(unit.RowBindings, RowBinding{RowID: row, Runner: RunnerGo})
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

// residents names the physical allocations a unit's services create: its
// own PostgreSQL generation, or the shared generation, and the shared
// object store.
func (c *compiler) residents(unitID string, services []string, isolation string) []Resident {
	var residents []Resident
	physical := c.in.Topology.PhysicalResourceClaims
	if slices.Contains(services, "object_store") {
		residents = append(residents, Resident{Key: "object_store:shared", Scope: ScopeGraph, Claims: maps.Clone(physical["object_store"])})
	}
	switch isolation {
	case "exclusive":
		residents = append(residents, Resident{Key: "postgres:unit:" + unitID, Scope: ScopeUnit, Claims: maps.Clone(physical["postgres"])})
	case "shared":
		residents = append(residents, Resident{Key: "postgres:shared", Scope: ScopeGraph, Claims: maps.Clone(physical["postgres"])})
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
