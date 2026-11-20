package workbookscenariotest

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

// IncidentEffects extends MutationEffects to every effect class a rejected
// request must not commit (Core 04 AC-616): source, history, projection,
// idempotency, object, job and Collaboration. Each class is a content hash,
// so a diagnostic names the class without retaining stored values.
//
// Selections are incident scoped, except Receipts, which covers every route
// receipt because several receipt scopes (mentions, observations, Blobs,
// import sessions) carry no incident key. Callers therefore use a test-owned
// database, as appsupport.StartServer provides. Asynchronous Collaboration
// delivery bookkeeping is excluded, as in MutationEffects.
type IncidentEffects struct {
	// Mutation covers record envelopes, change sets, mutations, revisions,
	// incident- and record-scoped receipts, and publication intents.
	Mutation MutationEffects
	// Source is the authoritative owner snapshot of every incident record.
	Source [32]byte
	// Incident covers the incident row, its memberships and its Timeline
	// time-conversion profile.
	Incident [32]byte
	// Mentions covers the entity mentions of every incident record.
	Mentions [32]byte
	// Children covers Indicator observations and state intervals, record
	// links and record tags.
	Children    [32]byte
	Projections [32]byte
	Receipts    [32]byte
	// Objects covers object Blobs, their verifications and cleanup claims,
	// and Evidence access handles and custody events.
	Objects      [32]byte
	UploadLeases [32]byte
	// Imports covers sessions, units, source streams, unit projections, apply
	// plans, the apply journal, unit outcomes and committed sources.
	Imports [32]byte
	// Jobs covers incident-scoped Jobs and their commit proofs.
	Jobs [32]byte
	// Workspace covers saved views and incident and user workbook preferences.
	Workspace [32]byte
	// Reporting covers report compositions, their versions and preview
	// attempts, snapshots, releases and approvals, Reporting Job payloads, and
	// incident bundle exports and Job payloads.
	Reporting [32]byte
}

