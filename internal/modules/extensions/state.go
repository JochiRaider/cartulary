package extensions

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

var (
	ErrStateMetadataMissing      = errors.New("extension_state_metadata_missing")
	ErrStateIncomplete           = errors.New("extension_state_incomplete")
	ErrStateLineageMismatch      = errors.New("extension_state_lineage_mismatch")
	ErrStateVersionUnsupported   = errors.New("extension_state_version_unsupported")
	ErrStateMigrationUnavailable = errors.New("extension_state_migration_unavailable")
	ErrStateValidationFailed     = errors.New("extension_state_validation_failed")
	ErrStateCommitIndeterminate  = errors.New("extension_state_commit_indeterminate")
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
		if policy == "allowed" {
			return StateInitialize, nil
		}
		if policy == "forbidden" {
			return "", ErrStateIncomplete
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

type StateValidator func(context.Context, extensionstore.FamilyCountReader) error
type StateMigration func(context.Context, *extensionstore.Tx) error

type StateRuntime struct {
	store       *extensionstore.Store
	validators  map[string]StateValidator
	migrations  map[string]StateMigration
	now         func() time.Time
	lockTimeout time.Duration
}

func NewStateRuntime(store *extensionstore.Store, validators map[string]StateValidator, migrations map[string]StateMigration, now func() time.Time, lockTimeout time.Duration) (*StateRuntime, error) {
	if store == nil {
		return nil, errors.New("extension state runtime requires a store")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &StateRuntime{store: store, validators: cloneStateValidators(validators), migrations: cloneStateMigrations(migrations), now: now, lockTimeout: lockTimeout}, nil
}

func cloneStateValidators(source map[string]StateValidator) map[string]StateValidator {
	result := make(map[string]StateValidator, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func cloneStateMigrations(source map[string]StateMigration) map[string]StateMigration {
	result := make(map[string]StateMigration, len(source))
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
	validator := r.validators[plan.FinalValidationAlgorithmID]
	if validator == nil {
		return ErrStateValidationFailed
	}
	familyIDs := append([]string(nil), plan.DatabaseFamilyIDs...)
	familyIDs = append(familyIDs, plan.ObjectReferenceFamilyIDs...)
	sort.Strings(familyIDs)
	return r.store.WithProfileLock(ctx, plan.ProfileID, r.lockTimeout, func(session *extensionstore.Session) error {
		for {
			tx, err := session.Begin(ctx)
			if err != nil {
				return err
			}
			metadata, err := tx.StateMetadata(ctx, plan.ProfileID)
			if err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			counts, err := tx.FamilyCounts(ctx, familyIDs)
			if err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			statePresent := false
			for _, count := range counts {
				statePresent = statePresent || count > 0
			}
			decision, err := DecideStatePresence(plan.EmptyStatePolicy, metadata != nil, statePresent)
			if err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			if decision == StateInitialize {
				if plan.InitializationKind != "empty" {
					_, _ = tx.Rollback(ctx)
					return ErrStateMigrationUnavailable
				}
				if err := validator(ctx, tx); err != nil {
					_, _ = tx.Rollback(ctx)
					return fmt.Errorf("%w", ErrStateValidationFailed)
				}
				now := r.now().UTC()
				if err := tx.InsertStateMetadata(ctx, extensionstore.StateMetadata{ProfileID: plan.ProfileID, MigrationLineageID: plan.MigrationLineageID, StateVersion: plan.CurrentStateVersion, MetadataVersion: 1, CreatedAt: now, UpdatedAt: now}); err != nil {
					_, _ = tx.Rollback(ctx)
					return err
				}
				return requireProvenStateCommit(ctx, tx)
			}
			if metadata.MigrationLineageID != plan.MigrationLineageID {
				_, _ = tx.Rollback(ctx)
				return ErrStateLineageMismatch
			}
			if metadata.StateVersion > plan.CurrentStateVersion || metadata.StateVersion < plan.MinimumMigratableStateVersion {
				_, _ = tx.Rollback(ctx)
				return ErrStateVersionUnsupported
			}
			if metadata.StateVersion == plan.CurrentStateVersion {
				if err := validator(ctx, tx); err != nil {
					_, _ = tx.Rollback(ctx)
					return fmt.Errorf("%w", ErrStateValidationFailed)
				}
				_, err = tx.Rollback(ctx)
				return err
			}
			definition, ok := nextMigration(plan.MigrationDefinitions, metadata.StateVersion)
			migration := r.migrations[definition.MigrationID]
			if !ok || migration == nil {
				_, _ = tx.Rollback(ctx)
				return ErrStateMigrationUnavailable
			}
			if err := migration(ctx, tx); err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			now := r.now().UTC()
			if err := tx.InsertMigrationLedger(ctx, extensionstore.MigrationLedgerEntry{ProfileID: plan.ProfileID, MigrationLineageID: plan.MigrationLineageID, MigrationID: definition.MigrationID, FromStateVersion: definition.FromVersion, ToStateVersion: definition.ToVersion, MigrationDefinitionSHA256: definition.DefinitionSHA256, CommittedAt: now, ResultingStateVersion: definition.ToVersion}); err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			if err := tx.UpdateStateMetadata(ctx, *metadata, definition.ToVersion, definition.MigrationID, now); err != nil {
				_, _ = tx.Rollback(ctx)
				return err
			}
			if err := requireProvenStateCommit(ctx, tx); err != nil {
				return err
			}
		}
	})
}

func nextMigration(definitions []MigrationDefinition, version int) (MigrationDefinition, bool) {
	for _, definition := range definitions {
		if definition.FromVersion == version && definition.ToVersion == version+1 {
			return definition, true
		}
	}
	return MigrationDefinition{}, false
}

func requireProvenStateCommit(ctx context.Context, tx *extensionstore.Tx) error {
	outcome, err := tx.Commit(ctx)
	if outcome != extensionstore.CommitProven {
		return fmt.Errorf("%w: %v", ErrStateCommitIndeterminate, err)
	}
	return nil
}
