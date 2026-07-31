import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { selectInspectorConfig } from "../models/workbookInspectorModel";
import { useWorkbookInspectorCoordinator } from "./useWorkbookInspectorCoordinator";

const config = selectInspectorConfig(
  requireViewContract("cartulary.view.timeline.v2"),
);
const subject = {
  recordId: "row-1",
  rowVersion: 1,
  viewSchemaId: config.viewSchemaId,
};

describe("useWorkbookInspectorCoordinator", () => {
  it("coordinates explicit open, retarget, lifecycle invalidation, action completion, and focus restoration", () => {
    const clearLocalForm = vi.fn();
    const clearMergePlan = vi.fn();
    const clearPendingConfirmation = vi.fn();
    const clearPreview = vi.fn();
    const clearWorkflowForm = vi.fn();
    const restoreFocus = vi.fn();
    const { result, rerender } = renderHook(
      ({ lifecycleKey, rowVersion }) =>
        useWorkbookInspectorCoordinator({
          actionPorts: {
            clearLocalForm,
            clearMergePlan,
            clearPendingConfirmation,
            clearPreview,
            clearWorkflowForm,
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

    act(() => result.current.commands.open("history"));
    expect(result.current.snapshot).toMatchObject({
      activePanelId: "history",
      isOpen: true,
      status: "ready",
    });

    const generation = result.current.snapshot.invalidationGeneration;
    rerender({ lifecycleKey: "base", rowVersion: 2 });
    expect(result.current.snapshot.invalidationGeneration).toBe(generation + 1);
    expect(clearPendingConfirmation).toHaveBeenCalled();
    expect(clearPreview).toHaveBeenCalled();
    expect(clearMergePlan).toHaveBeenCalled();
    expect(clearWorkflowForm).toHaveBeenCalled();
    expect(clearLocalForm).toHaveBeenCalled();

    act(() => result.current.commands.completeAction());
    expect(restoreFocus).toHaveBeenCalledOnce();
    expect(result.current.snapshot.isOpen).toBe(true);

    rerender({ lifecycleKey: "saved-view", rowVersion: 2 });
    expect(result.current.snapshot).toMatchObject({
      isOpen: false,
      status: "closed",
      subject: null,
    });
  });

  it("closes idempotently and restores focus only when requested", () => {
    const restoreFocus = vi.fn();
    const { result } = renderHook(() =>
      useWorkbookInspectorCoordinator({
        actionPorts: { restoreFocus },
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
