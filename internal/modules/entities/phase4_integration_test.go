package entities_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	entities "github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	platformws "github.com/JochiRaider/cartulary/internal/platform/ws"
	"github.com/JochiRaider/cartulary/internal/testutil/assertx"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

// I-4-01 / REQ-01-196..REQ-01-227, REQ-02-039..REQ-02-044 / AC-188..AC-190, AC-221..AC-225.
func TestPhase4_ResolveRoute_I_4_01(t *testing.T) {
	t.Run("resolve_item persists durable state, replays idempotently, and emits websocket invalidation", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-01-resolve")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-resolve-incident",
			"incident_key":  "IR-I401-R",
			"title":         "Phase 4 I-4-01 resolve route",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4test.SeedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

		socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.Phase4TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")

		payload := fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-resolve", golden.Phase4MentionActionResolve, uuidPointer(golden.Phase4CanonicalHostRecordID), nil)
		resp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			payload,
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := phase4test.RequireSuccessData(t, resp, http.StatusOK)
		if data["incident_id"] != incidentID.String() {
			t.Fatalf("unexpected incident_id in mention action payload: %#v", data)
		}
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after resolve_item, got %#v", data)
		}

		mention := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected resolve_item to point mention at the target record, got %#v", mention)
		}
		if mention.ResolutionMethod == nil || *mention.ResolutionMethod != "explicit_resolve_route" {
			t.Fatalf("expected resolve route provenance marker, got %#v", mention)
		}
		link := phase4test.LookupActiveLink(t, harness.DB, incidentID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)

		socketPayload := phase4test.RequireRecordChanged(t, socket, golden.Phase4TimelineRecordID.String(), 2)
		if socketPayload.ChangeSetID != data["change_set_id"] {
			t.Fatalf("expected websocket to carry the mention action change_set_id, got payload=%#v response=%#v", socketPayload, data)
		}

		replayResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			payload,
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		replayData := phase4test.RequireSuccessData(t, replayResp, http.StatusOK)
		if replayData["change_set_id"] != data["change_set_id"] {
			t.Fatalf("expected replay to reuse the original payload, got %#v %#v", data, replayData)
		}
		if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text = $2
   AND scope_key = $3
   AND client_txn_id = $4
`, "entities.entity_mentions.resolve", adminUserID.String(), golden.Phase4HostMentionID.String(), "txn-phase4-i-4-01-resolve"); got != 1 {
			t.Fatalf("expected one route idempotency row for a replayed mention action, got %d", got)
		}

		divergentResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-resolve", "dismiss_item", nil, nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		phase4test.RequireErrorBody(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("dismiss_item removes active links and fails closed on stale or illegal transitions", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-01-dismiss")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-dismiss-incident",
			"incident_key":  "IR-I401-D",
			"title":         "Phase 4 I-4-01 dismiss route",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4test.SeedResolvedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023")
		phase4test.SeedRecordLink(t, harness.DB, incidentID, adminUserID, golden.Phase4ManualLinkID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)

		socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.Phase4TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")

		resp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-dismiss", "dismiss_item", nil, nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := phase4test.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["entity_mention"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected mention row_version=2 after dismiss_item, got %#v", data)
		}

		dismissed := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, dismissed, golden.Phase4MentionStatusDismissed)
		if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'observed_on_host'
   AND deleted_at IS NULL
`, incidentID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID); got != 0 {
			t.Fatalf("expected dismiss_item to remove the active link, got %d active rows", got)
		}
		phase4test.RequireRecordChanged(t, socket, golden.Phase4TimelineRecordID.String(), 2)

		staleResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-stale", "revert_to_unresolved", nil, nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		staleBody := phase4test.RequireErrorBody(t, staleResp, http.StatusConflict, "row_version_conflict")
		if staleBody["error"].(map[string]any)["details"].(map[string]any)["current_mention_row_version"] != float64(2) {
			t.Fatalf("expected current_mention_row_version=2 on stale update, got %#v", staleBody)
		}

		illegalResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(2, "txn-phase4-i-4-01-illegal", "dismiss_item", nil, nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		illegalBody := phase4test.RequireErrorBody(t, illegalResp, http.StatusConflict, "illegal_transition")
		if illegalBody["error"].(map[string]any)["details"].(map[string]any)["from_status"] != "dismissed" {
			t.Fatalf("expected illegal transition details to report the dismissed source status, got %#v", illegalBody)
		}
	})

	t.Run("revert_to_unresolved restores unresolved mention state and emits websocket invalidation", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-01-revert")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-revert-incident",
			"incident_key":  "IR-I401-U",
			"title":         "Phase 4 I-4-01 revert route",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		phase4test.SeedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "dismissed", nil, nil)

		socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.Phase4TimelineViewSchemaID, adminLogin.SessionCookie.Value)
		defer socket.Close(1000, "test_complete")

		resp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-revert", "revert_to_unresolved", nil, nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		data := phase4test.RequireSuccessData(t, resp, http.StatusOK)
		if got := int64(data["source_record"].(map[string]any)["row_version"].(float64)); got != 2 {
			t.Fatalf("expected source record row_version=2 after revert_to_unresolved, got %#v", data)
		}

		mention := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
		if mention.RawText != "WS-023" || mention.ResolvedRecordID != nil || mention.ResolutionMethod != nil {
			t.Fatalf("expected revert_to_unresolved to preserve raw_text and clear resolution metadata, got %#v", mention)
		}
		phase4test.RequireRecordChanged(t, socket, golden.Phase4TimelineRecordID.String(), 2)
	})

	t.Run("authorization and target validation are re-derived from live state", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-01-access")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-access-incident",
			"incident_key":  "IR-I401-A",
			"title":         "Phase 4 I-4-01 access checks",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		phase4test.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		phase4test.SeedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "WS-023", "unresolved", nil, nil)

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
		deniedResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-denied", "resolve_item", uuidPointer(golden.Phase4CanonicalHostRecordID), nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		phase4test.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")

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

		otherActor := phase4test.SeedLocalUserFlags(t, harness.DB, "phase4-i401-other@example.test", "Phase4 I401 Other", "Phase4I401OtherPass1!", false, false, true)
		otherIncident := phase4test.CreateIncidentInStore(t, harness.Server.Runtime.Postgres, otherActor, "txn-phase4-i-4-01-hidden-incident", "IR-I401-H", "Phase 4 I-4-01 hidden")
		phase4test.SeedHostRecord(t, harness.DB, otherIncident.ID, otherActor.ID, golden.Phase4DuplicateHostRecordID, "Hidden WS-023", "HIDDEN-WS-023", "", "")

		hiddenResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-hidden", "resolve_item", uuidPointer(golden.Phase4DuplicateHostRecordID), nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		phase4test.RequireErrorBody(t, hiddenResp, http.StatusNotFound, "resolved_record_not_found")

		otherVisibleIncident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-01-visible-incident",
			"incident_key":  "IR-I401-V",
			"title":         "Phase 4 I-4-01 visible other incident",
		})
		otherVisibleIncidentID := phase4test.MustUUID(t, otherVisibleIncident["incident_id"].(string))
		phase4test.SeedHostRecord(t, harness.DB, otherVisibleIncidentID, adminUserID, golden.Phase4StubHostRecordID, "Visible WS-023", "VISIBLE-WS-023", "", "")

		crossIncidentResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-cross", "resolve_item", uuidPointer(golden.Phase4StubHostRecordID), nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		phase4test.RequireErrorBody(t, crossIncidentResp, http.StatusBadRequest, "invalid_mutation_payload")

		wrongTypeResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/entity-mentions/"+golden.Phase4HostMentionID.String()+"/resolve",
			fixtures.MentionResolveRoutePayload(1, "txn-phase4-i-4-01-wrong-type", "resolve_item", uuidPointer(golden.Phase4CanonicalIdentityID), nil),
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		phase4test.RequireErrorBody(t, wrongTypeResp, http.StatusBadRequest, "invalid_mutation_payload")

		mention := phase4test.LookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusUnresolved)
		assertx.RequireRowVersionStable(t, 1, mention.RowVersion)
		if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND deleted_at IS NULL
