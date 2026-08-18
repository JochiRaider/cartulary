package incidentbundles

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.incidentbundles",
		recoverystate.AuthoritativeTables(
			"incident_bundle_exports",
			"incident_bundle_imported_actors",
			"incident_bundle_imported_attributions",
			"incident_bundle_job_payloads",
			"incident_bundle_manifest_files",
		),
		recoverystate.AuthoritativeObjectFamily(
			"incident_bundles.files",
			"incidentbundles.snapshot_file_inventory.v1",
			"incidentbundles.validate_file_inventory.v1",
			"incidentbundles.restore_file_inventory.v1",
		),
	)
}
