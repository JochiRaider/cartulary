package networkflow_test

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

// The legacy patterns are the original row CHECK definitions. Every current
// definition must give the same three-valued result for every boundary value,
// so a cheaper predicate can never accept or reject anything the original did
// not.
const (
	legacyHexPattern   = `^[a-f0-9]{64}$`
	legacyRowIDPattern = `^nfr_[a-f0-9]{64}$`
)

func rowCheckBoundaryValues() []struct {
	name  string
	value *string
} {
	hex64 := strings.Repeat("0123456789abcdef", 4)
	text := func(value string) *string { return &value }
	return []struct {
		name  string
		value *string
	}{
		{"valid", text(hex64)},
		{"all_zero", text(strings.Repeat("0", 64))},
		{"all_f", text(strings.Repeat("f", 64))},
		{"63_chars", text(hex64[:63])},
		{"65_chars", text(hex64 + "0")},
		{"128_chars", text(hex64 + hex64)},
		{"empty", text("")},
		{"uppercase", text(hex64[:63] + "A")},
		{"non_hex_letter", text(hex64[:63] + "g")},
		{"trailing_newline", text(hex64 + "\n")},
		{"newline_replacing_last", text(hex64[:63] + "\n")},
		{"leading_newline", text("\n" + hex64[1:])},
		{"inner_space", text(hex64[:31] + " " + hex64[32:])},
		{"multibyte_64_chars", text(hex64[:63] + "é")},
		{"multibyte_63_chars_64_bytes", text(hex64[:62] + "é")},
		{"fullwidth_digit", text(hex64[:63] + "０")},
		{"nul_free_control", text(hex64[:63] + "\t")},
		{"sql_null", nil},
	}
}

func liveRowCheck(t *testing.T, db dbport.DB, name string) string {
	t.Helper()
	var definition string
	if err := db.QueryRow(t.Context(), `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='public.network_flow_rows'::regclass AND conname=$1`, name).Scan(&definition); err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return strings.TrimSuffix(strings.TrimPrefix(definition, "CHECK "), " NOT VALID")
}

func nullableBool(t *testing.T, db dbport.DB, query string, args ...any) sql.NullBool {
	t.Helper()
	var result sql.NullBool
	if err := db.QueryRow(t.Context(), query, args...).Scan(&result); err != nil {
		t.Fatalf("evaluate %q: %v", query, err)
	}
	return result
}

