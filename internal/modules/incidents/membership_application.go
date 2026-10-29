package incidents

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func (a *Application) ListMemberships(
	ctx context.Context,
	incidentID uuid.UUID,
	page MembershipListPageRequest,
) ([]MembershipRecord, error) {
	if page.Limit < 1 || page.Limit > 501 {
		return nil, errors.New("incidents: invalid membership page limit")
	}
	return a.repository.listMemberships(ctx, incidentID, page)
}

func (a *Application) CreateMembership(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	request MembershipCreateAdmission,
	requestID string,
) (MembershipCreateResult, error) {
	if !request.admitted {
		return MembershipCreateResult{}, errInvalidMutationAdmission
	}
	requestHash := request.requestHash[:]
	key := authn.RouteIdempotencyKey{
		RouteKey:    "incident.memberships.create",
		ActorUserID: actor.ID,
		ScopeKey:    incidentID.String(),
		ClientTxnID: request.clientTxnID,
	}

	tx, err := a.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MembershipCreateResult{}, fmt.Errorf("begin membership create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := authn.LockRouteIdempotencyTx(ctx, tx, key); err != nil {
		return MembershipCreateResult{}, err
	}

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
	if existing, err := authn.GetRouteIdempotencyTx(ctx, tx, key); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MembershipCreateResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MembershipCreateResult{}, fmt.Errorf("decode replayed membership payload: %w", err)
		}
		return MembershipCreateResult{
			Payload: payload,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MembershipCreateResult{}, fmt.Errorf("query membership create idempotency: %w", err)
	}
	targetUser, err := a.resolveMembershipTargetTx(ctx, tx, request)
	if err != nil {
		return MembershipCreateResult{}, err
	}
	mutationTime, err := a.recordedMutationTime()
	if err != nil {
		return MembershipCreateResult{}, err
	}
	current, err := txRepository.getMembershipForUpdate(ctx, incidentID, targetUser.ID)
	switch {
	case errors.Is(err, ErrMembershipNotFound):
	case err != nil:
		return MembershipCreateResult{}, err
	default:
		if current.Role != request.role {
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
		Role:          request.role,
		JoinedAt:      mutationTime,
		AddedByUserID: actor.ID,
		DisplayName:   targetUser.DisplayName,
	})
	if err != nil {
		return MembershipCreateResult{}, err
	}

	payload := BuildMembershipResource(created)
	if _, err := insertAuditEvent(ctx, tx, auditEvent{
		actorUserID:  &actor.ID,
		targetUserID: &targetUser.ID,
		incidentID:   &incidentID,
		kind:         auditMembershipCreated,
		source:       auditSourceAPI,
		clientTxnID:  &request.clientTxnID,
		requestID:    &requestID,
		afterJSON:    payload,
		roles:        auditRoleFacts{after: auditRole(created.Role)},
		occurredAt:   mutationTime,
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

func (a *Application) UpdateMembership(
	ctx context.Context,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	userID uuid.UUID,
	request MembershipPatchAdmission,
	requestID string,
) (MembershipRecord, bool, error) {
	if !request.admitted {
		return MembershipRecord{}, false, errInvalidMutationAdmission
	}
	mutationTime, err := a.recordedMutationTime()
	if err != nil {
		return MembershipRecord{}, false, err
	}
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
	if current.MembershipVersion != request.baseMembershipVersion {
		return MembershipRecord{}, false, ErrMembershipVersionConflict
	}
	if current.Role == request.role {
		if err := tx.Commit(ctx); err != nil {
			return MembershipRecord{}, false, fmt.Errorf("commit membership no-op patch transaction: %w", err)
		}
		return current, false, nil
	}

	adminCount, err := txRepository.countIncidentAdmins(ctx, incidentID)
	if err != nil {
		return MembershipRecord{}, false, fmt.Errorf("count incident admins: %w", err)
	}
	if wouldLeaveNoIncidentAdmins(current.Role, adminCount, &request.role, false) {
		return MembershipRecord{}, false, ErrLastIncidentAdmin
	}

	updated, err := txRepository.updateMembership(ctx, updateMembershipPersistenceParams{
		IncidentID:      incidentID,
		UserID:          userID,
		Role:            request.role,
		UpdatedAt:       mutationTime,
		UpdatedByUserID: actor.ID,
		DisplayName:     current.DisplayName,
	})
	if err != nil {
		return MembershipRecord{}, false, err
	}

	if _, err := insertAuditEvent(ctx, tx, auditEvent{
		actorUserID:  &actor.ID,
		targetUserID: &userID,
		incidentID:   &incidentID,
		kind:         auditMembershipUpdated,
		source:       auditSourceAPI,
		requestID:    &requestID,
		beforeJSON:   BuildMembershipResource(current),
		afterJSON:    BuildMembershipResource(updated),
		roles: auditRoleFacts{
			before: auditRole(current.Role),
			after:  auditRole(updated.Role),
		},
		occurredAt: mutationTime,
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
	request MembershipDeleteAdmission,
	requestID string,
) (MembershipDeleteResult, error) {
	if !request.admitted {
		return MembershipDeleteResult{}, errInvalidMutationAdmission
	}
	mutationTime, err := a.recordedMutationTime()
	if err != nil {
		return MembershipDeleteResult{}, err
	}
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
	if current.MembershipVersion != request.baseMembershipVersion {
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
		actorUserID:  &actor.ID,
		targetUserID: &userID,
		incidentID:   &incidentID,
		kind:         auditMembershipDeleted,
		source:       auditSourceAPI,
		requestID:    &requestID,
		beforeJSON:   BuildMembershipResource(current),
		afterJSON:    map[string]any{"deleted": true},
		roles:        auditRoleFacts{before: auditRole(current.Role)},
		occurredAt:   mutationTime,
	})
	if err != nil {
		return MembershipDeleteResult{}, err
	}
	commit, err := NewTerminalMutationCommit(auditEventID)
	if err != nil {
		return MembershipDeleteResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MembershipDeleteResult{}, fmt.Errorf("commit membership delete transaction: %w", err)
	}
	return MembershipDeleteResult{Commit: commit}, nil
}

// resolveMembershipTargetTx runs only for a fresh receipt, holding the target
// account lock until the membership and its receipt commit.
func (a *Application) resolveMembershipTargetTx(ctx context.Context, tx pgx.Tx, request MembershipCreateAdmission) (authn.UserRecord, error) {
	var target authn.UserRecord
	var err error
	if request.userID != nil {
		target, err = a.authStore.GetUserByIDForUpdateTx(ctx, tx, *request.userID)
	} else if request.email != nil {
		target, err = a.authStore.GetUserByNormalizedEmailForUpdateTx(ctx, tx, *request.email)
	} else {
		return target, errInvalidMutationAdmission
	}
	if err != nil {
		return target, err
	}
	if !target.IsActive {
		return target, ErrMembershipTargetInactive
	}
	if err := validateMembershipCreateTarget(request, target); err != nil {
		return target, err
	}
	return target, nil
}

func validateMembershipCreateTarget(request MembershipCreateAdmission, targetUser authn.UserRecord) error {
	if request.userID != nil {
		if request.email != nil || targetUser.ID == uuid.Nil || *request.userID != targetUser.ID {
			return errors.New("incidents: resolved membership target does not match user_id selector")
		}
		return nil
	}
	if request.email == nil || targetUser.ID == uuid.Nil {
		return errors.New("incidents: resolved membership target requires exactly one selector")
	}
	_, requestComparison, requestOK := authn.NormalizeEmailAddress(*request.email)
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
