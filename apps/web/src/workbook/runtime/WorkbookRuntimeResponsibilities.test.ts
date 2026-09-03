import { describe, expect, it, vi } from "vitest";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import { WorkbookClientTransactionLedger } from "./WorkbookClientTransactionLedger";
import { createWorkbookConflictStore } from "./WorkbookConflictStore";
import { createWorkbookMutationDriverRegistry } from "./WorkbookMutationDriverRegistry";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
import { WorkbookRetryScheduler } from "./WorkbookRetryScheduler";
import { WorkbookRuntimeLifecycle } from "./WorkbookRuntimeLifecycle";
import { WorkbookSurfaceRegistry } from "./WorkbookSurfaceRegistry";
import { projectWorkbookMutationStatus } from "./workbookMutationStatusProjector";
import { createWorkbookPendingQueueRuntime } from "./workbookPendingReplayRuntime";
import type { WorkbookSchedulerPort } from "./workbookRuntimePorts";

function controlledScheduler() {
  const microtasks: Array<() => void> = [];
  const delays: Array<{
    readonly cancel: ReturnType<typeof vi.fn>;
    readonly delayMilliseconds: number;
    readonly run: () => void;
  }> = [];
  const scheduler: WorkbookSchedulerPort = {
    enqueueMicrotask: (task) => microtasks.push(task),
    scheduleDelay: (delayMilliseconds, task) => {
      let active = true;
      const cancel = vi.fn(() => {
        active = false;
      });
      delays.push({
        cancel,
        delayMilliseconds,
        run: () => {
          if (active) task();
        },
      });
      return cancel;
    },
  };
  return { delays, microtasks, scheduler };
}

const scope = { clientInstanceId: "client-1", incidentId: "incident-1" };

