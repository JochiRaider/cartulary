package tasksdecisions

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func NewRecoveryContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.tasksdecisions",
		recoverystate.AuthoritativeTables("decisions", "task_requests"),
	)
}
