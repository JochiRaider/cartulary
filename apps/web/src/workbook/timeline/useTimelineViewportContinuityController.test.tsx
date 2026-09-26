import type { GridFocusResult, GridHandle } from "@cartulary/grid-adapter";
import { act, cleanup, renderHook } from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import {
  type TimelineViewportContinuityRequest,
  useTimelineViewportContinuityController,
} from "./hooks/useTimelineViewportContinuityController";
import { inputFocusKey } from "./models/timelineFieldRegistry";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  vi.useRealTimers();
  document.body.replaceChildren();
});

function harness() {
  vi.useFakeTimers();
  vi.spyOn(window, "focus").mockImplementation(() => {});
  const grid = document.createElement("div");
  grid.tabIndex = -1;
  grid.scrollTop = 120;
  grid.scrollLeft = 40;
  Object.defineProperty(grid, "getBoundingClientRect", {
    value: () => new DOMRect(0, 0, 600, 400),
  });
  const cell = document.createElement("div");
  cell.tabIndex = -1;
  cell.setAttribute("role", "gridcell");
  cell.dataset.gridRecordId = "saved";
  cell.dataset.gridFieldKey = "timeline.activity_synopsis_text";
  Object.defineProperty(cell, "getBoundingClientRect", {
    value: () => new DOMRect(80, 80, 200, 30),
  });
  const input = document.createElement("input");
  input.value = "Current native draft";
  document.body.append(grid, input);
  input.focus();
  input.setSelectionRange(2, 6, "backward");
  const registry = createTimelineEditorDraftRegistry();
  const scopeState = { key: "scope", readable: true };
  const scopeListeners = new Set<() => void>();
  const presentation = {
    surface: {
      kind: "view_schema",
      viewSchemaId: timelineViewSchemaId,
    } as const,
    rowIdentities: [{ kind: "core_record", recordId: "saved" }] as const,
    fieldKeys: ["timeline.activity_synopsis_text"],
    revision: 1,
  };
  const waiting = new Set<() => void>();
  const requestFocus = vi.fn<GridHandle["requestFocus"]>(
    (target, options) =>
      new Promise<GridFocusResult>((resolve) => {
        const finish = (result: GridFocusResult) => {
          waiting.delete(ready);
          options?.signal?.removeEventListener("abort", abort);
          resolve(result);
        };
        const abort = () => finish("cancelled");
        const ready = () => {
          if (options?.signal?.aborted) return abort();
          const element = target.kind === "root" ? grid : cell;
          if (!element.isConnected) return;
          element.focus({ preventScroll: true });
          finish("focused");
        };
        options?.signal?.addEventListener("abort", abort, { once: true });
        waiting.add(ready);
        ready();
      }),
  );
  const gridHandleRef = {
    current: {
      requestFocus,
      getScrollElement: () => grid,
      getAnchorRect: () =>
        cell.isConnected ? cell.getBoundingClientRect() : null,
      scrollToAnchor: vi.fn(() => false),
      presentation: {
        getSnapshot: () => presentation,
        subscribe: () => () => {},
      },
    } as unknown as GridHandle,
  };
  const gridShellRef = { current: grid };
  const viewportContinuityTokenRef = { current: 1 };
  const scope = {
    getSnapshot: () => scopeState,
    subscribe: (listener: () => void) => {
      scopeListeners.add(listener);
      return () => {
        scopeListeners.delete(listener);
      };
    },
  };
  const hook = renderHook(() => {
    const [request, setRequest] =
      useState<TimelineViewportContinuityRequest | null>(null);
    const controller = useTimelineViewportContinuityController({
      gridHandleRef,
      gridShellRef,
      scope,
      editorDraftRegistry: registry,
      viewportContinuityTokenRef,
      viewportContinuityRequest: request,
      setViewportContinuityRequest: setRequest,
    });
    return { ...controller.commands, request };
  });
  const begin = () => {
    let token = 0;
    act(() => {
      token = hook.result.current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: "saved",
      });
      hook.result.current.advanceViewportContinuity(token);
    });
    expect(requestFocus).toHaveBeenCalledTimes(1);
    expect(waiting.size).toBe(1);
    expect(cell.isConnected).toBe(false);
    return token;
  };
  const mount = async () => {
    await act(async () => {
      grid.append(cell);
      for (const ready of waiting) ready();
      await vi.runAllTimersAsync();
    });
  };
  const publishScope = () =>
    act(() => {
      for (const listener of scopeListeners) listener();
    });
  return {
    ...hook,
    grid,
    cell,
    input,
    registry,
    requestFocus,
    begin,
    mount,
    scopeState,
    publishScope,
    presentation,
    gridHandleRef,
  };
}

