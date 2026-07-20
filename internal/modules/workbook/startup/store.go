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
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrPreferencesNotFound = errors.New("workbook startup: preferences not found")
)

type Store struct {
	db                    postgres.DB
	preferences           preferenceRepository
	savedViews            savedViewResolver
	workspaceAvailability WorkspaceResolver
}

func NewStore(db postgres.DB, workspaceResolvers ...WorkspaceResolver) *Store {
	resolver := WorkspaceResolver(NewWorkspaceRegistry(nil))
	if len(workspaceResolvers) > 0 && workspaceResolvers[0] != nil {
		resolver = workspaceResolvers[0]
	}
	return &Store{
		db:                    db,
		preferences:           sqlPreferenceRepository{db: db},
		savedViews:            sqlSavedViewResolver{},
		workspaceAvailability: resolver,
	}
}

type preferenceRepository interface {
	GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error)
	PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error)
	GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error)
	PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error)
	GetUserPreferenceRefForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error)
	GetDefaultPreferenceRefForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]byte, error)
	ClearUserPreferenceRef(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, now time.Time) error
	ClearDefaultPreferenceRef(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) error
}

type savedViewResolver interface {
	ResolveStartupVisibleForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (SavedViewRecord, string, error)
}

type sqlPreferenceRepository struct {
	db postgres.DB
}

type sqlSavedViewResolver struct{}

func (s *Store) ValidatePreferenceSheetRef(raw []byte, role string, field string) *httpapi.APIError {
	if len(raw) == 0 {
		return nil
	}
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return invalidMutationPayload(field, reasonCode)
	}
	if ref.Kind != "extension_workspace" {
		return nil
	}
	if reasonCode = s.workspaceAvailability.ResolveWorkspace(ref, role); reasonCode == "" {
		return nil
	}
	return invalidMutationPayload(extensionReasonField(field, reasonCode), reasonCode)
}

func (s *Store) ValidateExplicitSheetRef(raw []byte, role string) *httpapi.APIError {
	if len(raw) == 0 {
		return nil
	}
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return invalidStartupRequest("sheet_ref", reasonCode)
	}
	if ref.Kind != "extension_workspace" {
		return nil
	}
	if reasonCode = s.workspaceAvailability.ResolveWorkspace(ref, role); reasonCode == "" {
		return nil
	}
	return invalidStartupRequest(extensionReasonField("sheet_ref", reasonCode), reasonCode)
}

func extensionReasonField(prefix string, reasonCode string) string {
	switch reasonCode {
	case "invalid_extension_profile_id", "extension_profile_not_claimed":
		if prefix == "sheet_ref" {
			return "extension_profile_id"
		}
		return prefix + ".extension_profile_id"
	case "invalid_extension_workspace_key", "extension_workspace_unavailable", "extension_workspace_not_visible":
		if prefix == "sheet_ref" {
			return "sheet_ref_id"
		}
		return prefix + ".workspace_key"
	default:
		return prefix
	}
}

func (s *Store) GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error) {
	return s.preferences.GetDefaultPreferences(ctx, incidentID)
}

func (s *Store) PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error) {
	return s.preferences.PutDefaultPreferences(ctx, incidentID, actorUserID, defaultSheetRef, now)
}

func (s *Store) GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error) {
	return s.preferences.GetUserPreferences(ctx, incidentID, userID)
}

func (s *Store) PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error) {
	return s.preferences.PutUserPreferences(ctx, incidentID, userID, homeSheetRef, now)
}

func (r sqlPreferenceRepository) GetDefaultPreferences(ctx context.Context, incidentID uuid.UUID) (DefaultPreferencesRecord, error) {
	row, err := sqlc.New(r.db).GetDefaultWorkbookPreferences(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return DefaultPreferencesRecord{}, ErrPreferencesNotFound
	}
	if err != nil {
		return DefaultPreferencesRecord{}, err
	}
	return defaultPreferencesRecordFromSQL(row)
}

