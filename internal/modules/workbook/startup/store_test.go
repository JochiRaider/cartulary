package startup_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	workbookstartup "github.com/JochiRaider/cartulary/internal/modules/workbook/startup"
	phase2storetest "github.com/JochiRaider/cartulary/internal/testutil/phase2storetest"
)

func TestSupportPhase2_WorkbookStartupPreferencesBootstrapAndUpsert(t *testing.T) {
	harness := phase2storetest.StartStore(t, "workbook-startup-prefs")
	store := workbookstartup.NewStore(harness.DB)
	actor := phase2storetest.SeedLocalUserRecord(
		t,
		harness.DB,
		"workbook-startup-prefs@example.test",
		"Workbook Startup Prefs",
		"WorkbookStartupPrefs1!",
		false,
		false,
		true,
	)

	result := phase2storetest.CreateIncidentInStore(t, harness.DB, actor, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-workbook-startup-prefs-create",
		IncidentKey: "IR-WORKBOOK-PREFS",
		Title:       "Workbook startup preferences",
	})

	defaultPrefs, err := store.GetDefaultPreferences(context.Background(), result.Incident.ID)
	if err != nil {
		t.Fatalf("lookup incident workbook preferences: %v", err)
	}
	if defaultPrefs.IncidentID != result.Incident.ID || defaultPrefs.UpdatedByUserID == nil || *defaultPrefs.UpdatedByUserID != actor.ID {
		t.Fatalf("unexpected incident workbook preferences: %#v", defaultPrefs)
	}

	userPrefs, err := store.GetUserPreferences(context.Background(), result.Incident.ID, actor.ID)
	if err != nil {
		t.Fatalf("lookup user workbook preferences: %v", err)
	}
	if userPrefs.IncidentID != result.Incident.ID || userPrefs.UserID != actor.ID {
		t.Fatalf("unexpected user workbook preferences: %#v", userPrefs)
	}

	secondActor := phase2storetest.SeedLocalUserRecord(
		t,
		harness.DB,
		"workbook-startup-prefs-second@example.test",
		"Workbook Startup Prefs Second",
		"WorkbookStartupPrefsSecond1!",
		false,
		false,
		true,
	)

	ctx := context.Background()
	timelineRef := []byte(`{"kind":"view_schema","id":"cartulary.view.timeline.v2"}`)
	timelineRefReordered := []byte(`{"id":"cartulary.view.timeline.v2","kind":"view_schema"}`)
	hostsRef := []byte(`{"kind":"view_schema","id":"cartulary.view.hosts.v1"}`)
	firstTime := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	noOpTime := firstTime.Add(time.Hour)
	changeTime := firstTime.Add(2 * time.Hour)

	if _, err := harness.DB.Exec(ctx, `
DELETE FROM user_workbook_preferences
 WHERE incident_id = $1
   AND user_id = $2
`, result.Incident.ID, actor.ID); err != nil {
		t.Fatalf("delete bootstrap user workbook preferences: %v", err)
	}
	userCreated, err := store.PutUserPreferences(ctx, result.Incident.ID, actor.ID, timelineRef, firstTime)
	if err != nil {
		t.Fatalf("create user workbook preferences: %v", err)
	}
	requireWorkbookSheetRefID(t, userCreated.HomeSheetRef, "cartulary.view.timeline.v2")
	if !userCreated.CreatedAt.Equal(firstTime) || !userCreated.UpdatedAt.Equal(firstTime) {
		t.Fatalf("unexpected created user preference timestamps: %#v", userCreated)
	}

	userNoOp, err := store.PutUserPreferences(ctx, result.Incident.ID, actor.ID, timelineRefReordered, noOpTime)
	if err != nil {
		t.Fatalf("no-op user workbook preferences: %v", err)
	}
	if !userNoOp.UpdatedAt.Equal(firstTime) {
		t.Fatalf("structural no-op must preserve user updated_at: before=%s after=%s", firstTime, userNoOp.UpdatedAt)
	}

	userChanged, err := store.PutUserPreferences(ctx, result.Incident.ID, actor.ID, hostsRef, changeTime)
	if err != nil {
		t.Fatalf("update user workbook preferences: %v", err)
	}
	requireWorkbookSheetRefID(t, userChanged.HomeSheetRef, "cartulary.view.hosts.v1")
	if !userChanged.UpdatedAt.Equal(changeTime) {
		t.Fatalf("effective user update must advance updated_at once: %#v", userChanged)
	}

	if _, err := harness.DB.Exec(ctx, `
DELETE FROM incident_workbook_preferences
 WHERE incident_id = $1
`, result.Incident.ID); err != nil {
		t.Fatalf("delete bootstrap default workbook preferences: %v", err)
	}
	defaultCreated, err := store.PutDefaultPreferences(ctx, result.Incident.ID, actor.ID, timelineRef, firstTime)
	if err != nil {
		t.Fatalf("create default workbook preferences: %v", err)
	}
	requireWorkbookSheetRefID(t, defaultCreated.DefaultSheetRef, "cartulary.view.timeline.v2")
	if defaultCreated.UpdatedByUserID == nil || *defaultCreated.UpdatedByUserID != actor.ID {
		t.Fatalf("default preference create must attribute admin actor: %#v", defaultCreated)
	}
	if !defaultCreated.CreatedAt.Equal(firstTime) || !defaultCreated.UpdatedAt.Equal(firstTime) {
		t.Fatalf("unexpected created default preference timestamps: %#v", defaultCreated)
	}

	defaultNoOp, err := store.PutDefaultPreferences(ctx, result.Incident.ID, secondActor.ID, timelineRefReordered, noOpTime)
	if err != nil {
		t.Fatalf("no-op default workbook preferences: %v", err)
	}
	if !defaultNoOp.UpdatedAt.Equal(firstTime) || defaultNoOp.UpdatedByUserID == nil || *defaultNoOp.UpdatedByUserID != actor.ID {
		t.Fatalf("structural no-op must preserve default updated_at and updated_by_user_id: %#v", defaultNoOp)
	}

	defaultChanged, err := store.PutDefaultPreferences(ctx, result.Incident.ID, secondActor.ID, hostsRef, changeTime)
	if err != nil {
		t.Fatalf("update default workbook preferences: %v", err)
	}
	requireWorkbookSheetRefID(t, defaultChanged.DefaultSheetRef, "cartulary.view.hosts.v1")
	if !defaultChanged.UpdatedAt.Equal(changeTime) || defaultChanged.UpdatedByUserID == nil || *defaultChanged.UpdatedByUserID != secondActor.ID {
		t.Fatalf("effective default update must advance timestamp and actor once: %#v", defaultChanged)
	}
}

func requireWorkbookSheetRefID(t testing.TB, raw []byte, wantID string) {
	t.Helper()
	var ref map[string]string
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("decode workbook sheet_ref: %v", err)
	}
	if ref["kind"] != "view_schema" || ref["id"] != wantID {
		t.Fatalf("unexpected workbook sheet_ref: got %#v want id %q", ref, wantID)
	}
}
