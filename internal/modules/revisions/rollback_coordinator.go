package revisions

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"slices"
)

type rollbackCoordinator struct {
	store      *commandStore
	repository rollbackQueryRepository
	planner    rollbackPlanner
	locker     rollbackRecordLocker
	applier    rollbackTransactionalApplier
	results    rollbackResultCoordinator
}

func newRollbackCoordinator(store *commandStore) rollbackCoordinator {
	repository := rollbackQueryRepository{store: store}
	publication := rollbackPublicationService{store: store}
	return rollbackCoordinator{
		store:      store,
		repository: repository,
		planner:    rollbackPlanner{targetSemantics: store.targetSemantics},
		locker:     rollbackRecordLocker{envelopes: store.envelopes},
		applier: rollbackTransactionalApplier{
			commandStore: store,
			repository:   repository,
			publication:  publication,
		},
		results: rollbackResultCoordinator{},
	}
}

func (s *commandStore) RollbackRecord(ctx context.Context, command RollbackCommand) (RollbackResult, error) {
	return newRollbackCoordinator(s).execute(ctx, command)
}

func (c rollbackCoordinator) execute(ctx context.Context, command RollbackCommand) (RollbackResult, error) {
	recordID := command.RecordID
	request := command.Request
	idempotencyKey := IdempotencyKey{
		RouteKey:    rollbackRouteKey,
		ActorID:     command.Actor,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := c.store.idempotency.Get(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return RollbackResult{}, ErrClientTxnConflict
		}
		payload, err := decodeStoredRollbackPayload(existing.ResponseJSON)
		if err != nil {
			return RollbackResult{}, err
		}
		return c.results.replayed(payload, request.ClientTxnID), nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return RollbackResult{}, fmt.Errorf("query rollback idempotency: %w", err)
	}

	tx, err := c.store.transactions.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return RollbackResult{}, fmt.Errorf("begin rollback transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	record, err := c.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		return RollbackResult{}, err
	}

	protected, err := c.repository.loadRollbackProtectedSetTx(ctx, tx, record, request.Target)
	if err != nil {
		return RollbackResult{}, err
	}
	var (
		preparedClaims  []preparedIdentifierClaimRestore
		claimPrepareErr error
	)
	if protected.DeferredErr == nil {
		preLockPlan, preLockErr := c.loadPlanTx(ctx, tx, record, request.Target)
		if preLockErr == nil {
			preLockEnvelopes, envelopeErr := c.repository.loadRollbackRecordEnvelopesTx(ctx, tx, preLockPlan.Affected)
			if envelopeErr == nil {
				preLockPlan, preLockErr = c.planner.finalize(preLockPlan, preLockEnvelopes)
			}
			if preLockErr == nil {
				preparedClaims, claimPrepareErr = c.applier.prepareIdentifierClaimRestoresTx(ctx, tx, record, preLockPlan)
			}
		}
	}
	if err := c.locker.lockTx(ctx, tx, protected.Affected); err != nil {
		return RollbackResult{}, err
	}
	record, err = c.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, true)
	if err != nil {
		return RollbackResult{}, err
	}
	if err := c.store.authorization.AuthorizeCommandTx(ctx, tx, record.IncidentID, command.Actor, CommandRollback); err != nil {
		return RollbackResult{}, err
	}
	if record.DeletedAt != nil {
		return RollbackResult{}, ErrRecordDeletedUseRestore
	}
	if record.RowVersion != request.BaseRowVersion {
		return RollbackResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: record.RowVersion}
	}
	if protected.DeferredErr != nil {
		return RollbackResult{}, protected.DeferredErr
	}
	plan, err := c.loadPlanTx(ctx, tx, record, request.Target)
	if err != nil {
		return RollbackResult{}, err
	}
	if request.Target.Kind != "row_restore" {
		if err := c.repository.validateCanonicalRowTargetsTx(ctx, tx, plan); err != nil {
			return RollbackResult{}, err
		}
	}
	envelopes, err := c.repository.loadRollbackRecordEnvelopesTx(ctx, tx, plan.Affected)
	if err != nil {
		return RollbackResult{}, err
	}
	plan, err = c.planner.finalize(plan, envelopes)
	if err != nil {
		return RollbackResult{}, err
	}
	if !slices.Equal(canonicalRecordIDs(protected.Affected), canonicalRecordIDs(plan.Affected)) {
		return RollbackResult{}, &RollbackPreconditionError{ReasonCode: "stale_target"}
	}
	if request.Target.Kind != "row_restore" {
		if err := c.repository.ensurePlanCurrentTx(ctx, tx, plan); err != nil {
			return RollbackResult{}, err
		}
	}
	if claimPrepareErr != nil {
		return RollbackResult{}, claimPrepareErr
	}
	var applied rollbackApplyResult
	if request.Target.Kind == "row_restore" {
		applied, err = c.applier.applyRowRestorePlanTx(ctx, tx, command.Actor, record, plan, request, command.RequestID, command.effectiveAt)
	} else {
		applied, err = c.applier.applyRollbackPlanTx(ctx, tx, command.Actor, record, plan, request, command.RequestID, command.effectiveAt)
	}
	if err != nil {
		return RollbackResult{}, err
	}
	if err := finalizeIdentifierClaimRestoresTx(ctx, tx, preparedClaims); err != nil {
		return RollbackResult{}, err
	}
	payload := c.results.payload(record, request.Target, plan, applied)
	if err := c.store.idempotency.PutSuccessTx(ctx, tx, idempotencyKey, command.RequestHash, payload); err != nil {
		return RollbackResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return RollbackResult{}, fmt.Errorf("commit rollback transaction: %w", err)
	}
	return c.results.committed(payload, record.IncidentID, request.ClientTxnID, applied.Changes), nil
}

func (c rollbackCoordinator) loadPlanTx(ctx context.Context, tx pgx.Tx, record rollbackRecordEnvelope, target RollbackTarget) (rollbackPlan, error) {
	switch target.Kind {
	case "history_entry":
		return c.repository.loadHistoryEntryRollbackPlanTx(ctx, tx, record, target.HistoryEntryRef)
	case "change_set":
		return c.repository.loadChangeSetRollbackPlanTx(ctx, tx, record, target.ChangeSetID)
	case "row_restore":
		return c.repository.loadRowRestorePlanTx(ctx, tx, record, target.RestoreToRevisionNo)
	default:
		return rollbackPlan{}, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
}