`, incidentID, golden.Phase4TimelineRecordID); got != 0 {
			t.Fatalf("expected invalid or unauthorized target attempts to leave record links untouched, got %d active rows", got)
		}
	})
}

// I-4-02 / REQ-02-035..REQ-02-036, REQ-02-054..REQ-02-055, REQ-02-059..REQ-02-063 / AC-022, AC-186.
func TestPhase4_EntityOriginUpsert_I_4_02(t *testing.T) {
	t.Run("host create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-02-host")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-host-incident",
			"incident_key":  "IR-I402-H",
			"title":         "Phase 4 I-4-02 host create",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

		hostCreate := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-create",
				"host.display_name": "Gateway record",
				"host.hostname":     "GATEWAY-01",
				"host.fqdn":         "gateway-01.corp.example",
				"host.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_token", "raw_text": "VPN Gateway"},
					},
				},
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostCreateData := phase4test.RequireSuccessData(t, hostCreate, http.StatusCreated)
		hostRow := hostCreateData["row"].(map[string]any)
		hostRecordID := phase4test.MustUUID(t, hostRow["record_id"].(string))

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
		if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("entity-origin host create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE hosts SET fqdn = NULL WHERE record_id = $1`, hostRecordID); err != nil {
			t.Fatalf("clear host fqdn to force preserved-identifier reuse: %v", err)
		}
		hostReuse := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-reuse",
				"host.display_name": "Gateway reused",
				"host.fqdn":         "gateway-01.corp.example",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReuseData := phase4test.RequireSuccessData(t, hostReuse, http.StatusOK)
		if got := phase4test.MustUUID(t, hostReuseData["row"].(map[string]any)["record_id"].(string)); got != hostRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original host record, got %#v", hostReuseData)
		}
		if state, mergedInto, rowVersion, restoredFQDN := phase4test.LookupHostState(t, harness.DB, hostRecordID); state != "stub" || mergedInto != nil || rowVersion != 2 || restoredFQDN != "gateway-01.corp.example" {
			t.Fatalf("unexpected reused host state: state=%s merged_into=%v row_version=%d fqdn=%q", state, mergedInto, rowVersion, restoredFQDN)
		}

		aliasOnly := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-alias-only",
				"host.display_name": "VPN Gateway",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := phase4test.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := phase4test.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == hostRecordID {
			t.Fatalf("expected alias-only create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "Conflict Host A", "COLLISION-01", "", "")
		phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "Conflict Host B", "COLLISION-01", "", "")
		conflictResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-host-conflict",
				"host.display_name": "Conflict Host",
				"host.hostname":     "COLLISION-01",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := phase4test.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "host" || details["identifier_class"] != "hostname" {
			t.Fatalf("unexpected host conflict details: %#v", details)
		}
		candidateIDs := details["candidate_record_ids"].([]any)
		if len(candidateIDs) != 2 {
			t.Fatalf("expected two conflict candidates, got %#v", details)
		}
		if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, hostRecordID); got != 0 {
			t.Fatalf("conflicted host create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("identity create covers direct create, preserved exact-match reuse, alias-only non-reuse, and conflict handling", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-02-identity")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-identity-incident",
			"incident_key":  "IR-I402-I",
			"title":         "Phase 4 I-4-02 identity create",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

		identityCreate := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":             "txn-phase4-i-4-02-identity-create",
				"identity.display_name":     "Alex Analyst",
				"identity.email":            "alex.analyst@example.test",
				"identity.sam_account_name": "ALEXA",
				"identity.aliases": map[string]any{
					"kind": "collection_actions_v1",
					"actions": []map[string]any{
						{"op": "add_token", "raw_text": "Case Owner"},
					},
				},
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityCreateData := phase4test.RequireSuccessData(t, identityCreate, http.StatusCreated)
		identityRecordID := phase4test.MustUUID(t, identityCreateData["row"].(map[string]any)["record_id"].(string))

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
		if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("entity-origin identity create must not synthesize mentions, got %d rows", got)
		}

		if _, err := harness.DB.ExecContext(context.Background(), `UPDATE identities SET email = NULL WHERE record_id = $1`, identityRecordID); err != nil {
			t.Fatalf("clear identity email to force preserved-identifier reuse: %v", err)
		}
		identityReuse := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-reuse",
				"identity.display_name": "Alex Analyst Reused",
				"identity.email":        "alex.analyst@example.test",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityReuseData := phase4test.RequireSuccessData(t, identityReuse, http.StatusOK)
		if got := phase4test.MustUUID(t, identityReuseData["row"].(map[string]any)["record_id"].(string)); got != identityRecordID {
			t.Fatalf("expected preserved exact-match reuse to return the original identity record, got %#v", identityReuseData)
		}
		if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM identities
 WHERE incident_id = $1
   AND record_id = $2
   AND row_version = 2
   AND email = 'alex.analyst@example.test'
