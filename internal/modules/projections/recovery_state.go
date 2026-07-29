package projections

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.projections",
		recoverystate.RebuildableTables(
			"workbook.restore_projections.v1",
			"artifact_grid_projection",
			"assessment_grid_projection",
			"decision_grid_projection",
			"evidence_grid_projection",
			"host_grid_projection",
			"identity_grid_projection",
			"indicator_grid_projection",
			"party_grid_projection",
			"task_request_grid_projection",
			"timeline_grid_projection",
		),
	)
}
