package configassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/config"
)

var reportingConfigurationKey = mustConfigurationKey[reporting.Configuration]("module.reporting.configuration")

func registerReportingConfigurationContribution(builder *config.CatalogBuilder) error {
	return config.Register(builder, config.Definition[reporting.Configuration]{
		Key: reportingConfigurationKey, Namespace: "snapshot_reporting", Paths: []string{"snapshot_reporting.claimed", "snapshot_reporting.runtime_receipt_path"},
		Decode: func(decoder config.NamespaceDecoder) (reporting.Configuration, []config.Diagnostic) {
			var value reporting.Configuration
			if err := decoder.Decode(&value); err != nil {
				return value, []config.Diagnostic{{Path: "snapshot_reporting", ReasonCode: "type_mismatch", Message: "invalid Reporting configuration shape"}}
			}
			return value, nil
		},
		ApplyOverlay: func(value reporting.Configuration, segments []string, raw string) (reporting.Configuration, *config.Diagnostic) {
			updated, err := reporting.ApplyConfigurationOverlay(value, segments, raw)
			if err != nil {
				return value, &config.Diagnostic{Path: "snapshot_reporting.runtime_receipt_path", ReasonCode: "invalid_value", Message: err.Error()}
			}
			return updated, nil
		},
		Project: func(value reporting.Configuration, presence config.NamespacePresence) (reporting.Configuration, []config.Diagnostic) {
			if err := reporting.ValidateConfiguration(value); err != nil || (presence.Defined("snapshot_reporting.runtime_receipt_path") && value.RuntimeReceiptPath == "") {
				return value, []config.Diagnostic{{Path: "snapshot_reporting.runtime_receipt_path", ReasonCode: "invalid_value", Message: "invalid Reporting runtime receipt path"}}
			}
			return value, nil
		},
		Clone: func(value reporting.Configuration) reporting.Configuration { return value },
	})
}
