package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func exerciseContextualNoteSourceLock(t *testing.T) {
	h := appsupport.StartStore(t, "context-note-source-lock")
	actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "note-source-lock@example.test", DisplayName: "Note source lock"})
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "note-source-lock-incident", "IR-NOTE-SOURCE-LOCK", "Note source lock")
	source := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, source, "Context host", "context-host-lock", "", "")
	facade, err := workbookassembly.NewArtifactMutationContribution(h.DB, conflicttest.NewCodec("context-note-source-lock"), revisionsupport.MustAppender(t), revisionsupport.MustConflictFieldResolver(t), appsupport.ArtifactProjectionRows(h.DB), appsupport.LinkProjectionRows(h.DB), collaborationsupport.NewRecordChangedAppender())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 3, 0, 0, 0, time.UTC)
	command := artifacts.ContextualNoteCreateCommand{ActorUserID: actor.ID, SourceRecordID: source, Admission: mustArtifactContextualNoteAdmission(t, "context-note-source-lock", map[string]any{"note.title": "Source race note"}, nil), RequestID: "context-note-source-lock", Now: now}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	state := func() map[string]any {
		t.Helper()
		result := map[string]any{"publication": collaborationsupport.SnapshotIntentContent(t, h.DB, incident.ID)}
		for _, table := range []string{"records", "hosts", "artifacts", "record_links", "record_tags", "artifact_grid_projection", "host_grid_projection", "change_sets", "change_set_mutations", "record_revisions", "record_revision_conflict_facts", "revision_history_orders", "route_idempotency"} {
			var value string
			if err := h.DB.QueryRow(ctx, "SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text),'[]'::jsonb)::text FROM "+table+" t").Scan(&value); err != nil {
				t.Fatal(err)
			}
			result[table] = value
		}
		return result
	}
	before := state()
	holder, err := h.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Rollback(context.Background())
	var holderPID int32
	if err := holder.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&holderPID); err != nil {
		t.Fatal(err)
	}
	if _, err := holder.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, source, now, actor.ID); err != nil {
		t.Fatal(err)
	}
	type outcome struct {
		result artifacts.MutationResult
		err    error
	}
	done := make(chan outcome, 1)
	go func() { result, err := facade.CreateContextualNote(ctx, command); done <- outcome{result, err} }()
	consumed := false
	defer func() {
		cancel()
		if !consumed {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("contextual Note worker did not drain")
			}
		}
	}()
	for {
		var waiting bool
		if err := h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND $1::integer=ANY(pg_blocking_pids(pid)))`, holderPID).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case result := <-done:
			consumed = true
			t.Fatalf("contextual Note did not wait for source deletion: %v", result.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var attempted outcome
	select {
	case attempted = <-done:
		consumed = true
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if !errors.Is(attempted.err, pgx.ErrNoRows) || attempted.result.RecordID != uuid.Nil {
		t.Fatalf("contextual Note accepted a source deleted during admission: result=%#v error=%v", attempted.result, attempted.err)
	}
	// Undo only the barrier's synthetic tombstone before checking all durable effects.
	if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1`, source); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, state()) {
		t.Fatal("rejected contextual Note changed durable state or consumed a receipt")
	}
	result, err := facade.CreateContextualNote(ctx, command)
	if err != nil {
		t.Fatal(err)
	}
	if result.RecordID == uuid.Nil || result.ChangeSetID == nil || result.RowVersion != 1 || result.Outcome == artifacts.MutationOutcomeReplayed {
		t.Fatalf("same-request contextual Note retry=%#v", result)
	}
	var exact bool
	if err := h.DB.QueryRow(ctx, `SELECT a.artifact_type='note' AND a.title='Source race note' AND r.row_version=1 AND r.deleted_at IS NULL AND p.title='Source race note' AND p.row_version=1 FROM artifacts a JOIN records r USING(record_id) JOIN artifact_grid_projection p USING(record_id) WHERE a.record_id=$1`, result.RecordID).Scan(&exact); err != nil || !exact {
		t.Fatalf("created Note source/projection exact=%v error=%v", exact, err)
	}
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1 AND src_record_id=$2 AND dst_record_id=$3 AND link_type='references_artifact' AND field_key IS NULL AND owner_user_id=$4 AND created_by_user_id=$4 AND decided_at=$5 AND created_at=$5 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, incident.ID, source, result.RecordID, actor.ID, now, 1)
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1`, *result.ChangeSetID, 2)
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=1`, result.RecordID, 1)
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=1 AND deleted_at IS NULL`, source, 1)
	if _, err := h.DB.Exec(ctx, `UPDATE records SET deleted_at=$2,deleted_by_user_id=$3 WHERE record_id=$1`, source, now, actor.ID); err != nil {
		t.Fatal(err)
	}
	after := state()
	replay, err := facade.CreateContextualNote(ctx, command)
	if err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.RecordID != result.RecordID || replay.RowVersion != 1 || replay.ChangeSetID == nil || *replay.ChangeSetID != *result.ChangeSetID || !reflect.DeepEqual(after, state()) {
		t.Fatalf("committed Note replay after source deletion=%#v error=%v", replay, err)
	}
}

// Gate the first current-source read after PostgreSQL has acquired its actual
// lock. The wrapper forwards SQL and values unchanged and owns no transaction.
type noteSourceGateDB struct {
	dbport.DB
	source  uuid.UUID
	begun   chan int32
	arrived chan int32
	release <-chan struct{}
}

func (d *noteSourceGateDB) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := d.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	var pid int32
	if err := tx.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&pid); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	d.begun <- pid
	return &noteSourceGateTx{Tx: tx, owner: d, pid: pid}, nil
}

type noteSourceGateTx struct {
	pgx.Tx
	owner *noteSourceGateDB
	pid   int32
	gated bool
}

func (tx *noteSourceGateTx) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	row := tx.Tx.QueryRow(ctx, query, args...)
	if !tx.gated && len(args) == 1 {
		if id, ok := args[0].(uuid.UUID); ok && id == tx.owner.source {
			tx.gated = true
			return noteSourceGateRow{Row: row, ctx: ctx, owner: tx.owner, pid: tx.pid}
		}
	}
	return row
}

type noteSourceGateRow struct {
	pgx.Row
	ctx   context.Context
	owner *noteSourceGateDB
	pid   int32
}

func (row noteSourceGateRow) Scan(dest ...any) error {
	if err := row.Row.Scan(dest...); err != nil {
		return err
	}
	row.owner.arrived <- row.pid
	select {
	case <-row.owner.release:
		return nil
	case <-row.ctx.Done():
		return row.ctx.Err()
	}
}

func exerciseContextualNoteWriterOverlap(t *testing.T) {
	h := appsupport.StartStore(t, "context-note-overlap")
	actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "note-overlap@example.test", DisplayName: "Note overlap"})
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "note-overlap-incident", "IR-NOTE-OVERLAP", "Note overlap")
	source := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, source, "Shared context", "shared-context", "", "")
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	release := make(chan struct{})
	db := &noteSourceGateDB{DB: h.DB, source: source, begun: make(chan int32, 2), arrived: make(chan int32, 2), release: release}
	facade, err := workbookassembly.NewArtifactMutationContribution(db, conflicttest.NewCodec("context-note-overlap"), revisionsupport.MustAppender(t), revisionsupport.MustConflictFieldResolver(t), appsupport.ArtifactProjectionRows(h.DB), appsupport.LinkProjectionRows(h.DB), collaborationsupport.NewRecordChangedAppender())
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC)
	commands := make([]artifacts.ContextualNoteCreateCommand, 2)
	for i := range commands {
		key := fmt.Sprintf("context-note-overlap-%d", i)
		commands[i] = artifacts.ContextualNoteCreateCommand{ActorUserID: actor.ID, SourceRecordID: source, Admission: mustArtifactContextualNoteAdmission(t, key, map[string]any{"note.title": fmt.Sprintf("Concurrent Note %d", i)}, nil), RequestID: key, Now: now}
	}
	type outcome struct {
		index  int
		result artifacts.MutationResult
		err    error
	}
	done := make(chan outcome, 2)
	started, consumed := 0, 0
	defer func() {
		cancel()
		for consumed < started {
			select {
			case <-done:
				consumed++
			case <-time.After(5 * time.Second):
				t.Error("contextual Note writers did not drain")
				return
			}
		}
	}()
	launch := func(i int) int32 {
		started++
		go func() { r, e := facade.CreateContextualNote(ctx, commands[i]); done <- outcome{i, r, e} }()
		select {
		case pid := <-db.begun:
			return pid
		case r := <-done:
			consumed++
			t.Fatalf("writer %d did not start a transaction: %v", i, r.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
		return 0
	}
	first := launch(0)
	select {
	case pid := <-db.arrived:
		if pid != first {
			t.Fatal("first admission used another transaction")
		}
	case r := <-done:
		consumed++
		t.Fatalf("first writer bypassed source admission: %v", r.err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	second := launch(1)
	for {
		var waiting bool
		if err := h.DB.QueryRow(ctx, `SELECT $1::integer=ANY(pg_blocking_pids($2::integer))`, first, second).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting {
			break
		}
		select {
		case pid := <-db.arrived:
			if pid != second {
				t.Fatal("second admission used another transaction")
			}
			waiting = true
		case r := <-done:
			consumed++
			t.Fatalf("second writer bypassed source admission: %v", r.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
		if waiting {
			break
		}
	}
	close(release)
	results := make([]artifacts.MutationResult, 2)
	for consumed < started {
		select {
		case r := <-done:
			consumed++
			if r.err != nil || r.result.Outcome != artifacts.MutationOutcomeCreated || r.result.RecordID == uuid.Nil || r.result.ChangeSetID == nil || r.result.RowVersion != 1 {
				t.Errorf("writer %d result=%#v error=%v", r.index, r.result, r.err)
				continue
			}
			results[r.index] = r.result
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	if t.Failed() {
		return
	}
	if results[0].RecordID == results[1].RecordID || *results[0].ChangeSetID == *results[1].ChangeSetID {
		t.Fatal("distinct creations shared an identity")
	}
	for i, result := range results {
		expectedRow := map[string]any{
			"record_id": result.RecordID.String(), "row_version": 1,
			"cells": map[string]any{
				"note.title":               map[string]any{"value": fmt.Sprintf("Concurrent Note %d", i)},
				"note.body":                map[string]any{"value": nil},
				"note.created_by_user_id":  map[string]any{"value": actor.ID.String()},
				"note.updated_at":          map[string]any{"value": "2026-11-14T09:00:00Z"},
				"note.linked_record_count": map[string]any{"value": 1},
				"note.tags":                map[string]any{"value": map[string]any{"items": []any{}, "kind": "collection_value_v1", "ordered": false}},
			},
		}
		canonicalJSON := func(value any) string {
			t.Helper()
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			return string(data)
		}
		wantRow := canonicalJSON(expectedRow)
		if got := canonicalJSON(result.Row); got != wantRow {
			t.Fatalf("created Note row=%s want=%s", got, wantRow)
		}
		var exact bool
		if err := h.DB.QueryRow(ctx, `SELECT a.artifact_type='note' AND a.title=$2 AND r.row_version=1 AND r.deleted_at IS NULL AND p.title=$2 AND p.row_version=1 FROM artifacts a JOIN records r USING(record_id) JOIN artifact_grid_projection p USING(record_id) WHERE a.record_id=$1`, result.RecordID, fmt.Sprintf("Concurrent Note %d", i)).Scan(&exact); err != nil || !exact {
			t.Fatalf("Note %d source/projection=%t error=%v", i, exact, err)
		}
		requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE incident_id=$1 AND src_record_id=$2 AND dst_record_id=$3 AND link_type='references_artifact' AND field_key IS NULL AND provenance='manual' AND confidence IS NULL AND owner_user_id=$4 AND created_by_user_id=$4 AND decided_at=$5 AND created_at=$5 AND deleted_at IS NULL AND deleted_by_user_id IS NULL`, incident.ID, source, result.RecordID, actor.ID, now, 1)
		requireLinkedNoteCount(t, h, `SELECT count(*) FROM change_set_mutations WHERE change_set_id=$1`, *result.ChangeSetID, 2)
		requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_revisions WHERE record_id=$1 AND row_version=1`, result.RecordID, 1)
		replay, err := facade.CreateContextualNote(ctx, commands[i])
		if err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.RecordID != result.RecordID || replay.ChangeSetID == nil || *replay.ChangeSetID != *result.ChangeSetID || replay.RowVersion != 1 || canonicalJSON(replay.Row) != wantRow {
			t.Fatalf("writer %d replay=%#v error=%v", i, replay, err)
		}
	}
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM records WHERE record_id=$1 AND row_version=1 AND deleted_at IS NULL`, source, 1)
	requireLinkedNoteCount(t, h, `SELECT count(*) FROM record_links WHERE src_record_id=$1`, source, 2)
}
