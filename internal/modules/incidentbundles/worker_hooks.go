package incidentbundles

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

const workerStartHookOverrideKey = "incidentbundles.worker_start_hook"

type testOptions struct {
	workerStartHook func(string)
}

type TestOption func(*testOptions)

func WithWorkerStartHookForTesting(hook func(string)) TestOption {
	return func(options *testOptions) {
		options.workerStartHook = hook
	}
}

func DependencySetForTesting(options ...TestOption) httpapi.DependencySet {
	var resolved testOptions
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}
	overrides := map[string]any{}
	if resolved.workerStartHook != nil {
		overrides[workerStartHookOverrideKey] = resolved.workerStartHook
	}
	return httpapi.DependencySet{ModuleOverrides: overrides}
}

func workerStartHookFromDependencies(deps httpapi.DependencySet) (func(string), error) {
	if deps.ModuleOverrides == nil {
		return nil, nil
	}
	override, ok := deps.ModuleOverrides[workerStartHookOverrideKey]
	if !ok {
		return nil, nil
	}
	if _, err := httpapi.NewTestRouteGuard(deps.Env); err != nil {
		return nil, fmt.Errorf("incident bundle worker hook override requires test runtime: %w", err)
	}
	hook, ok := override.(func(string))
	if !ok {
		return nil, fmt.Errorf("incident bundle worker hook override has type %T", override)
	}
	return hook, nil
}
