package authn

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
)

type administrativeProjection struct {
	ActorUserID  *uuid.UUID
	TargetUserID *uuid.UUID
	RawSource    string
	RawKind      string
	ReasonCode   *string
	ClientTxnID  *string
	RequestID    *string
	Before       any
	After        any
	OccurredAt   time.Time
	Source       string
	ActionCode   string
	TargetKind   string
	TargetID     string
	Changes      []administrativeaudit.Change
}

func appendAdministrativeProjectionTx(ctx context.Context, tx pgx.Tx, projection administrativeProjection) error {
	actorKind := administrativeaudit.ActorSystem
	if projection.ActorUserID != nil {
		actorKind = administrativeaudit.ActorUser
	}
	targetID := projection.TargetID
	_, err := administrativeaudit.AppendTx(ctx, tx, administrativeaudit.RawEvent{
		ActorUserID:  projection.ActorUserID,
		TargetUserID: projection.TargetUserID,
		EventSource:  projection.RawSource,
		EventKind:    projection.RawKind,
		ReasonCode:   projection.ReasonCode,
		ClientTxnID:  projection.ClientTxnID,
		RequestID:    projection.RequestID,
		Before:       projection.Before,
		After:        projection.After,
		OccurredAt:   projection.OccurredAt,
	}, administrativeaudit.Event{
		ScopeKind:   administrativeaudit.ScopeDeployment,
		OccurredAt:  projection.OccurredAt,
		ActorKind:   actorKind,
		ActorUserID: projection.ActorUserID,
		Source:      projection.Source,
		ActionCode:  projection.ActionCode,
		TargetKind:  projection.TargetKind,
		TargetID:    &targetID,
		Changes:     projection.Changes,
		ReasonCode:  projection.ReasonCode,
	})
	return err
}

func userUpdateAdministrativeChanges(before UserRecord, after UserRecord) (string, []administrativeaudit.Change) {
	actionCode := administrativeaudit.ActionUserProfileUpdated
	if before.IsDeploymentAdmin != after.IsDeploymentAdmin {
		if after.IsDeploymentAdmin {
			actionCode = administrativeaudit.ActionDeploymentAdminGranted
		} else {
			actionCode = administrativeaudit.ActionDeploymentAdminRevoked
		}
	} else if before.IsActive != after.IsActive {
		actionCode = administrativeaudit.ActionUserStatusChanged
	}
	changes := make([]administrativeaudit.Change, 0, 6)
	if before.DisplayName != after.DisplayName {
		changes = append(changes, administrativeaudit.Visible("display_name", before.DisplayName, after.DisplayName))
	}
	if before.Email != after.Email {
		changes = append(changes, administrativeaudit.Visible("email", before.Email, after.Email))
	}
	if before.IsActive != after.IsActive {
		changes = append(changes, administrativeaudit.Visible("is_active", before.IsActive, after.IsActive))
	}
	if before.IsDeploymentAdmin != after.IsDeploymentAdmin {
		changes = append(changes, administrativeaudit.Visible("is_deployment_admin", before.IsDeploymentAdmin, after.IsDeploymentAdmin))
	}
	if before.MFARequired != after.MFARequired {
		changes = append(changes, administrativeaudit.Visible("mfa_required", before.MFARequired, after.MFARequired))
	}
	changes = append(changes, administrativeaudit.Visible("user_version", before.UserVersion, after.UserVersion))
	return actionCode, changes
}
