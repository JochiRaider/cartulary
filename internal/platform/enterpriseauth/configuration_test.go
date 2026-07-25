package enterpriseauth

import "testing"

func TestConfigurationNormalizationAndValidation_Unit(t *testing.T) {
	t.Run("unclaimed configuration discards owner value without preflight", func(t *testing.T) {
		normalized, findings := NormalizeAndValidateConfiguration(Configuration{
			ProviderManifestPath: "relative/must-not-be-read.json",
		})
		if normalized.ProviderManifestPath != "" || len(findings) != 0 {
			t.Fatalf("unclaimed configuration = %#v, findings %#v", normalized, findings)
		}
	})

	t.Run("claimed configuration requires a manifest path", func(t *testing.T) {
		_, findings := NormalizeAndValidateConfiguration(Configuration{Claimed: true})
		requireConfigurationFinding(t, findings, providerManifestPathKey, "provider_manifest_path_missing")
	})

	t.Run("claimed configuration normalizes a valid path", func(t *testing.T) {
		normalized, findings := NormalizeAndValidateConfiguration(Configuration{
			Claimed:              true,
			ProviderManifestPath: "/etc//cartulary/enterprise-auth-providers.json",
		})
		if len(findings) != 0 || normalized.ProviderManifestPath != "/etc/cartulary/enterprise-auth-providers.json" {
			t.Fatalf("normalized configuration = %#v, findings %#v", normalized, findings)
		}
	})

	for name, path := range map[string]string{
		"relative":       "relative/providers.json",
		"parent":         "/etc/cartulary/../providers.json",
		"dot":            "/etc/./cartulary/providers.json",
		"shell variable": "/etc/$CONFIG/providers.json",
		"nul":            "/etc/cartulary/providers.json\x00",
	} {
		t.Run("rejects "+name, func(t *testing.T) {
			_, findings := NormalizeAndValidateConfiguration(Configuration{
				Claimed:              true,
				ProviderManifestPath: path,
			})
			if len(findings) != 1 ||
				findings[0].Path != providerManifestPathKey ||
				(findings[0].ReasonCode != "path_not_absolute" && findings[0].ReasonCode != "path_forbidden_segment") {
				t.Fatalf("findings = %#v", findings)
			}
		})
	}
}

func requireConfigurationFinding(t testing.TB, findings []ConfigurationFinding, path string, reason string) {
	t.Helper()
	for _, finding := range findings {
		if finding.Path == path && finding.ReasonCode == reason {
			return
		}
	}
	t.Fatalf("missing finding path=%q reason=%q in %#v", path, reason, findings)
}
