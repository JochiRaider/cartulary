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
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool postgres.DB
}

var (
	ErrSavedViewNotFound        = errors.New("savedviews: saved view not found")
	ErrSavedViewVersionConflict = errors.New("savedviews: saved view version conflict")
	ErrSavedViewMutationDenied  = errors.New("savedviews: saved view mutation denied")
)

type SavedViewVersionConflictError struct {
	SavedViewID             uuid.UUID
	BaseSavedViewVersion    int64
	CurrentSavedViewVersion int64
}

func (e *SavedViewVersionConflictError) Error() string {
	return ErrSavedViewVersionConflict.Error()
}

func (e *SavedViewVersionConflictError) Unwrap() error {
	return ErrSavedViewVersionConflict
}

func (e *SavedViewVersionConflictError) Details() map[string]any {
	return map[string]any{
		"saved_view_id":              e.SavedViewID.String(),
		"base_saved_view_version":    e.BaseSavedViewVersion,
		"current_saved_view_version": e.CurrentSavedViewVersion,
	}
}

type Record struct {
	SavedViewID      uuid.UUID
	IncidentID       uuid.UUID
	ViewSchemaID     string
	Scope            Scope
	DisplayName      string
	QueryJSON        []byte
	LayoutJSON       []byte
	OwnerUserID      *uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
	SavedViewVersion int64
}

type ListPosition struct {
	UpdatedAt   time.Time
	SavedViewID uuid.UUID
}

type ListPageRequest struct {
	AnchorUpdatedAt *time.Time
	After           *ListPosition
	Limit           int
}

func NewStore(pool postgres.DB) *Store {
	return &Store{pool: pool}
}

func (s *Store) Create(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, now time.Time) (Record, error) {
	row, err := sqlc.New(s.pool).CreateSavedView(ctx, sqlc.CreateSavedViewParams{
		IncidentID:   pgUUID(incidentID),
		ViewSchemaID: request.ViewSchemaID,
		Scope:        string(request.Scope),
		DisplayName:  request.DisplayName,
		Column5:      request.QueryJSON,
		Column6:      request.LayoutJSON,
		OwnerUserID:  pgUUID(actor.ID),
		CreatedAt:    pgtype.Timestamptz{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return Record{}, fmt.Errorf("create saved view: %w", err)
	}
	return recordFromSQL(row)
}

func (s *Store) CreateSystemFixture(ctx context.Context, incidentID uuid.UUID, request CreateRequest, now time.Time) (Record, error) {
	row, err := sqlc.New(s.pool).CreateSavedView(ctx, sqlc.CreateSavedViewParams{
		IncidentID:   pgUUID(incidentID),
		ViewSchemaID: request.ViewSchemaID,
		Scope:        string(ScopeSystem),
		DisplayName:  request.DisplayName,
		Column5:      request.QueryJSON,
		Column6:      request.LayoutJSON,
		OwnerUserID:  pgtype.UUID{},
		CreatedAt:    pgtype.Timestamptz{Time: now.UTC(), Valid: true},
	})
	if err != nil {
		return Record{}, fmt.Errorf("create system saved view fixture: %w", err)
	}
	return recordFromSQL(row)
}

func (s *Store) ListVisible(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID, page ListPageRequest) ([]Record, error) {
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
	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		record, err := recordFromSQL(row)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

func (s *Store) GetVisibleForUpdate(ctx context.Context, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (Record, error) {
	row, err := sqlc.New(s.pool).GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrSavedViewNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("get visible saved view: %w", err)
	}
	return recordFromSQL(row)
}

func GetVisibleForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (Record, error) {
	row, err := sqlc.New(tx).GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(userID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrSavedViewNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("get visible saved view: %w", err)
	}
	return recordFromSQL(row)
}

func ResolveStartupVisibleForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, savedViewID uuid.UUID, userID uuid.UUID) (Record, string, error) {
	record, err := GetVisibleForUpdateTx(ctx, tx, incidentID, savedViewID, userID)
	if err == nil {
		return record, "", nil
	}
	if !errors.Is(err, ErrSavedViewNotFound) {
		return Record{}, "", err
	}
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM saved_views
     WHERE incident_id = $1
       AND saved_view_id = $2
)`, incidentID, savedViewID).Scan(&exists); err != nil {
		return Record{}, "", fmt.Errorf("check startup saved view visibility: %w", err)
	}
	if exists {
		return Record{}, "saved_view_not_visible", nil
	}
	return Record{}, "saved_view_not_found", nil
}

func (s *Store) Patch(ctx context.Context, actor authn.UserRecord, membershipRole string, incidentID uuid.UUID, savedViewID uuid.UUID, request PatchRequest, now time.Time) (Record, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Record{}, fmt.Errorf("begin saved view patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := sqlc.New(tx)
	row, err := q.GetVisibleSavedViewForUpdate(ctx, sqlc.GetVisibleSavedViewForUpdateParams{
		IncidentID:  pgUUID(incidentID),
		SavedViewID: pgUUID(savedViewID),
		UserID:      pgUUID(actor.ID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrSavedViewNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("get saved view for patch: %w", err)
	}
	current, err := recordFromSQL(row)
	if err != nil {
		return Record{}, err
	}
	if !CanMutate(current, actor.ID, membershipRole) {
		return Record{}, ErrSavedViewMutationDenied
	}
	if current.SavedViewVersion != request.BaseSavedViewVersion {
		return Record{}, &SavedViewVersionConflictError{
			SavedViewID:             current.SavedViewID,
			BaseSavedViewVersion:    request.BaseSavedViewVersion,
			CurrentSavedViewVersion: current.SavedViewVersion,
		}
	}

	next, changed, err := ApplyPatch(current, request, now)
	if err != nil {
		return Record{}, fmt.Errorf("apply saved view patch: %w", err)
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
		return Record{}, fmt.Errorf("update saved view: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return Record{}, fmt.Errorf("commit saved view patch transaction: %w", err)
	}
	return recordFromSQL(updated)
}

func (s *Store) Delete(ctx context.Context, actor authn.UserRecord, membershipRole string, incidentID uuid.UUID, savedViewID uuid.UUID) error {
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
		UserID:      pgUUID(actor.ID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrSavedViewNotFound
	}
	if err != nil {
		return fmt.Errorf("get saved view for delete: %w", err)
	}
	current, err := recordFromSQL(row)
	if err != nil {
		return err
	}
	if !CanMutate(current, actor.ID, membershipRole) {
		return ErrSavedViewMutationDenied
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

func recordFromSQL(row sqlc.SavedView) (Record, error) {
	savedViewID, err := uuidFromPG(row.SavedViewID)
	if err != nil {
		return Record{}, fmt.Errorf("saved view id: %w", err)
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return Record{}, fmt.Errorf("saved view incident id: %w", err)
	}
	createdAt, err := timeFromPG(row.CreatedAt)
	if err != nil {
		return Record{}, fmt.Errorf("saved view created at: %w", err)
	}
	updatedAt, err := timeFromPG(row.UpdatedAt)
	if err != nil {
		return Record{}, fmt.Errorf("saved view updated at: %w", err)
	}
	return Record{
		SavedViewID:      savedViewID,
		IncidentID:       incidentID,
		ViewSchemaID:     row.ViewSchemaID,
		Scope:            Scope(row.Scope),
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
