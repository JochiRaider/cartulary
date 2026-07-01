package projections

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool     postgres.DB
	registry *providerRegistry
}

type TimelineProjectionInput struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
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
	RecordedAt            time.Time
	EditedAt              time.Time
	ActivitySortTS        *time.Time
	DateEnteredSortDay    *time.Time
	ActivityTimePairState string
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
}

func NewStore(pool postgres.DB) *Store {
	return &Store{pool: pool, registry: defaultProviderRegistry()}
}

func (s *Store) UpsertTimelineRowTx(ctx context.Context, tx pgx.Tx, input TimelineProjectionInput) error {
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_grid_projection (
    record_id,
    incident_id,
    row_version,
    date_entered_text,
    analyst_text,
    mitre_stage_text,
    device_object_text,
    ip_address_text,
    activity_utc_text,
    activity_local_text,
    raw_activity_text,
    activity_synopsis_text,
    data_source_text,
    recorded_at,
    edited_at,
    activity_sort_ts,
    date_entered_sort_day,
    activity_time_pair_state,
    capture_state,
    replacement_record_id,
    evidence_count,
    has_evidence,
    has_unresolved_mentions
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    $15,
    $16,
    $17,
    $18,
    $19,
    $20,
    $21,
    $22,
    $23
)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    date_entered_text = EXCLUDED.date_entered_text,
    analyst_text = EXCLUDED.analyst_text,
    mitre_stage_text = EXCLUDED.mitre_stage_text,
    device_object_text = EXCLUDED.device_object_text,
    ip_address_text = EXCLUDED.ip_address_text,
    activity_utc_text = EXCLUDED.activity_utc_text,
    activity_local_text = EXCLUDED.activity_local_text,
    raw_activity_text = EXCLUDED.raw_activity_text,
    activity_synopsis_text = EXCLUDED.activity_synopsis_text,
    data_source_text = EXCLUDED.data_source_text,
    recorded_at = EXCLUDED.recorded_at,
    edited_at = EXCLUDED.edited_at,
    activity_sort_ts = EXCLUDED.activity_sort_ts,
    date_entered_sort_day = EXCLUDED.date_entered_sort_day,
    activity_time_pair_state = EXCLUDED.activity_time_pair_state,
    capture_state = EXCLUDED.capture_state,
    replacement_record_id = EXCLUDED.replacement_record_id,
    evidence_count = EXCLUDED.evidence_count,
    has_evidence = EXCLUDED.has_evidence,
    has_unresolved_mentions = EXCLUDED.has_unresolved_mentions
`, input.RecordID, input.IncidentID, input.RowVersion, input.DateEnteredText, input.AnalystText, input.MitreStageText, input.DeviceObjectText, input.IPAddressText, input.ActivityUTCText, input.ActivityLocalText, input.RawActivityText, input.ActivitySynopsisText, input.DataSourceText, input.RecordedAt.UTC(), input.EditedAt.UTC(), input.ActivitySortTS, input.DateEnteredSortDay, input.ActivityTimePairState, input.CaptureState, input.ReplacementRecordID, input.EvidenceCount, input.HasEvidence, input.HasUnresolvedMentions); err != nil {
		return fmt.Errorf("upsert timeline projection row: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentTimeline(ctx context.Context, incidentID uuid.UUID) (err error) {
	ctx, finishTelemetry := s.startProjectionSpan(ctx, timelineViewSchemaID)
	defer func() { finishTelemetry(err) }()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin timeline projection rebuild: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := s.RebuildIncidentTimelineTx(ctx, tx, incidentID); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit timeline projection rebuild: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentTimelineTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return s.rebuildProjectionIncidentTx(ctx, tx, timelineViewSchemaID, incidentID)
}

func (s *Store) rebuildIncidentTimelineTxCore(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return timelineprojection.RebuildIncidentTimelineTx(ctx, tx, incidentID, func(ctx context.Context, tx pgx.Tx, input timelineprojection.ProjectionInput) error {
		return s.UpsertTimelineRowTx(ctx, tx, TimelineProjectionInput(input))
	})
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}
