package entities_test

import (
	"context"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"testing"

	"github.com/google/uuid"

	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	incidentstoretest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	collabtestprotocol "github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/protocoltest"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestEntityRelationshipContracts_Integration(t *testing.T) {
	for _, kind := range []string{"resolve", "merge"} {
		t.Run(kind, func(t *testing.T) {
			harness := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "entity-relationship-" + kind, TestRouteMode: httptestx.TestRouteModeDisabled})
			admin, adminID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
			incident := appsupport.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "incident", "incident_key": "RELATIONSHIP-" + kind, "title": "Relationship " + kind})
			incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
			actor := authflowtest.SeedLocalUserRecord(t, harness.DB, "relationship@example.test", "Relationship actor", "RelationshipContracts1!", false, false, true)
			incidentstoretest.SeedMembership(t, harness.DB, incidentID, actor.ID, actor.DisplayName, "admin", adminID)
			session, csrf := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, actor.Email, "RelationshipContracts1!", nil)
			login := appsupport.LoginResult{SessionCookie: session, CSRFCookie: csrf}
			event, mention, host := uuid.New(), uuid.New(), uuid.New()
			timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actor.ID, event)
			entitytest.SeedHostRecord(t, harness.DB, incidentID, actor.ID, host, "WS-023", "WS-023", "", "")
			path := "/api/v1/entity-mentions/" + mention.String() + "/resolve"
			body := entitytest.MentionResolveRoutePayload(1, "mutation", entitytest.MentionActionResolve, &host, nil)
			divergent := entitytest.MentionResolveRoutePayload(1, "mutation", entitytest.MentionActionDismiss, nil, nil)
			loser := uuid.Nil
			expectedVersion := int64(2)
			if kind == "resolve" {
				entitytest.SeedMention(t, harness.DB, actor.ID, mention, event, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)
			} else {
				loser = uuid.New()
				expectedVersion = 1
				entitytest.SeedHostRecord(t, harness.DB, incidentID, actor.ID, loser, "WS-024", "WS-024", "ws-024.corp.example.test", "")
				entitytest.SeedResolvedMention(t, harness.DB, actor.ID, mention, event, loser, timelinetest.FieldHostRefs, "host", "WS-024")
				linktest.SeedRecordLink(t, harness.DB, incidentID, actor.ID, uuid.New(), event, loser, "observed_on_host", "manual", nil)
				linktest.SeedRecordTag(t, harness.DB, incidentID, actor.ID, uuid.New(), host, "critical-host")
				linktest.SeedRecordTag(t, harness.DB, incidentID, actor.ID, uuid.New(), loser, "critical-host")
				assessmenttest.SeedAssessment(t, harness.DB, incidentID, actor.ID, uuid.New(), loser, "host", "confirmed")
				path = "/api/v1/records/" + host.String() + "/merge"
				body = map[string]any{"loser_record_id": loser.String(), "survivor_base_row_version": 1, "loser_base_row_version": 1, "client_txn_id": "mutation", "reason": "Merge duplicate host"}
				divergent = maps.Clone(body)
				divergent["reason"] = "Divergent merge"
			}
			// SQL owner fixtures need one initial projection synchronization.
			if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
				t.Fatal(err)
			}
			do := func(payload any, includeCSRF bool) *http.Response {
				opts := []func(*http.Request){appsupport.WithCookies(session, csrf)}
				if includeCSRF {
					opts = append(opts, appsupport.WithHeader(authn.CSRFHeaderName, csrf.Value))
				}
				return appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+path, payload, opts...)
			}
			appsupport.RequireErrorBody(t, do(body, false), http.StatusForbidden, "csrf_verification_failed")
			socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.TimelineViewSchemaID, session.Value)
			defer socket.Close(1000, "test_complete")
			first := appsupport.RequireSuccessData(t, do(body, true), http.StatusOK)
			changeID, ok := first["change_set_id"].(string)
			if !ok || changeID == "" {
				t.Fatalf("missing change identity: %#v", first)
			}
			if kind == "resolve" {
				if first["incident_id"] != incidentID.String() {
					t.Fatalf("wrong incident: %#v", first)
				}
				if _, ok := first["source_record"].(map[string]any); !ok {
					t.Fatalf("missing source record: %#v", first)
				}
			} else {
				if first["survivor_record_id"] != host.String() || first["loser_record_id"] != loser.String() {
					t.Fatalf("wrong merge identities: %#v", first)
				}
				if _, ok := first["merge_summary"].(map[string]any); !ok {
					t.Fatalf("missing merge summary: %#v", first)
				}
			}
			want := map[uuid.UUID]int64{event: expectedVersion, host: 1}
			if kind == "merge" {
				want[host] = 2
				want[loser] = 2
			}
			changes := incidentwstest.RequireRecordChangedEvents(t, socket, want)
			requireRelationshipSocket(t, changes[event], changeID, viewtest.TimelineViewSchemaID, nil)
			for id := range want {
				if id != event {
					requireRelationshipSocket(t, changes[id], changeID, viewtest.HostsViewSchemaID, nil)
				}
			}
			if kind == "resolve" {
				requireRelationshipSocket(t, changes[host], changeID, viewtest.HostsViewSchemaID, []string{"host.linked_event_count"})
			}

			incidentwstest.ExpectNoSocketMessage(t, socket)
			sourceIDs := []uuid.UUID{event, host}
			if loser != uuid.Nil {
				sourceIDs = append(sourceIDs, loser)
			}
			sourceBefore := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), sourceIDs...)
			query := func() map[string]any {
				if got := revisionsupport.SnapshotRecords(t, harness.Pool, harness.Revisions.Appender(), sourceIDs...); !reflect.DeepEqual(sourceBefore, got) {
					t.Fatal("relationship source changed during replay or rejection")
				}
				envelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.TimelineViewSchemaID, login)
				contractassert.RequireDefaultQueryMeta(t, envelope, viewtest.TimelineViewSchemaID)
				rows := envelope["data"].(map[string]any)["rows"].([]any)
				if len(rows) != 1 {
					t.Fatalf("unexpected timeline rows: %#v", rows)
				}
				row := rows[0].(map[string]any)
				if row["record_id"] != event.String() {
					t.Fatalf("wrong source row: %#v", row)
				}
				contractassert.RequireFieldKeyConformance(t, workbookscenariotest.SortedRowFieldKeys(t, row), workbookscenariotest.AllowedFieldKeys(t, kind, viewtest.TimelineViewSchemaID))
				return row
			}
			projected := query()
			before := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String())
			mentionBefore := entitytest.LookupMention(t, harness.DB, mention)
			if mentionBefore.ResolutionStatus != "resolved" || mentionBefore.ResolvedRecordID == nil || *mentionBefore.ResolvedRecordID != host {
				t.Fatalf("wrong resolved target: %#v", mentionBefore)
			}
			replay := appsupport.RequireSuccessData(t, do(body, true), http.StatusOK)
			if replay["change_set_id"] != changeID {
				t.Fatalf("changed replay identity: %#v", replay)
			}
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("exact replay changed mutation effects: %+v -> %+v", before, got)
			}
			for _, key := range []string{"source_record", "survivor_record_id", "loser_record_id", "merge_summary"} {
				if !reflect.DeepEqual(first[key], replay[key]) {
					t.Fatalf("replay changed %s", key)
				}
			}
			if got := entitytest.LookupMention(t, harness.DB, mention); !reflect.DeepEqual(got, mentionBefore) {
				t.Fatal("exact replay changed mention source")
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			incidentwstest.ExpectNoSocketMessage(t, socket)
			appsupport.RequireErrorBody(t, do(divergent, true), http.StatusConflict, "client_txn_conflict")
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("replay changed history: %+v -> %+v", before, got)
			}
			if got := entitytest.LookupMention(t, harness.DB, mention); !reflect.DeepEqual(got, mentionBefore) {
				t.Fatalf("replay changed mention: %#v -> %#v", mentionBefore, got)
			}
			contractassert.RequireProjectionDeterminism(t, projected, query())
			incidentwstest.ExpectNoSocketMessage(t, socket)
			if err := harness.Projections.RebuildIncident(context.Background(), incidentID); err != nil {
				t.Fatal(err)
			}
			contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
			if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role='viewer', membership_version=membership_version+1 WHERE incident_id=$1 AND user_id=$2`, incidentID, actor.ID); err != nil {
				t.Fatal(err)
			}
			denied := maps.Clone(body)
			denied["client_txn_id"] = "denied"
			appsupport.RequireErrorBody(t, do(denied, true), http.StatusForbidden, "authorization_denied")
			if got := workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()); got != before {
				t.Fatalf("denied mutation changed history: %+v -> %+v", before, got)
			}
			if got := entitytest.LookupMention(t, harness.DB, mention); !reflect.DeepEqual(got, mentionBefore) {
				t.Fatalf("denial changed mention: %#v", got)
			}
			contractassert.RequireProjectionDeterminism(t, projected["cells"], query()["cells"])
		})
	}
}

func requireRelationshipSocket(t testing.TB, change collabtestprotocol.RecordChangedEvent, changeID, view string, keys []string) {
	t.Helper()
	if change.ChangeSetID.String() != changeID {
		t.Fatalf("socket change identity: %#v", change)
	}
	found := false
	for _, affected := range change.AffectedViews {
		if affected.ViewSchemaID == view {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing affected view %s: %#v", view, change)
	}
	for _, key := range keys {
		if !slices.Contains(change.ChangedFieldKeys, key) {
			t.Fatalf("missing changed key %s: %#v", key, change)
		}
	}
}
