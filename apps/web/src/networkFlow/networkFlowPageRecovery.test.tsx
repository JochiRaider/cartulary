import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ExtensionAvailabilityUnavailableError } from "../extensions/extensionAvailability";
import {
  decodeNetworkFlowContributorResult,
  decodeNetworkFlowRejectedRowsQueryResult,
  decodeNetworkFlowTableQueryResult,
  NetworkFlowContractDecodeError,
  type NetworkFlowRow,
} from "../services/networkFlowContractAdapter";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { explorationFixture } from "./explorationTestFixtures";
import { NetworkFlowTableController } from "./NetworkFlowTableController";
import * as client from "./networkFlowClient";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import {
  defaultGraphQuerySettings,
  emptyNetworkFlowAcceptedQuery,
  emptyNetworkFlowRejectedQuery,
} from "./networkFlowQueryModel";
import {
  deferredSavedGraph,
  savedGraphResultFixture,
  savedGraphSelectorFixture,
} from "./savedGraphTestFixtures";
import {
  tableAuthority,
  tableFixture,
  tableIncidentId,
} from "./tableLifecycleTestFixtures";
import { useNetworkFlowGraphController } from "./useNetworkFlowGraphController";
import { useNetworkFlowRejectedRowsController } from "./useNetworkFlowRejectedRowsController";
import { useNetworkFlowRowsController } from "./useNetworkFlowRowsController";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});
const tableId = tableFixture().network_flow_table_id;
const paging = (next: string | null, count = 1) => ({
  limit: 200,
  returned_count: count,
  next_cursor_token: next,
});
const expired = () =>
  new NetworkFlowRequestError({
    code: "network_flow_cursor_invalid",
    reasonCode: "expired",
    retryAction: "restart_query",
    retryable: false,
    safeMessage: "Expired cursor.",
    status: 400,
  });

