package hostidentity

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MergeCapability is the immutable Host/Identity source boundary used by the
// Entities merge consumer. It owns no transaction and carries no mutable
// dependency state.
type MergeCapability struct{}

func NewMergeCapability() *MergeCapability {
	return &MergeCapability{}
}

func (*MergeCapability) HostExactMatchPrecedence() []string {
	return append([]string(nil), hostExactMatchPrecedence...)
}

func (*MergeCapability) IdentityExactMatchPrecedence() []string {
	return append([]string(nil), identityExactMatchPrecedence...)
}

func (*MergeCapability) HostCanonicalNormalized(record HostRecord, identifierClass string) string {
	return hostCanonicalNormalized(record, identifierClass)
}

func (*MergeCapability) IdentityCanonicalNormalized(record IdentityRecord, identifierClass string) string {
	return identityCanonicalNormalized(record, identifierClass)
}

func (*MergeCapability) LoadHostTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (HostRecord, error) {
	return loadHostByRecordIDTx(ctx, tx, recordID)
}

func (*MergeCapability) LoadIdentityTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (IdentityRecord, error) {
	return loadIdentityByRecordIDTx(ctx, tx, recordID)
}

func (*MergeCapability) PrepareIdentifierClaimsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	entityType string,
	survivorRecordID uuid.UUID,
	loserRecordID uuid.UUID,
) (*ActiveIdentifierTransitionConflict, error) {
	survivorTuples, err := recordIdentifierTuplesTx(ctx, tx, incidentID, entityType, survivorRecordID)
	if err != nil {
		return nil, err
	}
	loserTuples, err := recordIdentifierTuplesTx(ctx, tx, incidentID, entityType, loserRecordID)
	if err != nil {
		return nil, err
	}
	tuples := mergeNormalizedIdentifierTuples(survivorTuples, loserTuples)
	if err := lockIdentifierTuplesTx(ctx, tx, incidentID, entityType, tuples); err != nil {
		return nil, err
	}
	for _, tuple := range tuples {
		var claimedRecordID uuid.UUID
		err := tx.QueryRow(ctx, `
SELECT record_id
  FROM entity_active_identifier_claims
 WHERE incident_id = $1
   AND entity_type = $2
   AND identifier_type = $3
   AND normalized_value = $4
`, incidentID, entityType, tuple.IdentifierType, tuple.NormalizedValue).Scan(&claimedRecordID)
		if err == pgx.ErrNoRows {
			continue
		}
		if err != nil {
			return nil, err
		}
		if claimedRecordID != survivorRecordID && claimedRecordID != loserRecordID {
			return &ActiveIdentifierTransitionConflict{
				IdentifierClass:  tuple.IdentifierType,
				NormalizedValue:  tuple.NormalizedValue,
				BlockingRecordID: claimedRecordID,
			}, nil
		}
	}
	return nil, nil
}

func (*MergeCapability) UpdateHostTx(ctx context.Context, tx pgx.Tx, record HostRecord) error {
	return updateHostTx(ctx, tx, record)
}

func (*MergeCapability) UpdateIdentityTx(ctx context.Context, tx pgx.Tx, record IdentityRecord) error {
	return updateIdentityTx(ctx, tx, record)
}

func (*MergeCapability) SyncPreservedIdentifierTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	entityType string,
	identifierType string,
	rawValue string,
	classification string,
	actorUserID uuid.UUID,
	now time.Time,
) (bool, error) {
	return syncPreservedIdentifiersTx(ctx, tx, incidentID, recordID, entityType, []identifierSeed{{
		IdentifierType: identifierType,
		RawValue:       rawValue,
		Classification: classification,
	}}, actorUserID, now)
}

func (*MergeCapability) SyncAliasesTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
	entityType string,
	actions []CollectionAction,
	actorUserID uuid.UUID,
	now time.Time,
) (AliasSyncResult, error) {
	return syncEntityAliasesTx(ctx, tx, incidentID, recordID, entityType, actions, actorUserID, now)
}
