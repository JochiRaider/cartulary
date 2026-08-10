package revisions_test

import (
	"context"
	"database/sql"
	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	envelopetest "github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/app/revisionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestRollbackSelectorUnion_Unit(t *testing.T) {
	harness := appsupport.StartServer(t, "history_revision-u-7-05-rollback")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U705")
	changeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000501")
	seedRollbackHostPatch(t, harness.DB, incidentID, recordID, actorID, changeSetID, time.Date(2026, 5, 10, 17, 0, 0, 0, time.UTC), "before rollback", "after rollback")
	history := historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, ""))
	historyEntryRef := stringField(t, history[0], "history_entry_ref")
	if historyEntryRef == "" {
		t.Fatalf("seeded rollback target did not expose history_entry_ref: %#v", history[0])
	}
	requireHistoryActions(t, history[0].(map[string]any), "history_entry", "change_set", "row_restore")

	t.Run("openapi route and schema", func(t *testing.T) {
		document := contracttest.OpenAPIDocument(t)
		operation := historyOpenAPIObjectAt(t, document, "paths", "/api/v1/records/{record_id}/rollback", "post")
		if operation["operationId"] != "rollbackRecord" {
			t.Fatalf("rollback operationId = %#v", operation["operationId"])
		}
		schemas := historyOpenAPIObjectAt(t, document, "components", "schemas")
		request := historyOpenAPIObjectAt(t, schemas, "RecordRollbackRequest")
		required := historyOpenAPIStringArrayAt(t, request, "required")
		for _, field := range []string{"base_row_version", "client_txn_id", "target"} {
			if !slices.Contains(required, field) {
				t.Fatalf("RecordRollbackRequest missing required field %q: %v", field, required)
			}
		}
		_ = historyOpenAPIObjectAt(t, schemas, "RecordRollbackEnvelope")
		_ = historyOpenAPIObjectAt(t, schemas, "RecordRollbackData")
		errorsDocument := contracttest.ErrorRegistryDocument(t)
		requireErrorReasonRegistry(t, errorsDocument, "invalid_rollback_request", "request_not_object", "missing_required_field", "unknown_field", "invalid_base_row_version", "invalid_value", "target_not_object", "unsupported_target_kind")
		requireErrorReasonRegistry(t, errorsDocument, "rollback_precondition_failed", "target_not_reversible", "entry_requires_change_set", "dependent_later_changes", "stale_target")
	})

	t.Run("strict validation", func(t *testing.T) {
		cases := []struct {
			name string
			body map[string]any
		}{
			{name: "unknown top-level", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-unknown-top", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}, "display_text": "not a selector"}},
			{name: "unknown target member", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-unknown-target", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef, "label": "bad"}}},
			{name: "missing selector", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-missing-selector", "target": map[string]any{"kind": "history_entry"}}},
			{name: "wrong selector type", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-wrong-selector", "target": map[string]any{"kind": "history_entry", "history_entry_ref": 42}}},
			{name: "unsupported kind", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-unsupported", "target": map[string]any{"kind": "display_text", "history_entry_ref": historyEntryRef}}},
			{name: "change set selector wrong type", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-change-set-type", "target": map[string]any{"kind": "change_set", "change_set_id": 7}}},
			{name: "row restore selector wrong type", body: map[string]any{"base_row_version": 2, "client_txn_id": "txn-row-restore-type", "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": "2"}}},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := rollbackRecord(t, harness, login, recordID, tc.body)
				httptestx.RequireErrorEnvelope(t, resp, http.StatusBadRequest, "invalid_rollback_request")
			})
		}
	})

	t.Run("durable selector only", func(t *testing.T) {
		for _, selector := range []string{changeSetID.String(), "change_set_mutation:" + changeSetID.String() + ":1", "host.hostname", "cartulary.view.hosts.v1", "after rollback"} {
			resp := rollbackRecord(t, harness, login, recordID, map[string]any{
				"base_row_version": 2,
				"client_txn_id":    "txn-selector-" + selector,
				"target":           map[string]any{"kind": "history_entry", "history_entry_ref": selector},
			})
			httptestx.RequireErrorEnvelope(t, resp, http.StatusNotFound, "rollback_target_not_found")
		}
	})

	t.Run("authorization and target shape admission", func(t *testing.T) {
		setMembershipRole(t, harness.DB, incidentID, actorID, "editor")
		forbidden := rollbackRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-forbidden", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}})
		httptestx.RequireErrorEnvelope(t, forbidden, http.StatusForbidden, "authorization_denied")
		hiddenUser := authflowtest.SeedLocalUserRecord(t, harness.DB, "history_revision-hidden@example.test", "History Hidden", "HiddenPass1!", false, false, true)
		hiddenLogin := loginLocalUser(t, harness, hiddenUser.Email, "HiddenPass1!")
		hiddenExisting := rollbackRecord(t, harness, hiddenLogin, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-hidden-existing", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}})
		httptestx.RequireErrorEnvelope(t, hiddenExisting, http.StatusNotFound, "incident_not_found")
		missing := rollbackRecord(t, harness, login, uuid.New(), map[string]any{"base_row_version": 2, "client_txn_id": "txn-hidden", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}})
		httptestx.RequireErrorEnvelope(t, missing, http.StatusNotFound, "incident_not_found")
		setMembershipRole(t, harness.DB, incidentID, actorID, "reviewer")
		changeSet := rollbackRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-change-set-not-visible", "target": map[string]any{"kind": "change_set", "change_set_id": uuid.New().String()}})
		httptestx.RequireErrorEnvelope(t, changeSet, http.StatusNotFound, "rollback_target_not_found")
		rowRestoreMissing := rollbackRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-row-restore-not-visible", "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 99}})
		httptestx.RequireErrorEnvelope(t, rowRestoreMissing, http.StatusNotFound, "rollback_target_not_found")
	})

	t.Run("change set selectors must be visible through addressed record history", func(t *testing.T) {
		unrelatedRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, unrelatedRecordID, "Unrelated Rollback Host", "unrelated-rollback-host", "", "")
		unrelatedChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000536")
		seedRollbackHostPatch(t, harness.DB, incidentID, unrelatedRecordID, actorID, unrelatedChangeSetID, time.Date(2026, 5, 10, 17, 9, 0, 0, time.UTC), "unrelated before", "unrelated after")
		unrelated := rollbackRecord(t, harness, login, recordID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-change-set-unrelated-record",
			"target":           map[string]any{"kind": "change_set", "change_set_id": unrelatedChangeSetID.String()},
		})
		httptestx.RequireErrorEnvelope(t, unrelated, http.StatusNotFound, "rollback_target_not_found")

		foreignIncidentID, foreignRecordID := seedRecord(t, harness.DB, harness.Server, login, actorID, "IR-P7-U705-FOREIGN")
		foreignChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000537")
		seedRollbackHostPatch(t, harness.DB, foreignIncidentID, foreignRecordID, actorID, foreignChangeSetID, time.Date(2026, 5, 10, 17, 10, 0, 0, time.UTC), "foreign before", "foreign after")
		foreign := rollbackRecord(t, harness, login, recordID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-change-set-foreign-incident",
			"target":           map[string]any{"kind": "change_set", "change_set_id": foreignChangeSetID.String()},
		})
		httptestx.RequireErrorEnvelope(t, foreign, http.StatusNotFound, "rollback_target_not_found")

		displayDerived := rollbackRecord(t, harness, login, recordID, map[string]any{
			"base_row_version": 2,
			"client_txn_id":    "txn-change-set-display-derived",
			"target":           map[string]any{"kind": "change_set", "change_set_id": "after rollback"},
		})
		httptestx.RequireErrorEnvelope(t, displayDerived, http.StatusBadRequest, "invalid_rollback_request")
	})

	t.Run("soft deleted record fails through restore guidance", func(t *testing.T) {
		deletedRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, deletedRecordID, "Deleted Rollback Host", "deleted-rollback-host", "", "")
		deletedChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000528")
		seedRollbackHostPatch(t, harness.DB, incidentID, deletedRecordID, actorID, deletedChangeSetID, time.Date(2026, 5, 10, 17, 2, 0, 0, time.UTC), "deleted before", "deleted after")
		deletedRef := historyEntryRefForTarget(t, harness, login, deletedRecordID, "host", deletedRecordID.String())
		mustExec(t, harness.DB, `UPDATE records SET deleted_at = $2, deleted_by_user_id = $3 WHERE record_id = $1`, deletedRecordID, time.Now().UTC(), actorID)
		resp := rollbackRecord(t, harness, login, deletedRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-deleted-rollback", "target": map[string]any{"kind": "history_entry", "history_entry_ref": deletedRef}})
		httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "record_deleted_use_restore")
	})

	t.Run("successful single entry rollback and idempotency", func(t *testing.T) {
		httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 17, 1, 0, 0, time.UTC))
		body := map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-05-rollback", "reason": "  rollback reason\n", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}}
		success := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, recordID, body), http.StatusOK)["data"].(map[string]any)
		if success["record_id"] != recordID.String() || success["incident_id"] != incidentID.String() || success["row_version"] != float64(3) {
			t.Fatalf("unexpected rollback payload: %#v", success)
		}
		if success["target_change_set_id"] != changeSetID.String() || success["rollback_change_set_id"] == "" {
			t.Fatalf("rollback payload missing change-set ids: %#v", success)
		}
		target := success["target"].(map[string]any)
		if target["kind"] != "history_entry" || target["history_entry_ref"] != historyEntryRef {
			t.Fatalf("rollback did not echo normalized target: %#v", target)
		}
		affected := success["affected_record_ids"].([]any)
		if len(affected) != 1 || affected[0] != recordID.String() {
			t.Fatalf("unexpected affected records: %#v", affected)
		}
		if got := hostDisplayName(t, harness.DB, recordID); got != "before rollback" {
			t.Fatalf("rollback did not restore host source value, got %q", got)
		}
		if got := hostProjectionDisplayName(t, harness.DB, recordID); got != "before rollback" {
			t.Fatalf("rollback did not rebuild projection, got %q", got)
		}
		rollbackChangeSetID := success["rollback_change_set_id"].(string)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND actor_user_id = $2 AND reason = 'rollback reason'`, rollbackChangeSetID, actorID) != 1 {
			t.Fatalf("rollback change_set was not attributed with source rollback")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'host' AND target_id = $2 AND operation_kind = 'rollback'`, rollbackChangeSetID, recordID.String()) != 1 {
			t.Fatalf("rollback inverse mutation missing")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 3`, rollbackChangeSetID, recordID) != 1 {
			t.Fatalf("rollback record revision missing")
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, rollbackChangeSetID)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id = $1 AND operation_kind = 'field_update'`, changeSetID) != 1 {
			t.Fatalf("prior mutation history was rewritten")
		}
		replay := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, recordID, body), http.StatusOK)["data"].(map[string]any)
		if replay["rollback_change_set_id"] != rollbackChangeSetID || countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-u-7-05-rollback'`) != 1 {
			t.Fatalf("idempotent rollback replay changed payload or created a second change_set: replay=%#v", replay)
		}
		revisionsupport.RequireOneRecordChangeIntentPerRevisionSQL(t, harness.DB, rollbackChangeSetID)
		divergent := rollbackRecord(t, harness, login, recordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-05-rollback", "reason": "different", "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef}})
		httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("successful whole change set rollback and idempotency", func(t *testing.T) {
		left, right := seedRollbackHostPair(t, harness.DB, incidentID, actorID, "Change Set Left", "Change Set Right")
		wholeChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000534")
		seedRollbackTwoHostChangeSet(t, harness.DB, incidentID, actorID, wholeChangeSetID, left, right)
		history := historyItems(getHistory(t, harness.Server.HTTP.URL, login, left, ""))
		requireHistoryActions(t, history[0].(map[string]any), "history_entry", "change_set", "row_restore")

		httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 17, 7, 0, 0, time.UTC))
		body := map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-05-whole-change-set", "target": map[string]any{"kind": "change_set", "change_set_id": wholeChangeSetID.String()}}
		success := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, left, body), http.StatusOK)["data"].(map[string]any)
		if success["row_version"] != float64(3) || success["target_change_set_id"] != wholeChangeSetID.String() {
			t.Fatalf("unexpected whole change-set payload: %#v", success)
		}
		requireAffectedRecords(t, success, left, right)
		if got := hostDisplayName(t, harness.DB, left); got != "left before" {
			t.Fatalf("left source rollback got %q", got)
		}
		if got := hostDisplayName(t, harness.DB, right); got != "right before" {
			t.Fatalf("right source rollback got %q", got)
		}
		if got := hostProjectionDisplayName(t, harness.DB, left); got != "left before" {
			t.Fatalf("left projection rollback got %q", got)
		}
		if got := hostProjectionDisplayName(t, harness.DB, right); got != "right before" {
			t.Fatalf("right projection rollback got %q", got)
		}
		rollbackChangeSetID := success["rollback_change_set_id"].(string)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND client_txn_id = 'txn-u-7-05-whole-change-set'`, rollbackChangeSetID) != 1 {
			t.Fatalf("whole change-set rollback did not create exactly one rollback change_set")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND operation_kind = 'rollback'`, rollbackChangeSetID) != 2 {
			t.Fatalf("whole change-set rollback did not append two inverse mutations")
		}
		if first := stringScalar(t, harness.DB, `SELECT target_id FROM change_set_mutations WHERE change_set_id::text = $1 AND sequence_no = 1`, rollbackChangeSetID); first != right.String() {
			t.Fatalf("rollback sequence 1 = %s want right record %s", first, right)
		}
		if second := stringScalar(t, harness.DB, `SELECT target_id FROM change_set_mutations WHERE change_set_id::text = $1 AND sequence_no = 2`, rollbackChangeSetID); second != left.String() {
			t.Fatalf("rollback sequence 2 = %s want left record %s", second, left)
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND row_version = 3`, rollbackChangeSetID) != 2 {
			t.Fatalf("whole change-set rollback did not append one revision per affected record")
		}
		replay := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, left, body), http.StatusOK)["data"].(map[string]any)
		if replay["rollback_change_set_id"] != rollbackChangeSetID || countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-u-7-05-whole-change-set'`) != 1 {
			t.Fatalf("whole change-set replay changed payload or created another rollback: %#v", replay)
		}
		divergent := rollbackRecord(t, harness, login, left, map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-05-whole-change-set", "reason": "different", "target": map[string]any{"kind": "change_set", "change_set_id": wholeChangeSetID.String()}})
		httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("successful whole row restore and idempotency", func(t *testing.T) {
		rowRestoreRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, rowRestoreRecordID, "Row restore seed", "row-restore-host", "", "")
		seedHostProjection(t, harness.DB, incidentID, rowRestoreRecordID)
		rowRestoreChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000541")
		seedRollbackHostPatch(t, harness.DB, incidentID, rowRestoreRecordID, actorID, rowRestoreChangeSetID, time.Date(2026, 5, 10, 17, 11, 0, 0, time.UTC), "row restore before", "row restore snapshot")
		mustExec(t, harness.DB, `
UPDATE records
   SET row_version = 3,
       updated_at = $3,
       updated_by_user_id = $2
 WHERE record_id = $1
`, rowRestoreRecordID, actorID, time.Date(2026, 5, 10, 17, 12, 0, 0, time.UTC))
		mustExec(t, harness.DB, `
UPDATE hosts
   SET display_name = 'row restore current',
       row_version = 3,
       updated_at = $3,
       updated_by_user_id = $2
 WHERE record_id = $1
`, rowRestoreRecordID, actorID, time.Date(2026, 5, 10, 17, 12, 0, 0, time.UTC))
		seedHostProjection(t, harness.DB, incidentID, rowRestoreRecordID)
		linkID := seedRecordLinkForRowRestore(t, harness.DB, incidentID, rowRestoreRecordID, actorID)
		tagID := seedRecordTag(t, harness.DB, incidentID, rowRestoreRecordID, actorID)
		beforeLinks := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, linkID)
		beforeTags := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id = $1 AND deleted_at IS NULL`, tagID)

		history := historyItems(getHistory(t, harness.Server.HTTP.URL, login, rowRestoreRecordID, ""))
		requireHistoryActions(t, history[0].(map[string]any), "history_entry", "change_set", "row_restore")
		if history[0].(map[string]any)["revision_no"] != float64(2) {
			t.Fatalf("row restore history item did not expose revision_no=2: %#v", history[0])
		}
		beforeRefs := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, rowRestoreRecordID)

		httptestx.SetClockFixed(t, harness.Server, time.Date(2026, 5, 10, 17, 13, 0, 0, time.UTC))
		body := map[string]any{"base_row_version": 3, "client_txn_id": "txn-u-7-05-row-restore", "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 2}}
		success := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, rowRestoreRecordID, body), http.StatusOK)["data"].(map[string]any)
		if success["row_version"] != float64(4) || success["target_change_set_id"] != rowRestoreChangeSetID.String() {
			t.Fatalf("unexpected row restore payload: %#v", success)
		}
		requireAffectedRecords(t, success, rowRestoreRecordID)
		if got := hostDisplayName(t, harness.DB, rowRestoreRecordID); got != "row restore snapshot" {
			t.Fatalf("row restore did not restore selected row-backed snapshot, got %q", got)
		}
		if got := hostProjectionDisplayName(t, harness.DB, rowRestoreRecordID); got != "row restore snapshot" {
			t.Fatalf("row restore did not rebuild projection, got %q", got)
		}
		if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, linkID); got != beforeLinks {
			t.Fatalf("row restore mutated record link active state: got %d want %d", got, beforeLinks)
		}
		if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_tags WHERE record_tag_id = $1 AND deleted_at IS NULL`, tagID); got != beforeTags {
			t.Fatalf("row restore mutated record tag active state: got %d want %d", got, beforeTags)
		}
		rollbackChangeSetID := success["rollback_change_set_id"].(string)
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE change_set_id::text = $1 AND source = 'rollback' AND client_txn_id = 'txn-u-7-05-row-restore'`, rollbackChangeSetID) != 1 {
			t.Fatalf("row restore did not create attributed rollback change_set")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_set_mutations WHERE change_set_id::text = $1 AND target_kind = 'host' AND target_id = $2 AND operation_kind = 'row_restore'`, rollbackChangeSetID, rowRestoreRecordID.String()) != 1 {
			t.Fatalf("row restore mutation missing")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id::text = $1 AND record_id = $2 AND row_version = 4`, rollbackChangeSetID, rowRestoreRecordID) != 1 {
			t.Fatalf("row restore did not append a new record revision")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_revisions WHERE change_set_id = $1 AND record_id = $2 AND row_version = 2`, rowRestoreChangeSetID, rowRestoreRecordID) != 1 {
			t.Fatalf("row restore rewrote selected source revision")
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM record_history_entry_refs WHERE record_id = $1`, rowRestoreRecordID) != beforeRefs {
			t.Fatalf("row restore mutated prior history_entry_ref rows")
		}
		replay := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, rowRestoreRecordID, body), http.StatusOK)["data"].(map[string]any)
		if replay["rollback_change_set_id"] != rollbackChangeSetID || countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-u-7-05-row-restore'`) != 1 {
			t.Fatalf("row restore replay changed payload or created another rollback: %#v", replay)
		}
		divergent := rollbackRecord(t, harness, login, rowRestoreRecordID, map[string]any{"base_row_version": 3, "client_txn_id": "txn-u-7-05-row-restore", "reason": "different", "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 2}})
		httptestx.RequireErrorEnvelope(t, divergent, http.StatusConflict, "client_txn_conflict")
	})

	t.Run("whole change set rollback fails all or nothing", func(t *testing.T) {
		unsupportedRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, unsupportedRecordID, "Unsupported after", "unsupported-after", "", "")
		seedHostProjection(t, harness.DB, incidentID, unsupportedRecordID)
		unsupportedChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000535")
		seedRollbackHostAndTagChangeSet(t, harness.DB, incidentID, actorID, unsupportedChangeSetID, unsupportedRecordID)
		beforeCounts := StateCounts(t, harness.DB, unsupportedRecordID)
		resp := rollbackRecord(t, harness, login, unsupportedRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-u-7-05-whole-unsupported", "target": map[string]any{"kind": "change_set", "change_set_id": unsupportedChangeSetID.String()}})
		httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "rollback_precondition_failed")
		afterCounts := StateCounts(t, harness.DB, unsupportedRecordID)
		if beforeCounts != afterCounts {
			t.Fatalf("failed whole change-set rollback mutated state: before=%+v after=%+v", beforeCounts, afterCounts)
		}
		if got := hostDisplayName(t, harness.DB, unsupportedRecordID); got != "unsupported after" {
			t.Fatalf("failed whole change-set rollback changed host display name to %q", got)
		}
		if countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = 'txn-u-7-05-whole-unsupported'`) != 0 {
			t.Fatalf("failed whole change-set rollback inserted rollback change_set")
		}
		history := historyItems(getHistory(t, harness.Server.HTTP.URL, login, unsupportedRecordID, ""))
		requireHistoryActions(t, history[0].(map[string]any), "history_entry", "row_restore")
	})

	t.Run("additional reversible target families", func(t *testing.T) {
		setMembershipRole(t, harness.DB, incidentID, actorID, "reviewer")
		cases := []struct {
			name       string
			seed       func() (uuid.UUID, string, int64, func(map[string]any))
			clientTxn  string
			wantStatus int
			wantCode   string
		}{
			{
				name: "generic workbook record patch",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					recordID := seedRollbackPartyPatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000521"), "party before", "party after")
					return recordID, "href-party-rollback", 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT display_name FROM parties WHERE record_id = $1`, recordID); got != "party before" {
							t.Fatalf("party rollback display_name got %q", got)
						}
					}
				},
				clientTxn: "txn-party-rollback",
			},
			{
				name: "timeline record patch",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					recordID := seedRollbackTimelinePatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000522"), "timeline before", "timeline after")
					return recordID, "href-timeline-rollback", 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT activity_synopsis_text FROM timeline_events WHERE record_id = $1`, recordID); got != "timeline before" {
							t.Fatalf("timeline rollback summary got %q", got)
						}
					}
				},
				clientTxn: "txn-timeline-rollback",
			},
			{
				name: "evidence record patch",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					recordID := seedRollbackEvidencePatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000523"))
					return recordID, "href-evidence-rollback", 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT lifecycle_state FROM evidence WHERE record_id = $1`, recordID); got != "requested" {
							t.Fatalf("evidence rollback lifecycle got %q", got)
						}
					}
				},
				clientTxn: "txn-evidence-rollback",
			},
			{
				name: "record link create",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					src, dst, linkID := seedRollbackRecordLinkCreate(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000524"))
					historyRef := historyEntryRefForTarget(t, harness, login, src, "record_link", linkID.String())
					return src, historyRef, 1, func(data map[string]any) {
						if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NOT NULL`, linkID); got != 1 {
							t.Fatalf("record_link create rollback did not tombstone link")
						}
						requireAffectedRecords(t, data, src, dst)
					}
				},
				clientTxn: "txn-link-create-rollback",
			},
			{
				name: "record link delete",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					src, dst, linkID := seedRollbackRecordLinkDelete(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000525"))
					historyRef := historyEntryRefForTarget(t, harness, login, src, "record_link", linkID.String())
					return src, historyRef, 1, func(data map[string]any) {
						if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, linkID); got != 1 {
							t.Fatalf("record_link delete rollback did not restore link")
						}
						requireAffectedRecords(t, data, src, dst)
					}
				},
				clientTxn: "txn-link-delete-rollback",
			},
			{
				name: "entity mention patch with companion link",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					source, linkID, mentionID := seedRollbackMentionPatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000526"))
					historyRef := historyEntryRefForTarget(t, harness, login, source, "entity_mention", mentionID.String())
					return source, historyRef, 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT resolution_status FROM entity_mentions WHERE source_record_id = $1`, source); got != "unresolved" {
							t.Fatalf("mention rollback status got %q", got)
						}
						if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NOT NULL`, linkID); got != 1 {
							t.Fatalf("mention rollback did not tombstone companion link")
						}
					}
				},
				clientTxn: "txn-mention-rollback",
			},
			{
				name: "entity mention dismiss patch",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					source, mentionID, linkID := seedRollbackMentionDismissPatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000538"))
					historyRef := historyEntryRefForTarget(t, harness, login, source, "entity_mention", mentionID.String())
					return source, historyRef, 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT resolution_status FROM entity_mentions WHERE entity_mention_id = $1`, mentionID); got != "resolved" {
							t.Fatalf("mention dismiss rollback status got %q", got)
						}
						if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NULL`, linkID); got != 1 {
							t.Fatalf("mention dismiss rollback did not restore companion link")
						}
					}
				},
				clientTxn: "txn-mention-dismiss-rollback",
			},
			{
				name: "entity mention restore patch",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					source, mentionID := seedRollbackMentionRestorePatch(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000539"))
					historyRef := historyEntryRefForTarget(t, harness, login, source, "entity_mention", mentionID.String())
					return source, historyRef, 2, func(data map[string]any) {
						if got := stringScalar(t, harness.DB, `SELECT resolution_status FROM entity_mentions WHERE entity_mention_id = $1`, mentionID); got != "dismissed" {
							t.Fatalf("mention restore rollback status got %q", got)
						}
					}
				},
				clientTxn: "txn-mention-restore-rollback",
			},
			{
				name: "supersede link create",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					replacement, superseded, linkID := seedRollbackSupersedesLinkCreate(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000540"))
					historyRef := historyEntryRefForTarget(t, harness, login, replacement, "record_link", linkID.String())
					return replacement, historyRef, 1, func(data map[string]any) {
						if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM record_links WHERE record_link_id = $1 AND deleted_at IS NOT NULL`, linkID); got != 1 {
							t.Fatalf("supersede link rollback did not tombstone link")
						}
						requireAffectedRecords(t, data, replacement, superseded)
					}
				},
				clientTxn: "txn-supersede-link-rollback",
			},
			{
				name: "record tag unsupported",
				seed: func() (uuid.UUID, string, int64, func(map[string]any)) {
					recordID := seedRollbackRecordTagMutation(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000527"), "href-tag-rollback")
					return recordID, "href-tag-rollback", 1, nil
				},
				clientTxn:  "txn-tag-rollback",
				wantStatus: http.StatusConflict,
				wantCode:   "rollback_precondition_failed",
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				recordID, historyRef, baseVersion, verify := tc.seed()
				body := map[string]any{"base_row_version": baseVersion, "client_txn_id": tc.clientTxn, "target": map[string]any{"kind": "history_entry", "history_entry_ref": historyRef}}
				if tc.wantCode != "" {
					httptestx.RequireErrorEnvelope(t, rollbackRecord(t, harness, login, recordID, body), tc.wantStatus, tc.wantCode)
					return
				}
				data := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, recordID, body), http.StatusOK)["data"].(map[string]any)
				if data["rollback_change_set_id"] == "" {
					t.Fatalf("rollback response missing change set: %#v", data)
				}
				if verify != nil {
					verify(data)
				}
			})
		}
	})

	t.Run("precondition reason codes", func(t *testing.T) {
		tagRecordID := seedRollbackRecordTagMutation(t, harness.DB, incidentID, actorID, mustUUID(t, "77777777-0000-4000-8000-000000000529"), "href-tag-reason-rollback")
		requireRollbackReasonCode(t, rollbackRecord(t, harness, login, tagRecordID, map[string]any{"base_row_version": 1, "client_txn_id": "txn-reason-target-not-reversible", "target": map[string]any{"kind": "history_entry", "history_entry_ref": "href-tag-reason-rollback"}}), "target_not_reversible")

		timelineRecordID, evidenceRecordID, attachedChangeSetID, _ := createAttachedEvidencePatchTarget(t, harness, login, incidentID, "entry-requires")
		timelineItem := historyItemForChangeSetTarget(t, harness, login, timelineRecordID, attachedChangeSetID, "timeline_record", timelineRecordID.String())
		timelineRef := stringField(t, timelineItem, "history_entry_ref")
		requireHistoryActions(t, timelineItem, "change_set", "row_restore")
		requireRollbackReasonCode(t, rollbackRecord(t, harness, login, timelineRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-reason-entry-requires-change-set", "target": map[string]any{"kind": "history_entry", "history_entry_ref": timelineRef}}), "entry_requires_change_set")
		changeSetRollback := httptestx.RequireSuccessEnvelope(t, rollbackRecord(t, harness, login, timelineRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-reason-entry-change-set-success", "target": map[string]any{"kind": "change_set", "change_set_id": attachedChangeSetID.String()}}), http.StatusOK)["data"].(map[string]any)
		requireAffectedRecords(t, changeSetRollback, timelineRecordID, evidenceRecordID)

		detachTimelineID, _, detachChangeSetID, detachLinkID := createDetachedEvidencePatchTarget(t, harness, login, incidentID, "entry-requires-detach")
		detachItem := historyItemForChangeSetTarget(t, harness, login, detachTimelineID, detachChangeSetID, "timeline_record", detachTimelineID.String())
		requireHistoryActions(t, detachItem, "change_set", "row_restore")
		if countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no = 2
   AND target_kind = 'record_link'
   AND target_id = $2
   AND operation_kind = 'delete'
`, detachChangeSetID, detachLinkID.String()) != 1 {
			t.Fatalf("attached-evidence detach did not append record_link delete mutation")
		}

		dependentRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, dependentRecordID, "Dependent Host", "dependent-host", "", "")
		dependentChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000530")
		seedRollbackHostPatch(t, harness.DB, incidentID, dependentRecordID, actorID, dependentChangeSetID, time.Date(2026, 5, 10, 17, 3, 0, 0, time.UTC), "dependent before", "dependent after")
		dependentRef := historyEntryRefForTarget(t, harness, login, dependentRecordID, "host", dependentRecordID.String())
		laterChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000531")
		seedRollbackHostPatch(t, harness.DB, incidentID, dependentRecordID, actorID, laterChangeSetID, time.Date(2026, 5, 10, 17, 4, 0, 0, time.UTC), "dependent later before", "dependent later after")
		requireRollbackReasonCode(t, rollbackRecord(t, harness, login, dependentRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-reason-dependent", "target": map[string]any{"kind": "history_entry", "history_entry_ref": dependentRef}}), "dependent_later_changes")

		staleRecordID := uuid.New()
		entitytest.SeedHostRecord(t, harness.DB, incidentID, actorID, staleRecordID, "Stale Target Host", "stale-target-host", "", "")
		staleChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000532")
		seedRollbackHostPatch(t, harness.DB, incidentID, staleRecordID, actorID, staleChangeSetID, time.Date(2026, 5, 10, 17, 5, 0, 0, time.UTC), "stale before", "stale after")
		staleRef := historyEntryRefForTarget(t, harness, login, staleRecordID, "host", staleRecordID.String())
		rollbackChangeSetID := mustUUID(t, "77777777-0000-4000-8000-000000000533")
		seedRollbackHostPatchWithSource(t, harness.DB, incidentID, staleRecordID, actorID, rollbackChangeSetID, time.Date(2026, 5, 10, 17, 6, 0, 0, time.UTC), "stale after", "stale before", "rollback")
		requireRollbackReasonCode(t, rollbackRecord(t, harness, login, staleRecordID, map[string]any{"base_row_version": 2, "client_txn_id": "txn-reason-stale", "target": map[string]any{"kind": "history_entry", "history_entry_ref": staleRef}}), "stale_target")
	})
}

func rollbackRecord(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) *http.Response {
	t.Helper()
	return appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String()+"/rollback", body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
}

func loginLocalUser(t testing.TB, harness *appsupport.ServerHarness, username string, password string) appsupport.LoginResult {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
	var result appsupport.LoginResult
	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case authn.SessionCookieName:
			result.SessionCookie = cookie
		case authn.CSRFCookieName:
			result.CSRFCookie = cookie
		}
	}
	if result.SessionCookie == nil || result.CSRFCookie == nil {
		t.Fatalf("login did not set session and csrf cookies: %#v", resp.Cookies())
	}
	return result
}

func createAttachedEvidencePatchTarget(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, suffix string) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	timelineData := WorkbookCreate(t, harness, login, incidentID, "cartulary.view.timeline.v2", map[string]any{
		"client_txn_id":                   "txn-history_revision-" + suffix + "-timeline-create",
		"timeline.activity_synopsis_text": "Timeline " + suffix,
	})
	timelineRow := timelineData["row"].(map[string]any)
	timelineRecordID := mustUUID(t, timelineRow["record_id"].(string))
	evidenceData := WorkbookCreate(t, harness, login, incidentID, "cartulary.view.evidence.v1", map[string]any{
		"client_txn_id":  "txn-history_revision-" + suffix + "-evidence-create",
		"evidence.title": "Evidence " + suffix,
	})
	evidenceRow := evidenceData["row"].(map[string]any)
	evidenceRecordID := mustUUID(t, evidenceRow["record_id"].(string))

	patchData := WorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": int64(timelineRow["row_version"].(float64)),
		"client_txn_id":    "txn-history_revision-" + suffix + "-attach-evidence",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":               "add_record_ref",
					"linked_record_id": evidenceRecordID.String(),
				}},
			},
		}},
	})
	changeSetID := mustUUID(t, patchData["change_set_id"].(string))
	linkID := mustUUID(t, stringScalar(t, harness.DB, `
SELECT record_link_id::text
  FROM record_links
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
`, incidentID, timelineRecordID, evidenceRecordID))
	if countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations
 WHERE change_set_id = $1
   AND sequence_no = 2
   AND target_kind = 'record_link'
   AND target_id = $2
   AND operation_kind = 'create'
`, changeSetID, linkID.String()) != 1 {
		t.Fatalf("attached-evidence patch did not append record_link create mutation")
	}
	return timelineRecordID, evidenceRecordID, changeSetID, linkID
}

func createDetachedEvidencePatchTarget(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, suffix string) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	timelineRecordID, evidenceRecordID, _, linkID := createAttachedEvidencePatchTarget(t, harness, login, incidentID, suffix+"-attach")
	patchData := WorkbookPatch(t, harness, login, timelineRecordID, map[string]any{
		"view_schema_id":   "cartulary.view.timeline.v2",
		"base_row_version": 2,
		"client_txn_id":    "txn-history_revision-" + suffix + "-detach-evidence",
		"changes": []map[string]any{{
			"field_key": "timeline.attached_evidence_ids",
			"action_payload": map[string]any{
				"kind": "collection_actions_v1",
				"actions": []map[string]any{{
					"op":       "remove_record_ref",
					"item_ref": "record_ref:" + evidenceRecordID.String(),
				}},
			},
		}},
	})
	changeSetID := mustUUID(t, patchData["change_set_id"].(string))
	if countRows(t, harness.DB, `
SELECT COUNT(*)
  FROM record_links
 WHERE record_link_id = $1
   AND deleted_at IS NOT NULL
`, linkID) != 1 {
		t.Fatalf("attached-evidence detach did not tombstone link")
	}
	return timelineRecordID, evidenceRecordID, changeSetID, linkID
}

func WorkbookCreate(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/"+viewSchemaID+"/rows", body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusCreated)["data"].(map[string]any)
}

func WorkbookPatch(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, body map[string]any) map[string]any {
	t.Helper()
	resp := appsupport.DoJSON(t, http.MethodPatch, harness.Server.HTTP.URL+"/api/v1/records/"+recordID.String(), body, appsupport.WithCookies(login.SessionCookie, login.CSRFCookie), appsupport.WithHeader(authn.CSRFHeaderName, login.CSRFCookie.Value))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}

func requireRollbackReasonCode(t testing.TB, resp *http.Response, want string) {
	t.Helper()
	body := httptestx.RequireErrorEnvelope(t, resp, http.StatusConflict, "rollback_precondition_failed")
	errorObject := body["error"].(map[string]any)
	details := errorObject["details"].(map[string]any)
	if details["reason_code"] != want {
		t.Fatalf("rollback reason_code = %#v want %q", details["reason_code"], want)
	}
}

func requireErrorReasonRegistry(t testing.TB, document map[string]any, errorCode string, want ...string) {
	t.Helper()
	rawRegistries, ok := document["reason_registries"].([]any)
	if !ok {
		t.Fatalf("error registry missing reason_registries: %#v", document)
	}
	for _, rawRegistry := range rawRegistries {
		registry, _ := rawRegistry.(map[string]any)
		if registry["error_code"] != errorCode {
			continue
		}
		rawCodes, _ := registry["reason_codes"].([]any)
		got := make([]string, 0, len(rawCodes))
		for _, rawCode := range rawCodes {
			code, _ := rawCode.(map[string]any)
			got = append(got, code["code"].(string))
		}
		if !slices.Equal(got, want) {
			t.Fatalf("%s reason codes got %v want %v", errorCode, got, want)
		}
		return
	}
	t.Fatalf("missing reason registry for %s", errorCode)
}

func requireHistoryActions(t testing.TB, item map[string]any, want ...string) {
	t.Helper()
	raw, ok := item["available_rollback_actions"].([]any)
	if !ok {
		t.Fatalf("history item actions missing or invalid: %#v", item)
	}
	got := make([]string, 0, len(raw))
	for _, value := range raw {
		text, ok := value.(string)
		if !ok {
			t.Fatalf("history item action is not string: %#v", value)
		}
		got = append(got, text)
	}
	if !slices.Equal(got, want) {
		t.Fatalf("history item actions got %v want %v in item %#v", got, want, item)
	}
	if len(want) == 0 {
		if item["reversible"] != false {
			t.Fatalf("history item with no actions must be non-reversible: %#v", item)
		}
		return
	}
	if item["reversible"] != true {
		t.Fatalf("history item with actions must be reversible: %#v", item)
	}
	if slices.Contains(want, "row_restore") {
		if _, ok := item["revision_no"]; !ok {
			t.Fatalf("history item with row_restore must expose revision_no: %#v", item)
		}
		return
	}
	if _, ok := item["revision_no"]; ok {
		t.Fatalf("history item without row_restore must not advertise revision_no: %#v", item)
	}
}

func historyEntryRefForTarget(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, targetKind string, targetID string) string {
	t.Helper()
	for _, rawItem := range historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, "")) {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		diff, _ := item["diff_summary"].(map[string]any)
		units, _ := diff["units"].([]any)
		for _, rawUnit := range units {
			unit, _ := rawUnit.(map[string]any)
			if unit["target_kind"] == targetKind && unit["target_id"] == targetID {
				ref := stringField(t, item, "history_entry_ref")
				if ref == "" {
					t.Fatalf("history item for %s:%s did not expose history_entry_ref: %#v", targetKind, targetID, item)
				}
				return ref
			}
		}
	}
	t.Fatalf("history item for %s:%s not found on record %s", targetKind, targetID, recordID)
	return ""
}

func historyItemForChangeSetTarget(t testing.TB, harness *appsupport.ServerHarness, login appsupport.LoginResult, recordID uuid.UUID, changeSetID uuid.UUID, targetKind string, targetID string) map[string]any {
	t.Helper()
	for _, rawItem := range historyItems(getHistory(t, harness.Server.HTTP.URL, login, recordID, "")) {
		item, ok := rawItem.(map[string]any)
		if !ok || item["change_set_id"] != changeSetID.String() {
			continue
		}
		diff, _ := item["diff_summary"].(map[string]any)
		units, _ := diff["units"].([]any)
		for _, rawUnit := range units {
			unit, _ := rawUnit.(map[string]any)
			if unit["target_kind"] == targetKind && unit["target_id"] == targetID {
				return item
			}
		}
	}
	t.Fatalf("history item for change_set %s target %s:%s not found on record %s", changeSetID, targetKind, targetID, recordID)
	return nil
}

func seedRollbackHostPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time, beforeName string, afterName string) {
	t.Helper()
	seedRollbackHostPatchWithSource(t, db, incidentID, recordID, actorID, changeSetID, createdAt, beforeName, afterName, "workbook.records.patch")
}

