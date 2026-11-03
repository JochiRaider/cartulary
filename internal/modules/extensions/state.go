package extensions

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

var (
	ErrStateMetadataMissing            = errors.New("extension_state_metadata_missing")
	ErrStateIncomplete                 = errors.New("extension_state_incomplete")
	ErrStateLineageMismatch            = errors.New("extension_state_lineage_mismatch")
	ErrStateVersionUnsupported         = errors.New("extension_state_version_unsupported")
	ErrStateMigrationUnavailable       = errors.New("extension_state_migration_unavailable")
	ErrStateMigrationDefinitionChanged = errors.New("extension_migration_definition_changed")
	ErrStateValidationFailed           = errors.New("extension_state_validation_failed")
	ErrStateMigrationTimeout           = errors.New("extension_migration_timeout")
	ErrStateMigrationLockTimeout       = errors.New("extension_migration_lock_timeout")
	ErrStateCommitAbsent               = errors.New("extension_state_commit_absent")
	ErrStateCommitIndeterminate        = errors.New("extension_state_commit_indeterminate")
	ErrStateReadbackMismatch           = errors.New("extension_state_readback_mismatch")
)

type StatePresenceDecision string

const (
	StateInitialize StatePresenceDecision = "initialize"
	StateValidate   StatePresenceDecision = "validate"
)

// DecideStatePresence is the closed metadata/authoritative-state table. Generic
// coordination rows, migration ledgers, jobs, caches, projections, and staged
// objects are intentionally not inputs.
func DecideStatePresence(policy string, metadataPresent, authoritativeStatePresent bool) (StatePresenceDecision, error) {
	switch {
	case !metadataPresent && authoritativeStatePresent:
		return "", ErrStateMetadataMissing
	case !metadataPresent && !authoritativeStatePresent:
		if policy == "allowed" || policy == "forbidden" {
			return StateInitialize, nil
		}
	case metadataPresent && !authoritativeStatePresent:
		if policy == "allowed" {
			return StateValidate, nil
		}
		if policy == "forbidden" {
			return "", ErrStateIncomplete
		}
	case metadataPresent && authoritativeStatePresent:
		if policy == "allowed" || policy == "forbidden" {
			return StateValidate, nil
		}
	}
	return "", ErrStateIncomplete
}

type StateMetadata struct {
	ProfileID          string
	MigrationLineageID string
	StateVersion       int
	LastMigrationID    *string
	MetadataVersion    int
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type MigrationLedgerEntry struct {
	ProfileID                 string
	MigrationLineageID        string
	MigrationID               string
	FromStateVersion          int
	ToStateVersion            int
	MigrationDefinitionSHA256 string
	CommittedAt               time.Time
	ResultingStateVersion     int
}

type StateSnapshot struct {
	Metadata     *StateMetadata
	FamilyCounts map[string]int64
	Ledger       []MigrationLedgerEntry
}

type CommitOutcome string

const (
	CommitCommitted     CommitOutcome = "committed"
	CommitAbsent        CommitOutcome = "absent"
	CommitIndeterminate CommitOutcome = "indeterminate"
)

// StateReadCapability exposes only declared logical authoritative families.
// Physical tables, pgx values, and arbitrary SQL remain behind composition.
type StateReadCapability interface {
	CapabilityID() string
	StateFamilyIDs() []string
	FamilyCounts(context.Context, []string) (map[string]int64, error)
	ValidateFamilyState(context.Context, string) error
}

// StateWriteCapability is transaction-scoped. Profile owners extend this
// interface with owner-local typed mutation methods at the composition edge.
type StateWriteCapability interface {
	StateReadCapability
	IsStateWriteCapability() bool
}

// StateTransaction belongs to coordination. Callback views must have concrete
// method sets limited to their read or owner-local write authority.
type StateTransaction interface {
	StateReadCapability
	ReadCapability() StateReadCapability
	WriteCapability() StateWriteCapability
	// CapabilityError retains forbidden access even when a callback ignores it.
	CapabilityError() error
	StateMetadata(context.Context, string) (*StateMetadata, error)
	MigrationLedger(context.Context, string, string) ([]MigrationLedgerEntry, error)
	InsertStateMetadata(context.Context, StateMetadata) error
	UpdateStateMetadata(context.Context, StateMetadata, int, string, time.Time) error
	InsertMigrationLedger(context.Context, MigrationLedgerEntry) error
	Commit(context.Context) (CommitOutcome, error)
	Rollback(context.Context) (CommitOutcome, error)
}

type ProfileSession interface {
	Snapshot(context.Context, string, string, []string) (StateSnapshot, error)
	Begin(context.Context, string, []string) (StateTransaction, error)
}

type StateStore interface {
	WithProfileSession(context.Context, string, time.Duration, func(ProfileSession) error) error
}

type InitializationContext struct {
	ProfileID                               string
	MigrationLineageID                      string
	TargetStateVersion                      int
	AuthoritativeStateFamilyIDs             []string
	InitializationDefinitionSHA256          string
	InitializationAlgorithmID               string
	InitializationAlgorithmDefinitionSHA256 string
	DeadlineMonotonicNS                     int64
	ScopedWriteCapabilityID                 string
}

type InitializationResult struct {
	SchemaID string
	Status   string
}

type MigrationContext struct {
	ProfileID                 string
	MigrationLineageID        string
	MigrationID               string
	FromStateVersion          int
	ToStateVersion            int
	MigrationDefinitionSHA256 string
	StateFamilyIDs            []string
	MetadataBeforeSHA256      string
	DeadlineMonotonicNS       int64
	StateAccessCapabilityID   string
}

type MigrationApplyResult struct {
	SchemaID string
	Status   string
}

type MigrationValidationContext = MigrationContext

type FinalStateValidationContext struct {
	ProfileID                       string
	MigrationLineageID              string
	StateVersion                    int
	StateFamilyIDs                  []string
	StateMetadataSHA256             string
	StatePresenceManifestSHA256     string
	ReadOnlyStateAccessCapabilityID string
	DeadlineMonotonicNS             int64
}

type StateValidationResult struct {
	SchemaID string           `json:"schema_id"`
	Status   string           `json:"status"`
	Findings []StartupFinding `json:"findings"`
}

func ValidFinalStateValidationResult() StateValidationResult {
	return StateValidationResult{
		SchemaID: "cartulary.extension_final_state_validation_result.v1",
		Status:   "valid",
		Findings: []StartupFinding{},
	}
}

type StateInitializer func(context.Context, InitializationContext, StateWriteCapability) (InitializationResult, error)
type StateMigration func(context.Context, MigrationContext, StateWriteCapability) (MigrationApplyResult, error)
type PendingStateValidator func(context.Context, MigrationValidationContext, StateReadCapability) (StateValidationResult, error)
type FinalStateValidator func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error)

