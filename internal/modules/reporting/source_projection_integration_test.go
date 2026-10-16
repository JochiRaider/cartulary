package reporting_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"
	"slices"
	"strings"
	"testing"
	"time"
)

// Compare the frozen snapshot's Core identities with the dedicated fixture's
// active envelope roster. Auxiliary source tables cannot inflate this set.
func requireCoreRecordSummaries(t testing.TB, db *sql.DB, model map[string]any, incident string) {
	t.Helper()
	rows, err := db.Query("SELECT record_id::text, record_type FROM records WHERE incident_id=$1 AND deleted_at IS NULL", incident)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	expected := map[string]string{}
	for rows.Next() {
		var id, kind string
		if err := rows.Scan(&id, &kind); err != nil {
			t.Fatal(err)
		}
		expected[id] = kind
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	records, ok := model["records"].([]any)
	if !ok || len(records) != len(expected) {
		t.Fatal("snapshot Core roster differs from active envelopes")
	}
	seen := map[string]bool{}
	for _, value := range records {
		record, ok := value.(map[string]any)
		if !ok {
			t.Fatal("invalid Core summary")
		}
		id, ok := record["record_id"].(string)
		if !ok || seen[id] || expected[id] == "" || record["record_type"] != expected[id] {
			t.Fatal("duplicate, auxiliary or mistyped Core summary")
		}
		seen[id] = true
		ref, ok := record["source_record_ref"].(map[string]any)
		if !ok || ref["source_record_id"] != id {
			t.Fatal("summary lost its source envelope identity")
		}
		switch ref["source_family"] {
		case "record_tag", "record_link", "entity_mention":
			t.Fatal("auxiliary source became a Core record")
		}
	}
}

func seedReportingAssignment(t testing.TB, db *sql.DB, incident, actor string, ids map[string]string) string {
	t.Helper()
	const assignment = "00000000-0000-0000-0000-000000004007"
	if _, err := db.Exec(`INSERT INTO record_links(record_link_id,incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,owner_user_id,created_by_user_id,created_at,decided_at) VALUES($1,$2,$3,$4,'references_record','host.subject_party_ids','manual',$5,$5,'2026-05-01T01:02:03.123456Z','2026-05-01T01:02:03.123456Z')`, assignment, incident, ids["host"], ids["party"], actor); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO record_links(record_link_id,incident_id,src_record_id,dst_record_id,link_type,field_key,provenance,owner_user_id,created_by_user_id,created_at,decided_at) VALUES('00000000-0000-0000-0000-000000004008',$1,$2,$3,'references_record','identity.subject_party_ids','manual',$4,$4,'2026-05-01T01:02:03.123456Z','2026-05-01T01:02:03.123456Z')`, incident, ids["identity"], ids["party"], actor); err != nil {
		t.Fatal(err)
	}
	return assignment
}

func requireFrozenDisclosureFacts(t testing.TB, db *sql.DB, model map[string]any, assignment, snapshot string, ids map[string]string) {
	t.Helper()
	raw, err := json.Marshal(model["source_facts"])
	if err != nil {
		t.Fatal(err)
	}
	var facts exportprovider.SourceFacts
	if err := json.Unmarshal(raw, &facts); err != nil || facts.Validate() != nil || len(facts.Parties) != 1 || len(facts.Assignments) != 2 {
		t.Fatal("missing typed disclosure facts", err)
	}
	party, fact := facts.Parties[0], facts.Assignments[0]
	if party.PartyID != ids["party"] || party.PartitionSegment != ids["party"] || party.PublicDirectoryEligible || fact.AssignmentID.String() != assignment || fact.SubjectRecordID.String() != ids["host"] || fact.PartyID.String() != ids["party"] || fact.Provenance != "manual" || fact.CreatedAt != "2026-05-01T01:02:03.123456Z" {
		t.Fatal("source provenance or default privacy was lost")
	}
	identity := facts.Assignments[1]
	if identity.SubjectRecordType != "identity" || identity.SubjectRecordID.String() != ids["identity"] || identity.AssignmentKind != "subject_party" || identity.PartyID.String() != ids["party"] {
		t.Fatal("Identity disclosure assignment was not retained")
	}
	// A source change after snapshot admission must not alter the retained
	// disclosure input used by release creation, replay or recovery.
	if _, err := db.Exec(`UPDATE record_links SET provenance='import', field_key='host.observer_party_ids' WHERE record_link_id=$1`, assignment); err != nil {
		t.Fatal(err)
	}
	var retained []byte
	if err := db.QueryRow(`SELECT p.request_json->'source_capture'->'source_facts' FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id=$1`, snapshot).Scan(&retained); err != nil {
		t.Fatal(err)
	}
	var stored exportprovider.SourceFacts
	if err := json.Unmarshal(retained, &stored); err != nil {
		t.Fatal(err)
	}
	reencoded, _ := json.Marshal(stored)
	if !bytes.Equal(raw, reencoded) { // compare decoded values; JSONB changes member order
		original, _ := json.Marshal(facts)
		if !bytes.Equal(original, reencoded) {
			t.Fatal("retained disclosure facts followed mutable source")
		}
	}
	for path, field := range exportModelFieldsByPath(t, model) {
		if strings.HasPrefix(path, "/hosts/"+ids["host"]+"/") {
			if refs, ok := field["support_refs"].([]any); ok && len(refs) > 0 {
				t.Fatal("Party assignment was misclassified as evidentiary support")
			}
		}
	}
}

// Only the dedicated test database is changed. A future migration must not
// accidentally publish a newly stored value through a whole-row JSON query.
func addUnpublishedReportingColumns(t testing.TB, db *sql.DB) {
	t.Helper()
	for _, table := range []string{
		"records", "timeline_events", "parties", "entity_mentions",
		"host_grid_projection", "identity_grid_projection", "task_request_grid_projection",
		"decision_grid_projection", "artifact_grid_projection",
	} {
		if _, err := db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN future_private text NOT NULL DEFAULT 'unpublished-source-value'`); err != nil {
			t.Fatalf("add unpublished fixture column to %s: %v", table, err)
		}
	}
}

