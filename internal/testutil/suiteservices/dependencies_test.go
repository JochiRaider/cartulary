package suiteservices

import (
	"errors"
	"testing"
)

func TestCheckServiceDependenciesAllowsExplicitDirectInvocation(t *testing.T) {
	if err := CheckServiceDependencies(map[string]string{}, "object_store", "postgres"); err != nil {
		t.Fatalf("direct invocation declaration: %v", err)
	}
}

func TestCheckServiceDependenciesRejectsOmittedService(t *testing.T) {
	err := CheckServiceDependencies(
		map[string]string{HarnessServiceDependenciesEnv: "postgres"},
		"object_store",
		"postgres",
	)
	var dependencyErr *ServiceDependencyError
	if !errors.As(err, &dependencyErr) || dependencyErr.Service != "object_store" || dependencyErr.Reason != "omitted" {
		t.Fatalf("unexpected dependency error: %#v", err)
	}
}

func TestCheckServiceDependenciesRejectsMalformedAssignment(t *testing.T) {
	for name, assigned := range map[string]string{
		"duplicate": "postgres,postgres",
		"unknown":   "message_bus",
		"unsorted":  "postgres,object_store",
	} {
		t.Run(name, func(t *testing.T) {
			if err := CheckServiceDependencies(
				map[string]string{HarnessServiceDependenciesEnv: assigned},
				"postgres",
			); err == nil {
				t.Fatal("expected malformed assignment to fail")
			}
		})
	}
}
