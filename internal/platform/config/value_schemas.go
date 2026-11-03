package config

import (
	"fmt"
	"sort"
	"strings"
)

// ValueSchemaBinding binds a declared configuration key to the actual typed
// namespace decoder/projector registered with its owner catalog.
type ValueSchemaBinding struct {
	Path     string
	SchemaID string
}

type NamespaceSchemaBindings struct {
	Namespace string
	Values    []ValueSchemaBinding
}

// ValidateSchemaBindings admits metadata before reading deployment input. It
// must never invoke owner callbacks or resolve a value, file, secret or trust.
func (catalog Catalog) ValidateSchemaBindings(declared []NamespaceSchemaBindings) error {
	seen := make(map[string]struct{}, len(declared))
	for _, group := range declared {
		if !configPathPattern.MatchString(group.Namespace) {
			return fmt.Errorf("configuration schema namespace is invalid")
		}
		if _, duplicate := seen[group.Namespace]; duplicate {
			return fmt.Errorf("configuration schema namespace %q is repeated", group.Namespace)
		}
		seen[group.Namespace] = struct{}{}
		var expected []ValueSchemaBinding
		for _, entry := range catalog.entries {
			if entry.namespace == group.Namespace {
				expected = entry.valueSchemas
				break
			}
		}
		values, err := normalizeValueSchemas(group.Namespace, nil, group.Values)
		if err != nil {
			return err
		}
		if len(values) != len(expected) {
			return fmt.Errorf("configuration namespace %q value schema set differs from its typed owner", group.Namespace)
		}
		for i, value := range values {
			if value != expected[i] {
				return fmt.Errorf("configuration namespace %q value schema binding differs from its typed owner", group.Namespace)
			}
		}
	}
	return nil
}

func normalizeValueSchemas(namespace string, ownedPaths []string, supplied []ValueSchemaBinding) ([]ValueSchemaBinding, error) {
	values := append([]ValueSchemaBinding(nil), supplied...)
	sort.Slice(values, func(i, j int) bool { return values[i].Path < values[j].Path })
	for i, value := range values {
		if !configPathPattern.MatchString(value.Path) || !pathWithinNamespace(value.Path, namespace) || value.SchemaID == "" || strings.TrimSpace(value.SchemaID) != value.SchemaID {
			return nil, fmt.Errorf("configuration namespace %q has an invalid value schema binding", namespace)
		}
		if i > 0 && values[i-1].Path == value.Path {
			return nil, fmt.Errorf("configuration namespace %q repeats a value schema key", namespace)
		}
		if ownedPaths != nil {
			index := sort.SearchStrings(ownedPaths, value.Path)
			if index == len(ownedPaths) || ownedPaths[index] != value.Path {
				return nil, fmt.Errorf("configuration namespace %q value schema key is not owned", namespace)
			}
		}
	}
	return values, nil
}