func TestNetworkFlowRowCheckPredicateEquivalence_Integration(t *testing.T) {
	h, actor, incident := startNetworkFlowStoreTest(t, "row-check-equivalence")
	store := newTestNetworkFlowStore(t, h.DB, revisionsupport.MustAppender(t))
	table := createTestTable(t, h.DB, store, actor.ID, incident, "checks.csv", nil, 1)
	valid := strings.Repeat("0123456789abcdef", 4)

	t.Run("sha_ck", func(t *testing.T) {
		live := liveRowCheck(t, h.DB, "network_flow_rows_sha_ck")
		for _, tc := range rowCheckBoundaryValues() {
			for position := 0; position < 3; position++ {
				values := []any{valid, valid, valid}
				values[position] = tc.value
				want := nullableBool(t, h.DB, `SELECT ($1::text ~ '`+legacyHexPattern+`') AND ($2::text ~ '`+legacyHexPattern+`') AND ($3::text ~ '`+legacyHexPattern+`')`, values...)
				got := nullableBool(t, h.DB, `SELECT `+live+` FROM (VALUES ($1::text,$2::text,$3::text)) AS network_flow_rows(source_row_digest_sha256,normalized_row_digest_sha256,mapping_fingerprint)`, values...)
				if got != want {
					t.Errorf("%s at column %d: current=%v legacy=%v", tc.name, position, got, want)
				}
			}
		}
	})
	t.Run("id_ck", func(t *testing.T) {
		live := liveRowCheck(t, h.DB, "network_flow_rows_id_ck")
		cases := rowCheckBoundaryValues()
		for _, prefix := range []string{"nfr_", "NFR_", "nfr-", " nfr_", "xnfr_", "nfr_nfr_", ""} {
			for _, tc := range cases {
				var value *string
				if tc.value != nil {
					candidate := prefix + *tc.value
					value = &candidate
				}
				want := nullableBool(t, h.DB, `SELECT $1::text ~ '`+legacyRowIDPattern+`'`, value)
				got := nullableBool(t, h.DB, `SELECT `+live+` FROM (VALUES ($1::text)) AS network_flow_rows(network_flow_row_id)`, value)
				if got != want {
					t.Errorf("prefix %q %s: current=%v legacy=%v", prefix, tc.name, got, want)
				}
			}
		}
		if prefixOnly := nullableBool(t, h.DB, `SELECT `+live+` FROM (VALUES ('nfr_'::text)) AS network_flow_rows(network_flow_row_id)`); !prefixOnly.Valid || prefixOnly.Bool {
			t.Error("a bare prefix is accepted")
		}
	})
	t.Run("direction_evidence_hashes", func(t *testing.T) {
		for _, member := range []string{"raw_header_sha256", "raw_value_sha256"} {
			for _, input := range []int{0, 3} {
				for _, tc := range rowCheckBoundaryValues() {
					if tc.value == nil {
						continue // SQL NULL cannot be a JSON string; the type check owns it.
					}
					want := nullableBool(t, h.DB, `SELECT $1::text ~ '`+legacyHexPattern+`'`, *tc.value)
					path := fmt.Sprintf("{inputs,%d,%s}", input, member)
					got := nullableBool(t, h.DB, `SELECT public.network_flow_direction_evidence_is_valid_v2(jsonb_set(direction_evidence,$2::text[],to_jsonb($3::text)),direction_kind,from_endpoint,direction_reason,false) FROM network_flow_rows WHERE network_flow_table_id=$1`, table.TableID, path, *tc.value)
					if got != want {
						t.Errorf("%s input %d %s: current=%v legacy=%v", member, input, tc.name, got, want)
					}
				}
			}
		}
	})
	t.Run("inserts_violating_a_rewritten_check_are_rejected", func(t *testing.T) {
		short := valid[:63]
		for _, tc := range []struct {
			name, constraint, assignments, code string
		}{
			{"short_row_id", "network_flow_rows_id_ck", `network_flow_row_id := 'nfr_` + short + `'`, "23514"},
			{"short_source_digest", "network_flow_rows_sha_ck", `source_row_digest_sha256 := '` + short + `'`, "23514"},
			{"long_normalized_digest", "network_flow_rows_sha_ck", `normalized_row_digest_sha256 := '` + valid + `0'`, "23514"},
			{"uppercase_fingerprint", "network_flow_rows_sha_ck", `mapping_fingerprint := '` + strings.ToUpper(valid) + `'`, "23514"},
			{"short_evidence_hash", "network_flow_rows_direction_evidence_ck", `direction_evidence := jsonb_set(direction_evidence,'{inputs,1,raw_value_sha256}',to_jsonb('` + short + `'::text))`, "23514"},
			// The row-to-table key still proves the row's incident.
			{"foreign_incident", "network_flow_rows_table_key", `incident_id := '00000000-0000-4000-8000-000000000999'`, "23503"},
			{"unknown_table", "network_flow_rows_table_key", `network_flow_table_id := 'nft_` + strings.Repeat("e", 32) + `'`, "23503"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				assertRowInsertRejected(t, h.DB, table.TableID, tc.assignments, tc.code, tc.constraint)
			})
		}
	})
	t.Run("direction_evidence_read_once", func(t *testing.T) {
		// The predicate rebinds a non-null object argument to e #> '{}' before its
		// first member test. That must be the identity on every object, and every
		// NULL or non-object argument must still be rejected before it.
		for _, tc := range []struct {
			name, expression string
			want             bool
		}{
			{"stored", `direction_evidence`, true},
			{"reparsed", `direction_evidence::text::jsonb`, true},
			{"sql_null", `NULL::jsonb`, false},
			{"json_null", `'null'::jsonb`, false},
			{"number", `'1'::jsonb`, false},
			{"string", `'"object"'::jsonb`, false},
			{"boolean", `'true'::jsonb`, false},
			{"empty_array", `'[]'::jsonb`, false},
			{"wrapped", `jsonb_build_array(direction_evidence)`, false},
			{"empty_object", `'{}'::jsonb`, false},
			{"extra_member", `direction_evidence || '{"extra":1}'::jsonb`, false},
			{"missing_member", `direction_evidence - 'inputs'`, false},
			{"missing_input", `direction_evidence #- '{inputs,0}'`, false},
		} {
			var valid, identity sql.NullBool
			query := `SELECT public.network_flow_direction_evidence_is_valid_v2(` + tc.expression + `,direction_kind,from_endpoint,direction_reason,endpoint_a_ip = endpoint_b_ip AND endpoint_a_port IS NOT DISTINCT FROM endpoint_b_port),
 CASE WHEN jsonb_typeof(` + tc.expression + `) = 'object' THEN (` + tc.expression + ` #> '{}') IS NOT DISTINCT FROM ` + tc.expression + ` AND (` + tc.expression + ` #> '{}')::text = (` + tc.expression + `)::text END
 FROM network_flow_rows WHERE network_flow_table_id=$1`
			if err := h.DB.QueryRow(t.Context(), query, table.TableID).Scan(&valid, &identity); err != nil {
				t.Fatalf("%s: %v", tc.name, err)
			}
			if !valid.Valid || valid.Bool != tc.want {
				t.Errorf("%s: predicate=%v, want %t", tc.name, valid, tc.want)
			}
			if identity.Valid && !identity.Bool {
				t.Errorf("%s: #> '{}' changed the object", tc.name)
			}
		}
	})
	t.Run("evidence_columns_use_lz4", func(t *testing.T) {
		rows, err := h.DB.Query(t.Context(), `SELECT attname::text, attcompression::text FROM pg_attribute WHERE attrelid='public.network_flow_rows'::regclass AND attname IN ('direction_evidence','observation_source_ref') ORDER BY attname`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		seen := 0
		for rows.Next() {
			var column, method string
			if err := rows.Scan(&column, &method); err != nil {
				t.Fatal(err)
			}
			seen++
			if method != "l" {
				t.Errorf("%s compression = %q, want lz4", column, method)
			}
		}
		if err := rows.Err(); err != nil || seen != 2 {
			t.Fatalf("read column compression: err=%v columns=%d", err, seen)
		}
		// Compression is physical; "stored" and "reparsed" above read the same
		// logical value. A compressed fixture value must use the column method.
		var stored sql.NullString
		if err := h.DB.QueryRow(t.Context(), `SELECT pg_column_compression(direction_evidence) FROM network_flow_rows WHERE network_flow_table_id=$1`, table.TableID).Scan(&stored); err != nil {
			t.Fatal(err)
		}
		if stored.Valid && stored.String != "lz4" {
			t.Errorf("stored evidence compression = %q, want lz4", stored.String)
		}
	})
	other := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "row-check-other@example.test", DisplayName: "Other importer"})
	t.Run("creator_key", func(t *testing.T) {
		// The other user exists, so only the creator column of the composite
		// table key can reject a row whose creator is not its table's creator.
		assertRowInsertRejected(t, h.DB, table.TableID, `created_by_user_id := '`+other.ID.String()+`'::uuid`, "23503", "network_flow_rows_table_key")
		// Nor can the table's creator move away from its rows'.
		assertStatementRejected(t, h.DB, "23503", "network_flow_rows_table_key", `UPDATE network_flow_tables SET created_by_user_id=$2 WHERE network_flow_table_id=$1`, table.TableID, other.ID)
	})
	t.Run("deletion_semantics", func(t *testing.T) {
		// Only the table holds the creator. Rows carry no foreign key: statement
		// triggers tie them to their table and cascade table removal to them.
		assertForeignKeys(t, h.DB, "network_flow_rows", map[string]string{}, true)
		assertForeignKeys(t, h.DB, "network_flow_tables", map[string]string{
			"network_flow_tables_created_by_user_id_fkey": "users(created_by_user_id) delete=a update=a",
			"network_flow_tables_incident_id_fkey":        "incidents(incident_id) delete=c update=a",
		}, false)
		withSavepoint(t, h.DB, "creator_copy", func() {
			// Give the other user a table and row of its own, so that nothing
			// else references that user.
			copyID := "nft_" + strings.Repeat("c", 32)
			sourceID, unitID := uuid.New(), uuid.New()
			for _, statement := range []struct {
				query string
				args  []any
			}{
				{`INSERT INTO import_committed_sources SELECT (jsonb_populate_record(NULL::import_committed_sources, to_jsonb(s) || jsonb_build_object('committed_source_id', $2::uuid, 'origin_import_unit_id', $3::uuid))).* FROM import_committed_sources s JOIN network_flow_tables t USING (incident_id, committed_source_id) WHERE t.network_flow_table_id=$1`, []any{table.TableID, sourceID, unitID}},
				{`INSERT INTO network_flow_tables SELECT (jsonb_populate_record(NULL::network_flow_tables, to_jsonb(t) || jsonb_build_object('network_flow_table_id', $2::text, 'display_name', $2::text, 'committed_source_id', $3::uuid, 'created_by_user_id', $4::uuid))).* FROM network_flow_tables t WHERE network_flow_table_id=$1`, []any{table.TableID, copyID, sourceID, other.ID}},
				{`INSERT INTO network_flow_rows SELECT (jsonb_populate_record(NULL::network_flow_rows, to_jsonb(r) || jsonb_build_object('network_flow_row_id', 'nfr_' || repeat('d', 64), 'network_flow_table_id', $2::text, 'created_by_user_id', $3::uuid))).* FROM network_flow_rows r WHERE network_flow_table_id=$1 ORDER BY source_row_number LIMIT 1`, []any{table.TableID, copyID, other.ID}},
			} {
				if _, err := h.DB.Exec(t.Context(), statement.query, statement.args...); err != nil {
					t.Fatalf("copy fixture: %v", err)
				}
			}
			// While the row exists its creator cannot be removed; the table holds it.
			assertStatementRejected(t, h.DB, "23503", "network_flow_tables_created_by_user_id_fkey", `DELETE FROM users WHERE id=$1`, other.ID)
			// Removing the table removes its row and releases the creator.
			if _, err := h.DB.Exec(t.Context(), `DELETE FROM network_flow_tables WHERE network_flow_table_id=$1`, copyID); err != nil {
				t.Fatal(err)
			}
			var remaining int
			if err := h.DB.QueryRow(t.Context(), `SELECT count(*) FROM network_flow_rows WHERE network_flow_table_id=$1`, copyID).Scan(&remaining); err != nil || remaining != 0 {
				t.Fatalf("rows after table removal = %d, err=%v", remaining, err)
			}
			if _, err := h.DB.Exec(t.Context(), `DELETE FROM users WHERE id=$1`, other.ID); err != nil {
				t.Fatalf("remove released creator: %v", err)
			}
		})
	})
}

