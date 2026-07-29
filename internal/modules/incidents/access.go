package incidents

import (
	"context"
	"errors"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Access interface {
	GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error)
	GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error)
	EnsureOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error
	AuthorizeMutationTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, userID uuid.UUID, roles ...string) (MembershipRecord, error)
	IsIncidentClosed(err error) bool
	IsIncidentNotFound(err error) bool
	IsMembershipNotFound(err error) bool
}

type AccessService struct {
	repository *repository
}

func NewAccess(db postgres.DB) *AccessService {
	return &AccessService{repository: newRepository(db)}
}

func (s *AccessService) GetVisibleIncident(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (IncidentRecord, error) {
	return s.repository.getVisibleIncident(ctx, incidentID, userID)
}

func (s *AccessService) GetIncidentMembershipForUser(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) (MembershipRecord, error) {
	return s.repository.getMembership(ctx, incidentID, userID)
}

func (s *AccessService) EnsureOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return newRepository(tx).ensureOpen(ctx, incidentID)
}

func (s *AccessService) AuthorizeMutationTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	userID uuid.UUID,
	roles ...string,
) (MembershipRecord, error) {
	repository := newRepository(tx)
	if err := repository.ensureOpen(ctx, incidentID); err != nil {
		return MembershipRecord{}, err
	}
	membership, err := repository.getMembershipForUpdate(ctx, incidentID, userID)
	if err != nil {
		return MembershipRecord{}, err
	}
	if len(roles) > 0 && !slices.Contains(roles, membership.Role) {
		return MembershipRecord{}, ErrIncidentRoleDenied
	}
	return membership, nil
}

func (s *AccessService) IsIncidentClosed(err error) bool {
	return errors.Is(err, ErrIncidentClosed)
}

func (s *AccessService) IsIncidentNotFound(err error) bool {
	return errors.Is(err, ErrIncidentNotFound)
}

func (s *AccessService) IsMembershipNotFound(err error) bool {
	return errors.Is(err, ErrMembershipNotFound)
}
