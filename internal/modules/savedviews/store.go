package savedviews

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

type postgresSavedViewRepository struct {
	pool postgres.DB
}

var (
	errSavedViewNotFound        = errors.New("savedviews: saved view not found")
	errSavedViewVersionConflict = errors.New("savedviews: saved view version conflict")
	errSavedViewMutationDenied  = errors.New("savedviews: saved view mutation denied")
)

type savedViewVersionConflictError struct {
	SavedViewID             uuid.UUID
	BaseSavedViewVersion    int64
	CurrentSavedViewVersion int64
}

func (e *savedViewVersionConflictError) Error() string {
	return errSavedViewVersionConflict.Error()
}

func (e *savedViewVersionConflictError) Unwrap() error {
	return errSavedViewVersionConflict
}

func (e *savedViewVersionConflictError) Details() map[string]any {
	return map[string]any{
		"saved_view_id":              e.SavedViewID.String(),
		"base_saved_view_version":    e.BaseSavedViewVersion,
		"current_saved_view_version": e.CurrentSavedViewVersion,
	}
}

type savedViewRecord struct {
	SavedViewID      uuid.UUID
	IncidentID       uuid.UUID
	ViewSchemaID     string
	Scope            scope
	DisplayName      string
	QueryJSON        []byte
	LayoutJSON       []byte
	OwnerUserID      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SavedViewVersion int64
}

// StartupRecord is the purpose-specific Saved Views contribution consumed by
// workbook startup assembly. It intentionally exposes no persistence methods.
type StartupRecord struct {
	SavedViewID      uuid.UUID
	IncidentID       uuid.UUID
	ViewSchemaID     string
	Scope            string
	DisplayName      string
	QueryJSON        []byte
	LayoutJSON       []byte
	OwnerUserID      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SavedViewVersion int64
}

type listPosition struct {
	UpdatedAt   time.Time
	SavedViewID uuid.UUID
}

type listPageRequest struct {
	AnchorUpdatedAt *time.Time
	After           *listPosition
	Limit           int
}

func newPostgresSavedViewRepository(pool postgres.DB) *postgresSavedViewRepository {
	return &postgresSavedViewRepository{pool: pool}
}

