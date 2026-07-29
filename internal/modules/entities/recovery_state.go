package entities

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.entities", recoverystate.AuthoritativeTables(
		"entity_aliases",
		"entity_mentions",
		"entity_preserved_identifiers",
		"hosts",
		"identities",
	))
}
