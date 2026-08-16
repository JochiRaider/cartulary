package networkflow

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution("module.networkflow", recoverystate.AuthoritativeTables(
		"network_flow_graph_views",
		"network_flow_indicator_bindings",
		"network_flow_rejected_row_diagnostics",
		"network_flow_rows",
		"network_flow_tables",
	))
}
