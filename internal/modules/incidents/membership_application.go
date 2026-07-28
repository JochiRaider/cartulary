package incidents

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (a *Application) ListMemberships(
	ctx context.Context,
	incidentID uuid.UUID,
) ([]MembershipRecord, error) {
	return a.repository.listMemberships(ctx, incidentID)
}

func (a *Application) CreateMembership(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	targetUser authn.UserRecord,
	request MembershipCreateRequest,
	requestHash []byte,
	requestID string,
	now time.Time,
) (MembershipCreateResult, error) {
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incident.memberships.create",
		ActorUserID: actor.ID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := a.authStore.GetRouteIdempotency(ctx, key); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MembershipCreateResult{}, fmt.Errorf("decode replayed membership payload: %w", err)
		}
		record, err := a.GetMembership(ctx, incidentID, targetUser.ID)
		if err != nil {
			return MembershipCreateResult{}, err
		}
		return MembershipCreateResult{
			Membership: record,
			Payload:    payload,
			StatusCode: http.StatusOK,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MembershipCreateResult{}, fmt.Errorf("query membership create idempotency: %w", err)
	}

	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipCreateResult{}, fmt.Errorf("begin membership create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	current, err := txRepository.getMembershipForUpdate(ctx, incidentID, targetUser.ID)
	switch {
	case errors.Is(err, ErrMembershipNotFound):
	case err != nil:
		return MembershipCreateResult{}, err
	default:
		if current.Role != request.Role {
			return MembershipCreateResult{}, ErrMembershipExistsUsePatch
		}
		payload := BuildMembershipResource(current)
		if err := authn.InsertRouteIdempotencyPayload(
			ctx,
			tx,
			key,
			&targetUser.ID,
			requestHash,
			http.StatusOK,
			payload,
		); err != nil {
			if authn.IsUniqueViolation(err) {
				return MembershipCreateResult{}, authn.ErrClientTxnConflict
			}
			return MembershipCreateResult{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return MembershipCreateResult{}, fmt.Errorf("commit existing membership transaction: %w", err)
		}
		return MembershipCreateResult{
			Membership: current,
			Payload:    payload,
			StatusCode: http.StatusOK,
		}, nil
	}

	created, err := txRepository.createMembership(ctx, createMembershipPersistenceParams{
		IncidentID:    incidentID,
		UserID:        targetUser.ID,
		Role:          request.Role,
		JoinedAt:      now,
		AddedByUserID: actor.ID,
		DisplayName:   targetUser.DisplayName,
	})
	if err != nil {
		return MembershipCreateResult{}, err
	}

	payload := BuildMembershipResource(created)
	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &targetUser.ID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_created",
		ClientTxnID:  &request.ClientTxnID,
		RequestID:    &requestID,
		AfterJSON:    payload,
	}); err != nil {
		return MembershipCreateResult{}, err
	}
	if err := authn.InsertRouteIdempotencyPayload(
		ctx,
		tx,
		key,
		&targetUser.ID,
		requestHash,
		http.StatusCreated,
		payload,
	); err != nil {
		if authn.IsUniqueViolation(err) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		return MembershipCreateResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MembershipCreateResult{}, fmt.Errorf("commit membership create transaction: %w", err)
	}
	return MembershipCreateResult{
		Membership: created,
		Payload:    payload,
		StatusCode: http.StatusCreated,
	}, nil
}

func (a *Application) GetMembership(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (MembershipRecord, error) {
	return a.repository.getMembership(ctx, incidentID, userID)
}

func (a *Application) UpdateMembership(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	userID uuid.UUID,
	request MembershipPatchRequest,
	requestID string,
	now time.Time,
) (MembershipRecord, bool, error) {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("begin membership patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	current, err := txRepository.getMembershipForUpdate(ctx, incidentID, userID)
	if err != nil {
		return MembershipRecord{}, false, err
	}
	if current.MembershipVersion != request.BaseMembershipVersion {
		return MembershipRecord{}, false, ErrMembershipVersionConflict
	}
	if current.Role == request.Role {
		if err := tx.Commit(ctx); err != nil {
			return MembershipRecord{}, false, fmt.Errorf("commit membership no-op patch transaction: %w", err)
		}
		return current, false, nil
	}

	adminCount, err := txRepository.countIncidentAdmins(ctx, incidentID)
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("count incident admins: %w", err)
	}
	if WouldLeaveNoIncidentAdmins(current.Role, adminCount, &request.Role, false) {
		return MembershipRecord{}, false, ErrLastIncidentAdmin
	}

	updated, err := txRepository.updateMembership(ctx, updateMembershipPersistenceParams{
		IncidentID:      incidentID,
		UserID:          userID,
		Role:            request.Role,
		UpdatedAt:       now,
		UpdatedByUserID: actor.ID,
		DisplayName:     current.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, false, err
	}

	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &userID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_updated",
		RequestID:    &requestID,
		BeforeJSON:   BuildMembershipResource(current),
		AfterJSON:    BuildMembershipResource(updated),
	}); err != nil {
		return MembershipRecord{}, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MembershipRecord{}, false, fmt.Errorf("commit membership patch transaction: %w", err)
	}
	return updated, true, nil
}

func (a *Application) DeleteMembership(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	userID uuid.UUID,
	request MembershipDeleteRequest,
	requestID string,
) error {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin membership delete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	current, err := txRepository.getMembershipForUpdate(ctx, incidentID, userID)
	if err != nil {
		return err
	}
	if current.MembershipVersion != request.BaseMembershipVersion {
		return ErrMembershipVersionConflict
	}

	adminCount, err := txRepository.countIncidentAdmins(ctx, incidentID)
	if err != nil {
		return fmt.Errorf("count incident admins: %w", err)
	}
	if WouldLeaveNoIncidentAdmins(current.Role, adminCount, nil, true) {
		return ErrLastIncidentAdmin
	}

	if err := txRepository.deleteMembership(ctx, incidentID, userID); err != nil {
		return err
	}

	if err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &userID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_deleted",
		RequestID:    &requestID,
		BeforeJSON:   BuildMembershipResource(current),
		AfterJSON:    map[string]any{"deleted": true},
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit membership delete transaction: %w", err)
	}
	return nil
}
