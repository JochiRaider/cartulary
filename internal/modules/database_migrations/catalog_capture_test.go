package database_migrations_test

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// captureCatalogState preserves semantic endpoint evidence from a harness-owned
// fresh database. It excludes physical OIDs and never reads documentation.
func captureCatalogState(t *testing.T, db *sql.DB) {
	t.Helper()
	queries := map[string]string{
		"engine":      `SELECT current_setting('server_version_num') AS version, current_setting('data_checksums') AS checksums, current_user AS owner`,
		"relations":   `SELECT c.relname AS name, c.relkind, pg_get_userbyid(c.relowner) AS owner, c.relacl, c.reloptions, c.relrowsecurity, c.relforcerowsecurity FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relkind IN ('r','p','v','m','S') AND c.relname NOT LIKE 'goose_db_version%' ORDER BY c.relname`,
		"columns":     `SELECT c.relname AS relation, a.attname AS name, a.attnum AS position, format_type(a.atttypid,a.atttypmod) AS type, a.attnotnull, a.attidentity, a.attgenerated, pg_get_expr(d.adbin,d.adrelid) AS expression, col.collname AS collation, a.attacl FROM pg_attribute a JOIN pg_class c ON c.oid=a.attrelid JOIN pg_namespace n ON n.oid=c.relnamespace LEFT JOIN pg_attrdef d ON d.adrelid=a.attrelid AND d.adnum=a.attnum LEFT JOIN pg_collation col ON col.oid=a.attcollation WHERE n.nspname='public' AND c.relkind IN ('r','p','v','m') AND c.relname<>'goose_db_version' AND a.attnum>0 AND NOT a.attisdropped ORDER BY c.relname,a.attnum`,
		"constraints": `SELECT c.relname AS relation, k.conname AS name, k.contype, pg_get_constraintdef(k.oid,true) AS definition, k.convalidated, k.conenforced, k.condeferrable, k.condeferred FROM pg_constraint k JOIN pg_class c ON c.oid=k.conrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname<>'goose_db_version' ORDER BY c.relname,k.conname`,
		"indexes":     `SELECT c.relname AS relation, i.relname AS name, pg_get_indexdef(x.indexrelid) AS definition, x.indisvalid, x.indisready, x.indisunique, x.indisprimary, x.indnullsnotdistinct FROM pg_index x JOIN pg_class c ON c.oid=x.indrelid JOIN pg_class i ON i.oid=x.indexrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relname<>'goose_db_version' ORDER BY i.relname`,
		"routines":    `SELECT p.proname AS name, pg_get_function_identity_arguments(p.oid) AS arguments, pg_get_functiondef(p.oid) AS definition, pg_get_userbyid(p.proowner) AS owner, p.proacl, p.proconfig, p.prosecdef, p.provolatile FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public' AND NOT EXISTS (SELECT 1 FROM pg_depend d WHERE d.classid='pg_proc'::regclass AND d.objid=p.oid AND d.deptype='e') ORDER BY p.proname`,
		"triggers":    `SELECT c.relname AS relation, g.tgname AS name, pg_get_triggerdef(g.oid,true) AS definition, g.tgenabled FROM pg_trigger g JOIN pg_class c ON c.oid=g.tgrelid JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND NOT g.tgisinternal ORDER BY c.relname,g.tgname`,
		"views":       `SELECT c.relname AS name, pg_get_viewdef(c.oid,true) AS definition FROM pg_class c JOIN pg_namespace n ON n.oid=c.relnamespace WHERE n.nspname='public' AND c.relkind IN ('v','m') ORDER BY c.relname`,
		"sequences":   `SELECT c.relname AS name, format_type(s.seqtypid,NULL) AS type, s.seqstart, s.seqincrement, s.seqmax, s.seqmin, s.seqcache, s.seqcycle, t.relname AS owned_table, a.attname AS owned_column FROM pg_sequence s JOIN pg_class c ON c.oid=s.seqrelid JOIN pg_namespace n ON n.oid=c.relnamespace LEFT JOIN pg_depend d ON d.classid='pg_class'::regclass AND d.objid=c.oid AND d.deptype IN ('a','i') LEFT JOIN pg_class t ON t.oid=d.refobjid LEFT JOIN pg_attribute a ON a.attrelid=t.oid AND a.attnum=d.refobjsubid WHERE n.nspname='public' AND c.relname<>'goose_db_version_id_seq' ORDER BY c.relname`,
		"types":       `SELECT t.typname AS name, pg_get_userbyid(t.typowner) AS owner, t.typacl FROM pg_type t JOIN pg_namespace n ON n.oid=t.typnamespace WHERE n.nspname='public' AND t.typelem=0 ORDER BY t.typname`,
		"defaults":    `SELECT pg_get_userbyid(d.defaclrole) AS owner, n.nspname AS namespace, d.defaclobjtype, d.defaclacl FROM pg_default_acl d LEFT JOIN pg_namespace n ON n.oid=d.defaclnamespace ORDER BY owner,namespace,d.defaclobjtype`,
		"extensions":  `SELECT e.extname, e.extversion, n.nspname, pg_get_userbyid(e.extowner) AS owner FROM pg_extension e JOIN pg_namespace n ON n.oid=e.extnamespace ORDER BY e.extname`,
		"policies":    `SELECT * FROM pg_policies WHERE schemaname='public' ORDER BY tablename,policyname`,
		"roles":       `SELECT rolname, rolsuper, rolinherit, rolcreaterole, rolcreatedb, rolcanlogin, rolreplication, rolbypassrls FROM pg_roles WHERE rolname LIKE 'cartulary_%' ORDER BY rolname`,
	}
	state := make(map[string]json.RawMessage, len(queries)+1)
	for key, query := range queries {
		var value []byte
		if err := db.QueryRowContext(t.Context(), "SELECT COALESCE(jsonb_agg(to_jsonb(q)), '[]'::jsonb) FROM ("+query+") q").Scan(&value); err != nil {
			t.Fatalf("capture catalog %s: %v", key, err)
		}
		state[key] = value
	}
	var relations []struct {
		Name string `json:"name"`
		Kind string `json:"relkind"`
	}
	if err := json.Unmarshal(state["relations"], &relations); err != nil {
		t.Fatal(err)
	}
	seeds := map[string]json.RawMessage{}
	for _, relation := range relations {
		if relation.Kind != "r" {
			continue
		}
		var value []byte
		// Names come only from the controlled authored catalog, quoted by PostgreSQL.
		var qualified string
		if err := db.QueryRowContext(t.Context(), `SELECT format('%I.%I', 'public', $1::text)`, relation.Name).Scan(&qualified); err != nil {
			t.Fatal(err)
		}
		if err := db.QueryRowContext(t.Context(), "SELECT COALESCE(jsonb_agg(row ORDER BY row::text), '[]'::jsonb) FROM (SELECT to_jsonb(t) AS row FROM "+qualified+" t) q").Scan(&value); err != nil {
			t.Fatal(err)
		}
		if string(value) != "[]" {
			seeds[relation.Name] = value
		}
	}
	value, err := json.Marshal(seeds)
	if err != nil {
		t.Fatal(err)
	}
	state["seeds"] = value
	encoded, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	root := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
	runID := os.Getenv("CARTULARY_TEST_RUN_ID")
	if root == "" || runID == "" || filepath.Base(runID) != runID {
		t.Fatal("catalog evidence requires the Make harness run identity")
	}
	path := filepath.Join(root, runID, "postgres-catalog-state.json")
	if err := os.WriteFile(path, append(encoded, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Logf("semantic catalog evidence: %s", path)
}
