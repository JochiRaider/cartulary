package config

import (
	"fmt"
	"reflect"
)

// Snapshot is the immutable result of structural configuration admission.
type Snapshot struct {
	core   document
	values map[string]snapshotValue
}

type snapshotValue struct {
	valueType reflect.Type
	value     any
	clone     func(any) any
}

// Materialize projects every catalog contribution in deterministic ID order
// and aggregates its pure structural findings into the deployment-config
// envelope.
func (catalog Catalog) materialize(cfg document) (Snapshot, error) {
	values := make(map[string]snapshotValue, len(catalog.entries))
	diagnostics := make([]Diagnostic, 0)
	source := documentSource{document: cloneConfig(cfg)}
	for _, entry := range catalog.entries {
		value, findings := entry.project(source)
		diagnostics = append(diagnostics, findings...)
		values[entry.id] = snapshotValue{
			valueType: entry.valueType,
			value:     entry.clone(value),
			clone:     entry.clone,
		}
	}
	if len(diagnostics) > 0 {
		return Snapshot{}, newDiagnosticsError(diagnostics)
	}
	return Snapshot{
		core:   cloneConfig(cfg),
		values: values,
	}, nil
}

// LoadSnapshotWithOptions runs the existing strict artifact loader and then
// materializes typed owner values. Startup filesystem capability construction
// remains the next explicit phase.
func LoadSnapshotWithOptions(options LoadOptions, catalog Catalog) (Snapshot, error) {
	cfg, err := loadWithOptionsAndCatalog(options, catalog)
	if err != nil {
		return Snapshot{}, err
	}
	return catalog.materialize(cfg)
}

// LoadSnapshotFromTOML materializes an immutable snapshot from an in-memory
// deployment artifact. Application assembly uses this for composition tests;
// production startup uses LoadSnapshotWithOptions and its selected file.
func LoadSnapshotFromTOML(data []byte, options LoadOptions, catalog Catalog) (Snapshot, error) {
	cfg, err := decodeDocumentWithCatalog(append([]byte(nil), data...), options, catalog)
	if err != nil {
		return Snapshot{}, err
	}
	return catalog.materialize(cfg)
}

// Decode copies one normalized configuration subtree into caller-owned memory.
// It is intended for application facades that project narrow owner settings.
func (snapshot Snapshot) Decode(path string, destination any) error {
	return documentSource{document: cloneConfig(snapshot.core)}.Decode(path, destination)
}

// SnapshotBooleanValuesAtPaths projects exact normalized Boolean paths.
func SnapshotBooleanValuesAtPaths(snapshot Snapshot, paths []string) (map[string]bool, error) {
	return booleanValuesAtPaths(snapshot.core, paths)
}

// ValidateSnapshotForStartup verifies canonical root readiness without exposing the
// underlying deployment document.
func ValidateSnapshotForStartup(snapshot Snapshot) error {
	_, err := validateForStartup(snapshot.core)
	return err
}

// Value returns one typed owner value.
func Value[T any](snapshot Snapshot, key Key[T]) (T, error) {
	var zero T
	value, present := snapshot.values[key.id]
	if !present {
		return zero, fmt.Errorf("configuration snapshot has no contribution %q", key.id)
	}
	if value.valueType != key.valueType {
		return zero, fmt.Errorf(
			"configuration contribution %q has type %v, requested %v",
			key.id,
			value.valueType,
			key.valueType,
		)
	}
	typed, ok := value.clone(value.value).(T)
	if !ok {
		return zero, fmt.Errorf("configuration contribution %q value does not match %v", key.id, key.valueType)
	}
	return typed, nil
}

func cloneConfig(cfg document) document {
	cloned := cfg
	if cfg.Telemetry.Exporter.Headers != nil {
		cloned.Telemetry.Exporter.Headers = make(map[string]telemetrySecretDocument, len(cfg.Telemetry.Exporter.Headers))
		for name, reference := range cfg.Telemetry.Exporter.Headers {
			cloned.Telemetry.Exporter.Headers[name] = reference
		}
	}
	return cloned
}