func seedRollbackHostPatchWithSource(t testing.TB, db *sql.DB, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, createdAt time.Time, beforeName string, afterName string, source string) {
	t.Helper()
	beforeRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 1}
	afterRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 2}
	beforeSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": beforeName, "hostname": "history_revision-host", "host_state": "canonical", "row_version": 1}
	afterSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": afterName, "hostname": "history_revision-host", "host_state": "canonical", "row_version": 2}
	beforeValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": beforeRecord, "source": beforeSource}
	afterValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": afterRecord, "source": afterSource}
	if _, err := db.ExecContext(context.Background(), `
UPDATE records
   SET row_version = 2,
       updated_at = $4,
       updated_by_user_id = $3
 WHERE record_id = $1 AND incident_id = $2
`, recordID, incidentID, actorID, createdAt); err != nil {
		t.Fatalf("advance rollback host record: %v", err)
	}
	if _, err := db.ExecContext(context.Background(), `
	UPDATE hosts
	   SET display_name = $3,
	       row_version = 2,
       updated_at = $4,
       updated_by_user_id = $5
 WHERE record_id = $1 AND incident_id = $2
	`, recordID, incidentID, afterName, createdAt, actorID); err != nil {
		t.Fatalf("advance rollback host source: %v", err)
	}
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: changeSetID, CreatedAt: createdAt, Source: source})
	if _, err := db.ExecContext(context.Background(), `
	INSERT INTO change_set_mutations (
	    change_set_id, sequence_no, target_kind, target_id, operation_kind,
	    before_version_id, after_version_id, before_value, after_value,
	    history_record_ids, history_entry_record_ids
)
VALUES ($1, 1, 'host', $2::text, 'field_update', $3, $4, $5, $6, ARRAY[$2::uuid], ARRAY[$2::uuid])
`, changeSetID, recordID.String(), "host:"+recordID.String()+":1", "host:"+recordID.String()+":2", jsonOrNil(t, beforeValue), jsonOrNil(t, afterValue)); err != nil {
		t.Fatalf("seed rollback host mutation: %v", err)
	}
	seedHistoricalRecordRevision(
		t,
		db,
		changeSetID,
		recordID,
		jsonOrNil(t, beforeValue),
		jsonOrNil(t, afterValue),
		createdAt,
	)
	seedHostProjection(t, db, incidentID, recordID)
}

