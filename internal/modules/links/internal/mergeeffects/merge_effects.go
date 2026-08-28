package mergeeffects

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/links/internal/mutationvalue"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/valuecodec"
	"github.com/JochiRaider/cartulary/internal/modules/links/internal/vocabulary"
)

type RepointLinksCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointLinksResult struct {
	Mutations                 []mutationvalue.Value
	RepointedCount            int
	DedupedCount              int
	LinkTypesBySourceRecordID map[uuid.UUID][]string
}

type RepointTagsCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointTagsResult struct {
	Mutations      []mutationvalue.Value
	RepointedCount int
	DedupedCount   int
}

type mergeTagRecord struct {
	RecordTagID       uuid.UUID
	IncidentID        uuid.UUID
	RecordID          uuid.UUID
	TagName           string
	NormalizedTagName string
	CreatedByUserID   uuid.UUID
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
	DeletedByUserID   *uuid.UUID
}

type mergeLinkRecord struct {
	RecordLinkID    uuid.UUID
	IncidentID      uuid.UUID
	SrcRecordID     uuid.UUID
	DstRecordID     uuid.UUID
	LinkType        vocabulary.LinkType
	FieldKey        *string
	Provenance      vocabulary.LinkProvenance
	Confidence      *int
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

type LinkDependencies struct {
	Validate  func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, vocabulary.LinkType, vocabulary.LinkProvenance, *int) error
	Tombstone func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (*time.Time, error)
}

func RepointLinksTx(ctx context.Context, tx pgx.Tx, command RepointLinksCommand, dependencies LinkDependencies) (RepointLinksResult, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at,
    deleted_at,
    deleted_by_user_id
  FROM record_links
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND (src_record_id = $2 OR dst_record_id = $2)
 ORDER BY src_record_id ASC, dst_record_id ASC, link_type ASC, COALESCE(field_key, '') ASC, record_link_id ASC
 FOR UPDATE
`, command.IncidentID, command.LoserRecordID)
	if err != nil {
		return RepointLinksResult{}, fmt.Errorf("load merged links: %w", err)
	}
	defer rows.Close()

	records := make([]mergeLinkRecord, 0)
	for rows.Next() {
		record, err := scanMergeLinkRecord(rows)
		if err != nil {
			return RepointLinksResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointLinksResult{}, fmt.Errorf("iterate merged links: %w", err)
	}
	rows.Close()

	result := RepointLinksResult{
		Mutations:                 make([]mutationvalue.Value, 0),
		LinkTypesBySourceRecordID: make(map[uuid.UUID][]string),
	}
	for _, record := range records {
		if record.SrcRecordID != command.LoserRecordID && record.DstRecordID != command.LoserRecordID {
			continue
		}
		before := buildMergeLinkValue(record)
		nextSrc := record.SrcRecordID
		nextDst := record.DstRecordID
		if nextSrc == command.LoserRecordID {
			nextSrc = command.SurvivorRecordID
		}
		if nextDst == command.LoserRecordID {
			nextDst = command.SurvivorRecordID
		}
		addMergeLinkInvalidation(&result, record.SrcRecordID, record.LinkType.String())
		if nextSrc != record.SrcRecordID {
			addMergeLinkInvalidation(&result, nextSrc, record.LinkType.String())
		}
		if nextSrc == nextDst {
			deletedAt, err := dependencies.Tombstone(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointLinksResult{}, err
			}
			record.DeletedAt = deletedAt
			record.DeletedByUserID = &command.ActorUserID
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordLink, record.RecordLinkID.String(), mutationvalue.OperationDelete, before, buildMergeLinkValue(record)); err != nil {
				return RepointLinksResult{}, err
			}
			result.DedupedCount++
			continue
		}
		exists, err := activeMergeLinkExistsTx(ctx, tx, command.IncidentID, nextSrc, nextDst, record.LinkType, record.FieldKey)
		switch {
		case err == nil && !exists:
			deletedAt, err := dependencies.Tombstone(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointLinksResult{}, fmt.Errorf("tombstone merged link before repoint: %w", err)
			}
			record.DeletedAt = deletedAt
			record.DeletedByUserID = &command.ActorUserID
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordLink, record.RecordLinkID.String(), mutationvalue.OperationDelete, before, buildMergeLinkValue(record)); err != nil {
				return RepointLinksResult{}, err
			}
			created, err := insertRepointedMergeLinkTx(ctx, tx, record, nextSrc, nextDst, dependencies.Validate)
			if err != nil {
				return RepointLinksResult{}, fmt.Errorf("create repointed merged link: %w", err)
			}
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordLink, created.RecordLinkID.String(), mutationvalue.OperationCreate, nil, buildMergeLinkValue(created)); err != nil {
				return RepointLinksResult{}, err
			}
			result.RepointedCount++
		case err != nil:
			return RepointLinksResult{}, err
		default:
			deletedAt, err := dependencies.Tombstone(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointLinksResult{}, err
			}
			deletedRecord := mergeLinkRecordWithDeletedAt(record, deletedAt)
			deletedRecord.DeletedByUserID = &command.ActorUserID
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordLink, record.RecordLinkID.String(), mutationvalue.OperationDelete, before, buildMergeLinkValue(deletedRecord)); err != nil {
				return RepointLinksResult{}, err
			}
			result.DedupedCount++
		}
	}
	return result, nil
}

func activeMergeLinkExistsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, linkType vocabulary.LinkType, fieldKey *string) (bool, error) {
	var linkID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND field_key IS NOT DISTINCT FROM $5
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, src, dst, linkType.String(), fieldKey).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query active merged link: %w", err)
	}
	return true, nil
}

func insertRepointedMergeLinkTx(
	ctx context.Context,
	tx pgx.Tx,
	record mergeLinkRecord,
	src uuid.UUID,
	dst uuid.UUID,
	validate func(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, vocabulary.LinkType, vocabulary.LinkProvenance, *int) error,
) (mergeLinkRecord, error) {
	if err := validate(ctx, tx, record.IncidentID, src, dst, record.LinkType, record.Provenance, record.Confidence); err != nil {
		return mergeLinkRecord{}, err
	}
	row := tx.QueryRow(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9, $9)
RETURNING
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at,
    deleted_at,
    deleted_by_user_id
`, record.IncidentID, src, dst, record.LinkType.String(), record.FieldKey, record.Provenance.String(), record.Confidence, record.OwnerUserID, record.DecidedAt)
	created, err := scanMergeLinkRecord(row)
	if err != nil {
		return mergeLinkRecord{}, err
	}
	return created, nil
}

