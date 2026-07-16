import { describe, expect, it } from "vitest";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import {
  initialNetworkFlowControllerState,
  networkFlowControllerReducer,
} from "./networkFlowController";

function table(id: string): NetworkFlowTable {
  return {
    network_flow_table_id: id,
    incident_id: "11111111-1111-4111-8111-111111111111",
    display_name: id,
    table_version: 1,
    table_status: "active",
    source_import_session_id: "import-session-1",
    source_import_unit_id: "import-unit-1",
    source_content_sha256: "b".repeat(64),
    source_filename_display: `${id}.csv`,
    source_filename_digest: "c".repeat(64),
    source_filename_digest_key_id: "filename-key-1",
    mapping_fingerprint: "a".repeat(64),
    source_profile_id: "cisco_sna_netflow_csv_v1",
    parser_profile_id: "rfc4180_headered_csv_v1",
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_by_user_id: "user-1",
    created_at: "2026-07-13T00:00:00Z",
    updated_at: "2026-07-13T00:00:00Z",
    deleted_at: null,
  };
}

describe("networkFlowControllerReducer", () => {
  it("preserves selection on refresh and advances it on deletion", () => {
    const loaded = networkFlowControllerReducer(
      initialNetworkFlowControllerState,
      {
        type: "replace_tables",
        tables: [table("nft_a"), table("nft_b")],
      },
    );
    const selected = networkFlowControllerReducer(loaded, {
      type: "select_table",
      tableId: "nft_b",
    });
    const refreshed = networkFlowControllerReducer(selected, {
      type: "replace_tables",
      tables: [table("nft_a"), table("nft_b")],
    });
    expect(refreshed.activeTableId).toBe("nft_b");
    expect(
      networkFlowControllerReducer(refreshed, {
        type: "remove_table",
        tableId: "nft_b",
      }).activeTableId,
    ).toBe("nft_a");
  });

  it("preserves table identity on rename and selects next then previous on deletion", () => {
    const loaded = networkFlowControllerReducer(
      initialNetworkFlowControllerState,
      {
        type: "replace_tables",
        tables: [table("nft_a"), table("nft_b"), table("nft_c")],
      },
    );
    const selected = networkFlowControllerReducer(loaded, {
      type: "select_table",
      tableId: "nft_b",
    });
    const renamed = networkFlowControllerReducer(selected, {
      type: "replace_table",
      table: { ...table("nft_b"), display_name: "Renamed", table_version: 2 },
    });
    expect(renamed.activeTableId).toBe("nft_b");
    expect(renamed.tables.map((candidate) => candidate.display_name)).toEqual([
      "nft_a",
      "Renamed",
      "nft_c",
    ]);

    const removedMiddle = networkFlowControllerReducer(renamed, {
      type: "remove_table",
      tableId: "nft_b",
    });
    expect(removedMiddle.activeTableId).toBe("nft_c");
    const removedLast = networkFlowControllerReducer(removedMiddle, {
      type: "remove_table",
      tableId: "nft_c",
    });
    expect(removedLast.activeTableId).toBe("nft_a");
  });

  it("clears all table state when authorization is lost", () => {
    const loaded = networkFlowControllerReducer(
      initialNetworkFlowControllerState,
      {
        type: "replace_tables",
        tables: [table("nft_a")],
      },
    );
    expect(
      networkFlowControllerReducer(loaded, { type: "clear_authorization" }),
    ).toEqual(initialNetworkFlowControllerState);
  });
});
