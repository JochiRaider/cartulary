package auth_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDeploymentAdministrativeAuditRoute_Integration(t *testing.T) {
	runtime := flowtest.StartRuntime(t)
	server, db := startServer(t, runtime, "authentication-deployment-administrative-audit")
	defer db.Close()

	adminID := seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000811", "audit-route-admin@example.test", "Audit Route Admin", "AuditRouteAdmin1!", true)
	targetID := seedFixedLocalUser(t, db, "10000000-0000-0000-0000-000000000812", "audit-route-target@example.test", "Audit Route Target", "AuditRouteTarget1!", false)
	adminSession, _ := loginLocalUser(t, server, "audit-route-admin@example.test", "AuditRouteAdmin1!", nil)
	nonAdminSession, _ := loginLocalUser(t, server, "audit-route-target@example.test", "AuditRouteTarget1!", nil)
	seedLocalUser(t, db, "audit-route-bootstrap@example.test", "Audit Route Bootstrap", "AuditRouteBootstrap1!", true)
	bootstrapToken := requireBootstrapLogin(t, server, "audit-route-bootstrap@example.test", "AuditRouteBootstrap1!")

	anonymousInvalid := doJSON(t, http.MethodGet, server.HTTP.URL+"/api/v1/administrative-audit-events?target_id=invalid", nil)
	httptestx.RequireErrorEnvelope(t, anonymousInvalid, http.StatusUnauthorized, "session_required")
	nonAdminInvalid := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?target_id=invalid",
		nil,
		withCookies(nonAdminSession),
	)
	httptestx.RequireErrorEnvelope(t, nonAdminInvalid, http.StatusForbidden, "authorization_denied")
	bootstrapInvalid := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?target_id=invalid",
		nil,
		withHeader("Authorization", "Bearer "+bootstrapToken),
	)
	httptestx.RequireErrorEnvelope(t, bootstrapInvalid, http.StatusConflict, "credential_bootstrap_rejected")
	invalidAction := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?action_code=future_action",
		nil,
		withCookies(adminSession),
	)
	invalidActionBody := httptestx.RequireErrorEnvelope(t, invalidAction, http.StatusBadRequest, "invalid_list_query")
	if got := invalidActionBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != "invalid_filter_value" {
		t.Fatalf("unexpected invalid action filter reason: %v", got)
	}
	duplicateBeforeUnknown := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?unknown=value&action_code=user_created&action_code=password_reset",
		nil,
		withCookies(adminSession),
	)
	duplicateBody := httptestx.RequireErrorEnvelope(t, duplicateBeforeUnknown, http.StatusBadRequest, "invalid_list_query")
	if got := duplicateBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != "duplicate_query_member" {
		t.Fatalf("duplicate query member must precede unknown-member validation, got %v", got)
	}

	occurredAt := time.Date(2026, time.July, 25, 19, 30, 0, 123456000, time.UTC)
	insertAdministrativeAuditProjection(t, db, administrativeAuditFixture{
		AuditEventID: "20000000-0000-0000-0000-000000000811",
		ScopeKind:    administrativeaudit.ScopeDeployment,
		OccurredAt:   occurredAt,
		ActorKind:    administrativeaudit.ActorUser,
		ActorUserID:  adminID,
		Source:       administrativeaudit.SourceUI,
		ActionCode:   administrativeaudit.ActionUserProfileUpdated,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"email","value_state":"visible","before":"before@example.test","after":"after@example.test"},{"field_path":"display_name","value_state":"visible","before":"Before","after":"After"}]`,
		ReasonCode:   "profile_corrected",
	})
	insertAdministrativeAuditProjection(t, db, administrativeAuditFixture{
		AuditEventID: "20000000-0000-0000-0000-000000000812",
		ScopeKind:    administrativeaudit.ScopeDeployment,
		OccurredAt:   occurredAt,
		ActorKind:    administrativeaudit.ActorUser,
		ActorUserID:  adminID,
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionDeploymentAdminGranted,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"is_deployment_admin","value_state":"visible","before":false,"after":true}]`,
	})
	insertAdministrativeAuditProjection(t, db, administrativeAuditFixture{
		AuditEventID: "20000000-0000-0000-0000-000000000810",
		ScopeKind:    administrativeaudit.ScopeDeployment,
		OccurredAt:   occurredAt.Add(-time.Second),
		ActorKind:    administrativeaudit.ActorUser,
		ActorUserID:  adminID,
		Source:       administrativeaudit.SourceAPI,
		ActionCode:   administrativeaudit.ActionPasswordReset,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"password","value_state":"redacted","before":null,"after":null}]`,
	})
	insertAdministrativeAuditProjection(t, db, administrativeAuditFixture{
		AuditEventID: "20000000-0000-0000-0000-000000000813",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      "30000000-0000-0000-0000-000000000811",
		OccurredAt:   occurredAt.Add(time.Minute),
		ActorKind:    administrativeaudit.ActorUser,
		ActorUserID:  adminID,
		Source:       administrativeaudit.SourceUI,
		ActionCode:   administrativeaudit.ActionMembershipCreated,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"role","value_state":"visible","before":null,"after":"viewer"}]`,
	})

	list := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?actor_user_id="+url.QueryEscape(adminID),
		nil,
		withHeader("Authorization", "Bearer "+adminSession.Value),
	)
	listBody := httptestx.RequireSuccessEnvelope(t, list, http.StatusOK)
	rows := requireAdministrativeAuditRows(t, listBody)
	if len(rows) != 3 {
		t.Fatalf("expected only three deployment-scoped projections, got %#v", rows)
	}
	if got := rows[0]["audit_event_id"]; got != "20000000-0000-0000-0000-000000000812" {
		t.Fatalf("expected descending UUID tie-breaker, got first event %v", got)
	}
	requireExactMapKeys(t, rows[0],
		"action_code",
		"actor_kind",
		"actor_user_id",
		"audit_event_id",
		"changes",
		"occurred_at",
		"reason_code",
		"scope_id",
		"scope_kind",
		"source",
		"target_id",
		"target_kind",
	)
	if rows[0]["scope_kind"] != administrativeaudit.ScopeDeployment || rows[0]["scope_id"] != nil {
		t.Fatalf("deployment route returned an invalid scope: %#v", rows[0])
	}
	if occurredAt, ok := rows[0]["occurred_at"].(string); !ok || !strings.HasSuffix(occurredAt, "Z") {
		t.Fatalf("administrative-audit timestamps must serialize in UTC: %#v", rows[0]["occurred_at"])
	}
	if _, legacy := listBody["data"].(map[string]any)["administrative_audit_events"]; legacy {
		t.Fatal("obsolete administrative_audit_events response key must not be emitted")
	}

	filtered := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?action_code=user_profile_updated&target_kind=user&target_id="+url.QueryEscape(targetID),
		nil,
		withCookies(adminSession),
	)
	filteredRows := requireAdministrativeAuditRows(t, httptestx.RequireSuccessEnvelope(t, filtered, http.StatusOK))
	if len(filteredRows) != 1 || filteredRows[0]["audit_event_id"] != "20000000-0000-0000-0000-000000000811" {
		t.Fatalf("unexpected exact filtered projection: %#v", filteredRows)
	}
	changes := filteredRows[0]["changes"].([]any)
	if len(changes) != 2 ||
		changes[0].(map[string]any)["field_path"] != "display_name" ||
		changes[1].(map[string]any)["field_path"] != "email" {
		t.Fatalf("administrative-audit changes are not field_path ascending: %#v", changes)
	}
	change := changes[0].(map[string]any)
	if change["field_path"] != "display_name" || change["before"] != "Before" || change["after"] != "After" {
		t.Fatalf("unexpected safe projected changes: %#v", changes)
	}
	redacted := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?action_code=password_reset",
		nil,
		withCookies(adminSession),
	)
	redactedRows := requireAdministrativeAuditRows(t, httptestx.RequireSuccessEnvelope(t, redacted, http.StatusOK))
	redactedChange := redactedRows[0]["changes"].([]any)[0].(map[string]any)
	if redactedChange["value_state"] != administrativeaudit.ValueRedacted || redactedChange["before"] != nil || redactedChange["after"] != nil {
		t.Fatalf("redacted change leaked values: %#v", redactedChange)
	}

	firstPage := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?actor_user_id="+url.QueryEscape(adminID)+"&limit=1",
		nil,
		withCookies(adminSession),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := requireAdministrativeAuditRows(t, firstPageBody)
	if len(firstPageRows) != 1 || firstPageRows[0]["audit_event_id"] != "20000000-0000-0000-0000-000000000812" {
		t.Fatalf("unexpected first keyset page: %#v", firstPageRows)
	}
	cursor := requirePagingCursor(t, firstPageBody)

	insertAdministrativeAuditProjection(t, db, administrativeAuditFixture{
		AuditEventID: "20000000-0000-0000-0000-000000000814",
		ScopeKind:    administrativeaudit.ScopeDeployment,
		OccurredAt:   occurredAt.Add(time.Hour),
		ActorKind:    administrativeaudit.ActorSystem,
		Source:       administrativeaudit.SourceSystem,
		ActionCode:   administrativeaudit.ActionLegacyAdministrativeEvent,
		TargetKind:   administrativeaudit.TargetLegacyAdministrativeEvent,
		ChangesJSON:  `[]`,
	})
	continued := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?actor_user_id="+url.QueryEscape(adminID)+"&cursor_token="+url.QueryEscape(cursor),
		nil,
		withCookies(adminSession),
	)
	continuedRows := requireAdministrativeAuditRows(t, httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK))
	if len(continuedRows) != 1 || continuedRows[0]["audit_event_id"] != "20000000-0000-0000-0000-000000000811" {
		t.Fatalf("keyset continuation shifted after a newer insert: %#v", continuedRows)
	}

	mismatched := doJSON(
		t,
		http.MethodGet,
		server.HTTP.URL+"/api/v1/administrative-audit-events?actor_user_id="+url.QueryEscape(adminID)+"&action_code=user_profile_updated&cursor_token="+url.QueryEscape(cursor),
		nil,
		withCookies(adminSession),
	)
	mismatchBody := httptestx.RequireErrorEnvelope(t, mismatched, http.StatusBadRequest, "invalid_pagination_request")
	if got := mismatchBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"]; got != "cursor_query_mismatch" {
		t.Fatalf("unexpected cursor mismatch reason: %v", got)
	}
}

