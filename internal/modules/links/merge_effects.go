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
	Mutations             []MergeMutation
	RepointedCount        int
	DedupedCount          int
	TimelineInvalidations map[uuid.UUID][]string
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

func (s *Store) RepointMergedLinksTx(ctx context.Context, tx pgx.Tx, command RepointMergedLinksCommand) (RepointMergedLinksResult, error) {
	rows, err := tx.Query(ctx, `
SELECT
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    decided_at,
    created_at,
    deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND deleted_at IS NULL
   AND dst_record_id = $2
 ORDER BY src_record_id ASC, link_type ASC, record_link_id ASC
 FOR UPDATE
`, command.IncidentID, command.LoserRecordID)
	if err != nil {
		return RepointMergedLinksResult{}, fmt.Errorf("load merged links: %w", err)
	}
	defer rows.Close()

	records := make([]RecordLink, 0)
	for rows.Next() {
		record, err := scanRecordLink(rows)
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
		Mutations:             make([]MergeMutation, 0),
		TimelineInvalidations: make(map[uuid.UUID][]string),
	}
	for _, record := range records {
		if record.DstRecordID != command.LoserRecordID {
			continue
		}
		before := buildMergeLinkValue(record)
		_, err := s.GetActiveLinkTx(ctx, tx, command.IncidentID, record.SrcRecordID, command.SurvivorRecordID, record.LinkType)
		switch {
		case errors.Is(err, ErrRecordLinkNotFound):
			tombstoned, err := s.TombstoneLinkTx(ctx, tx, record.RecordLinkID, command.ActorUserID, command.Now.UTC())
			if err != nil {
				return RepointMergedLinksResult{}, fmt.Errorf("tombstone merged link before repoint: %w", err)
			}
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      tombstoned.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(tombstoned),
			})
			created, inserted, err := s.UpsertLinkTx(ctx, tx, command.IncidentID, record.SrcRecordID, command.SurvivorRecordID, record.LinkType, record.Provenance, record.Confidence, record.OwnerUserID, record.DecidedAt)
			if err != nil {
				return RepointMergedLinksResult{}, fmt.Errorf("create repointed merged link: %w", err)
			}
			if !inserted {
				return RepointMergedLinksResult{}, fmt.Errorf("create repointed merged link: expected insert for %s", record.RecordLinkID)
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
			result.Mutations = append(result.Mutations, MergeMutation{
				TargetKind:    "record_link",
				TargetID:      record.RecordLinkID.String(),
				OperationKind: "delete",
				BeforeValue:   before,
				AfterValue:    buildMergeLinkValue(tombstoned),
			})
			result.DedupedCount++
		}
		fieldKey := mergeLinkTypeFieldKey(record.LinkType)
		if fieldKey != "" {
			current := result.TimelineInvalidations[record.SrcRecordID]
			current = append(current, fieldKey)
			result.TimelineInvalidations[record.SrcRecordID] = current
		}
	}
	return result, nil
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

func mergeLinkTypeFieldKey(linkType string) string {
	switch linkType {
	case "observed_on_host":
		return "timeline.host_refs"
	case "observed_as_identity":
		return "timeline.identity_refs"
	default:
		return ""
	}
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

func buildMergeLinkValue(record RecordLink) map[string]any {
	return map[string]any{
		"record_link_id": record.RecordLinkID.String(),
		"incident_id":    record.IncidentID.String(),
		"src_record_id":  record.SrcRecordID.String(),
		"dst_record_id":  record.DstRecordID.String(),
		"link_type":      record.LinkType,
		"provenance":     record.Provenance,
		"confidence":     record.Confidence,
		"deleted_at":     formatTimestampPointer(record.DeletedAt),
	}
}

func buildMergeTagValue(record mergeTagRecord) map[string]any {
	return map[string]any{
		"record_tag_id":       record.RecordTagID.String(),
		"incident_id":         record.IncidentID.String(),
		"record_id":           record.RecordID.String(),
		"tag_name":            record.TagName,
		"normalized_tag_name": record.NormalizedTagName,
		"deleted_at":          formatTimestampPointer(record.DeletedAt),
		"deleted_by_user_id":  formatUUIDPointer(record.DeletedByUserID),
	}
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
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
