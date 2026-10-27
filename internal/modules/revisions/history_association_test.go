package revisions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestHistoryAssociationLookupAndIndex_Integration(t *testing.T) {
	fixture := newHistoryRepositoryFixture(t)
	ctx := context.Background()
	tx, err := fixture.database.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	other, unrelated := uuid.New(), uuid.New()
	if _, err := tx.Exec(ctx, "UPDATE change_set_mutations SET after_value=jsonb_build_object('record_id',$2::text) WHERE change_set_id=$1", fixture.changeSetID, other); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source,created_at) VALUES($1,$2,$3,'association-fixture',$4)", unrelated, fixture.record.IncidentID, fixture.actorID, fixture.now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,after_value,history_record_ids,history_entry_record_ids) VALUES($1,1,'record_tag','misleading-source','create',jsonb_build_object('record_id',$2::text),ARRAY[$3::uuid],ARRAY[$3::uuid])`, unrelated, fixture.record.RecordID, other); err != nil {
		t.Fatal(err)
	}
	positions, err := (historyQueryRepository{}).SelectDescriptorsTx(ctx, tx, fixture.record, HistoryQuery{RecordID: fixture.record.RecordID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 1 || positions[0].ChangeSetID != fixture.changeSetID || positions[0].Kind != HistoryMutation {
		t.Fatalf("membership followed source JSON instead of associations: %#v", positions)
	}
	rows, err := tx.Query(ctx, `SELECT index_class.relname FROM pg_index i
 JOIN pg_class table_class ON table_class.oid=i.indrelid
 JOIN pg_namespace n ON n.oid=table_class.relnamespace
 JOIN pg_class index_class ON index_class.oid=i.indexrelid
 JOIN pg_am am ON am.oid=index_class.relam
 JOIN pg_attribute a ON a.attrelid=table_class.oid AND a.attnum=ANY(i.indkey)
 WHERE n.nspname='public' AND table_class.relname='change_set_mutations'
 AND a.attname='history_record_ids' AND am.amname='gin'
 AND i.indisvalid AND i.indisready AND i.indislive AND i.indpred IS NULL`)
	if err != nil {
		t.Fatal(err)
	}
	indexes := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		indexes[name] = true
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(indexes) == 0 {
		t.Fatal("no valid ready GIN association index")
	}
	// Small fixtures naturally favor sequential scans. Force the planner to show
	// that the association predicate can use the catalog-discovered index, without a cost,
	// timing threshold, fixed index name or exact plan-tree assertion.
	if _, err := tx.Exec(ctx, "SET LOCAL enable_seqscan = off"); err != nil {
		t.Fatal(err)
	}
	var raw []byte
	if err := tx.QueryRow(ctx, "EXPLAIN (FORMAT JSON) SELECT change_set_id,sequence_no FROM change_set_mutations WHERE history_record_ids @> ARRAY[$1]::uuid[]", fixture.record.RecordID).Scan(&raw); err != nil {
		t.Fatal(err)
	}
	var plan any
	if err := json.Unmarshal(raw, &plan); err != nil {
		t.Fatal(err)
	}
	used := false
	var visit func(any)
	visit = func(v any) {
		switch v := v.(type) {
		case map[string]any:
			if name, ok := v["Index Name"].(string); ok && indexes[name] {
				used = true
			}
			for _, child := range v {
				visit(child)
			}
		case []any:
			for _, child := range v {
				visit(child)
			}
		}
	}
	visit(plan)
	if !used {
		t.Fatalf("history query cannot use association index: %s", raw)
	}
}
