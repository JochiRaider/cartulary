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
}

// StateWriteCapability is transaction-scoped. Profile owners extend this
// interface with owner-local typed mutation methods at the composition edge.
type StateWriteCapability interface {
	StateReadCapability
	IsStateWriteCapability() bool
}

type StateTransaction interface {
	StateWriteCapability
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

type StateFinding struct {
	Code string
	Path string
}

type StateValidationResult struct {
	SchemaID string
	Status   string
	Findings []StateFinding
}

func ValidMigrationValidationResult() StateValidationResult {
	return StateValidationResult{
		SchemaID: "cartulary.extension_migration_validation_result.v1",
		Status:   "valid",
		Findings: []StateFinding{},
	}
}

func ValidFinalStateValidationResult() StateValidationResult {
	return StateValidationResult{
		SchemaID: "cartulary.extension_final_state_validation_result.v1",
		Status:   "valid",
		Findings: []StateFinding{},
	}
}

type StateInitializer func(context.Context, InitializationContext, StateWriteCapability) (InitializationResult, error)
type StateMigration func(context.Context, MigrationContext, StateWriteCapability) (MigrationApplyResult, error)
type PendingStateValidator func(context.Context, MigrationValidationContext, StateReadCapability) (StateValidationResult, error)
type FinalStateValidator func(context.Context, FinalStateValidationContext, StateReadCapability) (StateValidationResult, error)

type StateRuntimeOptions struct {
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
	FatalIntegritySink func(error)
}

type StateRuntime struct {
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
	fatalIntegritySink func(error)
}

func NewStateRuntime(options StateRuntimeOptions) (*StateRuntime, error) {
	if options.Store == nil {
		return nil, errors.New("extension state runtime requires a store")
	}
	if options.LockTimeout <= 0 || options.StepTimeout <= 0 || options.ProfileTimeout <= 0 || options.ValidationTimeout <= 0 {
		return nil, errors.New("extension state runtime requires positive timeout policy")
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

func (r *StateRuntime) AdmitClaims(ctx context.Context, coordinator *Coordinator, claims ResolvedClaimSet) error {
	if r == nil || coordinator == nil {
		return ErrStateIncomplete
	}
	for _, profileID := range claims.ProfileIDs() {
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

func (r *StateRuntime) Admit(ctx context.Context, plan StatePlan) error {
	if r == nil || r.store == nil || plan.ProfileID == "" {
		return ErrStateIncomplete
	}
	families, err := stateFamilies(plan)
	if err != nil {
		return err
	}
	finalValidator := r.finalValidators[plan.FinalValidationAlgorithmID]
	if finalValidator == nil {
		return ErrStateValidationFailed
	}
	if err := r.preflightPackagedPlan(plan); err != nil {
		return err
	}
	profileDeadline := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(r.profileTimeout), nil)
	return r.store.WithProfileSession(ctx, plan.ProfileID, r.lockTimeout, func(session ProfileSession) error {
		if err := r.boundary(ctx, profileDeadline); err != nil {
			return err
		}
		snapshotCtx, cancel, err := r.invocationContext(ctx, profileDeadline)
		if err != nil {
			return err
		}
		snapshot, err := session.Snapshot(snapshotCtx, plan.ProfileID, plan.MigrationLineageID, families)
		cancel()
		if err != nil {
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
	return nil
}

func (r *StateRuntime) initialize(ctx context.Context, session ProfileSession, plan StatePlan, families []string, profileDeadline extensiondeadline.Deadline, finalValidator FinalStateValidator) error {
	stepDeadline := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(r.stepTimeout), &profileDeadline)
	stepCtx, cancelStep, err := r.invocationContext(ctx, stepDeadline)
	if err != nil {
		return err
	}
	defer cancelStep()
	tx, err := session.Begin(stepCtx, plan.ProfileID, families)
	if err != nil {
		return err
	}
	metadata, err := tx.StateMetadata(stepCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if metadata != nil {
		return r.rollbackFailure(ctx, tx, ErrStateIncomplete)
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
		invokeCtx, cancel, boundaryErr := r.invocationContext(ctx, stepDeadline)
		if boundaryErr != nil {
			return r.rollbackFailure(ctx, tx, boundaryErr)
		}
		result, invokeErr := r.initializers[plan.InitializationAlgorithmID](invokeCtx, invocation, tx)
		cancel()
		if invokeErr != nil || result.SchemaID != "cartulary.extension_state_initialization_result.v1" || result.Status != "ready_to_validate" {
			return r.rollbackFailure(ctx, tx, ErrStateValidationFailed)
		}
	}
	if err := r.boundary(ctx, stepDeadline); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	counts, err := tx.FamilyCounts(stepCtx, families)
	if err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	present, err := authoritativeStatePresent(counts, families)
	if err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if plan.EmptyStatePolicy == "forbidden" && !present {
		return r.rollbackFailure(ctx, tx, ErrStateIncomplete)
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
		return r.rollbackFailure(ctx, tx, err)
	}
	validationDeadline := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(r.validationTimeout), &stepDeadline)
	if err := r.invokeFinalValidator(ctx, validationDeadline, plan, pending, families, tx, finalValidator); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if err := r.boundary(ctx, stepDeadline); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if err := r.commit(stepCtx, tx); err != nil {
		return err
	}
	return r.verifyInitializationReadback(session, plan, families, pending)
}

func (r *StateRuntime) migrateStep(ctx context.Context, session ProfileSession, plan StatePlan, families []string, definition MigrationDefinition, profileDeadline extensiondeadline.Deadline) error {
	stepDeadline := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(r.stepTimeout), &profileDeadline)
	stepCtx, cancelStep, err := r.invocationContext(ctx, stepDeadline)
	if err != nil {
		return err
	}
	defer cancelStep()
	tx, err := session.Begin(stepCtx, plan.ProfileID, families)
	if err != nil {
		return err
	}
	metadata, err := tx.StateMetadata(stepCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if metadata == nil || metadata.MigrationLineageID != plan.MigrationLineageID || metadata.StateVersion != definition.FromVersion {
		return r.rollbackFailure(ctx, tx, ErrStateVersionUnsupported)
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
	invokeCtx, cancel, boundaryErr := r.invocationContext(ctx, stepDeadline)
	if boundaryErr != nil {
		return r.rollbackFailure(ctx, tx, boundaryErr)
	}
	applyResult, applyErr := r.migrations[definition.ApplyAlgorithmID](invokeCtx, invocation, tx)
	cancel()
	if applyErr != nil || applyResult.SchemaID != "cartulary.extension_migration_apply_result.v1" || applyResult.Status != "ready_to_validate" {
		return r.rollbackFailure(ctx, tx, ErrStateValidationFailed)
	}
	if err := r.boundary(ctx, stepDeadline); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	validationCtx, cancelValidation, boundaryErr := r.invocationContext(ctx, stepDeadline)
	if boundaryErr != nil {
		return r.rollbackFailure(ctx, tx, boundaryErr)
	}
	validationResult, validationErr := r.pendingValidators[definition.ValidationAlgorithmID](validationCtx, invocation, tx)
	cancelValidation()
	if validationErr != nil || validationResult.Status != "valid" || !validStateValidationResult(validationResult, "cartulary.extension_migration_validation_result.v1") {
		return r.rollbackFailure(ctx, tx, ErrStateValidationFailed)
	}
	if err := r.boundary(ctx, stepDeadline); err != nil {
		return r.rollbackFailure(ctx, tx, err)
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
		return r.rollbackFailure(ctx, tx, err)
	}
	if err := tx.UpdateStateMetadata(stepCtx, *metadata, definition.ToVersion, definition.MigrationID, now); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if err := r.boundary(ctx, stepDeadline); err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if err := r.commit(stepCtx, tx); err != nil {
		return err
	}
	expectedMetadata := *metadata
	expectedMetadata.StateVersion = definition.ToVersion
	expectedMetadata.LastMigrationID = cloneStateString(&definition.MigrationID)
	expectedMetadata.MetadataVersion++
	expectedMetadata.UpdatedAt = now
	return r.verifyMigrationReadback(session, plan, families, definition, expectedMetadata, entry)
}

func (r *StateRuntime) validateFinal(ctx context.Context, session ProfileSession, plan StatePlan, families []string, profileDeadline extensiondeadline.Deadline, validator FinalStateValidator) error {
	validationDeadline := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(r.validationTimeout), &profileDeadline)
	validationCtx, cancelValidation, err := r.invocationContext(ctx, validationDeadline)
	if err != nil {
		return err
	}
	defer cancelValidation()
	tx, err := session.Begin(validationCtx, plan.ProfileID, families)
	if err != nil {
		return err
	}
	metadata, err := tx.StateMetadata(validationCtx, plan.ProfileID)
	if err != nil {
		return r.rollbackFailure(ctx, tx, err)
	}
	if metadata == nil || metadata.StateVersion != plan.CurrentStateVersion || metadata.MigrationLineageID != plan.MigrationLineageID {
		return r.rollbackFailure(ctx, tx, ErrStateVersionUnsupported)
	}
	validationErr := r.invokeFinalValidator(ctx, validationDeadline, plan, *metadata, families, tx, validator)
	if validationErr != nil {
		return r.rollbackFailure(ctx, tx, validationErr)
	}
	outcome, rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
	if outcome == CommitIndeterminate {
		return r.fatal(fmt.Errorf("final validation rollback: %w", rollbackErr))
	}
	return rollbackErr
}

func (r *StateRuntime) invokeFinalValidator(ctx context.Context, deadlinePolicy extensiondeadline.Deadline, plan StatePlan, metadata StateMetadata, families []string, reader StateReadCapability, validator FinalStateValidator) error {
	invokeCtx, cancel, err := r.invocationContext(ctx, deadlinePolicy)
	if err != nil {
		return err
	}
	result, validationErr := validator(invokeCtx, FinalStateValidationContext{
		ProfileID:                       plan.ProfileID,
		MigrationLineageID:              plan.MigrationLineageID,
		StateVersion:                    metadata.StateVersion,
		StateFamilyIDs:                  append([]string(nil), families...),
		StateMetadataSHA256:             stateMetadataSHA256(metadata),
		StatePresenceManifestSHA256:     plan.StatePresenceManifestSHA256,
		ReadOnlyStateAccessCapabilityID: reader.CapabilityID(),
		DeadlineMonotonicNS:             deadlinePolicy.MonotonicNS,
	}, reader)
	cancel()
	if validationErr != nil || result.Status != "valid" || !validStateValidationResult(result, "cartulary.extension_final_state_validation_result.v1") {
		return ErrStateValidationFailed
	}
	return r.boundary(ctx, deadlinePolicy)
}

func (r *StateRuntime) commit(ctx context.Context, tx StateTransaction) error {
	outcome, commitErr := tx.Commit(ctx)
	switch outcome {
	case CommitCommitted:
		return nil
	case CommitAbsent:
		if boundaryErr := r.contextOutcome(ctx); boundaryErr != nil {
			return boundaryErr
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

func (r *StateRuntime) rollbackFailure(ctx context.Context, tx StateTransaction, cause error) error {
	outcome, rollbackErr := tx.Rollback(context.WithoutCancel(ctx))
	if outcome == CommitIndeterminate {
		return r.fatal(fmt.Errorf("%w: rollback after %v: %v", ErrStateCommitIndeterminate, cause, rollbackErr))
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

func (r *StateRuntime) boundary(ctx context.Context, deadlinePolicy extensiondeadline.Deadline) error {
	if deadlinePolicy.Expired(r.monotonicNowNS()) {
		return ErrStateMigrationTimeout
	}
	return r.contextOutcome(ctx)
}

func (r *StateRuntime) contextOutcome(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (r *StateRuntime) invocationContext(ctx context.Context, deadlinePolicy extensiondeadline.Deadline) (context.Context, context.CancelFunc, error) {
	now := r.monotonicNowNS()
	if deadlinePolicy.Expired(now) {
		return nil, func() {}, ErrStateMigrationTimeout
	}
	remaining := deadlinePolicy.MonotonicNS - now
	invokeCtx, cancel := context.WithTimeout(ctx, time.Duration(remaining))
	return invokeCtx, cancel, nil
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
	definitionsByID := make(map[string]MigrationDefinition, len(plan.MigrationDefinitions))
	for _, definition := range plan.MigrationDefinitions {
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

func validStateValidationResult(result StateValidationResult, schemaID string) bool {
	if result.SchemaID != schemaID || len(result.Findings) > 256 {
		return false
	}
	switch result.Status {
	case "valid":
		return len(result.Findings) == 0
	case "invalid":
		return len(result.Findings) > 0
	default:
		return false
	}
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
