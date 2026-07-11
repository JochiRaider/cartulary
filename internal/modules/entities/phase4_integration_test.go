package entities_test

import (
	"context"
	"database/sql"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/assertx"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/fixtures"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/auditassert"
	"github.com/JochiRaider/cartulary/internal/testutil/contractassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// I-4-01 / REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044 / AC-188..AC-190, AC-221..AC-225.
func TestPhase4_ResolveRoute_I_4_01(t *testing.T) {
	t.Run("resolve_item uses authoritative route replay and history conformance", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-01-resolve-conformance")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-resolve-conf-incident",
			"incident_key":  "IR-I401-RC",
			"title":         "Phase 4 I-4-01 resolve conformance",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		workbookscenariotest.SeedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		ctx := workbookscenariotest.RouteInventoryContext{
			IncidentID:       incidentID.String(),
			ActorUserID:      adminUserID.String(),
			TimelineRecordID: golden.RecordTimelineRecordID.String(),
			MentionID:        golden.RecordHostMentionID.String(),
			HostRecordID:     golden.RecordCanonicalHostRecordID.String(),
		}
		route := workbookscenariotest.MustRoute(t, workbookscenariotest.RouteMentionResolve, ctx)
		workbookscenariotest.RequireRouteReplayHistoryConformance(t, harness.DB, harness.Server.HTTP.URL, workbookscenariotest.RouteConformanceCase{
			Route:                  route,
			Context:                ctx,
			ClientTxnID:            "txn-phase4-i-4-01-resolve-conformance",
			Login:                  adminLogin,
			ActorUserID:            adminUserID.String(),
			ExpectedMutationSource: "entities.entity_mentions.resolve",
		})

		mention := workbookscenariotest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected route conformance resolve to leave mention resolved to target, got %#v", mention)
		}
	})

	t.Run("resolve_item persists durable state, replays idempotently, and emits websocket invalidation", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-01-resolve")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-resolve-incident",
			"incident_key":  "IR-I401-R",
			"title":         "Phase 4 I-4-01 resolve route",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		workbookscenariotest.SeedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordTimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		payload := fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-resolve", golden.RecordMentionActionResolve, uuidPointer(golden.RecordCanonicalHostRecordID), nil)
		resp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			payload,
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := workbookscenariotest.RequireSuccessData(t, resp, http.StatusOK)
		if data["incident_id"] != incidentID.String() {
			t.Fatalf("unexpected incident_id in mention action payload: %#v", data)
		}
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after resolve_item, got %#v", data)
		}

		mention := workbookscenariotest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected resolve_item to point mention at the target record, got %#v", mention)
		}
		if mention.ResolutionMethod == nil || *mention.ResolutionMethod != "explicit_resolve_route" {
			t.Fatalf("expected resolve route provenance marker, got %#v", mention)
		}
		link := workbookscenariotest.LookupActiveLink(t, harness.DB, incidentID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host", "manual", nil)

		socketPayload := workbookscenariotest.RequireRecordChanged(t, socket, golden.RecordTimelineRecordID.String(), 2)
		if socketPayload.ChangeSetID != data["change_set_id"] {
			t.Fatalf("expected websocket to carry the mention action change_set_id, got payload=%#v response=%#v", socketPayload, data)
		}
		hostSocketPayload := workbookscenariotest.RequireRecordChanged(t, hostSocket, golden.RecordCanonicalHostRecordID.String(), 1)
		if hostSocketPayload.ChangeSetID != data["change_set_id"] {
			t.Fatalf("expected host websocket invalidation to carry the mention action change_set_id, got payload=%#v response=%#v", hostSocketPayload, data)
		}
		if !stringSliceContains(hostSocketPayload.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected host linked-event count invalidation, got %#v", hostSocketPayload)
		}
		viewLogin := workbookscenariotest.LoginResult{
			SessionCookie: adminLogin.SessionCookie,
			CSRFCookie:    adminLogin.CSRFCookie,
		}
		hostRow := workbookscenariotest.FindRow(
			t,
			workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin),
			golden.RecordCanonicalHostRecordID.String(),
		)
		hostCells := hostRow["cells"].(map[string]any)
		if got := hostCells["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected resolved host to expose one linked event, got %#v row=%#v", got, hostRow)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE host_grid_projection SET linked_event_count = 0 WHERE record_id = $1`, golden.RecordCanonicalHostRecordID); err != nil {
			t.Fatalf("corrupt resolved host linked-event count: %v", err)
		}
		if err := projections.NewStore(harness.Server.Runtime.Postgres).RebuildIncidentHosts(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild resolved host projection: %v", err)
		}
		rebuiltHostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), golden.RecordCanonicalHostRecordID.String())
		if got := rebuiltHostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected rebuild to restore one linked event, got %#v row=%#v", got, rebuiltHostRow)
		}

		replayResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			payload,
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		replayData := workbookscenariotest.RequireSuccessData(t, replayResp, http.StatusOK)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("expected replay to reuse the original payload, got %#v %#v", data, replayData)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, "entities.entity_mentions.resolve", adminUserID.String(), golden.RecordHostMentionID.String(), "txn-phase4-i-4-01-resolve"); got != 1 {
			t.Fatalf("expected one route idempotency row for a replayed mention action, got %d", got)
		}

		divergentResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-resolve", "dismiss_item", nil, nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("dismiss_item removes active links and fails closed on stale or illegal transitions", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-01-dismiss")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-dismiss-incident",
			"incident_key":  "IR-I401-D",
			"title":         "Phase 4 I-4-01 dismiss route",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		workbookscenariotest.SeedResolvedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023")
		workbookscenariotest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, golden.RecordManualLinkID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host", "manual", nil)

		socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordTimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		resp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-dismiss", "dismiss_item", nil, nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := workbookscenariotest.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["entity_mention"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected mention row_version=2 after dismiss_item, got %#v", data)
		}

		dismissed := workbookscenariotest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, dismissed, golden.RecordMentionStatusDismissed)
		if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incidentID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID); got != 0 {
			t.Fatalf("expected dismiss_item to remove the active link, got %d active rows", got)
		}
		workbookscenariotest.RequireRecordChanged(t, socket, golden.RecordTimelineRecordID.String(), 2)
		hostChange := workbookscenariotest.RequireRecordChanged(t, hostSocket, golden.RecordCanonicalHostRecordID.String(), 1)
		if !stringSliceContains(hostChange.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected dismiss to invalidate host linked-event count, got %#v", hostChange)
		}
		viewLogin := workbookscenariotest.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), golden.RecordCanonicalHostRecordID.String())
		if got := hostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("expected dismiss to project zero linked events, got %#v row=%#v", got, hostRow)
		}

		staleResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-stale", "revert_to_unresolved", nil, nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		staleBody := workbookscenariotest.RequireErrorBody(t, staleResp, http.StatusConflict, "row_version_conflict")
		if staleBody["error"].(map[string]any)["details"].(map[string]any)["current_mention_row_version"] != float64(2) {
			t.Fatalf("expected current_mention_row_version=2 on stale update, got %#v", staleBody)
		}

		illegalResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(2, "txn-phase4-i-4-01-illegal", "dismiss_item", nil, nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		illegalBody := workbookscenariotest.RequireErrorBody(t, illegalResp, http.StatusConflict, "illegal_transition")
		if illegalBody["error"].(map[string]any)["details"].(map[string]any)["from_status"] != "dismissed" {
			t.Fatalf("expected illegal transition details to report the dismissed source status, got %#v", illegalBody)
		}
		hostRowAfterFailures := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), golden.RecordCanonicalHostRecordID.String())
		if got := hostRowAfterFailures["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("failed mention actions changed the projected linked-event count: %#v", hostRowAfterFailures)
		}
	})

	t.Run("revert_to_unresolved restores unresolved mention state and emits websocket invalidation", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-01-revert")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-revert-incident",
			"incident_key":  "IR-I401-U",
			"title":         "Phase 4 I-4-01 revert route",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		workbookscenariotest.SeedResolvedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023")
		workbookscenariotest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, golden.RecordManualLinkID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host", "manual", nil)

		socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordTimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")
		hostSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		resp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-revert", "revert_to_unresolved", nil, nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := workbookscenariotest.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after revert_to_unresolved, got %#v", data)
		}

		mention := workbookscenariotest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusUnresolved)
		if mention.RawText != "WS-023" || mention.ResolvedRecordID != nil || mention.ResolutionMethod != nil {
			t.Fatalf("expected revert_to_unresolved to preserve raw_text and clear resolution metadata, got %#v", mention)
		}
		workbookscenariotest.RequireRecordChanged(t, socket, golden.RecordTimelineRecordID.String(), 2)
		hostChange := workbookscenariotest.RequireRecordChanged(t, hostSocket, golden.RecordCanonicalHostRecordID.String(), 1)
		if !stringSliceContains(hostChange.ChangedFieldKeys, "host.linked_event_count") {
			t.Fatalf("expected revert to invalidate host linked-event count, got %#v", hostChange)
		}
		viewLogin := workbookscenariotest.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), golden.RecordCanonicalHostRecordID.String())
		if got := hostRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(0) {
			t.Fatalf("expected revert to project zero linked events, got %#v row=%#v", got, hostRow)
		}
	})

	t.Run("authorization and target validation are re-derived from live state", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-01-access")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-access-incident",
			"incident_key":  "IR-I401-A",
			"title":         "Phase 4 I-4-01 access checks",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		workbookscenariotest.SeedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

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
		deniedResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-denied", "resolve_item", uuidPointer(golden.RecordCanonicalHostRecordID), nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")

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

		otherActor := workbookscenariotest.SeedLocalUserFlags(t, harness.DB, "phase4-i401-other@example.test", "Phase4 I401 Other", "Phase4I401OtherPass1!", false, false, true)
		otherIncident := workbookscenariotest.CreateIncidentInStore(t, harness.Server.Runtime.Postgres, otherActor, "txn-phase4-i-4-01-hidden-incident", "IR-I401-H", "Phase 4 I-4-01 hidden")
		workbookscenariotest.SeedHostRecord(t, harness.DB, otherIncident.ID, otherActor.ID, golden.RecordDuplicateHostRecordID, "Hidden WS-023", "HIDDEN-WS-023", "", "")

		hiddenResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-hidden", "resolve_item", uuidPointer(golden.RecordDuplicateHostRecordID), nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, hiddenResp, http.StatusNotFound, "resolved_record_not_found")

		otherVisibleIncident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-visible-incident",
			"incident_key":  "IR-I401-V",
			"title":         "Phase 4 I-4-01 visible other incident",
		})
		otherVisibleIncidentID := workbookscenariotest.MustUUID(t, otherVisibleIncident["incident_id"].(string))
		workbookscenariotest.SeedHostRecord(t, harness.DB, otherVisibleIncidentID, adminUserID, golden.RecordStubHostRecordID, "Visible WS-023", "VISIBLE-WS-023", "", "")

		crossIncidentResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-cross", "resolve_item", uuidPointer(golden.RecordStubHostRecordID), nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, crossIncidentResp, http.StatusBadRequest, "invalid_mutation_payload")

		wrongTypeResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.RecordHostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-wrong-type", "resolve_item", uuidPointer(golden.RecordCanonicalIdentityID), nil),
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, wrongTypeResp, http.StatusBadRequest, "invalid_mutation_payload")

		mention := workbookscenariotest.LookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusUnresolved)
		assertx.RequireRowVersionStable(t, 1, mention.RowVersion)
		if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND deleted_at IS NULL
`, incidentID, golden.RecordTimelineRecordID); got != 0 {
			t.Fatalf("expected invalid or unauthorized target attempts to leave record links untouched, got %d active rows", got)
		}
	})
}

