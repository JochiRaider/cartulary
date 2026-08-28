package performancefixture

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type ProductionApplication struct {
	actor     authn.UserRecord
	incidents *incidents.Application
	users     *authn.Store
}

func NewProductionApplication(pool postgres.DB, actor authn.UserRecord) (*ProductionApplication, error) {
	if pool == nil {
		return nil, fmt.Errorf("incidents performance fixture Postgres is required")
	}
	if actor.ID == uuid.Nil || !actor.IsActive || !actor.IsDeploymentAdmin {
		return nil, fmt.Errorf("incidents performance fixture requires an active deployment-admin actor")
	}
	incidentApplication, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            pool,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("construct Incidents performance fixture application: %w", err)
	}
	return &ProductionApplication{
		actor: actor, incidents: incidentApplication, users: authn.NewStore(pool),
	}, nil
}

func (a *ProductionApplication) CreateFixtureWorkspaceIncident(ctx context.Context, seed int) (string, error) {
	clientTxnID := fmt.Sprintf("performance-fixture-incident-%d", seed)
	request, err := admitIncidentCreate(map[string]any{
		"client_txn_id": clientTxnID,
		"incident_key":  fmt.Sprintf("AC043-PERF-%d", seed),
		"title":         "Large-grid performance fixture",
	})
	if err != nil {
		return "", err
	}
	result, err := a.incidents.CreateIncident(ctx, a.actor, request, "req-"+clientTxnID)
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
	request, admissionErr := admitMembershipCreate(map[string]any{
		"client_txn_id": clientTxnID,
		"user_id":       userUUID.String(),
		"role":          role,
	})
	if admissionErr != nil {
		return admissionErr
	}
	_, err = a.incidents.CreateMembership(ctx, a.actor, incidentUUID, target, request, "req-"+clientTxnID)
	return err
}

func admitIncidentCreate(body map[string]any) (incidents.IncidentCreateAdmission, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return incidents.IncidentCreateAdmission{}, err
	}
	request, admissionErr := incidents.AdmitIncidentCreateJSON(bytes.NewReader(encoded))
	if admissionErr != nil {
		return incidents.IncidentCreateAdmission{}, admissionErr
	}
	return request, nil
}

func admitMembershipCreate(body map[string]any) (incidents.MembershipCreateAdmission, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return incidents.MembershipCreateAdmission{}, err
	}
	request, admissionErr := incidents.AdmitMembershipCreateJSON(bytes.NewReader(encoded))
	if admissionErr != nil {
		return incidents.MembershipCreateAdmission{}, admissionErr
	}
	return request, nil
}
