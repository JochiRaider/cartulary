package indicators

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables(
		"indicator_observations",
		"indicator_state_intervals",
		"indicators",
	)
	tables = append(tables, recoverystate.RebuildableTables(
		"indicators.restore_active_identities.v1",
		"indicator_active_identities",
	)...)
	return recoverystate.NewContribution("module.indicators", tables)
}
