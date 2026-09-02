import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { workbookInspectorStateIsOpen } from "../models/workbookInspectorModel";
import { useWorkbookInspectorCoordinator } from "./useWorkbookInspectorCoordinator";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "./workbookInspectorSubject";

const timeline = requireViewContract("cartulary.view.timeline.v2");
const config = timeline.inspectorConfig;

function subject(rowVersion = 1): WorkbookInspectorSubject {
  const value = buildWorkbookInspectorSubject({
    config,
    kind: "live",
    label: "Timeline row",
    recordId: "row-1",
    rowVersion,
    surfaceLabel: timeline.title,
  });
  if (value === null) throw new Error("Expected a valid subject");
  return value;
}

describe("useWorkbookInspectorCoordinator", () => {
  it("coordinates explicit open, retarget, lifecycle invalidation, action completion, and focus restoration", () => {
    const resetOwnerState = vi.fn();
    const restoreFocus = vi.fn();
    const { result, rerender } = renderHook(
      ({ lifecycleKey, rowVersion }) =>
        useWorkbookInspectorCoordinator({
          actionPorts: { resetOwnerState, restoreFocus },
          config,
          lifecycleKey,
          subject: subject(rowVersion),
        }),
      { initialProps: { lifecycleKey: "base", rowVersion: 1 } },
    );
    expect(result.current.snapshot).toMatchObject({
      phase: "closed",
      subject: subject(),
    });
    act(() => result.current.commands.open());
    expect(result.current.snapshot.phase).toBe("open_ready");

    const generation = result.current.snapshot.invalidationGeneration;
    rerender({ lifecycleKey: "base", rowVersion: 2 });
    expect(result.current.snapshot.invalidationGeneration).toBe(generation + 1);
    expect(resetOwnerState).toHaveBeenCalledWith({
      cause: "retarget",
      scope: "row_local",
    });

    act(() => result.current.commands.completeAction());
    expect(restoreFocus).toHaveBeenCalledOnce();
    expect(workbookInspectorStateIsOpen(result.current.snapshot)).toBe(true);

    rerender({ lifecycleKey: "saved-view", rowVersion: 2 });
    expect(result.current.snapshot).toMatchObject({
      phase: "closed",
      subject: null,
    });
    expect(resetOwnerState).toHaveBeenLastCalledWith({
      cause: "surface_changed",
      scope: "surface",
    });
    act(() => result.current.commands.open());
    expect(result.current.snapshot).toMatchObject({
      phase: "open_ready",
      subject: subject(2),
    });
  });

  it("coalesces config and lifecycle changes and rejects stale commands", () => {
    const resetOwnerState = vi.fn();
    const { result, rerender } = renderHook(
      ({ activeConfig, lifecycleKey }) =>
        useWorkbookInspectorCoordinator({
          actionPorts: { resetOwnerState, restoreFocus: vi.fn() },
          config: activeConfig,
          lifecycleKey,
          subject: subject(),
        }),
      { initialProps: { activeConfig: config, lifecycleKey: "base" } },
    );
    act(() => result.current.commands.open());
    const staleClose = result.current.commands.close;
    resetOwnerState.mockClear();
    const hostsConfig = requireViewContract(
      "cartulary.view.hosts.v1",
    ).inspectorConfig;
    rerender({ activeConfig: hostsConfig, lifecycleKey: "saved-view" });
    expect(resetOwnerState).toHaveBeenCalledTimes(1);
    expect(resetOwnerState).toHaveBeenCalledWith({
      cause: "surface_changed",
      scope: "surface",
    });
    act(() => staleClose());
    expect(resetOwnerState).toHaveBeenCalledTimes(1);
    expect(result.current.snapshot.phase).toBe("closed");
  });

  it("closes idempotently and restores focus only when requested", () => {
    const resetOwnerState = vi.fn();
    const restoreFocus = vi.fn();
    const { result, rerender } = renderHook(
      ({ label }) =>
        useWorkbookInspectorCoordinator({
          actionPorts: { resetOwnerState, restoreFocus },
          config,
          lifecycleKey: "base",
          subject: { ...subject(), label },
        }),
      { initialProps: { label: "Original" } },
    );
    resetOwnerState.mockClear();
    rerender({ label: "Renamed" });
    expect(resetOwnerState).not.toHaveBeenCalled();
    act(() => result.current.commands.open());
    act(() => result.current.commands.close({ restoreFocus: true }));
    const generation = result.current.snapshot.invalidationGeneration;
    act(() => result.current.commands.close());
    expect(result.current.snapshot.invalidationGeneration).toBe(generation);
    expect(restoreFocus).toHaveBeenCalledOnce();
  });
});