`, incidentID, identityRecordID); got != 1 {
			t.Fatalf("expected preserved-identifier reuse to restore the canonical email and increment row_version, got %d rows", got)
		}

		aliasOnly := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-alias-only",
				"identity.display_name": "Case Owner",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		aliasOnlyData := phase4test.RequireSuccessData(t, aliasOnly, http.StatusCreated)
		if got := phase4test.MustUUID(t, aliasOnlyData["row"].(map[string]any)["record_id"].(string)); got == identityRecordID {
			t.Fatalf("expected alias-only identity create to remain suggestion-only, got %#v", aliasOnlyData)
		}

		phase4test.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalIdentityID, "Conflict Identity A", "collision@example.test", "collision@example.test", "COLLISION-A")
		phase4test.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateIdentityID, "Conflict Identity B", "collision@example.test", "collision@example.test", "COLLISION-B")
		conflictResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-02-identity-conflict",
				"identity.display_name": "Conflict Identity",
				"identity.email":        "collision@example.test",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		conflictBody := phase4test.RequireErrorBody(t, conflictResp, http.StatusConflict, "entity_match_conflict")
		details := conflictBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "merge_required" || details["entity_type"] != "identity" || details["identifier_class"] != "email" {
			t.Fatalf("unexpected identity conflict details: %#v", details)
		}
		if got := len(details["candidate_record_ids"].([]any)); got != 2 {
			t.Fatalf("expected two identity conflict candidates, got %#v", details)
		}
		if got := phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM entity_mentions WHERE source_record_id = $1`, identityRecordID); got != 0 {
			t.Fatalf("conflicted identity create must not synthesize mentions, got %d rows", got)
		}
	})

	t.Run("create routes emit history, round-trip current-state query reads, and re-derive live authorization", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-02-query-auth")
		adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
		incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-02-query-auth-incident",
			"incident_key":  "IR-I402-Q",
			"title":         "Phase 4 I-4-02 query and auth",
		})
		incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
		viewLogin := phase4test.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}

		hostPayload := map[string]any{
			"client_txn_id":     "txn-phase4-i-4-02-query-host",
			"host.display_name": "Gateway query host",
			"host.hostname":     "GATEWAY-Q-01",
			"host.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_token", "raw_text": "Gateway Query Alias"},
				},
			},
		}
		hostResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			hostPayload,
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostData := phase4test.RequireSuccessData(t, hostResp, http.StatusCreated)
		hostRecordID := phase4test.MustUUID(t, hostData["row"].(map[string]any)["record_id"].(string))
		hostChangeSet := timelinetest.LookupChangeSet(t, harness.DB, hostData["change_set_id"].(string))
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: hostChangeSet.ActorUserID,
			Source:      hostChangeSet.Source,
			ClientTxnID: hostChangeSet.ClientTxnID,
			RequestID:   hostChangeSet.RequestID,
			CreatedAt:   hostChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.hosts.rows.create", "txn-phase4-i-4-02-query-host")
		if got := timelinetest.CountChangeSetMutations(t, harness.DB, hostData["change_set_id"].(string)); got != 1 {
			t.Fatalf("expected one host create mutation row, got %d", got)
		}

		identityPayload := map[string]any{
			"client_txn_id":             "txn-phase4-i-4-02-query-identity",
			"identity.display_name":     "Alex Query",
			"identity.email":            "alex.query@example.test",
			"identity.sam_account_name": "ALEXQ",
			"identity.aliases": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{
					{"op": "add_token", "raw_text": "Query Owner"},
				},
			},
		}
		identityResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			identityPayload,
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		identityData := phase4test.RequireSuccessData(t, identityResp, http.StatusCreated)
		identityRecordID := phase4test.MustUUID(t, identityData["row"].(map[string]any)["record_id"].(string))
		identityChangeSet := timelinetest.LookupChangeSet(t, harness.DB, identityData["change_set_id"].(string))
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: identityChangeSet.ActorUserID,
			Source:      identityChangeSet.Source,
			ClientTxnID: identityChangeSet.ClientTxnID,
			RequestID:   identityChangeSet.RequestID,
			CreatedAt:   identityChangeSet.CreatedAt,
		}, adminUserID.String(), "entities.identities.rows.create", "txn-phase4-i-4-02-query-identity")
		if got := timelinetest.CountChangeSetMutations(t, harness.DB, identityData["change_set_id"].(string)); got != 1 {
			t.Fatalf("expected one identity create mutation row, got %d", got)
		}

		hostEnvelope := phase4test.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4HostsViewSchemaID, viewLogin)
		httptestx.RequireDefaultQueryMeta(t, hostEnvelope, golden.Phase4HostsViewSchemaID)
		hostRow := phase4test.FindRow(t, phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4HostsViewSchemaID, viewLogin), hostRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", hostRow, golden.Phase4HostsViewSchemaID)
		hostAlias := phase4test.RequireSingleCollectionItem(t, hostRow, "host.aliases")
		if hostAlias["item_kind"] != "suggestion_only_alias" || hostAlias["raw_text"] != "Gateway Query Alias" {
			t.Fatalf("unexpected host alias readback: %#v", hostAlias)
		}

		identityEnvelope := phase4test.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IdentitiesViewSchemaID, viewLogin)
		httptestx.RequireDefaultQueryMeta(t, identityEnvelope, golden.Phase4IdentitiesViewSchemaID)
		identityRow := phase4test.FindRow(t, phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", identityRow, golden.Phase4IdentitiesViewSchemaID)
		identityAlias := phase4test.RequireSingleCollectionItem(t, identityRow, "identity.aliases")
		if identityAlias["item_kind"] != "suggestion_only_alias" || identityAlias["raw_text"] != "Query Owner" {
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
		httptestx.RequireProjectionDeterminism(t, hostProjectionBefore, hostProjectionAfter)
		httptestx.RequireProjectionDeterminism(t, identityProjectionBefore, identityProjectionAfter)

		hostRowAfterRebuild := phase4test.FindRow(t, phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4HostsViewSchemaID, viewLogin), hostRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", hostRowAfterRebuild, golden.Phase4HostsViewSchemaID)
		httptestx.RequireProjectionDeterminism(t, hostRow["cells"], hostRowAfterRebuild["cells"])
		if rebuiltHostAlias := phase4test.RequireSingleCollectionItem(t, hostRowAfterRebuild, "host.aliases"); rebuiltHostAlias["raw_text"] != "Gateway Query Alias" {
			t.Fatalf("unexpected rebuilt host alias readback: %#v", rebuiltHostAlias)
		}

		identityRowAfterRebuild := phase4test.FindRow(t, phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IdentitiesViewSchemaID, viewLogin), identityRecordID.String())
		requirePhase4ViewRowFieldSurface(t, "I-4-02", identityRowAfterRebuild, golden.Phase4IdentitiesViewSchemaID)
		httptestx.RequireProjectionDeterminism(t, identityRow["cells"], identityRowAfterRebuild["cells"])
		if rebuiltIdentityAlias := phase4test.RequireSingleCollectionItem(t, identityRowAfterRebuild, "identity.aliases"); rebuiltIdentityAlias["raw_text"] != "Query Owner" {
			t.Fatalf("unexpected rebuilt identity alias readback: %#v", rebuiltIdentityAlias)
		}

		replayStableBefore := httptestx.ReplayCounts{
			ChangeSets: phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		}
		hostReplay := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			hostPayload,
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostReplayData := phase4test.RequireSuccessData(t, hostReplay, http.StatusOK)
		if hostReplayData["change_set_id"] != hostData["change_set_id"] {
			t.Fatalf("expected host replay to reuse the original payload, got %#v %#v", hostData, hostReplayData)
		}
		httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
			FirstStatus:     http.StatusCreated,
			ReplayStatus:    http.StatusOK,
			DivergentStatus: http.StatusConflict,
			DivergentCode:   "client_txn_conflict",
			StableBefore:    replayStableBefore,
			StableAfter: httptestx.ReplayCounts{
				ChangeSets: phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
				MutationRows: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
			},
		})

		hostDivergent := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-query-host",
				"host.display_name": "Gateway query host divergent",
				"host.hostname":     "GATEWAY-Q-01",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		hostDivergentBody := phase4test.RequireErrorBody(t, hostDivergent, http.StatusConflict, "client_txn_conflict")
		httptestx.RequireDivergentReplayRejected(t, hostDivergent.StatusCode, hostDivergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

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
		deniedResp := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":     "txn-phase4-i-4-02-query-host-denied",
				"host.display_name": "Denied host",
				"host.hostname":     "DENIED-HOST",
			},
			phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		)
		deniedBody := phase4test.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")
		httptestx.RequireAuthorizationReDerived(
			t,
			httptestx.AuthorizationOutcome{Status: http.StatusCreated},
			httptestx.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
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
	harness := phase4test.StartServer(t, "phase4-entity-create-auth-csrf-order")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-entity-create-auth-order-incident",
		"incident_key":  "IR-AUTH-CSRF-ORDER",
		"title":         "Entity create auth csrf ordering",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.Phase4HostsViewSchemaID, adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	type entityCreateFailureCounts struct {
		Records        int
		ChangeSets     int
		MutationRows   int
		HostProjection int
	}
	counts := func() entityCreateFailureCounts {
		return entityCreateFailureCounts{
			Records:        phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM records WHERE incident_id = $1`, incidentID),
			ChangeSets:     phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows:   phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations m JOIN change_sets c ON c.change_set_id = m.change_set_id WHERE c.incident_id = $1`, incidentID),
			HostProjection: phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE incident_id = $1`, incidentID),
		}
	}
	before := counts()
	url := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + golden.Phase4HostsViewSchemaID + "/rows"

	unauthenticated := doEntitiesRawJSON(t, http.MethodPost, url, "{")
	phase4test.RequireErrorBody(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	phase4test.RequireErrorBody(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := doEntitiesRawJSON(
		t,
		http.MethodPost,
		url,
		"{",
		withCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		withHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	phase4test.RequireErrorBody(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	if after := counts(); after != before {
		t.Fatalf("auth/csrf failures must not mutate entity state: before=%#v after=%#v", before, after)
	}
	phase4test.ExpectNoSocketMessage(t, socket)
}

// I-4-03 / REQ-01-181..REQ-01-195, REQ-02-064..REQ-02-066 / AC-023, AC-186, AC-209.
func TestPhase4_ExplicitMergeRoute_I_4_03(t *testing.T) {
	t.Run("host merge repoints live fan-out and preserves survivor reuse", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-03")
		phase4test.RequireSchemaTables(t, harness.DB, "I-4-03", "hosts", "identities", "entity_mentions", "record_tags", "assessments")

		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-incident",
			"incident_key":  "IR-I403",
			"title":         "Entity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		timelineSocket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), golden.Phase4TimelineViewSchemaID, adminLogin.sessionCookie.Value)
		defer timelineSocket.Close(1000, "test_complete")

		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "ws-023.corp.example.test", "")
		seedEntityAlias(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "host", "Workstation 23")
		seedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		seedResolvedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, golden.Phase4FieldTimelineHostRefs, "WS-023")
		seedRecordLink(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateHostRecordID, "observed_on_host", "manual", nil)
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.Phase4TagIDSurvivor, golden.Phase4CanonicalHostRecordID, "critical-host")
		seedRecordTag(t, harness.DB, incidentID, adminUserID, golden.Phase4TagIDLoser, golden.Phase4DuplicateHostRecordID, "critical-host")
		seedAssessment(t, harness.DB, incidentID, adminUserID, golden.Phase4AssessmentHostID, golden.Phase4DuplicateHostRecordID, "host", "confirmed")

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateHostRecordID.String(),
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
		if mergeData["survivor_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("unexpected survivor_record_id: %#v", mergeData)
		}
		if mergeData["loser_record_id"] != golden.Phase4DuplicateHostRecordID.String() {
			t.Fatalf("unexpected loser_record_id: %#v", mergeData)
		}
		if got := int64(mergeData["survivor_row_version"].(float64)); got != 2 {
			t.Fatalf("expected survivor_row_version=2, got %d", got)
		}
		if got := int64(mergeData["loser_row_version"].(float64)); got != 2 {
			t.Fatalf("expected loser_row_version=2, got %d", got)
		}
		if mergeData["merged_into_record_id"] != golden.Phase4CanonicalHostRecordID.String() {
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

		survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN := lookupHostState(t, harness.DB, golden.Phase4CanonicalHostRecordID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorFQDN != "ws-023.corp.example.test" {
			t.Fatalf("unexpected survivor host state after merge: state=%s merged_into=%v row_version=%d fqdn=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorFQDN)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupHostState(t, harness.DB, golden.Phase4DuplicateHostRecordID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != golden.Phase4CanonicalHostRecordID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser host state after merge: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		mention := lookupMention(t, harness.DB, golden.Phase4HostMentionID)
		assertx.RequireMentionStatus(t, mention, golden.Phase4MentionStatusResolved)
		if mention.ResolvedRecordID == nil || *mention.ResolvedRecordID != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected merge to repoint mention to survivor, got %#v", mention)
		}
		if mention.RowVersion != 2 {
			t.Fatalf("expected merge to increment mention row_version, got %#v", mention)
		}

		link := lookupActiveLink(t, harness.DB, incidentID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalHostRecordID, "observed_on_host", "manual", nil)
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateLinkID); got != 0 {
			t.Fatalf("expected loser-targeted active link to disappear, got %d active rows", got)
		}

		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND normalized_tag_name = 'critical-host'
   AND deleted_at IS NULL
`, incidentID, golden.Phase4CanonicalHostRecordID); got != 1 {
			t.Fatalf("expected one active survivor tag after dedupe, got %d", got)
		}
		if got := queryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_tags
 WHERE record_id = $1
   AND deleted_at IS NULL
`, golden.Phase4DuplicateHostRecordID); got != 0 {
			t.Fatalf("expected loser active tags to be cleared, got %d", got)
		}
		if got := lookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentHostID); got != golden.Phase4CanonicalHostRecordID {
			t.Fatalf("expected loser assessment to repoint to survivor, got %s", got)
		}

		timelineChange := phase4test.RequireRecordChanged(t, timelineSocket, golden.Phase4TimelineRecordID.String(), 1)
		if timelineChange.ChangeSetID != mergeData["change_set_id"] {
			t.Fatalf("expected websocket invalidation to carry the merge change_set_id, got timeline=%#v merge=%#v", timelineChange, mergeData)
		}

		replayResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateHostRecordID.String(),
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
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateHostRecordID.String(),
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
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4HostsViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id": "txn-phase4-i-4-03-create-after-merge",
				"host.fqdn":     "ws-023.corp.example.test",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createData := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusOK)["data"].(map[string]any)
		row := createData["row"].(map[string]any)
		if row["record_id"] != golden.Phase4CanonicalHostRecordID.String() {
			t.Fatalf("expected carried-forward exact match to reuse survivor, got %#v", createData)
		}
		_ = link
	})

	t.Run("merge authorization re-derives current incident role", func(t *testing.T) {
		harness := phase4test.StartServer(t, "phase4-i-4-03-authz")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-authz-incident",
			"incident_key":  "IR-I403-A",
			"title":         "Entity merge authz",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "WS-023", "WS-023", "", "")
		seedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "WS-023 duplicate", "WS-023-DUP", "", "")

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
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalHostRecordID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateHostRecordID.String(),
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
		harness := phase4test.StartServer(t, "phase4-i-4-03-identity")
		adminLogin, adminUserID := provisionBootstrapAdmin(t, harness.Server)
		incident := createIncident(t, harness.Server, adminLogin, map[string]any{
			"client_txn_id": "txn-phase4-i-4-03-identity-incident",
			"incident_key":  "IR-I403-I",
			"title":         "Entity identity merge",
		})
		incidentID := mustUUID(t, incident["incident_id"].(string))
		viewLogin := phase4test.LoginResult{SessionCookie: adminLogin.sessionCookie, CSRFCookie: adminLogin.csrfCookie}

		phase4test.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalIdentityID, "Alex Analyst", "alex.survivor@example.test", "alex.survivor@example.test", "ALEXSURV")
		phase4test.SeedIdentityRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateIdentityID, "Alex Duplicate", "alex.analyst@example.test", "alex.analyst@example.test", "ALEXA")
		phase4test.SeedEntityAlias(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateIdentityID, "identity", "Case Owner")
		phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
		phase4test.SeedResolvedMention(t, harness.DB, adminUserID, golden.Phase4IdentityMentionID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateIdentityID, golden.Phase4FieldTimelineIdentityRefs, "identity", "Case Owner")
		phase4test.SeedRecordLink(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateLinkID, golden.Phase4TimelineRecordID, golden.Phase4DuplicateIdentityID, "observed_as_identity", "manual", nil)
		phase4test.SeedAssessment(t, harness.DB, incidentID, adminUserID, golden.Phase4AssessmentIdentID, golden.Phase4DuplicateIdentityID, "identity", "confirmed")
		beforeMention := lookupMention(t, harness.DB, golden.Phase4IdentityMentionID)

		mergeResp := doEntitiesJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+golden.Phase4CanonicalIdentityID.String()+"/merge",
			map[string]any{
				"loser_record_id":           golden.Phase4DuplicateIdentityID.String(),
				"survivor_base_row_version": 1,
				"loser_base_row_version":    1,
				"client_txn_id":             "txn-phase4-i-4-03-identity-merge",
				"reason":                    "merge duplicate identity",
			},
			withCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			withHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		mergeData := httptestx.RequireSuccessEnvelope(t, mergeResp, http.StatusOK)["data"].(map[string]any)
		if mergeData["survivor_record_id"] != golden.Phase4CanonicalIdentityID.String() || mergeData["loser_record_id"] != golden.Phase4DuplicateIdentityID.String() {
			t.Fatalf("unexpected identity merge payload: %#v", mergeData)
		}
		if mergeData["merge_summary"].(map[string]any)["record_type"] != "identity" {
			t.Fatalf("expected identity merge summary, got %#v", mergeData)
		}

		changeSet := timelinetest.LookupChangeSet(t, harness.DB, mergeData["change_set_id"].(string))
		httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
			ActorUserID: changeSet.ActorUserID,
			Source:      changeSet.Source,
			ClientTxnID: changeSet.ClientTxnID,
			RequestID:   changeSet.RequestID,
			CreatedAt:   changeSet.CreatedAt,
		}, adminUserID.String(), "entities.records.merge", "txn-phase4-i-4-03-identity-merge")
		if got := timelinetest.CountChangeSetMutations(t, harness.DB, mergeData["change_set_id"].(string)); got < 2 {
			t.Fatalf("expected identity merge to emit at least two mutation rows, got %d", got)
		}

		survivorState, survivorMergedInto, survivorRowVersion, survivorEmail := lookupIdentityState(t, harness.DB, golden.Phase4CanonicalIdentityID)
		if survivorState != "canonical" || survivorMergedInto != nil || survivorRowVersion != 2 || survivorEmail != "alex.survivor@example.test" {
			t.Fatalf("unexpected survivor identity state: state=%s merged_into=%v row_version=%d email=%q", survivorState, survivorMergedInto, survivorRowVersion, survivorEmail)
		}
		loserState, loserMergedInto, loserRowVersion, _ := lookupIdentityState(t, harness.DB, golden.Phase4DuplicateIdentityID)
		if loserState != "merged" || loserMergedInto == nil || *loserMergedInto != golden.Phase4CanonicalIdentityID || loserRowVersion != 2 {
			t.Fatalf("unexpected loser identity state: state=%s merged_into=%v row_version=%d", loserState, loserMergedInto, loserRowVersion)
		}

		afterMention := lookupMention(t, harness.DB, golden.Phase4IdentityMentionID)
		assertx.RequireMentionStatus(t, afterMention, golden.Phase4MentionStatusResolved)
		if afterMention.ResolvedRecordID == nil || *afterMention.ResolvedRecordID != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected identity merge to repoint mention resolution to survivor, got %#v", afterMention)
		}
		assertx.RequireRawTextPreserved(t, beforeMention.RawText, afterMention.RawText)

		link := phase4test.LookupActiveLink(t, harness.DB, incidentID, golden.Phase4TimelineRecordID, golden.Phase4CanonicalIdentityID, "observed_as_identity")
		assertx.RequireActiveLink(t, link, golden.Phase4TimelineRecordID, golden.Phase4CanonicalIdentityID, "observed_as_identity", "manual", nil)
		if got := phase4test.LookupAssessmentSubject(t, harness.DB, golden.Phase4AssessmentIdentID); got != golden.Phase4CanonicalIdentityID {
			t.Fatalf("expected identity assessment to repoint to survivor, got %s", got)
		}

		identityEnvelope := phase4test.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IdentitiesViewSchemaID, viewLogin)
		httptestx.RequireDefaultQueryMeta(t, identityEnvelope, golden.Phase4IdentitiesViewSchemaID)
		identityRows := phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IdentitiesViewSchemaID, viewLogin)
		phase4test.FindRow(t, identityRows, golden.Phase4CanonicalIdentityID.String())
		for _, row := range identityRows {
			if row["record_id"] == golden.Phase4DuplicateIdentityID.String() {
				t.Fatalf("expected merged loser to disappear from current-state identity rows, got %#v", identityRows)
			}
		}

		createAfterMerge := phase4test.DoJSON(
			t,
			http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IdentitiesViewSchemaID+"/rows",
			map[string]any{
				"client_txn_id":         "txn-phase4-i-4-03-identity-after-merge",
				"identity.email":        "alex.analyst@example.test",
				"identity.display_name": "Alex After Merge",
			},
			phase4test.WithCookies(adminLogin.sessionCookie, adminLogin.csrfCookie),
			phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.csrfCookie.Value),
		)
		createAfterMergeData := phase4test.RequireSuccessData(t, createAfterMerge, http.StatusOK)
		if createAfterMergeData["row"].(map[string]any)["record_id"] != golden.Phase4CanonicalIdentityID.String() {
			t.Fatalf("expected carried-forward identity exact match to reuse survivor, got %#v", createAfterMergeData)
		}
	})
}

func TestSupportPhase4Integration_EntityCreateIdempotencyIsActorScoped(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-entity-create-actor-scope")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	editor := phase4test.SeedLocalUserFlags(t, harness.DB, "phase4-entity-scope-editor@example.test", "Entity Scope Editor", "EntityScopeEditor1!", false, false, true)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-entity-scope-incident",
		"incident_key":  "IR-E-ACTOR-SCOPE",
		"title":         "Entity actor-scoped idempotency",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	phase4test.SeedIncidentMembership(t, harness.DB, incidentID, editor.ID, editor.DisplayName, "editor", adminUserID)
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
			viewSchema: golden.Phase4HostsViewSchemaID,
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
			viewSchema: golden.Phase4IdentitiesViewSchemaID,
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
			routeKey:   "entities.indicators.rows.create",
			viewSchema: golden.Phase4IndicatorsViewSchemaID,
			payload: func(label string) map[string]any {
				value := "198.51.100.10"
				if label == "editor" {
					value = "198.51.100.11"
				}
				return map[string]any{
					"client_txn_id":              "txn-phase4-shared-indicator-create",
					"indicator.indicator_type":   golden.Phase4IndicatorTypeIPv4,
					"indicator.value_kind":       golden.Phase4IndicatorValueKindAtomic,
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
			if got := phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = $1
   AND actor_user_id::text IN ($2, $3)
   AND scope_key = $4
   AND client_txn_id = $5
`, tc.routeKey, adminUserID.String(), editor.ID.String(), scopeKey, clientTxnID); got != 2 {
				t.Fatalf("expected two actor-scoped %s idempotency rows, got %d", tc.name, got)
			}
			if got := phase4test.QueryCount(t, harness.DB, `
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

// I-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestPhase4_IndicatorsRoute_I_4_07(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-i-4-07-indicators")
	store := entities.NewStore(harness.Server.Runtime.Postgres)
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-i-4-07-incident",
		"incident_key":  "IR-I407",
		"title":         "Phase 4 I-4-07 indicators route",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))
	viewLogin := phase4test.LoginResult{SessionCookie: adminLogin.SessionCookie, CSRFCookie: adminLogin.CSRFCookie}

	createPayload := map[string]any{
		"client_txn_id":              "txn-phase4-i-4-07-create",
		"indicator.indicator_type":   golden.Phase4IndicatorExamples[0].IndicatorType,
		"indicator.value_kind":       golden.Phase4IndicatorExamples[0].ValueKind,
		"indicator.display_value":    golden.Phase4IndicatorExamples[0].DisplayValue,
		"indicator.normalized_value": golden.Phase4IndicatorExamples[0].NormalizedValue,
		"indicator.defanged_value":   golden.Phase4IndicatorExamples[0].DefangedValue,
	}
	createResp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
		createPayload,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	createData := phase4test.RequireSuccessData(t, createResp, http.StatusCreated)
	recordID := phase4test.MustUUID(t, createData["row"].(map[string]any)["record_id"].(string))

	changeSet := timelinetest.LookupChangeSet(t, harness.DB, createData["change_set_id"].(string))
	httptestx.RequireMutationAttribution(t, httptestx.MutationAttribution{
		ActorUserID: changeSet.ActorUserID,
		Source:      changeSet.Source,
		ClientTxnID: changeSet.ClientTxnID,
		RequestID:   changeSet.RequestID,
		CreatedAt:   changeSet.CreatedAt,
	}, adminUserID.String(), "entities.indicators.rows.create", "txn-phase4-i-4-07-create")
	if got := timelinetest.CountChangeSetMutations(t, harness.DB, createData["change_set_id"].(string)); got != 1 {
		t.Fatalf("expected one indicator create mutation row, got %d", got)
	}

	record := lookupIndicatorRecord(t, harness.DB, recordID)
	if record.DisplayValue != golden.Phase4IndicatorExamples[0].DisplayValue || record.DedupeKey == "" {
		t.Fatalf("unexpected indicator record state: %#v", record)
	}
	projected := lookupIndicatorProjection(t, harness.DB, recordID)
	if projected.ObservationCount != 0 || projected.LifecycleSummary != nil {
		t.Fatalf("expected fresh indicator projection without observations or lifecycle summary, got %#v", projected)
	}

	queryEnvelope := phase4test.QueryViewEnvelope(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IndicatorsViewSchemaID, viewLogin)
	httptestx.RequireDefaultQueryMeta(t, queryEnvelope, golden.Phase4IndicatorsViewSchemaID)
	queryRows := phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IndicatorsViewSchemaID, viewLogin)
	queryRow := phase4test.FindRow(t, queryRows, recordID.String())
	requirePhase4ViewRowFieldSurface(t, "I-4-07", queryRow, golden.Phase4IndicatorsViewSchemaID)
	if queryRow["record_id"] != recordID.String() {
		t.Fatalf("unexpected indicator query row: %#v", queryRow)
	}

	phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
	phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineSiblingRecordID)
	if _, _, err := store.CreateIndicatorObservation(context.Background(), authn.UserRecord{ID: adminUserID}, entities.IndicatorObservationCreateParams{
		IncidentID:                incidentID,
		SourceRecordID:            golden.Phase4TimelineRecordID,
		SourceFieldKey:            golden.Phase4FieldTimelineSourceText,
		OriginKind:                "interactive_cell",
		OriginLocator:             "view:timeline/record:1/cell:timeline.source_text/span:1-9",
		ObservedText:              golden.Phase4IndicatorExamples[0].DefangedValue,
		ResolvedIndicatorRecordID: &recordID,
		CreatedAt:                 golden.Phase4PastTime,
	}); err != nil {
		t.Fatalf("create indicator observation one: %v", err)
	}
	if _, _, err := store.CreateIndicatorObservation(context.Background(), authn.UserRecord{ID: adminUserID}, entities.IndicatorObservationCreateParams{
		IncidentID:                incidentID,
		SourceRecordID:            golden.Phase4TimelineSiblingRecordID,
		SourceFieldKey:            golden.Phase4FieldTimelineSummary,
		OriginKind:                "interactive_cell",
		OriginLocator:             "view:timeline/record:2/cell:timeline.summary/span:1-9",
		ObservedText:              golden.Phase4IndicatorExamples[0].DefangedValue,
		ResolvedIndicatorRecordID: &recordID,
		CreatedAt:                 golden.Phase4BaseTime,
	}); err != nil {
		t.Fatalf("create indicator observation two: %v", err)
	}
	if _, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), authn.UserRecord{ID: adminUserID}, entities.IndicatorLifecycleAppendParams{
		IncidentID:        incidentID,
		IndicatorRecordID: recordID,
		LifecycleState:    "active",
		ValidFrom:         golden.Phase4PastTime,
		CreatedAt:         golden.Phase4PastTime,
	}); err != nil {
		t.Fatalf("append indicator lifecycle interval: %v", err)
	}

	queryRowsAfter := phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IndicatorsViewSchemaID, viewLogin)
	queryRowAfter := phase4test.FindRow(t, queryRowsAfter, recordID.String())
	requirePhase4ViewRowFieldSurface(t, "I-4-07", queryRowAfter, golden.Phase4IndicatorsViewSchemaID)
	cells := queryRowAfter["cells"].(map[string]any)
	if cells["indicator.observation_count"].(map[string]any)["value"] != float64(2) {
		t.Fatalf("expected indicator readback observation_count=2, got %#v", queryRowAfter)
	}
	if cells["indicator.lifecycle_summary"].(map[string]any)["value"] != "active" {
		t.Fatalf("expected indicator readback lifecycle_summary=active, got %#v", queryRowAfter)
	}
	if listIndicatorObservations(t, harness.DB, incidentID)[0].ResolvedIndicatorRecordID == nil {
		t.Fatalf("expected indicator observations to remain source-bound resolved rows")
	}

	indicatorProjectionBefore := lookupIndicatorProjection(t, harness.DB, recordID)
	if _, err := harness.DB.ExecContext(context.Background(), `DELETE FROM indicator_grid_projection WHERE incident_id = $1`, incidentID); err != nil {
		t.Fatalf("clear indicator projection rows: %v", err)
	}
	if err := projections.NewStore(harness.Server.Runtime.Postgres).RebuildIncidentIndicators(context.Background(), incidentID); err != nil {
		t.Fatalf("rebuild indicator projections: %v", err)
	}
	indicatorProjectionAfter := lookupIndicatorProjection(t, harness.DB, recordID)
	httptestx.RequireProjectionDeterminism(t, indicatorProjectionBefore, indicatorProjectionAfter)

	queryRowAfterRebuild := phase4test.FindRow(t, phase4test.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), golden.Phase4IndicatorsViewSchemaID, viewLogin), recordID.String())
	requirePhase4ViewRowFieldSurface(t, "I-4-07", queryRowAfterRebuild, golden.Phase4IndicatorsViewSchemaID)
	httptestx.RequireProjectionDeterminism(t, queryRowAfter["cells"], queryRowAfterRebuild["cells"])

	replayStableBefore := httptestx.ReplayCounts{
		ChangeSets: phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
		MutationRows: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
	}
	replayResp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
		createPayload,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	replayData := phase4test.RequireSuccessData(t, replayResp, http.StatusOK)
	if replayData["change_set_id"] != createData["change_set_id"] {
		t.Fatalf("expected indicator replay to reuse original payload, got %#v %#v", createData, replayData)
	}
	httptestx.RequireReplayScaffold(t, httptestx.ReplayExpectation{
		FirstStatus:     http.StatusCreated,
		ReplayStatus:    http.StatusOK,
		DivergentStatus: http.StatusConflict,
		DivergentCode:   "client_txn_conflict",
		StableBefore:    replayStableBefore,
		StableAfter: httptestx.ReplayCounts{
			ChangeSets: phase4test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			MutationRows: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		},
	})

	divergentResp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":            "txn-phase4-i-4-07-create",
			"indicator.indicator_type": golden.Phase4IndicatorExamples[0].IndicatorType,
			"indicator.value_kind":     golden.Phase4IndicatorExamples[0].ValueKind,
			"indicator.display_value":  "203.0.113.25",
		},
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	divergentBody := phase4test.RequireErrorBody(t, divergentResp, http.StatusConflict, "client_txn_conflict")
	httptestx.RequireDivergentReplayRejected(t, divergentResp.StatusCode, divergentBody["error"].(map[string]any)["code"].(string), "client_txn_conflict")

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incident_memberships
   SET role = 'viewer',
       updated_at = now(),
       updated_by_user_id = $3
 WHERE incident_id = $1
   AND user_id = $2
`, incidentID, adminUserID, adminUserID); err != nil {
		t.Fatalf("demote indicator actor membership: %v", err)
	}
	deniedResp := phase4test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+golden.Phase4IndicatorsViewSchemaID+"/rows",
		map[string]any{
			"client_txn_id":            "txn-phase4-i-4-07-denied",
			"indicator.indicator_type": golden.Phase4IndicatorExamples[1].IndicatorType,
			"indicator.value_kind":     golden.Phase4IndicatorExamples[1].ValueKind,
			"indicator.display_value":  golden.Phase4IndicatorExamples[1].DisplayValue,
		},
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	deniedBody := phase4test.RequireErrorBody(t, deniedResp, http.StatusForbidden, "authorization_denied")
	httptestx.RequireAuthorizationReDerived(
		t,
		httptestx.AuthorizationOutcome{Status: http.StatusCreated},
		httptestx.AuthorizationOutcome{Status: deniedResp.StatusCode, Code: deniedBody["error"].(map[string]any)["code"].(string)},
	)
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}

