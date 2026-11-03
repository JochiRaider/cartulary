package extensions

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func assertInitializationDiagnosticOwnership(t *testing.T) {
	for _, outcome := range []string{"error", "panic", "cancellation", "schema", "status", "scope", "rollback timeout", "rollback timeout with error"} {
		t.Run("initializer diagnostic "+outcome, func(t *testing.T) {
			var clock, deadline atomic.Int64
			clock.Store(1)
			faults := &stateTestFaults{}
			fixture := newSyntheticStateFixture(t, faults)
			plan := syntheticStatePlan(1)
			plan.InitializationKind = "algorithm"
			plan.InitializationAlgorithmID = "test_profile.initialize_v1"
			plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			finalCalls := 0
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				monotonicNowNS: clock.Load, stepTimeout: 3 * time.Second, profileTimeout: 9 * time.Second,
				initializers: map[string]StateInitializer{plan.InitializationAlgorithmID: func(ctx context.Context, input InitializationContext, capability StateWriteCapability) (InitializationResult, error) {
					deadline.Store(input.DeadlineMonotonicNS)
					if strings.HasPrefix(outcome, "rollback timeout") {
						faults.onRollback = func() { clock.Store(deadline.Load()) }
					}
					if outcome == "rollback timeout with error" {
						faults.rollbackError = errors.New("PRIVATE_ROLLBACK_PAYLOAD")
					}
					if err := putSyntheticMember(ctx, capability, "pending"); err != nil {
						return InitializationResult{}, err
					}
					switch outcome {
					case "error", "rollback timeout", "rollback timeout with error":
						return InitializationResult{}, errors.New("PRIVATE_INITIALIZER_PAYLOAD")
					case "panic":
						panic("PRIVATE_INITIALIZER_PAYLOAD")
					case "cancellation":
						cancel()
						return readyInitialization(), nil
					case "schema":
						return InitializationResult{SchemaID: "wrong", Status: "ready_to_validate"}, nil
					case "status":
						return InitializationResult{SchemaID: "cartulary.extension_state_initialization_result.v1", Status: "wrong"}, nil
					case "scope":
						_, err := capability.FamilyCounts(ctx, []string{"outside_profile.members"})
						return readyInitialization(), err
					}
					panic("unreachable initializer case")
				}},
				finalValidator: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
					finalCalls++
					return ValidFinalStateValidationResult(), nil
				},
			})
			err := runtime.Admit(ctx, plan)
			code, message := "extension_state_initialization_failed", "Extension state initialization failed."
			details := map[string]any{"profile_id": plan.ProfileID, "algorithm_id": plan.InitializationAlgorithmID}
			if strings.HasPrefix(outcome, "rollback timeout") {
				code, message = "extension_state_initialization_timeout", "Extension state initialization exceeded its timeout."
				details = map[string]any{"profile_id": plan.ProfileID, "timeout_seconds": int64(3)}
			}
			requireInitializationFinding(t, runtime, err, code, message, details)
			if finalCalls != 0 {
				t.Fatalf("failed initializer invoked final validation %d times", finalCalls)
			}
			requireNoSyntheticState(t, fixture.pool)
		})
	}
	for _, mode := range []string{"committed_with_error", "absent", "indeterminate"} {
		t.Run("initialization proof at expiry "+mode, func(t *testing.T) {
			var clock, deadline atomic.Int64
			clock.Store(1)
			faults := &stateTestFaults{commitModes: []string{mode}, onCommit: func() { clock.Store(deadline.Load()) }}
			fixture := newSyntheticStateFixture(t, faults)
			plan := syntheticStatePlan(1)
			plan.InitializationKind, plan.InitializationAlgorithmID = "algorithm", "test_profile.initialize_v1"
			plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
			fatals := 0
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				monotonicNowNS: clock.Load, stepTimeout: 3 * time.Second, profileTimeout: 9 * time.Second,
				fatalSink: func(error) { fatals++ },
				initializers: map[string]StateInitializer{plan.InitializationAlgorithmID: func(ctx context.Context, input InitializationContext, capability StateWriteCapability) (InitializationResult, error) {
					deadline.Store(input.DeadlineMonotonicNS)
					return readyInitialization(), putSyntheticMember(ctx, capability, "pending")
				}},
			})
			err := runtime.Admit(t.Context(), plan)
			switch mode {
			case "committed_with_error":
				if err != nil || fatals != 0 || syntheticMemberCount(t, fixture.pool) != 1 {
					t.Fatalf("proven initialization commit: %v / fatals=%d", err, fatals)
				}
				if readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 || ledgerCount(t, fixture.pool, plan.ProfileID) != 0 {
					t.Fatal("committed initialization metadata mismatch")
				}
			case "absent":
				requireInitializationFinding(t, runtime, err, "extension_state_initialization_timeout", "Extension state initialization exceeded its timeout.", map[string]any{"profile_id": plan.ProfileID, "timeout_seconds": int64(3)})
				requireNoSyntheticState(t, fixture.pool)
			case "indeterminate":
				var fatal *FatalStateIntegrityError
				var finding *StartupValidationError
				if !errors.As(err, &fatal) || errors.As(err, &finding) || fatals != 1 {
					t.Fatalf("unknown initialization outcome: %v / fatals=%d", err, fatals)
				}
			}
		})
	}
	t.Run("final validation and rollback errors remain secret safe", func(t *testing.T) {
		faults := &stateTestFaults{}
		fixture := newSyntheticStateFixture(t, faults)
		plan := syntheticStatePlan(1)
		runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{finalValidator: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
			faults.rollbackError = errors.New("PRIVATE_ROLLBACK_PAYLOAD")
			return StateValidationResult{}, errors.New("PRIVATE_FINAL_PAYLOAD")
		}})
		err := runtime.Admit(t.Context(), plan)
		finding := runtime.coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID)
		requireInitializationFinding(t, runtime, err, finding.ReasonCode, finding.Message, finding.Details)
		requireNoSyntheticState(t, fixture.pool)
	})

	for _, tc := range []struct {
		name, kind, phase, owner string
		step, profile, final     int64
	}{
		{"local initialization", "algorithm", "initialize", "initialize", 3, 9, 2},
		{"earlier profile", "algorithm", "initialize", "profile", 4, 2, 1},
		{"equal profile", "algorithm", "initialize", "profile", 3, 3, 1},
		{"empty nested initialization", "empty", "final", "initialize", 2, 8, 5},
		{"empty equal nested initialization", "empty", "final", "initialize", 2, 8, 2},
		{"algorithm nested profile", "algorithm", "final", "profile", 4, 2, 6},
		{"local final validation", "algorithm", "final", "final", 5, 9, 1},
	} {
		t.Run("selected initialization deadline "+tc.name, func(t *testing.T) {
			fixture := newSyntheticStateFixture(t, nil)
			plan := syntheticStatePlan(1)
			plan.InitializationKind = tc.kind
			if tc.kind == "algorithm" {
				plan.InitializationAlgorithmID = "test_profile.initialize_v1"
				plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
			}
			var clock atomic.Int64
			clock.Store(1)
			finalCalls := 0
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				monotonicNowNS: clock.Load, stepTimeout: time.Duration(tc.step) * time.Second, profileTimeout: time.Duration(tc.profile) * time.Second, validationTimeout: time.Duration(tc.final) * time.Second,
				initializers: map[string]StateInitializer{plan.InitializationAlgorithmID: func(ctx context.Context, input InitializationContext, capability StateWriteCapability) (InitializationResult, error) {
					if err := putSyntheticMember(ctx, capability, "pending"); err != nil {
						return InitializationResult{}, err
					}
					if tc.phase == "initialize" {
						clock.Store(input.DeadlineMonotonicNS)
					}
					return readyInitialization(), nil
				}},
				finalValidator: func(_ context.Context, input FinalStateValidationContext, _ StateReadCapability) (StateValidationResult, error) {
					finalCalls++
					clock.Store(input.DeadlineMonotonicNS)
					return ValidFinalStateValidationResult(), nil
				},
			})
			err := runtime.Admit(t.Context(), plan)
			code, message := "extension_state_initialization_timeout", "Extension state initialization exceeded its timeout."
			details := map[string]any{"profile_id": plan.ProfileID, "timeout_seconds": tc.step}
			switch tc.owner {
			case "profile":
				code, message = "extension_migration_timeout", "Extension state migration exceeded its timeout."
				details = map[string]any{"profile_id": plan.ProfileID, "operation_kind": "profile_migration", "migration_id": nil, "from_state_version": nil, "to_state_version": nil, "timeout_seconds": tc.profile}
			case "final":
				finding := runtime.coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID)
				code, message, details = finding.ReasonCode, finding.Message, finding.Details
			}
			requireInitializationFinding(t, runtime, err, code, message, details)
			wantFinal := 0
			if tc.phase == "final" {
				wantFinal = 1
			}
			if finalCalls != wantFinal {
				t.Fatalf("final calls=%d want=%d", finalCalls, wantFinal)
			}
			requireNoSyntheticState(t, fixture.pool)
		})
	}
}

func requireInitializationFinding(t *testing.T, runtime *StateRuntime, err error, code, message string, details map[string]any) {
	t.Helper()
	var failure *StartupValidationError
	if !errors.As(err, &failure) || len(failure.Findings) != 1 {
		t.Fatalf("missing initialization finding: %v", err)
	}
	finding := failure.Findings[0]
	if finding.Path != "$" || finding.ReasonCode != code || finding.Message != message || !equalJSONValue(finding.Details, details) || !runtime.coordinator.validStartupFinding(finding) {
		t.Fatalf("initialization finding=%#v, want %s/%s/%v", finding, code, message, details)
	}
	encoded, _ := json.Marshal(finding)
	if strings.Contains(string(encoded), "PRIVATE_") || strings.Contains(err.Error(), "PRIVATE_") {
		t.Fatal("initializer diagnostic exposed private failure")
	}
}
