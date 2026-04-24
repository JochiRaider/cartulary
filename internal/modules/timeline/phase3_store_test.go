package timeline_test

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/phase3test"
	"github.com/JochiRaider/cartulary/internal/testutil/timelinetest"
)

func TestPhase3_CreateCommitsAndAssignsIdentity_U_3_01(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-01")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U301", "txn-phase3-u-3-01-incident")

	summary := "First capture"
	request := timeline.CreateRequest{
		ClientTxnID: "txn-phase3-u-3-01-row",
		Summary:     &summary,
	}
	result, err := store.CreateRow(context.Background(), actor, incidentID, request, timeline.TimelineCreateRequestHash(request), "req-phase3-u-3-01-row", phase3BaseTime())
	if err != nil {
		t.Fatalf("create row: %v", err)
	}

	if result.StatusCode != http.StatusCreated {
		t.Fatalf("unexpected status code: got %d want %d", result.StatusCode, http.StatusCreated)
	}
	if result.RecordID == uuid.Nil || result.ChangeSetID == uuid.Nil {
		t.Fatalf("expected committed record and change set identifiers, got %#v", result)
	}
	if result.RowVersion != 1 {
		t.Fatalf("expected row_version=1, got %d", result.RowVersion)
	}

	row := result.Payload["row"].(map[string]any)
	if row["record_id"] != result.RecordID.String() || row["row_version"] != int64(1) {
		t.Fatalf("unexpected create payload row: %#v", row)
	}
	if row["cells"].(map[string]any)["timeline.summary"].(map[string]any)["value"] != summary {
		t.Fatalf("expected committed summary in payload row, got %#v", row)
	}

	if got := phase3test.QueryCount(t, harness.DB, `SELECT COUNT(*) FROM timeline_events WHERE record_id = $1`, result.RecordID); got != 1 {
		t.Fatalf("expected one durable timeline row, got %d", got)
	}
	projection := timelinetest.LookupProjectionRow(t, harness.DB, result.RecordID.String())
	if projection.RowVersion != 1 || projection.CaptureState != "rough" {
		t.Fatalf("unexpected projection row after create: %#v", projection)
	}
}

func TestPhase3_InitialCreateState_U_3_02(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-02")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U302", "txn-phase3-u-3-02-incident")

	summary := "Rough capture"
	request := timeline.CreateRequest{
		ClientTxnID: "txn-phase3-u-3-02-row",
		Summary:     &summary,
	}
	result, err := store.CreateRow(context.Background(), actor, incidentID, request, timeline.TimelineCreateRequestHash(request), "req-phase3-u-3-02-row", phase3BaseTime())
	if err != nil {
		t.Fatalf("create row: %v", err)
	}

	row := result.Payload["row"].(map[string]any)
	if got := row["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "rough" {
		t.Fatalf("expected create payload capture_state rough, got %#v", row)
	}
	substrate, err := store.SnapshotRecordSubstrate(context.Background(), result.RecordID)
	if err != nil {
		t.Fatalf("snapshot substrate: %v", err)
	}
	if substrate.CaptureState != "rough" || substrate.RowVersion != 1 || substrate.ReplacementRecordID != nil {
		t.Fatalf("unexpected substrate after create: %#v", substrate)
	}
}

func TestPhase3_CaptureStateLifecycle_U_3_03(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-03")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U303", "txn-phase3-u-3-03-incident")

	roughRow := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-03-rough", "rough row", phase3BaseTime())
	roughReview := timeline.ActionRequest{
		BaseRowVersion: 1,
		ClientTxnID:    "txn-phase3-u-3-03-review-rough",
	}
	reviewedRough, err := store.MarkReviewed(context.Background(), actor, roughRow.RecordID, roughReview, timeline.TimelineActionRequestHash(roughReview.BaseRowVersion, roughReview.ClientTxnID, roughReview.Reason, nil), "req-phase3-u-3-03-review-rough", phase3BaseTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("mark rough row reviewed: %v", err)
	}
	if reviewedRough.Payload["capture_state"] != "reviewed" || reviewedRough.RowVersion != 2 {
		t.Fatalf("unexpected reviewed rough payload: %#v", reviewedRough.Payload)
	}

	enrichedRow := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-03-enriched", "enriched row", phase3BaseTime().Add(2*time.Minute))
	patch := timeline.PatchRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		BaseRowVersion: 1,
		ClientTxnID:    "txn-phase3-u-3-03-patch",
		CanonicalChange: []timeline.PatchChange{
			{FieldKey: "timeline.details", TextValue: storeStringPtr("material edit")},
		},
	}
	patched, err := store.PatchRow(context.Background(), actor, enrichedRow.RecordID, patch, timeline.TimelinePatchRequestHash(patch), "req-phase3-u-3-03-patch", phase3BaseTime().Add(3*time.Minute))
	if err != nil {
		t.Fatalf("patch rough row: %v", err)
	}
	if got := patched.Payload["row"].(map[string]any)["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("expected rough -> enriched patch transition, got %#v", patched.Payload)
	}

	enrichedReview := timeline.ActionRequest{
		BaseRowVersion: patched.RowVersion,
		ClientTxnID:    "txn-phase3-u-3-03-review-enriched",
	}
	reviewedEnriched, err := store.MarkReviewed(context.Background(), actor, enrichedRow.RecordID, enrichedReview, timeline.TimelineActionRequestHash(enrichedReview.BaseRowVersion, enrichedReview.ClientTxnID, enrichedReview.Reason, nil), "req-phase3-u-3-03-review-enriched", phase3BaseTime().Add(4*time.Minute))
	if err != nil {
		t.Fatalf("mark enriched row reviewed: %v", err)
	}
	if reviewedEnriched.Payload["capture_state"] != "reviewed" || reviewedEnriched.RowVersion != 3 {
		t.Fatalf("unexpected reviewed enriched payload: %#v", reviewedEnriched.Payload)
	}
}

