package revisions

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func (r rollbackQueryRepository) affectedRecordsForRollbackTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, target rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	descriptor, describeErr := r.describeRollbackTargetTx(ctx, tx, incidentID, target, fallback, nil)
	affected, _, err := rollbackDescriptorAffected(descriptor, describeErr, fallback)
	return affected, err
}

func rollbackDescriptorAffected(descriptor rollbackcontract.TargetDescriptor, describeErr error, fallback uuid.UUID) ([]uuid.UUID, error, error) {
	if describeErr != nil {
		adapted := adaptRowRollbackProviderError(describeErr)
		if !deferableRollbackProviderError(adapted) {
			return nil, nil, adapted
		}
		affected := canonicalRecordIDs(descriptor.AffectedRecordIDs)
		if len(affected) == 0 && fallback != uuid.Nil {
			affected = []uuid.UUID{fallback}
		}
		return affected, adapted, nil
	}
	affected := canonicalRecordIDs(descriptor.AffectedRecordIDs)
	if len(affected) == 0 {
		return nil, nil, &RollbackPreconditionError{ReasonCode: "target_not_reversible"}
	}
	return affected, nil, nil
}

func (r rollbackQueryRepository) affectedRecordsForRollbackTargetsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, targets []rollbackMutationTarget, fallback uuid.UUID) ([]uuid.UUID, error) {
	recordIDs := map[uuid.UUID]struct{}{fallback: {}}
	for _, target := range targets {
		affected, err := r.affectedRecordsForRollbackTargetTx(ctx, tx, incidentID, target, fallback)
		if err != nil {
			return nil, err
		}
		for _, recordID := range affected {
			recordIDs[recordID] = struct{}{}
		}
	}
	values := make([]uuid.UUID, 0, len(recordIDs))
	for recordID := range recordIDs {
		if recordID != uuid.Nil {
			values = append(values, recordID)
		}
	}
	return canonicalRecordIDs(values), nil
}
