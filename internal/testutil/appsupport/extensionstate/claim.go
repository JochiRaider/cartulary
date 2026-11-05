// Package extensionstate composes real claim admission for owner integration tests.
// It deliberately avoids importing a profile implementation so package-local
// profile tests can supply their own continuing production ports without cycles.
package extensionstate

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

type Reader interface {
	FamilyCounts(context.Context, []string) (map[string]int64, error)
	ValidateFamilyState(context.Context, string) error
}

// NetworkFlowClaim constructs the current dependency-resolved claim lifecycle.
// Callers supply source-owned counters and semantic validation, never SQL hooks.
func NetworkFlowClaim(t testing.TB, db *pgxpool.Pool, counters []extensionstore.FamilyCounter, validate func(context.Context, Reader) error) func() error {
	t.Helper()
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	physical, err := extensionstore.New(db, counters)
	if err != nil {
		t.Fatal(err)
	}
	logical, err := extensionassembly.NewStateStore(physical)
	if err != nil {
		t.Fatal(err)
	}
	runtime, err := extensions.NewStateRuntime(extensions.StateRuntimeOptions{
		Coordinator: coordinator, Store: logical, LockTimeout: time.Second, StepTimeout: time.Second, ProfileTimeout: 10 * time.Second, ValidationTimeout: 5 * time.Second,
		FatalIntegritySink: func(err error) { t.Error(err) },
		FinalValidators: map[string]extensions.FinalStateValidator{"network_flow_activity.validate_state_v6": func(ctx context.Context, input extensions.FinalStateValidationContext, reader extensions.StateReadCapability) (extensions.StateValidationResult, error) {
			if err := validate(ctx, reader); err != nil {
				return extensions.StateValidationResult{SchemaID: "cartulary.extension_final_state_validation_result.v1", Status: "invalid", Findings: []extensions.StartupFinding{coordinator.StateValidationFailure(input.ProfileID, "state_validation", "network_flow_activity.validate_state_v6")}}, nil
			}
			return extensions.ValidFinalStateValidationResult(), nil
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	return func() error { return runtime.AdmitClaims(t.Context(), coordinator, resolution) }
}
