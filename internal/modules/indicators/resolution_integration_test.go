package indicators_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

// indicator-resolution / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestIndicatorsRoute_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "entity_linking-i-4-07-indicators")
	adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-entity_linking-i-4-07-incident",
		"incident_key":  "IR-I407",
		"title":         "Record relationships indicator route",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	payload := map[string]any{
		"client_txn_id":              "txn-entity_linking-i-4-07-create",
		"indicator.indicator_type":   indicatortest.Examples[0].IndicatorType,
		"indicator.value_kind":       indicatortest.Examples[0].ValueKind,
		"indicator.display_value":    indicatortest.Examples[0].DisplayValue,
		"indicator.normalized_value": indicatortest.Examples[0].NormalizedValue,
	}
	response := appsupport.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewtest.IndicatorsViewSchemaID+"/rows",
		payload,
		appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	data := appsupport.RequireSuccessData(t, response, http.StatusCreated)
	row := data["row"].(map[string]any)
	if row["record_id"] == "" || row["row_version"] == nil {
		t.Fatalf("indicator create row is incomplete: %#v", row)
	}
	changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), data["change_set_id"].(string))
	if changeSet.ActorUserID != adminUserID.String() || changeSet.Source != "indicators.rows.create" {
		t.Fatalf("indicator mutation attribution mismatch: %#v", changeSet)
	}
	login := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
	rows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login)
	queried := workbookscenariotest.FindRow(t, rows, row["record_id"].(string))
	if queried["record_id"] != row["record_id"] {
		t.Fatalf("indicator route readback mismatch: %#v", queried)
	}
	recordID := appsupport.MustUUID(t, row["record_id"].(string))
	store := indicators.NewStore(harness.Server.Runtime.Postgres, harness.Server.Runtime.Revisions.Appender())
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.SiblingRecordID)
	for index, sourceRecordID := range []struct {
		id    string
		field string
	}{
		{id: timelinetest.RecordID.String(), field: timelinetest.FieldSourceText},
		{id: timelinetest.SiblingRecordID.String(), field: timelinetest.FieldSummary},
	} {
		sourceID := appsupport.MustUUID(t, sourceRecordID.id)
		if _, _, err := store.CreateIndicatorObservation(context.Background(), authn.UserRecord{ID: adminUserID}, indicators.IndicatorObservationCreateParams{
			IncidentID:                incidentID,
			SourceRecordID:            sourceID,
			SourceFieldKey:            sourceRecordID.field,
			OriginKind:                "manual_entry",
			OriginLocator:             "entity_linking-i-4-07-observation-" + string(rune('1'+index)),
			ObservedText:              indicatortest.Examples[0].DefangedValue,
			ResolvedIndicatorRecordID: &recordID,
			CreatedAt:                 indicatortest.PastTime.Add(time.Duration(index) * time.Minute),
		}); err != nil {
			t.Fatalf("create observation %d: %v", index, err)
		}
	}
	if _, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), authn.UserRecord{ID: adminUserID}, indicators.IndicatorLifecycleAppendParams{
		IncidentID:        incidentID,
		IndicatorRecordID: recordID,
		LifecycleState:    "active",
		ValidFrom:         indicatortest.PastTime,
		CreatedAt:         indicatortest.PastTime,
	}); err != nil {
		t.Fatalf("append lifecycle: %v", err)
	}
	rowsBeforeRebuild := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login)
	rowBeforeRebuild := workbookscenariotest.FindRow(t, rowsBeforeRebuild, recordID.String())
	cells := rowBeforeRebuild["cells"].(map[string]any)
	if cells["indicator.observation_count"].(map[string]any)["value"] != float64(2) || cells["indicator.lifecycle_summary"].(map[string]any)["value"] != "active" {
		t.Fatalf("indicator projection readback mismatch: %#v", cells)
	}
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		t.Fatalf("clear indicator projections: %v", err)
	}
	if err := harness.Server.Runtime.Timeline.ProjectionCatalog.Rebuild.RebuildIndicators(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild indicator projections: %v", err)
	}
	rowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.IndicatorsViewSchemaID, login), recordID.String())
	if !reflect.DeepEqual(rowBeforeRebuild["cells"], rowAfterRebuild["cells"]) {
		t.Fatalf("indicator projection rebuild drifted: before=%#v after=%#v", rowBeforeRebuild, rowAfterRebuild)
	}
}
