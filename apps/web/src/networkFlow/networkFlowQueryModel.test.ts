import { describe, expect, it } from "vitest";
import type { NetworkFlowRow } from "../services/networkFlowContractAdapter";
import {
  acceptedContinuationRequest,
  acceptedInitialRequest,
  compileAcceptedFilters,
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowRows,
  rejectedContinuationRequest,
  rejectedInitialRequest,
} from "./networkFlowQueryModel";

describe("Network Flow query model", () => {
  it("omits limit on initial queries and compiles endpoint, ordered multi-sort, and overlap time filters", () => {
    const query: NetworkFlowAcceptedQuery = {
      filters: [
        {
          field_key: "network_flow.endpoint_ip",
          op: "cidr_contains",
          value: "192.0.2.0/24",
        },
      ],
      sort: [
        { field_key: "network_flow.bytes_count", direction: "desc" },
        { field_key: "network_flow.src_ip", direction: "asc" },
      ],
      timeWindow: {
        startUTC: "2026-07-10T12:00:00Z",
        endUTC: "2026-07-10T13:00:00Z",
      },
    };

    expect(acceptedInitialRequest(query)).toEqual({
      schema_id: "cartulary.network_flow.table_query_request.v1",
      filters: [
        {
          field_key: "network_flow.endpoint_ip",
          op: "cidr_contains",
          value: "192.0.2.0/24",
        },
        {
          field_key: "network_flow.flow_end_utc",
          op: "range",
          value: { gte: "2026-07-10T12:00:00Z", lt: null },
        },
        {
          field_key: "network_flow.flow_start_utc",
          op: "range",
          value: { gte: null, lt: "2026-07-10T13:00:00Z" },
        },
      ],
      sort: [
        { field_key: "network_flow.bytes_count", direction: "desc" },
        { field_key: "network_flow.src_ip", direction: "asc" },
      ],
    });
    expect(acceptedInitialRequest(query)).not.toHaveProperty("limit");
    expect(compileAcceptedFilters(query)).toHaveLength(3);
  });

  it("emits exact continuation bodies without initial query members", () => {
    expect(acceptedContinuationRequest("accepted-cursor")).toEqual({
      schema_id: "cartulary.network_flow.table_query_continuation.v1",
      cursor_token: "accepted-cursor",
    });
    expect(rejectedContinuationRequest("rejected-cursor")).toEqual({
      schema_id: "cartulary.network_flow.rejected_rows_query_continuation.v1",
      cursor_token: "rejected-cursor",
    });
  });

  it("compiles only registered diagnostic query members and omits limit", () => {
    const request = rejectedInitialRequest({
      errorCodes: ["network_flow_invalid_ip"],
      fieldKeys: ["network_flow.src_ip"],
      sourceRowRange: { gte: 2, lte: 40 },
    });
    expect(request).toEqual({
      schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
      error_codes: ["network_flow_invalid_ip"],
      field_keys: ["network_flow.src_ip"],
      source_row_range: { gte: 2, lte: 40 },
    });
    expect(request).not.toHaveProperty("limit");
  });

  it("reuses owner-identified row references only when every contract field is equal", () => {
    const previous = row("nfr_1", "10");
    const equalIncoming = {
      ...previous,
      "network_flow.observation_source_ref": {
        ...previous["network_flow.observation_source_ref"],
      },
    };
    const changedIncoming = row("nfr_1", "11");

    expect(reconcileNetworkFlowRows([previous], [equalIncoming])[0]).toBe(
      previous,
    );
    expect(reconcileNetworkFlowRows([previous], [changedIncoming])[0]).toBe(
      changedIncoming,
    );
  });
});

function row(id: string, bytes: string): NetworkFlowRow {
  return {
    network_flow_row_id: id,
    "network_flow.bytes_count": bytes,
    "network_flow.observation_source_ref": { import_unit_id: "unit-1" },
  } as NetworkFlowRow;
}
