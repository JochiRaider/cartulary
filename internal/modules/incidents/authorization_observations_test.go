package incidents_test

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// Source selection stays explicit and local to each operation family. Existing
// observers cover immutable history, idempotency receipts and publication content,
// excluding dispatcher bookkeeping and legitimate session activity.
type authorizationMutationState struct {
	Source  [32]byte
	Effects workbookscenariotest.MutationEffects
}

func snapshotIncidentResourceAuthorization(t testing.TB, f *authorizationFixture) authorizationMutationState {
	t.Helper()
	var content string
	if err := f.harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'incident',(SELECT jsonb_build_object(
  'id',i.id,
  'incident_key',i.incident_key,
  'incident_key_canonical',i.incident_key_canonical,
  'title',i.title,
  'description',i.description,
  'severity',i.severity,
  'status',i.status,
  'tlp',i.tlp,
  'current_phase',i.current_phase,
  'primary_external_case_ref',i.primary_external_case_ref,
  'incident_version',i.incident_version,
  'created_by_user_id',i.created_by_user_id,
  'updated_by_user_id',i.updated_by_user_id,
  'created_at',i.created_at,
  'updated_at',i.updated_at,
  'closed_at',i.closed_at
 ) FROM incidents i WHERE id::text=$1),
 'audit',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY id),'[]') FROM deployment_admin_audit_events a WHERE incident_id::text=$1)
 )::text`, f.incidentID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return authorizationMutationState{sha256.Sum256([]byte(content)), workbookscenariotest.SnapshotIncidentMutationEffects(t, f.harness.DB, f.incidentID)}
}

func snapshotMembershipAuthorization(t testing.TB, f *authorizationFixture, target string) authorizationMutationState {
	t.Helper()
	var content string
	if err := f.harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'incident',(SELECT jsonb_build_object(
  'id',i.id,
  'incident_key',i.incident_key,
  'incident_key_canonical',i.incident_key_canonical,
  'title',i.title,
  'description',i.description,
  'severity',i.severity,
  'status',i.status,
  'tlp',i.tlp,
  'current_phase',i.current_phase,
  'primary_external_case_ref',i.primary_external_case_ref,
  'incident_version',i.incident_version,
  'created_by_user_id',i.created_by_user_id,
  'updated_by_user_id',i.updated_by_user_id,
  'created_at',i.created_at,
  'updated_at',i.updated_at,
  'closed_at',i.closed_at
 ) FROM incidents i WHERE id::text=$1),
 'target',(SELECT to_jsonb(m) FROM incident_memberships m WHERE incident_id::text=$1 AND user_id::text=$2),
 'audit',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY id),'[]') FROM deployment_admin_audit_events a WHERE incident_id::text=$1)
 )::text`, f.incidentID, target).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return authorizationMutationState{sha256.Sum256([]byte(content)), workbookscenariotest.SnapshotIncidentMutationEffects(t, f.harness.DB, f.incidentID)}
}

func snapshotPreferenceAuthorization(t testing.TB, f *authorizationFixture) authorizationMutationState {
	t.Helper()
	var content string
	if err := f.harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'default',(SELECT to_jsonb(p) FROM incident_workbook_preferences p WHERE incident_id::text=$1),
 'personal',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY user_id),'[]') FROM user_workbook_preferences p WHERE incident_id::text=$1),
 'audit',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY id),'[]') FROM deployment_admin_audit_events a WHERE incident_id::text=$1)
 )::text`, f.incidentID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return authorizationMutationState{sha256.Sum256([]byte(content)), workbookscenariotest.SnapshotIncidentMutationEffects(t, f.harness.DB, f.incidentID)}
}

func snapshotTimelineAuthorization(t testing.TB, f *authorizationFixture) authorizationMutationState {
	t.Helper()
	var content string
	if err := f.harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'source',(SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY record_id),'[]') FROM timeline_events e WHERE incident_id::text=$1),
 'projection',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY record_id),'[]') FROM timeline_grid_projection p WHERE incident_id::text=$1),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]') FROM record_links l WHERE incident_id::text=$1),
 'mentions',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY to_jsonb(m)::text),'[]') FROM entity_mentions m WHERE source_record_id IN (SELECT record_id FROM records WHERE incident_id::text=$1))
 )::text`, f.incidentID).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return authorizationMutationState{sha256.Sum256([]byte(content)), workbookscenariotest.SnapshotIncidentMutationEffects(t, f.harness.DB, f.incidentID)}
}

