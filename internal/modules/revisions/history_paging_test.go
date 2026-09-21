package revisions

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Counts rows at the database boundary, independently of application counters.
type historyReadMeasurement struct {
	descriptors, mutations, revisions, dependencies, projected int
}
type historyMeasuredTransactions struct {
	TransactionRunner
	measurement *historyReadMeasurement
}

func (r historyMeasuredTransactions) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := r.TransactionRunner.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	return historyMeasuredTx{Tx: tx, measurement: r.measurement}, nil
}

type historyMeasuredTx struct {
	pgx.Tx
	measurement *historyReadMeasurement
}

func (tx historyMeasuredTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	rows, err := tx.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	var count *int
	switch {
	case sql == historyDescriptorsSQL:
		count = &tx.measurement.descriptors
	case strings.Contains(sql, "FROM unnest") && strings.Contains(sql, "csm.before_value"):
		count = &tx.measurement.mutations
	case strings.Contains(sql, "FROM unnest") && strings.Contains(sql, "rr.before_json"):
		count = &tx.measurement.revisions
	default:
		count = &tx.measurement.dependencies
	}
	return historyMeasuredRows{Rows: rows, count: count}, nil
}

type historyMeasuredRows struct {
	pgx.Rows
	count *int
}

func (rows historyMeasuredRows) Next() bool {
	next := rows.Rows.Next()
	if next {
		*rows.count++
	}
	return next
}

