package mentioneffects

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/rowsnapshot"
)

const TimelineViewSchemaID = projectionadapters.TimelineViewSchemaID

var ErrSourceRecordNotFound = errors.New("timeline mention effects: source record not found")

type SourceRecord struct {
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

type TimelineInvalidation struct {
	RecordID         uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
}

func LoadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (SourceRecord, error) {
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

	var record SourceRecord
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
			return SourceRecord{}, ErrSourceRecordNotFound
		}
		return SourceRecord{}, fmt.Errorf("load mention source record: %w", err)
	}
	record.RecordedAt = record.RecordedAt.UTC()
	record.EditedAt = record.EditedAt.UTC()
	record.ReviewedAt = normalizeTimePointer(record.ReviewedAt)
	record.SupersededAt = normalizeTimePointer(record.SupersededAt)
	return record, nil
}

func UpdateSourceRecordTx(ctx context.Context, tx pgx.Tx, record SourceRecord) error {
	if _, err := tx.Exec(ctx, `
UPDATE timeline_events
   SET row_version = $2,
       edited_at = $3,
       updated_by_user_id = $4
 WHERE record_id = $1
`, record.RecordID, record.RowVersion, record.EditedAt.UTC(), record.UpdatedByUserID); err != nil {
		return fmt.Errorf("update mention source record: %w", err)
	}
	return nil
}

func BuildRecordRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	snapshot, err := rowsnapshot.BuildRecordRowTx(ctx, tx, recordID)
	if errors.Is(err, rowsnapshot.ErrRecordNotFound) {
		return nil, ErrSourceRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return snapshot.Row, nil
}

func RebuildTimelineProjectionTx(ctx context.Context, tx pgx.Tx, projector *projectionadapters.RowProjector, incidentID uuid.UUID) error {
	return projector.RebuildIncidentViewTx(ctx, tx, projectionadapters.TimelineViewSchemaID, incidentID)
}

func LoadTimelineInvalidationsTx(ctx context.Context, tx pgx.Tx, fieldKeysByRecord map[uuid.UUID][]string) ([]TimelineInvalidation, error) {
	if len(fieldKeysByRecord) == 0 {
		return nil, nil
	}
	recordIDs := make([]uuid.UUID, 0, len(fieldKeysByRecord))
	for recordID := range fieldKeysByRecord {
		recordIDs = append(recordIDs, recordID)
	}
	slices.SortFunc(recordIDs, func(left uuid.UUID, right uuid.UUID) int {
		return strings.Compare(left.String(), right.String())
	})
	result := make([]TimelineInvalidation, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		var rowVersion int64
		if err := tx.QueryRow(ctx, `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&rowVersion); err != nil {
			return nil, fmt.Errorf("load record invalidation row_version: %w", err)
		}
		fieldKeys := append([]string(nil), fieldKeysByRecord[recordID]...)
		slices.Sort(fieldKeys)
		fieldKeys = slices.Compact(fieldKeys)
		result = append(result, TimelineInvalidation{
			RecordID:         recordID,
			RowVersion:       rowVersion,
			ChangedFieldKeys: fieldKeys,
		})
	}
	return result, nil
}

func VersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("timeline_record:%s:%d", recordID.String(), rowVersion)
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