func seedRollbackTwoHostChangeSet(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, left uuid.UUID, right uuid.UUID) {
	t.Helper()
	createdAt := time.Date(2026, 5, 10, 17, 6, 0, 0, time.UTC)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: left, ChangeSetID: changeSetID, CreatedAt: createdAt, Source: "workbook.records.patch"})
	seedRollbackHostPatchEntry(t, db, incidentID, actorID, changeSetID, left, 1, createdAt, "left before", "left after")
	seedRollbackHostPatchEntry(t, db, incidentID, actorID, changeSetID, right, 2, createdAt, "right before", "right after")
}

func seedRollbackHostAndTagChangeSet(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, recordID uuid.UUID) {
	t.Helper()
	createdAt := time.Date(2026, 5, 10, 17, 8, 0, 0, time.UTC)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: changeSetID, CreatedAt: createdAt, Source: "workbook.records.patch"})
	seedRollbackHostPatchEntry(t, db, incidentID, actorID, changeSetID, recordID, 1, createdAt, "unsupported before", "unsupported after")
	insertMutation(t, db, changeSetID, 2, "record_tag", uuid.New().String(), "create", nil, map[string]any{"tag_name": "history_revision", "record_id": recordID.String()})
}

func seedRollbackHostPatchEntry(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, recordID uuid.UUID, sequenceNo int, createdAt time.Time, beforeName string, afterName string) {
	t.Helper()
	beforeRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 1}
	afterRecord := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "record_type": "host", "row_version": 2}
	beforeSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": beforeName, "hostname": "history_revision-host", "host_state": "canonical", "row_version": 1}
	afterSource := map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String(), "display_name": afterName, "hostname": "history_revision-host", "host_state": "canonical", "row_version": 2}
	beforeValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": beforeRecord, "source": beforeSource}
	afterValue := map[string]any{"snapshot_schema_id": "cartulary.revisions.snapshot.host.v1", "record": afterRecord, "source": afterSource}
	mustExec(t, db, `
UPDATE records
   SET row_version = 2,
       updated_at = $4,
       updated_by_user_id = $3
 WHERE record_id = $1 AND incident_id = $2
`, recordID, incidentID, actorID, createdAt)
	mustExec(t, db, `
UPDATE hosts
   SET display_name = $3,
       row_version = 2,
       updated_at = $4,
       updated_by_user_id = $5
 WHERE record_id = $1 AND incident_id = $2
`, recordID, incidentID, afterName, createdAt, actorID)
	insertMutation(t, db, changeSetID, sequenceNo, "host", recordID.String(), "field_update", beforeValue, afterValue)
	seedHistoricalRecordRevision(
		t,
		db,
		changeSetID,
		recordID,
		jsonOrNil(t, beforeValue),
		jsonOrNil(t, afterValue),
		createdAt,
	)
	seedHostProjection(t, db, incidentID, recordID)
}

