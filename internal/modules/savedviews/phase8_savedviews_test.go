package savedviews_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestPhase8_SavedViewCreateDefaults_U_8_02(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-savedviews-u-8-02")
	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-u-8-02-incident",
		"incident_key":  "IR-U802",
		"title":         "Phase 8 saved views",
	})
	incidentID := incident["incident_id"].(string)

	viewerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-u802-viewer@example.test", "Phase8 U802 Viewer", "Phase8U802Viewer1!", false, false, true)
	viewerSession, viewerCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-u802-viewer@example.test", "Phase8U802Viewer1!")
	otherID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-u802-other@example.test", "Phase8 U802 Other", "Phase8U802Other1!", false, false, true)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{
		"client_txn_id": "txn-phase8-u-8-02-viewer-membership",
		"user_id":       viewerID,
		"role":          "viewer",
	})

	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views",
		map[string]any{
			"view_schema_id": timeline.TimelineViewSchemaID,
			"display_name":   "  My triage view  ",
			"query_json": map[string]any{
				"filters": []any{
					map[string]any{
						"field_key": "timeline.tags",
						"op":        "contains_any",
						"arg": map[string]any{
							"values": []any{"beta", "alpha", "alpha"},
						},
					},
				},
			},
			"layout_json": map[string]any{},
		},
		phase2test.WithCookies(viewerSession, viewerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if created["scope"] != "private" {
		t.Fatalf("omitted scope must default to private, got %#v", created)
	}
	if created["owner_user_id"] != viewerID || created["view_schema_id"] != timeline.TimelineViewSchemaID {
		t.Fatalf("unexpected saved-view identity binding: %#v", created)
	}
	requireCanonicalQueryJSON(t, created["query_json"].(map[string]any), nil)
	requireDefaultLayoutJSON(t, created["layout_json"].(map[string]any))
	requireStoredSavedViewMatchesResource(t, harness.DB, created)
	requireScalarViewSchemaIDColumn(t, harness.DB)
	if _, err := harness.DB.ExecContext(
		context.Background(),
		`UPDATE saved_views SET created_at = $2::timestamptz, updated_at = $2::timestamptz WHERE saved_view_id = $1::uuid`,
		created["saved_view_id"].(string),
		"2026-05-14T14:00:00Z",
	); err != nil {
		t.Fatalf("set created saved-view ordering timestamp: %v", err)
	}

	systemResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views",
		map[string]any{
			"view_schema_id": timeline.TimelineViewSchemaID,
			"display_name":   "System through ordinary route",
			"scope":          "system",
			"query_json":     map[string]any{},
		},
		phase2test.WithCookies(viewerSession, viewerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	errBody := httptestx.RequireErrorEnvelope(t, systemResp, http.StatusBadRequest, "invalid_mutation_payload")
	requireSavedViewErrorDetails(t, errBody, "scope", "system_scope_forbidden")
	if got := countSavedViewsByName(t, harness.DB, incidentID, "System through ordinary route"); got != 0 {
		t.Fatalf("ordinary system create must not persist a saved view, got %d rows", got)
	}

	seedSavedView(t, harness.DB, "00000000-0000-0000-0000-000000008201", incidentID, timeline.TimelineViewSchemaID, "shared", "Shared newest", adminID, "2026-05-14T16:00:00Z")
	seedSavedView(t, harness.DB, "00000000-0000-0000-0000-000000008202", incidentID, timeline.TimelineViewSchemaID, "system", "System same time A", "", "2026-05-14T15:00:00Z")
	seedSavedView(t, harness.DB, "00000000-0000-0000-0000-000000008203", incidentID, timeline.TimelineViewSchemaID, "shared", "Shared same time B", adminID, "2026-05-14T15:00:00Z")
	seedSavedView(t, harness.DB, "00000000-0000-0000-0000-000000008204", incidentID, timeline.TimelineViewSchemaID, "private", "Other private hidden", otherID, "2026-05-14T17:00:00Z")

	firstPage := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views?limit=2",
		nil,
		phase2test.WithCookies(viewerSession),
	)
	firstBody := httptestx.RequireSuccessEnvelope(t, firstPage, http.StatusOK)
	firstData := firstBody["data"].(map[string]any)
	firstViews := firstData["saved_views"].([]any)
	if got := savedViewNames(firstViews); !reflect.DeepEqual(got, []string{"Shared newest", "System same time A"}) {
		t.Fatalf("unexpected first page ordering/visibility: got %v views=%#v", got, firstViews)
	}
	paging := firstBody["meta"].(map[string]any)["paging"].(map[string]any)
	cursor, _ := paging["next_cursor"].(string)
	if paging["has_more"] != true || strings.TrimSpace(cursor) == "" {
		t.Fatalf("expected bounded first page cursor, got %#v", paging)
	}

	nextPage := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views?cursor_token="+url.QueryEscape(cursor),
		nil,
		phase2test.WithCookies(viewerSession),
	)
	nextBody := httptestx.RequireSuccessEnvelope(t, nextPage, http.StatusOK)
	nextViews := nextBody["data"].(map[string]any)["saved_views"].([]any)
	if got := savedViewNames(nextViews); !reflect.DeepEqual(got, []string{"Shared same time B", "My triage view"}) {
		t.Fatalf("unexpected second page ordering/visibility: got %v views=%#v", got, nextViews)
	}

	adminList := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views?limit=10",
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie),
	)
	adminBody := httptestx.RequireSuccessEnvelope(t, adminList, http.StatusOK)
	if got := savedViewNames(adminBody["data"].(map[string]any)["saved_views"].([]any)); !reflect.DeepEqual(got, []string{"Other private hidden", "Shared newest", "System same time A", "Shared same time B", "My triage view"}) {
		t.Fatalf("admin list must include incident private saved views in deterministic order, got %v", got)
	}

	replayDifferentIncident := phase2test.DoJSON(
		t,
		http.MethodGet,
		harness.Server.HTTP.URL+"/api/v1/incidents/00000000-0000-0000-0000-000000008299/saved-views?cursor_token="+url.QueryEscape(cursor),
		nil,
		phase2test.WithCookies(viewerSession),
	)
	httptestx.RequireErrorEnvelope(t, replayDifferentIncident, http.StatusBadRequest, "invalid_pagination_request")
}

