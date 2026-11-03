package extensions

import (
	"context"
	"errors"
	"testing"
	"time"
)

func assertStateCallbackAuthority(t *testing.T) {
	for _, phase := range []string{"initializer", "migration", "pending", "final"} {
		for _, violation := range []string{"none", "counts", "validation"} {
			t.Run(phase+" capability "+violation, func(t *testing.T) {
				fixture := newSyntheticStateFixture(t, nil)
				plan := syntheticStatePlan(1)
				plan.InitializationKind, plan.InitializationAlgorithmID = "algorithm", "test_profile.initialize_v1"
				plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
				var retained StateReadCapability
				invoked := 0
				inspect := func(ctx context.Context, access StateReadCapability, write bool) {
					invoked++
					retained = access
					if _, ok := access.(interface {
						Commit(context.Context) (CommitOutcome, error)
					}); ok {
						t.Error("callback can commit")
					}
					if _, ok := access.(interface {
						Rollback(context.Context) (CommitOutcome, error)
					}); ok {
						t.Error("callback can roll back")
					}
					if _, ok := access.(interface {
						InsertStateMetadata(context.Context, StateMetadata) error
					}); ok {
						t.Error("callback can create coordinator metadata")
					}
					if _, ok := access.(interface {
						InsertMigrationLedger(context.Context, MigrationLedgerEntry) error
					}); ok {
						t.Error("callback can write coordinator ledger")
					}
					if _, ok := access.(interface {
						MigrationLedger(context.Context, string, string) ([]MigrationLedgerEntry, error)
					}); ok {
						t.Error("callback can inspect coordinator ledger")
					}
					if _, ok := access.(interface {
						StateMetadata(context.Context, string) (*StateMetadata, error)
					}); ok {
						t.Error("callback can inspect coordinator metadata")
					}
					if _, ok := access.(interface {
						UpdateStateMetadata(context.Context, StateMetadata, int, string, time.Time) error
					}); ok {
						t.Error("callback can update coordinator metadata")
					}
					if _, ok := access.(StateWriteCapability); ok != write {
						t.Errorf("write capability=%t want=%t", ok, write)
					}
					if _, ok := access.(syntheticMemberWriter); ok != write {
						t.Errorf("owner mutation capability=%t want=%t", ok, write)
					}
					ids := access.StateFamilyIDs()
					ids[0] = "outside_profile.members"
					if access.StateFamilyIDs()[0] != testStateFamilyID {
						t.Error("callback altered family scope")
					}
					switch violation {
					case "counts":
						if _, err := access.FamilyCounts(ctx, []string{"outside_profile.members"}); err == nil {
							t.Error("forbidden family counts succeeded")
						}
					case "validation":
						if err := access.ValidateFamilyState(ctx, "outside_profile.members"); err == nil {
							t.Error("forbidden family validation succeeded")
						}
					}
				}
				options := stateRuntimeTestOptions{initializers: map[string]StateInitializer{plan.InitializationAlgorithmID: func(ctx context.Context, _ InitializationContext, access StateWriteCapability) (InitializationResult, error) {
					if err := putSyntheticMember(ctx, access, "pending"); err != nil {
						return InitializationResult{}, err
					}
					if phase == "initializer" {
						inspect(ctx, access, true)
					}
					return readyInitialization(), nil
				}}, finalValidator: func(ctx context.Context, _ FinalStateValidationContext, access StateReadCapability) (StateValidationResult, error) {
					if phase == "final" {
						inspect(ctx, access, false)
					}
					return ValidFinalStateValidationResult(), nil
				}}
				if phase == "migration" || phase == "pending" {
					plan = syntheticStatePlan(2)
					seedSyntheticMetadata(t, fixture.pool, syntheticMetadata(plan.MigrationLineageID, 1, 1, nil))
					options.migrations = map[string]StateMigration{"test_profile.apply_1_2_v1": func(ctx context.Context, _ MigrationContext, access StateWriteCapability) (MigrationApplyResult, error) {
						if err := putSyntheticMember(ctx, access, "pending"); err != nil {
							return MigrationApplyResult{}, err
						}
						if phase == "migration" {
							inspect(ctx, access, true)
						}
						return readyMigration(), nil
					}}
					options.pendingValidators = map[string]PendingStateValidator{"test_profile.validate_1_2_v1": func(ctx context.Context, _ MigrationValidationContext, access StateReadCapability) (StateValidationResult, error) {
						if phase == "pending" {
							inspect(ctx, access, false)
						}
						return validMigrationValidationResult(), nil
					}}
				}
				runtime := newSyntheticStateRuntime(t, fixture.store, options)
				err := runtime.Admit(t.Context(), plan)
				if invoked != 1 {
					t.Fatalf("callback invocations=%d", invoked)
				}
				if violation == "none" {
					if err != nil || syntheticMemberCount(t, fixture.pool) != 1 {
						t.Fatalf("valid owner mutation failed: %v", err)
					}
				} else {
					var failure *StartupValidationError
					if !errors.As(err, &failure) || len(failure.Findings) != 1 {
						t.Fatalf("ignored scope violation was admitted: %v", err)
					}
					want := "extension_migration_failed"
					if phase == "initializer" {
						want = "extension_state_initialization_failed"
					}
					if phase == "final" {
						want = "extension_state_validation_failed"
					}
					if failure.Findings[0].ReasonCode != want {
						t.Fatalf("scope diagnostic=%#v", failure.Findings)
					}
					if syntheticMemberCount(t, fixture.pool) != 0 || ledgerCount(t, fixture.pool, plan.ProfileID) != 0 {
						t.Fatal("scope violation committed pending state")
					}
					if phase == "migration" || phase == "pending" {
						if readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
							t.Fatal("scope violation advanced metadata")
						}
					} else {
						requireNoSyntheticState(t, fixture.pool)
					}
				}
				if _, err := retained.FamilyCounts(t.Context(), []string{testStateFamilyID}); err == nil {
					t.Fatal("capability survived transaction completion")
				}
			})
		}
	}
}
