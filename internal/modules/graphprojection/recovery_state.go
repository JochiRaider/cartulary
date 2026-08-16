package graphprojection

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.RebuildableTables(
		"graphprojection.restore_rebuild.v2",
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	)
	return recoverystate.NewContribution(
		"module.graphprojection",
		tables,
	)
}
