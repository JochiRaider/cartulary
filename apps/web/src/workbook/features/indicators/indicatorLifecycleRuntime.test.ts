import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../../testing/fetchMockTestSupport";
import {
  lifecycleAuthority,
  lifecycleDraft,
  lifecycleIndicator,
  lifecycleReceipt,
  lifecycleRow,
} from "../../../testing/indicatorLifecycleTestSupport";
import { createIndicatorLifecycleAdapter } from "../../adapters/createIndicatorLifecycleAdapter";
import type { WorkbookPendingMutationPort } from "../../ports/WorkbookPendingMutationPort";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { WorkbookMutationRuntimeRegistry } from "../../runtime/WorkbookMutationRuntimeRegistry";
import { indicatorLifecycleViewId } from "./indicatorLifecycleModel";
import type { IndicatorLifecycleTransportPort } from "./indicatorLifecycleOperation";

afterEach(() => vi.useRealTimers());
it("Lifecycle runtime waits for same-record edits and requires review of their accepted version", async () => {
  vi.useFakeTimers();
  const earlier =
    deferred<Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>>();
  const runtime = new WorkbookMutationRuntime(
    { incidentId: lifecycleAuthority.incidentId, clientInstanceId: "tab" },
    { create: () => "original" },
    { execute: () => earlier.promise },
  );
  const owner = runtime.indicatorLifecycle,
    send = vi.fn<IndicatorLifecycleTransportPort["send"]>(async () => ({
      kind: "uncertain",
    }));
  owner.configure({
    ...createIndicatorLifecycleAdapter({
      apiBase: undefined,
      incidentId: lifecycleAuthority.incidentId,
    }),
    send,
  });
  owner.setAuthority(lifecycleAuthority);
  owner.acceptRow(lifecycleRow());
  owner.drafts.open(lifecycleIndicator, "example.test", 1);
  owner.drafts.update(lifecycleIndicator, lifecycleDraft().values);
  const draft = owner.drafts.get(lifecycleIndicator);
  if (!draft) throw new Error("draft");
  const patch = {
    baseRowVersion: 1,
    changes: [{ field_key: "indicator.notes", value: "Earlier" }],
    fieldKey: "indicator.notes",
    localValue: "Earlier",
    recordId: lifecycleIndicator,
    rowLabel: "Indicator",
    surfaceLabel: "Indicators",
    viewSchemaId: indicatorLifecycleViewId,
  };
  runtime.enqueuePatch(patch);
  await vi.advanceTimersByTimeAsync(1);
  const attempt = owner.admit(draft, {
    isCurrent: () => true,
    reconcile: async () => {},
  });
  if (!attempt) throw new Error("admission");
  const running = owner.execute(attempt);
  await vi.advanceTimersByTimeAsync(1);
  expect(send).not.toHaveBeenCalled();
  expect(runtime.enqueuePatch(patch).kind).toBe("rejected_mutation");
  earlier.resolve({
    kind: "accepted",
    value: {
      changeSetId: "prior",
      viewSchemaId: indicatorLifecycleViewId,
      row: { ...lifecycleRow(2), view_schema_id: indicatorLifecycleViewId },
    },
  });
  await vi.advanceTimersByTimeAsync(32);
  await running;
  expect(send).not.toHaveBeenCalled();
  expect(owner.getSnapshot().entries[0]?.phase).toBe("rejected");
  expect(owner.drafts.get(lifecycleIndicator)?.baseRowVersion).toBe(1);
  runtime.invalidate({ kind: "runtime_disposed" });
});
it("Lifecycle runtime retains recovery across shell detachment and retires account data", async () => {
  const registry = new WorkbookMutationRuntimeRegistry(),
    scope = {
      incidentId: lifecycleAuthority.incidentId,
      clientInstanceId: "tab",
    };
  const create = () =>
    new WorkbookMutationRuntime(
      scope,
      { create: () => "original" },
      { execute: vi.fn() },
    );
  const runtime = registry.acquire(scope, create),
    owner = runtime.indicatorLifecycle;
  const send = vi.fn<IndicatorLifecycleTransportPort["send"]>(async () => ({
    kind: "uncertain",
  }));
  owner.configure({
    ...createIndicatorLifecycleAdapter({
      apiBase: undefined,
      incidentId: scope.incidentId,
    }),
    send,
  });
  owner.setAuthority(lifecycleAuthority);
  owner.acceptRow(lifecycleRow());
  owner.drafts.open(lifecycleIndicator, "example.test", 1);
  owner.drafts.update(lifecycleIndicator, lifecycleDraft().values);
  const draft = owner.drafts.get(lifecycleIndicator);
  if (!draft) throw new Error("draft");
  runtime.history.acceptVersion(lifecycleIndicator, 2);
  expect(
    owner.admit(draft, { isCurrent: () => true, reconcile: async () => {} }),
  ).toBeNull();
  owner.acceptRow(lifecycleRow(2));
  owner.drafts.review(lifecycleIndicator, 2);
  const reviewed = owner.drafts.get(lifecycleIndicator);
  if (!reviewed) throw new Error("review");
  const attempt = owner.admit(reviewed, {
    isCurrent: () => true,
    reconcile: async () => {},
  });
  if (!attempt) throw new Error("admit");
  await owner.execute(attempt);
  expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
  expect(runtime.getSnapshot().unresolvedConflictCount).toBe(0);
  registry.sessionUnavailable();
  expect(owner.getSnapshot().entries).toEqual([]);
  expect(registry.acquire(scope, create)).toBe(runtime);
  owner.setAuthority(lifecycleAuthority);
  expect(owner.getSnapshot().entries[0]?.attempt).toBe(attempt);
  send.mockResolvedValue({ kind: "acknowledged", receipt: lifecycleReceipt() });
  await owner.replay(attempt.id);
  expect(send).toHaveBeenCalledTimes(2);
  registry.replaceAccount();
  expect(owner.drafts.get(lifecycleIndicator)).toBeNull();
  expect(owner.getSnapshot().entries).toEqual([]);
  registry.dispose();
});
