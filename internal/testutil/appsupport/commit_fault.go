package appsupport

import (
	"context"
	"errors"
	"fmt"
	"reflect"
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

// IncidentCreateCommitFaultCapability is the test-support installation surface
// for the database decorator that faults the final incident-create commit.
type IncidentCreateCommitFaultCapability struct {
	application *incidents.Application
}

func (c *IncidentCreateCommitFaultCapability) Create(
	ctx context.Context,
	actor authn.UserRecord,
	request incidents.IncidentCreateAdmission,
	requestID string,
) error {
	if c == nil || c.application == nil {
		return errors.New("incident create commit-fault capability is not installed")
	}
	_, err := c.application.CreateIncident(
		ctx,
		actor,
		request,
		requestID,
	)
	return err
}

type incidentCreateCommitFaultDB struct {
	postgres.DB
}

func (db incidentCreateCommitFaultDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return incidentCreateCommitFaultTx{Tx: tx}, nil
}

type incidentCreateCommitFaultTx struct {
	pgx.Tx
}

func (tx incidentCreateCommitFaultTx) Commit(ctx context.Context) error {
	_ = tx.Tx.Rollback(ctx)
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
	if isNilIncidentCreateCommitFaultDB(db) {
		return nil, errors.New("incident create commit fault requires Postgres")
	}
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            incidentCreateCommitFaultDB{DB: db},
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		return nil, fmt.Errorf("construct incident create commit-fault application: %w", err)
	}
	return &IncidentCreateCommitFaultCapability{application: application}, nil
}

func isNilIncidentCreateCommitFaultDB(db postgres.DB) bool {
	if db == nil {
		return true
	}
	value := reflect.ValueOf(db)
	return value.Kind() == reflect.Pointer && value.IsNil()
}
