package parties

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables("parties")
	tables = append(tables, recoverystate.RebuildableTables(
		"parties.restore_active_key_claims.v1",
		"party_active_key_claims",
	)...)
	return recoverystate.NewContribution("module.parties", tables)
}
