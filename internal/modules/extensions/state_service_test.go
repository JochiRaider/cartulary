package extensions

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestStateRuntimeAdmit_ServiceBacked_StatePresenceMatrix(t *testing.T) {
	tests := []struct {
		name          string
		policy        string
		metadata      bool
		authoritative bool
		wantDecision  StatePresenceDecision
		wantErr       error
	}{
		{name: "allowed fresh", policy: "allowed", wantDecision: StateInitialize},
		{name: "allowed metadata missing", policy: "allowed", authoritative: true, wantErr: ErrStateMetadataMissing},
		{name: "allowed empty restart", policy: "allowed", metadata: true, wantDecision: StateValidate},
		{name: "allowed populated", policy: "allowed", metadata: true, authoritative: true, wantDecision: StateValidate},
		{name: "forbidden fresh initializer admission", policy: "forbidden", wantDecision: StateInitialize},
		{name: "forbidden metadata missing", policy: "forbidden", authoritative: true, wantErr: ErrStateMetadataMissing},
		{name: "forbidden incomplete", policy: "forbidden", metadata: true, wantErr: ErrStateIncomplete},
		{name: "forbidden populated", policy: "forbidden", metadata: true, authoritative: true, wantDecision: StateValidate},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := DecideStatePresence(test.policy, test.metadata, test.authoritative)
			if got != test.wantDecision || !errors.Is(err, test.wantErr) {
				t.Fatalf("decision = %q/%v, want %q/%v", got, err, test.wantDecision, test.wantErr)
			}
		})
	}
}

func TestStateRuntimeAdmit_ServiceBacked_FreshEmptyInitialization(t *testing.T) {
	fixture := newNetworkFlowStateFixture(t, nil)
	if _, err := fixture.pool.Exec(context.Background(), `DELETE FROM extension_state_metadata WHERE profile_id = $1`, fixture.plan.ProfileID); err != nil {
		t.Fatal(err)
	}
	validationCalls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &validationCalls)
	if err := runtime.Admit(context.Background(), fixture.plan); err != nil {
		t.Fatalf("fresh empty admission: %v", err)
	}
	metadata := readStateMetadata(t, fixture.pool, fixture.plan.ProfileID)
	if metadata.StateVersion != 1 || metadata.MetadataVersion != 1 || metadata.LastMigrationID != nil || validationCalls != 1 {
		t.Fatalf("fresh metadata = %#v validation_calls=%d", metadata, validationCalls)
	}
	if ledgerCount(t, fixture.pool, fixture.plan.ProfileID) != 0 {
		t.Fatal("fresh initialization wrote a migration ledger row")
	}
}

func TestStateRuntimeAdmit_ServiceBacked_AllowedEmptyRestart(t *testing.T) {
	fixture := newNetworkFlowStateFixture(t, nil)
	validationCalls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &validationCalls)
	if err := runtime.Admit(context.Background(), fixture.plan); err != nil {
		t.Fatalf("allowed-empty restart: %v", err)
	}
	if validationCalls != 1 {
		t.Fatalf("final validation calls = %d, want 1", validationCalls)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_ForbiddenEmptyState(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	plan := syntheticStatePlan(1)
	plan.EmptyStatePolicy = "forbidden"
	plan.InitializationKind = "algorithm"
	plan.InitializationAlgorithmID = "test_profile.initialize_v1"
	plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		initializers: map[string]StateInitializer{
			plan.InitializationAlgorithmID: func(context.Context, InitializationContext, StateWriteCapability) (InitializationResult, error) {
				return readyInitialization(), nil
			},
		},
	})
	if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateIncomplete) {
		t.Fatalf("forbidden-empty initialization error = %v", err)
	}
	requireNoSyntheticState(t, fixture.pool)
}

