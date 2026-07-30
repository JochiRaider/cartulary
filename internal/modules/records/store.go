package records

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	db postgres.DB
}

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

type InsertParams struct {
	RecordID        *uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	RowVersion      int64
}

func NewStore(database ...postgres.DB) *Store {
	store := &Store{}
	if len(database) > 0 {
		store.db = database[0]
	}
	return store
}

func (s *Store) InsertTx(ctx context.Context, tx pgx.Tx, params InsertParams) (uuid.UUID, error) {
	rowVersion := params.RowVersion
	if rowVersion == 0 {
		rowVersion = 1
	}
	if params.RecordID != nil {
		if _, err := tx.Exec(ctx, `
INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`, *params.RecordID, params.IncidentID, params.RecordType, params.CreatedByUserID, params.CreatedAt.UTC(), params.UpdatedByUserID, params.UpdatedAt.UTC(), rowVersion); err != nil {
			return uuid.UUID{}, fmt.Errorf("insert record envelope: %w", err)
		}
		return *params.RecordID, nil
	}
	var recordIDArg any
	var recordID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO records (
    record_id,
    incident_id,
    record_type,
    created_by_user_id,
    created_at,
    updated_by_user_id,
    updated_at,
    row_version
)
VALUES (
    COALESCE($1::uuid, gen_random_uuid()),
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING record_id
`, recordIDArg, params.IncidentID, params.RecordType, params.CreatedByUserID, params.CreatedAt.UTC(), params.UpdatedByUserID, params.UpdatedAt.UTC(), rowVersion).Scan(&recordID); err != nil {
		return uuid.UUID{}, fmt.Errorf("insert record envelope: %w", err)
	}
	return recordID, nil
}

func (s *Store) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	var rowVersion int64
	if err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID).Scan(&rowVersion); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrEnvelopeNotFound
		}
		return 0, fmt.Errorf("advance record envelope version: %w", err)
	}
	return rowVersion, nil
}

func (s *Store) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	var rowVersion int64
	if err := tx.QueryRow(ctx, `
SELECT row_version
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&rowVersion); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrEnvelopeNotFound
		}
		return 0, fmt.Errorf("load record envelope row version: %w", err)
	}
	return rowVersion, nil
}

func (s *Store) LoadEnvelope(ctx context.Context, recordID uuid.UUID) (Envelope, error) {
	if s == nil || s.db == nil {
		return Envelope{}, errors.New("records: envelope store database is required")
	}
	return loadEnvelopeRow(ctx, s.db.QueryRow(ctx, envelopeSelectSQL(false), recordID))
}

func (s *Store) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock bool) (Envelope, error) {
	return loadEnvelopeRow(ctx, tx.QueryRow(ctx, envelopeSelectSQL(lock), recordID))
}

func (s *Store) LoadEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID, lock bool) (map[uuid.UUID]Envelope, error) {
	if len(recordIDs) == 0 {
		return map[uuid.UUID]Envelope{}, nil
	}
	query := `
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
  FROM records
 WHERE record_id = ANY($1::uuid[])
 ORDER BY record_id`
	if lock {
		query += "\n FOR UPDATE"
	}
	rows, err := tx.Query(ctx, query, recordIDs)
	if err != nil {
		return nil, fmt.Errorf("load record envelopes: %w", err)
	}
	defer rows.Close()

	envelopes := make(map[uuid.UUID]Envelope, len(recordIDs))
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

func (s *Store) SetDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time, deleting bool) (int64, error) {
	var rowVersion int64
	if err := tx.QueryRow(ctx, `
UPDATE records
   SET row_version = row_version + 1,
       deleted_at = CASE WHEN $4::boolean THEN $2::timestamptz ELSE NULL END,
       deleted_by_user_id = CASE WHEN $4::boolean THEN $3::uuid ELSE NULL END,
       updated_at = $2,
       updated_by_user_id = $3
 WHERE record_id = $1
RETURNING row_version
`, recordID, now.UTC(), actorUserID, deleting).Scan(&rowVersion); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrEnvelopeNotFound
		}
		return 0, fmt.Errorf("set record envelope delete state: %w", err)
	}
	return rowVersion, nil
}

func envelopeSelectSQL(lock bool) string {
	query := `
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
  FROM records
 WHERE record_id = $1`
	if lock {
		query += "\n FOR UPDATE"
	}
	return query
}

func loadEnvelopeRow(_ context.Context, row pgx.Row) (Envelope, error) {
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
