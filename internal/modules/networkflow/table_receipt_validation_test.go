package networkflow

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
)

func assertTableReceiptAdmission(t *testing.T) {
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	id := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	table := tableRecord{TableID: "nft_11111111111141118111111111111111", IncidentID: id, DisplayName: "Flows", TableVersion: 2, TableStatus: "active", CommittedSourceID: id, SourceInterpretationSHA256: strings.Repeat("d", 64), SourceRawCellScalarLimit: 4096, SourceContentSHA256: strings.Repeat("a", 64), SourceFilenameDisplay: "flows.csv", SourceFilenameDigest: strings.Repeat("b", 64), SourceFilenameDigestKeyID: "fixture", MappingFingerprint: strings.Repeat("c", 64), SourceProfileID: sourceProfileCiscoSNANetFlowCSV, ParserProfileID: "rfc4180_headered_csv_v1", RowCountAccepted: 1, CreatedByUserID: id, CreatedAt: now, UpdatedAt: now}
	key := authn.RouteIdempotencyKey{RouteKey: routeKeyTablesPatch, ScopeKey: id.String() + ":" + table.TableID, ActorUserID: id, ClientTxnID: "rename"}
	payload := tableMutationPayload(table)
	if err := validateTableReceipt(key, 200, payload); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(map[string]any){
		"old schema":       func(v map[string]any) { v["schema_id"] = "cartulary.network_flow_table_mutation_result.v1" },
		"extra member":     func(v map[string]any) { v["ignored"] = true },
		"nested extra":     func(v map[string]any) { v["table"].(map[string]any)["ignored"] = true },
		"missing counters": func(v map[string]any) { delete(v["table"].(map[string]any), "counter_semantics") },
		"wrong counters":   func(v map[string]any) { v["table"].(map[string]any)["counter_semantics"] = "conversation_total" },
		"old profile":      func(v map[string]any) { v["table"].(map[string]any)["source_profile_id"] = "cisco_sna_netflow_csv_v1" },
		"wrong scope": func(v map[string]any) {
			v["table"].(map[string]any)["incident_id"] = "22222222-2222-4222-8222-222222222222"
		},
		"null count":            func(v map[string]any) { v["table"].(map[string]any)["row_count_rejected"] = nil },
		"fractional count":      func(v map[string]any) { v["table"].(map[string]any)["row_count_accepted"] = 1.5 },
		"negative count":        func(v map[string]any) { v["table"].(map[string]any)["row_count_rejected"] = -1 },
		"missing deletion time": func(v map[string]any) { v["table"].(map[string]any)["table_status"] = "soft_deleted" },
		"invalid timestamp":     func(v map[string]any) { v["table"].(map[string]any)["updated_at"] = "yesterday" },
	} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(payload)
			var value map[string]any
			json.Unmarshal(raw, &value)
			mutate(value)
			if validateTableReceipt(key, 200, value) == nil {
				t.Fatal("incompatible receipt admitted")
			}
		})
	}
	table.TableStatus = "soft_deleted"
	table.DeletedAt = &now
	key.RouteKey = routeKeyTablesDelete
	if err := validateTableReceipt(key, 200, tableMutationPayload(table)); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(payload)
	for _, invalid := range [][]byte{append(append([]byte{}, raw...), []byte(" {}")...), []byte(`{"schema_id":"a","schema_id":"b"}`)} {
		if _, err := decodeStoredNetworkFlowResponse(invalid); err == nil {
			t.Fatal("ambiguous receipt JSON admitted")
		}
	}
}