func requireClosedReportingSourceFields(t testing.TB, model map[string]any, ids map[string]string) {
	t.Helper()
	encoded, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "future_private") || strings.Contains(string(encoded), "unpublished-source-value") {
		t.Fatal("snapshot exposed an unadmitted source column")
	}
	fields := exportModelFieldsByPath(t, model)
	commonArtifact := "record_id row_version artifact_type title body timestamp_utc updated_at created_at created_by_user_id linked_record_count"
	wantMembers := map[string]string{
		"/record_envelopes/" + ids["timeline"]:    "record_id record_type created_by_user_id created_at updated_by_user_id updated_at row_version",
		"/timeline/" + ids["timeline"]:            "record_id capture_state row_version recorded_at edited_at created_by_user_id updated_by_user_id reviewed_by_user_id reviewed_at superseded_by_user_id superseded_at date_entered_text analyst_text mitre_stage_text device_object_text ip_address_text activity_utc_text activity_local_text raw_activity_text activity_synopsis_text data_source_text activity_utc_generated activity_local_generated activity_time_pair_state",
		"/hosts/" + ids["host"]:                   "record_id row_version display_name hostname host_state linked_event_count evidence_count location os_platform business_owner criticality containment_status edited_at",
		"/identities/" + ids["identity"]:          "record_id row_version display_name upn email sam_account_name identity_state linked_event_count evidence_count privilege_level mfa_state reset_status edited_at",
		"/parties/" + ids["party"]:                "record_id display_name party_kind organization_name role_title primary_email timezone_name external_ref notes created_at updated_at",
		"/task_requests/" + ids["task"]:           "record_id row_version title status owner_user_id priority task_kind workstream due_at requester_party_text requester_party_id blocked_reason completed_at external_ticket_ref closure_summary decision_record_id linked_record_count updated_at no_owner",
		"/decisions/" + ids["decision"]:           "record_id row_version summary status owner_user_id decision_type decided_at rationale affected_record_count supersedes_record_id updated_at is_superseded",
		"/notes/" + ids["note"]:                   commonArtifact,
		"/findings/" + ids["finding"]:             commonArtifact + " finding_statement finding_kind finding_state finding_owner_user_id finding_confidence_score finding_closed_at finding_updated_at finding_confidence_band",
		"/comm_log/" + ids["comm_log"]:            commonArtifact + " comm_id comm_type audience channel_or_meeting summary next_report_at privilege_tag",
		"/handoffs/" + ids["handoff"]:             commonArtifact + " handoff_id outgoing_owner_user_id incoming_owner_user_id current_state_summary next_checks acknowledged_at ack_state",
		"/status_reviews/" + ids["status_review"]: commonArtifact + " status_review_id review_owner_user_id current_state_summary active_risks_summary next_report_at",
		"/lessons/" + ids["lesson"]:               commonArtifact + " lesson_id owner_user_id closure_state",
	}
	mentionCount := 0
	for path := range fields {
		if strings.HasPrefix(path, "/entity_mentions/") && strings.HasSuffix(path, "/entity_mention_id") {
			wantMembers[strings.TrimSuffix(path, "/entity_mention_id")] = "entity_mention_id source_record_id entity_type source_field_key origin_kind origin_locator raw_text normalized_text resolution_status row_version ordinal created_by_user_id created_at resolved_record_id resolved_by_user_id resolved_at resolution_method"
			mentionCount++
		}
	}
	if mentionCount != 1 {
		t.Fatalf("source boundary fixture has %d mentions, want 1", mentionCount)
	}
	for prefix, members := range wantMembers {
		got := []string{}
		for path := range fields {
			if member, ok := strings.CutPrefix(path, prefix+"/"); ok {
				got = append(got, member)
			}
		}
		want := strings.Fields(members)
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("source %s exported members %v, want exact closed set %v", prefix, got, want)
		}
	}
}

