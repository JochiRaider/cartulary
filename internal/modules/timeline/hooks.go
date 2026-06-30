package timeline

import (
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const facadeFactoryOverrideKey = "timeline.facade_factory"

type storeHooks struct {
	beforeCommit func(routeKey string, recordID uuid.UUID) error
}

type testFacadeOptions struct {
	hooks storeHooks
}

type TestFacadeOption func(*testFacadeOptions)

func WithBeforeCommitHookForTesting(hook func(routeKey string, recordID uuid.UUID) error) TestFacadeOption {
	return func(options *testFacadeOptions) {
		options.hooks.beforeCommit = hook
	}
}

func NewFacadeForTesting(pool postgres.DB, options ...TestFacadeOption) *Facade {
	var resolved testFacadeOptions
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}
	return newFacadeWithStore(newStoreWithHooks(pool, resolved.hooks))
}

func DependencySetForTesting(options ...TestFacadeOption) httpapi.DependencySet {
	return httpapi.DependencySet{
		ModuleOverrides: map[string]any{
			facadeFactoryOverrideKey: func(pool postgres.DB) *Facade {
				return NewFacadeForTesting(pool, options...)
			},
		},
	}
}

func FacadeFromDependencies(deps httpapi.DependencySet) *Facade {
	if factory, ok := deps.ModuleOverrides[facadeFactoryOverrideKey].(func(postgres.DB) *Facade); ok && factory != nil {
		return factory(deps.PostgresHandle())
	}
	return NewFacade(deps.PostgresHandle())
}
