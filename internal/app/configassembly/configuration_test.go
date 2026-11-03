package configassembly

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestArtifactAdmissionDefaultsAndProjectionIsolation_Unit(t *testing.T) {
	t.Run("exact typed value schemas", checkTypedValueSchemaBindings)
	wantLoaded, err := loadProjection(t, nil)
	if err != nil {
		t.Fatalf("load deployment defaults: %v", err)
	}
	t.Run("extension budgets survive immutable composition", func(t *testing.T) {
		defaults := wantLoaded.Deployment().Timeouts.Extensions
		if defaults.ProfileMigrationSeconds != 3600 || defaults.ValidationSeconds != 60 || defaults.DependencyProbeSeconds != 30 {
			t.Fatalf("composed budget defaults = %#v", defaults)
		}
		loaded, err := loadProjection(t, map[string]string{
			"CARTULARY__TIMEOUTS__EXTENSIONS__PROFILE_MIGRATION_SECONDS": "14400",
			"CARTULARY__TIMEOUTS__EXTENSIONS__VALIDATION_SECONDS":        "600",
			"CARTULARY__TIMEOUTS__EXTENSIONS__DEPENDENCY_PROBE_SECONDS":  "300",
		})
		if err != nil {
			t.Fatal(err)
		}
		first := loaded.Deployment()
		if first.Timeouts.Extensions.ProfileMigrationSeconds != 14400 || first.Timeouts.Extensions.ValidationSeconds != 600 || first.Timeouts.Extensions.DependencyProbeSeconds != 300 {
			t.Fatal("admitted budget was lost in composition")
		}
		first.Timeouts.Extensions.DependencyProbeSeconds = 1
		if loaded.Deployment().Timeouts.Extensions.DependencyProbeSeconds != 300 {
			t.Fatal("composition exposed mutable timeout state")
		}
	})
	t.Run("inactive Reference Pack settings retain the registered profile identity", func(t *testing.T) {
		_, err := loadProjection(t, map[string]string{
			"CARTULARY__REFERENCE_PACK__CLAIMED":               "false",
			"CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED":        "true",
			"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": "/fixture/must-not-be-read.json",
		})
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 2 {
			t.Fatalf("inactive reference settings: %#v / %v", diagnostics, err)
		}
		for _, diagnostic := range diagnostics {
			if diagnostic.ReasonCode != "extension_config_without_claim" || diagnostic.Details["profile_id"] != "reference_pack" || diagnostic.Details["config_path"] != "$."+diagnostic.Path {
				t.Fatalf("namespace replaced registered owner identity: %#v", diagnostic)
			}
		}
	})
	wantLimits := wantLoaded.Deployment().Limits
	if wantLimits.ReferencePacks.MaxVerificationSeconds != 1800 || wantLimits.ReferencePacks.MaxContainerBytes != 536870912 || wantLoaded.Deployment().ReferencePacks.ClockTrusted {
		t.Fatal("Reference Pack defaults changed")
	}
	for _, value := range []string{"59", "86401", "null", "60.5"} {
		if _, err := loadProjection(t, map[string]string{"CARTULARY__LIMITS__REFERENCE_PACKS__MAX_VERIFICATION_SECONDS": value}); err == nil {
			t.Fatalf("invalid verification budget %s admitted", value)
		}
	}
	for _, value := range []string{"60", "86400"} {
		if _, err := loadProjection(t, map[string]string{"CARTULARY__LIMITS__REFERENCE_PACKS__MAX_VERIFICATION_SECONDS": value}); err != nil {
			t.Fatalf("inclusive verification budget %s: %v", value, err)
		}
	}
	trusted, err := loadProjection(t, map[string]string{"CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"})
	if err != nil || !trusted.Deployment().ReferencePacks.ClockTrusted || !trusted.Deployment().ReferencePacks.ClockTrustedExplicit {
		t.Fatal("explicit clock assertion lost", err)
	}
	content := string(fixtures.MustRead("config", "valid.toml"))
	content, _, _ = strings.Cut(content, "\n[limits.object_blobs]\n")
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write limits-omitted config: %v", err)
	}

	loaded, err := Load(LoadOptions{Path: path})
	if err != nil {
		t.Fatalf("load deployment with omitted limits: %v", err)
	}
	if got := loaded.Deployment().Limits; !reflect.DeepEqual(got, wantLimits) {
		t.Fatalf("omitted artifact limits = %#v, want %#v", got, wantLimits)
	}
	for _, profileID := range loaded.RequestedClaims().ProfileIDs() {
		if profileID == "enterprise_authentication" || profileID == "network_flow_activity" {
			t.Fatalf("unexpected contributed claim %q", profileID)
		}
	}

	first := loaded.Deployment()
	first.Telemetry.Exporter.Headers["mutated"] = first.Telemetry.Attribute.HMACSecretRef
	if _, retained := loaded.Deployment().Telemetry.Exporter.Headers["mutated"]; retained {
		t.Fatal("Deployment returned mutable telemetry owner state")
	}
}