type loginResult struct {
	sessionCookie *http.Cookie
	csrfCookie    *http.Cookie
}

type indicatorRecordRow struct {
	RecordID        uuid.UUID
	IncidentID      uuid.UUID
	IndicatorType   string
	ValueKind       string
	DisplayValue    string
	NormalizedValue *string
	DedupeKey       string
	DefangedValue   *string
	HashAlgorithm   *string
	HashValue       *string
	STIXPattern     *string
	RowVersion      int64
	CreatedByUser   uuid.UUID
	UpdatedByUser   uuid.UUID
}

type indicatorProjectionRow struct {
	RecordID            uuid.UUID
	RowVersion          int64
	IndicatorType       string
	ValueKind           string
	DisplayValue        string
	NormalizedValue     *string
	DefangedValue       *string
	HashAlgorithm       *string
	HashValue           *string
	STIXPattern         *string
	FirstObservedAt     *time.Time
	LastObservedAt      *time.Time
	ObservationCount    int
	LifecycleSummary    *string
	SupportingLinkCount int
}

type indicatorObservationRow struct {
	ObservationID             uuid.UUID
	SourceRecordID            uuid.UUID
	SourceFieldKey            string
	OriginKind                string
	OriginLocator             string
	ObservedText              string
	ParsedIndicatorType       *string
	NormalizedCandidate       *string
	ResolutionStatus          string
	ResolvedIndicatorRecordID *uuid.UUID
	RowVersion                int64
}

