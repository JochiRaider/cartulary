package entities_test

import (
	"context"
	"net/http"
	"testing"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// entity-resolution / REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044 / AC-188..AC-190, AC-221..AC-225.
func TestResolveRoute_Integration(t *testing.T) {
	t.Run("resolve_item uses authoritative route replay and history conformance", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-01-resolve-conformance")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-resolve-conf-incident",
			"incident_key":  "IR-I401-RC",
			"title":         "Record relationships entity-resolution resolve conformance",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)

		ctx := workbookscenariotest.RouteInventoryContext{
			IncidentID:       incidentID.String(),
			ActorUserID:      adminUserID.String(),
			TimelineRecordID: timelinetest.RecordID.String(),
			MentionID:        entitytest.HostMentionID.String(),
			HostRecordID:     entitytest.CanonicalHostRecordID.String(),
		}
		route := workbookscenariotest.MustRoute(t, workbookscenariotest.RouteMentionResolve, ctx)
		workbookscenariotest.RequireRouteReplayHistoryConformance(t, harness.DB, harness.Server.HTTP.URL, workbookscenariotest.RouteConformanceCase{
			Route:                  route,
			Context:                ctx,
			ClientTxnID:            "txn-entity_linking-i-4-01-resolve-conformance",
			Login:                  adminLogin,
			ActorUserID:            adminUserID.String(),
			ExpectedMutationSource: "entities.entity_mentions.resolve",
		})

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected route conformance resolve to leave mention resolved to target, got %#v", mention)
		}
	})

	t.Run("resolve_item persists durable state, replays idempotently, and emits websocket invalidation", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-01-resolve")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-resolve-incident",
			"incident_key":  "IR-I401-R",
			"title":         "Record relationships entity-resolution resolve route",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)

		socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		payload := entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-resolve", entitytest.MentionActionResolve, uuidPointer(entitytest.CanonicalHostRecordID), nil)
		resp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			payload,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := appsupport.RequireSuccessData(t, resp, http.StatusOK)
		if data["incident_id"] != incidentID.String() {
			t.Fatalf("unexpected incident_id in mention action payload: %#v", data)
		}
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after resolve_item, got %#v", data)
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, data["change_set_id"].(string))

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != entitytest.CanonicalHostRecordID {
			t.Fatalf("expected resolve_item to point mention at the target record, got %#v", mention)
		}
		if mention.ResolutionMethod == nil || *mention.ResolutionMethod != "explicit_resolve_route" {
			t.Fatalf("expected resolve route provenance marker, got %#v", mention)
		}
		link := linktest.LookupActiveLink(t, harness.DB, incidentID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host")
		linktest.RequireActiveLink(t, link, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)

		socketPayload := incidentwstest.RequireRecordChanged(t, socket, timelinetest.RecordID.String(), 2)
		if socketPayload.ChangeSetID != data["change_set_id"] {
			t.Fatalf("expected websocket to carry the mention action change_set_id, got payload=%#v response=%#v", socketPayload, data)
		}
		hostSocketPayload := incidentwstest.RequireRecordChanged(t, hostSocket, entitytest.CanonicalHostRecordID.String(), 1)
		if hostSocketPayload.ChangeSetID != data["change_set_id"] {
			t.Fatalf("expected host websocket invalidation to carry the mention action change_set_id, got payload=%#v response=%#v", hostSocketPayload, data)
		}
		if !stringSliceContains(hostSocketPayload.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected host linked-event count invalidation, got %#v", hostSocketPayload)
		}
		viewLogin := appsupport.LoginResult{
			SessionCookie: adminLogin.SessionCookie,
			CSRFCookie:    adminLogin.CSRFCookie,
		}
		hostRow := workbookscenariotest.FindRow(
			t,
			workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin),
			entitytest.CanonicalHostRecordID.String(),
		)
		hostCells := hostRow["cells"].(map[string]any)
		if got := hostCells["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected resolved host to expose one linked event, got %#v row=%#v", got, hostRow)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE host_grid_projection SET linked_event_count = 0 WHERE record_id = $1`, entitytest.CanonicalHostRecordID); err != nil {
			t.Fatalf("corrupt resolved host linked-event count: %v", err)
		}
		if err := harness.Projections.RebuildHosts(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild resolved host projection: %v", err)
		}
		rebuiltHostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), entitytest.CanonicalHostRecordID.String())
		if got := rebuiltHostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected rebuild to restore one linked event, got %#v row=%#v", got, rebuiltHostRow)
		}

		replayResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			payload,
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		replayData := appsupport.RequireSuccessData(t, replayResp, http.StatusOK)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("expected replay to reuse the original payload, got %#v %#v", data, replayData)
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, data["change_set_id"].(string))
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, "entities.entity_mentions.resolve", adminUserID.String(), entitytest.HostMentionID.String(), "txn-entity_linking-i-4-01-resolve"); got != 1 {
			t.Fatalf("expected one route idempotency row for a replayed mention action, got %d", got)
		}

		divergentResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-resolve", "dismiss_item", nil, nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("dismiss_item removes active links and fails closed on stale or illegal transitions", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-01-dismiss")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-dismiss-incident",
			"incident_key":  "IR-I401-D",
			"title":         "Record relationships entity-resolution dismiss route",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedResolvedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, timelinetest.FieldHostRefs, "host", "WS-023")
		linktest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, linktest.ManualLinkID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)

		socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		resp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-dismiss", "dismiss_item", nil, nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := appsupport.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["entity_mention"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected mention row_version=2 after dismiss_item, got %#v", data)
		}

		dismissed := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, dismissed, entitytest.MentionStatusDismissed)
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incidentID, timelinetest.RecordID, entitytest.CanonicalHostRecordID); got != 0 {
			t.Fatalf("expected dismiss_item to remove the active link, got %d active rows", got)
		}
		incidentwstest.RequireRecordChanged(t, socket, timelinetest.RecordID.String(), 2)
		hostChange := incidentwstest.RequireRecordChanged(t, hostSocket, entitytest.CanonicalHostRecordID.String(), 1)
		if !stringSliceContains(hostChange.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected dismiss to invalidate host linked-event count, got %#v", hostChange)
		}
		viewLogin := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), entitytest.CanonicalHostRecordID.String())
		if got := hostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("expected dismiss to project zero linked events, got %#v row=%#v", got, hostRow)
		}

		staleResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-stale", "revert_to_unresolved", nil, nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		staleBody := appsupport.RequireErrorBody(t, staleResp, http.StatusConflict, "row_version_conflict")
		if staleBody["error"].(map[string]any)["details"].(map[string]any)["current_mention_row_version"] != float64(2) {
			t.Fatalf("expected current_mention_row_version=2 on stale update, got %#v", staleBody)
		}

		illegalResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(2, "txn-entity_linking-i-4-01-illegal", "dismiss_item", nil, nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		illegalBody := appsupport.RequireErrorBody(t, illegalResp, http.StatusConflict, "illegal_transition")
		if illegalBody["error"].(map[string]any)["details"].(map[string]any)["from_status"] != "dismissed" {
			t.Fatalf("expected illegal transition details to report the dismissed source status, got %#v", illegalBody)
		}
		hostRowAfterFailures := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), entitytest.CanonicalHostRecordID.String())
		if got := hostRowAfterFailures["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("failed mention actions changed the projected linked-event count: %#v", hostRowAfterFailures)
		}
	})

	t.Run("revert_to_unresolved restores unresolved mention state and emits websocket invalidation", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-01-revert")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-revert-incident",
			"incident_key":  "IR-I401-U",
			"title":         "Record relationships entity-resolution revert route",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedResolvedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, timelinetest.FieldHostRefs, "host", "WS-023")
		linktest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, linktest.ManualLinkID, timelinetest.RecordID, entitytest.CanonicalHostRecordID, "observed_on_host", "manual", nil)

		socket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := incidentwstest.ConnectViewSocket(t, harness.Server, incidentID.String(), viewtest.HostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		resp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-revert", "revert_to_unresolved", nil, nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := appsupport.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after revert_to_unresolved, got %#v", data)
		}

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
		if mention.RawText != "WS-023" || mention.ResolvedRecordID != nil || mention.ResolutionMethod != nil {
			t.Fatalf("expected revert_to_unresolved to preserve raw_text and clear resolution metadata, got %#v", mention)
		}
		incidentwstest.RequireRecordChanged(t, socket, timelinetest.RecordID.String(), 2)
		hostChange := incidentwstest.RequireRecordChanged(t, hostSocket, entitytest.CanonicalHostRecordID.String(), 1)
		if !stringSliceContains(hostChange.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected revert to invalidate host linked-event count, got %#v", hostChange)
		}
		viewLogin := appsupport.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), viewtest.HostsViewSchemaID, viewLogin), entitytest.CanonicalHostRecordID.String())
		if got := hostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("expected revert to project zero linked events, got %#v row=%#v", got, hostRow)
		}
	})

	t.Run("authorization and target validation are re-derived from live state", func(t *testing.T) {
		harness := appsupport.StartServer(t, "entity_linking-i-4-01-access")
		adminLogin, adminUserID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
		incident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-access-incident",
			"incident_key":  "IR-I401-A",
			"title":         "Record relationships entity-resolution access checks",
		})
		incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
		timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, timelinetest.RecordID)
		entitytest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		entitytest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, entitytest.CanonicalIdentityRecordID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		entitytest.SeedMention(t, harness.DB, adminUserID, entitytest.HostMentionID, timelinetest.RecordID, timelinetest.FieldHostRefs, "host", "WS-023", "unresolved", nil, nil)

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote incident membership: %v", err)
		}
		deniedResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-denied", "resolve_item", uuidPointer(entitytest.CanonicalHostRecordID), nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'admin',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("restore incident membership: %v", err)
		}

		otherActor := authflowtest.SeedLocalUserRecord(t, harness.DB, "entity_linking-i401-other@example.test", "EntityLinking I401 Other", "EntityLinkingI401OtherPass1!", false, false, true)
		otherIncident := appsupport.CreateIncidentInStore(t, harness.Pool, otherActor, "txn-entity_linking-i-4-01-hidden-incident", "IR-I401-H", "Record relationships entity-resolution hidden")
		entitytest.SeedHostRecord(t, harness.DB, otherIncident.ID, otherActor.ID, entitytest.DuplicateHostRecordID, "Hidden WS-023", "HIDDEN-WS-023", "", "")

		hiddenResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-hidden", "resolve_item", uuidPointer(entitytest.DuplicateHostRecordID), nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, hiddenResp, http.StatusNotFound, "resolved_record_not_found")

		otherVisibleIncident := appsupport.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-entity_linking-i-4-01-visible-incident",
			"incident_key":  "IR-I401-V",
			"title":         "Record relationships entity-resolution visible other incident",
		})
		otherVisibleIncidentID := appsupport.MustUUID(t, otherVisibleIncident["incident_id"].(string))
		entitytest.SeedHostRecord(t, harness.DB, otherVisibleIncidentID, adminUserID, entitytest.StubHostRecordID, "Visible WS-023", "VISIBLE-WS-023", "", "")

		crossIncidentResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-cross", "resolve_item", uuidPointer(entitytest.StubHostRecordID), nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, crossIncidentResp, http.StatusBadRequest, "invalid_mutation_payload")

		wrongTypeResp := appsupport.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+entitytest.HostMentionID.String()+"/resolve",
			entitytest.MentionResolveRoutePayload(1, "txn-entity_linking-i-4-01-wrong-type", "resolve_item", uuidPointer(entitytest.CanonicalIdentityRecordID), nil),
			appsupport.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			appsupport.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		appsupport.RequireErrorBody(t, wrongTypeResp, http.StatusBadRequest, "invalid_mutation_payload")

		mention := entitytest.LookupMention(t, harness.DB, entitytest.HostMentionID)
		entitytest.RequireMentionStatus(t, mention, entitytest.MentionStatusUnresolved)
		envelopetest.RequireRowVersionStable(t, 1, mention.RowVersion)
		if got := appsupport.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND deleted_at IS NULL
`, incidentID, timelinetest.RecordID); got != 0 {
			t.Fatalf("expected invalid or unauthorized target attempts to leave record links untouched, got %d active rows", got)
		}
	})
}

// entity-resolution / REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063 / AC-022, AC-186.
