import { describe, expect, it } from "vitest";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowRow,
} from "../services/networkFlowContractAdapter";
import {
  acceptedContinuationRequest,
  acceptedInitialRequest,
  compileAcceptedFilters,
  type NetworkFlowAcceptedQuery,
  reconcileNetworkFlowContributors,
  reconcileNetworkFlowDiagnostics,
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

  it("preserves server order while reusing only unchanged rows and dropping removed rows", () => {
    const first = row("nfr_1", "10");
    const second = row("nfr_2", "20");
    const third = row("nfr_3", "30");
    const equalThird = { ...third };
    const changedSecond = row("nfr_2", "21");
    const added = row("nfr_4", "40");

    const reconciled = reconcileNetworkFlowRows(
      [first, second, third],
      [equalThird, changedSecond, added],
    );

    expect(reconciled.map((value) => value.network_flow_row_id)).toEqual([
      "nfr_3",
      "nfr_2",
      "nfr_4",
    ]);
    expect(reconciled[0]).toBe(third);
    expect(reconciled[1]).toBe(changedSecond);
    expect(reconciled[2]).toBe(added);
    expect(reconciled).not.toContain(first);
  });

  it("reconciles rejected diagnostics and contributors at their owner identities", () => {
    const priorDiagnostic = diagnostic("nfd_1", "invalid_ipv4");
    const equalDiagnostic = { ...priorDiagnostic, message_args: {} };
    const changedDiagnostic = diagnostic("nfd_1", "invalid_ipv6");
    expect(
      reconcileNetworkFlowDiagnostics([priorDiagnostic], [equalDiagnostic])[0],
    ).toBe(priorDiagnostic);
    expect(
      reconcileNetworkFlowDiagnostics(
        [priorDiagnostic],
        [changedDiagnostic],
      )[0],
    ).toBe(changedDiagnostic);

    const priorContributor = contributor(row("nfr_1", "10"));
    const equalContributor = {
      row: { ...priorContributor.row },
      row_ref: { ...priorContributor.row_ref },
    };
    const changedContributor = contributor(row("nfr_1", "11"));
    expect(
      reconcileNetworkFlowContributors(
        [priorContributor],
        [equalContributor],
      )[0],
    ).toBe(priorContributor);
    expect(
      reconcileNetworkFlowContributors(
        [priorContributor],
        [changedContributor],
      )[0],
    ).toBe(changedContributor);
  });
});

function row(id: string, bytes: string): NetworkFlowRow {
  return {
    network_flow_row_id: id,
    "network_flow.bytes_count": bytes,
    "network_flow.observation_source_ref": { import_unit_id: "unit-1" },
  } as NetworkFlowRow;
}

function diagnostic(id: string, reasonCode: string): NetworkFlowDiagnostic {
  return {
    diagnostic_id: id,
    error_code: "network_flow_invalid_ip",
    field_key: "network_flow.src_ip",
    message: "Invalid IP.",
    message_args: {},
    message_key: `network_flow.diagnostic.network_flow_invalid_ip.${reasonCode}`,
    reason_code: reasonCode,
    source_column_ordinal: 3,
    source_row_number: 2,
  } as NetworkFlowDiagnostic;
}

function contributor(row: NetworkFlowRow): NetworkFlowContributor {
  return {
    row,
    row_ref: {
      mapping_fingerprint: "a".repeat(64),
      network_flow_row_id: row.network_flow_row_id,
      network_flow_table_id: "nft_1",
      source_row_number: 2,
    },
  };
}