// flowRowColumns are network_flow_rows' columns in COPY order.
var flowRowColumns = []string{"network_flow_row_id", "network_flow_table_id", "incident_id", "source_row_number",
	"source_row_digest_sha256", "normalized_row_digest_sha256", "mapping_fingerprint",
	"flow_start_utc", "flow_end_utc", "endpoint_a_ip", "endpoint_b_ip", "endpoint_a_port", "endpoint_b_port", "ip_protocol",
	"bytes_count", "packets_count", "exporter_id", "input_interface", "output_interface",
	"tcp_flags", "application_label", "unmapped_raw", "observation_source_ref", "created_at", "created_by_user_id",
	"direction_kind", "from_endpoint", "direction_reason", "direction_evidence"}

// assertRowInsertRejected copies the table's first row with one assignment
// changed, under a savepoint, and requires the named constraint to reject it.
func assertRowInsertRejected(t *testing.T, db dbport.DB, tableID, assignment, code, constraint string) {
	t.Helper()
	column, expression, _ := strings.Cut(assignment, " := ")
	columns := flowRowColumns
	selected := make([]string, len(columns))
	for i, name := range columns {
		switch name {
		case column:
			selected[i] = expression
		case "network_flow_row_id":
			selected[i] = `'nfr_' || repeat('e', 64)`
		case "source_row_number":
			selected[i] = `source_row_number + 1000`
		default:
			selected[i] = name
		}
	}
	assertStatementRejected(t, db, code, constraint, `INSERT INTO network_flow_rows (`+strings.Join(columns, ",")+`) SELECT `+strings.Join(selected, ",")+` FROM network_flow_rows WHERE network_flow_table_id=$1 ORDER BY source_row_number LIMIT 1`, tableID)
}

