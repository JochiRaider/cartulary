package incidents

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Access interface {
	GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error)
	GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error)
	EnsureOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
	IsIncidentClosed(err error) bool
}

type AccessService struct {
	store *Store
}

func NewAccess(db postgres.DB) *AccessService {
	return &AccessService{store: NewStore(db)}
}

func AccessFromStore(store *Store) *AccessService {
	return &AccessService{store: store}
}

func (s *AccessService) GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error) {
	return s.store.GetVisibleIncident(ctx, incidentID, userID)
}

func (s *AccessService) GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	return s.store.GetIncidentMembershipForUser(ctx, incidentID, userID)
}

func (s *AccessService) EnsureOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return EnsureIncidentOpenTx(ctx, tx, incidentID)
}

func (s *AccessService) IsIncidentClosed(err error) bool {
	return errors.Is(err, ErrIncidentClosed)
}
