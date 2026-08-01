import { afterEach, describe, expect, it, vi } from "vitest";
import {
  successEnvelope,
  timelineRow,
} from "../../testing/timelineWorkbookTestSupport";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";

const incidentId = "10000000-0000-4000-8000-000000000001";
const recordId = "20000000-0000-4000-8000-000000000001";

const transactionIds = {
  create: (prefix: string) => `${prefix}-txn`,
};

function successResponse(rowVersion: number): Response {
  return successEnvelope({
    change_set_id: "30000000-0000-4000-8000-000000000001",
    row: timelineRow({ captureState: "rough", recordId, rowVersion }),
    view_schema_id: timelineViewSchemaId,
  });
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
        incidentId,
      },
      {
        create: () => {
          throw new Error("randomness unavailable");
        },
      },
      createWorkbookPendingMutationAdapter({
        apiBase: undefined,
        incidentId,
      }),
    );

    expect(
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [
          {
            field_key: "timeline.activity_synopsis_text",
            value: "Local title",
          },
        ],
        fieldKey: "timeline.activity_synopsis_text",
        localValue: "Local title",
        recordId,
        rowLabel: "Task 1",
        surfaceLabel: "Tasks",
        viewSchemaId: timelineViewSchemaId,
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
            releaseRequest.current = () => resolve(successResponse(2));
          }),
      ),
    );
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-1",
        incidentId,
      },
      transactionIds,
      createWorkbookPendingMutationAdapter({
        apiBase: undefined,
        incidentId,
      }),
    );
    const firstRefresh = vi.fn();
    const unregister = runtime.registerSurface(
      timelineViewSchemaId,
      firstRefresh,
    );

    expect(
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [
          {
            field_key: "timeline.activity_synopsis_text",
            value: "Local title",
          },
        ],
        fieldKey: "timeline.activity_synopsis_text",
        localValue: "Local title",
        recordId,
        rowLabel: "Task 1",
        surfaceLabel: "Tasks",
        viewSchemaId: timelineViewSchemaId,
      }),
    ).toEqual({ kind: "accepted" });
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    expect(
      runtime.visibleEdit(
        timelineViewSchemaId,
        recordId,
        "timeline.activity_synopsis_text",
      ),
    ).toBe("Local title");

    unregister();
    await vi.waitFor(() => expect(releaseRequest.current).not.toBeNull());
    releaseRequest.current?.();
    await vi.waitFor(() =>
      expect(runtime.getSnapshot().primaryLabel).toBe("Saved"),
    );
    expect(firstRefresh).not.toHaveBeenCalled();

    const returnRefresh = vi.fn();
    runtime.registerSurface(timelineViewSchemaId, returnRefresh);
    await vi.waitFor(() => expect(returnRefresh).toHaveBeenCalledOnce());
    expect(
      runtime.visibleEdit(
        timelineViewSchemaId,
        recordId,
        "timeline.activity_synopsis_text",
      ),
    ).toBeUndefined();
  });

  it("coalesces current-surface autosaves through one shell queue", async () => {
    const releaseRequest: { current: (() => void) | null } = {
      current: null,
    };
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) =>
        new Promise<Response>((resolve) => {
          releaseRequest.current = () => resolve(successResponse(2));
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const runtime = new WorkbookMutationRuntime(
      {
        clientInstanceId: "client-1",
        incidentId,
      },
      transactionIds,
      createWorkbookPendingMutationAdapter({
        apiBase: undefined,
        incidentId,
      }),
    );
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "First",
        },
      ],
      fieldKey: "timeline.activity_synopsis_text",
      localValue: "First",
      recordId,
      rowLabel: "Task 1",
      surfaceLabel: "Tasks",
      viewSchemaId: timelineViewSchemaId,
    });
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Second",
        },
      ],
      fieldKey: "timeline.activity_synopsis_text",
      localValue: "Second",
      recordId,
      rowLabel: "Task 1",
      surfaceLabel: "Tasks",
      viewSchemaId: timelineViewSchemaId,
    });

    expect(runtime.pending().model.snapshot().units).toHaveLength(1);
    expect(
      runtime.visibleEdit(
        timelineViewSchemaId,
        recordId,
        "timeline.activity_synopsis_text",
      ),
    ).toBe("Second");
    await vi.waitFor(() => expect(releaseRequest.current).not.toBeNull());
    releaseRequest.current?.();
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalledOnce());
    const request = fetchMock.mock.calls[0]?.[1];
    if (request === undefined) throw new Error("missing request init");
    expect(JSON.parse(String(request.body))).toMatchObject({
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Second",
        },
      ],
    });
  });
});
