package config

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
)

// NamespacePresence is the read-only, namespace-scoped presence view supplied
// to one catalog contribution. Paths outside the contribution's registered
// namespace are never observable.
type NamespacePresence interface {
	Defined(path ...string) bool
}

type namespacePresence struct {
	namespace string
	presence  configPresence
}

func (source namespacePresence) Defined(path ...string) bool {
	if len(path) == 0 {
		return false
	}
	if len(path) == 1 {
		path = strings.Split(path[0], ".")
	}
	joined := strings.Join(path, ".")
	if !pathWithinNamespace(joined, source.namespace) {
		return false
	}
	return source.presence.Defined(path...)
}

// NamespaceDecoder is the closed, namespace-scoped artifact decoder supplied
// to one statically registered owner. It exposes neither the complete
// deployment document nor any other owner's namespace.
type NamespaceDecoder interface {
	Decode(destination any) error
}

type namespaceDecoder struct {
	namespace string
	raw       map[string]any
	decoded   bool
	undecoded []string
}

func (decoder *namespaceDecoder) Decode(destination any) error {
	if decoder == nil {
		return fmt.Errorf("configuration namespace decoder is required")
	}
	if decoder.decoded {
		return fmt.Errorf("configuration namespace %q was decoded more than once", decoder.namespace)
	}
	decoder.decoded = true
	target := reflect.ValueOf(destination)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return fmt.Errorf("configuration namespace decode destination must be a non-nil pointer")
	}

	value, present := decoder.raw[decoder.namespace]
	if !present {
		value = map[string]any{}
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("configuration namespace %q must be a table", decoder.namespace)
	}
	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(object); err != nil {
		return fmt.Errorf("encode configuration namespace %q: %w", decoder.namespace, err)
	}
	metadata, err := toml.Decode(encoded.String(), destination)
	if err != nil {
		return fmt.Errorf("decode configuration namespace %q: %w", decoder.namespace, err)
	}
	for _, key := range metadata.Undecoded() {
		path := decoder.namespace
		if len(key) > 0 {
			path += "." + strings.Join(key, ".")
		}
		decoder.undecoded = append(decoder.undecoded, path)
	}
	sort.Strings(decoder.undecoded)
	return nil
}

