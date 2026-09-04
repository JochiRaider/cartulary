import type { GridCellAnchor, GridCellRange } from "@cartulary/grid-adapter";
import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import { type CSSProperties, useCallback, useMemo, useState } from "react";
import type {
  NetworkFlowContributor,
  NetworkFlowDiagnostic,
  NetworkFlowRow,
  NetworkFlowTable,
} from "../services/networkFlowContractAdapter";
import {
  NetworkFlowButton,
  NetworkFlowChromeStyles,
  networkFlowChromeRootClassName,
} from "./NetworkFlowControls";
import {
  NetworkFlowSavedGraphPanel,
  type NetworkFlowSavedGraphPanelController,
} from "./NetworkFlowSavedGraphPanel";
import {
  NetworkFlowAcceptedGrid,
  NetworkFlowContributorGrid,
  NetworkFlowRejectedGrid,
} from "./NetworkFlowSemanticGrid";
import type {
  NetworkFlowGraphResult,
  NetworkFlowSavedGraph,
  NetworkFlowSavedGraphResult,
} from "./networkFlowClient";
import {
  reconcileNetworkFlowContributors,
  reconcileNetworkFlowDiagnostics,
  reconcileNetworkFlowRows,
} from "./networkFlowQueryModel";

type FixtureSurface = "accepted" | "contributors" | "rejected" | "saved";

export function NetworkFlowGridLoadFixture() {
  const logicalRowCount = fixtureRowCount();
  const [surface, setSurface] = useState<FixtureSurface>("accepted");
  const [rows, setRows] = useState<readonly NetworkFlowRow[]>(() =>
    fixtureRows(logicalRowCount),
  );
  const [diagnostics, setDiagnostics] = useState<
    readonly NetworkFlowDiagnostic[]
  >(() => fixtureDiagnostics(logicalRowCount));
  const [contributors, setContributors] = useState<
    readonly NetworkFlowContributor[]
  >(() => fixtureContributors(logicalRowCount));
  const tables = useMemo(() => fixtureTables(), []);
  const savedGraphController = useMemo(() => fixtureSavedGraphController(), []);
  const [refreshCount, setRefreshCount] = useState(0);
  const [selectionSummary, setSelectionSummary] = useState("No selection");
  const handleSelectionChange = useCallback(
    (active: GridCellAnchor | null, range: GridCellRange | null) => {
      setSelectionSummary(
        active === null
          ? "No selection"
          : `${anchorSummary(active)}; range ${range === null ? "none" : `${anchorSummary(range.start)} to ${anchorSummary(range.end)}`}`,
      );
    },
    [],
  );

  const refreshEquivalentResources = () => {
    setRows((current) =>
      reconcileNetworkFlowRows(current, fixtureRows(logicalRowCount)),
    );
    setDiagnostics((current) =>
      reconcileNetworkFlowDiagnostics(
        current,
        fixtureDiagnostics(logicalRowCount),
      ),
    );
    setContributors((current) =>
      reconcileNetworkFlowContributors(
        current,
        fixtureContributors(logicalRowCount),
      ),
    );
    setRefreshCount((current) => current + 1);
  };

  return (
    <section
      aria-label="Network Flow supported-load fixture"
      className={networkFlowChromeRootClassName}
      data-logical-row-count={logicalRowCount}
      data-testid={networkAnalysisTestId("load-fixture")}
      style={fixtureStyle}
    >
      <NetworkFlowChromeStyles />
      <header style={headerStyle}>
        <div>
          <strong>Network Flow supported-load fixture</strong>
          <div>
            {logicalRowCount.toLocaleString("en-US")} deterministic resources;
            refresh generation {refreshCount}
          </div>
        </div>
        <NetworkFlowButton
          variant="secondary"
          onClick={refreshEquivalentResources}
        >
          Refresh equivalent resources
        </NetworkFlowButton>
        <output aria-label="Fixture semantic selection">
          {selectionSummary}
        </output>
      </header>
      <nav aria-label="Fixture grid surface" style={surfaceControlsStyle}>
        {(["accepted", "rejected", "contributors", "saved"] as const).map(
          (candidate) => (
            <NetworkFlowButton
              aria-pressed={surface === candidate}
              key={candidate}
              selected={surface === candidate}
              variant="mode"
              onClick={() => setSurface(candidate)}
            >
              {fixtureSurfaceLabel(candidate)}
            </NetworkFlowButton>
          ),
        )}
      </nav>
      <div style={gridHostStyle}>
        {surface === "accepted" ? (
          <NetworkFlowAcceptedGrid
            error={null}
            filtered={false}
            loadState="ready"
            resetKey="supported-load-fixture"
            rows={rows}
            sort={[]}
            onResetQuery={() => undefined}
            onRetry={() => undefined}
            onSelectionChange={handleSelectionChange}
            onSortChange={() => undefined}
          />
        ) : null}
        {surface === "rejected" ? (
          <NetworkFlowRejectedGrid
            diagnostics={diagnostics}
            error={null}
            filtered={false}
            loadState="ready"
            resetKey="supported-load-fixture"
            onResetQuery={() => undefined}
            onRetry={() => undefined}
          />
        ) : null}
        {surface === "contributors" ? (
          <NetworkFlowContributorGrid
            contributors={contributors}
            error={null}
            loadState="ready"
            tables={tables}
            onRetry={() => undefined}
          />
        ) : null}
        {surface === "saved" ? (
          <NetworkFlowSavedGraphPanel
            canCreate={false}
            canRetire={false}
            controller={savedGraphController}
            currentGraph={null}
          />
        ) : null}
      </div>
    </section>
  );
}

