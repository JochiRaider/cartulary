package evidence_test

import (
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestEvidenceAccountableOwner_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-accountable-owner", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{"client_txn_id": "owner-incident", "incident_key": "OWNER", "title": "Accountable Evidence"})
	incidentID := uuid.MustParse(incident["incident_id"].(string))
	owner := flowtest.SeedLocalUserRecord(t, harness.DB, "owner@example.test", "Assigned owner", "OwnerAssignment1!", false, false, true)
	inactive := flowtest.SeedLocalUserRecord(t, harness.DB, "inactive-owner@example.test", "Inactive owner", "OwnerAssignment1!", false, false, false)
	foreign := flowtest.SeedLocalUserRecord(t, harness.DB, "foreign-owner@example.test", "Nonmember", "OwnerAssignment1!", false, false, true)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, owner.ID, "viewer", adminID)
	appsupport.SeedMembershipSQL(t, harness.DB, incidentID, inactive.ID, "viewer", adminID)
	const schema = "cartulary.view.evidence.v1"
	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/" + schema + "/rows"
	call := func(method, url string, body map[string]any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, method, url, body, authOptions(login)...)
	}
	checkRow := func(row map[string]any, want any, version int) {
		t.Helper()
		cells := row["cells"].(map[string]any)
		cell, present := cells["evidence.owner_user_id"].(map[string]any)
		if !present || cell["value"] != want || row["row_version"] != float64(version) {
			t.Fatalf("owner row=%#v version=%v want=%v/%d", cell, row["row_version"], want, version)
		}
		var sourceOwner, projectionOwner *string
		var storedVersion int
		if err := harness.DB.QueryRow(`SELECT e.owner_user_id::text,p.owner_user_id::text,r.row_version FROM evidence e JOIN evidence_grid_projection p USING(record_id) JOIN records r USING(record_id) WHERE e.record_id=$1`, row["record_id"]).Scan(&sourceOwner, &projectionOwner, &storedVersion); err != nil {
			t.Fatal(err)
		}
		for _, actual := range []*string{sourceOwner, projectionOwner} {
			if (want == nil && actual != nil) || (want != nil && (actual == nil || *actual != want)) {
				t.Fatalf("owner source/projection=%v want=%v", actual, want)
			}
		}
		if storedVersion != version {
			t.Fatalf("stored version=%d want=%d", storedVersion, version)
		}
	}
	for index, value := range []any{owner.ID.String(), nil} {
		before := snapshotEvidenceContractState(t, harness, incidentID)
		httptestx.RequireErrorEnvelope(t, call(http.MethodPost, createURL, map[string]any{"client_txn_id": fmt.Sprintf("owner-only-%d", index), "evidence.owner_user_id": value}), http.StatusBadRequest, "invalid_mutation_payload")
		if snapshotEvidenceContractState(t, harness, incidentID) != before {
			t.Fatal("owner-only create changed state")
		}
	}
	for index, value := range []any{inactive.ID.String(), foreign.ID.String(), uuid.Nil.String(), uuid.NewString(), "not-a-uuid"} {
		before := snapshotEvidenceContractState(t, harness, incidentID)
		httptestx.RequireErrorEnvelope(t, call(http.MethodPost, createURL, map[string]any{"client_txn_id": fmt.Sprintf("invalid-owner-%d", index), "evidence.title": "Rejected", "evidence.owner_user_id": value}), http.StatusBadRequest, "invalid_mutation_payload")
		if snapshotEvidenceContractState(t, harness, incidentID) != before {
			t.Fatal("rejected assignment changed state")
		}
	}
	for _, explicitNull := range []bool{false, true} {
		body := map[string]any{"client_txn_id": fmt.Sprintf("unassigned-%t", explicitNull), "evidence.title": "Unassigned"}
		if explicitNull {
			body["evidence.owner_user_id"] = nil
		}
		created := httptestx.RequireSuccessData(t, call(http.MethodPost, createURL, body), http.StatusCreated)
		checkRow(created["row"].(map[string]any), nil, 1)
	}
	body := map[string]any{"client_txn_id": "assigned-create", "evidence.title": "Collect memory", "evidence.owner_user_id": owner.ID.String()}
	created := httptestx.RequireSuccessData(t, call(http.MethodPost, createURL, body), http.StatusCreated)
	row := created["row"].(map[string]any)
	checkRow(row, owner.ID.String(), 1)
	recordID := row["record_id"].(string)
	recordURL := harness.Server.HTTP.URL + "/api/v1/records/" + recordID
	before := snapshotEvidenceContractState(t, harness, incidentID)
	replay := httptestx.RequireSuccessData(t, call(http.MethodPost, createURL, body), http.StatusOK)
	if !reflect.DeepEqual(created, replay) || snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("assignment create replay changed response or state")
	}
	divergent := maps.Clone(body)
	divergent["evidence.owner_user_id"] = adminID.String()
	httptestx.RequireErrorEnvelope(t, call(http.MethodPost, createURL, divergent), http.StatusConflict, "client_txn_conflict")
	if snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("divergent assignment replay changed state")
	}
	patch := func(version int, txn string, field string, value any) map[string]any {
		return map[string]any{"view_schema_id": schema, "base_row_version": version, "client_txn_id": txn, "changes": []map[string]any{{"field_key": field, "value": value}}}
	}
	clearBody := patch(1, "clear-owner", "evidence.owner_user_id", nil)
	cleared := httptestx.RequireSuccessData(t, call(http.MethodPatch, recordURL, clearBody), http.StatusOK)
	checkRow(cleared["row"].(map[string]any), nil, 2)
	before = snapshotEvidenceContractState(t, harness, incidentID)
	replay = httptestx.RequireSuccessData(t, call(http.MethodPatch, recordURL, clearBody), http.StatusOK)
	if !reflect.DeepEqual(cleared, replay) || snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("clear replay changed response or state")
	}
	for index, value := range []string{inactive.ID.String(), foreign.ID.String()} {
		httptestx.RequireErrorEnvelope(t, call(http.MethodPatch, recordURL, patch(2, fmt.Sprintf("invalid-owner-patch-%d", index), "evidence.owner_user_id", value)), http.StatusBadRequest, "invalid_mutation_payload")
		if snapshotEvidenceContractState(t, harness, incidentID) != before {
			t.Fatal("rejected owner patch changed state")
		}
	}
	// A historical assignment remains restorable after membership changes.
	if _, err := harness.DB.Exec(`DELETE FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, incidentID, owner.ID); err != nil {
		t.Fatal(err)
	}
	rollbackBody := map[string]any{"base_row_version": 2, "client_txn_id": "owner-rollback", "target": map[string]any{"kind": "change_set", "change_set_id": cleared["change_set_id"]}}
	rolledBack := httptestx.RequireSuccessData(t, call(http.MethodPost, recordURL+"/rollback", rollbackBody), http.StatusOK)
	if rolledBack["row_version"] != float64(3) {
		t.Fatalf("rollback version=%#v", rolledBack)
	}
	queryRow := requireEvidenceProjectionRow(t, harness, login, incidentID, uuid.MustParse(recordID))
	checkRow(queryRow, owner.ID.String(), 3)
	before = snapshotEvidenceContractState(t, harness, incidentID)
	replay = httptestx.RequireSuccessData(t, call(http.MethodPost, recordURL+"/rollback", rollbackBody), http.StatusOK)
	if !reflect.DeepEqual(rolledBack, replay) || snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("owner rollback replay changed response or state")
	}
	received := httptestx.RequireSuccessData(t, call(http.MethodPatch, recordURL, patch(3, "owner-received", "evidence.lifecycle_state", "received")), http.StatusOK)
	checkRow(received["row"].(map[string]any), owner.ID.String(), 4)
	// A stale replacement must conflict using the owner field's retained values.
	changed := httptestx.RequireSuccessData(t, call(http.MethodPatch, recordURL, patch(4, "owner-reassigned", "evidence.owner_user_id", adminID.String())), http.StatusOK)
	checkRow(changed["row"].(map[string]any), adminID.String(), 5)
	conflictEnvelope := httptestx.RequireErrorEnvelope(t, call(http.MethodPatch, recordURL, patch(4, "owner-stale", "evidence.owner_user_id", nil)), http.StatusConflict, "same_field_conflict")
	conflict := conflictEnvelope["error"].(map[string]any)["conflict"].(map[string]any)
	if conflict["field_key"] != "evidence.owner_user_id" || conflict["conflict_resolution_class"] != "atomic_replace" || conflict["server_value"] != adminID.String() || conflict["client_value"] != nil {
		t.Fatalf("owner conflict=%#v", conflict)
	}
	resolveURL := recordURL + "/conflicts/" + conflict["conflict_token"].(string) + "/resolve"
	resolveBody := map[string]any{"client_txn_id": "owner-resolve-invalid", "conflict_token": conflict["conflict_token"], "resolution_kind": "use_unsaved", "resolved_value": owner.ID.String()}
	before = snapshotEvidenceContractState(t, harness, incidentID)
	httptestx.RequireErrorEnvelope(t, call(http.MethodPost, resolveURL, resolveBody), http.StatusBadRequest, "invalid_mutation_payload")
	if snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("conflict resolution accepted a former member")
	}
	resolveBody["client_txn_id"], resolveBody["resolved_value"] = "owner-resolve-clear", nil
	resolved := httptestx.RequireSuccessData(t, call(http.MethodPost, resolveURL, resolveBody), http.StatusOK)
	checkRow(resolved["row"].(map[string]any), nil, 6)
	before = snapshotEvidenceContractState(t, harness, incidentID)
	replay = httptestx.RequireSuccessData(t, call(http.MethodPost, resolveURL, resolveBody), http.StatusOK)
	if !reflect.DeepEqual(resolved, replay) || snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("owner conflict replay changed response or state")
	}
	var snapshots int
	if err := harness.DB.QueryRow(`SELECT count(*) FROM record_revisions WHERE record_id=$1 AND after_json->>'snapshot_schema_id'='cartulary.revisions.snapshot.evidence.v2' AND after_json->'source' ? 'owner_user_id'`, recordID).Scan(&snapshots); err != nil || snapshots != 6 {
		t.Fatalf("complete owner snapshots=%d err=%v", snapshots, err)
	}
	// History must publish the actual clear transition, independently of the row response.
	history := httptestx.RequireSuccessData(t, call(http.MethodGet, recordURL+"/history?limit=100", nil), http.StatusOK)
	ownerChanges := []any{}
	for _, raw := range history["items"].([]any) {
		item := raw.(map[string]any)
		if item["change_set_id"] != cleared["change_set_id"] {
			continue
		}
		for _, rawUnit := range item["diff_summary"].(map[string]any)["units"].([]any) {
			for _, rawChange := range rawUnit.(map[string]any)["changes"].([]any) {
				if rawChange.(map[string]any)["field_key"] == "evidence.owner_user_id" {
					ownerChanges = append(ownerChanges, rawChange)
				}
			}
		}
	}
	wantChanges := []any{map[string]any{"field_key": "evidence.owner_user_id", "before": map[string]any{"state": "present", "value": owner.ID.String()}, "after": map[string]any{"state": "null"}}}
	if !reflect.DeepEqual(ownerChanges, wantChanges) {
		t.Fatalf("owner History changes=%#v want=%#v", ownerChanges, wantChanges)
	}
	reassigned := httptestx.RequireSuccessData(t, call(http.MethodPatch, recordURL, patch(6, "owner-before-delete", "evidence.owner_user_id", adminID.String())), http.StatusOK)
	checkRow(reassigned["row"].(map[string]any), adminID.String(), 7)
	deleted := httptestx.RequireSuccessData(t, call(http.MethodDelete, recordURL, map[string]any{"base_row_version": 7, "client_txn_id": "owner-delete"}), http.StatusOK)
	if deleted["deleted"] != true || deleted["row_version"] != float64(8) {
		t.Fatalf("delete=%#v", deleted)
	}
	var retainedOwner string
	if err := harness.DB.QueryRow(`SELECT owner_user_id::text FROM evidence WHERE record_id=$1`, recordID).Scan(&retainedOwner); err != nil || retainedOwner != adminID.String() {
		t.Fatalf("deleted owner=%q err=%v", retainedOwner, err)
	}
	restoreBody := map[string]any{"base_row_version": 8, "client_txn_id": "owner-restore"}
	restored := httptestx.RequireSuccessData(t, call(http.MethodPost, recordURL+"/restore", restoreBody), http.StatusOK)
	if restored["deleted"] != false || restored["row_version"] != float64(9) {
		t.Fatalf("restore=%#v", restored)
	}
	checkRow(requireEvidenceProjectionRow(t, harness, login, incidentID, uuid.MustParse(recordID)), adminID.String(), 9)
	before = snapshotEvidenceContractState(t, harness, incidentID)
	replay = httptestx.RequireSuccessData(t, call(http.MethodPost, recordURL+"/restore", restoreBody), http.StatusOK)
	if !reflect.DeepEqual(restored, replay) || snapshotEvidenceContractState(t, harness, incidentID) != before {
		t.Fatal("restore replay changed owner response or state")
	}

}