func addMergeLinkInvalidation(result *RepointLinksResult, sourceRecordID uuid.UUID, linkType string) {
	result.LinkTypesBySourceRecordID[sourceRecordID] = append(result.LinkTypesBySourceRecordID[sourceRecordID], linkType)
}

func RepointTagsTx(ctx context.Context, tx pgx.Tx, command RepointTagsCommand) (RepointTagsResult, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_tag_id,
    incident_id,
    record_id,
    tag_name,
    normalized_tag_name,
    created_by_user_id,
    created_at,
    updated_at,
    deleted_at,
    deleted_by_user_id
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
 FOR UPDATE
`, command.IncidentID, command.LoserRecordID)
	if err != nil {
		return RepointTagsResult{}, fmt.Errorf("load merged tags: %w", err)
	}
	defer rows.Close()

	records := make([]mergeTagRecord, 0)
	for rows.Next() {
		record, err := scanMergeTagRecord(rows)
		if err != nil {
			return RepointTagsResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointTagsResult{}, fmt.Errorf("iterate merged tags: %w", err)
	}
	rows.Close()

	result := RepointTagsResult{Mutations: make([]mutationvalue.Value, 0)}
	for _, record := range records {
		before := buildMergeTagValue(record)
		var existingID uuid.UUID
		err = tx.QueryRow(ctx, `
SELECT record_tag_id
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = $3
   AND deleted_at IS NULL
 LIMIT 1
`, command.IncidentID, command.SurvivorRecordID, record.NormalizedTagName).Scan(&existingID)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			targetID := recordTagTargetID(record.RecordID, record.RecordTagID)
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET record_id = $2,
       updated_at = $3
 WHERE record_tag_id = $1
`, record.RecordTagID, command.SurvivorRecordID, command.Now.UTC()); err != nil {
				return RepointTagsResult{}, fmt.Errorf("repoint merged tag: %w", err)
			}
			record.RecordID = command.SurvivorRecordID
			record.UpdatedAt = command.Now.UTC()
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordTag, targetID, mutationvalue.OperationPatch, before, buildMergeTagValue(record)); err != nil {
				return RepointTagsResult{}, err
			}
			result.RepointedCount++
		case err != nil:
			return RepointTagsResult{}, fmt.Errorf("lookup survivor tag collision: %w", err)
		default:
			targetID := recordTagTargetID(record.RecordID, record.RecordTagID)
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = COALESCE(deleted_at, $2),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $3),
       updated_at = $2
 WHERE record_tag_id = $1
