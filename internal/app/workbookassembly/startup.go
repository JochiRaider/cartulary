package workbookassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/workbookpreferences"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewStartupStoreFromDependencies(deps httpapi.DependencySet) (*workbookstartup.Store, error) {
	return NewStartupStore(
		deps.PostgresHandle(),
		workbookstartup.NewWorkspaceRegistryFromPublication(httpapi.ExtensionWorkspacesFromDependencies(deps)),
	), nil
}

func NewStartupStore(db postgres.DB, workspaceResolver workbookstartup.WorkspaceResolver) *workbookstartup.Store {
	repository := workbookpreferences.NewRepository(db)
	preferences := preferenceAdapter{repository: repository}
	unitOfWork := startupUnitOfWork{db: db}
	return workbookstartup.NewStore(preferences, unitOfWork, workspaceResolver)
}

type preferenceAdapter struct {
	repository *workbookpreferences.Repository
}

func (a preferenceAdapter) GetDefaultPreferences(
	ctx context.Context,
	incidentID uuid.UUID,
) (workbookstartup.DefaultPreferencesRecord, error) {
	record, err := a.repository.GetDefault(ctx, incidentID)
	if errors.Is(err, workbookpreferences.ErrNotFound) {
		return workbookstartup.DefaultPreferencesRecord{}, workbookstartup.ErrPreferencesNotFound
	}
	if err != nil {
		return workbookstartup.DefaultPreferencesRecord{}, err
	}
	return defaultRecord(record), nil
}

func (a preferenceAdapter) PutDefaultPreferences(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	defaultSheetRef []byte,
	now time.Time,
) (workbookstartup.DefaultPreferencesRecord, error) {
	record, err := a.repository.PutDefault(ctx, incidentID, actorUserID, defaultSheetRef, now)
	if err != nil {
		return workbookstartup.DefaultPreferencesRecord{}, err
	}
	return defaultRecord(record), nil
}

func (a preferenceAdapter) GetUserPreferences(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) (workbookstartup.UserPreferencesRecord, error) {
	record, err := a.repository.GetUser(ctx, incidentID, userID)
	if errors.Is(err, workbookpreferences.ErrNotFound) {
		return workbookstartup.UserPreferencesRecord{}, workbookstartup.ErrPreferencesNotFound
	}
	if err != nil {
		return workbookstartup.UserPreferencesRecord{}, err
	}
	return userRecord(record), nil
}

func (a preferenceAdapter) PutUserPreferences(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	homeSheetRef []byte,
	now time.Time,
) (workbookstartup.UserPreferencesRecord, error) {
	record, err := a.repository.PutUser(ctx, incidentID, userID, homeSheetRef, now)
	if err != nil {
		return workbookstartup.UserPreferencesRecord{}, err
	}
	return userRecord(record), nil
}

type startupUnitOfWork struct {
	db postgres.DB
}

func (u startupUnitOfWork) Run(
	ctx context.Context,
	operation func(workbookstartup.Session) (workbookstartup.Record, error),
) (workbookstartup.Record, error) {
	tx, err := u.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return workbookstartup.Record{}, fmt.Errorf("begin workbook startup transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	result, err := operation(startupSession{
		preferences: workbookpreferences.NewSession(tx),
		tx:          tx,
	})
	if err != nil {
		return workbookstartup.Record{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return workbookstartup.Record{}, fmt.Errorf("commit workbook startup transaction: %w", err)
	}
	return result, nil
}

type startupSession struct {
	preferences workbookpreferences.Session
	tx          pgx.Tx
}

func (s startupSession) UserPreferenceRef(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) ([]byte, error) {
	return s.preferences.UserRefForUpdate(ctx, incidentID, userID)
}

func (s startupSession) DefaultPreferenceRef(
	ctx context.Context,
	incidentID uuid.UUID,
) ([]byte, error) {
	return s.preferences.DefaultRefForUpdate(ctx, incidentID)
}

func (s startupSession) ClearUserPreferenceIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	return s.preferences.ClearUserIfCurrent(ctx, incidentID, userID, expected, now)
}

func (s startupSession) ClearDefaultPreferenceIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	return s.preferences.ClearDefaultIfCurrent(ctx, incidentID, actorUserID, expected, now)
}

func (s startupSession) ResolveSavedView(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	userID uuid.UUID,
) (workbookstartup.SavedViewRecord, string, error) {
	record, reasonCode, err := savedviews.ResolveStartupVisibleForUpdateTx(
		ctx,
		s.tx,
		incidentID,
		savedViewID,
		userID,
	)
	if err != nil {
		return workbookstartup.SavedViewRecord{}, "", err
	}
	if reasonCode != "" {
		return workbookstartup.SavedViewRecord{}, reasonCode, nil
	}
	return workbookstartup.SavedViewRecord{
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
	}, "", nil
}

func defaultRecord(record workbookpreferences.DefaultRecord) workbookstartup.DefaultPreferencesRecord {
	return workbookstartup.DefaultPreferencesRecord{
		IncidentID:      record.IncidentID,
		DefaultSheetRef: cloneBytes(record.DefaultSheetRef),
		CreatedAt:       record.CreatedAt,
		UpdatedAt:       record.UpdatedAt,
		UpdatedByUserID: record.UpdatedByUserID,
	}
}

func userRecord(record workbookpreferences.UserRecord) workbookstartup.UserPreferencesRecord {
	return workbookstartup.UserPreferencesRecord{
		IncidentID:   record.IncidentID,
		UserID:       record.UserID,
		HomeSheetRef: cloneBytes(record.HomeSheetRef),
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}
