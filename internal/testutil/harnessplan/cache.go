package harnessplan

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"maps"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
)

const (
	cacheRecordSchema = "cartulary.harness_cache_record.v2"
	// Dependency strategies recorded in cache facts.
	strategyBroad         = "broad"
	strategyBroadFallback = "broad_fallback"
	strategyGoPackages    = "go_packages"
	strategyWorkspaces    = "typescript_workspaces"
)

// CacheContext is what every cache key binds besides the unit and its
// dependency closure.
type CacheContext struct {
	ToolchainDigest string
	// Platform is GOOS/GOARCH of the preparing host.
	Platform string
	// Cryptographic execution settings inherited by units.
	FIPSSelector, GoDebug, GoToolchain string
}

// closure is a dependency closure: the source inputs a cached result depends
// on, as indices into the run's source snapshot.
type closure struct {
	strategy string
	indices  []int
	digest   string
}

// cachePlanner computes cache facts for one graph over one source snapshot.
type cachePlanner struct {
	root     string
	in       *Inputs
	source   []SourceInput
	context  CacheContext
	closures map[string]closure
	imports  map[string][]string
	// workspaces are the frontend workspaces by package name, once read.
	workspaces map[string]workspace
	// entries holds each snapshot entry's canonical JSON once computed.
	entries map[int][]byte
	// runners holds each profile's runner closure, modules each Node entry
	// module's closure, and literals each Go file's named repository files,
	// all computed once.
	runners  map[string]runnerClosure
	modules  map[string]moduleSet
	literals map[int][]string
}

// CacheFacts decides each unit's cache eligibility and, for an eligible unit,
// its input digest and dependency closure. It returns the facts in unit order
// and the closures' source indices by dependency digest.
func CacheFacts(root string, in *Inputs, graph *Graph, source []SourceInput, context CacheContext) ([]CacheFact, map[string][]int, error) {
	p := &cachePlanner{root: root, in: in, source: source, context: context, closures: map[string]closure{}, imports: map[string][]string{}, entries: map[int][]byte{}, runners: map[string]runnerClosure{}, modules: map[string]moduleSet{}, literals: map[int][]string{}}
	inputs := map[string][]int{}
	var facts []CacheFact
	for _, unit := range graph.Sorted() {
		fact, closure, err := p.fact(unit)
		if err != nil {
			return nil, nil, err
		}
		if fact.Eligible {
			inputs[closure.digest] = closure.indices
		}
		facts = append(facts, fact)
	}
	return facts, inputs, nil
}

func (p *cachePlanner) fact(unit *Unit) (CacheFact, closure, error) {
	ineligible := func(reason string, profile *CacheProfile) (CacheFact, closure, error) {
		fact := CacheFact{UnitID: unit.ID, Reason: &reason}
		if profile != nil {
			fact.ProfileID, fact.Policy = &profile.ProfileID, &profile.Policy
		}
		return fact, closure{}, nil
	}
	if unit.CachePolicy == "none" {
		return ineligible("policy_none", nil)
	}
	profile, err := p.profile(unit)
	if err != nil {
		return CacheFact{}, closure{}, err
	}
	if len(unit.ArtifactOutputs) == 0 && len(unit.RowBindings) == 0 {
		return ineligible("reusable_outputs_incomplete", nil)
	}
	for _, output := range unit.EvidenceOutputs {
		if !strings.HasPrefix(output, "rows/") && !strings.HasPrefix(output, "unit-results/") {
			return ineligible("fresh_outputs_incomplete", nil)
		}
	}
	// The planner proves no vulnerability-database revision, so findings
	// keyed on one are never reused and the scan always runs.
	if profile.RequiresVulnerabilityDatabaseRevision {
		return ineligible("freshness_unknown", profile)
	}
	closure := p.closure(profile, unit)
	inputDigest := semanticDigest(map[string]any{
		"cache_schema_id":     cacheRecordSchema,
		"profile_id":          profile.ProfileID,
		"dependency_strategy": closure.strategy,
		"unit_digest":         resultIdentity(unit),
		"inputs":              closure.digest,
		"platform":            p.context.Platform,
		"toolchain_digest":    p.context.ToolchainDigest,
		"cryptographic_execution": map[string]string{
			"selector": p.context.FIPSSelector, "mode": p.context.GoDebug, "toolchain": p.context.GoToolchain,
		},
	}, "")
	fact := CacheFact{UnitID: unit.ID, Eligible: true, ProfileID: &profile.ProfileID, Policy: &profile.Policy,
		InputDigest: &inputDigest, DependencyStrategy: &closure.strategy, DependencyDigest: &closure.digest}
	return fact, closure, nil
}

