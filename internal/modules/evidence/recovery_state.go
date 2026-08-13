package evidence

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables(
		"evidence",
		"evidence_custody_events",
		"object_blobs",
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"evidence.invalidate_access_handles.v1",
			"evidence_access_handles",
		)...,
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"evidence.invalidate_upload_leases.v1",
			"evidence_object_upload_leases",
		)...,
	)
	tables = append(tables,
		recoverystate.TransientStateTables(
			"evidence.invalidate_blob_cleanup_claims.v1",
			"evidence_blob_cleanup_claims",
		)...,
	)
	return recoverystate.NewContribution(
		"module.evidence",
		tables,
		recoverystate.AuthoritativeObjectFamily(
			"evidence.blobs",
			"evidence.snapshot_blob_inventory.v1",
			"evidence.validate_blob_inventory.v1",
			"evidence.restore_blob_inventory.v1",
		),
	)
}
