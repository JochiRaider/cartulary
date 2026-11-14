package entities_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestHostRollbackRestoresPromotedFields_Integration(t *testing.T) {
	for _, kind := range []string{"change_set", "row_restore"} {
		for _, nullable := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/null=%t", kind, nullable), func(t *testing.T) {
				h := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "host-rollback-fields", TestRouteMode: httptestx.TestRouteModeDisabled})
				login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
				incident := appsupport.CreateIncident(t, h.Server, login, map[string]any{
					"client_txn_id": "host-rollback-incident", "incident_key": "IR-HOST-ROLLBACK", "title": "Host source restoration",
				})
				incidentID := incident["incident_id"].(string)
				fields := []string{"location", "os_platform", "business_owner", "criticality", "containment_status"}
				before := []any{"HQ", "Linux", "Security", "high", "contained"}
				if nullable {
					before = []any{nil, nil, nil, nil, nil}
				}
				payload := map[string]any{"client_txn_id": "host-rollback-create", "host.display_name": "Rollback Host", "host.hostname": "rollback-host"}
				for i, key := range fields {
					if before[i] != nil {
						payload["host."+key] = before[i]
					}
				}
				created := createEntityRow(t, h.Server.HTTP.URL, incidentID, viewtest.HostsViewSchemaID, login, payload, http.StatusCreated)
				recordID := created["row"].(map[string]any)["record_id"].(string)
				request := func(method, path string, body map[string]any) map[string]any {
					t.Helper()
					return httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, method, h.Server.HTTP.URL+path, body,
						httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)), http.StatusOK)
				}
				assertState := func(version int, expected []any) {
					t.Helper()
					var values [5]sql.NullString
					var sourceVersion, recordVersion int
					var updatedBy string
					if err := h.DB.QueryRowContext(context.Background(), `
SELECT h.location, h.os_platform, h.business_owner, h.criticality, h.containment_status,
       h.row_version, r.row_version, h.updated_by_user_id::text
  FROM hosts h JOIN records r USING (record_id) WHERE h.record_id = $1
`, recordID).Scan(&values[0], &values[1], &values[2], &values[3], &values[4], &sourceVersion, &recordVersion, &updatedBy); err != nil {
						t.Fatalf("read Host source: %v", err)
					}
					if sourceVersion != version || recordVersion != version || updatedBy != actorID.String() {
						t.Fatalf("source/envelope attribution mismatch: source=%d record=%d actor=%s", sourceVersion, recordVersion, updatedBy)
					}
					row := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, h.Server.HTTP.URL, incidentID, viewtest.HostsViewSchemaID, login), recordID)
					if row["row_version"] != float64(version) {
						t.Fatalf("public version = %#v, want %d", row["row_version"], version)
					}
					cells := row["cells"].(map[string]any)
					for i, key := range fields {
						var got any
						if values[i].Valid {
							got = values[i].String
						}
						if got != expected[i] || cells["host."+key].(map[string]any)["value"] != expected[i] {
							t.Fatalf("%s source=%#v public=%#v want=%#v", key, got, cells["host."+key], expected[i])
						}
					}
					if cells["host.hostname"].(map[string]any)["value"] != "rollback-host" || cells["host.display_name"].(map[string]any)["value"] != "Rollback Host" {
						t.Fatalf("untouched identity fields changed: %#v", row)
					}
				}
				assertState(1, before)
				after := []any{"Branch", "Windows", "Operations", "low", "uncontained"}
				changes := make([]map[string]any, 0, len(fields))
				for i, key := range fields {
					changes = append(changes, map[string]any{"field_key": "host." + key, "value": after[i]})
				}
				patched := request(http.MethodPatch, "/api/v1/records/"+recordID, map[string]any{
					"view_schema_id": viewtest.HostsViewSchemaID, "base_row_version": 1, "client_txn_id": "host-rollback-patch", "changes": changes,
				})
				assertState(2, after)
				target := map[string]any{"kind": kind}
				if kind == "change_set" {
					target["change_set_id"] = patched["change_set_id"]
				} else {
					target["restore_to_revision_no"] = 1
				}
				body := map[string]any{"base_row_version": 2, "client_txn_id": "host-rollback", "target": target}
				rolledBack := request(http.MethodPost, "/api/v1/records/"+recordID+"/rollback", body)
				if rolledBack["record_id"] != recordID || rolledBack["incident_id"] != incidentID || rolledBack["row_version"] != float64(3) || !reflect.DeepEqual(rolledBack["affected_record_ids"], []any{recordID}) {
					t.Fatalf("unexpected rollback summary: %#v", rolledBack)
				}
				assertState(3, before)
				rollbackID, ok := rolledBack["rollback_change_set_id"].(string)
				if !ok || rollbackID == "" {
					t.Fatalf("missing rollback change set: %#v", rolledBack)
				}
				if n := appsupport.QueryCount(t, h.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID); n != 3 {
					t.Fatalf("revision count = %d, want 3", n)
				}
				operation := "rollback"
				if kind == "row_restore" {
					operation = "row_restore"
				}
				if n := appsupport.QueryCount(t, h.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'host' AND target_id = $2 AND operation_kind = $3`, rollbackID, recordID, operation); n != 1 {
					t.Fatalf("inverse Host mutation count = %d, want 1", n)
				}
				revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, h.DB, rollbackID)
				snapshot := workbookscenariotest.SnapshotIncidentMutationEffects(t, h.DB, incidentID)
				replayed := request(http.MethodPost, "/api/v1/records/"+recordID+"/rollback", body)
				if !reflect.DeepEqual(replayed, rolledBack) || workbookscenariotest.SnapshotIncidentMutationEffects(t, h.DB, incidentID) != snapshot {
					t.Fatalf("rollback replay changed result or effects: first=%#v replay=%#v", rolledBack, replayed)
				}
				assertState(3, before)
			})
		}
	}
}
