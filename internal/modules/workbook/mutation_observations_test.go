package workbook_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// Workbook rejection/replay observations keep source, history, and the public
// query result separate. Hashes avoid retaining source values in diagnostics.
// Publication observations exclude asynchronous delivery bookkeeping.
type workbookState struct {
	Source       []revisions.RecordSnapshot
	History      [32]byte
	Effects      workbookscenariotest.MutationEffects
	Rows         [32]byte
	Preferences  [32]byte
	Associations [32]byte
}

func snapshotWorkbookState(t testing.TB, h *appsupport.ServerHarness, login appsupport.LoginResult, incidentID uuid.UUID, viewSchemaID string) workbookState {
	t.Helper()
	rows, err := h.DB.QueryContext(context.Background(), `SELECT record_id FROM records WHERE incident_id=$1 ORDER BY record_id`, incidentID)
	if err != nil {
		t.Fatal(err)
	}
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatal(err)
	}
	rows.Close()
	var preferences, associations, history string
	if err := h.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'incident',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM incident_workbook_preferences p WHERE incident_id=$1),
 'user',(SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY to_jsonb(p)::text),'[]') FROM user_workbook_preferences p WHERE incident_id=$1)
 )::text`, incidentID).Scan(&preferences); err != nil {
		t.Fatal(err)
	}
	if err := h.DB.QueryRowContext(context.Background(), `SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY to_jsonb(l)::text),'[]')::text FROM record_links l WHERE incident_id=$1`, incidentID).Scan(&associations); err != nil {
		t.Fatal(err)
	}
	if err := h.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'changes',(SELECT coalesce(jsonb_agg(to_jsonb(c) ORDER BY change_set_id),'[]') FROM change_sets c WHERE incident_id=$1),
 'mutations',(SELECT coalesce(jsonb_agg(to_jsonb(m) ORDER BY m.change_set_id,sequence_no),'[]') FROM change_set_mutations m JOIN change_sets c USING(change_set_id) WHERE c.incident_id=$1),
 'revisions',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM record_revisions v JOIN records r USING(record_id) WHERE r.incident_id=$1),
 'history_refs',(SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM record_history_entry_refs v JOIN records r USING(record_id) WHERE r.incident_id=$1)
 )::text`, incidentID).Scan(&history); err != nil {
		t.Fatal(err)
	}
	return workbookState{
		Source:       revisionsupport.SnapshotRecords(t, h.Pool, h.Revisions.Appender(), ids...),
		History:      sha256.Sum256([]byte(history)),
		Effects:      workbookscenariotest.SnapshotIncidentMutationEffects(t, h.DB, incidentID.String()),
		Rows:         workbookContentHash(t, workbookscenariotest.QueryViewRows(t, h.Server.HTTP.URL, incidentID.String(), viewSchemaID, login)),
		Preferences:  sha256.Sum256([]byte(preferences)),
		Associations: sha256.Sum256([]byte(associations)),
	}
}

func workbookContentHash(t testing.TB, value any) [32]byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(data)
}

func requireWorkbookStateUnchanged(t testing.TB, before, after workbookState) {
	t.Helper()
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("workbook state changed: source=%t history=%t history/receipts/publication=%t query=%t preferences=%t associations=%t", !reflect.DeepEqual(before.Source, after.Source), before.History != after.History, before.Effects != after.Effects, before.Rows != after.Rows, before.Preferences != after.Preferences, before.Associations != after.Associations)
	}
}

// The first keep_saved success may add its receipt, but cannot alter history,
// source, query results, associations, preferences, or publication content.
func requireWorkbookKeepSaved(t testing.TB, before, after workbookState) {
	t.Helper()
	if before.History != after.History {
		t.Fatal("keep_saved changed history")
	}
	after.Effects.History = before.Effects.History
	requireWorkbookStateUnchanged(t, before, after)
}
