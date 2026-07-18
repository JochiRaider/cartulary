package revisions_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	workbookscenariotest "github.com/JochiRaider/cartulary/internal/modules/workbook/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestDestructiveOperationLocks_Unit(t *testing.T) {
	harness := workbookscenariotest.StartServer(t, "phase7-u-7-06-rollback-lock")
	login, actorID := workbookscenariotest.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U706")
	changeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000601")
	seedRollbackHostPatch(t, harness.DB, incidentID, recordID, actorID, changeSetID, time.Date(2026, 5, 10, 18, 0, 0, 0, time.UTC), "lock before", "lock after")
	historyEntryRef := stringField(t, historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))[0], "history_entry_ref")
	if historyEntryRef == "" {
		t.Fatalf("seeded rollback target did not expose history_entry_ref")
	}
	setMembershipRole(t, harness.DB, incidentID, actorID, "reviewer")

	lockTx, err := harness.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin rollback lock holder: %v", err)
	}
	if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, recordID); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("hold rollback lock: %v", err)
	}

	locked := rollbackRecord(t, harness, login, recordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-u-7-06-locked-stale",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef},
	})
	lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
	if lockedErr["error"].(map[string]any)["retryable"] != true {
		t.Fatalf("record_locked must be retryable: %#v", lockedErr)
	}
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("release rollback lock: %v", err)
	}

	stale := rollbackRecord(t, harness, login, recordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    "txn-u-7-06-locked-stale",
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef},
	})
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")

	t.Run("row restore record lock wins before stale row version", func(t *testing.T) {
		requireRollbackTargetLockPrecedence(t, harness, login, recordID, recordID, map[string]any{"kind": "row_restore", "restore_to_revision_no": 2}, "txn-u-7-06-row-restore-lock", "row_version_conflict", 1)
	})

	t.Run("row restore record lock wins before selector visibility", func(t *testing.T) {
		requireRollbackTargetLockPrecedence(t, harness, login, recordID, recordID, map[string]any{"kind": "row_restore", "restore_to_revision_no": 999}, "txn-u-7-06-row-restore-target-lock", "rollback_target_not_found", 2)
	})

	t.Run("link protected set lock wins before stale row version", func(t *testing.T) {
		src, dst, linkID := seedRollbackRecordLinkCreate(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000602"))
		mustExec(t, harness.DB, `UPDATE records SET row_version = 2 WHERE record_id = $1`, src)
		historyEntryRef := historyEntryRefForTarget(t, harness, login, src, "record_link", linkID.String())
		requireRollbackLockPrecedence(t, harness, login, dst, src, historyEntryRef, "txn-u-7-06-link-lock")
	})

	t.Run("mention source lock wins before stale row version", func(t *testing.T) {
		source, _, mentionID := seedRollbackMentionPatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000603"))
		historyEntryRef := historyEntryRefForTarget(t, harness, login, source, "entity_mention", mentionID.String())
		requireRollbackLockPrecedence(t, harness, login, source, source, historyEntryRef, "txn-u-7-06-mention-lock")
	})

	t.Run("whole change set protected record lock wins before stale row version", func(t *testing.T) {
		left, right := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Lock Whole Left", "Lock Whole Right")
		changeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000605")
		seedRollbackTwoHostChangeSet(t, harness.DB, incidentID, actorID, changeSetID, left, right)
		requireRollbackTargetLockPrecedence(t, harness, login, right, left, map[string]any{"kind": "change_set", "change_set_id": changeSetID.String()}, "txn-u-7-06-whole-lock", "row_version_conflict", 1)
	})

	t.Run("target precondition waits until lock release", func(t *testing.T) {
		tagRecordID := seedRollbackRecordTagMutation(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000604"), "href-tag-lock-rollback")
		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin rollback lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, tagRecordID); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold rollback lock: %v", err)
		}
		locked := rollbackRecord(t, harness, login, tagRecordID, map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-u-7-06-tag-lock",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": "href-tag-lock-rollback"},
		})
		httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release rollback lock: %v", err)
		}
		released := rollbackRecord(t, harness, login, tagRecordID, map[string]any{
			"base_row_version": 1,
			"client_txn_id":    "txn-u-7-06-tag-lock",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": "href-tag-lock-rollback"},
		})
		requireRollbackReasonCode(t, released, "target_not_reversible")
	})

	t.Run("whole change set target precondition waits until lock release", func(t *testing.T) {
		tagRecordID := uuid.New()
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, actorID, tagRecordID, "Whole Tag Host", "whole-tag-host", "", "")
		changeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000606")
		seedRollbackHostAndTagChangeSet(t, harness.DB, incidentID, actorID, changeSetID, tagRecordID)
		requireRollbackTargetLockPrecedence(t, harness, login, tagRecordID, tagRecordID, map[string]any{"kind": "change_set", "change_set_id": changeSetID.String()}, "txn-u-7-06-whole-tag-lock", "rollback_precondition_failed", 2)
	})

	t.Run("merge record lock wins before stale row version", func(t *testing.T) {
		survivor, loser := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Merge Lock Survivor", "Merge Lock Loser")
		mustExec(t, harness.DB, `UPDATE records SET row_version = 2 WHERE record_id = $1`, survivor)
		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin merge lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, survivor); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold merge lock: %v", err)
		}
		body := map[string]any{
			"loser_record_id":           loser.String(),
			"survivor_base_row_version": 1,
			"loser_base_row_version":    1,
			"client_txn_id":             "txn-u-7-06-merge-lock",
		}
		locked := mergeRecords(t, harness, login, survivor, body)
		lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
		if lockedErr["error"].(map[string]any)["retryable"] != true {
			t.Fatalf("merge record_locked must be retryable: %#v", lockedErr)
		}
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release merge lock: %v", err)
		}
		released := mergeRecords(t, harness, login, survivor, body)
		httptestx.RequireErrorEnvelope(t, released, http.StatusConflict, "row_version_conflict")
	})

	t.Run("locked soft deleted rollback returns lock before restore guidance", func(t *testing.T) {
		deletedRecordID := uuid.New()
		workbookscenariotest.SeedHostRecord(t, harness.DB, incidentID, actorID, deletedRecordID, "Deleted Lock Host", "deleted-lock-host", "", "")
		deletedChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000607")
		seedRollbackHostPatch(t, harness.DB, incidentID, deletedRecordID, actorID, deletedChangeSetID, time.Date(2026, 5, 10, 18, 6, 0, 0, time.UTC), "deleted lock before", "deleted lock after")
		deletedRef := historyEntryRefForTarget(t, harness, login, deletedRecordID, "host", deletedRecordID.String())
		mustExec(t, harness.DB, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedRecordID, time.Now().UTC(), actorID)

		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin deleted rollback lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, deletedRecordID); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold deleted rollback lock: %v", err)
		}
		body := map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-u-7-06-deleted-lock",
			"target":           map[string]any{"kind": "history_entry", "history_entry_ref": deletedRef},
		}
		locked := rollbackRecord(t, harness, login, deletedRecordID, body)
		lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
		if lockedErr["error"].(map[string]any)["retryable"] != true {
			t.Fatalf("deleted rollback record_locked must be retryable: %#v", lockedErr)
		}
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release deleted rollback lock: %v", err)
		}
		released := rollbackRecord(t, harness, login, deletedRecordID, body)
		httptestx.RequireErrorEnvelope(t, released, http.StatusConflict, "record_deleted_use_restore")
	})

	t.Run("invalid rollback selector waits until lock release", func(t *testing.T) {
		requireRollbackTargetLockPrecedence(t, harness, login, recordID, recordID, map[string]any{"kind": "history_entry", "history_entry_ref": "missing-history-entry-ref"}, "txn-u-7-06-invalid-ref-lock", "rollback_target_not_found", 2)
	})

	t.Run("locked merge same record waits until lock release", func(t *testing.T) {
		survivor, _ := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Merge Same Survivor", "Merge Same Other")
		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin merge same-record lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, survivor); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold merge same-record lock: %v", err)
		}
		body := map[string]any{
			"loser_record_id":           survivor.String(),
			"survivor_base_row_version": 1,
			"loser_base_row_version":    1,
			"client_txn_id":             "txn-u-7-06-merge-same-lock",
		}
		locked := mergeRecords(t, harness, login, survivor, body)
		lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
		if lockedErr["error"].(map[string]any)["retryable"] != true {
			t.Fatalf("same-record merge lock must be retryable: %#v", lockedErr)
		}
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release merge same-record lock: %v", err)
		}
		requireMergeReasonCode(t, mergeRecords(t, harness, login, survivor, body), "same_record")
	})

	t.Run("locked merge type mismatch waits until lock release", func(t *testing.T) {
		survivor, _ := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Merge Mismatch Survivor", "Merge Mismatch Other")
		loser := uuid.New()
		workbookscenariotest.SeedIdentityRecord(t, harness.DB, incidentID, actorID, loser, "Mismatch Identity", "mismatch-lock@example.test", "mismatch-lock@example.test", "MISMATCHLOCK")
		lockTx, err := harness.DB.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin merge mismatch lock holder: %v", err)
		}
		if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, loser); err != nil {
			_ = lockTx.Rollback()
			t.Fatalf("hold merge mismatch lock: %v", err)
		}
		body := map[string]any{
			"loser_record_id":           loser.String(),
			"survivor_base_row_version": 1,
			"loser_base_row_version":    1,
			"client_txn_id":             "txn-u-7-06-merge-mismatch-lock",
		}
		locked := mergeRecords(t, harness, login, survivor, body)
		lockedErr := httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
		if lockedErr["error"].(map[string]any)["retryable"] != true {
			t.Fatalf("mismatch merge lock must be retryable: %#v", lockedErr)
		}
		if err := lockTx.Rollback(); err != nil {
			t.Fatalf("release merge mismatch lock: %v", err)
		}
		requireMergeReasonCode(t, mergeRecords(t, harness, login, survivor, body), "record_type_mismatch")
	})

	_ = uuid.Nil
}

