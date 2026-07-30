package revisions_test

import (
	"context"
	"database/sql"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	assessmenttest "github.com/JochiRaider/cartulary/internal/modules/assessments/testsupport"
	platformws "github.com/JochiRaider/cartulary/internal/modules/collaboration"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	linktest "github.com/JochiRaider/cartulary/internal/modules/links/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestDeleteRestoreRollbackAtomicConsequences_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-i-7-01-delete-restore")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-I701")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	indicatorID := seedIndicatorRecord(t, harness.DB, incidentID, actorID)
	hubChanges, unsubscribe := harness.Server.Runtime.CollaborationHub.SubscribeIncident(incidentID, 16)
	defer unsubscribe()

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 0, 0, 0, time.UTC))
	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-01-delete-host"}), http.StatusOK)["data"].(map[string]any)
	requireDeleteRestoreRecordChange(t, asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second), recordID, 2, "remove", "cartulary.view.hosts.v1")
	deleteChangeSetID := deletePayload["change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'records.delete' AND actor_user_id = $2`, deleteChangeSetID, actorID) != 1 {
		t.Fatalf("delete did not create attributed change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record' AND target_id = $2 AND operation_kind = 'soft_delete'`, deleteChangeSetID, recordID.String()) != 1 {
		t.Fatalf("delete did not create reversible soft_delete mutation")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 2`, deleteChangeSetID, recordID) != 1 {
		t.Fatalf("delete did not append row revision")
	}
	if rows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), "cartulary.view.hosts.v1", login); len(rows) != 0 {
		t.Fatalf("deleted host remained in ordinary view rows: %#v", rows)
	}
	historyAfterDelete := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	if historyAfterDelete[0].(map[string]any)["operation"] != "soft_delete" {
		t.Fatalf("latest history item must be soft_delete, got %#v", historyAfterDelete[0])
	}

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 1, 0, 0, time.UTC))
	restorePayload := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-i-7-01-restore-host"}), http.StatusOK)["data"].(map[string]any)
	requireDeleteRestoreRecordChange(t, asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second), recordID, 3, "invalidate", "cartulary.view.hosts.v1")
	restoreChangeSetID := restorePayload["change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'records.restore' AND actor_user_id = $2`, restoreChangeSetID, actorID) != 1 {
		t.Fatalf("restore did not create attributed change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'record' AND target_id = $2 AND operation_kind = 'restore'`, restoreChangeSetID, recordID.String()) != 1 {
		t.Fatalf("restore did not create reversible restore mutation")
	}
	if rows := workbookscenariotest.QueryViewRows(t, harness.Server.HTTP.URL, incidentID.String(), "cartulary.view.hosts.v1", login); len(rows) != 1 || rows[0]["record_id"] != recordID.String() {
		t.Fatalf("restored host was not returned to ordinary view rows: %#v", rows)
	}
	historyAfterRestore := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	if historyAfterRestore[0].(map[string]any)["operation"] != "restore" || historyAfterRestore[1].(map[string]any)["operation"] != "soft_delete" {
		t.Fatalf("delete/restore history was not append-only newest-first: %#v", historyAfterRestore)
	}

	rollbackRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, rollbackRecordID, "Rollback Host", "rollback-host", "", "")
	rollbackTargetChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000701")
	seedRollbackHostPatch(t, harness.DB, incidentID, rollbackRecordID, actorID, rollbackTargetChangeSet, time.Date(2026, 5, 10, 13, 2, 0, 0, time.UTC), "rollback before", "rollback after")
	rollbackRef := stringField(t, historyItems(getHistory(t, harness.Server.HTTP.URL, login, rollbackRecordID, ""))[0], "history_entry_ref")
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 3, 0, 0, time.UTC))
	beforeHistoryRefRows := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, rollbackRecordID)
	rollbackPayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, rollbackRecordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-i-7-01-rollback-host",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": rollbackRef},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChange(t, asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second), rollbackRecordID, 3, "cartulary.view.hosts.v1")
	rollbackChangeSetID := rollbackPayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2`, rollbackChangeSetID, actorID) != 1 {
		t.Fatalf("rollback did not create attributed rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'host' AND target_id = $2 AND operation_kind = 'rollback'`, rollbackChangeSetID, rollbackRecordID.String()) != 1 {
		t.Fatalf("rollback did not append inverse mutation")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 3`, rollbackChangeSetID, rollbackRecordID) != 1 {
		t.Fatalf("rollback did not append record revision")
	}
	if got := hostDisplayName(t, harness.DB, rollbackRecordID); got != "rollback before" {
		t.Fatalf("rollback did not update source row, got %q", got)
	}
	if got := hostProjectionDisplayName(t, harness.DB, rollbackRecordID); got != "rollback before" {
		t.Fatalf("rollback did not update projection row, got %q", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, rollbackRecordID) != beforeHistoryRefRows {
		t.Fatalf("rollback mutated prior history_entry_ref rows")
	}

	linkSrc, linkDst, linkID := seedRollbackRecordLinkCreate(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000702"))
	linkRollbackRef := historyEntryRefForTarget(t, harness, login, linkSrc, "record_link", linkID.String())
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 4, 0, 0, time.UTC))
	linkRollbackPayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, linkSrc, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-i-7-01-rollback-link",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": linkRollbackRef},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChangesAnyOrder(t, []platformws.RecordChange{
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
	}, map[uuid.UUID]int64{linkSrc: 2, linkDst: 2}, "cartulary.view.hosts.v1")
	requireAffectedRecords(t, linkRollbackPayload, linkSrc, linkDst)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NOT NULL`, linkID) != 1 {
		t.Fatalf("link rollback did not tombstone active link")
	}
	linkRollbackChangeSetID := linkRollbackPayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1`, linkRollbackChangeSetID) != 2 {
		t.Fatalf("link rollback did not append revisions for both affected records")
	}

	wholeLeft, wholeRight := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Integration Whole Left", "Integration Whole Right")
	wholeChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000704")
	seedRollbackTwoHostChangeSet(t, harness.DB, incidentID, actorID, wholeChangeSetID, wholeLeft, wholeRight)
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 5, 0, 0, time.UTC))
	wholeRollbackPayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, wholeLeft, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-i-7-01-rollback-whole-change-set",
		"target":           map[string]any{"kind": "change_set", "change_set_id": wholeChangeSetID.String()},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChangesAnyOrder(t, []platformws.RecordChange{
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
	}, map[uuid.UUID]int64{wholeLeft: 3, wholeRight: 3}, "cartulary.view.hosts.v1")
	requireAffectedRecords(t, wholeRollbackPayload, wholeLeft, wholeRight)
	if got := hostDisplayName(t, harness.DB, wholeLeft); got != "left before" {
		t.Fatalf("whole change-set rollback did not update left source row, got %q", got)
	}
	if got := hostDisplayName(t, harness.DB, wholeRight); got != "right before" {
		t.Fatalf("whole change-set rollback did not update right source row, got %q", got)
	}
	wholeRollbackChangeSetID := wholeRollbackPayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2`, wholeRollbackChangeSetID, actorID) != 1 {
		t.Fatalf("whole change-set rollback did not create attributed rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND operation_kind = 'rollback'`, wholeRollbackChangeSetID) != 2 {
		t.Fatalf("whole change-set rollback did not append inverse mutations")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND row_version = 3`, wholeRollbackChangeSetID) != 2 {
		t.Fatalf("whole change-set rollback did not append revisions for all affected records")
	}

	rowRestoreID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, rowRestoreID, "Integration row restore seed", "integration-row-restore", "", "")
	seedHostProjection(t, harness.DB, incidentID, rowRestoreID)
	rowRestoreTargetChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000705")
	seedRollbackHostPatch(t, harness.DB, incidentID, rowRestoreID, actorID, rowRestoreTargetChangeSetID, time.Date(2026, 5, 10, 13, 5, 30, 0, time.UTC), "integration row before", "integration row snapshot")
	mustExec(t, harness.DB, `UPDATE records SET row_version = 3, updated_by_user_id = $2 WHERE record_id = $1`, rowRestoreID, actorID)
	mustExec(t, harness.DB, `UPDATE hosts SET display_name = 'integration row current', row_version = 3, updated_by_user_id = $2 WHERE record_id = $1`, rowRestoreID, actorID)
	seedHostProjection(t, harness.DB, incidentID, rowRestoreID)
	rowRestoreLinkID := seedRecordLinkForRowRestore(t, harness.DB, incidentID, rowRestoreID, actorID)
	rowRestoreTagID := seedRecordTag(t, harness.DB, incidentID, rowRestoreID, actorID)
	beforeRowRestoreLinks := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, rowRestoreLinkID)
	beforeRowRestoreTags := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id = $1 AND deleted_at IS NULL`, rowRestoreTagID)
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 5, 45, 0, time.UTC))
	rowRestorePayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, rowRestoreID, map[string]any{
		"base_row_version": 3,
		"client_txn_id":    "txn-i-7-01-row-restore",
		"target":           map[string]any{"kind": "row_restore", "restore_to_revision_no": 2},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChange(t, asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second), rowRestoreID, 4, "cartulary.view.hosts.v1")
	requireAffectedRecords(t, rowRestorePayload, rowRestoreID)
	if got := hostDisplayName(t, harness.DB, rowRestoreID); got != "integration row snapshot" {
		t.Fatalf("row restore did not update source row, got %q", got)
	}
	if got := hostProjectionDisplayName(t, harness.DB, rowRestoreID); got != "integration row snapshot" {
		t.Fatalf("row restore did not update projection row, got %q", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, rowRestoreLinkID) != beforeRowRestoreLinks {
		t.Fatalf("row restore mutated active link state")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id = $1 AND deleted_at IS NULL`, rowRestoreTagID) != beforeRowRestoreTags {
		t.Fatalf("row restore mutated active tag state")
	}
	rowRestoreRollbackChangeSetID := rowRestorePayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2`, rowRestoreRollbackChangeSetID, actorID) != 1 {
		t.Fatalf("row restore did not create attributed rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND operation_kind = 'row_restore'`, rowRestoreRollbackChangeSetID) != 1 {
		t.Fatalf("row restore did not append row_restore mutation")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 4`, rowRestoreRollbackChangeSetID, rowRestoreID) != 1 {
		t.Fatalf("row restore did not append row revision")
	}

	attachedCreateSrc, attachedCreateEvidence, attachedCreateLink := seedRollbackAttachedEvidenceLinkCreate(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000706"))
	attachedCreateRef := historyEntryRefForTarget(t, harness, login, attachedCreateSrc, "record_link", attachedCreateLink.String())
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 6, 0, 0, time.UTC))
	attachedCreatePayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, attachedCreateSrc, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-i-7-01-rollback-attached-evidence-create",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": attachedCreateRef},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChangesByRecord(t, []platformws.RecordChange{
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
	}, map[uuid.UUID]rollbackRecordChangeExpectation{
		attachedCreateSrc: {
			rowVersion:       2,
			viewSchemaID:     "cartulary.view.timeline.v2",
			changedFieldKeys: []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"},
		},
		attachedCreateEvidence: {
			rowVersion:       2,
			viewSchemaID:     "cartulary.view.evidence.v1",
			changedFieldKeys: []string{"evidence.linked_record_count"},
		},
	})
	requireAffectedRecordsCanonical(t, attachedCreatePayload, attachedCreateSrc, attachedCreateEvidence)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NOT NULL`, attachedCreateLink) != 1 {
		t.Fatalf("attached-evidence create rollback did not tombstone link")
	}
	if gotCount, gotHasEvidence := timelineProjectionEvidenceState(t, harness.DB, attachedCreateSrc); gotCount != 0 || gotHasEvidence {
		t.Fatalf("attached-evidence create rollback projection got count=%d has_evidence=%v want count=0 has_evidence=false", gotCount, gotHasEvidence)
	}
	attachedCreateRollbackChangeSetID := attachedCreatePayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2`, attachedCreateRollbackChangeSetID, actorID) != 1 {
		t.Fatalf("attached-evidence create rollback did not create attributed rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND row_version = 2`, attachedCreateRollbackChangeSetID) != 2 {
		t.Fatalf("attached-evidence create rollback did not append revisions for both records")
	}

	attachedDeleteSrc, attachedDeleteEvidence, attachedDeleteLink := seedRollbackAttachedEvidenceLinkDelete(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000707"))
	attachedDeleteRef := historyEntryRefForTarget(t, harness, login, attachedDeleteSrc, "record_link", attachedDeleteLink.String())
	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 13, 7, 0, 0, time.UTC))
	attachedDeletePayload := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, attachedDeleteSrc, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-i-7-01-rollback-attached-evidence-delete",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": attachedDeleteRef},
	}), http.StatusOK)["data"].(map[string]any)
	requireRollbackRecordChangesByRecord(t, []platformws.RecordChange{
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
		asserttest.AwaitRecordChange(t, hubChanges, 5*time.Second),
	}, map[uuid.UUID]rollbackRecordChangeExpectation{
		attachedDeleteSrc: {
			rowVersion:       2,
			viewSchemaID:     "cartulary.view.timeline.v2",
			changedFieldKeys: []string{"timeline.attached_evidence_ids", "timeline.evidence_count", "timeline.has_evidence"},
		},
		attachedDeleteEvidence: {
			rowVersion:       2,
			viewSchemaID:     "cartulary.view.evidence.v1",
			changedFieldKeys: []string{"evidence.linked_record_count"},
		},
	})
	requireAffectedRecordsCanonical(t, attachedDeletePayload, attachedDeleteSrc, attachedDeleteEvidence)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, attachedDeleteLink) != 1 {
		t.Fatalf("attached-evidence delete rollback did not restore link")
	}
	if gotCount, gotHasEvidence := timelineProjectionEvidenceState(t, harness.DB, attachedDeleteSrc); gotCount != 1 || !gotHasEvidence {
		t.Fatalf("attached-evidence delete rollback projection got count=%d has_evidence=%v want count=1 has_evidence=true", gotCount, gotHasEvidence)
	}
	attachedDeleteRollbackChangeSetID := attachedDeletePayload["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2`, attachedDeleteRollbackChangeSetID, actorID) != 1 {
		t.Fatalf("attached-evidence delete rollback did not create attributed rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND row_version = 2`, attachedDeleteRollbackChangeSetID) != 2 {
		t.Fatalf("attached-evidence delete rollback did not append revisions for both records")
	}

	indicatorDelete := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, indicatorID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-01-delete-indicator"}), http.StatusOK)["data"].(map[string]any)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE record_id = $1 AND deleted_at IS NOT NULL AND deleted_by_user_id = $2`, indicatorID, actorID) != 1 {
		t.Fatalf("indicator source tombstone was not set")
	}
	httptestx.RequireSuccessEnvelope(t, restoreRecord(t, harness, login, indicatorID, map[string]any{"base_row_version": int64(indicatorDelete["row_version"].(float64)), "client_txn_id": "txn-i-7-01-restore-indicator"}), http.StatusOK)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM indicators WHERE record_id = $1 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, indicatorID) != 1 {
		t.Fatalf("indicator source tombstone was not cleared")
	}
}