type administrativeAuditFixture struct {
	AuditEventID string
	ScopeKind    string
	ScopeID      string
	OccurredAt   time.Time
	ActorKind    string
	ActorUserID  string
	Source       string
	ActionCode   string
	TargetKind   string
	TargetID     string
	ChangesJSON  string
	ReasonCode   string
}

func insertAdministrativeAuditProjection(t testing.TB, db *sql.DB, fixture administrativeAuditFixture) {
	t.Helper()
	var actorUserID any
	if fixture.ActorUserID != "" {
		actorUserID = fixture.ActorUserID
	}
	var scopeID any
	if fixture.ScopeID != "" {
		scopeID = fixture.ScopeID
	}
	var targetID any
	if fixture.TargetID != "" {
		targetID = fixture.TargetID
	}
	var reasonCode any
	if fixture.ReasonCode != "" {
		reasonCode = fixture.ReasonCode
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO deployment_admin_audit_events (
    id, actor_user_id, event_source, event_kind, reason_code, created_at
) VALUES ($1, $2, 'test.fixture', $3, $4, $5)
`, fixture.AuditEventID, actorUserID, fixture.ActionCode, reasonCode, fixture.OccurredAt); err != nil {
		t.Fatalf("insert raw administrative audit fixture: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO administrative_audit_projections (
    audit_event_id, scope_kind, scope_id, occurred_at, actor_kind, actor_user_id,
    source, action_code, target_kind, target_id, changes, reason_code
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12)
`,
		fixture.AuditEventID,
		fixture.ScopeKind,
		scopeID,
		fixture.OccurredAt,
		fixture.ActorKind,
		actorUserID,
		fixture.Source,
		fixture.ActionCode,
		fixture.TargetKind,
		targetID,
		fixture.ChangesJSON,
		reasonCode,
	); err != nil {
		t.Fatalf("insert administrative audit projection fixture: %v", err)
	}
}

func requireAdministrativeAuditRows(t testing.TB, body map[string]any) []map[string]any {
	t.Helper()
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("administrative audit response data has type %T", body["data"])
	}
	rawRows, ok := data["audit_events"].([]any)
	if !ok {
		t.Fatalf("administrative audit response rows have type %T", data["audit_events"])
	}
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row, ok := rawRow.(map[string]any)
		if !ok {
			t.Fatalf("administrative audit row has type %T", rawRow)
		}
		rows = append(rows, row)
	}
	return rows
}