type indicatorLifecycleIntervalRow struct {
	IntervalID     uuid.UUID
	IndicatorID    uuid.UUID
	LifecycleState string
	ValidFrom      time.Time
	ValidTo        *time.Time
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

func createPhase4EntityRow(t testing.TB, serverURL string, incidentID string, viewSchemaID string, actor phase4test.LoginResult, body map[string]any, wantStatus int) map[string]any {
	t.Helper()

	resp := phase4test.DoJSON(
		t,
		http.MethodPost,
		serverURL+"/api/v1/incidents/"+incidentID+"/views/"+viewSchemaID+"/rows",
		body,
		phase4test.WithCookies(actor.SessionCookie, actor.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value),
	)
	return phase4test.RequireSuccessData(t, resp, wantStatus)
}

func loginPhase4LocalUser(t testing.TB, server *httptestx.Server, username string, password string) phase4test.LoginResult {
	t.Helper()

	resp := phase4test.DoJSON(t, http.MethodPost, server.HTTP.URL+"/api/v1/auth/login", map[string]any{
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
	return phase4test.LoginResult{SessionCookie: sessionCookie, CSRFCookie: csrfCookie}
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
	phase4test.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "host")

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
	phase4test.SeedRecordEnvelope(t, db, incidentID, actorUserID, recordID, "timeline_event")

	if _, err := db.ExecContext(context.Background(), `
INSERT INTO timeline_events (record_id, incident_id, summary, capture_state, created_by_user_id, updated_by_user_id)
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
	phase4test.SeedRecordEnvelope(t, db, incidentID, actorUserID, assessmentID, "assessment")

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

func collectRecordChanges(t testing.TB, changes <-chan platformws.RecordChange, want int, timeout time.Duration) []platformws.RecordChange {
	t.Helper()

	deadline := time.After(timeout)
	collected := make([]platformws.RecordChange, 0, want)
	for len(collected) < want {
		select {
		case change := <-changes:
			collected = append(collected, change)
		case <-deadline:
			t.Fatalf("timed out waiting for %d record changes, got %#v", want, collected)
		}
	}
	return collected
}

func requireRecordChange(t testing.TB, changes []platformws.RecordChange, recordID uuid.UUID, viewSchemaID string) {
	t.Helper()

	for _, change := range changes {
		if change.RecordID == recordID && change.ViewSchemaID == viewSchemaID {
			return
		}
	}
	payload, _ := json.Marshal(changes)
	t.Fatalf("expected record change for record=%s view=%s, got %s", recordID, viewSchemaID, string(payload))
}

func httptestSuccess(t testing.TB, resp *http.Response, wantStatus int) map[string]any {
	t.Helper()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status: got %d want %d body=%#v", resp.StatusCode, wantStatus, httptestx.ReadJSONBody(t, resp))
	}
	return httptestx.RequireSuccessEnvelope(t, resp, wantStatus)["data"].(map[string]any)
}

func httptestError(t testing.TB, resp *http.Response, wantStatus int, wantCode string) map[string]any {
	t.Helper()
	return httptestx.RequireErrorEnvelope(t, resp, wantStatus, wantCode)
}

func requirePhase4ViewRowFieldSurface(t testing.TB, testID string, row map[string]any, viewSchemaID string) {
	t.Helper()

	httptestx.RequireFieldKeyConformance(
		t,
		phase4test.SortedRowFieldKeys(t, row),
		phase4test.AllowedFieldKeys(t, testID, viewSchemaID),
	)
}

func requireIndicatorCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		t.Fatalf("missing indicator cell %s in %#v", fieldKey, row)
	}
	if got := cell["value"]; got != want {
		t.Fatalf("unexpected indicator cell %s: got %#v want %#v", fieldKey, got, want)
	}
}

func lookupIndicatorRecord(t testing.TB, db *sql.DB, recordID uuid.UUID) indicatorRecordRow {
	t.Helper()

	var (
		row             indicatorRecordRow
		recordIDRaw     string
		incidentIDRaw   string
		normalizedValue sql.NullString
		defangedValue   sql.NullString
		hashAlgorithm   sql.NullString
		hashValue       sql.NullString
		stixPattern     sql.NullString
		createdByRaw    string
		updatedByRaw    string
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    record_id::text,
    incident_id::text,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    dedupe_key,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    row_version,
    created_by_user_id::text,
    updated_by_user_id::text
  FROM indicators
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &incidentIDRaw, &row.IndicatorType, &row.ValueKind, &row.DisplayValue, &normalizedValue, &row.DedupeKey, &defangedValue, &hashAlgorithm, &hashValue, &stixPattern, &row.RowVersion, &createdByRaw, &updatedByRaw); err != nil {
		t.Fatalf("lookup indicator record: %v", err)
	}
	row.RecordID = mustUUID(t, recordIDRaw)
	row.IncidentID = mustUUID(t, incidentIDRaw)
	row.NormalizedValue = nullStringPointer(normalizedValue)
	row.DefangedValue = nullStringPointer(defangedValue)
	row.HashAlgorithm = nullStringPointer(hashAlgorithm)
	row.HashValue = nullStringPointer(hashValue)
	row.STIXPattern = nullStringPointer(stixPattern)
	row.CreatedByUser = mustUUID(t, createdByRaw)
	row.UpdatedByUser = mustUUID(t, updatedByRaw)
	return row
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

func lookupIndicatorProjection(t testing.TB, db *sql.DB, recordID uuid.UUID) indicatorProjectionRow {
	t.Helper()

	var (
		row              indicatorProjectionRow
		recordIDRaw      string
		normalizedValue  sql.NullString
		defangedValue    sql.NullString
		hashAlgorithm    sql.NullString
		hashValue        sql.NullString
		stixPattern      sql.NullString
		firstObservedAt  sql.NullTime
		lastObservedAt   sql.NullTime
		lifecycleSummary sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    record_id::text,
    row_version,
    indicator_type,
    value_kind,
    display_value,
    normalized_value,
    defanged_value,
    hash_algorithm,
    hash_value,
    stix_pattern,
    first_observed_at,
    last_observed_at,
    observation_count,
    lifecycle_summary,
    supporting_link_count
  FROM indicator_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&recordIDRaw, &row.RowVersion, &row.IndicatorType, &row.ValueKind, &row.DisplayValue, &normalizedValue, &defangedValue, &hashAlgorithm, &hashValue, &stixPattern, &firstObservedAt, &lastObservedAt, &row.ObservationCount, &lifecycleSummary, &row.SupportingLinkCount); err != nil {
		t.Fatalf("lookup indicator projection: %v", err)
	}
	row.RecordID = mustUUID(t, recordIDRaw)
	row.NormalizedValue = nullStringPointer(normalizedValue)
	row.DefangedValue = nullStringPointer(defangedValue)
	row.HashAlgorithm = nullStringPointer(hashAlgorithm)
	row.HashValue = nullStringPointer(hashValue)
	row.STIXPattern = nullStringPointer(stixPattern)
	row.FirstObservedAt = nullTimePointer(firstObservedAt)
	row.LastObservedAt = nullTimePointer(lastObservedAt)
	row.LifecycleSummary = nullStringPointer(lifecycleSummary)
	return row
}

func listIndicatorObservations(t testing.TB, db *sql.DB, incidentID uuid.UUID) []indicatorObservationRow {
	t.Helper()

	rows, err := db.QueryContext(context.Background(), `
SELECT
    indicator_observation_id::text,
    source_record_id::text,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id::text,
    row_version
  FROM indicator_observations
 WHERE incident_id = $1
 ORDER BY created_at ASC, indicator_observation_id ASC
`, incidentID)
	if err != nil {
		t.Fatalf("list indicator observations: %v", err)
	}
	defer rows.Close()

	result := make([]indicatorObservationRow, 0)
	for rows.Next() {
		var (
			row               indicatorObservationRow
			observationIDRaw  string
			sourceRecordIDRaw string
			parsedType        sql.NullString
			normalized        sql.NullString
			resolvedID        sql.NullString
		)
		if err := rows.Scan(&observationIDRaw, &sourceRecordIDRaw, &row.SourceFieldKey, &row.OriginKind, &row.OriginLocator, &row.ObservedText, &parsedType, &normalized, &row.ResolutionStatus, &resolvedID, &row.RowVersion); err != nil {
			t.Fatalf("scan indicator observation: %v", err)
		}
		row.ObservationID = mustUUID(t, observationIDRaw)
		row.SourceRecordID = mustUUID(t, sourceRecordIDRaw)
		row.ParsedIndicatorType = nullStringPointer(parsedType)
		row.NormalizedCandidate = nullStringPointer(normalized)
		if resolvedID.Valid {
			value := mustUUID(t, resolvedID.String)
			row.ResolvedIndicatorRecordID = &value
		}
		result = append(result, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate indicator observations: %v", err)
	}
	return result
}

func lookupIndicatorLifecycleInterval(t testing.TB, db *sql.DB, intervalID uuid.UUID) indicatorLifecycleIntervalRow {
	t.Helper()

	var (
		row            indicatorLifecycleIntervalRow
		intervalIDRaw  string
		indicatorIDRaw string
		validTo        sql.NullTime
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT
    indicator_state_interval_id::text,
    indicator_record_id::text,
    lifecycle_state,
    valid_from,
    valid_to
  FROM indicator_state_intervals
 WHERE indicator_state_interval_id = $1
`, intervalID).Scan(&intervalIDRaw, &indicatorIDRaw, &row.LifecycleState, &row.ValidFrom, &validTo); err != nil {
		t.Fatalf("lookup indicator lifecycle interval: %v", err)
	}
	row.IntervalID = mustUUID(t, intervalIDRaw)
	row.IndicatorID = mustUUID(t, indicatorIDRaw)
	row.ValidTo = nullTimePointer(validTo)
	return row
}

func queryCount(t testing.TB, db *sql.DB, query string, args ...any) int {
	t.Helper()

	var count int
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query count: %v", err)
	}
	return count
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

func nullStringPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func nullTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func derefStringPointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
