package timeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
)

func TestTimelineProvenanceWriter_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "timeline-provenance-writer")
	incident, foreign, actor, record := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Provenance writer','test-only',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	for i, id := range []uuid.UUID{incident, foreign} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Provenance writer','active',$3,$3)`, id, fmt.Sprintf("TL-PROV-WRITER-%d", i), actor); err != nil {
			t.Fatal(err)
		}
	}
	begin := func(t testing.TB) pgx.Tx {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	seed := func(t testing.TB, tx pgx.Tx, scope uuid.UUID, kind string, event bool) {
		t.Helper()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$4,$4)`, record, scope, kind, actor); err != nil {
			t.Fatal(err)
		}
		if event {
			if _, err := tx.Exec(ctx, `INSERT INTO timeline_events(record_id,incident_id,capture_state,created_by_user_id,updated_by_user_id) VALUES($1,$2,'rough',$3,$3)`, record, scope, actor); err != nil {
				t.Fatal(err)
			}
		}
	}
	clipboard := ClipboardRawImportColumn{SourceKind: "clipboard_paste", PasteClientTxnID: "txn<>&\u2028\u2029界", MappingFingerprint: strings.Repeat("a", 64), SourceRowOrdinal: 1, SourceColumnOrdinal: 2, SourceHeaderText: " raw e\u0301\r\n<>& ", RawValue: "raw\r\n\t<>&\u200d"}
	file := ClipboardRawImportColumn{SourceKind: "file_import", ImportSessionID: "11111111-1111-4111-8111-111111111111", ImportUnitID: "22222222-2222-4222-8222-222222222222", MappingFingerprint: strings.Repeat("a", 64), SourceFileKind: "csv", SourceContentSHA256: strings.Repeat("b", 64), ParserProfileID: "cartulary.import.workbook.v2", ParserVersion: "workbook_import_adapter_v2", LocatorKind: "csv_file", Locator: "file", SourceRectA1: "A1:B2", SourceRowOrdinal: 1, SourceColumnOrdinal: 2, RawValue: "raw\r\n\t<>&\u200d", CellKind: "string"}
	for _, tc := range []struct {
		name    string
		column  ClipboardRawImportColumn
		literal string
	}{
		{"clipboard", clipboard, `{"mapping_fingerprint":"` + strings.Repeat("a", 64) + `","paste_client_txn_id":"txn\u003c\u003e\u0026\u2028\u2029界","source_kind":"clipboard_paste"}`},
		{"file", file, `{"import_session_id":"11111111-1111-4111-8111-111111111111","import_unit_id":"22222222-2222-4222-8222-222222222222","locator":"file","locator_kind":"csv_file","mapping_fingerprint":"` + strings.Repeat("a", 64) + `","parser_profile_id":"cartulary.import.workbook.v2","parser_version":"workbook_import_adapter_v2","source_content_sha256":"` + strings.Repeat("b", 64) + `","source_file_kind":"csv","source_kind":"file_import","source_rect_a1":"A1:B2"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, incident, "timeline_event", true)
			additions := []ClipboardRawImportColumn{tc.column}
			before := append([]ClipboardRawImportColumn(nil), additions...)
			if err := insertSourceProvenanceTx(ctx, tx, record, additions); err != nil {
				t.Fatal(err)
			}
			var id, digest, kind, raw string
			var row, column int
			var metadataJSON, headerJSON []byte
			var cell *string
			var created, expectedCreated time.Time
			if err := tx.QueryRow(ctx, `SELECT record_id::text,encode(source_identity_hash,'hex'),source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,cell_kind,created_at,transaction_timestamp() FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&id, &digest, &row, &column, &kind, &metadataJSON, &headerJSON, &raw, &cell, &created, &expectedCreated); err != nil {
				t.Fatal(err)
			}
			hash := sha256.Sum256([]byte(tc.literal))
			var expectedMetadata, metadata, header any
			for _, pair := range []struct {
				raw []byte
				out *any
			}{{[]byte(tc.literal), &expectedMetadata}, {metadataJSON, &metadata}, {headerJSON, &header}} {
				if err := json.Unmarshal(pair.raw, pair.out); err != nil {
					t.Fatal(err)
				}
			}
			if id != record.String() || digest != hex.EncodeToString(hash[:]) || row != 1 || column != 2 || kind != tc.column.SourceKind || !reflect.DeepEqual(metadata, expectedMetadata) || !reflect.DeepEqual(header, tc.column.SourceHeaderText) || raw != "raw\r\n\t<>&\u200d" || !created.Equal(expectedCreated) {
				t.Fatal("ordinary writer changed retained provenance")
			}
			if tc.name == "clipboard" {
				if cell != nil {
					t.Fatal("empty internal cell kind must retain null")
				}
			} else if cell == nil || *cell != "string" {
				t.Fatal("file cell kind changed")
			}
			if err := insertSourceProvenanceTx(ctx, tx, record, additions); err != nil {
				t.Fatal("exact replay", err)
			}
			var count int
			var replayCreated time.Time
			if err := tx.QueryRow(ctx, `SELECT count(*),min(created_at) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count, &replayCreated); err != nil || count != 1 || !replayCreated.Equal(created) {
				t.Fatalf("replay changed identity/count/clock: %d %v", count, err)
			}
			for _, field := range []string{"raw", "header", "cell"} {
				changed := tc.column
				switch field {
				case "raw":
					changed.RawValue = "different"
				case "header":
					changed.SourceHeaderText = "different"
				case "cell":
					changed.CellKind = "different"
				}
				if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{changed}); !errors.Is(err, ErrProvenanceIdentityCollision) {
					t.Fatalf("%s collision=%v", field, err)
				}
			}
			if !reflect.DeepEqual(additions, before) {
				t.Fatal("writer mutated caller input")
			}
		})
	}
	for name, mutate := range map[string]func(*ClipboardRawImportColumn){
		"zero row": func(v *ClipboardRawImportColumn) { v.SourceRowOrdinal = 0 }, "column overflow": func(v *ClipboardRawImportColumn) { v.SourceColumnOrdinal = 2147483648 },
		"nontext header": func(v *ClipboardRawImportColumn) { v.SourceHeaderText = map[string]any{"text": "value"} }, "invalid Unicode": func(v *ClipboardRawImportColumn) { v.RawValue = string([]byte{255}) },
		"NUL raw": func(v *ClipboardRawImportColumn) { v.RawValue = "a\x00b" }, "raw bound": func(v *ClipboardRawImportColumn) { v.RawValue = strings.Repeat("a", 65537) },
		"header compact bound": func(v *ClipboardRawImportColumn) { v.SourceHeaderText = strings.Repeat("<", 10923) }, "metadata digest": func(v *ClipboardRawImportColumn) { v.MappingFingerprint = "bad" },
	} {
		t.Run("reject/"+name, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, incident, "timeline_event", true)
			value := clipboard
			mutate(&value)
			if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{value}); !errors.Is(err, ErrInvalidProvenance) {
				t.Fatal("invalid provenance admitted")
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count); err != nil || count != 0 {
				t.Fatalf("invalid value reached storage: count=%d err=%v", count, err)
			}
		})
	}
	for _, extra := range []int{0, 1} {
		t.Run(fmt.Sprintf("JSONB metadata bound/%d", extra), func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, incident, "timeline_event", true)
			prefix, suffix := `{"paste_client_txn_id":"`, `","source_kind":"capture"}`
			value := ClipboardRawImportColumn{SourceKind: "capture", PasteClientTxnID: strings.Repeat("a", 65536-len(prefix)-len(suffix)-3+extra), SourceRowOrdinal: 1, SourceColumnOrdinal: 1, SourceHeaderText: strings.Repeat("a", 65534), RawValue: strings.Repeat("a", 65536)}
			err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{value})
			if extra == 0 {
				if err != nil {
					t.Fatal(err)
				}
				var metadata, header, raw int
				if err := tx.QueryRow(ctx, `SELECT octet_length(source_metadata::text),octet_length(source_header_json::text),octet_length(raw_value) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&metadata, &header, &raw); err != nil || metadata != 65536 || header != 65536 || raw != 65536 {
					t.Fatalf("storage bounds %d/%d/%d err=%v", metadata, header, raw, err)
				}
			} else {
				if !errors.Is(err, ErrInvalidProvenance) {
					t.Fatal("overbound metadata admitted")
				}
				var count int
				if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count); err != nil || count != 0 {
					t.Fatalf("overbound value reached storage: %d %v", count, err)
				}
			}
		})
	}
	// Exercise the private database phase independently: public Apply rejects an
	// occupied event first, while this phase must still enforce its own bindings.
	for _, variant := range []string{"valid", "missing event", "foreign event", "wrong type", "collision"} {
		t.Run("portable phase/"+variant, func(t *testing.T) {
			tx := begin(t)
			scope, kind, event := incident, "timeline_event", true
			switch variant {
			case "missing event":
				event = false
			case "foreign event":
				scope = foreign
			case "wrong type":
				kind = "host"
			}
			seed(t, tx, scope, kind, event)
			hash := sha256.Sum256([]byte(`{"source_kind":"capture"}`))
			row := map[string]any{"record_id": record.String(), "source_identity_sha256": hex.EncodeToString(hash[:]), "source_row_ordinal": 1, "source_column_ordinal": 1, "source_kind": "capture", "source_metadata": map[string]any{"source_kind": "capture"}, "source_header": nil, "raw_value": "raw", "cell_kind": nil, "created_at": "2026-01-02T03:04:05.123456Z"}
			raw, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			members, err := incidentportability.BindMembers(portabilitysupport.MapBundle{timelineBundleProvenancePath: append(raw, '\n')}, []string{timelineBundleProvenancePath})
			if err != nil {
				t.Fatal(err)
			}
			prepared := preparedTimelineImport{members: members, incidentID: incident}
			descriptor := NewIncidentBundleSourcePort().Descriptor()
			if variant == "collision" {
				if _, err := tx.Exec(ctx, `INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,created_at) VALUES($1,$2,1,1,'capture','{"source_kind":"capture"}','null','existing',now())`, record, hash[:]); err != nil {
					t.Fatal(err)
				}
			}
			err = prepared.applyProvenance(ctx, tx, descriptor)
			if variant == "valid" {
				if err != nil {
					t.Fatal(err)
				}
				if err := prepared.validateProvenance(ctx, tx, descriptor); err != nil {
					t.Fatal(err)
				}
			} else {
				want := "timeline.provenance_non_orphaned"
				if variant == "collision" {
					want = "timeline.provenance_unique"
				}
				var failure *sourceport.Failure
				if !errors.As(err, &failure) || failure.InvariantID() != want {
					t.Fatalf("phase failure=%v want=%s", err, want)
				}
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count); err != nil {
				t.Fatal(err)
			}
			want := 0
			if variant == "valid" || variant == "collision" {
				want = 1
			}
			if count != want {
				t.Fatalf("phase mutated rejected state: count=%d want=%d", count, want)
			}
		})
	}
	t.Run("capacity counts distinct retained identities", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, incident, "timeline_event", true)
		hash := sha256.Sum256([]byte(`{"source_kind":"capture"}`))
		if _, err := tx.Exec(ctx, `INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,created_at) SELECT $1,$2,n,1,'capture','{"source_kind":"capture"}','null','raw','2026-01-02T03:04:05.123456Z'::timestamptz FROM generate_series(1,4095) n`, record, hash[:]); err != nil {
			t.Fatal(err)
		}
		item := func(row int) ClipboardRawImportColumn {
			return ClipboardRawImportColumn{SourceKind: "capture", SourceRowOrdinal: row, SourceColumnOrdinal: 1, RawValue: "raw"}
		}
		state := func() string {
			t.Helper()
			var state string
			if err := tx.QueryRow(ctx, `SELECT jsonb_agg(to_jsonb(p) ORDER BY source_row_ordinal)::text FROM timeline_source_provenance p WHERE record_id=$1`, record).Scan(&state); err != nil {
				t.Fatal(err)
			}
			return state
		}
		before := state()
		bad := item(4097)
		bad.RawValue = "bad\x00value"
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(4096), bad}); !errors.Is(err, ErrInvalidProvenance) || state() != before {
			t.Fatal("later invalid value changed earlier provenance")
		}
		conflict := item(4096)
		conflict.RawValue = "different"
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(4096), conflict}); !errors.Is(err, ErrProvenanceIdentityCollision) || state() != before {
			t.Fatal("conflicting repeated identity changed provenance", err)
		}
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(1), item(4096), item(4096)}); err != nil {
			t.Fatal("4095 plus one distinct identity", err)
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count); err != nil || count != 4096 {
			t.Fatalf("capacity count=%d err=%v", count, err)
		}
		before = state()
		replay := make([]ClipboardRawImportColumn, 4096)
		for i := range replay {
			replay[i] = item(i + 1)
		}
		if err := insertSourceProvenanceTx(ctx, tx, record, replay); err != nil || state() != before {
			t.Fatal("full replay changed or exceeded capacity", err)
		}
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(1), item(1)}); err != nil || state() != before {
			t.Fatal("duplicate replay changed capacity", err)
		}
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(4097), item(1)}); !errors.Is(err, ErrInvalidProvenance) || state() != before {
			t.Fatal("new identity at capacity mutated state", err)
		}
		conflict = item(1)
		conflict.RawValue = "different"
		if err := insertSourceProvenanceTx(ctx, tx, record, []ClipboardRawImportColumn{item(4097), conflict}); !errors.Is(err, ErrProvenanceIdentityCollision) || state() != before {
			t.Fatal("collision at capacity was hidden or mutated state", err)
		}
		if err := insertSourceProvenanceTx(ctx, tx, record, append(replay, item(1))); !errors.Is(err, ErrInvalidProvenance) || state() != before {
			t.Fatal("unbounded append call admitted", err)
		}
	})
	var count int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM records WHERE incident_id IN ($1,$2)`, incident, foreign).Scan(&count); err != nil || count != 0 {
		t.Fatalf("writer savepoint state survived: count=%d err=%v", count, err)
	}
}