func TestStateRuntimeAdmit_ServiceBacked_AlgorithmInitialization(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	plan := syntheticStatePlan(1)
	plan.EmptyStatePolicy = "forbidden"
	plan.InitializationKind = "algorithm"
	plan.InitializationAlgorithmID = "test_profile.initialize_v1"
	plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
	initializerCalls := 0
	finalCalls := 0
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		initializers: map[string]StateInitializer{
			plan.InitializationAlgorithmID: func(ctx context.Context, invocation InitializationContext, capability StateWriteCapability) (InitializationResult, error) {
				initializerCalls++
				if invocation.ProfileID != plan.ProfileID ||
					invocation.InitializationDefinitionSHA256 != plan.InitializationDefinitionSHA256 ||
					invocation.InitializationAlgorithmID != plan.InitializationAlgorithmID ||
					invocation.InitializationAlgorithmDefinitionSHA256 != plan.InitializationAlgorithmDefinitionSHA256 ||
					invocation.ScopedWriteCapabilityID != capability.CapabilityID() ||
					!sameTestStrings(invocation.AuthoritativeStateFamilyIDs, []string{testStateFamilyID}) {
					t.Fatalf("initialization context = %#v", invocation)
				}
				return readyInitialization(), putSyntheticMember(ctx, capability, "initialized")
			},
		},
		finalValidator: countingSyntheticFinalValidator(t, &finalCalls, 1),
	})
	if err := runtime.Admit(context.Background(), plan); err != nil {
		t.Fatalf("algorithm initialization: %v", err)
	}
	if initializerCalls != 1 || finalCalls != 1 || syntheticMemberCount(t, fixture.pool) != 1 {
		t.Fatalf("initializer=%d final=%d members=%d", initializerCalls, finalCalls, syntheticMemberCount(t, fixture.pool))
	}
	metadata := readStateMetadata(t, fixture.pool, plan.ProfileID)
	if metadata.StateVersion != 1 || metadata.MetadataVersion != 1 || metadata.LastMigrationID != nil {
		t.Fatalf("initialized metadata = %#v", metadata)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_InitializationValidationRollback(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	plan := syntheticStatePlan(1)
	plan.InitializationKind = "algorithm"
	plan.InitializationAlgorithmID = "test_profile.initialize_v1"
	plan.InitializationAlgorithmDefinitionSHA256 = testDigest("b")
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		initializers: map[string]StateInitializer{
			plan.InitializationAlgorithmID: func(ctx context.Context, _ InitializationContext, capability StateWriteCapability) (InitializationResult, error) {
				return readyInitialization(), putSyntheticMember(ctx, capability, "rollback")
			},
		},
		finalValidator: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
			return StateValidationResult{
				SchemaID: "cartulary.extension_final_state_validation_result.v1",
				Status:   "invalid",
				Findings: []StateFinding{{Code: "invalid", Path: "/"}},
			}, nil
		},
	})
	if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateValidationFailed) {
		t.Fatalf("initialization validation error = %v", err)
	}
	requireNoSyntheticState(t, fixture.pool)
}

func TestStateRuntimeAdmit_ServiceBacked_LineageAndVersionPreflight(t *testing.T) {
	tests := []struct {
		name     string
		metadata StateMetadata
		plan     StatePlan
		wantErr  error
	}{
		{
			name:     "wrong lineage",
			metadata: syntheticMetadata("other.lineage", 1, 1, nil),
			plan:     syntheticStatePlan(1),
			wantErr:  ErrStateLineageMismatch,
		},
		{
			name:     "newer state",
			metadata: syntheticMetadata("test_profile.state_v1", 4, 1, nil),
			plan:     syntheticStatePlan(3),
			wantErr:  ErrStateVersionUnsupported,
		},
		{
			name:     "older than minimum",
			metadata: syntheticMetadata("test_profile.state_v1", 1, 1, nil),
			plan: func() StatePlan {
				plan := syntheticStatePlan(3)
				plan.MinimumMigratableStateVersion = 2
				return plan
			}(),
			wantErr: ErrStateVersionUnsupported,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fixture := newSyntheticStateFixture(t, nil)
			seedSyntheticMetadata(t, fixture.pool, test.metadata)
			runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
				migrations:        standardSyntheticMigrations(t, nil),
				pendingValidators: standardSyntheticPendingValidators(),
			})
			if err := runtime.Admit(context.Background(), test.plan); !errors.Is(err, test.wantErr) {
				t.Fatalf("preflight error = %v, want %v", err, test.wantErr)
			}
			if syntheticMemberCount(t, fixture.pool) != 0 || ledgerCount(t, fixture.pool, test.plan.ProfileID) != 0 {
				t.Fatal("lineage/version preflight mutated state")
			}
		})
	}
}

func TestStateRuntimeAdmit_ServiceBacked_MigrationDefinitionClosure(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
	plan := syntheticStatePlan(3)
	plan.MigrationDefinitions = plan.MigrationDefinitions[:1]
	applyCalls := 0
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations: map[string]StateMigration{
			"test_profile.apply_1_2_v1": func(context.Context, MigrationContext, StateWriteCapability) (MigrationApplyResult, error) {
				applyCalls++
				return readyMigration(), nil
			},
		},
		pendingValidators: map[string]PendingStateValidator{
			"test_profile.validate_1_2_v1": validPendingValidator,
		},
	})
	if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateMigrationUnavailable) {
		t.Fatalf("migration closure error = %v", err)
	}
	if applyCalls != 0 || readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
		t.Fatal("incomplete migration sequence began mutation")
	}
}

