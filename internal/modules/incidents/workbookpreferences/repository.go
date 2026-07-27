package workbookpreferences

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var ErrNotFound = errors.New("incidents workbook preferences: not found")

type DefaultRecord struct {
	IncidentID      uuid.UUID
	DefaultSheetRef []byte
	CreatedAt       time.Time
	UpdatedAt       time.Time
	UpdatedByUserID *uuid.UUID
}

type UserRecord struct {
	IncidentID   uuid.UUID
	UserID       uuid.UUID
	HomeSheetRef []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Repository struct {
	db postgres.DB
}

func NewRepository(db postgres.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetDefault(ctx context.Context, incidentID uuid.UUID) (DefaultRecord, error) {
	row, err := sqlc.New(r.db).GetDefaultWorkbookPreferences(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultRecord{}, ErrNotFound
	}
	if err != nil {
		return DefaultRecord{}, err
	}
	return defaultRecordFromSQL(row)
}

func (r *Repository) PutDefault(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	defaultSheetRef []byte,
	now time.Time,
) (DefaultRecord, error) {
	row, err := sqlc.New(r.db).PutDefaultWorkbookPreferences(ctx, sqlc.PutDefaultWorkbookPreferencesParams{
		IncidentID:      pgUUID(incidentID),
		Column2:         sheetRefParam(defaultSheetRef),
		CreatedAt:       pgTimestamptz(now),
		UpdatedByUserID: pgUUID(actorUserID),
	})
	if err != nil {
		return DefaultRecord{}, err
	}
	return defaultRecordFromSQL(row)
}

func (r *Repository) GetUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserRecord, error) {
	row, err := sqlc.New(r.db).GetUserWorkbookPreferences(ctx, sqlc.GetUserWorkbookPreferencesParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return UserRecord{}, ErrNotFound
	}
	if err != nil {
		return UserRecord{}, err
	}
	return userRecordFromSQL(row)
}

func (r *Repository) PutUser(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	homeSheetRef []byte,
	now time.Time,
) (UserRecord, error) {
	row, err := sqlc.New(r.db).PutUserWorkbookPreferences(ctx, sqlc.PutUserWorkbookPreferencesParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		Column3:    sheetRefParam(homeSheetRef),
		CreatedAt:  pgTimestamptz(now),
	})
	if err != nil {
		return UserRecord{}, err
	}
	return userRecordFromSQL(row)
}

type Session struct {
	tx pgx.Tx
}

func NewSession(tx pgx.Tx) Session {
	return Session{tx: tx}
}

func (s Session) UserRefForUpdate(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error) {
	ref, err := sqlc.New(s.tx).GetUserWorkbookPreferenceRefForUpdate(ctx, sqlc.GetUserWorkbookPreferenceRefForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup user preference: %w", err)
	}
	return cloneBytes(ref), nil
}

func (s Session) DefaultRefForUpdate(ctx context.Context, incidentID uuid.UUID) ([]byte, error) {
	ref, err := sqlc.New(s.tx).GetDefaultWorkbookPreferenceRefForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup default preference: %w", err)
	}
	return cloneBytes(ref), nil
}

func (s Session) ClearUserIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	rows, err := sqlc.New(s.tx).ClearUserWorkbookPreferenceRef(ctx, sqlc.ClearUserWorkbookPreferenceRefParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		Column3:    expected,
		UpdatedAt:  pgTimestamptz(now),
	})
	if err != nil {
		return false, fmt.Errorf("clear startup home pointer: %w", err)
	}
	return rows == 1, nil
}

func (s Session) ClearDefaultIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	rows, err := sqlc.New(s.tx).ClearDefaultWorkbookPreferenceRef(ctx, sqlc.ClearDefaultWorkbookPreferenceRefParams{
		IncidentID:      pgUUID(incidentID),
		Column2:         expected,
		UpdatedByUserID: pgUUID(actorUserID),
		UpdatedAt:       pgTimestamptz(now),
	})
	if err != nil {
		return false, fmt.Errorf("clear startup default pointer: %w", err)
	}
	return rows == 1, nil
}

func defaultRecordFromSQL(row sqlc.IncidentWorkbookPreference) (DefaultRecord, error) {
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return DefaultRecord{}, err
	}
	return DefaultRecord{
		IncidentID:      incidentID,
		DefaultSheetRef: cloneBytes(row.DefaultSheetRef),
		CreatedAt:       row.CreatedAt.Time.UTC(),
		UpdatedAt:       row.UpdatedAt.Time.UTC(),
		UpdatedByUserID: optionalUUIDFromPG(row.UpdatedByUserID),
	}, nil
}

func userRecordFromSQL(row sqlc.UserWorkbookPreference) (UserRecord, error) {
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return UserRecord{}, err
	}
	userID, err := uuidFromPG(row.UserID)
	if err != nil {
		return UserRecord{}, err
	}
	return UserRecord{
		IncidentID:   incidentID,
		UserID:       userID,
		HomeSheetRef: cloneBytes(row.HomeSheetRef),
		CreatedAt:    row.CreatedAt.Time.UTC(),
		UpdatedAt:    row.UpdatedAt.Time.UTC(),
	}, nil
}

func sheetRefParam(raw []byte) []byte {
	if len(raw) == 0 {
		return nil
	}
	return cloneBytes(raw)
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.Nil, errors.New("incidents workbook preferences: UUID is null")
	}
	return uuid.UUID(value.Bytes), nil
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func cloneBytes(raw []byte) []byte {
	if raw == nil {
		return nil
	}
	return append([]byte(nil), raw...)
}
