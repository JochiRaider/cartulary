package viewschemas_test

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
)

func TestViewSchemasDiscoveryHTTP(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "view-schemas-discovery")
	login, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)

	t.Run("allows any active authenticated deployment user", func(t *testing.T) {
		phase2test.SeedLocalUserFlags(t, harness.DB, "analyst@example.test", "Analyst", "AnalystPass1!", false, false, true)
		sessionCookie, _ := phase2test.LoginLocalUser(t, harness.Server, "analyst@example.test", "AnalystPass1!")

		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas",
			nil,
			phase2test.WithCookies(sessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		if got := len(body["data"].(map[string]any)["view_schemas"].([]any)); got != 14 {
			t.Fatalf("expected active non-admin user to see fourteen schemas, got %d", got)
		}
	})

	t.Run("lists the exact Base registry with default terminal paging", func(t *testing.T) {
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas",
			nil,
			phase2test.WithCookies(login.SessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		data := body["data"].(map[string]any)
		items := data["view_schemas"].([]any)
		if len(items) != 14 {
			t.Fatalf("expected fourteen Base view schemas, got %d", len(items))
		}
		gotIDs := make([]string, 0, len(items))
		for _, item := range items {
			resource := item.(map[string]any)
			gotIDs = append(gotIDs, resource["view_schema_id"].(string))
			requirePublicResource(t, resource)
		}
		wantIDs := []string{
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
			"cartulary.view.timeline.v1",
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
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas?limit=5",
			nil,
			phase2test.WithCookies(login.SessionCookie),
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

		next := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas?cursor_token="+cursor,
			nil,
			phase2test.WithCookies(login.SessionCookie),
		)
		nextBody := httptestx.RequireSuccessEnvelope(t, next, http.StatusOK)
		if got := len(nextBody["data"].(map[string]any)["view_schemas"].([]any)); got != 5 {
			t.Fatalf("expected cursor page to preserve bound limit, got %d", got)
		}

		replay := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/incidents?cursor_token="+cursor,
			nil,
			phase2test.WithCookies(login.SessionCookie),
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
			resp := phase2test.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+target, nil, phase2test.WithCookies(login.SessionCookie))
			httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
		}
	})

	t.Run("fetches one schema and rejects singleton pagination", func(t *testing.T) {
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.timeline.v1",
			nil,
			phase2test.WithCookies(login.SessionCookie),
		)
		body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		resource := body["data"].(map[string]any)
		if resource["view_schema_id"] != "cartulary.view.timeline.v1" {
			t.Fatalf("unexpected singleton resource: %#v", resource)
		}
		requirePublicResource(t, resource)

		paginated := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.timeline.v1?limit=1",
			nil,
			phase2test.WithCookies(login.SessionCookie),
		)
		errBody := httptestx.RequireErrorEnvelope(t, paginated, http.StatusBadRequest, "invalid_pagination_request")
		details := errBody["error"].(map[string]any)["details"].(map[string]any)
		if details["reason_code"] != "pagination_not_supported" {
			t.Fatalf("unexpected singleton pagination reason: %#v", details)
		}
	})

	t.Run("returns a canonical missing-schema error", func(t *testing.T) {
		resp := phase2test.DoJSON(
			t,
			http.MethodGet,
			harness.Server.HTTP.URL+"/api/v1/view-schemas/cartulary.view.not_real.v1",
			nil,
			phase2test.WithCookies(login.SessionCookie),
		)
		httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "view_schema_not_found")
	})
}

func requirePublicResource(t testing.TB, resource map[string]any) {
	t.Helper()

	if !reflect.DeepEqual(resource["technical_fields"], []any{"record_id", "row_version"}) {
		t.Fatalf("unexpected technical fields for %s: %#v", resource["view_schema_id"], resource["technical_fields"])
	}
	for _, forbidden := range []string{"write_target", "write_action", "base_projection", "read_model", "create_writable", "writable"} {
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
		for _, key := range []string{"field_key", "label", "default_hidden", "sortable", "header_sort_field_key", "filter_ops", "groupable", "read_kind", "write_kind", "conflict_resolution_class", "entity_binding_mode", "string_contract_id", "direct_scalar_contract_id", "direct_reference_contract_id", "clearable", "enum_values"} {
			if _, exists := entry[key]; !exists {
				t.Fatalf("field entry for %s missing %s: %#v", resource["view_schema_id"], key, entry)
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
