package sourcerepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const timelineRecordType = "timeline_event"

var (
	ErrNotFound         = errors.New("timeline source repository: record not found")
	ErrEnvelopeNotFound = errors.New("timeline source repository: record envelope not found")
)

type Envelope struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	RowVersion      int64
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	DeletedAt       *time.Time
}

type EnvelopeReader interface {
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID) (Envelope, error)
	LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID) (map[uuid.UUID]Envelope, error)
}

type Snapshot struct {
	RecordID               uuid.UUID
	IncidentID             uuid.UUID
	DateEnteredText        *string
	AnalystText            *string
	MitreStageText         *string
	DeviceObjectText       *string
	IPAddressText          *string
	ActivityUTCText        *string
	ActivityLocalText      *string
	RawActivityText        *string
	ActivitySynopsisText   *string
	DataSourceText         *string
	ActivityUTCGenerated   bool
	ActivityLocalGenerated bool
	ActivityTimePairState  string
	CaptureState           string
	RowVersion             int64
	RecordedAt             time.Time
	EditedAt               time.Time
	CreatedByUserID        uuid.UUID
	UpdatedByUserID        uuid.UUID
	ReviewedByUserID       *uuid.UUID
	ReviewedAt             *time.Time
	SupersededByUserID     *uuid.UUID
	SupersededAt           *time.Time
}

type Repository struct {
	envelopes EnvelopeReader
}

type Page struct {
	Snapshots    []Snapshot
	NextRecordID *uuid.UUID
}

func New(envelopes EnvelopeReader) *Repository {
	return &Repository{envelopes: envelopes}
}

func (r *Repository) LoadTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (Snapshot, error) {
	return r.loadTx(ctx, tx, nil, recordID)
}

func (r *Repository) ListIncidentPageTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, afterRecordID *uuid.UUID, limit int) (Page, error) {
	if limit < 1 {
		return Page{}, fmt.Errorf("timeline source page limit must be positive")
	}
	rows, err := tx.Query(ctx, SelectSQL+`
 WHERE incident_id = $1
   AND ($2::uuid IS NULL OR record_id > $2)
 ORDER BY record_id
 LIMIT $3
`, incidentID, afterRecordID, limit+1)
	if err != nil {
		return Page{}, fmt.Errorf("list timeline source page: %w", err)
	}
	defer rows.Close()
	snapshots := make([]Snapshot, 0, limit+1)
	for rows.Next() {
		snapshot, err := ScanSource(rows)
		if err != nil {
			return Page{}, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return Page{}, fmt.Errorf("iterate timeline source page: %w", err)
	}
	hasMore := len(snapshots) > limit
	if hasMore {
		snapshots = snapshots[:limit]
	}
	recordIDs := make([]uuid.UUID, 0, len(snapshots))
	for _, snapshot := range snapshots {
		recordIDs = append(recordIDs, snapshot.RecordID)
	}
	envelopes, err := r.envelopes.LoadEnvelopesTx(ctx, tx, recordIDs)
	if err != nil {
		return Page{}, err
	}
	filtered := snapshots[:0]
	for _, snapshot := range snapshots {
		envelope, ok := envelopes[snapshot.RecordID]
		if !ok || !EligibleEnvelope(snapshot, envelope, &incidentID) {
			continue
		}
		ApplyEnvelope(&snapshot, envelope)
		filtered = append(filtered, snapshot)
	}
	page := Page{Snapshots: filtered}
	if hasMore && len(snapshots) > 0 {
		next := snapshots[len(snapshots)-1].RecordID
		page.NextRecordID = &next
	}
	return page, nil
}

func (r *Repository) loadTx(ctx context.Context, tx pgx.Tx, incidentID *uuid.UUID, recordID uuid.UUID) (Snapshot, error) {
	query := SelectSQL + "\n WHERE record_id = $1"
	args := []any{recordID}
	if incidentID != nil {
		query += "\n   AND incident_id = $2"
		args = append(args, *incidentID)
	}
	snapshot, err := ScanSource(tx.QueryRow(ctx, query, args...))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Snapshot{}, ErrNotFound
		}
		return Snapshot{}, err
	}
	envelope, err := r.envelopes.LoadEnvelopeTx(ctx, tx, recordID)
	if err != nil {
		if errors.Is(err, ErrEnvelopeNotFound) {
			return Snapshot{}, ErrNotFound
		}
		return Snapshot{}, err
	}
	if !EligibleEnvelope(snapshot, envelope, incidentID) {
		return Snapshot{}, ErrNotFound
	}
	ApplyEnvelope(&snapshot, envelope)
	return snapshot, nil
}

const SelectSQL = `
SELECT
    record_id,
    incident_id,
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
    activity_utc_generated,
    activity_local_generated,
    activity_time_pair_state,
    capture_state,
    recorded_at,
    edited_at,
    reviewed_by_user_id,
    reviewed_at,
    superseded_by_user_id,
    superseded_at
  FROM timeline_events`

type rowScanner interface {
	Scan(...any) error
}

func ScanSource(row rowScanner) (Snapshot, error) {
	var snapshot Snapshot
	if err := row.Scan(
		&snapshot.RecordID,
		&snapshot.IncidentID,
		&snapshot.DateEnteredText,
		&snapshot.AnalystText,
		&snapshot.MitreStageText,
		&snapshot.DeviceObjectText,
		&snapshot.IPAddressText,
		&snapshot.ActivityUTCText,
		&snapshot.ActivityLocalText,
		&snapshot.RawActivityText,
		&snapshot.ActivitySynopsisText,
		&snapshot.DataSourceText,
		&snapshot.ActivityUTCGenerated,
		&snapshot.ActivityLocalGenerated,
		&snapshot.ActivityTimePairState,
		&snapshot.CaptureState,
		&snapshot.RecordedAt,
		&snapshot.EditedAt,
		&snapshot.ReviewedByUserID,
		&snapshot.ReviewedAt,
		&snapshot.SupersededByUserID,
		&snapshot.SupersededAt,
	); err != nil {
		return Snapshot{}, err
	}
	snapshot.RecordedAt = snapshot.RecordedAt.UTC()
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	snapshot.ReviewedAt = normalizeTimePointer(snapshot.ReviewedAt)
	snapshot.SupersededAt = normalizeTimePointer(snapshot.SupersededAt)
	return snapshot, nil
}

func EligibleEnvelope(snapshot Snapshot, envelope Envelope, incidentID *uuid.UUID) bool {
	if envelope.RecordType != timelineRecordType || envelope.DeletedAt != nil || envelope.IncidentID != snapshot.IncidentID {
		return false
	}
	return incidentID == nil || envelope.IncidentID == *incidentID
}

func ApplyEnvelope(snapshot *Snapshot, envelope Envelope) {
	snapshot.RowVersion = envelope.RowVersion
	snapshot.CreatedByUserID = envelope.CreatedByUserID
	snapshot.UpdatedByUserID = envelope.UpdatedByUserID
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	normalized := value.UTC()
	return &normalized
}
