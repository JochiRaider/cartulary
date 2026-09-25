import { renderHook } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { pendingReplayCapacity } from "../runtime/pending/workbookPendingQueue";
import {
  beginWorkbookPendingRefreshBlock,
  finishWorkbookPendingRefreshBlock,
} from "../runtime/workbookPendingReplayRuntime";
import { useTimelineCommittedRecordIdle } from "./hooks/useTimelineCommittedRecordIdle";

const runtimes: ReturnType<typeof createWorkbookMutationRuntime>[] = [];
afterEach(() => {
  for (const runtime of runtimes.splice(0))
    runtime.invalidate({ kind: "runtime_disposed" });
  vi.restoreAllMocks();
});

function fixture() {
  const microtasks: (() => void)[] = [];
  const scheduleDelay = vi.fn(() => () => {});
  let sequence = 0;
  const runtime = createWorkbookMutationRuntime(
    { clientInstanceId: "client-1", incidentId: "incident-1" },
    { create: () => `txn-${++sequence}` },
    { execute: vi.fn() },
    {
      clock: { now: () => 1 },
      scheduler: {
        enqueueMicrotask: (work) => microtasks.push(work),
        scheduleDelay,
      },
    },
  );
  runtimes.push(runtime);
  const flush = () => {
    while (microtasks.length) microtasks.shift()?.();
  };
  const signal = new AbortController().signal;
  const wait = (recordId = "record-1", currentSignal = signal) =>
    runtime.waitForPendingRecordIdle({
      recordId,
      viewSchemaId: timelineViewSchemaId,
      signal: currentSignal,
    });
  const enqueue = (recordId = "record-1") => {
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [{ field_key: "summary", value: "draft" }],
      fieldKey: "summary",
      localValue: "draft",
      recordId,
      rowLabel: "Row",
      surfaceLabel: "Timeline",
      viewSchemaId: timelineViewSchemaId,
    });
    return (
      runtime.pendingQueue().model.snapshot().units.at(-1)?.id ?? "missing"
    );
  };
  return { runtime, flush, wait, enqueue, scheduleDelay, signal };
}

it("useTimelineCommittedRecordIdle refreshes a missing committed version at most once", async () => {
  const { runtime, signal } = fixture();
  let rowVersion: number | null = null;
  const loadRows = vi.fn(async () => {
    rowVersion = 7;
  });
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      mutationRuntime: runtime,
      latestCommittedRowVersion: () => rowVersion,
      latestCommittedTimelineRow: () => null,
      loadRows,
    }),
  );
  await expect(result.current("record-1", { signal })).resolves.toEqual({
    row: null,
    rowVersion: 7,
  });
  expect(loadRows).toHaveBeenCalledTimes(1);
  rowVersion = null;
  await expect(
    result.current("record-1", { signal, refreshIfMissing: false }),
  ).resolves.toBeNull();
  await expect(
    result.current("record-1", { signal, fallbackRowVersion: 8 }),
  ).resolves.toEqual({ row: null, rowVersion: 8 });
  expect(loadRows).toHaveBeenCalledTimes(1);
  loadRows.mockImplementation(async () => {});
  await expect(result.current("record-1", { signal })).resolves.toBeNull();
  expect(loadRows).toHaveBeenCalledTimes(2);
});

it("pending record readiness follows queue and scoped refresh publications without timers", async () => {
  const { runtime, enqueue, wait, flush, scheduleDelay } = fixture();
  await expect(wait()).resolves.toBe("idle");
  const unit = enqueue();
  const completed = vi.fn();
  const waiting = wait().then(completed);
  await expect(wait("unrelated-record")).resolves.toBe("idle");
  const pending = runtime.pendingQueue();
  const scope = { kind: "record" as const, recordId: "record-1" };
  beginWorkbookPendingRefreshBlock(pending, scope);
  runtime.notifyPendingChanged();
  pending.model.settleUnchanged(unit);
  flush();
  await Promise.resolve();
  expect(completed).not.toHaveBeenCalled();
  finishWorkbookPendingRefreshBlock(pending, scope);
  runtime.notifyPendingChanged();
  flush();
  await waiting;
  expect(completed).toHaveBeenCalledExactlyOnceWith("idle");
  expect(scheduleDelay).not.toHaveBeenCalled();
  const global = { kind: "all" as const };
  beginWorkbookPendingRefreshBlock(pending, global);
  const globallyWaiting = wait("other-record").then(completed);
  flush();
  await Promise.resolve();
  expect(completed).toHaveBeenCalledTimes(1);
  finishWorkbookPendingRefreshBlock(pending, global);
  runtime.notifyPendingChanged();
  flush();
  await globallyWaiting;
  expect(completed).toHaveBeenCalledTimes(2);
});

