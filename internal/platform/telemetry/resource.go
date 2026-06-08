package telemetry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/JochiRaider/cartulary/internal/platform/config"
)

type ResourceIdentity struct {
	SchemaURL  string
	Attributes []attribute.KeyValue
}

func BuildResourceIdentity(cfg config.Config, claimedExtensionProfiles []string) (ResourceIdentity, error) {
	profileClaims, err := SerializeProfileClaims(claimedExtensionProfiles)
	if err != nil {
		return ResourceIdentity{}, err
	}

	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.Telemetry.Resource.ServiceName),
		attribute.String("service.namespace", cfg.Telemetry.Resource.ServiceNamespace),
		attribute.String("service.version", cfg.Telemetry.Resource.ServiceVersion),
		attribute.String("service.instance.id", cfg.Telemetry.Resource.ServiceInstanceID),
		attribute.String("cartulary.deployment.profile", cfg.DeploymentProfile),
		attribute.String("cartulary.profile.claims", profileClaims),
	}
	if deploymentEnvironment := strings.TrimSpace(cfg.Telemetry.Resource.DeploymentEnvironmentName); deploymentEnvironment != "" {
		attrs = append(attrs, attribute.String("deployment.environment.name", deploymentEnvironment))
	}

	for _, attr := range attrs {
		if strings.TrimSpace(attr.Value.AsString()) == "" {
			return ResourceIdentity{}, config.NewDiagnosticsError(config.Diagnostic{
				Path:       "telemetry.resource",
				ReasonCode: "invalid_telemetry_config",
				Message:    fmt.Sprintf("resource attribute %s must not be empty", attr.Key),
			})
		}
	}
	safeAttrs := SafeAttributes(attrs...)
	if len(safeAttrs) != len(attrs) {
		return ResourceIdentity{}, config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.resource",
			ReasonCode: "invalid_telemetry_config",
			Message:    "resource identity includes an attribute outside the adopted telemetry registry",
		})
	}

	return ResourceIdentity{
		SchemaURL:  "",
		Attributes: safeAttrs,
	}, nil
}

func ValidateExternalResourceContribution(candidate ResourceIdentity) error {
	if strings.TrimSpace(candidate.SchemaURL) != "" {
		return config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.resource.schema_url",
			ReasonCode: "invalid_telemetry_config",
			Message:    "external resource schema URL contributions are not adopted",
		})
	}
	safeAttrs := SafeAttributes(candidate.Attributes...)
	if len(safeAttrs) != len(candidate.Attributes) {
		return config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.resource",
			ReasonCode: "invalid_telemetry_config",
			Message:    "external resource attributes are outside the adopted telemetry registry",
		})
	}
	return nil
}

func SerializeProfileClaims(claimedExtensionProfiles []string) (string, error) {
	claims := map[string]struct{}{"base": {}}
	for _, profileID := range claimedExtensionProfiles {
		profileID = strings.TrimSpace(profileID)
		if profileID == "" || profileID == "base" {
			continue
		}
		if !knownExtensionProfile(profileID) {
			return "", config.NewDiagnosticsError(config.Diagnostic{
				Path:       "telemetry.resource.profile_claims",
				ReasonCode: "invalid_telemetry_config",
				Message:    "profile claim is outside the current Core 00/Core 01 profile vocabulary",
			})
		}
		claims[profileID] = struct{}{}
	}

	tokens := make([]string, 0, len(claims))
	for token := range claims {
		tokens = append(tokens, token)
	}
	sort.Strings(tokens)
	return strings.Join(tokens, ","), nil
}

func IncidentHash64(secret string, incidentID string) (string, error) {
	if secret == "" {
		return "", config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.attribute.hmac_secret_ref",
			ReasonCode: "invalid_telemetry_config",
			Message:    "incident correlation secret is required",
		})
	}
	parsed, err := uuid.Parse(incidentID)
	if err != nil || parsed == uuid.Nil {
		return "", config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.attribute.incident_id",
			ReasonCode: "invalid_telemetry_config",
			Message:    "incident correlation input must be a canonical incident UUID",
		})
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strings.ToLower(parsed.String())))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum[:8]), nil
}

func knownExtensionProfile(profileID string) bool {
	switch profileID {
	case "enterprise_authentication", "import", "incident_portability", "reference_pack", "snapshot_reporting":
		return true
	default:
		return false
	}
}
