package records

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.records",
		recoverystate.AuthoritativeTables("records"),
	)
}