func requireFrozenTimelineSort(t testing.TB, db *sql.DB, model map[string]any, recordID string) {
	t.Helper()
	events, ok := model["timeline_events"].([]any)
	if !ok || len(events) != 1 {
		t.Fatal("snapshot does not have exactly one Timeline event")
	}
	event := events[0].(map[string]any)
	var createdAt time.Time
	if err := db.QueryRow(`SELECT created_at FROM records WHERE record_id=$1`, recordID).Scan(&createdAt); err != nil {
		t.Fatal(err)
	}
	if event["record_id"] != recordID || event["record_created_at"] != createdAt.UTC().Format("2006-01-02T15:04:05.000000Z") || event["activity_sort_ts"] != "2026-05-02T01:02:03.000001Z" || event["activity_parse_state"] != "parsed" || event["activity_precision_rank"] != float64(6) || event["date_entered_sort_key"] != "2026-05-01" {
		t.Fatal("snapshot lost source-owned sort facts", event)
	}
	wantKeys := strings.Fields("schema_id record_id activity_sort_key activity_sort_ts activity_parse_state activity_precision_rank date_entered_sort_key record_created_at display_times fields source_refs support_refs disclosure_partition_refs")
	if len(event) != len(wantKeys) {
		t.Fatal("Timeline event contains noncontract members")
	}
	for _, key := range wantKeys {
		if _, exists := event[key]; !exists {
			t.Fatal("Timeline event missing member", key)
		}
	}
	display, ok := event["display_times"].(map[string]any)
	if !ok || len(display) != 3 {
		t.Fatal("invalid display-times shape")
	}
	for _, key := range []string{"primary_display", "precision_label", "source_text_display"} {
		if value, exists := display[key]; !exists || value != nil {
			t.Fatal("snapshot inferred presentation", key)
		}
	}
}