// resultIdentity digests what a unit's result depends on: the unit without
// the fields that only schedule it (P5-3b). Estimates, resource claims,
// residents, locks and the timeout decide when and where a unit runs, and a Go
// unit's CPU budget (GOMAXPROCS and -p) only how many of its tests run at
// once. None of them may change its cache key, so an estimate refresh or a
// host with other capacity keeps every result.
func resultIdentity(unit *Unit) string {
	identity := *unit
	identity.EstimatedWorkMS, identity.TimeoutMS, identity.SemanticDigest = 0, 0, ""
	identity.Claims, identity.Residents, identity.SharedLocks, identity.ExclusiveLocks = nil, nil, nil, nil
	if unit.Command.Executable == "go" {
		identity.Command.Environment = maps.Clone(unit.Command.Environment)
		delete(identity.Command.Environment, "GOMAXPROCS")
		identity.Command.Args = slices.DeleteFunc(slices.Clone(unit.Command.Args), func(argument string) bool { return strings.HasPrefix(argument, "-p=") })
	}
	return semanticDigest(identity, "")
}

// profile finds the registered profile of a cacheable unit: its operation's
// profile, or its rows' runner's. A cacheable unit without one is an
// authoring error.
func (p *cachePlanner) profile(unit *Unit) (*CacheProfile, error) {
	match := func(bound func(CacheProfile) bool) *CacheProfile {
		index := slices.IndexFunc(p.in.Cache.Profiles, bound)
		if index < 0 {
			return nil
		}
		return &p.in.Cache.Profiles[index]
	}
	var profile *CacheProfile
	if operation, ok := strings.CutPrefix(unit.ID, "operation:"); ok {
		profile = match(func(c CacheProfile) bool { return slices.Contains(c.OperationIDs, operation) })
	} else if len(unit.RowBindings) != 0 {
		runner := unit.RowBindings[0].Runner
		profile = match(func(c CacheProfile) bool { return slices.Contains(c.RunnerTypes, runner) })
	}
	if profile == nil || profile.Policy != unit.CachePolicy {
		return nil, fmt.Errorf("%w: unit %s has cache policy %s but no matching cache profile", ErrInput, unit.ID, unit.CachePolicy)
	}
	return profile, nil
}

// closure resolves a unit's dependency closure. A unit-aware profile narrows
// Go units to their packages and Vitest units to their workspaces; any
// incomplete knowledge selects the profile's broad closure, recorded as a
// fallback.
func (p *cachePlanner) closure(profile *CacheProfile, unit *Unit) closure {
	if profile.DependencyStrategy == "unit_aware" {
		if closure, err := p.unitClosure(profile, unit); err == nil {
			return closure
		}
		return p.broad(profile, strategyBroadFallback, "resolution_incomplete")
	}
	return p.broad(profile, strategyBroad, "registered_broad")
}

func (p *cachePlanner) broad(profile *CacheProfile, strategy, reason string) closure {
	key := strategy + "\x00" + profile.ProfileID
	if cached, ok := p.closures[key]; ok {
		return cached
	}
	roots := dedupe(profile.InputRoots)
	set := map[int]bool{}
	for _, root := range roots {
		p.under(root, set)
	}
	closure := p.digest(strategy, map[string]any{"reason": reason, "registered_roots": roots}, set)
	p.closures[key] = closure
	return closure
}

