package savedviews_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	gencontracts "github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/phase2test"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
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

	emptyDefaultsResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views",
		map[string]any{
			"view_schema_id": timeline.TimelineViewSchemaID,
			"display_name":   "Empty query defaults",
			"query_json":     map[string]any{},
		},
		phase2test.WithCookies(viewerSession, viewerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
	)
	emptyCreated := httptestx.RequireSuccessEnvelope(t, emptyDefaultsResp, http.StatusCreated)["data"].(map[string]any)
	requireEmptyCanonicalQueryJSON(t, emptyCreated["query_json"].(map[string]any))
	requireDefaultLayoutJSON(t, emptyCreated["layout_json"].(map[string]any))
	requireStoredSavedViewMatchesResource(t, harness.DB, emptyCreated)

	invalidCases := map[string]struct {
		displayName string
		queryJSON   map[string]any
		layoutJSON  map[string]any
		field       string
		reasonCode  string
	}{
		"group by null": {
			displayName: "Invalid group by null",
			queryJSON:   map[string]any{"group_by": nil},
			field:       "query_json.group_by",
			reasonCode:  "invalid_group_by",
		},
		"query sort record id": {
			displayName: "Invalid query sort record id",
			queryJSON: map[string]any{
				"sort": []any{map[string]any{"field_key": "record_id", "direction": "asc"}},
			},
			field:      "query_json.sort[0].field_key",
			reasonCode: "forbidden_field",
		},
		"query filter row version": {
			displayName: "Invalid query filter row version",
			queryJSON: map[string]any{
				"filters": []any{map[string]any{"field_key": "row_version", "op": "eq", "arg": map[string]any{"value": 1}}},
			},
			field:      "query_json.filters[0].field_key",
			reasonCode: "forbidden_field",
		},
		"query group by record id": {
			displayName: "Invalid query group by record id",
			queryJSON:   map[string]any{"group_by": "record_id"},
			field:       "query_json.group_by",
			reasonCode:  "forbidden_field",
		},
		"layout column order record id": {
			displayName: "Invalid layout column order record id",
			queryJSON:   map[string]any{},
			layoutJSON:  map[string]any{"layout_schema_id": "cartulary.layout.v1", "column_order": []any{"record_id"}, "hidden_field_keys": []any{}, "column_widths": []any{}},
			field:       "layout_json.column_order[0]",
			reasonCode:  "forbidden_field",
		},
		"layout hidden row version": {
			displayName: "Invalid layout hidden row version",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["hidden_field_keys"] = []any{"row_version"} }),
			field:       "layout_json.hidden_field_keys[0]",
			reasonCode:  "forbidden_field",
		},
		"layout width record id": {
			displayName: "Invalid layout width record id",
			queryJSON:   map[string]any{},
			layoutJSON: savedViewLayoutWith(t, func(layout map[string]any) {
				layout["column_widths"] = []any{map[string]any{"field_key": "record_id", "width_px": 160}}
			}),
			field:      "layout_json.column_widths[0].field_key",
			reasonCode: "forbidden_field",
		},
		"layout width row version": {
			displayName: "Invalid layout width row version",
			queryJSON:   map[string]any{},
			layoutJSON: savedViewLayoutWith(t, func(layout map[string]any) {
				layout["column_widths"] = []any{map[string]any{"field_key": "row_version", "width_px": 160}}
			}),
			field:      "layout_json.column_widths[0].field_key",
			reasonCode: "forbidden_field",
		},
		"layout inspector open state": {
			displayName: "Invalid layout inspector open state",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["inspector_open"] = true }),
			field:       "layout_json.inspector_open",
			reasonCode:  "unknown_field",
		},
		"layout inspector active panel": {
			displayName: "Invalid layout inspector active panel",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["active_panel"] = "history" }),
			field:       "layout_json.active_panel",
			reasonCode:  "unknown_field",
		},
		"layout inspector preview state": {
			displayName: "Invalid layout inspector preview state",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["preview_state"] = map[string]any{"record_id": "row-1"} }),
			field:       "layout_json.preview_state",
			reasonCode:  "unknown_field",
		},
		"layout inspector local form state": {
			displayName: "Invalid layout inspector local form state",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["local_form_state"] = map[string]any{"dirty": true} }),
			field:       "layout_json.local_form_state",
			reasonCode:  "unknown_field",
		},
		"layout inspector stale confirmation state": {
			displayName: "Invalid layout inspector stale confirmation state",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["stale_confirmation_state"] = map[string]any{"delete": true} }),
			field:       "layout_json.stale_confirmation_state",
			reasonCode:  "unknown_field",
		},
		"layout inspector rollback previews": {
			displayName: "Invalid layout inspector rollback previews",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["rollback_previews"] = []any{"row-1"} }),
			field:       "layout_json.rollback_previews",
			reasonCode:  "unknown_field",
		},
		"layout inspector merge plans": {
			displayName: "Invalid layout inspector merge plans",
			queryJSON:   map[string]any{},
			layoutJSON:  savedViewLayoutWith(t, func(layout map[string]any) { layout["merge_plans"] = []any{"row-1"} }),
			field:       "layout_json.merge_plans",
			reasonCode:  "unknown_field",
		},
	}
	for name, testCase := range invalidCases {
		t.Run(name, func(t *testing.T) {
			body := map[string]any{
				"view_schema_id": timeline.TimelineViewSchemaID,
				"display_name":   testCase.displayName,
				"query_json":     testCase.queryJSON,
			}
			if testCase.layoutJSON != nil {
				body["layout_json"] = testCase.layoutJSON
			}
			resp := phase2test.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views",
				body,
				phase2test.WithCookies(viewerSession, viewerCSRF),
				phase2test.WithHeader(authn.CSRFHeaderName, viewerCSRF.Value),
			)
			errBody := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_mutation_payload")
			requireSavedViewErrorDetails(t, errBody, testCase.field, testCase.reasonCode)
			if got := countSavedViewsByName(t, harness.DB, incidentID, testCase.displayName); got != 0 {
				t.Fatalf("invalid saved-view create must not persist %q, got %d rows", testCase.displayName, got)
			}
		})
	}
}