func TestReportingConfigurationProjection_Unit(t *testing.T) {
	const key = "CARTULARY__SNAPSHOT_REPORTING__RUNTIME_RECEIPT_PATH"
	const path = "/fixture/report-renderer/installation.json"
	loaded, err := loadProjection(t, map[string]string{key: path})
	if err != nil {
		t.Fatal(err)
	}
	if value := loaded.Deployment().Reporting; !value.Claimed || value.RuntimeReceiptPath != path {
		t.Fatal("Reporting claim or receipt projection lost")
	}
	copy := loaded.Deployment()
	copy.Reporting.RuntimeReceiptPath = "/changed/installation.json"
	if loaded.Deployment().Reporting.RuntimeReceiptPath != path {
		t.Fatal("Reporting configuration is mutable")
	}
	for _, value := range []string{"", "relative/installation.json", "/fixture/../installation.json", "/fixture//installation.json", "/fixture/other.json", "/installation.json"} {
		if _, err := loadProjection(t, map[string]string{key: value}); err == nil {
			t.Fatalf("admitted invalid receipt %q", value)
		}
	}
	_, err = loadProjection(t, map[string]string{key: path, "CARTULARY__SNAPSHOT_REPORTING__CLAIMED": "false"})
	diagnostics, ok := config.DiagnosticsFromError(err)
	if !ok || len(diagnostics) != 1 || diagnostics[0].ReasonCode != "extension_config_without_claim" || strings.Contains(err.Error(), path) {
		t.Fatalf("inactive receipt diagnostic: %v", err)
	}
}

func TestEnterpriseAuthenticationConfigurationProjection_Unit(t *testing.T) {
	t.Run("claimed configuration is owner-validated before preflight", func(t *testing.T) {
		_, err := loadProjection(t, map[string]string{
			"CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED":                "true",
			"CARTULARY__ENTERPRISE_AUTHENTICATION__PROVIDER_MANIFEST_PATH": "",
		})
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 ||
			diagnostics[0].ReasonCode != "provider_manifest_path_missing" {
			t.Fatalf("claimed configuration diagnostics = %#v / %v", diagnostics, err)
		}
	})

	t.Run("normalized owner value is immutable and mapped into the temporary projection", func(t *testing.T) {
		loaded, err := loadProjection(t, map[string]string{
			"CARTULARY__ENTERPRISE_AUTHENTICATION__CLAIMED":                "true",
			"CARTULARY__ENTERPRISE_AUTHENTICATION__PROVIDER_MANIFEST_PATH": "/etc//cartulary/enterprise-auth-providers.json",
		})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		configuration := loaded.Deployment().EnterpriseAuthentication
		const expected = "/etc/cartulary/enterprise-auth-providers.json"
		if !configuration.Claimed || configuration.ProviderManifestPath != expected {
			t.Fatalf("owner configuration = %#v", configuration)
		}
		configuration.ProviderManifestPath = "/mutated"
		again := loaded.Deployment().EnterpriseAuthentication
		if again.ProviderManifestPath != expected {
			t.Fatalf("snapshot owner value was mutable: %#v", again)
		}
		if projection := loaded.Deployment(); projection.EnterpriseAuthentication.ProviderManifestPath != expected {
			t.Fatalf("application projection did not use owner value: %#v", projection.EnterpriseAuthentication)
		}
	})
}

