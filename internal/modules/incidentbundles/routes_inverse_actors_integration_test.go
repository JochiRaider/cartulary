package incidentbundles_test

import (
	"context"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func seedPortableDeletedTagHistory(t *testing.T, harness *appsupport.ServerHarness, incident, record string, actor uuid.UUID) (string, string) {
	t.Helper()
	id, created, deleted := uuid.New(), uuid.New(), uuid.New()
	start := time.Date(2026, 5, 25, 17, 2, 0, 0, time.UTC)
	end := start.Add(time.Minute)
	before := map[string]any{"record_tag_id": id.String(), "incident_id": incident, "record_id": record, "tag_name": "Portable deleted tag", "normalized_tag_name": "portable deleted tag", "created_by_user_id": actor.String(), "created_at": start.Format(time.RFC3339Nano), "updated_at": start.Format(time.RFC3339Nano), "deleted_at": nil, "deleted_by_user_id": nil}
	after := map[string]any{"record_tag_id": id.String(), "incident_id": incident, "record_id": record, "tag_name": "Portable deleted tag", "normalized_tag_name": "portable deleted tag", "created_by_user_id": actor.String(), "created_at": start.Format(time.RFC3339Nano), "updated_at": end.Format(time.RFC3339Nano), "deleted_at": end.Format(time.RFC3339Nano), "deleted_by_user_id": actor.String()}
	if _, err := harness.DB.Exec(`INSERT INTO record_tags(record_tag_id,incident_id,record_id,tag_name,normalized_tag_name,created_by_user_id,created_at,updated_at,deleted_at,deleted_by_user_id) VALUES($1,$2,$3,'Portable deleted tag','portable deleted tag',$4,$5,$6,$6,$4)`, id, incident, record, actor, start, end); err != nil {
		t.Fatal(err)
	}
	for _, action := range []struct {
		id            uuid.UUID
		at            time.Time
		operation     string
		before, after map[string]any
	}{{created, start, "create", nil, before}, {deleted, end, "delete", before, after}} {
		if _, err := harness.DB.Exec(`INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source,created_at) VALUES($1,$2,$3,'fixture.portable.tag',$4)`, action.id, incident, actor, action.at); err != nil {
			t.Fatal(err)
		}
		if _, err := harness.DB.Exec(`INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,before_value,after_value,history_record_ids,history_entry_record_ids) VALUES($1,1,'record_tag',$2,$3,$4,$5,ARRAY[$6::uuid],ARRAY[$6::uuid])`, action.id, "record_tag:"+record+":"+id.String(), action.operation, jsonRaw(t, action.before), jsonRaw(t, action.after), record); err != nil {
			t.Fatal(err)
		}
	}
	return id.String(), deleted.String()
}

func exerciseImportedInverseActors(t *testing.T, harness *appsupport.ServerHarness, login authcookietest.AuthCookies, local, source uuid.UUID, incident, record, deletedTagID, deletedChangeSet, createChangeSet, nativeChangeSet string) {
	t.Helper()
	ctx := context.Background()
	tx, err := harness.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	resolver := incidentbundles.ImportedAttributionResolver()
	for _, test := range []struct {
		name                string
		incident, changeSet uuid.UUID
		actors              []string
		want                map[string]string
		wantError           bool
	}{
		{"imported", uuid.MustParse(incident), uuid.MustParse(deletedChangeSet), []string{source.String()}, map[string]string{source.String(): local.String()}, false},
		{"native", uuid.MustParse(incident), uuid.MustParse(nativeChangeSet), []string{local.String()}, nil, false},
		{"foreign incident", uuid.New(), uuid.MustParse(deletedChangeSet), []string{source.String()}, nil, false},
		{"unknown actor", uuid.MustParse(incident), uuid.MustParse(deletedChangeSet), []string{source.String(), uuid.NewString()}, nil, true},
	} {
		got, err := resolver.ResolveImportedHistoryActorsTx(ctx, tx, test.incident, test.changeSet, test.actors)
		if (err != nil) != test.wantError || !reflect.DeepEqual(got, test.want) {
			t.Fatalf("%s actor scope: %#v %v", test.name, got, err)
		}
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	requireSuccess := func(phase string, response *http.Response) map[string]any {
		t.Helper()
		if response.StatusCode != http.StatusOK {
			t.Fatalf("%s: status=%d body=%#v", phase, response.StatusCode, httptestx.ReadJSONBody(t, response))
		}
		return httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
	}
	retained := stringScalar(t, harness.DB, `SELECT jsonb_agg(to_jsonb(m) ORDER BY change_set_id,sequence_no)::text FROM change_set_mutations m WHERE change_set_id IN ($1,$2)`, deletedChangeSet, createChangeSet)
	performer := flowtest.SeedLocalUserRecord(t, harness.DB, "inverse-reviewer@example.test", "Inverse reviewer", "InversePass1!", false, false, true)
	scenariotest.CreateMembership(t, harness.Server, login, incident, map[string]any{"client_txn_id": "txn-inverse-reviewer-membership", "user_id": performer.ID.String(), "role": "reviewer"})
	login = flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, performer.Email, "InversePass1!", nil)
	other := performer.ID
	if _, err := harness.DB.Exec(`UPDATE record_tags SET created_by_user_id=$2 WHERE record_tag_id=$1`, deletedTagID, other); err != nil {
		t.Fatal(err)
	}
	beforeFailure := importedInverseState(t, harness, incident)
	items := getRecordHistoryItems(t, harness.Server, login, record)
	if item := requireHistoryItemForChangeSet(t, items, deletedChangeSet); item["reversible"] != false {
		t.Fatalf("actor drift remains reversible: %#v", item)
	}
	failed := postRollback(t, harness.Server, login, record, map[string]any{"base_row_version": 3, "client_txn_id": "txn-imported-tag-delete-inverse", "target": map[string]any{"kind": "change_set", "change_set_id": deletedChangeSet}})
	failure := httptestx.RequireErrorEnvelope(t, failed, http.StatusConflict, "rollback_precondition_failed")
	if details := failure["error"].(map[string]any)["details"].(map[string]any); details["reason_code"] != "stale_target" {
		t.Fatalf("actor drift reason: %#v", failure)
	}
	if got := importedInverseState(t, harness, incident); got != beforeFailure {
		t.Fatal("rejected actor drift changed source/history/receipts/publications")
	}
	if _, err := harness.DB.Exec(`UPDATE record_tags SET created_by_user_id=$2 WHERE record_tag_id=$1`, deletedTagID, local); err != nil {
		t.Fatal(err)
	}
	// Reuse the rejected request identity: failure must not have committed a receipt.
	response := postRollback(t, harness.Server, login, record, map[string]any{"base_row_version": 3, "client_txn_id": "txn-imported-tag-delete-inverse", "target": map[string]any{"kind": "change_set", "change_set_id": deletedChangeSet}})
	data := requireSuccess("imported delete inverse", response)
	if data["row_version"] != float64(4) {
		t.Fatalf("imported delete inverse version: %#v", data)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM record_tags WHERE record_tag_id=$1 AND created_by_user_id=$2 AND deleted_at IS NULL AND deleted_by_user_id IS NULL AND tag_name='Portable deleted tag'`, deletedTagID, local) != 1 {
		t.Fatal("imported before image did not restore destination actor and tag")
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id=$1 AND source_table='record_tags' AND source_row_id=$2 AND source_column='created_by_user_id' AND source_actor_id=$3`, incident, deletedTagID, source.String()) != 1 {
		t.Fatal("inverse lost immutable creation attribution")
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id=$1 AND source_table='record_tags' AND source_row_id=$2 AND source_column='deleted_by_user_id'`, incident, deletedTagID) != 0 {
		t.Fatal("inverse retained stale deletion attribution")
	}
	createItem := requireHistoryItemForChangeSet(t, getRecordHistoryItems(t, harness.Server, login, record), createChangeSet)
	response = postRollback(t, harness.Server, login, record, map[string]any{"base_row_version": 4, "client_txn_id": "txn-imported-tag-create-inverse", "target": map[string]any{"kind": "history_entry", "history_entry_ref": createItem["history_entry_ref"]}})
	data = requireSuccess("imported create inverse", response)
	if data["row_version"] != float64(5) {
		t.Fatalf("imported create inverse version: %#v", data)
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM record_tags WHERE incident_id=$1 AND record_id=$2 AND normalized_tag_name='extensionprofile history' AND deleted_at IS NOT NULL AND deleted_by_user_id=$3 AND created_by_user_id=$4`, incident, record, performer.ID, local) != 1 {
		t.Fatal("imported create inverse did not retain creation and attribute deletion locally")
	}
	response = postRollback(t, harness.Server, login, record, map[string]any{"base_row_version": 5, "client_txn_id": "txn-imported-native-inverse", "target": map[string]any{"kind": "change_set", "change_set_id": nativeChangeSet}})
	data = requireSuccess("native inverse after import", response)
	if data["row_version"] != float64(6) || countRows(t, harness.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND created_by_user_id=$2 AND updated_by_user_id=$3`, record, local, performer.ID) != 1 || stringScalar(t, harness.DB, `SELECT display_name FROM hosts WHERE record_id=$1`, record) != "portable host after" {
		t.Fatal("native inverse after import changed actor semantics")
	}
	response = postRollback(t, harness.Server, login, record, map[string]any{"base_row_version": 6, "client_txn_id": "txn-imported-row-restore", "target": map[string]any{"kind": "row_restore", "restore_to_revision_no": 2}})
	data = requireSuccess("imported row restore", response)
	if data["row_version"] != float64(7) || stringScalar(t, harness.DB, `SELECT display_name FROM hosts WHERE record_id=$1`, record) != "portable host after" {
		t.Fatal("imported row restore failed")
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM record_tags WHERE incident_id=$1 AND record_id=$2 AND normalized_tag_name='extensionprofile history' AND deleted_at IS NOT NULL AND deleted_by_user_id=$3 AND created_by_user_id=$4`, incident, record, performer.ID, local) != 1 {
		t.Fatal("row inverses changed the separate tag association")
	}
	if got := stringScalar(t, harness.DB, `SELECT jsonb_agg(to_jsonb(m) ORDER BY change_set_id,sequence_no)::text FROM change_set_mutations m WHERE change_set_id IN ($1,$2)`, deletedChangeSet, createChangeSet); got != retained {
		t.Fatal("inverse rewrote portable retained history")
	}
	if countRows(t, harness.DB, `SELECT count(*) FROM users WHERE id=$1`, source) != 0 {
		t.Fatal("inverse materialized a portable login")
	}
}

type importedInverseSnapshot struct {
	retained string
	intents  [32]byte
}

func importedInverseState(t *testing.T, harness *appsupport.ServerHarness, incident string) importedInverseSnapshot {
	t.Helper()
	retained := stringScalar(t, harness.DB, `SELECT jsonb_build_object(
'records',(SELECT jsonb_agg(to_jsonb(r) ORDER BY record_id) FROM records r WHERE incident_id=$1),
'tags',(SELECT jsonb_agg(to_jsonb(r) ORDER BY record_tag_id) FROM record_tags r WHERE incident_id=$1),
'sets',(SELECT jsonb_agg(to_jsonb(r) ORDER BY change_set_id) FROM change_sets r WHERE incident_id=$1),
'mutations',(SELECT jsonb_agg(to_jsonb(r) ORDER BY change_set_id,sequence_no) FROM change_set_mutations r WHERE change_set_id IN(SELECT change_set_id FROM change_sets WHERE incident_id=$1)),
'revisions',(SELECT jsonb_agg(to_jsonb(r) ORDER BY revision_id) FROM record_revisions r WHERE record_id IN(SELECT record_id FROM records WHERE incident_id=$1)),
'receipts',(SELECT count(*) FROM route_idempotency)
)::text`, incident)
	return importedInverseSnapshot{retained: retained, intents: collaborationsupport.SnapshotIntentContent(t, harness.DB, uuid.MustParse(incident))}
}

func exerciseMutatedImportedHistoryRoundTrip(t *testing.T, runtime *appsupport.Runtime, source *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, record string) {
	t.Helper()
	const retainedQuery = `SELECT jsonb_agg(jsonb_build_object('row_version',row_version,'before',before_json,'after',after_json) ORDER BY row_version)::text FROM record_revisions WHERE record_id=$1`
	retained := stringScalar(t, source.DB, retainedQuery, record)
	bundle := exportBundleBytes(t, source, login, incident, "txn-export-mutated-import")
	destination := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-mutated-import-third")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, destination.Server.HTTP.URL)
	importBundleAndWait(t, destination.Server, admin, bundle, "txn-import-mutated-history")
	if got := stringScalar(t, destination.DB, retainedQuery, record); got != retained {
		t.Fatal("second transfer changed retained history values")
	}
	if countRows(t, destination.DB, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=7`, record) != 1 || stringScalar(t, destination.DB, `SELECT display_name FROM hosts WHERE record_id=$1`, record) != "portable host after" {
		t.Fatal("second transfer changed the mutated current record")
	}
	if countRows(t, destination.DB, `SELECT count(*) FROM record_tags WHERE record_id=$1 AND normalized_tag_name='extensionprofile history' AND deleted_at IS NOT NULL`, record) != 1 {
		t.Fatal("second transfer lost the inverted Tag tombstone")
	}
	// A further transfer with no local mutation adds a common context suffix.
	// It must retain the previous import boundary and both actors' descriptors.
	const pathsQuery = `SELECT jsonb_object_agg(change_set_id,actor_import_contexts)::text FROM change_sets WHERE incident_id=$1`
	beforePaths := stringScalar(t, destination.DB, pathsQuery, incident)
	reexported := exportBundleBytes(t, destination, admin, incident, "txn-export-third-history")
	fourth := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-mutated-import-fourth")
	fourthAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, fourth.Server.HTTP.URL)
	fourthJob := importBundleAndWait(t, fourth.Server, fourthAdmin, reexported, "txn-import-fourth-history")
	if got := stringScalar(t, fourth.DB, retainedQuery, record); got != retained {
		t.Fatal("repeated transfer changed retained snapshots")
	}
	var prior map[string][]map[string]any
	var final map[string][]map[string]any
	if err := json.Unmarshal([]byte(beforePaths), &prior); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal([]byte(stringScalar(t, fourth.DB, pathsQuery, incident)), &final); err != nil {
		t.Fatal(err)
	}
	if len(prior) != len(final) {
		t.Fatal("repeated transfer changed change-set inventory")
	}
	var added map[string]any
	for id, path := range prior {
		next := final[id]
		if len(next) != len(path)+1 || !reflect.DeepEqual(next[:len(path)], path) {
			t.Fatal("repeated transfer changed prior actor contexts")
		}
		step := next[len(path)]
		if added == nil {
			added = step
		} else if !reflect.DeepEqual(added, step) {
			t.Fatal("one import appended different contexts")
		}
	}
	if added == nil {
		t.Fatal("missing imported history context")
	}
	if countRows(t, fourth.DB, `SELECT count(*) FROM users WHERE id::text=$1`, added["actor_user_id"]) != 1 {
		t.Fatal("context did not bind the destination actor")
	}
	if added["import_id"] != fourthJob["job_id"] {
		t.Fatal("context did not bind the import operation")
	}
}
