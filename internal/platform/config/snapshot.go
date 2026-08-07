package config

import (
	"errors"
	"fmt"
	"reflect"
)

var errSnapshotNotAdmitted = errors.New("configuration snapshot is not structurally admitted")

// CoreConfiguration is the closed Core 04 projection retained after
// structural admission. It excludes source-owner namespaces and transient
// decoder/presence state.
type CoreConfiguration struct {
	ConfigSchemaID    string
	DeploymentProfile string
	Application       ApplicationConfig
	Roots             RootBindings
	Bootstrap         BootstrapConfig
	Timeouts          TimeoutConfig
	Intervals         IntervalConfig
	Limits            LimitConfig
}

// Snapshot is the immutable result of structural configuration admission.
type Snapshot struct {
	admitted                      bool
	core                          CoreConfiguration
	requestedClaimRegistrationIDs []string
	values                        map[string]snapshotValue
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
	for _, entry := range catalog.entries {
		presence := namespacePresence{namespace: entry.namespace, presence: cfg.presence}
		value, findings := entry.project(&cfg, presence)
		diagnostics = append(diagnostics, findings...)
		if value == nil {
			continue
		}
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
		admitted:                      true,
		core:                          coreConfigurationFromDocument(cfg),
		requestedClaimRegistrationIDs: requestedClaimRegistrationIDs(cfg),
		values:                        values,
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

// Core returns a defensive copy of the admitted Core 04 configuration.
func (snapshot Snapshot) Core() CoreConfiguration {
	return cloneCoreConfiguration(snapshot.core)
}

// RequestedClaimRegistrationIDs returns only registered claims whose admitted
// effective value is true. The returned identity list is defensive and sorted.
func RequestedClaimRegistrationIDs(snapshot Snapshot) []string {
	return append([]string(nil), snapshot.requestedClaimRegistrationIDs...)
}

// ValidateSnapshotForStartup verifies canonical root readiness without exposing the
// underlying deployment document.
func ValidateSnapshotForStartup(snapshot Snapshot) error {
	if !snapshot.admitted {
		return errSnapshotNotAdmitted
	}
	diagnostics := validateStartupFilesystemRoots(snapshot.core.Roots)
	if len(diagnostics) > 0 {
		return newDiagnosticsError(diagnostics)
	}
	return nil
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

func coreConfigurationFromDocument(cfg document) CoreConfiguration {
	return CoreConfiguration{
		ConfigSchemaID:    cfg.ConfigSchemaID,
		DeploymentProfile: cfg.DeploymentProfile,
		Application:       cfg.Application,
		Roots:             cfg.Roots,
		Bootstrap:         cfg.Bootstrap,
		Timeouts:          cfg.Timeouts,
		Intervals:         cfg.Intervals,
		Limits:            cfg.Limits,
	}
}

func cloneCoreConfiguration(core CoreConfiguration) CoreConfiguration {
	return core
}
