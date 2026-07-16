import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useNetworkFlowGridLayout } from "./useNetworkFlowGridLayout";

describe("useNetworkFlowGridLayout", () => {
  it("keeps show, move, and resize state for the mounted session and resets explicitly", () => {
    const { result } = renderHook(() =>
      useNetworkFlowGridLayout("network_flow.accepted_rows.v1"),
    );

    expect(
      result.current.visibleFieldKeys.has("network_flow.exporter_id"),
    ).toBe(false);
    act(() =>
      result.current.setColumnVisible("network_flow.exporter_id", true),
    );
    act(() =>
      result.current.onColumnReorder(
        "network_flow.dst_ip",
        "network_flow.flow_start_utc",
      ),
    );
    act(() =>
      result.current.onColumnWidthChange("network_flow.flow_start_utc", 40),
    );

    expect(
      result.current.visibleFieldKeys.has("network_flow.exporter_id"),
    ).toBe(true);
    expect(result.current.orderedVisibleFieldKeys[0]).toBe(
      "network_flow.dst_ip",
    );
    expect(result.current.columnWidths["network_flow.flow_start_utc"]).toBe(
      144,
    );

    act(() => result.current.reset());
    expect(
      result.current.visibleFieldKeys.has("network_flow.exporter_id"),
    ).toBe(false);
    expect(result.current.orderedVisibleFieldKeys[0]).toBe(
      "network_flow.flow_start_utc",
    );
    expect(result.current.columnWidths["network_flow.flow_start_utc"]).toBe(
      184,
    );
  });

  it("persists layout across component remounts in one browser session", () => {
    const first = renderHook(() =>
      useNetworkFlowGridLayout("network_flow.accepted_rows.v1"),
    );
    act(() =>
      first.result.current.setColumnVisible("network_flow.exporter_id", true),
    );
    expect(
      first.result.current.visibleFieldKeys.has("network_flow.exporter_id"),
    ).toBe(true);
    first.unmount();

    const reloaded = renderHook(() =>
      useNetworkFlowGridLayout("network_flow.accepted_rows.v1"),
    );
    expect(
      reloaded.result.current.visibleFieldKeys.has("network_flow.exporter_id"),
    ).toBe(true);
    act(() => reloaded.result.current.reset());
  });
});