// I-4-02 / REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063 / AC-022, AC-186.
func TestPhase4_EntityOriginUpsert_I_4_02(t *testing.T) {
	t.Run("host create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-02-host")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-host-incident",
			"incident_key":  "IR-I402-H",
			"title":         "Phase 4 I-4-02 host create",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

		hostCreate := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-create",
				"host.display_name": "Gateway record",
				"host.hostname":     "GATEWAY-01",
				"host.fqdn":         "gateway-01.corp.example",
				"host.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "VPN Gateway"},
					},
				},
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostCreateData := workbookscenariotest.RequireSuccessData(t, hostCreate, http.StatusCreated)
		hostRow := hostCreateData["row"].(map[string]any)
		hostRecordID := workbookscenariotest.MustUUID(t, hostRow["record_id"].(string))

		var (
			hostState        string
			entityOrigin     string
			seedMentionID    sql.NullString
			displayName      string
			hostname         string
			fqdn             sql.NullString
			suggestionAliasN int
		)
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    host_state,
    entity_origin,
    seed_entity_mention_id::text,
    display_name,
    hostname,
    fqdn,
    (
      SELECT COUNT(*)
        FROM entity_aliases
       WHERE incident_id = h.incident_id
         AND record_id = h.record_id
         AND entity_type = 'host'
         AND classification = 'suggestion_only'
         AND deleted_at IS NULL
    )
  FROM hosts h
 WHERE record_id = $1
