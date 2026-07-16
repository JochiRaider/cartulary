package viewschemas_test

import (
	"net/http"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestViewSchemasDiscoveryHTTP(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "view-schemas-discovery")
	login, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)

	t.Run("allows any active authenticated deployment user", func(t *testing.T) {
		flowtest.SeedLocalUserFlags(t, harness.DB, "analyst@example.test", "Analyst", "AnalystPass1!", false, false, true)
		sessionCookie, _ := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "analyst@example.test", "AnalystPass1!", nil)

		resp := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas",
			nil,
			httptestx.WithCookies(sessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		wantIDs := currentProfileStandardizedViewSchemaIDs()
		if got := len(body["data"].(map[string]any)["view_schemas"].([]any)); got != len(wantIDs) {
			t.Fatalf("expected active non-admin user to see %d current-profile schemas, got %d", len(wantIDs), got)
		}
	})

	t.Run("lists the exact current-profile standardized registry with default terminal paging", func(t *testing.T) {
		resp := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas",
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		data := body["data"].(map[string]any)
		items := data["view_schemas"].([]any)
		wantIDs := currentProfileStandardizedViewSchemaIDs()
		if len(items) != len(wantIDs) {
			t.Fatalf("expected %d current-profile view schemas, got %d", len(wantIDs), len(items))
		}
		gotIDs := make([]string, 0, len(items))
		for _, item := range items {
			resource := item.(map[string]any)
			gotIDs = append(gotIDs, resource["view_schema_id"].(string))
			requirePublicResource(t, resource)
		}
		if !reflect.DeepEqual(gotIDs, wantIDs) {
			t.Fatalf("unexpected view schema ids:\ngot  %v\nwant %v", gotIDs, wantIDs)
		}
		paging := body["meta"].(map[string]any)["paging"].(map[string]any)
		if paging["limit"] != float64(100) || paging["has_more"] != false || paging["next_cursor"] != nil {
			t.Fatalf("unexpected default paging: %#v", paging)
		}
	})

	t.Run("supports bounded pagination and rejects cursor binding replay", func(t *testing.T) {
		resp := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas?limit=5",
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		items := body["data"].(map[string]any)["view_schemas"].([]any)
		if len(items) != 5 {
			t.Fatalf("expected first page to contain five resources, got %d", len(items))
		}
		paging := body["meta"].(map[string]any)["paging"].(map[string]any)
		cursor, _ := paging["next_cursor"].(string)
		if paging["has_more"] != true || strings.TrimSpace(cursor) == "" {
			t.Fatalf("expected non-terminal first page, got paging %#v", paging)
		}

		next := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas?cursor_token="+cursor,
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		nextBody := httptestx.RequireSuccessEnvelope(t, next, http.StatusOK)
		if got := len(nextBody["data"].(map[string]any)["view_schemas"].([]any)); got != 5 {
			t.Fatalf("expected cursor page to preserve bound limit, got %d", got)
		}

		replay := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+cursor,
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		errBody := httptestx.RequireErrorEnvelope(t, replay, http.StatusBadRequest, "invalid_pagination_request")
		details := errBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "invalid_cursor_token" {
			t.Fatalf("unexpected cursor replay reason: %#v", details)
		}
	})

	t.Run("fails closed for invalid list pagination", func(t *testing.T) {
		for _, target := range []string{
			"/api/v1/view-schemas?page=1",
			"/api/v1/view-schemas?offset=0",
			"/api/v1/view-schemas?page_size=10",
			"/api/v1/view-schemas?block_size=10",
			"/api/v1/view-schemas?limit=0",
			"/api/v1/view-schemas?limit=501",
		} {
			resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+target, nil, httptestx.WithCookies(login.SessionCookie))
			httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
		}
	})

	t.Run("fetches one schema and rejects singleton pagination", func(t *testing.T) {
		resp := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.timeline.v2",
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		resource := body["data"].(map[string]any)
		if resource["view_schema_id"] != "cartulary.view.timeline.v2" {
			t.Fatalf("unexpected singleton resource: %#v", resource)
		}
		requirePublicResource(t, resource)

		paginated := httptestx.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.timeline.v2?limit=1",
			nil,
			httptestx.WithCookies(login.SessionCookie),
		)
		errBody := httptestx.RequireErrorEnvelope(t, paginated, http.StatusBadRequest, "invalid_pagination_request")
		details := errBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "pagination_not_supported" {
			t.Fatalf("unexpected singleton pagination reason: %#v", details)
		}
	})

	t.Run("returns a canonical missing-schema error", func(t *testing.T) {
		for _, viewSchemaID := range []string{
			"cartulary.view.not_real.v1",
			"cartulary.view.hypotheses.v1",
		} {
			resp := httptestx.DoJSON(
				t,
				http.MethodGet,
				harness.Server.HTTP.URL+"/api/v1/view-schemas/"+viewSchemaID,
				nil,
				httptestx.WithCookies(login.SessionCookie),
			)
			httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "view_schema_not_found")
		}
	})
}