// SnapshotIncidentEffects observes every IncidentEffects class for one
// incident of the harness's test-owned database.
func SnapshotIncidentEffects(t testing.TB, h *appsupport.ServerHarness, incidentID uuid.UUID) IncidentEffects {
	t.Helper()
	ctx := context.Background()
	effects := IncidentEffects{Mutation: SnapshotIncidentMutationEffects(t, h.DB, incidentID.String())}

	rows, err := h.DB.QueryContext(ctx, `SELECT record_id FROM records WHERE incident_id=$1 ORDER BY record_id`, incidentID)
	if err != nil {
		t.Fatalf("list incident records: %v", err)
	}
	var recordIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			t.Fatalf("scan incident record: %v", err)
		}
		recordIDs = append(recordIDs, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatalf("list incident records: %v", err)
	}
	rows.Close()
	source, err := json.Marshal(revisionsupport.SnapshotRecords(t, h.Pool, h.Revisions.Appender(), recordIDs...))
	if err != nil {
		t.Fatalf("encode incident source snapshots: %v", err)
	}
	effects.Source = sha256.Sum256(source)

	classes := []struct {
		name   string
		target *[32]byte
		query  string
	}{
		{"incident", &effects.Incident, `SELECT jsonb_build_object(
 'incident',(SELECT coalesce(jsonb_agg(to_jsonb(i)),'[]') FROM incidents i WHERE id=$1),
 'memberships',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY to_jsonb(m)::text),'[]') FROM incident_memberships m WHERE incident_id=$1),
 'membership_versions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM incident_membership_versions v WHERE incident_id=$1),
 'time_conversion',(SELECT coalesce(jsonb_agg(to_jsonb(p)),'[]') FROM timeline_time_conversion_profiles p WHERE incident_id=$1)
)::text`},
		{"mentions", &effects.Mentions, `SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY m.entity_mention_id),'[]')::text
  FROM entity_mentions m JOIN records r ON r.record_id=m.source_record_id WHERE r.incident_id=$1`},
		{"children", &effects.Children, `SELECT jsonb_build_object(
 'observations',(SELECT coalesce(jsonb_agg(to_jsonb(o) ORDER BY to_jsonb(o)::text),'[]') FROM indicator_observations o WHERE incident_id=$1),
 'intervals',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM indicator_state_intervals v WHERE incident_id=$1),
 'links',(SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]') FROM record_links l WHERE incident_id=$1),
 'tags',(SELECT coalesce(jsonb_agg(to_jsonb(g) ORDER BY to_jsonb(g)::text),'[]') FROM record_tags g WHERE incident_id=$1)
)::text`},
		{"projections", &effects.Projections, `SELECT jsonb_build_object(
 'timeline',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM timeline_grid_projection p WHERE incident_id=$1),
 'hosts',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM host_grid_projection p WHERE incident_id=$1),
 'identities',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM identity_grid_projection p WHERE incident_id=$1),
 'indicators',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM indicator_grid_projection p WHERE incident_id=$1),
 'assessments',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM assessment_grid_projection p WHERE incident_id=$1),
 'artifacts',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM artifact_grid_projection p WHERE incident_id=$1),
 'evidence',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM evidence_grid_projection p WHERE incident_id=$1),
 'parties',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM party_grid_projection p WHERE incident_id=$1),
 'tasks',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM task_request_grid_projection p WHERE incident_id=$1),
 'decisions',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM decision_grid_projection p WHERE incident_id=$1)
)::text`},
		{"receipts", &effects.Receipts, `SELECT coalesce(jsonb_agg(to_jsonb(i) ORDER BY id),'[]')::text FROM route_idempotency i WHERE $1::uuid IS NOT NULL`},
		{"objects", &effects.Objects, `SELECT jsonb_build_object(
 'blobs',(SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY b.object_blob_id),'[]') FROM object_blobs b WHERE incident_id=$1),
 'verifications',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM evidence_blob_verifications v JOIN object_blobs b USING(object_blob_id) WHERE b.incident_id=$1),
 'cleanup',(SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY to_jsonb(c)::text),'[]') FROM evidence_blob_cleanup_claims c JOIN object_blobs b USING(object_blob_id) WHERE b.incident_id=$1),
 'handles',(SELECT coalesce(jsonb_agg(to_jsonb(h) ORDER BY to_jsonb(h)::text),'[]') FROM evidence_access_handles h WHERE incident_id=$1),
 'custody',(SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY to_jsonb(e)::text),'[]') FROM evidence_custody_events e WHERE incident_id=$1)
)::text`},
		{"upload leases", &effects.UploadLeases, `SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]')::text FROM evidence_object_upload_leases l WHERE incident_id=$1`},
		{"imports", &effects.Imports, `SELECT jsonb_build_object(
 'sessions',(SELECT coalesce(jsonb_agg(to_jsonb(s) ORDER BY s.import_session_id),'[]') FROM import_sessions s WHERE incident_id=$1),
 'units',(SELECT coalesce(jsonb_agg(to_jsonb(u) ORDER BY u.import_unit_id),'[]') FROM import_units u JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'streams',(SELECT coalesce(jsonb_agg(to_jsonb(x) ORDER BY x.source_stream_ref),'[]') FROM import_source_streams x JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'unit_projection',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM import_unit_projection_rows p JOIN import_units u USING(import_unit_id) JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'plans',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM import_apply_unit_plans p JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'journal',(SELECT coalesce(jsonb_agg(to_jsonb(j) ORDER BY to_jsonb(j)::text),'[]') FROM import_apply_journal j JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'outcomes',(SELECT coalesce(jsonb_agg(to_jsonb(o) ORDER BY to_jsonb(o)::text),'[]') FROM import_unit_apply_outcomes o JOIN import_sessions s USING(import_session_id) WHERE s.incident_id=$1),
 'committed',(SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY to_jsonb(c)::text),'[]') FROM import_committed_sources c WHERE incident_id=$1),
 'committed_records',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY to_jsonb(r)::text),'[]') FROM import_committed_source_records r WHERE incident_id=$1)
)::text`},
		{"workspace", &effects.Workspace, `SELECT jsonb_build_object(
 'saved_views',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY v.saved_view_id),'[]') FROM saved_views v WHERE incident_id=$1),
 'incident_preferences',(SELECT coalesce(jsonb_agg(to_jsonb(p)),'[]') FROM incident_workbook_preferences p WHERE incident_id=$1),
 'user_preferences',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.user_id),'[]') FROM user_workbook_preferences p WHERE incident_id=$1)
)::text`},
		{"reporting", &effects.Reporting, `SELECT jsonb_build_object(
 'compositions',(SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY c.composition_id),'[]') FROM report_compositions c WHERE incident_id=$1),
 'versions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM report_composition_versions v JOIN report_compositions c USING(composition_id) WHERE c.incident_id=$1),
 'previews',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.preview_attempt_id),'[]') FROM report_composition_preview_attempts p WHERE incident_id=$1),
 'snapshots',(SELECT coalesce(jsonb_agg(to_jsonb(s) ORDER BY s.snapshot_id),'[]') FROM reporting_snapshots s WHERE incident_id=$1),
 'releases',(SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY r.release_id),'[]') FROM reporting_releases r WHERE incident_id=$1),
 'approvals',(SELECT coalesce(jsonb_agg(to_jsonb(a) ORDER BY a.approval_id),'[]') FROM reporting_release_approvals a JOIN reporting_releases r USING(release_id) WHERE r.incident_id=$1),
 'payloads',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.job_id),'[]') FROM reporting_job_payloads p WHERE incident_id=$1),
 'bundle_exports',(SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY b.bundle_id),'[]') FROM incident_bundle_exports b WHERE incident_id=$1),
 'bundle_payloads',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.job_id),'[]') FROM incident_bundle_job_payloads p WHERE incident_id=$1)
)::text`},
		{"jobs", &effects.Jobs, `SELECT jsonb_build_object(
 'jobs',(SELECT coalesce(jsonb_agg(to_jsonb(j) ORDER BY j.job_id),'[]') FROM jobs j WHERE incident_id=$1),
 'proofs',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY p.job_id),'[]') FROM extension_job_commit_proofs p JOIN jobs j USING(job_id) WHERE j.incident_id=$1)
)::text`},
	}
	for _, class := range classes {
		var content string
		if err := h.DB.QueryRowContext(ctx, class.query, incidentID).Scan(&content); err != nil {
			t.Fatalf("snapshot incident %s effects: %v", class.name, err)
		}
		*class.target = sha256.Sum256([]byte(content))
	}
	return effects
}