`, hostRecordID).Scan(&hostState, &entityOrigin, &seedMentionID, &displayName, &hostname, &fqdn, &suggestionAliasN); err != nil {
			t.Fatalf("lookup created host row: %v", err)
		}
		if hostState != "stub" || entityOrigin != "entity_sheet" || seedMentionID.Valid {
			t.Fatalf("expected entity_sheet host provenance without seed mention, got state=%q origin=%q seed=%v", hostState, entityOrigin, seedMentionID)
		}
		requireEntityOriginDefault(t, harness.DB, "hosts", "entity_sheet")
		requireEntityOriginRejected(t, harness.DB, "hosts", hostRecordID, "direct_create")
		requireEntityOriginRejected(t, harness.DB, "hosts", hostRecordID, "not_a_core02_origin")
		if displayName != "Gateway record" || hostname != "GATEWAY-01" || !fqdn.Valid || fqdn.String != "gateway-01.corp.example" || suggestionAliasN != 1 {
			t.Fatalf("unexpected created host state: display=%q hostname=%q fqdn=%v aliases=%d", displayName, hostname, fqdn, suggestionAliasN)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("entity-origin host create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE hosts SET fqdn = NULL WHERE record_id = $1`, hostRecordID); err != nil {
			t.Fatalf("clear host fqdn to force preserved-identifier reuse: %v", err)
		}
		hostReuse := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-reuse",
				"host.display_name": "Gateway reused",
				"host.fqdn":         "gateway-01.corp.example",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReuseData := workbookscenariotest.RequireSuccessData(t, hostReuse, http.StatusOK)
		if got := workbookscenariotest.MustUUID(t, hostReuseData["row"].(map[string]any)["record_id"].(string)); got != hostRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original host record, got %#v", hostReuseData)
		}
		if state, mergedInto, rowVersion, restoredFQDN := workbookscenariotest.LookupHostState(t, harness.DB, hostRecordID); state != "stub" || mergedInto != nil || rowVersion != 2 || restoredFQDN != "gateway-01.corp.example" {
			t.Fatalf("unexpected reused host state: state=%s merged_into=%v row_version=%d fqdn=%q", state, mergedInto, rowVersion, restoredFQDN)
		}

		aliasOnly := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-alias-only",
				"host.display_name": "VPN Gateway",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := workbookscenariotest.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := workbookscenariotest.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == hostRecordID {
			t.Fatalf("expected alias-only create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		aliasPayloadOnly := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-02-host-alias-payload-only",
				"host.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "VPN Gateway Only"},
					},
				},
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, aliasPayloadOnly, http.StatusBadRequest, "invalid_mutation_payload")

		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "Conflict Host A", "COLLISION-01", "", "")
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "Conflict Host B", "COLLISION-01", "", "")
		conflictResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-conflict",
				"host.display_name": "Conflict Host",
				"host.hostname":     "COLLISION-01",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := workbookscenariotest.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "host" || details["identifier_class"] != "hostname" {
			t.Fatalf("unexpected host conflict details: %#v", details)
		}
		candidateIDs := details["candidate_record_ids"].([]any)
		if len(candidateIDs) != 2 {
			t.Fatalf("expected two conflict candidates, got %#v", details)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("conflicted host create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("identity create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-02-identity")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-identity-incident",
			"incident_key":  "IR-I402-I",
			"title":         "Phase 4 I-4-02 identity create",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))

		identityCreate := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":             "txn-phase4-i-4-02-identity-create",
				"identity.display_name":     "Alex Analyst",
				"identity.email":            "alex.analyst@example.test",
				"identity.sam_account_name": "ALEXA",
				"identity.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "Case Owner"},
					},
				},
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityCreateData := workbookscenariotest.RequireSuccessData(t, identityCreate, http.StatusCreated)
		identityRecordID := workbookscenariotest.MustUUID(t, identityCreateData["row"].(map[string]any)["record_id"].(string))

		var (
			identityState   string
			entityOrigin    string
			seedMentionID   sql.NullString
			displayName     string
			email           sql.NullString
			samAccountName  sql.NullString
			suggestionCount int
		)
		if err := harness.DB.QueryRowContext(context.Background(), `
SELECT
    identity_state,
    entity_origin,
    seed_entity_mention_id::text,
    display_name,
    email::text,
    sam_account_name,
    (
      SELECT COUNT(*)
        FROM entity_aliases
       WHERE incident_id = i.incident_id
         AND record_id = i.record_id
         AND entity_type = 'identity'
         AND classification = 'suggestion_only'
         AND deleted_at IS NULL
    )
  FROM identities i
 WHERE record_id = $1
`, identityRecordID).Scan(&identityState, &entityOrigin, &seedMentionID, &displayName, &email, &samAccountName, &suggestionCount); err != nil {
			t.Fatalf("lookup created identity row: %v", err)
		}
		if identityState != "stub" || entityOrigin != "entity_sheet" || seedMentionID.Valid {
			t.Fatalf("expected entity_sheet identity provenance without seed mention, got state=%q origin=%q seed=%v", identityState, entityOrigin, seedMentionID)
		}
		requireEntityOriginDefault(t, harness.DB, "identities", "entity_sheet")
		requireEntityOriginRejected(t, harness.DB, "identities", identityRecordID, "direct_create")
		requireEntityOriginRejected(t, harness.DB, "identities", identityRecordID, "not_a_core02_origin")
		if displayName != "Alex Analyst" || !email.Valid || email.String != "alex.analyst@example.test" || !samAccountName.Valid || samAccountName.String != "ALEXA" || suggestionCount != 1 {
			t.Fatalf("unexpected created identity state: display=%q email=%v sam=%v aliases=%d", displayName, email, samAccountName, suggestionCount)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("entity-origin identity create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE identities SET email = NULL WHERE record_id = $1`, identityRecordID); err != nil {
			t.Fatalf("clear identity email to force preserved-identifier reuse: %v", err)
		}
		identityReuse := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-reuse",
				"identity.display_name": "Alex Analyst Reused",
				"identity.email":        "alex.analyst@example.test",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityReuseData := workbookscenariotest.RequireSuccessData(t, identityReuse, http.StatusOK)
		if got := workbookscenariotest.MustUUID(t, identityReuseData["row"].(map[string]any)["record_id"].(string)); got != identityRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original identity record, got %#v", identityReuseData)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM identities
 WHERE incident_id = $1
   AND record_id = $2
   AND row_version = 2
   AND email = 'alex.analyst@example.test'
`, incidentID, identityRecordID); got != 1 {
			t.Fatalf("expected preserved-identifier reuse to restore the canonical email and increment row_version, got %d rows", got)
		}

		aliasOnly := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-alias-only",
				"identity.display_name": "Case Owner",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := workbookscenariotest.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := workbookscenariotest.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == identityRecordID {
			t.Fatalf("expected alias-only identity create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		aliasPayloadOnly := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-02-identity-alias-payload-only",
				"identity.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_alias", "alias_text": "Case Owner Only"},
					},
				},
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		workbookscenariotest.RequireErrorBody(t, aliasPayloadOnly, http.StatusBadRequest, "invalid_mutation_payload")

		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalIdentityID, "Conflict Identity A", "collision@example.test", "collision@example.test", "COLLISION-A")
		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateIdentityID, "Conflict Identity B", "collision@example.test", "collision@example.test", "COLLISION-B")
		conflictResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-conflict",
				"identity.display_name": "Conflict Identity",
				"identity.email":        "collision@example.test",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := workbookscenariotest.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "identity" || details["identifier_class"] != "email" {
			t.Fatalf("unexpected identity conflict details: %#v", details)
		}
		if got := len(details["candidate_record_ids"].([]any)); got != 2 {
			t.Fatalf("expected two identity conflict candidates, got %#v", details)
		}
		if got := workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("conflicted identity create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("create routes emit history, round-trip current-state query reads, and re-derive live authorization", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-02-query-auth")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-query-auth-incident",
			"incident_key":  "IR-I402-Q",
			"title":         "Phase 4 I-4-02 query and auth",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		viewLogin := workbookscenariotest.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}

		hostPayload := map[string]any{
			"client_txn_id":     "txn-phase4-i-4-02-query-host",
			"host.display_name": "Gateway query host",
			"host.hostname":     "GATEWAY-Q-01",
			"host.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "Gateway Query Alias"},
				},
			},
		}
		hostResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			hostPayload,
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostData := workbookscenariotest.RequireSuccessData(t, hostResp, http.StatusCreated)
		hostRecordID := workbookscenariotest.MustUUID(t, hostData["row"].(map[string]any)["record_id"].(string))
		hostChangeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), hostData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: hostChangeSet.ActorUserID,
			Source:      hostChangeSet.Source,
			ClientTxnID: hostChangeSet.ClientTxnID,
			RequestID:   hostChangeSet.RequestID,
			CreatedAt:   hostChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.hosts.rows.create", "txn-phase4-i-4-02-query-host")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), hostData["change_set_id"].(string)); got != 2 {
			t.Fatalf("expected host and alias create mutation rows, got %d", got)
		}

		identityPayload := map[string]any{
			"client_txn_id":             "txn-phase4-i-4-02-query-identity",
			"identity.display_name":     "Alex Query",
			"identity.email":            "alex.query@example.test",
			"identity.sam_account_name": "ALEXQ",
			"identity.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_alias", "alias_text": "Query Owner"},
				},
			},
		}
		identityResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			identityPayload,
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityData := workbookscenariotest.RequireSuccessData(t, identityResp, http.StatusCreated)
		identityRecordID := workbookscenariotest.MustUUID(t, identityData["row"].(map[string]any)["record_id"].(string))
		identityChangeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), identityData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: identityChangeSet.ActorUserID,
			Source:      identityChangeSet.Source,
			ClientTxnID: identityChangeSet.ClientTxnID,
			RequestID:   identityChangeSet.RequestID,
			CreatedAt:   identityChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.identities.rows.create", "txn-phase4-i-4-02-query-identity")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), identityData["change_set_id"].(string)); got != 2 {
			t.Fatalf("expected identity and alias create mutation rows, got %d", got)
		}

		hostEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, hostEnvelope, golden.RecordHostsViewSchemaID)
		hostRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), hostRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", hostRow, golden.RecordHostsViewSchemaID)
		hostAlias := workbookscenariotest.RequireSingleCollectionItem(t, hostRow, "host.aliases")
		if hostAlias["item_kind"] != "alias" || hostAlias["alias_text"] != "Gateway Query Alias" || !strings.HasPrefix(hostAlias["item_ref"].(string), "entity_alias:") {
			t.Fatalf("unexpected host alias readback: %#v", hostAlias)
		}

		identityEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordIdentitiesViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, identityEnvelope, golden.RecordIdentitiesViewSchemaID)
		identityRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordIdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", identityRow, golden.RecordIdentitiesViewSchemaID)
		identityAlias := workbookscenariotest.RequireSingleCollectionItem(t, identityRow, "identity.aliases")
		if identityAlias["item_kind"] != "alias" || identityAlias["alias_text"] != "Query Owner" || !strings.HasPrefix(identityAlias["item_ref"].(string), "entity_alias:") {
			t.Fatalf("unexpected identity alias readback: %#v", identityAlias)
		}

		hostProjectionBefore := lookupHostProjectionSnapshot(t, harness.DB, hostRecordID)
		identityProjectionBefore := lookupIdentityProjectionSnapshot(t, harness.DB, identityRecordID)
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM host_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
			t.Fatalf("clear host projection rows: %v", err)
		}
		if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM identity_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
			t.Fatalf("clear identity projection rows: %v", err)
		}
		projectionStore := projections.NewStore(harness.Server.Runtime.Postgres)
		if err := projectionStore.RebuildIncidentHosts(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild host projections: %v", err)
		}
		if err := projectionStore.RebuildIncidentIdentities(context.Background(), incidentID); err != nil {
			t.Fatalf("rebuild identity projections: %v", err)
		}
		hostProjectionAfter := lookupHostProjectionSnapshot(t, harness.DB, hostRecordID)
		identityProjectionAfter := lookupIdentityProjectionSnapshot(t, harness.DB, identityRecordID)
		contractassert.RequireProjectionDeterminism(t, hostProjectionBefore, hostProjectionAfter)
		contractassert.RequireProjectionDeterminism(t, identityProjectionBefore, identityProjectionAfter)

		hostRowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), hostRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", hostRowAfterRebuild, golden.RecordHostsViewSchemaID)
		contractassert.RequireProjectionDeterminism(t, hostRow["cells"], hostRowAfterRebuild["cells"])
		if rebuiltHostAlias := workbookscenariotest.RequireSingleCollectionItem(t, hostRowAfterRebuild, "host.aliases"); rebuiltHostAlias["alias_text"] != "Gateway Query Alias" {
			t.Fatalf("unexpected rebuilt host alias readback: %#v", rebuiltHostAlias)
		}

		identityRowAfterRebuild := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordIdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", identityRowAfterRebuild, golden.RecordIdentitiesViewSchemaID)
		contractassert.RequireProjectionDeterminism(t, identityRow["cells"], identityRowAfterRebuild["cells"])
		if rebuiltIdentityAlias := workbookscenariotest.RequireSingleCollectionItem(t, identityRowAfterRebuild, "identity.aliases"); rebuiltIdentityAlias["alias_text"] != "Query Owner" {
			t.Fatalf("unexpected rebuilt identity alias readback: %#v", rebuiltIdentityAlias)
		}

		replayStableBefore := contractassert.ReplayCounts{
			ChangeSets: workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		}
		hostReplay := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			hostPayload,
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReplayData := workbookscenariotest.RequireSuccessData(t, hostReplay, http.StatusOK)
		if hostReplayData["change_set_id"] != hostData["change_set_id"] {
			t.Fatalf("expected host replay to reuse the original payload, got %#v %#v", hostData, hostReplayData)
		}
		contractassert.RequireReplayScaffold(t, contractassert.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore:    replayStableBefore,
			StableAfter: contractassert.ReplayCounts{
				ChangeSets: workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
				MutationRows: workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
			},
		})

		hostDivergent := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-query-host",
				"host.display_name": "Gateway query host divergent",
				"host.hostname":     "GATEWAY-Q-01",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostDivergentBody := workbookscenariotest.RequireErrorBody(t, hostDivergent, http.StatusConflict, "client_txn_conflict")
		contractassert.RequireDivergentReplayRejected(t, hostDivergent.StatusCode, hostDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote entity create actor membership: %v", err)
		}
		deniedResp := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-query-host-denied",
				"host.display_name": "Denied host",
				"host.hostname":     "DENIED-HOST",
			},
			workbookscenariotest.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		deniedBody := workbookscenariotest.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")
		contractassert.RequireAuthorizationReDerived(
			t,
			contractassert.AuthorizationOutcome{Status: http.StatusCreated},
			contractassert.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
		)
	})
}

func requireEntityOriginDefault(t *testing.T, db *sql.DB, tableName string, want string) {
	t.Helper()

	if tableName != "hosts" && tableName != "identities" {
		t.Fatalf("unsupported entity_origin table %q", tableName)
	}

	var defaultExpression string
	if err := db.QueryRowContext(context.Background(), `
SELECT column_default
  FROM information_schema.columns
 WHERE table_schema = 'public'
   AND table_name = $1
   AND column_name = 'entity_origin'
`, tableName).Scan(&defaultExpression); err != nil {
		t.Fatalf("lookup %s.entity_origin default: %v", tableName, err)
	}
	if !strings.Contains(defaultExpression, "'"+want+"'") {
		t.Fatalf("expected %s.entity_origin default %q, got %q", tableName, want, defaultExpression)
	}
}

func requireEntityOriginRejected(t *testing.T, db *sql.DB, tableName string, recordID uuid.UUID, origin string) {
	t.Helper()

	var query string
	switch tableName {
	case "hosts":
		query = `UPDATE hosts SET entity_origin = $1 WHERE record_id = $2`
	case "identities":
		query = `UPDATE identities SET entity_origin = $1 WHERE record_id = $2`
	default:
		t.Fatalf("unsupported entity_origin table %q", tableName)
	}

	if _, err := db.ExecContext(context.Background(), query, origin, recordID); err == nil {
		t.Fatalf("expected %s.entity_origin to reject %q", tableName, origin)
	}
}

func TestPhase4_EntityCreateAuthAndCSRFFailBeforeMalformedBody_I_4_02(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase4-entity-create-auth-csrf-order")
	adminLogin, _ := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-entity-create-auth-order-incident",
		"incident_key":  "IR-AUTH-CSRF-ORDER",
		"title":         "Entity create auth csrf ordering",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
	socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	type entityCreateFailureCounts struct {
		Records        int
		ChangeSets     int
		MutationRows   int
		HostProjection int
	}
	counts := func() entityCreateFailureCounts {
		return entityCreateFailureCounts{
			Records:        workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID),
			ChangeSets:     workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows:   workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, incidentID),
			HostProjection: workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE incident_id = $1`, incidentID),
		}
	}
	before := counts()
	url := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + golden.RecordHostsViewSchemaID + "/rows"

	unauthenticated := doEntitiesRawJSON(t, http.MethodPost, url, "{")
	workbookscenariotest.RequireErrorBody(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	workbookscenariotest.RequireErrorBody(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		withHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	workbookscenariotest.RequireErrorBody(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	if after := counts(); after != before {
		t.Fatalf("auth/csrf failures must not mutate entity state: before=%#v after=%#v", before, after)
	}
	workbookscenariotest.ExpectNoSocketMessage(t, socket)
}

// I-4-03 / REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitMergeRoute_I_4_03(t *testing.T) {
	t.Run("entity route failures enforce authentication csrf visibility role and body precedence", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-entity-route-failure-precedence")
		adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
		incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-route-precedence-incident",
			"incident_key":  "IR-ROUTE-PRECEDENCE",
			"title":         "Entity route failure precedence",
		})
		incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "Survivor", "SURVIVOR", "", "")
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "Loser", "LOSER", "", "")
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordFieldTimelineHostRefs, "host", "SURVIVOR", "unresolved", nil, nil)

		type failureCounts struct {
			changeSets      int
			mutations       int
			idempotencyRows int
			hostProjection  int
		}
		counts := func() failureCounts {
			return failureCounts{
				changeSets:      workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
				mutations:       workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, incidentID),
				idempotencyRows: workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM route_idempotency WHERE scope_key IN ($1, $2, $3)`, golden.RecordCanonicalHostRecordID.String(), golden.RecordDuplicateHostRecordID.String(), golden.RecordHostMentionID.String()),
				hostProjection:  workbookscenariotest.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE incident_id = $1`, incidentID),
			}
		}
		before := counts()
		socket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")

		mergeMalformedPath := harness.Server.HTTP.URL + "/api/v1/records/not-a-uuid/merge"
		mentionMalformedPath := harness.Server.HTTP.URL + "/api/v1/entity-mentions/not-a-uuid/resolve"
		for _, url := range []string{mergeMalformedPath, mentionMalformedPath} {
			workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{"), http.StatusUnauthorized, "session_required")
			workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie)), http.StatusForbidden, "csrf_verification_failed")
		}
		workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, mergeMalformedPath, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "incident_not_found")
		workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, mentionMalformedPath, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "entity_mention_not_found")

		hiddenRecordID := uuid.New()
		hiddenMentionID := uuid.New()
		workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+hiddenRecordID.String()+"/merge", "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "incident_not_found")
		workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+hiddenMentionID.String()+"/resolve", "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusNotFound, "entity_mention_not_found")

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role = 'viewer', updated_at = now(), updated_by_user_id = $3 WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote entity route actor: %v", err)
		}
		mergeURL := harness.Server.HTTP.URL + "/api/v1/records/" + golden.RecordCanonicalHostRecordID.String() + "/merge"
		mentionURL := harness.Server.HTTP.URL + "/api/v1/entity-mentions/" + golden.RecordHostMentionID.String() + "/resolve"
		for _, url := range []string{mergeURL, mentionURL} {
			workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusForbidden, "authorization_denied")
		}
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incident_memberships SET role = 'admin', updated_at = now(), updated_by_user_id = $3 WHERE incident_id = $1 AND user_id = $2`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("restore entity route actor: %v", err)
		}
		for _, url := range []string{mergeURL, mentionURL} {
			workbookscenariotest.RequireErrorBody(t, doEntitiesRawJSON(t, http.MethodPost, url, "{", withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value)), http.StatusBadRequest, "invalid_mutation_payload")
		}

		otherActor := workbookscenariotest.SeedLocalUserFlags(t, harness.DB, "phase4-route-precedence-other@example.test", "Phase4 Route Other", "Phase4RouteOtherPass1!", false, false, true)
		otherIncident := workbookscenariotest.CreateIncidentInStore(t, harness.Server.Runtime.Postgres, otherActor, "txn-phase4-route-precedence-hidden-incident", "IR-ROUTE-PRECEDENCE-HIDDEN", "Entity route hidden incident")
		hiddenLoserID := uuid.New()
		workbookscenariotest.SeedHostRecord(t, harness.DB, otherIncident.ID, otherActor.ID, hiddenLoserID, "Hidden loser", "HIDDEN-LOSER", "", "")

		for _, loserID := range []string{"not-a-uuid", uuid.New().String(), hiddenLoserID.String()} {
			resp := doEntitiesJSON(t, http.MethodPost, mergeURL, map[string]any{
				"loser_record_id":           loserID,
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-route-precedence-" + loserID,
			}, withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie), withHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value))
			workbookscenariotest.RequireErrorBody(t, resp, http.StatusNotFound, "incident_not_found")
		}

		if after := counts(); after != before {
			t.Fatalf("entity route failures mutated durable state: before=%#v after=%#v", before, after)
		}
		workbookscenariotest.ExpectNoSocketMessage(t, socket)
	})

	t.Run("host merge repoints live fan-out and preserves survivor reuse", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-03")
		workbookscenariotest.RequireSchemaTables(t, harness.DB, "I-4-03", "hosts", "identities", "entity_mentions", "record_tags", "assessments")

		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-incident",
			"incident_key":  "IR-I403",
			"title":         "Entity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		timelineSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordTimelineViewSchemaID, adminLogin.sessionCookie.Value)
		defer timelineSocket.Close(1000, "test_complete")
		hostSocket := workbookscenariotest.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.RecordHostsViewSchemaID, adminLogin.sessionCookie.Value)
		defer hostSocket.Close(1000, "test_complete")

		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		seedEntityAlias(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "host", "Workstation 23")
		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		seedResolvedMention(t, harness.DB, adminUserID, golden.RecordHostMentionID, golden.RecordTimelineRecordID, golden.RecordDuplicateHostRecordID, golden.RecordFieldTimelineHostRefs, "WS-023")
		seedRecordLink(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateLinkID, golden.RecordTimelineRecordID, golden.RecordDuplicateHostRecordID, "observed_on_host", "manual", nil)
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.RecordTagIDSurvivor, golden.RecordCanonicalHostRecordID, "critical-host")
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.RecordTagIDLoser, golden.RecordDuplicateHostRecordID, "critical-host")
		seedAssessment(t, harness.DB, incidentID, adminUserID, golden.RecordAssessmentHostID, golden.RecordDuplicateHostRecordID, "host", "confirmed")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-merge",
				"reason":                    "  merge duplicate host  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		if mergeResp.StatusCode != http.StatusOK {
			t.Fatalf("unexpected status: got %d want %d body=%#v", mergeResp.StatusCode, http.StatusOK, httptestx.ReadJSONBody(t, mergeResp))
		}
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != golden.RecordCanonicalHostRecordID.String() {
			t.Fatalf("unexpected survivor_record_id: %#v", mergeData)
		}
		if mergeData["loser_record_id"] != golden.RecordDuplicateHostRecordID.String() {
			t.Fatalf("unexpected loser_record_id: %#v", mergeData)
		}
		if mergeData["record_type"] != "host" {
			t.Fatalf("unexpected canonical record_type: %#v", mergeData)
		}
		if got := int64(mergeData["survivor_row_version"].(float64)); got != 2 {
			t.Fatalf("expected survivor_row_version=2, got %d", got)
		}
		if got := int64(mergeData["loser_row_version"].(float64)); got != 2 {
			t.Fatalf("expected loser_row_version=2, got %d", got)
		}
		if mergeData["merged_into_record_id"] != golden.RecordCanonicalHostRecordID.String() {
			t.Fatalf("expected merged_into_record_id to echo survivor, got %#v", mergeData)
		}

		summary := mergeData["merge_summary"].(map[string]any)
		if summary["record_type"] != "host" {
			t.Fatalf("unexpected merge summary record_type: %#v", summary)
		}
		if got := int(summary["repointed_mention_resolution_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed mention resolution, got %d", got)
		}
		if got := int(summary["repointed_link_count"].(float64)); got != 1 {
			t.Fatalf("expected one repointed link, got %d", got)
		}
		if got := int(summary["deduped_tag_count"].(float64)); got != 1 {
			t.Fatalf("expected one deduped tag, got %d", got)
		}
		if got := int(summary["suggestion_aliases_copied_count"].(float64)); got != 1 {
			t.Fatalf("expected one copied suggestion alias, got %#v", summary)
		}
		if got := int(summary["suggestion_alias_duplicate_noop_count"].(float64)); got != 0 {
			t.Fatalf("expected zero duplicate alias no-ops, got %#v", summary)
		}
		if got := int(summary["provenance_only_retained_count"].(float64)); got != 0 {
			t.Fatalf("expected zero provenance-only retained identifiers, got %#v", summary)
		}
		exactMatchClasses := summary["exact_match_classes"].([]any)
		if len(exactMatchClasses) != 3 {
			t.Fatalf("expected three host exact-match classes, got %#v", exactMatchClasses)
		}
		if exactMatchClasses[0].(map[string]any)["identifier_class"] != "aad_device_id" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if exactMatchClasses[1].(map[string]any)["identifier_class"] != "fqdn" {
			t.Fatalf("unexpected host exact-match precedence: %#v", exactMatchClasses)
		}
		if got := int(exactMatchClasses[1].(map[string]any)["promoted_count"].(float64)); got != 1 {
			t.Fatalf("expected fqdn promoted_count=1, got %#v", exactMatchClasses[1])
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN := lookupHostState(t, harness.DB, golden.RecordCanonicalHostRecordID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorFQDN != "ws-023.corp.example.test" {
			t.Fatalf("unexpected survivor host state after merge: state=%s merged_into=%v row_version=%d fqdn=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupHostState(t, harness.DB, golden.RecordDuplicateHostRecordID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != golden.RecordCanonicalHostRecordID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser host state after merge: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		mention := lookupMention(t, harness.DB, golden.RecordHostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.RecordMentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention to survivor, got %#v", mention)
		}
		if mention.RowVersion != 2 {
			t.Fatalf("expected merge to increment mention row_version, got %#v", mention)
		}

		link := lookupActiveLink(t, harness.DB, incidentID, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.RecordTimelineRecordID, golden.RecordCanonicalHostRecordID, "observed_on_host", "manual", nil)
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.RecordDuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d active rows", got)
		}

		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incidentID, golden.RecordCanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active survivor tag after dedupe, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id = $1
   AND deleted_at IS NULL
`, golden.RecordDuplicateHostRecordID); got != 0 {
			t.Fatalf("expected loser active tags to be cleared, got %d", got)
		}
		if got := lookupAssessmentSubject(t, harness.DB, golden.RecordAssessmentHostID); got != golden.RecordCanonicalHostRecordID {
			t.Fatalf("expected loser assessment to repoint to survivor, got %s", got)
		}

		timelineChange := workbookscenariotest.RequireRecordChanged(t, timelineSocket, golden.RecordTimelineRecordID.String(), 1)
		if timelineChange.ChangeSetID != mergeData["change_set_id"] {
			t.Fatalf("expected websocket invalidation to carry the merge change_set_id, got timeline=%#v merge=%#v", timelineChange, mergeData)
		}
		survivorChange := workbookscenariotest.RequireRecordChanged(t, hostSocket, golden.RecordCanonicalHostRecordID.String(), 2)
		if len(survivorChange.AffectedViews) != 1 || survivorChange.AffectedViews[0].ChangeKind != "invalidate" {
			t.Fatalf("expected survivor invalidation, got %#v", survivorChange)
		}
		loserChange := workbookscenariotest.RequireRecordChanged(t, hostSocket, golden.RecordDuplicateHostRecordID.String(), 2)
		if len(loserChange.AffectedViews) != 1 || loserChange.AffectedViews[0].ChangeKind != "remove" {
			t.Fatalf("expected explicit loser removal, got %#v", loserChange)
		}
		viewLogin := workbookscenariotest.LoginResult{SessionCookie: adminLogin.sessionCookie, CSRFCookie: adminLogin.csrfCookie}
		survivorRow := workbookscenariotest.FindRow(t, workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordHostsViewSchemaID, viewLogin), golden.RecordCanonicalHostRecordID.String())
		if got := survivorRow["cells"].(map[string]any)["host.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected merge to project the repointed linked event on the survivor, got %#v row=%#v", got, survivorRow)
		}

		replayResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-merge",
				"reason":                    "  merge duplicate host  ",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		replayData := httptestx.RequireSuccessEnvelope(t, replayResp, http.StatusOK)["data"].(map[string]any)
		if replayData["change_set_id"] != mergeData["change_set_id"] {
			t.Fatalf("expected replayed merge to return the stored payload, got %#v %#v", mergeData, replayData)
		}

		divergentResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-merge",
				"reason":                    "different replay payload",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, divergentResp, http.StatusConflict, "client_txn_conflict")

		createResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordHostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-03-create-after-merge",
				"host.fqdn":     "ws-023.corp.example.test",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusOK)["data"].(map[string]any)
		row := createData["row"].(map[string]any)
		if row["record_id"] != golden.RecordCanonicalHostRecordID.String() {
			t.Fatalf("expected carried-forward exact match to reuse survivor, got %#v", createData)
		}
		_ = link
	})

	t.Run("merge conflict exposes both supplied and current versions", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-03-version-conflict")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-version-conflict-incident",
			"incident_key":  "IR-I403-V",
			"title":         "Entity merge version conflict",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "Survivor Host", "SURVIVOR", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "Loser Host", "LOSER", "", "")
		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE records SET row_version = CASE record_id WHEN $1 THEN 2 ELSE 3 END WHERE record_id IN ($1, $2)`, golden.RecordCanonicalHostRecordID, golden.RecordDuplicateHostRecordID); err != nil {
			t.Fatalf("advance merge fixture versions: %v", err)
		}

		response := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    2,
				"client_txn_id":             "txn-phase4-i-4-03-version-conflict",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		body := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "row_version_conflict")
		details := body["error"].(map[string]any)["details"].(map[string]any)
		expected := map[string]any{
			"survivor_record_id":           golden.RecordCanonicalHostRecordID.String(),
			"loser_record_id":              golden.RecordDuplicateHostRecordID.String(),
			"survivor_base_row_version":    float64(1),
			"loser_base_row_version":       float64(2),
			"survivor_current_row_version": float64(2),
			"loser_current_row_version":    float64(3),
		}
		if !reflect.DeepEqual(details, expected) {
			t.Fatalf("unexpected complete merge conflict details: got=%#v want=%#v", details, expected)
		}
	})

	t.Run("host merge collision uses owner precondition detail shape", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-03-collision-detail")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-collision-detail-incident",
			"incident_key":  "IR-I403-COLLISION",
			"title":         "Entity merge collision",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		blockingRecordID := uuid.New()

		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "Survivor Host", "SURVIVOR-HOST", "survivor-host.corp.example.test", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "Loser Host", "LOSER-HOST", "blocked-host.corp.example.test", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, blockingRecordID, "Blocking Host", "BLOCKING-HOST", "blocked-host.corp.example.test", "")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-collision-detail-merge",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		errBody := httptestx.RequireErrorEnvelope(t, mergeResp, http.StatusConflict, "merge_precondition_failed")
		details := errBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "carry_forward_identifier_collision" {
			t.Fatalf("unexpected merge collision reason_code: %#v", details)
		}
		if details["identifier_class"] != "fqdn" || details["normalized_value"] != "blocked-host.corp.example.test" {
			t.Fatalf("unexpected merge collision identifier details: %#v", details)
		}
		if details["blocking_record_id"] != blockingRecordID.String() {
			t.Fatalf("expected blocking_record_id=%s, got %#v", blockingRecordID, details)
		}
		if _, exists := details["conflicting_record_id"]; exists {
			t.Fatalf("merge collision details must not expose non-owner conflicting_record_id: %#v", details)
		}
	})

	t.Run("merge authorization re-derives current incident role", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-03-authz")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-authz-incident",
			"incident_key":  "IR-I403-A",
			"title":         "Entity merge authz",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "", "")

		if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
			t.Fatalf("demote merge actor membership: %v", err)
		}

		resp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateHostRecordID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-authz-merge",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusForbidden, "authorization_denied")
	})

	t.Run("identity merge preserves loser lineage, raw mention text, and current-state readback", func(t *testing.T) {
		harness := workbookscenariotest.StartServer(t, "phase4-i-4-03-identity")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-identity-incident",
			"incident_key":  "IR-I403-I",
			"title":         "Entity identity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		viewLogin := workbookscenariotest.LoginResult{SessionCookie: adminLogin.sessionCookie, CSRFCookie: adminLogin.csrfCookie}

		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.RecordCanonicalIdentityID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateIdentityID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		workbookscenariotest.SeedEntityAlias(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateIdentityID, "identity", "Case Owner")
		workbookscenariotest.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.RecordTimelineRecordID)
		workbookscenariotest.SeedResolvedMention(t, harness.DB, adminUserID, golden.RecordIdentityMentionID, golden.RecordTimelineRecordID, golden.RecordDuplicateIdentityID, golden.RecordFieldTimelineIdentityRefs, "identity", "Case Owner")
		workbookscenariotest.SeedRecordLink(t, harness.DB, incidentID, adminUserID, golden.RecordDuplicateLinkID, golden.RecordTimelineRecordID, golden.RecordDuplicateIdentityID, "observed_as_identity", "manual", nil)
		workbookscenariotest.SeedAssessment(t, harness.DB, incidentID, adminUserID, golden.RecordAssessmentIdentID, golden.RecordDuplicateIdentityID, "identity", "confirmed")
		beforeMention := lookupMention(t, harness.DB, golden.RecordIdentityMentionID)

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.RecordCanonicalIdentityID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.RecordDuplicateIdentityID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-identity-merge",
				"reason":                    "merge duplicate identity",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != golden.RecordCanonicalIdentityID.String() || mergeData["loser_record_id"] != golden.RecordDuplicateIdentityID.String() {
			t.Fatalf("unexpected identity merge payload: %#v", mergeData)
		}
		if mergeData["merge_summary"].(map[string]any)["record_type"] != "identity" {
			t.Fatalf("expected identity merge summary, got %#v", mergeData)
		}

		changeSet := asserttest.LookupChangeSet(t, asserttest.SQLDatabase(harness.DB), mergeData["change_set_id"].(string))
		auditassert.RequireMutationAttribution(t, auditassert.MutationAttribution{
			ActorUserID: changeSet.ActorUserID,
			Source:      changeSet.Source,
			ClientTxnID: changeSet.ClientTxnID,
			RequestID:   changeSet.RequestID,
			CreatedAt:   changeSet.CreatedAt,
		}, adminUserID.String(), "entities.records.merge", "txn-phase4-i-4-03-identity-merge")
		if got := asserttest.CountChangeSetMutations(t, asserttest.SQLDatabase(harness.DB), mergeData["change_set_id"].(string)); got < 2 {
			t.Fatalf("expected identity merge to emit at least two mutation rows, got %d", got)
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorEmail := lookupIdentityState(t, harness.DB, golden.RecordCanonicalIdentityID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorEmail != "alex.survivor@example.test" {
			t.Fatalf("unexpected survivor identity state: state=%s merged_into=%v row_version=%d email=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorEmail)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupIdentityState(t, harness.DB, golden.RecordDuplicateIdentityID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != golden.RecordCanonicalIdentityID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser identity state: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		afterMention := lookupMention(t, harness.DB, golden.RecordIdentityMentionID)
		assertx.RequireMentionStatus(t, afterMention, golden.RecordMentionStatusResolved)
		if afterMention.ResolvedRecordID == nil || *afterMention.ResolvedRecordID != golden.RecordCanonicalIdentityID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", afterMention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, afterMention.RawText)

		link := workbookscenariotest.LookupActiveLink(t, harness.DB, incidentID, golden.RecordTimelineRecordID, golden.RecordCanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(t, link, golden.RecordTimelineRecordID, golden.RecordCanonicalIdentityID, "observed_as_identity", "manual", nil)
		if got := workbookscenariotest.LookupAssessmentSubject(t, harness.DB, golden.RecordAssessmentIdentID); got != golden.RecordCanonicalIdentityID {
			t.Fatalf("expected identity assessment to repoint to survivor, got %s", got)
		}

		identityEnvelope := workbookscenariotest.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordIdentitiesViewSchemaID, viewLogin)
		contractassert.RequireDefaultQueryMeta(t, identityEnvelope, golden.RecordIdentitiesViewSchemaID)
		identityRows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.RecordIdentitiesViewSchemaID, viewLogin)
		identitySurvivorRow := workbookscenariotest.FindRow(t, identityRows, golden.RecordCanonicalIdentityID.String())
		if got := identitySurvivorRow["cells"].(map[string]any)["identity.linked_event_count"].(map[string]any)["value"]; got != float64(1) {
			t.Fatalf("expected identity merge to project the repointed linked event on the survivor, got %#v row=%#v", got, identitySurvivorRow)
		}
		for _, row := range identityRows {
			if row["record_id"] == golden.RecordDuplicateIdentityID.String() {
				t.Fatalf("expected merged loser to disappear from current-state identity rows, got %#v", identityRows)
			}
		}

		createAfterMerge := workbookscenariotest.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.RecordIdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-03-identity-after-merge",
				"identity.email":        "alex.analyst@example.test",
				"identity.display_name": "Alex After Merge",
			},
			workbookscenariotest.WithCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			workbookscenariotest.WithHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createAfterMergeData := workbookscenariotest.RequireSuccessData(t, createAfterMerge, http.StatusOK)
		if createAfterMergeData["row"].(map[string]any)["record_id"] != golden.RecordCanonicalIdentityID.String() {
			t.Fatalf("expected carried-forward identity exact match to reuse survivor, got %#v", createAfterMergeData)
		}
	})
}

func TestSupportPhase4Integration_EntityCreateIdempotencyIsActorScoped(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase4-entity-create-actor-scope")
	adminLogin, adminUserID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	editor := workbookscenariotest.SeedLocalUserFlags(t, harness.DB, "phase4-entity-scope-editor@example.test", "Entity Scope Editor", "EntityScopeEditor1!", false, false, true)
	incident := workbookscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-entity-scope-incident",
		"incident_key":  "IR-E-ACTOR-SCOPE",
		"title":         "Entity actor-scoped idempotency",
	})
	incidentID := workbookscenariotest.MustUUID(t, incident["incident_id"].(string))
	workbookscenariotest.SeedIncidentMembership(t, harness.DB, incidentID, editor.ID, editor.DisplayName, "editor", adminUserID)
	editorLogin := loginPhase4LocalUser(t, harness.Server, editor.Email, "EntityScopeEditor1!")

	cases := []struct {
		name       string
		routeKey   string
		viewSchema string
		payload    func(label string) map[string]any
	}{
		{
			name:       "hosts",
			routeKey:   "entities.hosts.rows.create",
			viewSchema: golden.RecordHostsViewSchemaID,
			payload: func(label string) map[string]any {
				return map[string]any{
					"client_txn_id":     "txn-phase4-shared-host-create",
					"host.display_name": "Actor scoped host " + label,
					"host.hostname":     "ACTOR-SCOPE-" + label,
				}
			},
		},
		{
			name:       "identities",
			routeKey:   "entities.identities.rows.create",
			viewSchema: golden.RecordIdentitiesViewSchemaID,
			payload: func(label string) map[string]any {
				return map[string]any{
					"client_txn_id":         "txn-phase4-shared-identity-create",
					"identity.display_name": "Actor Scoped " + label,
					"identity.email":        "actor-scope-" + label + "@example.test",
				}
			},
		},
		{
			name:       "indicators",
			routeKey:   "indicators.rows.create",
			viewSchema: golden.RecordIndicatorsViewSchemaID,
			payload: func(label string) map[string]any {
				value := "198.51.100.10"
				if label == "editor" {
					value = "198.51.100.11"
				}
				return map[string]any{
					"client_txn_id":              "txn-phase4-shared-indicator-create",
					"indicator.indicator_type":   golden.RecordIndicatorTypeIPv4,
					"indicator.value_kind":       golden.RecordIndicatorValueKindAtomic,
					"indicator.display_value":    value,
					"indicator.normalized_value": value,
					"indicator.defanged_value":   strings.ReplaceAll(value, ".", "[.]"),
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			adminPayload := tc.payload("admin")
			editorPayload := tc.payload("editor")
			adminCreate := createPhase4EntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, adminLogin, adminPayload, http.StatusCreated)
			editorCreate := createPhase4EntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, editorLogin, editorPayload, http.StatusCreated)
			adminRecordID := adminCreate["row"].(map[string]any)["record_id"].(string)
			editorRecordID := editorCreate["row"].(map[string]any)["record_id"].(string)
			if adminRecordID == editorRecordID {
				t.Fatalf("cross-actor %s create must not replay another actor's row, got %s", tc.name, adminRecordID)
			}

			adminReplay := createPhase4EntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, adminLogin, adminPayload, http.StatusOK)
			if adminReplay["change_set_id"] != adminCreate["change_set_id"] {
				t.Fatalf("admin %s replay returned wrong payload: got %#v want %#v", tc.name, adminReplay, adminCreate)
			}
			editorReplay := createPhase4EntityRow(t, harness.Server.HTTP.URL, incidentID.String(), tc.viewSchema, editorLogin, editorPayload, http.StatusOK)
			if editorReplay["change_set_id"] != editorCreate["change_set_id"] {
				t.Fatalf("editor %s replay returned wrong payload: got %#v want %#v", tc.name, editorReplay, editorCreate)
			}

			clientTxnID := adminPayload["client_txn_id"].(string)
			scopeKey := incidentID.String() + ":" + tc.viewSchema
			if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text IN ($2, $3)
   AND scope_key = $4
   AND client_txn_id = $5
`, tc.routeKey, adminUserID.String(), editor.ID.String(), scopeKey, clientTxnID); got != 2 {
				t.Fatalf("expected two actor-scoped %s idempotency rows, got %d", tc.name, got)
			}
			if got := workbookscenariotest.QueryCount(t, harness.DB, `
SELECT COUNT(DISTINCT actor_user_id)
  FROM route_idempotency
 WHERE route_key = $1
   AND scope_key = $2
   AND client_txn_id = $3
   AND actor_user_id::text IN ($4, $5)
`, tc.routeKey, scopeKey, clientTxnID, adminUserID.String(), editor.ID.String()); got != 2 {
				t.Fatalf("expected both actors represented for %s idempotency, got %d", tc.name, got)
			}
		})
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

