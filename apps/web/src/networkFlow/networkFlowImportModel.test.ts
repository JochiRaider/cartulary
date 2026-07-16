import { describe, expect, it } from "vitest";
import type { DiscoveredImportColumn } from "../shared/importCoordinator";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  ignoredColumnChoice,
  networkFlowMappingDraftReadyForPreview,
  sourceColumnLabel,
  withNetworkFlowColumnChoice,
} from "./networkFlowImportModel";

describe("Network Flow import mapping model", () => {
  it("keeps duplicate aliases distinct and blocks preview until every collision is explicit", () => {
    const columns = [
      column(1, "Source IP"),
      column(2, "Source IP"),
      column(3, null),
      column(4, "Destination IP"),
    ];
    const draft = createNetworkFlowMappingDraft(columns);

    expect(draft.columnChoices).toEqual({
      1: null,
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
): DiscoveredImportColumn {
  return {
    source_column_ordinal: sourceColumnOrdinal,
    source_header_text: sourceHeaderText,
  };
}