func (s *postgresSavedViewRepository) create(ctx context.Context, actorUserID uuid.UUID, incidentID uuid.UUID, request createRequest, now time.Time) (savedViewRecord, error) {
	row, err := sqlc.New(s.pool).CreateSavedView(ctx, sqlc.CreateSavedViewParams{
		IncidentID:   pgUUID(incidentID),
		ViewSchemaID: request.ViewSchemaID,
		Scope:        string(request.Scope),
		DisplayName:  request.DisplayName,
		Column5:      request.QueryJSON,
		Column6:      request.LayoutJSON,
		OwnerUserID:  pgUUID(actorUserID),
		CreatedAt:    pgtype.Timestamptz{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("create saved view: %w", err)
	}
	return recordFromSQL(row)
}

func (s *postgresSavedViewRepository) createSystemFixture(ctx context.Context, incidentID uuid.UUID, request createRequest, now time.Time) (savedViewRecord, error) {
	row, err := sqlc.New(s.pool).CreateSavedView(ctx, sqlc.CreateSavedViewParams{
		IncidentID:   pgUUID(incidentID),
		ViewSchemaID: request.ViewSchemaID,
		Scope:        string(scopeSystem),
		DisplayName:  request.DisplayName,
		Column5:      request.QueryJSON,
		Column6:      request.LayoutJSON,
		OwnerUserID:  pgtype.UUID{},
		CreatedAt:    pgtype.Timestamptz{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("create system saved view fixture: %w", err)
	}
	return recordFromSQL(row)
}

func (s *postgresSavedViewRepository) listVisible(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, page listPageRequest) ([]savedViewRecord, error) {
	if page.Limit < 1 {
		page.Limit = 1
	}
	params := sqlc.ListVisibleSavedViewsParams{
		IncidentID: pgUUID(incidentID),
		UserID:     pgUUID(userID),
		Limit:      int32(page.Limit),
	}
	if page.AnchorUpdatedAt != nil {
		params.Column3 = pgtype.Timestamptz{Time: page.AnchorUpdatedAt.UTC(), Valid: true}
	}
	if page.After != nil {
		params.Column4 = pgtype.Timestamptz{Time: page.After.UpdatedAt.UTC(), Valid: true}
		params.Column5 = pgUUID(page.After.SavedViewID)
	}
	rows, err := sqlc.New(s.pool).ListVisibleSavedViews(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list visible saved views: %w", err)
	}
	records := make([]savedViewRecord, 0, len(rows))
	for _, row := range rows {
		record, err := recordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *postgresSavedViewRepository) getVisibleForUpdate(ctx context.Context, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (savedViewRecord, error) {
	row, err := sqlc.New(s.pool).GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return savedViewRecord{}, errSavedViewNotFound
	}
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("get visible saved view: %w", err)
	}
	return recordFromSQL(row)
}

func getVisibleForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (savedViewRecord, error) {
	row, err := sqlc.New(tx).GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return savedViewRecord{}, errSavedViewNotFound
	}
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("get visible saved view: %w", err)
	}
	return recordFromSQL(row)
}

func ResolveStartupVisibleForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (StartupRecord, string, error) {
	record, err := getVisibleForUpdateTx(ctx, tx, incidentID, savedViewID, userID)
	if err == nil {
		return startupRecord(record), "", nil
	}
	if !errors.Is(err, errSavedViewNotFound) {
		return StartupRecord{}, "", err
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM saved_views
     WHERE incident_id = $1
       AND saved_view_id = $2
)`, incidentID, savedViewID).Scan(&exists); err != nil {
		return StartupRecord{}, "", fmt.Errorf("check startup saved view visibility: %w", err)
	}
	if exists {
		return StartupRecord{}, "saved_view_not_visible", nil
	}
	return StartupRecord{}, "saved_view_not_found", nil
}

func startupRecord(record savedViewRecord) StartupRecord {
	return StartupRecord{
		SavedViewID:      record.SavedViewID,
		IncidentID:       record.IncidentID,
		ViewSchemaID:     record.ViewSchemaID,
		Scope:            string(record.Scope),
		DisplayName:      record.DisplayName,
		QueryJSON:        append([]byte(nil), record.QueryJSON...),
		LayoutJSON:       append([]byte(nil), record.LayoutJSON...),
		OwnerUserID:      record.OwnerUserID,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
		SavedViewVersion: record.SavedViewVersion,
	}
}

func (s *postgresSavedViewRepository) patchVisible(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	actorUserID uuid.UUID,
	mutate func(savedViewRecord) (savedViewRecord, bool, error),
) (savedViewRecord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("begin saved view patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(tx)
	row, err := q.GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(actorUserID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return savedViewRecord{}, errSavedViewNotFound
	}
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("get saved view for patch: %w", err)
	}
	current, err := recordFromSQL(row)
	if err != nil {
		return savedViewRecord{}, err
	}
	next, changed, err := mutate(current)
	if err != nil {
		return savedViewRecord{}, err
	}
	if !changed {
		return current, nil
	}
	updated, err := q.UpdateSavedView(ctx, sqlc.UpdateSavedViewParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		Scope:       string(next.Scope),
		DisplayName: next.DisplayName,
		Column5:     next.QueryJSON,
		Column6:     next.LayoutJSON,
		UpdatedAt:   pgtype.Timestamptz{Time: next.UpdatedAt.UTC(), Valid: true},
	})
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("update saved view: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return savedViewRecord{}, fmt.Errorf("commit saved view patch transaction: %w", err)
	}
	return recordFromSQL(updated)
}

func (s *postgresSavedViewRepository) deleteVisible(
	ctx context.Context,
	incidentID uuid.UUID,
	savedViewID uuid.UUID,
	actorUserID uuid.UUID,
	authorize func(savedViewRecord) error,
) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin saved view delete transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(tx)
	row, err := q.GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(actorUserID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return errSavedViewNotFound
	}
	if err != nil {
		return fmt.Errorf("get saved view for delete: %w", err)
	}
	current, err := recordFromSQL(row)
	if err != nil {
		return err
	}
	if err := authorize(current); err != nil {
		return err
	}
	if err := q.DeleteSavedView(ctx, sqlc.DeleteSavedViewParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
	}); err != nil {
		return fmt.Errorf("delete saved view: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit saved view delete transaction: %w", err)
	}
	return nil
}

func recordFromSQL(row sqlc.SavedView) (savedViewRecord, error) {
	savedViewID, err := uuidFromPG(row.SavedViewID)
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("saved view id: %w", err)
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("saved view incident id: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("saved view created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return savedViewRecord{}, fmt.Errorf("saved view updated at: %w", err)
	}
	return savedViewRecord{
		SavedViewID:      savedViewID,
		IncidentID:       incidentID,
		ViewSchemaID:     row.ViewSchemaID,
		Scope:            scope(row.Scope),
		DisplayName:      row.DisplayName,
		QueryJSON:        append([]byte(nil), row.QueryJson...),
		LayoutJSON:       append([]byte(nil), row.LayoutJson...),
		OwnerUserID:      optionalUUIDFromPG(row.OwnerUserID),
		CreatedAt:        createdAt,
		UpdatedAt:        updatedAt,
		SavedViewVersion: row.SavedViewVersion,
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