func TestHistoryPaginationRecordBinding_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-i-7-02-pagination")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordA := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-I702A")
	incidentB, recordB := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-I702B")
	base := time.Date(2026, 5, 10, 15, 0, 0, 0, time.UTC)
	firstChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000301")
	secondChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000302")
	thirdChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000303")

	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: firstChangeSet,
		CreatedAt: base, Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "oldest", RowVersion: 2,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: secondChangeSet,
		CreatedAt: base.Add(time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "record", Operation: "middle", RowVersion: 3,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordA, ChangeSetID: thirdChangeSet,
		CreatedAt: base.Add(2 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "identity", Operation: "newest", RowVersion: 4,
	})
	seedHistoryChangeSet(t, harness.DB, historySeed{
		IncidentID: incidentB, ActorID: actorID, RecordID: recordB, ChangeSetID: mustUUID(t, "77777777-0000-4000-8000-000000000304"),
		CreatedAt: base.Add(3 * time.Minute), Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "other-record", RowVersion: 2,
	})

	firstPage := getHistory(t, harness.Server.HTTP.URL, login, recordA, "?limit=1")
	firstItems := historyItems(firstPage)
	if len(firstItems) != 1 || firstItems[0].(map[string]any)["change_set_id"] != thirdChangeSet.String() {
		t.Fatalf("first page did not preserve newest-first order: %#v", firstItems)
	}
	paging := firstPage["meta"].(map[string]any)["paging"].(map[string]any)
	cursor := paging["next_cursor"].(string)
	if paging["limit"] != float64(1) || paging["has_more"] != true || cursor == "" {
		t.Fatalf("unexpected first page cursor metadata: %#v", paging)
	}

	secondPage := getHistory(t, harness.Server.HTTP.URL, login, recordA, "?cursor_token="+cursor)
	secondItems := historyItems(secondPage)
	if len(secondItems) != 1 || secondItems[0].(map[string]any)["change_set_id"] != secondChangeSet.String() {
		t.Fatalf("second page did not preserve order: %#v", secondItems)
	}

	crossRecord := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordB.String()+"/history?cursor_token="+cursor, nil, appsupport.WithCookies(login.SessionCookie))
	errBody := httptestx.RequireErrorEnvelope(t, crossRecord, http.StatusBadRequest, "invalid_pagination_request")
	if errBody["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_cursor_token" {
		t.Fatalf("unexpected cross-record cursor reason: %#v", errBody)
	}

	for _, query := range []string{"?page=1", "?offset=1", "?page_size=1", "?block_size=1", "?limit=0"} {
		resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordA.String()+"/history"+query, nil, appsupport.WithCookies(login.SessionCookie))
		httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
	}
	for _, query := range []string{"?limit=-1", "?limit=abc", "?limit=501"} {
		resp := appsupport.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/records/"+recordA.String()+"/history"+query, nil, appsupport.WithCookies(login.SessionCookie))
		body := httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_pagination_request")
		if body["error"].(map[string]any)["details"].(map[string]any)["reason_code"] != "invalid_limit" {
			t.Fatalf("unexpected invalid limit reason for %s: %#v", query, body)
		}
	}
}

