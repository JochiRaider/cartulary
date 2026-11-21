package incidentbundle

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
)

// validateAttributions keeps the capability's rank; the attribution itself is
// flushed by the core in step 9 and checked there by counts.
func (p preparedTasksDecisionsImport) validateAttributions(_ context.Context, input sourceport.ImportContext) error {
	if input.Attributions == nil {
		return tasksDecisionsInvariantFailure(policy.InvariantAttribution)
	}
	return nil
}