func TestPhase3_ReviewedDemotionAndSupersedeTerminality_U_3_04(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-04")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U304", "txn-phase3-u-3-04-incident")

	row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-04-row", "lifecycle row", phase3BaseTime())
	review := timeline.ActionRequest{
		BaseRowVersion: 1,
		ClientTxnID:    "txn-phase3-u-3-04-review",
	}
	reviewed, err := store.MarkReviewed(context.Background(), actor, row.RecordID, review, timeline.TimelineActionRequestHash(review.BaseRowVersion, review.ClientTxnID, review.Reason, nil), "req-phase3-u-3-04-review", phase3BaseTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("mark reviewed: %v", err)
	}
	if reviewed.Payload["capture_state"] != "reviewed" {
		t.Fatalf("expected reviewed state, got %#v", reviewed.Payload)
	}

	demotionPatch := timeline.PatchRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		BaseRowVersion: reviewed.RowVersion,
		ClientTxnID:    "txn-phase3-u-3-04-demote",
		CanonicalChange: []timeline.PatchChange{
			{FieldKey: "timeline.summary", TextValue: storeStringPtr("demoted after edit")},
		},
	}
	demoted, err := store.PatchRow(context.Background(), actor, row.RecordID, demotionPatch, timeline.TimelinePatchRequestHash(demotionPatch), "req-phase3-u-3-04-demote", phase3BaseTime().Add(2*time.Minute))
	if err != nil {
		t.Fatalf("patch reviewed row: %v", err)
	}
	if got := demoted.Payload["row"].(map[string]any)["cells"].(map[string]any)["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
		t.Fatalf("expected reviewed row demotion to enriched, got %#v", demoted.Payload)
	}

	reviewAgain := timeline.ActionRequest{
		BaseRowVersion: demoted.RowVersion,
		ClientTxnID:    "txn-phase3-u-3-04-review-again",
	}
	reviewedAgain, err := store.MarkReviewed(context.Background(), actor, row.RecordID, reviewAgain, timeline.TimelineActionRequestHash(reviewAgain.BaseRowVersion, reviewAgain.ClientTxnID, reviewAgain.Reason, nil), "req-phase3-u-3-04-review-again", phase3BaseTime().Add(3*time.Minute))
	if err != nil {
		t.Fatalf("mark reviewed again: %v", err)
	}

	replacement := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-04-replacement", "replacement row", phase3BaseTime().Add(4*time.Minute))
	supersede := timeline.SupersedeRequest{
		BaseRowVersion:      reviewedAgain.RowVersion,
		ClientTxnID:         "txn-phase3-u-3-04-supersede",
		Reason:              "superseded by a better row",
		ReplacementRecordID: &replacement.RecordID,
	}
	superseded, err := store.Supersede(context.Background(), actor, row.RecordID, supersede, timeline.TimelineActionRequestHash(supersede.BaseRowVersion, supersede.ClientTxnID, &supersede.Reason, supersede.ReplacementRecordID), "req-phase3-u-3-04-supersede", phase3BaseTime().Add(5*time.Minute))
	if err != nil {
		t.Fatalf("supersede row: %v", err)
	}
	if superseded.Payload["capture_state"] != "superseded" {
		t.Fatalf("expected superseded action payload, got %#v", superseded.Payload)
	}

	patchAfterSupersede := timeline.PatchRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		BaseRowVersion: superseded.RowVersion,
		ClientTxnID:    "txn-phase3-u-3-04-after-supersede",
		CanonicalChange: []timeline.PatchChange{
			{FieldKey: "timeline.details", TextValue: storeStringPtr("blocked")},
		},
	}
	if _, err := store.PatchRow(context.Background(), actor, row.RecordID, patchAfterSupersede, timeline.TimelinePatchRequestHash(patchAfterSupersede), "req-phase3-u-3-04-after-supersede", phase3BaseTime().Add(6*time.Minute)); !errors.Is(err, timeline.ErrIllegalTransition) {
		t.Fatalf("expected superseded rows to reject ordinary patch semantics, got %v", err)
	}
}

