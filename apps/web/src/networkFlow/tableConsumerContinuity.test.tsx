import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { explorationFixture } from "./explorationTestFixtures";
import { NetworkFlowTableController } from "./NetworkFlowTableController";
import * as client from "./networkFlowClient";
import { defaultGraphQuerySettings } from "./networkFlowQueryModel";
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
describe("Table consumer continuity", () => {
  it("preserves graph identity through rename and unrelated removal then requires affected recomputation", async () => {
    const graph = explorationFixture(2, 1);
    const request = vi
      .spyOn(client, "queryNetworkFlowGraph")
      .mockResolvedValue(graph);
    const owner = new NetworkFlowTableController();
    owner.bind(
      {
        list: async () => [tableFixture(), tableFixture("b", 1, "Other")],
        submit: async () => tableFixture(),
      },
      tableAuthority,
    );
    await owner.loadTables();
    const fixed = {
      settings: defaultGraphQuerySettings,
      applicationRevision: 0,
      revision: 0,
      onQueryResult: vi.fn(),
      availability: readyExtensionAvailability(tableIncidentId),
      tableLifecycle: owner,
      activeTableId: tableFixture().network_flow_table_id,
      apiBase: undefined,
      enabled: true,
      incidentId: tableIncidentId,
      onError: vi.fn(),
      onIncidentAccessLost: undefined,
      query: { filters: [], sort: [], timeWindow: null },
    };
    const hook = renderHook(
      ({ tables }) => useNetworkFlowGraphController({ ...fixed, tables }),
      {
        initialProps: {
          tables: [tableFixture(), tableFixture("b", 1, "Other")],
        },
      },
    );
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    hook.rerender({
      tables: [
        tableFixture("a", 2, "Peer name"),
        tableFixture("b", 1, "Other"),
      ],
    });
    expect(request).toHaveBeenCalledTimes(1);
    expect(hook.result.current.graph).toBe(graph);
    await act(() =>
      owner.onResourceChange({
        resourceKind: "network_flow_table",
        resourceId: tableFixture("b").network_flow_table_id,
        changeKind: "remove",
        reasonCode: "soft_deleted",
      }),
    );
    hook.rerender({ tables: [tableFixture("a", 2, "Peer name")] });
    expect(hook.result.current.graphStale).toBe(false);
    expect(request).toHaveBeenCalledTimes(1);
    await act(() =>
      owner.onResourceChange({
        resourceKind: "network_flow_table",
        resourceId: tableFixture().network_flow_table_id,
        changeKind: "remove",
        reasonCode: "soft_deleted",
      }),
    );
    fixed.revision += 1;
    hook.rerender({ tables: [tableFixture("a", 2, "Peer name")] });
    expect(hook.result.current.graphStale).toBe(true);
    expect(hook.result.current.graph).toBeNull();
    expect(hook.result.current.selection).toBeNull();
    expect(request).toHaveBeenCalledTimes(1);
    act(() => hook.result.current.refreshGraph());
    await waitFor(() => expect(request).toHaveBeenCalledTimes(2));
    expect(hook.result.current.graphStale).toBe(false);
  });
  it("tracks all-active membership independently of names and resolves fresh sources on recompute", async () => {
    const graph = explorationFixture(2, 1);
    const request = vi
      .spyOn(client, "queryNetworkFlowGraph")
      .mockResolvedValue(graph);
    const owner = new NetworkFlowTableController();
    const fixed = {
      settings: defaultGraphQuerySettings,
      applicationRevision: 0,
      revision: 0,
      onQueryResult: vi.fn(),
      availability: readyExtensionAvailability(tableIncidentId),
      tableLifecycle: owner,
      activeTableId: tableFixture().network_flow_table_id,
      apiBase: undefined,
      enabled: true,
      incidentId: tableIncidentId,
      onError: vi.fn(),
      onIncidentAccessLost: undefined,
      query: { filters: [], sort: [], timeWindow: null },
    };
    const hook = renderHook(
      ({ tables }) => useNetworkFlowGraphController({ ...fixed, tables }),
      { initialProps: { tables: [tableFixture()] } },
    );
    await waitFor(() => expect(hook.result.current.graph).toBe(graph));
    fixed.settings = {
      ...defaultGraphQuerySettings,
      scopeMode: "all_active_tables",
    };
    fixed.revision += 1;
    hook.rerender({ tables: [tableFixture()] });
    await waitFor(() => expect(request).toHaveBeenCalledTimes(2));
    hook.rerender({ tables: [tableFixture("a", 2, "Metadata")] });
    expect(request).toHaveBeenCalledTimes(2);
    hook.rerender({
      tables: [
        tableFixture("a", 2, "Metadata"),
        tableFixture("b", 1, "Imported"),
      ],
    });
    expect(hook.result.current.graphStale).toBe(true);
    expect(request).toHaveBeenCalledTimes(2);
    act(() => hook.result.current.refreshGraph());
    await waitFor(() => expect(request).toHaveBeenCalledTimes(3));
    expect(request.mock.calls[2]?.[0].tableScope).toEqual({
      mode: "all_active_tables",
    });
  });
});
