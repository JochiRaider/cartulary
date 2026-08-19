package workbook

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.workbook", recoverystate.AuthoritativeTables(
		"incident_workbook_preferences",
		"user_workbook_preferences",
	))
}
