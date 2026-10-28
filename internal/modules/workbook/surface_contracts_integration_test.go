package workbook_test

import (
	"context"
	"maps"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// Each case exercises the same Workbook create/query contract. Only assessment
// needs a source prerequisite; other domains do not seed each other's records.
func TestWorkbookSurfaceContracts_Integration(t *testing.T) {
	harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "workbook-surface-contracts", TestRouteMode: httptestx.TestRouteModeDisabled})
	admin, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	actor := authflowtest.SeedLocalUserRecord(t, harness.DB, "surface-contracts@example.test", "Surface actor", "SurfaceContracts1!", false, false, true)
	login := LoginLocalUserNoMFA(t, harness, actor.Email, "SurfaceContracts1!")
	for _, tc := range []struct {
		name         string
		body         map[string]any
		changedField string
		changedValue any
	}{
		{"hosts", entitytest.HostCreatePayload("create"), "host.display_name", "Changed host"},
		{"identities", entitytest.IdentityCreatePayload("create"), "identity.display_name", "Changed identity"},
		{"indicators", indicatortest.CreatePayload("create"), "indicator.display_value", "203.0.113.25"},
		{"timeline", map[string]any{"timeline.activity_synopsis_text": "Surface event"}, "timeline.activity_synopsis_text", "Changed event"},
		{"parties", map[string]any{"party.display_name": "Surface party", "party.party_kind": "organization"}, "party.display_name", "Changed party"},
		{"assessments", map[string]any{"assessment.subject_type": "host", "assessment.assessment_state": "confirmed", "assessment.confidence_score": 55, "assessment.rationale": "Surface assessment"}, "assessment.rationale", "Changed rationale"},
		{"evidence", map[string]any{"evidence.title": "Surface evidence"}, "evidence.title", "Changed evidence"},
		{"notes", map[string]any{"note.title": "Surface note"}, "note.title", "Changed note"},
		{"task_requests", map[string]any{"task.title": "Surface task", "task.task_kind": "collection"}, "task.title", "Changed task"},
		{"decisions", map[string]any{"decision.summary": "Surface decision", "decision.decision_type": "containment", "decision.rationale": "Surface rationale"}, "decision.summary", "Changed decision"},
		{"comm_log", map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Surface communication"}, "comm_log.summary", "Changed communication"},
		{"handoff", map[string]any{"handoff.incoming_owner_user_id": actor.ID.String(), "handoff.current_state_summary": "Surface handoff"}, "handoff.current_state_summary", "Changed handoff"},
		{"status_review", map[string]any{"status_review.current_state_summary": "Surface status"}, "status_review.current_state_summary", "Changed status"},
		{"lesson", map[string]any{"lesson.summary": "Surface lesson"}, "lesson.summary", "Changed lesson"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			incident := appsupport.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "incident-" + tc.name, "incident_key": "SURFACE-" + tc.name, "title": "Surface " + tc.name})
			incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
			incidentstoretest.SeedMembership(t, harness.DB, incidentID, actor.ID, actor.DisplayName, "admin", adminID)
			view := "cartulary.view." + tc.name + ".v1"
			if tc.name == "timeline" {
				view = "cartulary.view.timeline.v2"
			}
			body := maps.Clone(tc.body)
			body["client_txn_id"] = "create-" + tc.name
			if tc.name == "assessments" {
				hostID := uuid.New()
				entitytest.SeedHostRecord(t, harness.DB, incidentID, actor.ID, hostID, "Subject host", "SUBJECT-HOST", "", "")
				body["assessment.subject_ref"] = hostID.String()
			}
			path := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + view
			doCreate := func(payload any, csrf bool) *http.Response {
				opts := []func(*http.Request){httptestx.WithCookies(login.SessionCookie, login.CSRFCookie)}
				if csrf {
					opts = append(opts, httptestx.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
				}
				return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path+"/rows", payload, opts...)
			}
			httptestx.RequireErrorEnvelope(t, doCreate(body, false), http.StatusForbidden, "csrf_verification_failed")
			first := httptestx.RequireSuccessData(t, doCreate(body, true), http.StatusCreated)
			row := first["row"].(map[string]any)
			id := row["record_id"].(string)
			if id == "" || first["change_set_id"] == nil || first["change_set_id"] == "" {
				t.Fatalf("missing mutation identity: %#v", first)
			}
			source := "workbook.rows.create"
			switch tc.name {
			case "hosts", "identities":
				source = "entities." + tc.name + ".rows.create"
			case "indicators", "timeline", "assessments":
				source = tc.name + ".rows.create"
			}
			workbookscenariotest.RequireChangeSetAttribution(t, harness.DB, first["change_set_id"].(string), actor.ID.String(), source, body["client_txn_id"].(string))
			before := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
			sourceBefore := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), appsupport.MustUUID(t, id))
			query := func() map[string]any {
				if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), appsupport.MustUUID(t, id)); !reflect.DeepEqual(sourceBefore, got) {
					t.Fatal("source record changed without an authorized mutation")
				}
				envelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), view, login)
				contractassert.RequireDefaultQueryMeta(t, envelope, view)
				rows := envelope["data"].(map[string]any)["rows"].([]any)
				if len(rows) != 1 {
					t.Fatalf("expected only the case's row, got %#v", rows)
				}
				got := rows[0].(map[string]any)
				if got["record_id"] != id {
					t.Fatalf("unexpected query record: %#v", got)
				}
				contractassert.RequireFieldKeyConformance(t, workbookscenariotest.SortedRowFieldKeys(t, got), workbookscenariotest.AllowedFieldKeys(t, tc.name, view))
				return got
			}
			projected := query()
			replay := httptestx.RequireSuccessData(t, doCreate(body, true), http.StatusOK)
			if replay["change_set_id"] != first["change_set_id"] || !reflect.DeepEqual(replay["row"], row) {
				t.Fatalf("replay changed result: first=%#v replay=%#v", first, replay)
			}
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("replay changed history: %+v -> %+v", before, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			divergent := maps.Clone(body)
			divergent[tc.changedField] = tc.changedValue
			if tc.name == "indicators" {
				delete(divergent, "indicator.normalized_value")
				delete(divergent, "indicator.defanged_value")
			}
			httptestx.RequireErrorEnvelope(t, doCreate(divergent, true), http.StatusConflict, "client_txn_conflict")
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("divergent request changed history: %+v -> %+v", before, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			switch tc.name {
			case "hosts", "identities", "indicators", "timeline":
				if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
					t.Fatal(err)
				}
				contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
			}
			// The same authenticated session performed the allowed create above.
			if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer', membership_version=membership_version+1 WHERE incident_id=$1 AND user_id=$2`, incidentID, actor.ID); err != nil {
				t.Fatal(err)
			}
			denied := maps.Clone(body)
			denied["client_txn_id"] = "denied-" + tc.name
			httptestx.RequireErrorEnvelope(t, doCreate(denied, true), http.StatusForbidden, "authorization_denied")
			contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("denied write changed history: %+v -> %+v", before, got)
			}
			if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, actor.ID); err != nil {
				t.Fatal(err)
			}
			resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, path+"/query", map[string]any{}, httptestx.WithCookies(login.SessionCookie))
			httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "incident_not_found")
		})
	}
}

