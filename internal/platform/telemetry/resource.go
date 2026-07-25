package telemetry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	telemetryconfiguration "github.com/JochiRaider/cartulary/internal/platform/telemetry/configuration"
)

type ResourceIdentity struct {
	SchemaURL  string
	Attributes []attribute.KeyValue
}

type ResolvedClaimIdentity struct {
	ProfileIDs []string
	SHA256     string
}

var resolvedClaimProfileIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

func BuildResourceIdentity(cfg telemetryconfiguration.Config, deploymentProfile string, resolvedClaims ResolvedClaimIdentity) (ResourceIdentity, error) {
	profileClaims, err := SerializeProfileClaims(resolvedClaims)
	if err != nil {
		return ResourceIdentity{}, err
	}

	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.Resource.ServiceName),
		attribute.String("service.namespace", cfg.Resource.ServiceNamespace),
		attribute.String("service.version", cfg.Resource.ServiceVersion),
		attribute.String("service.instance.id", cfg.Resource.ServiceInstanceID),
		attribute.String("cartulary.deployment.profile", deploymentProfile),
		attribute.String("cartulary.profile.claims", profileClaims),
	}
	if deploymentEnvironment := strings.TrimSpace(cfg.Resource.DeploymentEnvironmentName); deploymentEnvironment != "" {
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

func SerializeProfileClaims(resolvedClaims ResolvedClaimIdentity) (string, error) {
	if err := validateResolvedClaimIdentity(resolvedClaims); err != nil {
		return "", err
	}
	tokens := append([]string{"base"}, resolvedClaims.ProfileIDs...)
	sort.Strings(tokens)
	return strings.Join(tokens, ","), nil
}

func validateResolvedClaimIdentity(identity ResolvedClaimIdentity) error {
	fail := func(message string) error {
		return config.NewDiagnosticsError(config.Diagnostic{
			Path:       "telemetry.resource.profile_claims",
			ReasonCode: "invalid_telemetry_config",
			Message:    message,
		})
	}
	if len(identity.SHA256) != sha256.Size*2 {
		return fail("resolved claim identity requires a lowercase SHA-256 digest")
	}
	digestBytes, err := hex.DecodeString(identity.SHA256)
	if err != nil || hex.EncodeToString(digestBytes) != identity.SHA256 {
		return fail("resolved claim identity requires a lowercase SHA-256 digest")
	}
	for index, profileID := range identity.ProfileIDs {
		if profileID == "base" || !resolvedClaimProfileIDPattern.MatchString(profileID) {
			return fail("resolved claim identity contains an invalid profile identifier")
		}
		if index > 0 && identity.ProfileIDs[index-1] >= profileID {
			return fail("resolved claim identity profile identifiers must be unique and canonically ordered")
		}
	}
	canonicalBytes, err := json.Marshal(struct {
		ProfileIDs []string `json:"profile_ids"`
	}{ProfileIDs: append([]string{}, identity.ProfileIDs...)})
	if err != nil {
		return fail("resolved claim identity cannot be encoded")
	}
	computed := sha256.Sum256(canonicalBytes)
	if !hmac.Equal(computed[:], digestBytes) {
		return fail("resolved claim identity digest does not match its canonical profile set")
	}
	return nil
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
