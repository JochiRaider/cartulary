package networkflow_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestNetworkFlowExtensionStateV5Admission_Integration(t *testing.T) {
	t.Run("retained diagnostic admission and restore rollback", func(t *testing.T) {
		harness, actor, incident := startNetworkFlowStoreTest(t, "network-flow-state5-diagnostics")
		store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
		session, unit := seedImportSessionUnit(t, harness.DB, incident, actor.ID, "diagnostics.csv")
		column := int64(5)
		field := "network_flow.src_port"
		header := testSHA2
		digest := fmt.Sprintf("%x", sha256.Sum256([]byte("99999")))
		message := "network_flow.diagnostic.network_flow_invalid_port.out_of_range"
		id := fmt.Sprintf("nfd_%x", sha256.Sum256([]byte(strings.Join([]string{"cartulary.network_flow.diagnostic_id.v1", testSHA1, testSHA3, "3", "p:5", "p:" + field, "network_flow_invalid_port", "out_of_range", ""}, "\x00"))))
		diagnostic := RejectedRowDiagnostic{DiagnosticID: id, SourceRowNumber: 3, SourceColumnOrdinal: &column, RawHeaderSHA256: &header, FieldKey: &field, ErrorCode: "network_flow_invalid_port", ReasonCode: "out_of_range", SafeSample: stringPtr("99999"), RawValueSHA256: &digest, MessageKey: message, Message: message, MessageArgs: json.RawMessage(`{}`)}
		table, err := store.CreateTable(t.Context(), CreateTableParams{IncidentID: incident, ActorUserID: actor.ID, ImportSessionID: session, ImportUnitID: unit, SourceContentSHA256: testSHA1, OriginalFilename: "diagnostics.csv", SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "state5-key", MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(2, "1")}, RowCountRejected: 1, Diagnostics: []RejectedRowDiagnostic{diagnostic}, Now: time.Now().UTC()})
		if err != nil {
			t.Fatal(err)
		}
		if err := ValidateExtensionState(t.Context(), newExtensionStateV3Reader(harness.DB)); err != nil {
			t.Fatalf("valid current state rejected: %v", err)
		}
		if err := ValidateRestoredExtensionState(t.Context(), harness.DB); err == nil {
			t.Fatal("populated restore without metadata admitted")
		}
		if _, err := harness.DB.Exec(t.Context(), `INSERT INTO extension_state_metadata(profile_id,migration_lineage_id,state_version,metadata_version,created_at,updated_at) VALUES ('network_flow_activity','network_flow_activity.state_v1',5,1,now(),now())`); err != nil {
			t.Fatal(err)
		}
		if err := ValidateRestoredExtensionState(t.Context(), harness.DB); err != nil {
			t.Fatalf("valid restored state rejected: %v", err)
		}
		catalog, err := extensionassembly.GeneratedRecoveryCatalog()
		if err != nil {
			t.Fatal(err)
		}
		bindings := catalog.Bindings()
		if len(bindings) == 0 {
			t.Fatal("missing Recovery owner bindings")
		}
		validateRestore := bindings[0].ValidateRestoredState
		if validateRestore == nil || bindings[0].PostRestoreValidationAlgorithm != "network_flow_activity.validate_state_v5" {
			t.Fatal("Recovery validator not bound")
		}
		for key, value := range map[string]any{"message": "private-source", "diagnostic_id": "nfd_" + strings.Repeat("0", 64), "safe_sample": "private-source", "reason_code": "unknown", "raw_value_sha256": nil, "field_key": "network_flow.src_ip"} {
			replacement, _ := json.Marshal(map[string]any{key: value})

			tx, err := harness.DB.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(t.Context(), `WITH original AS (DELETE FROM network_flow_rejected_row_diagnostics WHERE network_flow_table_id = $1 RETURNING *) INSERT INTO network_flow_rejected_row_diagnostics SELECT (jsonb_populate_record(NULL::network_flow_rejected_row_diagnostics, to_jsonb(original) || $2::jsonb)).* FROM original`, table.TableID, string(replacement)); err != nil {
				_ = tx.Rollback(t.Context())
				t.Fatal(err)
			}
			if err := validateRestore(t.Context(), tx); err == nil {
				_ = tx.Rollback(t.Context())
				t.Fatalf("invalid restored diagnostic admitted: %s", key)
			}
			if err := tx.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
		}
		if err := ValidateRestoredExtensionState(t.Context(), harness.DB); err != nil {
			t.Fatal("failed restore modified prior state", err)
		}
	})

	harness, actor, incidentID := startNetworkFlowStoreTest(t, "network-flow-state-v3-mixed-graphs")
	store := newTestNetworkFlowStore(t, harness.DB, revisionsupport.MustAppender(t))
	ctx := context.Background()
	now := time.Date(2026, 8, 16, 18, 0, 0, 0, time.UTC)
	sessionID, unitID := seedImportSessionUnit(t, harness.DB, incidentID, actor.ID, "state-v3.csv")
	table, err := store.CreateTable(ctx, CreateTableParams{
		IncidentID: incidentID, ActorUserID: actor.ID, ImportSessionID: sessionID, ImportUnitID: unitID,
		SourceContentSHA256: testSHA1, OriginalFilename: "state-v3.csv",
		SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "state-v3-key",
		MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV,
		ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(1, "3")}, Now: now,
	})
	if err != nil {
		t.Fatalf("create state-v3 source table: %v", err)
	}
	queries := [][]byte{
		unsupportedDefaultGraphSemanticQuery(table.TableID),
		[]byte(fmt.Sprintf(`{"aggregation":{"include_example_row_refs":true,"mode":"default_flow_edge_v1"},"filters":[],"schema_id":"cartulary.network_flow.graph_semantic_query.v2","selected_table_ids":[%q],"time_range":{"end_utc":null,"start_utc":null}}`, table.TableID)),
	}
	for index, query := range queries {
		declaration := graphViewDeclarationFixture(
			fmt.Sprintf("nfgv_%032x", index+1), incidentID, actor.ID, now.Add(time.Duration(index)*time.Second),
		)
		declaration.DisplayName = fmt.Sprintf("State graph %d", index+1)
		declaration.NormalizedDisplayName = fmt.Sprintf("state graph %d", index+1)
		declaration.SemanticQueryJSON = query
		declaration.SemanticQuerySHA256 = GraphViewSemanticQuerySHA256(query)
		tx, beginErr := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if beginErr != nil {
			t.Fatal(beginErr)
		}
		if insertErr := store.InsertGraphViewDeclarationTx(ctx, tx, declaration); insertErr != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("insert mixed graph declaration: %v", insertErr)
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			t.Fatal(commitErr)
		}
	}
	before := persistedGraphSemanticBytes(t, harness.DB, incidentID)
	reader := newExtensionStateV3Reader(harness.DB)
	if err := ValidateExtensionState(ctx, reader); err == nil {
		t.Fatal("state-5 validation admitted a semantic-query-v1 declaration")
	}
	after := persistedGraphSemanticBytes(t, harness.DB, incidentID)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("state-5 rejection rewrote authoritative graph bytes: before=%q after=%q", before, after)
	}

	current := queries[1]
	if _, err := harness.DB.Exec(ctx, `
UPDATE network_flow_graph_views
   SET semantic_query_json = $2::jsonb,
       semantic_query_sha256 = $3
 WHERE graph_view_id = $1
`, "nfgv_00000000000000000000000000000001", current, GraphViewSemanticQuerySHA256(current)); err != nil {
		t.Fatal(err)
	}
	if err := ValidateExtensionState(ctx, reader); err != nil {
		t.Fatalf("state-5 validation rejected semantic-query-v2-only state: %v", err)
	}
}

