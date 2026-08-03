package configassembly

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestEnterpriseAuthenticationConfigurationProjection_Unit(t *testing.T) {
	t.Run("claimed configuration is owner-validated before preflight", func(t *testing.T) {
		deployment := validProjectionDeployment(t)
		deployment.EnterpriseAuthentication.Claimed = true
		deployment.EnterpriseAuthentication.ProviderManifestPath = ""
		_, err := Admit(deployment)
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 1 ||
			diagnostics[0].ReasonCode != "provider_manifest_path_missing" {
			t.Fatalf("claimed configuration diagnostics = %#v / %v", diagnostics, err)
		}
	})

	t.Run("normalized owner value is immutable and mapped into the temporary projection", func(t *testing.T) {
		deployment := validProjectionDeployment(t)
		deployment.EnterpriseAuthentication.Claimed = true
		deployment.EnterpriseAuthentication.ProviderManifestPath = "/etc//cartulary/enterprise-auth-providers.json"
		loaded, err := Admit(deployment)
		if err != nil {
			t.Fatalf("Admit: %v", err)
		}
		configuration, err := loaded.EnterpriseAuthentication()
		if err != nil {
			t.Fatalf("EnterpriseAuthentication: %v", err)
		}
		const expected = "/etc/cartulary/enterprise-auth-providers.json"
		if !configuration.Claimed || configuration.ProviderManifestPath != expected {
			t.Fatalf("owner configuration = %#v", configuration)
		}
		configuration.ProviderManifestPath = "/mutated"
		again, err := loaded.EnterpriseAuthentication()
		if err != nil || again.ProviderManifestPath != expected {
			t.Fatalf("snapshot owner value was mutable: %#v / %v", again, err)
		}
		if projection := loaded.Deployment(); projection.EnterpriseAuthentication.ProviderManifestPath != expected {
			t.Fatalf("application projection did not use owner value: %#v", projection.EnterpriseAuthentication)
		}
	})
}

func TestRevisionsConfigurationProjection_Unit(t *testing.T) {
	deployment := validProjectionDeployment(t)
	deployment.Revisions.ConflictTokenKeyRingManifestPath = ""
	_, err := Admit(deployment)
	diagnostics, ok := config.DiagnosticsFromError(err)
	if !ok || len(diagnostics) != 1 || diagnostics[0].ReasonCode != "revisions_conflict_token_manifest_missing" {
		t.Fatalf("missing Revisions configuration diagnostics = %#v / %v", diagnostics, err)
	}

	deployment.Revisions.ConflictTokenKeyRingManifestPath = "/etc//cartulary/revisions-conflict-token-key-ring.json"
	if _, err := Admit(deployment); err == nil {
		t.Fatal("unnormalized Revisions key-ring path was admitted")
	}
	deployment.Revisions.ConflictTokenKeyRingManifestPath = "/etc/cartulary/revisions-conflict-token-key-ring.json"
	loaded, err := Admit(deployment)
	if err != nil {
		t.Fatalf("admit Revisions configuration: %v", err)
	}
	configuration, err := loaded.Revisions()
	if err != nil || configuration.ConflictTokenKeyRingManifestPath != "/etc/cartulary/revisions-conflict-token-key-ring.json" {
		t.Fatalf("Revisions configuration projection = %#v / %v", configuration, err)
	}
	configuration.ConflictTokenKeyRingManifestPath = "/mutated"
	again, err := loaded.Revisions()
	if err != nil || again.ConflictTokenKeyRingManifestPath != "/etc/cartulary/revisions-conflict-token-key-ring.json" {
		t.Fatalf("Revisions configuration snapshot was mutable: %#v / %v", again, err)
	}
}

func TestNetworkFlowConfigurationProjection_Unit(t *testing.T) {
	t.Run("claimed configuration is owner-validated before preflight", func(t *testing.T) {
		deployment := validProjectionDeployment(t)
		deployment.NetworkFlowActivity.Claimed = true
		deployment.NetworkFlowActivity.KeyRingManifestPath = ""
		_, err := Admit(deployment)
		diagnostics, ok := config.DiagnosticsFromError(err)
		if !ok || len(diagnostics) != 2 ||
			diagnostics[0].ReasonCode != "network_flow_cursor_key_missing" ||
			diagnostics[1].ReasonCode != "network_flow_safe_digest_key_missing" {
			t.Fatalf("claimed configuration diagnostics = %#v / %v", diagnostics, err)
		}
	})

	t.Run("normalized owner value is immutable and mapped into the temporary projection", func(t *testing.T) {
		deployment := validProjectionDeployment(t)
		deployment.NetworkFlowActivity.Claimed = true
		deployment.NetworkFlowActivity.KeyRingManifestPath = "/etc//cartulary/network-flow-key-rings.json"
		loaded, err := Admit(deployment)
		if err != nil {
			t.Fatalf("Admit: %v", err)
		}
		configuration, err := loaded.NetworkFlow()
		if err != nil {
			t.Fatalf("NetworkFlow: %v", err)
		}
		const expected = "/etc/cartulary/network-flow-key-rings.json"
		if !configuration.Claimed || configuration.KeyRingManifestPath != expected {
			t.Fatalf("owner configuration = %#v", configuration)
		}
		configuration.KeyRingManifestPath = "/mutated"
		again, err := loaded.NetworkFlow()
		if err != nil || again.KeyRingManifestPath != expected {
			t.Fatalf("snapshot owner value was mutable: %#v / %v", again, err)
		}
		if projection := loaded.Deployment(); projection.NetworkFlowActivity.KeyRingManifestPath != expected {
			t.Fatalf("application projection did not use owner value: %#v", projection.NetworkFlowActivity)
		}
	})
}

func validProjectionDeployment(t testing.TB) Deployment {
	t.Helper()
	policy, err := extensionassembly.GeneratedInactiveConfigurationPolicy()
	if err != nil {
		t.Fatalf("build inactive configuration policy: %v", err)
	}
	loaded, err := Load(config.LoadOptions{
		Path:           fixtures.Path("config", "valid.toml"),
		InactivePolicy: policy,
	})
	if err != nil {
		t.Fatalf("load valid deployment fixture: %v", err)
	}
	return loaded.Deployment()
}