func TestWorkbookPatchSurfaceContracts_Integration(t *testing.T) {
	for _, name := range []string{"timeline", "parties"} {
		t.Run(name, func(t *testing.T) {
			harness, admin, adminID, incidentID := ConflictFixture(t, "workbook-patch-"+name, "PATCH-"+name)
			actor := authflowtest.SeedLocalUserRecord(t, harness.DB, "patch-"+name+"@example.test", "Patch actor", "SurfaceContracts1!", false, false, true)
			incidentstoretest.SeedMembership(t, harness.DB, incidentID, actor.ID, actor.DisplayName, "editor", adminID)
			login := LoginLocalUserNoMFA(t, harness, actor.Email, "SurfaceContracts1!")
			view := "cartulary.view." + name + ".v1"
			if name == "timeline" {
				view = "cartulary.view.timeline.v2"
			}
			create := map[string]any{"client_txn_id": "create", "party.display_name": "Patch party", "party.party_kind": "organization"}
			if name == "timeline" {
				create = map[string]any{"client_txn_id": "create", "timeline.activity_synopsis_text": "Patch event"}
			}
			data := requireWorkbookCreate(t, harness, login, incidentID, view, create)
			id := appsupport.MustUUID(t, data["row"].(map[string]any)["record_id"].(string))
			body := map[string]any{"view_schema_id": view, "base_row_version": 1, "client_txn_id": "patch", "changes": []map[string]any{{"field_key": "party.primary_email", "value": "support@example.test"}}}
			divergent := maps.Clone(body)
			divergent["changes"] = []map[string]any{{"field_key": "party.primary_email", "value": "changed@example.test"}}
			if name == "timeline" {
				host, other := uuid.New(), uuid.New()
				entitytest.SeedHostRecord(t, harness.DB, incidentID, actor.ID, host, "WS-023", "WS-023", "", "")
				entitytest.SeedHostRecord(t, harness.DB, incidentID, actor.ID, other, "WS-024", "WS-024", "", "")
				body = timelinetest.TimelineCollectionPatchPayload(timelinetest.FieldHostRefs, 1, "patch", timelinetest.CollectionActions(timelinetest.AddResolvedRefAction("WS-023", host)))
				divergent = timelinetest.TimelineCollectionPatchPayload(timelinetest.FieldHostRefs, 1, "patch", timelinetest.CollectionActions(timelinetest.AddResolvedRefAction("WS-024", other)))
			}
			url := harness.Server.HTTP.URL + "/api/v1/records/" + id.String()
			httptestx.RequireErrorEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, url, body, httptestx.WithCookies(login.SessionCookie, login.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
			first := httptestx.RequireSuccessData(t, doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", id, body), http.StatusOK)
			before := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
			sourceBefore := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), id)
			query := func() map[string]any {
				if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), id); !reflect.DeepEqual(sourceBefore, got) {
					t.Fatal("patch source changed during replay or rejection")
				}
				return workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), view, admin), id.String())
			}
			projected := query()
			contractassert.RequireFieldKeyConformance(t, workbookscenariotest.SortedRowFieldKeys(t, projected), workbookscenariotest.AllowedFieldKeys(t, name, view))
			replay := httptestx.RequireSuccessData(t, doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", id, body), http.StatusOK)
			if replay["change_set_id"] != first["change_set_id"] || !reflect.DeepEqual(first["row"], replay["row"]) {
				t.Fatalf("patch replay changed result: %#v %#v", first, replay)
			}
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("exact patch replay changed effects: %+v -> %+v", before, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			httptestx.RequireErrorEnvelope(t, doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", id, divergent), http.StatusConflict, "client_txn_conflict")
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("patch replay changed history: %+v -> %+v", before, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			if name == "timeline" {
				if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
					t.Fatal(err)
				}
				contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
			}
			if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer', membership_version=membership_version+1 WHERE incident_id=$1 AND user_id=$2`, incidentID, actor.ID); err != nil {
				t.Fatal(err)
			}
			denied := maps.Clone(body)
			denied["client_txn_id"] = "denied"
			denied["base_row_version"] = 2
			httptestx.RequireErrorEnvelope(t, doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", id, denied), http.StatusForbidden, "authorization_denied")
			contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("denied patch changed history: %+v -> %+v", before, got)
			}
		})
	}
}
