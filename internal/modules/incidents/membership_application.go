package incidents

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
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
	requestID string,
	now time.Time,
) (MembershipCreateResult, error) {
	if err := validateMembershipCreateTarget(request, targetUser); err != nil {
		return MembershipCreateResult{}, err
	}
	requestHash := membershipCreateRequestHash(request)
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
	if _, err := txRepository.getIncidentForUpdate(ctx, incidentID); err != nil {
		return MembershipCreateResult{}, err
	}
	if _, err := a.admission.CheckTx(ctx, tx, incidentID, actor.ID, admission.Requirement{
		AllowedRoles: admission.RolesAdmin,
		Lifecycle:    admission.LifecycleAny,
	}); err != nil {
		return MembershipCreateResult{}, err
	}
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
			persistedSuccessStatus,
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
	if _, err := insertAuditEvent(ctx, tx, auditEvent{
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
		persistedCreatedStatus,
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
		Created:    true,
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
	if _, err := txRepository.getIncidentForUpdate(ctx, incidentID); err != nil {
		return MembershipRecord{}, false, err
	}
	if _, err := a.admission.CheckTx(ctx, tx, incidentID, actor.ID, admission.Requirement{
		AllowedRoles: admission.RolesAdmin,
		Lifecycle:    admission.LifecycleAny,
	}); err != nil {
		return MembershipRecord{}, false, err
	}
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
	if wouldLeaveNoIncidentAdmins(current.Role, adminCount, &request.Role, false) {
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

	if _, err := insertAuditEvent(ctx, tx, auditEvent{
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
) (MembershipDeleteResult, error) {
	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipDeleteResult{}, fmt.Errorf("begin membership delete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	txRepository := newRepository(tx)
	if _, err := txRepository.getIncidentForUpdate(ctx, incidentID); err != nil {
		return MembershipDeleteResult{}, err
	}
	if _, err := a.admission.CheckTx(ctx, tx, incidentID, actor.ID, admission.Requirement{
		AllowedRoles: admission.RolesAdmin,
		Lifecycle:    admission.LifecycleAny,
	}); err != nil {
		return MembershipDeleteResult{}, err
	}
	current, err := txRepository.getMembershipForUpdate(ctx, incidentID, userID)
	if err != nil {
		return MembershipDeleteResult{}, err
	}
	if current.MembershipVersion != request.BaseMembershipVersion {
		return MembershipDeleteResult{}, ErrMembershipVersionConflict
	}

	adminCount, err := txRepository.countIncidentAdmins(ctx, incidentID)
	if err != nil {
		return MembershipDeleteResult{}, fmt.Errorf("count incident admins: %w", err)
	}
	if wouldLeaveNoIncidentAdmins(current.Role, adminCount, nil, true) {
		return MembershipDeleteResult{}, ErrLastIncidentAdmin
	}

	if err := txRepository.deleteMembership(ctx, incidentID, userID); err != nil {
		return MembershipDeleteResult{}, err
	}

	auditEventID, err := insertAuditEvent(ctx, tx, auditEvent{
		ActorUserID:  &actor.ID,
		TargetUserID: &userID,
		IncidentID:   &incidentID,
		EventSource:  "incidents",
		EventKind:    "incident_membership_deleted",
		RequestID:    &requestID,
		BeforeJSON:   BuildMembershipResource(current),
		AfterJSON:    map[string]any{"deleted": true},
	})
	if err != nil {
		return MembershipDeleteResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MembershipDeleteResult{}, fmt.Errorf("commit membership delete transaction: %w", err)
	}
	return MembershipDeleteResult{Commit: NewTerminalMutationCommit(auditEventID)}, nil
}

func validateMembershipCreateTarget(request MembershipCreateRequest, targetUser authn.UserRecord) error {
	if request.UserID != nil {
		if request.Email != nil || targetUser.ID == uuid.Nil || *request.UserID != targetUser.ID {
			return errors.New("incidents: resolved membership target does not match user_id selector")
		}
		return nil
	}
	if request.Email == nil || targetUser.ID == uuid.Nil {
		return errors.New("incidents: resolved membership target requires exactly one selector")
	}
	_, requestComparison, requestOK := authn.NormalizeEmailAddress(*request.Email)
	_, targetComparison, targetOK := authn.NormalizeEmailAddress(targetUser.Email)
	if !requestOK || !targetOK || requestComparison != targetComparison {
		return errors.New("incidents: resolved membership target does not match email selector")
	}
	return nil
}

func wouldLeaveNoIncidentAdmins(currentRole string, adminCount int, nextRole *string, deleting bool) bool {
	if currentRole != "admin" {
		return false
	}
	if !deleting && nextRole != nil && *nextRole == "admin" {
		return false
	}
	return adminCount <= 1
}