function anchorSummary(anchor: GridCellAnchor): string {
  const resourceID =
    anchor.rowIdentity.kind === "extension_resource"
      ? anchor.rowIdentity.resourceId
      : anchor.rowIdentity.recordId;
  return `${resourceID}:${anchor.fieldKey}`;
}

function fixtureRowCount(): number {
  const value = new URLSearchParams(window.location.search).get("fixture_rows");
  return value === "100" ? 100 : 1_000;
}

function fixtureRows(count: number): NetworkFlowRow[] {
  return Array.from({ length: count }, (_, index) => {
    const ordinal = index + 1;
    const suffix = String(ordinal).padStart(4, "0");
    return {
      mapping_fingerprint: "a".repeat(64),
      network_flow_row_id: `nfr_load_${suffix}`,
      network_flow_table_id: `nft_load_${(index % 4) + 1}`,
      source_row_number: ordinal,
      unmapped_raw: {},
      "network_flow.application_label": `application-${ordinal % 17}`,
      "network_flow.bytes_count": String(ordinal * 1_024),
      "network_flow.dst_ip": `203.0.113.${(index % 250) + 1}`,
      "network_flow.dst_port": 8_000 + (index % 100),
      "network_flow.exporter_id": `exporter-${index % 8}`,
      "network_flow.flow_end_utc": "2026-07-16T00:01:00Z",
      "network_flow.flow_start_utc": "2026-07-16T00:00:00Z",
      "network_flow.input_interface": `ingress-${index % 4}`,
      "network_flow.ip_protocol": 6,
      "network_flow.observation_source_ref": {
        import_unit_id: `niu_load_${(index % 4) + 1}`,
      },
      "network_flow.output_interface": `egress-${index % 4}`,
      "network_flow.packets_count": String(ordinal * 4),
      "network_flow.src_ip": `192.0.2.${(index % 250) + 1}`,
      "network_flow.src_port": 1_000 + (index % 1_000),
      "network_flow.tcp_flags": index % 256,
    } as NetworkFlowRow;
  });
}

function fixtureDiagnostics(count: number): NetworkFlowDiagnostic[] {
  return Array.from({ length: count }, (_, index) => {
    const ordinal = index + 1;
    const suffix = String(ordinal).padStart(4, "0");
    return {
      actual_value: null,
      diagnostic_id: `nfd_load_${suffix}`,
      error_code: "network_flow_invalid_ip",
      field_key: "network_flow.src_ip",
      limit_name: null,
      limit_value: null,
      message: "The value is not a valid IP address.",
      message_args: {},
      message_key:
        "network_flow.diagnostic.network_flow_invalid_ip.invalid_ipv4",
      raw_header_sha256: null,
      raw_value_sha256: null,
      reason_code: "invalid_ipv4",
      safe_sample: `invalid-ip-${ordinal}`,
      source_column_ordinal: 3,
      source_row_number: ordinal,
    } as NetworkFlowDiagnostic;
  });
}

function fixtureContributors(count: number): NetworkFlowContributor[] {
  return fixtureRows(count).map((row) => ({
    row,
    row_ref: {
      mapping_fingerprint: row.mapping_fingerprint,
      network_flow_row_id: row.network_flow_row_id,
      network_flow_table_id: row.network_flow_table_id,
      source_row_number: row.source_row_number,
    },
  }));
}

function fixtureTables(): NetworkFlowTable[] {
  return Array.from({ length: 4 }, (_, index) => ({
    display_name: `Load fixture table ${index + 1}`,
    network_flow_table_id: `nft_load_${index + 1}`,
  })) as NetworkFlowTable[];
}