func provisionBootstrapAdmin(t testing.TB, server *httptestx.Server) (loginResult, uuid.UUID) {
	t.Helper()

	bootstrapToken := requireBootstrapLogin(t, server, "bootstrap-admin@example.test", "BootstrapPass1!")
	begin := beginTOTPEnrollment(t, server, bootstrapToken, map[string]any{
		"client_txn_id": "txn-phase4-bootstrap-admin-begin",
	})
	secretBase32 := begin["totp_setup"].(map[string]any)["secret_base32"].(string)
	completeInitialEnrollment(t, server, bootstrapToken, begin["enrollment_id"].(string), secretBase32, "txn-phase4-bootstrap-admin-complete")
	login := loginLocalUserWithSecondFactor(t, server, "bootstrap-admin@example.test", "BootstrapPass1!", generateTOTPCode(t, secretBase32))

	sessionResp := doEntitiesJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/auth/session", nil, withCookies(login.sessionCookie))
	sessionData := httptestx.RequireSuccessEnvelope(t, sessionResp, http.StatusOK)["data"].(map[string]any)
	return login, mustUUID(t, sessionData["user_id"].(string))
}

func createIncident(t testing.TB, server *httptestx.Server, admin loginResult, body map[string]any) map[string]any {
	t.Helper()

	resp := doEntitiesJSON(
		t,
		http.MethodPost,
		server.HTTP.URL+"/api/v1/incidents",
		body,
		withCookies(admin.sessionCookie, admin.csrfCookie),
		withHeader(authn.CSRFHeaderName, admin.csrfCookie.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func createPhase4EntityRow(t testing.TB, serverURL string, incidentID string, viewSchemaID string, actor workbookscenariotest.LoginResult, body map[string]any, wantStatus int) map[string]any {
	t.Helper()

	resp := workbookscenariotest.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/rows",
		body,
		workbookscenariotest.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		workbookscenariotest.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return workbookscenariotest.RequireSuccessData(t, resp, wantStatus)
}

func loginPhase4LocalUser(t testing.TB, server *httptestx.Server, username string, password string) workbookscenariotest.LoginResult {
	t.Helper()

	resp := workbookscenariotest.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return workbookscenariotest.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
}

func requireBootstrapLogin(t testing.TB, server *httptestx.Server, username string, password string) string {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusUnauthorized, "mfa_setup_required")
	details := body["error"].(map[string]any)["details"].(map[string]any)
	return details["bootstrap_token"].(string)
}

func beginTOTPEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, body map[string]any) map[string]any {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/begin", body, withHeader("Authorization", "Bearer "+bootstrapToken))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func completeInitialEnrollment(t testing.TB, server *httptestx.Server, bootstrapToken string, enrollmentID string, secretBase32 string, clientTxnID string) {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/mfa/totp/complete", map[string]any{
		"client_txn_id": clientTxnID,
		"enrollment_id": enrollmentID,
		"code":          generateTOTPCode(t, secretBase32),
	}, withHeader("Authorization", "Bearer "+bootstrapToken))
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
}

