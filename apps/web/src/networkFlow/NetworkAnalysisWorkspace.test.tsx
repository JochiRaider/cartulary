import { networkFlowDecoders } from "@cartulary/protocol-ts/network-flow";
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
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ExtensionAvailabilityProvider } from "../extensions/ExtensionAvailabilityContext";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { NetworkAnalysisWorkspace as ProductionNetworkAnalysisWorkspace } from "./NetworkAnalysisWorkspace";

function NetworkAnalysisWorkspace(
  props: ComponentProps<typeof ProductionNetworkAnalysisWorkspace>,
) {
  return (
    <ExtensionAvailabilityProvider
      controller={readyExtensionAvailability(props.incidentId)}
    >
      <ProductionNetworkAnalysisWorkspace {...props} />
    </ExtensionAvailabilityProvider>
  );
}

const tableId = "nft_11111111111111111111111111111111";
const rowId =
  "nfr_1111111111111111111111111111111111111111111111111111111111111111";
const edgeId =
  "nff_2222222222222222222222222222222222222222222222222222222222222222";
const temporalEdgeId =
  "nfbe_9999999999999999999999999999999999999999999999999999999999999999";
const srcEndpointId =
  "nfe_4444444444444444444444444444444444444444444444444444444444444444";
const dstEndpointId =
  "nfe_5555555555555555555555555555555555555555555555555555555555555555";
const projectedSrcVertexId =
  "vx_6666666666666666666666666666666666666666666666666666666666666666";
const projectedDstVertexId =
  "vx_7777777777777777777777777777777777777777777777777777777777777777";
const projectedEdgeId =
  "ed_8888888888888888888888888888888888888888888888888888888888888888";
const graphViewId = "nfgv_99999999999999999999999999999999";
const projectionResultId =
  "gpres_0000000000000000000000000000000000000000000000000000000000000000";
const diagnosticId =
  "nfd_3333333333333333333333333333333333333333333333333333333333333333";
const mappingFingerprint =
  "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
const graphDigest =
  "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
const incidentResourceId = "11111111-1111-4111-8111-111111111111";
const importSessionId = "22222222-2222-4222-8222-222222222222";
const importUnitId = "33333333-3333-4333-8333-333333333333";
const importActorId = "44444444-4444-4444-8444-444444444444";
const uploadImportJobId = "55555555-5555-4555-8555-555555555555";
const applyImportJobId = "66666666-6666-4666-8666-666666666666";
const sourceDigest =
  "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc";
const sourceRowDigest =
  "dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd";

