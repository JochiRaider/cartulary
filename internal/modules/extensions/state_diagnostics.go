package extensions

import "errors"

// Storage failures are operational failures, not owner-algorithm results. Keep
// their cause for internal handling without making driver text a diagnostic.
type stateAccessError struct{ cause error }

func (e *stateAccessError) Error() string { return "extension state storage admission failed" }
func (e *stateAccessError) Unwrap() error { return e.cause }

func startupFailure(cause error, finding StartupFinding) error {
	return &StartupValidationError{Findings: []StartupFinding{finding}, cause: cause}
}

func (r *StateRuntime) stateDiagnostic(plan StatePlan, snapshot StateSnapshot, err error) error {
	if err == nil {
		return nil
	}
	var diagnosed *StartupValidationError
	var fatal *FatalStateIntegrityError
	var storage *stateAccessError
	if errors.As(err, &fatal) {
		return err
	}
	var expired *stateBudgetTimeout
	if errors.As(err, &expired) {
		return r.stateBudgetDiagnostic(plan, expired, err)
	}
	if errors.As(err, &diagnosed) {
		// Keep secondary storage causes internal, even when they wrap a finding.
		return &StartupValidationError{Findings: diagnosed.Findings, cause: err}
	}
	if errors.As(err, &storage) {
		return &stateAccessError{cause: err}
	}
	var initialization *initializerFailure
	if errors.As(err, &initialization) {
		return startupFailure(err, r.coordinator.registeredResultFinding("extension_state_initialization_failed", StartupFinding{Path: "$", ReasonCode: "extension_state_initialization_failed", Message: "Extension state initialization failed.", Details: map[string]any{"profile_id": plan.ProfileID, "algorithm_id": plan.InitializationAlgorithmID}}))
	}
	code, message := "", ""
	details := map[string]any{"profile_id": plan.ProfileID, "migration_lineage_id": plan.MigrationLineageID}
	switch {
	case errors.Is(err, ErrStateMetadataMissing):
		code, message = "extension_state_metadata_missing", "Authoritative extension state is missing required state metadata."
	case errors.Is(err, ErrStateIncomplete) && plan.EmptyStatePolicy == "forbidden":
		code, message = "extension_state_incomplete", "Extension state metadata requires authoritative state under the declared empty-state policy."
	case errors.Is(err, ErrStateMigrationDefinitionChanged):
		for _, entry := range snapshot.Ledger {
			for _, definition := range plan.MigrationLedgerDefinitions {
				if definition.MigrationID == entry.MigrationID && definition.DefinitionSHA256 != entry.MigrationDefinitionSHA256 && validSHA256(entry.MigrationDefinitionSHA256) {
					return startupFailure(err, r.coordinator.registeredResultFinding("extension_migration_definition_changed", StartupFinding{Path: "$", ReasonCode: "extension_migration_definition_changed", Message: "A committed extension migration definition changed.", Details: map[string]any{"profile_id": plan.ProfileID, "migration_id": definition.MigrationID, "expected_digest": definition.DefinitionSHA256, "actual_digest": entry.MigrationDefinitionSHA256}}))
				}
			}
		}
		return &stateAccessError{cause: err}
	case errors.Is(err, ErrStateVersionUnsupported), errors.Is(err, ErrStateLineageMismatch), errors.Is(err, ErrStateMigrationUnavailable):
		if snapshot.Metadata == nil {
			return &stateAccessError{cause: err}
		}
		code, message = "extension_state_version_unsupported", "Stored extension state is not supported by the packaged implementation."
		details["minimum_migratable_state_version"] = plan.MinimumMigratableStateVersion
		details["current_state_version"] = plan.CurrentStateVersion
		details["stored_state_version"] = snapshot.Metadata.StateVersion
	case errors.Is(err, ErrStateMigrationLockTimeout):
		code, message = "extension_migration_lock_timeout", "The extension migration lock could not be acquired before timeout."
		details = map[string]any{"profile_id": plan.ProfileID, "timeout_seconds": durationSeconds(r.lockTimeout)}
	case errors.Is(err, ErrStateMigrationTimeout):
		return r.migrationTimeoutDiagnostic(plan, nil, durationSeconds(r.profileTimeout), err)
	case errors.Is(err, ErrStateValidationFailed):
		return startupFailure(err, r.coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID))
	default:
		return &stateAccessError{cause: err}
	}
	return startupFailure(err, r.coordinator.registeredResultFinding(code, StartupFinding{Path: "$", ReasonCode: code, Message: message, Details: details}))
}

func (r *StateRuntime) migrationTimeoutDiagnostic(plan StatePlan, definition *MigrationDefinition, seconds int64, cause error) error {
	details := map[string]any{"profile_id": plan.ProfileID, "operation_kind": "profile_migration", "migration_id": nil, "from_state_version": nil, "to_state_version": nil, "timeout_seconds": seconds}
	if definition != nil {
		details["operation_kind"] = "migration_step"
		details["migration_id"] = definition.MigrationID
		details["from_state_version"] = definition.FromVersion
		details["to_state_version"] = definition.ToVersion
	}
	return startupFailure(cause, r.coordinator.registeredResultFinding("extension_migration_timeout", StartupFinding{Path: "$", ReasonCode: "extension_migration_timeout", Message: "Extension state migration exceeded its timeout.", Details: details}))
}
