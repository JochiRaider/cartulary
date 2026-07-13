import { requireContractArtifactJSON } from "@cartulary/protocol-ts";

const schemaBundle = JSON.parse(
  requireContractArtifactJSON("contracts/network-flow/schemas.v1.json"),
) as { readonly $defs?: Readonly<Record<string, unknown>> };

for (const requiredDefinition of [
  "NetworkFlowTable",
  "NetworkFlowRow",
  "RejectedRowDiagnostic",
  "GraphQueryResult",
  "GraphContributorQueryResult",
  "IndicatorLinkResult",
]) {
  if (schemaBundle.$defs?.[requiredDefinition] === undefined) {
    throw new Error(
      `missing Network Flow generated contract ${requiredDefinition}`,
    );
  }
}

// These are the sole feature-local TypeScript projections of generated wire
// definitions. API, controller, collaboration, and presentation code import
// them through this adapter rather than redeclaring response shapes.
export type NetworkFlowTable = {
  readonly network_flow_table_id: string;
  readonly display_name: string;
  readonly table_version: number;
  readonly table_status: "active" | "soft_deleted";
  readonly source_filename_display: string;
  readonly mapping_fingerprint: string;
  readonly row_count_accepted: number;
  readonly row_count_rejected: number;
  readonly diagnostics_truncated: boolean;
  readonly created_at: string;
  readonly updated_at: string;
};

export type NetworkFlowRowRef = {
  readonly network_flow_table_id: string;
  readonly network_flow_row_id: string;
  readonly source_row_number: number;
  readonly mapping_fingerprint: string;
};

export type NetworkFlowRow = {
  readonly network_flow_row_id: string;
  readonly network_flow_table_id: string;
  readonly source_row_number: number;
  readonly mapping_fingerprint: string;
  readonly "network_flow.flow_start_utc": string;
  readonly "network_flow.flow_end_utc": string;
  readonly "network_flow.src_ip": string;
  readonly "network_flow.dst_ip": string;
  readonly "network_flow.src_port": number | null;
  readonly "network_flow.dst_port": number | null;
  readonly "network_flow.ip_protocol": number;
  readonly "network_flow.bytes_count": string;
  readonly "network_flow.packets_count": string;
  readonly "network_flow.exporter_id": string | null;
  readonly "network_flow.input_interface": string | null;
  readonly "network_flow.output_interface": string | null;
  readonly "network_flow.application_label": string | null;
};

export type NetworkFlowDiagnostic = {
  readonly diagnostic_id: string;
  readonly source_row_number: number;
  readonly source_column_ordinal: number | null;
  readonly field_key: string | null;
  readonly error_code: string;
  readonly reason_code: string;
  readonly safe_sample: string | null;
  readonly message: string;
};

export type NetworkFlowPaging = {
  readonly limit: number;
  readonly returned_count: number;
  readonly next_cursor_token: string | null;
};

export type NetworkFlowGraphSemanticQuery = {
  readonly schema_id: "cartulary.network_flow.graph_semantic_query.v1";
  readonly selected_table_ids: string[];
  readonly filters: unknown[];
  readonly time_range: {
    readonly start_utc: string | null;
    readonly end_utc: string | null;
  };
  readonly aggregation: {
    readonly mode: "default_flow_edge_v1";
    readonly include_example_row_refs: boolean;
  };
  readonly result_limits: {
    readonly max_vertices: number;
    readonly max_edges: number;
    readonly max_example_row_refs_per_edge: number;
    readonly max_aggregate_counter_digits: number;
  };
};

export type NetworkFlowEdgeAnnotation = {
  readonly edge_id: string;
  readonly example_row_refs: NetworkFlowRowRef[];
  readonly example_refs_truncated: boolean;
  readonly example_refs_total_count: number;
};

export type NetworkFlowGraphResult = {
  readonly schema_id: "cartulary.network_flow.graph_query_result.v1";
  readonly graph_query_digest: string;
  readonly semantic_query: NetworkFlowGraphSemanticQuery;
  readonly graph_projection_result: {
    readonly projection_schema_id: "graph_projection.v1";
    readonly graph_view_id: string;
    readonly graph_view_key: string;
    readonly ephemeral_projection_id: string;
    readonly source_snapshot_id: string;
    readonly projection_version: string;
    readonly generated_at: string;
    readonly state: "ephemeral_available";
    readonly properties: Readonly<Record<string, unknown>>;
    readonly metadata: Readonly<Record<string, unknown>>;
    readonly schema_registry: Readonly<Record<string, unknown>>;
    readonly vertices: ReadonlyArray<Readonly<Record<string, unknown>>>;
    readonly edges: ReadonlyArray<Readonly<Record<string, unknown>>>;
    readonly validation_summary: {
      readonly status: "valid";
      readonly fatal_count: 0;
      readonly error_count: 0;
      readonly warning_count: 0;
      readonly info_count: 0;
      readonly issues: readonly [];
    };
    readonly consumer_capabilities: Readonly<Record<string, unknown>>;
  };
  readonly edge_annotations: NetworkFlowEdgeAnnotation[];
  readonly source_table_refs: Array<{
    readonly network_flow_table_id: string;
    readonly table_version: number;
    readonly mapping_fingerprint: string;
    readonly row_count_accepted: number;
    readonly row_count_rejected: number;
  }>;
};

export type NetworkFlowContributor = {
  readonly row_ref: NetworkFlowRowRef;
  readonly row: NetworkFlowRow;
};

export type NetworkFlowContributorResult = {
  readonly schema_id: "cartulary.network_flow.graph_contributor_query_result.v1";
  readonly graph_query_digest: string;
  readonly selector:
    | { readonly kind: "edge"; readonly edge_id: string }
    | { readonly kind: "vertex"; readonly vertex_id: string };
  readonly contributors: NetworkFlowContributor[];
  readonly meta: { readonly paging: NetworkFlowPaging };
};

export type NetworkFlowIndicatorLinkResult = {
  readonly schema_id: "cartulary.network_flow.indicator_link_result.v1";
  readonly duplicate: boolean;
  readonly binding: {
    readonly network_flow_indicator_binding_id: string;
    readonly target_indicator_ref: {
      readonly indicator_id: string;
      readonly indicator_type: "ipv4_addr" | "ipv6_addr";
      readonly normalized_value: string;
    };
    readonly source_row_refs_total_count: number;
  };
};