func mergeRecords(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, survivorRecordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return workbookscenariotest.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+survivorRecordID.String()+"/merge", body, workbookscenariotest.WithCookies(login.SessionCookie, login.CSRFCookie), workbookscenariotest.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func requireMergeReasonCode(t testing.TB, resp *http.Response, want string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "merge_precondition_failed")
	errorObject := body["error"].(map[string]any)
	details := errorObject["details"].(map[string]any)
	if details["reason_code"] != want {
		t.Fatalf("merge reason_code = %#v want %q", details["reason_code"], want)
	}
}

func requireRollbackLockPrecedence(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, lockedRecordID uuid.UUID, addressedRecordID uuid.UUID, historyEntryRef string, clientTxnID string) {
	t.Helper()
	lockTx, err := harness.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin rollback lock holder: %v", err)
	}
	if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, lockedRecordID); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("hold rollback lock: %v", err)
	}
	locked := rollbackRecord(t, harness, login, addressedRecordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    clientTxnID,
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef},
	})
	httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("release rollback lock: %v", err)
	}
	stale := rollbackRecord(t, harness, login, addressedRecordID, map[string]any{
		"base_row_version": 1,
		"client_txn_id":    clientTxnID,
		"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef},
	})
	httptestx.RequireErrorEnvelope(t, stale, http.StatusConflict, "row_version_conflict")
}