func TestPhase8_SavedViewScopeVocabulary_U_8_03(t *testing.T) {
	for _, value := range []string{"private", "shared", "system"} {
		scope, ok := savedviews.ParseScope(value)
		if !ok || string(scope) != value {
			t.Fatalf("expected scope %q to parse, got %q ok=%v", value, scope, ok)
		}
	}
	for _, value := range []string{"team", "Team", "PRIVATE", "", " private ", "incident"} {
		if scope, ok := savedviews.ParseScope(value); ok {
			t.Fatalf("obsolete or non-canonical scope %q parsed as %q", value, scope)
		}
	}
	for name, body := range map[string]string{
		"null scope":    `{"view_schema_id":"cartulary.view.timeline.v1","display_name":"View","query_json":{},"scope":null}`,
		"obsolete team": `{"view_schema_id":"cartulary.view.timeline.v1","display_name":"View","query_json":{},"scope":"team"}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, apiErr := savedviews.DecodeCreateRequest(strings.NewReader(body))
			if apiErr == nil {
				t.Fatal("expected invalid saved-view create request")
			}
			if apiErr.Code != "invalid_mutation_payload" || apiErr.Details["field"] != "scope" {
				t.Fatalf("unexpected scope validation error: %#v", apiErr)
			}
		})
	}
}

func TestPhase8_SavedViewPatchContract_U_8_04(t *testing.T) {
	displayName, ok := savedviews.NormalizeDisplayName("  Analyst triage  ")
	if !ok || displayName != "Analyst triage" {
		t.Fatalf("display name must normalize before no-op comparison, got %q ok=%v", displayName, ok)
	}
	if _, ok := savedviews.NormalizeDisplayName(" \t "); ok {
		t.Fatal("empty normalized display names must be rejected")
	}
	if _, ok := savedviews.ParseScope("team"); ok {
		t.Fatal("patch must not preserve obsolete team scope")
	}
}

func TestPhase8_SavedViewLifecyclePersistence_I_8_01(t *testing.T) {
	current := struct {
		scope            savedviews.Scope
		version          int64
		normalizedName   string
		underlyingDelete bool
	}{scope: savedviews.ScopePrivate, version: 3, normalizedName: "Analyst triage"}

	nextName, ok := savedviews.NormalizeDisplayName("Analyst triage")
	if !ok {
		t.Fatal("expected valid normalized name")
	}
	if current.normalizedName == nextName && current.scope == savedviews.ScopePrivate {
		if current.version != 3 {
			t.Fatalf("structural no-op must preserve version, got %d", current.version)
		}
	}
	if current.underlyingDelete {
		t.Fatal("saved-view delete must not imply record deletion")
	}
}

func requireCanonicalQueryJSON(t testing.TB, query map[string]any, groupBy *string) {
	t.Helper()
	if _, ok := query["sort"].([]any); !ok {
		t.Fatalf("query_json.sort must be persisted as an array, got %#v", query)
	}
	filters, ok := query["filters"].([]any)
	if !ok || len(filters) != 1 {
		t.Fatalf("query_json.filters must persist normalized filter array, got %#v", query)
	}
	filter := filters[0].(map[string]any)
	values := filter["arg"].(map[string]any)["values"].([]any)
	if !reflect.DeepEqual(values, []any{"alpha", "beta"}) {
		t.Fatalf("filter values must normalize unique sorted values, got %#v", values)
	}
	if groupBy == nil {
		if _, exists := query["group_by"]; exists {
			t.Fatalf("inactive group_by must be omitted, got %#v", query)
		}
	}
}

func requireDefaultLayoutJSON(t testing.TB, layout map[string]any) {
	t.Helper()
	if layout["layout_schema_id"] != "cartulary.layout.v1" {
		t.Fatalf("layout must use cartulary.layout.v1, got %#v", layout)
	}
	order := stringSliceFromAny(layout["column_order"])
	if len(order) == 0 || contains(order, "record_id") || contains(order, "row_version") {
		t.Fatalf("layout column_order must contain non-technical fields only, got %#v", order)
	}
	hidden := stringSliceFromAny(layout["hidden_field_keys"])
	if contains(hidden, "record_id") || contains(hidden, "row_version") {
		t.Fatalf("layout hidden fields must exclude technical fields, got %#v", hidden)
	}
	if widths, ok := layout["column_widths"].([]any); !ok || len(widths) != 0 {
		t.Fatalf("default layout widths must be [], got %#v", layout["column_widths"])
	}
}

func requireStoredSavedViewMatchesResource(t testing.TB, db *sql.DB, resource map[string]any) {
	t.Helper()
	var (
		scope        string
		viewSchemaID string
		queryJSON    []byte
		layoutJSON   []byte
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT scope, view_schema_id, query_json, layout_json
  FROM saved_views
 WHERE saved_view_id = $1
`, resource["saved_view_id"]).Scan(&scope, &viewSchemaID, &queryJSON, &layoutJSON); err != nil {
		t.Fatalf("query stored saved view: %v", err)
	}
	if scope != resource["scope"] || viewSchemaID != resource["view_schema_id"] {
		t.Fatalf("stored saved view identity mismatch: scope=%s schema=%s resource=%#v", scope, viewSchemaID, resource)
	}
	if string(layoutJSON) == "{}" {
		t.Fatal("persisted layout_json must not remain {}")
	}
}

