package links

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.links",
		recoverystate.AuthoritativeTables("record_links", "record_tags"),
	)
}