// unitClosure is the closure of a Go or Vitest unit: what its rows run
// (Go packages and their repository-local imports, or frontend workspaces and
// their declared workspace dependencies) with everything beneath them, the
// profile's unit inputs, the selected owners' manifests, and the rows'
// fixture roots. An imported package's own tests are not part of what a row
// runs, so they are left out.
func (p *cachePlanner) unitClosure(profile *CacheProfile, unit *Unit) (closure, error) {
	var selected, owners, fixtures, services []string
	runner := ""
	for _, binding := range unit.RowBindings {
		row := p.in.row(binding.RowID)
		if row.RowID == "" || runner != "" && row.Runner != runner {
			return closure{}, errors.New("a selected row is unknown or the rows mix runners")
		}
		runner = row.Runner
		switch runner {
		case RunnerGo:
			selected = append(selected, strings.TrimPrefix(strings.TrimSuffix(row.Selector.Package, "/"), "./"))
		case RunnerVitest:
			selected = append(selected, row.Selector.File)
		default:
			return closure{}, fmt.Errorf("%s rows have no unit-aware closure", runner)
		}
		owners = append(append(owners, row.OwnerID), row.CollaboratorIDs...)
		fixtures = append(fixtures, row.FixtureInputRoots...)
		services = append(services, row.ServiceDependencies...)
	}
	selected, owners, fixtures, services = dedupe(selected), dedupe(owners), dedupe(fixtures), dedupe(services)
	key := digest(map[string]any{"profile": profile.ProfileID, "runner": runner, "selected": selected, "owners": owners, "fixtures": fixtures, "services": services})
	if cached, ok := p.closures[key]; ok {
		return cached, nil
	}
	set := map[int]bool{}
	strategy, metadata := strategyGoPackages, map[string]any{}
	switch runner {
	case RunnerGo:
		directories, err := p.packageDirectories(selected, true)
		if err != nil {
			return closure{}, err
		}
		for _, directory := range directories {
			if slices.Contains(selected, directory) {
				p.under(directory, set)
			} else {
				p.withoutTests(directory, set)
			}
			for module := directory; module != "."; module = path.Dir(module) {
				if _, ok := p.index(module + "/go.mod"); ok {
					p.under(module+"/go.mod", set)
					p.under(module+"/go.sum", set)
					break
				}
			}
		}
		metadata["packages"] = directories
	case RunnerVitest:
		roots, err := p.workspaceRoots(selected)
		if err != nil {
			return closure{}, err
		}
		for _, root := range roots {
			p.under(root, set)
		}
		strategy, metadata["workspaces"] = strategyWorkspaces, roots
	}
	for _, root := range fixtures {
		members := map[int]bool{}
		p.under(root, members)
		if len(members) == 0 {
			return closure{}, fmt.Errorf("fixture root %s is absent", root)
		}
		for index := range members {
			if p.source[index].Kind != "file" {
				return closure{}, fmt.Errorf("fixture root %s holds a non-regular member", root)
			}
			set[index] = true
		}
	}
	for _, root := range profile.UnitInputs.Paths {
		p.under(root, set)
	}
	// A service-backed row's result also depends on the services it ran
	// against: the migrated template its databases are cloned from, and the
	// helper that starts and seeds them. Service image pins live in the test
	// support packages its Go closure already holds.
	if len(services) != 0 {
		for _, root := range profile.UnitInputs.ServicePaths {
			p.under(root, set)
		}
		metadata["services"] = services
		metadata["migration_template_digest"] = p.in.MigrationDigest
	}
	for _, prefix := range profile.UnitInputs.Prefixes {
		p.prefixed(prefix, set)
	}
	runnerInputs, err := p.runnerClosure(profile)
	if err != nil {
		return closure{}, err
	}
	for index := range runnerInputs {
		set[index] = true
	}
	if runner == RunnerGo {
		if err := p.namedInputs(profile.UnitInputs, set); err != nil {
			return closure{}, err
		}
	}
	// Planning estimates order work; they never change a result.
	if index, ok := p.index(estimatesPath); ok {
		delete(set, index)
	}
	manifests := map[string]string{}
	for _, owner := range p.in.Catalog.Owners {
		manifests[owner.OwnerID] = owner.ManifestPath
	}
	for _, owner := range owners {
		if manifest := manifests[owner]; manifest != "" {
			p.under(manifest, set)
		}
	}
	closure := p.digest(strategy, metadata, set)
	p.closures[key] = closure
	return closure, nil
}

// workspace is one frontend workspace package.
type workspace struct {
	root         string
	dependencies []string
}

var workspaceEntry = regexp.MustCompile(`^\s+-\s+["']?([A-Za-z0-9._/*-]+)["']?\s*$`)

// workspaceRoots are the workspaces that contain the selected files and,
// transitively, every workspace they declare as a dependency. pnpm resolves
// only declared dependencies and the frontend import boundary forbids
// reaching another workspace by relative path, so declarations close the
// graph without parsing sources.
func (p *cachePlanner) workspaceRoots(files []string) ([]string, error) {
	if p.workspaces == nil {
		workspaces, err := p.readWorkspaces()
		if err != nil {
			return nil, err
		}
		p.workspaces = workspaces
	}
	pending := []string{}
	for _, file := range files {
		owner := ""
		for name, workspace := range p.workspaces {
			if strings.HasPrefix(file, workspace.root+"/") && (owner == "" || len(workspace.root) > len(p.workspaces[owner].root)) {
				owner = name
			}
		}
		if _, ok := p.index(file); !ok || owner == "" {
			return nil, fmt.Errorf("selected file %s has no workspace", file)
		}
		pending = append(pending, owner)
	}
	selected := map[string]bool{}
	for len(pending) != 0 {
		name := pending[0]
		pending = pending[1:]
		if selected[name] {
			continue
		}
		selected[name] = true
		pending = append(pending, p.workspaces[name].dependencies...)
	}
	var roots []string
	for name := range selected {
		roots = append(roots, p.workspaces[name].root)
	}
	return dedupe(roots), nil
}

