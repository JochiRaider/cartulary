import { afterEach, expect, it, vi } from "vitest";
import {
  taskAuthority,
  taskIntent,
  taskReceipt,
  taskRecordId,
  taskRow,
} from "../../testing/taskWorkbookTestSupport";
import { createWorkbookOperationExecutor } from "../adapters/workbookOperationExecutor";
import {
  captureRecordPatch,
  createRecordPatchTransport,
  type RecordPatchOutcome,
  type RecordPatchTransport,
} from "../adapters/workbookRecordPatchTransport";
import { taskExplicitPatchContribution } from "../features/coordination/taskExplicitPatchContribution";
import {
  TaskLifecycleDraftStore,
  taskViewId,
} from "../features/coordination/taskLifecycleModel";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookSourceWriteSettlement } from "../ports/WorkbookSourceWriteCoordination";
import { WorkbookExplicitPatchOwner } from "./WorkbookExplicitPatchOwner";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

function deferred<T>() {
  let resolve: (value: T) => void = () => {
    throw new Error("uninitialized deferred");
  };
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
function fixture() {
  const coordinate = vi.fn<() => Promise<WorkbookSourceWriteSettlement>>(
      async () => ({ kind: "settled", minimumRowVersion: 0 }),
    ),
    registerConflict = vi.fn(),
    accepted = vi.fn(),
    create = vi.fn((prefix: string) => `${prefix}-${create.mock.calls.length}`);
  const drafts = new TaskLifecycleDraftStore();
  const refresh = vi.fn(async () => {});
  const owner = new WorkbookExplicitPatchOwner(
    taskAuthority.incidentId,
    { create },
    {
      coordinate,
      registerConflict,
      accepted,
      refresh,
      contribute: (intent) => taskExplicitPatchContribution(intent, drafts),
    },
    100,
  );
  const send = vi.fn<RecordPatchTransport["send"]>(
    async (): Promise<RecordPatchOutcome> => ({
      kind: "acknowledged",
      receipt: taskReceipt(),
    }),
  );
  owner.configure(
    { send },
    undefined,
    async (_view, id) => owner.latestRow(id) ?? taskRow(),
  );
  owner.setAuthority(taskAuthority);
  owner.observeQuery(taskRow());
  return {
    owner,
    drafts,
    send,
    coordinate,
    registerConflict,
    accepted,
    create,
    refresh,
  };
}
afterEach(() => {
  vi.unstubAllGlobals();
  vi.useRealTimers();
});
it("reserves explicit Task patches synchronously outside the autosave queue", async () => {
  const f = fixture(),
    gate = deferred<WorkbookSourceWriteSettlement>();
  f.coordinate.mockReturnValue(gate.promise);
  const first = f.owner.submit(taskIntent());
  expect(f.owner.blocksRecord(taskRecordId)).toBe(true);
  expect(await f.owner.submit(taskIntent())).toBeNull();
  expect(f.send).not.toHaveBeenCalled();
  gate.resolve({ kind: "settled", minimumRowVersion: 0 });
  expect((await first)?.phase).toBe("acknowledged");
  expect(f.create).toHaveBeenCalledTimes(1);
  expect(f.send).toHaveBeenCalledTimes(1);
});
it("coordinates disjoint writes but preserves drafts when guard siblings moved", async () => {
  const f = fixture();
  f.coordinate.mockImplementation(async () => {
    f.owner.observeQuery({
      ...taskRow(8),
      cells: { ...taskRow().cells, "task.title": { value: "New title" } },
    });
    return { kind: "settled", minimumRowVersion: 0 };
  });
  await f.owner.submit(taskIntent());
  expect(
    JSON.parse(f.send.mock.calls[0]?.[0]?.body ?? "{}").base_row_version,
  ).toBe(8);
  const next = fixture();
  next.coordinate.mockImplementation(async () => {
    next.owner.observeQuery({
      ...taskRow(8),
      cells: {
        ...taskRow().cells,
        "task.owner_user_id": { value: "other-member" },
      },
    });
    return { kind: "settled", minimumRowVersion: 0 };
  });
  const rejected = await next.owner.submit(taskIntent());
  expect(rejected?.phase).toBe("preparation_failed");
  expect(next.send).not.toHaveBeenCalled();
  expect(rejected?.intent.changes).toEqual(taskIntent().changes);
});
it("retains exact identity for precommit loss and committed lost responses", async () => {
  for (const loss of ["precommit", "committed"]) {
    const f = fixture();
    f.send.mockResolvedValueOnce({ kind: "uncertain" });
    const first = await f.owner.submit(taskIntent());
    expect(first?.phase, loss).toBe("uncertain");
    if (!first) throw new Error("missing operation");
    f.owner.acceptRow(taskRow(9, "done"));
    await Promise.all([f.owner.replay(first.id), f.owner.replay(first.id)]);
    expect(f.send).toHaveBeenCalledTimes(2);
    expect(f.send.mock.calls[0]?.[0]).toBe(f.send.mock.calls[1]?.[0]);
    expect(f.create).toHaveBeenCalledTimes(1);
    expect(f.owner.latestVersion(taskRecordId)).toBe(9);
    expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(taskReceipt());
  }
});
it("retains receipt before refresh and retries refresh without a mutation", async () => {
  const f = fixture(),
    gate = deferred<void>();
  f.refresh.mockReturnValueOnce(gate.promise);
  const submitted = f.owner.submit(taskIntent());
  await vi.waitFor(() =>
    expect(f.owner.getSnapshot().entries[0]?.phase).toBe("acknowledged"),
  );
  expect(f.owner.latestRow(taskRecordId)?.cells["task.status"]?.value).toBe(
    "blocked",
  );
  expect(f.owner.blocksRecord(taskRecordId)).toBe(true);
  gate.resolve();
  await submitted;
  f.refresh.mockRejectedValueOnce(new Error("read failure"));
  const id = f.owner.getSnapshot().entries[0]?.id;
  if (!id) throw new Error("missing operation");
  await f.owner.refresh(id);
  expect(f.owner.getSnapshot().entries[0]?.reconciliation).toBe("required");
  await f.owner.refresh(id);
  expect(f.owner.getSnapshot().entries[0]?.reconciliation).toBe("complete");
  expect(f.send).toHaveBeenCalledTimes(1);
});
it("keeps admitted Task work while detached and conceals it during session loss", async () => {
  const f = fixture(),
    pending = deferred<RecordPatchOutcome>();
  f.send.mockReturnValueOnce(pending.promise);
  const result = f.owner.submit(taskIntent());
  await vi.waitFor(() => expect(f.send).toHaveBeenCalledTimes(1));
  f.owner.suspend();
  expect(f.owner.getSnapshot().entries).toEqual([]);
  expect(f.owner.latestRow(taskRecordId)).toBeNull();
  pending.resolve({ kind: "acknowledged", receipt: taskReceipt() });
  await result;
  f.owner.setAuthority({ ...taskAuthority, sessionIdentity: "next-session" });
  expect(f.owner.getSnapshot().entries[0]?.receipt).toEqual(taskReceipt());
  f.owner.setAuthority({ ...taskAuthority, actorId: "replacement" });
  expect(f.owner.getSnapshot().entries).toEqual([]);
  expect(f.owner.canSubmit()).toBe(false);
});
it("blocks replay on failed current requery and session replacement", async () => {
  const f = fixture();
  f.send.mockResolvedValueOnce({ kind: "uncertain" });
  const result = await f.owner.submit(taskIntent());
  if (!result) throw new Error("missing operation");
  f.owner.suspend();
  await f.owner.replay(result.id);
  expect(f.send).toHaveBeenCalledTimes(1);
  f.owner.setAuthority({ ...taskAuthority, sessionIdentity: "recovered" });
  f.refresh.mockRejectedValueOnce(new Error("query failed"));
  await f.owner.replay(result.id);
  expect(f.send).toHaveBeenCalledTimes(1);
  expect(f.owner.getSnapshot().entries[0]?.request).toBe(result.request);
  await f.owner.replay(result.id);
  expect(f.send).toHaveBeenCalledTimes(2);
});
it("registers the actual compound conflict field and retains full intent after Keep saved", async () => {
  const f = fixture();
  f.drafts.update(taskRow(), "task.blocked_reason", "Waiting");
  f.send.mockResolvedValueOnce({
    kind: "rejected",
    failure: {
      kind: "same_field_conflict",
      message: "conflict",
      conflict: {
        record_id: taskRecordId,
        field_key: "task.blocked_reason",
        conflict_token: "token",
        conflict_resolution_class: "atomic_replace",
        base_row_version: 7,
        current_row_version: 8,
        client_value: "Waiting",
        server_value: "Saved",
      },
    },
  });
  const result = await f.owner.submit(taskIntent());
  expect(result?.phase).toBe("conflict");
  expect(f.registerConflict).toHaveBeenCalledWith(
    expect.objectContaining({
      compoundOperationId: result?.id,
      focusKey: `${taskRecordId}:task.blocked_reason`,
      focusOrigin: "inspector",
    }),
  );
  f.owner.conflictResolved(taskRecordId);
  expect(f.owner.blocksRecord(taskRecordId)).toBe(false);
  expect(f.drafts.read(taskRow()).values["task.blocked_reason"]).toBe(
    "Waiting",
  );
  expect(f.owner.getSnapshot().entries[0]?.intent.changes).toEqual(
    taskIntent().changes,
  );
});
it("classifies timeout and transport exceptions as uncertain without losing identity", async () => {
  vi.useFakeTimers();
  const f = fixture();
  f.send.mockImplementationOnce(() => new Promise(() => {}));
  const submitted = f.owner.submit(taskIntent());
  await vi.advanceTimersByTimeAsync(101);
  expect((await submitted)?.phase).toBe("uncertain");
  const h = fixture();
  h.coordinate.mockImplementationOnce(() => new Promise(() => {}));
  const coordinated = h.owner.submit(taskIntent());
  await vi.advanceTimersByTimeAsync(101);
  expect((await coordinated)?.phase).toBe("preparation_failed");
  expect(h.send).not.toHaveBeenCalled();
  const g = fixture();
  g.send.mockRejectedValueOnce(new Error("network"));
  expect((await g.owner.submit(taskIntent()))?.phase).toBe("uncertain");
});
it("validates complete ordinary patch receipts and treats malformed success as uncertain", async () => {
  const captured = captureRecordPatch({
    recordId: taskRecordId,
    viewSchemaId: taskViewId,
    baseRowVersion: 7,
    clientTxnId: "secure-task-id",
    changes: taskIntent().changes,
  });
  if (!captured) throw new Error("capture failed");
  const transport = createRecordPatchTransport(
    createWorkbookOperationExecutor({ apiBase: undefined }),
  );
  const receipt = taskReceipt();
  const data = {
    change_set_id: receipt.changeSetId,
    view_schema_id: taskViewId,
    row: receipt.row,
  };
  for (const invalid of [
    { ...data, change_set_id: "" },
    { ...data, row: { ...receipt.row, cells: {} } },
    { ...data, row: taskRow(7) },
    { ...data, view_schema_id: "wrong" },
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(
        async () =>
          new Response(
            JSON.stringify({
              data: invalid,
              meta: { request_id: "request-task" },
            }),
            { status: 200, headers: { "content-type": "application/json" } },
          ),
      ),
    );
    expect(
      (await transport.send(captured, new AbortController().signal)).kind,
    ).toBe("uncertain");
  }
  const fetch = vi.fn<typeof globalThis.fetch>(
    async () =>
      new Response(
        JSON.stringify({ data, meta: { request_id: "request-task" } }),
        { status: 200, headers: { "content-type": "application/json" } },
      ),
  );
  vi.stubGlobal("fetch", fetch);
  expect(await transport.send(captured, new AbortController().signal)).toEqual({
    kind: "acknowledged",
    receipt,
  });
  expect(fetch.mock.calls[0]?.[1]?.body).toBe(captured.body);
});

it("waits for Task autosave receipt and source verification without waiting for presentation", async () => {
  const pending =
    deferred<Awaited<ReturnType<WorkbookPendingMutationPort["execute"]>>>();
  let sequence = 0;
  const runtime = new WorkbookMutationRuntime(
    { incidentId: taskAuthority.incidentId, clientInstanceId: "task-test" },
    { create: (prefix) => `${prefix}-${++sequence}` },
    { execute: () => pending.promise },
  );
  runtime.explicitPatches.setAuthority(taskAuthority);
  runtime.explicitPatches.observeQuery(taskRow());
  const send = vi.fn<RecordPatchTransport["send"]>(async () => ({
    kind: "acknowledged",
    receipt: { ...taskReceipt(), row: taskRow(9, "blocked") },
  }));
  const verified = deferred<ReturnType<typeof taskRow>>();
  const readSource = vi.fn(async () => verified.promise);
  runtime.explicitPatches.configure({ send }, undefined, readSource);
  const refreshing = deferred<void>();
  let refreshStarted = false;
  runtime.registerSurface(taskViewId, async () => {
    refreshStarted = true;
    await refreshing.promise;
  });
  expect(
    runtime.enqueuePatch({
      recordId: taskRecordId,
      viewSchemaId: taskViewId,
      baseRowVersion: 7,
      fieldKey: "task.title",
      changes: [{ field_key: "task.title", value: "Queued title" }],
      localValue: "Queued title",
      rowLabel: "Task",
      surfaceLabel: "Tasks",
    }).kind,
  ).toBe("accepted");
  const explicit = runtime.explicitPatches.submit(taskIntent());
  expect(runtime.pendingQueue().model.snapshot().units).toHaveLength(1);
  await vi.waitFor(() =>
    expect(runtime.pendingQueue().model.snapshot().units[0]?.status).toBe(
      "in_flight",
    ),
  );
  pending.resolve({
    kind: "accepted",
    value: {
      ...taskReceipt(),
      row: {
        ...taskRow(8),
        cells: { ...taskRow().cells, "task.title": { value: "Queued title" } },
      },
    },
  });
  await vi.waitFor(() => expect(refreshStarted).toBe(true));
  expect(send).not.toHaveBeenCalled();
  await vi.waitFor(() => expect(readSource).toHaveBeenCalledOnce());
  verified.resolve(taskRow(8));
  await vi.waitFor(() => expect(send).toHaveBeenCalledOnce());
  refreshing.resolve();
  await explicit;
  expect(send).toHaveBeenCalledTimes(1);
  expect(JSON.parse(send.mock.calls[0]?.[0]?.body ?? "{}")).toMatchObject({
    base_row_version: 8,
    changes: taskIntent().changes,
  });
  expect(runtime.pendingQueue().model.snapshot().units).toHaveLength(0);
  runtime.invalidate({ kind: "runtime_disposed" });
});
