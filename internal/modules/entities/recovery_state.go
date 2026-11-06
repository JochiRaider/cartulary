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
	tables = append(tables, recoverystate.TransientStateTables("entities.invalidate_bundle_validation.v1", "entities_bundle_validation_associations", "entities_bundle_validation_binding", "entities_bundle_validation_claims", "entities_bundle_validation_mentions", "entities_bundle_validation_nodes")...)
	return recoverystate.NewContribution("module.entities", tables)
}
