package rollback

import (
	"encoding/json"
	"maps"
	"reflect"
	"testing"
)

func TestSourceForRollbackValueMapsAllArtifactVariantsWithoutCollections(t *testing.T) {
	t.Parallel()
	for _, typ := range []string{"note", "comm_log", "handoff", "status_review", "lesson", "finding", "investigative_query", "forensic_keyword"} {
		value := artifactRollbackFixture(typ)
		original, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		got, ok := extractRollbackSource(value)
		if !ok {
			t.Fatalf("complete %s source rejected", typ)
		}
		source := value["source"].(map[string]any)
		want := source
		if child, nested := source[typ].(map[string]any); nested {
			want = maps.Clone(child)
			want["artifact_type"] = typ
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s active source = %#v, want %#v", typ, got, want)
		}
		after, err := json.Marshal(value)
		if err != nil || string(original) != string(after) {
			t.Fatal("retained source was rewritten")
		}
		if _, present := got["open_risk_refs"]; present {
			t.Fatal("separately revisioned collection entered row source")
		}
	}
	for _, value := range []map[string]any{
		{"source": map[string]any{"body": "note", "summary": "brief", "next_checks": nil, "active_risks_summary": "risk", "closure_state": "closed", "kind": "hypothesis", "platform": "KQL", "case_sensitive": true}},
		{"cells": map[string]any{"note.body": map[string]any{"value": "legacy"}}},
	} {
		if _, ok := extractRollbackSource(value); ok {
			t.Fatal("partial or projection source accepted")
		}
	}
	old := artifactRollbackFixture("note")
	old["snapshot_schema_id"] = "cartulary.revisions.snapshot.artifact.v1"
	if _, ok := extractRollbackSource(old); ok {
		t.Fatal("retired Artifact snapshot schema accepted")
	}
}

func TestValidSourceRejectsSubtypeInvariantViolations(t *testing.T) {
	t.Parallel()
	for _, score := range []json.Number{"0", "25", "100"} {
		value := artifactRollbackFixture("finding")
		value["source"].(map[string]any)["finding"].(map[string]any)["confidence_score"] = score
		if err := (Provider{}).ValidateRollbackValue(value); err != nil {
			t.Fatalf("valid exact score %s rejected: %v", score, err)
		}
	}
	for _, score := range []any{json.Number("-1"), json.Number("101"), json.Number("0.5"), json.Number("9223372036854775807"), float64(101)} {
		value := artifactRollbackFixture("finding")
		value["source"].(map[string]any)["finding"].(map[string]any)["confidence_score"] = score
		if err := (Provider{}).ValidateRollbackValue(value); err == nil {
			t.Fatalf("invalid exact score %v accepted", score)
		}
	}
	for _, test := range []struct{ typ, field, invalid string }{{"comm_log", "comm_type", "invalid"}, {"forensic_keyword", "match_mode", "glob"}} {
		value := artifactRollbackFixture(test.typ)
		source := value["source"].(map[string]any)
		if child, ok := source[test.typ].(map[string]any); ok {
			source = child
		}
		source[test.field] = test.invalid
		if err := (Provider{}).ValidateRollbackValue(value); err == nil {
			t.Fatalf("invalid %s accepted", test.field)
		}
	}
}

func artifactRollbackFixture(typ string) map[string]any {
	const id = "10000000-0000-4000-8000-000000000001"
	const incident = "20000000-0000-4000-8000-000000000002"
	const actor = "30000000-0000-4000-8000-000000000003"
	const created = "2026-01-01T00:00:00Z"
	const updated = "2026-01-01T01:00:00Z"
	source := map[string]any{
		"record_id": id, "incident_id": incident, "artifact_type": typ, "title": "Title", "body": "Body",
		"timestamp_utc": created, "updated_at": updated, "created_at": created,
		"comm_id": "COMM-1", "comm_type": "briefing", "audience": "Responders", "channel_or_meeting": "Bridge", "summary": "Summary", "next_report_at": nil, "privilege_tag": nil,
		"handoff_id": "HANDOFF-1", "outgoing_owner_user_id": actor, "incoming_owner_user_id": actor, "current_state_summary": "Current", "next_checks": nil, "acknowledged_at": nil,
		"status_review_id": "REVIEW-1", "review_owner_user_id": actor, "active_risks_summary": nil,
		"lesson_id": "LESSON-1", "owner_user_id": actor, "closure_state": "closed", "created_by_user_id": actor,
		"finding": nil, "investigative_query": nil, "forensic_keyword": nil,
	}
	switch typ {
	case "finding":
		source[typ] = map[string]any{"record_id": id, "incident_id": incident, "kind": "hypothesis", "statement": "Statement", "state": "open", "confidence_score": json.Number("25"), "owner_user_id": actor, "closed_at": nil, "created_at": created, "updated_at": updated}
	case "investigative_query":
		source[typ] = map[string]any{"record_id": id, "incident_id": incident, "query_id": "QUERY-1", "platform": "KQL", "purpose": "Purpose", "query_text": "Query", "created_by_user_id": actor, "created_at": created, "updated_at": updated}
	case "forensic_keyword":
		source[typ] = map[string]any{"record_id": id, "incident_id": incident, "keyword_id": "KEYWORD-1", "pattern": "Pattern", "reason": "Reason", "match_mode": "literal", "case_sensitive": true, "created_at": created, "updated_at": updated}
	}
	return map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.artifact.v2", "record": map[string]any{"record_id": id, "incident_id": incident}, "source": source}
}
