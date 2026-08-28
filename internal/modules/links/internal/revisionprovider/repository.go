package revisionprovider

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/vocabulary"
)

func loadRecordLinkMutationValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (valuecodec.RecordLinkMutationValue, error) {
	var (
		input           valuecodec.RecordLinkMutationInput
		linkType        string
		fieldKey        pgtype.Text
		provenance      string
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := tx.QueryRow(ctx, `
SELECT
    record_link_id, incident_id, src_record_id, dst_record_id, link_type,
    field_key, provenance, confidence, owner_user_id, created_by_user_id,
    decided_at, created_at, deleted_at, deleted_by_user_id
  FROM record_links
 WHERE record_link_id = $1
`, recordLinkID).Scan(
		&input.RecordLinkID,
		&input.IncidentID,
		&input.SrcRecordID,
		&input.DstRecordID,
		&linkType,
		&fieldKey,
		&provenance,
		&confidence,
		&input.OwnerUserID,
		&input.CreatedByUserID,
		&input.DecidedAt,
		&input.CreatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return valuecodec.RecordLinkMutationValue{}, err
	}
	parsedLinkType, err := vocabulary.ParseLinkType(linkType)
	if err != nil {
		return valuecodec.RecordLinkMutationValue{}, fmt.Errorf("load record link mutation value: %w", err)
	}
	parsedProvenance, err := vocabulary.ParseLinkProvenance(provenance)
	if err != nil {
		return valuecodec.RecordLinkMutationValue{}, fmt.Errorf("load record link mutation value: %w", err)
	}
	input.LinkType = parsedLinkType.String()
	input.Provenance = parsedProvenance.String()
	if fieldKey.Valid {
		input.FieldKey = &fieldKey.String
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		input.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		input.DeletedAt = &value
	}
	if deletedByUserID.Valid {
		value := uuid.UUID(deletedByUserID.Bytes)
		input.DeletedByUserID = &value
	}
	return valuecodec.BuildRecordLinkMutationValue(input), nil
}

func loadRecordTagMutationValueTx(ctx context.Context, tx pgx.Tx, recordTagID uuid.UUID) (valuecodec.RecordTagMutationValue, error) {
	var (
		input           valuecodec.RecordTagMutationInput
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := tx.QueryRow(ctx, `
SELECT
    record_tag_id, incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at, deleted_at, deleted_by_user_id
  FROM record_tags
 WHERE record_tag_id = $1
`, recordTagID).Scan(
		&input.RecordTagID,
		&input.IncidentID,
		&input.RecordID,
		&input.TagName,
		&input.NormalizedTagName,
		&input.CreatedByUserID,
		&input.CreatedAt,
		&input.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return valuecodec.RecordTagMutationValue{}, err
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		input.DeletedAt = &value
	}
	if deletedByUserID.Valid {
		value := uuid.UUID(deletedByUserID.Bytes)
		input.DeletedByUserID = &value
	}
	return valuecodec.BuildRecordTagMutationValue(input), nil
}
