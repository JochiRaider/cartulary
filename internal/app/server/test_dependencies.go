package server

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const postgresDBDecoratorOverrideKey = "app.server.postgres_db_decorator"

type postgresDBDecorator func(postgres.DB) postgres.DB

func DependencySetWithPostgresDBDecoratorForTesting(decorator func(postgres.DB) postgres.DB) httpapi.DependencySet {
	return httpapi.DependencySet{
		ModuleOverrides: map[string]any{
			postgresDBDecoratorOverrideKey: postgresDBDecorator(decorator),
		},
	}
}

func decoratePostgresForTestRuntime(env map[string]string, overrides map[string]any, base postgres.DB) (postgres.DB, error) {
	override, ok := overrides[postgresDBDecoratorOverrideKey]
	if !ok {
		return base, nil
	}
	if _, err := httpapi.NewTestRouteGuard(env); err != nil {
		return nil, fmt.Errorf("application Postgres decorator requires test runtime: %w", err)
	}
	decorator, ok := override.(postgresDBDecorator)
	if !ok || decorator == nil {
		return nil, fmt.Errorf("application Postgres decorator has type %T", override)
	}
	decorated := decorator(base)
	if decorated == nil {
		return nil, fmt.Errorf("application Postgres decorator returned nil")
	}
	return decorated, nil
}