func TestPhase8_SavedViewSystemFixtureRouteUnavailableByDefault_U_8_02(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	fixtureBody := map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"display_name":   "  Harness system view  ",
		"query_json":     map[string]any{},
		"layout_json":    map[string]any{},
	}

	harness := runtime.StartServer(t, "phase8-savedviews-system-fixture-unregistered")
	unregisteredResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/test/incidents/00000000-0000-0000-0000-000000008901/saved-views/system",
		fixtureBody,
		phase2test.WithHeader(httpapi.TestRouteTokenHeader, httptestx.TestRouteToken),
	)
	httptestx.RequireStatus(t, unregisteredResp, http.StatusNotFound)
	_ = unregisteredResp.Body.Close()
}

func TestPhase8_SavedViewSystemFixtureRoute_U_8_02(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServerWithRoutes(t, "phase8-savedviews-system-fixture", savedviews.RegisterTestRoutes())
	adminLogin, _ := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-system-fixture-incident",
		"incident_key":  "IR-U802-SYSTEM",
		"title":         "Phase 8 system saved-view fixture",
	})
	incidentID := incident["incident_id"].(string)
	route := harness.Server.HTTP.URL + "/api/v1/test/incidents/" + incidentID + "/saved-views/system"
	fixtureBody := map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"display_name":   "  Harness system view  ",
		"query_json":     map[string]any{},
		"layout_json":    map[string]any{},
	}

	missingTokenResp := phase2test.DoJSON(t, http.MethodPost, route, fixtureBody)
	httptestx.RequireErrorEnvelope(t, missingTokenResp, http.StatusForbidden, "test_route_forbidden")

	createResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		route,
		fixtureBody,
		phase2test.WithHeader(httpapi.TestRouteTokenHeader, httptestx.TestRouteToken),
	)
	created := httptestx.RequireSuccessEnvelope(t, createResp, http.StatusCreated)["data"].(map[string]any)
	if created["scope"] != "system" || created["owner_user_id"] != nil || created["display_name"] != "Harness system view" {
		t.Fatalf("unexpected system fixture resource: %#v", created)
	}
	requireEmptyCanonicalQueryJSON(t, created["query_json"].(map[string]any))
	requireDefaultLayoutJSON(t, created["layout_json"].(map[string]any))
	requireStoredSavedViewMatchesResource(t, harness.DB, created)

	visible := visibleSavedViewByName(t, harness.Server.HTTP.URL, incidentID, adminLogin.SessionCookie, "Harness system view")
	if visible["saved_view_id"] != created["saved_view_id"] || visible["scope"] != "system" {
		t.Fatalf("system fixture must be visible through ordinary listing: created=%#v visible=%#v", created, visible)
	}
}