it("pending record readiness observes canonical conflicts without a React projection", async () => {
  const { runtime, wait, flush } = fixture();
  const register = (viewSchemaId: string) =>
    runtime.registerConflict({
      viewSchemaId,
      rowLabel: "Row",
      surfaceLabel: "Surface",
      conflict: {
        record_id: "record-2",
        field_key: "summary",
        base_row_version: 1,
        current_row_version: 2,
        client_value: "draft",
        server_value: "saved",
        conflict_token: "token",
        conflict_resolution_class: "text_compare_merge",
      },
    });
  register("other-view");
  await expect(wait()).resolves.toBe("idle");
  register(timelineViewSchemaId);
  await expect(wait()).resolves.toBe("blocked");
  runtime.clearConflict(runtime.getSnapshot().conflicts[0]?.key ?? "missing");
  flush();
  await expect(wait()).resolves.toBe("idle");
  runtime.pendingQueue().model.pauseForAuthRecovery();
  await expect(wait()).resolves.toBe("blocked");
});

it("pending record readiness cancels independent waiters on abort authority change and retirement", async () => {
  const { runtime, enqueue, wait, flush, scheduleDelay } = fixture();
  enqueue();
  const first = new AbortController();
  const remove = vi.spyOn(first.signal, "removeEventListener");
  const cancelled = wait("record-1", first.signal);
  const second = wait();
  first.abort();
  await expect(cancelled).resolves.toBe("cancelled");
  expect(remove).toHaveBeenCalledWith("abort", expect.any(Function));
  await expect(wait("record-1", first.signal)).resolves.toBe("cancelled");
  runtime.setAuthority({
    actorId: "actor",
    sessionIdentity: "session",
    incidentId: "incident-1",
    role: "editor",
    closed: false,
  });
  flush();
  await expect(second).resolves.toBe("cancelled");
  const scheduledBeforeRetirement = scheduleDelay.mock.calls.length;
  const retired = wait();
  runtime.invalidate({ kind: "runtime_disposed" });
  await expect(retired).resolves.toBe("cancelled");
  await expect(wait()).resolves.toBe("cancelled");
  flush();
  expect(scheduleDelay).toHaveBeenCalledTimes(scheduledBeforeRetirement);
});

it("committed record waiting accepts the latest version before resolving a dependent action", async () => {
  const { runtime, enqueue, flush, signal } = fixture();
  const unit = enqueue();
  let version = 1;
  const loadRows = vi.fn();
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      mutationRuntime: runtime,
      latestCommittedRowVersion: () => version,
      latestCommittedTimelineRow: () => null,
      loadRows,
    }),
  );
  const action = vi.fn();
  const timeout = vi.spyOn(globalThis, "setTimeout");
  const interval = vi.spyOn(globalThis, "setInterval");
  const waiting = result.current("record-1", { signal }).then(action);
  version = 9;
  runtime.pendingQueue().model.settleUnchanged(unit);
  flush();
  await waiting;
  expect(action).toHaveBeenCalledExactlyOnceWith({ row: null, rowVersion: 9 });
  expect(loadRows).not.toHaveBeenCalled();
  action.mockClear();
  // Readiness initially resolves synchronously. A later publication before its
  // promise continuation must still prevent accepting stale committed evidence.
  const raced = result.current("record-1", { signal }).then(action);
  const next = enqueue();
  await Promise.resolve();
  expect(action).not.toHaveBeenCalled();
  version = 10;
  runtime.pendingQueue().model.settleUnchanged(next);
  flush();
  await raced;
  expect(action).toHaveBeenCalledExactlyOnceWith({ row: null, rowVersion: 10 });
  expect(timeout).not.toHaveBeenCalled();
  expect(interval).not.toHaveBeenCalled();
});

it("committed record waiting cancels during a shared refresh and ignores late rejection", async () => {
  const { runtime } = fixture();
  let version: number | null = null;
  const refresh = deferred<void>();
  const entered = deferred<void>();
  const loadRows = vi.fn(() => {
    entered.resolve();
    return refresh.promise;
  });
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      mutationRuntime: runtime,
      latestCommittedRowVersion: () => version,
      latestCommittedTimelineRow: () => null,
      loadRows,
    }),
  );
  const controller = new AbortController();
  const waiting = result.current("record-1", { signal: controller.signal });
  await entered.promise;
  controller.abort();
  await expect(waiting).resolves.toBeNull();
  refresh.reject(new Error("late read failure"));
  await Promise.resolve();
  expect(loadRows).toHaveBeenCalledOnce();
  const shared = deferred<void>();
  const bothReading = deferred<void>();
  let reads = 0;
  loadRows.mockImplementation(() => {
    if (++reads === 2) bothReading.resolve();
    return shared.promise;
  });
  const cancelled = new AbortController();
  const first = result.current("record-1", { signal: cancelled.signal });
  const second = result.current("record-1", {
    signal: new AbortController().signal,
  });
  await bothReading.promise;
  cancelled.abort();
  await expect(first).resolves.toBeNull();
  version = 11;
  shared.resolve();
  await expect(second).resolves.toEqual({ row: null, rowVersion: 11 });
});