func loginLocalUserWithSecondFactor(t testing.TB, server *httptestx.Server, username string, password string, code string) loginResult {
	t.Helper()

	resp := doEntitiesJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
		"second_factor": map[string]any{
			"kind": "totp",
			"assertion": map[string]any{
				"code": code,
			},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login with second factor failed: status=%d body=%#v", resp.StatusCode, httptestx.ReadJSONBody(t, resp))
	}
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)

	var sessionCookie *http.Cookie
	var csrfCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			sessionCookie = cookie
		case authn.CSRFCookieName:
			csrfCookie = cookie
		}
	}
	if sessionCookie == nil || csrfCookie == nil {
		t.Fatalf("expected login to set both session and csrf cookies, got %#v", resp.Cookies())
	}
	return loginResult{sessionCookie: sessionCookie, csrfCookie: csrfCookie}
}

func generateTOTPCode(t testing.TB, secretBase32 string) string {
	t.Helper()

	code, err := totp.GenerateCodeCustom(secretBase32, time.Now().UTC(), totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil {
		t.Fatalf("generate totp code: %v", err)
	}
	return code
}

func doEntitiesJSON(t testing.TB, method string, url string, body any, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req := httptestx.NewJSONRequest(t, method, url, body)
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func doEntitiesRawJSON(t testing.TB, method string, url string, body string, options ...func(*http.Request)) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create raw json request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(req)
	}
	return httptestx.Do(t, http.DefaultClient, req)
}

