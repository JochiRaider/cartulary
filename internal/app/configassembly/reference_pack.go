package configassembly

import (
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

var referencePackConfigurationKey = mustConfigurationKey[reference_data.Configuration]("module.reference_data.configuration")

func registerReferencePackConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[reference_data.Configuration]{
		Key: referencePackConfigurationKey, Namespace: "reference_packs",
		Paths: []string{"reference_packs.clock_trusted", "reference_packs.trust_bootstrap_path"},
		Decode: func(decoder config.NamespaceDecoder) (reference_data.Configuration, []config.Diagnostic) {
			var c reference_data.Configuration
			if err := decoder.Decode(&c); err != nil {
				return c, []config.Diagnostic{{Path: "reference_packs", ReasonCode: "type_mismatch", Message: "invalid Reference Pack configuration shape"}}
			}
			return c, nil
		},
		ApplyOverlay: func(c reference_data.Configuration, segments []string, raw string) (reference_data.Configuration, *config.Diagnostic) {
			value, err := reference_data.ApplyConfigurationOverlay(c, segments, raw)
			if err != nil {
				return c, &config.Diagnostic{Path: strings.Join(segments, "."), ReasonCode: "type_mismatch", Message: err.Error()}
			}
			return value, nil
		},
		Project: func(c reference_data.Configuration, presence config.NamespacePresence) (reference_data.Configuration, []config.Diagnostic) {
			c.ClockTrustedExplicit = presence.Defined("reference_packs.clock_trusted")
			if err := reference_data.ValidateReferencePackConfiguration(c); err != nil {
				return c, []config.Diagnostic{{Path: "reference_packs.trust_bootstrap_path", ReasonCode: "invalid_value", Message: err.Error()}}
			}
			return c, nil
		},
		Clone: func(c reference_data.Configuration) reference_data.Configuration { return c },
	})
}

func ReferencePackAdmissionConfiguration(c reference_data.Configuration) []AdmissionConfigurationValue {
	source := "default"
	if c.ClockTrustedExplicit {
		source = "explicit"
	}
	return []AdmissionConfigurationValue{
		{Key: "reference_packs.clock_trusted", Source: source, Value: c.ClockTrusted},
		{Key: "reference_packs.trust_bootstrap_path", Source: "explicit", Value: map[string]any{"kind": "regular_file_ref", "key": "reference_packs.trust_bootstrap_path"}},
	}
}