type StateRuntimeOptions struct {
	Coordinator        *Coordinator
	Store              StateStore
	Initializers       map[string]StateInitializer
	Migrations         map[string]StateMigration
	PendingValidators  map[string]PendingStateValidator
	FinalValidators    map[string]FinalStateValidator
	Now                func() time.Time
	MonotonicNowNS     func() int64
	LockTimeout        time.Duration
	StepTimeout        time.Duration
	ProfileTimeout     time.Duration
	ValidationTimeout  time.Duration
	CancellationGrace  time.Duration
	FatalIntegritySink func(error)
}

type StateRuntime struct {
	coordinator        *Coordinator
	store              StateStore
	initializers       map[string]StateInitializer
	migrations         map[string]StateMigration
	pendingValidators  map[string]PendingStateValidator
	finalValidators    map[string]FinalStateValidator
	now                func() time.Time
	monotonicNowNS     func() int64
	lockTimeout        time.Duration
	stepTimeout        time.Duration
	profileTimeout     time.Duration
	validationTimeout  time.Duration
	cancellationGrace  time.Duration
	fatalIntegritySink func(error)
}

func NewStateRuntime(options StateRuntimeOptions) (*StateRuntime, error) {
	if options.Coordinator == nil || options.Coordinator.validateFinding == nil {
		return nil, errors.New("extension state runtime requires admitted validation contracts")
	}
	if options.Store == nil {
		return nil, errors.New("extension state runtime requires a store")
	}
	if options.LockTimeout <= 0 || options.StepTimeout <= 0 || options.ProfileTimeout <= 0 || options.ValidationTimeout <= 0 {
		return nil, errors.New("extension state runtime requires positive timeout policy")
	}
	if options.CancellationGrace < 0 {
		return nil, errors.New("extension state runtime requires nonnegative cancellation grace")
	}
	if options.FatalIntegritySink == nil {
		return nil, errors.New("extension state runtime requires a fatal-integrity sink")
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	if options.MonotonicNowNS == nil {
		started := time.Now()
		options.MonotonicNowNS = func() int64 { return time.Since(started).Nanoseconds() }
	}
	return &StateRuntime{
		coordinator:        options.Coordinator,
		store:              options.Store,
		initializers:       cloneInitializers(options.Initializers),
		migrations:         cloneMigrations(options.Migrations),
		pendingValidators:  clonePendingValidators(options.PendingValidators),
		finalValidators:    cloneFinalValidators(options.FinalValidators),
		now:                options.Now,
		monotonicNowNS:     options.MonotonicNowNS,
		lockTimeout:        options.LockTimeout,
		stepTimeout:        options.StepTimeout,
		profileTimeout:     options.ProfileTimeout,
		validationTimeout:  options.ValidationTimeout,
		cancellationGrace:  options.CancellationGrace,
		fatalIntegritySink: options.FatalIntegritySink,
	}, nil
}

func cloneInitializers(source map[string]StateInitializer) map[string]StateInitializer {
	result := make(map[string]StateInitializer, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneMigrations(source map[string]StateMigration) map[string]StateMigration {
	result := make(map[string]StateMigration, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func clonePendingValidators(source map[string]PendingStateValidator) map[string]PendingStateValidator {
	result := make(map[string]PendingStateValidator, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneFinalValidators(source map[string]FinalStateValidator) map[string]FinalStateValidator {
	result := make(map[string]FinalStateValidator, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func (r *StateRuntime) AdmitClaims(ctx context.Context, coordinator *Coordinator, resolution ClaimResolution) error {
	if r == nil || coordinator == nil {
		return ErrStateIncomplete
	}
	order, err := coordinator.validateClaimResolution(resolution)
	if err != nil {
		return err
	}
	for _, profileID := range order {
		plan, ok := coordinator.StatePlan(profileID)
		if !ok {
			continue
		}
		if err := r.Admit(ctx, plan); err != nil {
			return fmt.Errorf("%s: %w", profileID, err)
		}
	}
	return nil
}

func (r *StateRuntime) Admit(ctx context.Context, plan StatePlan) (resultErr error) {
	if r == nil || r.store == nil || plan.ProfileID == "" {
		return ErrStateIncomplete
	}
	var snapshot StateSnapshot
	defer func() { resultErr = r.stateDiagnostic(plan, snapshot, resultErr) }()
	families, err := stateFamilies(plan)
	if err != nil {
		return err
	}
	finalValidator := r.finalValidators[plan.FinalValidationAlgorithmID]
	if finalValidator == nil {
		return ErrStateValidationFailed
	}
	if err := r.preflightPackagedPlan(plan); err != nil {
		return startupFailure(err, registryArtifactFinding("state."+plan.ProfileID))
	}
	return r.store.WithProfileSession(ctx, plan.ProfileID, r.lockTimeout, func(session ProfileSession) error {
		profileDeadline := r.newStateBudget(r.profileTimeout, stateBudgetProfile, nil)
		if err := r.boundary(ctx, profileDeadline); err != nil {
			return err
		}
		snapshotCtx, cancel, err := r.invocationContext(ctx, profileDeadline)
		if err != nil {
			return err
		}
		snapshot, err = session.Snapshot(snapshotCtx, plan.ProfileID, plan.MigrationLineageID, families)
		err = r.stateStorageFailure(snapshotCtx, profileDeadline, err)
		cancel()
		if err != nil {
			return err
		}
		if err := r.boundary(ctx, profileDeadline); err != nil {
			return err
		}
		present, err := authoritativeStatePresent(snapshot.FamilyCounts, families)
		if err != nil {
			return err
		}
		decision, err := DecideStatePresence(plan.EmptyStatePolicy, snapshot.Metadata != nil, present)
		if err != nil {
			return err
		}
		if decision == StateInitialize {
			return r.initialize(ctx, session, plan, families, profileDeadline, finalValidator)
		}
		if err := preflightStoredState(plan, snapshot); err != nil {
			return err
		}
		definitions, err := requiredMigrations(plan, snapshot.Metadata.StateVersion)
		if err != nil {
			return err
		}
		for _, definition := range definitions {
			if err := r.boundary(ctx, profileDeadline); err != nil {
				return err
			}
			if err := r.migrateStep(ctx, session, plan, families, definition, profileDeadline); err != nil {
				return err
			}
		}
		return r.validateFinal(ctx, session, plan, families, profileDeadline, finalValidator)
	})
}

func (r *StateRuntime) preflightPackagedPlan(plan StatePlan) error {
	if plan.InitializationKind != "empty" && plan.InitializationKind != "algorithm" {
		return ErrStateMigrationUnavailable
	}
	if plan.InitializationKind == "empty" {
		if plan.EmptyStatePolicy == "forbidden" || plan.InitializationAlgorithmID != "" || plan.InitializationAlgorithmDefinitionSHA256 != "" {
			return ErrStateMigrationUnavailable
		}
	} else {
		if plan.InitializationAlgorithmID == "" || !validSHA256(plan.InitializationAlgorithmDefinitionSHA256) || r.initializers[plan.InitializationAlgorithmID] == nil {
			return ErrStateMigrationUnavailable
		}
	}
	if !validSHA256(plan.InitializationDefinitionSHA256) || !validSHA256(plan.StatePresenceManifestSHA256) || !validSHA256(plan.PhysicalStateBindingSHA256) || !validSHA256(plan.ImplementationBindingSHA256) {
		return ErrStateMigrationUnavailable
	}
	previousFrom := 0
	seenIDs := map[string]struct{}{}
	for index, definition := range plan.MigrationDefinitions {
		if definition.MigrationLineageID != plan.MigrationLineageID ||
			definition.ToVersion != definition.FromVersion+1 ||
			definition.FromVersion < 1 ||
			definition.ToVersion > plan.CurrentStateVersion ||
			!validSHA256(definition.DefinitionSHA256) ||
			definition.ApplyAlgorithmID == "" ||
			definition.ValidationAlgorithmID == "" ||
			definition.ImplementationBindingProfileID != plan.ProfileID ||
			definition.ImplementationBindingSHA256 != plan.ImplementationBindingSHA256 ||
			r.migrations[definition.ApplyAlgorithmID] == nil ||
			r.pendingValidators[definition.ValidationAlgorithmID] == nil {
			return ErrStateMigrationUnavailable
		}
		if index > 0 && definition.FromVersion <= previousFrom {
			return ErrStateMigrationUnavailable
		}
		previousFrom = definition.FromVersion
		if _, duplicate := seenIDs[definition.MigrationID]; definition.MigrationID == "" || duplicate {
			return ErrStateMigrationUnavailable
		}
		seenIDs[definition.MigrationID] = struct{}{}
	}
	previousFrom = 0
	seenLedgerIDs := map[string]struct{}{}
	ledgerByID := make(map[string]MigrationLedgerDefinition, len(plan.MigrationLedgerDefinitions))
	for index, definition := range plan.MigrationLedgerDefinitions {
		if definition.MigrationLineageID != plan.MigrationLineageID ||
			definition.ToVersion != definition.FromVersion+1 ||
			definition.FromVersion < 1 || definition.ToVersion > plan.CurrentStateVersion ||
			!validSHA256(definition.DefinitionSHA256) {
			return ErrStateMigrationUnavailable
		}
		if index > 0 && definition.FromVersion <= previousFrom {
			return ErrStateMigrationUnavailable
		}
		previousFrom = definition.FromVersion
		if _, duplicate := seenLedgerIDs[definition.MigrationID]; definition.MigrationID == "" || duplicate {
			return ErrStateMigrationUnavailable
		}
		seenLedgerIDs[definition.MigrationID] = struct{}{}
		ledgerByID[definition.MigrationID] = definition
	}
	for _, definition := range plan.MigrationDefinitions {
		ledger, ok := ledgerByID[definition.MigrationID]
		if !ok || ledger.MigrationLineageID != definition.MigrationLineageID ||
			ledger.FromVersion != definition.FromVersion || ledger.ToVersion != definition.ToVersion ||
			ledger.DefinitionSHA256 != definition.DefinitionSHA256 {
			return ErrStateMigrationUnavailable
		}
	}
	return nil
}

func (r *StateRuntime) initialize(ctx context.Context, session ProfileSession, plan StatePlan, families []string, profileDeadline stateBudget, finalValidator FinalStateValidator) error {
	stepDeadline := r.newStateBudget(r.stepTimeout, stateBudgetInitialization, &profileDeadline)
	stepCtx, cancelStep, err := r.invocationContext(ctx, stepDeadline)
	if err != nil {
		return err
	}
	defer cancelStep()
	stepCtx = extensiondeadline.WithCancellationTracking(stepCtx)
	tx, err := session.Begin(stepCtx, plan.ProfileID, families)
	if err != nil {
		return r.stateStorageFailure(stepCtx, stepDeadline, err)
	}
	metadata, err := tx.StateMetadata(stepCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if metadata != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, ErrStateIncomplete)
	}
	if plan.InitializationKind == "algorithm" {
		invocation := InitializationContext{
			ProfileID:                               plan.ProfileID,
			MigrationLineageID:                      plan.MigrationLineageID,
			TargetStateVersion:                      plan.CurrentStateVersion,
			AuthoritativeStateFamilyIDs:             append([]string(nil), families...),
			InitializationDefinitionSHA256:          plan.InitializationDefinitionSHA256,
			InitializationAlgorithmID:               plan.InitializationAlgorithmID,
			InitializationAlgorithmDefinitionSHA256: plan.InitializationAlgorithmDefinitionSHA256,
			DeadlineMonotonicNS:                     stepDeadline.MonotonicNS,
			ScopedWriteCapabilityID:                 tx.CapabilityID(),
		}
		invokeCtx, cancel, boundaryErr := r.invocationContext(stepCtx, stepDeadline)
		if boundaryErr != nil {
			return r.rollbackFailure(stepCtx, stepDeadline, tx, boundaryErr)
		}
		result, invokeErr := invokeStateCallback(r, invokeCtx, stepDeadline, func(callCtx context.Context) (InitializationResult, error) {
			return r.initializers[plan.InitializationAlgorithmID](callCtx, invocation, tx.WriteCapability())
		})
		cancel()
		invokeErr = errors.Join(invokeErr, tx.CapabilityError())
		if invokeErr != nil {
			return r.rollbackFailure(stepCtx, stepDeadline, tx, &initializerFailure{cause: stateMutationInvocationFailure(invokeErr, stepDeadline)})
		}
		if result.SchemaID != "cartulary.extension_state_initialization_result.v1" || result.Status != "ready_to_validate" {
			return r.rollbackFailure(stepCtx, stepDeadline, tx, &initializerFailure{cause: ErrStateValidationFailed})
		}
	}
	if err := r.boundary(stepCtx, stepDeadline); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	counts, err := tx.FamilyCounts(stepCtx, families)
	if err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	present, err := authoritativeStatePresent(counts, families)
	if err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if plan.EmptyStatePolicy == "forbidden" && !present {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, ErrStateIncomplete)
	}
	now := r.now().UTC().Truncate(time.Microsecond)
	pending := StateMetadata{
		ProfileID:          plan.ProfileID,
		MigrationLineageID: plan.MigrationLineageID,
		StateVersion:       plan.CurrentStateVersion,
		MetadataVersion:    1,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := tx.InsertStateMetadata(stepCtx, pending); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	validationDeadline := r.newStateBudget(r.validationTimeout, stateBudgetFinalValidation, &stepDeadline)
	if err := r.invokeFinalValidator(stepCtx, validationDeadline, plan, pending, families, tx, finalValidator); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if err := r.boundary(stepCtx, stepDeadline); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if err := r.commit(stepCtx, tx, stepDeadline); err != nil {
		return err
	}
	return r.verifyInitializationReadback(session, plan, families, pending)
}

func (r *StateRuntime) migrateStep(ctx context.Context, session ProfileSession, plan StatePlan, families []string, definition MigrationDefinition, profileDeadline stateBudget) (resultErr error) {
	stepDeadline := r.newStateBudget(r.stepTimeout, stateBudgetMigration, &profileDeadline)
	if stepDeadline.kind == stateBudgetMigration {
		stepDeadline.migration = &definition
	}
	defer func() {
		var diagnosed *StartupValidationError
		var fatal *FatalStateIntegrityError
		var expired *stateBudgetTimeout
		if resultErr == nil || errors.As(resultErr, &diagnosed) || errors.As(resultErr, &fatal) || errors.As(resultErr, &expired) {
			return
		}
		if errors.Is(resultErr, ErrStateValidationFailed) {
			resultErr = startupFailure(resultErr, r.coordinator.migrationValidationFailure(plan.ProfileID, definition.MigrationID, definition.FromVersion, definition.ToVersion))
		}

	}()
	stepCtx, cancelStep, err := r.invocationContext(ctx, stepDeadline)
	if err != nil {
		return err
	}
	defer cancelStep()
	stepCtx = extensiondeadline.WithCancellationTracking(stepCtx)
	tx, err := session.Begin(stepCtx, plan.ProfileID, families)
	if err != nil {
		return r.stateStorageFailure(stepCtx, stepDeadline, err)
	}
	metadata, err := tx.StateMetadata(stepCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if metadata == nil || metadata.MigrationLineageID != plan.MigrationLineageID || metadata.StateVersion != definition.FromVersion {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, ErrStateVersionUnsupported)
	}
	invocation := MigrationContext{
		ProfileID:                 plan.ProfileID,
		MigrationLineageID:        plan.MigrationLineageID,
		MigrationID:               definition.MigrationID,
		FromStateVersion:          definition.FromVersion,
		ToStateVersion:            definition.ToVersion,
		MigrationDefinitionSHA256: definition.DefinitionSHA256,
		StateFamilyIDs:            append([]string(nil), families...),
		MetadataBeforeSHA256:      stateMetadataSHA256(*metadata),
		DeadlineMonotonicNS:       stepDeadline.MonotonicNS,
		StateAccessCapabilityID:   tx.CapabilityID(),
	}
	invokeCtx, cancel, boundaryErr := r.invocationContext(stepCtx, stepDeadline)
	if boundaryErr != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, boundaryErr)
	}
	applyResult, applyErr := invokeStateCallback(r, invokeCtx, stepDeadline, func(callCtx context.Context) (MigrationApplyResult, error) {
		return r.migrations[definition.ApplyAlgorithmID](callCtx, invocation, tx.WriteCapability())
	})
	cancel()
	applyErr = errors.Join(applyErr, tx.CapabilityError())
	if applyErr != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, stateMutationInvocationFailure(applyErr, stepDeadline))
	}
	if applyResult.SchemaID != "cartulary.extension_migration_apply_result.v1" || applyResult.Status != "ready_to_validate" {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, ErrStateValidationFailed)
	}
	if err := r.boundary(stepCtx, stepDeadline); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	validationCtx, cancelValidation, boundaryErr := r.invocationContext(stepCtx, stepDeadline)
	if boundaryErr != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, boundaryErr)
	}
	validationResult, validationErr := invokeStateCallback(r, validationCtx, stepDeadline, func(callCtx context.Context) (StateValidationResult, error) {
		return r.pendingValidators[definition.ValidationAlgorithmID](callCtx, invocation, tx.ReadCapability())
	})
	defer cancelValidation()
	validationErr = errors.Join(validationErr, tx.CapabilityError())
	if extensiondeadline.Interrupted(validationErr) && errors.Is(validationErr, context.DeadlineExceeded) {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, &stateBudgetTimeout{budget: stepDeadline, cause: validationErr})
	}
	resultErr = r.classifyStateResult(plan.ProfileID, "migration_validation", definition.ValidationAlgorithmID, "cartulary.extension_migration_validation_result.v1", validationResult, validationErr, r.coordinator.migrationValidationFailure(plan.ProfileID, definition.MigrationID, definition.FromVersion, definition.ToVersion))
	var fatal *FatalStateIntegrityError
	if errors.As(resultErr, &fatal) {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, resultErr)
	}
	if boundaryErr := extensiondeadline.Sample(validationCtx, r.invocationPolicy(stepDeadline.Deadline)); boundaryErr != nil {
		if errors.Is(boundaryErr, context.DeadlineExceeded) {
			return r.rollbackFailure(stepCtx, stepDeadline, tx, &stateBudgetTimeout{budget: stepDeadline, cause: boundaryErr})
		}
		resultErr = r.classifyStateResult(plan.ProfileID, "migration_validation", definition.ValidationAlgorithmID, "cartulary.extension_migration_validation_result.v1", StateValidationResult{}, boundaryErr, r.coordinator.migrationValidationFailure(plan.ProfileID, definition.MigrationID, definition.FromVersion, definition.ToVersion))
	}
	if resultErr != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, resultErr)
	}
	if err := r.boundary(stepCtx, stepDeadline); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	now := r.now().UTC().Truncate(time.Microsecond)
	entry := MigrationLedgerEntry{
		ProfileID:                 plan.ProfileID,
		MigrationLineageID:        plan.MigrationLineageID,
		MigrationID:               definition.MigrationID,
		FromStateVersion:          definition.FromVersion,
		ToStateVersion:            definition.ToVersion,
		MigrationDefinitionSHA256: definition.DefinitionSHA256,
		CommittedAt:               now,
		ResultingStateVersion:     definition.ToVersion,
	}
	if err := tx.InsertMigrationLedger(stepCtx, entry); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if err := tx.UpdateStateMetadata(stepCtx, *metadata, definition.ToVersion, definition.MigrationID, now); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if err := r.boundary(stepCtx, stepDeadline); err != nil {
		return r.rollbackFailure(stepCtx, stepDeadline, tx, err)
	}
	if err := r.commit(stepCtx, tx, stepDeadline); err != nil {
		return err
	}
	expectedMetadata := *metadata
	expectedMetadata.StateVersion = definition.ToVersion
	expectedMetadata.LastMigrationID = cloneStateString(&definition.MigrationID)
	expectedMetadata.MetadataVersion++
	expectedMetadata.UpdatedAt = now
	return r.verifyMigrationReadback(session, plan, families, definition, expectedMetadata, entry)
}