func TestPhase8_SavedViewOpenAPICreateInputIsLenient_U_8_02(t *testing.T) {
	artifact, ok := gencontracts.OpenAPIArtifactsIndex["contracts/openapi/cartulary.openapi.yaml"]
	if !ok {
		t.Fatal("generated OpenAPI artifact missing from internal/gen/contracts")
	}
	var document map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &document); err != nil {
		t.Fatalf("decode generated OpenAPI artifact JSON: %v", err)
	}
	schemas := objectAt(t, objectAt(t, document, "components"), "schemas")

	createRequest := schemaAt(t, schemas, "SavedViewCreateRequest")
	createProps := objectAt(t, createRequest, "properties")
	queryRef := stringAt(t, objectAt(t, createProps, "query_json"), "$ref")
	if queryRef != "#/components/schemas/SavedViewCreateQueryJSON" {
		t.Fatalf("create query_json must use create-input schema, got %q", queryRef)
	}
	layoutRef := stringAt(t, objectAt(t, createProps, "layout_json"), "$ref")
	if layoutRef != "#/components/schemas/SavedViewCreateLayoutJSON" {
		t.Fatalf("create layout_json must use create-input schema, got %q", layoutRef)
	}

	createQuery := schemaAt(t, schemas, "SavedViewCreateQueryJSON")
	if required, ok := createQuery["required"]; ok {
		t.Fatalf("create query_json must not require sort or filters, got required=%#v", required)
	}
	createQueryProps := objectAt(t, createQuery, "properties")
	if _, ok := createQueryProps["sort"]; !ok {
		t.Fatal("create query_json must still declare optional sort")
	}
	if _, ok := createQueryProps["filters"]; !ok {
		t.Fatal("create query_json must still declare optional filters")
	}
	groupBy := objectAt(t, createQueryProps, "group_by")
	if stringAt(t, groupBy, "type") != "string" {
		t.Fatalf("create group_by must be string-only and non-nullable, got %#v", groupBy)
	}

	resourceQuery := schemaAt(t, schemas, "SavedViewQueryJSON")
	requireStringSet(t, resourceQuery["required"], []string{"sort", "filters"})

	createLayout := schemaAt(t, schemas, "SavedViewCreateLayoutJSON")
	variants, ok := createLayout["oneOf"].([]any)
	if !ok || len(variants) != 2 {
		t.Fatalf("create layout_json must allow empty object or canonical layout, got %#v", createLayout)
	}
	emptyVariant, ok := variants[0].(map[string]any)
	if !ok || emptyVariant["maxProperties"] != float64(0) {
		t.Fatalf("create layout_json first variant must be {}, got %#v", variants[0])
	}
	fullVariant, ok := variants[1].(map[string]any)
	if !ok || fullVariant["$ref"] != "#/components/schemas/SavedViewLayoutJSON" {
		t.Fatalf("create layout_json second variant must be canonical layout ref, got %#v", variants[1])
	}

	resourceLayout := schemaAt(t, schemas, "SavedViewLayoutJSON")
	requireStringSet(t, resourceLayout["required"], []string{"layout_schema_id", "column_order", "hidden_field_keys", "column_widths"})
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
		"null scope":    `{"view_schema_id":"cartulary.view.timeline.v2","display_name":"View","query_json":{},"scope":null}`,
		"obsolete team": `{"view_schema_id":"cartulary.view.timeline.v2","display_name":"View","query_json":{},"scope":"team"}`,
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
	harness := phase2storetest.StartStore(t, "phase8-savedviews-u-8-04")
	admin := phase2storetest.SeedLocalUserRecord(t, harness.DB, "phase8-u804-admin@example.test", "Phase8 U804 Admin", "Phase8U804Admin1!", false, false, true)
	incident := phase2storetest.CreateIncidentInStore(t, harness.DB, admin, incidents.CreateIncidentRequest{
		ClientTxnID: "txn-phase8-u-8-04-incident",
		IncidentKey: "IR-U804",
		Title:       "Phase 8 saved-view patch",
	})
	incidentID := incident.Incident.ID
	owner := phase2storetest.SeedLocalUserRecord(t, harness.DB, "phase8-u804-owner@example.test", "Phase8 U804 Owner", "Phase8U804Owner1!", false, false, true)
	phase2storetest.CreateMembershipInStore(t, harness.DB, admin, incidentID, owner, incidents.MembershipCreateRequest{
		ClientTxnID: "txn-phase8-u-8-04-owner-membership",
		UserID:      &owner.ID,
		Role:        "viewer",
	})

	createRequest, apiErr := savedviews.DecodeCreateRequest(strings.NewReader(`{
		"view_schema_id":"cartulary.view.timeline.v2",
		"display_name":"Analyst triage",
		"query_json":{},
		"layout_json":{}
	}`))
	if apiErr != nil {
		t.Fatalf("decode create request: %#v", apiErr)
	}
	store := savedviews.NewStore(harness.DB)
	createdAt := time.Date(2026, 5, 14, 14, 0, 0, 0, time.UTC)
	created, err := store.Create(context.Background(), owner, incidentID, createRequest, createdAt)
	if err != nil {
		t.Fatalf("create saved view through store: %v", err)
	}

	patchRequest := decodeSavedViewPatchMap(t, created.ViewSchemaID, map[string]any{
		"base_saved_view_version": 1,
		"display_name":            "  Analyst triage shared  ",
		"scope":                   "shared",
		"query_json": map[string]any{
			"filters": []any{map[string]any{"field_key": "timeline.tags", "op": "contains_any", "arg": map[string]any{"values": []any{"beta", "alpha", "alpha"}}}},
			"sort":    []any{},
		},
		"layout_json": savedViewLayoutWith(t, func(layout map[string]any) {
			layout["column_widths"] = []any{map[string]any{"field_key": "timeline.activity_synopsis_text", "width_px": 240}}
		}),
	})
	patchedAt := createdAt.Add(5 * time.Minute)
	patched, err := store.Patch(context.Background(), owner, "viewer", incidentID, created.SavedViewID, patchRequest, patchedAt)
	if err != nil {
		t.Fatalf("patch saved view through store: %v", err)
	}
	if patched.DisplayName != "Analyst triage shared" || patched.Scope != savedviews.ScopeShared {
		t.Fatalf("patch did not normalize mutable fields: %#v", patched)
	}
	if patched.SavedViewVersion != created.SavedViewVersion+1 || !patched.UpdatedAt.Equal(patchedAt) {
		t.Fatalf("material patch must advance version and updated_at once: before=%#v after=%#v", created, patched)
	}
	var patchedQuery map[string]any
	if err := json.Unmarshal(patched.QueryJSON, &patchedQuery); err != nil {
		t.Fatalf("decode patched query: %v", err)
	}
	requireCanonicalQueryJSON(t, patchedQuery, nil)

	noOpRequest := decodeSavedViewPatchMap(t, patched.ViewSchemaID, map[string]any{
		"base_saved_view_version": 2,
		"display_name":            "Analyst triage shared",
		"scope":                   "shared",
		"query_json": map[string]any{
			"sort":    []any{},
			"filters": []any{map[string]any{"field_key": "timeline.tags", "op": "contains_any", "arg": map[string]any{"values": []any{"alpha", "beta"}}}},
		},
		"layout_json": savedViewLayoutWith(t, func(layout map[string]any) {
			layout["column_widths"] = []any{map[string]any{"field_key": "timeline.activity_synopsis_text", "width_px": 240}}
		}),
	})
	noOp, err := store.Patch(context.Background(), owner, "viewer", incidentID, patched.SavedViewID, noOpRequest, patchedAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("no-op patch through store: %v", err)
	}
	if noOp.SavedViewVersion != patched.SavedViewVersion || !noOp.UpdatedAt.Equal(patched.UpdatedAt) {
		t.Fatalf("structural no-op advanced version or timestamp: before=%#v after=%#v", patched, noOp)
	}

	staleRequest := noOpRequest
	staleRequest.BaseSavedViewVersion = 1
	_, err = store.Patch(context.Background(), owner, "viewer", incidentID, patched.SavedViewID, staleRequest, patchedAt.Add(2*time.Hour))
	var conflict *savedviews.SavedViewVersionConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("stale saved-view patch got %v want SavedViewVersionConflictError", err)
	}
	if conflict.BaseSavedViewVersion != 1 || conflict.CurrentSavedViewVersion != 2 {
		t.Fatalf("unexpected saved-view conflict details: %#v", conflict)
	}

	for name, body := range map[string]string{
		"incident id":    `{"base_saved_view_version":2,"incident_id":"` + incidentID.String() + `"}`,
		"saved view id":  `{"base_saved_view_version":2,"saved_view_id":"` + created.SavedViewID.String() + `"}`,
		"view schema id": `{"base_saved_view_version":2,"view_schema_id":"` + timeline.TimelineViewSchemaID + `"}`,
		"owner":          `{"base_saved_view_version":2,"owner_user_id":"` + owner.ID.String() + `"}`,
		"created at":     `{"base_saved_view_version":2,"created_at":"2026-05-14T15:00:00Z"}`,
		"updated at":     `{"base_saved_view_version":2,"updated_at":"2026-05-14T15:00:00Z"}`,
		"version":        `{"base_saved_view_version":2,"saved_view_version":3}`,
		"unknown":        `{"base_saved_view_version":2,"bogus":true}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, apiErr := savedviews.DecodePatchRequest(strings.NewReader(body), timeline.TimelineViewSchemaID)
			if apiErr == nil || apiErr.Code != "invalid_mutation_payload" {
				t.Fatalf("expected invalid_mutation_payload for %s, got %#v", name, apiErr)
			}
		})
	}
}

func TestPhase8_SavedViewLifecyclePersistence_I_8_01(t *testing.T) {
	runtime := phase2test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase8-savedviews-i-8-01")
	adminLogin, adminID := phase2test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase2test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase8-i-8-01-incident",
		"incident_key":  "IR-I801",
		"title":         "Phase 8 saved-view lifecycle",
	})
	incidentID := incident["incident_id"].(string)
	incidentUUID := phase4test.MustUUID(t, incidentID)
	adminUUID := phase4test.MustUUID(t, adminID)

	ownerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-i801-owner@example.test", "Phase8 I801 Owner", "Phase8I801Owner1!", false, false, true)
	peerID := phase2test.SeedLocalUserFlags(t, harness.DB, "phase8-i801-peer@example.test", "Phase8 I801 Peer", "Phase8I801Peer1!", false, false, true)
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-01-owner-membership", "user_id": ownerID, "role": "viewer"})
	phase2test.CreateMembership(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-01-peer-membership", "user_id": peerID, "role": "viewer"})
	ownerSession, ownerCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-i801-owner@example.test", "Phase8I801Owner1!")
	peerSession, peerCSRF := phase2test.LoginLocalUser(t, harness.Server, "phase8-i801-peer@example.test", "Phase8I801Peer1!")

	timelineOne := phase2test.CreateTimelineRow(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-01-row-one", "timeline.activity_synopsis_text": "Saved-view delete keeps records"})
	timelineTwo := phase2test.CreateTimelineRow(t, harness.Server, adminLogin, incidentID, map[string]any{"client_txn_id": "txn-phase8-i-8-01-row-two", "timeline.activity_synopsis_text": "Saved-view delete keeps linked records"})
	recordOneID := phase4test.MustUUID(t, timelineOne["row"].(map[string]any)["record_id"].(string))
	recordTwoID := phase4test.MustUUID(t, timelineTwo["row"].(map[string]any)["record_id"].(string))
	phase4test.SeedRecordLink(t, harness.DB, incidentUUID, adminUUID, uuid.MustParse("00000000-0000-0000-0000-000000008151"), recordOneID, recordTwoID, "references_record", "manual", nil)
	phase4test.SeedRecordTag(t, harness.DB, incidentUUID, adminUUID, uuid.MustParse("00000000-0000-0000-0000-000000008152"), recordOneID, "sprint3")
	evidenceResp := phase2test.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.evidence.v1/rows",
		map[string]any{"client_txn_id": "txn-phase8-i-8-01-evidence", "evidence.title": "Saved-view delete keeps evidence"},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireSuccessEnvelope(t, evidenceResp, http.StatusCreated)

	beforeCounts := savedViewUnderlyingCounts(t, harness.DB, incidentID)
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 14, 15, 0, 0, 0, time.UTC))
	created := createSavedViewHTTP(t, harness.Server.HTTP.URL, incidentID, ownerSession, ownerCSRF, map[string]any{
		"view_schema_id": timeline.TimelineViewSchemaID,
		"display_name":   "  Owner private  ",
		"query_json": map[string]any{
			"filters": []any{map[string]any{"field_key": "timeline.tags", "op": "contains_any", "arg": map[string]any{"values": []any{"beta", "alpha", "alpha"}}}},
		},
		"layout_json": map[string]any{},
	})
	requireCanonicalQueryJSON(t, created["query_json"].(map[string]any), nil)
	requireDefaultLayoutJSON(t, created["layout_json"].(map[string]any))

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 14, 15, 5, 0, 0, time.UTC))
	patched := patchSavedViewHTTP(t, harness.Server.HTTP.URL, incidentID, created["saved_view_id"].(string), ownerSession, ownerCSRF, map[string]any{
		"base_saved_view_version": int64FromResource(t, created, "saved_view_version"),
		"display_name":            " Shared triage ",
		"scope":                   "shared",
		"query_json": map[string]any{
			"sort":    []any{},
			"filters": []any{map[string]any{"field_key": "timeline.tags", "op": "contains_any", "arg": map[string]any{"values": []any{"beta", "alpha", "alpha"}}}},
		},
		"layout_json": savedViewLayoutWith(t, func(layout map[string]any) {
			layout["column_widths"] = []any{map[string]any{"field_key": "timeline.activity_synopsis_text", "width_px": 240}}
		}),
	})
	if patched["display_name"] != "Shared triage" || patched["scope"] != "shared" {
		t.Fatalf("patch did not persist normalized public fields: %#v", patched)
	}
	if int64FromResource(t, patched, "saved_view_version") != int64FromResource(t, created, "saved_view_version")+1 {
		t.Fatalf("patch did not advance saved_view_version exactly once: before=%#v after=%#v", created, patched)
	}
	if patched["updated_at"] == created["updated_at"] {
		t.Fatalf("material patch did not refresh updated_at: before=%#v after=%#v", created, patched)
	}
	requireStoredSavedViewMatchesResource(t, harness.DB, patched)

	noOp := patchSavedViewHTTP(t, harness.Server.HTTP.URL, incidentID, created["saved_view_id"].(string), ownerSession, ownerCSRF, map[string]any{
		"base_saved_view_version": int64FromResource(t, patched, "saved_view_version"),
		"display_name":            "Shared triage",
		"scope":                   "shared",
		"query_json": map[string]any{
			"filters": []any{map[string]any{"field_key": "timeline.tags", "op": "contains_any", "arg": map[string]any{"values": []any{"alpha", "beta"}}}},
			"sort":    []any{},
		},
		"layout_json": patched["layout_json"],
	})
	if int64FromResource(t, noOp, "saved_view_version") != int64FromResource(t, patched, "saved_view_version") || noOp["updated_at"] != patched["updated_at"] {
		t.Fatalf("route no-op advanced saved_view_version or updated_at: before=%#v after=%#v", patched, noOp)
	}

	staleResp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+created["saved_view_id"].(string),
		map[string]any{"base_saved_view_version": 1, "display_name": "stale overwrite"},
		phase2test.WithCookies(ownerSession, ownerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, ownerCSRF.Value),
	)
	staleBody := httptestx.RequireErrorEnvelope(t, staleResp, http.StatusConflict, "saved_view_version_conflict")
	staleDetails := staleBody["error"].(map[string]any)["details"].(map[string]any)
	if staleDetails["base_saved_view_version"] != float64(1) || staleDetails["current_saved_view_version"] != float64(2) {
		t.Fatalf("unexpected stale conflict details: %#v", staleDetails)
	}

	peerPatch := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+created["saved_view_id"].(string),
		map[string]any{"base_saved_view_version": int64FromResource(t, patched, "saved_view_version"), "display_name": "peer overwrite"},
		phase2test.WithCookies(peerSession, peerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, peerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, peerPatch, http.StatusForbidden, "authorization_denied")

	duplicate := createSavedViewHTTP(t, harness.Server.HTTP.URL, incidentID, peerSession, peerCSRF, map[string]any{
		"view_schema_id": patched["view_schema_id"],
		"display_name":   "Peer duplicate from visible shared",
		"scope":          "private",
		"query_json":     patched["query_json"],
		"layout_json":    patched["layout_json"],
	})
	if duplicate["saved_view_id"] == patched["saved_view_id"] || duplicate["owner_user_id"] != peerID {
		t.Fatalf("duplicate must be a new ordinary saved view owned by caller: source=%#v duplicate=%#v", patched, duplicate)
	}
	if !reflect.DeepEqual(duplicate["query_json"], patched["query_json"]) || !reflect.DeepEqual(duplicate["layout_json"], patched["layout_json"]) {
		t.Fatalf("duplicate must persist normalized source query/layout copy:\nsource=%#v\nduplicate=%#v", patched, duplicate)
	}

	systemID := "00000000-0000-0000-0000-000000008153"
	seedSavedView(t, harness.DB, systemID, incidentID, timeline.TimelineViewSchemaID, "system", "System visible source", "", "2026-05-14T15:10:00Z")
	systemPatch := phase2test.DoJSON(
		t,
		http.MethodPatch,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+systemID,
		map[string]any{"base_saved_view_version": 1, "display_name": "system overwrite"},
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, systemPatch, http.StatusForbidden, "authorization_denied")
	systemResource := visibleSavedViewByName(t, harness.Server.HTTP.URL, incidentID, peerSession, "System visible source")
	systemDuplicate := createSavedViewHTTP(t, harness.Server.HTTP.URL, incidentID, peerSession, peerCSRF, map[string]any{
		"view_schema_id": systemResource["view_schema_id"],
		"display_name":   "Peer duplicate from visible system",
		"query_json":     systemResource["query_json"],
		"layout_json":    systemResource["layout_json"],
	})
	if systemDuplicate["scope"] != "private" || systemDuplicate["saved_view_id"] == systemID {
		t.Fatalf("system duplicate must be a new ordinary private saved view: %#v", systemDuplicate)
	}

	peerDelete := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+created["saved_view_id"].(string),
		nil,
		phase2test.WithCookies(peerSession, peerCSRF),
		phase2test.WithHeader(authn.CSRFHeaderName, peerCSRF.Value),
	)
	httptestx.RequireErrorEnvelope(t, peerDelete, http.StatusForbidden, "authorization_denied")
	systemDelete := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+systemID,
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	httptestx.RequireErrorEnvelope(t, systemDelete, http.StatusForbidden, "authorization_denied")

	deleteResp := phase2test.DoJSON(
		t,
		http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/saved-views/"+created["saved_view_id"].(string),
		nil,
		phase2test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase2test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	)
	deleteData := httptestx.RequireSuccessEnvelope(t, deleteResp, http.StatusOK)["data"].(map[string]any)
	if deleteData["saved_view_id"] != created["saved_view_id"] || deleteData["deleted"] != true {
		t.Fatalf("unexpected delete response: %#v", deleteData)
	}
	if got := phase2test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM saved_views WHERE saved_view_id::text = $1`, created["saved_view_id"]); got != 0 {
		t.Fatalf("delete must remove only the saved-view configuration row, got %d rows", got)
	}
	requireSavedViewUnderlyingCounts(t, harness.DB, incidentID, beforeCounts)
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

