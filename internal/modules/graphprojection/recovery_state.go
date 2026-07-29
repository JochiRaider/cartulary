package graphprojection

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.graphprojection",
		recoverystate.RebuildableTables(
			"graphprojection.restore_rebuild.v1",
			"graph_projection_edges",
			"graph_projection_idempotency",
			"graph_projection_runs",
			"graph_projection_vertices",
			"graph_projection_views",
		),
	)
}
