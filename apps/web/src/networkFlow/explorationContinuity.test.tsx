import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { explorationFixture } from "./explorationTestFixtures";
import { NetworkFlowTableController } from "./NetworkFlowTableController";
import * as client from "./networkFlowClient";
import { NetworkFlowRequestError } from "./networkFlowErrors";
import { networkFlowVertexLinkCandidate } from "./networkFlowIndicatorLinkModel";
import { defaultGraphQuerySettings } from "./networkFlowQueryModel";
import { deferredSavedGraph } from "./savedGraphTestFixtures";
import {
  tableAuthority,
  tableFixture,
  tableIncidentId,
} from "./tableLifecycleTestFixtures";
import { useNetworkFlowGraphController } from "./useNetworkFlowGraphController";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});
type ContributorResult = Awaited<
  ReturnType<typeof client.queryNetworkFlowContributors>
>;
function page(
  graph: client.NetworkFlowGraphResult,
  selector: client.NetworkFlowGraphSelector,
): ContributorResult {
  return {
    schema_id: "cartulary.network_flow.graph_contributor_query_result.v2",
    graph_query_digest: graph.graph_query_digest,
    selector,
    contributors: [],
    meta: {
      paging: { limit: 500, returned_count: 0, next_cursor_token: "next-page" },
    },
  };
}
function options(graph: client.NetworkFlowGraphResult) {
  return {
    settings: {
      ...defaultGraphQuerySettings,
      aggregation: graph.semantic_query.aggregation,
    },
    applicationRevision: 0,
    revision: 0,
    onQueryResult: vi.fn(),
    onNavigationContextChange: vi.fn(),
    tableLifecycle: new NetworkFlowTableController(),
    availability: readyExtensionAvailability(tableIncidentId),
    activeTableId: tableFixture().network_flow_table_id,
    apiBase: undefined,
    enabled: true,
    incidentId: tableIncidentId,
    onError: vi.fn(),
    onIncidentAccessLost: vi.fn(),
    query: {
      filters: [],
      sort: [],
      timeWindow:
        graph.semantic_query.aggregation.mode === "time_bucket_v1"
          ? {
              startUTC: "2026-07-10T00:00:00.000Z",
              endUTC: "2026-07-10T03:00:00.000Z",
            }
          : null,
    },
    tables: [tableFixture()],
  };
}

