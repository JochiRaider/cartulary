package phase2test

import (
	"errors"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const incidentsBeforeCommitHookOverrideKey = "incidents.store_hooks.before_commit"

func IncidentCreateRollbackFaultDependencies() httpapi.DependencySet {
	return httpapi.DependencySet{
		ModuleOverrides: map[string]any{
			incidentsBeforeCommitHookOverrideKey: func(routeKey string, _ uuid.UUID) error {
				if routeKey == "incidents.create" {
					return errors.New("forced incidents rollback")
				}
				return nil
			},
		},
	}
}