describe("Network Flow page recovery boundaries", () => {
  it("discards late contributor success and authorization failure after selector replacement", async () => {
    for (const failure of [false, true]) {
      const graph = explorationFixture(2, 1);
      const selected = graph.vertex_selectors[0]?.selector;
      if (!selected) throw new Error("Missing fixture vertex");
      vi.spyOn(client, "queryNetworkFlowGraph").mockResolvedValue(graph);
      const pending =
        deferredSavedGraph<
          Awaited<ReturnType<typeof client.queryNetworkFlowContributors>>
        >();
      const basePage = {
        schema_id:
          "cartulary.network_flow.graph_contributor_query_result.v2" as const,
        graph_query_digest: graph.graph_query_digest,
        selector: selected,
        contributors: [],
        meta: { paging: paging(null, 0) },
      };
      const pages = vi
        .spyOn(client, "queryNetworkFlowContributors")
        .mockImplementationOnce(() => pending.promise)
        .mockResolvedValue(basePage);
      const onError = vi.fn(),
        onIncidentAccessLost = vi.fn();
      const options = {
        settings: defaultGraphQuerySettings,
        applicationRevision: 0,
        revision: 1,
        onQueryResult: vi.fn(),
        tableLifecycle: new NetworkFlowTableController(),
        availability: readyExtensionAvailability(tableIncidentId),
        activeTableId: tableId,
        apiBase: undefined,
        enabled: true,
        incidentId: tableIncidentId,
        onError,
        onIncidentAccessLost,
        query: emptyNetworkFlowAcceptedQuery,
        tables: [tableFixture()],
      };
      const hook = renderHook(() => useNetworkFlowGraphController(options));
      await waitFor(() => expect(hook.result.current.graph).toBe(graph));
      act(() => hook.result.current.selectGraphObject(selected));
      await waitFor(() => expect(pages).toHaveBeenCalledTimes(1));
      const next = graph.vertex_selectors[1]?.selector;
      if (!next) throw new Error("Missing second fixture vertex");
      act(() => hook.result.current.selectGraphObject(next));
      await waitFor(() =>
        expect(hook.result.current.contributorLoadState).toBe("ready"),
      );
      const count = onError.mock.calls.length;
      await act(async () => {
        if (failure)
          pending.reject(
            new NetworkFlowRequestError({
              code: "authorization_denied",
              retryAction: "do_not_retry",
              retryable: false,
              safeMessage: "Denied.",
              status: 403,
            }),
          );
        else pending.resolve(basePage);
      });
      expect(onError).toHaveBeenCalledTimes(count);
      expect(onIncidentAccessLost).not.toHaveBeenCalled();
      expect(hook.result.current.selection).toEqual(next);
      expect(hook.result.current.graph).toBe(graph);
      hook.unmount();
      vi.restoreAllMocks();
    }
  });
  it("gives accepted rows and diagnostics equivalent recovery without notifying authoring for pagination", async () => {
    for (const rejected of [false, true]) {
      const rows = vi.spyOn(client, "queryNetworkFlowTable").mockResolvedValue({
        rows: [pageRow()],
        paging: paging("next"),
        metadata: acceptedMeta,
      });
      const diagnostics = vi
        .spyOn(client, "queryNetworkFlowRejectedRows")
        .mockResolvedValue({
          diagnostics: [pageDiagnostic()],
          paging: paging("next"),
          metadata: rejectedMeta,
        });
      const onQueryResult = vi.fn();
      const options = {
        availability: readyExtensionAvailability(tableIncidentId),
        activeTableId: tableId,
        apiBase: undefined,
        enabled: true,
        incidentId: tableIncidentId,
        onError: vi.fn(),
        onIncidentAccessLost: vi.fn(),
        revision: 1,
        onQueryResult,
        readIdentity: "scope",
        isCurrentRead: () => true,
        onProtectedStateLoss: vi.fn(),
      };
      const hook = renderHook(() => {
        const a = useNetworkFlowRowsController({
          ...options,
          enabled: !rejected,
          query: emptyNetworkFlowAcceptedQuery,
        });
        const b = useNetworkFlowRejectedRowsController({
          ...options,
          enabled: rejected,
          query: emptyNetworkFlowRejectedQuery,
        });
        return rejected ? b : a;
      });
      await waitFor(() => expect(hook.result.current.loadState).toBe("ready"));
      const request = rejected ? diagnostics : rows;
      request.mockRejectedValueOnce(new Error("offline"));
      act(() => hook.result.current.nextPage());
      await waitFor(() => expect(hook.result.current.loadState).toBe("error"));
      expect(hook.result.current.pageNumber).toBe(1);
      act(() => hook.result.current.retry());
      await waitFor(() => expect(hook.result.current.pageNumber).toBe(2));
      request.mockRejectedValueOnce(expired());
      act(() => hook.result.current.nextPage());
      await waitFor(() => expect(hook.result.current.pageNumber).toBe(1));
      expect(onQueryResult).toHaveBeenCalledTimes(1);
      expect(request.mock.calls.at(-1)?.[0].request).toEqual({
        schema_id: rejected
          ? "cartulary.network_flow.rejected_rows_query_request.v1"
          : "cartulary.network_flow.table_query_request.v1",
      });
      hook.unmount();
      vi.restoreAllMocks();
    }
  });

  it("recovers contributor expiry on the same graph and routes stale composition through recomputation", async () => {
    const graph = explorationFixture(2, 1);
    const selected = graph.vertex_selectors[0]?.selector;
    if (!selected) throw new Error("Missing fixture vertex");
    const graphRequest = vi
      .spyOn(client, "queryNetworkFlowGraph")
      .mockResolvedValue(graph);
    const pages = vi
      .spyOn(client, "queryNetworkFlowContributors")
      .mockResolvedValue({
        schema_id: "cartulary.network_flow.graph_contributor_query_result.v2",
        graph_query_digest: graph.graph_query_digest,
        selector: selected,
        contributors: [
          {
            row: pageRow(),
            row_ref: {
              network_flow_table_id: tableId,
              network_flow_row_id: pageRow().network_flow_row_id,
              source_row_number: 1,
              mapping_fingerprint: pageRow().mapping_fingerprint,
            },
          },
        ],
        meta: { paging: paging("next") },
      });
    const options = {
      settings: defaultGraphQuerySettings,
      applicationRevision: 0,
      revision: 1,
      onQueryResult: vi.fn(),
      tableLifecycle: new NetworkFlowTableController(),
      availability: readyExtensionAvailability(tableIncidentId),
      activeTableId: tableId,
      apiBase: undefined,
      enabled: true,
      incidentId: tableIncidentId,
      onError: vi.fn(),
      onIncidentAccessLost: vi.fn(),
      query: emptyNetworkFlowAcceptedQuery,
      tables: [tableFixture()],
    };
    const hook = renderHook(() => useNetworkFlowGraphController(options));
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    act(() => hook.result.current.selectGraphObject(selected));
    await waitFor(() =>
      expect(hook.result.current.contributorLoadState).toBe("ready"),
    );
    pages.mockRejectedValueOnce(expired());
    act(() => hook.result.current.nextContributorPage());
    await waitFor(() => expect(pages).toHaveBeenCalledTimes(3));
    expect(hook.result.current.graph).toBe(graph);
    expect(hook.result.current.graphStale).toBe(false);
    expect(pages.mock.calls[2]?.[0].request).toEqual(
      pages.mock.calls[0]?.[0].request,
    );
    pages.mockRejectedValueOnce(
      new NetworkFlowRequestError({
        code: "network_flow_graph_query_stale",
        reasonCode: "digest_mismatch",
        retryAction: "refresh_resource",
        retryable: false,
        safeMessage: "Graph changed.",
        status: 409,
      }),
    );
    act(() => hook.result.current.nextContributorPage());
    await waitFor(() => expect(hook.result.current.graphStale).toBe(true));
    expect(hook.result.current.graph).toBeNull();
    expect(hook.result.current.selection).toBeNull();
    expect(graphRequest).toHaveBeenCalledTimes(1);
    act(() => hook.result.current.refreshGraph());
    await waitFor(() => expect(graphRequest).toHaveBeenCalledTimes(2));
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    const resourceRefresh = vi
      .spyOn(options.tableLifecycle, "loadTables")
      .mockResolvedValue(true);
    act(() => hook.result.current.selectGraphObject(selected));
    await waitFor(() =>
      expect(hook.result.current.contributorLoadState).toBe("ready"),
    );
    pages.mockRejectedValueOnce(
      new NetworkFlowRequestError({
        code: "network_flow_table_not_active",
        status: 409,
        retryable: false,
        retryAction: "refresh_resource",
        safeMessage: "Source is inactive.",
      }),
    );
    act(() => hook.result.current.nextContributorPage());
    await waitFor(() => expect(hook.result.current.graphStale).toBe(true));
    expect(resourceRefresh).toHaveBeenCalledOnce();
    expect(options.onIncidentAccessLost).not.toHaveBeenCalled();
  });

  it("blocks queued page dispatch after current authority changes", async () => {
    const availability = readyExtensionAvailability(tableIncidentId);
    const gate = deferredSavedGraph<void>();
    const prior = availability.runProfileRequest(
      "network_flow_activity",
      "/api/v1/incidents/{incident_id}/network-flow",
      () => gate.promise,
    );
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    let allowed = true;
    const request = client.queryNetworkFlowTable({
      availability,
      incidentId: tableIncidentId,
      tableId,
      request: { schema_id: "cartulary.network_flow.table_query_request.v1" },
      authorizeDispatch: () => {
        if (!allowed) throw new Error("obsolete");
      },
    });
    const assertion = expect(request).rejects.toThrow("obsolete");
    allowed = false;
    gate.resolve();
    await prior;
    await assertion;
    expect(fetch).not.toHaveBeenCalled();
    const stopped =
      deferredSavedGraph<
        Awaited<ReturnType<typeof client.queryNetworkFlowGraph>>
      >();
    vi.spyOn(client, "queryNetworkFlowGraph").mockImplementation(
      () => stopped.promise,
    );
    let current = true;
    const options = {
      settings: defaultGraphQuerySettings,
      applicationRevision: 0,
      revision: 1,
      onQueryResult: vi.fn(),
      tableLifecycle: new NetworkFlowTableController(),
      availability,
      activeTableId: tableId,
      apiBase: undefined,
      enabled: true,
      incidentId: tableIncidentId,
      onError: vi.fn(),
      onIncidentAccessLost: vi.fn(),
      query: emptyNetworkFlowAcceptedQuery,
      tables: [tableFixture()],
      readIdentity: "captured-authority",
      isCurrentRead: () => current,
    };
    const graph = renderHook(() => useNetworkFlowGraphController(options));
    await waitFor(() =>
      expect(graph.result.current.graphLoadState).toBe("loading"),
    );
    current = false;
    await act(async () =>
      stopped.reject(new ExtensionAvailabilityUnavailableError()),
    );
    await waitFor(() =>
      expect(graph.result.current.graphLoadState).toBe("idle"),
    );
    expect(options.onQueryResult).not.toHaveBeenCalled();
    expect(options.onIncidentAccessLost).not.toHaveBeenCalled();
  });

  it("keeps draft context stable while fencing replaced role session and availability read identities", async () => {
    let authority = tableAuthority();
    const owner = new NetworkFlowTableController();
    owner.bind(
      {
        list: async () => [tableFixture()],
        submit: async () => tableFixture(),
      },
      () => authority,
    );
    await owner.loadTables();
    const draft = owner.queryContextIdentity(),
      read = owner.readContextIdentity();
    authority = { ...authority, role: "viewer" };
    owner.revalidateAuthority();
    expect(owner.queryContextIdentity()).toBe(draft);
    expect(owner.readContextIdentity()).not.toBe(read);
    const roleRead = owner.readContextIdentity();
    authority = {
      ...authority,
      availabilityTag: { epochId: "epoch", generation: 2n },
    };
    owner.revalidateAuthority();
    expect(owner.queryContextIdentity()).toBe(draft);
    expect(owner.readContextIdentity()).not.toBe(roleRead);
    authority = { ...authority, sessionIdentity: "session-2" };
    owner.revalidateAuthority();
    expect(owner.queryContextIdentity()).not.toBe(draft);
  });

  it("fences late rows and diagnostics through incident actor role session and availability authority changes", async () => {
    for (const rejected of [false, true])
      for (const outcome of ["success", "expired", "denied"] as const)
        for (const change of [
          "incident",
          "actor",
          "session",
          "role",
          "availability",
        ] as const) {
          let authority = tableAuthority();
          const owner = new NetworkFlowTableController();
          owner.bind(
            {
              list: async () => [tableFixture()],
              submit: async () => tableFixture(),
            },
            () => authority,
          );
          await owner.loadTables();
          const nextRows =
            deferredSavedGraph<
              Awaited<ReturnType<typeof client.queryNetworkFlowTable>>
            >();
          const nextDiagnostics =
            deferredSavedGraph<
              Awaited<ReturnType<typeof client.queryNetworkFlowRejectedRows>>
            >();
          const rows = vi
            .spyOn(client, "queryNetworkFlowTable")
            .mockResolvedValue({
              rows: [pageRow()],
              paging: paging("next"),
              metadata: acceptedMeta,
            });
          const diagnostics = vi
            .spyOn(client, "queryNetworkFlowRejectedRows")
            .mockResolvedValue({
              diagnostics: [pageDiagnostic()],
              paging: paging("next"),
              metadata: rejectedMeta,
            });
          const options = {
            availability: readyExtensionAvailability(tableIncidentId),
            activeTableId: tableId,
            apiBase: undefined,
            enabled: true,
            onError: vi.fn(),
            onIncidentAccessLost: vi.fn(),
            onProtectedStateLoss: vi.fn(),
            onQueryResult: vi.fn(),
            revision: 1,
          };
          const hook = renderHook(() => {
            const readIdentity = owner.readContextIdentity();
            const context = {
              ...options,
              incidentId: authority.incidentId,
              readIdentity,
              isCurrentRead: () => owner.readContextIdentity() === readIdentity,
            };
            const accepted = useNetworkFlowRowsController({
              ...context,
              enabled: !rejected,
              query: emptyNetworkFlowAcceptedQuery,
            });
            const rejectedPage = useNetworkFlowRejectedRowsController({
              ...context,
              enabled: rejected,
              query: emptyNetworkFlowRejectedQuery,
            });
            return rejected ? rejectedPage : accepted;
          });
          await waitFor(() =>
            expect(hook.result.current.loadState).toBe("ready"),
          );
          rows.mockImplementationOnce(() => nextRows.promise);
          diagnostics.mockImplementationOnce(() => nextDiagnostics.promise);
          act(() => hook.result.current.nextPage());
          authority = {
            ...authority,
            ...(change === "incident"
              ? { incidentId: "44444444-4444-4444-8444-444444444444" }
              : {}),
            ...(change === "actor"
              ? { actorId: "55555555-5555-4555-8555-555555555555" }
              : {}),
            ...(change === "session" ? { sessionIdentity: "new-session" } : {}),
            ...(change === "role" ? { role: "viewer" as const } : {}),
            ...(change === "availability"
              ? { availabilityTag: { epochId: "new-epoch", generation: 2n } }
              : {}),
          };
          owner.revalidateAuthority();
          await owner.loadTables();
          hook.rerender();
          await waitFor(() =>
            expect(hook.result.current.loadState).toBe("ready"),
          );
          const callbacks = options.onError.mock.calls.length;
          await act(async () => {
            if (outcome === "success") {
              nextRows.resolve({
                rows: [pageRow()],
                paging: paging(null),
                metadata: acceptedMeta,
              });
              nextDiagnostics.resolve({
                diagnostics: [pageDiagnostic()],
                paging: paging(null),
                metadata: rejectedMeta,
              });
            } else
              (rejected ? nextDiagnostics : nextRows).reject(
                outcome === "expired"
                  ? expired()
                  : new NetworkFlowRequestError({
                      code: "authorization_denied",
                      status: 403,
                      retryable: false,
                      retryAction: "do_not_retry",
                      safeMessage: "Denied.",
                    }),
              );
          });
          expect(options.onError).toHaveBeenCalledTimes(callbacks);
          expect(options.onQueryResult).toHaveBeenCalledTimes(2);
          expect(options.onProtectedStateLoss).not.toHaveBeenCalled();
          expect(options.onIncidentAccessLost).not.toHaveBeenCalled();
          expect(hook.result.current.pageNumber).toBe(1);
          expect(hook.result.current.error).toBeNull();
          expect(rejected ? diagnostics : rows).toHaveBeenCalledTimes(3);
          hook.unmount();
          vi.restoreAllMocks();
        }
  });

  it("rejects malformed counts table identities contributor bindings and row references", async () => {
    for (const status of [400, 401, 403]) {
      vi.stubGlobal(
        "fetch",
        vi.fn(
          async () =>
            new Response(
              JSON.stringify({
                error: {
                  code: "network_flow_cursor_invalid",
                  details: {
                    reason_code: "expired",
                    retry_action: "restart_query",
                  },
                },
              }),
              { status, headers: { "Content-Type": "application/json" } },
            ),
        ),
      );
      const read = client.queryNetworkFlowTable({
        availability: readyExtensionAvailability(tableIncidentId),
        incidentId: tableIncidentId,
        tableId,
        request: { schema_id: "cartulary.network_flow.table_query_request.v1" },
      });
      if (status === 400)
        await expect(read).rejects.toBeInstanceOf(SyntaxError);
      else
        await expect(read).rejects.toMatchObject({
          status,
          code: status === 401 ? "session_required" : "authorization_denied",
          retryAction: "do_not_retry",
        });
    }
    const row = pageRow();
    const accepted = {
      schema_id: "cartulary.network_flow.table_query_result.v1",
      network_flow_table_id: tableId,
      rows: [row],
      meta: { query: acceptedMeta, paging: paging(null) },
    };
    expect(
      decodeNetworkFlowTableQueryResult(accepted, tableId).rows,
    ).toHaveLength(1);
    for (const patch of [
      { network_flow_table_id: tableFixture("b").network_flow_table_id },
      {
        rows: [row, row],
        meta: { query: acceptedMeta, paging: paging(null, 2) },
      },
      { meta: { query: acceptedMeta, paging: paging(null, 2) } },
      { rows: [], meta: { query: acceptedMeta, paging: paging("cursor", 0) } },
    ])
      expect(() =>
        decodeNetworkFlowTableQueryResult({ ...accepted, ...patch }, tableId),
      ).toThrow(NetworkFlowContractDecodeError);
    const rejected = {
      schema_id: "cartulary.network_flow.rejected_rows_query_result.v1",
      network_flow_table_id: tableId,
      diagnostics: [],
      meta: { query: rejectedMeta, paging: paging(null, 0) },
    };
    expect(
      decodeNetworkFlowRejectedRowsQueryResult(rejected, tableId).diagnostics,
    ).toEqual([]);
    expect(() =>
      decodeNetworkFlowRejectedRowsQueryResult(
        rejected,
        tableFixture("b").network_flow_table_id,
      ),
    ).toThrow(NetworkFlowContractDecodeError);
    const graph = savedGraphResultFixture().result;
    const context = {
      schema_id:
        "cartulary.network_flow.graph_contributor_query_request.v2" as const,
      graph_query: {
        ...graph.semantic_query,
        selected_table_ids: [tableId] as [string],
      },
      graph_query_digest: graph.graph_query_digest,
      selector: savedGraphSelectorFixture,
      limit: 200,
    };
    const contributor = {
      row,
      row_ref: {
        network_flow_row_id: row.network_flow_row_id,
        network_flow_table_id: tableId,
        source_row_number: 1,
        mapping_fingerprint: row.mapping_fingerprint,
      },
    };
    const response = {
      schema_id: "cartulary.network_flow.graph_contributor_query_result.v2",
      graph_query_digest: graph.graph_query_digest,
      selector: savedGraphSelectorFixture,
      contributors: [contributor],
      meta: { paging: paging(null) },
    };
    expect(
      decodeNetworkFlowContributorResult(response, context).contributors,
    ).toHaveLength(1);
    for (const patch of [
      { graph_query_digest: "b".repeat(64) },
      {
        selector: { ...savedGraphSelectorFixture, endpoint_value: "192.0.2.9" },
      },
      {
        contributors: [
          {
            ...contributor,
            row_ref: { ...contributor.row_ref, source_row_number: 2 },
          },
        ],
      },
    ])
      expect(() =>
        decodeNetworkFlowContributorResult({ ...response, ...patch }, context),
      ).toThrow(NetworkFlowContractDecodeError);
  });
});