func requireEmptyCanonicalQueryJSON(t testing.TB, query map[string]any) {
	t.Helper()
	sortEntries, ok := query["sort"].([]any)
	if !ok || len(sortEntries) != 0 {
		t.Fatalf("query_json.sort must normalize to [], got %#v", query)
	}
	filters, ok := query["filters"].([]any)
	if !ok || len(filters) != 0 {
		t.Fatalf("query_json.filters must normalize to [], got %#v", query)
	}
	if _, exists := query["group_by"]; exists {
		t.Fatalf("inactive group_by must be omitted, got %#v", query)
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
	var storedQuery any
	if err := json.Unmarshal(queryJSON, &storedQuery); err != nil {
		t.Fatalf("decode stored query_json: %v", err)
	}
	if !reflect.DeepEqual(storedQuery, resource["query_json"]) {
		t.Fatalf("stored query_json must match resource:\nstored=%#v\nresource=%#v", storedQuery, resource["query_json"])
	}
	var storedLayout any
	if err := json.Unmarshal(layoutJSON, &storedLayout); err != nil {
		t.Fatalf("decode stored layout_json: %v", err)
	}
	if !reflect.DeepEqual(storedLayout, resource["layout_json"]) {
		t.Fatalf("stored layout_json must match resource:\nstored=%#v\nresource=%#v", storedLayout, resource["layout_json"])
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
	layoutJSON, layoutErr := viewschema.DefaultLayout(viewSchemaID)
	if layoutErr != nil {
		t.Fatalf("build seed saved-view layout: %+v", layoutErr)
	}
	if _, err := db.ExecContext(context.Background(), `
INSERT INTO saved_views (
    saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json,
    owner_user_id, created_at, updated_at, saved_view_version
)
VALUES ($1, $2, $3, $4, $5, '{"sort":[],"filters":[]}'::jsonb, $6::jsonb, $7, $8, $8, 1)
`, savedViewID, incidentID, viewSchemaID, scope, name, layoutJSON, ownerExpr, ts); err != nil {
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

func decodeSavedViewPatchMap(t testing.TB, viewSchemaID string, body map[string]any) savedviews.PatchRequest {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal saved-view patch body: %v", err)
	}
	request, apiErr := savedviews.DecodePatchRequest(strings.NewReader(string(payload)), viewSchemaID)
	if apiErr != nil {
		t.Fatalf("decode saved-view patch body: %#v", apiErr)
	}
	return request
}

func createSavedViewHTTP(t testing.TB, baseURL string, incidentID string, session *http.Cookie, csrf *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPost,
		baseURL+"/api/v1/incidents/"+incidentID+"/saved-views",
		body,
		phase2test.WithCookies(session, csrf),
		phase2test.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func patchSavedViewHTTP(t testing.TB, baseURL string, incidentID string, savedViewID string, session *http.Cookie, csrf *http.Cookie, body map[string]any) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodPatch,
		baseURL+"/api/v1/incidents/"+incidentID+"/saved-views/"+savedViewID,
		body,
		phase2test.WithCookies(session, csrf),
		phase2test.WithHeader(authn.CSRFHeaderName, csrf.Value),
	)
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func visibleSavedViewByName(t testing.TB, baseURL string, incidentID string, session *http.Cookie, name string) map[string]any {
	t.Helper()
	resp := phase2test.DoJSON(
		t,
		http.MethodGet,
		baseURL+"/api/v1/incidents/"+incidentID+"/saved-views?limit=100",
		nil,
		phase2test.WithCookies(session),
	)
	body := httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	for _, item := range body["data"].(map[string]any)["saved_views"].([]any) {
		resource := item.(map[string]any)
		if resource["display_name"] == name {
			return resource
		}
	}
	t.Fatalf("saved view %q not visible in list %#v", name, body)
	return nil
}

func int64FromResource(t testing.TB, resource map[string]any, key string) int64 {
	t.Helper()
	value, ok := resource[key].(float64)
	if !ok {
		t.Fatalf("resource[%q] is %T, want JSON number", key, resource[key])
	}
	return int64(value)
}

type savedViewUnderlyingRowCounts struct {
	Records                     int
	TimelineProjection          int
	RecordLinks                 int
	RecordTags                  int
	Evidence                    int
	IncidentWorkbookPreferences int
	UserWorkbookPreferences     int
}

func savedViewUnderlyingCounts(t testing.TB, db *sql.DB, incidentID string) savedViewUnderlyingRowCounts {
	t.Helper()
	return savedViewUnderlyingRowCounts{
		Records:                     phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM records WHERE incident_id::text = $1`, incidentID),
		TimelineProjection:          phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id::text = $1`, incidentID),
		RecordLinks:                 phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM record_links WHERE incident_id::text = $1`, incidentID),
		RecordTags:                  phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM record_tags WHERE incident_id::text = $1`, incidentID),
		Evidence:                    phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM evidence WHERE incident_id::text = $1`, incidentID),
		IncidentWorkbookPreferences: phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM incident_workbook_preferences WHERE incident_id::text = $1`, incidentID),
		UserWorkbookPreferences:     phase2test.QueryCount(t, db, `SELECT COUNT(*) FROM user_workbook_preferences WHERE incident_id::text = $1`, incidentID),
	}
}

