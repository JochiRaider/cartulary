package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

const autoResolutionMethod = "auto_match"

type timelineSourceRecord struct {
	RecordID           uuid.UUID
	IncidentID         uuid.UUID
	OccurredAt         *time.Time
	Summary            *string
	Details            *string
	SourceText         *string
	CaptureState       string
	RowVersion         int64
	RecordedAt         time.Time
	EditedAt           time.Time
	CreatedByUserID    uuid.UUID
	UpdatedByUserID    uuid.UUID
	ReviewedByUserID   *uuid.UUID
	ReviewedAt         *time.Time
	SupersededByUserID *uuid.UUID
	SupersededAt       *time.Time
}

type timelineProjectedRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	OccurredAt            *time.Time
	Summary               *string
	Details               *string
	SourceText            *string
	RecordedAt            time.Time
	EditedAt              time.Time
	SortTS                time.Time
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	OccurredDay           *time.Time
	RecordedDay           time.Time
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
	HostRefs              []map[string]any
	IdentityRefs          []map[string]any
	Tags                  []map[string]any
}

type mentionQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type projectedMentionItem struct {
	MentionID        uuid.UUID
	EntityType       string
	SourceFieldKey   string
	RawText          string
	ResolutionStatus string
	RowVersion       int64
	ResolvedRecordID *uuid.UUID
	ResolutionMethod *string
}

func loadTimelineSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (timelineSourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    e.record_id,
    e.incident_id,
    e.occurred_at,
    e.summary,
    e.details,
    e.source_text,
    e.capture_state,
    r.row_version,
    e.recorded_at,
    e.edited_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    e.reviewed_by_user_id,
    e.reviewed_at,
    e.superseded_by_user_id,
    e.superseded_at
  FROM timeline_events e
  JOIN records r ON r.record_id = e.record_id
 WHERE e.record_id = $1
 FOR UPDATE OF e, r
`, recordID)

	var record timelineSourceRecord
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.OccurredAt,
		&record.Summary,
		&record.Details,
		&record.SourceText,
		&record.CaptureState,
		&record.RowVersion,
		&record.RecordedAt,
		&record.EditedAt,
		&record.CreatedByUserID,
		&record.UpdatedByUserID,
		&record.ReviewedByUserID,
		&record.ReviewedAt,
		&record.SupersededByUserID,
		&record.SupersededAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return timelineSourceRecord{}, ErrSourceRecordNotFound
		}
		return timelineSourceRecord{}, fmt.Errorf("load mention source record: %w", err)
	}
	record.RecordedAt = record.RecordedAt.UTC()
	record.EditedAt = record.EditedAt.UTC()
	record.OccurredAt = normalizeTimePointer(record.OccurredAt)
	record.ReviewedAt = normalizeTimePointer(record.ReviewedAt)
	record.SupersededAt = normalizeTimePointer(record.SupersededAt)
	return record, nil
}

func projectTimelineRecord(record timelineSourceRecord, replacementRecordID *uuid.UUID) timelineProjectedRecord {
	sortTS := record.RecordedAt.UTC()
	if record.OccurredAt != nil {
		sortTS = record.OccurredAt.UTC()
	}

	var occurredDay *time.Time
	if record.OccurredAt != nil {
		day := time.Date(record.OccurredAt.UTC().Year(), record.OccurredAt.UTC().Month(), record.OccurredAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
		occurredDay = &day
	}
	recordedDay := time.Date(record.RecordedAt.UTC().Year(), record.RecordedAt.UTC().Month(), record.RecordedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)

	return timelineProjectedRecord{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		OccurredAt:            normalizeTimePointer(record.OccurredAt),
		Summary:               cloneStringPointer(record.Summary),
		Details:               cloneStringPointer(record.Details),
		SourceText:            cloneStringPointer(record.SourceText),
		RecordedAt:            record.RecordedAt.UTC(),
		EditedAt:              record.EditedAt.UTC(),
		SortTS:                sortTS,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   replacementRecordID,
		OccurredDay:           occurredDay,
		RecordedDay:           recordedDay,
		EvidenceCount:         0,
		HasEvidence:           false,
		HasUnresolvedMentions: false,
	}
}

func hydrateTimelineCollections(ctx context.Context, querier mentionQueryer, record *timelineProjectedRecord) error {
	if record == nil {
		return nil
	}
	rows, err := querier.Query(ctx, `
SELECT entity_mention_id, entity_type, source_field_key, raw_text, resolution_status, row_version, resolved_record_id, resolution_method, ordinal
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status IN ('unresolved', 'resolved')
 ORDER BY ordinal ASC, entity_mention_id ASC
`, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline mention collections: %w", err)
	}

	mentions := make([]projectedMentionItem, 0)
	for rows.Next() {
		var (
			mentionID        uuid.UUID
			entityType       string
			sourceFieldKey   string
			rawText          string
			resolutionStatus string
			rowVersion       int64
			resolvedRecordID pgtype.UUID
			resolutionMethod pgtype.Text
			ordinal          int
		)
		if err := rows.Scan(&mentionID, &entityType, &sourceFieldKey, &rawText, &resolutionStatus, &rowVersion, &resolvedRecordID, &resolutionMethod, &ordinal); err != nil {
			rows.Close()
			return fmt.Errorf("scan timeline mention collection row: %w", err)
		}
		mention := projectedMentionItem{
			MentionID:        mentionID,
			EntityType:       entityType,
			SourceFieldKey:   sourceFieldKey,
			RawText:          rawText,
			ResolutionStatus: resolutionStatus,
			RowVersion:       rowVersion,
		}
		if resolvedRecordID.Valid {
			resolved := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
			mention.ResolvedRecordID = &resolved
		}
		if resolutionMethod.Valid {
			value := resolutionMethod.String
			mention.ResolutionMethod = &value
		}
		mentions = append(mentions, mention)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate timeline mention collection rows: %w", err)
	}
	rows.Close()

	hostRefs := make([]map[string]any, 0)
	identityRefs := make([]map[string]any, 0)
	hasUnresolved := false
	for _, mention := range mentions {
		item := map[string]any{
			"item_ref":            "entity_mention:" + mention.MentionID.String(),
			"entity_type":         mention.EntityType,
			"display_text":        mention.RawText,
			"raw_text":            mention.RawText,
			"mention_row_version": mention.RowVersion,
		}
		if mention.ResolutionStatus == "resolved" && mention.ResolvedRecordID != nil {
			item["item_kind"] = "resolved_ref"
			item["resolved_record_id"] = mention.ResolvedRecordID.String()
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod != "" {
				item["resolution_method"] = *mention.ResolutionMethod
				if *mention.ResolutionMethod == autoResolutionMethod {
					item["auto_resolved"] = true
				}
			}
			if linkType, ok := mentionLinkType(mention.SourceFieldKey); ok {
				linkMetadata, err := loadActiveTimelineCollectionLinkMetadata(ctx, querier, record.IncidentID, record.RecordID, *mention.ResolvedRecordID, linkType)
				if err != nil {
					return err
				}
				if linkMetadata != nil {
					item["provenance"] = linkMetadata.Provenance
					item["confidence"] = linkMetadata.Confidence
				}
			}
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod == autoResolutionMethod {
				matchedAliasText, err := loadMatchedTimelineAliasText(ctx, querier, *mention.ResolvedRecordID, mention.EntityType, mention.RawText)
				if err != nil {
					return err
				}
				if matchedAliasText != nil {
					item["matched_alias_text"] = *matchedAliasText
				}
			}
		} else {
			item["item_kind"] = "unresolved_mention"
			hasUnresolved = true
		}

		switch mention.EntityType {
		case "host":
			hostRefs = append(hostRefs, item)
		case "identity":
			identityRefs = append(identityRefs, item)
		}
	}
	record.HostRefs = hostRefs
	record.IdentityRefs = identityRefs
	record.HasUnresolvedMentions = hasUnresolved

	tagRows, err := querier.Query(ctx, `
SELECT record_tag_id, tag_name
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline tags: %w", err)
	}

	tags := make([]map[string]any, 0)
	for tagRows.Next() {
		var (
			recordTagID uuid.UUID
			tagName     string
		)
		if err := tagRows.Scan(&recordTagID, &tagName); err != nil {
			tagRows.Close()
			return fmt.Errorf("scan timeline tag row: %w", err)
		}
		tags = append(tags, map[string]any{
			"item_ref":     "record_tag:" + record.RecordID.String() + ":" + recordTagID.String(),
			"item_kind":    "tag",
			"display_text": tagName,
			"tag_id":       recordTagID.String(),
		})
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return fmt.Errorf("iterate timeline tags: %w", err)
	}
	tagRows.Close()
	record.Tags = tags
	return nil
}