func (r *StateRuntime) validateFinal(ctx context.Context, session ProfileSession, plan StatePlan, families []string, profileDeadline stateBudget, validator FinalStateValidator) error {
	validationCtx, cancelValidation, err := r.invocationContext(ctx, profileDeadline)
	if err != nil {
		return err
	}
	defer cancelValidation()
	validationCtx = extensiondeadline.WithCancellationTracking(validationCtx)
	tx, err := session.Begin(validationCtx, plan.ProfileID, families)
	if err != nil {
		return r.stateStorageFailure(validationCtx, profileDeadline, err)
	}
	metadata, err := tx.StateMetadata(validationCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(validationCtx, profileDeadline, tx, err)
	}
	if metadata == nil || metadata.StateVersion != plan.CurrentStateVersion || metadata.MigrationLineageID != plan.MigrationLineageID {
		return r.rollbackFailure(validationCtx, profileDeadline, tx, ErrStateVersionUnsupported)
	}
	validationDeadline := r.newStateBudget(r.validationTimeout, stateBudgetFinalValidation, &profileDeadline)
	validationErr := r.invokeFinalValidator(validationCtx, validationDeadline, plan, *metadata, families, tx, validator)
	if validationErr != nil {
		return r.rollbackFailure(validationCtx, profileDeadline, tx, validationErr)
	}
	return r.rollbackFailure(validationCtx, profileDeadline, tx, nil)
}

