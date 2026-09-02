import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useWorkbookInspectorCoordinator } from "./useWorkbookInspectorCoordinator";

const config = requireViewContract(
  "cartulary.view.timeline.v2",
).inspectorConfig;
const subject = {
  recordId: "row-1",
  rowVersion: 1,
  viewSchemaId: config.viewSchemaId,
};

describe("useWorkbookInspectorCoordinator", () => {
  it("coordinates explicit open, retarget, lifecycle invalidation, action completion, and focus restoration", () => {
    const resetOwnerState = vi.fn();
    const restoreFocus = vi.fn();
    const { result, rerender } = renderHook(
      ({ lifecycleKey, rowVersion }) =>
        useWorkbookInspectorCoordinator({
          actionPorts: {
            resetOwnerState,
            restoreFocus,
          },
          config,
          lifecycleKey,
          subject: { ...subject, rowVersion },
        }),
      { initialProps: { lifecycleKey: "base", rowVersion: 1 } },
    );

    expect(result.current.snapshot.isOpen).toBe(false);
    expect(result.current.snapshot.subject).toEqual(subject);

    act(() => result.current.commands.open());
    expect(result.current.snapshot).toMatchObject({
      isOpen: true,
      status: "ready",
    });

    const generation = result.current.snapshot.invalidationGeneration;
    rerender({ lifecycleKey: "base", rowVersion: 2 });
    expect(result.current.snapshot.invalidationGeneration).toBe(generation + 1);
    expect(resetOwnerState).toHaveBeenCalledWith({
      cause: "retarget",
      scope: "row_local",
    });

    act(() => result.current.commands.completeAction());
    expect(restoreFocus).toHaveBeenCalledOnce();
    expect(result.current.snapshot.isOpen).toBe(true);

    rerender({ lifecycleKey: "saved-view", rowVersion: 2 });
    expect(result.current.snapshot).toMatchObject({
      isOpen: false,
      status: "closed",
      subject: null,
    });
    expect(resetOwnerState).toHaveBeenLastCalledWith({
      cause: "surface_changed",
      scope: "surface",
    });
  });

  it("closes idempotently and restores focus only when requested", () => {
    const restoreFocus = vi.fn();
    const { result } = renderHook(() =>
      useWorkbookInspectorCoordinator({
        actionPorts: { resetOwnerState: vi.fn(), restoreFocus },
        config,
        lifecycleKey: "base",
        subject: null,
      }),
    );

    act(() => result.current.commands.open());
    expect(result.current.snapshot.status).toBe("no_row_selected");
    act(() => result.current.commands.close({ restoreFocus: true }));
    expect(result.current.snapshot.status).toBe("closed");
    expect(restoreFocus).toHaveBeenCalledOnce();
  });
});
