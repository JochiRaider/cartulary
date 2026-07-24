package config

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config/extensioninactive"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestInactiveExtensionConfiguration_Unit(t *testing.T) {
	schema := map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"mode": map[string]any{"type": "string", "enum": []any{"strict", "relaxed"}},
		},
		"required": []any{"mode"},
	}
	catalog, err := extensioninactive.NewCatalog([]extensioninactive.Policy{
		{
			ProfileID: "future_profile",
			ClaimKey:  "enterprise_authentication.claimed",
			Key:       "future_profile.syntax",
			Kind:      extensioninactive.PolicySyntaxOnly,
			Schema:    schema,
		},
		{
			ProfileID: "future_profile",
			ClaimKey:  "enterprise_authentication.claimed",
			Key:       "future_profile.forbidden",
			Kind:      extensioninactive.PolicyForbidden,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("validates and discards an authored syntax-only table", func(t *testing.T) {
		content := string(fixtures.MustRead("config", "valid.toml")) + "\n[future_profile.syntax]\nmode = \"strict\"\n"
		cfg, loadErr := LoadWithOptions(LoadOptions{
			Path:                     writeTempConfig(t, content),
			ExtensionInactiveCatalog: catalog,
		})
		if loadErr != nil {
			t.Fatalf("load inert syntax-only configuration: %v", loadErr)
		}
		if cfg.EnterpriseAuthentication.Claimed {
			t.Fatal("inactive syntax-only configuration changed claim state")
		}
	})

	t.Run("validates and discards a syntax-only overlay", func(t *testing.T) {
		_, loadErr := LoadWithOptions(LoadOptions{
			Path: writeTempConfig(t, string(fixtures.MustRead("config", "valid.toml"))),
			Env: map[string]string{
				"CARTULARY__FUTURE_PROFILE__SYNTAX": `{"mode":"relaxed"}`,
			},
			ExtensionInactiveCatalog: catalog,
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
			_, loadErr := LoadWithOptions(LoadOptions{
				Path:                     writeTempConfig(t, content),
				ExtensionInactiveCatalog: catalog,
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
