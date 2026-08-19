package collaboration

import (
	"context"

	"github.com/google/uuid"
)

type IncidentSessionNotifier struct {
	hub *Hub
}

func NewIncidentSessionNotifier(hub *Hub) *IncidentSessionNotifier {
	return &IncidentSessionNotifier{hub: hub}
}

func (n *IncidentSessionNotifier) NotifyIncidentClosed(ctx context.Context, effectKey uuid.UUID, incidentID uuid.UUID) {
	_ = ctx
	_ = effectKey
	if n == nil || n.hub == nil {
		return
	}
	n.hub.TerminateIncident(incidentID, IncidentTerminalClosed)
}

func (n *IncidentSessionNotifier) NotifyIncidentMembershipRevoked(ctx context.Context, effectKey uuid.UUID, incidentID uuid.UUID, userID uuid.UUID) {
	_ = ctx
	_ = effectKey
	if n == nil || n.hub == nil {
		return
	}
	n.hub.RevokeIncidentAccess(incidentID, userID)
}
