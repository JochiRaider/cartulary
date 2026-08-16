// Package configassembly owns application-level construction and
// materialization of the deployment-configuration contribution catalog.
package configassembly

import (
	"fmt"

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
	snapshot        config.Snapshot
	deployment      Deployment
	requestedClaims extensionassembly.RequestedClaims
}

// LoadOptions contains only operator-selected artifact inputs. Application
// assembly owns the generated Extensions policy supplied to the kernel.
type LoadOptions struct {
	Path string
	Env  map[string]string
}

// Load builds the application catalog and materializes one immutable snapshot.
func Load(options LoadOptions) (Loaded, error) {
	policy, err := extensionassembly.GeneratedConfigurationPolicy()
	if err != nil {
		return Loaded{}, err
	}
	catalog, err := applicationCatalog()
	if err != nil {
		return Loaded{}, err
	}
	snapshot, err := config.LoadSnapshotWithOptions(config.LoadOptions{
		Path:            options.Path,
		Env:             options.Env,
		ExtensionPolicy: policy,
	}, catalog)
	if err != nil {
		return Loaded{}, err
	}
	return loadedFromSnapshot(snapshot, policy)
}

func loadedFromSnapshot(snapshot config.Snapshot, policy extensionassembly.ConfigurationPolicy) (Loaded, error) {
	requestedClaims, err := policy.MaterializeRequestedClaims(config.RequestedClaimRegistrationIDs(snapshot))
	if err != nil {
		return Loaded{}, fmt.Errorf("materialize requested extension claims: %w", err)
	}
	deployment, err := deploymentFromSnapshot(snapshot)
	if err != nil {
		return Loaded{}, fmt.Errorf("project deployment configuration: %w", err)
	}
	return Loaded{snapshot: snapshot, deployment: deployment, requestedClaims: requestedClaims}, nil
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
		Decode: func(decoder config.NamespaceDecoder) (conflicts.Configuration, []config.Diagnostic) {
			var configuration conflicts.Configuration
			if err := decoder.Decode(&configuration); err != nil {
				return conflicts.Configuration{}, []config.Diagnostic{{
					Path:       "revisions",
					ReasonCode: "revisions_conflict_token_manifest_invalid",
					Message:    err.Error(),
				}}
			}
			return configuration, nil
		},
		ApplyOverlay: func(configuration conflicts.Configuration, segments []string, raw string) (conflicts.Configuration, *config.Diagnostic) {
			updated, finding := conflicts.ApplyConfigurationOverlay(configuration, segments, raw)
			if finding == nil {
				return updated, nil
			}
			return updated, &config.Diagnostic{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message}
		},
		Project: func(configuration conflicts.Configuration, _ config.NamespacePresence) (conflicts.Configuration, []config.Diagnostic) {
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
		Decode: func(decoder config.NamespaceDecoder) (enterpriseauth.Configuration, []config.Diagnostic) {
			var configuration enterpriseauth.Configuration
			if err := decoder.Decode(&configuration); err != nil {
				return enterpriseauth.Configuration{}, []config.Diagnostic{{
					Path:       "enterprise_authentication",
					ReasonCode: "invalid_enterprise_authentication_config",
					Message:    err.Error(),
				}}
			}
			return configuration, nil
		},
		ApplyOverlay: func(configuration enterpriseauth.Configuration, segments []string, raw string) (enterpriseauth.Configuration, *config.Diagnostic) {
			updated, finding := enterpriseauth.ApplyConfigurationOverlay(configuration, segments, raw)
			if finding == nil {
				return updated, nil
			}
			return updated, &config.Diagnostic{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message}
		},
		Project: func(configuration enterpriseauth.Configuration, _ config.NamespacePresence) (enterpriseauth.Configuration, []config.Diagnostic) {
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
			"network_flow_activity.resource_limits",
		},
		Decode: func(decoder config.NamespaceDecoder) (networkflow.Configuration, []config.Diagnostic) {
			var configuration networkflow.Configuration
			if err := decoder.Decode(&configuration); err != nil {
				return networkflow.Configuration{}, []config.Diagnostic{{
					Path:       "network_flow_activity",
					ReasonCode: "invalid_network_flow_config",
					Message:    err.Error(),
				}}
			}
			return configuration, nil
		},
		ApplyOverlay: func(configuration networkflow.Configuration, segments []string, raw string) (networkflow.Configuration, *config.Diagnostic) {
			updated, finding := networkflow.ApplyConfigurationOverlay(configuration, segments, raw)
			if finding == nil {
				return updated, nil
			}
			return updated, &config.Diagnostic{Path: finding.Path, ReasonCode: finding.ReasonCode, Message: finding.Message}
		},
		Project: func(configuration networkflow.Configuration, _ config.NamespacePresence) (networkflow.Configuration, []config.Diagnostic) {
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
			return networkflow.CloneConfiguration(configuration)
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

// ValidateForStartup performs the root-readiness phase on the retained
// immutable snapshot.
func (loaded Loaded) ValidateForStartup() error {
	return config.ValidateSnapshotForStartup(loaded.snapshot)
}

// RequestedClaims returns an immutable typed request distinct from coordinator
// resolution and the published resolved-claim-set identity.
func (loaded Loaded) RequestedClaims() extensionassembly.RequestedClaims {
	return loaded.requestedClaims
}
