package performancefixture

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	collaborationfixture "github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/performancefixture"
	evidencefixture "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/performancefixture"
	projectionsfixture "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport/performancefixture"
	timelinefixture "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/performancefixture"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	fixture "github.com/JochiRaider/cartulary/internal/testutil/performancefixture"
)

func TestEvidenceInspectorFixturePreservesOwnerAttachments_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "evidence_inspector_fixture")
	objects, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer objects.Close()
	owners, err := NewOwners(harness.DB, objects, conflicttest.NewCodec("fixture-evidence"))
	if err != nil {
		t.Fatal(err)
	}
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "fixture-inspector@example.test", "Inspector fixture", "EvidenceFixture1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-INSPECTOR-FIXTURE", "Inspector fixture")
	application, err := timelinefixture.NewEvidenceInspectorApplication(owners.Timeline, owners.TimelineMutations, actor)
	if err != nil {
		t.Fatal(err)
	}
	// These are physical objects, but deliberately small: this test qualifies
	// owner construction and attachment scale, not the 20,971,520,000-byte performance load.
	descriptor := fixture.Descriptor{ContributionID: "evidence.objects.v1", Version: "evidence.objects.v1", OwnerID: "module.evidence", ExpectedCounts: map[string]int{"evidence_records": 100, "object_blobs": 100, "total_object_bytes": 25600}}
	evidenceProvider, err := evidencefixture.New(owners.Evidence, actor.ID, descriptor, 256, "Inspector Evidence")
	if err != nil {
		t.Fatal(err)
	}
	descriptor.ExpectedCounts["evidence_records"] = 1
	if evidenceProvider.Descriptor().ExpectedCounts["evidence_records"] != 100 {
		t.Fatal("provider retained mutable descriptor")
	}
	state := &fixture.BuildState{IncidentID: incident.ID.String()}
	started := time.Now()
	receipt, err := evidenceProvider.Apply(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Evidence owner construction: %s", time.Since(started))
	if receipt.Counts["evidence_records"] != 100 || receipt.Counts["total_object_bytes"] != 25600 {
		t.Fatal("incorrect Evidence receipt")
	}
	recordIDs := evidenceProvider.RecordIDs()
	copyIDs := evidenceProvider.RecordIDs()
	copyIDs[0] = uuid.Nil
	if len(recordIDs) != 100 || evidenceProvider.RecordIDs()[0] != recordIDs[0] {
		t.Fatal("private identities are missing or mutable")
	}
	if _, err := evidenceProvider.Apply(ctx, state); err == nil {
		t.Fatal("Evidence contribution retry accepted")
	}
	for _, invalid := range [][]uuid.UUID{nil, {uuid.Nil}, {recordIDs[0], recordIDs[0]}, make([]uuid.UUID, 101)} {
		if _, err := application.Create(ctx, incident.ID, 5000, invalid); err == nil {
			t.Fatal("invalid attachment population accepted")
		}
	}
	for _, invalid := range []int{0, 1, 100001} {
		if _, err := application.Create(ctx, incident.ID, invalid, recordIDs); err == nil {
			t.Fatal("invalid Timeline population accepted")
		}
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := application.Create(cancelled, incident.ID, 5000, recordIDs); err == nil {
		t.Fatal("cancelled fixture accepted")
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE incident_id=$1`, incident.ID); got != 0 {
		t.Fatal("invalid construction mutated Timeline")
	}
	inspectorProvider, err := timelinefixture.NewInspectorProvider(application, evidenceProvider, fixture.Descriptor{
		ContributionID: "timeline.evidence_inspector.v1", Version: "timeline.evidence_inspector.v1", OwnerID: "module.timeline",
		Dependencies: []string{"evidence.objects.v1"}, ExpectedCounts: map[string]int{"timeline_rows": 5000, "attached_evidence": 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	started = time.Now()
	inspectorReceipt, err := inspectorProvider.Apply(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Timeline owner construction: %s", time.Since(started))
	if inspectorReceipt.Counts["timeline_rows"] != 5000 || inspectorReceipt.Counts["attached_evidence"] != 100 {
		t.Fatal("incorrect Timeline receipt")
	}
	if _, err := inspectorProvider.Apply(ctx, state); err == nil {
		t.Fatal("Timeline contribution retry accepted")
	}
	target := inspectorProvider.Target()
	started = time.Now()
	if target.RecordID == uuid.Nil || target.RowVersion != 2 || target.Summary == "" {
		t.Fatal("missing target identity")
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE incident_id=$1`, incident.ID); got != 5000 {
		t.Fatalf("Timeline count = %d", got)
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM timeline_grid_projection WHERE incident_id=$1`, incident.ID); got != 5000 {
		t.Fatalf("Timeline projection count = %d", got)
	}
	var summary string
	var version int64
	var evidenceCount, refCount int
	if err := harness.DB.QueryRow(ctx, `SELECT activity_synopsis_text,row_version,evidence_count,jsonb_array_length(attached_evidence_refs)
 FROM timeline_grid_projection WHERE incident_id=$1 AND record_id=$2`, incident.ID, target.RecordID).Scan(&summary, &version, &evidenceCount, &refCount); err != nil {
		t.Fatal(err)
	}
	if summary != target.Summary || version != target.RowVersion || evidenceCount != 100 || refCount != 100 {
		t.Fatal("Timeline target projection diverged")
	}
	for _, id := range recordIDs {
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE incident_id=$1 AND src_record_id=$2 AND dst_record_id=$3 AND link_type='attached_evidence' AND field_key='timeline.attached_evidence_ids' AND deleted_at IS NULL`, incident.ID, target.RecordID, id); got != 1 {
			t.Fatal("missing or duplicate owner-created Evidence link")
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM evidence_grid_projection WHERE record_id=$1 AND incident_id=$2 AND linked_record_count=1 AND row_version=1`, id, incident.ID); got != 1 {
			t.Fatal("Evidence attachment projection diverged")
		}
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE record_id=$1 AND row_version IN (1,2)`, target.RecordID); got != 2 {
		t.Fatal("missing create or attachment patch revision")
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id IN (SELECT change_set_id FROM record_revisions WHERE record_id=$1) AND target_kind='record_link'`, target.RecordID); got != 100 {
		t.Fatalf("attachment mutation history = %d", got)
	}
	validator := evidenceHeavyValidator{pool: harness.DB, evidence: evidenceProvider, timeline: inspectorProvider, workload: evidenceHeavyWorkload{
		TimelineRows: 5000, EvidenceRows: 100, ObjectCount: 100, ObjectBytes: 256, TotalObjectBytes: 25600,
		TargetSummary: target.Summary, LinkedEvidenceCount: 100,
	}}
	before, err := validator.Validate(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	pending := before.Counts["pending_publication_intents"]
	if pending < 100 || before.Conditions["construction_publications_sequenced"] {
		t.Fatal("unsequenced construction events qualified as quiescent")
	}
	publications, err := collaborationfixture.New(harness.DB, fixture.Descriptor{
		ContributionID: "collaboration.fixture_publications.v1", Version: "collaboration.fixture_publications.v1", OwnerID: "module.collaboration",
		Dependencies: []string{"projections.evidence_timeline.v1"}, ExpectedCounts: map[string]int{"pending_intents": 0},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := publications.Apply(cancelled, state); err == nil {
		t.Fatal("cancelled publication finalization accepted")
	}
	// Unready work must reject immediately rather than advance time or retry.
	if _, err := harness.DB.Exec(ctx, `UPDATE collaboration_event_intents SET next_attempt_at=now()+interval '1 hour' WHERE dispatch_state='pending'`); err != nil {
		t.Fatal(err)
	}
	if _, err := publications.Apply(ctx, state); err == nil {
		t.Fatal("stalled publication sequencing accepted")
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_replay_events`); got != 0 {
		t.Fatal("stalled finalization sequenced construction history")
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE collaboration_event_intents SET next_attempt_at=now() WHERE dispatch_state='pending'`); err != nil {
		t.Fatal(err)
	}
	publicationReceipt, err := publications.Apply(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	if publicationReceipt.Counts["pending_intents"] != 0 {
		t.Fatal("publication receipt retained pending work")
	}
	for range 2 {
		if _, err := publications.Apply(ctx, state); err != nil {
			t.Fatal(err)
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE dispatch_state='sequenced'`); got != pending {
			t.Fatal("construction intents lost or duplicated")
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_replay_events`); got != pending {
			t.Fatal("construction replay events lost or duplicated")
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT high_water_stream_seq FROM collaboration_incident_stream_cursors WHERE incident_id=$1`, incident.ID); got != pending {
			t.Fatal("restored live-tail boundary differs from construction history")
		}
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE collaboration_incident_stream_cursors SET quarantined_at=now(),quarantine_reason='invalid_event_payload' WHERE incident_id=$1`, incident.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := publications.Apply(ctx, state); err == nil {
		t.Fatal("quarantined construction stream accepted")
	}
	quarantined, err := validator.Validate(ctx, state)
	if err != nil || quarantined.Conditions["construction_publications_sequenced"] {
		t.Fatal("quarantined stream passed independent validation", err)
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE collaboration_incident_stream_cursors SET quarantined_at=NULL,quarantine_reason=NULL WHERE incident_id=$1`, incident.ID); err != nil {
		t.Fatal(err)
	}
	t.Logf("Construction publications retained and sequenced before sealing: %d", pending)
	validation, err := validator.Validate(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	for condition, passed := range validation.Conditions {
		if !passed {
			t.Fatalf("Evidence-heavy semantic condition %s failed", condition)
		}
	}
	for key, expected := range map[string]int{"timeline_rows": 5000, "evidence_rows": 100, "object_blobs": 100, "total_object_bytes": 25600, "attached_evidence": 100, "analyst_accounts": 0, "active_sessions": 0, "pending_publication_intents": 0} {
		if validation.Counts[key] != expected {
			t.Fatalf("Evidence-heavy %s=%d want=%d", key, validation.Counts[key], expected)
		}
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE incident_workbook_preferences SET default_sheet_ref='{}'::jsonb WHERE incident_id=$1`, incident.ID); err != nil {
		t.Fatal(err)
	}
	altered, err := validator.Validate(ctx, state)
	if err != nil {
		t.Fatal(err)
	}
	if altered.Conditions["default_view"] {
		t.Fatal("non-default workbook preferences qualified as the default view")
	}
	if _, err := harness.DB.Exec(ctx, `UPDATE incident_workbook_preferences SET default_sheet_ref=NULL WHERE incident_id=$1`, incident.ID); err != nil {
		t.Fatal(err)
	}
	projectionApplication, err := projectionsfixture.NewProductionApplication(owners.Projections)
	if err != nil {
		t.Fatal(err)
	}
	if err := projectionApplication.ValidateFixtureProjectionSets(ctx, incident.ID.String(), []projectionsfixture.SetExpectation{
		{ViewSchemaID: "cartulary.view.evidence.v1", ExactRows: 100},
		{ViewSchemaID: "cartulary.view.timeline.v2", ExactRows: 5000},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := application.Create(ctx, incident.ID, 5000, recordIDs); err == nil {
		t.Fatal("existing fixture overwritten")
	}

	// A valid UUID is insufficient: the ordinary Evidence owner rejects missing
	// and cross-incident references atomically with the selected-row create.
	for _, id := range []uuid.UUID{uuid.New(), recordIDs[0]} {
		other := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-"+uuid.NewString(), "Rejected inspector fixture")
		if _, err := application.Create(ctx, other.ID, 2, []uuid.UUID{id}); err == nil {
			t.Fatal("invalid Evidence reference accepted")
		}
		// Only the explicit background prestate exists; no target, link or history
		// commits. The complete fixture builder must dispose of this prestate.
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE incident_id=$1`, other.ID); got != 1 {
			t.Fatal("failed target create partially committed")
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE incident_id=$1`, other.ID); got != 0 {
			t.Fatal("failed target retained attachment")
		}
		if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM record_revisions v JOIN records r USING (record_id) WHERE r.incident_id=$1`, other.ID); got != 0 {
			t.Fatal("failed target retained history")
		}
	}
	other := appsupport.CreateIncidentInStore(t, harness.DB, actor, uuid.NewString(), "IR-OTHER-FIXTURE", "Unrelated fixture")
	otherEvidence, err := evidencefixture.New(owners.Evidence, actor.ID, fixture.Descriptor{
		ContributionID: "evidence.objects.v1", Version: "evidence.objects.v1", OwnerID: "module.evidence",
		ExpectedCounts: map[string]int{"evidence_records": 1, "object_blobs": 1, "total_object_bytes": 256},
	}, 256, "Other Evidence")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := otherEvidence.Apply(ctx, &fixture.BuildState{IncidentID: other.ID.String()}); err != nil {
		t.Fatal(err)
	}
	if _, err := publications.Apply(ctx, state); err == nil {
		t.Fatal("unrelated pending publications accepted")
	}
	if got := appsupport.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM collaboration_event_intents WHERE incident_id=$1 AND dispatch_state='pending'`, other.ID); got != 1 {
		t.Fatal("finalization modified unrelated publications")
	}
	t.Logf("Independent fixture validation: %s", time.Since(started))
}
