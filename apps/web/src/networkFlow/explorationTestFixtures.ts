import type {
  NetworkFlowGraphResult,
  NetworkFlowGraphSelector,
} from "./networkFlowClient";
import { savedGraphResultFixture } from "./savedGraphTestFixtures";
import { tableFixture } from "./tableLifecycleTestFixtures";

export function explorationFixture(
  vertexCount = 501,
  edgeCount = 1001,
  temporal = false,
): NetworkFlowGraphResult {
  const base = savedGraphResultFixture().result;
  const table = tableFixture();
  const identity = (prefix: string, n: number) =>
    `${prefix}_${n.toString(16).padStart(64, "0")}`;
  const endpoint = (n: number) => `192.0.${Math.floor(n / 256)}.${n % 256}`;
  const buckets = [0, 1, 2].map((n) => ({
    start_utc: `2026-07-10T0${n}:00:00.000Z`,
    end_utc: `2026-07-10T0${n + 1}:00:00.000Z`,
    unique_vertex_count: n === 1 ? 0 : vertexCount,
    edge_count: n === 1 ? 0 : edgeCount,
    contributing_row_count: n === 1 ? 0 : edgeCount,
  }));
  const firstBucket = buckets[0],
    lastBucket = buckets[2];
  if (!firstBucket || !lastBucket)
    throw new Error("Missing temporal fixture buckets");
  const vertices = Array.from({ length: vertexCount }, (_, n) => ({
    vertex_id: identity("vx", n),
    vertex_kind: "network_flow.ip_endpoint.v1",
    vertex_family: "network_flow.ip_endpoint.v1",
    labels: [],
    properties: {
      endpoint_kind: "ip",
      endpoint_value: endpoint(n),
      contributing_table_ids: [table.network_flow_table_id],
      flow_row_count: temporal ? 2 : 1,
      indicator_candidate_value: endpoint(n),
    },
    metadata: {
      mapping_rule_id: "nf.map.ip_endpoint.v1",
      aggregation_rule_id: null,
      aggregation_source_refs: [],
      mapped_metadata: {},
    },
    source_entity_ref: {
      source_entity_id: identity("nfe", n),
      source_entity_kind: "network_flow.ip_endpoint.v1",
      mapping_rule_id: "nf.map.ip_endpoint.v1",
    },
    sort_key: identity("nfe", n),
  }));
  const edges = (temporal ? [0, 2] : [0]).flatMap((bucket) =>
    Array.from({ length: vertexCount ? edgeCount : 0 }, (_, n) => {
      const id = identity(temporal ? "nfbe" : "nff", bucket * edgeCount + n);
      return {
        edge_id: identity("ed", bucket * edgeCount + n),
        edge_kind: temporal
          ? "network_flow.bucketed_flow_edge.v1"
          : "network_flow.flow_edge.v1",
        edge_family: "direct",
        direction: "directed" as const,
        labels: [],
        src_vertex_id: identity("vx", n % vertexCount),
        dst_vertex_id: identity("vx", (n + 1) % vertexCount),
        properties: {
          edge_id: id,
          src_endpoint_id: identity("nfe", n % vertexCount),
          dst_endpoint_id: identity("nfe", (n + 1) % vertexCount),
          ip_protocol: 6,
          dst_port: n,
          flow_row_count: 1,
          contributing_table_ids: [table.network_flow_table_id],
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
          source_relationship_id: id,
          source_relationship_kind: temporal
            ? "network_flow.bucketed_flow_edge.v1"
            : "network_flow.flow_edge.v1",
          mapping_rule_id: "nf.map.flow_edge.v1",
        },
        sort_key: id,
      };
    }),
  );
  const edge_annotations = edges.map((edge, n) => {
    const position = n % edgeCount;
    const key = {
      source_edge_id: edge.source_relationship_ref.source_relationship_id,
      source_endpoint_value: endpoint(position % vertexCount),
      destination_endpoint_value: endpoint((position + 1) % vertexCount),
      protocol: 6,
      destination_port_present: true as const,
      destination_port: position,
    };
    const bucket = buckets[n < edgeCount ? 0 : 2];
    const selector: NetworkFlowGraphSelector =
      temporal && bucket
        ? {
            ...key,
            kind: "time_bucket_edge",
            bucket_start_utc: bucket.start_utc,
            bucket_end_utc: bucket.end_utc,
          }
        : { ...key, kind: "default_edge" };
    return {
      projected_edge_id: edge.edge_id,
      selector,
      example_row_refs: [],
      example_refs_truncated: true,
      example_refs_total_count: 1,
    };
  });
  return {
    ...base,
    semantic_query: {
      ...base.semantic_query,
      selected_table_ids: [table.network_flow_table_id],
      ...(temporal
        ? {
            aggregation: {
              mode: "time_bucket_v1",
              bucket_width_seconds: 3600,
              include_example_row_refs: false,
            },
            time_range: {
              start_utc: firstBucket.start_utc,
              end_utc: lastBucket.end_utc,
            },
          }
        : {
            aggregation: {
              mode: "default_flow_edge_v1",
              include_example_row_refs: false,
            },
            time_range: { start_utc: null, end_utc: null },
          }),
    },
    source_table_refs: [
      {
        network_flow_table_id: table.network_flow_table_id,
        table_version: 1,
        mapping_fingerprint: table.mapping_fingerprint,
        row_count_accepted: edgeCount,
        row_count_rejected: 0,
      },
    ],
    graph_projection_result: {
      ...base.graph_projection_result,
      vertices,
      edges,
    },
    vertex_selectors: vertices.map((v) => ({
      projected_vertex_id: v.vertex_id,
      selector: {
        kind: "vertex",
        source_vertex_id: v.source_entity_ref.source_entity_id,
        endpoint_value: String(v.properties.endpoint_value),
      },
    })),
    edge_annotations,
    result_variant: temporal
      ? {
          kind: "time_bucket_v1",
          time_buckets: [firstBucket, ...buckets.slice(1)],
        }
      : { kind: "default_flow_edge_v1" },
  };
}
