package incidents

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const storeBeforeCommitOverrideKey = "incidents.store_hooks.before_commit"

type storeHooks struct {
	beforeCommit func(routeKey string, incidentID uuid.UUID) error
}

func storeHooksFromDependencies(deps httpapi.DependencySet) (storeHooks, error) {
	if deps.ModuleOverrides == nil {
		return storeHooks{}, nil
	}
	override, ok := deps.ModuleOverrides[storeBeforeCommitOverrideKey]
	if !ok {
		return storeHooks{}, nil
	}
	if _, err := httpapi.NewTestRouteGuard(deps.Env); err != nil {
		return storeHooks{}, fmt.Errorf("incidents store hook override requires test runtime: %w", err)
	}
	hook, ok := override.(func(string, uuid.UUID) error)
	if !ok {
		return storeHooks{}, fmt.Errorf("incidents store hook override has type %T", override)
	}
	return storeHooks{beforeCommit: hook}, nil
}