func seedHistoricalRecordRevision(
	t testing.TB,
	db *sql.DB,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
	beforeValue any,
	afterValue any,
	createdAt time.Time,
) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin historical revision seed: %v", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()
	if err := collaboration.NewHistoricalIntentPolicy().SuppressSQLTx(ctx, tx); err != nil {
		t.Fatalf("suppress seeded historical revision intent: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO record_revisions (change_set_id, record_id, row_version, before_json, after_json, created_at)
VALUES ($1, $2, 2, $3, $4, $5)
ON CONFLICT (record_id, row_version) DO NOTHING
`, changeSetID, recordID, beforeValue, afterValue, createdAt); err != nil {
		t.Fatalf("seed rollback host revision: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit historical revision seed: %v", err)
	}
}

func seedRollbackPartyPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, beforeName string, afterName string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, recordID, "party")
	mustExec(t, db, `INSERT INTO parties (record_id, incident_id, display_name, party_kind, updated_at) VALUES ($1, $2, $3, 'person', $4)`, recordID, incidentID, afterName, time.Now().UTC())
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, recordID)
	before := canonicalRowSnapshot(recordID, incidentID, "party", "cartulary.revisions.snapshot.party.v1", 1, map[string]any{"display_name": beforeName, "party_kind": "person"})
	after := canonicalRowSnapshot(recordID, incidentID, "party", "cartulary.revisions.snapshot.party.v1", 2, map[string]any{"display_name": afterName, "party_kind": "person"})
	seedRollbackMutationWithRef(t, db, incidentID, actorID, recordID, changeSetID, 1, "record", recordID.String(), "patch", before, after, "href-party-rollback")
	return recordID
}

func seedRollbackTimelinePatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, beforeSummary string, afterSummary string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, recordID, "timeline_event")
	mustExec(t, db, `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, raw_activity_text, data_source_text, capture_state, row_version, recorded_at, edited_at, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, 'details', 'source', 'rough', 2, $4, $4, $5, $5)
`, recordID, incidentID, afterSummary, time.Now().UTC(), actorID)
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, recordID)
	before := canonicalRowSnapshot(recordID, incidentID, "timeline_event", "cartulary.revisions.snapshot.timeline_event.v1", 1, map[string]any{"activity_synopsis_text": beforeSummary, "raw_activity_text": "details", "data_source_text": "source", "capture_state": "rough"})
	after := canonicalRowSnapshot(recordID, incidentID, "timeline_event", "cartulary.revisions.snapshot.timeline_event.v1", 2, map[string]any{"activity_synopsis_text": afterSummary, "raw_activity_text": "details", "data_source_text": "source", "capture_state": "rough"})
	seedRollbackMutationWithRef(t, db, incidentID, actorID, recordID, changeSetID, 1, "timeline_record", recordID.String(), "patch", before, after, "href-timeline-rollback")
	return recordID
}

func seedRollbackEvidencePatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, recordID, "evidence")
	mustExec(t, db, `INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, updated_at) VALUES ($1, $2, 'Evidence after', 'available', 'available', $3)`, recordID, incidentID, time.Now().UTC())
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, recordID)
	before := canonicalRowSnapshot(recordID, incidentID, "evidence", "cartulary.revisions.snapshot.evidence.v1", 1, map[string]any{"title": "Evidence before", "lifecycle_state": "requested", "upload_state": "pending"})
	after := canonicalRowSnapshot(recordID, incidentID, "evidence", "cartulary.revisions.snapshot.evidence.v1", 2, map[string]any{"title": "Evidence after", "lifecycle_state": "available", "upload_state": "available"})
	seedRollbackMutationWithRef(t, db, incidentID, actorID, recordID, changeSetID, 1, "record", recordID.String(), "patch", before, after, "href-evidence-rollback")
	return recordID
}

func seedRollbackRecordLinkCreate(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	src, dst := seedRollbackHostPair(t, db, incidentID, actorID, "Link Src", "Link Dst")
	linkID := uuid.New()
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id, decided_at, created_at)
VALUES ($1, $2, $3, $4, 'references_record', 'manual', $5, $5, $6, $6)
	`, linkID, incidentID, src, dst, actorID, time.Now().UTC())
	after := linkValue(linkID, incidentID, src, dst, "references_record", nil, "manual", actorID, nil)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: src, ChangeSetID: changeSetID, CreatedAt: time.Now().UTC(), Source: "history_revision.rollback.fixture"})
	insertMutation(t, db, changeSetID, 1, "record_link", linkID.String(), "create", nil, after)
	return src, dst, linkID
}

