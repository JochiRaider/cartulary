package projections

import (
	"context"
	"errors"
	"fmt"
	"time"

	sqlc "example.com/todo/cartulary/internal/gen/sql"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

type TimelineProjectionInput struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	OccurredAt            *time.Time
	Summary               *string
	Details               *string
	SourceText            *string
	RecordedAt            time.Time
	EditedAt              time.Time
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) UpsertTimelineRowTx(ctx context.Context, tx pgx.Tx, input TimelineProjectionInput) error {
	sortTS := input.RecordedAt.UTC()
	if input.OccurredAt != nil {
		sortTS = input.OccurredAt.UTC()
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_grid_projection (
    record_id,
    incident_id,
    row_version,
    occurred_at,
    summary,
    details,
    source_text,
    recorded_at,
    edited_at,
    sort_ts,
    capture_state,
    replacement_record_id,
    occurred_day,
    recorded_day,
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
    CASE WHEN $4::timestamptz IS NULL THEN NULL ELSE ($4::timestamptz AT TIME ZONE 'UTC')::date END,
    ($8::timestamptz AT TIME ZONE 'UTC')::date,
    $13,
    $14,
    $15
)
ON CONFLICT (record_id) DO UPDATE
SET incident_id = EXCLUDED.incident_id,
    row_version = EXCLUDED.row_version,
    occurred_at = EXCLUDED.occurred_at,
    summary = EXCLUDED.summary,
    details = EXCLUDED.details,
    source_text = EXCLUDED.source_text,
    recorded_at = EXCLUDED.recorded_at,
    edited_at = EXCLUDED.edited_at,
    sort_ts = EXCLUDED.sort_ts,
    capture_state = EXCLUDED.capture_state,
    replacement_record_id = EXCLUDED.replacement_record_id,
    occurred_day = EXCLUDED.occurred_day,
    recorded_day = EXCLUDED.recorded_day,
    evidence_count = EXCLUDED.evidence_count,
    has_evidence = EXCLUDED.has_evidence,
    has_unresolved_mentions = EXCLUDED.has_unresolved_mentions
`, input.RecordID, input.IncidentID, input.RowVersion, input.OccurredAt, input.Summary, input.Details, input.SourceText, input.RecordedAt.UTC(), input.EditedAt.UTC(), sortTS, input.CaptureState, input.ReplacementRecordID, input.EvidenceCount, input.HasEvidence, input.HasUnresolvedMentions); err != nil {
		return fmt.Errorf("upsert timeline projection row: %w", err)
	}
	return nil
}

func (s *Store) RebuildIncidentTimeline(ctx context.Context, incidentID uuid.UUID) error {
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
	if _, err := tx.Exec(ctx, `DELETE FROM timeline_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		return fmt.Errorf("clear timeline projection rows: %w", err)
	}

	queries := sqlc.New(tx)
	rows, err := queries.ListTimelineProjectionSourceRows(ctx, pgUUID(incidentID))
	if err != nil {
		return fmt.Errorf("list timeline projection source rows: %w", err)
	}

	for _, row := range rows {
		input, err := projectionInputFromSQL(row)
		if err != nil {
			return err
		}
		if err := s.UpsertTimelineRowTx(ctx, tx, input); err != nil {
			return err
		}
	}
	return nil
}

func projectionInputFromSQL(row sqlc.ListTimelineProjectionSourceRowsRow) (TimelineProjectionInput, error) {
	recordID, err := uuidFromPG(row.RecordID)
	if err != nil {
		return TimelineProjectionInput{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return TimelineProjectionInput{}, err
	}
	recordedAt, err := timeFromPG(row.RecordedAt)
	if err != nil {
		return TimelineProjectionInput{}, err
	}
	editedAt, err := timeFromPG(row.EditedAt)
	if err != nil {
		return TimelineProjectionInput{}, err
	}

	return TimelineProjectionInput{
		RecordID:              recordID,
		IncidentID:            incidentID,
		RowVersion:            row.RowVersion,
		OccurredAt:            optionalTimeFromPG(row.OccurredAt),
		Summary:               optionalTextFromPG(row.Summary),
		Details:               optionalTextFromPG(row.Details),
		SourceText:            optionalTextFromPG(row.SourceText),
		RecordedAt:            recordedAt,
		EditedAt:              editedAt,
		CaptureState:          row.CaptureState,
		ReplacementRecordID:   optionalUUIDFromPG(row.ReplacementRecordID),
		EvidenceCount:         int(row.EvidenceCount),
		HasEvidence:           row.HasEvidence,
		HasUnresolvedMentions: row.HasUnresolvedMentions,
	}, nil
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func optionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}
