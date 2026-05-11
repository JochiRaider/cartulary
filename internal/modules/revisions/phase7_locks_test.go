package revisions_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestPhase7_DestructiveOperationLocks_U_7_06(t *testing.T) {
	harness := phase4test.StartServer(t, "phase7-u-7-06-rollback-lock")
	login, actorID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedPhase7Record(t, harness.DB, harness.Server, login, actorID, "IR-P7-U706")
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
		phase4test.SeedHostRecord(t, harness.DB, incidentID, actorID, tagRecordID, "Whole Tag Host", "whole-tag-host", "", "")
		changeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000606")
		seedRollbackHostAndTagChangeSet(t, harness.DB, incidentID, actorID, changeSetID, tagRecordID)
		requireRollbackTargetLockPrecedence(t, harness, login, tagRecordID, tagRecordID, map[string]any{"kind": "change_set", "change_set_id": changeSetID.String()}, "txn-u-7-06-whole-tag-lock", "rollback_precondition_failed", 2)
	})

	_ = uuid.Nil
}

func requireRollbackLockPrecedence(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, lockedRecordID uuid.UUID, addressedRecordID uuid.UUID, historyEntryRef string, clientTxnID string) {
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

func requireRollbackTargetLockPrecedence(t testing.TB, harness *phase4test.ServerHarness, login phase4test.LoginResult, lockedRecordID uuid.UUID, addressedRecordID uuid.UUID, target map[string]any, clientTxnID string, releasedCode string, baseRowVersion int64) {
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
	if releasedCode == "rollback_precondition_failed" {
		requireRollbackReasonCode(t, released, "target_not_reversible")
		return
	}
	httptestx.RequireErrorEnvelope(t, released, http.StatusConflict, releasedCode)
}
