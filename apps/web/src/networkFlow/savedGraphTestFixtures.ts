import type { NetworkFlowSavedGraph } from "../services/networkFlowContractAdapter";
import type {
  SavedGraphAuthority,
  SavedGraphReceipt,
} from "./savedGraphOperation";
export const savedGraphTestAuthority: SavedGraphAuthority = {
  incidentId: "00000000-0000-4000-8000-000000000001",
  actorId: "00000000-0000-4000-8000-000000000002",
  session: {},
  role: "editor",
  open: true,
  available: true,
  availabilityTag: { epochId: "epoch", generation: 1n },
};
export function savedGraphFixture(
  patch: Partial<NetworkFlowSavedGraph> = {},
): NetworkFlowSavedGraph {
  const graph = {
    schema_id: "cartulary.network_flow.graph_view.v4" as const,
    graph_view_id: `nfgv_${"a".repeat(32)}`,
    incident_id: savedGraphTestAuthority.incidentId,
    display_name: "Graph A",
    state: "active" as const,
    semantic_query: {
      schema_id: "cartulary.network_flow.graph_semantic_query.v2" as const,
      selected_table_ids: [`nft_${"c".repeat(32)}`] as [string],
      filters: [] as [],
      time_range: { start_utc: null, end_utc: null },
      aggregation: {
        mode: "default_flow_edge_v1" as const,
        include_example_row_refs: false,
      },
    },
    semantic_query_sha256: "a".repeat(64),
    desired_source_snapshot_id: "snapshot-1",
    selected_result_binding: null,
    graph_view_version: 1,
    materialization_generation: 1,
    created_by: savedGraphTestAuthority.actorId ?? "",
    created_at: "2026-09-09T12:00:00Z",
    updated_at: "2026-09-09T12:00:00Z",
    latest_job_id: "00000000-0000-4000-8000-000000000004",
    last_failure_code: null,
    last_failed_at: null,
    ...patch,
  };
  return graph.last_failure_code === null
    ? { ...graph, last_failure_code: null, last_failed_at: null }
    : {
        ...graph,
        last_failure_code: graph.last_failure_code,
        last_failed_at: graph.last_failed_at ?? graph.updated_at,
      };
}
export function savedGraphAccepted(
  graph = savedGraphFixture(),
): SavedGraphReceipt {
  return {
    kind: "accepted",
    value: {
      schema_id: "cartulary.network_flow.graph_view_accepted.v4",
      graph_view: graph,
      job: {
        job_id: graph.latest_job_id ?? "",
        status_route: `/api/v1/jobs/${graph.latest_job_id}`,
      },
    },
  };
}
export function deferredSavedGraph<T>() {
  let resolve!: (value: T) => void, reject!: (reason: unknown) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}

export const savedGraphBindingFixture = {
  projection_result_id: `gpres_${"d".repeat(64)}`,
  source_snapshot_id: "snapshot-1",
  projection_schema_id: "graph_projection.v2" as const,
  projection_version: "network_flow_activity.v1",
  normalized_configuration_sha256: "a".repeat(64),
  normalized_source_sha256: "b".repeat(64),
  canonical_output_sha256: "c".repeat(64),
};
export const savedGraphSelectorFixture = {
  kind: "vertex" as const,
  source_vertex_id: `nfe_${"1".repeat(64)}`,
  endpoint_value: "192.0.2.1",
};
export function savedGraphResultFixture(
  graph = savedGraphFixture({
    selected_result_binding: savedGraphBindingFixture,
  }),
): import("../services/networkFlowContractAdapter").NetworkFlowSavedGraphResult {
  if (graph.selected_result_binding === null)
    throw new Error("Fixture requires selected binding");
  return {
    schema_id: "cartulary.network_flow.graph_view_result.v4",
    graph_view: graph,
    result: {
      schema_id: "cartulary.network_flow.graph_query_result.v2",
      graph_query_digest: graph.semantic_query_sha256,
      semantic_query: graph.semantic_query,
      graph_projection_result: {
        ...graph.selected_result_binding,
        graph_view_id: graph.graph_view_id,
        source_owner_id: "network_flow_activity",
        properties: {},
        mapped_metadata: {},
        schema_registry: {
          vertex_kinds: [],
          edge_kinds: [],
          property_keys: [],
          metadata_keys: [],
        },
        vertices: [],
        edges: [],
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
      },
      vertex_selectors: [],
      edge_annotations: [],
      source_table_refs: [],
      result_limits: {
        max_vertices: 1000,
        max_edges: 1000,
        max_example_row_refs_per_edge: 10,
        max_aggregate_counter_digits: 20,
        max_contributing_rows_per_graph: 250000,
        max_time_buckets_per_graph: 256,
      },
      result_variant: { kind: "default_flow_edge_v1" },
    },
  };
}
export function savedGraphJobFixture(
  status: import("../services/commonJobContract").CommonJobResource["status"] = "queued",
): import("../services/commonJobContract").CommonJobResource {
  const graph = savedGraphFixture(),
    jobId = graph.latest_job_id ?? "";
  const terminal = ["succeeded", "failed", "canceled"].includes(status);
  return {
    job_id: jobId,
    scope: { kind: "incident", incident_id: graph.incident_id },
    status_route: `/api/v1/jobs/${jobId}`,
    status,
    cancelable: status === "queued" || status === "running",
    submitted_by_user_id: graph.created_by,
    submitted_at: "2026-09-09T12:00:00Z",
    updated_at: terminal ? "2026-09-09T12:00:02Z" : "2026-09-09T12:00:01Z",
    started_at: status === "queued" ? null : "2026-09-09T12:00:01Z",
    finished_at: terminal ? "2026-09-09T12:00:02Z" : null,
    retained_until: terminal ? "2026-09-16T12:00:02Z" : null,
    progress: { completed: status === "succeeded" ? 1 : 0, total: 1 },
    result_summary:
      status === "succeeded"
        ? {
            code: "network_flow_graph_view_materialized",
            message: "Materialized",
            resource_refs: [
              {
                kind: "network_flow_graph_view",
                id: graph.graph_view_id,
                route: `/api/v1/incidents/${graph.incident_id}/network-flow/graph-views/${graph.graph_view_id}`,
              },
            ],
          }
        : status === "canceled"
          ? { code: "job_canceled", message: "Canceled", resource_refs: [] }
          : null,
    error_summary:
      status === "failed"
        ? {
            code: "network_flow_graph_materialization_failed",
            message: "Materialization failed",
            retryable: false,
            details: { reason_code: "timeout" },
          }
        : null,
  };
}