func TestStateRuntimeAdmit_ServiceBacked_MigrationSequence(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
	plan := syntheticStatePlan(3)
	events := []string{}
	finalCalls := 0
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        standardSyntheticMigrations(t, &events),
		pendingValidators: standardSyntheticPendingValidators(),
		finalValidator:    countingSyntheticFinalValidator(t, &finalCalls, 2),
	})
	if err := runtime.Admit(context.Background(), plan); err != nil {
		t.Fatalf("migration sequence: %v", err)
	}
	if !sameTestStrings(events, []string{"apply-1-2", "apply-2-3"}) || finalCalls != 1 {
		t.Fatalf("events=%v final_calls=%d", events, finalCalls)
	}
	metadata := readStateMetadata(t, fixture.pool, plan.ProfileID)
	if metadata.StateVersion != 3 || metadata.MetadataVersion != 3 || metadata.LastMigrationID == nil || *metadata.LastMigrationID != "test_profile.migrate_2_3_v1" {
		t.Fatalf("migrated metadata = %#v", metadata)
	}
	if ledgerCount(t, fixture.pool, plan.ProfileID) != 2 || syntheticMemberCount(t, fixture.pool) != 2 {
		t.Fatalf("ledger=%d members=%d", ledgerCount(t, fixture.pool, plan.ProfileID), syntheticMemberCount(t, fixture.pool))
	}
}

func TestStateRuntimeAdmit_ServiceBacked_MigrationStepRollback(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
	plan := syntheticStatePlan(2)
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations: standardSyntheticMigrations(t, nil),
		pendingValidators: map[string]PendingStateValidator{
			"test_profile.validate_1_2_v1": func(context.Context, MigrationValidationContext, StateReadCapability) (StateValidationResult, error) {
				return StateValidationResult{
					SchemaID: "cartulary.extension_migration_validation_result.v1",
					Status:   "invalid",
					Findings: []StateFinding{{Code: "pending_invalid", Path: "/members"}},
				}, nil
			},
		},
	})
	if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateValidationFailed) {
		t.Fatalf("migration rollback error = %v", err)
	}
	if syntheticMemberCount(t, fixture.pool) != 0 || ledgerCount(t, fixture.pool, plan.ProfileID) != 0 || readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
		t.Fatal("failed pending-state validation committed a partial step")
	}
}

