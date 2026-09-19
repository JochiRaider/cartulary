package networkflow

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

type tableApplication struct {
	store          *store
	incidentAccess incidentAdmissionChecker
	receipts       tableReceiptAdapter
	safeDigester   safeDigester
	now            func() time.Time
}

type tableMutationOutcome struct {
	table   tableRecord
	receipt *storedMutationReceipt
}

func (s *tableApplication) rename(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableRenameRequest, requestHash []byte, requestID string) (tableMutationOutcome, error) {
	key := tableMutationIdempotencyKey(routeKeyTablesPatch, actorUserID, incidentID, tableID, request.ClientTxnID)
	return s.commitTableMutation(ctx, key, incidentID, actorUserID, admission.RolesEditorAdmin, requestHash, func(tx pgx.Tx) (tableRecord, error) {
		return s.store.renameTableTx(ctx, tx, renameTableParams{
			IncidentID: incidentID, ActorUserID: actorUserID, TableID: tableID,
			BaseTableVersion: request.BaseTableVersion, DisplayName: request.DisplayName,
			ClientTxnID: request.ClientTxnID, RequestID: requestID, SafeDigester: s.safeDigester, Now: s.now(),
		})
	})
}

func (s *tableApplication) softDelete(ctx context.Context, incidentID uuid.UUID, tableID string, actorUserID uuid.UUID, request tableSoftDeleteRequest, requestHash []byte, requestID string) (tableMutationOutcome, error) {
	key := tableMutationIdempotencyKey(routeKeyTablesDelete, actorUserID, incidentID, tableID, request.ClientTxnID)
	return s.commitTableMutation(ctx, key, incidentID, actorUserID, admission.RolesReviewerAdmin, requestHash, func(tx pgx.Tx) (tableRecord, error) {
		table, err := s.store.softDeleteTableTx(ctx, tx, softDeleteTableParams{
			IncidentID: incidentID, ActorUserID: actorUserID, TableID: tableID,
			BaseTableVersion: request.BaseTableVersion, ClientTxnID: request.ClientTxnID, RequestID: requestID, Now: s.now(),
		})
		if err != nil {
			return tableRecord{}, err
		}
		if err := s.store.InvalidateGraphViewsForTableTx(ctx, tx, incidentID, tableID, table.UpdatedAt); err != nil {
			return tableRecord{}, err
		}
		return table, nil
	})
}

// Incident serialization orders current admission, immutable receipt lookup and
// fresh table mutation. Replay never consults the current table lifecycle/version.
func (s *tableApplication) commitTableMutation(ctx context.Context, key authn.RouteIdempotencyKey, incidentID, actorID uuid.UUID, roles admission.RoleSet, requestHash []byte, mutate func(pgx.Tx) (tableRecord, error)) (tableMutationOutcome, error) {
	var outcome tableMutationOutcome
	err := withinTransaction(ctx, s.store.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if err := s.store.lockIncidentTx(ctx, tx, incidentID); err != nil {
			return err
		}
		if _, err := s.incidentAccess.CheckTx(ctx, tx, incidentID, actorID, admission.Requirement{AllowedRoles: roles, Lifecycle: admission.LifecycleOpen}); err != nil {
			return err
		}
		replay, found, err := s.receipts.replayTx(ctx, tx, key, requestHash)
		if err != nil || found {
			outcome = replay
			return err
		}
		table, err := mutate(tx)
		if err != nil {
			return err
		}
		outcome = tableMutationOutcome{table: table}
		return s.receipts.saveTx(ctx, tx, key, requestHash, outcome)
	})
	if err != nil {
		return tableMutationOutcome{}, err
	}
	return outcome, nil
}