func withCookies(cookies ...*http.Cookie) func(*http.Request) {
	return func(req *http.Request) {
		for _, cookie := range cookies {
			if cookie != nil {
				req.AddCookie(cookie)
			}
		}
	}
}

func withHeader(key string, value string) func(*http.Request) {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func seedHostRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, displayName string, hostname string, fqdn string, aadDeviceID string) {
	t.Helper()
	workbookscenariotest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

	var (
		fqdnValue      any
		aadDeviceValue any
	)
	if fqdn != "" {
		fqdnValue = fqdn
	}
	if aadDeviceID != "" {
		aadDeviceValue = aadDeviceID
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO hosts (record_id, incident_id, display_name, hostname, fqdn, aad_device_id, host_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6, 'canonical', $7, $7)
`, recordID, incidentID, displayName, hostname, fqdnValue, aadDeviceValue, actorUserID); err != nil {
		t.Fatalf("seed host record: %v", err)
	}
}

func seedEntityAlias(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID, entityType string, rawText string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_aliases (incident_id, record_id, entity_type, raw_text, normalized_text, classification, created_by_user_id, created_at)
VALUES ($1, $2, $3, $4, $5, 'suggestion_only', $6, now())
`, incidentID, recordID, entityType, rawText, rawText, actorUserID); err != nil {
		t.Fatalf("seed entity alias: %v", err)
	}
}

