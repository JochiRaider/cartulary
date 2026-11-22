package database_migrations_test

import (
	"database/sql"
	"strings"
	"testing"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"

	"github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	linkportability "github.com/JochiRaider/cartulary/internal/modules/links/testsupport/portability"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

// Exercise the increment against data that could have been committed at v29.
func TestTimelineTagLabelRepair_Integration(t *testing.T) {
	ctx := t.Context()
	h := pgtest.Start(t)
	database := h.MigrationDatabaseThroughT(t, 29)
	db := database.SQL()
	actor, incident, record := uuid.New(), uuid.New(), uuid.New()
	active, deleted, invalid := uuid.New(), uuid.New(), uuid.New()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	exec(`INSERT INTO users(id,email,display_name,password_hash) VALUES ($1,'repair@example.test','Repair','unused')`, actor)
	exec(`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ($1,'NFC','nfc','Label repair','active',$2,$2)`, incident, actor)
	exec(`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id)
VALUES ($1,$2,'timeline_event',$3,$3)`, record, incident, actor)
	exec(`INSERT INTO timeline_events(record_id,incident_id,capture_state,created_by_user_id,updated_by_user_id)
VALUES ($1,$2,'rough',$3,$3)`, record, incident, actor)
	exec(`INSERT INTO timeline_grid_projection(record_id,incident_id,row_version,recorded_at,edited_at,capture_state)
VALUES ($1,$2,1,now(),now(),'rough')`, record, incident)
	exec(`INSERT INTO record_tags(record_tag_id,record_id,incident_id,created_by_user_id,tag_name,normalized_tag_name,deleted_by_user_id,deleted_at)
VALUES ($1,$4,$5,$6,$7,'café',NULL,NULL), ($2,$4,$5,$6,$7,'café',$6,now()), ($3,$4,$5,$6,' Unrelated ','unrelated',NULL,NULL)`,
		active, deleted, invalid, record, incident, actor, "Cafe\u0301")

	// Retained receipts and mutation envelopes are historical evidence of the
	// original spelling. Verify their complete stored values after both upgrades.
	change := uuid.New()
	exec(`INSERT INTO route_idempotency(route_key,scope_key,client_txn_id,actor_user_id,request_hash,status_code,response_json)
VALUES ('timeline_clipboard_paste',$1,'legacy',$2,decode(repeat('01',32),'hex'),200,jsonb_build_object('display_text',$3::text))`,
		incident.String(), actor, "Cafe\u0301")
	exec(`INSERT INTO change_sets(change_set_id,incident_id,actor_user_id,source) VALUES ($1,$2,$3,'clipboard_paste')`, change, incident, actor)
	exec(`INSERT INTO change_set_mutations(change_set_id,sequence_no,target_kind,target_id,operation_kind,after_value,history_record_ids,history_entry_record_ids)
SELECT $1,1,'record_tag','record_tag:'||record_id::text||':'||record_tag_id::text,'create',to_jsonb(t),ARRAY[record_id],ARRAY[record_id]
FROM record_tags t WHERE record_tag_id=$2`, change, active)

	auditTimelineTags(t, db, 1, 1, 1)
	if err := database.ApplyThrough(ctx, 30); err == nil || !strings.Contains(err.Error(), "invalid bindings outside the NFC repair") {
		t.Fatalf("invalid binding failed to block repair: %v", err)
	}
	assertMigrationHead(t, db, 29)
	assertCount(t, db, `SELECT count(*) FROM record_tags WHERE tag_name IS NOT NFC NORMALIZED`, 2)
	exec(`DELETE FROM record_tags WHERE record_tag_id=$1`, invalid)

	// PostgreSQL NFC alone accepts arbitrarily long combining sequences. The
	// writer inserts a CGJ, which is outside a composition-only repair.
	unsafeLabel := strings.Repeat("\u035b", 31)
	safeLabel, safeKey, ok := fieldnorm.NormalizeTagLabel(unsafeLabel)
	if !ok || !strings.Contains(safeLabel, "\u034f") {
		t.Fatal("invalid stream-safety fixture")
	}
	for _, key := range []string{safeKey, unsafeLabel} {
		exec(`INSERT INTO record_tags(record_tag_id,record_id,incident_id,created_by_user_id,tag_name,normalized_tag_name)
VALUES ($1,$2,$3,$4,$5,$6)`, invalid, record, incident, actor, unsafeLabel, key)
		auditTimelineTags(t, db, 1, 1, 1)
		if err := database.ApplyThrough(ctx, 30); err == nil || !strings.Contains(err.Error(), "non-stream-safe label") {
			t.Fatalf("non-compositional change passed audit: %v", err)
		}
		assertMigrationHead(t, db, 29)
		exec(`DELETE FROM record_tags WHERE record_tag_id=$1`, invalid)
	}
	exec(`INSERT INTO record_tags(record_id,incident_id,created_by_user_id,tag_name,normalized_tag_name,deleted_by_user_id,deleted_at)
VALUES ($1,$2,$3,$4,$5,$3,now())`, record, incident, actor, safeLabel, safeKey)

	// The database's built-in Unicode version predates these letter folds
	// and combining classes. Include a literal stand-in used by the upgrade.
	const raw17 = "\U00016ea0A\u035b\u1acf\u0323"
	label17, key17, ok := fieldnorm.NormalizeTagLabel(raw17)
	if !ok || raw17 == label17 {
		t.Fatal("invalid Unicode 17 normalization fixture")
	}
	exec(`INSERT INTO record_tags(record_id,incident_id,created_by_user_id,tag_name,normalized_tag_name,deleted_by_user_id,deleted_at)
VALUES ($1,$2,$3,$4,$5,$3,now())`, record, incident, actor, raw17, key17)

	auditTimelineTags(t, db, 1, 2, 0)

	// Keep an independent physical copy of the entire pre-repair database.
	// Close idle connections so PostgreSQL can clone a consistent snapshot.
	db.SetMaxIdleConns(0)
	backup := "tag_repair_backup_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	source := migrationScratchDatabaseName(t, db)
	created, err := pgtest.CloneDatabase(ctx, h.AdminDSN(), backup, source)
	if created {
		h.RetireClonedDatabaseT(t, backup)
	}
	if err != nil {
		t.Fatal(err)
	}
	verify := func(db *sql.DB) {
		t.Helper()
		assertCount(t, db, `SELECT count(*) FROM record_tags WHERE tag_name='Café' AND normalized_tag_name='café'`, 2)
		auditTimelineTags(t, db, 0, 0, 0)
		var unicodeCount int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM record_tags WHERE tag_name=$1 AND normalized_tag_name=$2`, label17, key17).Scan(&unicodeCount); err != nil || unicodeCount != 1 {
			t.Fatalf("Unicode 17 repair count=%d, err=%v", unicodeCount, err)
		}
		var matches bool
		err := db.QueryRowContext(ctx, `SELECT tags = jsonb_build_array(jsonb_build_object(
'tag_id',$1::uuid,'item_ref','record_tag:'||record_id::text||':'||$1::text,'item_kind','tag','display_text','Café'))
FROM timeline_grid_projection WHERE record_id=$2 AND row_version=1`, active.String(), record).Scan(&matches)
		if err != nil || !matches {
			t.Fatalf("incorrect live projection after repair: matches=%v err=%v", matches, err)
		}
	}
	restored := openAdminDatabase(t, h.AdminDSN(), backup)
	metadata := func(db *sql.DB) string {
		t.Helper()
		var state string
		if err := db.QueryRowContext(ctx, `SELECT jsonb_build_array(
    (SELECT jsonb_agg(to_jsonb(t)-'tag_name' ORDER BY record_tag_id) FROM record_tags t),
    (SELECT to_jsonb(c) FROM change_sets c WHERE change_set_id=$1),
    (SELECT to_jsonb(m) FROM change_set_mutations m WHERE change_set_id=$1),
    (SELECT to_jsonb(r) FROM route_idempotency r WHERE client_txn_id='legacy')
)::text`, change).Scan(&state); err != nil {
			t.Fatal(err)
		}
		return state
	}
	before := metadata(restored)
	if err := database.ApplyThrough(ctx, 30); err != nil {
		t.Fatal(err)
	}
	verify(db)
	if err := database.ApplyThrough(ctx, 30); err != nil || metadata(db) != before {
		t.Fatalf("repair retry or preservation failed: %v", err)
	}
	if err := database.RollbackThrough(ctx, 29); err == nil {
		t.Fatal("irreversible repair allowed Down")
	}
	assertMigrationHead(t, db, 30)
	assertCount(t, restored, `SELECT count(*) FROM record_tags WHERE tag_name IS NOT NFC NORMALIZED`, 2)
	if err := database_migrations.Apply(ctx, restored, canonicalMigrationSource(t)); err != nil {
		t.Fatal(err)
	}
	verify(restored)
	if metadata(restored) != before {
		t.Fatal("restored backup repair changed tag identity, attribution or deletion state")
	}
	linkportability.RequireTagExportImportRoundTrip(t, openSchemaReadinessPoolForDatabase(t, h.AdminDSN(), backup), incident, actor)
}

func auditTimelineTags(t *testing.T, db *sql.DB, wantActive, wantDeleted, wantInvalid int) {
	t.Helper()
	ctx := t.Context()
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, `SELECT t.tag_name, t.normalized_tag_name, t.deleted_at IS NOT NULL
FROM public.record_tags t JOIN public.records r USING (incident_id, record_id) WHERE r.record_type='timeline_event'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var active, deleted, invalid int
	for rows.Next() {
		var name, key string
		var tombstone bool
		if err := rows.Scan(&name, &key, &tombstone); err != nil {
			t.Fatal(err)
		}
		label, comparison, valid := fieldnorm.NormalizeTagLabel(name)
		switch {
		case !valid || key != comparison || norm.NFC.String(name) != label || strings.Count(name, "\u034f") != strings.Count(label, "\u034f"):
			invalid++
		case name == label:
		case tombstone:
			deleted++
		default:
			active++
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if active != wantActive || deleted != wantDeleted || invalid != wantInvalid {
		t.Fatalf("read-only audit active=%d tombstoned=%d invalid=%d; want %d/%d/%d", active, deleted, invalid, wantActive, wantDeleted, wantInvalid)
	}
}
