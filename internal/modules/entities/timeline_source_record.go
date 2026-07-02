package entities

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type timelineSourceRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	DateEnteredText       *string
	AnalystText           *string
	MitreStageText        *string
	DeviceObjectText      *string
	IPAddressText         *string
	ActivityUTCText       *string
	ActivityLocalText     *string
	RawActivityText       *string
	ActivitySynopsisText  *string
	DataSourceText        *string
	ActivityTimePairState string
	CaptureState          string
	RowVersion            int64
	RecordedAt            time.Time
	EditedAt              time.Time
	CreatedByUserID       uuid.UUID
	UpdatedByUserID       uuid.UUID
	ReviewedByUserID      *uuid.UUID
	ReviewedAt            *time.Time
	SupersededByUserID    *uuid.UUID
	SupersededAt          *time.Time
}

func loadTimelineSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (timelineSourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    e.record_id,
    e.incident_id,
    e.date_entered_text,
    e.analyst_text,
    e.mitre_stage_text,
    e.device_object_text,
    e.ip_address_text,
    e.activity_utc_text,
    e.activity_local_text,
    e.raw_activity_text,
    e.activity_synopsis_text,
    e.data_source_text,
    e.activity_time_pair_state,
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
		&record.DateEnteredText,
		&record.AnalystText,
		&record.MitreStageText,
		&record.DeviceObjectText,
		&record.IPAddressText,
		&record.ActivityUTCText,
		&record.ActivityLocalText,
		&record.RawActivityText,
		&record.ActivitySynopsisText,
		&record.DataSourceText,
		&record.ActivityTimePairState,
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
	record.ReviewedAt = normalizeTimePointer(record.ReviewedAt)
	record.SupersededAt = normalizeTimePointer(record.SupersededAt)
	return record, nil
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
