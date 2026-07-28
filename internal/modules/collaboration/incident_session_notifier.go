package collaboration

import (
	"context"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type IncidentSessionNotifier struct {
	authStore *authn.Store
	hub       *Hub
}

func NewIncidentSessionNotifier(db postgres.DB, hub *Hub) *IncidentSessionNotifier {
	return &IncidentSessionNotifier{
		authStore: authn.NewStore(db),
		hub:       hub,
	}
}

func (n *IncidentSessionNotifier) NotifyIncidentClosed(ctx context.Context, incidentID uuid.UUID) {
	_ = ctx
	if n == nil || n.hub == nil {
		return
	}
	n.hub.TerminateIncident(incidentID, IncidentTerminalClosed)
}

func (n *IncidentSessionNotifier) NotifyIncidentMembershipRevoked(ctx context.Context, incidentID uuid.UUID, userID uuid.UUID) {
	if n == nil || n.hub == nil || n.authStore == nil {
		return
	}
	sessions, err := n.authStore.ListActiveSessionsForUser(ctx, userID)
	if err != nil {
		return
	}
	for _, session := range sessions {
		n.hub.RevokeIncidentAccess(incidentID, session.ID)
	}
}