func TestMergeChangeSetRollback_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-i-7-05-merge-rollback")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-history_revision-i-7-05-incident",
		"incident_key":  "IR-P7-I705",
		"title":         "History Merge Rollback",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))

	survivor := uuid.New()
	loser := uuid.New()
	timeline := uuid.New()
	outgoingTarget := uuid.New()
	mentionID := uuid.New()
	linkID := uuid.New()
	outgoingLinkID := uuid.New()
	survivorTag := uuid.New()
	loserTag := uuid.New()
	assessment := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, survivor, "Survivor Host", "survivor-host", "", "")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, loser, "Loser Host", "loser-host", "loser.example.test", "")
	entitytest.SeedEntityAlias(t, harness.DB, incidentID, actorID, loser, "host", "Loser Alias")
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actorID, timeline)
	timelinetest.SeedTimelineRecord(t, harness.DB, incidentID, actorID, outgoingTarget)
	entitytest.SeedResolvedMention(t, harness.DB, actorID, mentionID, timeline, loser, "timeline.host_refs", "host", "loser-host")
	linktest.SeedRecordLink(t, harness.DB, incidentID, actorID, linkID, timeline, loser, "observed_on_host", "manual", nil)
	linktest.SeedRecordLink(t, harness.DB, incidentID, actorID, outgoingLinkID, loser, outgoingTarget, "references_record", "manual", nil)
	linktest.SeedRecordTag(t, harness.DB, incidentID, actorID, survivorTag, survivor, "duplicate-merge-tag")
	linktest.SeedRecordTag(t, harness.DB, incidentID, actorID, loserTag, loser, "duplicate-merge-tag")
	assessmenttest.SeedAssessment(t, harness.DB, incidentID, actorID, assessment, loser, "host", "suspected")
	seedHostProjection(t, harness.DB, incidentID, survivor)
	seedHostProjection(t, harness.DB, incidentID, loser)

	mergeData := httptestx.RequireSuccessEnvelope(t, mergeRecords(t, harness, login, survivor, map[string]any{
		"loser_record_id":           loser.String(),
		"survivor_base_row_version": 1,
		"loser_base_row_version":    1,
		"client_txn_id":             "txn-history_revision-i-7-05-merge-hosts",
		"reason":                    "merge rollback fixture",
	}), http.StatusOK)["data"].(map[string]any)
	mergeChangeSetID := mergeData["change_set_id"].(string)
	if countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id::text = $1
   AND target_kind IN ('host', 'record_link', 'record_tag', 'entity_mention', 'entity_preserved_identifier', 'entity_alias', 'assessment')
