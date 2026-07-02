package startup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

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
	row := s.db.QueryRow(ctx, `
SELECT incident_id, default_sheet_ref, created_at, updated_at, updated_by_user_id
  FROM incident_workbook_preferences
 WHERE incident_id = $1
`, incidentID)
	var record DefaultPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.DefaultSheetRef, &record.CreatedAt, &record.UpdatedAt, &record.UpdatedByUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DefaultPreferencesRecord{}, ErrPreferencesNotFound
		}
		return DefaultPreferencesRecord{}, err
	}
	return record, nil
}

func (s *Store) PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error) {
	row := s.db.QueryRow(ctx, `
INSERT INTO incident_workbook_preferences (incident_id, default_sheet_ref, created_at, updated_at, updated_by_user_id)
VALUES ($1, $2::jsonb, $3, $3, $4)
ON CONFLICT (incident_id) DO UPDATE
   SET default_sheet_ref = EXCLUDED.default_sheet_ref,
       updated_at = CASE
           WHEN incident_workbook_preferences.default_sheet_ref IS NOT DISTINCT FROM EXCLUDED.default_sheet_ref
           THEN incident_workbook_preferences.updated_at
           ELSE EXCLUDED.updated_at
       END,
       updated_by_user_id = CASE
           WHEN incident_workbook_preferences.default_sheet_ref IS NOT DISTINCT FROM EXCLUDED.default_sheet_ref
           THEN incident_workbook_preferences.updated_by_user_id
           ELSE EXCLUDED.updated_by_user_id
       END
RETURNING incident_id, default_sheet_ref, created_at, updated_at, updated_by_user_id
`, incidentID, sheetRefParam(defaultSheetRef), now.UTC(), actorUserID)
	var record DefaultPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.DefaultSheetRef, &record.CreatedAt, &record.UpdatedAt, &record.UpdatedByUserID); err != nil {
		return DefaultPreferencesRecord{}, err
	}
	return record, nil
}

func (s *Store) GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error) {
	row := s.db.QueryRow(ctx, `
SELECT incident_id, user_id, home_sheet_ref, created_at, updated_at
  FROM user_workbook_preferences
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, userID)
	var record UserPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.UserID, &record.HomeSheetRef, &record.CreatedAt, &record.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserPreferencesRecord{}, ErrPreferencesNotFound
		}
		return UserPreferencesRecord{}, err
	}
	return record, nil
}

func (s *Store) PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error) {
	row := s.db.QueryRow(ctx, `
INSERT INTO user_workbook_preferences (incident_id, user_id, home_sheet_ref, created_at, updated_at)
VALUES ($1, $2, $3::jsonb, $4, $4)
ON CONFLICT (incident_id, user_id) DO UPDATE
   SET home_sheet_ref = EXCLUDED.home_sheet_ref,
       updated_at = CASE
           WHEN user_workbook_preferences.home_sheet_ref IS NOT DISTINCT FROM EXCLUDED.home_sheet_ref
           THEN user_workbook_preferences.updated_at
           ELSE EXCLUDED.updated_at
       END
RETURNING incident_id, user_id, home_sheet_ref, created_at, updated_at
`, incidentID, userID, sheetRefParam(homeSheetRef), now.UTC())
	var record UserPreferencesRecord
	if err := row.Scan(&record.IncidentID, &record.UserID, &record.HomeSheetRef, &record.CreatedAt, &record.UpdatedAt); err != nil {
		return UserPreferencesRecord{}, err
	}
	return record, nil
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
	var ref []byte
	err := tx.QueryRow(ctx, `
	SELECT home_sheet_ref
	  FROM user_workbook_preferences
	 WHERE incident_id = $1
	   AND user_id = $2
	 FOR UPDATE
	`, incidentID, userID).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup user preference: %w", err)
	}
	return ref, nil
}

func getDefaultPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]byte, error) {
	var ref []byte
	err := tx.QueryRow(ctx, `
	SELECT default_sheet_ref
	  FROM incident_workbook_preferences
	 WHERE incident_id = $1
	 FOR UPDATE
	`, incidentID).Scan(&ref)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup default preference: %w", err)
	}
	return ref, nil
}

func clearUserPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
	UPDATE user_workbook_preferences
	   SET home_sheet_ref = NULL,
	       updated_at = $3
	 WHERE incident_id = $1
	   AND user_id = $2
	`, incidentID, userID, now.UTC()); err != nil {
		return fmt.Errorf("clear startup home pointer: %w", err)
	}
	return nil
}

func clearDefaultPreferenceRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
	UPDATE incident_workbook_preferences
	   SET default_sheet_ref = NULL,
	       updated_at = $2
	 WHERE incident_id = $1
	`, incidentID, now.UTC()); err != nil {
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

func sheetRefParam(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return value
}

func cloneBytes(value []byte) []byte {
	if len(value) == 0 {
		return nil
	}
	return append([]byte(nil), value...)
}
