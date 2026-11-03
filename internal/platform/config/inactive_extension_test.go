package config

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestInactiveExtensionConfiguration_Unit(t *testing.T) {
	t.Run("actual typed owner remains inactive", checkInactiveTypedOwner)
	policy := syntaxOnlyTestExtensionPolicy()

	t.Run("validates and discards an authored syntax-only table", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml")) + "\n[future_profile.syntax]\nmode = \"strict\"\n"
		cfg, loadErr := loadWithTestCatalog(t, LoadOptions{
			Path:            writeTempConfig(t, content),
			ExtensionPolicy: policy,
		})
		if loadErr != nil {
			t.Fatalf("load inert syntax-only configuration: %v", loadErr)
		}
		if testOwnerValue[testEnterpriseConfiguration](t, cfg, "enterprise_authentication").Claimed {
			t.Fatal("inactive syntax-only configuration changed claim state")
		}
	})

	t.Run("validates and discards a syntax-only overlay", func(t *testing.T) {
		_, loadErr := loadWithTestCatalog(t, LoadOptions{
			Path: writeTempConfig(t, string(fixtures.MustRead("config", "valid.toml"))),
			Env: map[string]string{
				"CARTULARY__FUTURE_PROFILE__SYNTAX": `{"mode":"relaxed"}`,
			},
			ExtensionPolicy: policy,
		})
		if loadErr != nil {
			t.Fatalf("load inert syntax-only overlay: %v", loadErr)
		}
	})

	for name, suffix := range map[string]string{
		"invalid enum":     "\n[future_profile.syntax]\nmode = \"active\"\n",
		"reference shaped": "\n[future_profile.syntax]\nmode = \"strict\"\nref = \"vault://do-not-resolve\"\n",
		"forbidden":        "\n[future_profile]\nforbidden = \"vault://do-not-resolve\"\n",
	} {
		t.Run(name, func(t *testing.T) {
			content := string(fixtures.MustRead("config", "valid.toml")) + suffix
			_, loadErr := loadWithTestCatalog(t, LoadOptions{
				Path:            writeTempConfig(t, content),
				ExtensionPolicy: policy,
			})
			if loadErr == nil {
				t.Fatal("invalid inactive configuration was accepted")
			}
			rendered := loadErr.Error()
			if strings.Contains(rendered, "vault://") {
				t.Fatalf("inactive value leaked into diagnostic: %s", rendered)
			}
			reason := "extension_validation_result_invalid"
			if name == "forbidden" {
				reason = "extension_config_without_claim"
			}
			requireDiagnostic(t, loadErr, "future_profile."+map[string]string{
				"invalid enum":     "syntax",
				"reference shaped": "syntax",
				"forbidden":        "forbidden",
			}[name], reason)
		})
	}
}