`, mergeChangeSetID) < 8 {
		t.Fatalf("merge did not record complete reversible mutation families")
	}
	mergeHistory := historyItemForChangeSetTarget(t, harness, login, survivor, mustUUID(t, mergeChangeSetID), "host", survivor.String())
	requireHistoryActionContains(t, mergeHistory, "change_set")

	requireHostState(t, harness.DB, survivor, "canonical", nil, 2, "loser.example.test")
	requireHostState(t, harness.DB, loser, "merged", &survivor, 2, "loser.example.test")
	if got := stringScalar(t, harness.DB, `SELECT resolved_record_id::text FROM entity_mentions WHERE entity_mention_id = $1`, mentionID); got != survivor.String() {
		t.Fatalf("merge did not repoint mention to survivor, got %s", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND link_type = 'observed_on_host' AND deleted_at IS NULL`, timeline, survivor) != 1 {
		t.Fatalf("merge did not repoint active link to survivor")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE src_record_id = $1 AND dst_record_id = $2 AND link_type = 'references_record' AND deleted_at IS NULL`, survivor, outgoingTarget) != 1 {
		t.Fatalf("merge did not repoint loser-sourced active link to survivor")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, outgoingLinkID) != 0 {
		t.Fatalf("merge did not tombstone original loser-sourced link")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id = $1 AND deleted_at IS NOT NULL`, loserTag) != 1 {
		t.Fatalf("merge did not dedupe loser tag")
	}
	if got := stringScalar(t, harness.DB, `SELECT subject_record_id::text FROM assessments WHERE record_id = $1`, assessment); got != survivor.String() {
		t.Fatalf("merge did not repoint assessment subject, got %s", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_aliases WHERE incident_id = $1 AND record_id = $2 AND normalized_text = 'Loser Alias' AND deleted_at IS NULL`, incidentID, survivor) != 1 {
		t.Fatalf("merge did not carry loser alias to survivor")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_preserved_identifiers WHERE incident_id = $1 AND record_id = $2 AND identifier_type = 'fqdn' AND normalized_value = 'loser.example.test' AND deleted_at IS NULL`, incidentID, survivor) != 1 {
		t.Fatalf("merge did not preserve loser fqdn for survivor")
	}

	httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 19, 0, 0, 0, time.UTC))
	rollbackBody := map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-history_revision-i-7-05-rollback-merge",
		"target":           map[string]any{"kind": "change_set", "change_set_id": mergeChangeSetID},
	}
	rollbackData := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, survivor, rollbackBody), http.StatusOK)["data"].(map[string]any)
	requireAffectedRecords(t, rollbackData, survivor, loser, timeline, outgoingTarget, assessment)
	rollbackChangeSetID := rollbackData["rollback_change_set_id"].(string)
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND client_txn_id = 'txn-history_revision-i-7-05-rollback-merge'`, rollbackChangeSetID) != 1 {
		t.Fatalf("merge rollback did not append rollback change_set")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND operation_kind = 'rollback'`, rollbackChangeSetID) < 8 {
		t.Fatalf("merge rollback did not record inverse mutation families")
	}

	requireHostState(t, harness.DB, survivor, "canonical", nil, 3, "")
	requireHostState(t, harness.DB, loser, "canonical", nil, 3, "loser.example.test")
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM host_grid_projection WHERE record_id IN ($1, $2)`, survivor, loser) != 2 {
		t.Fatalf("merge rollback did not rebuild active host projections")
	}
	if got := stringScalar(t, harness.DB, `SELECT resolved_record_id::text FROM entity_mentions WHERE entity_mention_id = $1`, mentionID); got != loser.String() {
		t.Fatalf("merge rollback did not restore mention target, got %s", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND dst_record_id = $2 AND deleted_at IS NULL`, linkID, loser) != 1 {
		t.Fatalf("merge rollback did not restore original active link target")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND src_record_id = $2 AND dst_record_id = $3 AND deleted_at IS NULL`, outgoingLinkID, loser, outgoingTarget) != 1 {
		t.Fatalf("merge rollback did not restore original loser-sourced active link")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id IN ($1, $2) AND deleted_at IS NULL`, survivorTag, loserTag) != 2 {
		t.Fatalf("merge rollback did not restore both duplicate tags")
	}
	if got := stringScalar(t, harness.DB, `SELECT subject_record_id::text FROM assessments WHERE record_id = $1`, assessment); got != loser.String() {
		t.Fatalf("merge rollback did not restore assessment subject, got %s", got)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_aliases WHERE incident_id = $1 AND record_id = $2 AND normalized_text = 'Loser Alias' AND deleted_at IS NULL`, incidentID, survivor) != 0 {
		t.Fatalf("merge rollback did not tombstone carried survivor alias")
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM entity_preserved_identifiers WHERE incident_id = $1 AND record_id = $2 AND identifier_type = 'fqdn' AND normalized_value = 'loser.example.test' AND deleted_at IS NULL`, incidentID, survivor) != 0 {
		t.Fatalf("merge rollback did not tombstone carried survivor preserved identifier")
	}

	replayData := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, survivor, rollbackBody), http.StatusOK)["data"].(map[string]any)
	if replayData["rollback_change_set_id"] != rollbackChangeSetID || countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-history_revision-i-7-05-rollback-merge'`) != 1 {
		t.Fatalf("merge rollback replay was not idempotent: first=%#v replay=%#v", rollbackData, replayData)
	}

	staleSurvivor := uuid.New()
	staleLoser := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, staleSurvivor, "Stale Survivor", "stale-survivor", "", "")
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, staleLoser, "Stale Loser", "stale-loser", "stale.example.test", "")
	seedHostProjection(t, harness.DB, incidentID, staleSurvivor)
	seedHostProjection(t, harness.DB, incidentID, staleLoser)
	staleMerge := httptestx.RequireSuccessEnvelope(t, mergeRecords(t, harness, login, staleSurvivor, map[string]any{
		"loser_record_id":           staleLoser.String(),
		"survivor_base_row_version": 1,
		"loser_base_row_version":    1,
		"client_txn_id":             "txn-history_revision-i-7-05-stale-merge-hosts",
	}), http.StatusOK)["data"].(map[string]any)
	staleRollback := rollbackRecord(t, harness, login, staleSurvivor, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-history_revision-i-7-05-stale-rollback-merge",
		"target":           map[string]any{"kind": "change_set", "change_set_id": staleMerge["change_set_id"].(string)},
	})
	httptestx.RequireErrorEnvelope(t, staleRollback, http.StatusConflict, "row_version_conflict")
}