var requiredPackIndependentViewSchemaIDs = []string{
	"cartulary.view.assessments.v1",
	"cartulary.view.comm_log.v1",
	"cartulary.view.decisions.v1",
	"cartulary.view.evidence.v1",
	"cartulary.view.handoff.v1",
	"cartulary.view.hosts.v1",
	"cartulary.view.identities.v1",
	"cartulary.view.indicators.v1",
	"cartulary.view.lesson.v1",
	"cartulary.view.notes.v1",
	"cartulary.view.parties.v1",
	"cartulary.view.status_review.v1",
	"cartulary.view.task_requests.v1",
	"cartulary.view.timeline.v2",
}

var implementedStandardizedOptionalViewSchemaIDs = []string{
	"cartulary.view.findings.v1",
	"cartulary.view.forensic_keywords.v1",
	"cartulary.view.investigative_queries.v1",
}

func currentProfileStandardizedViewSchemaIDs() []string {
	ids := make([]string, 0, len(requiredPackIndependentViewSchemaIDs)+len(implementedStandardizedOptionalViewSchemaIDs))
	ids = append(ids, requiredPackIndependentViewSchemaIDs...)
	ids = append(ids, implementedStandardizedOptionalViewSchemaIDs...)
	slices.Sort(ids)
	return ids
}

func requirePublicResource(t testing.TB, resource map[string]any) {
	t.Helper()

	if !reflect.DeepEqual(resource["technical_fields"], []any{"record_id", "row_version"}) {
		t.Fatalf("unexpected technical fields for %s: %#v", resource["view_schema_id"], resource["technical_fields"])
	}
	requireInspectorConfig(t, resource)
	for _, forbidden := range []string{"write_target", "write_action", "base_projection", "canonical_source_filter", "read_model", "create_writable", "writable"} {
		if containsKey(resource, forbidden) {
			t.Fatalf("public resource %s leaked %s: %#v", resource["view_schema_id"], forbidden, resource)
		}
	}
	fields, ok := resource["fields"].([]any)
	if !ok || len(fields) == 0 {
		t.Fatalf("resource %s must expose fields[]", resource["view_schema_id"])
	}
	for _, field := range fields {
		entry := field.(map[string]any)
		for _, key := range []string{"field_key", "label", "default_hidden", "sortable", "header_sort_field_key", "filter_ops", "groupable", "read_kind", "write_kind", "grid_editable", "conflict_resolution_class", "entity_binding_mode", "string_contract_id", "direct_scalar_contract_id", "direct_reference_contract_id", "clearable", "enum_values"} {
			if _, exists := entry[key]; !exists {
				t.Fatalf("field entry for %s missing %s: %#v", resource["view_schema_id"], key, entry)
			}
		}
	}
}

