package appsupport

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

var ErrIncidentCreateCommitFault = errors.New("forced incident create final commit failure")

// IncidentCreateCommitFaultCapability is the only test-support installation
// surface for the Incidents-owned final-create-commit seam.
type IncidentCreateCommitFaultCapability struct {
	application *incidents.Application
}

func (c *IncidentCreateCommitFaultCapability) Create(
	ctx context.Context,
	actor authn.UserRecord,
	request incidents.CreateIncidentRequest,
	requestID string,
	now time.Time,
) error {
	if c == nil || c.application == nil {
		return errors.New("incident create commit-fault capability is not installed")
	}
	_, err := c.application.CreateIncident(
		ctx,
		actor,
		request,
		incidents.IncidentCreateRequestHash(request),
		requestID,
		now,
	)
	return err
}

type incidentCreateCommitFault struct{}

func (incidentCreateCommitFault) CommitIncidentCreate(ctx context.Context, tx pgx.Tx) error {
	_ = tx.Rollback(ctx)
	return ErrIncidentCreateCommitFault
}

func newIncidentCreateCommitFaultCapability(
	env map[string]string,
	mode httptestx.TestRouteMode,
	db postgres.DB,
) (*IncidentCreateCommitFaultCapability, error) {
	if mode != httptestx.TestRouteModeHarnessOwned && mode != httptestx.TestRouteModeCustomEnv {
		return nil, fmt.Errorf("incident create commit fault requires an admitted test-runtime mode, got %q", mode)
	}
	admissionEnv := make(map[string]string, len(env)+3)
	for key, value := range env {
		admissionEnv[key] = value
	}
	if mode == httptestx.TestRouteModeHarnessOwned {
		admissionEnv[httpapi.TestRoutesEnabledEnv] = "1"
		admissionEnv[httpapi.TestRuntimeMarkerEnv] = httpapi.TestRuntimeMarkerValue
		admissionEnv[httpapi.TestRouteTokenEnv] = httptestx.TestRouteToken
	}
	if _, err := httpapi.NewTestRouteGuard(admissionEnv); err != nil {
		return nil, fmt.Errorf("incident create commit fault requires validated test-runtime admission: %w", err)
	}
	return &IncidentCreateCommitFaultCapability{
		application: incidents.NewApplicationWithOptions(db, incidents.ApplicationOptions{
			PreferenceBootstrap:  workbookstartuppostgres.NewWriter(),
			IncidentCreateCommit: incidentCreateCommitFault{},
		}),
	}, nil
}
