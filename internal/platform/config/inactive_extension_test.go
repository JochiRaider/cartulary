package config

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
)

func TestInactiveExtensionConfiguration_Unit(t *testing.T) {
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