func (r *StateRuntime) invokeFinalValidator(ctx context.Context, deadlinePolicy stateBudget, plan StatePlan, metadata StateMetadata, families []string, tx StateTransaction, validator FinalStateValidator) error {
	invokeCtx, cancel, err := r.invocationContext(ctx, deadlinePolicy)
	if err != nil {
		return err
	}
	reader := tx.ReadCapability()
	result, validationErr := invokeStateCallback(r, invokeCtx, deadlinePolicy, func(callCtx context.Context) (StateValidationResult, error) {
		return validator(callCtx, FinalStateValidationContext{
			ProfileID:                       plan.ProfileID,
			MigrationLineageID:              plan.MigrationLineageID,
			StateVersion:                    metadata.StateVersion,
			StateFamilyIDs:                  append([]string(nil), families...),
			StateMetadataSHA256:             stateMetadataSHA256(metadata),
			StatePresenceManifestSHA256:     plan.StatePresenceManifestSHA256,
			ReadOnlyStateAccessCapabilityID: reader.CapabilityID(),
			DeadlineMonotonicNS:             deadlinePolicy.MonotonicNS,
		}, reader)
	})
	defer cancel()
	validationErr = errors.Join(validationErr, tx.CapabilityError())

	failure := r.coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID)
	resultErr := r.classifyStateResult(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID, "cartulary.extension_final_state_validation_result.v1", result, validationErr, failure)
	var fatal *FatalStateIntegrityError
	if errors.As(resultErr, &fatal) {
		return resultErr
	}
	if boundaryErr := extensiondeadline.Sample(invokeCtx, r.invocationPolicy(deadlinePolicy.Deadline)); boundaryErr != nil {
		if errors.Is(boundaryErr, context.DeadlineExceeded) {
			return &stateBudgetTimeout{budget: deadlinePolicy, cause: boundaryErr}
		}
		return r.classifyStateResult(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID, "cartulary.extension_final_state_validation_result.v1", StateValidationResult{}, boundaryErr, failure)
	}
	return resultErr
}

