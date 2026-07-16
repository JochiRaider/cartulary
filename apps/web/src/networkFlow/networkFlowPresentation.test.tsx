import { describe, expect, it } from "vitest";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowRow,
} from "../services/networkFlowContractAdapter";
import {
  compileNetworkFlowColumns,
  localizedNetworkFlowDiagnosticMessage,
  networkFlowClipboardValue,
  networkFlowContributorsForGrid,
  networkFlowDiagnosticsForGrid,
  networkFlowDisplayValue,
  networkFlowPresentationColumns,
  networkFlowRowsForGrid,
} from "./networkFlowPresentation";

describe("Network Flow presentation metadata", () => {
  it("FE-U-P12-01 Verify Network Analysis presentation, query, layout, lifecycle, and reconciliation units use extension-owned semantic contracts.", () => {
    const metadata = networkFlowPresentationColumns(
      "network_flow.accepted_rows.v1",
    );
    const visible = metadata
      .filter(
        (column) =>
          column.default_visible &&
          !column.inspector_only &&
          column.field_key !== "source_row_number",
      )
      .map((column) => column.field_key);
    const columns = compileNetworkFlowColumns<NetworkFlowRow>({
      gridSchemaId: "network_flow.accepted_rows.v1",
      orderedVisibleFieldKeys: visible,
      widths: {},
    });

    expect(columns.map((column) => column.fieldKey)).toEqual([
      "network_flow.flow_start_utc",
      "network_flow.flow_end_utc",
      "network_flow.src_ip",
      "network_flow.src_port",
      "network_flow.dst_ip",
      "network_flow.dst_port",
      "network_flow.ip_protocol",
      "network_flow.bytes_count",
      "network_flow.packets_count",
      "network_flow.input_interface",
      "network_flow.output_interface",
    ]);
    expect(
      columns.some((column) => column.fieldKey === "network_flow_row_id"),
    ).toBe(false);
  });

  it("keeps display formatting separate from canonical clipboard scalars", () => {
    const row = acceptedRow();
    const metadata = networkFlowPresentationColumns(
      "network_flow.accepted_rows.v1",
    );
    const bytes = requiredColumn(metadata, "network_flow.bytes_count");
    const protocol = requiredColumn(metadata, "network_flow.ip_protocol");
    const port = requiredColumn(metadata, "network_flow.src_port");
    const timestamp = requiredColumn(metadata, "network_flow.flow_start_utc");

    expect(networkFlowDisplayValue(row, bytes)).toBe("1,234,567");
    expect(networkFlowClipboardValue(row, bytes)).toBe("1234567");
    expect(networkFlowDisplayValue(row, protocol)).toBe("TCP (6)");
    expect(networkFlowClipboardValue(row, protocol)).toBe(6);
    expect(networkFlowDisplayValue(row, port)).toBe("—");
    expect(networkFlowClipboardValue(row, port)).toBe("");
    expect(networkFlowDisplayValue(row, timestamp)).toBe(
      "2026-07-16T00:00:00.123456Z",
    );
  });

  it("uses extension resource identities for accepted and diagnostic rows", () => {
    const accepted = networkFlowRowsForGrid([acceptedRow()])[0];
    const rejected = networkFlowDiagnosticsForGrid([diagnostic()])[0];

    expect(accepted?.rowIdentity).toEqual({
      kind: "extension_resource",
      extensionProfileId: "network_flow_activity",
      resourceKind: "network_flow_row",
      resourceId: "nfr_1",
    });
    expect(accepted?.mutationIdentity).toBeUndefined();
    expect(rejected?.rowIdentity).toEqual({
      kind: "extension_resource",
      extensionProfileId: "network_flow_activity",
      resourceKind: "network_flow_rejected_row",
      resourceId: "nfd_1",
    });
  });

  it("localizes registered diagnostics and safely falls back to the server message", () => {
    expect(localizedNetworkFlowDiagnosticMessage(diagnostic())).toBe(
      "The value is not a valid IP address.",
    );
    expect(
      localizedNetworkFlowDiagnosticMessage({
        ...diagnostic(),
        error_code: "network_flow_future_error",
        reason_code: "future_reason",
        message_key:
          "network_flow.diagnostic.network_flow_future_error.future_reason",
        message: "Safe server message.",
      } as NetworkFlowDiagnostic),
    ).toBe("Safe server message.");
    expect(
      localizedNetworkFlowDiagnosticMessage({
        ...diagnostic(),
        message_key:
          "network_flow.diagnostic.network_flow_invalid_port.invalid_syntax",
        message: "Safe mismatched-key fallback.",
      }),
    ).toBe("Safe mismatched-key fallback.");
  });

  it("groups contributors by workspace table order without changing server order within a group", () => {
    const tableOneRows = [
      contributorRow("nft_1", "nfr_1", 1),
      contributorRow("nft_1", "nfr_2", 2),
    ];
    const tableTwoRows = [
      contributorRow("nft_2", "nfr_10", 10),
      contributorRow("nft_2", "nfr_11", 11),
    ];
    const serverContributors = [
      contributor(tableOneRows[0]),
      contributor(tableOneRows[1]),
      contributor(tableTwoRows[0]),
      contributor(tableTwoRows[1]),
    ];

    const dataRows = networkFlowContributorsForGrid(serverContributors);

    expect(dataRows.map((row) => row.data.network_flow_row_id)).toEqual([
      "nfr_1",
      "nfr_2",
      "nfr_10",
      "nfr_11",
    ]);
    expect(dataRows).toHaveLength(serverContributors.length);
    expect(dataRows[0]?.data).toBe(tableOneRows[0]);
    expect(dataRows[1]?.data).toBe(tableOneRows[1]);
    expect(dataRows[2]?.data).toBe(tableTwoRows[0]);
    expect(dataRows[3]?.data).toBe(tableTwoRows[1]);
  });

  it("reuses grid-row wrappers only for unchanged owner objects and derived metadata", () => {
    const first = contributorRow("nft_1", "nfr_1", 1);
    const second = contributorRow("nft_1", "nfr_2", 2);
    const initial = networkFlowRowsForGrid([first, second]);

    const reordered = networkFlowRowsForGrid([second, first], initial);
    expect(reordered[0]).toBe(initial[1]);
    expect(reordered[1]).toBe(initial[0]);

    const changed = { ...first, "network_flow.bytes_count": "2" };
    const replaced = networkFlowRowsForGrid([changed], reordered);
    expect(replaced[0]).not.toBe(initial[0]);
    expect(replaced[0]?.data).toBe(changed);

    const initialContributor = contributor(first);
    const contributorRows = networkFlowContributorsForGrid([
      initialContributor,
    ]);
    const sameContributorRows = networkFlowContributorsForGrid(
      [initialContributor],
      contributorRows,
    );
    expect(sameContributorRows[0]).toBe(contributorRows[0]);

    const changedReference = {
      ...initialContributor,
      row_ref: { ...initialContributor.row_ref, source_row_number: 99 },
    };
    const changedContributorRows = networkFlowContributorsForGrid(
      [changedReference],
      contributorRows,
    );
    expect(changedContributorRows[0]).not.toBe(contributorRows[0]);
    expect(changedContributorRows[0]?.gutterContent).toBe(99);
  });
});

