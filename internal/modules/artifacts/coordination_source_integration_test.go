package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestCoordinationSourceAtomicCreation(t *testing.T) {
	ctx := context.Background()
	h := appsupport.StartStore(t, "coordination-source-atomic")
	actor := authstoretest.SeedLocalUserRecord(t, h.DB, "coordination@example.test", "Coordination Author", "CoordinationAuthor1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "coordination-incident", "IR-CCA", "Coordination")
	facade, err := workbookassembly.NewArtifactMutationContribution(h.DB, conflicttest.NewCodec("coordination"), revisionsupport.MustAppender(t), revisionsupport.MustConflictFieldResolver(t), appsupport.ArtifactProjectionRows(h.DB), collaborationsupport.NewRecordChangedAppender())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	command := func(view, id string, source any, supplied bool) artifacts.CreateCommand {
		fields := coordinationMinimum(view, actor.ID)
		fields["client_txn_id"] = id
		if supplied {
			fields["coordination.source_record_id"] = source
		}
		raw, err := json.Marshal(fields)
		if err != nil {
			t.Fatal(err)
		}
		admission, admissionErr := artifacts.AdmitCreate(view, strings.NewReader(string(raw)))
		if admissionErr != nil {
			t.Fatal(admissionErr)
		}
		return artifacts.CreateCommand{IncidentID: incident.ID, ActorUserID: actor.ID, Admission: admission, RequestID: "req-" + id, Now: now}
	}
	sourceArtifact := func(view string) uuid.UUID {
		result, err := facade.Create(ctx, command(view, "source-"+uuid.NewString(), nil, false))
		if err != nil {
			t.Fatal(err)
		}
		return result.RecordID
	}
	sources := map[string]uuid.UUID{"timeline": seedLinkedNoteSource(t, h, incident.ID, actor.ID, "timeline_event", now), "task": seedLinkedNoteSource(t, h, incident.ID, actor.ID, "task_request", now), "decision": seedLinkedNoteSource(t, h, incident.ID, actor.ID, "decision", now), "comm_log": sourceArtifact(artifacts.CommLogViewSchemaID), "handoff": sourceArtifact(artifacts.HandoffViewSchemaID), "status_review": sourceArtifact(artifacts.StatusReviewViewSchemaID)}
	matrix := map[string][]string{"timeline": {artifacts.CommLogViewSchemaID, artifacts.HandoffViewSchemaID, artifacts.StatusReviewViewSchemaID, artifacts.LessonViewSchemaID}, "task": {artifacts.CommLogViewSchemaID, artifacts.StatusReviewViewSchemaID, artifacts.LessonViewSchemaID}, "decision": {artifacts.CommLogViewSchemaID, artifacts.StatusReviewViewSchemaID}, "comm_log": {artifacts.StatusReviewViewSchemaID}, "handoff": {artifacts.StatusReviewViewSchemaID}, "status_review": {artifacts.CommLogViewSchemaID}}
	count := 0
	for surface, views := range matrix {
		for _, view := range views {
			count++
			t.Run(surface+"_"+view, func(t *testing.T) {
				cmd := command(view, "matrix-"+surface+view, sources[surface].String(), true)
				result, err := facade.Create(ctx, cmd)
				if err != nil {
					t.Fatal(err)
				}
				if result.ContextualLink == nil || result.ContextualLink.SourceRecordID != sources[surface] || result.ContextualLink.LinkType != "references_artifact" || result.Row == nil || result.ChangeSetID == nil {
					t.Fatalf("incomplete acceptance: %#v", result)
				}
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1 AND src_record_id=$2 AND dst_record_id=$3 AND field_key IS NULL AND link_type='references_artifact' AND provenance='manual' AND confidence IS NULL AND deleted_at IS NULL`, incident.ID, sources[surface], result.RecordID, 1)
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, result.ChangeSetID, 1)
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=1`, result.RecordID, 1)
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM artifact_grid_projection WHERE record_id=$1`, result.RecordID, 1)
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM change_sets WHERE change_set_id=$1 AND actor_user_id=$2 AND source='workbook.rows.create'`, result.ChangeSetID, actor.ID, 1)
				replay, err := facade.Create(ctx, cmd)
				if err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.RecordID != result.RecordID || replay.ContextualLink == nil || *replay.ContextualLink != *result.ContextualLink {
					t.Fatalf("replay mismatch: %#v, %v", replay, err)
				}
				conflicting := command(view, cmd.Admission.ClientTxnID(), nil, true)
				if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
					t.Fatalf("changed context did not conflict: %v", err)
				}
			})
		}
	}
	if count != 12 {
		t.Fatalf("matrix size %d", count)
	}
	t.Run("cleared and omitted context remain unlinked", func(t *testing.T) {
		for _, supplied := range []bool{false, true} {
			result, err := facade.Create(ctx, command(artifacts.LessonViewSchemaID, fmt.Sprintf("unlinked-%v", supplied), nil, supplied))
			if err != nil || result.ContextualLink != nil {
				t.Fatalf("unlinked: %#v %v", result, err)
			}
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE dst_record_id=$1`, result.RecordID, 0)
		}
	})
	t.Run("invalid foreign deleted and unsupported source fail atomically", func(t *testing.T) {
		foreign := appsupport.CreateIncidentInStore(t, h.DB, actor, "coordination-foreign", "IR-CCA-FOREIGN", "Foreign")
		foreignSource := seedLinkedNoteSource(t, h, foreign.ID, actor.ID, "timeline_event", now)
		deleted := seedLinkedNoteSource(t, h, incident.ID, actor.ID, "timeline_event", now)
		if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, deleted, now, actor.ID); err != nil {
			t.Fatal(err)
		}
		invalid := []struct {
			source uuid.UUID
			target string
		}{{uuid.New(), artifacts.LessonViewSchemaID}, {foreignSource, artifacts.LessonViewSchemaID}, {deleted, artifacts.LessonViewSchemaID}, {sources["task"], artifacts.HandoffViewSchemaID}, {sources["decision"], artifacts.LessonViewSchemaID}, {sources["comm_log"], artifacts.CommLogViewSchemaID}, {seedLinkedNoteSource(t, h, incident.ID, actor.ID, "host", now), artifacts.CommLogViewSchemaID}}
		before := linkedNoteCount(t, h, `SELECT count(*) FROM artifacts WHERE incident_id=$1`, incident.ID)
		for i, item := range invalid {
			cmd := command(item.target, fmt.Sprintf("invalid-%d", i), item.source.String(), true)
			_, err := facade.Create(ctx, cmd)
			var validation *artifacts.ValidationError
			if !errors.As(err, &validation) || validation.Field != "coordination.source_record_id" {
				t.Fatalf("invalid source: %v", err)
			}
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM route_idempotency WHERE client_txn_id=$1`, cmd.Admission.ClientTxnID(), 0)
		}
		requireLinkedNoteCount(t, h, `SELECT count(*) FROM artifacts WHERE incident_id=$1`, incident.ID, before)
	})
	t.Run("source association and analyst references share one transaction", func(t *testing.T) {
		for _, valid := range []bool{true, false} {
			targetID := sources["task"]
			if !valid {
				targetID = uuid.New()
			}
			fields := coordinationMinimum(artifacts.LessonViewSchemaID, actor.ID)
			fields["client_txn_id"] = fmt.Sprintf("combined-references-%v", valid)
			fields["coordination.source_record_id"] = sources["timeline"].String()
			fields["lesson.follow_up_task_ids"] = map[string]any{"kind": "collection_actions_v1", "actions": []any{
				map[string]any{"op": "add_record_ref", "linked_record_id": targetID.String()},
			}}
			raw, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			admission, admissionErr := artifacts.AdmitCreate(artifacts.LessonViewSchemaID, strings.NewReader(string(raw)))
			if admissionErr != nil {
				t.Fatal(admissionErr)
			}
			before := linkedNoteCount(t, h, `SELECT count(*) FROM records WHERE incident_id=$1`, incident.ID)
			links := linkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1`, incident.ID)
			result, err := facade.Create(ctx, artifacts.CreateCommand{IncidentID: incident.ID, ActorUserID: actor.ID, Admission: admission, RequestID: "combined-references", Now: now})
			if !valid {
				if err == nil {
					t.Fatal("unavailable target reference accepted")
				}
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM records WHERE incident_id=$1`, incident.ID, before)
				requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1`, incident.ID, links)
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2 AND field_key='lesson.follow_up_task_ids' AND deleted_at IS NULL`, result.RecordID, targetID, 1)
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2 AND field_key IS NULL AND deleted_at IS NULL`, sources["timeline"], result.RecordID, 1)
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1 AND target_kind='record_link'`, result.ChangeSetID, 2)
		}
	})
	t.Run("concurrent exact requests converge on one receipt", func(t *testing.T) {
		cmd := command(artifacts.LessonViewSchemaID, "concurrent", sources["timeline"].String(), true)
		var wg sync.WaitGroup
		results := make([]artifacts.MutationResult, 4)
		errs := make([]error, 4)
		for i := range results {
			wg.Add(1)
			go func() { defer wg.Done(); results[i], errs[i] = facade.Create(ctx, cmd) }()
		}
		wg.Wait()
		for i, result := range results {
			if errs[i] != nil || result.RecordID != results[0].RecordID {
				t.Fatalf("concurrent result %d: %#v %v", i, result, errs[i])
			}
		}
		requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE src_record_id=$1 AND dst_record_id=$2`, sources["timeline"], results[0].RecordID, 1)
	})
	for _, table := range []string{"record_links", "artifact_grid_projection", "change_sets", "record_revisions", "route_idempotency"} {
		t.Run("rollback_"+table, func(t *testing.T) {
			before := linkedNoteCount(t, h, `SELECT count(*) FROM records WHERE incident_id=$1`, incident.ID)
			links := linkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1`, incident.ID)
			installLinkedNoteFailureTrigger(t, h, table, "coordination_"+table)
			cmd := command(artifacts.LessonViewSchemaID, "fault-"+table, sources["timeline"].String(), true)
			if _, err := facade.Create(ctx, cmd); err == nil {
				t.Fatal("fault committed")
			}
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM records WHERE incident_id=$1`, incident.ID, before)
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1`, incident.ID, links)
			requireLinkedNoteCount(t, h, `SELECT count(*) FROM route_idempotency WHERE client_txn_id=$1`, cmd.Admission.ClientTxnID(), 0)
		})
	}
	t.Run("committed replay precedes closure and deleted source validation", func(t *testing.T) {
		cmd := command(artifacts.LessonViewSchemaID, "late-replay", sources["timeline"].String(), true)
		result, err := facade.Create(ctx, cmd)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, sources["timeline"], now, actor.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := h.DB.Exec(ctx, `UPDATE incidents SET status='closed',closed_at=now() WHERE id=$1`, incident.ID); err != nil {
			t.Fatal(err)
		}
		replay, err := facade.Create(ctx, cmd)
		if err != nil || replay.RecordID != result.RecordID || replay.ContextualLink == nil {
			t.Fatalf("late replay %#v %v", replay, err)
		}
		if _, err := facade.Create(ctx, command(artifacts.LessonViewSchemaID, "fresh-after-close", sources["timeline"].String(), true)); err == nil {
			t.Fatal("fresh closed create admitted")
		}
	})
}

func coordinationMinimum(view string, actor uuid.UUID) map[string]any {
	switch view {
	case artifacts.CommLogViewSchemaID:
		return map[string]any{"comm_log.comm_type": "briefing", "comm_log.audience": "Team", "comm_log.channel_or_meeting": "Call", "comm_log.summary": "Summary"}
	case artifacts.HandoffViewSchemaID:
		return map[string]any{"handoff.incoming_owner_user_id": actor.String(), "handoff.current_state_summary": "State"}
	case artifacts.StatusReviewViewSchemaID:
		return map[string]any{"status_review.current_state_summary": "State"}
	case artifacts.LessonViewSchemaID:
		return map[string]any{"lesson.summary": "Lesson"}
	default:
		panic("unsupported coordination target")
	}
}
