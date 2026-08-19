package workbookassembly

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
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
	preferences := workbookstartuppostgres.NewRepository(db)
	unitOfWork := startupUnitOfWork{db: db}
	return workbookstartup.NewStore(preferences, unitOfWork, workspaceResolver)
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
		preferences: workbookstartuppostgres.NewSession(tx),
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
	preferences workbookstartuppostgres.Session
	tx          pgx.Tx
}

func (s startupSession) UserPreferenceRef(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
) ([]byte, error) {
	return s.preferences.UserPreferenceRef(ctx, incidentID, userID)
}

func (s startupSession) DefaultPreferenceRef(
	ctx context.Context,
	incidentID uuid.UUID,
) ([]byte, error) {
	return s.preferences.DefaultPreferenceRef(ctx, incidentID)
}

func (s startupSession) ClearUserPreferenceIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	userID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	return s.preferences.ClearUserPreferenceIfCurrent(ctx, incidentID, userID, expected, now)
}

func (s startupSession) ClearDefaultPreferenceIfCurrent(
	ctx context.Context,
	incidentID uuid.UUID,
	actorUserID uuid.UUID,
	expected []byte,
	now time.Time,
) (bool, error) {
	return s.preferences.ClearDefaultPreferenceIfCurrent(ctx, incidentID, actorUserID, expected, now)
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

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}