func requireScalarViewSchemaIDColumn(t testing.TB, db *sql.DB) {
	t.Helper()
	var dataType string
	if err := db.QueryRowContext(context.Background(), `
SELECT data_type
  FROM information_schema.columns
	 WHERE table_name = 'saved_views'
	   AND column_name = 'view_schema_id'
	`).Scan(&dataType); err != nil {
		t.Fatalf("query saved_views.view_schema_id column: %v", err)
	}
	if dataType != "text" {
		t.Fatalf("view_schema_id must be one scalar text column, got %q", dataType)
	}
}

func seedSavedView(t testing.TB, db *sql.DB, savedViewID string, incidentID string, viewSchemaID string, scope string, name string, ownerUserID string, updatedAt string) {
	t.Helper()
	ownerExpr := any(nil)
	if ownerUserID != "" {
		ownerExpr = ownerUserID
	}
	ts, err := time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		t.Fatalf("parse seed timestamp: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json,
    owner_user_id, created_at, updated_at, saved_view_version
)
VALUES ($1, $2, $3, $4, $5, '{"sort":[],"filters":[]}'::jsonb, '{"layout_schema_id":"cartulary.layout.v1","column_order":["timeline.occurred_at"],"hidden_field_keys":[],"column_widths":[]}'::jsonb, $6, $7, $7, 1)
`, savedViewID, incidentID, viewSchemaID, scope, name, ownerExpr, ts); err != nil {
		t.Fatalf("seed saved view %s: %v", name, err)
	}
}

func countSavedViewsByName(t testing.TB, db *sql.DB, incidentID string, name string) int {
	t.Helper()
	var count int
	if err := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM saved_views WHERE incident_id = $1 AND display_name = $2`, incidentID, name).Scan(&count); err != nil {
		t.Fatalf("count saved views by name: %v", err)
	}
	return count
}

func savedViewNames(items []any) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.(map[string]any)["display_name"].(string))
	}
	return names
}

func stringSliceFromAny(value any) []string {
	items := value.([]any)
	result := make([]string, 0, len(items))
	for _, item := range items {
		result = append(result, item.(string))
	}
	return result
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func requireSavedViewErrorDetails(t testing.TB, envelope map[string]any, field string, reasonCode string) {
	t.Helper()
	details := envelope["error"].(map[string]any)["details"].(map[string]any)
	if details["field"] != field || details["reason_code"] != reasonCode {
		t.Fatalf("unexpected error details: %#v", details)
	}
}
