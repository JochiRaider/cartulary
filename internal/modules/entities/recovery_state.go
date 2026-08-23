package entities

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables(
		"entity_aliases",
		"entity_mentions",
		"entity_preserved_identifiers",
		"hosts",
		"identities",
	)
	tables = append(tables, recoverystate.RebuildableTables(
		"entities.restore_active_identifier_claims.v1",
		"entity_active_identifier_claims",
	)...)
	return recoverystate.NewContribution("module.entities", tables)
}