func TestRevisionsConfigurationProjection_Unit(t *testing.T) {
	_, err := loadProjection(t, map[string]string{
		"CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH": "",
	})
	diagnostics, ok := config.DiagnosticsFromError(err)
	if !ok || len(diagnostics) != 1 || diagnostics[0].ReasonCode != "revisions_conflict_token_manifest_missing" {
		t.Fatalf("missing Revisions configuration diagnostics = %#v / %v", diagnostics, err)
	}

	if _, err := loadProjection(t, map[string]string{
		"CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH": "/etc//cartulary/revisions-conflict-token-key-ring.json",
	}); err == nil {
		t.Fatal("unnormalized Revisions key-ring path was admitted")
	}
	loaded, err := loadProjection(t, map[string]string{
		"CARTULARY__REVISIONS__CONFLICT_TOKEN_KEY_RING_MANIFEST_PATH": "/etc/cartulary/revisions-conflict-token-key-ring.json",
	})
	if err != nil {
		t.Fatalf("admit Revisions configuration: %v", err)
	}
	configuration := loaded.Deployment().Revisions
	if configuration.ConflictTokenKeyRingManifestPath != "/etc/cartulary/revisions-conflict-token-key-ring.json" {
		t.Fatalf("Revisions configuration projection = %#v", configuration)
	}
	configuration.ConflictTokenKeyRingManifestPath = "/mutated"
	again := loaded.Deployment().Revisions
	if again.ConflictTokenKeyRingManifestPath != "/etc/cartulary/revisions-conflict-token-key-ring.json" {
		t.Fatalf("Revisions configuration snapshot was mutable: %#v", again)
	}
}

