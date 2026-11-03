package recovery

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.recovery",
		append(recoverystate.AuthoritativeTables("operator_recovery_journal"),
			recoverystate.RecoveryMetadataTables("recovery_target_state")...),
	)
}

func DeploymentAdminRecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables(
		"deployment_admin_audit_events",
		"deployment_bootstrap_state",
	)
	tables = append(tables, recoverystate.RecoveryMetadataTables(
		"backup_sets",
		"restore_verification_runs",
	)...)
	return recoverystate.NewContribution("module.deployment_admin", tables)
}
