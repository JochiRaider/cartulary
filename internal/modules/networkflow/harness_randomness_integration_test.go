package networkflow_test

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	hc "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func assertNetworkFlowTableEntropyTransactions(t *testing.T) {
	h, actor, incident := startNetworkFlowStoreTest(t, "network-flow-id-retry")
	controls := hc.NewControls()
	mux := http.NewServeMux()
	if err := hc.RegisterNetworkFlowRandomnessRoutes(controls.Randomness)(mux, httpapi.DependencySet{Env: map[string]string{"CARTULARY_ENABLE_TEST_ROUTES": "1", "CARTULARY_TEST_RUNTIME_MARKER": "harness-owned", "CARTULARY_TEST_ROUTE_TOKEN": httptestx.TestRouteToken}}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	store := newTestNetworkFlowStore(t, h.DB, revisionsupport.MustAppender(t), WithTableIDEntropy(controls.TableIDEntropy()))
	first := createTestTable(t, h.DB, store, actor.ID, incident, "first.csv", nil, 1)
	raw, err := hex.DecodeString(strings.TrimPrefix(first.TableID, "nft_"))
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.FromBytes(raw)
	if err != nil {
		t.Fatal(err)
	}
	arm := func(values []string) {
		t.Helper()
		controls.Randomness.Clear()
		httptestx.RequireStatus(t, httptestx.DoJSON(t, http.MethodPost, server.URL+"/api/v1/test/runtime/network-flow-randomness", map[string]any{"stream": hc.NetworkFlowRandomStreamTableID, "value_kind": "uuid", "values": values, "consume_once": true, "exhaustion": "fail_closed"}, httptestx.WithHeader("X-Cartulary-Test-Route-Token", httptestx.TestRouteToken)), http.StatusCreated)
	}
	snapshot := func() string {
		t.Helper()
		var v string
		if err := h.DB.QueryRow(context.Background(), `SELECT jsonb_build_array((SELECT count(*) FROM network_flow_tables),(SELECT count(*) FROM network_flow_rows),(SELECT count(*) FROM deployment_admin_audit_events WHERE event_kind='network_flow_table_created'))::text`).Scan(&v); err != nil {
			t.Fatal(err)
		}
		return v
	}
	t.Run("collision then success in same transaction", func(t *testing.T) {
		next := uuid.NewString()
		arm([]string{id.String(), next})
		second := createTestTable(t, h.DB, store, actor.ID, incident, "second.csv", nil, 2)
		if second.TableID != "nft_"+strings.ReplaceAll(next, "-", "") {
			t.Fatalf("injected UUID bytes lost: %s", second.TableID)
		}
		if state, _ := controls.Randomness.NetworkFlowRandomnessState(hc.NetworkFlowRandomStreamTableID); state.RemainingCount != 0 {
			t.Fatal("collision was not retried")
		}
	})
	t.Run("eight collisions roll back", func(t *testing.T) {
		values := make([]string, 8)
		for i := range values {
			values[i] = id.String()
		}
		arm(values)
		before := snapshot()
		_, err := createTestTableResult(t, h.DB, store, actor.ID, incident, "collision.csv", nil, 3)
		if err == nil || snapshot() != before {
			t.Fatalf("collision bound/rollback: %v", err)
		}
		if state, _ := controls.Randomness.NetworkFlowRandomnessState(hc.NetworkFlowRandomStreamTableID); state.RemainingCount != 0 {
			t.Fatal("did not attempt eight allocations")
		}
		if _, err := createTestTableResult(t, h.DB, store, actor.ID, incident, "exhausted.csv", nil, 4); !errors.Is(err, hc.ErrNetworkFlowRandomnessExhausted) || snapshot() != before {
			t.Fatalf("exhaustion must fail closed: %v", err)
		}
	})
	t.Run("other constraints do not retry", func(t *testing.T) {
		arm([]string{uuid.NewString(), uuid.NewString()})
		before := snapshot()
		_, err := store.CreateTable(context.Background(), CreateTableParams{IncidentID: incident, ActorUserID: actor.ID, ImportSessionID: first.SourceImportSessionID, ImportUnitID: first.SourceImportUnitID, SourceContentSHA256: testSHA1, OriginalFilename: "duplicate-source.csv", SourceFilenameDigest: testSHA2, SourceFilenameDigestKeyID: "nf-test-key", MappingFingerprint: testSHA3, SourceProfileID: SourceProfileCiscoSNANetFlowCSV, ParserProfileID: ParserProfileRFC4180HeaderedCSV, Rows: []FlowRow{testFlowRow(5, "e")}})
		var pgErr *pgconn.PgError
		if !errors.As(err, &pgErr) || pgErr.ConstraintName != "network_flow_tables_source_import_unit_id_key" {
			t.Fatalf("unrelated constraint changed: %v", err)
		}
		if snapshot() != before {
			t.Fatal("constraint failure committed partial effects")
		}
		if state, _ := controls.Randomness.NetworkFlowRandomnessState(hc.NetworkFlowRandomStreamTableID); state.RemainingCount != 1 {
			t.Fatal("unrelated constraint retried")
		}
	})
}
