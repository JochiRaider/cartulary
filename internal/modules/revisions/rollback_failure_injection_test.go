package revisions_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestRollbackTransactionalFailureInjection_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, "revisions-rollback-failure-injection")
	login, actorID := appsupport.ProvisionBootstrapAdmin(t, harness.Server)
	stages := []struct {
		name   string
		table  string
		events string
	}{
		{name: "projection", table: "host_grid_projection", events: "DELETE OR INSERT OR UPDATE"},
		{name: "mutation_history", table: "change_set_mutations", events: "INSERT"},
		{name: "revision_history", table: "record_revisions", events: "INSERT"},
		{name: "idempotency_result", table: "route_idempotency", events: "INSERT"},
	}
	for index, stage := range stages {
		t.Run(stage.name, func(t *testing.T) {
			incidentID, recordID := seedRecord(t, harness.DB, harness.Server, login, actorID, fmt.Sprintf("IR-ROLLBACK-FAILURE-%d", index))
			setMembershipRole(t, harness.DB, incidentID, actorID, "reviewer")
			changeSetID := uuid.New()
			seedRollbackHostPatch(t, harness.DB, incidentID, recordID, actorID, changeSetID, time.Date(2026, 8, 3, 17, index, 0, 0, time.UTC), "failure before", "failure after")
			historyEntryRef := historyEntryRefForTarget(t, harness, login, recordID, "host", recordID.String())
			before := StateCounts(t, harness.DB, recordID)

			functionName := "revisions_rollback_fail_" + stage.name
			triggerName := functionName + "_trigger"
			mustExec(t, harness.DB, fmt.Sprintf(`
CREATE FUNCTION %s() RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'injected rollback failure at %s';
END
$$
`, functionName, stage.name))
			mustExec(t, harness.DB, fmt.Sprintf(`
CREATE TRIGGER %s
BEFORE %s ON %s
FOR EACH ROW EXECUTE FUNCTION %s()
`, triggerName, stage.events, stage.table, functionName))
			t.Cleanup(func() {
				mustExec(t, harness.DB, fmt.Sprintf("DROP TRIGGER IF EXISTS %s ON %s", triggerName, stage.table))
				mustExec(t, harness.DB, fmt.Sprintf("DROP FUNCTION IF EXISTS %s()", functionName))
			})

			clientTxnID := "txn-rollback-failure-" + stage.name
			response := rollbackRecord(t, harness, login, recordID, map[string]any{
				"base_row_version": 2,
				"client_txn_id":    clientTxnID,
				"target":           map[string]any{"kind": "history_entry", "history_entry_ref": historyEntryRef},
			})
			if response.StatusCode != http.StatusInternalServerError {
				t.Fatalf("injected %s status = %d, want %d", stage.name, response.StatusCode, http.StatusInternalServerError)
			}
			_ = response.Body.Close()

			after := StateCounts(t, harness.DB, recordID)
			if after != before {
				t.Fatalf("%s failure left partial state: before=%+v after=%+v", stage.name, before, after)
			}
			if got := hostDisplayName(t, harness.DB, recordID); got != "failure after" {
				t.Fatalf("%s failure changed source row to %q", stage.name, got)
			}
			if got := hostProjectionDisplayName(t, harness.DB, recordID); got != "failure after" {
				t.Fatalf("%s failure changed projection row to %q", stage.name, got)
			}
			if got := countRows(t, harness.DB, `SELECT COUNT(*) FROM change_sets WHERE source = 'rollback' AND client_txn_id = $1`, clientTxnID); got != 0 {
				t.Fatalf("%s failure retained %d rollback change sets", stage.name, got)
			}
		})
	}
}
