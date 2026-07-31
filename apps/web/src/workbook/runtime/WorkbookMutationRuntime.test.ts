import { afterEach, describe, expect, it, vi } from "vitest";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

const transactionIds = {
  create: (prefix: string) => `${prefix}-txn`,
};

function successResponse(recordId: string, rowVersion: number): Response {
  return new Response(
    JSON.stringify({
      data: {
        view_schema_id: "cartulary.view.tasks.v1",
        row: { record_id: recordId, row_version: rowVersion, cells: {} },
      },
    }),
    {
      status: 200,
      headers: { "content-type": "application/json" },
    },
  );
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("WorkbookMutationRuntime", () => {
  it("keeps secure transaction identity failure local without queue admission", () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-1",
        incidentId: "incident-1",
      },
      {
        create: () => {
          throw new Error("randomness unavailable");
        },
      },
    );

    expect(
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [{ field_key: "task.title", value: "Local title" }],
        fieldKey: "task.title",
        localValue: "Local title",
        recordId: "task-1",
        rowLabel: "Task 1",
        surfaceLabel: "Tasks",
        viewSchemaId: "cartulary.view.tasks.v1",
      }),
    ).toEqual({
      kind: "rejected_mutation",
      message:
        "This edit remains local because a secure transaction ID could not be created.",
    });
    expect(runtime.pending().model.snapshot().units).toEqual([]);
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("retains queued drafts and refresh debt across surface unmounts", async () => {
    const releaseRequest: { current: (() => void) | null } = {
      current: null,
    };
    vi.stubGlobal(
      "fetch",
      vi.fn(
        () =>
          new Promise<Response>((resolve) => {
            releaseRequest.current = () =>
              resolve(successResponse("task-1", 2));
          }),
      ),
    );
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-1",
        incidentId: "incident-1",
      },
      transactionIds,
    );
    const firstRefresh = vi.fn();
    const unregister = runtime.registerSurface(
      "cartulary.view.tasks.v1",
      firstRefresh,
    );

    expect(
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [{ field_key: "task.title", value: "Local title" }],
        fieldKey: "task.title",
        localValue: "Local title",
        recordId: "task-1",
        rowLabel: "Task 1",
        surfaceLabel: "Tasks",
        viewSchemaId: "cartulary.view.tasks.v1",
      }),
    ).toEqual({ kind: "accepted" });
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    expect(
      runtime.visibleEdit("cartulary.view.tasks.v1", "task-1", "task.title"),
    ).toBe("Local title");

    unregister();
    await vi.waitFor(() => expect(releaseRequest.current).not.toBeNull());
    releaseRequest.current?.();
    await vi.waitFor(() =>
      expect(runtime.getSnapshot().primaryLabel).toBe("Saved"),
    );
    expect(firstRefresh).not.toHaveBeenCalled();

    const returnRefresh = vi.fn();
    runtime.registerSurface("cartulary.view.tasks.v1", returnRefresh);
    await vi.waitFor(() => expect(returnRefresh).toHaveBeenCalledOnce());
    expect(
      runtime.visibleEdit("cartulary.view.tasks.v1", "task-1", "task.title"),
    ).toBeUndefined();
  });

  it("coalesces current-surface autosaves through one shell queue", async () => {
    const releaseRequest: { current: (() => void) | null } = {
      current: null,
    };
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Promise<Response>((resolve) => {
          releaseRequest.current = () => resolve(successResponse("task-1", 2));
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-1",
        incidentId: "incident-1",
      },
      transactionIds,
    );
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [{ field_key: "task.title", value: "First" }],
      fieldKey: "task.title",
      localValue: "First",
      recordId: "task-1",
      rowLabel: "Task 1",
      surfaceLabel: "Tasks",
      viewSchemaId: "cartulary.view.tasks.v1",
    });
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [{ field_key: "task.title", value: "Second" }],
      fieldKey: "task.title",
      localValue: "Second",
      recordId: "task-1",
      rowLabel: "Task 1",
      surfaceLabel: "Tasks",
      viewSchemaId: "cartulary.view.tasks.v1",
    });

    expect(runtime.pending().model.snapshot().units).toHaveLength(1);
    expect(
      runtime.visibleEdit("cartulary.view.tasks.v1", "task-1", "task.title"),
    ).toBe("Second");
    await vi.waitFor(() => expect(releaseRequest.current).not.toBeNull());
    releaseRequest.current?.();
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    const request = fetchMock.mock.calls[0]?.[1];
    if (request === undefined) throw new Error("missing request init");
    expect(JSON.parse(String(request.body))).toMatchObject({
      changes: [{ field_key: "task.title", value: "Second" }],
    });
  });
});
