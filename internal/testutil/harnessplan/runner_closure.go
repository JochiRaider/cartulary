package harnessplan

import (
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
)

// The row runner's closure (G36, TH-HARNESS-REQ-807 amended 2026-11-25). A
// test row's result depends on the runner that executes it, not on the rest
// of the harness. Go rows run under the Go row runner, whose closure is the
// repository-local import closure of its packages without tests. Vitest rows
// run under the JavaScript row runner, whose closure is its entry modules and
// every repository module they reach through a relative static import,
// re-export, literal dynamic import or literal require. Installed modules and
// packages are bound by the lockfile and go.sum, which every unit-aware
// closure already holds.
var (
	moduleFrom     = regexp.MustCompile(`(?:^|[\s;}])(?:import|export)\s[^'"]*?\bfrom\s*["']([^"']+)["']`)
	moduleBare     = regexp.MustCompile(`(?:^|[\s;}])import\s*["']([^"']+)["']`)
	moduleDynamic  = regexp.MustCompile(`\bimport\(\s*["']([^"']+)["']\s*\)`)
	moduleRequire  = regexp.MustCompile(`\brequire\w*\(\s*["']([^"']+)["']\s*[,)]`)
	moduleComputed = regexp.MustCompile(`\bimport\(\s*[^"'\s)]`)
	modulePatterns = []*regexp.Regexp{moduleFrom, moduleBare, moduleDynamic, moduleRequire}
)

type runnerClosure struct {
	set map[int]bool
	err error
}

type moduleSet = runnerClosure

// runnerClosure returns the snapshot indices of a profile's runner closure,
// computed once per profile. An unresolvable relative specifier, or a
// computed dynamic import in a module not declared external, is an error, so
// the caller falls back to the broad closure rather than reuse a result on an
// incomplete key.
func (p *cachePlanner) runnerClosure(profile *CacheProfile) (map[int]bool, error) {
	if known, ok := p.runners[profile.ProfileID]; ok {
		return known.set, known.err
	}
	set, err := p.resolveRunner(profile.UnitInputs)
	p.runners[profile.ProfileID] = runnerClosure{set: set, err: err}
	return set, err
}

func (p *cachePlanner) resolveRunner(inputs *UnitInputs) (map[int]bool, error) {
	set := map[int]bool{}
	directories, err := p.packageDirectories(inputs.RunnerPackages, false)
	if err != nil {
		return nil, err
	}
	for _, directory := range directories {
		p.withoutTests(directory, set)
	}
	modules, err := p.moduleClosure(inputs.RunnerEntries, inputs.RunnerExternalImports)
	if err != nil {
		return nil, err
	}
	maps.Copy(set, modules)
	return set, nil
}

// moduleClosure is the closure of Node entry modules: each entry and every
// repository module it reaches through a relative static import, re-export,
// literal dynamic import or literal require. Each entry's closure is computed
// once.
func (p *cachePlanner) moduleClosure(entries, externals []string) (map[int]bool, error) {
	set := map[int]bool{}
	for _, entry := range entries {
		known, ok := p.modules[entry]
		if !ok {
			known.set, known.err = p.resolveModules(entry, externals)
			p.modules[entry] = known
		}
		if known.err != nil {
			return nil, known.err
		}
		maps.Copy(set, known.set)
	}
	return set, nil
}

func (p *cachePlanner) resolveModules(entry string, externals []string) (map[int]bool, error) {
	set := map[int]bool{}
	pending := []string{entry}
	for len(pending) != 0 {
		module := pending[0]
		pending = pending[1:]
		index, ok := p.index(module)
		if !ok {
			return nil, fmt.Errorf("module %s is not in the source snapshot", module)
		}
		if set[index] {
			continue
		}
		set[index] = true
		scan, err := p.moduleScan(index)
		if err != nil {
			return nil, err
		}
		if scan.Computed && !slices.Contains(externals, module) {
			return nil, fmt.Errorf("module %s has a computed import", module)
		}
		for _, specifier := range scan.Specifiers {
			resolved, ok := p.resolveModule(path.Join(path.Dir(module), specifier))
			if !ok {
				return nil, fmt.Errorf("module %s imports unresolvable %s", module, specifier)
			}
			pending = append(pending, resolved)
		}
	}
	return set, nil
}

// goPathLiteral is a Go string literal shaped like a repository-relative path.
var goPathLiteral = regexp.MustCompile(`"([A-Za-z0-9_.@+-]+(?:/[A-Za-z0-9_.@+-]+)+)"`)

// namedInputs adds the repository files a Go closure's code names by a string
// literal, because a test or its support may read or launch them; a named
// Node module brings its module closure. Only exact snapshot files count: a
// named directory is left to the profile's unit inputs. A named module whose
// closure cannot be resolved makes the closure broad.
func (p *cachePlanner) namedInputs(inputs *UnitInputs, set map[int]bool) error {
	var files []int
	for index := range set {
		if strings.HasSuffix(p.source[index].Path, ".go") {
			files = append(files, index)
		}
	}
	slices.Sort(files)
	var modules []string
	for _, index := range files {
		literals, ok := p.literals[index]
		if !ok {
			candidates, err := p.goLiterals(index)
			if err != nil {
				return err
			}
			for _, candidate := range candidates {
				if target, ok := p.index(candidate); ok && p.source[target].Kind == "file" {
					literals = append(literals, candidate)
				}
			}
			p.literals[index] = literals
		}
		for _, literal := range literals {
			target, _ := p.index(literal)
			set[target] = true
			if strings.HasSuffix(literal, ".mjs") || strings.HasSuffix(literal, ".js") || strings.HasSuffix(literal, ".cjs") {
				modules = append(modules, literal)
			}
		}
	}
	named, err := p.moduleClosure(dedupe(modules), inputs.RunnerExternalImports)
	if err != nil {
		return err
	}
	maps.Copy(set, named)
	return nil
}

// resolveModule finds a module the way Node resolves a relative specifier:
// the exact file, or the file with a JavaScript extension.
func (p *cachePlanner) resolveModule(candidate string) (string, bool) {
	for _, suffix := range []string{"", ".mjs", ".js", ".cjs", "/index.mjs", "/index.js"} {
		if index, ok := p.index(candidate + suffix); ok && p.source[index].Kind == "file" {
			return candidate + suffix, true
		}
	}
	return "", false
}
