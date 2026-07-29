package recoverycontribution

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.extensions",
		recoverystate.AuthoritativeTables(
			"extension_job_cancellation_observations",
			"extension_job_commit_proofs",
			"extension_migration_ledger",
			"extension_staged_object_references",
			"extension_staged_objects",
			"extension_state_metadata",
		),
		recoverystate.AuthoritativeObjectFamily(
			"extensions.staged_objects",
			"extensions.snapshot_staged_object_inventory.v1",
			"extensions.validate_staged_object_inventory.v1",
			"extensions.restore_staged_object_inventory.v1",
		),
	)
}
