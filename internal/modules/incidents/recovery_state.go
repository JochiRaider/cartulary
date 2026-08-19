package incidents

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.incidents", recoverystate.AuthoritativeTables(
		"incident_memberships",
		"incidents",
	))
}