func requireSavedViewUnderlyingCounts(t testing.TB, db *sql.DB, incidentID string, want savedViewUnderlyingRowCounts) {
	t.Helper()
	got := savedViewUnderlyingCounts(t, db, incidentID)
	if got != want {
		t.Fatalf("saved-view delete changed underlying workbook/record/evidence/link/tag data: got %+v want %+v", got, want)
	}
}

func requireSavedViewErrorDetails(t testing.TB, envelope map[string]any, field string, reasonCode string) {
	t.Helper()
	details := envelope["error"].(map[string]any)["details"].(map[string]any)
	if details["field"] != field || details["reason_code"] != reasonCode {
		t.Fatalf("unexpected error details: %#v", details)
	}
}

func savedViewLayoutWith(t testing.TB, mutate func(map[string]any)) map[string]any {
	t.Helper()
	raw, layoutErr := viewschema.DefaultLayout(timeline.TimelineViewSchemaID)
	if layoutErr != nil {
		t.Fatalf("build default layout: %+v", layoutErr)
	}
	var layout map[string]any
	if err := json.Unmarshal(raw, &layout); err != nil {
		t.Fatalf("decode default layout: %v", err)
	}
	mutate(layout)
	return layout
}

func schemaAt(t testing.TB, schemas map[string]any, name string) map[string]any {
	t.Helper()
	value, ok := schemas[name]
	if !ok {
		t.Fatalf("OpenAPI schema %q missing", name)
	}
	typed, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("OpenAPI schema %q is %T, want object", name, value)
	}
	return typed
}

func objectAt(t testing.TB, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := any(root)
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("path %v: parent for %q is %T, want object", path, key, current)
		}
		value, ok := object[key]
		if !ok {
			t.Fatalf("path %v missing key %q", path, key)
		}
		current = value
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("path %v is %T, want object", path, current)
	}
	return object
}

func stringAt(t testing.TB, root map[string]any, key string) string {
	t.Helper()
	value, ok := root[key].(string)
	if !ok {
		t.Fatalf("key %q is %T, want string", key, root[key])
	}
	return value
}

func requireStringSet(t testing.TB, value any, want []string) {
	t.Helper()
	items, ok := value.([]any)
	if !ok {
		t.Fatalf("required must be array, got %#v", value)
	}
	got := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			t.Fatalf("required item is %T, want string", item)
		}
		got = append(got, text)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected required fields: got %v want %v", got, want)
	}
}
