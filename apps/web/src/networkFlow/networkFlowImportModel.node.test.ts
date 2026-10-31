import { describe, expect, it } from "vitest";
import type { NetworkFlowMappingColumn } from "./networkFlowImportModel";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  ignoredColumnChoice,
  networkFlowMappingColumnsForDisplay,
  networkFlowMappingDraftReadyForPreview,
  sourceColumnLabel,
  withNetworkFlowColumnChoice,
} from "./networkFlowImportModel";

describe("Network Flow import mapping model", () => {
  it("maps original Cisco Flow Search headers with its own defaults", () => {
    const headers = [
      "firstActiveTime",
      "lastActiveTime",
      "searchSubject.ipAddress",
      "peer.ipAddress",
      "searchSubject.portProtocol.port",
      "peer.portProtocol.port",
      "searchSubject.portProtocol.protocol",
      "connection.transferBytes",
      "connection.transferPackets",
      "peer.portProtocol.protocol",
      "searchSubject.interface",
    ];
    const columns = headers.map((header, index) => column(index + 1, header));
    const draft = createNetworkFlowMappingDraft(
      columns,
      "cisco_sna_flow_search_csv_v1",
    );
    const candidate = buildNetworkFlowMappingCandidate(draft, columns);
    expect(candidate.source_profile_id).toBe("cisco_sna_flow_search_csv_v1");
    expect(candidate.timestamp_profile).toEqual({
      schema_id: "cartulary.network_flow.timestamp_profile.v1",
      mode: "cisco_sna_datetime_v1",
      precision: "milliseconds",
    });
    expect(candidate.field_mappings).toHaveLength(9);
    expect(
      candidate.field_mappings.filter(
        (field) =>
          field.mapping_kind === "source_column" &&
          ["network_flow.src_port", "network_flow.dst_port"].includes(
            field.field_key,
          ),
      ),
    ).toEqual([
      {
        mapping_kind: "source_column",
        field_key: "network_flow.src_port",
        source_column_ordinal: 5,
        transform_id: "port_number_v1",
        empty_value_policy: "empty_string_is_null",
        combinability: "single_source_only",
      },
      {
        mapping_kind: "source_column",
        field_key: "network_flow.dst_port",
        source_column_ordinal: 6,
        transform_id: "port_number_v1",
        empty_value_policy: "empty_string_is_null",
        combinability: "single_source_only",
      },
    ]);
    expect(candidate.unknown_column_policy).toBe("preserve_unmapped_raw");
    for (const field of candidate.field_mappings) {
      if (field.mapping_kind === "source_column") {
        expect(field.empty_value_policy).toBe(
          ["network_flow.src_port", "network_flow.dst_port"].includes(
            field.field_key,
          )
            ? "empty_string_is_null"
            : "empty_string_is_invalid",
        );
      }
    }
    const originalColumns = [
      "Start Time",
      "End Time",
      "Source IP",
      "Destination IP",
      "Source Port",
      "Destination Port",
      "Protocol",
      "Bytes",
      "Packets",
      "Input Interface",
      "Output Interface",
    ].map((header, index) => column(index + 1, header));
    const original = buildNetworkFlowMappingCandidate(
      createNetworkFlowMappingDraft(originalColumns),
      originalColumns,
    );
    expect(original.field_mappings).toHaveLength(11);
    for (const field of original.field_mappings) {
      if (field.mapping_kind === "source_column") {
        expect(field.empty_value_policy).toBe(
          [
            "network_flow.input_interface",
            "network_flow.output_interface",
          ].includes(field.field_key)
            ? "empty_string_is_null"
            : "empty_string_is_invalid",
        );
      }
    }
    expect(createNetworkFlowMappingDraft(columns).timestampMode).toBe(
      "rfc3339",
    );
  });

  it("keeps duplicate aliases distinct and blocks preview until every collision is explicit", () => {
    const columns = [
      column(1, "Source IP"),
      column(2, "Source IP"),
      column(3, null),
      column(4, "Destination IP"),
    ];
    const draft = createNetworkFlowMappingDraft(columns);

    expect(draft.columnChoices).toEqual({
      1: "network_flow.src_ip",
      2: null,
      3: null,
      4: "network_flow.dst_ip",
    });
    expect(draft.unresolvedAliasCollisionOrdinals).toEqual([1, 2]);
    expect(networkFlowMappingDraftReadyForPreview(draft)).toBe(false);
    const firstResolved = withNetworkFlowColumnChoice(
      draft,
      1,
      "network_flow.src_ip",
    );
    const resolved = withNetworkFlowColumnChoice(
      firstResolved,
      2,
      ignoredColumnChoice,
    );
    expect(resolved.unresolvedAliasCollisionOrdinals).toEqual([]);
    expect(networkFlowMappingDraftReadyForPreview(resolved)).toBe(true);
    expect(sourceColumnLabel(column(2, "Source IP"))).toBe(
      "Source IP · column 2",
    );
    expect(sourceColumnLabel(column(3, null))).toBe("(unnamed) · column 3");
    expect(sourceColumnLabel(column(4, " Source IP "))).toBe(
      " Source IP  · column 4",
    );
    const unsorted = [
      column(9, "Bytes"),
      column(5, "Source IP"),
      column(3, "Source IP"),
      column(8, "PacKets"),
      column(1, "Start Time"),
    ];
    const sortedDraft = createNetworkFlowMappingDraft(unsorted);
    expect(sortedDraft.columnChoices[3]).toBe("network_flow.src_ip");
    expect(sortedDraft.columnChoices[5]).toBeNull();
    expect(
      networkFlowMappingColumnsForDisplay(
        unsorted,
        sortedDraft.sourceProfileId,
      ).map((item) => item.source_column_ordinal),
    ).toEqual([1, 3, 5, 9, 8]);
  });

  it("builds mappings from generated transforms without a compatibility schema_id", () => {
    const columns = [column(1, "Source IP"), column(2, "unused")];
    let draft = createNetworkFlowMappingDraft(columns);
    draft = withNetworkFlowColumnChoice(
      draft,
      2,
      "network_flow.input_interface",
    );
    const candidate = buildNetworkFlowMappingCandidate(draft, columns);

    expect(candidate).not.toHaveProperty("schema_id");
    expect(candidate.field_mappings).toEqual([
      {
        mapping_kind: "source_column",
        field_key: "network_flow.src_ip",
        source_column_ordinal: 1,
        transform_id: "ip_literal_v1",
        empty_value_policy: "empty_string_is_invalid",
        combinability: "single_source_only",
      },
      {
        mapping_kind: "source_column",
        field_key: "network_flow.input_interface",
        source_column_ordinal: 2,
        transform_id: "trim_ascii_space_v1",
        empty_value_policy: "empty_string_is_null",
        combinability: "single_source_only",
      },
    ]);
  });

  it("accounts for ignored columns only through an explicit choice or ignore policy", () => {
    const columns = [column(1, "unrecognized")];
    const preserve = createNetworkFlowMappingDraft(columns);
    const reject = {
      ...preserve,
      unknownColumnPolicy: "reject_unmapped_columns" as const,
    };
    const ignore = {
      ...preserve,
      unknownColumnPolicy: "ignore_unmapped_columns" as const,
    };
    const explicit = withNetworkFlowColumnChoice(
      reject,
      1,
      ignoredColumnChoice,
    );

    expect(
      buildNetworkFlowMappingCandidate(preserve, columns).field_mappings,
    ).toHaveLength(0);
    expect(
      buildNetworkFlowMappingCandidate(reject, columns).field_mappings,
    ).toHaveLength(0);
    expect(
      buildNetworkFlowMappingCandidate(ignore, columns).field_mappings,
    ).toEqual([
      {
        mapping_kind: "ignored_source_column",
        source_column_ordinal: 1,
        ignore_reason: "user_ignored",
      },
    ]);
    expect(
      buildNetworkFlowMappingCandidate(explicit, columns).field_mappings,
    ).toEqual([
      {
        mapping_kind: "ignored_source_column",
        source_column_ordinal: 1,
        ignore_reason: "user_ignored",
      },
    ]);
  });

  it("requires distinct ordinal-aware NetFlow uptime timestamp inputs", () => {
    const base = createNetworkFlowMappingDraft([
      column(1, "Export Time"),
      column(2, "Exporter Uptime"),
    ]);
    const incomplete = {
      ...base,
      timestampMode: "netflow_sys_uptime_milliseconds" as const,
    };
    const complete = {
      ...incomplete,
      netflowExportTimeColumnOrdinal: 1,
      netflowExporterUptimeColumnOrdinal: 2,
    };

    expect(networkFlowMappingDraftReadyForPreview(incomplete)).toBe(false);
    expect(networkFlowMappingDraftReadyForPreview(complete)).toBe(true);
    expect(
      buildNetworkFlowMappingCandidate(complete, []).timestamp_profile,
    ).toMatchObject({
      mode: "netflow_sys_uptime_milliseconds",
      netflow_export_time_column_ordinal: 1,
      netflow_exporter_uptime_at_export_column_ordinal: 2,
    });
  });
});

function column(
  sourceColumnOrdinal: number,
  sourceHeaderText: string | null,
): NetworkFlowMappingColumn {
  return {
    source_column_ordinal: sourceColumnOrdinal,
    raw_header_text: sourceHeaderText ?? "",
  };
}