func seedTimelineRecord(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	workbookscenariotest.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "timeline_event")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, capture_state, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'merge-source-row', 'reviewed', $3, $3)
`, recordID, incidentID, actorUserID); err != nil {
		t.Fatalf("seed timeline record: %v", err)
	}
}

func seedResolvedMention(t testing.TB, db *sql.DB, actorUserID uuid.UUID, mentionID uuid.UUID, sourceRecordID uuid.UUID, resolvedRecordID uuid.UUID, sourceFieldKey string, rawText string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO entity_mentions (
    entity_mention_id,
    source_record_id,
    entity_type,
    source_field_key,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    ordinal,
    created_by_user_id,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, 'host', $3, 'interactive_cell', 'merge-test', $4, $4, 'resolved', 1, 1, $5, $6, $5, now(), 'explicit_resolve_route')
`, mentionID, sourceRecordID, sourceFieldKey, rawText, actorUserID, resolvedRecordID); err != nil {
		t.Fatalf("seed resolved mention: %v", err)
	}
}

func seedRecordLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordLinkID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_links (
    record_link_id,
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, now(), now())
`, recordLinkID, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, actorUserID); err != nil {
		t.Fatalf("seed record link: %v", err)
	}
}

func seedRecordTag(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, recordTagID uuid.UUID, recordID uuid.UUID, tagName string) {
	t.Helper()

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO record_tags (record_tag_id, incident_id, record_id, tag_name, normalized_tag_name, created_by_user_id)
VALUES ($1, $2, $3, $4, $5, $6)
`, recordTagID, incidentID, recordID, tagName, tagName, actorUserID); err != nil {
		t.Fatalf("seed record tag: %v", err)
	}
}

