package startup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrPreferencesNotFound   = errors.New("workbook startup: preferences not found")
	errEmptySheetRef         = errors.New("empty sheet ref")
	errMissingSheetRefMember = errors.New("missing sheet ref member")
)

type Store struct {
	db postgres.DB
}

func NewStore(db postgres.DB) *Store {
	return &Store{db: db}
}

func (s *Store) GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error) {
	row, err := sqlc.New(s.db).GetDefaultWorkbookPreferences(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultPreferencesRecord{}, ErrPreferencesNotFound
	}
	if err != nil {
		return DefaultPreferencesRecord{}, err
	}
	return defaultPreferencesRecordFromSQL(row)
}

func (s *Store) PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error) {
	row, err := sqlc.New(s.db).PutDefaultWorkbookPreferences(ctx, sqlc.PutDefaultWorkbookPreferencesParams{
		IncidentID:      pgUUID(incidentID),
		Column2:         sheetRefParam(defaultSheetRef),
		CreatedAt:       pgTimestamptz(now),
		UpdatedByUserID: pgUUID(actorUserID),
	})
	if err != nil {
		return DefaultPreferencesRecord{}, err
	}
	return defaultPreferencesRecordFromSQL(row)
}

func (s *Store) GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error) {
	row, err := sqlc.New(s.db).GetUserWorkbookPreferences(ctx, sqlc.GetUserWorkbookPreferencesParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return UserPreferencesRecord{}, ErrPreferencesNotFound
	}
	if err != nil {
		return UserPreferencesRecord{}, err
	}
	return userPreferencesRecordFromSQL(row)
}

func (s *Store) PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error) {
	row, err := sqlc.New(s.db).PutUserWorkbookPreferences(ctx, sqlc.PutUserWorkbookPreferencesParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		Column3:    sheetRefParam(homeSheetRef),
		CreatedAt:  pgTimestamptz(now),
	})
	if err != nil {
		return UserPreferencesRecord{}, err
	}
	return userPreferencesRecordFromSQL(row)
}

func (s *Store) Resolve(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, _ string, explicitSheetRef []byte, now time.Time) (Record, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Record{}, fmt.Errorf("begin workbook startup transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	homeRef, err := getUserPreferenceRefTx(ctx, tx, incidentID, userID)
	if err != nil {
		return Record{}, err
	}
	defaultRef, err := getDefaultPreferenceRefTx(ctx, tx, incidentID)
	if err != nil {
		return Record{}, err
	}

	record := Record{
		IncidentID:      incidentID,
		HomeSheetRef:    cloneBytes(homeRef),
		DefaultSheetRef: cloneBytes(defaultRef),
	}
	for _, candidate := range []struct {
		source string
		ref    []byte
	}{
		{source: SourceExplicit, ref: explicitSheetRef},
		{source: SourceHome, ref: homeRef},
		{source: SourceDefault, ref: defaultRef},
	} {
		if len(candidate.ref) == 0 {
			continue
		}
		resolved, invalidReason, err := resolveCandidate(ctx, tx, incidentID, userID, candidate.ref)
		if err != nil {
			return Record{}, err
		}
		if invalidReason == "" {
			record.SelectedSheetRef = cloneBytes(candidate.ref)
			record.SelectedViewSchemaID = resolved.ViewSchemaID
			record.SelectedSavedView = resolved.SavedView
			record.Source = candidate.source
			if err := tx.Commit(ctx); err != nil {
				return Record{}, fmt.Errorf("commit workbook startup transaction: %w", err)
			}
			return record, nil
		}
		if candidate.source == SourceExplicit {
			continue
		}
		record.ClearedPointers = append(record.ClearedPointers, ClearedPointer{
			Source:     candidate.source,
			SheetRef:   cloneBytes(candidate.ref),
			ReasonCode: invalidReason,
		})
		if candidate.source == SourceHome {
			if err := clearUserPreferenceRefTx(ctx, tx, incidentID, userID, now); err != nil {
				return Record{}, err
			}
			record.HomeSheetRef = nil
		}
		if candidate.source == SourceDefault {
			if err := clearDefaultPreferenceRefTx(ctx, tx, incidentID, now); err != nil {
				return Record{}, err
			}
			record.DefaultSheetRef = nil
		}
	}

	record.SelectedSheetRef = mustSheetRefJSON(SheetRef{Kind: "view_schema", ID: TimelineViewSchemaID})
	record.SelectedViewSchemaID = TimelineViewSchemaID
	record.Source = SourceTimeline
	if err := tx.Commit(ctx); err != nil {
		return Record{}, fmt.Errorf("commit workbook startup fallback transaction: %w", err)
	}
	return record, nil
}

type resolvedCandidate struct {
	ViewSchemaID string
	SavedView    *SavedViewRecord
}

func resolveCandidate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, raw []byte) (resolvedCandidate, string, error) {
	ref, err := decodeSheetRef(raw)
	if err != nil {
		return resolvedCandidate{}, "invalid_sheet_ref", nil
	}
	switch ref.Kind {
	case "view_schema":
		if _, ok := viewschema.Lookup(ref.ID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		return resolvedCandidate{ViewSchemaID: ref.ID}, "", nil
	case "saved_view":
		savedViewID, err := uuid.Parse(ref.ID)
		if err != nil {
			return resolvedCandidate{}, "invalid_saved_view_id", nil
		}
		savedView, reasonCode, err := savedviews.ResolveStartupVisibleForUpdateTx(ctx, tx, incidentID, savedViewID, userID)
		if err != nil {
			return resolvedCandidate{}, "", err
		}
		if reasonCode != "" {
			return resolvedCandidate{}, reasonCode, nil
		}
		if _, ok := viewschema.Lookup(savedView.ViewSchemaID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		record := savedViewRecordFromSavedviews(savedView)
		return resolvedCandidate{ViewSchemaID: savedView.ViewSchemaID, SavedView: &record}, "", nil
	default:
		return resolvedCandidate{}, "unsupported_sheet_ref_kind", nil
	}
}

func getUserPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error) {
	ref, err := sqlc.New(tx).GetUserWorkbookPreferenceRefForUpdate(ctx, sqlc.GetUserWorkbookPreferenceRefForUpdateParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup user preference: %w", err)
	}
	return ref, nil
}

func getDefaultPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]byte, error) {
	ref, err := sqlc.New(tx).GetDefaultWorkbookPreferenceRefForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup default preference: %w", err)
	}
	return ref, nil
}

func clearUserPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, now time.Time) error {
	if err := sqlc.New(tx).ClearUserWorkbookPreferenceRef(ctx, sqlc.ClearUserWorkbookPreferenceRefParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		UpdatedAt:  pgTimestamptz(now),
	}); err != nil {
		return fmt.Errorf("clear startup home pointer: %w", err)
	}
	return nil
}

func clearDefaultPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) error {
	if err := sqlc.New(tx).ClearDefaultWorkbookPreferenceRef(ctx, sqlc.ClearDefaultWorkbookPreferenceRefParams{
		IncidentID: pgUUID(incidentID),
		UpdatedAt:  pgTimestamptz(now),
	}); err != nil {
		return fmt.Errorf("clear startup default pointer: %w", err)
	}
	return nil
}

func savedViewRecordFromSavedviews(record savedviews.Record) SavedViewRecord {
	return SavedViewRecord{
		SavedViewID:      record.SavedViewID,
		IncidentID:       record.IncidentID,
		ViewSchemaID:     record.ViewSchemaID,
		Scope:            string(record.Scope),
		DisplayName:      record.DisplayName,
		QueryJSON:        cloneBytes(record.QueryJSON),
		LayoutJSON:       cloneBytes(record.LayoutJSON),
		OwnerUserID:      record.OwnerUserID,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
		SavedViewVersion: record.SavedViewVersion,
	}
}

func sheetRefParam(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return value
}

func defaultPreferencesRecordFromSQL(row sqlc.IncidentWorkbookPreference) (DefaultPreferencesRecord, error) {
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return DefaultPreferencesRecord{}, fmt.Errorf("default preferences incident id: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return DefaultPreferencesRecord{}, fmt.Errorf("default preferences created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return DefaultPreferencesRecord{}, fmt.Errorf("default preferences updated at: %w", err)
	}
	return DefaultPreferencesRecord{
		IncidentID:      incidentID,
		DefaultSheetRef: cloneBytes(row.DefaultSheetRef),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		UpdatedByUserID: optionalUUIDFromPG(row.UpdatedByUserID),
	}, nil
}

func userPreferencesRecordFromSQL(row sqlc.UserWorkbookPreference) (UserPreferencesRecord, error) {
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return UserPreferencesRecord{}, fmt.Errorf("user preferences incident id: %w", err)
	}
	userID, err := uuidFromPG(row.UserID)
	if err != nil {
		return UserPreferencesRecord{}, fmt.Errorf("user preferences user id: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return UserPreferencesRecord{}, fmt.Errorf("user preferences created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return UserPreferencesRecord{}, fmt.Errorf("user preferences updated at: %w", err)
	}
	return UserPreferencesRecord{
		IncidentID:   incidentID,
		UserID:       userID,
		HomeSheetRef: cloneBytes(row.HomeSheetRef),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
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
	id, err := uuid.FromBytes(value.Bytes[:])
	if err != nil {
		return nil
	}
	return &id
}

func pgTimestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return append([]byte(nil), value...)
}
