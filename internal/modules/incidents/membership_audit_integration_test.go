package incidents_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentMembershipAuditRouteAuthorizationScopeFiltersAndKeyset_Integration(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "incident-membership-audit-route")
	adminLogin, adminID := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	firstIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-audit-route-first",
		"incident_key":  "IR-MAUDIT-ONE",
		"title":         "Membership audit one",
	})
	secondIncident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-membership-audit-route-second",
		"incident_key":  "IR-MAUDIT-TWO",
		"title":         "Membership audit two",
	})
	firstIncidentID := firstIncident["incident_id"].(string)
	secondIncidentID := secondIncident["incident_id"].(string)
	firstPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + firstIncidentID + "/membership-audit-events"
	secondPath := harness.Server.HTTP.URL + "/api/v1/incidents/" + secondIncidentID + "/membership-audit-events"

	const viewerSecret = "JBSWY3DPEHPK3RAB"
	viewerID := flowtest.SeedLocalUserWithActiveTOTP(
		t,
		harness.DB,
		"membership-audit-viewer@example.test",
		"Membership Audit Viewer",
		"MembershipAuditViewer1!",
		true,
		false,
		viewerSecret,
	)
	scenariotest.CreateMembership(t, harness.Server, adminLogin, firstIncidentID, map[string]any{
		"client_txn_id": "txn-membership-audit-route-viewer",
		"user_id":       viewerID,
		"role":          "viewer",
	})
	viewerLogin := flowtest.LoginLocalUserWithSecondFactor(
		t,
		harness.Server.HTTP.URL,
		"membership-audit-viewer@example.test",
		"MembershipAuditViewer1!",
		flowtest.GenerateTOTPCode(t, viewerSecret),
	)

	const deploymentAdminSecret = "JBSWY3DPEHPK3RAC"
	flowtest.SeedLocalUserWithActiveTOTP(
		t,
		harness.DB,
		"membership-audit-deployment-admin@example.test",
		"Membership Audit Deployment Admin",
		"MembershipAuditDeploymentAdmin1!",
		true,
		true,
		deploymentAdminSecret,
	)
	deploymentAdminLogin := flowtest.LoginLocalUserWithSecondFactor(
		t,
		harness.Server.HTTP.URL,
		"membership-audit-deployment-admin@example.test",
		"MembershipAuditDeploymentAdmin1!",
		flowtest.GenerateTOTPCode(t, deploymentAdminSecret),
	)

	const nonmemberSecret = "JBSWY3DPEHPK3RAD"
	flowtest.SeedLocalUserWithActiveTOTP(
		t,
		harness.DB,
		"membership-audit-nonmember@example.test",
		"Membership Audit Nonmember",
		"MembershipAuditNonmember1!",
		true,
		false,
		nonmemberSecret,
	)
	nonmemberLogin := flowtest.LoginLocalUserWithSecondFactor(
		t,
		harness.Server.HTTP.URL,
		"membership-audit-nonmember@example.test",
		"MembershipAuditNonmember1!",
		flowtest.GenerateTOTPCode(t, nonmemberSecret),
	)

	anonymousInvalid := httptestx.DoJSON(t, http.MethodGet, firstPath+"?action_code=future_action", nil)
	httptestx.RequireErrorEnvelope(t, anonymousInvalid, http.StatusUnauthorized, "session_required")
	deploymentAdminInvalid := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?action_code=future_action",
		nil,
		httptestx.WithCookies(deploymentAdminLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, deploymentAdminInvalid, http.StatusNotFound, "incident_not_found")
	nonmemberInvalid := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?action_code=future_action",
		nil,
		httptestx.WithCookies(nonmemberLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, nonmemberInvalid, http.StatusNotFound, "incident_not_found")
	viewerInvalid := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?action_code=future_action",
		nil,
		httptestx.WithCookies(viewerLogin.SessionCookie),
	)
	httptestx.RequireErrorEnvelope(t, viewerInvalid, http.StatusForbidden, "authorization_denied")

	flowtest.SeedLocalUserFlags(
		t,
		harness.DB,
		"membership-audit-bootstrap@example.test",
		"Membership Audit Bootstrap",
		"MembershipAuditBootstrap1!",
		true,
		false,
		true,
	)
	bootstrapToken := flowtest.RequireBootstrapLogin(
		t,
		harness.Server.HTTP.URL,
		"membership-audit-bootstrap@example.test",
		"MembershipAuditBootstrap1!",
	)
	bootstrapInvalid := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?action_code=future_action",
		nil,
		httptestx.WithHeader("Authorization", "Bearer "+bootstrapToken),
	)
	httptestx.RequireErrorEnvelope(t, bootstrapInvalid, http.StatusConflict, "credential_bootstrap_rejected")

	invalidAction := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?action_code=future_action",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	requireAuditReasonCode(t, invalidAction, "invalid_list_query", "invalid_filter_value")
	duplicateBeforeUnknown := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?unknown=value&action_code=membership_created&action_code=membership_deleted",
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	requireAuditReasonCode(t, duplicateBeforeUnknown, "invalid_list_query", "duplicate_query_member")

	occurredAt := time.Date(2035, time.July, 25, 19, 30, 0, 123456000, time.UTC)
	targetID := viewerID
	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000933",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      firstIncidentID,
		OccurredAt:   occurredAt,
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionMembershipCreated,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"role","value_state":"visible","before":null,"after":"viewer"}]`,
	})
	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000932",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      firstIncidentID,
		OccurredAt:   occurredAt,
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionMembershipRoleChanged,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"role","value_state":"visible","before":"viewer","after":"reviewer"}]`,
	})
	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000931",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      firstIncidentID,
		OccurredAt:   occurredAt,
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionMembershipDeleted,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"password","value_state":"redacted","before":null,"after":null},{"field_path":"alpha","value_state":"visible","before":null,"after":"safe"}]`,
	})
	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000934",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      secondIncidentID,
		OccurredAt:   occurredAt.Add(time.Minute),
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionMembershipDeleted,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"role","value_state":"visible","before":"viewer","after":null}]`,
	})
	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000935",
		ScopeKind:    administrativeaudit.ScopeDeployment,
		OccurredAt:   occurredAt.Add(2 * time.Minute),
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionUserProfileUpdated,
		TargetKind:   administrativeaudit.TargetUser,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"display_name","value_state":"visible","before":"Before","after":"After"}]`,
	})

	rangeQuery := url.Values{
		"occurred_at_gte": {occurredAt.Add(-time.Second).Format(time.RFC3339Nano)},
		"occurred_at_lt":  {occurredAt.Add(time.Second).Format(time.RFC3339Nano)},
	}
	list := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?"+rangeQuery.Encode(),
		nil,
		httptestx.WithHeader("Authorization", "Bearer "+adminLogin.SessionCookie.Value),
	)
	listBody := httptestx.RequireSuccessEnvelope(t, list, http.StatusOK)
	rows := requireMembershipAuditRows(t, listBody)
	if len(rows) != 3 {
		t.Fatalf("expected only three addressed-incident projections, got %#v", rows)
	}
	if got := rows[0]["audit_event_id"]; got != "50000000-0000-0000-0000-000000000933" {
		t.Fatalf("expected descending UUID tie-breaker, got first event %v", got)
	}
	requireMembershipAuditResource(t, rows[0], firstIncidentID)
	safeChanges := rows[2]["changes"].([]any)
	if got := []string{
		safeChanges[0].(map[string]any)["field_path"].(string),
		safeChanges[1].(map[string]any)["field_path"].(string),
	}; !reflect.DeepEqual(got, []string{"alpha", "password"}) {
		t.Fatalf("changes must be field_path sorted, got %v", got)
	}
	redacted := safeChanges[1].(map[string]any)
	if redacted["value_state"] != administrativeaudit.ValueRedacted || redacted["before"] != nil || redacted["after"] != nil {
		t.Fatalf("redacted change exposed a value: %#v", redacted)
	}

	filterQuery := url.Values{
		"action_code":     {administrativeaudit.ActionMembershipCreated},
		"target_kind":     {administrativeaudit.TargetIncidentMembership},
		"target_id":       {targetID},
		"occurred_at_gte": rangeQuery["occurred_at_gte"],
		"occurred_at_lt":  rangeQuery["occurred_at_lt"],
	}
	filtered := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?"+filterQuery.Encode(),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	filteredRows := requireMembershipAuditRows(t, httptestx.RequireSuccessEnvelope(t, filtered, http.StatusOK))
	if len(filteredRows) != 1 || filteredRows[0]["audit_event_id"] != "50000000-0000-0000-0000-000000000933" {
		t.Fatalf("exact membership audit filters returned %#v", filteredRows)
	}

	pageQuery := url.Values{
		"limit":           {"1"},
		"occurred_at_gte": rangeQuery["occurred_at_gte"],
		"occurred_at_lt":  rangeQuery["occurred_at_lt"],
	}
	firstPage := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?"+pageQuery.Encode(),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	firstPageBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstPageRows := requireMembershipAuditRows(t, firstPageBody)
	if len(firstPageRows) != 1 || firstPageRows[0]["audit_event_id"] != "50000000-0000-0000-0000-000000000933" {
		t.Fatalf("unexpected first membership-audit page: %#v", firstPageRows)
	}
	cursor := requireMembershipAuditCursor(t, firstPageBody)

	insertIncidentAuditProjection(t, harness.DB, incidentAuditFixture{
		AuditEventID: "50000000-0000-0000-0000-000000000936",
		ScopeKind:    administrativeaudit.ScopeIncident,
		ScopeID:      firstIncidentID,
		OccurredAt:   occurredAt.Add(time.Minute),
		ActorUserID:  adminID,
		ActionCode:   administrativeaudit.ActionMembershipDeleted,
		TargetKind:   administrativeaudit.TargetIncidentMembership,
		TargetID:     targetID,
		ChangesJSON:  `[{"field_path":"role","value_state":"visible","before":"reviewer","after":null}]`,
	})
	continuedQuery := cloneAuditQuery(pageQuery)
	continuedQuery.Set("cursor_token", cursor)
	continued := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?"+continuedQuery.Encode(),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	continuedRows := requireMembershipAuditRows(t, httptestx.RequireSuccessEnvelope(t, continued, http.StatusOK))
	if len(continuedRows) != 1 || continuedRows[0]["audit_event_id"] != "50000000-0000-0000-0000-000000000932" {
		t.Fatalf("newer inserts must not shift keyset continuation: %#v", continuedRows)
	}

	mismatchedQuery := cloneAuditQuery(pageQuery)
	mismatchedQuery.Set("action_code", administrativeaudit.ActionMembershipCreated)
	mismatchedQuery.Set("cursor_token", cursor)
	mismatched := httptestx.DoJSON(
		t,
		http.MethodGet,
		firstPath+"?"+mismatchedQuery.Encode(),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	requireAuditReasonCode(t, mismatched, "invalid_pagination_request", "cursor_query_mismatch")

	otherIncidentQuery := cloneAuditQuery(pageQuery)
	otherIncidentQuery.Set("cursor_token", cursor)
	otherIncident := httptestx.DoJSON(
		t,
		http.MethodGet,
		secondPath+"?"+otherIncidentQuery.Encode(),
		nil,
		httptestx.WithCookies(adminLogin.SessionCookie),
	)
	requireAuditReasonCode(t, otherIncident, "invalid_pagination_request", "cursor_query_mismatch")
}

type incidentAuditFixture struct {
	AuditEventID string
	ScopeKind    string
	ScopeID      string
	OccurredAt   time.Time
	ActorUserID  string
	ActionCode   string
	TargetKind   string
	TargetID     string
	ChangesJSON  string
}

func insertIncidentAuditProjection(t testing.TB, db *sql.DB, fixture incidentAuditFixture) {
	t.Helper()
	var scopeID any
	if fixture.ScopeID != "" {
		scopeID = fixture.ScopeID
	}
	var targetID any
	if fixture.TargetID != "" {
		targetID = fixture.TargetID
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO deployment_admin_audit_events (
    id, actor_user_id, target_user_id, incident_id, event_source, event_kind, created_at
) VALUES ($1, $2, $3, $4, 'test.fixture', $5, $6)
`,
		fixture.AuditEventID,
		fixture.ActorUserID,
		targetID,
		scopeID,
		fixture.ActionCode,
		fixture.OccurredAt,
	); err != nil {
		t.Fatalf("insert raw incident administrative-audit fixture: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO administrative_audit_projections (
    audit_event_id, scope_kind, scope_id, occurred_at, actor_kind, actor_user_id,
    source, action_code, target_kind, target_id, changes, reason_code
) VALUES ($1, $2, $3, $4, 'user', $5, 'api', $6, $7, $8, $9::jsonb, NULL)
`,
		fixture.AuditEventID,
		fixture.ScopeKind,
		scopeID,
		fixture.OccurredAt,
		fixture.ActorUserID,
		fixture.ActionCode,
		fixture.TargetKind,
		targetID,
		fixture.ChangesJSON,
	); err != nil {
		t.Fatalf("insert incident administrative-audit projection fixture: %v", err)
	}
}

func requireMembershipAuditRows(t testing.TB, body map[string]any) []map[string]any {
	t.Helper()
	data, ok := body["data"].(map[string]any)
	if !ok {
		t.Fatalf("membership audit data has type %T", body["data"])
	}
	if len(data) != 1 {
		t.Fatalf("membership audit data members must be exact: %#v", data)
	}
	rawRows, ok := data["audit_events"].([]any)
	if !ok {
		t.Fatalf("membership audit rows have type %T", data["audit_events"])
	}
	rows := make([]map[string]any, 0, len(rawRows))
	for _, rawRow := range rawRows {
		row, ok := rawRow.(map[string]any)
		if !ok {
			t.Fatalf("membership audit row has type %T", rawRow)
		}
		rows = append(rows, row)
	}
	return rows
}

func requireMembershipAuditResource(t testing.TB, row map[string]any, incidentID string) {
	t.Helper()
	gotKeys := make([]string, 0, len(row))
	for key := range row {
		gotKeys = append(gotKeys, key)
	}
	sort.Strings(gotKeys)
	wantKeys := []string{
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
	}
	if !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("membership audit resource keys got %v want %v", gotKeys, wantKeys)
	}
	if row["scope_kind"] != administrativeaudit.ScopeIncident || row["scope_id"] != incidentID {
		t.Fatalf("membership audit resource escaped addressed incident: %#v", row)
	}
	occurredAt, ok := row["occurred_at"].(string)
	if !ok || !strings.HasSuffix(occurredAt, "Z") {
		t.Fatalf("membership audit timestamp must be UTC: %#v", row["occurred_at"])
	}
}

func requireMembershipAuditCursor(t testing.TB, body map[string]any) string {
	t.Helper()
	meta := body["meta"].(map[string]any)
	paging := meta["paging"].(map[string]any)
	cursor, ok := paging["next_cursor"].(string)
	if !ok || cursor == "" {
		t.Fatalf("missing membership audit cursor: %#v", paging)
	}
	return cursor
}

func requireAuditReasonCode(t testing.TB, response *http.Response, code string, reasonCode string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, code)
	errorBody := body["error"].(map[string]any)
	details := errorBody["details"].(map[string]any)
	if details["reason_code"] != reasonCode {
		t.Fatalf("%s reason_code got %#v want %q", code, details["reason_code"], reasonCode)
	}
}

func cloneAuditQuery(source url.Values) url.Values {
	cloned := make(url.Values, len(source))
	for key, values := range source {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
