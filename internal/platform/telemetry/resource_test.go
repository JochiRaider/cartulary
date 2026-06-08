package telemetry

import (
	"regexp"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestResourceIdentityClosedRegistry(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	cfg.Telemetry.Resource.DeploymentEnvironmentName = "test"
	resource, err := BuildResourceIdentity(cfg, []string{
		"snapshot_reporting",
		"import",
		"reference_pack",
		"incident_portability",
		"import",
	})
	if err != nil {
		t.Fatalf("build resource identity: %v", err)
	}
	if resource.SchemaURL != "" {
		t.Fatalf("resource schema URL must be empty, got %q", resource.SchemaURL)
	}

	attrs := attributesByKey(resource)
	wantKeys := []string{
		"cartulary.deployment.profile",
		"cartulary.profile.claims",
		"deployment.environment.name",
		"service.instance.id",
		"service.name",
		"service.namespace",
		"service.version",
	}
	if len(attrs) != len(wantKeys) {
		t.Fatalf("unexpected resource attributes: %#v", attrs)
	}
	for _, key := range wantKeys {
		if attrs[key] == "" {
			t.Fatalf("missing or empty resource attribute %s in %#v", key, attrs)
		}
	}
	if attrs["cartulary.profile.claims"] != "base,import,incident_portability,reference_pack,snapshot_reporting" {
		t.Fatalf("unexpected profile claim serialization: %q", attrs["cartulary.profile.claims"])
	}
	if strings.Contains(strings.Join(mapValues(attrs), ","), "host.") || strings.Contains(strings.Join(mapValues(attrs), ","), "process.") {
		t.Fatalf("resource leaked non-adopted detector family values: %#v", attrs)
	}
}

func TestResourceIdentityOmitsOptionalNullDeploymentEnvironment(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	cfg.Telemetry.Resource.DeploymentEnvironmentName = ""
	resource, err := BuildResourceIdentity(cfg, nil)
	if err != nil {
		t.Fatalf("build resource identity: %v", err)
	}
	if attrs := attributesByKey(resource); attrs["deployment.environment.name"] != "" {
		t.Fatalf("optional null deployment environment must be omitted, got %#v", attrs)
	}
}

func TestExternalResourceContributionRejectsSchemaURLAndDetectorAttributes(t *testing.T) {
	if err := ValidateExternalResourceContribution(ResourceIdentity{}); err != nil {
		t.Fatalf("empty external resource contribution should be harmless: %v", err)
	}
	if err := ValidateExternalResourceContribution(ResourceIdentity{SchemaURL: "https://opentelemetry.io/schemas/1.41.0"}); err == nil {
		t.Fatal("non-empty external resource schema URL should fail")
	}
	if err := ValidateExternalResourceContribution(ResourceIdentity{
		Attributes: []attribute.KeyValue{
			attribute.String("host.name", "prod-node-1"),
		},
	}); err == nil {
		t.Fatal("host/process/container/cloud detector attributes should fail")
	}
	if err := ValidateExternalResourceContribution(ResourceIdentity{
		Attributes: []attribute.KeyValue{
			attribute.String("service.name", "cartulary.app"),
			attribute.String("service.namespace", "cartulary"),
		},
	}); err != nil {
		t.Fatalf("closed external resource attributes should validate: %v", err)
	}
}

func TestResourceInstanceIDOpacityPredicate(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	resource, err := BuildResourceIdentity(cfg, nil)
	if err != nil {
		t.Fatalf("build resource identity with generated instance id: %v", err)
	}
	attrs := attributesByKey(resource)
	if !safeUUIDV4(attrs["service.instance.id"]) {
		t.Fatalf("generated service.instance.id must be canonical lowercase non-nil UUID v4, got %q", attrs["service.instance.id"])
	}

	cfg.Telemetry.Resource.ServiceInstanceID = "10000000-0000-4000-8000-000000000001"
	resource, err = BuildResourceIdentity(cfg, nil)
	if err != nil {
		t.Fatalf("build resource identity with configured opaque instance id: %v", err)
	}
	if attrs := attributesByKey(resource); attrs["service.instance.id"] != cfg.Telemetry.Resource.ServiceInstanceID {
		t.Fatalf("configured service.instance.id was not exported exactly once: %#v", attrs)
	}

	for _, invalid := range []string{
		"host-prod-1",
		"00000000-0000-0000-0000-000000000000",
		"10000000-0000-1000-8000-000000000001",
		"10000000-0000-4000-8000-0000000000AA",
	} {
		cfg.Telemetry.Resource.ServiceInstanceID = invalid
		if _, err := BuildResourceIdentity(cfg, nil); err == nil {
			t.Fatalf("unsafe service.instance.id %q should fail resource construction", invalid)
		}
	}
}

func TestResourceIdentityRejectsUnknownProfileClaim(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	_, err := BuildResourceIdentity(cfg, []string{"unsupported_profile"})
	if err == nil {
		t.Fatal("expected unknown profile claim to fail")
	}
}

func TestRuntimeResourceIdentityIsCopied(t *testing.T) {
	cfg := validTelemetryBootstrapConfig(t)
	runtime, err := Bootstrap(t.Context(), cfg, nil, WithClaimedExtensionProfiles([]string{"import"}))
	if err != nil {
		t.Fatalf("bootstrap telemetry runtime: %v", err)
	}

	first := runtime.ResourceIdentity()
	first.Attributes[0] = first.Attributes[1]
	second := runtime.ResourceIdentity()
	attrs := attributesByKey(second)
	if attrs["service.name"] != cfg.Telemetry.Resource.ServiceName {
		t.Fatalf("runtime resource identity was mutated through returned slice: %#v", attrs)
	}
	if attrs["cartulary.profile.claims"] != "base,import" {
		t.Fatalf("unexpected runtime profile claims: %q", attrs["cartulary.profile.claims"])
	}
}

func TestIncidentHash64Shape(t *testing.T) {
	hash, err := IncidentHash64("safe-test-secret", "10000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatalf("hash incident id: %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(hash) {
		t.Fatalf("incident hash must be 16 lowercase hex characters, got %q", hash)
	}
	if strings.Contains(hash, "10000000") || strings.Contains(hash, "safe-test-secret") {
		t.Fatalf("incident hash leaked raw input: %q", hash)
	}
}

func attributesByKey(resource ResourceIdentity) map[string]string {
	attrs := make(map[string]string, len(resource.Attributes))
	for _, attr := range resource.Attributes {
		attrs[string(attr.Key)] = attr.Value.AsString()
	}
	return attrs
}

func mapValues(values map[string]string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}