`, record.RecordTagID, command.Now.UTC(), command.ActorUserID); err != nil {
				return RepointTagsResult{}, fmt.Errorf("dedupe merged tag: %w", err)
			}
			record.DeletedAt = timePointer(command.Now.UTC())
			record.DeletedByUserID = &command.ActorUserID
			record.UpdatedAt = command.Now.UTC()
			if err := appendMutation(&result.Mutations, mutationvalue.TargetRecordTag, targetID, mutationvalue.OperationDelete, before, buildMergeTagValue(record)); err != nil {
				return RepointTagsResult{}, err
			}
			result.DedupedCount++
			_ = existingID
		}
	}
	return result, nil
}

func scanMergeTagRecord(row pgx.Row) (mergeTagRecord, error) {
	var (
		record          mergeTagRecord
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&record.RecordTagID,
		&record.IncidentID,
		&record.RecordID,
		&record.TagName,
		&record.NormalizedTagName,
		&record.CreatedByUserID,
		&record.CreatedAt,
		&record.UpdatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return mergeTagRecord{}, fmt.Errorf("scan merge tag record: %w", err)
	}
	record.CreatedAt = record.CreatedAt.UTC()
	record.UpdatedAt = record.UpdatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DeletedByUserID = uuidPointerFromPG(deletedByUserID)
	return record, nil
}

func scanMergeLinkRecord(row pgx.Row) (mergeLinkRecord, error) {
	var (
		record          mergeLinkRecord
		linkType        string
		fieldKey        pgtype.Text
		provenance      string
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&record.RecordLinkID,
		&record.IncidentID,
		&record.SrcRecordID,
		&record.DstRecordID,
		&linkType,
		&fieldKey,
		&provenance,
		&confidence,
		&record.OwnerUserID,
		&record.CreatedByUserID,
		&record.DecidedAt,
		&record.CreatedAt,
		&deletedAt,
		&deletedByUserID,
	); err != nil {
		return mergeLinkRecord{}, err
	}
	parsedLinkType, err := vocabulary.ParseLinkType(linkType)
	if err != nil {
		return mergeLinkRecord{}, fmt.Errorf("scan merge link record: %w", err)
	}
	parsedProvenance, err := vocabulary.ParseLinkProvenance(provenance)
	if err != nil {
		return mergeLinkRecord{}, fmt.Errorf("scan merge link record: %w", err)
	}
	record.LinkType = parsedLinkType
	record.Provenance = parsedProvenance
	if fieldKey.Valid {
		value := fieldKey.String
		record.FieldKey = &value
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		record.Confidence = &value
	}
	record.DecidedAt = record.DecidedAt.UTC()
	record.CreatedAt = record.CreatedAt.UTC()
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		record.DeletedAt = &value
	}
	record.DeletedByUserID = uuidPointerFromPG(deletedByUserID)
	return record, nil
}

func buildMergeLinkValue(record mergeLinkRecord) map[string]any {
	return valuecodec.BuildRecordLinkMutationValue(valuecodec.RecordLinkMutationInput{
		RecordLinkID:    record.RecordLinkID,
		IncidentID:      record.IncidentID,
		SrcRecordID:     record.SrcRecordID,
		DstRecordID:     record.DstRecordID,
		LinkType:        record.LinkType.String(),
		FieldKey:        record.FieldKey,
		Provenance:      record.Provenance.String(),
		Confidence:      record.Confidence,
		OwnerUserID:     record.OwnerUserID,
		CreatedByUserID: record.CreatedByUserID,
		DecidedAt:       record.DecidedAt,
		CreatedAt:       record.CreatedAt,
		DeletedAt:       record.DeletedAt,
		DeletedByUserID: record.DeletedByUserID,
	}).Map()
}

func mergeLinkRecordWithDeletedAt(record mergeLinkRecord, deletedAt *time.Time) mergeLinkRecord {
	record.DeletedAt = deletedAt
	return record
}

func buildMergeTagValue(record mergeTagRecord) map[string]any {
	return valuecodec.BuildRecordTagMutationValue(valuecodec.RecordTagMutationInput{
		RecordTagID:       record.RecordTagID,
		IncidentID:        record.IncidentID,
		RecordID:          record.RecordID,
		TagName:           record.TagName,
		NormalizedTagName: record.NormalizedTagName,
		CreatedByUserID:   record.CreatedByUserID,
		CreatedAt:         record.CreatedAt,
		UpdatedAt:         record.UpdatedAt,
		DeletedAt:         record.DeletedAt,
		DeletedByUserID:   record.DeletedByUserID,
	}).Map()
}

func timePointer(value time.Time) *time.Time {
	copy := value.UTC()
	return &copy
}

func appendMutation(result *[]mutationvalue.Value, targetKind string, targetID string, operationKind string, beforeValue any, afterValue any) error {
	mutation, err := mutationvalue.New(targetKind, targetID, operationKind, beforeValue, afterValue)
	if err != nil {
		return err
	}
	*result = append(*result, mutation)
	return nil
}

func recordTagTargetID(recordID uuid.UUID, recordTagID uuid.UUID) string {
	return "record_tag:" + recordID.String() + ":" + recordTagID.String()
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}