// readWorkspaces reads the workspace globs from pnpm-workspace.yaml (its
// packages list: exact directories or "<directory>/*") and each workspace's
// package name and declared workspace dependencies.
func (p *cachePlanner) readWorkspaces() (map[string]workspace, error) {
	manifest, err := p.text("pnpm-workspace.yaml")
	if err != nil {
		return nil, err
	}
	var patterns []string
	listing := false
	for _, line := range strings.Split(manifest, "\n") {
		switch {
		case strings.TrimSpace(line) == "packages:":
			listing = true
		case listing && workspaceEntry.MatchString(line):
			patterns = append(patterns, workspaceEntry.FindStringSubmatch(line)[1])
		case listing && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, " "):
			listing = false
		}
	}
	if len(patterns) == 0 {
		return nil, errors.New("pnpm-workspace.yaml lists no workspaces")
	}
	var manifests []string
	for _, pattern := range patterns {
		directory, wildcard := strings.CutSuffix(pattern, "/*")
		if strings.ContainsAny(directory, "*?[") {
			return nil, fmt.Errorf("unsupported workspace pattern %s", pattern)
		}
		for _, index := range p.rangeOf(directory + "/") {
			relative := strings.TrimPrefix(p.source[index].Path, directory+"/")
			if wildcard && strings.Count(relative, "/") == 1 && strings.HasSuffix(relative, "/package.json") || !wildcard && relative == "package.json" {
				manifests = append(manifests, p.source[index].Path)
			}
		}
	}
	type packageJSON struct {
		Name                 string            `json:"name"`
		Dependencies         map[string]string `json:"dependencies"`
		DevDependencies      map[string]string `json:"devDependencies"`
		OptionalDependencies map[string]string `json:"optionalDependencies"`
		PeerDependencies     map[string]string `json:"peerDependencies"`
	}
	parsed := map[string]packageJSON{}
	names := map[string]bool{}
	for _, manifest := range manifests {
		text, err := p.text(manifest)
		if err != nil {
			return nil, err
		}
		var value packageJSON
		if err := json.Unmarshal([]byte(text), &value); err != nil || value.Name == "" {
			return nil, fmt.Errorf("%s has no package name", manifest)
		}
		if names[value.Name] {
			return nil, fmt.Errorf("workspace package %s is declared twice", value.Name)
		}
		names[value.Name] = true
		parsed[manifest] = value
	}
	workspaces := map[string]workspace{}
	for manifest, value := range parsed {
		var dependencies []string
		for _, section := range []map[string]string{value.Dependencies, value.DevDependencies, value.OptionalDependencies, value.PeerDependencies} {
			for name := range section {
				if names[name] && name != value.Name {
					dependencies = append(dependencies, name)
				}
			}
		}
		workspaces[value.Name] = workspace{root: path.Dir(manifest), dependencies: dedupe(dependencies)}
	}
	return workspaces, nil
}

var moduleLine = regexp.MustCompile(`(?m)^module\s+(\S+)\s*$`)

// packageDirectories follows the repository-local imports from the given
// packages: of every Go file (all build constraints) in a given package when
// tested is set, otherwise of its non-test files, and of every non-test Go
// file in an imported one, since a package's tests are never compiled into
// another package's tests.
func (p *cachePlanner) packageDirectories(packages []string, tested bool) ([]string, error) {
	if len(packages) == 0 {
		return nil, nil
	}
	goMod, err := p.text("go.mod")
	if err != nil {
		return nil, err
	}
	match := moduleLine.FindStringSubmatch(goMod)
	if match == nil {
		return nil, errors.New("go.mod has no module path")
	}
	module := match[1]
	pending := slices.Clone(packages)
	selected := map[string]bool{}
	for len(pending) != 0 {
		directory := pending[0]
		pending = pending[1:]
		if selected[directory] {
			continue
		}
		if directory == "" || !safeRelative(directory) {
			return nil, fmt.Errorf("package %q is unsafe", directory)
		}
		selected[directory] = true
		files := 0
		for _, index := range p.rangeOf(directory + "/") {
			entry := p.source[index]
			name := strings.TrimPrefix(entry.Path, directory+"/")
			if strings.Contains(name, "/") || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") && !(tested && slices.Contains(packages, directory)) {
				continue
			}
			if entry.Kind != "file" {
				return nil, fmt.Errorf("%s is not a regular file", entry.Path)
			}
			files++
			imports, err := p.goImports(entry)
			if err != nil {
				return nil, err
			}
			for _, imported := range imports {
				if imported == module {
					return nil, errors.New("the module root cannot be imported as a package")
				}
				if local, ok := strings.CutPrefix(imported, module+"/"); ok && !selected[local] {
					pending = append(pending, local)
				}
			}
		}
		if files == 0 {
			return nil, fmt.Errorf("package %s has no Go files", directory)
		}
	}
	return dedupe(slicesKeys(selected)), nil
}

