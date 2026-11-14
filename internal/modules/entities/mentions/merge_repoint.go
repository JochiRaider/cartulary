package mentions

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions/historyvalue"
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
	Ordinal          int32
	CreatedByUserID  uuid.UUID
	CreatedAt        time.Time
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
    ordinal,
    created_by_user_id,
    created_at,
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
	return (historyvalue.Value{
		MentionID: record.EntityMentionID, SourceRecordID: record.SourceRecordID,
		EntityType: record.EntityType, SourceFieldKey: record.SourceFieldKey,
		OriginKind: record.OriginKind, OriginLocator: record.OriginLocator,
		RawText: record.RawText, NormalizedText: record.NormalizedText,
		ResolutionStatus: record.ResolutionStatus, RowVersion: record.RowVersion,
		Ordinal: record.Ordinal, CreatedByUserID: record.CreatedByUserID, CreatedAt: record.CreatedAt,
		ResolvedRecordID: record.ResolvedRecordID, ResolvedByUserID: record.ResolvedByUserID,
		ResolvedAt: record.ResolvedAt, ResolutionMethod: record.ResolutionMethod,
	}).Map()
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
		&record.Ordinal,
		&record.CreatedByUserID,
		&record.CreatedAt,
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