func TestStaleRestoreRollbackFailsClosed_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-i-7-03-stale-restore")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-I703")
	seedHostProjection(t, harness.DB, incidentID, recordID)

	deletePayload := httptestx.RequireSuccessEnvelope(t, deleteRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-03-delete"}), http.StatusOK)["data"].(map[string]any)
	if deletePayload["row_version"] != float64(2) {
		t.Fatalf("unexpected delete payload: %#v", deletePayload)
	}
	asserttest.AwaitIncidentStreamIdle(t, asserttest.SQLDatabase(harness.DB), incidentID.String())
	before := StateCounts(t, harness.DB, recordID)
	hubChanges, unsubscribe := harness.Server.Runtime.CollaborationHub.SubscribeIncident(incidentID, 4)
	defer unsubscribe()
	stale := restoreRecord(t, harness, login, recordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-i-7-03-stale-restore"})
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")
	asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	after := StateCounts(t, harness.DB, recordID)
	if before != after {
		t.Fatalf("stale restore mutated state: before=%+v after=%+v", before, after)
	}

	rollbackRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, rollbackRecordID, "Stale Rollback Host", "stale-rollback-host", "", "")
	rollbackTargetChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000703")
	seedRollbackHostPatch(t, harness.DB, incidentID, rollbackRecordID, actorID, rollbackTargetChangeSet, time.Date(2026, 5, 10, 15, 2, 0, 0, time.UTC), "stale rollback before", "stale rollback after")
	rollbackRef := stringField(t, historyItems(getHistory(t, harness.Server.HTTP.URL, login, rollbackRecordID, ""))[0], "history_entry_ref")
	beforeRollback := StateCounts(t, harness.DB, rollbackRecordID)
	staleRollback := rollbackRecord(t, harness, login, rollbackRecordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-i-7-03-stale-rollback",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": rollbackRef},
	})
	httptestx.RequireErrorEnvelope(t, staleRollback, http.StatusConflict, "row_version_conflict")
	asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	afterRollback := StateCounts(t, harness.DB, rollbackRecordID)
	if beforeRollback != afterRollback {
		t.Fatalf("stale rollback mutated state: before=%+v after=%+v", beforeRollback, afterRollback)
	}

	rowRestoreRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, rowRestoreRecordID, "Stale Row Restore Host", "stale-row-restore", "", "")
	seedHostProjection(t, harness.DB, incidentID, rowRestoreRecordID)
	rowRestoreTargetChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000708")
	seedRollbackHostPatch(t, harness.DB, incidentID, rowRestoreRecordID, actorID, rowRestoreTargetChangeSet, time.Date(2026, 5, 10, 15, 2, 30, 0, time.UTC), "stale row before", "stale row snapshot")
	mustExec(t, harness.DB, `UPDATE records SET row_version = 3, updated_by_user_id = $2 WHERE record_id = $1`, rowRestoreRecordID, actorID)
	mustExec(t, harness.DB, `UPDATE hosts SET display_name = 'stale row current', row_version = 3, updated_by_user_id = $2 WHERE record_id = $1`, rowRestoreRecordID, actorID)
	seedHostProjection(t, harness.DB, incidentID, rowRestoreRecordID)
	beforeRowRestore := StateCounts(t, harness.DB, rowRestoreRecordID)
	staleRowRestore := rollbackRecord(t, harness, login, rowRestoreRecordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-i-7-03-stale-row-restore",
		"target":           map[string]any{"kind": "row_restore", "restore_to_revision_no": 2},
	})
	httptestx.RequireErrorEnvelope(t, staleRowRestore, http.StatusConflict, "row_version_conflict")
	asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	afterRowRestore := StateCounts(t, harness.DB, rowRestoreRecordID)
	if beforeRowRestore != afterRowRestore {
		t.Fatalf("stale row restore mutated state: before=%+v after=%+v", beforeRowRestore, afterRowRestore)
	}
	if got := hostDisplayName(t, harness.DB, rowRestoreRecordID); got != "stale row current" {
		t.Fatalf("stale row restore changed host display name to %q", got)
	}

	wholeLeft, wholeRight := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Stale Whole Left", "Stale Whole Right")
	wholeChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000704")
	seedRollbackTwoHostChangeSet(t, harness.DB, incidentID, actorID, wholeChangeSetID, wholeLeft, wholeRight)
	beforeWholeLeft := StateCounts(t, harness.DB, wholeLeft)
	beforeWholeRight := StateCounts(t, harness.DB, wholeRight)
	staleWhole := rollbackRecord(t, harness, login, wholeLeft, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-i-7-03-stale-whole-rollback",
		"target":           map[string]any{"kind": "change_set", "change_set_id": wholeChangeSetID.String()},
	})
	httptestx.RequireErrorEnvelope(t, staleWhole, http.StatusConflict, "row_version_conflict")
	asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	if afterWholeLeft := StateCounts(t, harness.DB, wholeLeft); beforeWholeLeft != afterWholeLeft {
		t.Fatalf("stale whole rollback mutated left state: before=%+v after=%+v", beforeWholeLeft, afterWholeLeft)
	}
	if afterWholeRight := StateCounts(t, harness.DB, wholeRight); beforeWholeRight != afterWholeRight {
		t.Fatalf("stale whole rollback mutated right state: before=%+v after=%+v", beforeWholeRight, afterWholeRight)
	}

	unsupportedRecordID := uuid.New()
	entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, unsupportedRecordID, "Unsupported after", "unsupported-after", "", "")
	seedHostProjection(t, harness.DB, incidentID, unsupportedRecordID)
	unsupportedChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000705")
	seedRollbackHostAndTagChangeSet(t, harness.DB, incidentID, actorID, unsupportedChangeSetID, unsupportedRecordID)
	beforeUnsupported := StateCounts(t, harness.DB, unsupportedRecordID)
	unsupported := rollbackRecord(t, harness, login, unsupportedRecordID, map[string]any{
		"base_row_version": 2,
		"client_txn_id":    "txn-i-7-03-unsupported-whole-rollback",
		"target":           map[string]any{"kind": "change_set", "change_set_id": unsupportedChangeSetID.String()},
	})
	httptestx.RequireErrorEnvelope(t, unsupported, http.StatusConflict, "rollback_precondition_failed")
	asserttest.RequireNoRecordChange(t, hubChanges, 300*time.Millisecond)
	afterUnsupported := StateCounts(t, harness.DB, unsupportedRecordID)
	if beforeUnsupported != afterUnsupported {
		t.Fatalf("unsupported whole rollback mutated state: before=%+v after=%+v", beforeUnsupported, afterUnsupported)
	}
	if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-i-7-03-unsupported-whole-rollback'`) != 0 {
		t.Fatalf("unsupported whole rollback inserted idempotent rollback state")
	}
}

