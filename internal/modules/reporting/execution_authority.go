package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const reportingJobAuthorizationDeniedCode = "authorization_denied"

var errReportingAuthorityUnavailable = errors.New("reporting: initiating actor is no longer authorized")

// authorizeExecutionTx derives authority from the fenced Jobs admission, before
// private input reads and again in the final success transaction. The shared
// incident/membership/account locks order publication against revocation. They
// never span source materialization or rendering.
func (s *Store) authorizeExecutionTx(ctx context.Context, tx pgx.Tx, execution jobs.Execution) error {
	if s == nil || s.pool == nil || s.jobTransactions == nil || tx == nil {
		return jobs.ErrNotConfigured
	}
	metadata, err := s.jobTransactions.ExtensionFinalizationContextTx(ctx, tx, execution)
	if err != nil {
		return err
	}
	if metadata.Definition.Extension == nil || metadata.Definition.Extension.OwnerProfileID != ProfileID ||
		metadata.Scope.Kind != jobs.ScopeKindIncident || metadata.Scope.IncidentID == nil ||
		metadata.ActorUserID == uuid.Nil || metadata.OperatorOperationID != uuid.Nil {
		return jobs.ErrInvalidJobDefinition
	}
	roles := admission.RolesEditorReviewerAdmin
	switch metadata.Definition.JobKind {
	case SnapshotCreateJobKind, ReleaseCreateJobKind:
	case CompositionPreviewJobKind:
		roles = admission.RolesMember
	default:
		return jobs.ErrInvalidJobDefinition
	}
	_, err = admission.NewChecker(s.pool).CheckTx(ctx, tx, *metadata.Scope.IncidentID, metadata.ActorUserID,
		admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleAny})
	var denied *admission.Denied
	if errors.As(err, &denied) {
		return errReportingAuthorityUnavailable
	}
	if err != nil {
		return err
	}
	active, err := authn.NewStore(s.pool).IsActiveUserTx(ctx, tx, metadata.ActorUserID)
	if err != nil {
		return err
	}
	if !active {
		return errReportingAuthorityUnavailable
	}
	return nil
}