func snapshotIncidentCreationAuthorization(t testing.TB, f *authorizationFixture, txn string) [32]byte {
	t.Helper()
	var content string
	// No incident exists before this operation; the isolated database contains
	// only this scenario's accounts. Observe every bootstrap-owned table.
	if err := f.harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'incidents',(SELECT coalesce(jsonb_agg(jsonb_build_object(
  'id',i.id,
  'incident_key',i.incident_key,
  'incident_key_canonical',i.incident_key_canonical,
  'title',i.title,
  'description',i.description,
  'severity',i.severity,
  'status',i.status,
  'tlp',i.tlp,
  'current_phase',i.current_phase,
  'primary_external_case_ref',i.primary_external_case_ref,
  'incident_version',i.incident_version,
  'created_by_user_id',i.created_by_user_id,
  'updated_by_user_id',i.updated_by_user_id,
  'created_at',i.created_at,
  'updated_at',i.updated_at,
  'closed_at',i.closed_at
 ) ORDER BY id),'[]') FROM incidents i),
 'memberships',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY incident_id,user_id),'[]') FROM incident_memberships m),
 'defaults',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY incident_id),'[]') FROM incident_workbook_preferences p),
 'personal',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY incident_id,user_id),'[]') FROM user_workbook_preferences p),
 'receipts',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY id),'[]') FROM route_idempotency r WHERE client_txn_id=$1),
 'audit',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY id),'[]') FROM deployment_admin_audit_events a WHERE event_kind IN ('incident_created','incident_membership_created'))
 )::text`, txn).Scan(&content); err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256([]byte(content))
}

func requireIncidentAuthorizationAudit(t testing.TB, f *authorizationFixture, requestID, kind, target string) {
	t.Helper()
	var count int
	if err := f.harness.DB.QueryRowContext(context.Background(), `
SELECT count(*) FROM deployment_admin_audit_events
 WHERE incident_id::text=$1 AND request_id=$2 AND event_kind=$3
   AND actor_user_id::text=$4
   AND ($5='' OR before_json->>'user_id'=$5 OR after_json->>'user_id'=$5)
`, f.incidentID, requestID, kind, f.subjectID, target).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("attributed %s audit events = %d, want 1", kind, count)
	}
}

func requireAuthorizationReceipt(t testing.TB, f *authorizationFixture, txn, scope string, status int) {
	t.Helper()
	var count int
	if err := f.harness.DB.QueryRowContext(context.Background(), `
SELECT count(*) FROM route_idempotency
 WHERE client_txn_id=$1 AND scope_key=$2 AND actor_user_id::text=$3 AND status_code=$4
`, txn, scope, f.subjectID, status).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("committed idempotency receipts = %d, want 1", count)
	}
}

type authorizationTimelineRecord struct {
	RecordVersion int64
	SourceVersion int64
	CaptureState  string
	Synopsis      string
	Actor         string
}

func readAuthorizationTimelineRecord(t testing.TB, f *authorizationFixture, recordID string) authorizationTimelineRecord {
	t.Helper()
	var record authorizationTimelineRecord
	if err := f.harness.DB.QueryRowContext(context.Background(), `
SELECT r.row_version,e.row_version,e.capture_state,e.activity_synopsis_text,e.updated_by_user_id::text
 FROM records r JOIN timeline_events e USING(record_id)
 WHERE r.incident_id::text=$1 AND r.record_id::text=$2
`, f.incidentID, recordID).Scan(&record.RecordVersion, &record.SourceVersion, &record.CaptureState, &record.Synopsis, &record.Actor); err != nil {
		t.Fatal(err)
	}
	return record
}

func requireTimelineAuthorizationHistory(t testing.TB, f *authorizationFixture, recordID, txn, source, requestID string, version int64) string {
	t.Helper()
	var changeSetID string
	if err := f.harness.DB.QueryRowContext(context.Background(), `
SELECT change_set_id::text FROM change_sets WHERE incident_id::text=$1 AND client_txn_id=$2 AND request_id=$3
`, f.incidentID, txn, requestID).Scan(&changeSetID); err != nil {
		t.Fatal(err)
	}
	workbookscenariotest.RequireChangeSetAttribution(t, f.harness.DB, changeSetID, f.subjectID, source, txn)
	var revisions int
	if err := f.harness.DB.QueryRowContext(context.Background(), `
SELECT count(*) FROM record_revisions WHERE record_id::text=$1 AND change_set_id::text=$2 AND row_version=$3
`, recordID, changeSetID, version).Scan(&revisions); err != nil {
		t.Fatal(err)
	}
	if revisions != 1 {
		t.Fatalf("attributed record revisions = %d, want 1", revisions)
	}
	collaborationsupport.RequireIntentCount(t, f.harness.DB, collaborationsupport.IntentSelector{
		IncidentID: f.incidentID, SourceRecordID: recordID, SourceChangeSetID: changeSetID, EventFamily: "record_changed", SourceRowVersion: &version,
	}, 1)
	return changeSetID
}