func requireDeleteRestoreRecordChange(t testing.TB, change platformws.RecordChange, recordID uuid.UUID, rowVersion int64, changeKind string, viewSchemaID string) {
	t.Helper()
	if change.RecordID != recordID || change.RowVersion != rowVersion || change.ChangeKind != changeKind || change.ViewSchemaID != viewSchemaID {
		t.Fatalf("unexpected record_changed event: %+v", change)
	}
	if len(change.ChangedFieldKeys) != 0 {
		t.Fatalf("delete/restore changed_field_keys must be present and empty, got %#v", change.ChangedFieldKeys)
	}
	payload := platformws.RecordChangePayload(change)
	affectedViews, ok := payload["affected_views"].([]map[string]any)
	if !ok || len(affectedViews) != 1 {
		t.Fatalf("delete/restore affected_views must be a single view, got %#v", payload["affected_views"])
	}
	if affectedViews[0]["view_schema_id"] != viewSchemaID || affectedViews[0]["change_kind"] != changeKind {
		t.Fatalf("unexpected affected view payload: %#v", affectedViews[0])
	}
}

func requireRollbackRecordChange(t testing.TB, change platformws.RecordChange, recordID uuid.UUID, rowVersion int64, viewSchemaID string) {
	t.Helper()
	if change.RecordID != recordID || change.RowVersion != rowVersion || change.ChangeKind != "invalidate" || change.ViewSchemaID != viewSchemaID {
		t.Fatalf("unexpected rollback record_changed event: %+v", change)
	}
	payload := platformws.RecordChangePayload(change)
	affectedViews, ok := payload["affected_views"].([]map[string]any)
	if !ok || len(affectedViews) != 1 {
		t.Fatalf("rollback affected_views must be a single view, got %#v", payload["affected_views"])
	}
	if affectedViews[0]["view_schema_id"] != viewSchemaID || affectedViews[0]["change_kind"] != "invalidate" {
		t.Fatalf("unexpected rollback affected view payload: %#v", affectedViews[0])
	}
}

