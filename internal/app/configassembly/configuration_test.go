package configassembly

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestArtifactAdmissionDefaultsAndProjectionIsolation_Unit(t *testing.T) {
	wantLoaded, err := loadProjection(t, nil)
	if err != nil {
		t.Fatalf("load deployment defaults: %v", err)
	}
	wantLimits := wantLoaded.Deployment().Limits
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
