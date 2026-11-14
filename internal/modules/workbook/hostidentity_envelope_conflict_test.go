package workbook_test

import (
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
	"net/http"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func exerciseHostIdentityEnvelopeConflict(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "hostidentity-envelope-conflicts", "IR-ENTITY-ENVELOPE")
	for _, tc := range []struct{ kind, view, identifier string }{
		{"host", "cartulary.view.hosts.v1", "host.hostname"},
		{"identity", "cartulary.view.identities.v1", "identity.upn"},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			field := tc.kind + ".display_name"
			row := requireWorkbookCreate(t, harness, login, incidentID, tc.view, map[string]any{"client_txn_id": "envelope-create-" + tc.kind, field: "Original display", tc.identifier: "envelope-" + tc.kind + "@example.test"})["row"].(map[string]any)
			id := appsupport.MustUUID(t, row["record_id"].(string))
			patch := map[string]any{"view_schema_id": tc.view, "base_row_version": 1, "client_txn_id": "envelope-patch-" + tc.kind, "changes": []map[string]any{{"field_key": field, "value": "Saved display"}}}
			applied := requireWorkbookPatch(t, harness, login, id, patch)
			requireCellValue(t, applied["row"].(map[string]any), field, "Saved display")
			if applied["row"].(map[string]any)["row_version"] != float64(2) {
				t.Fatalf("patch version=%#v", applied)
			}
			state := snapshotWorkbookState(t, harness, login, incidentID, tc.view)
			replay := requireWorkbookPatch(t, harness, login, id, patch)
			if !reflect.DeepEqual(replay, applied) {
				t.Fatalf("patch replay changed: %#v versus %#v", replay, applied)
			}
			requireWorkbookStateUnchanged(t, state, snapshotWorkbookState(t, harness, login, incidentID, tc.view))
			// A future base version is a row conflict independent of the
			// separate stale-field rebase/conflict behavior.
			future := map[string]any{"view_schema_id": tc.view, "base_row_version": 3, "client_txn_id": "envelope-future-" + tc.kind, "changes": []map[string]any{{"field_key": field, "value": "Future display"}}}
			failure := httptestx.RequireErrorEnvelope(t, doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", id, future), http.StatusConflict, "row_version_conflict")
			details := failure["error"].(map[string]any)["details"].(map[string]any)
			if details["record_id"] != id.String() || details["base_row_version"] != float64(3) || details["current_row_version"] != float64(2) {
				t.Fatalf("future conflict=%#v", details)
			}
			requireWorkbookStateUnchanged(t, state, snapshotWorkbookState(t, harness, login, incidentID, tc.view))

		})
	}
}
