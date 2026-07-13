import { describe, expect, it } from "vitest";
import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import {
  initialNetworkFlowControllerState,
  networkFlowControllerReducer,
} from "./networkFlowController";

function table(id: string): NetworkFlowTable {
  return {
    network_flow_table_id: id,
    display_name: id,
    table_version: 1,
    table_status: "active",
    source_filename_display: `${id}.csv`,
    mapping_fingerprint: "a".repeat(64),
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_at: "2026-07-13T00:00:00Z",
    updated_at: "2026-07-13T00:00:00Z",
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