func TestPhase3_PatchReplayStability_U_3_07(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-07")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U307", "txn-phase3-u-3-07-incident")

	row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-07-row", "replay row", phase3BaseTime())
	patch := timeline.PatchRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		BaseRowVersion: 1,
		ClientTxnID:    "txn-phase3-u-3-07-patch",
		CanonicalChange: []timeline.PatchChange{
			{FieldKey: "timeline.details", TextValue: storeStringPtr("details")},
			{FieldKey: "timeline.summary", TextValue: storeStringPtr("summary")},
		},
	}
	first, err := store.PatchRow(context.Background(), actor, row.RecordID, patch, timeline.TimelinePatchRequestHash(patch), "req-phase3-u-3-07-patch", phase3BaseTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("initial patch: %v", err)
	}
	beforeReplay := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())

	replay, err := store.PatchRow(context.Background(), actor, row.RecordID, patch, timeline.TimelinePatchRequestHash(patch), "req-phase3-u-3-07-patch-replay", phase3BaseTime().Add(2*time.Minute))
	if err != nil {
		t.Fatalf("replay patch: %v", err)
	}
	if !replay.Replayed || replay.StatusCode != http.StatusOK || replay.Payload["change_set_id"] != first.Payload["change_set_id"] {
		t.Fatalf("expected replay to return original committed payload, got %#v", replay)
	}
	afterReplay := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())
	if beforeReplay != afterReplay {
		t.Fatalf("replay must not create additional history rows: before=%#v after=%#v", beforeReplay, afterReplay)
	}

	divergent := patch
	divergent.CanonicalChange = []timeline.PatchChange{
		{FieldKey: "timeline.summary", TextValue: storeStringPtr("different")},
	}
	if _, err := store.PatchRow(context.Background(), actor, row.RecordID, divergent, timeline.TimelinePatchRequestHash(divergent), "req-phase3-u-3-07-divergent", phase3BaseTime().Add(3*time.Minute)); !errors.Is(err, authn.ErrClientTxnConflict) {
		t.Fatalf("expected divergent replay conflict, got %v", err)
	}
}