func seedRollbackRecordLinkDelete(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	src, dst := seedRollbackHostPair(t, db, incidentID, actorID, "Deleted Link Src", "Deleted Link Dst")
	linkID := uuid.New()
	deletedAt := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id, decided_at, created_at, deleted_at, deleted_by_user_id)
VALUES ($1, $2, $3, $4, 'references_record', 'manual', $5, $5, $6, $6, $7, $5)
	`, linkID, incidentID, src, dst, actorID, deletedAt.Add(-time.Minute), deletedAt)
	before := linkValue(linkID, incidentID, src, dst, "references_record", nil, "manual", actorID, nil)
	after := linkValue(linkID, incidentID, src, dst, "references_record", nil, "manual", actorID, &deletedAt)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: src, ChangeSetID: changeSetID, CreatedAt: time.Now().UTC(), Source: "history_revision.rollback.fixture"})
	insertMutation(t, db, changeSetID, 1, "record_link", linkID.String(), "delete", before, after)
	return src, dst, linkID
}

func seedRollbackAttachedEvidenceLinkCreate(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	src, dst := seedRollbackTimelineEvidencePair(t, db, incidentID, actorID, "Attached Evidence Create", 1)
	linkID := uuid.New()
	now := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at)
VALUES ($1, $2, $3, $4, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', $5, $5, $6, $6)
`, linkID, incidentID, src, dst, actorID, now)
	after := linkValueWithTimestamps(linkID, incidentID, src, dst, "attached_evidence", stringPtr("timeline.attached_evidence_ids"), "manual", actorID, now, nil)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: src, ChangeSetID: changeSetID, CreatedAt: now, Source: "history_revision.rollback.fixture"})
	insertMutation(t, db, changeSetID, 1, "record_link", linkID.String(), "create", nil, after)
	return src, dst, linkID
}

