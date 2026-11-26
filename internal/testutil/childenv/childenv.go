// Package childenv builds the environment harness test support gives a
// process it starts for another component, such as a product server or
// operator (TH-HARNESS Section 5.5, single writer). The child receives only:
//   - the parent's process and tool basics: the environment registry's
//     inherited names and prefixes and the task surface's global inputs;
//   - the harness names the launcher forwards by name;
//   - the values the launcher sets.
//
// A launcher never passes its whole environment on, so harness and test
// context (fixture bindings, the Go build context, run identity) cannot reach
// a product process unless the launcher names it. It is the Go counterpart of
// tools/harness/runtime/child-environment.mjs and reads the same declarations.
package childenv

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
)

type basics struct {
	names    map[string]bool
	prefixes []string
}

var (
	mu    sync.Mutex
	cache = map[string]basics{}
)

func load(root string) (basics, error) {
	mu.Lock()
	defer mu.Unlock()
	if loaded, ok := cache[root]; ok {
		return loaded, nil
	}
	var registry struct {
		Inherited struct {
			Names    []string `json:"names"`
			Prefixes []string `json:"prefixes"`
		} `json:"inherited"`
	}
	var surface struct {
		GlobalInputs []struct {
			Name string `json:"name"`
		} `json:"global_inputs"`
	}
	for path, into := range map[string]any{"tools/harness_environment_registry.json": &registry, "tools/task_surface_manifest.json": &surface} {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			return basics{}, err
		}
		if err := json.Unmarshal(data, into); err != nil {
			return basics{}, fmt.Errorf("%s: %w", path, err)
		}
	}
	loaded := basics{names: map[string]bool{}, prefixes: slices.Clone(registry.Inherited.Prefixes)}
	for _, name := range registry.Inherited.Names {
		loaded.names[name] = true
	}
	for _, input := range surface.GlobalInputs {
		loaded.names[input.Name] = true
	}
	if len(loaded.names) == 0 {
		return basics{}, fmt.Errorf("%s declares no inherited names", root)
	}
	cache[root] = loaded
	return loaded, nil
}

// Build returns the child's environment as sorted NAME=value entries: the
// basics present in parent, the parent's values of the forwarded names, then
// set (which wins over both).
func Build(root string, parent []string, forward []string, set map[string]string) ([]string, error) {
	declared, err := load(root)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, entry := range parent {
		name, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		if declared.names[name] || slices.ContainsFunc(declared.prefixes, func(prefix string) bool { return strings.HasPrefix(name, prefix) }) || slices.Contains(forward, name) {
			values[name] = value
		}
	}
	maps.Copy(values, set)
	child := make([]string, 0, len(values))
	for _, name := range slices.Sorted(maps.Keys(values)) {
		child = append(child, name+"="+values[name])
	}
	return child, nil
}
