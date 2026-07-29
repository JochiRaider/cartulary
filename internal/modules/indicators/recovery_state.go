package indicators

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.indicators", recoverystate.AuthoritativeTables(
		"indicator_observations",
		"indicator_state_intervals",
		"indicators",
	))
}