// assertStatementRejected runs one statement under a savepoint and requires
// the named constraint to reject it with the given SQLSTATE.
func assertStatementRejected(t *testing.T, db dbport.DB, code, constraint, statement string, args ...any) {
	t.Helper()
	withSavepoint(t, db, "row_check_case", func() {
		_, err := db.Exec(t.Context(), statement, args...)
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != code || pgErr.ConstraintName != constraint {
			t.Fatalf("%s: err=%v, want %s on %s", statement, err, code, constraint)
		}
	})
}

func withSavepoint(t *testing.T, db dbport.DB, name string, body func()) {
	t.Helper()
	if _, err := db.Exec(t.Context(), `SAVEPOINT `+name); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := db.Exec(t.Context(), `ROLLBACK TO SAVEPOINT `+name); err != nil {
			t.Errorf("roll back %s: %v", name, err)
		}
	}()
	body()
}

// assertForeignKeys compares a table's named foreign keys, as
// "parent(child columns) delete=<action> update=<action>", with the expected
// set. When exact is true the table may have no other foreign key.
func assertForeignKeys(t *testing.T, db dbport.DB, table string, want map[string]string, exact bool) {
	t.Helper()
	rows, err := db.Query(t.Context(), `
SELECT k.conname::text,
       (SELECT p.relname::text FROM pg_class p WHERE p.oid=k.confrelid) || '(' || (SELECT string_agg(a.attname::text, ',' ORDER BY c.ordinality) FROM unnest(k.conkey) WITH ORDINALITY c(attnum, ordinality) JOIN pg_attribute a ON a.attrelid=k.conrelid AND a.attnum=c.attnum) || ')'
       || ' delete=' || k.confdeltype::text || ' update=' || k.confupdtype::text
  FROM pg_constraint k WHERE k.contype='f' AND k.conrelid=('public.' || $1)::regclass ORDER BY 1`, table)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var name, definition string
		if err := rows.Scan(&name, &definition); err != nil {
			t.Fatal(err)
		}
		if _, named := want[name]; named || exact {
			got[name] = definition
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("%s foreign keys = %v, want %v", table, got, want)
	}
}

