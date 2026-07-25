package networkflow

import (
	"strings"
	"testing"
)

func TestNetworkFlowConfigurationContribution_Unit(t *testing.T) {
	t.Run("unclaimed is inert and clears owner-local path material", func(t *testing.T) {
		configuration, findings := NormalizeAndValidateConfiguration(Configuration{
			KeyRingManifestPath: "/must/not/be/read",
		})
		if configuration.Claimed || configuration.KeyRingManifestPath != "" || len(findings) != 0 {
			t.Fatalf("unclaimed configuration = %#v, %#v", configuration, findings)
		}
	})

	t.Run("claimed requires both key-ring purposes", func(t *testing.T) {
		_, findings := NormalizeAndValidateConfiguration(Configuration{Claimed: true})
		if len(findings) != 2 ||
			findings[0].ReasonCode != "network_flow_cursor_key_missing" ||
			findings[1].ReasonCode != "network_flow_safe_digest_key_missing" {
			t.Fatalf("missing manifest findings = %#v", findings)
		}
	})

	t.Run("claimed path validation preserves diagnostics and normalization", func(t *testing.T) {
		for name, path := range map[string]string{
			"relative": "relative/manifest.json",
			"NUL":      "/etc/cartulary/\x00manifest.json",
			"variable": "/etc/$CARTULARY/manifest.json",
			"parent":   "/etc/cartulary/../manifest.json",
		} {
			t.Run(name, func(t *testing.T) {
				_, findings := NormalizeAndValidateConfiguration(Configuration{
					Claimed:             true,
					KeyRingManifestPath: path,
				})
				if len(findings) != 1 || findings[0].Path != keyRingConfigPath {
					t.Fatalf("path findings = %#v", findings)
				}
			})
		}
		configuration, findings := NormalizeAndValidateConfiguration(Configuration{
			Claimed:             true,
			KeyRingManifestPath: "/etc//cartulary/network-flow-key-rings.json",
		})
		if len(findings) != 0 ||
			configuration.KeyRingManifestPath != "/etc/cartulary/network-flow-key-rings.json" {
			t.Fatalf("normalized configuration = %#v, %#v", configuration, findings)
		}
	})

	t.Run("manifest read failures remain owner-defined and redacted", func(t *testing.T) {
		for failure, reason := range map[ManifestReadFailure]string{
			ManifestUnreadable: "network_flow_cursor_key_missing",
			ManifestUnsafe:     "network_flow_cursor_key_invalid",
			ManifestTooLarge:   "network_flow_cursor_key_invalid",
		} {
			err := KeyRingManifestReadError(failure)
			if err == nil || !strings.Contains(err.Error(), reason) {
				t.Fatalf("failure %q error = %v; want %q", failure, err, reason)
			}
		}
	})
}