var (
	contributionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,126}$`)
	configPathPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

// Key is an opaque typed identity for one owner contribution.
type Key[T any] struct {
	id        string
	valueType reflect.Type
}

// NewKey constructs a stable typed contribution identity.
func NewKey[T any](id string) (Key[T], error) {
	if !contributionIDPattern.MatchString(id) {
		return Key[T]{}, fmt.Errorf("invalid configuration contribution id %q", id)
	}
	return Key[T]{
		id:        id,
		valueType: reflect.TypeOf((*T)(nil)).Elem(),
	}, nil
}

// Definition declares the exact namespace and keys owned by one pure
// contribution. Project must not perform filesystem, secret, network,
// database, process, or other external effects.
type Definition[T any] struct {
	Key          Key[T]
	Namespace    string
	Paths        []string
	Decode       func(NamespaceDecoder) (T, []Diagnostic)
	ApplyOverlay func(T, []string, string) (T, *Diagnostic)
	Project      func(T, NamespacePresence) (T, []Diagnostic)
	Clone        func(T) T
}

type catalogEntry struct {
	id        string
	valueType reflect.Type
	namespace string
	paths     []string
	decode    func(map[string]any) (any, []string, []Diagnostic)
	overlay   func(*document, []string, string) *Diagnostic
	project   func(*document, NamespacePresence) (any, []Diagnostic)
	clone     func(any) any
}

// CatalogBuilder validates contribution ownership before a catalog becomes
// immutable.
type CatalogBuilder struct {
	entries []catalogEntry
	sealed  bool
}

// Register adds one typed owner contribution.
func Register[T any](builder *CatalogBuilder, definition Definition[T]) error {
	if builder == nil {
		return fmt.Errorf("configuration catalog builder is required")
	}
	if builder.sealed {
		return fmt.Errorf("configuration catalog builder is sealed")
	}
	if definition.Key.id == "" || definition.Key.valueType == nil {
		return fmt.Errorf("configuration contribution key is required")
	}
	if !configPathPattern.MatchString(definition.Namespace) {
		return fmt.Errorf("configuration contribution %q has invalid namespace %q", definition.Key.id, definition.Namespace)
	}
	if len(definition.Paths) == 0 {
		return fmt.Errorf("configuration contribution %q must own at least one path", definition.Key.id)
	}
	if definition.Decode == nil {
		return fmt.Errorf("configuration contribution %q namespace decoder is required", definition.Key.id)
	}
	if definition.Project == nil {
		return fmt.Errorf("configuration contribution %q projector is required", definition.Key.id)
	}
	if definition.Clone == nil {
		return fmt.Errorf("configuration contribution %q immutable-value clone is required", definition.Key.id)
	}

	paths := append([]string(nil), definition.Paths...)
	sort.Strings(paths)
	for index, path := range paths {
		if !configPathPattern.MatchString(path) || !pathWithinNamespace(path, definition.Namespace) {
			return fmt.Errorf("configuration contribution %q path %q is outside namespace %q", definition.Key.id, path, definition.Namespace)
		}
		if index > 0 && paths[index-1] == path {
			return fmt.Errorf("configuration contribution %q repeats path %q", definition.Key.id, path)
		}
	}
	entry := catalogEntry{
		id:        definition.Key.id,
		valueType: definition.Key.valueType,
		namespace: definition.Namespace,
		paths:     paths,
		decode: func(raw map[string]any) (any, []string, []Diagnostic) {
			decoder := &namespaceDecoder{namespace: definition.Namespace, raw: raw}
			value, diagnostics := definition.Decode(decoder)
			if !decoder.decoded {
				diagnostics = append(diagnostics, Diagnostic{
					Path:       definition.Namespace,
					ReasonCode: "type_mismatch",
					Message:    "owner namespace decoder did not consume its namespace",
				})
			}
			copy := definition.Clone(value)
			return &copy, append([]string(nil), decoder.undecoded...), diagnostics
		},
		project: func(cfg *document, presence NamespacePresence) (any, []Diagnostic) {
			current, present := cfg.namespaces[definition.Namespace].(*T)
			if !present || current == nil {
				return nil, []Diagnostic{{
					Path:       definition.Namespace,
					ReasonCode: "type_mismatch",
					Message:    "owner namespace was not decoded",
				}}
			}
			value, diagnostics := definition.Project(definition.Clone(*current), presence)
			return value, diagnostics
		},
		clone: func(value any) any {
			return definition.Clone(value.(T))
		},
	}
	if definition.ApplyOverlay != nil {
		entry.overlay = func(cfg *document, segments []string, raw string) *Diagnostic {
			current, present := cfg.namespaces[definition.Namespace].(*T)
			if !present || current == nil {
				return &Diagnostic{
					Path:       strings.Join(segments, "."),
					ReasonCode: "type_mismatch",
					Message:    fmt.Sprintf("configuration namespace %q was not decoded", definition.Namespace),
				}
			}
			updated, diagnostic := definition.ApplyOverlay(definition.Clone(*current), segments, raw)
			if diagnostic != nil {
				return diagnostic
			}
			*current = definition.Clone(updated)
			return nil
		}
	}
	builder.entries = append(builder.entries, entry)
	return nil
}

func (catalog Catalog) decodeNamespaces(cfg *document, raw map[string]any) ([]string, []Diagnostic) {
	if cfg.namespaces == nil {
		cfg.namespaces = make(map[string]any, len(catalog.entries))
	}
	undecoded := make([]string, 0)
	diagnostics := make([]Diagnostic, 0)
	for _, entry := range catalog.entries {
		value, unknownPaths, findings := entry.decode(raw)
		cfg.namespaces[entry.namespace] = value
		undecoded = append(undecoded, unknownPaths...)
		diagnostics = append(diagnostics, findings...)
	}
	sort.Strings(undecoded)
	return undecoded, diagnostics
}

func (catalog Catalog) ownsNamespacePath(path string) bool {
	for _, entry := range catalog.entries {
		if pathWithinNamespace(path, entry.namespace) {
			return true
		}
	}
	return false
}

// Catalog is an immutable, deterministically ordered owner registry.
type Catalog struct {
	entries []catalogEntry
}

// Build validates duplicate identities, namespace overlap, and key overlap.
func (builder *CatalogBuilder) Build() (Catalog, error) {
	if builder == nil {
		return Catalog{}, fmt.Errorf("configuration catalog builder is required")
	}
	if builder.sealed {
		return Catalog{}, fmt.Errorf("configuration catalog builder is sealed")
	}
	builder.sealed = true

	entries := append([]catalogEntry(nil), builder.entries...)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].id < entries[j].id
	})
	ownedPaths := map[string]string{}
	for index, entry := range entries {
		if index > 0 && entries[index-1].id == entry.id {
			return Catalog{}, fmt.Errorf("duplicate configuration contribution id %q", entry.id)
		}
		for priorIndex := 0; priorIndex < index; priorIndex++ {
			prior := entries[priorIndex]
			if namespacesOverlap(prior.namespace, entry.namespace) {
				return Catalog{}, fmt.Errorf(
					"configuration contribution namespaces %q and %q overlap",
					prior.namespace,
					entry.namespace,
				)
			}
		}
		for _, path := range entry.paths {
			if priorID, duplicate := ownedPaths[path]; duplicate {
				return Catalog{}, fmt.Errorf("configuration path %q is owned by both %q and %q", path, priorID, entry.id)
			}
			ownedPaths[path] = entry.id
		}
	}
	return Catalog{entries: entries}, nil
}

func (catalog Catalog) applyOverlay(cfg *document, segments []string, raw string) (bool, *Diagnostic) {
	path := strings.Join(segments, ".")
	for _, entry := range catalog.entries {
		if !pathWithinNamespace(path, entry.namespace) {
			continue
		}
		if entry.overlay == nil {
			return false, nil
		}
		return true, entry.overlay(cfg, segments, raw)
	}
	return false, nil
}

func pathWithinNamespace(path string, namespace string) bool {
	return path == namespace || strings.HasPrefix(path, namespace+".")
}

func namespacesOverlap(left string, right string) bool {
	return pathWithinNamespace(left, right) || pathWithinNamespace(right, left)
}
