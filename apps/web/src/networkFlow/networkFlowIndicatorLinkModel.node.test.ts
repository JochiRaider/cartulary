import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import { describe, expect, it } from "vitest";
import type { NetworkFlowRow } from "./networkFlowClient";
import {
  networkFlowRowLinkCandidate,
  resolveNetworkFlowRowLinkSelection,
} from "./networkFlowIndicatorLinkModel";

const tableId = "nft_11111111111111111111111111111111";
const mappingFingerprint = "a".repeat(64);

describe("networkFlowIndicatorLinkModel", () => {
  it("compiles one active IP cell to its immutable owner row identity", () => {
    const rows = [flowRow(1, "192.0.2.10")];
    const activeAnchor = anchor(rows[0], "network_flow.src_ip");
    const selection = resolveNetworkFlowRowLinkSelection({
      activeAnchor,
      bindingSourceRowLimit: 10,
      cellRange: null,
      rows,
    });

    expect(selection).not.toBeNull();
    expect(
      networkFlowRowLinkCandidate(selection as NonNullable<typeof selection>),
    ).toMatchObject({
      candidateValue: "192.0.2.10",
      selector: {
        kind: "row_field_value",
        network_flow_table_id: tableId,
        network_flow_row_id: rows[0]?.network_flow_row_id,
        field_key: "network_flow.src_ip",
      },
    });
  });

  it("compiles a same-value IP range to ordered owner row refs", () => {
    const rows = [
      flowRow(1, "192.0.2.10"),
      flowRow(2, "192.0.2.10"),
      flowRow(3, "192.0.2.10"),
    ];
    const end = anchor(rows[0], "network_flow.src_ip");
    const start = anchor(rows[2], "network_flow.src_ip");
    const selection = resolveNetworkFlowRowLinkSelection({
      activeAnchor: start,
      bindingSourceRowLimit: 3,
      cellRange: { start, end },
      rows,
    });
    const candidate = networkFlowRowLinkCandidate(
      selection as NonNullable<typeof selection>,
    );

    expect(candidate.selector).toEqual({
      kind: "row_refs",
      field_key: "network_flow.src_ip",
      row_refs: rows.map((row) => ({
        network_flow_table_id: tableId,
        network_flow_row_id: row.network_flow_row_id,
        source_row_number: row.source_row_number,
        mapping_fingerprint: mappingFingerprint,
      })),
    });
  });

  it("accepts the active cell anywhere inside its same-field semantic range", () => {
    const rows = [
      flowRow(1, "192.0.2.10"),
      flowRow(2, "192.0.2.10"),
      flowRow(3, "192.0.2.10"),
    ];
    const selection = resolveNetworkFlowRowLinkSelection({
      activeAnchor: anchor(rows[1], "network_flow.src_ip"),
      bindingSourceRowLimit: 3,
      cellRange: {
        start: anchor(rows[0], "network_flow.src_ip"),
        end: anchor(rows[2], "network_flow.src_ip"),
      },
      rows,
    });

    expect(selection?.rows).toEqual(rows);
  });

  it("rejects mixed values, mixed fields, non-owner anchors, and ranges over the owner limit", () => {
    const rows = [flowRow(1, "192.0.2.10"), flowRow(2, "198.51.100.20")];
    const first = anchor(rows[0], "network_flow.src_ip");
    const second = anchor(rows[1], "network_flow.src_ip");
    const mixedField = anchor(rows[1], "network_flow.dst_ip");
    const range = (
      start: GridCellAnchor,
      end: GridCellAnchor,
    ): GridCellRange => ({
      start,
      end,
    });

    expect(
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: first,
        bindingSourceRowLimit: 2,
        cellRange: range(first, second),
        rows,
      }),
    ).toBeNull();
    expect(
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: first,
        bindingSourceRowLimit: 2,
        cellRange: range(first, mixedField),
        rows,
      }),
    ).toBeNull();
    expect(
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: first,
        bindingSourceRowLimit: 1,
        cellRange: range(first, second),
        rows: [rows[0] as NetworkFlowRow, flowRow(2, "192.0.2.10")],
      }),
    ).toBeNull();
    expect(
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: {
          ...first,
          rowIdentity: { kind: "core_record", recordId: "record-1" },
        },
        bindingSourceRowLimit: 2,
        cellRange: null,
        rows,
      }),
    ).toBeNull();
    expect(
      resolveNetworkFlowRowLinkSelection({
        activeAnchor: {
          ...first,
          surface: {
            kind: "extension_grid",
            extensionProfileId: "another_extension",
            workspaceKey: "network_analysis",
            gridSchemaId: "network_flow.accepted_rows.v1",
          },
        },
        bindingSourceRowLimit: 2,
        cellRange: null,
        rows,
      }),
    ).toBeNull();
  });
});

function anchor(
  row: NetworkFlowRow | undefined,
  fieldKey: "network_flow.src_ip" | "network_flow.dst_ip",
): GridCellAnchor {
  if (row === undefined) {
    throw new Error("A row is required for a semantic test anchor.");
  }
  return {
    surface: {
      kind: "extension_grid",
      extensionProfileId: "network_flow_activity",
      workspaceKey: "network_analysis",
      gridSchemaId: "network_flow.accepted_rows.v1",
    },
    rowIdentity: {
      kind: "extension_resource",
      extensionProfileId: "network_flow_activity",
      resourceKind: "network_flow_row",
      resourceId: row.network_flow_row_id,
    },
    fieldKey,
  };
}

function flowRow(sourceRowNumber: number, sourceIP: string): NetworkFlowRow {
  const suffix = String(sourceRowNumber).padStart(64, "0");
  return {
    network_flow_row_id: `nfr_${suffix}`,
    network_flow_table_id: tableId,
    incident_id: "11111111-1111-4111-8111-111111111111",
    source_row_number: sourceRowNumber,
    source_row_digest_sha256: "b".repeat(64),
    normalized_row_digest_sha256: "c".repeat(64),
    mapping_fingerprint: mappingFingerprint,
    "network_flow.flow_start_utc": "2026-07-10T12:00:00Z",
    "network_flow.flow_end_utc": "2026-07-10T12:01:00Z",
    "network_flow.src_ip": sourceIP,
    "network_flow.dst_ip": "203.0.113.30",
    "network_flow.src_port": 443,
    "network_flow.dst_port": 8443,
    "network_flow.ip_protocol": 6,
    "network_flow.bytes_count": "100",
    "network_flow.packets_count": "2",
    "network_flow.exporter_id": null,
    "network_flow.input_interface": null,
    "network_flow.output_interface": null,
    "network_flow.tcp_flags": null,
    "network_flow.application_label": null,
    unmapped_raw: {},
    "network_flow.observation_source_ref": {
      import_session_id: "import-session-1",
      import_unit_id: "import-unit-1",
      source_content_sha256: "d".repeat(64),
      source_profile_id: "cisco_sna_netflow_csv_v1",
      parser_profile_id: "rfc4180_headered_csv_v1",
      mapping_fingerprint: mappingFingerprint,
      source_row_number: sourceRowNumber,
      source_row_digest_sha256: "b".repeat(64),
    },
    created_at: "2026-07-10T12:02:00Z",
    created_by_user_id: "user-1",
  };
}
