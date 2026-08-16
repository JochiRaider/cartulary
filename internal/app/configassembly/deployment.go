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
	ConfigSchemaID           string
	DeploymentProfile        string
	Application              config.ApplicationConfig
	Roots                    config.RootBindings
	Bootstrap                config.BootstrapConfig
	EnterpriseAuthentication enterpriseauth.Configuration
	NetworkFlowActivity      networkflow.Configuration
	Revisions                conflicts.Configuration
	Timeouts                 config.TimeoutConfig
	Intervals                config.IntervalConfig
	Limits                   config.LimitConfig
	Telemetry                telemetryconfiguration.Config
}

func deploymentFromSnapshot(snapshot config.Snapshot) (Deployment, error) {
	core := snapshot.Core()
	deployment := Deployment{
		ConfigSchemaID:    core.ConfigSchemaID,
		DeploymentProfile: core.DeploymentProfile,
		Application:       core.Application,
		Roots:             core.Roots,
		Bootstrap:         core.Bootstrap,
		Timeouts:          core.Timeouts,
		Intervals:         core.Intervals,
		Limits:            core.Limits,
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
	cloned.NetworkFlowActivity = networkflow.CloneConfiguration(source.NetworkFlowActivity)
	cloned.Telemetry = telemetryconfiguration.Clone(source.Telemetry)
	return cloned
}