// TestNetworkFlowRowTableKey_Integration proves the statement-level row-to-table
// key: a multi-row statement and COPY are checked as a whole, truncating tables
// alone cannot leave rows, and Recovery's owner predicate finds a row without
// its table. The row check test covers the cascade from table removal.
func TestNetworkFlowRowTableKey_Integration(t *testing.T) {
	h, actor, incident := startNetworkFlowStoreTest(t, "row-table-key")
	store := newTestNetworkFlowStore(t, h.DB, revisionsupport.MustAppender(t))
	table := createTestTable(t, h.DB, store, actor.ID, incident, "key.csv", nil, 1)
	t.Run("statements_are_checked_as_a_whole", func(t *testing.T) {
		// One foreign row beside a valid copy rejects the whole statement.
		assertStatementRejected(t, h.DB, "23503", "network_flow_rows_table_key", `INSERT INTO network_flow_rows SELECT (jsonb_populate_record(NULL::network_flow_rows, to_jsonb(r) || jsonb_build_object('network_flow_row_id', 'nfr_' || repeat(x.c, 64), 'source_row_number', r.source_row_number + x.n, 'network_flow_table_id', x.t))).* FROM network_flow_rows r CROSS JOIN (VALUES ('a', 1000, $1::text), ('b', 2000, 'nft_' || repeat('e', 32))) x(c, n, t) WHERE r.network_flow_table_id=$1`, table.TableID)
		tx, err := h.DB.BeginTx(t.Context(), pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		_, err = tx.CopyFrom(t.Context(), pgx.Identifier{"network_flow_rows"}, flowRowColumns, pgx.CopyFromRows([][]any{foreignRowValues(t, h.DB, table.TableID)}))
		_ = tx.Rollback(t.Context())
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.Code != "23503" || pgErr.ConstraintName != "network_flow_rows_table_key" {
			t.Fatalf("COPY of a row without its table: err=%v", err)
		}
	})
	t.Run("tables_cannot_be_truncated_from_under_rows", func(t *testing.T) {
		// TRUNCATE refuses to run while the fixture's deferred trigger events
		// are pending, so fire them first.
		if _, err := h.DB.Exec(t.Context(), `SET CONSTRAINTS ALL IMMEDIATE`); err != nil {
			t.Fatal(err)
		}
		assertStatementRejected(t, h.DB, "23503", "network_flow_rows_table_key", `TRUNCATE network_flow_tables CASCADE`)
	})
	t.Run("recovery_predicate_finds_a_row_without_its_table", func(t *testing.T) {
		withSavepoint(t, h.DB, "row_key_restore", func() {
			var valid bool
			if err := h.DB.QueryRow(t.Context(), `SELECT public.network_flow_rows_belong_to_tables_v1()`).Scan(&valid); err != nil || !valid {
				t.Fatalf("intact rows: valid=%v err=%v", valid, err)
			}
			// A restore loads with triggers disabled; simulate one orphan.
			if _, err := h.DB.Exec(t.Context(), `SET LOCAL session_replication_role = replica`); err != nil {
				t.Fatal(err)
			}
			if _, err := h.DB.Exec(t.Context(), `DELETE FROM network_flow_rejected_row_diagnostics WHERE network_flow_table_id=$1`, table.TableID); err != nil {
				t.Fatal(err)
			}
			if _, err := h.DB.Exec(t.Context(), `DELETE FROM network_flow_tables WHERE network_flow_table_id=$1`, table.TableID); err != nil {
				t.Fatal(err)
			}
			if err := h.DB.QueryRow(t.Context(), `SELECT public.network_flow_rows_belong_to_tables_v1()`).Scan(&valid); err != nil || valid {
				t.Fatalf("orphaned rows: valid=%v err=%v", valid, err)
			}
		})
	})
}

// foreignRowValues is the table's first row, renamed and pointed at a table
// that does not exist, in COPY column order.
func foreignRowValues(t *testing.T, db dbport.DB, tableID string) []any {
	t.Helper()
	values := make([]any, len(flowRowColumns))
	pointers := make([]any, len(values))
	for i := range values {
		pointers[i] = &values[i]
	}
	if err := db.QueryRow(t.Context(), `SELECT `+strings.Join(flowRowColumns, ",")+` FROM network_flow_rows WHERE network_flow_table_id=$1 ORDER BY source_row_number LIMIT 1`, tableID).Scan(pointers...); err != nil {
		t.Fatal(err)
	}
	values[0], values[1] = "nfr_"+strings.Repeat("f", 64), "nft_"+strings.Repeat("e", 32)
	return values
}
