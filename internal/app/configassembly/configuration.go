// Package configassembly owns application-level construction and
// materialization of the deployment-configuration contribution catalog.
package configassembly

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
)

var (
	enterpriseAuthenticationConfigurationKey = mustConfigurationKey[enterpriseauth.Configuration]("platform.enterpriseauth.configuration")
	networkFlowConfigurationKey              = mustConfigurationKey[networkflow.Configuration]("module.networkflow.configuration")
	revisionsConfigurationKey                = mustConfigurationKey[conflicts.Configuration]("module.revisions.configuration")
)

// Loaded is a structurally admitted deployment-configuration snapshot.
type Loaded struct {
	snapshot   config.Snapshot
	deployment Deployment
}

// Load builds the application catalog and materializes one immutable snapshot.
func Load(options config.LoadOptions) (Loaded, error) {
	catalog, err := applicationCatalog()
	if err != nil {
		return Loaded{}, err
	}
	snapshot, err := config.LoadSnapshotWithOptions(options, catalog)
	if err != nil {
		return Loaded{}, err
	}
	return loadedFromSnapshot(snapshot)
}

// LoadPath selects an explicit deployment artifact with the generated
// Extensions-owned inactive-key policy.
func LoadPath(path string) (Loaded, error) {
	policy, err := extensionassembly.GeneratedInactiveConfigurationPolicy()
	if err != nil {
		return Loaded{}, err
	}
	return Load(config.LoadOptions{Path: path, InactivePolicy: policy})
}

// Admit strictly parses and materializes an in-memory application projection.
// It exists for composition tests; production startup selects the deployment
// artifact through Load.
func Admit(deployment Deployment) (Loaded, error) {
	catalog, err := applicationCatalog()
	if err != nil {
		return Loaded{}, err
	}
	policy, err := extensionassembly.GeneratedInactiveConfigurationPolicy()
	if err != nil {
		return Loaded{}, err
	}
	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(deployment); err != nil {
		return Loaded{}, fmt.Errorf("encode in-memory deployment configuration: %w", err)
	}
	snapshot, err := config.LoadSnapshotFromTOML(
		encoded.Bytes(),
		config.LoadOptions{InactivePolicy: policy},
		catalog,
	)
	if err != nil {
		return Loaded{}, err
	}
	return loadedFromSnapshot(snapshot)
}

func loadedFromSnapshot(snapshot config.Snapshot) (Loaded, error) {
	deployment, err := deploymentFromSnapshot(snapshot)
	if err != nil {
		return Loaded{}, fmt.Errorf("project deployment configuration: %w", err)
	}
	return Loaded{snapshot: snapshot, deployment: deployment}, nil
}

func applicationCatalog() (config.Catalog, error) {
	builder := &config.CatalogBuilder{}
	if err := telemetry.RegisterConfigurationContribution(builder); err != nil {
		return config.Catalog{}, err
	}
	if err := registerEnterpriseAuthenticationConfigurationContribution(builder); err != nil {
		return config.Catalog{}, err
	}
	if err := registerNetworkFlowConfigurationContribution(builder); err != nil {
		return config.Catalog{}, err
	}
	if err := registerRevisionsConfigurationContribution(builder); err != nil {
		return config.Catalog{}, err
	}
	return builder.Build()
}

func registerRevisionsConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[conflicts.Configuration]{
		Key:       revisionsConfigurationKey,
		Namespace: "revisions",
		Paths: []string{
			"revisions.conflict_token_key_ring_manifest_path",
		},
		Project: func(source config.Source) (conflicts.Configuration, []config.Diagnostic) {
			var configuration conflicts.Configuration
			if err := source.Decode("revisions", &configuration); err != nil {
				return conflicts.Configuration{}, []config.Diagnostic{{
					Path:       "revisions",
					ReasonCode: "revisions_conflict_token_manifest_invalid",
					Message:    err.Error(),
				}}
			}
			normalized, findings := conflicts.NormalizeAndValidateConfiguration(configuration)
			diagnostics := make([]config.Diagnostic, len(findings))
			for index, finding := range findings {
				diagnostics[index] = config.Diagnostic{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message}
			}
			return normalized, diagnostics
		},
		Clone: func(configuration conflicts.Configuration) conflicts.Configuration { return configuration },
	})
}

func registerEnterpriseAuthenticationConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[enterpriseauth.Configuration]{
		Key:       enterpriseAuthenticationConfigurationKey,
		Namespace: "enterprise_authentication",
		Paths: []string{
			"enterprise_authentication.claimed",
			"enterprise_authentication.provider_manifest_path",
		},
		ClaimPath: "enterprise_authentication.claimed",
		Project: func(source config.Source) (enterpriseauth.Configuration, []config.Diagnostic) {
			var configuration enterpriseauth.Configuration
			if err := source.Decode("enterprise_authentication", &configuration); err != nil {
				return enterpriseauth.Configuration{}, []config.Diagnostic{{
					Path:       "enterprise_authentication",
					ReasonCode: "invalid_enterprise_authentication_config",
					Message:    err.Error(),
				}}
			}
			normalized, findings := enterpriseauth.NormalizeAndValidateConfiguration(configuration)
			diagnostics := make([]config.Diagnostic, len(findings))
			for index, finding := range findings {
				diagnostics[index] = config.Diagnostic{
					Path:       finding.Path,
					ReasonCode: finding.ReasonCode,
					Message:    finding.Message,
				}
			}
			return normalized, diagnostics
		},
		Clone: func(configuration enterpriseauth.Configuration) enterpriseauth.Configuration {
			return configuration
		},
	})
}

func registerNetworkFlowConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[networkflow.Configuration]{
		Key:       networkFlowConfigurationKey,
		Namespace: "network_flow_activity",
		Paths: []string{
			"network_flow_activity.claimed",
			"network_flow_activity.key_ring_manifest_path",
		},
		ClaimPath: "network_flow_activity.claimed",
		Project: func(source config.Source) (networkflow.Configuration, []config.Diagnostic) {
			var configuration networkflow.Configuration
			if err := source.Decode("network_flow_activity", &configuration); err != nil {
				return networkflow.Configuration{}, []config.Diagnostic{{
					Path:       "network_flow_activity",
					ReasonCode: "invalid_network_flow_config",
					Message:    err.Error(),
				}}
			}
			normalized, findings := networkflow.NormalizeAndValidateConfiguration(configuration)
			diagnostics := make([]config.Diagnostic, len(findings))
			for index, finding := range findings {
				diagnostics[index] = config.Diagnostic{
					Path:       finding.Path,
					ReasonCode: finding.ReasonCode,
					Message:    finding.Message,
				}
			}
			return normalized, diagnostics
		},
		Clone: func(configuration networkflow.Configuration) networkflow.Configuration {
			return configuration
		},
	})
}

func mustConfigurationKey[T any](id string) config.Key[T] {
	key, err := config.NewKey[T](id)
	if err != nil {
		panic(fmt.Sprintf("construct application configuration key: %v", err))
	}
	return key
}

// Deployment returns a defensive application-composition projection.
func (loaded Loaded) Deployment() Deployment {
	return cloneDeployment(loaded.deployment)
}

func (loaded Loaded) Revisions() (conflicts.Configuration, error) {
	return config.Value(loaded.snapshot, revisionsConfigurationKey)
}

// ValidateForStartup performs the root-readiness phase on the retained
// immutable snapshot.
func (loaded Loaded) ValidateForStartup() error {
	return config.ValidateSnapshotForStartup(loaded.snapshot)
}

// BooleanValuesAtPaths returns the exact generated claim projection.
func (loaded Loaded) BooleanValuesAtPaths(paths []string) (map[string]bool, error) {
	return config.SnapshotBooleanValuesAtPaths(loaded.snapshot, paths)
}

func (loaded Loaded) EnterpriseAuthentication() (enterpriseauth.Configuration, error) {
	return config.Value(loaded.snapshot, enterpriseAuthenticationConfigurationKey)
}

func (loaded Loaded) NetworkFlow() (networkflow.Configuration, error) {
	return config.Value(loaded.snapshot, networkFlowConfigurationKey)
}
