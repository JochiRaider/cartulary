package reportcomposition_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCompositionPublicLifecycleAndAuthority_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "composition-lifecycle")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{
		"client_txn_id": "composition-lifecycle-incident", "incident_key": "IR-COMPOSITION", "title": "Composition lifecycle",
	})
	incidentID := incident["incident_id"].(string)
	collection := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/report-compositions"
	type actor struct{ session, csrf *http.Cookie }
	userIDs := map[string]string{}
	actors := map[string]actor{"admin": {admin.SessionCookie, admin.CSRFCookie}}
	for _, role := range []string{"viewer", "editor", "reviewer", "outsider", "deployment_admin"} {
		email := "composition-" + role + "@example.test"
		user := flowtest.SeedLocalUserFlags(t, harness.DB, email, role, "CompositionLifecycle1!", false, role == "deployment_admin", true)
		session, csrf := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, email, "CompositionLifecycle1!", nil)
		actors[role] = actor{session, csrf}
		userIDs[role] = user
		if role != "outsider" && role != "deployment_admin" {
			scenariotest.CreateMembership(t, harness.Server, admin, incidentID, map[string]any{
				"client_txn_id": "composition-membership-" + role, "user_id": user, "role": role,
			})
		}
	}
	do := func(role, method, url string, body any) *http.Response {
		t.Helper()
		a := actors[role]
		return httptestx.DoJSON(t, method, url, body, httptestx.WithCookies(a.session, a.csrf), httptestx.WithHeader(authn.CSRFHeaderName, a.csrf.Value))
	}
	data := func(response *http.Response, status int) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, response, status)["data"].(map[string]any)
	}
	denied := func(response *http.Response, status int, code, validation string) {
		t.Helper()
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		if validation != "" && body["error"].(map[string]any)["details"].(map[string]any)["validation_code"] != validation {
			t.Fatalf("wrong Composition failure: %#v", body)
		}
	}
	create := map[string]any{"client_txn_id": "create", "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion}
	for _, role := range []string{"viewer", "reviewer", "outsider", "deployment_admin"} {
		status, code := http.StatusForbidden, "authorization_denied"
		if role == "outsider" || role == "deployment_admin" {
			status, code = http.StatusNotFound, "incident_not_found"
		}
		denied(do(role, http.MethodPost, collection, create), status, code, "")
	}
	denied(httptestx.DoJSON(t, http.MethodPost, collection, create, httptestx.WithCookies(actors["editor"].session)), http.StatusForbidden, "csrf_verification_failed", "")
	listed := data(do("viewer", http.MethodGet, collection, nil), http.StatusOK)
	if len(listed["composition_resources"].([]any)) != 0 {
		t.Fatal("denied requests created a resource")
	}
	created := data(do("editor", http.MethodPost, collection, create), http.StatusCreated)
	resource := collection + "/" + created["composition_id"].(string)
	if created["draft_version"] != float64(1) || created["latest_composition_version"] != nil || created["retired_at"] != nil {
		t.Fatal("invalid initial lifecycle", created)
	}
	for _, key := range []string{"deck_ops", "diagram_decls", "authored_texts", "release_bound_versions"} {
		if values, ok := created[key].([]any); !ok || len(values) != 0 {
			t.Fatal("missing materialized empty array", key)
		}
	}
	if replay := data(do("editor", http.MethodPost, collection, create), http.StatusCreated); !reflect.DeepEqual(created, replay) {
		t.Fatal("create replay changed the committed resource")
	}
	for _, role := range []string{"viewer", "editor", "reviewer", "admin"} {
		if read := data(do(role, http.MethodGet, resource, nil), http.StatusOK); !reflect.DeepEqual(created, read) {
			t.Fatal("member read changed the resource", role)
		}
	}
	for _, role := range []string{"outsider", "deployment_admin"} {
		denied(do(role, http.MethodGet, resource, nil), http.StatusNotFound, "incident_not_found", "")
		denied(do(role, http.MethodPost, resource+"/validate", map[string]any{}), http.StatusNotFound, "incident_not_found", "")
		denied(do(role, http.MethodPost, resource+"/preview", map[string]any{}), http.StatusNotFound, "incident_not_found", "")
	}
	text := map[string]any{"authored_text_id": "title", "text_role": "title_override", "body": "Review title", "disclosure_partition_ref": "internal_only"}
	update := map[string]any{"client_txn_id": "update", "base_draft_version": 1, "authored_texts": []any{text}}
	for _, role := range []string{"viewer", "reviewer"} {
		denied(do(role, http.MethodPatch, resource, update), http.StatusForbidden, "authorization_denied", "")
		denied(do(role, http.MethodPost, resource+"/versions", map[string]any{"client_txn_id": "freeze-denied", "base_draft_version": 1}), http.StatusForbidden, "authorization_denied", "")
		denied(do(role, http.MethodDelete, resource, map[string]any{"client_txn_id": "retire-denied", "base_draft_version": 1}), http.StatusForbidden, "authorization_denied", "")
	}
	updated := data(do("editor", http.MethodPatch, resource, update), http.StatusOK)
	originalUpdate := updated
	if updated["draft_version"] != float64(2) || !reflect.DeepEqual(updated["authored_texts"], []any{text}) {
		t.Fatal("update did not replace exactly one draft array", updated)
	}
	if replay := data(do("editor", http.MethodPatch, resource, update), http.StatusOK); !reflect.DeepEqual(updated, replay) {
		t.Fatal("update replay revalidated the stale base or changed output")
	}
	denied(do("editor", http.MethodPatch, resource, map[string]any{"client_txn_id": "update", "base_draft_version": 2}), http.StatusConflict, "client_txn_conflict", "")
	denied(do("editor", http.MethodPatch, resource, map[string]any{"client_txn_id": "stale", "base_draft_version": 1}), http.StatusConflict, "conflict", "composition_draft_version_conflict")
	for _, raw := range []string{
		`{"client_txn_id":"invalid","base_draft_version":2,"deck_ops":null}`,
		`{"client_txn_id":"invalid","base_draft_version":2,"base_draft_version":2}`,
		`{"client_txn_id":"invalid","base_draft_version":2.0}`,
		`{"client_txn_id":"invalid","base_draft_version":2,"raw_markdown":"# forbidden"}`,
	} {
		denied(httptestx.DoRawJSON(t, http.MethodPatch, resource, raw, httptestx.WithCookies(actors["editor"].session, actors["editor"].csrf), httptestx.WithHeader(authn.CSRFHeaderName, actors["editor"].csrf.Value)), http.StatusBadRequest, "invalid_request", "composition_schema_invalid")
	}
	if read := data(do("viewer", http.MethodGet, resource, nil), http.StatusOK); !reflect.DeepEqual(updated, read) {
		t.Fatal("rejected mutations changed the resource")
	}
	freeze := map[string]any{"client_txn_id": "freeze", "base_draft_version": 2}
	v1 := data(do("editor", http.MethodPost, resource+"/versions", freeze), http.StatusCreated)
	if v1["composition_version"] != "v1" {
		t.Fatal("version allocation did not start at v1", v1)
	}
	if replay := data(do("editor", http.MethodPost, resource+"/versions", freeze), http.StatusCreated); !reflect.DeepEqual(v1, replay) {
		t.Fatal("freeze replay allocated or changed a version")
	}
	updated = data(do("editor", http.MethodPatch, resource, map[string]any{"client_txn_id": "clear", "base_draft_version": 2, "authored_texts": []any{}}), http.StatusOK)
	if updated["draft_version"] != float64(3) {
		t.Fatal("draft revision skipped or regressed")
	}
	if read := data(do("viewer", http.MethodGet, resource+"/versions/v1", nil), http.StatusOK); !reflect.DeepEqual(v1, read) {
		t.Fatal("newer draft changed immutable version")
	}
	v2 := data(do("editor", http.MethodPost, resource+"/versions", map[string]any{"client_txn_id": "freeze-next", "base_draft_version": 3}), http.StatusCreated)
	if v2["composition_version"] != "v2" || v2["composition_sha256"] == v1["composition_sha256"] {
		t.Fatal("gapless versions lost distinct canonical content")
	}
	beforeValidation := data(do("viewer", http.MethodGet, resource, nil), http.StatusOK)
	retainedCounts := func() [6]int64 {
		t.Helper()
		var counts [6]int64
		err := harness.DB.QueryRowContext(t.Context(), `SELECT
        (SELECT count(*) FROM report_composition_versions),
        (SELECT count(*) FROM report_composition_preview_attempts),
        (SELECT count(*) FROM report_composition_release_bindings),
        (SELECT count(*) FROM jobs),
        (SELECT count(*) FROM reporting_snapshots),
        (SELECT count(*) FROM artifacts)`).Scan(&counts[0], &counts[1], &counts[2], &counts[3], &counts[4], &counts[5])
		if err != nil {
			t.Fatal(err)
		}
		return counts
	}
	beforeCounts := retainedCounts()
	for _, source := range []map[string]any{{}, {"source_kind": "inline", "inline_composition": v1["canonical_composition"]}, {"source_kind": "version", "composition_version": "v1"}} {
		validated := data(do("viewer", http.MethodPost, resource+"/validate", source), http.StatusOK)
		if validated["validation_scope"] != "local_authoring" || validated["valid"] != true || validated["stage"] != nil || len(validated["issues"].([]any)) != 0 {
			t.Fatal("valid local source rejected", validated)
		}
	}
	for _, source := range []map[string]any{{"source_kind": "draft", "composition_version": "v1"}, {"source_kind": "version"}, {"source_kind": "inline"}} {
		denied(do("viewer", http.MethodPost, resource+"/validate", source), http.StatusBadRequest, "invalid_request", "composition_source_invalid")
	}
	for _, retiredInput := range []map[string]any{{"validation_context": nil}, {"validation_context": map[string]any{}}, {"release_scope": "external_release"}} {
		denied(do("viewer", http.MethodPost, resource+"/validate", retiredInput), http.StatusBadRequest, "invalid_request", "composition_schema_invalid")
	}
	for _, invalid := range []struct{ member, value, code, stage string }{
		{"incident_id", "another-incident", "composition_incident_mismatch", "resource_binding_validation"},
		{"composition_id", "another-composition", "composition_id_mismatch", "resource_binding_validation"},
		{"template_id", "another-template", "composition_template_mismatch", "resource_binding_validation"},
		{"composition_sha256", strings.Repeat("0", 64), "composition_digest_mismatch", "canonical_digest_validation"},
	} {
		encoded, err := json.Marshal(v1["canonical_composition"])
		if err != nil {
			t.Fatal(err)
		}
		var inline map[string]any
		if err := json.Unmarshal(encoded, &inline); err != nil {
			t.Fatal(err)
		}
		inline[invalid.member] = invalid.value
		result := data(do("viewer", http.MethodPost, resource+"/validate", map[string]any{"source_kind": "inline", "inline_composition": inline}), http.StatusOK)
		issues := result["issues"].([]any)
		if result["valid"] != false || result["stage"] != invalid.stage || len(issues) != 1 || issues[0].(map[string]any)["code"] != invalid.code {
			t.Fatal("inline binding/digest validation lost first-stage precedence", invalid.member, result)
		}
	}
	if after := data(do("viewer", http.MethodGet, resource, nil), http.StatusOK); !reflect.DeepEqual(beforeValidation, after) || retainedCounts() != beforeCounts {
		t.Fatal("local validation mutated the resource or allocated execution state")
	}
	retire := map[string]any{"client_txn_id": "retire", "base_draft_version": 3}
	retired := data(do("editor", http.MethodDelete, resource, retire), http.StatusOK)
	if retired["retired_at"] == nil || retired["latest_composition_version"] != "v2" {
		t.Fatal("retirement lost audit metadata", retired)
	}
	if replay := data(do("editor", http.MethodDelete, resource, retire), http.StatusOK); !reflect.DeepEqual(retired, replay) {
		t.Fatal("retirement replay changed its timestamp")
	}
	for _, operation := range []struct{ method, suffix string }{{http.MethodPatch, ""}, {http.MethodPost, "/versions"}, {http.MethodDelete, ""}} {
		denied(do("editor", operation.method, resource+operation.suffix, map[string]any{"client_txn_id": "after-retire", "base_draft_version": 3}), http.StatusConflict, "conflict", "composition_resource_retired")
	}
	if read := data(do("viewer", http.MethodGet, resource, nil), http.StatusOK); !reflect.DeepEqual(retired, read) {
		t.Fatal("retired resource is not audit-readable")
	}
	if replay := data(do("editor", http.MethodPatch, resource, update), http.StatusOK); !reflect.DeepEqual(replay, originalUpdate) {
		t.Fatal("exact replay was changed by retirement", replay)
	}
	// A committed replay is not authority: a removed member cannot retrieve it.
	scenariotest.DeleteMembershipVersion(t, harness.Server, admin, incidentID, userIDs["editor"], 1)
	denied(do("editor", http.MethodPatch, resource, update), http.StatusNotFound, "incident_not_found", "")
	denied(do("editor", http.MethodPost, collection, create), http.StatusNotFound, "incident_not_found", "")
	denied(do("editor", http.MethodGet, resource, nil), http.StatusNotFound, "incident_not_found", "")

}
