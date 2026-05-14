package savedviews

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool postgres.DB
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
