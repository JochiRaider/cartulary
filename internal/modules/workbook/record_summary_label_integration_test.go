package workbook_test

import (
	"fmt"
	"maps"
	"net/http"
	"regexp"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

var recordLabelUUIDPattern = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// TestRecordRefLabelsServeLinkedRecordSummary_Integration verifies AC-605 for
// served record_ref items. Each linked record type displays its REQ-01-333a
// label, an empty label source reads "Untitled <noun>", an assessment reads
// "Assessment of <subject label>", no label carries a record identifier, and a
// changed label appears on the referring row's next read. The expected labels
// are authored here, independently of the label owner.
func TestRecordRefLabelsServeLinkedRecordSummary_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "workbook-record-ref-labels", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "record-ref-labels-incident",
		"incident_key":  "RECORD-REF-LABELS",
		"title":         "Record ref labels",
	})
	incidentID := incident["incident_id"].(string)
	authOptions := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	creates := 0
	create := func(name string, fields map[string]any) map[string]any {
		t.Helper()
		creates++
		body := maps.Clone(fields)
		body["client_txn_id"] = fmt.Sprintf("record-ref-labels-%d", creates)
		view := "cartulary.view." + name + ".v1"
		if name == "timeline" {
			view = "cartulary.view.timeline.v2"
		}
		resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+view+"/rows", body, authOptions...)
		return httptestx.RequireSuccessData(t, resp, http.StatusCreated)["row"].(map[string]any)
	}
	id := func(row map[string]any) string { return row["record_id"].(string) }

	host := create("hosts", map[string]any{"host.hostname": "label-host.example.test", "host.display_name": "Label host"})
	timeline := create("timeline", map[string]any{"timeline.activity_synopsis_text": "Label timeline synopsis"})
	evidence := create("evidence", map[string]any{"evidence.collector_party_text": "Label collector", "evidence.title": "Label evidence"})
	untitledEvidence := create("evidence", map[string]any{"evidence.collector_party_text": "Untitled collector"})
	want := map[string]string{
		id(evidence):         "Label evidence",
		id(untitledEvidence): "Untitled evidence",
		id(host):             "Label host",
		id(timeline):         "Label timeline synopsis",
		id(create("identities", map[string]any{"identity.email": "label-identity@example.test", "identity.display_name": "Label identity"})):                                                            "Label identity",
		id(create("parties", map[string]any{"party.party_kind": "person", "party.display_name": "Label party"})):                                                                                        "Label party",
		id(create("indicators", map[string]any{"indicator.indicator_type": "domain_name", "indicator.value_kind": "atomic", "indicator.display_value": "label-indicator.example.test"})):                "label-indicator.example.test",
		id(create("task_requests", map[string]any{"task.task_kind": "question", "task.title": "Label task"})):                                                                                           "Label task",
		id(create("findings", map[string]any{"finding.statement": "Label finding"})):                                                                                                                    "Label finding",
		id(create("notes", map[string]any{"note.body": "Label body", "note.title": "Label note"})):                                                                                                      "Label note",
		id(create("comm_log", map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "Operations", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Label communication"})): "Label communication",
		id(create("lesson", map[string]any{"lesson.summary": "Label lesson"})):                                                                                                                          "Label lesson",
		id(create("handoff", map[string]any{"handoff.incoming_owner_user_id": actorID.String(), "handoff.current_state_summary": "Label handoff"})):                                                     "Label handoff",
		id(create("status_review", map[string]any{"status_review.current_state_summary": "Label status review"})):                                                                                       "Label status review",
		id(create("forensic_keywords", map[string]any{"forensic_keyword.reason": "Label fixture", "forensic_keyword.pattern": "label-keyword"})):                                                        "label-keyword",
		id(create("investigative_queries", map[string]any{"investigative_query.platform": "SQL", "investigative_query.query_text": "select value", "investigative_query.purpose": "Label query"})):      "Label query",
		id(create("timeline", map[string]any{"timeline.analyst_text": "Untitled fixture"})):                                                                                                             "Untitled timeline record",
	}
	decision := create("decisions", map[string]any{"decision.decision_type": "scope", "decision.rationale": "Label fixture", "decision.summary": "Label decision"})
	want[id(decision)] = "Label decision"
	assessment := create("assessments", map[string]any{
		"assessment.subject_ref": id(host), "assessment.subject_type": "host", "assessment.assessment_state": "confirmed",
		"assessment.confidence_score": 85, "assessment.assessed_at": "2026-09-10T00:00:00Z", "assessment.rationale": "Label assessment",
		"assessment.support_refs": recordRefActions(id(timeline)),
	})
	want[id(assessment)] = "Assessment of Label host"
	requireRecordRefLabels(t, assessment, "assessment.support_refs", map[string]string{id(timeline): "Label timeline synopsis"})

	linked := make([]string, 0, len(want))
	for recordID := range want {
		linked = append(linked, recordID)
	}
	referrer := create("task_requests", map[string]any{"task.task_kind": "question", "task.title": "Label referrer", "task.linked_record_ids": recordRefActions(linked...)})
	requireRecordRefLabels(t, referrer, "task.linked_record_ids", want)

	resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+id(decision), map[string]any{
		"view_schema_id": "cartulary.view.decisions.v1", "base_row_version": 1, "client_txn_id": "record-ref-labels-relabel",
		"changes": []map[string]any{{"field_key": "decision.summary", "value": "Relabelled decision"}},
	}, authOptions...)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	want[id(decision)] = "Relabelled decision"
	reread := queryWorkbookRow(t, harness, login, uuid.MustParse(incidentID), "cartulary.view.task_requests.v1", uuid.MustParse(id(referrer)))
	requireRecordRefLabels(t, reread, "task.linked_record_ids", want)

	// Timeline stores attached Evidence labels from the Evidence owner. They
	// equal the record-ref labels above, and an Evidence title patch refreshes
	// the referring Timeline row before its next read.
	holder := create("timeline", map[string]any{"timeline.activity_synopsis_text": "Evidence holder", "timeline.attached_evidence_ids": recordRefActions(id(evidence), id(untitledEvidence))})
	requireRecordRefLabels(t, holder, "timeline.attached_evidence_ids", map[string]string{id(evidence): want[id(evidence)], id(untitledEvidence): want[id(untitledEvidence)]})
	resp = httptestx.DoJSON(t, http.DefaultClient, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+id(evidence), map[string]any{
		"view_schema_id": "cartulary.view.evidence.v1", "base_row_version": 1, "client_txn_id": "record-ref-labels-evidence-title",
		"changes": []map[string]any{{"field_key": "evidence.title", "value": "Relabelled evidence"}},
	}, authOptions...)
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	holderRow := queryWorkbookRow(t, harness, login, uuid.MustParse(incidentID), "cartulary.view.timeline.v2", uuid.MustParse(id(holder)))
	requireRecordRefLabels(t, holderRow, "timeline.attached_evidence_ids", map[string]string{id(evidence): "Relabelled evidence", id(untitledEvidence): "Untitled evidence"})
}

func recordRefActions(recordIDs ...string) map[string]any {
	actions := make([]map[string]any, 0, len(recordIDs))
	for _, recordID := range recordIDs {
		actions = append(actions, map[string]any{"op": "add_record_ref", "linked_record_id": recordID})
	}
	return map[string]any{"kind": "collection_actions_v1", "actions": actions}
}

func requireRecordRefLabels(t testing.TB, row map[string]any, fieldKey string, want map[string]string) {
	t.Helper()
	cell := row["cells"].(map[string]any)[fieldKey].(map[string]any)
	items := cell["value"].(map[string]any)["items"].([]any)
	got := make(map[string]string, len(items))
	for _, raw := range items {
		item := raw.(map[string]any)
		label, _ := item["display_text"].(string)
		if recordLabelUUIDPattern.MatchString(label) {
			t.Fatalf("%s label %q carries a record identifier", fieldKey, label)
		}
		got[item["linked_record_id"].(string)] = label
	}
	if !maps.Equal(got, want) {
		t.Fatalf("%s labels by linked record:\n got  %v\n want %v", fieldKey, got, want)
	}
}