func TestHistoryBoundedProjectionAndCoalescing_Integration(t *testing.T) {
	f := newHistoryRepositoryFixture(t)
	ctx := context.Background()
	measurement := &historyReadMeasurement{}
	catalog := validTargetSemanticsCatalog(t, validProviderContributions())
	tag := catalog.byTargetKind["record_tag"]
	tag.historyProjector = func(facts historycontract.Facts) ([]historycontract.Unit, error) {
		return historycontract.Collection(facts, "tag", historycontract.Fields("record_tag", "text", "tag_name"), []string{"record_tag_id"}, []string{"record_id"})
	}
	catalog.byTargetKind["record_tag"] = tag

	for kind, entry := range catalog.byTargetKind {
		if entry.historyProjector != nil {
			project := entry.historyProjector
			entry.historyProjector = func(facts historycontract.Facts) ([]historycontract.Unit, error) {
				measurement.projected++
				return project(facts)
			}
		}
		for recordType, projection := range entry.rowHistory {
			if recordType == "host" {
				projection.project = func(facts historycontract.Facts) ([]historycontract.Unit, error) {
					return historycontract.Row(facts, historycontract.Fields("host", "text", "display_name"))
				}
			}
			project := projection.project
			projection.project = func(facts historycontract.Facts) ([]historycontract.Unit, error) {
				measurement.projected++
				return project(facts)
			}
			entry.rowHistory[recordType] = projection
		}
		catalog.byTargetKind[kind] = entry
	}
	store := historyStore{transactions: historyMeasuredTransactions{f.database, measurement}, materializer: historyRowMaterializer{catalog: catalog}}
	// Tombstones still read every semantic fact; no reversal dependencies are needed.
	record := f.record
	record.Deleted, record.DeletedAt = true, &f.now
	query := HistoryQuery{RecordID: record.RecordID, Limit: 1}
	observations := make([]map[string]any, 0, 3)
	previous := 0
	for _, retained := range []int{10, 100, 1000} {
		_, err := f.database.Exec(ctx, `INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source,created_at)
   SELECT md5($1::uuid::text || '-growth-' || n::text)::uuid,$1::uuid,$2::uuid,'history-growth',$3::timestamptz - n * interval '1 second'
   FROM generate_series($4::int,$5::int) n`, record.IncidentID, f.actorID, f.now, previous+1, retained)
		if err != nil {
			t.Fatal(err)
		}
		_, err = f.database.Exec(ctx, `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,before_value,after_value,history_record_ids,history_entry_record_ids)
   SELECT md5($1::uuid::text || '-growth-' || n::text)::uuid,1,'host',$2::uuid::text,'field_update',
    jsonb_set($3::jsonb,'{source,display_name}',to_jsonb(repeat('retained snapshot ',1024))),$4::jsonb,ARRAY[$2::uuid],ARRAY[$2::uuid]
   FROM generate_series($5::int,$6::int) n`, record.IncidentID, record.RecordID, f.before, f.after, previous+1, retained)
		if err != nil {
			t.Fatal(err)
		}
		previous = retained
		for _, statement := range []string{"ANALYZE change_sets", "ANALYZE change_set_mutations", "ANALYZE record_revisions"} {
			if _, err := f.database.Exec(ctx, statement); err != nil {
				t.Fatal(err)
			}
		}
		*measurement = historyReadMeasurement{}
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		page, err := store.ListRecordHistory(ctx, record, query)
		runtime.ReadMemStats(&after)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) != 1 || page.Next == nil || measurement.descriptors != 2 || measurement.mutations != 1 || measurement.revisions != 1 || measurement.projected != 2 || measurement.dependencies != 0 {
			t.Fatalf("unbounded page at %d: page=%+v counts=%+v", retained, page, measurement)
		}
		var plan json.RawMessage
		if err := f.database.QueryRow(ctx, "EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "+historyDescriptorsSQL, record.RecordID, record.IncidentID, (*time.Time)(nil), uuid.Nil, 0, 0, int64(0), 2).Scan(&plan); err != nil {
			t.Fatal(err)
		}
		var compactPlan bytes.Buffer
		if err := json.Compact(&compactPlan, plan); err != nil {
			t.Fatal(err)
		}
		observations = append(observations, map[string]any{"retained_events": retained + 1, "selected_items": len(page.Items), "descriptor_rows": measurement.descriptors, "selected_mutation_rows": measurement.mutations, "selected_revision_rows": measurement.revisions, "dependency_rows": measurement.dependencies, "projector_calls": measurement.projected, "total_alloc_bytes": after.TotalAlloc - before.TotalAlloc, "heap_delta_bytes": int64(after.HeapAlloc) - int64(before.HeapAlloc), "plan": plan})
		t.Logf("history-scaling retained=%d selected=%d counts=%+v total_alloc_bytes=%d heap_delta_bytes=%d plan=%s", retained, len(page.Items), *measurement, after.TotalAlloc-before.TotalAlloc, int64(after.HeapAlloc)-int64(before.HeapAlloc), compactPlan.String())
	}
	// A malformed lookahead is not projected; once selected it must fail closed.
	var poisoned uuid.UUID
	if err := f.database.QueryRow(ctx, `SELECT md5($1::uuid::text || '-growth-1')::uuid`, record.IncidentID).Scan(&poisoned); err != nil {
		t.Fatal(err)
	}
	if _, err := f.database.Exec(ctx, `UPDATE change_set_mutations SET after_value='{}' WHERE change_set_id=$1`, poisoned); err != nil {
		t.Fatal(err)
	}
	first, err := store.ListRecordHistory(ctx, record, query)
	if err != nil {
		t.Fatalf("lookahead projected: %v", err)
	}
	if _, err := store.ListRecordHistory(ctx, record, HistoryQuery{RecordID: record.RecordID, Limit: 1, After: first.Next}); err == nil {
		t.Fatal("malformed selected facts accepted")
	}
	if _, err := f.database.Exec(ctx, `UPDATE change_set_mutations SET after_value=$2 WHERE change_set_id=$1`, poisoned, f.after); err != nil {
		t.Fatal(err)
	}

	// A row mutation later in the same change set suppresses supplemental revision
	// projection even when it is beyond the page boundary.
	if _, err := f.database.Exec(ctx, `INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,before_value,after_value,history_record_ids,history_entry_record_ids)
 VALUES($1,2,'host',$2::uuid::text,'field_update',$3,$4,ARRAY[$2::uuid],ARRAY[$2::uuid])`, f.changeSetID, record.RecordID.String(), f.before, f.after); err != nil {
		t.Fatal(err)
	}
	*measurement = historyReadMeasurement{}
	first, err = store.ListRecordHistory(ctx, record, query)
	if err != nil {
		t.Fatal(err)
	}
	if measurement.projected != 1 || measurement.revisions != 0 {
		t.Fatalf("off-page row projection ignored: %+v", measurement)
	}
	second, err := store.ListRecordHistory(ctx, record, HistoryQuery{RecordID: record.RecordID, Limit: 1, After: first.Next})
	if err != nil || len(second.Items) != 1 || second.Items[0].ChangeSetID != f.changeSetID || second.Items[0].HistoryItemRef == first.Items[0].HistoryItemRef {
		t.Fatalf("split change set: %+v %v", second, err)
	}
	all, err := store.ListRecordHistory(ctx, record, HistoryQuery{RecordID: record.RecordID, Limit: 2})
	if err != nil || !reflect.DeepEqual(all.Items, append(first.Items, second.Items...)) {
		t.Fatalf("logical page content changed: %v", err)
	}
	// Legal reversal planning may need all mutations of a selected change set.
	// Account for those separately from page projection and unrelated history.
	*measurement = historyReadMeasurement{}
	tx, err := (historyMeasuredTransactions{f.database, measurement}).BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	repository := rollbackQueryRepository{store: &commandStore{targetSemantics: catalog}}
	_, err = repository.loadChangeSetRollbackPlanTx(ctx, tx, rollbackRecordEnvelope{IncidentID: record.IncidentID, RecordID: record.RecordID, RecordType: record.RecordType, RowVersion: record.RowVersion}, f.changeSetID.String())
	if !errors.Is(err, ErrRollbackPreconditionFailed) {
		t.Fatalf("expected ineligible fixture reversal, got %v", err)
	}
	if measurement.dependencies != 2 {
		t.Fatalf("selected change-set dependency count: %+v", measurement)
	}
	root, err := suiteservices.ResolveResultsRoot(nil)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(root, suiteservices.ResolveRunID(nil), "module.revisions")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	evidence, err := json.MarshalIndent(map[string]any{"schema_id": "cartulary.history_paging_measurements.v1", "observations": observations, "selected_change_set_dependency_rows": measurement.dependencies, "dependency_eligibility": "precondition_failed"}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "history-paging-measurements.json"), append(evidence, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Logf("history-selected-change-set-dependencies retained=%d selected=1 dependency_rows=%d unrelated_history_rows=0", previous, measurement.dependencies)

}
