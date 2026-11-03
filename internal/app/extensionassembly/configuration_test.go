package extensionassembly

import (
	"bytes"
	"github.com/BurntSushi/toml"
	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
)

func TestConfigurationPolicySupportsFutureProfileWithoutKernelChange_Unit(t *testing.T) {
	t.Run("recognized unclaimable retains policy", func(t *testing.T) {
		policy, err := newConfigurationPolicy([]extensions.Descriptor{{ProfileID: "retired_profile", Claimable: false, ClaimConfigKey: "retired_profile.claimed"}}, []extensions.InactiveConfigurationPolicy{{ProfileID: "retired_profile", Namespace: "retained_settings", ClaimKey: "retired_profile.claimed", Key: "retained_settings.path", Kind: extensions.PolicyForbidden}})
		if err != nil {
			t.Fatal("recognized unclaimable policy prevented Base configuration", err)
		}
		checkRetainedUnclaimableLoading(t, policy)
		if len(policy.ClaimRegistrations()) != 1 || len(policy.Keys()) != 1 {
			t.Fatal("retained profile identity/policy lost")
		}
		requested, err := policy.MaterializeRequestedClaims([]string{"retired_profile"})
		if err != nil || !reflect.DeepEqual(requested.ProfileIDs(), []string{"retired_profile"}) {
			t.Fatal("true request did not reach claim admission", err)
		}
	})
	policy, err := newConfigurationPolicy([]extensions.Descriptor{
		{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"},
		{ProfileID: "existing_profile", Claimable: true, ClaimConfigKey: "existing_profile.claimed"},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	registrations := policy.ClaimRegistrations()
	if got, want := []string{registrations[0].Path, registrations[1].Path}, []string{"existing_profile.claimed", "future_profile.claimed"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
	claims, err := policy.MaterializeRequestedClaims([]string{"future_profile"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"future_profile"}; !reflect.DeepEqual(claims.ProfileIDs(), want) {
		t.Fatalf("claims = %#v, want %#v", claims.ProfileIDs(), want)
	}
	copy := claims.ProfileIDs()
	copy[0] = "mutated"
	if got := claims.ProfileIDs(); !reflect.DeepEqual(got, []string{"future_profile"}) {
		t.Fatalf("requested claims were mutable: %#v", got)
	}
}

func TestGeneratedConfigurationPolicyHasExactClaimRegistrationParity_Unit(t *testing.T) {
	policy, err := GeneratedConfigurationPolicy()
	if err != nil {
		t.Fatal(err)
	}
	registrations := policy.ClaimRegistrations()
	got := make([]string, len(registrations))
	for index, registration := range registrations {
		got[index] = registration.ID + "=" + registration.Path
	}
	want := []string{
		"enterprise_authentication=enterprise_authentication.claimed",
		"import=import.claimed",
		"incident_portability=incident_portability.claimed",
		"network_flow_activity=network_flow_activity.claimed",
		"reference_pack=reference_pack.claimed",
		"snapshot_reporting=snapshot_reporting.claimed",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("generated claim registrations = %#v, want %#v", got, want)
	}
}

func TestConfigurationPolicyRejectsMalformedCatalog_Unit(t *testing.T) {
	if err := (ConfigurationPolicy{}).ValidateAdmission(); err == nil {
		t.Fatal("zero policy admitted")
	}
	if _, err := newConfigurationPolicy([]extensions.Descriptor{{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"}}, []extensions.InactiveConfigurationPolicy{{ProfileID: "future_profile", Namespace: "future_settings", ClaimKey: "future_profile.claimed", Key: "future_settings.setting", Kind: extensions.PolicyForbidden}}); err != nil {
		t.Fatal("declared configuration namespace rejected", err)
	}
	tests := map[string]struct {
		descriptors []extensions.Descriptor
		inactive    []extensions.InactiveConfigurationPolicy
	}{
		"missing identity": {
			descriptors: []extensions.Descriptor{{Claimable: true}},
		},
		"noncanonical path": {
			descriptors: []extensions.Descriptor{{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "other.claimed"}},
		},
		"duplicate profile": {
			descriptors: []extensions.Descriptor{
				{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"},
				{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"},
			},
		},
		"stale inactive claim": {
			descriptors: []extensions.Descriptor{{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"}},
			inactive: []extensions.InactiveConfigurationPolicy{{
				ProfileID: "future_profile", ClaimKey: "stale.claimed", Key: "future_profile.setting", Kind: extensions.PolicyForbidden,
			}},
		},
		"foreign inactive path": {
			descriptors: []extensions.Descriptor{{ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed"}},
			inactive: []extensions.InactiveConfigurationPolicy{{
				ProfileID: "future_profile", Namespace: "future_profile", ClaimKey: "future_profile.claimed", Key: "other.setting", Kind: extensions.PolicyForbidden,
			}},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := newConfigurationPolicy(test.descriptors, test.inactive); err == nil {
				t.Fatal("expected catalog rejection")
			}
		})
	}
}

func TestRequestedClaimsRejectUnknownOrDuplicateRegistration_Unit(t *testing.T) {
	policy, err := newConfigurationPolicy([]extensions.Descriptor{{
		ProfileID: "future_profile", Claimable: true, ClaimConfigKey: "future_profile.claimed",
	}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	for name, ids := range map[string][]string{
		"unknown":   {"other"},
		"duplicate": {"future_profile", "future_profile"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := policy.MaterializeRequestedClaims(ids); err == nil {
				t.Fatal("expected request rejection")
			}
		})
	}
}

func TestRequestedClaimsRemainDistinctAfterDependencyFailure_Unit(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	policy, err := ConfigurationPolicyFromCoordinator(coordinator)
	if err != nil {
		t.Fatal(err)
	}
	requested, err := policy.MaterializeRequestedClaims([]string{"network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims(requested.ProfileIDs())
	if err == nil {
		t.Fatal("dependency-incomplete request was resolved")
	}
	if got := requested.ProfileIDs(); !reflect.DeepEqual(got, []string{"network_flow_activity"}) {
		t.Fatalf("request was mutated by dependency resolution: %#v", got)
	}
	if got := resolution.AdmissionOrder(); len(got) != 0 {
		t.Fatalf("partial admission order escaped: %#v", got)
	}
	if got := resolution.Claims().ProfileIDs(); len(got) != 0 {
		t.Fatalf("partial resolved claim identity escaped: %#v", got)
	}
}

func checkRetainedUnclaimableLoading(t *testing.T, policy ConfigurationPolicy) {
	var document map[string]any
	if _, err := toml.Decode(string(fixtures.MustRead("config", "valid.toml")), &document); err != nil {
		t.Fatal(err)
	}
	core := map[string]any{}
	for _, name := range []string{"config_schema_id", "deployment_profile", "application", "roots", "bootstrap", "timeouts", "intervals", "limits"} {
		if value, present := document[name]; present {
			core[name] = value
		}
	}
	var encoded bytes.Buffer
	if err := toml.NewEncoder(&encoded).Encode(core); err != nil {
		t.Fatal(err)
	}
	catalog, err := (&config.CatalogBuilder{}).Build()
	if err != nil {
		t.Fatal(err)
	}
	for _, claim := range []string{"omitted", "false", "true"} {
		t.Run("loader "+claim, func(t *testing.T) {
			content := encoded.String()
			if claim != "omitted" {
				content += "\n[retired_profile]\nclaimed = " + claim + "\n"
			}
			path := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(path, []byte(content), 0600); err != nil {
				t.Fatal(err)
			}
			snapshot, err := config.LoadSnapshotWithOptions(config.LoadOptions{Path: path, ExtensionPolicy: policy}, catalog)
			if err != nil {
				t.Fatal("inactive loading required an owner implementation", err)
			}
			requested, err := policy.MaterializeRequestedClaims(config.RequestedClaimRegistrationIDs(snapshot))
			if err != nil {
				t.Fatal(err)
			}
			if (len(requested.ProfileIDs()) == 1) != (claim == "true") {
				t.Fatal("claim request changed before owned admission")
			}
			_, err = config.LoadSnapshotWithOptions(config.LoadOptions{Path: path, ExtensionPolicy: policy, Env: map[string]string{"CARTULARY__RETAINED_SETTINGS__PATH": "vault://never-resolve"}}, catalog)
			diagnostics, ok := config.DiagnosticsFromError(err)
			if !ok || len(diagnostics) != 1 || diagnostics[0].ReasonCode != "extension_config_without_claim" || diagnostics[0].Details["profile_id"] != "retired_profile" {
				t.Fatalf("retained forbidden policy lost: %#v / %v", diagnostics, err)
			}
		})
	}
}
