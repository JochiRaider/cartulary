package performancefixture

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ProductionApplication struct {
	actor     authn.UserRecord
	incidents *incidents.Application
	users     *authn.Store
	now       func() time.Time
}

func NewProductionApplication(pool postgres.DB, actor authn.UserRecord) (*ProductionApplication, error) {
	if pool == nil {
		return nil, fmt.Errorf("incidents performance fixture Postgres is required")
	}
	if actor.ID == uuid.Nil || !actor.IsActive || !actor.IsDeploymentAdmin {
		return nil, fmt.Errorf("incidents performance fixture requires an active deployment-admin actor")
	}
	return &ProductionApplication{
		actor: actor, incidents: incidents.NewApplication(pool), users: authn.NewStore(pool), now: time.Now,
	}, nil
}

func (a *ProductionApplication) CreateFixtureWorkspaceIncident(ctx context.Context, seed int) (string, error) {
	clientTxnID := fmt.Sprintf("performance-fixture-incident-%d", seed)
	request := incidents.CreateIncidentRequest{
		ClientTxnID: clientTxnID,
		IncidentKey: fmt.Sprintf("AC043-PERF-%d", seed),
		Title:       "Large-grid performance fixture",
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256(payload)
	result, err := a.incidents.CreateIncident(ctx, a.actor, request, hash[:], "req-"+clientTxnID, a.now().UTC())
	if err != nil {
		return "", err
	}
	return result.Incident.ID.String(), nil
}

func (a *ProductionApplication) AddFixtureMembership(ctx context.Context, incidentID string, userID string, role string) error {
	incidentUUID, err := uuid.Parse(incidentID)
	if err != nil {
		return err
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}
	target, err := a.users.GetUserByID(ctx, userUUID)
	if err != nil {
		return err
	}
	clientTxnID := "performance-fixture-membership-" + userUUID.String()
	request := incidents.MembershipCreateRequest{ClientTxnID: clientTxnID, Role: role}
	payload, err := json.Marshal(request)
	if err != nil {
		return err
	}
	hash := sha256.Sum256(payload)
	_, err = a.incidents.CreateMembership(ctx, a.actor, incidentUUID, target, request, hash[:], "req-"+clientTxnID, a.now().UTC())
	return err
}