func seedAssessment(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorUserID uuid.UUID, assessmentID uuid.UUID, subjectID uuid.UUID, subjectType string, state string) {
	t.Helper()
	workbookscenariotest.SeedRecordEnvelope(t, db, incidentID, actorUserID, assessmentID, "assessment")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO assessments (record_id, incident_id, subject_record_id, subject_type, assessment_state, rationale, assessor_user_id)
VALUES ($1, $2, $3, $4, $5, 'Seeded test assessment rationale.', $6)
`, assessmentID, incidentID, subjectID, subjectType, state, actorUserID); err != nil {
		t.Fatalf("seed assessment: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT a.record_id, a.incident_id, r.row_version, a.subject_record_id, a.subject_type, a.assessment_state, 'unset', a.rationale, a.assessor_user_id, a.assessed_at, 0
  FROM assessments a
  JOIN records r ON r.record_id = a.record_id
 WHERE a.record_id = $1
ON CONFLICT (record_id) DO NOTHING
`, assessmentID); err != nil {
		t.Fatalf("seed assessment projection: %v", err)
	}
}

func lookupHostState(t testing.TB, db *sql.DB, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		fqdn          sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT host_state, merged_into_record_id::text, row_version, COALESCE(fqdn, '')
  FROM hosts
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &fqdn); err != nil {
		t.Fatalf("lookup host state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := mustUUID(t, mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, fqdn.String
}

func lookupMention(t testing.TB, db *sql.DB, mentionID uuid.UUID) fixtures.EntityMentionFixture {
	t.Helper()

	var mention fixtures.EntityMentionFixture
	var (
		sourceRecordID   string
		resolvedRecordID sql.NullString
		resolvedByUserID sql.NullString
		resolvedAt       sql.NullTime
		resolutionMethod sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT source_record_id::text, raw_text, resolution_status, row_version, resolved_record_id::text, resolved_by_user_id::text, resolved_at, resolution_method
  FROM entity_mentions
 WHERE entity_mention_id = $1
`, mentionID).Scan(
		&sourceRecordID,
		&mention.RawText,
		&mention.ResolutionStatus,
		&mention.RowVersion,
		&resolvedRecordID,
		&resolvedByUserID,
		&resolvedAt,
		&resolutionMethod,
	); err != nil {
		t.Fatalf("lookup mention: %v", err)
	}
	mention.EntityMentionID = mentionID
	mention.SourceRecordID = mustUUID(t, sourceRecordID)
	if resolvedRecordID.Valid {
		value := mustUUID(t, resolvedRecordID.String)
		mention.ResolvedRecordID = &value
	}
	if resolvedByUserID.Valid {
		value := mustUUID(t, resolvedByUserID.String)
		mention.ResolvedByUserID = &value
	}
	if resolvedAt.Valid {
		value := resolvedAt.Time.UTC()
		mention.ResolvedAt = &value
	}
	if resolutionMethod.Valid {
		value := resolutionMethod.String
		mention.ResolutionMethod = &value
	}
	return mention
}

func lookupActiveLink(t testing.TB, db *sql.DB, incidentID uuid.UUID, sourceID uuid.UUID, targetID uuid.UUID, linkType string) fixtures.LinkFixture {
	t.Helper()

	var (
		link        fixtures.LinkFixture
		confidence  sql.NullInt64
		deletedAt   sql.NullTime
		recordLink  string
		incidentRaw string
		sourceRaw   string
		targetRaw   string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_link_id::text, incident_id::text, src_record_id::text, dst_record_id::text, link_type, provenance, confidence, deleted_at
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $4
   AND deleted_at IS NULL
`, incidentID, sourceID, targetID, linkType).Scan(&recordLink, &incidentRaw, &sourceRaw, &targetRaw, &link.LinkType, &link.Provenance, &confidence, &deletedAt); err != nil {
		t.Fatalf("lookup active link: %v", err)
	}
	link.RecordLinkID = mustUUID(t, recordLink)
	link.IncidentID = mustUUID(t, incidentRaw)
	link.SourceID = mustUUID(t, sourceRaw)
	link.TargetID = mustUUID(t, targetRaw)
	if confidence.Valid {
		value := int(confidence.Int64)
		link.Confidence = &value
	}
	if deletedAt.Valid {
		value := deletedAt.Time.UTC()
		link.DeletedAt = &value
	}
	return link
}

func lookupAssessmentSubject(t testing.TB, db *sql.DB, assessmentID uuid.UUID) uuid.UUID {
	t.Helper()

	var subjectID string
	if err := db.QueryRowContext(context.Background(), `
SELECT subject_record_id::text
  FROM assessments
 WHERE record_id = $1
`, assessmentID).Scan(&subjectID); err != nil {
		t.Fatalf("lookup assessment subject: %v", err)
	}
	return mustUUID(t, subjectID)
}

func requirePhase4ViewRowFieldSurface(t testing.TB, testID string, row map[string]any, viewSchemaID string) {
	t.Helper()

	contractassert.RequireFieldKeyConformance(
		t,
		workbookscenariotest.SortedRowFieldKeys(t, row),
		workbookscenariotest.AllowedFieldKeys(t, testID, viewSchemaID),
	)
}

type hostProjectionSnapshot struct {
	RecordID    uuid.UUID
	IncidentID  uuid.UUID
	RowVersion  int64
	DisplayName string
	Hostname    string
	HostState   string
	EditedAt    time.Time
}

type identityProjectionSnapshot struct {
	RecordID       uuid.UUID
	IncidentID     uuid.UUID
	RowVersion     int64
	DisplayName    string
	Email          *string
	SamAccountName *string
	IdentityState  string
	EditedAt       time.Time
}

func lookupHostProjectionSnapshot(t testing.TB, db *sql.DB, recordID uuid.UUID) hostProjectionSnapshot {
	t.Helper()

	var (
		snapshot      hostProjectionSnapshot
		recordIDRaw   string
		incidentIDRaw string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, hostname, host_state, edited_at
  FROM host_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &snapshot.Hostname, &snapshot.HostState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup host projection snapshot: %v", err)
	}
	snapshot.RecordID = mustUUID(t, recordIDRaw)
	snapshot.IncidentID = mustUUID(t, incidentIDRaw)
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

func lookupIdentityProjectionSnapshot(t testing.TB, db *sql.DB, recordID uuid.UUID) identityProjectionSnapshot {
	t.Helper()

	var (
		snapshot       identityProjectionSnapshot
		recordIDRaw    string
		incidentIDRaw  string
		email          *string
		samAccountName *string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT record_id::text, incident_id::text, row_version, display_name, email::text, sam_account_name, identity_state, edited_at
  FROM identity_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &snapshot.RowVersion, &snapshot.DisplayName, &email, &samAccountName, &snapshot.IdentityState, &snapshot.EditedAt); err != nil {
		t.Fatalf("lookup identity projection snapshot: %v", err)
	}
	snapshot.RecordID = mustUUID(t, recordIDRaw)
	snapshot.IncidentID = mustUUID(t, incidentIDRaw)
	snapshot.Email = email
	snapshot.SamAccountName = samAccountName
	snapshot.EditedAt = snapshot.EditedAt.UTC()
	return snapshot
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func lookupIdentityState(t testing.TB, db *sql.DB, recordID uuid.UUID) (string, *uuid.UUID, int64, string) {
	t.Helper()

	var (
		state         string
		mergedIntoRaw sql.NullString
		rowVersion    int64
		email         sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT identity_state, merged_into_record_id::text, row_version, email::text
  FROM identities
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedIntoRaw, &rowVersion, &email); err != nil {
		t.Fatalf("lookup identity state: %v", err)
	}
	var mergedInto *uuid.UUID
	if mergedIntoRaw.Valid {
		value := mustUUID(t, mergedIntoRaw.String)
		mergedInto = &value
	}
	return state, mergedInto, rowVersion, email.String
}

func mustUUID(t testing.TB, value string) uuid.UUID {
	t.Helper()

	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", value, err)
	}
	return parsed
}
