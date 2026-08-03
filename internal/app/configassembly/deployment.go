package configassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/enterpriseauth"
	"github.com/JochiRaider/cartulary/internal/platform/telemetry"
	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

// Deployment is the application-composition projection of one immutable
// configuration snapshot. It is not passed to modules, HTTP transport, or
// platform adapters; application facades map it into their narrow settings.
type Deployment struct {
	ConfigSchemaID           string                        `toml:"config_schema_id,omitempty"`
	DeploymentProfile        string                        `toml:"deployment_profile,omitempty"`
	Application              config.ApplicationConfig      `toml:"application,omitempty"`
	Roots                    config.RootBindings           `toml:"roots,omitempty"`
	Bootstrap                config.BootstrapConfig        `toml:"bootstrap,omitempty"`
	EnterpriseAuthentication enterpriseauth.Configuration  `toml:"enterprise_authentication,omitempty"`
	Import                   config.ClaimConfig            `toml:"import,omitempty"`
	IncidentPortability      config.ClaimConfig            `toml:"incident_portability,omitempty"`
	NetworkFlowActivity      networkflow.Configuration     `toml:"network_flow_activity,omitempty"`
	ReferencePack            config.ClaimConfig            `toml:"reference_pack,omitempty"`
	Revisions                conflicts.Configuration       `toml:"revisions,omitempty"`
	SnapshotReporting        config.ClaimConfig            `toml:"snapshot_reporting,omitempty"`
	Timeouts                 config.TimeoutConfig          `toml:"timeouts,omitempty"`
	Intervals                config.IntervalConfig         `toml:"intervals,omitempty"`
	Limits                   config.LimitConfig            `toml:"limits,omitempty"`
	Telemetry                telemetryconfiguration.Config `toml:"telemetry,omitempty"`
}

func deploymentFromSnapshot(snapshot config.Snapshot) (Deployment, error) {
	var deployment Deployment
	for _, section := range []struct {
		path        string
		destination any
	}{
		{"config_schema_id", &deployment.ConfigSchemaID},
		{"deployment_profile", &deployment.DeploymentProfile},
		{"application", &deployment.Application},
		{"roots", &deployment.Roots},
		{"bootstrap", &deployment.Bootstrap},
		{"import", &deployment.Import},
		{"incident_portability", &deployment.IncidentPortability},
		{"reference_pack", &deployment.ReferencePack},
		{"snapshot_reporting", &deployment.SnapshotReporting},
		{"timeouts", &deployment.Timeouts},
		{"intervals", &deployment.Intervals},
		{"limits", &deployment.Limits},
	} {
		if err := snapshot.Decode(section.path, section.destination); err != nil {
			return Deployment{}, err
		}
	}

	var err error
	deployment.EnterpriseAuthentication, err = config.Value(snapshot, enterpriseAuthenticationConfigurationKey)
	if err != nil {
		return Deployment{}, err
	}
	deployment.NetworkFlowActivity, err = config.Value(snapshot, networkFlowConfigurationKey)
	if err != nil {
		return Deployment{}, err
	}
	deployment.Revisions, err = config.Value(snapshot, revisionsConfigurationKey)
	if err != nil {
		return Deployment{}, err
	}
	deployment.Telemetry, err = telemetry.ConfigurationValue(snapshot)
	if err != nil {
		return Deployment{}, err
	}
	return deployment, nil
}

func cloneDeployment(source Deployment) Deployment {
	cloned := source
	cloned.Telemetry = telemetryconfiguration.Clone(source.Telemetry)
	return cloned
}
