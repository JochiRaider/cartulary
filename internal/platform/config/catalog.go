package config

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// Source is the read-only, presence-aware configuration view supplied to one
// catalog contribution. Decode copies only the requested owned subtree into a
// caller-owned value; it never exposes the mutable deployment document.
type Source interface {
	Decode(path string, destination any) error
	Defined(path ...string) bool
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

// ValidationPhase names the fixed deployment-configuration lifecycle. Owner
// contributions participate only in decode and structural validation; later
// phases remain application-owned effects.
type ValidationPhase string

const (
	ValidationPhaseParse            ValidationPhase = "parse"
	ValidationPhaseClaimOverlays    ValidationPhase = "claim_overlays"
	ValidationPhaseInactiveKeys     ValidationPhase = "inactive_keys"
	ValidationPhaseKeyOwnership     ValidationPhase = "key_ownership"
	ValidationPhaseDecodeNormalize  ValidationPhase = "decode_normalize"
	ValidationPhaseStructural       ValidationPhase = "structural_validation"
	ValidationPhaseRootCapabilities ValidationPhase = "root_capabilities"
	ValidationPhaseOwnerPreflight   ValidationPhase = "owner_preflight"
)

var validationPhases = []ValidationPhase{
	ValidationPhaseParse,
	ValidationPhaseClaimOverlays,
	ValidationPhaseInactiveKeys,
	ValidationPhaseKeyOwnership,
	ValidationPhaseDecodeNormalize,
	ValidationPhaseStructural,
	ValidationPhaseRootCapabilities,
	ValidationPhaseOwnerPreflight,
}

var (
	contributionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,126}$`)
	configPathPattern     = regexp.MustCompile(`^[a-z][a-z0-9_]*(?:\.[a-z][a-z0-9_]*)*$`)
)

// ValidationPhases returns a copy of the normative implementation order.
func ValidationPhases() []ValidationPhase {
	return append([]ValidationPhase(nil), validationPhases...)
}

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

// ID returns the stable catalog identity.
func (key Key[T]) ID() string {
	return key.id
}

// Definition declares the exact namespace and keys owned by one pure
// contribution. Project must not perform filesystem, secret, network,
// database, process, or other external effects.
type Definition[T any] struct {
	Key          Key[T]
	Namespace    string
	Paths        []string
	ClaimPath    string
	ApplyOverlay func(T, []string, string) (T, *Diagnostic)
	Project      func(Source) (T, []Diagnostic)
	Clone        func(T) T
}

type catalogEntry struct {
	id        string
	valueType reflect.Type
	namespace string
	paths     []string
	claimPath string
	overlay   func(*document, []string, string) *Diagnostic
	project   func(Source) (any, []Diagnostic)
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
	if definition.ClaimPath != "" {
		if !configPathPattern.MatchString(definition.ClaimPath) || !pathWithinNamespace(definition.ClaimPath, definition.Namespace) {
			return fmt.Errorf("configuration contribution %q claim path %q is outside namespace %q", definition.Key.id, definition.ClaimPath, definition.Namespace)
		}
	}

	entry := catalogEntry{
		id:        definition.Key.id,
		valueType: definition.Key.valueType,
		namespace: definition.Namespace,
		paths:     paths,
		claimPath: definition.ClaimPath,
		project: func(source Source) (any, []Diagnostic) {
			value, diagnostics := definition.Project(source)
			return value, diagnostics
		},
		clone: func(value any) any {
			return definition.Clone(value.(T))
		},
	}
	if definition.ApplyOverlay != nil {
		entry.overlay = func(cfg *document, segments []string, raw string) *Diagnostic {
			var current T
			source := documentSource{document: cloneConfig(*cfg)}
			if err := source.Decode(definition.Namespace, &current); err != nil {
				return &Diagnostic{
					Path:       strings.Join(segments, "."),
					ReasonCode: "type_mismatch",
					Message:    err.Error(),
				}
			}
			updated, diagnostic := definition.ApplyOverlay(current, segments, raw)
			if diagnostic != nil {
				return diagnostic
			}
			target, present := configFieldAtPath(cfg, definition.Namespace)
			if !present {
				return &Diagnostic{
					Path:       strings.Join(segments, "."),
					ReasonCode: "unknown_key",
					Message:    fmt.Sprintf("configuration namespace %q is unresolved", definition.Namespace),
				}
			}
			if err := copyConfigurationValue(target, reflect.ValueOf(updated)); err != nil {
				return &Diagnostic{
					Path:       strings.Join(segments, "."),
					ReasonCode: "type_mismatch",
					Message:    err.Error(),
				}
			}
			return nil
		}
	}
	builder.entries = append(builder.entries, entry)
	return nil
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

// IDs returns contribution identities in deterministic execution order.
func (catalog Catalog) IDs() []string {
	ids := make([]string, len(catalog.entries))
	for index, entry := range catalog.entries {
		ids[index] = entry.id
	}
	return ids
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