func requireRollbackTargetLockPrecedence(t testing.TB, harness *workbookscenariotest.ServerHarness, login workbookscenariotest.LoginResult, lockedRecordID uuid.UUID, addressedRecordID uuid.UUID, target map[string]any, clientTxnID string, releasedCode string, baseRowVersion int64) {
	t.Helper()
	lockTx, err := harness.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin rollback lock holder: %v", err)
	}
	if _, err := lockTx.ExecContext(context.Background(), `SELECT record_id FROM records WHERE record_id = $1 FOR UPDATE`, lockedRecordID); err != nil {
		_ = lockTx.Rollback()
		t.Fatalf("hold rollback lock: %v", err)
	}
	locked := rollbackRecord(t, harness, login, addressedRecordID, map[string]any{
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"target":           target,
	})
	httptestx.RequireErrorEnvelope(t, locked, http.StatusConflict, "record_locked")
	if err := lockTx.Rollback(); err != nil {
		t.Fatalf("release rollback lock: %v", err)
	}
	released := rollbackRecord(t, harness, login, addressedRecordID, map[string]any{
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"target":           target,
	})
	if releasedCode == "rollback_target_not_found" {
		httptestx.RequireErrorEnvelope(t, released, http.StatusNotFound, releasedCode)
		return
	}
	if releasedCode == "rollback_precondition_failed" {
		requireRollbackReasonCode(t, released, "target_not_reversible")
		return
	}
	httptestx.RequireErrorEnvelope(t, released, http.StatusConflict, releasedCode)
}
