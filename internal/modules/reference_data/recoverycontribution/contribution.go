package recoverycontribution

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.reference_data",
		recoverystate.AuthoritativeTables(
			"reference_pack_activation_state",
			"reference_pack_attestations",
			"reference_pack_job_payloads",
			"reference_packs",
		),
		recoverystate.AuthoritativeObjectFamily(
			"reference_packs.members",
			"reference_data.snapshot_member_inventory.v1",
			"reference_data.validate_member_inventory.v1",
			"reference_data.restore_member_inventory.v1",
		),
	)
}
