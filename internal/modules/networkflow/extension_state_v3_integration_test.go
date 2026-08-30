package networkflow_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestNetworkFlowExtensionStateV4RejectsV1WithoutRewritingBytes_Integration(t *testing.T) {
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
		t.Fatal("state-4 validation admitted a semantic-query-v1 declaration")
	}
	after := persistedGraphSemanticBytes(t, harness.DB, incidentID)
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("state-4 rejection rewrote authoritative graph bytes: before=%q after=%q", before, after)
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
		t.Fatalf("state-4 validation rejected semantic-query-v2-only state: %v", err)
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
