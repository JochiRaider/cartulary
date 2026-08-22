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