func requireInspectorConfig(t testing.TB, resource map[string]any) {
	t.Helper()

	viewSchemaID, _ := resource["view_schema_id"].(string)
	config, ok := resource["inspector_config"].(map[string]any)
	if !ok {
		t.Fatalf("resource %s must expose inspector_config: %#v", viewSchemaID, resource)
	}
	if config["inspector_config_schema_id"] != "cartulary.inspector_config.v1" {
		t.Fatalf("%s inspector schema id: %#v", viewSchemaID, config)
	}
	if config["view_schema_id"] != viewSchemaID {
		t.Fatalf("%s inspector view_schema_id mismatch: %#v", viewSchemaID, config)
	}
	if config["default_open"] != false {
		t.Fatalf("%s inspector default_open must be false: %#v", viewSchemaID, config)
	}
	subject, ok := config["subject_binding"].(map[string]any)
	if !ok || subject["kind"] != "selected_record" {
		t.Fatalf("%s inspector subject_binding: %#v", viewSchemaID, config["subject_binding"])
	}
	if config["no_row_state"] != "no_row_selected" || config["unsupported_feature_behavior"] != "omit_feature" {
		t.Fatalf("%s inspector fixed values invalid: %#v", viewSchemaID, config)
	}
	panels, ok := config["panels"].([]any)
	if !ok || len(panels) == 0 || len(panels) > 5 {
		t.Fatalf("%s inspector panels bound: %#v", viewSchemaID, config["panels"])
	}
	allowedPanels := map[string]struct{}{"details": {}, "relationships": {}, "evidence": {}, "history": {}, "workflow": {}}
	declaredPanels := map[string]struct{}{}
	for _, panelValue := range panels {
		panel := panelValue.(map[string]any)
		panelID, _ := panel["panel_id"].(string)
		if _, ok := allowedPanels[panelID]; !ok || panel["label"] == "" {
			t.Fatalf("%s inspector panel invalid: %#v", viewSchemaID, panel)
		}
		if _, exists := declaredPanels[panelID]; exists {
			t.Fatalf("%s duplicate inspector panel %s", viewSchemaID, panelID)
		}
		declaredPanels[panelID] = struct{}{}
	}
	groups, ok := config["feature_groups"].([]any)
	if !ok || len(groups) > 64 {
		t.Fatalf("%s inspector feature group bound: %#v", viewSchemaID, config["feature_groups"])
	}
	allowedRoutes := map[string]struct{}{
		"panel_read": {}, "view_row_create": {}, "record_patch": {}, "record_action": {},
		"entity_mention_action": {}, "evidence_access": {}, "surface_pivot": {},
	}
	allowedConditions := map[string]struct{}{
		"no_row_selected": {}, "incident_closed": {}, "authorization_lost": {}, "row_version_changed": {},
		"record_deleted": {}, "record_merged": {}, "evidence_preview_unavailable": {}, "merge_target_unavailable": {},
	}
	featureKeys := map[string]struct{}{}
	for _, groupValue := range groups {
		group := groupValue.(map[string]any)
		key, _ := group["feature_group_key"].(string)
		if key == "" {
			t.Fatalf("%s inspector feature group missing key: %#v", viewSchemaID, group)
		}
		if _, exists := featureKeys[key]; exists {
			t.Fatalf("%s duplicate inspector feature group %s", viewSchemaID, key)
		}
		featureKeys[key] = struct{}{}
		panelID, _ := group["panel_id"].(string)
		if _, ok := declaredPanels[panelID]; !ok {
			t.Fatalf("%s inspector feature group references unknown panel: %#v", viewSchemaID, group)
		}
		route, ok := group["route_binding"].(map[string]any)
		if !ok {
			t.Fatalf("%s inspector feature group missing route binding: %#v", viewSchemaID, group)
		}
		routeKind, _ := route["kind"].(string)
		if _, ok := allowedRoutes[routeKind]; !ok {
			t.Fatalf("%s inspector route kind invalid: %#v", viewSchemaID, route)
		}
		if seedBindings, ok := group["seed_bindings"].([]any); !ok || len(seedBindings) > 16 {
			t.Fatalf("%s inspector seed binding bound: %#v", viewSchemaID, group["seed_bindings"])
		}
		conditions, ok := group["disabled_when"].([]any)
		if !ok || len(conditions) > 16 {
			t.Fatalf("%s inspector disabled_when bound: %#v", viewSchemaID, group["disabled_when"])
		}
		for _, conditionValue := range conditions {
			condition, _ := conditionValue.(string)
			if _, ok := allowedConditions[condition]; !ok {
				t.Fatalf("%s inspector disabled_when invalid: %#v", viewSchemaID, group["disabled_when"])
			}
		}
	}
}

func containsKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		for existing, nested := range typed {
			if existing == key || containsKey(nested, key) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if containsKey(nested, key) {
				return true
			}
		}
	}
	return false
}