func TestStateRuntimeAdmit_ServiceBacked_FinalValidationFailureResumable(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
	plan := syntheticStatePlan(3)
	events := []string{}
	first := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        standardSyntheticMigrations(t, &events),
		pendingValidators: standardSyntheticPendingValidators(),
		finalValidator: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
			return StateValidationResult{
				SchemaID: "cartulary.extension_final_state_validation_result.v1",
				Status:   "invalid",
				Findings: []StateFinding{{Code: "final_invalid", Path: "/"}},
			}, nil
		},
	})
	if err := first.Admit(context.Background(), plan); !errors.Is(err, ErrStateValidationFailed) {
		t.Fatalf("first final validation error = %v", err)
	}
	if readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 3 || ledgerCount(t, fixture.pool, plan.ProfileID) != 2 {
		t.Fatal("final validation failure did not preserve resumable committed steps")
	}
	secondFinalCalls := 0
	second := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        standardSyntheticMigrations(t, &events),
		pendingValidators: standardSyntheticPendingValidators(),
		finalValidator:    countingSyntheticFinalValidator(t, &secondFinalCalls, 2),
	})
	if err := second.Admit(context.Background(), plan); err != nil {
		t.Fatalf("resumed final validation: %v", err)
	}
	if !sameTestStrings(events, []string{"apply-1-2", "apply-2-3"}) || secondFinalCalls != 1 {
		t.Fatalf("resume reapplied migration: events=%v final=%d", events, secondFinalCalls)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_AlreadyCurrentValidation(t *testing.T) {
	fixture := newNetworkFlowStateFixture(t, nil)
	validationCalls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &validationCalls)
	if err := runtime.Admit(context.Background(), fixture.plan); err != nil {
		t.Fatal(err)
	}
	if validationCalls != 1 {
		t.Fatalf("validation calls = %d, want 1", validationCalls)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_ConcurrentAdmissionAndLockLifetime(t *testing.T) {
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	attempted := make(chan struct{}, 2)
	faults := &stateTestFaults{onAttempt: func() { attempted <- struct{}{} }}
	fixture := newNetworkFlowStateFixture(t, faults)
	var validationCalls atomic.Int32
	runtime := newStateRuntimeForTest(t, fixture.store, StateRuntimeOptions{
		FinalValidators: map[string]FinalStateValidator{
			fixture.plan.FinalValidationAlgorithmID: func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
				if validationCalls.Add(1) == 1 {
					close(firstEntered)
					<-releaseFirst
				}
				return ValidFinalStateValidationResult(), nil
			},
		},
		LockTimeout: 5 * time.Second,
	})
	errs := make(chan error, 2)
	go func() { errs <- runtime.Admit(context.Background(), fixture.plan) }()
	<-attempted
	<-firstEntered
	go func() { errs <- runtime.Admit(context.Background(), fixture.plan) }()
	<-attempted
	acquired, _, maxActive := faults.lifecycle()
	if acquired != 1 || maxActive != 1 {
		t.Fatalf("concurrent lock lifecycle before release = acquired:%d max:%d", acquired, maxActive)
	}
	close(releaseFirst)
	for range 2 {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
	acquired, released, maxActive := faults.lifecycle()
	if acquired != 2 || released != 2 || maxActive != 1 {
		t.Fatalf("concurrent lock lifecycle = acquired:%d released:%d max:%d", acquired, released, maxActive)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_StepAndProfileDeadlines(t *testing.T) {
	t.Run("step deadline rolls back active step", func(t *testing.T) {
		fixture := newSyntheticStateFixture(t, nil)
		seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
		plan := syntheticStatePlan(2)
		var monotonic atomic.Int64
		migrations := standardSyntheticMigrations(t, nil)
		migrations["test_profile.apply_1_2_v1"] = func(ctx context.Context, _ MigrationContext, capability StateWriteCapability) (MigrationApplyResult, error) {
			if err := putSyntheticMember(ctx, capability, "timed-out"); err != nil {
				return MigrationApplyResult{}, err
			}
			monotonic.Store(int64(2 * time.Second))
			return readyMigration(), nil
		}
		runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
			migrations:        migrations,
			pendingValidators: standardSyntheticPendingValidators(),
			monotonicNowNS:    monotonic.Load,
			stepTimeout:       time.Second,
			profileTimeout:    10 * time.Second,
		})
		if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateMigrationTimeout) {
			t.Fatalf("step deadline error = %v", err)
		}
		if syntheticMemberCount(t, fixture.pool) != 0 || readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
			t.Fatal("timed-out step was not rolled back")
		}
	})

	t.Run("profile deadline prevents later step", func(t *testing.T) {
		var monotonic atomic.Int64
		faults := &stateTestFaults{
			onSnapshot: func(ordinal int) {
				if ordinal == 2 {
					monotonic.Store(int64(2 * time.Second))
				}
			},
		}
		fixture := newSyntheticStateFixture(t, faults)
		seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
		plan := syntheticStatePlan(3)
		events := []string{}
		runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
			migrations:        standardSyntheticMigrations(t, &events),
			pendingValidators: standardSyntheticPendingValidators(),
			monotonicNowNS:    monotonic.Load,
			stepTimeout:       10 * time.Second,
			profileTimeout:    time.Second,
		})
		if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateMigrationTimeout) {
			t.Fatalf("profile deadline error = %v", err)
		}
		if !sameTestStrings(events, []string{"apply-1-2"}) || readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 2 {
			t.Fatalf("profile deadline events=%v metadata=%#v", events, readStateMetadata(t, fixture.pool, plan.ProfileID))
		}
	})
}

func TestStateRuntimeAdmit_ServiceBacked_CancellationDeadlineTie(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata("test_profile.state_v1", 1, 1, nil))
	plan := syntheticStatePlan(2)
	var monotonic atomic.Int64
	ctx, cancel := context.WithCancel(context.Background())
	migrations := standardSyntheticMigrations(t, nil)
	migrations["test_profile.apply_1_2_v1"] = func(invokeCtx context.Context, _ MigrationContext, capability StateWriteCapability) (MigrationApplyResult, error) {
		if err := putSyntheticMember(invokeCtx, capability, "tie"); err != nil {
			return MigrationApplyResult{}, err
		}
		monotonic.Store(int64(time.Second))
		cancel()
		return readyMigration(), nil
	}
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        migrations,
		pendingValidators: standardSyntheticPendingValidators(),
		monotonicNowNS:    monotonic.Load,
		stepTimeout:       time.Second,
		profileTimeout:    10 * time.Second,
	})
	if err := runtime.Admit(ctx, plan); !errors.Is(err, ErrStateMigrationTimeout) {
		t.Fatalf("cancellation/deadline tie error = %v", err)
	}
	if syntheticMemberCount(t, fixture.pool) != 0 {
		t.Fatal("tie outcome did not roll back the active step")
	}
}

func TestStateRuntimeAdmit_ServiceBacked_CommitOutcomeCommitted(t *testing.T) {
	faults := &stateTestFaults{commitModes: []string{"committed_with_error"}}
	fixture := newSyntheticStateFixture(t, faults)
	plan := syntheticStatePlan(1)
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{})
	if err := runtime.Admit(context.Background(), plan); err != nil {
		t.Fatalf("proven committed outcome with lost acknowledgment: %v", err)
	}
	if readStateMetadata(t, fixture.pool, plan.ProfileID).StateVersion != 1 {
		t.Fatal("committed outcome did not pass exact readback")
	}
}

