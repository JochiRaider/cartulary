package reporting

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.reporting",
		recoverystate.AuthoritativeTables(
			"reporting_composition_preview_output_files",
			"reporting_composition_preview_outputs",
			"reporting_job_payloads",
			"reporting_release_approvals",
			"reporting_releases",
			"reporting_render_bundle_files",
			"reporting_render_bundles",
			"reporting_snapshots",
		),
		recoverystate.AuthoritativeObjectFamily(
			"reporting.render_preview_members",
			"reporting.snapshot_output_inventory.v1",
			"reporting.validate_output_inventory.v1",
			"reporting.restore_output_inventory.v1",
		),
	)
}