func seedRollbackAttachedEvidenceLinkDelete(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	src, dst := seedRollbackTimelineEvidencePair(t, db, incidentID, actorID, "Attached Evidence Delete", 0)
	linkID := uuid.New()
	now := time.Now().UTC()
	deletedAt := now.Add(time.Minute)
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at, deleted_at, deleted_by_user_id)
VALUES ($1, $2, $3, $4, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', $5, $5, $6, $6, $7, $5)
`, linkID, incidentID, src, dst, actorID, now, deletedAt)
	before := linkValueWithTimestamps(linkID, incidentID, src, dst, "attached_evidence", stringPtr("timeline.attached_evidence_ids"), "manual", actorID, now, nil)
	after := linkValueWithTimestamps(linkID, incidentID, src, dst, "attached_evidence", stringPtr("timeline.attached_evidence_ids"), "manual", actorID, now, &deletedAt)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: src, ChangeSetID: changeSetID, CreatedAt: deletedAt, Source: "history_revision.rollback.fixture"})
	insertMutation(t, db, changeSetID, 1, "record_link", linkID.String(), "delete", before, after)
	return src, dst, linkID
}

func seedRollbackTimelineEvidencePair(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, label string, evidenceCount int) (uuid.UUID, uuid.UUID) {
	t.Helper()
	now := time.Now().UTC()
	timelineID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, timelineID, "timeline_event")
	mustExec(t, db, `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, raw_activity_text, data_source_text, capture_state, row_version, recorded_at, edited_at, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, $3, 'details', 'source', 'rough', 1, $4, $4, $5, $5)