func TestStateRuntimeAdmit_ServiceBacked_CommitOutcomeAbsent(t *testing.T) {
	faults := &stateTestFaults{commitModes: []string{"absent"}}
	fixture := newSyntheticStateFixture(t, faults)
	plan := syntheticStatePlan(1)
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{})
	if err := runtime.Admit(context.Background(), plan); !errors.Is(err, ErrStateCommitAbsent) {
		t.Fatalf("proven absent commit error = %v", err)
	}
	requireNoSyntheticState(t, fixture.pool)
}

func TestStateRuntimeAdmit_ServiceBacked_CommitOutcomeIndeterminate(t *testing.T) {
	faults := &stateTestFaults{commitModes: []string{"indeterminate"}}
	fixture := newSyntheticStateFixture(t, faults)
	plan := syntheticStatePlan(1)
	var fatalCalls atomic.Int32
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		fatalSink: func(error) { fatalCalls.Add(1) },
	})
	err := runtime.Admit(context.Background(), plan)
	var fatalErr *FatalStateIntegrityError
	if !errors.As(err, &fatalErr) || fatalCalls.Load() != 1 {
		t.Fatalf("indeterminate outcome = %v fatal_calls=%d", err, fatalCalls.Load())
	}
}

func TestStateRuntimeAdmit_ServiceBacked_BorrowedSessionLifecycle(t *testing.T) {
	faults := &stateTestFaults{}
	fixture := newNetworkFlowStateFixture(t, faults)
	validationCalls := 0
	runtime := newNetworkFlowStateRuntime(t, fixture.store, &validationCalls)
	if err := runtime.Admit(context.Background(), fixture.plan); err != nil {
		t.Fatal(err)
	}
	acquired, released, maxActive := faults.lifecycle()
	if acquired != 1 || released != 1 || maxActive != 1 {
		t.Fatalf("borrowed session lifecycle = acquired:%d released:%d max:%d", acquired, released, maxActive)
	}
	if err := fixture.pool.Ping(context.Background()); err != nil {
		t.Fatalf("state admission closed borrowed PostgreSQL pool: %v", err)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_NetworkFlowFinalValidator(t *testing.T) {
	fixture := newNetworkFlowStateFixture(t, nil)
	physical, err := extensionstore.New(fixture.pool, networkflow.ExtensionStateFamilyCounters())
	if err != nil {
		t.Fatal(err)
	}
	if err := physical.WithProfileLock(context.Background(), fixture.plan.ProfileID, time.Second, func(session *extensionstore.Session) error {
		tx, err := session.Begin(context.Background())
		if err != nil {
			return err
		}
		defer func() { _, _ = tx.Rollback(context.Background()) }()
		return networkflow.ValidateExtensionState(context.Background(), tx)
	}); err != nil {
		t.Fatalf("validate Network Flow state through owner reader: %v", err)
	}
}

func TestStateRuntimeAdmit_ServiceBacked_CrashResumeLedger(t *testing.T) {
	fixture := newSyntheticStateFixture(t, nil)
	plan := syntheticStatePlan(3)
	lastMigration := plan.MigrationDefinitions[0].MigrationID
	seedSyntheticMetadata(t, fixture.pool, syntheticMetadata(plan.MigrationLineageID, 2, 2, &lastMigration))
	seedSyntheticLedger(t, fixture.pool, plan.MigrationDefinitions[0])
	if _, err := fixture.pool.Exec(context.Background(), `INSERT INTO extension_state_test_members (member_id) VALUES ('step-1-2')`); err != nil {
		t.Fatal(err)
	}
	events := []string{}
	runtime := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        standardSyntheticMigrations(t, &events),
		pendingValidators: standardSyntheticPendingValidators(),
	})
	if err := runtime.Admit(context.Background(), plan); err != nil {
		t.Fatalf("crash resume: %v", err)
	}
	if !sameTestStrings(events, []string{"apply-2-3"}) || ledgerCount(t, fixture.pool, plan.ProfileID) != 2 {
		t.Fatalf("crash resume events=%v ledger=%d", events, ledgerCount(t, fixture.pool, plan.ProfileID))
	}
	changed := cloneStatePlan(plan)
	changed.MigrationDefinitions[0].DefinitionSHA256 = testDigest("9")
	second := newSyntheticStateRuntime(t, fixture.store, stateRuntimeTestOptions{
		migrations:        standardSyntheticMigrations(t, nil),
		pendingValidators: standardSyntheticPendingValidators(),
	})
	if err := second.Admit(context.Background(), changed); !errors.Is(err, ErrStateMigrationDefinitionChanged) {
		t.Fatalf("changed committed migration digest error = %v", err)
	}
}