func TestPhase3_PatchFieldLevelConcurrency_U_3_11(t *testing.T) {
	t.Run("stale different-field patch rebases onto current row", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-11-rebase")
		store, actor, incidentID := newPhase3StoreFixture(t, harness, "U311R", "txn-phase3-u-3-11-rebase-incident")

		row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-11-rebase-row", "base summary", phase3BaseTime())
		serverPatch := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-rebase-server",
			CanonicalChange: []timeline.PatchChange{
				{FieldKey: "timeline.summary", TextValue: storeStringPtr("server summary")},
			},
		}
		serverResult, err := store.PatchRow(context.Background(), actor, row.RecordID, serverPatch, timeline.TimelinePatchRequestHash(serverPatch), "req-phase3-u-3-11-rebase-server", phase3BaseTime().Add(time.Minute))
		if err != nil {
			t.Fatalf("server patch: %v", err)
		}

		stalePatch := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-rebase-client",
			CanonicalChange: []timeline.PatchChange{
				{FieldKey: "timeline.details", TextValue: storeStringPtr("client details")},
			},
		}
		rebased, err := store.PatchRow(context.Background(), actor, row.RecordID, stalePatch, timeline.TimelinePatchRequestHash(stalePatch), "req-phase3-u-3-11-rebase-client", phase3BaseTime().Add(2*time.Minute))
		if err != nil {
			t.Fatalf("stale different-field patch should rebase: %v", err)
		}
		if rebased.RowVersion != serverResult.RowVersion+1 {
			t.Fatalf("expected rebased patch to advance once from current version, got %d after %d", rebased.RowVersion, serverResult.RowVersion)
		}
		cells := rebased.Payload["row"].(map[string]any)["cells"].(map[string]any)
		if got := cells["timeline.summary"].(map[string]any)["value"]; got != "server summary" {
			t.Fatalf("expected rebased row to preserve server summary, got %#v", cells["timeline.summary"])
		}
		if got := cells["timeline.details"].(map[string]any)["value"]; got != "client details" {
			t.Fatalf("expected rebased row to include client details, got %#v", cells["timeline.details"])
		}
	})

	t.Run("stale same-field text patch returns conflict payload without writes", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-11-text-conflict")
		store, actor, incidentID := newPhase3StoreFixture(t, harness, "U311T", "txn-phase3-u-3-11-text-conflict-incident")

		row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-11-text-conflict-row", "base summary", phase3BaseTime())
		serverPatch := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-text-conflict-server",
			CanonicalChange: []timeline.PatchChange{
				{FieldKey: "timeline.summary", TextValue: storeStringPtr("server summary")},
			},
		}
		serverResult, err := store.PatchRow(context.Background(), actor, row.RecordID, serverPatch, timeline.TimelinePatchRequestHash(serverPatch), "req-phase3-u-3-11-text-conflict-server", phase3BaseTime().Add(time.Minute))
		if err != nil {
			t.Fatalf("server patch: %v", err)
		}
		beforeConflict := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())

		stalePatch := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-text-conflict-client",
			CanonicalChange: []timeline.PatchChange{
				{FieldKey: "timeline.summary", TextValue: storeStringPtr("client summary")},
			},
		}
		_, err = store.PatchRow(context.Background(), actor, row.RecordID, stalePatch, timeline.TimelinePatchRequestHash(stalePatch), "req-phase3-u-3-11-text-conflict-client", phase3BaseTime().Add(2*time.Minute))
		var conflict *timeline.SameFieldConflictError
		if !errors.As(err, &conflict) {
			t.Fatalf("expected same-field conflict, got %v", err)
		}
		if conflict.Conflict["record_id"] != row.RecordID.String() ||
			conflict.Conflict["field_key"] != "timeline.summary" ||
			conflict.Conflict["conflict_resolution_class"] != "text_compare_merge" ||
			conflict.Conflict["base_value"] != "base summary" ||
			conflict.Conflict["server_value"] != "server summary" ||
			conflict.Conflict["client_value"] != "client summary" ||
			conflict.Conflict["server_updated_by"] != actor.ID.String() {
			t.Fatalf("unexpected same-field conflict payload: %#v", conflict.Conflict)
		}
		if conflict.Conflict["base_row_version"] != row.RowVersion || conflict.Conflict["current_row_version"] != serverResult.RowVersion || conflict.Conflict["conflict_token"] == "" {
			t.Fatalf("missing same-field conflict version/token fields: %#v", conflict.Conflict)
		}
		afterConflict := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())
		if beforeConflict != afterConflict {
			t.Fatalf("same-field conflict must not create writes: before=%#v after=%#v", beforeConflict, afterConflict)
		}
		if got := phase3test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM route_idempotency
 WHERE route_key = 'timeline.records.patch'
   AND actor_user_id::text = $1
   AND scope_key = $2
   AND client_txn_id = $3