func (r *StateRuntime) commit(ctx context.Context, tx StateTransaction, bound stateBudget) error {
	completion, lifetimeErr := extensiondeadline.Complete(ctx, r.invocationPolicy(bound.Deadline), tx.Commit)
	if lifetimeErr != nil {
		return &FatalStateIntegrityError{Cause: lifetimeErr}
	}
	outcome, commitErr := completion.Value, completion.Err
	switch outcome {
	case CommitCommitted:
		return nil
	case CommitAbsent:
		if completion.Interruption != nil {
			if errors.Is(completion.Interruption, context.DeadlineExceeded) {
				return &stateBudgetTimeout{budget: bound, cause: completion.Interruption}
			}
			return completion.Interruption
		}
		if commitErr != nil {
			return fmt.Errorf("%w: %v", ErrStateCommitAbsent, commitErr)
		}
		return ErrStateCommitAbsent
	case CommitIndeterminate:
		return r.fatal(fmt.Errorf("%w: %v", ErrStateCommitIndeterminate, commitErr))
	default:
		return r.fatal(fmt.Errorf("%w: unknown commit outcome %q", ErrStateCommitIndeterminate, outcome))
	}
}

func (r *StateRuntime) rollbackFailure(ctx context.Context, bound stateBudget, tx StateTransaction, cause error) error {
	completion, lifetimeErr := extensiondeadline.Complete(ctx, r.invocationPolicy(bound.Deadline), func(callCtx context.Context) (CommitOutcome, error) {
		return tx.Rollback(context.WithoutCancel(callCtx))
	})
	if lifetimeErr != nil {
		return &FatalStateIntegrityError{Cause: lifetimeErr}
	}
	outcome, rollbackErr := completion.Value, completion.Err
	if outcome != CommitAbsent {
		return r.fatal(fmt.Errorf("%w: rollback after %v: %v", ErrStateCommitIndeterminate, cause, rollbackErr))
	}
	if completion.Interruption != nil {
		var selected *stateBudgetTimeout
		if errors.As(cause, &selected) || extensiondeadline.Interrupted(cause) {
			return cause
		}
		if errors.Is(completion.Interruption, context.DeadlineExceeded) {
			return errors.Join(&stateBudgetTimeout{budget: bound, cause: completion.Interruption}, cause)
		}
		if cause == nil {
			return completion.Interruption
		}
	}
	if rollbackErr != nil {
		return fmt.Errorf("%w: rollback: %v", cause, rollbackErr)
	}
	return cause
}

