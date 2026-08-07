package config

import (
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

var characterizedClaimPaths = []string{
	"enterprise_authentication.claimed",
	"import.claimed",
	"incident_portability.claimed",
	"network_flow_activity.claimed",
	"reference_pack.claimed",
	"snapshot_reporting.claimed",
}

func TestRegisteredClaimsProjectOnlyRequestedIdentities_Unit(t *testing.T) {
	base := configWithoutClaimSections(t)
	content := base + "\n[future_profile]\nclaimed = true\n"
	policy := testExtensionPolicy{registrations: []ClaimRegistration{{ID: "future_profile", Path: "future_profile.claimed"}}}
	snapshot, err := LoadSnapshotFromTOML([]byte(content), LoadOptions{ExtensionPolicy: policy}, testOwnerNamespaceCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := RequestedClaimRegistrationIDs(snapshot), []string{"future_profile"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("requested registrations = %#v, want %#v", got, want)
	}
	copy := RequestedClaimRegistrationIDs(snapshot)
	copy[0] = "mutated"
	if got := RequestedClaimRegistrationIDs(snapshot); !reflect.DeepEqual(got, []string{"future_profile"}) {
		t.Fatalf("snapshot claim projection was mutable: %#v", got)
	}
}

func TestClaimRegistrationCatalogFailsClosedBeforeConfigurationRead_Unit(t *testing.T) {
	tests := map[string][]ClaimRegistration{
		"missing id":         {{Path: "future_profile.claimed"}},
		"path mismatch":      {{ID: "future_profile", Path: "other.claimed"}},
		"duplicate":          {{ID: "future_profile", Path: "future_profile.claimed"}, {ID: "future_profile", Path: "future_profile.claimed"}},
		"noncanonical order": {{ID: "z_profile", Path: "z_profile.claimed"}, {ID: "a_profile", Path: "a_profile.claimed"}},
	}
	for name, registrations := range tests {
		t.Run(name, func(t *testing.T) {
			policy := testExtensionPolicy{registrations: registrations}
			_, err := LoadSnapshotWithOptions(LoadOptions{
				Path:            "/path/that/must/not/be/read.toml",
				ExtensionPolicy: policy,
			}, testOwnerNamespaceCatalog(t))
			if err == nil || strings.Contains(err.Error(), "read config") {
				t.Fatalf("registration catalog was not rejected before configuration read: %v", err)
			}
		})
	}
}

func TestClaimConfigurationTruthTable_Unit(t *testing.T) {
	base := configWithoutClaimSections(t)

	t.Run("omitted and explicit false are not requested", func(t *testing.T) {
		for name, env := range map[string]map[string]string{
			"omitted": nil,
			"false":   claimOverlayValues("false"),
		} {
			t.Run(name, func(t *testing.T) {
				cfg := mustLoadConfig(t, base, env)
				if got := requestedClaimRegistrationIDs(cfg); len(got) != 0 {
					t.Fatalf("claims were requested: %#v", got)
				}
			})
		}
	})

	t.Run("explicit true is retained only as registered identities", func(t *testing.T) {
		cfg := mustLoadConfig(t, base, claimOverlayValues("true"))
		if got, want := requestedClaimRegistrationIDs(cfg), []string{
			"enterprise_authentication", "import", "incident_portability",
			"network_flow_activity", "reference_pack", "snapshot_reporting",
		}; !reflect.DeepEqual(got, want) {
			t.Fatalf("requested registrations = %#v, want %#v", got, want)
		}
	})

	for _, invalid := range []string{"null", "truthy"} {
		t.Run("rejects "+invalid, func(t *testing.T) {
			for _, path := range characterizedClaimPaths {
				envName := "CARTULARY__" + strings.ToUpper(strings.ReplaceAll(path, ".", "__"))
				err := loadInvalidConfig(t, base, map[string]string{envName: invalid})
				requireDiagnostic(t, err, path, "type_mismatch")
			}
		})
	}
}

func TestUnregisteredClaimPathIsRejected_Unit(t *testing.T) {
	content := configWithoutClaimSections(t) + "\n[unregistered_profile]\nclaimed = true\n"
	err := loadInvalidConfig(t, content, nil)
	requireDiagnostic(t, err, "unregistered_profile.claimed", "unknown_key")
}

func configWithoutClaimSections(t testing.TB) string {
	t.Helper()
	base := string(fixtures.MustRead("config", "valid.toml"))
	for _, section := range []string{"[import]", "[incident_portability]", "[reference_pack]", "[snapshot_reporting]"} {
		base = stripSection(t, base, section)
	}
	return base
}

func claimOverlayValues(value string) map[string]string {
	env := make(map[string]string, len(characterizedClaimPaths))
	for _, path := range characterizedClaimPaths {
		env["CARTULARY__"+strings.ToUpper(strings.ReplaceAll(path, ".", "__"))] = value
	}
	return env
}
