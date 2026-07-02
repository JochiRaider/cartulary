package mentions

import (
	"context"
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

type RepointMergedMentionsCommand struct {
	IncidentID       uuid.UUID
	EntityType       string
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	ActorUserID      uuid.UUID
	Now              time.Time
}

type RepointMergedMentionsResult struct {
	Mutations             []MergeMutation
	RepointedCount        int
	TimelineInvalidations map[uuid.UUID][]string
}

type mergeMentionRecord struct {
	EntityMentionID  uuid.UUID
	SourceRecordID   uuid.UUID
	SourceFieldKey   string
	EntityType       string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
}

func (s *Store) RepointMergedMentionsTx(ctx context.Context, tx pgx.Tx, command RepointMergedMentionsCommand) (RepointMergedMentionsResult, error) {
	rows, err := tx.Query(ctx, `
SELECT
    entity_mention_id,
    source_record_id,
    source_field_key,
    entity_type,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
  FROM entity_mentions
 WHERE entity_type = $1
   AND resolution_status = 'resolved'
   AND resolved_record_id = $2
 ORDER BY source_record_id ASC, source_field_key ASC, entity_mention_id ASC
 FOR UPDATE
`, command.EntityType, command.LoserRecordID)
	if err != nil {
		return RepointMergedMentionsResult{}, fmt.Errorf("load merged mentions: %w", err)
	}
	defer rows.Close()

	records := make([]mergeMentionRecord, 0)
	for rows.Next() {
		record, err := scanMergeMentionRecord(rows)
		if err != nil {
			return RepointMergedMentionsResult{}, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return RepointMergedMentionsResult{}, fmt.Errorf("iterate merged mentions: %w", err)
	}
	rows.Close()

	result := RepointMergedMentionsResult{
		Mutations:             make([]MergeMutation, 0, len(records)),
		TimelineInvalidations: make(map[uuid.UUID][]string),
	}
	for _, record := range records {
		before := buildMergeMentionValue(record)
		beforeVersion := mentionVersionID(record.EntityMentionID, record.RowVersion)
		record.RowVersion++
		record.ResolvedRecordID = &command.SurvivorRecordID
		if _, err := tx.Exec(ctx, `
UPDATE entity_mentions
   SET resolved_record_id = $2,
       row_version = $3
 WHERE entity_mention_id = $1
`, record.EntityMentionID, command.SurvivorRecordID, record.RowVersion); err != nil {
			return RepointMergedMentionsResult{}, fmt.Errorf("repoint merged mention: %w", err)
		}
		afterVersion := mentionVersionID(record.EntityMentionID, record.RowVersion)
		result.Mutations = append(result.Mutations, MergeMutation{
			TargetKind:      "entity_mention",
			TargetID:        record.EntityMentionID.String(),
			OperationKind:   "patch",
			BeforeVersionID: &beforeVersion,
			AfterVersionID:  &afterVersion,
			BeforeValue:     before,
			AfterValue:      buildMergeMentionValue(record),
		})
		current := result.TimelineInvalidations[record.SourceRecordID]
		current = append(current, record.SourceFieldKey)
		result.TimelineInvalidations[record.SourceRecordID] = current
		result.RepointedCount++
	}
	return result, nil
}

func buildMergeMentionValue(record mergeMentionRecord) map[string]any {
	return map[string]any{
		"entity_mention_id":   record.EntityMentionID.String(),
		"source_record_id":    record.SourceRecordID.String(),
		"source_field_key":    record.SourceFieldKey,
		"entity_type":         record.EntityType,
		"origin_kind":         record.OriginKind,
		"origin_locator":      record.OriginLocator,
		"raw_text":            record.RawText,
		"normalized_text":     record.NormalizedText,
		"resolution_status":   record.ResolutionStatus,
		"row_version":         record.RowVersion,
		"resolved_record_id":  formatUUIDPointer(record.ResolvedRecordID),
		"resolved_by_user_id": formatUUIDPointer(record.ResolvedByUserID),
		"resolved_at":         formatTimestampPointer(record.ResolvedAt),
		"resolution_method":   derefString(record.ResolutionMethod),
	}
}

func scanMergeMentionRecord(scanner interface{ Scan(dest ...any) error }) (mergeMentionRecord, error) {
	var (
		record           mergeMentionRecord
		resolvedRecordID pgtype.UUID
		resolvedByUserID pgtype.UUID
		resolvedAt       pgtype.Timestamptz
		resolutionMethod pgtype.Text
	)
	if err := scanner.Scan(
		&record.EntityMentionID,
		&record.SourceRecordID,
		&record.SourceFieldKey,
		&record.EntityType,
		&record.OriginKind,
		&record.OriginLocator,
		&record.RawText,
		&record.NormalizedText,
		&record.ResolutionStatus,
		&record.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
	); err != nil {
		return mergeMentionRecord{}, fmt.Errorf("scan merged mention: %w", err)
	}
	record.ResolvedRecordID = uuidPointerFromPG(resolvedRecordID)
	record.ResolvedByUserID = uuidPointerFromPG(resolvedByUserID)
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		record.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		record.ResolutionMethod = &value
	}
	return record, nil
}