function fixtureSurfaceLabel(surface: FixtureSurface): string {
  switch (surface) {
    case "accepted":
      return "Accepted rows";
    case "rejected":
      return "Rejected diagnostics";
    case "contributors":
      return "Graph contributors";
    case "saved":
      return "Saved graph result";
  }
}

function fixtureSavedGraphController(): NetworkFlowSavedGraphPanelController {
  const result = fixtureSavedGraphResult();
  const graph = result.graph_view;
  return {
    contributorState: "idle",
    contributors: [],
    createGraph: async () => true,
    graphs: [graph],
    listState: "ready",
    loadGraphs: async () => undefined,
    loadResult: async () => undefined,
    mutationPending: false,
    notice: null,
    refreshGraph: async () => true,
    renameGraph: async () => true,
    result,
    resultState: "ready",
    retireGraph: async () => true,
    selectedGraph: graph,
    selectedGraphViewId: graph.graph_view_id,
    selection: null,
    selectGraphView: () => undefined,
    selectObject: async () => undefined,
  };
}

function fixtureSavedGraphResult(): NetworkFlowSavedGraphResult {
  const vertices = Array.from({ length: 501 }, (_, index) => {
    const identity = index.toString(16).padStart(64, "0");
    return {
      vertex_id: `vx_${identity}`,
      vertex_kind: "network_flow.ip_endpoint.v1",
      vertex_family: "network_flow.ip_endpoint.v1",
      labels: [],
      properties: {
        endpoint_kind: "ip",
        endpoint_value: `fixture-endpoint-${String(index + 1).padStart(4, "0")}`,
        contributing_table_ids: ["nft_load_1"],
        flow_row_count: 1,
        indicator_candidate_value: `fixture-endpoint-${String(index + 1).padStart(4, "0")}`,
      },
      metadata: {
        mapping_rule_id: "nf.map.ip_endpoint.v1",
        aggregation_rule_id: null,
        aggregation_source_refs: [],
        mapped_metadata: {},
      },
      source_entity_ref: {
        source_entity_id: `nfe_${identity}`,
        source_entity_kind: "network_flow.ip_endpoint.v1",
        mapping_rule_id: "nf.map.ip_endpoint.v1",
      },
      sort_key: `nfe_${identity}`,
    };
  });
  const edges = Array.from({ length: 1_001 }, (_, index) => {
    const identity = index.toString(16).padStart(64, "0");
    const sourceIndex = index % vertices.length;
    const destinationIndex = (index + 1) % vertices.length;
    const sourceIdentity = sourceIndex.toString(16).padStart(64, "0");
    const destinationIdentity = destinationIndex.toString(16).padStart(64, "0");
    return {
      edge_id: `ed_${identity}`,
      edge_kind: "network_flow.flow_edge.v1",
      edge_family: "direct",
      src_vertex_id: `vx_${sourceIdentity}`,
      dst_vertex_id: `vx_${destinationIdentity}`,
      direction: "directed",
      labels: [],
      properties: {
        edge_id: `nff_${identity}`,
        src_endpoint_id: `nfe_${sourceIdentity}`,
        dst_endpoint_id: `nfe_${destinationIdentity}`,
        ip_protocol: 6,
        flow_row_count: 1,
      },
      metadata: {
        mapping_rule_id: "nf.map.flow_edge.v1",
        aggregation_rule_id: null,
        is_reverse_edge: false,
        reverse_of_edge_id: null,
        aggregation_source_refs: [],
        mapped_metadata: {},
      },
      source_relationship_ref: {
        source_relationship_id: `nff_${identity}`,
        source_relationship_kind: "network_flow.flow_edge.v1",
        mapping_rule_id: "nf.map.flow_edge.v1",
      },
      sort_key: `nff_${identity}`,
    };
  });
  const semanticQuery = {
    schema_id: "cartulary.network_flow.graph_semantic_query.v2" as const,
    selected_table_ids: ["nft_load_1"] as [string],
    filters: [] as [],
    time_range: { start_utc: null, end_utc: null },
    aggregation: {
      mode: "default_flow_edge_v1" as const,
      include_example_row_refs: false,
    },
  };
  const graph = {
    schema_id: "cartulary.network_flow.graph_view.v3",
    graph_view_id: "nfgv_load_fixture_0000000000000001",
    incident_id: "00000000-0000-0000-0000-000000000000",
    display_name: "Supported-load saved graph",
    normalized_display_name: "supported-load saved graph",
    graph_view_version: 1,
    materialization_generation: 1,
    state: "active",
    semantic_query: semanticQuery,
    selected_result: {
      projection_result_id: `gpres_${"1".repeat(64)}`,
      source_snapshot_id: "supported-load-snapshot",
      projection_schema_id: "graph_projection.v2",
      projection_version: "network_flow_activity.v1",
      normalized_configuration_sha256: "2".repeat(64),
      normalized_source_sha256: "3".repeat(64),
      canonical_output_sha256: "4".repeat(64),
    },
    last_materialization_job_id: "supported-load-job",
    last_materialization_status: "succeeded",
    last_failure_code: null,
    created_at: "2026-07-16T00:00:00Z",
    updated_at: "2026-07-16T00:00:00Z",
  } as NetworkFlowSavedGraph;
  const projectionResult: NetworkFlowGraphResult["graph_projection_result"] = {
    projection_schema_id: "graph_projection.v2" as const,
    projection_result_id: graph.selected_result?.projection_result_id ?? "",
    graph_view_id: graph.graph_view_id,
    source_owner_id: "network_flow_activity" as const,
    source_snapshot_id: "supported-load-snapshot",
    projection_version: "network_flow_activity.v1",
    normalized_configuration_sha256: "2".repeat(64),
    normalized_source_sha256: "3".repeat(64),
    canonical_output_sha256: "4".repeat(64),
    properties: {},
    mapped_metadata: {},
    schema_registry: {
      vertex_kinds: [],
      edge_kinds: [],
      property_keys: [],
      metadata_keys: [],
    },
    vertices,
    edges,
    validation_summary: {
      status: "passed",
      fatal_count: 0,
      error_count: 0,
      warning_count: 0,
      info_count: 0,
      issues: [],
    },
    consumer_capabilities: {
      query_shapes: [],
      supports_direct_vertex_lookup: false,
      supports_direct_edge_lookup: false,
      supports_breadth_first_traversal: false,
      supports_alternate_traversal_order: [],
      max_traversal_depth: 0,
      max_traversal_seed_vertices: 0,
      max_kind_filters: 0,
    },
  };
  return {
    schema_id: "cartulary.network_flow.graph_view_result.v3",
    graph_view: graph,
    result: {
      schema_id: "cartulary.network_flow.graph_query_result.v2",
      graph_query_digest: "6".repeat(64),
      semantic_query: semanticQuery,
      graph_projection_result: projectionResult,
      vertex_selectors: vertices.map((vertex) => ({
        projected_vertex_id: vertex.vertex_id,
        selector: {
          kind: "vertex" as const,
          source_vertex_id: vertex.source_entity_ref.source_entity_id,
          endpoint_value: String(vertex.properties.endpoint_value),
        },
      })),
      edge_annotations: edges.map((edge) => ({
        projected_edge_id: edge.edge_id,
        selector: {
          kind: "default_edge" as const,
          source_edge_id: edge.source_relationship_ref.source_relationship_id,
          source_endpoint_value: String(
            vertices.find((vertex) => vertex.vertex_id === edge.src_vertex_id)
              ?.properties.endpoint_value,
          ),
          destination_endpoint_value: String(
            vertices.find((vertex) => vertex.vertex_id === edge.dst_vertex_id)
              ?.properties.endpoint_value,
          ),
          protocol: 6,
          destination_port_present: false as const,
        },
        example_row_refs: [],
        example_refs_truncated: false,
        example_refs_total_count: 1,
      })),
      source_table_refs: [
        {
          network_flow_table_id: "nft_load_1",
          table_version: 1,
          mapping_fingerprint: "5".repeat(64),
          row_count_accepted: 1_001,
          row_count_rejected: 0,
        },
      ],
      result_limits: {
        max_vertices: 5_000,
        max_edges: 10_000,
        max_example_row_refs_per_edge: 10,
        max_aggregate_counter_digits: 39,
        max_contributing_rows_per_graph: 250_000,
        max_time_buckets_per_graph: 256,
      },
      result_variant: { kind: "default_flow_edge_v1" },
    },
  } as NetworkFlowSavedGraphResult;
}

const fixtureStyle = {
  blockSize: "min(48rem, calc(100vh - 2rem))",
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  gridTemplateRows: "auto auto minmax(0, 1fr)",
  inlineSize: "min(100%, calc(100vw - 4rem))",
  maxInlineSize: "96rem",
  minBlockSize: "36rem",
  minWidth: 0,
  overflow: "hidden",
} satisfies CSSProperties;

const headerStyle = {
  alignItems: "center",
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-md)",
  justifyContent: "space-between",
} satisfies CSSProperties;

const surfaceControlsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const gridHostStyle = {
  display: "grid",
  minBlockSize: 0,
  minWidth: 0,
} satisfies CSSProperties;
