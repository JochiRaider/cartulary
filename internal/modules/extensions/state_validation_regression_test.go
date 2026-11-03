package extensions

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func assertStateResultPrecedence(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	for _, test := range []struct {
		name     string
		count    int
		mutation string
		want     string
	}{
		{"empty", 0, "", ""}, {"absent findings", 0, "nil", "extension_validation_result_invalid"},
		{"schema limit", 256, "", "extension_state_validation_failed"},
		{"257 is schema invalid", 257, "", "extension_validation_result_invalid"},
		{"4096 is schema invalid", 4096, "", "extension_validation_result_invalid"},
		{"overflow precedes schema", 4097, "schema", "extension_diagnostic_overflow"},
		{"overflow bounded actual", 5000, "schema", "extension_diagnostic_overflow"},
		{"invocation precedes overflow", 4097, "invoke", "extension_state_validation_failed"},
		{"panic", 0, "panic", "extension_state_validation_failed"},
		{"cancel without result", 0, "cancel", "extension_state_validation_failed"},
		{"private message", 1, "private", "extension_validation_result_invalid"},
		{"private path", 1, "private_path", "extension_validation_result_invalid"},
		{"unknown details", 1, "extra", "extension_validation_result_invalid"},
		{"unsorted", 2, "order", "extension_validation_result_invalid"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fixture := newSyntheticStateFixture(t, nil)
			plan := syntheticStatePlan(1)
			plan.InitializationKind = "algorithm"
			plan.InitializationAlgorithmID = "test_profile.initialize_v1"
			plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				initializers: map[string]StateInitializer{plan.InitializationAlgorithmID: func(ctx context.Context, _ InitializationContext, write StateWriteCapability) (InitializationResult, error) {
					return readyInitialization(), putSyntheticMember(ctx, write, "validation-rollback")
				}},
				finalValidator: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
					result := ValidFinalStateValidationResult()
					for i := 0; i < test.count; i++ {
						result.Findings = append(result.Findings, coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID))
					}
					if test.count > 0 {
						result.Status = "invalid"
					}
					switch test.mutation {
					case "nil":
						result.Findings = nil
					case "schema":
						result.SchemaID = "private"
						result.Findings[0].Message = "private bytes"
					case "invoke":
						return result, errors.New("private bytes")
					case "panic":
						panic("private bytes")
					case "cancel":
						cancel()
						return StateValidationResult{}, nil
					case "private":
						result.Findings[0].Message = "private bytes"
					case "private_path":
						result.Findings[0].Path = `$["/private/storage"]`
					case "extra":
						result.Findings[0].Details["private"] = "private bytes"
					case "order":
						result.Findings[0].Path = "$.z"
						result.Findings[1].Path = "$.a"
					}
					return result, nil
				},
			})
			err := runtime.Admit(ctx, plan)
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var failure *StartupValidationError
			if !errors.Is(err, ErrStateValidationFailed) || !errors.As(err, &failure) || len(failure.Findings) == 0 || failure.Findings[0].ReasonCode != test.want {
				t.Fatalf("classification = %v; want %s", err, test.want)
			}
			if strings.Contains(fmt.Sprint(failure.Findings), "private") {
				t.Fatal("private callback value escaped")
			}
			if test.want == "extension_diagnostic_overflow" && (len(failure.Findings) != 1 || failure.Findings[0].Details["actual"] != 4097) {
				t.Fatal("overflow did not replace ordinary findings")
			}
			requireNoSyntheticState(t, fixture.pool)
		})
	}
}

func assertPendingResultPrecedence(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	for _, test := range []struct {
		name     string
		count    int
		mutation string
		want     string
	}{
		{"nil", 0, "nil", "extension_validation_result_invalid"},
		{"semantic", 256, "", "extension_migration_failed"},
		{"257", 257, "", "extension_validation_result_invalid"},
		{"4096", 4096, "", "extension_validation_result_invalid"},
		{"overflow", 4097, "schema", "extension_diagnostic_overflow"},
		{"invocation", 4097, "invoke", "extension_migration_failed"},
		{"panic", 0, "panic", "extension_migration_failed"},
	} {
		t.Run("pending "+test.name, func(t *testing.T) {
			fixture := newSyntheticStateFixture(t, nil)
			seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
			plan := syntheticStatePlan(2)
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				migrations: standardSyntheticMigrations(t, nil),
				pendingValidators: map[string]PendingStateValidator{"test_profile.validate_1_2_v1": func(context.Context, MigrationValidationContext, StateReadCapability) (StateValidationResult, error) {
					result := validMigrationValidationResult()
					for i := 0; i < test.count; i++ {
						result.Findings = append(result.Findings, coordinator.migrationValidationFailure(plan.ProfileID, "test_profile.migrate_1_2_v1", 1, 2))
					}
					if test.count > 0 {
						result.Status = "invalid"
					}
					switch test.mutation {
					case "nil":
						result.Findings = nil
					case "schema":
						result.SchemaID = "private"
						result.Findings[0].Message = "private bytes"
					case "invoke":
						return result, errors.New("private bytes")
					case "panic":
						panic("private bytes")
					}
					return result, nil
				}},
			})
			err := runtime.Admit(context.Background(), plan)
			var failure *StartupValidationError
			if !errors.Is(err, ErrStateValidationFailed) || !errors.As(err, &failure) || len(failure.Findings) == 0 || failure.Findings[0].ReasonCode != test.want {
				t.Fatalf("pending classification = %v; want %s", err, test.want)
			}
			if strings.Contains(fmt.Sprint(failure.Findings), "private") || syntheticMemberCount(t, fixture.pool) != 0 || ledgerCount(t, fixture.pool, plan.ProfileID) != 0 || readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
				t.Fatal("pending failure escaped or committed state")
			}
		})
	}
}
