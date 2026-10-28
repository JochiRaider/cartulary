package appsupport

import (
	"context"
	"database/sql"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func CreateIncident(t testing.TB, server *httptestx.Server, admin authcookietest.AuthCookies, body map[string]any) map[string]any {
	t.Helper()

	resp := httptestx.DoJSON(t, http.DefaultClient,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		httptestx.WithCookies(admin.Session, admin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func CreateIncidentInStore(t testing.TB, pool postgres.DB, actor authn.UserRecord, clientTxnID string, incidentKey string, title string) incidents.IncidentRecord {
	t.Helper()

	store, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            pool,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	request := admissiontest.IncidentCreate(t, admissiontest.IncidentCreateInput{
		ClientTxnID: clientTxnID,
		IncidentKey: incidentKey,
		Title:       title,
	})
	result, err := store.CreateIncident(context.Background(), actor, request, "req-"+clientTxnID)
	if err != nil {
		t.Fatalf("create incident in store: %v", err)
	}
	return result.Incident
}

func MustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}

func QueryCount(t testing.TB, db any, query string, args ...any) int {
	t.Helper()

	var count int
	var err error
	switch typed := db.(type) {
	case *sql.DB:
		err = typed.QueryRowContext(context.Background(), query, args...).Scan(&count)
	case postgres.DB:
		err = typed.QueryRow(context.Background(), query, args...).Scan(&count)
	default:
		t.Fatalf("query count requires *sql.DB or postgres.DB, got %T", db)
	}
	if err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}