`, timelineID, incidentID, label, now, actorID)
	mustExec(t, db, `
INSERT INTO timeline_grid_projection (
    record_id, incident_id, row_version, activity_synopsis_text, raw_activity_text, data_source_text, recorded_at, edited_at,
    capture_state, evidence_count, has_evidence
) VALUES ($1, $2, 1, $3, 'details', 'source', $4, $4, 'rough', $5, $6)
`, timelineID, incidentID, label, now, evidenceCount, evidenceCount > 0)

	evidenceID := uuid.New()
	blobID := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, evidenceID, "evidence")
	mustExec(t, db, `
INSERT INTO object_blobs (
    object_blob_id, incident_id, created_by_user_id, storage_key, upload_state,
    byte_size, target_expires_at, pending_expires_at, finalized_at, created_at, updated_at
) VALUES ($1, $2, $3, $4, 'available', 1, $5, $5, $6, $6, $6)
`, blobID, incidentID, actorID, "history_revision/rollback/"+blobID.String(), now.Add(time.Hour), now)
	mustExec(t, db, `
INSERT INTO evidence (record_id, incident_id, title, lifecycle_state, upload_state, object_blob_id, requested_at, received_at, created_at, updated_at)
VALUES ($1, $2, $3, 'available', 'available', $4, $5, $5, $5, $5)
`, evidenceID, incidentID, label+" Evidence", blobID, now)
	mustExec(t, db, `
INSERT INTO evidence_grid_projection (
    record_id, incident_id, row_version, title, lifecycle_state, requested_at, received_at,
    blob_hash, upload_state, linked_record_count, edited_at
) VALUES ($1, $2, 1, $3, 'available', $4, $4, NULL, 'available', $5, $4)
`, evidenceID, incidentID, label+" Evidence", now, evidenceCount)
	if evidenceCount > 0 {
		mustExec(t, db, `
UPDATE timeline_grid_projection
   SET attached_evidence_refs = jsonb_build_array(jsonb_build_object(
           'item_ref', 'record_ref:' || $2::text,
           'item_kind', 'record_ref',
           'display_text', $3::text,
           'linked_record_id', $2::text
       ))
 WHERE record_id = $1
`, timelineID, evidenceID, label+" Evidence")
	}
	return timelineID, evidenceID
}

func seedRollbackMentionPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	source := uuid.New()
	envelopetest.SeedRecordEnvelope(t, db, incidentID, actorID, source, "timeline_event")
	mustExec(t, db, `
INSERT INTO timeline_events (record_id, incident_id, activity_synopsis_text, raw_activity_text, data_source_text, capture_state, row_version, recorded_at, edited_at, created_by_user_id, updated_by_user_id)
VALUES ($1, $2, 'mention source', 'details', 'source', 'rough', 2, $3, $3, $4, $4)
`, source, incidentID, time.Now().UTC(), actorID)
	target, _ := seedRollbackHostPair(t, db, incidentID, actorID, "Mention Host", "Mention Other")
	mentionID := uuid.New()
	linkID := uuid.New()
	now := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO entity_mentions (
    entity_mention_id, source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal, created_by_user_id, created_at,
    resolved_record_id, resolved_by_user_id, resolved_at, resolution_method
) VALUES ($1, $2, 'host', 'timeline.host_refs', 'manual', 'history_revision', 'Mention Host', 'mention host', 'resolved', 2, 1, $3, $4, $5, $3, $4, 'explicit_resolve_route')
`, mentionID, source, actorID, now, target)
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at)
VALUES ($1, $2, $3, $4, 'observed_on_host', 'timeline.host_refs', 'manual', $5, $5, $6, $6)
`, linkID, incidentID, source, target, actorID, now)
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, source)
	beforeMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision", "raw_text": "Mention Host", "normalized_text": "mention host", "resolution_status": "unresolved", "row_version": 1, "ordinal": 1}
	afterMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision", "raw_text": "Mention Host", "normalized_text": "mention host", "resolution_status": "resolved", "row_version": 2, "ordinal": 1, "resolved_record_id": target.String(), "resolved_by_user_id": actorID.String(), "resolved_at": now.Format(time.RFC3339Nano), "resolution_method": "explicit_resolve_route"}
	linkAfter := linkValue(linkID, incidentID, source, target, "observed_on_host", stringPtr("timeline.host_refs"), "manual", actorID, nil)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: source, ChangeSetID: changeSetID, CreatedAt: now, Source: "entities.mentions.action"})
	insertMutation(t, db, changeSetID, 1, "entity_mention", mentionID.String(), "patch", beforeMention, afterMention)
	insertMutation(t, db, changeSetID, 2, "record_link", linkID.String(), "create", nil, linkAfter)
	return source, linkID, mentionID
}

func seedRollbackMentionDismissPatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	source := uuid.New()
	timelinetest.SeedTimelineRecord(t, db, incidentID, actorID, source)
	target, _ := seedRollbackHostPair(t, db, incidentID, actorID, "Dismiss Mention Host", "Dismiss Mention Other")
	mentionID := uuid.New()
	linkID := uuid.New()
	now := time.Now().UTC()
	deletedAt := now.Add(time.Minute)
	mustExec(t, db, `
INSERT INTO entity_mentions (
    entity_mention_id, source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal, created_by_user_id, created_at,
    resolved_record_id, resolved_by_user_id, resolved_at, resolution_method
) VALUES ($1, $2, 'host', 'timeline.host_refs', 'manual', 'history_revision-dismiss', 'Dismiss Host', 'dismiss host', 'dismissed', 2, 1, $3, $4, NULL, NULL, NULL, NULL)
`, mentionID, source, actorID, now)
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, field_key, provenance, owner_user_id, created_by_user_id, decided_at, created_at, deleted_at, deleted_by_user_id)
VALUES ($1, $2, $3, $4, 'observed_on_host', 'timeline.host_refs', 'manual', $5, $5, $6, $6, $7, $5)
`, linkID, incidentID, source, target, actorID, now, deletedAt)
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, source)
	beforeMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision-dismiss", "raw_text": "Dismiss Host", "normalized_text": "dismiss host", "resolution_status": "resolved", "row_version": 1, "ordinal": 1, "resolved_record_id": target.String(), "resolved_by_user_id": actorID.String(), "resolved_at": now.Format(time.RFC3339Nano), "resolution_method": "explicit_resolve_route"}
	afterMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision-dismiss", "raw_text": "Dismiss Host", "normalized_text": "dismiss host", "resolution_status": "dismissed", "row_version": 2, "ordinal": 1}
	linkBefore := linkValueWithTimestamps(linkID, incidentID, source, target, "observed_on_host", stringPtr("timeline.host_refs"), "manual", actorID, now, nil)
	linkAfter := linkValueWithTimestamps(linkID, incidentID, source, target, "observed_on_host", stringPtr("timeline.host_refs"), "manual", actorID, now, &deletedAt)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: source, ChangeSetID: changeSetID, CreatedAt: now, Source: "entities.mentions.action"})
	insertMutation(t, db, changeSetID, 1, "entity_mention", mentionID.String(), "patch", beforeMention, afterMention)
	insertMutation(t, db, changeSetID, 2, "record_link", linkID.String(), "delete", linkBefore, linkAfter)
	return source, mentionID, linkID
}

func seedRollbackMentionRestorePatch(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID) {
	t.Helper()
	source := uuid.New()
	timelinetest.SeedTimelineRecord(t, db, incidentID, actorID, source)
	mentionID := uuid.New()
	now := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO entity_mentions (
    entity_mention_id, source_record_id, entity_type, source_field_key, origin_kind, origin_locator,
    raw_text, normalized_text, resolution_status, row_version, ordinal, created_by_user_id, created_at
) VALUES ($1, $2, 'host', 'timeline.host_refs', 'manual', 'history_revision-restore', 'Restore Host', 'restore host', 'unresolved', 2, 1, $3, $4)
`, mentionID, source, actorID, now)
	mustExec(t, db, `UPDATE records SET row_version = 2 WHERE record_id = $1`, source)
	beforeMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision-restore", "raw_text": "Restore Host", "normalized_text": "restore host", "resolution_status": "dismissed", "row_version": 1, "ordinal": 1}
	afterMention := map[string]any{"entity_mention_id": mentionID.String(), "source_record_id": source.String(), "entity_type": "host", "source_field_key": "timeline.host_refs", "origin_kind": "manual", "origin_locator": "history_revision-restore", "raw_text": "Restore Host", "normalized_text": "restore host", "resolution_status": "unresolved", "row_version": 2, "ordinal": 1}
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: source, ChangeSetID: changeSetID, CreatedAt: now, Source: "entities.mentions.action"})
	insertMutation(t, db, changeSetID, 1, "entity_mention", mentionID.String(), "patch", beforeMention, afterMention)
	return source, mentionID
}

func seedRollbackSupersedesLinkCreate(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID) (uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	replacement := uuid.New()
	superseded := uuid.New()
	timelinetest.SeedTimelineRecord(t, db, incidentID, actorID, replacement)
	timelinetest.SeedTimelineRecord(t, db, incidentID, actorID, superseded)
	linkID := uuid.New()
	now := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id, decided_at, created_at)
VALUES ($1, $2, $3, $4, 'supersedes', 'manual', $5, $5, $6, $6)
`, linkID, incidentID, replacement, superseded, actorID, now)
	after := linkValueWithTimestamps(linkID, incidentID, replacement, superseded, "supersedes", nil, "manual", actorID, now, nil)
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: replacement, ChangeSetID: changeSetID, CreatedAt: now, Source: "timeline.records.supersede"})
	insertMutation(t, db, changeSetID, 1, "record_link", linkID.String(), "create", nil, after)
	return replacement, superseded, linkID
}

