import {
  networkAnalysisEdgeTestId,
  networkAnalysisTableTabTestId,
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
const incidentResourceId = "11111111-1111-4111-8111-111111111111";
const sourceDigest =
  "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc";
const sourceRowDigest =
  "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd";

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

    const initialRowsCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith(
        `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/query`,
      ),
    );
    expect(JSON.parse(String(initialRowsCall?.[1]?.body))).toEqual({
      schema_id: "cartulary.network_flow.table_query_request.v1",
    });

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

  it("previews an ordinal-aware mapping and selects the returned table resource", async () => {
    const returnedTableId = "nft_22222222222222222222222222222222";
    const fetchSpy = installImportFlowFetchMock(returnedTableId);

    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );

    await screen.findAllByText("flows.csv");
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("import-input")),
      {
        target: {
          files: [
            new File(["header\nvalue\n"], "new-flows.csv", {
              type: "text/csv",
            }),
          ],
        },
      },
    );

    expect(
      await screen.findByTestId(networkAnalysisTestId("mapping-dialog")),
    ).toBeTruthy();
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("mapping-preview")),
    );
    expect(
      (
        await screen.findByTestId(
          networkAnalysisTestId("mapping-preview-summary"),
        )
      ).textContent,
    ).toContain("1 accepted");
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("mapping-apply")));

    await waitFor(() => {
      expect(
        screen
          .getByTestId(networkAnalysisTableTabTestId(returnedTableId))
          .getAttribute("aria-selected"),
      ).toBe("true");
    });
    const previewCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith("/mapping-preview"),
    );
    expect(JSON.parse(String(previewCall?.[1]?.body))).toEqual({
      target_kind: "network_flow_table",
      extension_profile_id: "network_flow_activity",
      owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
      owner_mapping: expect.objectContaining({
        source_profile_id: "cisco_sna_netflow_csv_v1",
      }),
    });
  });
});

function installImportFlowFetchMock(returnedTableId: string) {
  let tableListRequests = 0;
  const columns = importColumns();
  const fetchSpy = vi.fn(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = requestURL(input);
      const method = init?.method ?? "GET";
      if (
        method === "GET" &&
        url.endsWith("/api/v1/incidents/incident-1/network-flow/tables")
      ) {
        tableListRequests += 1;
        const tables =
          tableListRequests === 1
            ? [tableResource()]
            : [
                tableResource(),
                {
                  ...tableResource(),
                  network_flow_table_id: returnedTableId,
                  display_name: "new-flows.csv",
                  source_filename_display: "new-flows.csv",
                },
              ];
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_list.v1",
          tables,
          meta: { count: tables.length },
        });
      }
      if (method === "POST" && url.endsWith("/query")) {
        const responseTableId = url.includes(returnedTableId)
          ? returnedTableId
          : tableId;
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_query_result.v1",
          network_flow_table_id: responseTableId,
          rows: [],
          meta: {
            query: {
              filters: [],
              sort: [],
              effective_sort: [],
              table_ids: [responseTableId],
            },
            paging: {
              limit: 200,
              returned_count: 0,
              next_cursor_token: null,
            },
          },
        });
      }
      if (method === "POST" && url.endsWith("/api/v1/import-sessions")) {
        return jsonResponse({ data: importJob("upload-job") });
      }
      if (method === "GET" && url.endsWith("/api/v1/jobs/upload-job")) {
        return jsonResponse({
          data: {
            ...importJob("upload-job"),
            result_summary: {
              resource_refs: [{ kind: "import_session", id: "session-1" }],
            },
          },
        });
      }
      if (method === "GET" && url.endsWith("/session-1/units")) {
        return jsonResponse({
          data: {
            import_units: [
              { import_session_id: "session-1", import_unit_id: "unit-1" },
            ],
          },
        });
      }
      if (method === "GET" && url.endsWith("/unit-1/preview")) {
        return jsonResponse({
          data: {
            import_session_id: "session-1",
            import_unit_id: "unit-1",
            header_row_ref: 1,
            data_start_row_ref: 2,
            columns,
            preview_rows: [],
          },
        });
      }
      if (method === "POST" && url.endsWith("/mapping-preview")) {
        const request = JSON.parse(String(init?.body)) as {
          owner_mapping: Record<string, unknown>;
        };
        return jsonResponse({
          data: {
            schema_id: "cartulary.imports.extension_mapping_preview_result.v1",
            import_session_id: "session-1",
            import_unit_id: "unit-1",
            target_kind: "network_flow_table",
            extension_profile_id: "network_flow_activity",
            owner_result_schema_id:
              "cartulary.network_flow.import_preview_result.v1",
            owner_result: importPreviewResult(request.owner_mapping),
          },
        });
      }
      if (method === "PUT" && url.endsWith("/mapping")) {
        return jsonResponse({
          data: {
            import_session_id: "session-1",
            import_unit_id: "unit-1",
            mapping_fingerprint: mappingFingerprint,
          },
        });
      }
      if (method === "POST" && url.endsWith("/select")) {
        return jsonResponse({ data: { selected: true } });
      }
      if (method === "POST" && url.endsWith("/session-1/apply")) {
        return jsonResponse({ data: importJob("apply-job") });
      }
      if (method === "GET" && url.endsWith("/api/v1/jobs/apply-job")) {
        return jsonResponse({
          data: {
            ...importJob("apply-job"),
            result_summary: {
              resource_refs: [
                { kind: "network_flow_table", id: returnedTableId },
              ],
            },
          },
        });
      }
      return jsonResponse({ error: { code: "unexpected_request" } }, 404);
    },
  );
  vi.stubGlobal("fetch", fetchSpy);
  return fetchSpy;
}