describe("Exploration continuity", () => {
  it("fences late contributor success and failure after bucket and selection changes", async () => {
    for (const failure of [false, true]) {
      const graph = explorationFixture(2, 1, true);
      vi.spyOn(client, "queryNetworkFlowGraph").mockResolvedValue(graph);
      const pending = deferredSavedGraph<ContributorResult>();
      const calls = vi
        .spyOn(client, "queryNetworkFlowContributors")
        .mockImplementationOnce(() => pending.promise)
        .mockImplementation(async (request) => {
          if (!request.context) throw new Error("Missing contributor context");
          return page(graph, request.context.selector);
        });
      const fixed = options(graph);
      const hook = renderHook(() => useNetworkFlowGraphController(fixed));
      await waitFor(() => expect(hook.result.current.graph).toBe(graph));
      const first = hook.result.current.presentation.edges[0];
      if (!first) throw new Error("Missing first temporal edge");
      act(() => hook.result.current.selectGraphObject(first.selector));
      await waitFor(() => expect(calls).toHaveBeenCalledTimes(1));
      act(() => {
        hook.result.current.navigate({ type: "bucket", index: 1 });
        expect(calls.mock.calls[0]?.[0].signal?.aborted).toBe(true);
        expect(fixed.onNavigationContextChange.mock.lastCall?.[0]).toContain(
          "null",
        );
      });
      expect(hook.result.current.selection).toBeNull();
      expect(hook.result.current.contributorPage.committed).toBeNull();
      act(() => hook.result.current.navigate({ type: "bucket", index: 2 }));
      const current = hook.result.current.presentation.edges[0];
      if (!current) throw new Error("Missing later temporal edge");
      act(() => hook.result.current.selectGraphObject(current.selector));
      await waitFor(() =>
        expect(hook.result.current.contributorLoadState).toBe("ready"),
      );
      expect(hook.result.current.selection).toEqual(current.selector);
      act(() => {
        hook.result.current.navigate({ type: "bucket", index: 0 });
        hook.result.current.selectGraphObject(first.selector);
      });
      await waitFor(() => expect(calls).toHaveBeenCalledTimes(3));
      await waitFor(() =>
        expect(hook.result.current.contributorLoadState).toBe("ready"),
      );
      const accepted = hook.result.current.contributorPage.committed;
      const errors = fixed.onError.mock.calls.length;
      await act(async () => {
        if (failure)
          pending.reject(
            new NetworkFlowRequestError({
              code: "authorization_denied",
              status: 403,
              safeMessage: "Denied",
              retryAction: "do_not_retry",
              retryable: false,
            }),
          );
        else pending.resolve(page(graph, first.selector));
      });
      expect(hook.result.current.selection).toEqual(first.selector);
      expect(hook.result.current.contributorPage.committed).toBe(accepted);
      expect(accepted?.initialRequest).toMatchObject({
        selector: first.selector,
      });
      expect(fixed.onError).toHaveBeenCalledTimes(errors);
      expect(fixed.onIncidentAccessLost).not.toHaveBeenCalled();
      hook.unmount();
      vi.restoreAllMocks();
    }
  });
  it("pauses hidden reads and restores a noninitial bucket page selection and committed contributors", async () => {
    const graph = explorationFixture(501, 1001, true);
    const graphs = vi
      .spyOn(client, "queryNetworkFlowGraph")
      .mockResolvedValue(graph);
    const initial = graph.vertex_selectors[0]?.selector;
    if (!initial) throw new Error("Missing fixture selector");
    const pending = deferredSavedGraph<ContributorResult>();
    const calls = vi
      .spyOn(client, "queryNetworkFlowContributors")
      .mockResolvedValue(page(graph, initial));
    const fixed = options(graph);
    const hook = renderHook(
      ({ active, readIdentity }) =>
        useNetworkFlowGraphController({ ...fixed, active, readIdentity }),
      { initialProps: { active: true, readIdentity: "reader" } },
    );
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    act(() => {
      hook.result.current.navigate({ type: "bucket", index: 2 });
      hook.result.current.navigate({ type: "page", kind: "vertices", page: 1 });
    });
    const selected = hook.result.current.presentation.vertices[0];
    if (!selected) throw new Error("Missing page-two vertex");
    act(() => hook.result.current.selectGraphObject(selected.selector));
    await waitFor(() =>
      expect(hook.result.current.contributorLoadState).toBe("ready"),
    );
    calls.mockImplementationOnce(() => pending.promise);
    act(() => hook.result.current.nextContributorPage());
    await waitFor(() => expect(calls).toHaveBeenCalledTimes(2));
    act(() => hook.result.current.setActive(false));
    hook.rerender({ active: false, readIdentity: "reader" });
    expect(calls.mock.calls[1]?.[0].signal?.aborted).toBe(true);
    await act(async () => pending.resolve(page(graph, initial)));
    hook.rerender({ active: true, readIdentity: "reader" });
    expect([
      hook.result.current.navigation.bucketIndex,
      hook.result.current.navigation.vertexPage,
    ]).toEqual([2, 1]);
    expect(hook.result.current.selection).toEqual(selected.selector);
    expect(hook.result.current.contributorPage.pageNumber).toBe(1);
    expect(
      networkFlowVertexLinkCandidate(
        hook.result.current.graph,
        hook.result.current.selectedVertex,
      )?.candidateValue,
    ).toBe(selected.selector.endpoint_value);
    expect(graphs).toHaveBeenCalledTimes(1);
    expect(calls).toHaveBeenCalledTimes(2);
    hook.rerender({ active: false, readIdentity: "new-session" });
    expect(hook.result.current.graph).toBeNull();
    expect(hook.result.current.selection).toBeNull();
    expect(hook.result.current.contributors).toEqual([]);
    expect(graphs).toHaveBeenCalledTimes(1);
  });
  it("cancels an unfinished hidden graph and resumes only the missing initial read", async () => {
    const graph = explorationFixture(2, 1);
    const pending = deferredSavedGraph<client.NetworkFlowGraphResult>();
    const graphs = vi
      .spyOn(client, "queryNetworkFlowGraph")
      .mockImplementationOnce(() => pending.promise)
      .mockResolvedValue(graph);
    const fixed = options(graph);
    const hook = renderHook(
      ({ active }) => useNetworkFlowGraphController({ ...fixed, active }),
      { initialProps: { active: true } },
    );
    await waitFor(() => expect(graphs).toHaveBeenCalledTimes(1));
    act(() => hook.result.current.setActive(false));
    hook.rerender({ active: false });
    await act(async () => pending.resolve(graph));
    expect(hook.result.current.graph).toBeNull();
    expect(fixed.onQueryResult).not.toHaveBeenCalled();
    hook.rerender({ active: true });
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    expect(graphs).toHaveBeenCalledTimes(2);
    const initial = graph.vertex_selectors[0]?.selector;
    if (!initial) throw new Error("Missing selection");
    const unfinished = deferredSavedGraph<ContributorResult>();
    const reads = vi
      .spyOn(client, "queryNetworkFlowContributors")
      .mockImplementationOnce(() => unfinished.promise)
      .mockResolvedValue(page(graph, initial));
    act(() => hook.result.current.selectGraphObject(initial));
    await waitFor(() => expect(reads).toHaveBeenCalledTimes(1));
    act(() => hook.result.current.setActive(false));
    hook.rerender({ active: false });
    expect(reads.mock.calls[0]?.[0].signal?.aborted).toBe(true);
    hook.rerender({ active: true });
    await waitFor(() =>
      expect(hook.result.current.contributorLoadState).toBe("ready"),
    );
    const accepted = hook.result.current.contributorPage.committed;
    await act(async () => unfinished.resolve(page(graph, initial)));
    expect(hook.result.current.contributorPage.committed).toBe(accepted);
    expect(reads).toHaveBeenCalledTimes(2);
    expect(graphs).toHaveBeenCalledTimes(2);
  });
  it("withdraws navigation and pending contributors for replacement source staleness and authority changes", async () => {
    for (const reason of [
      "result",
      "source",
      "stale",
      "session",
      "incident",
      "role",
    ] as const) {
      const graph = explorationFixture(501, 1001, true);
      const pending = deferredSavedGraph<ContributorResult>();
      const graphs = vi
        .spyOn(client, "queryNetworkFlowGraph")
        .mockResolvedValue(graph);
      const reads = vi
        .spyOn(client, "queryNetworkFlowContributors")
        .mockImplementation(() => pending.promise);
      const fixed = options(graph);
      fixed.tableLifecycle.bind(
        {
          list: async () => [tableFixture()],
          submit: async () => tableFixture(),
        },
        tableAuthority,
      );
      await fixed.tableLifecycle.loadTables();
      const hook = renderHook(
        ({ readIdentity, incidentId, enabled }) =>
          useNetworkFlowGraphController({
            ...fixed,
            readIdentity,
            incidentId,
            enabled,
          }),
        {
          initialProps: {
            readIdentity: "editor-session",
            incidentId: tableIncidentId,
            enabled: true,
          },
        },
      );
      await waitFor(() => expect(hook.result.current.graph).toBe(graph));
      act(() => {
        hook.result.current.navigate({ type: "bucket", index: 2 });
        hook.result.current.navigate({
          type: "page",
          kind: "vertices",
          page: 1,
        });
      });
      const selected = hook.result.current.presentation.vertices[0];
      if (!selected) throw new Error("Missing bounded selection");
      act(() => hook.result.current.selectGraphObject(selected.selector));
      await waitFor(() => expect(reads).toHaveBeenCalledTimes(1));
      act(() => hook.result.current.navigate({ type: "reveal" }));
      const focus = hook.result.current.navigation.focus;
      if (!focus) throw new Error("Missing focus intent");
      const replacement = explorationFixture(2, 1);
      graphs.mockResolvedValue(replacement);
      if (reason === "result") act(() => hook.result.current.refreshGraph());
      else if (reason === "source")
        await act(() =>
          fixed.tableLifecycle.onResourceChange({
            resourceKind: "network_flow_table",
            resourceId: fixed.activeTableId,
            changeKind: "remove",
            reasonCode: "soft_deleted",
          }),
        );
      else if (reason === "stale")
        act(() => hook.result.current.markGraphStale());
      else
        hook.rerender({
          readIdentity: reason === "session" ? "new-session" : "editor-session",
          incidentId: reason === "incident" ? "new-incident" : tableIncidentId,
          enabled: reason !== "role",
        });
      expect(hook.result.current.selection, reason).toBeNull();
      expect(hook.result.current.navigation.focus).toBeNull();
      expect(hook.result.current.isFocusCurrent(focus)).toBe(false);
      expect(hook.result.current.contributors).toEqual([]);
      expect(reads.mock.calls[0]?.[0].signal?.aborted).toBe(true);
      await act(async () => pending.resolve(page(graph, selected.selector)));
      expect(hook.result.current.contributorPage.committed).toBeNull();
      expect(
        networkFlowVertexLinkCandidate(
          hook.result.current.graph,
          hook.result.current.selectedVertex,
        ),
      ).toBeNull();
      if (reason === "result") {
        await waitFor(() =>
          expect(hook.result.current.graph).toBe(replacement),
        );
        expect(hook.result.current.presentation.vertices).toHaveLength(2);
        expect([
          hook.result.current.navigation.bucketIndex,
          hook.result.current.navigation.vertexPage,
          hook.result.current.navigation.edgePage,
        ]).toEqual([0, 0, 0]);
      }
      if (reason === "stale" || reason === "source" || reason === "role")
        expect(hook.result.current.graph).toBeNull();
      hook.unmount();
      vi.restoreAllMocks();
    }
  });
});