func requireRollbackRecordChangesAnyOrder(t testing.TB, changes []platformws.RecordChange, expected map[uuid.UUID]int64, viewSchemaID string) {
	t.Helper()
	seen := map[uuid.UUID]bool{}
	for _, change := range changes {
		wantVersion, ok := expected[change.RecordID]
		if !ok {
			t.Fatalf("unexpected rollback changed record: %+v", change)
		}
		requireRollbackRecordChange(t, change, change.RecordID, wantVersion, viewSchemaID)
		seen[change.RecordID] = true
	}
	if len(seen) != len(expected) {
		t.Fatalf("rollback changed records got %v want %v", seen, expected)
	}
}

type rollbackRecordChangeExpectation struct {
	rowVersion       int64
	viewSchemaID     string
	changedFieldKeys []string
}

func requireRollbackRecordChangesByRecord(t testing.TB, changes []platformws.RecordChange, expected map[uuid.UUID]rollbackRecordChangeExpectation) {
	t.Helper()
	seen := map[uuid.UUID]bool{}
	for _, change := range changes {
		want, ok := expected[change.RecordID]
		if !ok {
			t.Fatalf("unexpected rollback changed record: %+v", change)
		}
		if change.RowVersion != want.rowVersion || change.ChangeKind != "invalidate" || change.ViewSchemaID != want.viewSchemaID {
			t.Fatalf("unexpected rollback record_changed event: got %+v want row_version=%d view_schema_id=%s", change, want.rowVersion, want.viewSchemaID)
		}
		if !slices.Equal(change.ChangedFieldKeys, want.changedFieldKeys) {
			t.Fatalf("rollback changed_field_keys got %v want %v for %s", change.ChangedFieldKeys, want.changedFieldKeys, change.RecordID)
		}
		payload := platformws.RecordChangePayload(change)
		affectedViews, ok := payload["affected_views"].([]map[string]any)
		if !ok || len(affectedViews) != 1 {
			t.Fatalf("rollback affected_views must be a single view, got %#v", payload["affected_views"])
		}
		if affectedViews[0]["view_schema_id"] != want.viewSchemaID || affectedViews[0]["change_kind"] != "invalidate" {
			t.Fatalf("unexpected rollback affected view payload: %#v", affectedViews[0])
		}
		seen[change.RecordID] = true
	}
	if len(seen) != len(expected) {
		t.Fatalf("rollback changed records got %v want %v", seen, expected)
	}
}