describe("Workbook runtime responsibilities", () => {
  it("projects save state from explicit queue, conflict, and surface facts", () => {
    const pending = createWorkbookPendingQueueRuntime(scope);
    const saved = projectWorkbookMutationStatus({
      conflictPanelOpen: false,
      conflicts: [],
      explicitInFlightCount: 0,
      queue: pending.model.snapshot(),
      surfaceSaveStates: new Map(),
    });
    expect(saved.primaryLabel).toBe("Saved");

    const syncing = projectWorkbookMutationStatus({
      conflictPanelOpen: false,
      conflicts: [],
      explicitInFlightCount: 0,
      queue: pending.model.snapshot(),
      surfaceSaveStates: new Map([
        [
          "surface-1",
          { primaryLabel: "Syncing", secondaryMessage: "Refresh pending" },
        ],
      ]),
    });
    expect(syncing.primaryLabel).toBe("Syncing");
    expect(syncing.secondaryMessage).toBe("Refresh pending");
    expect(syncing.secondaryCandidates).toContainEqual({
      kind: "refresh_paused",
      message: "Refresh pending",
      surfaceId: "surface-1",
    });
  });

  it("coalesces lifecycle scheduling and permanently closes on disposal", async () => {
    const controlled = controlledScheduler();
    const lifecycle = new WorkbookRuntimeLifecycle(controlled.scheduler);
    const managedDrain = vi.fn(async () => undefined);
    const listener = vi.fn();
    lifecycle.subscribe(listener);

    lifecycle.requestDrain(managedDrain);
    lifecycle.requestDrain(managedDrain);
    expect(controlled.microtasks).toHaveLength(1);
    controlled.microtasks.shift()?.();
    await Promise.resolve();
    expect(managedDrain).toHaveBeenCalledOnce();
    lifecycle.emit();
    expect(listener).toHaveBeenCalledOnce();

    expect(lifecycle.dispose()).toBe(true);
    expect(lifecycle.dispose()).toBe(false);
    lifecycle.emit();
    lifecycle.requestDrain(managedDrain);
    expect(listener).toHaveBeenCalledOnce();
    expect(controlled.microtasks).toHaveLength(0);
  });

  it("runs at most one retry and supports deterministic cancellation", () => {
    const controlled = controlledScheduler();
    const retries = new WorkbookRetryScheduler(controlled.scheduler);
    const task = vi.fn();
    expect(retries.schedule(750, task)).toBe(true);
    expect(retries.schedule(750, task)).toBe(false);
    expect(controlled.delays).toHaveLength(1);
    expect(controlled.delays[0]?.delayMilliseconds).toBe(750);
    controlled.delays[0]?.run();
    expect(task).toHaveBeenCalledOnce();
    expect(retries.pending).toBe(false);

    expect(retries.schedule(750, task)).toBe(true);
    retries.cancel();
    controlled.delays[1]?.run();
    expect(controlled.delays[1]?.cancel).toHaveBeenCalledOnce();
    expect(task).toHaveBeenCalledOnce();
  });

  it("pauses owner work without a driver and rejects duplicate registration", async () => {
    const pending = createWorkbookPendingQueueRuntime(scope);
    const admission = pending.model.admit({
      id: "timeline-unit-1",
      kind: "patch",
      source: "autosave",
      incidentId: scope.incidentId,
      clientInstanceId: scope.clientInstanceId,
      viewSchemaId: "cartulary.view.timeline.v2",
      rowKey: "row-1",
      recordId: "record-1",
      payloadIntent: {
        base_row_version: 1,
        changes: [
          { field_key: "timeline.activity_synopsis_text", value: "Local" },
        ],
      },
      clientTxnId: "txn-1",
      coalesceKey: "record-1",
      enqueueOrder: 1,
    });
    expect(admission.accepted).toBe(true);
    if (!admission.accepted) return;

    const drivers = createWorkbookMutationDriverRegistry();
    drivers.claim(admission.unit.id, {
      kind: "timeline_row",
      viewSchemaId: admission.unit.viewSchemaId,
    });
    await expect(drivers.drain(admission.unit)).resolves.toEqual({
      status: "driver_absent",
      kind: "timeline_row",
    });
    expect(pending.model.snapshot().queuedCount).toBe(1);

    const firstDrain = vi.fn(async () => undefined);
    const first = drivers.register({ kind: "timeline_row", drain: firstDrain });
    expect(first.accepted).toBe(true);
    const duplicateDrain = vi.fn(async () => undefined);
    expect(
      drivers.register({ kind: "timeline_row", drain: duplicateDrain }),
    ).toEqual({
      accepted: false,
      status: "duplicate",
      kind: "timeline_row",
    });
    await expect(drivers.drain(admission.unit)).resolves.toEqual({
      status: "dispatched",
    });
    expect(firstDrain).toHaveBeenCalledOnce();
    expect(duplicateDrain).not.toHaveBeenCalled();

    if (first.accepted) first.unregister();
    await expect(drivers.drain(admission.unit)).resolves.toEqual({
      status: "driver_absent",
      kind: "timeline_row",
    });
  });

  it("retains refresh debt and ignores cleanup from a replaced registration", async () => {
    const onDebtChanged = vi.fn();
    const surfaces = new WorkbookSurfaceRegistry(onDebtChanged);
    await surfaces.refresh("surface-1");
    const firstRefresh = vi.fn(async () => undefined);
    const unregisterFirst = surfaces.register("surface-1", firstRefresh);
    await vi.waitFor(() => expect(firstRefresh).toHaveBeenCalledOnce());

    const replacementRefresh = vi.fn(async () => undefined);
    surfaces.register("surface-1", replacementRefresh);
    unregisterFirst();
    await surfaces.refresh("surface-1");
    expect(replacementRefresh).toHaveBeenCalledOnce();
    expect(onDebtChanged).not.toHaveBeenCalled();
  });

  it("preserves compatible conflict drafts and tracks transaction settlement", () => {
    const conflicts = createWorkbookConflictStore();
    const registration = {
      conflict: {
        conflict_token: "cft3.active.token",
        record_id: "record-1",
        field_key: "timeline.activity_synopsis_text",
        conflict_resolution_class: "text_compare_merge" as const,
        base_row_version: 1,
        current_row_version: 2,
        base_value: "Base",
        client_value: "Local",
        server_value: "Remote",
      },
      rowLabel: "Row 1",
      surfaceLabel: "Timeline",
      viewSchemaId: "cartulary.view.timeline.v2",
    };
    const first = conflicts.register(registration);
    expect(conflicts.updateDraft(first.key, "Merged")).toBe(true);
    conflicts.register({
      ...registration,
      conflict: { ...registration.conflict, current_row_version: 3 },
    });
    expect(conflicts.get(first.key)?.mergedDraft).toBe("Merged");
    expect(conflicts.panelOpen).toBe(true);

    const pending = createWorkbookPendingQueueRuntime(scope);
    const ledger = new WorkbookClientTransactionLedger();
    ledger.remember("txn-1");
    expect(ledger.settle("txn-1", pending)).toBe(true);
    expect(ledger.settle("txn-1", pending)).toBe(false);
    expect(ledger.settle(null, pending)).toBe(false);
  });

  it("uses injected time and cancels a pending transport retry on disposal", async () => {
    const controlled = controlledScheduler();
    const execute = vi.fn(async () => {
      throw new Error("offline");
    });
    const runtime = new WorkbookMutationRuntime(
      scope,
      { create: () => "txn-1" },
      { execute } satisfies WorkbookPendingMutationPort,
      {
        clock: { now: () => 4_242 },
        scheduler: controlled.scheduler,
      },
    );
    expect(
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [{ field_key: "field-1", value: "Local" }],
        fieldKey: "field-1",
        localValue: "Local",
        recordId: "record-1",
        rowLabel: "Row 1",
        surfaceLabel: "Surface 1",
        viewSchemaId: "surface-1",
      }),
    ).toEqual({ kind: "accepted" });
    expect(runtime.pendingQueue().model.snapshot().units[0]?.enqueueOrder).toBe(
      4_242,
    );
    expect(controlled.microtasks).toHaveLength(1);
    controlled.microtasks.shift()?.();
    await vi.waitFor(() => expect(controlled.delays).toHaveLength(1));
    expect(execute).toHaveBeenCalledOnce();

    runtime.invalidate({ kind: "runtime_disposed" });
    expect(controlled.delays[0]?.cancel).toHaveBeenCalledOnce();
    controlled.delays[0]?.run();
    expect(controlled.microtasks).toHaveLength(0);
  });
});
