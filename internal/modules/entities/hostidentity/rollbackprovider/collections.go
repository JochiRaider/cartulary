package rollbackprovider

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/aliashistory"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/preservedhistory"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/records/enveloperead"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type CollectionProvider struct{}

var _ rollbackcontract.NonRowTargetProvider = CollectionProvider{}
var _ rollbackcontract.IdentifierClaimNonRowProvider = CollectionProvider{}

func NewCollectionProvider() CollectionProvider { return CollectionProvider{} }

func (CollectionProvider) IdentifierClaimRecordTx(_ context.Context, _ pgx.Tx, target rollbackcontract.NonRowTarget) (uuid.UUID, error) {
	identity, err := parseCollectionTarget(target)
	if err != nil {
		return uuid.Nil, err
	}
	if identity.targetKind != "entity_preserved_identifier" || identity.classification != "exact_match_reuse" {
		return uuid.Nil, nil
	}
	return identity.recordID, nil
}

type collectionIdentity struct {
	targetKind                  string
	rowID, incidentID, recordID uuid.UUID
	entityType, classification  string
}

func (p CollectionProvider) DescribeTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.DescribeRequest) (rollbackcontract.TargetDescriptor, error) {
	identity, err := parseCollectionTarget(request.Target)
	if err != nil {
		return rollbackcontract.TargetDescriptor{}, err
	}
	descriptor := rollbackcontract.TargetDescriptor{
		AffectedRecordIDs:      []uuid.UUID{identity.recordID},
		RequiresWholeChangeSet: true,
	}
	if request.Target.IncidentID != identity.incidentID {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	envelope, err := records.NewStore().LoadEnvelopeTx(ctx, tx, identity.recordID, false)
	if err != nil {
		if errors.Is(err, enveloperead.ErrEnvelopeNotFound) {
			return descriptor, rollbackcontract.ErrTargetNotFound
		}
		return descriptor, err
	}
	if envelope.IncidentID != identity.incidentID || envelope.RecordType != identity.entityType {
		return descriptor, rollbackcontract.ErrTargetNotFound
	}
	current, err := loadCollectionValueByIDTx(ctx, tx, identity.targetKind, identity.rowID)
	if err != nil {
		return descriptor, err
	}
	if !sameCompleteCollectionState(identity.targetKind, current, request.Target.AfterValue) {
		return descriptor, rollbackcontract.ErrStaleTarget
	}

	return descriptor, nil
}

func sameCompleteCollectionState(kind string, current, expected map[string]any) bool {
	switch kind {
	case "entity_alias":
		left, leftErr := aliashistory.Decode(current)
		right, rightErr := aliashistory.Decode(expected)
		return leftErr == nil && rightErr == nil && reflect.DeepEqual(left, right)
	case "entity_preserved_identifier":
		left, leftErr := preservedhistory.Decode(current)
		right, rightErr := preservedhistory.Decode(expected)
		return leftErr == nil && rightErr == nil && reflect.DeepEqual(left, right)
	default:
		return false
	}
}

func (p CollectionProvider) ApplyInverseTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.ApplyInverseRequest) (rollbackcontract.ApplyInverseResult, error) {
	descriptor, err := p.DescribeTx(ctx, tx, rollbackcontract.DescribeRequest{Target: request.Target})
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	identity, err := parseCollectionTarget(request.Target)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	rowID := identity.rowID
	before, err := loadCollectionValueByIDTx(ctx, tx, identity.targetKind, rowID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	switch request.Target.OperationKind {
	case "create":
		err = tombstoneCollectionTx(ctx, tx, identity.targetKind, rowID, request.Now)
	case "delete":
		err = restoreAliasCollectionTx(ctx, tx, rowID)
	default:
		err = rollbackcontract.ErrTargetNotReversible
	}
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	after, err := loadCollectionValueByIDTx(ctx, tx, identity.targetKind, rowID)
	if err != nil {
		return rollbackcontract.ApplyInverseResult{}, err
	}
	return rollbackcontract.ApplyInverseResult{
		AffectedRecordIDs: descriptor.AffectedRecordIDs,
		BeforeValue:       before,
		AfterValue:        after,
	}, nil
}

func parseCollectionTarget(target rollbackcontract.NonRowTarget) (collectionIdentity, error) {
	if target.TargetKind == "entity_preserved_identifier" {
		return parsePreservedTarget(target)
	}
	if target.TargetKind != "entity_alias" || (target.OperationKind != "create" && target.OperationKind != "delete") ||
		aliashistory.ValidateMutation(target.TargetID, target.OperationKind, target.BeforeValue, target.AfterValue) != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	value, err := aliashistory.Decode(target.AfterValue)
	if err != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	return collectionIdentity{targetKind: target.TargetKind, rowID: value.ID, incidentID: value.IncidentID, recordID: value.RecordID,
		entityType: value.EntityType, classification: value.Classification}, nil
}

func restoreAliasCollectionTx(ctx context.Context, tx pgx.Tx, rowID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at = NULL WHERE entity_alias_id = $1 AND deleted_at IS NOT NULL`, rowID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func tombstoneCollectionTx(ctx context.Context, tx pgx.Tx, targetKind string, rowID uuid.UUID, now time.Time) error {
	var rowsAffected int64
	switch targetKind {
	case "entity_alias":
		tag, err := tx.Exec(ctx, `UPDATE entity_aliases SET deleted_at = $2 WHERE entity_alias_id = $1 AND deleted_at IS NULL`, rowID, now.UTC())
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	case "entity_preserved_identifier":
		tag, err := tx.Exec(ctx, `UPDATE entity_preserved_identifiers SET deleted_at = $2 WHERE entity_preserved_identifier_id = $1 AND deleted_at IS NULL`, rowID, now.UTC())
		if err != nil {
			return err
		}
		rowsAffected = tag.RowsAffected()
	default:
		return rollbackcontract.ErrTargetNotReversible
	}
	if rowsAffected != 1 {
		return rollbackcontract.ErrStaleTarget
	}
	return nil
}

func loadCollectionValueByIDTx(ctx context.Context, tx pgx.Tx, targetKind string, rowID uuid.UUID) (map[string]any, error) {
	switch targetKind {
	case "entity_alias":
		return loadAliasValueTx(ctx, tx, rowID)
	case "entity_preserved_identifier":
		return loadPreservedValueTx(ctx, tx, rowID)
	default:
		return nil, rollbackcontract.ErrTargetNotReversible
	}
}

func parsePreservedTarget(target rollbackcontract.NonRowTarget) (collectionIdentity, error) {
	if target.OperationKind != "create" || preservedhistory.ValidateMutation(target.TargetID, target.OperationKind, target.BeforeValue, target.AfterValue) != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	value, err := preservedhistory.Decode(target.AfterValue)
	if err != nil {
		return collectionIdentity{}, rollbackcontract.ErrTargetNotReversible
	}
	return collectionIdentity{targetKind: target.TargetKind, rowID: value.ID, incidentID: value.IncidentID, recordID: value.RecordID,
		entityType: value.EntityType, classification: value.Classification}, nil
}

func loadPreservedValueTx(ctx context.Context, tx pgx.Tx, rowID uuid.UUID) (map[string]any, error) {
	var value preservedhistory.Value
	if err := tx.QueryRow(ctx, `
SELECT entity_preserved_identifier_id,incident_id,record_id,entity_type,identifier_type,
       raw_value,normalized_value,classification,created_by_user_id,created_at,deleted_at
FROM entity_preserved_identifiers WHERE entity_preserved_identifier_id=$1`, rowID).Scan(
		&value.ID, &value.IncidentID, &value.RecordID, &value.EntityType, &value.IdentifierType,
		&value.RawValue, &value.NormalizedValue, &value.Classification, &value.CreatedByUserID, &value.CreatedAt, &value.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rollbackcontract.ErrTargetNotFound
		}
		return nil, err
	}
	encoded := value.Map()
	if _, err := preservedhistory.Decode(encoded); err != nil {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	return encoded, nil
}

func loadAliasValueTx(ctx context.Context, tx pgx.Tx, rowID uuid.UUID) (map[string]any, error) {
	var value aliashistory.Value
	if err := tx.QueryRow(ctx, `
SELECT entity_alias_id,incident_id,record_id,entity_type,raw_text,normalized_text::text,
       classification,created_by_user_id,created_at,deleted_at
FROM entity_aliases WHERE entity_alias_id=$1`, rowID).Scan(
		&value.ID, &value.IncidentID, &value.RecordID, &value.EntityType, &value.RawText, &value.NormalizedText,
		&value.Classification, &value.CreatedByUserID, &value.CreatedAt, &value.DeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, rollbackcontract.ErrTargetNotFound
		}
		return nil, err
	}
	encoded := value.Map()
	if _, err := aliashistory.Decode(encoded); err != nil {
		return nil, rollbackcontract.ErrTargetNotReversible
	}
	return encoded, nil
}
