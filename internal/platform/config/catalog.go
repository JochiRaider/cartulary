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

// Source is the read-only, presence-aware configuration view supplied to one
// catalog contribution. Decode copies only the requested owned subtree into a
// caller-owned value; it never exposes the mutable deployment document.
type Source interface {
	Decode(path string, destination any) error
	Defined(path ...string) bool
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

type documentSource struct {
	document document
}

func (source documentSource) Decode(path string, destination any) error {
	target := reflect.ValueOf(destination)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return fmt.Errorf("configuration decode destination must be a non-nil pointer")
	}

	value := reflect.ValueOf(source.document)
	if path != "" {
		resolved, present := configFieldAtPath(&source.document, path)
		if !present {
			return fmt.Errorf("configuration path %q is unresolved", path)
		}
		value = resolved
	}
	if err := copyConfigurationValue(target.Elem(), value); err != nil {
		return fmt.Errorf("decode configuration path %q: %w", path, err)
	}
	return nil
}

func (source documentSource) Defined(path ...string) bool {
	if len(path) == 0 {
		return false
	}
	if len(path) == 1 {
		return source.document.presence.Defined(strings.Split(path[0], ".")...)
	}
	return source.document.presence.Defined(path...)
}

func copyConfigurationValue(target reflect.Value, source reflect.Value) error {
	for source.Kind() == reflect.Pointer {
		if source.IsNil() {
			target.SetZero()
			return nil
		}
		source = source.Elem()
	}
	if target.Kind() == reflect.Pointer {
		if target.IsNil() {
			target.Set(reflect.New(target.Type().Elem()))
		}
		return copyConfigurationValue(target.Elem(), source)
	}
	if source.Type().AssignableTo(target.Type()) {
		target.Set(cloneReflectValue(source))
		return nil
	}

	switch target.Kind() {
	case reflect.Struct:
		if source.Kind() != reflect.Struct {
			return fmt.Errorf("cannot copy %s into %s", source.Type(), target.Type())
		}
		for index := 0; index < target.NumField(); index++ {
			targetFieldType := target.Type().Field(index)
			if !targetFieldType.IsExported() {
				continue
			}
			sourceField, present := fieldByTOMLName(source, configurationFieldName(targetFieldType))
			if !present {
				continue
			}
			if err := copyConfigurationValue(target.Field(index), sourceField); err != nil {
				return err
			}
		}
		return nil
	case reflect.Slice:
		if source.Kind() != reflect.Slice {
			return fmt.Errorf("cannot copy %s into %s", source.Type(), target.Type())
		}
		copied := reflect.MakeSlice(target.Type(), source.Len(), source.Len())
		for index := 0; index < source.Len(); index++ {
			if err := copyConfigurationValue(copied.Index(index), source.Index(index)); err != nil {
				return err
			}
		}
		target.Set(copied)
		return nil
	case reflect.Map:
		if source.Kind() != reflect.Map {
			return fmt.Errorf("cannot copy %s into %s", source.Type(), target.Type())
		}
		copied := reflect.MakeMapWithSize(target.Type(), source.Len())
		iterator := source.MapRange()
		for iterator.Next() {
			key := reflect.New(target.Type().Key()).Elem()
			if err := copyConfigurationValue(key, iterator.Key()); err != nil {
				return err
			}
			value := reflect.New(target.Type().Elem()).Elem()
			if err := copyConfigurationValue(value, iterator.Value()); err != nil {
				return err
			}
			copied.SetMapIndex(key, value)
		}
		target.Set(copied)
		return nil
	default:
		if source.Type().ConvertibleTo(target.Type()) {
			target.Set(source.Convert(target.Type()))
			return nil
		}
		return fmt.Errorf("cannot copy %s into %s", source.Type(), target.Type())
	}
}

func cloneReflectValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		copy := reflect.New(value.Type().Elem())
		copy.Elem().Set(cloneReflectValue(value.Elem()))
		return copy
	case reflect.Slice:
		copy := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			copy.Index(index).Set(cloneReflectValue(value.Index(index)))
		}
		return copy
	case reflect.Map:
		copy := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			copy.SetMapIndex(cloneReflectValue(iterator.Key()), cloneReflectValue(iterator.Value()))
		}
		return copy
	case reflect.Struct:
		copy := reflect.New(value.Type()).Elem()
		for index := 0; index < value.NumField(); index++ {
			if copy.Field(index).CanSet() {
				copy.Field(index).Set(cloneReflectValue(value.Field(index)))
			}
		}
		return copy
	default:
		return value
	}
}

func fieldByTOMLName(value reflect.Value, name string) (reflect.Value, bool) {
	valueType := value.Type()
	for index := 0; index < value.NumField(); index++ {
		fieldType := valueType.Field(index)
		if configurationFieldName(fieldType) == name {
			return value.Field(index), true
		}
	}
	return reflect.Value{}, false
}

func configurationFieldName(field reflect.StructField) string {
	name := strings.Split(field.Tag.Get("toml"), ",")[0]
	if name == "" {
		return strings.ToLower(field.Name)
	}
	return name
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
	Project      func(T, Source) (T, []Diagnostic)
	Clone        func(T) T
}

type catalogEntry struct {
	id        string
	valueType reflect.Type
	namespace string
	paths     []string
	decode    func(map[string]any) (any, []string, []Diagnostic)
	overlay   func(*document, []string, string) *Diagnostic
	project   func(*document, Source) (any, []Diagnostic)
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
		project: func(cfg *document, source Source) (any, []Diagnostic) {
			current, present := cfg.namespaces[definition.Namespace].(*T)
			if !present || current == nil {
				return nil, []Diagnostic{{
					Path:       definition.Namespace,
					ReasonCode: "type_mismatch",
					Message:    "owner namespace was not decoded",
				}}
			}
			value, diagnostics := definition.Project(definition.Clone(*current), source)
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