type extensionStateV3Reader struct {
	db       postgres.DB
	counters map[string]extensionstore.FamilyCounter
}

func newExtensionStateV3Reader(db postgres.DB) extensionStateV3Reader {
	counters := make(map[string]extensionstore.FamilyCounter)
	for _, counter := range ExtensionStateFamilyCounters() {
		counters[counter.FamilyID] = counter
	}
	return extensionStateV3Reader{db: db, counters: counters}
}

func (reader extensionStateV3Reader) FamilyCounts(ctx context.Context, familyIDs []string) (map[string]int64, error) {
	counts := make(map[string]int64, len(familyIDs))
	for _, familyID := range familyIDs {
		counter, ok := reader.counters[familyID]
		if !ok {
			return nil, fmt.Errorf("missing family counter %s", familyID)
		}
		count, err := counter.Count(ctx, reader.db)
		if err != nil {
			return nil, err
		}
		counts[familyID] = count
	}
	return counts, nil
}

func (reader extensionStateV3Reader) ValidateFamilyState(ctx context.Context, familyID string) error {
	counter, ok := reader.counters[familyID]
	if !ok {
		return fmt.Errorf("missing family validator %s", familyID)
	}
	if counter.Validate == nil {
		return nil
	}
	return counter.Validate(ctx, reader.db)
}

func persistedGraphSemanticBytes(t testing.TB, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, incidentID uuid.UUID) []string {
	t.Helper()
	rows, err := db.Query(context.Background(), `
SELECT semantic_query_json::text, semantic_query_sha256
  FROM network_flow_graph_views
 WHERE incident_id = $1
 ORDER BY graph_view_id
`, incidentID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	values := []string{}
	for rows.Next() {
		var body, digest string
		if err := rows.Scan(&body, &digest); err != nil {
			t.Fatal(err)
		}
		values = append(values, body+"\n"+digest)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return values
}