// RequireIncidentEffectsUnchanged fails with the changed effect classes.
func RequireIncidentEffectsUnchanged(t testing.TB, name string, before, after IncidentEffects) {
	t.Helper()
	if changed := ChangedIncidentEffects(before, after); len(changed) != 0 {
		t.Fatalf("%s committed effects: %s", name, strings.Join(changed, ", "))
	}
}

// ChangedIncidentEffects names the classes that differ between two snapshots.
func ChangedIncidentEffects(before, after IncidentEffects) []string {
	var changed []string
	for _, class := range []struct {
		name          string
		before, after any
	}{
		{"source/history/receipts", before.Mutation.History, after.Mutation.History},
		{"collaboration publication", before.Mutation.PublicationIntents, after.Mutation.PublicationIntents},
		{"record source", before.Source, after.Source},
		{"incident", before.Incident, after.Incident},
		{"entity mentions", before.Mentions, after.Mentions},
		{"child objects/links/tags", before.Children, after.Children},
		{"projections", before.Projections, after.Projections},
		{"idempotency receipts", before.Receipts, after.Receipts},
		{"object Blobs", before.Objects, after.Objects},
		{"upload leases", before.UploadLeases, after.UploadLeases},
		{"imports", before.Imports, after.Imports},
		{"jobs", before.Jobs, after.Jobs},
		{"workspace", before.Workspace, after.Workspace},
		{"reporting", before.Reporting, after.Reporting},
	} {
		if class.before != class.after {
			changed = append(changed, class.name)
		}
	}
	return changed
}
