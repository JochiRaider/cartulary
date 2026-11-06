package imports_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	evidenceflow "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/irfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

type irPublicClient struct {
	t                *testing.T
	origin, incident string
	login            authcookietest.AuthCookies
}

func (c irPublicClient) request(method, path string, body any, status int) map[string]any {
	c.t.Helper()
	response := doImportJSON(c.t, c.origin, c.login, method, path, body)
	if response.StatusCode != status {
		txn := ""
		if payload, ok := body.(map[string]any); ok {
			txn, _ = payload["client_txn_id"].(string)
		}
		c.t.Fatalf("%s %s (transaction %s): status %d want %d: %#v", method, path, txn, response.StatusCode, status, httptestx.ReadJSONBody(c.t, response))
	}
	return httptestx.RequireSuccessEnvelope(c.t, response, status)["data"].(map[string]any)
}
func (c irPublicClient) create(view, txn string, fields map[string]any) map[string]any {
	c.t.Helper()
	fields["client_txn_id"] = txn
	return c.request(http.MethodPost, "/api/v1/incidents/"+c.incident+"/views/"+view+"/rows", fields, http.StatusCreated)["row"].(map[string]any)
}
func (c irPublicClient) patch(view, txn string, row map[string]any, changes ...map[string]any) map[string]any {
	c.t.Helper()
	return c.request(http.MethodPatch, "/api/v1/records/"+row["record_id"].(string), map[string]any{"view_schema_id": view, "client_txn_id": txn, "base_row_version": row["row_version"], "changes": changes}, http.StatusOK)["row"].(map[string]any)
}
func irLinks(ids ...string) map[string]any {
	actions := []map[string]any{}
	for _, id := range ids {
		actions = append(actions, map[string]any{"op": "add_record_ref", "linked_record_id": id})
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}
func irValue(field string, value any) map[string]any {
	return map[string]any{"field_key": field, "value": value}
}

type irInvestigation struct {
	evidenceID, noteID, indicatorID, savedViewID string
	replacement                                  []byte
	views                                        []string
}

func irCreateInvestigation(c irPublicClient, corpus *irfixtures.Corpus) irInvestigation {
	t := c.t
	t.Helper()
	raw, err := corpus.Read("ir-incident/v1/supplement/records.json")
	if err != nil {
		t.Fatal(err)
	}
	var supplement struct {
		SchemaID  string   `json:"schema_id"`
		Synthetic bool     `json:"synthetic"`
		Label     string   `json:"label"`
		Mentions  []string `json:"mentions"`
		Evidence  []string `json:"evidence"`
		Pending   string   `json:"pending_request"`
		Note      string   `json:"note"`
		Tasks     []struct {
			ID, Title string
			DependsOn []string `json:"depends_on"`
		} `json:"tasks"`
		Decision    string                       `json:"decision"`
		Assessment  string                       `json:"assessment"`
		Indicator   struct{ Kind, Value string } `json:"indicator"`
		Party       string                       `json:"party"`
		SavedView   string                       `json:"saved_view"`
		Composition string                       `json:"report_composition"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&supplement); err != nil {
		t.Fatal(err)
	}
	if !supplement.Synthetic || len(supplement.Evidence) != 2 || len(supplement.Tasks) != 2 || len(supplement.Mentions) != 2 {
		t.Fatal("incompatible synthetic supplement")
	}
	state := irInvestigation{views: []string{"cartulary.view.evidence.v1", "cartulary.view.notes.v1", "cartulary.view.task_requests.v1", "cartulary.view.decisions.v1", "cartulary.view.assessments.v1", "cartulary.view.indicators.v1", "cartulary.view.parties.v1"}}
	party := c.create("cartulary.view.parties.v1", "ir-party", map[string]any{"party.display_name": supplement.Party, "party.party_kind": "organization"})
	partyID := party["record_id"].(string)
	c.create("cartulary.view.evidence.v1", "ir-pending-evidence", map[string]any{"evidence.title": supplement.Pending, "evidence.lifecycle_state": "requested", "evidence.collector_party_id": partyID})
	var evidence map[string]any
	for index, name := range supplement.Evidence {
		content, err := corpus.Read("ir-incident/v1/supplement/" + name)
		if err != nil {
			t.Fatal(err)
		}
		slot := c.request(http.MethodPost, "/api/v1/object-blobs", map[string]any{"incident_id": c.incident, "client_txn_id": fmt.Sprintf("ir-blob-%d", index), "byte_size": len(content), "filename_hint": name, "content_type_hint": "text/plain", "sha256_hex": fmt.Sprintf("%x", sha256.Sum256(content))}, http.StatusCreated)
		target := slot["upload_target"].(map[string]any)
		href := target["href"].(string)
		if strings.HasPrefix(href, "/") {
			href = c.origin + href
		}
		request, err := http.NewRequestWithContext(t.Context(), http.MethodPut, href, bytes.NewReader(content))
		if err != nil {
			t.Fatal(err)
		}
		request.Header.Set("Content-Type", "text/plain")
		request.AddCookie(c.login.Session)
		request.AddCookie(c.login.CSRF)
		request.Header.Set(authn.CSRFHeaderName, c.login.CSRF.Value)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			t.Fatalf("upload status=%d", response.StatusCode)
		}
		evidenceflow.VerifyUploadedBlob(t, http.DefaultClient, c.origin, c.login.Session, c.login.CSRF, slot["object_blob_id"].(string), fmt.Sprintf("ir-blob-verify-%d", index))
		if index == 0 {
			evidence = c.create("cartulary.view.evidence.v1", "ir-evidence-original", map[string]any{"evidence.title": supplement.Label + " attachment", "evidence.initial_object_blob_id": slot["object_blob_id"], "evidence.source_party_id": partyID})
			state.evidenceID = evidence["record_id"].(string)
		} else {
			evidence = c.request(http.MethodPost, "/api/v1/evidence-records/"+state.evidenceID+"/attach-blob", map[string]any{"client_txn_id": "ir-evidence-replacement", "base_row_version": evidence["row_version"], "object_blob_id": slot["object_blob_id"]}, http.StatusOK)["row"].(map[string]any)
			state.replacement = content
		}
		if index == 0 {
			evidence = c.patch("cartulary.view.evidence.v1", "ir-evidence-available", evidence, irValue("evidence.lifecycle_state", "available"))
		} else if irText(t, evidence, "evidence.lifecycle_state") != "available" {
			t.Fatal("attachment replacement changed availability")
		}
		irVerifyEvidenceContent(c, state.evidenceID, content)
	}
	noteBody := supplement.Note + " Explicit observation: " + supplement.Indicator.Value + "."
	note := c.request(http.MethodPost, "/api/v1/records/"+state.evidenceID+"/linked-notes", map[string]any{"client_txn_id": "ir-linked-note", "note.title": supplement.Label + " linked note", "note.body": noteBody}, http.StatusCreated)
	state.noteID = note["row"].(map[string]any)["record_id"].(string)
	hosts := irQueryAll(t, c.origin, c.login, c.incident, "cartulary.view.hosts.v1")
	if len(hosts) != 30 {
		t.Fatal("supplement host roster mismatch")
	}
	hostSource, err := corpus.Records("ir-incident/v1/hosts.csv", ',')
	if err != nil {
		t.Fatal(err)
	}
	var subjectID any
	for _, host := range hosts {
		if irText(t, host, "host.hostname") == hostSource[1][1] {
			subjectID = host["record_id"]
		}
	}
	if subjectID == nil {
		t.Fatal("representative host source identity missing")
	}
	assessment := c.create("cartulary.view.assessments.v1", "ir-assessment", map[string]any{"assessment.subject_ref": subjectID, "assessment.subject_type": "host", "assessment.assessment_state": "unknown", "assessment.rationale": supplement.Assessment, "assessment.support_refs": irLinks(state.evidenceID, state.noteID)})
	reviewAssessment := c.create("cartulary.view.assessments.v1", "ir-assessment-review", map[string]any{"assessment.subject_ref": subjectID, "assessment.subject_type": "host", "assessment.assessment_state": "unknown", "assessment.rationale": supplement.Assessment + " Synthetic review recorded; status remains unknown.", "assessment.support_refs": irLinks(state.evidenceID, state.noteID)})
	if assessment["record_id"] == reviewAssessment["record_id"] || assessment["row_version"] != float64(1) || reviewAssessment["row_version"] != float64(1) {
		t.Fatal("assessment review must append a distinct judgment")
	}
	decision := c.create("cartulary.view.decisions.v1", "ir-decision", map[string]any{"decision.summary": supplement.Decision, "decision.decision_type": "scope", "decision.rationale": supplement.Assessment, "decision.support_refs": irLinks(state.evidenceID)})
	collect := c.create("cartulary.view.task_requests.v1", "ir-task-collect", map[string]any{"task.title": supplement.Tasks[0].Title, "task.task_kind": "collection", "task.requester_party_id": partyID, "task.decision_record_id": decision["record_id"], "task.linked_record_ids": irLinks(state.evidenceID)})
	review := c.create("cartulary.view.task_requests.v1", "ir-task-review", map[string]any{"task.title": supplement.Tasks[1].Title, "task.task_kind": "follow_up", "task.status": "blocked", "task.blocked_reason": "Await collection task " + collect["record_id"].(string), "task.linked_record_ids": irLinks(collect["record_id"].(string), state.evidenceID), "task.decision_record_id": decision["record_id"]})
	done := c.patch("cartulary.view.task_requests.v1", "ir-task-collected", collect, irValue("task.status", "done"), irValue("task.closure_summary", supplement.Label+": collected and verified both attachment versions."))
	if irText(t, done, "task.completed_at") == "" {
		t.Fatal("completed collection has no completion time")
	}
	review = c.patch("cartulary.view.task_requests.v1", "ir-task-unblocked", review, irValue("task.status", "in_progress"))
	if irText(t, review, "task.blocked_reason") != "" {
		t.Fatal("unblocking retained its blocking reason")
	}
	c.patch("cartulary.view.task_requests.v1", "ir-task-reviewed", review, irValue("task.status", "done"), irValue("task.closure_summary", supplement.Label+": reviewed after collection."))
	indicator := c.create("cartulary.view.indicators.v1", "ir-indicator", map[string]any{"indicator.indicator_type": "ipv4_addr", "indicator.value_kind": "atomic", "indicator.display_value": supplement.Indicator.Value, "indicator.normalized_value": supplement.Indicator.Value})
	state.indicatorID = indicator["record_id"].(string)
	start := strings.Index(noteBody, supplement.Indicator.Value)
	c.request(http.MethodPost, "/api/v1/records/"+state.noteID+"/indicator-observations", map[string]any{"client_txn_id": "ir-indicator-observation", "base_row_version": 1, "source_field_key": "note.body", "span_start_byte": start, "span_end_byte": start + len(supplement.Indicator.Value), "parsed_indicator_type": "ipv4_addr", "resolved_indicator_record_id": state.indicatorID}, http.StatusCreated)
	events := irQueryAll(t, c.origin, c.login, c.incident, "cartulary.view.timeline.v2")
	// Freeze the source identity by its authored first logical record, never a grid position.
	sourceRows, err := corpus.Records("ir-incident/v1/timeline.csv", ',')
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	for _, row := range events {
		if irText(t, row, "timeline.raw_activity_text") == sourceRows[1][7] {
			event = row
		}
	}
	if event == nil {
		t.Fatal("representative logical source record missing")
	}
	changes := []map[string]any{}
	for index, field := range []string{"timeline.host_refs", "timeline.identity_refs"} {
		changes = append(changes, map[string]any{"field_key": field, "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": []map[string]any{{"op": "add_token", "raw_text": supplement.Mentions[index]}}}})
	}
	event = c.patch("cartulary.view.timeline.v2", "ir-explicit-mentions", event, changes...)
	for _, field := range []string{"timeline.host_refs", "timeline.identity_refs"} {
		items := workbookscenariotest.CollectionItems(t, event, field)
		if len(items) != 1 || items[0]["resolved_record_id"] != nil {
			t.Fatal("ambiguous mention silently resolved", field)
		}
	}
	layout, layoutErr := viewschema.DefaultLayout("cartulary.view.timeline.v2")
	if layoutErr != nil {
		t.Fatal(layoutErr)
	}
	saved := c.request(http.MethodPost, "/api/v1/incidents/"+c.incident+"/saved-views", map[string]any{"view_schema_id": "cartulary.view.timeline.v2", "display_name": supplement.SavedView, "scope": "shared", "query_json": map[string]any{}, "layout_json": json.RawMessage(layout)}, http.StatusCreated)
	state.savedViewID = saved["saved_view_id"].(string)
	history := c.request(http.MethodGet, "/api/v1/records/"+state.evidenceID+"/history", nil, http.StatusOK)
	if len(history["items"].([]any)) < 2 {
		t.Fatal("attachment replacement lost revision history")
	}
	return state
}

func irVerifyEvidenceContent(c irPublicClient, id string, expected []byte) {
	c.t.Helper()
	handle := c.request(http.MethodPost, "/api/v1/evidence-records/"+id+"/download-handle", map[string]any{}, http.StatusOK)
	href := handle["href"].(string)
	if strings.HasPrefix(href, "/") {
		href = c.origin + href
	}
	request, err := http.NewRequestWithContext(c.t.Context(), http.MethodGet, href, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	request.AddCookie(c.login.Session)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		c.t.Fatal(err)
	}
	defer response.Body.Close()
	actual, err := io.ReadAll(io.LimitReader(response.Body, int64(len(expected)+1)))
	if err != nil || response.StatusCode != http.StatusOK || !bytes.Equal(actual, expected) {
		c.t.Fatalf("evidence content changed, status=%d err=%v", response.StatusCode, err)
	}
}
func irVerifyInvestigationTransfer(left, right irPublicClient, state irInvestigation) {
	left.t.Helper()
	sourceActor := left.request(http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK)["user_id"]
	destinationActor := right.request(http.MethodGet, "/api/v1/auth/session", nil, http.StatusOK)["user_id"]
	if sourceActor == destinationActor {
		left.t.Fatal("deployments must have independent actors")
	}
	// These current local-user columns follow the source owners' actor remap.
	// Their original attribution is asserted in public history and exact re-export.
	localActorFields := map[string]bool{"note.created_by_user_id": true, "task.owner_user_id": true, "decision.owner_user_id": true, "assessment.assessor": true}
	for _, view := range state.views {
		before, after := irQueryAll(left.t, left.origin, left.login, left.incident, view), irQueryAll(right.t, right.origin, right.login, right.incident, view)
		byID := map[string]map[string]any{}
		for _, row := range after {
			byID[row["record_id"].(string)] = row
		}
		if len(before) != len(after) {
			left.t.Fatal("supplement inventory changed", view)
		}
		for _, row := range before {
			other := byID[row["record_id"].(string)]
			if other == nil {
				left.t.Errorf("supplement record missing in %s", view)
				continue
			}
			if row["row_version"] != other["row_version"] {
				left.t.Errorf("supplement version changed in %s: %v -> %v", view, row["row_version"], other["row_version"])
			}
			beforeCells, afterCells := row["cells"].(map[string]any), other["cells"].(map[string]any)
			if len(beforeCells) != len(afterCells) {
				left.t.Errorf("supplement cell inventory changed in %s", view)
			}
			for key, value := range beforeCells {
				if localActorFields[key] && value.(map[string]any)["value"] != nil {
					if value.(map[string]any)["value"] != sourceActor || afterCells[key].(map[string]any)["value"] != destinationActor {
						left.t.Errorf("incorrect local actor remap for %s", key)
					}
					continue
				}
				if !reflect.DeepEqual(value, afterCells[key]) {
					left.t.Errorf("supplement field %s changed: %#v -> %#v", key, value, afterCells[key])
				}
			}
		}
	}
	if left.t.Failed() {
		left.t.FailNow()
	}
	history := right.request(http.MethodGet, "/api/v1/records/"+state.evidenceID+"/history", nil, http.StatusOK)
	for _, raw := range history["items"].([]any) {
		item := raw.(map[string]any)
		if item["source_actor_id"] != sourceActor || item["actor_user_id"] != destinationActor {
			left.t.Fatal("imported revision lost historical attribution")
		}
	}
	irVerifyEvidenceContent(right, state.evidenceID, state.replacement)
}

func irBindFlowIndicator(c irPublicClient, tableID, indicatorID string) {
	c.t.Helper()
	root := "/api/v1/incidents/" + c.incident + "/network-flow"
	for _, field := range []string{"network_flow.endpoint_a_ip", "network_flow.endpoint_b_ip"} {
		page := c.request(http.MethodPost, root+"/tables/"+tableID+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v2", "limit": 1, "filters": []any{map[string]any{"field_key": field, "op": "eq", "value": "10.20.20.10"}}}, http.StatusOK)
		rows := page["rows"].([]any)
		if len(rows) == 0 {
			continue
		}
		row := rows[0].(map[string]any)
		c.request(http.MethodPost, root+"/indicator-links", map[string]any{"schema_id": "cartulary.network_flow.indicator_link_request.v2", "client_txn_id": "ir-explicit-flow-binding", "selector": map[string]any{"kind": "row_field_value", "network_flow_table_id": tableID, "network_flow_row_id": row["network_flow_row_id"], "field_key": field}, "target": map[string]any{"mode": "existing_indicator", "indicator_id": indicatorID}, "observation_mode": "binding_only", "confirm_exact_value": "10.20.20.10"}, http.StatusCreated)
		return
	}
	c.t.Fatal("supplement indicator has no matching source observation")
}