func (r *StateRuntime) verifyInitializationReadback(session ProfileSession, plan StatePlan, families []string, expected StateMetadata) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.validationTimeout)
	defer cancel()
	snapshot, err := session.Snapshot(ctx, plan.ProfileID, plan.MigrationLineageID, families)
	if err != nil || snapshot.Metadata == nil ||
		stateMetadataSHA256(*snapshot.Metadata) != stateMetadataSHA256(expected) ||
		len(snapshot.Ledger) != 0 {
		return r.fatal(fmt.Errorf("%w: initialization: %v", ErrStateReadbackMismatch, err))
	}
	return nil
}

func (r *StateRuntime) verifyMigrationReadback(session ProfileSession, plan StatePlan, families []string, definition MigrationDefinition, expectedMetadata StateMetadata, expectedLedger MigrationLedgerEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.validationTimeout)
	defer cancel()
	snapshot, err := session.Snapshot(ctx, plan.ProfileID, plan.MigrationLineageID, families)
	if err != nil || snapshot.Metadata == nil ||
		stateMetadataSHA256(*snapshot.Metadata) != stateMetadataSHA256(expectedMetadata) {
		return r.fatal(fmt.Errorf("%w: migration metadata %s: %v", ErrStateReadbackMismatch, definition.MigrationID, err))
	}
	for _, entry := range snapshot.Ledger {
		if migrationLedgerEntryEqual(entry, expectedLedger) {
			return nil
		}
	}
	return r.fatal(fmt.Errorf("%w: committed migration ledger row %s missing", ErrStateReadbackMismatch, definition.MigrationID))
}

