package jobs

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.platform_jobs",
		recoverystate.AuthoritativeTables("jobs"),
	)
}