func buildTimelineRow(record timelineProjectedRecord) map[string]any {
	cells := map[string]any{
		"timeline.occurred_at":             map[string]any{"value": formatTimestampPointer(record.OccurredAt)},
		"timeline.summary":                 map[string]any{"value": derefString(record.Summary)},
		"timeline.details":                 map[string]any{"value": derefString(record.Details)},
		"timeline.source_text":             map[string]any{"value": derefString(record.SourceText)},
		"timeline.host_refs":               map[string]any{"value": collectionValue(true, record.HostRefs)},
		"timeline.identity_refs":           map[string]any{"value": collectionValue(true, record.IdentityRefs)},
		"timeline.evidence_count":          map[string]any{"value": record.EvidenceCount},
		"timeline.tags":                    map[string]any{"value": collectionValue(false, record.Tags)},
		"timeline.edited_at":               map[string]any{"value": formatTimestamp(record.EditedAt)},
		"timeline.recorded_at":             map[string]any{"value": formatTimestamp(record.RecordedAt)},
		"timeline.sort_ts":                 map[string]any{"value": formatTimestamp(record.SortTS)},
		"timeline.capture_state":           map[string]any{"value": record.CaptureState},
		"timeline.replacement_record_id":   map[string]any{"value": formatUUIDPointer(record.ReplacementRecordID)},
		"timeline.occurred_day":            map[string]any{"value": formatDatePointer(record.OccurredDay)},
		"timeline.recorded_day":            map[string]any{"value": formatDate(record.RecordedDay)},
		"timeline.has_evidence":            map[string]any{"value": record.HasEvidence},
		"timeline.has_unresolved_mentions": map[string]any{"value": record.HasUnresolvedMentions},
	}

	row := map[string]any{
		"record_id":   record.RecordID.String(),
		"row_version": record.RowVersion,
		"cells":       cells,
	}
	row["group_values"] = map[string]any{
		"timeline.occurred_day":            formatDatePointer(record.OccurredDay),
		"timeline.recorded_day":            formatDate(record.RecordedDay),
		"timeline.capture_state":           record.CaptureState,
		"timeline.has_evidence":            record.HasEvidence,
		"timeline.has_unresolved_mentions": record.HasUnresolvedMentions,
	}
	return row
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func formatDate(value time.Time) string {
	return value.UTC().Format("2006-01-02")
}

func formatDatePointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format("2006-01-02")
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

type timelineCollectionLinkMetadata struct {
	Provenance string
	Confidence *int
}

func loadActiveTimelineCollectionLinkMetadata(ctx context.Context, querier mentionQueryer, incidentID uuid.UUID, sourceRecordID uuid.UUID, targetRecordID uuid.UUID, linkType string) (*timelineCollectionLinkMetadata, error) {
	rows, err := querier.Query(ctx, `
SELECT provenance, confidence
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
`, incidentID, sourceRecordID, targetRecordID, linkType)
	if err != nil {
		return nil, fmt.Errorf("query active timeline collection link metadata: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("iterate active timeline collection link metadata: %w", err)
		}
		return nil, nil
	}

	var (
		metadata   timelineCollectionLinkMetadata
		confidence pgtype.Int4
	)
	if err := rows.Scan(&metadata.Provenance, &confidence); err != nil {
		return nil, fmt.Errorf("scan active timeline collection link metadata: %w", err)
	}
	if confidence.Valid {
		value := int(confidence.Int32)
		metadata.Confidence = &value
	}
	return &metadata, nil
}

func loadMatchedTimelineAliasText(ctx context.Context, querier mentionQueryer, recordID uuid.UUID, entityType string, rawText string) (*string, error) {
	candidateText, ok := fieldnorm.AutoResolutionCandidateText(rawText)
	if !ok {
		return nil, nil
	}

	rows, err := querier.Query(ctx, `
SELECT raw_text
  FROM entity_aliases
 WHERE record_id = $1
   AND entity_type = $2
   AND deleted_at IS NULL
 ORDER BY created_at ASC, entity_alias_id ASC
`, recordID, entityType)
	if err != nil {
		return nil, fmt.Errorf("query matched timeline alias text: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var aliasText string
		if err := rows.Scan(&aliasText); err != nil {
			return nil, fmt.Errorf("scan matched timeline alias text: %w", err)
		}
		aliasCandidateText, ok := fieldnorm.AutoResolutionCandidateText(aliasText)
		if ok && aliasCandidateText == candidateText {
			value := aliasText
			return &value, nil
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate matched timeline alias texts: %w", err)
	}
	return nil, nil
}