func (r *StateRuntime) boundary(ctx context.Context, bound stateBudget) error {
	err := extensiondeadline.Sample(ctx, r.invocationPolicy(bound.Deadline))
	if errors.Is(err, context.DeadlineExceeded) {
		return &stateBudgetTimeout{budget: bound, cause: err}
	}
	return err
}

func (r *StateRuntime) invocationContext(ctx context.Context, deadlinePolicy stateBudget) (context.Context, context.CancelFunc, error) {
	now := r.monotonicNowNS()
	if deadlinePolicy.Expired(now) {
		return nil, func() {}, &stateBudgetTimeout{budget: deadlinePolicy, cause: context.DeadlineExceeded}
	}
	remaining := deadlinePolicy.MonotonicNS - now
	invokeCtx, cancel := context.WithTimeout(ctx, time.Duration(remaining))
	return invokeCtx, cancel, nil
}

func invokeStateCallback[T any](r *StateRuntime, ctx context.Context, bound stateBudget, invoke func(context.Context) (T, error)) (T, error) {
	result, err := extensiondeadline.Invoke(ctx, r.invocationPolicy(bound.Deadline), invoke)
	if errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		return result, &FatalStateIntegrityError{Cause: err}
	}
	return result, err
}

func stateMutationInvocationFailure(err error, budget stateBudget) error {
	var fatal *FatalStateIntegrityError
	if errors.As(err, &fatal) {
		return err
	}
	if extensiondeadline.Interrupted(err) && errors.Is(err, context.DeadlineExceeded) {
		return &stateBudgetTimeout{budget: budget, cause: err}
	}
	return errors.Join(ErrStateValidationFailed, err)
}

func (r *StateRuntime) invocationPolicy(bound extensiondeadline.Deadline) extensiondeadline.InvocationPolicy {
	return extensiondeadline.InvocationPolicy{Grace: r.cancellationGrace, Exhausted: r.fatalIntegritySink, Deadline: &bound, MonotonicNowNS: r.monotonicNowNS}
}

func (r *StateRuntime) fatal(cause error) error {
	r.fatalIntegritySink(cause)
	return &FatalStateIntegrityError{Cause: cause}
}

type FatalStateIntegrityError struct {
	Cause error
}

func (e *FatalStateIntegrityError) Error() string {
	return fmt.Sprintf("fatal extension state integrity failure: %v", e.Cause)
}

func (e *FatalStateIntegrityError) Unwrap() error {
	return e.Cause
}

func (e *FatalStateIntegrityError) FatalReasonCode() string {
	if e != nil && errors.Is(e.Cause, extensiondeadline.ErrGraceExhausted) {
		return "operation_cancellation_grace_exhausted"
	}
	if e != nil && errors.Is(e.Cause, ErrStateReadbackMismatch) {
		return "migration_ledger_state_mismatch"
	}
	return "indeterminate_database_commit"
}