type stateServiceFixture struct {
	pool  *pgxpool.Pool
	store *liveStateTestStore
	plan  StatePlan
}

func newNetworkFlowStateFixture(t testing.TB, faults *stateTestFaults) stateServiceFixture {
	t.Helper()
	pool := newStateServicePool(t)
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := coordinator.StatePlan("network_flow_activity")
	if !ok {
		t.Fatal("generated Network Flow state plan missing")
	}
	physical, err := extensionstore.New(pool, networkflow.ExtensionStateFamilyCounters())
	if err != nil {
		t.Fatal(err)
	}
	return stateServiceFixture{pool: pool, store: newLiveStateTestStore(physical, faults), plan: plan}
}

func newSyntheticStateFixture(t testing.TB, faults *stateTestFaults) stateServiceFixture {
	t.Helper()
	pool := newStateServicePool(t)
	if _, err := pool.Exec(context.Background(), `
CREATE TABLE extension_state_test_members (
    member_id text PRIMARY KEY
)
`); err != nil {
		t.Fatal(err)
	}
	physical, err := extensionstore.New(pool, []extensionstore.FamilyCounter{{
		FamilyID: testStateFamilyID,
		Count: func(ctx context.Context, querier extensionstore.Querier) (int64, error) {
			var count int64
			err := querier.QueryRow(ctx, `SELECT COUNT(*) FROM extension_state_test_members`).Scan(&count)
			return count, err
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
	return stateServiceFixture{pool: pool, store: newLiveStateTestStore(physical, faults), plan: syntheticStatePlan(1)}
}

func newStateServicePool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	harness := pgtest.Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "extensions-state-remediation")
	pool, err := pgxpool.New(context.Background(), testDB.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

type stateRuntimeTestOptions struct {
	initializers      map[string]StateInitializer
	migrations        map[string]StateMigration
	pendingValidators map[string]PendingStateValidator
	finalValidator    FinalStateValidator
	monotonicNowNS    func() int64
	lockTimeout       time.Duration
	stepTimeout       time.Duration
	profileTimeout    time.Duration
	validationTimeout time.Duration
	fatalSink         func(error)
}

func newSyntheticStateRuntime(t testing.TB, store StateStore, options stateRuntimeTestOptions) *StateRuntime {
	t.Helper()
	finalValidator := options.finalValidator
	if finalValidator == nil {
		finalValidator = func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error) {
			return ValidFinalStateValidationResult(), nil
		}
	}
	return newStateRuntimeForTest(t, store, StateRuntimeOptions{
		Initializers:      options.initializers,
		Migrations:        options.migrations,
		PendingValidators: options.pendingValidators,
		FinalValidators: map[string]FinalStateValidator{
			"test_profile.final_state_v1": finalValidator,
		},
		MonotonicNowNS:     options.monotonicNowNS,
		LockTimeout:        options.lockTimeout,
		StepTimeout:        options.stepTimeout,
		ProfileTimeout:     options.profileTimeout,
		ValidationTimeout:  options.validationTimeout,
		FatalIntegritySink: options.fatalSink,
	})
}

func newNetworkFlowStateRuntime(t testing.TB, store StateStore, validationCalls *int) *StateRuntime {
	t.Helper()
	return newStateRuntimeForTest(t, store, StateRuntimeOptions{
		FinalValidators: map[string]FinalStateValidator{
			"network_flow_activity.validate_state_v1": func(ctx context.Context, _ FinalStateValidationContext, reader StateReadCapability) (StateValidationResult, error) {
				(*validationCalls)++
				if err := networkflow.ValidateExtensionState(ctx, reader); err != nil {
					return StateValidationResult{
						SchemaID: "cartulary.extension_final_state_validation_result.v1",
						Status:   "invalid",
						Findings: []StateFinding{{Code: "network_flow_activity_state_invalid", Path: "/"}},
					}, nil
				}
				return ValidFinalStateValidationResult(), nil
			},
		},
	})
}

func newStateRuntimeForTest(t testing.TB, store StateStore, options StateRuntimeOptions) *StateRuntime {
	t.Helper()
	options.Store = store
	if options.Now == nil {
		options.Now = func() time.Time { return time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC) }
	}
	if options.LockTimeout <= 0 {
		options.LockTimeout = time.Second
	}
	if options.StepTimeout <= 0 {
		options.StepTimeout = time.Second
	}
	if options.ProfileTimeout <= 0 {
		options.ProfileTimeout = 10 * time.Second
	}
	if options.ValidationTimeout <= 0 {
		options.ValidationTimeout = time.Second
	}
	if options.FatalIntegritySink == nil {
		options.FatalIntegritySink = func(err error) { t.Fatalf("unexpected fatal integrity sink: %v", err) }
	}
	runtime, err := NewStateRuntime(options)
	if err != nil {
		t.Fatal(err)
	}
	return runtime
}

func syntheticStatePlan(currentVersion int) StatePlan {
	plan := StatePlan{
		ProfileID:                      "test_profile",
		ContractMajor:                  1,
		MigrationLineageID:             "test_profile.state_v1",
		CurrentStateVersion:            currentVersion,
		MinimumMigratableStateVersion:  1,
		EmptyStatePolicy:               "allowed",
		DatabaseFamilyIDs:              []string{testStateFamilyID},
		ObjectReferenceFamilyIDs:       []string{},
		InitializationKind:             "empty",
		InitializationDefinitionSHA256: testDigest("a"),
		FinalValidationAlgorithmID:     "test_profile.final_state_v1",
		PhysicalStateBindingSHA256:     testDigest("c"),
		StatePresenceManifestSHA256:    testDigest("d"),
		ImplementationBindingSHA256:    testDigest("e"),
	}
	if currentVersion >= 2 {
		plan.MigrationDefinitions = append(plan.MigrationDefinitions, syntheticMigrationDefinition(1, 2, "f"))
	}
	if currentVersion >= 3 {
		plan.MigrationDefinitions = append(plan.MigrationDefinitions, syntheticMigrationDefinition(2, 3, "1"))
	}
	return plan
}

func syntheticMigrationDefinition(fromVersion, toVersion int, digestCharacter string) MigrationDefinition {
	return MigrationDefinition{
		MigrationLineageID:             "test_profile.state_v1",
		MigrationID:                    "test_profile.migrate_" + string(rune('0'+fromVersion)) + "_" + string(rune('0'+toVersion)) + "_v1",
		FromVersion:                    fromVersion,
		ToVersion:                      toVersion,
		DefinitionSHA256:               testDigest(digestCharacter),
		ApplyAlgorithmID:               "test_profile.apply_" + string(rune('0'+fromVersion)) + "_" + string(rune('0'+toVersion)) + "_v1",
		ValidationAlgorithmID:          "test_profile.validate_" + string(rune('0'+fromVersion)) + "_" + string(rune('0'+toVersion)) + "_v1",
		ImplementationBindingProfileID: "test_profile",
		ImplementationBindingSHA256:    testDigest("e"),
	}
}

func standardSyntheticMigrations(t testing.TB, events *[]string) map[string]StateMigration {
	t.Helper()
	return map[string]StateMigration{
		"test_profile.apply_1_2_v1": func(ctx context.Context, invocation MigrationContext, capability StateWriteCapability) (MigrationApplyResult, error) {
			if invocation.MigrationID != "test_profile.migrate_1_2_v1" || invocation.StateAccessCapabilityID != capability.CapabilityID() {
				t.Fatalf("step 1 context = %#v", invocation)
			}
			if events != nil {
				*events = append(*events, "apply-1-2")
			}
			return readyMigration(), putSyntheticMember(ctx, capability, "step-1-2")
		},
		"test_profile.apply_2_3_v1": func(ctx context.Context, invocation MigrationContext, capability StateWriteCapability) (MigrationApplyResult, error) {
			if invocation.MigrationID != "test_profile.migrate_2_3_v1" || invocation.StateAccessCapabilityID != capability.CapabilityID() {
				t.Fatalf("step 2 context = %#v", invocation)
			}
			if events != nil {
				*events = append(*events, "apply-2-3")
			}
			return readyMigration(), putSyntheticMember(ctx, capability, "step-2-3")
		},
	}
}

func standardSyntheticPendingValidators() map[string]PendingStateValidator {
	return map[string]PendingStateValidator{
		"test_profile.validate_1_2_v1": validPendingValidator,
		"test_profile.validate_2_3_v1": validPendingValidator,
	}
}

func validPendingValidator(_ context.Context, invocation MigrationValidationContext, reader StateReadCapability) (StateValidationResult, error) {
	if invocation.StateAccessCapabilityID != reader.CapabilityID() || !sameTestStrings(invocation.StateFamilyIDs, reader.StateFamilyIDs()) {
		return StateValidationResult{}, errors.New("pending validation capability mismatch")
	}
	return ValidMigrationValidationResult(), nil
}

func countingSyntheticFinalValidator(t testing.TB, calls *int, wantMembers int64) FinalStateValidator {
	t.Helper()
	return func(ctx context.Context, invocation FinalStateValidationContext, reader StateReadCapability) (StateValidationResult, error) {
		(*calls)++
		if invocation.ReadOnlyStateAccessCapabilityID != reader.CapabilityID() ||
			!sameTestStrings(invocation.StateFamilyIDs, []string{testStateFamilyID}) ||
			invocation.StateMetadataSHA256 == "" {
			t.Fatalf("final validation context = %#v", invocation)
		}
		counts, err := reader.FamilyCounts(ctx, invocation.StateFamilyIDs)
		if err != nil {
			return StateValidationResult{}, err
		}
		if counts[testStateFamilyID] != wantMembers {
			return StateValidationResult{
				SchemaID: "cartulary.extension_final_state_validation_result.v1",
				Status:   "invalid",
				Findings: []StateFinding{{Code: "member_count", Path: "/members"}},
			}, nil
		}
		return ValidFinalStateValidationResult(), nil
	}
}

type syntheticMemberWriter interface {
	StateWriteCapability
	PutTestMember(context.Context, string) error
}

func putSyntheticMember(ctx context.Context, capability StateWriteCapability, memberID string) error {
	writer, ok := capability.(syntheticMemberWriter)
	if !ok {
		return errors.New("synthetic profile write capability unavailable")
	}
	return writer.PutTestMember(ctx, memberID)
}

func readyInitialization() InitializationResult {
	return InitializationResult{
		SchemaID: "cartulary.extension_state_initialization_result.v1",
		Status:   "ready_to_validate",
	}
}
func readyMigration() MigrationApplyResult {
	return MigrationApplyResult{
		SchemaID: "cartulary.extension_migration_apply_result.v1",
		Status:   "ready_to_validate",
	}
}

func syntheticMetadata(lineageID string, stateVersion, metadataVersion int, lastMigrationID *string) StateMetadata {
	now := time.Date(2026, 7, 24, 10, 0, 0, 0, time.UTC)
	return StateMetadata{
		ProfileID:          "test_profile",
		MigrationLineageID: lineageID,
		StateVersion:       stateVersion,
		LastMigrationID:    cloneTestString(lastMigrationID),
		MetadataVersion:    metadataVersion,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func seedSyntheticMetadata(t testing.TB, pool *pgxpool.Pool, metadata StateMetadata) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO extension_state_metadata (
    profile_id, migration_lineage_id, state_version, last_migration_id,
    metadata_version, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
`, metadata.ProfileID, metadata.MigrationLineageID, metadata.StateVersion, metadata.LastMigrationID,
		metadata.MetadataVersion, metadata.CreatedAt, metadata.UpdatedAt); err != nil {
		t.Fatal(err)
	}
}

func seedSyntheticLedger(t testing.TB, pool *pgxpool.Pool, definition MigrationDefinition) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
INSERT INTO extension_migration_ledger (
    profile_id, migration_lineage_id, migration_id, from_state_version,
    to_state_version, migration_definition_sha256, committed_at,
    resulting_state_version
) VALUES ('test_profile', $1, $2, $3, $4, $5, $6, $4)
`, definition.MigrationLineageID, definition.MigrationID, definition.FromVersion,
		definition.ToVersion, definition.DefinitionSHA256,
		time.Date(2026, 7, 24, 11, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
}

func readStateMetadata(t testing.TB, pool *pgxpool.Pool, profileID string) StateMetadata {
	t.Helper()
	var metadata StateMetadata
	if err := pool.QueryRow(context.Background(), `
SELECT profile_id, migration_lineage_id, state_version, last_migration_id,
       metadata_version, created_at, updated_at
  FROM extension_state_metadata
 WHERE profile_id = $1
`, profileID).Scan(
		&metadata.ProfileID,
		&metadata.MigrationLineageID,
		&metadata.StateVersion,
		&metadata.LastMigrationID,
		&metadata.MetadataVersion,
		&metadata.CreatedAt,
		&metadata.UpdatedAt,
	); err != nil {
		t.Fatal(err)
	}
	return metadata
}

func syntheticMemberCount(t testing.TB, pool *pgxpool.Pool) int64 {
	t.Helper()
	var count int64
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM extension_state_test_members`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func ledgerCount(t testing.TB, pool *pgxpool.Pool, profileID string) int {
	t.Helper()
	var count int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM extension_migration_ledger WHERE profile_id = $1`, profileID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}

func requireNoSyntheticState(t testing.TB, pool *pgxpool.Pool) {
	t.Helper()
	if syntheticMemberCount(t, pool) != 0 {
		t.Fatal("synthetic authoritative state was not rolled back")
	}
	var metadataCount int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM extension_state_metadata WHERE profile_id = 'test_profile'`).Scan(&metadataCount); err != nil {
		t.Fatal(err)
	}
	if metadataCount != 0 || ledgerCount(t, pool, "test_profile") != 0 {
		t.Fatalf("synthetic coordination state remains: metadata=%d ledger=%d", metadataCount, ledgerCount(t, pool, "test_profile"))
	}
}

func testDigest(character string) string {
	return strings.Repeat(character, 64)
}