function importColumns() {
  return [
    "Flow Start Time",
    "Flow End Time",
    "Source IP Address",
    "Destination IP Address",
    "Source Port",
    "Destination Port",
    "Protocol",
    "Bytes",
    "Packets",
  ].map((sourceHeaderText, index) => ({
    source_column_ordinal: index + 1,
    source_header_text: sourceHeaderText,
  }));
}

function importPreviewResult(ownerMapping: Record<string, unknown>) {
  const sourceColumns = importColumns().map((column) => ({
    source_column_ordinal: column.source_column_ordinal,
    raw_header_text: column.source_header_text,
    normalized_header_for_suggestion: column.source_header_text.toLowerCase(),
    raw_header_sha256: sourceDigest,
    sample_values: [{ safe_sample: "sample", raw_value_sha256: sourceDigest }],
    detected_empty_count: 0,
  }));
  return {
    schema_id: "cartulary.network_flow.import_preview_result.v1",
    source_content_sha256: sourceDigest,
    source_columns: sourceColumns,
    materialized_mapping: { ...ownerMapping, source_columns: sourceColumns },
    mapping_fingerprint: mappingFingerprint,
    preview_record_count: 1,
    preview_accepted_count: 1,
    preview_rejected_count: 0,
    diagnostics: [],
    diagnostics_truncated: false,
  };
}

function importJob(jobId: string) {
  return { job_id: jobId, status: "succeeded", result_summary: null };
}

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
            query: {
              filters: [],
              sort: [],
              effective_sort: [],
              table_ids: [tableId],
            },
            paging: {
              limit: 200,
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
            incident_id: incidentResourceId,
            target_indicator_ref: {
              indicator_id: "44444444-4444-4444-8444-444444444444",
              indicator_type: "ipv4_addr",
              value_kind: "atomic",
              normalized_value: "192.0.2.10",
            },
            selector_kind: "graph_edge",
            candidate_value: "192.0.2.10",
            source_row_refs: [rowRefResource()],
            source_row_refs_truncated: false,
            source_row_refs_total_count: 1,
            created_observation_refs: [],
            created_by_user_id: "user-1",
            created_at: "2026-07-10T12:00:00Z",
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
    incident_id: incidentResourceId,
    display_name: "flows.csv",
    table_version: 1,
    table_status: "active",
    source_import_session_id: "import-session-1",
    source_import_unit_id: "import-unit-1",
    source_content_sha256: sourceDigest,
    source_filename_display: "flows.csv",
    source_filename_digest: sourceDigest,
    source_filename_digest_key_id: "filename-key-1",
    mapping_fingerprint: mappingFingerprint,
    source_profile_id: "cisco_sna_netflow_csv_v1",
    parser_profile_id: "rfc4180_headered_csv_v1",
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_by_user_id: "user-1",
    created_at: "2026-07-10T12:00:00Z",
    updated_at: "2026-07-10T12:00:00Z",
    deleted_at: null,
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
    incident_id: incidentResourceId,
    source_row_number: 2,
    source_row_digest_sha256: sourceRowDigest,
    normalized_row_digest_sha256: sourceRowDigest,
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
    "network_flow.tcp_flags": null,
    "network_flow.application_label": null,
    unmapped_raw: {},
    "network_flow.observation_source_ref": {
      import_session_id: "import-session-1",
      import_unit_id: "import-unit-1",
      source_content_sha256: sourceDigest,
      source_profile_id: "cisco_sna_netflow_csv_v1",
      parser_profile_id: "rfc4180_headered_csv_v1",
      mapping_fingerprint: mappingFingerprint,
      source_row_number: 2,
      source_row_digest_sha256: sourceRowDigest,
    },
    created_at: "2026-07-10T12:00:00Z",
    created_by_user_id: "user-1",
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
      projection_schema_id: "graph_projection.v1",
      graph_view_id:
        "gv_0000000000000000000000000000000000000000000000000000000000000000",
      graph_view_key: "network-flow-test",
      state: "ephemeral_available",
      ephemeral_projection_id:
        "gpe_0000000000000000000000000000000000000000000000000000000000000000",
      source_snapshot_id: "snapshot-1",
      projection_version: "v1",
      generated_at: "2026-05-30T00:00:00Z",
      properties: {},
      metadata: {
        previous_projection_run_id: null,
        projection_config_digest: graphDigest,
        projection_source_digest: sourceDigest,
        mapped_metadata: {},
        invalidation: null,
      },
      schema_registry: {
        vertex_kinds: [],
        edge_kinds: [],
        property_keys: [],
        metadata_keys: [],
      },
      vertices: [],
      edges: [],
      validation_summary: {
        status: "valid",
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