func seedRollbackRecordTagMutation(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, changeSetID uuid.UUID, historyRef string) uuid.UUID {
	t.Helper()
	recordID := uuid.New()
	entitytest.SeedHostRecord(t, db, incidentID, actorID, recordID, "Tag Host", "tag-host", "", "")
	seedRollbackMutationWithRef(t, db, incidentID, actorID, recordID, changeSetID, 1, "record_tag", uuid.New().String(), "create", nil, map[string]any{
		"record_id": recordID.String(),
		"tag_name":  "history_revision",
	}, historyRef)
	return recordID
}

func seedRecordLinkForRowRestore(t testing.TB, db *sql.DB, incidentID uuid.UUID, src uuid.UUID, actorID uuid.UUID) uuid.UUID {
	t.Helper()
	dst := uuid.New()
	entitytest.SeedHostRecord(t, db, incidentID, actorID, dst, "Row Restore Linked Host", "row-restore-linked", "", "")
	seedHostProjection(t, db, incidentID, dst)
	linkID := uuid.New()
	now := time.Now().UTC()
	mustExec(t, db, `
INSERT INTO record_links (record_link_id, incident_id, src_record_id, dst_record_id, link_type, provenance, owner_user_id, created_by_user_id, decided_at, created_at)
VALUES ($1, $2, $3, $4, 'references_record', 'manual', $5, $5, $6, $6)
`, linkID, incidentID, src, dst, actorID, now)
	return linkID
}

func seedRollbackHostPair(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, leftName string, rightName string) (uuid.UUID, uuid.UUID) {
	t.Helper()
	left := uuid.New()
	right := uuid.New()
	entitytest.SeedHostRecord(t, db, incidentID, actorID, left, leftName, leftName, "", "")
	entitytest.SeedHostRecord(t, db, incidentID, actorID, right, rightName, rightName, "", "")
	seedHostProjection(t, db, incidentID, left)
	seedHostProjection(t, db, incidentID, right)
	return left, right
}

func seedRollbackMutationWithRef(t testing.TB, db *sql.DB, incidentID uuid.UUID, actorID uuid.UUID, recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int, targetKind string, targetID string, operation string, before any, after any, historyRef string) {
	t.Helper()
	seedChangeSet(t, db, historySeed{IncidentID: incidentID, ActorID: actorID, RecordID: recordID, ChangeSetID: changeSetID, CreatedAt: time.Now().UTC(), Source: "history_revision.rollback.fixture"})
	insertMutation(t, db, changeSetID, sequenceNo, targetKind, targetID, operation, before, after)
	seedHistoryRef(t, db, recordID, changeSetID, sequenceNo, historyRef)
}

func insertMutation(t testing.TB, db *sql.DB, changeSetID uuid.UUID, sequenceNo int, targetKind string, targetID string, operation string, before any, after any) {
	t.Helper()
	catalog, err := revisionassembly.CurrentTargetSemanticsCatalog()
	if err != nil {
		t.Fatalf("build seeded target-semantics catalog: %v", err)
	}
	history, err := catalog.DescribeValues(targetKind, targetID, before, after)
	if err != nil {
		t.Fatalf("describe seeded mutation history: %v", err)
	}
	mustExec(t, db, `
INSERT INTO change_set_mutations (
    change_set_id, sequence_no, target_kind, target_id, operation_kind,
    before_value, after_value, history_record_ids, history_entry_record_ids
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, changeSetID, sequenceNo, targetKind, targetID, operation, jsonOrNil(t, before), jsonOrNil(t, after), history.HistoryRecordIDs, history.HistoryEntryRecordIDs)
}

func seedHistoryRef(t testing.TB, db *sql.DB, recordID uuid.UUID, changeSetID uuid.UUID, sequenceNo int, ref string) {
	t.Helper()
	mustExec(t, db, `
INSERT INTO record_history_entry_refs (history_entry_ref, record_id, change_set_id, mutation_sequence_no)
VALUES ($1, $2, $3, $4)
`, ref, recordID, changeSetID, sequenceNo)
}

func linkValue(linkID uuid.UUID, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, linkType string, fieldKey *string, provenance string, actorID uuid.UUID, deletedAt *time.Time) map[string]any {
	value := map[string]any{"record_link_id": linkID.String(), "incident_id": incidentID.String(), "src_record_id": src.String(), "dst_record_id": dst.String(), "link_type": linkType, "provenance": provenance, "owner_user_id": actorID.String(), "created_by_user_id": actorID.String()}
	if fieldKey != nil {
		value["field_key"] = *fieldKey
	}
	if deletedAt != nil {
		value["deleted_at"] = deletedAt.Format(time.RFC3339Nano)
		value["deleted_by_user_id"] = actorID.String()
	}
	return value
}

func linkValueWithTimestamps(linkID uuid.UUID, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, linkType string, fieldKey *string, provenance string, actorID uuid.UUID, createdAt time.Time, deletedAt *time.Time) map[string]any {
	value := linkValue(linkID, incidentID, src, dst, linkType, fieldKey, provenance, actorID, deletedAt)
	value["decided_at"] = createdAt.Format(time.RFC3339Nano)
	value["created_at"] = createdAt.Format(time.RFC3339Nano)
	return value
}

func mustExec(t testing.TB, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.ExecContext(context.Background(), query, args...); err != nil {
		t.Fatalf("exec fixture SQL: %v\nSQL: %s", err, query)
	}
}

func stringScalar(t testing.TB, db *sql.DB, query string, args ...any) string {
	t.Helper()
	var value string
	if err := db.QueryRowContext(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query scalar: %v", err)
	}
	return value
}

func requireAffectedRecords(t testing.TB, data map[string]any, want ...uuid.UUID) {
	t.Helper()
	raw := data["affected_record_ids"].([]any)
	got := make([]string, 0, len(raw))
	for _, value := range raw {
		got = append(got, value.(string))
	}
	wantText := make([]string, 0, len(want))
	for _, id := range want {
		wantText = append(wantText, id.String())
	}
	slices.Sort(got)
	slices.Sort(wantText)
	if !slices.Equal(got, wantText) {
		t.Fatalf("affected records got %v want %v", got, wantText)
	}
}

func requireAffectedRecordsCanonical(t testing.TB, data map[string]any, want ...uuid.UUID) {
	t.Helper()
	raw := data["affected_record_ids"].([]any)
	got := make([]string, 0, len(raw))
	for _, value := range raw {
		got = append(got, value.(string))
	}
	wantText := make([]string, 0, len(want))
	for _, id := range want {
		wantText = append(wantText, id.String())
	}
	slices.Sort(wantText)
	if !slices.Equal(got, wantText) {
		t.Fatalf("affected records got %v want canonical %v", got, wantText)
	}
}

func stringPtr(value string) *string {
	return &value
}

func hostDisplayName(t testing.TB, db *sql.DB, recordID uuid.UUID) string {
	t.Helper()
	var value string
	if err := db.QueryRowContext(context.Background(), `SELECT display_name FROM hosts WHERE record_id = $1`, recordID).Scan(&value); err != nil {
		t.Fatalf("load host display_name: %v", err)
	}
	return value
}

func hostProjectionDisplayName(t testing.TB, db *sql.DB, recordID uuid.UUID) string {
	t.Helper()
	var value string
	if err := db.QueryRowContext(context.Background(), `SELECT display_name FROM host_grid_projection WHERE record_id = $1`, recordID).Scan(&value); err != nil {
		t.Fatalf("load host projection display_name: %v", err)
	}
	return value
}

func timelineProjectionEvidenceState(t testing.TB, db *sql.DB, recordID uuid.UUID) (int, bool) {
	t.Helper()
	var count int
	var hasEvidence bool
	if err := db.QueryRowContext(context.Background(), `
SELECT evidence_count, has_evidence
  FROM timeline_grid_projection
 WHERE record_id = $1
`, recordID).Scan(&count, &hasEvidence); err != nil {
		t.Fatalf("load timeline evidence projection: %v", err)
	}
	return count, hasEvidence
}