func (p *cachePlanner) goImports(entry SourceInput) ([]string, error) {
	if imports, ok := p.imports[entry.Path]; ok {
		return imports, nil
	}
	text, err := p.text(entry.Path)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(token.NewFileSet(), entry.Path, text, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var imports []string
	for _, spec := range file.Imports {
		value, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return nil, err
		}
		imports = append(imports, value)
	}
	p.imports[entry.Path] = imports
	return imports, nil
}

// index finds a snapshot path.
func (p *cachePlanner) index(relative string) (int, bool) {
	index := sort.Search(len(p.source), func(i int) bool { return p.source[i].Path >= relative })
	return index, index < len(p.source) && p.source[index].Path == relative
}

// text reads a snapshot file and proves it still has the snapshot's content.
func (p *cachePlanner) text(relative string) (string, error) {
	index, ok := p.index(relative)
	if !ok || p.source[index].Kind != "file" {
		return "", fmt.Errorf("%s is not a snapshot file", relative)
	}
	data, err := readBounded(filepath.Join(p.root, relative))
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	if "sha256:"+hex.EncodeToString(sum[:]) != p.source[index].ByteDigest {
		return "", fmt.Errorf("%s changed after the source snapshot", relative)
	}
	return string(data), nil
}

// rangeOf returns the snapshot indices of paths that start with prefix. The
// snapshot is sorted, so they are contiguous.
func (p *cachePlanner) rangeOf(prefix string) []int {
	start := sort.Search(len(p.source), func(i int) bool { return p.source[i].Path >= prefix })
	var indices []int
	for i := start; i < len(p.source) && strings.HasPrefix(p.source[i].Path, prefix); i++ {
		indices = append(indices, i)
	}
	return indices
}

// under adds the snapshot path root and everything beneath it.
func (p *cachePlanner) under(root string, set map[int]bool) {
	root = strings.TrimSuffix(root, "/")
	for _, index := range p.rangeOf(root) {
		if path := p.source[index].Path; path == root || strings.HasPrefix(path, root+"/") {
			set[index] = true
		}
	}
}

// withoutTests adds everything beneath root except Go test files.
func (p *cachePlanner) withoutTests(root string, set map[int]bool) {
	members := map[int]bool{}
	p.under(root, members)
	for index := range members {
		if !strings.HasSuffix(p.source[index].Path, "_test.go") {
			set[index] = true
		}
	}
}

func (p *cachePlanner) prefixed(prefix string, set map[int]bool) {
	for _, index := range p.rangeOf(prefix) {
		set[index] = true
	}
}

// digest names a closure: the semantic digest of {entries, metadata,
// strategy}. The canonical document is assembled from each entry's canonical
// JSON, encoded once per snapshot entry, in canonical key order.
func (p *cachePlanner) digest(strategy string, metadata map[string]any, set map[int]bool) closure {
	indices := slicesKeys(set)
	slices.Sort(indices)
	hash := sha256.New()
	hash.Write([]byte(`{"entries":[`))
	for i, index := range indices {
		if i > 0 {
			hash.Write([]byte(","))
		}
		entry, ok := p.entries[index]
		if !ok {
			entry = canonicalBytes(p.source[index], "")
			p.entries[index] = entry
		}
		hash.Write(entry)
	}
	hash.Write([]byte(`],"metadata":`))
	hash.Write(canonicalBytes(metadata, ""))
	hash.Write([]byte(`,"strategy":`))
	hash.Write(canonicalBytes(strategy, ""))
	hash.Write([]byte("}"))
	return closure{strategy: strategy, indices: indices, digest: "sha256:" + hex.EncodeToString(hash.Sum(nil))}
}

func slicesKeys[K comparable](set map[K]bool) []K {
	keys := make([]K, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	return keys
}