describe("Timeline deferred continuity", () => {
  it("cancels an already pending semantic target for wheel native input and composition before it mounts", async () => {
    for (const event of ["wheel", "input", "compositionstart"]) {
      const h = harness();
      h.begin();
      const signal = h.requestFocus.mock.calls[0]?.[1]?.signal;
      act(() => {
        h.grid.scrollTop = 250;
        h.grid.scrollLeft = 90;
        h.input.dispatchEvent(new Event(event, { bubbles: true }));
      });
      await h.mount();
      expect(document.activeElement).toBe(h.input);
      expect([h.grid.scrollTop, h.grid.scrollLeft]).toEqual([250, 90]);
      expect([
        h.input.value,
        h.input.selectionStart,
        h.input.selectionEnd,
        h.input.selectionDirection,
      ]).toEqual(["Current native draft", 2, 6, "backward"]);
      expect(signal?.aborted).toBe(true);
      expect(h.requestFocus).toHaveBeenCalledTimes(1);
      expect(h.result.current.request).toBeNull();
      h.unmount();
      h.grid.remove();
      h.input.remove();
    }
  });

  it("yields a pending row destination to external focus without a pointer or key event", async () => {
    const h = harness();
    h.begin();
    const shell = document.createElement("button");
    document.body.append(shell);
    act(() => shell.focus());
    await h.mount();
    expect(document.activeElement).toBe(shell);
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    expect(h.result.current.request).toBeNull();
  });

  it("cancels pending focus for pointer keyboard scope authority replacement and unmount", async () => {
    for (const cause of [
      "pointerdown",
      "keydown",
      "scope",
      "authority",
      "replacement",
      "unmount",
    ]) {
      const h = harness();
      h.begin();
      const signal = h.requestFocus.mock.calls[0]?.[1]?.signal;
      act(() => {
        if (cause === "scope") {
          h.scopeState.key = "new-scope";
          h.rerender();
        } else if (cause === "authority") {
          h.scopeState.readable = false;
          h.publishScope();
        } else if (cause === "unmount") h.unmount();
        else if (cause === "replacement")
          h.result.current.beginViewportContinuity({
            kind: "input",
            focusKey: "draft-2:activitySynopsisText:grid",
          });
        else h.input.dispatchEvent(new Event(cause, { bubbles: true }));
      });
      expect(signal?.aborted).toBe(true);
      await h.mount();
      expect(document.activeElement).toBe(h.input);
      expect(h.requestFocus).toHaveBeenCalledTimes(1);
      if (cause !== "unmount") h.unmount();
      h.grid.remove();
      h.input.remove();
    }
  });

  it("checks ownership after focus acknowledgement before changing geometry", async () => {
    const h = harness();
    let acknowledge: (result: GridFocusResult) => void = () => {};
    h.requestFocus.mockImplementationOnce(
      (_target, options) =>
        new Promise((resolve) => {
          acknowledge = resolve;
          expect(options?.signal?.aborted).toBe(false);
        }),
    );
    act(() => {
      const token = h.result.current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: "saved",
      });
      h.result.current.advanceViewportContinuity(token);
    });
    act(() => {
      h.grid.scrollTop = 310;
      h.grid.scrollLeft = 170;
      h.input.dispatchEvent(new Event("input", { bubbles: true }));
    });
    await act(async () => {
      acknowledge("focused");
      await vi.runAllTimersAsync();
    });
    expect([h.grid.scrollTop, h.grid.scrollLeft]).toEqual([310, 170]);
    expect(document.activeElement).toBe(h.input);
    expect(h.result.current.request).toBeNull();
  });

  it("cancels queued stabilization and late named follow-ups without resurrecting restoration", async () => {
    const h = harness();
    let token = 0;
    await act(async () => {
      token = h.result.current.beginViewportContinuity(
        { kind: "scroll-only" },
        { requirements: ["entity-refresh"] },
      );
      h.result.current.advanceViewportContinuity(token);
    });
    expect(vi.getTimerCount()).toBeGreaterThan(0);
    act(() => {
      h.grid.scrollTop = 360;
      h.grid.scrollLeft = 180;
      h.input.dispatchEvent(new Event("wheel", { bubbles: true }));
    });
    await act(async () => {
      h.result.current.settleViewportContinuityFollowUp(
        token,
        "entity-refresh",
        "settled",
      );
      h.result.current.advanceViewportContinuity(token);
      await vi.runAllTimersAsync();
    });
    expect([h.grid.scrollTop, h.grid.scrollLeft]).toEqual([360, 180]);
    expect(h.result.current.request).toBeNull();
    expect(h.requestFocus).not.toHaveBeenCalled();
    expect(vi.getTimerCount()).toBe(0);
  });

  it("keeps owned focus through authoritative row and named follow-up settlement", async () => {
    const h = harness();
    let token = 0;
    act(() => {
      token = h.result.current.beginViewportContinuity(
        { kind: "row-inspect", recordId: "saved" },
        { requirements: ["entity-refresh"] },
      );
      h.result.current.requireViewportContinuitySourceRecord(token, {
        recordId: "saved",
        minimumRowVersion: 4,
      });
      h.result.current.advanceViewportContinuity(token, {
        sourceRecord: { recordId: "saved", rowVersion: 3 },
      });
    });
    await h.mount();
    expect(document.activeElement).toBe(h.cell);
    expect(h.result.current.request).not.toBeNull();
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    expect(h.requestFocus.mock.calls[0]?.[1]?.preserveSelection).toBe(true);
    await act(async () => {
      h.result.current.advanceViewportContinuity(token, {
        sourceRecord: { recordId: "saved", rowVersion: 4 },
      });
      h.result.current.settleViewportContinuityFollowUp(
        token,
        "entity-refresh",
        "settled",
      );
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(h.result.current.request).toBeNull();
    expect(document.activeElement).toBe(h.cell);
    expect([h.grid.scrollTop, h.grid.scrollLeft]).toEqual([120, 40]);
    expect(h.requestFocus).toHaveBeenCalledTimes(2);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("recovers a newly registered input but cancels its bounded retry for newer authoring", async () => {
    for (const interrupt of [false, true]) {
      const h = harness();
      const identity = {
        rowKey: "saved",
        field: "activitySynopsisText",
        surface: "grid",
      } as const;
      const focusKey = inputFocusKey(
        identity.rowKey,
        identity.field,
        identity.surface,
      );
      h.registry.setDraft(identity, "Retained exact Ω");
      act(() => {
        const token = h.result.current.beginViewportContinuity({
          kind: "input",
          focusKey,
        });
        h.result.current.advanceViewportContinuity(token);
      });
      await act(async () => {});
      expect(vi.getTimerCount()).toBeGreaterThan(0);
      if (interrupt)
        act(() => h.input.dispatchEvent(new Event("input", { bubbles: true })));
      const recovered = document.createElement("input");
      recovered.value = "Retained exact Ω";
      h.grid.append(recovered);
      h.registry.registerInput(identity, recovered);
      await act(async () => {
        await vi.runAllTimersAsync();
      });
      expect(document.activeElement).toBe(interrupt ? h.input : recovered);
      expect(h.registry.draftValue(identity)).toBe("Retained exact Ω");
      expect(h.result.current.request).toBeNull();
      expect(vi.getTimerCount()).toBe(0);
      h.unmount();
      h.grid.remove();
      h.input.remove();
    }
  });

  it("uses a visible fallback after bounded editor readiness and never reopens the editor", async () => {
    const h = harness();
    const identity = {
      rowKey: "saved",
      field: "activitySynopsisText",
      surface: "grid",
    } as const;
    h.grid.append(h.cell);
    act(() => {
      const token = h.result.current.beginViewportContinuity({
        kind: "input",
        focusKey: inputFocusKey(
          identity.rowKey,
          identity.field,
          identity.surface,
        ),
      });
      h.result.current.advanceViewportContinuity(token);
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(document.activeElement).toBe(h.cell);
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    expect(h.result.current.request).toBeNull();
    const late = document.createElement("input");
    h.grid.append(late);
    h.registry.registerInput(identity, late);
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(document.activeElement).toBe(h.cell);
    expect(vi.getTimerCount()).toBe(0);
  });

  it("retires removed targets and cancels the whole unavailable fallback chain", async () => {
    const h = harness();
    let unavailable: (result: GridFocusResult) => void = () => {};
    h.requestFocus.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          unavailable = resolve;
        }),
    );
    act(() => {
      const token = h.result.current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: "saved",
      });
      h.result.current.advanceViewportContinuity(token);
    });
    act(() => h.result.current.interruptViewportContinuity());
    await act(async () => {
      unavailable("unavailable");
      await vi.runAllTimersAsync();
    });
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    expect(h.result.current.request).toBeNull();
    await act(async () => {
      const token = h.result.current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: "removed",
      });
      h.result.current.advanceViewportContinuity(token);
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(h.requestFocus).toHaveBeenLastCalledWith(
      { kind: "root" },
      expect.objectContaining({ preserveSelection: true }),
    );
    expect(document.activeElement).toBe(h.grid);
    expect(h.result.current.request).toBeNull();
    await h.mount();
    expect(document.activeElement).toBe(h.grid);
  });

  it("cancels acceptance-time fresh draft continuation independently of the accepted result", async () => {
    const h = harness();
    let token = 0;
    act(() => {
      token = h.result.current.beginViewportContinuity({ kind: "scroll-only" });
    });
    act(() => h.input.dispatchEvent(new Event("wheel", { bubbles: true })));
    act(() =>
      h.result.current.completeAcceptedViewportContinuity(token, {
        kind: "fresh_draft",
        focusKey: "draft-2:activitySynopsisText:grid",
        recordId: "saved",
      }),
    );
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(h.requestFocus).not.toHaveBeenCalled();
    expect(h.result.current.request).toBeNull();
    expect(document.activeElement).toBe(h.input);
    expect([h.grid.scrollTop, h.grid.scrollLeft]).toEqual([120, 40]);
  });

  it("reuses acknowledged focus while transient geometry becomes ready", async () => {
    const h = harness();
    let ready = false;
    Object.defineProperty(h.gridHandleRef.current, "getAnchorRect", {
      value: () => (ready ? h.cell.getBoundingClientRect() : null),
    });
    h.grid.append(h.cell);
    await act(async () => {
      const token = h.result.current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: "saved",
      });
      h.result.current.advanceViewportContinuity(token);
    });
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(100);
    });
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    ready = true;
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
    expect(document.activeElement).toBe(h.cell);
    expect(h.result.current.request).toBeNull();
  });

  it("continues an uninterrupted create through adapter draft readiness without pre-create scroll", async () => {
    const h = harness();
    const identity = {
      rowKey: "draft-2",
      field: "activitySynopsisText",
      surface: "grid",
    } as const;
    const input = document.createElement("input");
    h.grid.append(input);
    Object.defineProperty(input, "getBoundingClientRect", {
      value: () => new DOMRect(80, 80, 200, 30),
    });
    h.registry.registerInput(identity, input);
    h.requestFocus.mockImplementationOnce(async (target) => {
      expect(target).toEqual({
        kind: "draft",
        fieldKey: "timeline.activity_synopsis_text",
      });
      h.grid.scrollTop = 500;
      input.focus();
      return "focused";
    });
    await act(async () => {
      const token = h.result.current.beginViewportContinuity({
        kind: "scroll-only",
      });
      h.result.current.completeAcceptedViewportContinuity(token, {
        kind: "fresh_draft",
        focusKey: inputFocusKey(
          identity.rowKey,
          identity.field,
          identity.surface,
        ),
        recordId: "saved",
      });
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(document.activeElement).toBe(input);
    expect(h.grid.scrollTop).toBe(500);
    expect(h.result.current.request).toBeNull();
    expect(h.requestFocus).toHaveBeenCalledTimes(1);
  });

  it("settles terminal follow-up failure at a visible row fallback without stealing nested control focus", async () => {
    const h = harness();
    let token = 0;
    act(() => {
      token = h.result.current.beginViewportContinuity(
        { kind: "row-inspect", recordId: "saved" },
        { requirements: ["entity-refresh"] },
      );
      h.result.current.advanceViewportContinuity(token);
    });
    await h.mount();
    expect(h.result.current.request).not.toBeNull();
    await act(async () => {
      h.result.current.failViewportContinuity(token);
    });
    await act(async () => {
      await vi.runAllTimersAsync();
    });
    expect(h.result.current.request).toBeNull();
    expect(document.activeElement).toBe(h.cell);
    act(() => {
      token = h.result.current.beginViewportContinuity(
        { kind: "row-inspect", recordId: "saved" },
        { requirements: ["entity-refresh"] },
      );
      h.result.current.advanceViewportContinuity(token);
    });
    await act(async () => {});
    const nested = document.createElement("button");
    h.cell.append(nested);
    act(() => nested.focus());
    await act(async () => {
      h.result.current.settleViewportContinuityFollowUp(
        token,
        "entity-refresh",
        "settled",
      );
      await vi.runAllTimersAsync();
    });
    expect(document.activeElement).toBe(nested);
    expect(h.result.current.request).toBeNull();
  });
});
