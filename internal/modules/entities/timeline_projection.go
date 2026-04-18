package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

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
}

type mentionQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func loadTimelineSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (timelineSourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    record_id,
    incident_id,
    occurred_at,
    summary,
    details,
    source_text,
    capture_state,
    row_version,
    recorded_at,
    edited_at,
    created_by_user_id,
    updated_by_user_id,
    reviewed_by_user_id,
    reviewed_at,
    superseded_by_user_id,
    superseded_at
  FROM timeline_events
 WHERE record_id = $1
 FOR UPDATE
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

func hydrateTimelineMentionCollections(ctx context.Context, querier mentionQueryer, record *timelineProjectedRecord) error {
	if record == nil {
		return nil
	}
	rows, err := querier.Query(ctx, `
SELECT entity_mention_id, entity_type, raw_text, resolution_status, resolved_record_id, ordinal
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status IN ('unresolved', 'resolved')
 ORDER BY ordinal ASC, entity_mention_id ASC
`, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline mention collections: %w", err)
	}
	defer rows.Close()

	hostRefs := make([]map[string]any, 0)
	identityRefs := make([]map[string]any, 0)
	hasUnresolved := false
	for rows.Next() {
		var (
			mentionID        uuid.UUID
			entityType       string
			rawText          string
			resolutionStatus string
			resolvedRecordID pgtype.UUID
			ordinal          int
		)
		if err := rows.Scan(&mentionID, &entityType, &rawText, &resolutionStatus, &resolvedRecordID, &ordinal); err != nil {
			return fmt.Errorf("scan timeline mention collection row: %w", err)
		}

		item := map[string]any{
			"item_ref":     "entity_mention:" + mentionID.String(),
			"entity_type":  entityType,
			"display_text": rawText,
			"raw_text":     rawText,
		}
		if resolutionStatus == "resolved" && resolvedRecordID.Valid {
			item["item_kind"] = "resolved_ref"
			resolved := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
			item["resolved_record_id"] = resolved.String()
		} else {
			item["item_kind"] = "unresolved_mention"
			hasUnresolved = true
		}

		switch entityType {
		case "host":
			hostRefs = append(hostRefs, item)
		case "identity":
			identityRefs = append(identityRefs, item)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate timeline mention collection rows: %w", err)
	}

	record.HostRefs = hostRefs
	record.IdentityRefs = identityRefs
	record.HasUnresolvedMentions = hasUnresolved
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
		"timeline.tags":                    map[string]any{"value": collectionValue(false, nil)},
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
