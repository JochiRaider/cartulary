package imports

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.imports",
		recoverystate.AuthoritativeTables(
			"import_apply_journal",
			"import_apply_unit_plans",
			"import_sessions",
			"import_source_streams",
			"import_unit_apply_outcomes",
			"import_units",
		),
		recoverystate.AuthoritativeObjectFamily(
			"imports.source_streams",
			"imports.snapshot_source_stream_inventory.v1",
			"imports.validate_source_stream_inventory.v1",
			"imports.restore_source_stream_inventory.v1",
		),
	)
}
