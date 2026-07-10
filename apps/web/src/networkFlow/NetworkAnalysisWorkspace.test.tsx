import {
  networkAnalysisEdgeTestId,
  networkAnalysisTestId,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { NetworkAnalysisWorkspace } from "./NetworkAnalysisWorkspace";

const tableId = "nft_11111111111111111111111111111111";
const rowId =
  "nfr_1111111111111111111111111111111111111111111111111111111111111111";
const edgeId =
  "nff_2222222222222222222222222222222222222222222222222222222222222222";
const mappingFingerprint =
  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
const graphDigest =
  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";

describe("NetworkAnalysisWorkspace", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("loads tables, opens graph contributors, and links an edge endpoint", async () => {
    const fetchSpy = installNetworkFlowFetchMock();

    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );

    expect(await screen.findAllByText("flows.csv")).toHaveLength(2);
    expect(await screen.findByText("192.0.2.10:443")).toBeTruthy();

    fireEvent.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));

    expect(
      await screen.findByTestId(networkAnalysisEdgeTestId(edgeId)),
    ).toBeTruthy();
    expect(
      await screen.findByTestId(networkAnalysisTestId("contributor-drawer")),
    ).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: /link source/i }));

    await waitFor(() => {
      expect(
        fetchSpy.mock.calls.some(([input, init]) => {
          const url = requestURL(input);
          return (
            url.endsWith(
              "/api/v1/incidents/incident-1/network-flow/indicator-links",
            ) && init?.method === "POST"
          );
        }),
      ).toBe(true);
    });

    const linkCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/indicator-links",
      ),
    );
    expect(JSON.parse(String(linkCall?.[1]?.body))).toMatchObject({
      selector: {
        kind: "graph_edge",
        edge_id: edgeId,
        field_key: "network_flow.src_ip",
        graph_query_digest: graphDigest,
      },
      target: {
        mode: "create_indicator",
        indicator_type: "ipv4_addr",
      },
      confirm_exact_value: "192.0.2.10",
    });
  });

  it("keeps the import entry role-gated", async () => {
    installNetworkFlowFetchMock({ tables: [] });

    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="viewer"
        incidentId="incident-1"
      />,
    );

    await screen.findByText("No active Network Flow tables.");
    expect(
      screen.queryByTestId(networkAnalysisTestId("import-trigger")),
    ).toBeNull();
  });
});