func requireHostState(t testing.TB, db *sql.DB, recordID uuid.UUID, wantState string, wantMergedInto *uuid.UUID, wantRowVersion int64, wantFQDN string) {
	t.Helper()
	var (
		state     string
		mergedRaw sql.NullString
		row       int64
		fqdn      sql.NullString
	)
	if err := db.QueryRowContext(context.Background(), `
SELECT host_state, merged_into_record_id::text, row_version, fqdn
  FROM hosts
 WHERE record_id = $1
`, recordID).Scan(&state, &mergedRaw, &row, &fqdn); err != nil {
		t.Fatalf("load host state: %v", err)
	}
	if state != wantState || row != wantRowVersion || fqdn.String != wantFQDN || fqdn.Valid != (wantFQDN != "") {
		t.Fatalf("host %s state got state=%q row=%d fqdn=%q valid=%v", recordID, state, row, fqdn.String, fqdn.Valid)
	}
	if wantMergedInto == nil {
		if mergedRaw.Valid {
			t.Fatalf("host %s merged_into got %s want null", recordID, mergedRaw.String)
		}
		return
	}
	if !mergedRaw.Valid || mergedRaw.String != wantMergedInto.String() {
		t.Fatalf("host %s merged_into got %q valid=%v want %s", recordID, mergedRaw.String, mergedRaw.Valid, wantMergedInto)
	}
}

func requireHistoryActionContains(t testing.TB, item map[string]any, want string) {
	t.Helper()
	raw, ok := item["available_rollback_actions"].([]any)
	if !ok {
		t.Fatalf("history item actions missing or invalid: %#v", item)
	}
	for _, value := range raw {
		if text, ok := value.(string); ok && text == want {
			return
		}
	}
	t.Fatalf("history item actions %#v do not contain %q in item %#v", raw, want, item)
}

func TestRetainedHistoryAcrossRestartAndClosure_Integration(t *testing.T) {
	server, db, env := startReusableServer(t, "history_revision-i-7-04-restart")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, server)
	incidentID, recordID := seedRecord(t, db, server, login, actorID, "IR-P7-I704")
	base := time.Date(2026, 5, 10, 16, 0, 0, 0, time.UTC)
	originalChangeSet := mustUUID(t, "77777777-0000-4000-8000-000000000401")

	seedHistoryChangeSet(t, db, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: originalChangeSet,
		CreatedAt: base, Source: "workbook.records.patch", SequenceNo: 1,
		TargetKind: "host", Operation: "field_update", RowVersion: 2,
	})
	beforeItem := historyItems(getHistory(t, server.HTTP.URL, login, recordID, ""))[0]
	refBefore := stringField(t, beforeItem, "history_entry_ref")
	itemRefBefore := stringField(t, beforeItem, "history_item_ref")
	server.Close()

	restarted := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
	afterRestartItem := historyItems(getHistory(t, restarted.HTTP.URL, login, recordID, ""))[0]
	refAfterRestart := stringField(t, afterRestartItem, "history_entry_ref")
	if refAfterRestart != refBefore {
		t.Fatalf("history_entry_ref changed across restart: before=%q after=%q", refBefore, refAfterRestart)
	}
	if itemRefAfterRestart := stringField(t, afterRestartItem, "history_item_ref"); itemRefAfterRestart != itemRefBefore {
		t.Fatalf("history_item_ref changed across restart: before=%q after=%q", itemRefBefore, itemRefAfterRestart)
	}
	if _, err := db.Exec(`
	UPDATE incidents
	   SET status = 'closed',
	       closed_at = $1
	 WHERE id = $2
	`, base.Add(time.Minute), incidentID); err != nil {
		t.Fatalf("seed incident closure: %v", err)
	}
	seedHistoryChangeSet(t, db, historySeed{
		IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: mustUUID(t, "77777777-0000-4000-8000-000000000402"),
		CreatedAt: base.Add(2 * time.Minute), Source: "rollback", SequenceNo: 1,
		TargetKind: "host", Operation: "rollback", RowVersion: 3,
	})
	items := collectHistoryPages(t, restarted.HTTP.URL, login, recordID, 1)
	if len(items) != 2 {
		t.Fatalf("expected full retained history after restart and later change: %#v", items)
	}
	if stringField(t, items[1], "history_entry_ref") != refBefore {
		t.Fatalf("older selector changed after closure and later change: %#v", items[1])
	}
	if stringField(t, items[1], "history_item_ref") != itemRefBefore {
		t.Fatalf("older display selector changed after closure and later change: %#v", items[1])
	}
}

func startReusableServer(t testing.TB, prefix string) (*httptestx.Server, *sql.DB, map[string]string) {
	t.Helper()
	postgresHarness := pgtest.Start(t)
	testDB := postgresHarness.PrepareIsolatedDatabaseT(t, prefix)
	s3Harness := s3test.Start(t)
	bucket := s3Harness.PreparePackageBucketT(t, prefix)

	env := testDB.Env()
	for key, value := range s3Harness.Env(bucket) {
		env[key] = value
	}
	env["CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH"] = fixtures.Path("bootstrap-admin", "canonical.json")

	server := httptestx.StartServer(t, httptestx.ServerOptions{Env: env, TestRouteMode: httptestx.TestRouteModeDisabled})
	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open reusable sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return server, db, env
}