function requiredColumn(
  columns: ReturnType<typeof networkFlowPresentationColumns>,
  fieldKey: string,
) {
  const column = columns.find((candidate) => candidate.field_key === fieldKey);
  if (column === undefined) {
    throw new Error(`missing column ${fieldKey}`);
  }
  return column;
}

function acceptedRow(): NetworkFlowRow {
  return {
    network_flow_row_id: "nfr_1",
    network_flow_table_id: "nft_1",
    source_row_number: 2,
    "network_flow.flow_start_utc": "2026-07-16T00:00:00.123456Z",
    "network_flow.src_port": null,
    "network_flow.ip_protocol": 6,
    "network_flow.bytes_count": "1234567",
  } as NetworkFlowRow;
}

function diagnostic(): NetworkFlowDiagnostic {
  return {
    diagnostic_id: "nfd_1",
    source_row_number: 2,
    source_column_ordinal: 3,
    field_key: "network_flow.src_ip",
    error_code: "network_flow_invalid_ip",
    reason_code: "invalid_ipv4",
    safe_sample: "999.0.2.1",
    raw_header_sha256: null,
    raw_value_sha256: null,
    message_key: "network_flow.diagnostic.network_flow_invalid_ip.invalid_ipv4",
    message_args: {},
    message: "Safe server message.",
    limit_name: null,
    limit_value: null,
    actual_value: null,
  } as NetworkFlowDiagnostic;
}

function contributor(row: NetworkFlowRow | undefined): NetworkFlowContributor {
  if (row === undefined) {
    throw new Error("A contributor row is required.");
  }
  return {
    row,
    row_ref: {
      network_flow_table_id: row.network_flow_table_id,
      network_flow_row_id: row.network_flow_row_id,
      source_row_number: row.source_row_number,
      mapping_fingerprint: "a".repeat(64),
    },
  };
}

function contributorRow(
  tableId: string,
  rowId: string,
  sourceRowNumber: number,
): NetworkFlowRow {
  return {
    network_flow_table_id: tableId,
    network_flow_row_id: rowId,
    source_row_number: sourceRowNumber,
  } as NetworkFlowRow;
}
