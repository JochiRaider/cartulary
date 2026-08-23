package revisions

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type preparedIdentifierClaimRestore struct {
	recordIDs []uuid.UUID
	provider  rollbackcontract.IdentifierClaimRestoreProvider
}

type identifierClaimPreparation struct {
	providerID string
	provider   rollbackcontract.IdentifierClaimRestoreProvider
	records    map[uuid.UUID]map[string]any
}

func (a rollbackTransactionalApplier) prepareIdentifierClaimRestoresTx(
	ctx context.Context,
	tx pgx.Tx,
	record rollbackRecordEnvelope,
	plan rollbackPlan,
) ([]preparedIdentifierClaimRestore, error) {
	preparations := make(map[string]identifierClaimPreparation)
	add := func(targetKind string, recordID uuid.UUID, retainedValue map[string]any) error {
		if recordID == uuid.Nil {
			return nil
		}
		envelope, err := a.repository.loadRollbackRecordEnvelopeTx(ctx, tx, recordID, false)
		if err != nil {
			return err
		}
		provider, err := a.targetSemantics.rowProvider(targetKind, envelope.RecordType)
		if err != nil {
			return &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
		}
		claimProvider, ok := provider.(rollbackcontract.IdentifierClaimRestoreProvider)
		if !ok {
			return nil
		}
		providerID := "row/" + envelope.RecordType
		preparation, found := preparations[providerID]
		if !found {
			preparation = identifierClaimPreparation{
				providerID: providerID,
				provider:   claimProvider,
				records:    make(map[uuid.UUID]map[string]any),
			}
		}
		if _, found := preparation.records[recordID]; found && retainedValue == nil {
			return nil
		}
		preparation.records[recordID] = retainedValue
		preparations[providerID] = preparation
		return nil
	}
	if plan.RestoreSnapshot != nil {
		if err := add(plan.Target.TargetKind, record.RecordID, plan.RestoreSnapshot); err != nil {
			return nil, err
		}
	} else {
		for _, step := range plan.ApplyOrder {
			dispatch, err := a.targetSemantics.dispatchClass(step.Target.TargetKind)
			if err != nil {
				continue
			}
			switch dispatch {
			case rollbackcontract.DispatchRow:
				recordID, err := uuid.Parse(step.Target.TargetID)
				if err != nil {
					return nil, ErrRollbackTargetNotFound
				}
				if err := add(step.Target.TargetKind, recordID, step.Target.BeforeValue); err != nil {
					return nil, err
				}
			case rollbackcontract.DispatchNonRow:
				provider, err := a.targetSemantics.nonRowProvider(step.Target.TargetKind)
				if err != nil {
					return nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
				}
				claimProvider, ok := provider.(rollbackcontract.IdentifierClaimNonRowProvider)
				if !ok {
					continue
				}
				recordID, err := claimProvider.IdentifierClaimRecordTx(ctx, tx, nonRowContractTarget(record.IncidentID, step.Target))
				if err != nil {
					return nil, adaptRowRollbackProviderError(err)
				}
				if err := add("record", recordID, nil); err != nil {
					return nil, err
				}
			}
		}
	}
	ordered := make([]identifierClaimPreparation, 0, len(preparations))
	for _, preparation := range preparations {
		ordered = append(ordered, preparation)
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].providerID < ordered[j].providerID
	})
	prepared := make([]preparedIdentifierClaimRestore, 0, len(ordered))
	affected := canonicalRecordIDs(plan.Affected)
	for _, preparation := range ordered {
		recordIDs := make([]uuid.UUID, 0, len(preparation.records))
		for recordID := range preparation.records {
			recordIDs = append(recordIDs, recordID)
		}
		recordIDs = canonicalRecordIDs(recordIDs)
		records := make([]rollbackcontract.IdentifierClaimRestoreRecord, 0, len(recordIDs))
		for _, recordID := range recordIDs {
			records = append(records, rollbackcontract.IdentifierClaimRestoreRecord{
				RecordID:      recordID,
				RetainedValue: preparation.records[recordID],
			})
		}
		err := preparation.provider.PrepareIdentifierClaimRestoreTx(ctx, tx, rollbackcontract.IdentifierClaimRestoreRequest{
			IncidentID:        record.IncidentID,
			AffectedRecordIDs: affected,
			Records:           records,
		})
		if err != nil {
			return nil, adaptRowRollbackProviderError(err)
		}
		prepared = append(prepared, preparedIdentifierClaimRestore{
			recordIDs: recordIDs,
			provider:  preparation.provider,
		})
	}
	return prepared, nil
}

func finalizeIdentifierClaimRestoresTx(ctx context.Context, tx pgx.Tx, prepared []preparedIdentifierClaimRestore) error {
	if len(prepared) == 0 {
		return nil
	}
	if _, err := tx.Exec(ctx, `SELECT pg_catalog.set_config('cartulary.entities_defer_active_identifier_claims', 'off', true)`); err != nil {
		return err
	}
	for _, preparation := range prepared {
		if err := preparation.provider.FinalizeIdentifierClaimRestoreTx(ctx, tx, preparation.recordIDs); err != nil {
			return adaptRowRollbackProviderError(err)
		}
	}
	return nil
}
