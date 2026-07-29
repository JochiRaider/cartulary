package savedviews

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.savedviews",
		recoverystate.AuthoritativeTables("saved_views"),
	)
}
