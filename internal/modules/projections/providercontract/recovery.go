package providercontract

import (
	"slices"

	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

const RecoveryRebuildAlgorithmID = "workbook.restore_projections.v1"

var recoveryProjectionTableIDs = []string{
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
}

// RecoveryProjectionTableIDs returns the closed physical derived-state set
// that restore excludes and Projections rebuilds. Callers receive a copy.
func RecoveryProjectionTableIDs() []string {
	return slices.Clone(recoveryProjectionTableIDs)
}

// RecoveryStateContribution projects Projections-owned table facts into the
// Recovery-owned state vocabulary without exposing a runtime catalog or store.
func RecoveryStateContribution() recoverystate.Contribution {
	return recoverystate.NewContribution(
		"module.projections",
		recoverystate.RebuildableTables(
			RecoveryRebuildAlgorithmID,
			RecoveryProjectionTableIDs()...,
		),
	)
}