describe("NetworkAnalysisWorkspace", () => {
  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it("Verify production Network Flow grids, read-only behavior, stable references, virtualization, and adapter boundaries compose through semantic identities.", async () => {
    const decodedGraph = networkFlowDecoders.graphQueryResult.decode(
      graphResource(),
    );
    if (!decodedGraph.ok) {
      throw new Error(JSON.stringify(decodedGraph.error));
    }
    const decodedSourceProfiles = networkFlowDecoders.sourceProfileList.decode(
      sourceProfileListResource(),
    );
    if (!decodedSourceProfiles.ok) {
      throw new Error(JSON.stringify(decodedSourceProfiles.error));
    }
    const fetchSpy = installNetworkFlowFetchMock({
      contributorNextCursor: "contributor-cursor-2",
    });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );

    expect(await screen.findAllByText("flows.csv")).toHaveLength(3);
    expect(await screen.findByText("192.0.2.10")).toBeTruthy();
    expect(
      screen.getByTestId(networkAnalysisTestId("workspace-header")),
    ).toBeTruthy();
    expect(
      screen.getByTestId(networkAnalysisTestId("diagnostics-summary")),
    ).toBeTruthy();
    expect(screen.getByTestId(networkAnalysisTestId("filters"))).toBeTruthy();
    expect(
      screen.getByTestId(networkAnalysisTestId("accepted-grid")),
    ).toBeTruthy();
    fireEvent.click(screen.getByText("Columns"));
    fireEvent.click(screen.getByLabelText("Exporter"));
    expect(
      screen.getByRole("columnheader", { name: /Exporter/u }),
    ).toBeTruthy();
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("layout-reset")));
    expect(
      screen.queryByRole("columnheader", { name: /Exporter/u }),
    ).toBeNull();
    fireEvent.click(
      screen.getByRole("gridcell", { name: /Source IP: 192\.0\.2\.10/u }),
    );
    const inspector = await screen.findByTestId(
      networkAnalysisTestId("inspector"),
    );
    expect(inspector).toBeTruthy();
    await waitFor(() => {
      const currentRowLinkButton = screen.getByRole("button", {
        name: /Link \d+ selected row|Select one IP cell/u,
      });
      expect(currentRowLinkButton.textContent).toContain("Link 1 selected row");
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Link 1 selected row" }),
    );
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(
          networkAnalysisTestId("indicator-link-confirmation"),
        ),
      ),
    );
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("indicator-link-confirmation")),
      { target: { value: "192.0.2.10" } },
    );
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("indicator-link-submit")),
    );
    await waitFor(() => {
      expect(indicatorLinkRequestBodies(fetchSpy)).toHaveLength(1);
    });
    expect(indicatorLinkRequestBodies(fetchSpy)[0]).toMatchObject({
      selector: {
        kind: "row_field_value",
        network_flow_table_id: tableId,
        network_flow_row_id: rowId,
        field_key: "network_flow.src_ip",
      },
      confirm_exact_value: "192.0.2.10",
    });

    const initialRowsCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith(
        `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/query`,
      ),
    );
    expect(JSON.parse(String(initialRowsCall?.[1]?.body))).toEqual({
      schema_id: "cartulary.network_flow.table_query_request.v1",
    });

    fireEvent.change(screen.getByLabelText("Endpoint IP operator"), {
      target: { value: "cidr_contains" },
    });
    fireEvent.change(screen.getByLabelText("Endpoint IP value"), {
      target: { value: "192.0.2.0/24" },
    });
    fireEvent.change(screen.getByLabelText("Flow overlap starts at"), {
      target: { value: "2026-07-10T00:00:00Z" },
    });
    fireEvent.change(screen.getByLabelText("Flow overlap ends before"), {
      target: { value: "2026-07-11T00:00:00Z" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply query" }));

    await waitFor(() => {
      const rowQueries = fetchSpy.mock.calls.filter(([input]) =>
        requestURL(input).endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/query`,
        ),
      );
      expect(rowQueries.length).toBeGreaterThanOrEqual(2);
      expect(JSON.parse(String(rowQueries.at(-1)?.[1]?.body))).toEqual({
        schema_id: "cartulary.network_flow.table_query_request.v1",
        filters: [
          {
            field_key: "network_flow.endpoint_ip",
            op: "cidr_contains",
            value: "192.0.2.0/24",
          },
          {
            field_key: "network_flow.flow_end_utc",
            op: "range",
            value: { gte: "2026-07-10T00:00:00Z", lt: null },
          },
          {
            field_key: "network_flow.flow_start_utc",
            op: "range",
            value: { gte: null, lt: "2026-07-11T00:00:00Z" },
          },
        ],
      });
    });

    fireEvent.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));

    expect(
      await screen.findByTestId(networkAnalysisEdgeTestId(edgeId)),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(networkAnalysisTestId("contributor-drawer")),
    ).toBeNull();
    expect(
      fetchSpy.mock.calls.some(([input]) =>
        requestURL(input).endsWith(
          "/api/v1/incidents/incident-1/network-flow/graphs/contributors/query",
        ),
      ),
    ).toBe(false);

    const graphCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/graphs/query",
      ),
    );
    expect(JSON.parse(String(graphCall?.[1]?.body))).toEqual({
      schema_id: "cartulary.network_flow.graph_query_request.v2",
      table_scope: { mode: "active_table", active_table_id: tableId },
      filters: [
        {
          field_key: "network_flow.endpoint_ip",
          op: "cidr_contains",
          value: "192.0.2.0/24",
        },
      ],
      time_range: {
        start_utc: "2026-07-10T00:00:00Z",
        end_utc: "2026-07-11T00:00:00Z",
      },
      aggregation: {
        mode: "default_flow_edge_v1",
        include_example_row_refs: true,
      },
    });

    fireEvent.click(screen.getByLabelText("Selected tables"));
    await waitFor(() => {
      expect(graphRequestBodies(fetchSpy).at(-1)?.table_scope).toEqual({
        mode: "selected_tables",
        selected_table_ids: [tableId],
      });
    });
    fireEvent.click(screen.getByLabelText("All active tables"));
    await waitFor(() => {
      expect(graphRequestBodies(fetchSpy).at(-1)?.table_scope).toEqual({
        mode: "all_active_tables",
      });
    });
    fireEvent.click(screen.getByLabelText("Active table"));
    await screen.findByTestId(networkAnalysisEdgeTestId(edgeId));

    const selectEdgeButton = screen.getByRole("button", {
      name: "Select edge",
    });
    fireEvent.click(selectEdgeButton);
    expect(
      await screen.findByTestId(networkAnalysisTestId("contributor-drawer")),
    ).toBeTruthy();
    expect(
      await screen.findByTestId(networkAnalysisTestId("contributor-grid")),
    ).toBeTruthy();

    const contributorCall = fetchSpy.mock.calls.find(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/graphs/contributors/query",
      ),
    );
    expect(JSON.parse(String(contributorCall?.[1]?.body))).toEqual({
      schema_id: "cartulary.network_flow.graph_contributor_query_request.v2",
      graph_query: graphSemanticQueryResource(),
      graph_query_digest: graphDigest,
      selector: graphDefaultEdgeSelectorResource(),
      limit: 500,
    });
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("page-next")));
    await waitFor(() => {
      expect(contributorRequestBodies(fetchSpy).at(-1)).toEqual({
        schema_id:
          "cartulary.network_flow.graph_contributor_query_continuation.v1",
        cursor_token: "contributor-cursor-2",
      });
    });
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("page-previous")));
    await waitFor(() => {
      expect(contributorRequestBodies(fetchSpy).at(-1)).toEqual({
        schema_id: "cartulary.network_flow.graph_contributor_query_request.v2",
        graph_query: graphSemanticQueryResource(),
        graph_query_digest: graphDigest,
        selector: graphDefaultEdgeSelectorResource(),
        limit: 500,
      });
    });

    fireEvent.click(screen.getByRole("button", { name: /link source/i }));
    expect(
      screen.getByTestId(networkAnalysisTestId("indicator-link-dialog")),
    ).toBeTruthy();
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("indicator-link-confirmation")),
      { target: { value: "192.0.2.10" } },
    );
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("indicator-link-submit")),
    );

    await waitFor(() => {
      expect(
        fetchSpy.mock.calls.some(([input, init]) => {
          const url = requestURL(input);
          return (
            url.endsWith(
              "/api/v1/incidents/incident-1/network-flow/indicator-links",
            ) &&
            init?.method === "POST" &&
            (
              JSON.parse(String(init.body)) as {
                selector: { kind: string };
              }
            ).selector.kind === "graph_edge"
          );
        }),
      ).toBe(true);
    });

    const linkCall = fetchSpy.mock.calls.find(
      ([input, init]) =>
        requestURL(input).endsWith(
          "/api/v1/incidents/incident-1/network-flow/indicator-links",
        ) &&
        (JSON.parse(String(init?.body)) as { selector: { kind: string } })
          .selector.kind === "graph_edge",
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

    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("contributor-close")),
    );
    await waitFor(() => expect(document.activeElement).toBe(selectEdgeButton));
    fireEvent.click(
      screen.getAllByRole("button", { name: "Select vertex" })[0] as Element,
    );
    expect(
      await screen.findByTestId(networkAnalysisTestId("contributor-drawer")),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Link vertex" }));
    fireEvent.click(screen.getByLabelText("Existing Indicator"));
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("indicator-link-existing-id")),
      {
        target: {
          value: "44444444-4444-4444-8444-444444444444",
        },
      },
    );
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("indicator-link-confirmation")),
      { target: { value: "192.0.2.10" } },
    );
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("indicator-link-submit")),
    );

    await waitFor(() => {
      expect(indicatorLinkRequestBodies(fetchSpy)).toHaveLength(3);
    });
    expect(indicatorLinkRequestBodies(fetchSpy).at(-1)).toMatchObject({
      selector: {
        kind: "graph_vertex",
        graph_query_digest: graphDigest,
        vertex_id: srcEndpointId,
      },
      target: {
        mode: "existing_indicator",
        indicator_id: "44444444-4444-4444-8444-444444444444",
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

  it("enforces the viewer, editor, reviewer, and admin command matrix", async () => {
    const expectations = [
      {
        role: "viewer" as const,
        canImport: false,
        canRename: false,
        canDelete: false,
      },
      {
        role: "editor" as const,
        canImport: true,
        canRename: true,
        canDelete: false,
      },
      {
        role: "reviewer" as const,
        canImport: true,
        canRename: true,
        canDelete: true,
      },
      {
        role: "admin" as const,
        canImport: true,
        canRename: true,
        canDelete: true,
      },
    ];

    for (const expectation of expectations) {
      const fetchSpy = installNetworkFlowFetchMock();
      const rendered = render(
        <NetworkAnalysisWorkspace
          currentIncidentRole={expectation.role}
          incidentId="incident-1"
        />,
      );
      await screen.findByTestId(networkAnalysisTestId("accepted-grid"));
      expect(
        screen.queryByTestId(networkAnalysisTestId("import-trigger")) !== null,
      ).toBe(expectation.canImport);
      expect(
        screen.queryByTestId(networkAnalysisTestId("rename-trigger")) !== null,
      ).toBe(expectation.canRename);
      expect(
        screen.queryByTestId(networkAnalysisTestId("delete-trigger")) !== null,
      ).toBe(expectation.canDelete);

      fireEvent.paste(
        screen.getByTestId(networkAnalysisTestId("accepted-grid")),
        { clipboardData: { getData: () => "replacement" } },
      );
      expect(
        fetchSpy.mock.calls.some(
          ([, init]) => init?.method === "PATCH" || init?.method === "DELETE",
        ),
      ).toBe(false);
      rendered.unmount();
    }
  });

  it("removes mutation surfaces immediately when the role is downgraded", async () => {
    installNetworkFlowFetchMock();
    const rendered = render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );
    await screen.findByTestId(networkAnalysisTestId("accepted-grid"));
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("rename-trigger")),
    );
    expect(
      screen.getByTestId(networkAnalysisTestId("rename-dialog")),
    ).toBeTruthy();

    rendered.rerender(
      <NetworkAnalysisWorkspace
        currentIncidentRole="viewer"
        incidentId="incident-1"
      />,
    );

    await waitFor(() => {
      expect(
        screen.queryByTestId(networkAnalysisTestId("rename-dialog")),
      ).toBeNull();
    });
    expect(
      screen.queryByTestId(networkAnalysisTestId("rename-trigger")),
    ).toBeNull();
    expect(
      screen.queryByTestId(networkAnalysisTestId("delete-trigger")),
    ).toBeNull();
    expect(
      screen.getByTestId(networkAnalysisTestId("accepted-grid")),
    ).toBeTruthy();
  });

  it("renames by stable table identity and soft-deletes only after exact confirmation", async () => {
    const fetchSpy = installNetworkFlowFetchMock();
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );
    await screen.findByTestId(networkAnalysisTestId("accepted-grid"));

    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("rename-trigger")),
    );
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("rename-input")),
      {
        target: { value: "Renamed flows" },
      },
    );
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("rename-submit")));
    await screen.findAllByText("Renamed flows");

    const renameCall = fetchSpy.mock.calls.find(
      ([input, init]) =>
        requestURL(input).endsWith(`/network-flow/tables/${tableId}`) &&
        init?.method === "PATCH",
    );
    expect(JSON.parse(String(renameCall?.[1]?.body))).toMatchObject({
      base_table_version: 1,
      display_name: "Renamed flows",
    });

    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("delete-trigger")),
    );
    const confirm = screen.getByTestId(
      networkAnalysisTestId("delete-confirm"),
    ) as HTMLButtonElement;
    expect(confirm.disabled).toBe(true);
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("delete-confirmation")),
      { target: { value: "Renamed flows" } },
    );
    expect(confirm.disabled).toBe(false);
    fireEvent.click(confirm);

    await waitFor(() => {
      expect(
        screen.queryByTestId(networkAnalysisTableTabTestId(tableId)),
      ).toBeNull();
      expect(
        screen.queryByTestId(networkAnalysisTestId("accepted-grid")),
      ).toBeNull();
    });
    const deleteCall = fetchSpy.mock.calls.find(
      ([input, init]) =>
        requestURL(input).endsWith(`/network-flow/tables/${tableId}`) &&
        init?.method === "DELETE",
    );
    expect(JSON.parse(String(deleteCall?.[1]?.body))).toMatchObject({
      base_table_version: 2,
    });
  });

  it("refreshes table metadata after a rename version conflict before retry", async () => {
    const fetchSpy = installNetworkFlowFetchMock({ renameConflictOnce: true });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );
    await screen.findByTestId(networkAnalysisTestId("accepted-grid"));
    fireEvent.click(
      screen.getByTestId(networkAnalysisTestId("rename-trigger")),
    );
    fireEvent.change(
      screen.getByTestId(networkAnalysisTestId("rename-input")),
      {
        target: { value: "Analyst flows" },
      },
    );
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("rename-submit")));

    await screen.findAllByText("server-flows.csv");
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("rename-submit")));
    await screen.findAllByText("Analyst flows");

    const renameBodies = fetchSpy.mock.calls
      .filter(
        ([input, init]) =>
          requestURL(input).endsWith(`/network-flow/tables/${tableId}`) &&
          init?.method === "PATCH",
      )
      .map(
        ([, init]) => JSON.parse(String(init?.body)) as Record<string, unknown>,
      );
    expect(renameBodies).toHaveLength(2);
    expect(renameBodies.map((body) => body.base_table_version)).toEqual([1, 2]);
  });

  it("clears protected data immediately when authorization is lost", async () => {
    const onIncidentAccessLost = vi.fn();
    installNetworkFlowFetchMock({ rowFailureAfter: 1 });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="viewer"
        incidentId="incident-1"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );
    await screen.findByText("192.0.2.10");
    fireEvent.change(screen.getByLabelText("Endpoint IP value"), {
      target: { value: "192.0.2.10" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply query" }));

    await screen.findByRole("region", {
      name: "Network Analysis permission state",
    });
    expect(screen.queryByText("192.0.2.10")).toBeNull();
    expect(
      screen.queryByTestId(networkAnalysisTestId("accepted-grid")),
    ).toBeNull();
    expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
  });

  it("clears table-scoped state when the active table becomes inactive", async () => {
    installNetworkFlowFetchMock({
      removeTablesOnRowFailure: true,
      rowFailureAfter: 1,
      rowFailureCode: "network_flow_table_not_active",
      rowFailureReason: "soft_deleted",
      rowFailureStatus: 409,
    });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );
    await screen.findByText("192.0.2.10");
    fireEvent.change(screen.getByLabelText("Endpoint IP value"), {
      target: { value: "192.0.2.10" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Apply query" }));

    await screen.findByRole("region", {
      name: "Network Analysis lifecycle state",
    });
    expect(screen.queryByText("192.0.2.10")).toBeNull();
    expect(screen.queryByTestId(networkAnalysisTestId("inspector"))).toBeNull();
  });

  it("renders rejected diagnostics through the semantic grid and applies owner filters", async () => {
    const fetchSpy = installNetworkFlowFetchMock();
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );

    await screen.findAllByText("flows.csv");
    fireEvent.click(screen.getByTestId(networkAnalysisTestId("mode-rejected")));

    expect(
      await screen.findByTestId(networkAnalysisTestId("rejected-grid")),
    ).toBeTruthy();
    expect(
      await screen.findByText("The value is not a valid IP address."),
    ).toBeTruthy();
    fireEvent.change(screen.getByLabelText("Error codes"), {
      target: { value: "network_flow_invalid_ip" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "Apply diagnostics query" }),
    );

    await waitFor(() => {
      const matching = fetchSpy.mock.calls.filter(([input]) =>
        requestURL(input).endsWith("/rejected-rows/query"),
      );
      expect(matching.length).toBeGreaterThanOrEqual(2);
      expect(JSON.parse(String(matching.at(-1)?.[1]?.body))).toEqual({
        schema_id: "cartulary.network_flow.rejected_rows_query_request.v1",
        error_codes: ["network_flow_invalid_ip"],
      });
    });
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
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(networkAnalysisTestId("mapping-profile")),
      ),
    );
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

  it("supports keyboard table navigation and restores focus from lifecycle dialogs", async () => {
    const user = userEvent.setup();
    const secondTableId = "nft_22222222222222222222222222222222";
    installNetworkFlowFetchMock({
      tables: [
        tableResource(),
        {
          ...tableResource(),
          network_flow_table_id: secondTableId,
          display_name: "second-flows.csv",
          source_filename_display: "second-flows.csv",
          created_at: "2026-07-10T12:01:00Z",
        },
      ],
    });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="admin"
        incidentId="incident-1"
      />,
    );

    const firstTab = await screen.findByTestId(
      networkAnalysisTableTabTestId(tableId),
    );
    const secondTab = screen.getByTestId(
      networkAnalysisTableTabTestId(secondTableId),
    );
    firstTab.focus();
    await user.keyboard("{ArrowRight}");
    expect(document.activeElement).toBe(secondTab);
    expect(secondTab.getAttribute("aria-selected")).toBe("true");
    await user.keyboard("{Home}");
    expect(document.activeElement).toBe(firstTab);
    expect(firstTab.getAttribute("aria-selected")).toBe("true");

    const renameTrigger = screen.getByTestId(
      networkAnalysisTestId("rename-trigger"),
    );
    renameTrigger.focus();
    await user.keyboard("{Enter}");
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(networkAnalysisTestId("rename-input")),
      ),
    );
    await user.keyboard("{Shift>}{Tab}{/Shift}");
    expect(document.activeElement).toBe(
      screen.getByTestId(networkAnalysisTestId("rename-submit")),
    );
    await user.keyboard("{Tab}");
    expect(document.activeElement).toBe(
      screen.getByTestId(networkAnalysisTestId("rename-input")),
    );
    await user.keyboard("{Escape}");
    await waitFor(() => expect(document.activeElement).toBe(renameTrigger));

    const deleteTrigger = screen.getByTestId(
      networkAnalysisTestId("delete-trigger"),
    );
    deleteTrigger.focus();
    await user.keyboard("{Enter}");
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByTestId(networkAnalysisTestId("delete-confirmation")),
      ),
    );
    await user.keyboard("{Escape}");
    await waitFor(() => expect(document.activeElement).toBe(deleteTrigger));
  });

  it("completes the saved graph lifecycle with exact results, last-safe refresh, contributors, bounded rendering, and role-aware controls", async () => {
    const user = userEvent.setup();
    const fetchSpy = installNetworkFlowFetchMock({
      savedGraphs: [savedGraphResource()],
    });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="admin"
        incidentId="incident-1"
      />,
    );

    await screen.findByTestId(networkAnalysisTableTabTestId(tableId));
    await user.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));
    expect(await screen.findByText("Graph ready")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Saved graphs" }));
    const savedGraphPanel = screen.getByRole("region", {
      name: "Saved Network Flow graphs",
    });

    expect(
      await screen.findByRole("heading", { name: "Investigation graph" }),
    ).toBeTruthy();
    expect(await screen.findByText(/Result gpres_/u)).toBeTruthy();
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-vertex-/u),
    ).toHaveLength(2);
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-edge-/u),
    ).toHaveLength(1);

    await user.click(screen.getByRole("button", { name: "192.0.2.10" }));
    expect(
      await screen.findByRole("complementary", {
        name: "Saved graph contributors",
      }),
    ).toBeTruthy();
    expect(await screen.findByText(/Row 2/u)).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Close" }));

    await user.click(
      within(savedGraphPanel).getByRole("button", { name: "Rename" }),
    );
    const renameInput = screen.getByRole("textbox", { name: "Display name" });
    await user.clear(renameInput);
    await user.type(renameInput, "Renamed investigation");
    await user.click(screen.getByRole("button", { name: "Rename graph" }));
    expect(
      await screen.findByRole("heading", { name: "Renamed investigation" }),
    ).toBeTruthy();

    await user.click(
      within(savedGraphPanel).getByRole("button", { name: "Refresh" }),
    );
    expect(
      await screen.findByText(
        "Showing the last successful result while refresh continues.",
      ),
    ).toBeTruthy();
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-vertex-/u),
    ).toHaveLength(2);

    await user.click(
      within(savedGraphPanel).getByRole("button", { name: "Retire" }),
    );
    await user.click(screen.getByRole("button", { name: "Retire graph" }));
    expect(await screen.findByText("No saved graphs yet.")).toBeTruthy();

    await user.click(
      screen.getByRole("button", { name: "Save current graph" }),
    );
    await user.type(
      screen.getByRole("textbox", { name: "Display name" }),
      "New saved graph",
    );
    await user.click(screen.getByRole("button", { name: "Save graph" }));
    expect(
      await screen.findByRole("heading", { name: "New saved graph" }),
    ).toBeTruthy();
    expect(await screen.findByText("Materialization queued.")).toBeTruthy();

    const savedGraphCalls = fetchSpy.mock.calls.filter(([input]) =>
      requestURL(input).includes("/network-flow/graph-views"),
    );
    expect(savedGraphCalls.some(([, init]) => init?.method === "PATCH")).toBe(
      true,
    );
    expect(
      savedGraphCalls.some(
        ([input, init]) =>
          (init?.method ?? "GET") === "POST" &&
          requestURL(input).endsWith("/refresh"),
      ),
    ).toBe(true);
    expect(savedGraphCalls.some(([, init]) => init?.method === "DELETE")).toBe(
      true,
    );
  });

  it("validates temporal ranges, echoes canonical bucket selectors, and navigates empty buckets without mounting unrelated objects", async () => {
    const user = userEvent.setup();
    const fetchSpy = installNetworkFlowFetchMock();
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );

    await screen.findByTestId(networkAnalysisTableTabTestId(tableId));
    await user.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));
    await user.click(screen.getByLabelText("Time buckets"));
    expect(
      screen.getByRole("alert", {
        name: "",
      }).textContent,
    ).toContain("require both UTC range bounds");

    await user.type(
      screen.getByLabelText("Flow overlap starts at"),
      "2026-07-10T00:00:00Z",
    );
    await user.type(
      screen.getByLabelText("Flow overlap ends before"),
      "2026-07-10T02:00:00Z",
    );
    await user.click(
      screen.getByTestId(networkAnalysisTestId("accepted-query-apply")),
    );

    expect(
      await screen.findByRole("navigation", {
        name: "Time bucket navigation",
      }),
    ).toBeTruthy();
    expect(screen.getByText("Bucket 1 of 2")).toBeTruthy();
    expect(
      screen.getByTestId(networkAnalysisEdgeTestId(temporalEdgeId)),
    ).toBeTruthy();
    expect(graphRequestBodies(fetchSpy).at(-1)).toMatchObject({
      schema_id: "cartulary.network_flow.graph_query_request.v2",
      aggregation: {
        mode: "time_bucket_v1",
        bucket_width_seconds: 3600,
      },
      time_range: {
        start_utc: "2026-07-10T00:00:00Z",
        end_utc: "2026-07-10T02:00:00Z",
      },
    });

    await user.click(screen.getByRole("button", { name: "Select edge" }));
    await waitFor(() => {
      expect(contributorRequestBodies(fetchSpy).at(-1)).toMatchObject({
        schema_id: "cartulary.network_flow.graph_contributor_query_request.v2",
        selector: {
          kind: "time_bucket_edge",
          source_edge_id: temporalEdgeId,
          bucket_start_utc: "2026-07-10T00:00:00.000Z",
          bucket_end_utc: "2026-07-10T01:00:00.000Z",
        },
      });
    });
    await user.click(screen.getByRole("button", { name: "Next bucket" }));
    expect(screen.getByText("Bucket 2 of 2")).toBeTruthy();
    expect(screen.getByText(/0 vertices · 0 edges · 0 rows/u)).toBeTruthy();
    expect(
      screen.queryByTestId(networkAnalysisEdgeTestId(temporalEdgeId)),
    ).toBeNull();

    await user.selectOptions(screen.getByLabelText("Bucket width"), "300");
    await waitFor(() => {
      expect(graphRequestBodies(fetchSpy).at(-1)).toMatchObject({
        aggregation: {
          mode: "time_bucket_v1",
          bucket_width_seconds: 300,
        },
      });
    });
  });

  it("enforces viewer, editor, reviewer, and administrator controls for saved graphs", async () => {
    const user = userEvent.setup();
    installNetworkFlowFetchMock({ savedGraphs: [savedGraphResource()] });
    const rendered = render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="viewer"
        incidentId="incident-1"
      />,
    );

    await screen.findByTestId(networkAnalysisTableTabTestId(tableId));
    await user.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));
    await user.click(screen.getByRole("button", { name: "Saved graphs" }));
    const savedGraphPanel = screen.getByRole("region", {
      name: "Saved Network Flow graphs",
    });
    const savedControls = () => within(savedGraphPanel);

    expect(
      savedControls().queryByRole("button", { name: "Save current graph" }),
    ).toBeNull();
    expect(
      savedControls().queryByRole("button", { name: "Rename" }),
    ).toBeNull();
    expect(
      savedControls().queryByRole("button", { name: "Refresh" }),
    ).toBeNull();
    expect(
      savedControls().queryByRole("button", { name: "Retire" }),
    ).toBeNull();

    rendered.rerender(
      <NetworkAnalysisWorkspace
        currentIncidentRole="editor"
        incidentId="incident-1"
      />,
    );
    await screen.findByRole("heading", { name: "Investigation graph" });
    expect(
      savedControls().getByRole("button", { name: "Save current graph" }),
    ).toBeTruthy();
    expect(
      savedControls().getByRole("button", { name: "Rename" }),
    ).toBeTruthy();
    expect(
      savedControls().getByRole("button", { name: "Refresh" }),
    ).toBeTruthy();
    expect(
      savedControls().queryByRole("button", { name: "Retire" }),
    ).toBeNull();

    rendered.rerender(
      <NetworkAnalysisWorkspace
        currentIncidentRole="reviewer"
        incidentId="incident-1"
      />,
    );
    await screen.findByRole("heading", { name: "Investigation graph" });
    expect(
      savedControls().queryByRole("button", { name: "Save current graph" }),
    ).toBeNull();
    expect(
      savedControls().queryByRole("button", { name: "Rename" }),
    ).toBeNull();
    expect(
      savedControls().queryByRole("button", { name: "Refresh" }),
    ).toBeNull();
    expect(
      savedControls().getByRole("button", { name: "Retire" }),
    ).toBeTruthy();

    rendered.rerender(
      <NetworkAnalysisWorkspace
        currentIncidentRole="admin"
        incidentId="incident-1"
      />,
    );
    await screen.findByRole("heading", { name: "Investigation graph" });
    expect(
      savedControls().getByRole("button", { name: "Save current graph" }),
    ).toBeTruthy();
    expect(
      savedControls().getByRole("button", { name: "Rename" }),
    ).toBeTruthy();
    expect(
      savedControls().getByRole("button", { name: "Refresh" }),
    ).toBeTruthy();
    expect(
      savedControls().getByRole("button", { name: "Retire" }),
    ).toBeTruthy();
  });

  it("bounds saved graph rendering to 500 vertices and 1000 edges with paged navigation", async () => {
    const user = userEvent.setup();
    installNetworkFlowFetchMock({
      savedGraphProjectionResult: largeSavedGraphProjectionResult(),
      savedGraphs: [savedGraphResource()],
    });
    render(
      <NetworkAnalysisWorkspace
        currentIncidentRole="viewer"
        incidentId="incident-1"
      />,
    );

    await screen.findByTestId(networkAnalysisTableTabTestId(tableId));
    await user.click(screen.getByTestId(networkAnalysisTestId("mode-graph")));
    await user.click(screen.getByRole("button", { name: "Saved graphs" }));

    expect(
      await screen.findAllByTestId(/^network-flow-saved-graph-vertex-/u),
    ).toHaveLength(500);
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-edge-/u),
    ).toHaveLength(1_000);
    expect(
      screen.getByText(
        "Large results are paged: at most 500 vertices and 1000 edges are mounted at once.",
      ),
    ).toBeTruthy();
    await user.click(
      within(
        screen.getByRole("navigation", { name: "vertices navigation" }),
      ).getByRole("button", { name: "Next" }),
    );
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-vertex-/u),
    ).toHaveLength(1);
    await user.click(
      within(
        screen.getByRole("navigation", { name: "edges navigation" }),
      ).getByRole("button", { name: "Next" }),
    );
    expect(
      screen.getAllByTestId(/^network-flow-saved-graph-edge-/u),
    ).toHaveLength(1);
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
      if (
        method === "GET" &&
        url.endsWith(
          "/api/v1/incidents/incident-1/network-flow/source-profiles",
        )
      ) {
        return jsonResponse(sourceProfileListResource());
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
        return jsonResponse(
          importEnvelope(
            importJob(uploadImportJobId, {
              code: "import_session_discovered",
              message: "Import discovery completed.",
              resource_refs: [
                {
                  kind: "import_session",
                  id: importSessionId,
                  route: `/api/v1/import-sessions/${importSessionId}`,
                },
              ],
            }),
          ),
        );
      }
      if (
        method === "GET" &&
        url.endsWith(`/api/v1/jobs/${uploadImportJobId}`)
      ) {
        return jsonResponse(
          importEnvelope(
            importJob(uploadImportJobId, {
              code: "import_session_discovered",
              message: "Import discovery completed.",
              resource_refs: [
                {
                  kind: "import_session",
                  id: importSessionId,
                  route: `/api/v1/import-sessions/${importSessionId}`,
                },
              ],
            }),
          ),
        );
      }
      if (method === "GET" && url.endsWith(`/${importSessionId}`)) {
        return jsonResponse(importEnvelope(importSessionResource()));
      }
      if (
        method === "GET" &&
        url.endsWith(`/${importSessionId}/units?limit=50`)
      ) {
        return jsonResponse({
          data: {
            import_units: [importUnitResource()],
          },
          meta: {
            request_id: "request-import",
            paging: { limit: 50, has_more: false, next_cursor: null },
          },
        });
      }
      if (method === "GET" && url.endsWith(`/${importUnitId}/preview`)) {
        return jsonResponse(importEnvelope(importPreviewResource(columns)));
      }
      if (method === "POST" && url.endsWith("/mapping-preview")) {
        const request = JSON.parse(String(init?.body)) as {
          owner_mapping: Record<string, unknown>;
        };
        return jsonResponse(
          importEnvelope({
            schema_id: "cartulary.imports.extension_mapping_preview_result.v1",
            import_session_id: importSessionId,
            import_unit_id: importUnitId,
            target_kind: "network_flow_table",
            extension_profile_id: "network_flow_activity",
            owner_result_schema_id:
              "cartulary.network_flow.import_preview_result.v1",
            owner_result: importPreviewResult(request.owner_mapping),
          }),
        );
      }
      if (method === "PUT" && url.endsWith("/mapping")) {
        return jsonResponse(
          importEnvelope({
            ...importUnitResource("mapped"),
            mapping_fingerprint: mappingFingerprint,
            approved_mapping: {
              target_kind: "network_flow_table",
              extension_profile_id: "network_flow_activity",
              owner_mapping_schema_id:
                "cartulary.network_flow.mapping_candidate.v1",
              owner_mapping: {},
              source_columns: columns.map((column) => ({
                ...column,
                field_key: null,
                entity_binding_mode: null,
                transform_id: null,
                transform_options: {},
                empty_value_policy: "omit_field",
              })),
            },
          }),
        );
      }
      if (method === "POST" && url.endsWith("/select")) {
        return jsonResponse(
          importEnvelope({
            import_session_id: importSessionId,
            session_status: "ready_to_apply",
            selected_unit_ids: [importUnitId],
            unit: importUnitResource("ready"),
          }),
        );
      }
      if (method === "POST" && url.endsWith(`/${importSessionId}/apply`)) {
        return jsonResponse(importEnvelope(importJob(applyImportJobId)));
      }
      if (
        method === "GET" &&
        url.endsWith(`/api/v1/jobs/${applyImportJobId}`)
      ) {
        return jsonResponse(
          importEnvelope(
            importJob(applyImportJobId, {
              code: "import_session_applied",
              message: "Import session applied.",
              resource_refs: [
                {
                  kind: "network_flow_table",
                  id: returnedTableId,
                  route: `/api/v1/incidents/incident-1/network-flow/tables/${returnedTableId}`,
                },
              ],
            }),
          ),
        );
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

function importSessionResource() {
  return {
    import_session_id: importSessionId,
    incident_id: incidentResourceId,
    created_by_user_id: importActorId,
    created_at: "2026-07-28T12:00:00Z",
    source_file_kind: "csv",
    original_filename: "new-flows.csv",
    source_content_sha256: sourceDigest,
    parser_profile_id: "tabular_default",
    parser_version: "1.0.0",
    assistant_profile: "phase2_workbook_import_v1",
    session_status: "discovered",
    selected_unit_ids: [],
    blocking_diagnostics: [],
    nonblocking_warning_codes: [],
  } as const;
}

function importUnitResource(
  unitStatus: "discovered" | "mapped" | "ready" = "discovered",
) {
  return {
    import_session_id: importSessionId,
    import_unit_id: importUnitId,
    locator_kind: "csv_file",
    locator: { file: "source" },
    source_rect_a1: "A1:I2",
    header_row_ref: 1,
    data_start_row_ref: 2,
    inferred_row_count: 1,
    inferred_column_count: 9,
    warning_codes: [],
    unit_status: unitStatus,
  } as const;
}

function importPreviewResource(columns: ReturnType<typeof importColumns>) {
  return {
    ...importUnitResource(),
    columns,
    preview_rows: [],
    truncated: false,
  };
}

function importJob(
  jobId: string,
  resultSummary: {
    readonly code: string;
    readonly message: string;
    readonly resource_refs: readonly {
      readonly kind: string;
      readonly id: string;
      readonly route: string;
    }[];
  } | null = null,
) {
  return {
    job_id: jobId,
    scope: { kind: "incident", incident_id: incidentResourceId },
    status_route: `/api/v1/jobs/${jobId}`,
    submitted_by_user_id: importActorId,
    status: "succeeded",
    cancelable: false,
    progress: { completed: 1, total: 1 },
    result_summary: resultSummary,
    error_summary: null,
    submitted_at: "2026-07-28T12:00:00Z",
    started_at: "2026-07-28T12:00:01Z",
    finished_at: "2026-07-28T12:00:02Z",
    retained_until: "2026-08-04T12:00:02Z",
    updated_at: "2026-07-28T12:00:02Z",
  } as const;
}

function importEnvelope<T>(data: T) {
  return { data, meta: { request_id: "request-import" } };
}

function installNetworkFlowFetchMock(
  options: {
    readonly contributorNextCursor?: string;
    readonly renameConflictOnce?: boolean;
    readonly removeTablesOnRowFailure?: boolean;
    readonly rowFailureAfter?: number;
    readonly rowFailureCode?: string;
    readonly rowFailureReason?: string;
    readonly rowFailureStatus?: number;
    readonly savedGraphProjectionResult?: ReturnType<
      typeof graphResource
    >["graph_projection_result"];
    readonly savedGraphs?: Array<Record<string, unknown>>;
    readonly tables?: Array<Record<string, unknown>>;
  } = {},
) {
  let tables = options.tables ?? [tableResource()];
  let savedGraphs = options.savedGraphs ?? [];
  let renameConflictUsed = false;
  let rowQueryCount = 0;
  let contributorSelector: Record<string, unknown> =
    graphDefaultEdgeSelectorResource();
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
        method === "GET" &&
        url.endsWith("/api/v1/incidents/incident-1/network-flow/graph-views")
      ) {
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_view_list.v2",
          graph_views: savedGraphs,
        });
      }
      if (
        method === "POST" &&
        url.endsWith("/api/v1/incidents/incident-1/network-flow/graph-views")
      ) {
        const request = JSON.parse(String(init?.body)) as {
          display_name: string;
          semantic_query: ReturnType<typeof graphSemanticQueryResource>;
        };
        const created = savedGraphResource({
          displayName: request.display_name,
          graphViewID: "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
          semanticQuery: request.semantic_query,
          selected: false,
          status: "queued",
        });
        savedGraphs = [...savedGraphs, created];
        return jsonResponse(
          {
            schema_id: "cartulary.network_flow.graph_view_accepted.v2",
            graph_view: created,
            job_id: "graph-job-create",
            job_kind: "network_flow_activity.graph_view_materialize_v1",
          },
          202,
        );
      }
      if (
        method === "GET" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/graph-views/${graphViewId}/result`,
        )
      ) {
        const graphView = savedGraphs.find(
          (candidate) => candidate.graph_view_id === graphViewId,
        );
        const projection = {
          ...(options.savedGraphProjectionResult ??
            graphResource().graph_projection_result),
          graph_view_id: graphViewId,
        };
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_view_result.v2",
          graph_view: graphView,
          result: graphResultForProjection(projection),
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/graph-views/${graphViewId}/contributors/query`,
        )
      ) {
        const request = JSON.parse(String(init?.body)) as {
          selector: Record<string, unknown>;
        };
        return jsonResponse({
          schema_id:
            "cartulary.network_flow.graph_view_contributor_query_result.v2",
          graph_view_id: graphViewId,
          projection_result_id: projectionResultId,
          selector: request.selector,
          contributors: [{ row_ref: rowRefResource(), row: rowResource() }],
          meta: {
            paging: {
              limit: 100,
              returned_count: 1,
              next_cursor_token: null,
            },
          },
        });
      }
      if (
        method === "PATCH" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/graph-views/${graphViewId}`,
        )
      ) {
        const request = JSON.parse(String(init?.body)) as {
          display_name: string;
        };
        const current = savedGraphs.find(
          (candidate) => candidate.graph_view_id === graphViewId,
        );
        const renamed = {
          ...current,
          display_name: request.display_name,
          normalized_display_name: request.display_name.toLowerCase(),
          graph_view_version: Number(current?.graph_view_version ?? 1) + 1,
        };
        savedGraphs = savedGraphs.map((candidate) =>
          candidate.graph_view_id === graphViewId ? renamed : candidate,
        );
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_view_mutation_result.v2",
          graph_view: renamed,
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/graph-views/${graphViewId}/refresh`,
        )
      ) {
        const current = savedGraphs.find(
          (candidate) => candidate.graph_view_id === graphViewId,
        );
        const refreshing = {
          ...current,
          graph_view_version: Number(current?.graph_view_version ?? 1) + 1,
          materialization_generation:
            Number(current?.materialization_generation ?? 1) + 1,
          last_materialization_job_id: "graph-job-refresh",
          last_materialization_status: "running",
        };
        savedGraphs = savedGraphs.map((candidate) =>
          candidate.graph_view_id === graphViewId ? refreshing : candidate,
        );
        return jsonResponse(
          {
            schema_id: "cartulary.network_flow.graph_view_accepted.v2",
            graph_view: refreshing,
            job_id: "graph-job-refresh",
            job_kind: "network_flow_activity.graph_view_materialize_v1",
          },
          202,
        );
      }
      if (
        method === "DELETE" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/graph-views/${graphViewId}`,
        )
      ) {
        const current = savedGraphs.find(
          (candidate) => candidate.graph_view_id === graphViewId,
        );
        const retired = {
          ...current,
          state: "retired",
          graph_view_version: Number(current?.graph_view_version ?? 1) + 1,
          materialization_generation:
            Number(current?.materialization_generation ?? 1) + 1,
          selected_result: null,
        };
        savedGraphs = savedGraphs.filter(
          (candidate) => candidate.graph_view_id !== graphViewId,
        );
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_view_mutation_result.v2",
          graph_view: retired,
        });
      }
      if (
        method === "GET" &&
        url.endsWith(
          "/api/v1/incidents/incident-1/network-flow/source-profiles",
        )
      ) {
        return jsonResponse(sourceProfileListResource());
      }
      if (
        method === "PATCH" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}`,
        )
      ) {
        if (options.renameConflictOnce && !renameConflictUsed) {
          renameConflictUsed = true;
          tables = tables.map((table) =>
            table.network_flow_table_id === tableId
              ? {
                  ...table,
                  display_name: "server-flows.csv",
                  table_version: 2,
                  updated_at: "2026-07-10T12:01:00Z",
                }
              : table,
          );
          return jsonResponse(
            {
              error: {
                code: "network_flow_table_version_conflict",
                message: "The table changed.",
                details: {
                  field: "base_table_version",
                  reason_code: "stale_version",
                },
              },
            },
            409,
          );
        }
        const request = JSON.parse(String(init?.body)) as {
          base_table_version: number;
          display_name: string;
        };
        const current = tables.find(
          (table) => table.network_flow_table_id === tableId,
        );
        if (current === undefined) {
          return jsonResponse({ error: { code: "unexpected_table" } }, 404);
        }
        const unchanged = current.display_name === request.display_name.trim();
        const renamed = {
          ...current,
          display_name: request.display_name.trim(),
          table_version: unchanged
            ? current.table_version
            : Number(current.table_version) + 1,
          updated_at: unchanged ? current.updated_at : "2026-07-10T12:02:00Z",
        };
        tables = tables.map((table) =>
          table.network_flow_table_id === tableId ? renamed : table,
        );
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_mutation_result.v1",
          table: renamed,
        });
      }
      if (
        method === "DELETE" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}`,
        )
      ) {
        const current = tables.find(
          (table) => table.network_flow_table_id === tableId,
        );
        if (current === undefined) {
          return jsonResponse({ error: { code: "unexpected_table" } }, 404);
        }
        const deleted = {
          ...current,
          table_status: "soft_deleted",
          table_version: Number(current.table_version) + 1,
          updated_at: "2026-07-10T12:03:00Z",
          deleted_at: "2026-07-10T12:03:00Z",
        };
        tables = tables.filter(
          (table) => table.network_flow_table_id !== tableId,
        );
        return jsonResponse({
          schema_id: "cartulary.network_flow.table_mutation_result.v1",
          table: deleted,
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/rejected-rows/query`,
        )
      ) {
        rowQueryCount += 1;
        if (
          options.rowFailureAfter !== undefined &&
          rowQueryCount > options.rowFailureAfter
        ) {
          if (options.removeTablesOnRowFailure) {
            tables = [];
          }
          return jsonResponse(
            {
              error: {
                code: options.rowFailureCode ?? "authorization_denied",
                message: "Network Flow query access changed.",
                details: {
                  reason_code: options.rowFailureReason ?? "role_changed",
                },
              },
            },
            options.rowFailureStatus ?? 403,
          );
        }
        return jsonResponse({
          schema_id: "cartulary.network_flow.rejected_rows_query_result.v1",
          network_flow_table_id: tableId,
          diagnostics: [diagnosticResource()],
          meta: {
            query: {
              error_codes: [],
              field_keys: [],
              source_row_range: null,
              effective_sort: [],
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
        url.endsWith(
          `/api/v1/incidents/incident-1/network-flow/tables/${tableId}/query`,
        )
      ) {
        rowQueryCount += 1;
        if (
          options.rowFailureAfter !== undefined &&
          rowQueryCount > options.rowFailureAfter
        ) {
          if (options.removeTablesOnRowFailure) {
            tables = [];
          }
          return jsonResponse(
            {
              error: {
                code: options.rowFailureCode ?? "authorization_denied",
                message: "Network Flow query access changed.",
                details: {
                  reason_code: options.rowFailureReason ?? "role_changed",
                },
              },
            },
            options.rowFailureStatus ?? 403,
          );
        }
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
        const request = JSON.parse(String(init?.body)) as {
          aggregation: ReturnType<
            typeof graphSemanticQueryResource
          >["aggregation"];
          filters?: ReturnType<typeof graphSemanticQueryResource>["filters"];
          table_scope:
            | { mode: "active_table"; active_table_id: string }
            | { mode: "selected_tables"; selected_table_ids: string[] }
            | { mode: "all_active_tables" };
          time_range?: ReturnType<
            typeof graphSemanticQueryResource
          >["time_range"];
        };
        const selectedTableIds =
          request.table_scope.mode === "selected_tables"
            ? request.table_scope.selected_table_ids
            : request.table_scope.mode === "active_table"
              ? [request.table_scope.active_table_id]
              : tables.map((table) => String(table.network_flow_table_id));
        return jsonResponse(
          graphResource({
            aggregation: request.aggregation,
            filters: request.filters ?? [],
            selectedTableIds,
            timeRange: request.time_range ?? {
              start_utc: null,
              end_utc: null,
            },
          }),
        );
      }
      if (
        method === "POST" &&
        url.endsWith(
          "/api/v1/incidents/incident-1/network-flow/graphs/contributors/query",
        )
      ) {
        const request = JSON.parse(String(init?.body)) as {
          schema_id: string;
          selector?: Record<string, unknown>;
        };
        if (request.selector !== undefined) {
          contributorSelector = request.selector;
        }
        return jsonResponse({
          schema_id: "cartulary.network_flow.graph_contributor_query_result.v2",
          graph_query_digest: graphDigest,
          selector: contributorSelector,
          contributors: [{ row_ref: rowRefResource(), row: rowResource() }],
          meta: {
            paging: {
              limit: 50,
              returned_count: 1,
              next_cursor_token:
                request.schema_id ===
                "cartulary.network_flow.graph_contributor_query_request.v2"
                  ? (options.contributorNextCursor ?? null)
                  : null,
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
            selector_kind: (
              JSON.parse(String(init?.body)) as {
                selector: { kind: string };
              }
            ).selector.kind,
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

function diagnosticResource() {
  return {
    diagnostic_id: diagnosticId,
    source_row_number: 2,
    source_column_ordinal: 3,
    raw_header_sha256: sourceDigest,
    field_key: "network_flow.src_ip",
    error_code: "network_flow_invalid_ip",
    reason_code: "invalid_ipv4",
    safe_sample: "999.0.2.1",
    raw_value_sha256: sourceDigest,
    message_key: "network_flow.diagnostic.network_flow_invalid_ip.invalid_ipv4",
    message_args: {},
    message: "network_flow.diagnostic.network_flow_invalid_ip.invalid_ipv4",
    limit_name: null,
    limit_value: null,
    actual_value: null,
  };
}

function graphResource(
  options: {
    readonly aggregation?:
      | {
          readonly mode: "default_flow_edge_v1";
          readonly include_example_row_refs?: boolean;
        }
      | {
          readonly mode: "time_bucket_v1";
          readonly bucket_width_seconds: 60 | 300 | 900 | 3600 | 21600 | 86400;
          readonly include_example_row_refs?: boolean;
        };
    readonly filters?: ReturnType<typeof graphSemanticQueryResource>["filters"];
    readonly selectedTableIds?: readonly string[];
    readonly timeRange?: ReturnType<
      typeof graphSemanticQueryResource
    >["time_range"];
  } = {},
) {
  const semanticQuery = graphSemanticQueryResource(options);
  const aggregation = semanticQuery.aggregation;
  const temporal = aggregation.mode === "time_bucket_v1";
  const buckets =
    aggregation.mode === "time_bucket_v1"
      ? timeBucketResources(
          semanticQuery.time_range.start_utc as string,
          semanticQuery.time_range.end_utc as string,
          aggregation.bucket_width_seconds,
        )
      : [];
  const firstBucket = buckets[0];
  const sourceEdgeId = temporal ? temporalEdgeId : edgeId;
  const edgeSelector =
    temporal && firstBucket !== undefined
      ? {
          kind: "time_bucket_edge" as const,
          source_edge_id: temporalEdgeId,
          bucket_start_utc: firstBucket.start_utc,
          bucket_end_utc: firstBucket.end_utc,
          source_endpoint_value: "192.0.2.10",
          destination_endpoint_value: "198.51.100.20",
          protocol: 6,
          destination_port_present: false as const,
        }
      : graphDefaultEdgeSelectorResource();
  return {
    schema_id: "cartulary.network_flow.graph_query_result.v2",
    graph_query_digest: graphDigest,
    semantic_query: semanticQuery,
    graph_projection_result: {
      projection_schema_id: "graph_projection.v2",
      projection_result_id:
        "gpres_0000000000000000000000000000000000000000000000000000000000000000",
      graph_view_id:
        "gv_0000000000000000000000000000000000000000000000000000000000000000",
      source_owner_id: "network_flow_activity",
      source_snapshot_id: "snapshot-1",
      projection_version: temporal
        ? "network_flow_activity.time_bucket.v1"
        : "network_flow_activity.v1",
      normalized_configuration_sha256: graphDigest,
      normalized_source_sha256: sourceDigest,
      canonical_output_sha256:
        "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
      properties: {},
      mapped_metadata: {},
      schema_registry: {
        vertex_kinds: [],
        edge_kinds: [],
        property_keys: [],
        metadata_keys: [],
      },
      vertices: [
        graphVertexResource({
          candidateValue: "192.0.2.10",
          endpointId: srcEndpointId,
          projectedVertexId: projectedSrcVertexId,
        }),
        graphVertexResource({
          candidateValue: "198.51.100.20",
          endpointId: dstEndpointId,
          projectedVertexId: projectedDstVertexId,
        }),
      ],
      edges: [
        {
          edge_id: projectedEdgeId,
          edge_kind: temporal
            ? "network_flow.bucketed_flow_edge.v1"
            : "network_flow.flow_edge.v1",
          edge_family: "direct",
          src_vertex_id: projectedSrcVertexId,
          dst_vertex_id: projectedDstVertexId,
          direction: "directed",
          labels: [],
          properties: {
            edge_id: sourceEdgeId,
            src_endpoint_id: srcEndpointId,
            dst_endpoint_id: dstEndpointId,
            ip_protocol: 6,
            flow_row_count: 1,
            ...(firstBucket === undefined
              ? {}
              : {
                  bucket_start_utc: firstBucket.start_utc,
                  bucket_end_utc: firstBucket.end_utc,
                }),
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
            source_relationship_id: sourceEdgeId,
            source_relationship_kind: temporal
              ? "network_flow.bucketed_flow_edge.v1"
              : "network_flow.flow_edge.v1",
            mapping_rule_id: "nf.map.flow_edge.v1",
          },
          sort_key: sourceEdgeId,
        },
      ],
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
    edge_annotations: [
      {
        projected_edge_id: projectedEdgeId,
        selector: edgeSelector,
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
      max_contributing_rows_per_graph: 250000,
      max_time_buckets_per_graph: 256,
    },
    vertex_selectors: [
      {
        projected_vertex_id: projectedSrcVertexId,
        selector: graphVertexSelectorResource(),
      },
      {
        projected_vertex_id: projectedDstVertexId,
        selector: {
          kind: "vertex",
          source_vertex_id: dstEndpointId,
          endpoint_value: "198.51.100.20",
        },
      },
    ],
    result_variant: temporal
      ? { kind: "time_bucket_v1", time_buckets: buckets }
      : { kind: "default_flow_edge_v1" },
  };
}

function savedGraphResource(
  options: {
    readonly displayName?: string;
    readonly graphViewID?: string;
    readonly selected?: boolean;
    readonly semanticQuery?: ReturnType<typeof graphSemanticQueryResource>;
    readonly status?:
      | "not_started"
      | "queued"
      | "running"
      | "succeeded"
      | "failed"
      | "cancelled";
  } = {},
) {
  const selected = options.selected ?? true;
  return {
    schema_id: "cartulary.network_flow.graph_view.v2",
    graph_view_id: options.graphViewID ?? graphViewId,
    incident_id: incidentResourceId,
    display_name: options.displayName ?? "Investigation graph",
    normalized_display_name: (
      options.displayName ?? "Investigation graph"
    ).toLowerCase(),
    graph_view_version: 1,
    materialization_generation: 1,
    state: "active",
    semantic_query: options.semanticQuery ?? graphSemanticQueryResource(),
    selected_result: selected
      ? {
          projection_result_id: projectionResultId,
          source_snapshot_id: "snapshot-1",
          projection_schema_id: "graph_projection.v2",
          projection_version: "network_flow_activity.v1",
          normalized_configuration_sha256: graphDigest,
          normalized_source_sha256: sourceDigest,
          canonical_output_sha256:
            "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
        }
      : null,
    last_materialization_job_id: selected ? "graph-job-1" : "graph-job-create",
    last_materialization_status:
      options.status ?? (selected ? "succeeded" : "queued"),
    last_failure_code: null,
    created_at: "2026-07-10T12:00:00Z",
    updated_at: "2026-07-10T12:00:00Z",
  };
}

function graphResultForProjection(
  projection: ReturnType<typeof graphResource>["graph_projection_result"],
) {
  const base = graphResource();
  const endpointByProjectedVertex = new Map(
    projection.vertices.map((vertex) => [
      vertex.vertex_id,
      {
        value: vertex.properties.endpoint_value,
      },
    ]),
  );
  return {
    ...base,
    graph_projection_result: projection,
    vertex_selectors: projection.vertices.map((vertex) => ({
      projected_vertex_id: vertex.vertex_id,
      selector: {
        kind: "vertex" as const,
        source_vertex_id: vertex.source_entity_ref.source_entity_id,
        endpoint_value: vertex.properties.endpoint_value,
      },
    })),
    edge_annotations: projection.edges.map((edge) => {
      const source = endpointByProjectedVertex.get(edge.src_vertex_id);
      const destination = endpointByProjectedVertex.get(edge.dst_vertex_id);
      if (source === undefined || destination === undefined) {
        throw new Error("saved graph fixture edge has an unknown endpoint");
      }
      const destinationPort = (edge.properties as Record<string, unknown>)
        .dst_port;
      return {
        projected_edge_id: edge.edge_id,
        selector: {
          kind: "default_edge" as const,
          source_edge_id: edge.source_relationship_ref.source_relationship_id,
          source_endpoint_value: source.value,
          destination_endpoint_value: destination.value,
          protocol: edge.properties.ip_protocol,
          ...(typeof destinationPort === "number"
            ? {
                destination_port_present: true as const,
                destination_port: destinationPort,
              }
            : { destination_port_present: false as const }),
        },
        example_row_refs: [],
        example_refs_truncated: false,
        example_refs_total_count: edge.properties.flow_row_count,
      };
    }),
  };
}

function largeSavedGraphProjectionResult() {
  const base = graphResource().graph_projection_result;
  const baseEdge = base.edges[0];
  if (baseEdge === undefined) {
    throw new Error("saved graph fixture requires one base edge");
  }
  const vertices = Array.from({ length: 501 }, (_, index) => {
    const identity = index.toString(16).padStart(64, "0");
    return graphVertexResource({
      candidateValue: `192.0.${Math.floor(index / 256)}.${index % 256}`,
      endpointId: `nfe_${identity}`,
      projectedVertexId: `vx_${identity}`,
    });
  });
  const edges = Array.from({ length: 1_001 }, (_, index) => {
    const identity = index.toString(16).padStart(64, "0");
    const sourceIdentity = (index % vertices.length)
      .toString(16)
      .padStart(64, "0");
    const destinationIdentity = ((index + 1) % vertices.length)
      .toString(16)
      .padStart(64, "0");
    return {
      ...baseEdge,
      edge_id: `ed_${identity}`,
      src_vertex_id: `vx_${sourceIdentity}`,
      dst_vertex_id: `vx_${destinationIdentity}`,
      properties: {
        ...baseEdge.properties,
        edge_id: `nff_${identity}`,
        src_endpoint_id: `nfe_${sourceIdentity}`,
        dst_endpoint_id: `nfe_${destinationIdentity}`,
      },
      source_relationship_ref: {
        ...baseEdge.source_relationship_ref,
        source_relationship_id: `nff_${identity}`,
      },
      sort_key: `nff_${identity}`,
    };
  });
  return { ...base, vertices, edges };
}

function graphSemanticQueryResource(
  options: {
    readonly aggregation?:
      | {
          readonly mode: "default_flow_edge_v1";
          readonly include_example_row_refs?: boolean;
        }
      | {
          readonly mode: "time_bucket_v1";
          readonly bucket_width_seconds: 60 | 300 | 900 | 3600 | 21600 | 86400;
          readonly include_example_row_refs?: boolean;
        };
    readonly filters?: readonly Record<string, unknown>[];
    readonly selectedTableIds?: readonly string[];
    readonly timeRange?: {
      readonly end_utc: string | null;
      readonly start_utc: string | null;
    };
  } = {},
) {
  return {
    schema_id: "cartulary.network_flow.graph_semantic_query.v2" as const,
    selected_table_ids: [...(options.selectedTableIds ?? [tableId])],
    filters: [
      ...(options.filters ?? [
        {
          field_key: "network_flow.endpoint_ip",
          op: "cidr_contains",
          value: "192.0.2.0/24",
        },
      ]),
    ],
    time_range: options.timeRange ?? {
      start_utc: "2026-07-10T00:00:00Z",
      end_utc: "2026-07-11T00:00:00Z",
    },
    aggregation:
      options.aggregation ??
      ({
        mode: "default_flow_edge_v1" as const,
        include_example_row_refs: true,
      } as const),
  };
}

function timeBucketResources(
  startUTC: string,
  endUTC: string,
  widthSeconds: number,
) {
  const widthMilliseconds = widthSeconds * 1000;
  const first =
    Math.floor(Date.parse(startUTC) / widthMilliseconds) * widthMilliseconds;
  const endExclusive =
    Math.ceil(Date.parse(endUTC) / widthMilliseconds) * widthMilliseconds;
  return Array.from(
    { length: (endExclusive - first) / widthMilliseconds },
    (_, index) => ({
      start_utc: new Date(first + index * widthMilliseconds).toISOString(),
      end_utc: new Date(first + (index + 1) * widthMilliseconds).toISOString(),
      unique_vertex_count: index === 0 ? 2 : 0,
      edge_count: index === 0 ? 1 : 0,
      contributing_row_count: index === 0 ? 1 : 0,
    }),
  );
}

function graphVertexSelectorResource() {
  return {
    kind: "vertex" as const,
    source_vertex_id: srcEndpointId,
    endpoint_value: "192.0.2.10",
  };
}

function graphDefaultEdgeSelectorResource() {
  return {
    kind: "default_edge" as const,
    source_edge_id: edgeId,
    source_endpoint_value: "192.0.2.10",
    destination_endpoint_value: "198.51.100.20",
    protocol: 6,
    destination_port_present: false as const,
  };
}

function graphVertexResource(options: {
  readonly candidateValue: string;
  readonly endpointId: string;
  readonly projectedVertexId: string;
}) {
  return {
    vertex_id: options.projectedVertexId,
    vertex_kind: "network_flow.ip_endpoint.v1",
    vertex_family: "network_flow.ip_endpoint.v1",
    labels: [],
    properties: {
      endpoint_kind: "ip",
      endpoint_value: options.candidateValue,
      contributing_table_ids: [tableId],
      flow_row_count: 1,
      indicator_candidate_value: options.candidateValue,
    },
    metadata: {
      mapping_rule_id: "nf.map.ip_endpoint.v1",
      aggregation_rule_id: null,
      aggregation_source_refs: [],
      mapped_metadata: {},
    },
    source_entity_ref: {
      source_entity_id: options.endpointId,
      source_entity_kind: "network_flow.ip_endpoint.v1",
      mapping_rule_id: "nf.map.ip_endpoint.v1",
    },
    sort_key: options.endpointId,
  };
}

function sourceProfileListResource() {
  return {
    schema_id: "cartulary.network_flow.source_profile_list.v2",
    source_profiles: [],
    effective_limits: {
      "network_flow.max_active_tables_per_incident": 8,
      "network_flow.max_retained_tables_per_incident": 32,
      "network_flow.max_selected_tables_per_query": 8,
      "network_flow.max_columns_per_csv": 128,
      "network_flow.max_header_scalar_length": 256,
      "network_flow.max_raw_cell_scalar_length": 8192,
      "network_flow.max_rows_per_csv": 100000,
      "network_flow.max_accepted_rows_per_table": 100000,
      "network_flow.max_rejected_row_diagnostics": 10000,
      "network_flow.max_filters_per_query": 16,
      "network_flow.max_sorts_per_query": 4,
      "network_flow.max_query_limit": 1000,
      "network_flow.max_graph_vertices": 1000,
      "network_flow.max_graph_edges": 1000,
      "network_flow.max_example_row_refs_per_edge": 10,
      "network_flow.max_binding_source_row_refs": 100,
      "network_flow.max_aggregate_counter_digits": 20,
      "network_flow.max_active_graph_views_per_incident": 32,
      "network_flow.max_retained_graph_views_per_incident": 128,
      "network_flow.max_nonterminal_graph_jobs_per_incident": 4,
      "network_flow.max_contributing_rows_per_graph": 250000,
      "network_flow.max_time_buckets_per_graph": 256,
      "network_flow.graph_materialization_timeout_seconds": 300,
    },
    meta: { count: 0 },
  };
}

function jsonResponse(body: unknown, status = 200): Response {
  const envelope =
    status >= 200 &&
    status < 300 &&
    body !== null &&
    typeof body === "object" &&
    !Array.isArray(body) &&
    "schema_id" in body &&
    typeof body.schema_id === "string" &&
    body.schema_id.startsWith("cartulary.network_flow.")
      ? { data: body, meta: { request_id: "req-network-flow-unit" } }
      : body;
  return new Response(JSON.stringify(envelope), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function graphRequestBodies(
  fetchSpy: ReturnType<typeof installNetworkFlowFetchMock>,
): Array<Record<string, unknown>> {
  return fetchSpy.mock.calls
    .filter(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/graphs/query",
      ),
    )
    .map(
      ([, init]) => JSON.parse(String(init?.body)) as Record<string, unknown>,
    );
}

function indicatorLinkRequestBodies(
  fetchSpy: ReturnType<typeof installNetworkFlowFetchMock>,
): Array<Record<string, unknown>> {
  return fetchSpy.mock.calls
    .filter(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/indicator-links",
      ),
    )
    .map(
      ([, init]) => JSON.parse(String(init?.body)) as Record<string, unknown>,
    );
}

function contributorRequestBodies(
  fetchSpy: ReturnType<typeof installNetworkFlowFetchMock>,
): Array<Record<string, unknown>> {
  return fetchSpy.mock.calls
    .filter(([input]) =>
      requestURL(input).endsWith(
        "/api/v1/incidents/incident-1/network-flow/graphs/contributors/query",
      ),
    )
    .map(
      ([, init]) => JSON.parse(String(init?.body)) as Record<string, unknown>,
    );
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