func (r sqlPreferenceRepository) PutDefaultPreferences(ctx context.Context, incidentID uuid.UUID, actorUserID uuid.UUID, defaultSheetRef []byte, now time.Time) (DefaultPreferencesRecord, error) {
	row, err := sqlc.New(r.db).PutDefaultWorkbookPreferences(ctx, sqlc.PutDefaultWorkbookPreferencesParams{
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

func (r sqlPreferenceRepository) GetUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (UserPreferencesRecord, error) {
	row, err := sqlc.New(r.db).GetUserWorkbookPreferences(ctx, sqlc.GetUserWorkbookPreferencesParams{
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

func (r sqlPreferenceRepository) PutUserPreferences(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, homeSheetRef []byte, now time.Time) (UserPreferencesRecord, error) {
	row, err := sqlc.New(r.db).PutUserWorkbookPreferences(ctx, sqlc.PutUserWorkbookPreferencesParams{
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

func (s *Store) Resolve(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, role string, explicitSheetRef []byte, now time.Time) (Record, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Record{}, fmt.Errorf("begin workbook startup transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	homeRef, err := s.preferences.GetUserPreferenceRefForUpdate(ctx, tx, incidentID, userID)
	if err != nil {
		return Record{}, err
	}
	defaultRef, err := s.preferences.GetDefaultPreferenceRefForUpdate(ctx, tx, incidentID)
	if err != nil {
		return Record{}, err
	}

	record := Record{
		IncidentID:      incidentID,
		HomeSheetRef:    cloneBytes(homeRef),
		DefaultSheetRef: cloneBytes(defaultRef),
		ExtensionWorkspaceAvailability: ExtensionWorkspaceAvailability{
			SchemaID:   ExtensionWorkspaceAvailabilitySchemaID,
			IncidentID: incidentID.String(),
			Workspaces: s.workspaceAvailability.AvailableWorkspaces(role),
		},
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
		resolved, invalidReason, err := resolveCandidate(ctx, tx, incidentID, userID, role, candidate.ref, s.savedViews, s.workspaceAvailability)
		if err != nil {
			return Record{}, err
		}
		if invalidReason == "" {
			record.SelectedSheetRef = mustSheetRefJSON(resolved.SheetRef)
			if resolved.ViewSchemaID != "" {
				viewSchemaID := resolved.ViewSchemaID
				record.SelectedViewSchemaID = &viewSchemaID
			}
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
			if err := s.preferences.ClearUserPreferenceRef(ctx, tx, incidentID, userID, now); err != nil {
				return Record{}, err
			}
			record.HomeSheetRef = nil
		}
		if candidate.source == SourceDefault {
			if err := s.preferences.ClearDefaultPreferenceRef(ctx, tx, incidentID, now); err != nil {
				return Record{}, err
			}
			record.DefaultSheetRef = nil
		}
	}

	record.SelectedSheetRef = mustSheetRefJSON(SheetRef{Kind: "view_schema", ID: TimelineViewSchemaID})
	timelineViewSchemaID := TimelineViewSchemaID
	record.SelectedViewSchemaID = &timelineViewSchemaID
	record.Source = SourceTimeline
	if err := tx.Commit(ctx); err != nil {
		return Record{}, fmt.Errorf("commit workbook startup fallback transaction: %w", err)
	}
	return record, nil
}

type resolvedCandidate struct {
	ViewSchemaID string
	SavedView    *SavedViewRecord
	SheetRef     SheetRef
}

func resolveCandidate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, role string, raw []byte, savedViews savedViewResolver, workspaceAvailability WorkspaceResolver) (resolvedCandidate, string, error) {
	ref, reasonCode := decodeSheetRef(raw)
	if reasonCode != "" {
		return resolvedCandidate{}, reasonCode, nil
	}
	switch ref.Kind {
	case "view_schema":
		if _, ok := viewschema.Lookup(ref.ID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		return resolvedCandidate{ViewSchemaID: ref.ID, SheetRef: ref}, "", nil
	case "saved_view":
		savedViewID, err := uuid.Parse(ref.ID)
		if err != nil {
			return resolvedCandidate{}, "invalid_saved_view_id", nil
		}
		savedView, reasonCode, err := savedViews.ResolveStartupVisibleForUpdate(ctx, tx, incidentID, savedViewID, userID)
		if err != nil {
			return resolvedCandidate{}, "", err
		}
		if reasonCode != "" {
			return resolvedCandidate{}, reasonCode, nil
		}
		if _, ok := viewschema.Lookup(savedView.ViewSchemaID); !ok {
			return resolvedCandidate{}, "unknown_view_schema", nil
		}
		return resolvedCandidate{ViewSchemaID: savedView.ViewSchemaID, SavedView: &savedView, SheetRef: ref}, "", nil
	case "extension_workspace":
		if reasonCode := workspaceAvailability.ResolveWorkspace(ref, role); reasonCode != "" {
			return resolvedCandidate{}, reasonCode, nil
		}
		return resolvedCandidate{SheetRef: ref}, "", nil
	default:
		return resolvedCandidate{}, "unsupported_sheet_ref_kind", nil
	}
}

func (r sqlPreferenceRepository) GetUserPreferenceRefForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID) ([]byte, error) {
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

func (r sqlPreferenceRepository) GetDefaultPreferenceRefForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) ([]byte, error) {
	ref, err := sqlc.New(tx).GetDefaultWorkbookPreferenceRefForUpdate(ctx, pgUUID(incidentID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query startup default preference: %w", err)
	}
	return ref, nil
}

func (r sqlPreferenceRepository) ClearUserPreferenceRef(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, now time.Time) error {
	if err := sqlc.New(tx).ClearUserWorkbookPreferenceRef(ctx, sqlc.ClearUserWorkbookPreferenceRefParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		UpdatedAt:  pgTimestamptz(now),
	}); err != nil {
		return fmt.Errorf("clear startup home pointer: %w", err)
	}
	return nil
}

func (r sqlPreferenceRepository) ClearDefaultPreferenceRef(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, now time.Time) error {
	if err := sqlc.New(tx).ClearDefaultWorkbookPreferenceRef(ctx, sqlc.ClearDefaultWorkbookPreferenceRefParams{
		IncidentID: pgUUID(incidentID),
		UpdatedAt:  pgTimestamptz(now),
	}); err != nil {
		return fmt.Errorf("clear startup default pointer: %w", err)
	}
	return nil
}

func (r sqlSavedViewResolver) ResolveStartupVisibleForUpdate(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (SavedViewRecord, string, error) {
	record, reasonCode, err := savedviews.ResolveStartupVisibleForUpdateTx(ctx, tx, incidentID, savedViewID, userID)
	if err != nil || reasonCode != "" {
		return SavedViewRecord{}, reasonCode, err
	}
	return savedViewRecordFromSavedviews(record), "", nil
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
