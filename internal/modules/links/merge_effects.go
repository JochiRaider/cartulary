package links

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MergeMutation struct {
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type RepointMergedLinksCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedLinksResult struct {
	Mutations                 []MergeMutation
	RepointedCount            int
	DedupedCount              int
	LinkTypesBySourceRecordID map[uuid.UUID][]string
}

type RepointMergedTagsCommand struct {
	IncidentID       uuid.UUID
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedTagsResult struct {
	Mutations      []MergeMutation
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
	LinkType        string
	FieldKey        *string
	Provenance      string
	Confidence      *int
	OwnerUserID     uuid.UUID
	CreatedByUserID uuid.UUID
	DecidedAt       time.Time
	CreatedAt       time.Time
	DeletedAt       *time.Time
	DeletedByUserID *uuid.UUID
}

func (s *Store) RepointMergedLinksTx(ctx context.Context, tx pgx.Tx, command RepointMergedLinksCommand) (RepointMergedLinksResult, error) {
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
		return RepointMergedLinksResult{}, fmt.Errorf("load merged links: %w", err)
	}
	defer rows.Close()

	records := make([]mergeLinkRecord, 0)
	for rows.Next() {
		record, err := scanMergeLinkRecord(rows)
		if err != nil {
			return RepointMergedLinksResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointMergedLinksResult{}, fmt.Errorf("iterate merged links: %w", err)
	}
	rows.Close()

	result := RepointMergedLinksResult{
		Mutations:                 make([]MergeMutation, 0),
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
		addMergeLinkInvalidation(&result, record.SrcRecordID, record.LinkType)
		if nextSrc != record.SrcRecordID {
			addMergeLinkInvalidation(&result, nextSrc, record.LinkType)
		}
		if nextSrc == nextDst {
			tombstoned, err := s.TombstoneLinkTx(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointMergedLinksResult{}, err
			}
			record.DeletedAt = tombstoned.DeletedAt
			record.DeletedByUserID = &command.ActorUserID
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      record.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(record),
			})
			result.DedupedCount++
			continue
		}
		exists, err := s.activeMergeLinkExistsTx(ctx, tx, command.IncidentID, nextSrc, nextDst, record.LinkType, record.FieldKey)
		switch {
		case err == nil && !exists:
			tombstoned, err := s.TombstoneLinkTx(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointMergedLinksResult{}, fmt.Errorf("tombstone merged link before repoint: %w", err)
			}
			record.DeletedAt = tombstoned.DeletedAt
			record.DeletedByUserID = &command.ActorUserID
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      tombstoned.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(record),
			})
			created, err := s.insertRepointedMergeLinkTx(ctx, tx, record, nextSrc, nextDst)
			if err != nil {
				return RepointMergedLinksResult{}, fmt.Errorf("create repointed merged link: %w", err)
			}
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      created.RecordLinkID.String(),
				OperationKind: "create",
				AfterValue:    buildMergeLinkValue(created),
			})
			result.RepointedCount++
		case err != nil:
			return RepointMergedLinksResult{}, err
		default:
			tombstoned, err := s.TombstoneLinkTx(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointMergedLinksResult{}, err
			}
			deletedRecord := mergeLinkRecordWithDeletedAt(record, tombstoned.DeletedAt)
			deletedRecord.DeletedByUserID = &command.ActorUserID
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      record.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(deletedRecord),
			})
			result.DedupedCount++
		}
	}
	return result, nil
}

func (s *Store) activeMergeLinkExistsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, linkType string, fieldKey *string) (bool, error) {
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
`, incidentID, src, dst, linkType, fieldKey).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query active merged link: %w", err)
	}
	return true, nil
}

func (s *Store) insertRepointedMergeLinkTx(ctx context.Context, tx pgx.Tx, record mergeLinkRecord, src uuid.UUID, dst uuid.UUID) (mergeLinkRecord, error) {
	if err := validateRecordLinkCommand(record.LinkType, record.Provenance, record.Confidence, src, dst); err != nil {
		return mergeLinkRecord{}, err
	}
	if err := validateActiveLinkEndpointsTx(ctx, tx, record.IncidentID, src, dst); err != nil {
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
`, record.IncidentID, src, dst, record.LinkType, record.FieldKey, record.Provenance, record.Confidence, record.OwnerUserID, record.DecidedAt)
	created, err := scanMergeLinkRecord(row)
	if err != nil {
		return mergeLinkRecord{}, err
	}
	return created, nil
}

