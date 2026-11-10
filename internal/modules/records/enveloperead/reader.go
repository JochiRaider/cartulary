package enveloperead

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"time"
)

var ErrEnvelopeNotFound = errors.New("records: envelope not found")

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
	DeletedByUserID *uuid.UUID
}

const SelectSQL = `
SELECT
    record_id,
    incident_id,
    record_type,
    row_version,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    deleted_at,
    deleted_by_user_id
  FROM records`

type Reader struct{}

func (Reader) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (Envelope, error) {
	return ScanRow(tx.QueryRow(ctx, SelectSQL+" WHERE record_id = $1", recordID))
}
func (Reader) LoadEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) (map[uuid.UUID]Envelope, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]Envelope{}, nil
	}
	rows, err := tx.Query(ctx, SelectSQL+" WHERE record_id = ANY($1::uuid[]) ORDER BY record_id", recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load record envelopes: %w", err)
	}
	defer rows.Close()
	return Collect(rows, len(recordIDs))
}
func Collect(rows pgx.Rows, capacity int) (map[uuid.UUID]Envelope, error) {
	envelopes := make(map[uuid.UUID]Envelope, capacity)
	for rows.Next() {
		var envelope Envelope
		if err := rows.Scan(
			&envelope.RecordID,
			&envelope.IncidentID,
			&envelope.RecordType,
			&envelope.RowVersion,
			&envelope.CreatedByUserID,
			&envelope.CreatedAt,
			&envelope.UpdatedByUserID,
			&envelope.UpdatedAt,
			&envelope.DeletedAt,
			&envelope.DeletedByUserID,
		); err != nil {
			return nil, fmt.Errorf("scan record envelope: %w", err)
		}
		envelope.CreatedAt = envelope.CreatedAt.UTC()
		envelope.UpdatedAt = envelope.UpdatedAt.UTC()
		if envelope.DeletedAt != nil {
			deletedAt := envelope.DeletedAt.UTC()
			envelope.DeletedAt = &deletedAt
		}
		envelopes[envelope.RecordID] = envelope
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate record envelopes: %w", err)
	}
	return envelopes, nil
}
func ScanRow(row pgx.Row) (Envelope, error) {
	var envelope Envelope
	if err := row.Scan(
		&envelope.RecordID,
		&envelope.IncidentID,
		&envelope.RecordType,
		&envelope.RowVersion,
		&envelope.CreatedByUserID,
		&envelope.CreatedAt,
		&envelope.UpdatedByUserID,
		&envelope.UpdatedAt,
		&envelope.DeletedAt,
		&envelope.DeletedByUserID,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Envelope{}, ErrEnvelopeNotFound
		}
		return Envelope{}, fmt.Errorf("load record envelope: %w", err)
	}
	envelope.CreatedAt = envelope.CreatedAt.UTC()
	envelope.UpdatedAt = envelope.UpdatedAt.UTC()
	if envelope.DeletedAt != nil {
		deletedAt := envelope.DeletedAt.UTC()
		envelope.DeletedAt = &deletedAt
	}
	return envelope, nil
}