it("committed record waiting rechecks blockers after refresh and cancels on authority replacement", async () => {
  const { runtime, signal } = fixture();
  const refresh = deferred<void>();
  const entered = deferred<void>();
  let version: number | null = null;
  const loadRows = vi.fn(() => {
    entered.resolve();
    return refresh.promise;
  });
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      mutationRuntime: runtime,
      latestCommittedRowVersion: () => version,
      latestCommittedTimelineRow: () => null,
      loadRows,
    }),
  );
  const waiting = result.current("record-1", { signal });
  await entered.promise;
  runtime.setAuthority({
    actorId: "actor",
    sessionIdentity: "new-session",
    incidentId: "incident-1",
    role: "editor",
    closed: false,
  });
  await expect(waiting).resolves.toBeNull();
  version = 8;
  refresh.resolve();
  await Promise.resolve();
  version = null;
  loadRows.mockImplementation(async () => {
    version = 9;
    runtime.pendingQueue().model.pauseForAuthRecovery();
  });
  await expect(result.current("record-1", { signal })).resolves.toBeNull();
});

it("pending record readiness blocks on queue halt overflow and queue conflicts", async () => {
  const halted = fixture();
  const unit = halted.enqueue();
  const waiting = halted.wait();
  halted.runtime
    .pendingQueue()
    .model.haltBeforeDispatch(unit, "Cannot dispatch");
  halted.flush();
  await expect(waiting).resolves.toBe("blocked");
  await expect(halted.wait("unrelated")).resolves.toBe("blocked");
  const overflow = fixture();
  for (let index = 0; index <= pendingReplayCapacity; index++)
    overflow.enqueue(`record-${index}`);
  expect(
    overflow.runtime.pendingQueue().model.snapshot().overflow,
  ).not.toBeNull();
  await expect(overflow.wait("unrelated")).resolves.toBe("blocked");
  const conflicting = fixture();
  conflicting.enqueue();
  const queue = conflicting.runtime.pendingQueue().model;
  queue.dispatchNext();
  queue.settleDispatched({
    ok: false,
    status: 409,
    error: {
      code: "same_field_conflict",
      message: "Conflict",
      conflict: {
        conflict_token: "test-conflict",
        record_id: "record-1",
        field_key: "summary",
        base_row_version: 1,
        current_row_version: 2,
        conflict_resolution_class: "text_compare_merge",
      },
    },
  });
  expect(queue.snapshot().sameFieldConflicts).toHaveLength(1);
  await expect(conflicting.wait("unrelated")).resolves.toBe("blocked");
});

it("committed record waiting removes observers on pre-abort idle blocked and refresh failure", async () => {
  const { runtime, signal } = fixture();
  const listeners = new Set<() => void>();
  const subscribe = runtime.subscribe;
  vi.spyOn(runtime, "subscribe").mockImplementation((listener) => {
    listeners.add(listener);
    const off = subscribe(listener);
    return () => {
      listeners.delete(listener);
      off();
    };
  });
  const failure = new Error("refresh failed");
  const loadRows = vi.fn(async () => {
    throw failure;
  });
  let version: number | null = 1;
  const { result } = renderHook(() =>
    useTimelineCommittedRecordIdle({
      mutationRuntime: runtime,
      latestCommittedRowVersion: () => version,
      latestCommittedTimelineRow: () => null,
      loadRows,
    }),
  );
  const abort = new AbortController();
  abort.abort();
  await expect(
    result.current("record-1", { signal: abort.signal }),
  ).resolves.toBeNull();
  expect(listeners.size).toBe(0);
  await expect(result.current("record-1", { signal })).resolves.toEqual({
    row: null,
    rowVersion: 1,
  });
  expect(listeners.size).toBe(0);
  version = null;
  await expect(result.current("record-1", { signal })).rejects.toBe(failure);
  expect(listeners.size).toBe(0);
  runtime.pendingQueue().model.pauseForAuthRecovery();
  await expect(result.current("record-1", { signal })).resolves.toBeNull();
  expect(listeners.size).toBe(0);
});
