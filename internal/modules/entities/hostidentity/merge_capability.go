package hostidentity

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/preservedhistory"
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

// MergeIdentifierClaimPreparation binds the discovered claim set to the two
// source identities. Consumers cannot replace or extend its protected tuples.
type MergeIdentifierClaimPreparation struct {
	incidentID uuid.UUID
	entityType string
	survivorID uuid.UUID
	loserID    uuid.UUID
	tuples     []normalizedIdentifierTuple
	Conflict   *ActiveIdentifierTransitionConflict
}

func (*MergeCapability) PrepareIdentifierClaimsTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	entityType string,
	survivorRecordID uuid.UUID,
	loserRecordID uuid.UUID,
) (*MergeIdentifierClaimPreparation, error) {
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
	prepared := &MergeIdentifierClaimPreparation{incidentID: incidentID, entityType: entityType, survivorID: survivorRecordID, loserID: loserRecordID, tuples: tuples}
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
			prepared.Conflict = &ActiveIdentifierTransitionConflict{
				IdentifierClass:  tuple.IdentifierType,
				NormalizedValue:  tuple.NormalizedValue,
				BlockingRecordID: claimedRecordID,
			}
			return prepared, nil
		}
	}
	return prepared, nil
}

// RevalidateIdentifierClaimsTx only reads. The caller already holds both source
// envelopes and typed rows; acquiring new claims here would reverse lock order.
func (*MergeCapability) RevalidateIdentifierClaimsTx(ctx context.Context, tx pgx.Tx, prepared *MergeIdentifierClaimPreparation) (bool, error) {
	if prepared == nil {
		return false, fmt.Errorf("missing Merge identifier claim preparation")
	}
	survivor, err := recordIdentifierTuplesTx(ctx, tx, prepared.incidentID, prepared.entityType, prepared.survivorID)
	if err != nil {
		return false, err
	}
	loser, err := recordIdentifierTuplesTx(ctx, tx, prepared.incidentID, prepared.entityType, prepared.loserID)
	if err != nil {
		return false, err
	}
	return slices.Equal(prepared.tuples, mergeNormalizedIdentifierTuples(survivor, loser)), nil
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
) (*preservedhistory.Value, error) {
	values, err := syncPreservedIdentifiersTx(ctx, tx, incidentID, recordID, entityType, []identifierSeed{{
		IdentifierType: identifierType,
		RawValue:       rawValue,
		Classification: classification,
	}}, actorUserID, now)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	return &values[0], nil
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
