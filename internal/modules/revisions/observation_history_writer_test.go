package revisions_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
)

func expectedObservationHistory(id, incident, source, actor uuid.UUID, field string, at time.Time, target *uuid.UUID) map[string]any {
	value := map[string]any{"indicator_observation_id": id.String(), "incident_id": incident.String(), "source_record_id": source.String(), "source_field_key": field, "origin_kind": "manual_entry", "origin_locator": fmt.Sprintf("record:%s:field:%s:bytes:0-25", source, field), "observed_text": "record-support-source-row", "parsed_indicator_type": "text", "normalized_candidate": "record-support-source-row", "resolution_status": "unresolved", "resolved_indicator_record_id": nil, "row_version": 1, "created_by_user_id": actor.String(), "created_at": at.UTC().Format(time.RFC3339Nano), "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil, "deleted_at": nil, "deleted_by_user_id": nil}
	if target != nil {
		value["resolution_status"], value["resolved_indicator_record_id"], value["resolved_by_user_id"], value["resolved_at"], value["resolution_method"] = "resolved", target.String(), actor.String(), at.UTC().Format(time.RFC3339Nano), "indicators.observations.capture"
	}
	return value
}

func requireObservationHistory(t testing.TB, db *sql.DB, changeSet, operation string, id uuid.UUID, before, after map[string]any) {
	t.Helper()
	encode := func(v map[string]any) any {
		if v == nil {
			return nil
		}
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var count int
	var exact bool
	if err := db.QueryRow(`SELECT count(*),coalesce(bool_and(operation_kind=$3 AND coalesce(before_value,'null'::jsonb)=coalesce($4::jsonb,'null'::jsonb) AND after_value=$5::jsonb),false) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='indicator_observation' AND target_id=$2`, changeSet, id.String(), operation, encode(before), encode(after)).Scan(&count, &exact); err != nil || count != 1 || !exact {
		var differences string
		diagnosticErr := db.QueryRow(`SELECT coalesce(jsonb_object_agg(side||'.'||key,jsonb_build_object('actual',actual->key,'expected',expected->key)),'{}'::jsonb)::text
FROM change_set_mutations m
CROSS JOIN LATERAL (VALUES ('before',coalesce(m.before_value,'{}'::jsonb),coalesce($3::jsonb,'{}'::jsonb)),('after',coalesce(m.after_value,'{}'::jsonb),coalesce($4::jsonb,'{}'::jsonb))) pair(side,actual,expected)
CROSS JOIN LATERAL jsonb_object_keys(actual||expected) keys(key)
WHERE m.change_set_id=$1 AND m.target_kind='indicator_observation' AND m.target_id=$2 AND actual->key IS DISTINCT FROM expected->key`, changeSet, id.String(), encode(before), encode(after)).Scan(&differences)
		t.Fatalf("complete Observation history count=%d exact=%v error=%v differences=%s diagnostic=%v", count, exact, err, differences, diagnosticErr)
	}
	if err := db.QueryRow(`SELECT (to_jsonb(o)-'created_at'-'resolved_at'-'deleted_at')=($2::jsonb-'created_at'-'resolved_at'-'deleted_at') AND created_at=($2::jsonb->>'created_at')::timestamptz AND resolved_at IS NOT DISTINCT FROM ($2::jsonb->>'resolved_at')::timestamptz AND deleted_at IS NOT DISTINCT FROM ($2::jsonb->>'deleted_at')::timestamptz FROM indicator_observations o WHERE indicator_observation_id=$1`, id, encode(after)).Scan(&exact); err != nil || !exact {
		t.Fatalf("complete Observation source exact=%v error=%v", exact, err)
	}
}