func addMergeLinkInvalidation(result *RepointMergedLinksResult, sourceRecordID uuid.UUID, linkType string) {
	result.LinkTypesBySourceRecordID[sourceRecordID] = append(result.LinkTypesBySourceRecordID[sourceRecordID], linkType)
}

func (s *Store) RepointMergedTagsTx(ctx context.Context, tx pgx.Tx, command RepointMergedTagsCommand) (RepointMergedTagsResult, error) {
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
		return RepointMergedTagsResult{}, fmt.Errorf("load merged tags: %w", err)
	}
	defer rows.Close()

	records := make([]mergeTagRecord, 0)
	for rows.Next() {
		record, err := scanMergeTagRecord(rows)
		if err != nil {
			return RepointMergedTagsResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointMergedTagsResult{}, fmt.Errorf("iterate merged tags: %w", err)
	}
	rows.Close()

	result := RepointMergedTagsResult{Mutations: make([]MergeMutation, 0)}
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
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET record_id = $2,
       updated_at = $3
 WHERE record_tag_id = $1
`, record.RecordTagID, command.SurvivorRecordID, command.Now.UTC()); err != nil {
				return RepointMergedTagsResult{}, fmt.Errorf("repoint merged tag: %w", err)
			}
			record.RecordID = command.SurvivorRecordID
			record.UpdatedAt = command.Now.UTC()
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_tag",
				TargetID:      record.RecordTagID.String(),
				OperationKind: "patch",
				BeforeValue:   before,
				AfterValue:    buildMergeTagValue(record),
			})
			result.RepointedCount++
		case err != nil:
			return RepointMergedTagsResult{}, fmt.Errorf("lookup survivor tag collision: %w", err)
		default:
			if _, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = COALESCE(deleted_at, $2),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $3),
       updated_at = $2
 WHERE record_tag_id = $1
`, record.RecordTagID, command.Now.UTC(), command.ActorUserID); err != nil {
				return RepointMergedTagsResult{}, fmt.Errorf("dedupe merged tag: %w", err)
			}
			record.DeletedAt = timePointer(command.Now.UTC())
			record.DeletedByUserID = &command.ActorUserID
			record.UpdatedAt = command.Now.UTC()
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_tag",
				TargetID:      record.RecordTagID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeTagValue(record),
			})
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
		fieldKey        pgtype.Text
		confidence      pgtype.Int4
		deletedAt       pgtype.Timestamptz
		deletedByUserID pgtype.UUID
	)
	if err := row.Scan(
		&record.RecordLinkID,
		&record.IncidentID,
		&record.SrcRecordID,
		&record.DstRecordID,
		&record.LinkType,
		&fieldKey,
		&record.Provenance,
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
	value := map[string]any{
		"record_link_id":     record.RecordLinkID.String(),
		"incident_id":        record.IncidentID.String(),
		"src_record_id":      record.SrcRecordID.String(),
		"dst_record_id":      record.DstRecordID.String(),
		"link_type":          record.LinkType,
		"field_key":          nil,
		"provenance":         record.Provenance,
		"confidence":         record.Confidence,
		"owner_user_id":      record.OwnerUserID.String(),
		"created_by_user_id": record.CreatedByUserID.String(),
		"decided_at":         record.DecidedAt.UTC().Format(time.RFC3339Nano),
		"created_at":         record.CreatedAt.UTC().Format(time.RFC3339Nano),
		"deleted_at":         formatMutationTimestampPointer(record.DeletedAt),
		"deleted_by_user_id": formatMutationUUIDPointer(record.DeletedByUserID),
	}
	if record.FieldKey != nil {
		value["field_key"] = *record.FieldKey
	}
	return value
}

func mergeLinkRecordWithDeletedAt(record mergeLinkRecord, deletedAt *time.Time) mergeLinkRecord {
	record.DeletedAt = deletedAt
	return record
}

func buildMergeTagValue(record mergeTagRecord) map[string]any {
	return map[string]any{
		"record_tag_id":       record.RecordTagID.String(),
		"incident_id":         record.IncidentID.String(),
		"record_id":           record.RecordID.String(),
		"tag_name":            record.TagName,
		"normalized_tag_name": record.NormalizedTagName,
		"created_by_user_id":  record.CreatedByUserID.String(),
		"created_at":          record.CreatedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":          record.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"deleted_at":          formatMutationTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":  formatMutationUUIDPointer(record.DeletedByUserID),
	}
}

func timePointer(value time.Time) *time.Time {
	copy := value.UTC()
	return &copy
}

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}