func preflightStoredState(plan StatePlan, snapshot StateSnapshot) error {
	metadata := snapshot.Metadata
	if metadata == nil {
		return ErrStateIncomplete
	}
	if metadata.ProfileID != plan.ProfileID ||
		metadata.MetadataVersion < 1 ||
		metadata.StateVersion < 1 ||
		metadata.CreatedAt.IsZero() ||
		metadata.UpdatedAt.Before(metadata.CreatedAt) {
		return ErrStateVersionUnsupported
	}
	if metadata.MigrationLineageID != plan.MigrationLineageID {
		return ErrStateLineageMismatch
	}
	if metadata.StateVersion > plan.CurrentStateVersion || metadata.StateVersion < plan.MinimumMigratableStateVersion {
		return ErrStateVersionUnsupported
	}
	definitionsByID := make(map[string]MigrationLedgerDefinition, len(plan.MigrationLedgerDefinitions))
	for _, definition := range plan.MigrationLedgerDefinitions {
		definitionsByID[definition.MigrationID] = definition
	}
	var latest *MigrationLedgerEntry
	var previous *MigrationLedgerEntry
	for index := range snapshot.Ledger {
		entry := snapshot.Ledger[index]
		definition, ok := definitionsByID[entry.MigrationID]
		if !ok || definition.FromVersion != entry.FromStateVersion || definition.ToVersion != entry.ToStateVersion {
			return ErrStateMigrationUnavailable
		}
		if definition.DefinitionSHA256 != entry.MigrationDefinitionSHA256 {
			return ErrStateMigrationDefinitionChanged
		}
		if entry.ProfileID != plan.ProfileID ||
			entry.MigrationLineageID != plan.MigrationLineageID ||
			entry.ResultingStateVersion != entry.ToStateVersion ||
			entry.ToStateVersion > metadata.StateVersion ||
			entry.CommittedAt.IsZero() {
			return ErrStateVersionUnsupported
		}
		if previous != nil &&
			(entry.FromStateVersion != previous.ToStateVersion ||
				entry.CommittedAt.Before(previous.CommittedAt)) {
			return ErrStateVersionUnsupported
		}
		copy := entry
		previous = &copy
		if latest == nil || entry.ToStateVersion > latest.ToStateVersion {
			latest = &copy
		}
	}
	if metadata.MetadataVersion != 1+len(snapshot.Ledger) {
		return ErrStateVersionUnsupported
	}
	if latest == nil {
		if metadata.LastMigrationID != nil {
			return ErrStateVersionUnsupported
		}
	} else {
		if latest.ToStateVersion != metadata.StateVersion ||
			metadata.LastMigrationID == nil ||
			*metadata.LastMigrationID != latest.MigrationID {
			return ErrStateVersionUnsupported
		}
	}
	_, err := requiredMigrations(plan, metadata.StateVersion)
	return err
}

func requiredMigrations(plan StatePlan, storedVersion int) ([]MigrationDefinition, error) {
	byFrom := make(map[int]MigrationDefinition, len(plan.MigrationDefinitions))
	for _, definition := range plan.MigrationDefinitions {
		if _, duplicate := byFrom[definition.FromVersion]; duplicate {
			return nil, ErrStateMigrationUnavailable
		}
		byFrom[definition.FromVersion] = definition
	}
	result := make([]MigrationDefinition, 0, plan.CurrentStateVersion-storedVersion)
	for version := storedVersion; version < plan.CurrentStateVersion; version++ {
		definition, ok := byFrom[version]
		if !ok || definition.ToVersion != version+1 {
			return nil, ErrStateMigrationUnavailable
		}
		result = append(result, definition)
	}
	return result, nil
}

func stateFamilies(plan StatePlan) ([]string, error) {
	families := append([]string(nil), plan.DatabaseFamilyIDs...)
	families = append(families, plan.ObjectReferenceFamilyIDs...)
	sort.Strings(families)
	if len(families) == 0 {
		return nil, ErrStateIncomplete
	}
	for index, familyID := range families {
		if familyID == "" || (index > 0 && families[index-1] == familyID) {
			return nil, ErrStateIncomplete
		}
	}
	return families, nil
}

func authoritativeStatePresent(counts map[string]int64, familyIDs []string) (bool, error) {
	if len(counts) != len(familyIDs) {
		return false, ErrStateIncomplete
	}
	present := false
	for _, familyID := range familyIDs {
		count, ok := counts[familyID]
		if !ok || count < 0 {
			return false, ErrStateIncomplete
		}
		present = present || count > 0
	}
	return present, nil
}

func stateMetadataSHA256(metadata StateMetadata) string {
	payload := struct {
		SchemaID           string  `json:"schema_id"`
		ProfileID          string  `json:"profile_id"`
		MigrationLineageID string  `json:"migration_lineage_id"`
		StateVersion       int     `json:"state_version"`
		LastMigrationID    *string `json:"last_migration_id"`
		MetadataVersion    int     `json:"metadata_version"`
	}{
		SchemaID:           "cartulary.extension_state_metadata.v1",
		ProfileID:          metadata.ProfileID,
		MigrationLineageID: metadata.MigrationLineageID,
		StateVersion:       metadata.StateVersion,
		LastMigrationID:    metadata.LastMigrationID,
		MetadataVersion:    metadata.MetadataVersion,
	}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(append(encoded, '\n'))
	return hex.EncodeToString(sum[:])
}

func migrationLedgerEntryEqual(left, right MigrationLedgerEntry) bool {
	return left.ProfileID == right.ProfileID &&
		left.MigrationLineageID == right.MigrationLineageID &&
		left.MigrationID == right.MigrationID &&
		left.FromStateVersion == right.FromStateVersion &&
		left.ToStateVersion == right.ToStateVersion &&
		left.MigrationDefinitionSHA256 == right.MigrationDefinitionSHA256 &&
		left.CommittedAt.Equal(right.CommittedAt) &&
		left.ResultingStateVersion == right.ResultingStateVersion
}

func cloneStateString(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}

func durationSeconds(value time.Duration) int64 {
	seconds := int64(value / time.Second)
	if value%time.Second != 0 {
		seconds++
	}
	return seconds
}
