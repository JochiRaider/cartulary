package database_migrations_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"

	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestSchemaBootstrap_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	t.Run("Current Imports source storage and copy preserve uncompressed bytes", func(t *testing.T) {
		db := postgresHarness.MigrationDatabaseThroughT(t, canonicalRepositoryHead(t)).SQL()
		assertCurrentImportSourceStorage(t, db)
	})
	for _, boundary := range []int64{0, canonicalRepositoryHead(t) - 1} {
		t.Run(fmt.Sprintf("from_%d", boundary), func(t *testing.T) {
			var database *pgtest.MigrationDatabase
			if boundary == 0 {
				database = emptyMigrationDatabase(t, postgresHarness)
			} else {
				database = postgresHarness.MigrationDatabaseThroughT(t, boundary)
			}
			assertSchemaBootstrap(t, database.SQL())
		})
	}
}

func assertCurrentImportSourceStorage(t *testing.T, db *sql.DB) {
	t.Helper()
	ctx := context.Background()
	exec := func(query string, args ...any) {
		t.Helper()
		if _, err := db.ExecContext(ctx, query, args...); err != nil {
			t.Fatal(err)
		}
	}
	const actor = "10000000-0000-0000-0000-000000000001"
	const incident = "10000000-0000-0000-0000-000000000002"
	const session = "10000000-0000-0000-0000-000000000003"
	const unit = "10000000-0000-0000-0000-000000000004"
	const copiedUnit = "10000000-0000-0000-0000-000000000005"
	const ref = "impsrc_10000000000000000000000000000001"
	content := bytes.Repeat([]byte("current source\x00\r\n"), 32768)
	digest := fmt.Sprintf("%x", sha256.Sum256(content))
	exec(`INSERT INTO users (id,email,display_name,password_hash) VALUES ($1,'storage@example.test','Storage','unused')`, actor)
	exec(`INSERT INTO incidents (id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id)
VALUES ($1,'STORAGE','storage','Storage','active',$2,$2)`, incident, actor)
	exec(`INSERT INTO import_sessions (import_session_id,incident_id,created_by_user_id,client_txn_id,assistant_profile,
source_file_kind,original_filename,source_content_sha256,source_media_type,source_byte_size,parser_profile_id,parser_version,session_status)
VALUES ($1,$2,$3,'storage','test','csv','storage.csv',$4,'text/csv',$5,'test','1','discovered')`, session, incident, actor, digest, len(content))
	for i, id := range []string{unit, copiedUnit} {
		exec(`INSERT INTO import_units (import_unit_id,import_session_id,unit_status,locator_kind,locator,discovery_sequence)
VALUES ($1,$2,'discovered','csv_file','file',$3)`, id, session, i+1)
	}
	exec(`INSERT INTO import_source_streams (source_stream_ref,import_session_id,import_unit_id,source_content_sha256,
source_media_type,source_byte_size,source_bytes) VALUES ($1,$2,$3,$4,'text/csv',$5,$6)`, ref, session, unit, digest, len(content), content)
	var compressed bool
	if err := db.QueryRowContext(ctx, `SELECT pg_column_compression(source_bytes) IS NOT NULL FROM import_source_streams WHERE source_stream_ref=$1`, ref).Scan(&compressed); err != nil || compressed {
		t.Fatalf("current source storage must be uncompressed: compressed=%v err=%v", compressed, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE import_units SET source_rows_json='[]' WHERE import_unit_id=$1`, unit); err == nil {
		t.Fatal("partially present projection admitted")
	}
	var storage string
	if err := db.QueryRowContext(ctx, `SELECT attstorage FROM pg_attribute WHERE attrelid='public.import_source_streams'::regclass AND attname='source_bytes'`).Scan(&storage); err != nil || storage != "e" {
		t.Fatalf("external storage policy: %q %v", storage, err)
	}
	// The operator-region path copies the bytea through INSERT SELECT. Verify the
	// copied value and the original source both remain uncompressed and identical.
	exec(`INSERT INTO import_source_streams (source_stream_ref,import_session_id,import_unit_id,source_content_sha256,source_media_type,source_byte_size,source_bytes)
SELECT 'impsrc_10000000000000000000000000000002',import_session_id,$2,source_content_sha256,source_media_type,source_byte_size,source_bytes
FROM import_source_streams WHERE source_stream_ref=$1`, ref, copiedUnit)
	rows, err := db.QueryContext(ctx, `SELECT source_bytes,source_content_sha256,source_byte_size,pg_column_compression(source_bytes) IS NOT NULL FROM import_source_streams`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var got []byte
		var gotDigest string
		var size int
		if err := rows.Scan(&got, &gotDigest, &size, &compressed); err != nil {
			t.Fatal(err)
		}
		if compressed || size != len(content) || gotDigest != digest || !bytes.Equal(got, content) || fmt.Sprintf("%x", sha256.Sum256(got)) != digest {
			t.Fatal("source storage or copy changed logical content or retained compression")
		}
		count++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("source count: %d", count)
	}
}

func assertSchemaBootstrap(t *testing.T, db *sql.DB) {
	source := canonicalMigrationSource(t)

	if err := database_migrations.Apply(context.Background(), db, source); err != nil {
		t.Fatalf("apply canonical head: %v", err)
	}
	if err := database_migrations.Apply(context.Background(), db, source); err != nil {
		t.Fatalf("repeat canonical head apply: %v", err)
	}
	assertCount(t, db, `SELECT COUNT(*) FROM schema_migration_lineage WHERE lineage_id = 'cartulary.prod_ddl_rebaseline.v3'`, 1)

	if err := db.PingContext(context.Background()); err != nil {
		t.Fatalf("borrowed database handle was closed by migration apply: %v", err)
	}

	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'pgcrypto'`)
	assertExists(t, db, `SELECT 1 FROM pg_extension WHERE extname = 'citext'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'users'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_bootstrap_state'`)
	assertExists(t, db, `SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'deployment_admin_audit_events'`)
	assertCount(t, db, `SELECT COUNT(*) FROM pg_extension WHERE extname IN ('pgcrypto', 'citext')`, 2)
	assertCount(t, db, `SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('users', 'deployment_bootstrap_state', 'deployment_admin_audit_events')`, 3)
}

func assertExists(t testing.TB, db *sql.DB, query string) {
	t.Helper()

	var marker int
	if err := db.QueryRowContext(context.Background(), query).Scan(&marker); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
}

func assertCount(t testing.TB, db *sql.DB, query string, want int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(context.Background(), query).Scan(&got); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}