function installNetworkFlowFetchMock(options: { tables?: unknown[] } = {}) {
  const tables = options.tables ?? [tableResource()];
  const fetchSpy = vi.fn(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = requestURL(input);
      const method = init?.method ?? "GET";
      if (
        method === "GET" &&
        url.endsWith("/api/v1/incidents/incident-1/network-flow/tables")
      ) {
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_list.v1",
          tables,
          meta: { count: tables.length },
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/query`,
        )
      ) {
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_query_result.v1",
          network_flow_table_id: tableId,
          rows: [rowResource()],
          meta: {
            paging: {
              limit: 50,
              returned_count: 1,
              next_cursor_token: null,
            },
          },
        });
      }
      if (
        method === "POST" &&
        url.endsWith("/api/v1/incidents/incident-1/network-flow/graphs/query")
      ) {
        return jsonResponse(graphResource());
      }
      if (
        method === "POST" &&
        url.endsWith(
          "/api/v1/incidents/incident-1/network-flow/graphs/contributors/query",
        )
      ) {
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_contributor_query_result.v1",
          graph_query_digest: graphDigest,
          selector: { kind: "edge", edge_id: edgeId },
          contributors: [{ row_ref: rowRefResource(), row: rowResource() }],
          meta: {
            paging: {
              limit: 50,
              returned_count: 1,
              next_cursor_token: null,
            },
          },
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          "/api/v1/incidents/incident-1/network-flow/indicator-links",
        )
      ) {
        return jsonResponse({
          schema_id: "cartulary.network_flow.indicator_link_result.v1",
          duplicate: false,
          binding: {
            network_flow_indicator_binding_id:
              "nfb_33333333333333333333333333333333",
            target_indicator_ref: {
              indicator_id: "44444444-4444-4444-8444-444444444444",
              indicator_type: "ipv4_addr",
              normalized_value: "192.0.2.10",
            },
            source_row_refs_total_count: 1,
          },
        });
      }
      return jsonResponse({ error: { code: "unexpected_request" } }, 404);
    },
  );
  vi.stubGlobal("fetch", fetchSpy);
  return fetchSpy;
}

function tableResource() {
  return {
    network_flow_table_id: tableId,
    incident_id: "incident-1",
    display_name: "flows.csv",
    table_version: 1,
    table_status: "active",
    source_filename_display: "flows.csv",
    mapping_fingerprint: mappingFingerprint,
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_at: "2026-07-10T12:00:00Z",
    updated_at: "2026-07-10T12:00:00Z",
  };
}

function rowRefResource() {
  return {
    network_flow_table_id: tableId,
    network_flow_row_id: rowId,
    source_row_number: 2,
    mapping_fingerprint: mappingFingerprint,
  };
}

function rowResource() {
  return {
    network_flow_row_id: rowId,
    network_flow_table_id: tableId,
    incident_id: "incident-1",
    source_row_number: 2,
    mapping_fingerprint: mappingFingerprint,
    "network_flow.flow_start_utc": "2026-07-10T12:00:00Z",
    "network_flow.flow_end_utc": "2026-07-10T12:00:05Z",
    "network_flow.src_ip": "192.0.2.10",
    "network_flow.dst_ip": "192.0.2.20",
    "network_flow.src_port": 443,
    "network_flow.dst_port": 51515,
    "network_flow.ip_protocol": 6,
    "network_flow.bytes_count": "1234",
    "network_flow.packets_count": "12",
    "network_flow.exporter_id": null,
    "network_flow.input_interface": "Gi0/1",
    "network_flow.output_interface": "Gi0/2",
    "network_flow.application_label": null,
  };
}

function graphResource() {
  return {
    schema_id: "cartulary.network_flow.graph_query_result.v1",
    graph_query_digest: graphDigest,
    semantic_query: {
      schema_id: "cartulary.network_flow.graph_semantic_query.v1",
      selected_table_ids: [tableId],
      filters: [],
      time_range: { start_utc: null, end_utc: null },
      aggregation: {
        mode: "default_flow_edge_v1",
        include_example_row_refs: true,
      },
      result_limits: {
        max_vertices: 1000,
        max_edges: 1000,
        max_example_row_refs_per_edge: 10,
        max_aggregate_counter_digits: 20,
      },
    },
    graph_projection_result: {
      schema_id: "graph_projection.ephemeral_projection_result.v1",
      state: "ephemeral_available",
      ephemeral_projection_id: "projection-1",
      validation_summary: {
        fatal_count: 0,
        error_count: 0,
        warning_count: 0,
        info_count: 0,
        issues: [],
      },
    },
    edge_annotations: [
      {
        edge_id: edgeId,
        example_row_refs: [rowRefResource()],
        example_refs_truncated: false,
        example_refs_total_count: 1,
      },
    ],
    source_table_refs: [
      {
        network_flow_table_id: tableId,
        table_version: 1,
        mapping_fingerprint: mappingFingerprint,
        row_count_accepted: 1,
        row_count_rejected: 0,
      },
    ],
    result_limits: {
      max_vertices: 1000,
      max_edges: 1000,
      max_example_row_refs_per_edge: 10,
      max_aggregate_counter_digits: 20,
    },
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function requestURL(input: RequestInfo | URL): string {
  if (typeof input === "string") {
    return input;
  }
  if (input instanceof URL) {
    return input.toString();
  }
  return input.url;
}