func TestNetworkFlowConfigurationProjection_Unit(t *testing.T) {
	t.Run("claimed configuration is owner-validated before preflight", func(t *testing.T) {
		_, err := loadProjection(t, map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": "",
		})
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 2 ||
			diagnostics[0].ReasonCode != "network_flow_cursor_key_missing" ||
			diagnostics[1].ReasonCode != "network_flow_safe_digest_key_missing" {
			t.Fatalf("claimed configuration diagnostics = %#v / %v", diagnostics, err)
		}
	})

	t.Run("normalized owner value is immutable and mapped into the temporary projection", func(t *testing.T) {
		loaded, err := loadProjection(t, map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": "/etc//cartulary/network-flow-key-rings.json",
		})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		configuration := loaded.Deployment().NetworkFlowActivity
		const expected = "/etc/cartulary/network-flow-key-rings.json"
		if !configuration.Claimed || configuration.KeyRingManifestPath != expected {
			t.Fatalf("owner configuration = %#v", configuration)
		}
		configuration.KeyRingManifestPath = "/mutated"
		again := loaded.Deployment().NetworkFlowActivity
		if again.KeyRingManifestPath != expected {
			t.Fatalf("snapshot owner value was mutable: %#v", again)
		}
		if projection := loaded.Deployment(); projection.NetworkFlowActivity.KeyRingManifestPath != expected {
			t.Fatalf("application projection did not use owner value: %#v", projection.NetworkFlowActivity)
		}
	})

	t.Run("resource limits resolve once and remain immutable", func(t *testing.T) {
		loaded, err := loadProjection(t, map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": "/etc/cartulary/network-flow-key-rings.json",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS":        `{"max_graph_vertices":7000,"max_graph_edges":0,"max_rejected_row_diagnostics":0}`,
		})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		configuration := loaded.Deployment().NetworkFlowActivity
		limits := configuration.EffectiveResourceLimits()
		if limits.MaxGraphVertices != 7000 || limits.MaxGraphEdges != 0 ||
			limits.MaxRejectedRowDiagnostics != 0 || limits.MaxHeaderScalarLength != 256 {
			t.Fatalf("effective limits = %#v", limits)
		}
		*configuration.ResourceLimits.MaxGraphVertices = 8000
		again := loaded.Deployment().NetworkFlowActivity
		if *again.ResourceLimits.MaxGraphVertices != 7000 || again.EffectiveResourceLimits().MaxGraphVertices != 7000 {
			t.Fatalf("configuration snapshot resource limits were mutable: %#v", again)
		}
	})

	t.Run("TOML resource-limit table projects through the same owner policy", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml")) + `

[network_flow_activity]
claimed = true
key_ring_manifest_path = "/etc/cartulary/network-flow-key-rings.json"

[network_flow_activity.resource_limits]
max_graph_vertices = 9000
max_graph_edges = 0
max_time_buckets_per_graph = 512
`
		path := filepath.Join(t.TempDir(), "network-flow-resource-limits.toml")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("write resource-limit config: %v", err)
		}
		loaded, err := Load(LoadOptions{Path: path})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		limits := loaded.Deployment().NetworkFlowActivity.EffectiveResourceLimits()
		if limits.MaxGraphVertices != 9000 || limits.MaxGraphEdges != 0 || limits.MaxTimeBucketsPerGraph != 512 || limits.MaxHeaderScalarLength != 256 {
			t.Fatalf("TOML effective limits = %#v", limits)
		}
	})

	t.Run("invalid resource-limit objects fail before preflight", func(t *testing.T) {
		for name, raw := range map[string]string{
			"null":         "null",
			"unknown":      `{"unknown":1}`,
			"non-integer":  `{"max_graph_vertices":1.5}`,
			"out-of-range": `{"max_graph_vertices":100001}`,
			"cross-limit":  `{"max_active_tables_per_incident":2,"max_retained_tables_per_incident":1}`,
		} {
			t.Run(name, func(t *testing.T) {
				_, err := loadProjection(t, map[string]string{
					"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
					"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": "/etc/cartulary/network-flow-key-rings.json",
					"CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS":        raw,
				})
				if err == nil {
					t.Fatalf("invalid resource-limit object %s was accepted", raw)
				}
			})
		}
	})

	t.Run("resource limits are forbidden while unclaimed", func(t *testing.T) {
		_, err := loadProjection(t, map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":         "false",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__RESOURCE_LIMITS": `{"max_graph_vertices":7000}`,
		})
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 ||
			diagnostics[0].Path != "network_flow_activity.resource_limits" ||
			diagnostics[0].ReasonCode != "extension_config_without_claim" {
			t.Fatalf("inactive resource-limit diagnostics = %#v / %v", diagnostics, err)
		}
	})
}

func loadProjection(t testing.TB, env map[string]string) (Loaded, error) {
	t.Helper()
	return Load(LoadOptions{
		Path: fixtures.Path("config", "valid.toml"),
		Env:  env,
	})
}

func checkTypedValueSchemaBindings(t *testing.T) {
	catalog, err := applicationCatalog()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := extensionassembly.GeneratedConfigurationPolicy()
	if err != nil {
		t.Fatal(err)
	}
	baseline := policy.ValueSchemaBindings()
	if len(baseline) != 6 {
		t.Fatal("recognized configuration contracts were omitted")
	}
	if err := catalog.ValidateSchemaBindings(baseline); err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"missing key", "extra key", "wrong schema", "wrong namespace", "duplicate namespace", "duplicate key"} {
		t.Run(mutation, func(t *testing.T) {
			groups := policy.ValueSchemaBindings()
			for i := range groups {
				if groups[i].Namespace != "network_flow_activity" {
					continue
				}
				switch mutation {
				case "missing key":
					groups[i].Values = groups[i].Values[:1]
				case "extra key":
					groups[i].Values = append(groups[i].Values, config.ValueSchemaBinding{Path: "network_flow_activity.unregistered", SchemaID: "cartulary.unregistered.v1"})
				case "wrong schema":
					groups[i].Values[0].SchemaID = "cartulary.reference_pack_clock_trusted.v1"
				case "wrong namespace":
					groups[i].Namespace = "other"
				case "duplicate namespace":
					groups = append(groups, groups[i])
				case "duplicate key":
					groups[i].Values = append(groups[i].Values, groups[i].Values[0])
				}
				break
			}
			if catalog.ValidateSchemaBindings(groups) == nil {
				t.Fatal("mismatched actual owner schema was admitted")
			}
			if !reflect.DeepEqual(policy.ValueSchemaBindings(), baseline) {
				t.Fatal("schema references escaped immutable policy")
			}
		})
	}
}
