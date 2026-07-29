package assessments

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.assessments",
		recoverystate.AuthoritativeTables("assessments"),
	)
}