func checkInactiveTypedOwner(t *testing.T) {
	type settings struct {
		Syntax int `toml:"syntax"`
	}
	key, err := NewKey[settings]("future.settings")
	if err != nil {
		t.Fatal(err)
	}
	for _, scenario := range []struct {
		name, extra string
		namespace   string
		env         map[string]string
		active      bool
		unclaimable bool
		reason      string
	}{
		{name: "omitted"},
		{name: "Core owns the claim field", namespace: "future_profile", extra: "\n[future_profile]\nclaimed = true\nsyntax = 9\n", active: true},
		{name: "unclaimable true remains a request", extra: "\n[future_profile]\nclaimed = true\n", unclaimable: true},
		{name: "inactive unknown file key", extra: "\n[future_settings]\nunknown = \"vault://never-read\"\n", reason: "unknown_key"},
		{name: "inactive unknown overlay", env: map[string]string{"CARTULARY__FUTURE_SETTINGS__UNKNOWN": "vault://never-read"}, reason: "unknown_key"},
		{name: "namespace is not a table", extra: "\nfuture_settings = \"vault://never-read\"\n", reason: "type_mismatch"},
		{name: "inert type differs from active", extra: "\n[future_settings.syntax]\nmode = \"strict\"\n"},
		{name: "inactive overlay", env: map[string]string{"CARTULARY__FUTURE_SETTINGS__SYNTAX": `{"mode":"strict"}`}},
		{name: "true overridden false", extra: "\n[future_profile]\nclaimed = true\n[future_settings.syntax]\nmode = \"strict\"\n", env: map[string]string{"CARTULARY__FUTURE_PROFILE__CLAIMED": "false"}},
		{name: "active", extra: "\n[future_profile]\nclaimed = true\n[future_settings]\nsyntax = 9\n", active: true},
		{name: "false overridden true", extra: "\n[future_profile]\nclaimed = false\n[future_settings]\nsyntax = 9\n", env: map[string]string{"CARTULARY__FUTURE_PROFILE__CLAIMED": "true"}, active: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			calls := 0
			namespace := scenario.namespace
			if namespace == "" {
				namespace = "future_settings"
			}
			syntaxKey := namespace + ".syntax"
			builder := &CatalogBuilder{entries: append([]catalogEntry(nil), testOwnerNamespaceCatalog(t).entries...)}
			err := Register(builder, Definition[settings]{Key: key, Namespace: namespace, Paths: []string{syntaxKey},
				Decode: func(decoder NamespaceDecoder) (settings, []Diagnostic) {
					calls++
					var value settings
					if err := decoder.Decode(&value); err != nil {
						return value, []Diagnostic{{Path: "future_settings.syntax", ReasonCode: "type_mismatch", Message: "active owner decoder was invoked"}}
					}
					return value, nil
				},
				Project: func(value settings, _ NamespacePresence) (settings, []Diagnostic) {
					calls++
					if value.Syntax == 0 {
						value.Syntax = 42
					}
					return value, nil
				},
				Clone: func(value settings) settings { calls++; return value },
			})
			if err != nil {
				t.Fatal(err)
			}
			catalog, err := builder.Build()
			if err != nil {
				t.Fatal(err)
			}
			registrations := append(characterizedClaimRegistrations(), ClaimRegistration{ID: "future_profile", Path: "future_profile.claimed", Namespace: namespace, Claimable: true})
			sort.Slice(registrations, func(i, j int) bool { return registrations[i].ID < registrations[j].ID })
			inert := syntaxOnlyTestExtensionPolicy()
			policy := testExtensionPolicy{registrations: registrations, claims: map[string]string{syntaxKey: "future_profile.claimed"}, parse: inert.ParseOverlay, validate: func(values map[string]any) [][2]string {
				return inert.ValidateAndDiscard(map[string]any{"future_profile.syntax": values[syntaxKey]})
			}}
			content := configWithoutClaimSections(t) + scenario.extra
			if scenario.name == "namespace is not a table" {
				content = scenario.extra + configWithoutClaimSections(t)
			}
			if scenario.unclaimable {
				for i := range policy.registrations {
					if policy.registrations[i].ID == "future_profile" {
						policy.registrations[i].Claimable = false
					}
				}
			}
			snapshot, err := LoadSnapshotWithOptions(LoadOptions{Path: writeTempConfig(t, content), Env: scenario.env, ExtensionPolicy: policy}, catalog)
			if scenario.reason != "" {
				if err == nil || strings.Contains(err.Error(), "vault://") {
					t.Fatalf("inactive malformed key classification: %v", err)
				}
				findings, ok := DiagnosticsFromError(err)
				if !ok || len(findings) != 1 || findings[0].ReasonCode != scenario.reason {
					t.Fatalf("inactive findings: %#v", findings)
				}
				if calls != 0 {
					t.Fatal("invalid inactive input invoked owner callbacks")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if scenario.unclaimable && !reflect.DeepEqual(RequestedClaimRegistrationIDs(snapshot), []string{"future_profile"}) {
				t.Fatal("unclaimable true request was lost")
			}
			if !scenario.active {
				if calls != 0 {
					t.Fatalf("inactive owner callbacks executed %d times", calls)
				}
				if _, present := snapshot.values[key.id]; present {
					t.Fatal("inactive configuration view was retained")
				}
			} else {
				value, err := Value(snapshot, key)
				if err != nil || value.Syntax != 9 || calls == 0 {
					t.Fatalf("claimed value lost: %#v / %v", value, err)
				}
			}
		})
	}
}