`, actor.ID.String(), row.RecordID.String(), stalePatch.ClientTxnID); got != 0 {
			t.Fatalf("same-field conflict must not persist idempotency row, got %d", got)
		}
	})

	t.Run("stale same-field collection patch reports collection values", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-11-collection-conflict")
		store, actor, incidentID := newPhase3StoreFixture(t, harness, "U311C", "txn-phase3-u-3-11-collection-conflict-incident")

		row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-11-collection-conflict-row", "collection row", phase3BaseTime())
		serverPatch := decodeStorePatchRequest(t, `{
			"view_schema_id": "cartulary.view.timeline.v1",
			"base_row_version": 1,
			"client_txn_id": "txn-phase3-u-3-11-collection-conflict-server",
			"changes": [
				{
					"field_key": "timeline.host_refs",
					"action_payload": {
						"kind": "collection_actions_v1",
						"actions": [{ "op": "add_token", "raw_text": "Server Host" }]
					}
				}
			]
		}`)
		serverResult, err := store.PatchRow(context.Background(), actor, row.RecordID, serverPatch, timeline.TimelinePatchRequestHash(serverPatch), "req-phase3-u-3-11-collection-conflict-server", phase3BaseTime().Add(time.Minute))
		if err != nil {
			t.Fatalf("server collection patch: %v", err)
		}

		stalePatch := decodeStorePatchRequest(t, `{
			"view_schema_id": "cartulary.view.timeline.v1",
			"base_row_version": 1,
			"client_txn_id": "txn-phase3-u-3-11-collection-conflict-client",
			"changes": [
				{
					"field_key": "timeline.host_refs",
					"action_payload": {
						"kind": "collection_actions_v1",
						"actions": [{ "op": "add_token", "raw_text": "Client Host" }]
					}
				}
			]
		}`)
		_, err = store.PatchRow(context.Background(), actor, row.RecordID, stalePatch, timeline.TimelinePatchRequestHash(stalePatch), "req-phase3-u-3-11-collection-conflict-client", phase3BaseTime().Add(2*time.Minute))
		var conflict *timeline.SameFieldConflictError
		if !errors.As(err, &conflict) {
			t.Fatalf("expected collection same-field conflict, got %v", err)
		}
		if conflict.Conflict["field_key"] != "timeline.host_refs" ||
			conflict.Conflict["conflict_resolution_class"] != "collection_review" ||
			conflict.Conflict["base_row_version"] != row.RowVersion ||
			conflict.Conflict["current_row_version"] != serverResult.RowVersion {
			t.Fatalf("unexpected collection conflict metadata: %#v", conflict.Conflict)
		}
		requireCollectionConflictValue(t, conflict.Conflict["base_value"], "")
		requireCollectionConflictValue(t, conflict.Conflict["server_value"], "Server Host")
		requireCollectionConflictValue(t, conflict.Conflict["client_value"], "Client Host")
	})

	t.Run("stale patch after lifecycle-only change applies against current lifecycle", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-11-lifecycle-rebase")
		store, actor, incidentID := newPhase3StoreFixture(t, harness, "U311L", "txn-phase3-u-3-11-lifecycle-rebase-incident")

		row := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-11-lifecycle-rebase-row", "lifecycle row", phase3BaseTime())
		review := timeline.ActionRequest{
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-lifecycle-rebase-review",
		}
		reviewed, err := store.MarkReviewed(context.Background(), actor, row.RecordID, review, timeline.TimelineActionRequestHash(review.BaseRowVersion, review.ClientTxnID, review.Reason, nil), "req-phase3-u-3-11-lifecycle-rebase-review", phase3BaseTime().Add(time.Minute))
		if err != nil {
			t.Fatalf("mark reviewed: %v", err)
		}

		stalePatch := timeline.PatchRequest{
			ViewSchemaID:   timeline.TimelineViewSchemaID,
			BaseRowVersion: row.RowVersion,
			ClientTxnID:    "txn-phase3-u-3-11-lifecycle-rebase-patch",
			CanonicalChange: []timeline.PatchChange{
				{FieldKey: "timeline.details", TextValue: storeStringPtr("stale lifecycle edit")},
			},
		}
		rebased, err := store.PatchRow(context.Background(), actor, row.RecordID, stalePatch, timeline.TimelinePatchRequestHash(stalePatch), "req-phase3-u-3-11-lifecycle-rebase-patch", phase3BaseTime().Add(2*time.Minute))
		if err != nil {
			t.Fatalf("stale lifecycle-only patch should apply: %v", err)
		}
		if rebased.RowVersion != reviewed.RowVersion+1 {
			t.Fatalf("expected lifecycle rebase to advance once, got %d after %d", rebased.RowVersion, reviewed.RowVersion)
		}
		cells := rebased.Payload["row"].(map[string]any)["cells"].(map[string]any)
		if got := cells["timeline.capture_state"].(map[string]any)["value"]; got != "enriched" {
			t.Fatalf("expected patch against reviewed current state to demote to enriched, got %#v", cells["timeline.capture_state"])
		}
	})
}

func TestPhase3_CreateAndPatchWriteHistory_U_3_09(t *testing.T) {
	harness := phase3test.StartStore(t, "phase3-u-3-09")
	store, actor, incidentID := newPhase3StoreFixture(t, harness, "U309", "txn-phase3-u-3-09-incident")

	createSummary := "history row"
	createRequest := timeline.CreateRequest{
		ClientTxnID: "txn-phase3-u-3-09-create",
		Summary:     &createSummary,
	}
	created, err := store.CreateRow(context.Background(), actor, incidentID, createRequest, timeline.TimelineCreateRequestHash(createRequest), "req-phase3-u-3-09-create", phase3BaseTime())
	if err != nil {
		t.Fatalf("create row: %v", err)
	}
	requirePhase3MutationRecorded(t, harness.DB, created.ChangeSetID.String(), created.RecordID.String(), actor.ID.String(), "timeline.rows.create", createRequest.ClientTxnID, 1, 1)

	patch := timeline.PatchRequest{
		ViewSchemaID:   timeline.TimelineViewSchemaID,
		BaseRowVersion: 1,
		ClientTxnID:    "txn-phase3-u-3-09-patch",
		CanonicalChange: []timeline.PatchChange{
			{FieldKey: "timeline.details", TextValue: storeStringPtr("patched")},
		},
	}
	patched, err := store.PatchRow(context.Background(), actor, created.RecordID, patch, timeline.TimelinePatchRequestHash(patch), "req-phase3-u-3-09-patch", phase3BaseTime().Add(time.Minute))
	if err != nil {
		t.Fatalf("patch row: %v", err)
	}
	requirePhase3MutationRecorded(t, harness.DB, patched.ChangeSetID.String(), created.RecordID.String(), actor.ID.String(), "timeline.records.patch", patch.ClientTxnID, 1, 2)
}

func TestPhase3_SupersedeReplayAndRollbackCoupling_U_3_10(t *testing.T) {
	t.Run("illegal targets and replay stay fail-closed and single-write", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-10")
		store, actor, incidentID := newPhase3StoreFixture(t, harness, "U310", "txn-phase3-u-3-10-incident")
		otherIncident := phase3test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase3-u-3-10-other-incident", "IR-U310X", "Phase 3 U-3-10 other")

		row := createReviewedTimelineRow(t, store, actor, incidentID, "txn-phase3-u-3-10-row", "row", phase3BaseTime())
		replacement := createTimelineSummaryRow(t, store, actor, incidentID, "txn-phase3-u-3-10-replacement", "replacement", phase3BaseTime().Add(time.Minute))
		crossIncidentReplacement := createTimelineSummaryRow(t, store, actor, otherIncident.ID, "txn-phase3-u-3-10-cross", "cross", phase3BaseTime().Add(2*time.Minute))

		self := timeline.SupersedeRequest{
			BaseRowVersion:      row.RowVersion,
			ClientTxnID:         "txn-phase3-u-3-10-self",
			Reason:              "self replacement",
			ReplacementRecordID: &row.RecordID,
		}
		if _, err := store.Supersede(context.Background(), actor, row.RecordID, self, timeline.TimelineActionRequestHash(self.BaseRowVersion, self.ClientTxnID, &self.Reason, self.ReplacementRecordID), "req-phase3-u-3-10-self", phase3BaseTime().Add(3*time.Minute)); !errors.Is(err, timeline.ErrIllegalTransition) {
			t.Fatalf("expected self supersede to fail, got %v", err)
		}

		crossIncident := timeline.SupersedeRequest{
			BaseRowVersion:      row.RowVersion,
			ClientTxnID:         "txn-phase3-u-3-10-cross-incident",
			Reason:              "cross incident replacement",
			ReplacementRecordID: &crossIncidentReplacement.RecordID,
		}
		if _, err := store.Supersede(context.Background(), actor, row.RecordID, crossIncident, timeline.TimelineActionRequestHash(crossIncident.BaseRowVersion, crossIncident.ClientTxnID, &crossIncident.Reason, crossIncident.ReplacementRecordID), "req-phase3-u-3-10-cross-incident", phase3BaseTime().Add(4*time.Minute)); !errors.Is(err, timeline.ErrIllegalTransition) {
			t.Fatalf("expected cross-incident supersede to fail, got %v", err)
		}

		supersede := timeline.SupersedeRequest{
			BaseRowVersion:      row.RowVersion,
			ClientTxnID:         "txn-phase3-u-3-10-supersede",
			Reason:              "superseded by a better row",
			ReplacementRecordID: &replacement.RecordID,
		}
		first, err := store.Supersede(context.Background(), actor, row.RecordID, supersede, timeline.TimelineActionRequestHash(supersede.BaseRowVersion, supersede.ClientTxnID, &supersede.Reason, supersede.ReplacementRecordID), "req-phase3-u-3-10-supersede", phase3BaseTime().Add(5*time.Minute))
		if err != nil {
			t.Fatalf("supersede row: %v", err)
		}
		timelinetest.RequireSupersedeCoupledChangeSet(t, harness.DB, first.ChangeSetID.String(), row.RecordID.String(), replacement.RecordID.String(), first.RowVersion)
		if got := timelinetest.CountActiveSupersedesLinks(t, harness.DB, incidentID.String(), replacement.RecordID.String(), row.RecordID.String()); got != 1 {
			t.Fatalf("expected one active supersedes link, got %d", got)
		}
		beforeReplay := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())

		replay, err := store.Supersede(context.Background(), actor, row.RecordID, supersede, timeline.TimelineActionRequestHash(supersede.BaseRowVersion, supersede.ClientTxnID, &supersede.Reason, supersede.ReplacementRecordID), "req-phase3-u-3-10-supersede-replay", phase3BaseTime().Add(6*time.Minute))
		if err != nil {
			t.Fatalf("replay supersede: %v", err)
		}
		if !replay.Replayed || replay.Payload["change_set_id"] != first.Payload["change_set_id"] {
			t.Fatalf("expected replay to return original committed payload, got %#v", replay)
		}
		afterReplay := timelinetest.SnapshotCounters(t, harness.DB, incidentID.String(), row.RecordID.String())
		if beforeReplay != afterReplay {
			t.Fatalf("replay must not create additional history rows: before=%#v after=%#v", beforeReplay, afterReplay)
		}
		if got := timelinetest.CountActiveSupersedesLinks(t, harness.DB, incidentID.String(), replacement.RecordID.String(), row.RecordID.String()); got != 1 {
			t.Fatalf("replay must keep one active supersedes link, got %d", got)
		}

		divergent := supersede
		divergent.ReplacementRecordID = &crossIncidentReplacement.RecordID
		if _, err := store.Supersede(context.Background(), actor, row.RecordID, divergent, timeline.TimelineActionRequestHash(divergent.BaseRowVersion, divergent.ClientTxnID, &divergent.Reason, divergent.ReplacementRecordID), "req-phase3-u-3-10-divergent", phase3BaseTime().Add(7*time.Minute)); !errors.Is(err, authn.ErrClientTxnConflict) {
			t.Fatalf("expected divergent supersede replay conflict, got %v", err)
		}
	})

	t.Run("rollback removes lifecycle and replacement link together", func(t *testing.T) {
		harness := phase3test.StartStore(t, "phase3-u-3-10-rollback")
		actor := phase3test.SeedLocalUserFlags(t, harness.DB, "phase3-U310R@example.test", "Phase3 U310R", "Phase3Pass1!", false, false, true)
		incident := phase3test.CreateIncidentInStore(t, harness.Pool, actor, "txn-phase3-u-3-10-rollback-incident", "IR-U310R", "Phase 3 U310R")

		restoreHooks := timeline.SetStoreHooksForTesting(timeline.StoreHooks{
			BeforeCommit: func(routeKey string, recordID uuid.UUID) error {
				if routeKey == "timeline.records.supersede" {
					return errors.New("forced rollback")
				}
				return nil
			},
		})
		defer restoreHooks()
		store := timeline.NewStore(harness.Pool)
		row := createReviewedTimelineRow(t, store, actor, incident.ID, "txn-phase3-u-3-10-rollback-row", "row", phase3BaseTime())
		replacement := createTimelineSummaryRow(t, store, actor, incident.ID, "txn-phase3-u-3-10-rollback-replacement", "replacement", phase3BaseTime().Add(time.Minute))

		request := timeline.SupersedeRequest{
			BaseRowVersion:      row.RowVersion,
			ClientTxnID:         "txn-phase3-u-3-10-rollback",
			Reason:              "rollback supersede",
			ReplacementRecordID: &replacement.RecordID,
		}
		beforeRollback := timelinetest.SnapshotCounters(t, harness.DB, incident.ID.String(), row.RecordID.String())
		if _, err := store.Supersede(context.Background(), actor, row.RecordID, request, timeline.TimelineActionRequestHash(request.BaseRowVersion, request.ClientTxnID, &request.Reason, request.ReplacementRecordID), "req-phase3-u-3-10-rollback", phase3BaseTime().Add(2*time.Minute)); err == nil {
			t.Fatal("expected supersede rollback error")
		}
		afterRollback := timelinetest.SnapshotCounters(t, harness.DB, incident.ID.String(), row.RecordID.String())
		if beforeRollback != afterRollback {
			t.Fatalf("rollback must keep counters stable, before=%+v after=%+v", beforeRollback, afterRollback)
		}

		substrate, err := store.SnapshotRecordSubstrate(context.Background(), row.RecordID)
		if err != nil {
			t.Fatalf("snapshot substrate after rollback: %v", err)
		}
		if substrate.CaptureState != "reviewed" || substrate.ReplacementRecordID != nil {
			t.Fatalf("rollback must preserve reviewed substrate without replacement link, got %#v", substrate)
		}
		if got := timelinetest.CountActiveSupersedesLinks(t, harness.DB, incident.ID.String(), replacement.RecordID.String(), row.RecordID.String()); got != 0 {
			t.Fatalf("rollback must not leave an active supersedes link, got %d", got)
		}
	})
}

func newPhase3StoreFixture(t testing.TB, harness *phase3test.StoreHarness, suffix string, incidentTxn string) (*timeline.Store, authn.UserRecord, uuid.UUID) {
	t.Helper()

	actor := phase3test.SeedLocalUserFlags(t, harness.DB, "phase3-"+suffix+"@example.test", "Phase3 "+suffix, "Phase3Pass1!", false, false, true)
	incident := phase3test.CreateIncidentInStore(t, harness.Pool, actor, incidentTxn, "IR-"+suffix, "Phase 3 "+suffix)
	return timeline.NewStore(harness.Pool), actor, incident.ID
}

func createTimelineSummaryRow(t testing.TB, store *timeline.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, summary string, now time.Time) timeline.MutationResult {
	t.Helper()

	request := timeline.CreateRequest{
		ClientTxnID: clientTxnID,
		Summary:     &summary,
	}
	result, err := store.CreateRow(context.Background(), actor, incidentID, request, timeline.TimelineCreateRequestHash(request), "req-"+clientTxnID, now)
	if err != nil {
		t.Fatalf("create summary row: %v", err)
	}
	return result
}

func createReviewedTimelineRow(t testing.TB, store *timeline.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, summary string, now time.Time) timeline.MutationResult {
	t.Helper()

	created := createTimelineSummaryRow(t, store, actor, incidentID, clientTxnID, summary, now)
	review := timeline.ActionRequest{
		BaseRowVersion: created.RowVersion,
		ClientTxnID:    clientTxnID + "-review",
	}
	reviewed, err := store.MarkReviewed(context.Background(), actor, created.RecordID, review, timeline.TimelineActionRequestHash(review.BaseRowVersion, review.ClientTxnID, review.Reason, nil), "req-"+review.ClientTxnID, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("mark reviewed: %v", err)
	}
	return reviewed
}

func requirePhase3MutationRecorded(t testing.TB, db *sql.DB, changeSetID string, recordID string, wantActorUserID string, wantSource string, wantClientTxnID string, wantMutationRows int, wantRevisions int) {
	t.Helper()

	changeSet := timelinetest.LookupChangeSet(t, db, changeSetID)
	if changeSet.ActorUserID != wantActorUserID || changeSet.Source != wantSource || changeSet.ClientTxnID != wantClientTxnID || changeSet.RequestID == "" || changeSet.CreatedAt.IsZero() {
		t.Fatalf("unexpected mutation attribution: %#v", changeSet)
	}
	if got := timelinetest.CountChangeSetMutations(t, db, changeSetID); got != wantMutationRows {
		t.Fatalf("unexpected mutation row count for %s: got %d want %d", changeSetID, got, wantMutationRows)
	}
	if got := timelinetest.CountRecordRevisions(t, db, recordID); got != wantRevisions {
		t.Fatalf("unexpected record revision count for %s: got %d want %d", recordID, got, wantRevisions)
	}
}

func decodeStorePatchRequest(t testing.TB, body string) timeline.PatchRequest {
	t.Helper()

	request, apiErr := timeline.DecodeTimelinePatchRequest(strings.NewReader(body))
	if apiErr != nil {
		t.Fatalf("decode patch request: %#v", apiErr)
	}
	return request
}

func requireCollectionConflictValue(t testing.TB, value any, wantRawText string) {
	t.Helper()

	collection, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("expected collection conflict value object, got %T", value)
	}
	if collection["kind"] != "collection_value_v1" {
		t.Fatalf("expected collection_value_v1, got %#v", collection)
	}
	items, ok := collection["items"].([]map[string]any)
	if !ok {
		rawItems, rawOK := collection["items"].([]any)
		if !rawOK {
			t.Fatalf("expected collection items array, got %T", collection["items"])
		}
		items = make([]map[string]any, 0, len(rawItems))
		for _, rawItem := range rawItems {
			item, itemOK := rawItem.(map[string]any)
			if !itemOK {
				t.Fatalf("expected collection item object, got %T", rawItem)
			}
			items = append(items, item)
		}
	}
	if wantRawText == "" {
		if len(items) != 0 {
			t.Fatalf("expected empty base collection, got %#v", collection)
		}
		return
	}
	if len(items) != 1 {
		t.Fatalf("expected one collection item, got %#v", collection)
	}
	if items[0]["raw_text"] != wantRawText || items[0]["item_kind"] != "unresolved_mention" {
		t.Fatalf("unexpected collection item: %#v", items[0])
	}
}

func phase3BaseTime() time.Time {
	return time.Date(2026, 4, 20, 12, 0, 0, 0, time.UTC)
}

func storeStringPtr(value string) *string {
	return &value
}
