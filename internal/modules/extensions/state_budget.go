package extensions

import (
	"context"
	"errors"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

type stateBudgetKind uint8

const (
	stateBudgetProfile stateBudgetKind = iota
	stateBudgetInitialization
	stateBudgetMigration
	stateBudgetFinalValidation
)

// State budgets retain the selected owner's configured interval across nesting.
// Local/inherited describes one clock comparison, not the originating operation.
type stateBudget struct {
	extensiondeadline.Deadline
	kind      stateBudgetKind
	seconds   int64
	migration *MigrationDefinition
}

func (r *StateRuntime) newStateBudget(interval time.Duration, kind stateBudgetKind, parent *stateBudget) stateBudget {
	var inherited *extensiondeadline.Deadline
	if parent != nil {
		inherited = &parent.Deadline
	}
	selected := extensiondeadline.New(r.monotonicNowNS(), durationSeconds(interval), inherited)
	if selected.Source == extensiondeadline.SourceInherited {
		result := *parent
		result.Deadline = selected
		return result
	}
	return stateBudget{Deadline: selected, kind: kind, seconds: durationSeconds(interval)}
}

type stateBudgetTimeout struct {
	budget stateBudget
	cause  error
}

func (e *stateBudgetTimeout) Error() string { return "extension state budget expired" }
func (e *stateBudgetTimeout) Unwrap() error { return e.cause }
func (e *stateBudgetTimeout) Is(target error) bool {
	if e.budget.kind == stateBudgetFinalValidation {
		return target == ErrStateValidationFailed
	}
	return target == ErrStateMigrationTimeout
}

type initializerFailure struct{ cause error }

func (e *initializerFailure) Error() string { return "extension state initializer failed" }
func (e *initializerFailure) Unwrap() error { return e.cause }

// A prior sampled interruption owns the outcome. Otherwise a budget expiring
// while storage access settles remains part of that operation's interval.
func (r *StateRuntime) stateStorageFailure(ctx context.Context, budget stateBudget, cause error) error {
	var fatal *FatalStateIntegrityError
	var selected *stateBudgetTimeout
	if cause == nil || errors.As(cause, &fatal) || errors.As(cause, &selected) || extensiondeadline.Interrupted(cause) {
		return cause
	}
	if boundary := extensiondeadline.Sample(ctx, r.invocationPolicy(budget.Deadline)); errors.Is(boundary, context.DeadlineExceeded) {
		return errors.Join(&stateBudgetTimeout{budget: budget, cause: boundary}, cause)
	}
	return cause
}

func (r *StateRuntime) stateBudgetDiagnostic(plan StatePlan, timeout *stateBudgetTimeout, cause error) error {
	switch timeout.budget.kind {
	case stateBudgetInitialization:
		return startupFailure(cause, r.coordinator.registeredResultFinding("extension_state_initialization_timeout", StartupFinding{
			Path: "$", ReasonCode: "extension_state_initialization_timeout", Message: "Extension state initialization exceeded its timeout.",
			Details: map[string]any{"profile_id": plan.ProfileID, "timeout_seconds": timeout.budget.seconds},
		}))
	case stateBudgetFinalValidation:
		return startupFailure(cause, r.coordinator.StateValidationFailure(plan.ProfileID, "state_validation", plan.FinalValidationAlgorithmID))
	default:
		return r.migrationTimeoutDiagnostic(plan, timeout.budget.migration, timeout.budget.seconds, cause)
	}
}