const acceptedMeta = {
  filters: [],
  sort: [],
  effective_sort: [],
  table_ids: [tableId],
};
const rejectedMeta = {
  error_codes: [],
  field_keys: [],
  source_row_range: null,
  effective_sort: [],
};
function pageRow(): NetworkFlowRow {
  const table = tableFixture(),
    digest = "a".repeat(64);
  return {
    network_flow_row_id: `nfr_${digest}`,
    network_flow_table_id: tableId,
    incident_id: tableIncidentId,
    source_row_number: 1,
    source_row_digest_sha256: digest,
    normalized_row_digest_sha256: digest,
    mapping_fingerprint: table.mapping_fingerprint,
    "network_flow.flow_start_utc": table.created_at,
    "network_flow.flow_end_utc": table.created_at,
    "network_flow.src_ip": "192.0.2.1",
    "network_flow.dst_ip": "192.0.2.2",
    "network_flow.src_port": null,
    "network_flow.dst_port": null,
    "network_flow.ip_protocol": 6,
    "network_flow.bytes_count": "1",
    "network_flow.packets_count": "1",
    "network_flow.exporter_id": null,
    "network_flow.input_interface": null,
    "network_flow.output_interface": null,
    "network_flow.tcp_flags": null,
    "network_flow.application_label": null,
    unmapped_raw: {},
    "network_flow.observation_source_ref": {
      import_session_id: table.source_import_session_id,
      import_unit_id: table.source_import_unit_id,
      source_content_sha256: table.source_content_sha256,
      source_profile_id: table.source_profile_id,
      parser_profile_id: table.parser_profile_id,
      mapping_fingerprint: table.mapping_fingerprint,
      source_row_number: 1,
      source_row_digest_sha256: digest,
    },
    created_at: table.created_at,
    created_by_user_id: table.created_by_user_id,
  };
}

function pageDiagnostic(): import("../services/networkFlowContractAdapter").NetworkFlowDiagnostic {
  return {
    diagnostic_id: `nfd_${"a".repeat(64)}`,
    source_row_number: 1,
    source_column_ordinal: 3,
    raw_header_sha256: null,
    field_key: "network_flow.src_ip",
    error_code: "network_flow_invalid_ip",
    reason_code: "invalid_syntax",
    safe_sample: null,
    raw_value_sha256: null,
    message_key: "network_flow_invalid_ip",
    message_args: {},
    message: "Invalid IP.",
    limit_name: null,
    limit_value: null,
    actual_value: null,
  };
}
