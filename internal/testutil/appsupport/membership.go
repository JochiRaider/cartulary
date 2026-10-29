package appsupport

import (
	"context"
	"database/sql"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	workbookstoretest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"testing"
	"time"
)

// SeedMembership composes owner fixtures for a usable application member.
func SeedMembership(t testing.TB, db postgres.DB, incidentID, userID uuid.UUID, role string, addedByUserID uuid.UUID) {
	t.Helper()
	incidentstoretest.SeedMembership(t, db, incidentID, userID, role, addedByUserID)
	workbookstoretest.SeedUserPreferences(t, db, incidentID, userID)
}

// SeedMembershipSQL composes owner fixtures for a usable application member.
func SeedMembershipSQL(t testing.TB, db *sql.DB, incidentID, userID uuid.UUID, role string, addedByUserID uuid.UUID) {
	t.Helper()
	incidentstoretest.SeedMembershipSQL(t, db, incidentID, userID, role, addedByUserID)
	workbookstoretest.SeedUserPreferencesSQL(t, db, incidentID, userID)
}

func CreateMembershipInStore(
	t testing.TB,
	pool postgres.DB,
	actor authn.UserRecord,
	incidentID uuid.UUID,
	targetUser authn.UserRecord,
	clientTxnID string,
	role string,
) incidents.MembershipCreateResult {
	t.Helper()
	request := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{
		ClientTxnID: clientTxnID,
		UserID:      &targetUser.ID,
		Role:        role,
	})

	store, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            pool,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	result, err := store.CreateMembership(
		context.Background(),
		actor,
		incidentID,
		request,
		"req-"+clientTxnID,
	)
	if err != nil {
		t.Fatalf("create membership in store: %v", err)
	}
	return result
}
