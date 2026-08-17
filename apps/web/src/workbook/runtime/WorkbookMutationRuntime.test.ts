import { afterEach, describe, expect, it, vi } from "vitest";
import {
  errorEnvelope,
  successEnvelope,
  timelineRow,
} from "../../testing/timelineWorkbookTestSupport";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { WorkbookMutationRuntime } from "./WorkbookMutationRuntime";
import { WorkbookMutationRuntimeRegistry } from "./WorkbookMutationRuntimeRegistry";

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

  it("preserves the draft and refreshes a rejected legacy conflict token", async () => {
    const conflict = (token: string, serverValue: string) => ({
      conflict_token: token,
      record_id: recordId,
      field_key: "timeline.activity_synopsis_text",
      conflict_resolution_class: "text_compare_merge",
      base_row_version: 1,
      current_row_version: 2,
      base_value: "Base",
      client_value: "Local draft",
      server_value: serverValue,
    });
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce(
        errorEnvelope(
          "same_field_conflict",
          409,
          conflict("cft2.retired-token", "Remote saved"),
        ),
      )
      .mockResolvedValueOnce(errorEnvelope("invalid_mutation_payload", 400))
      .mockResolvedValueOnce(
        errorEnvelope(
          "same_field_conflict",
          409,
          conflict("cft3.active.fresh-token", "Remote saved"),
        ),
      );
    vi.stubGlobal("fetch", fetchMock);
    const runtime = new WorkbookMutationRuntime(
      { clientInstanceId: "client-1", incidentId },
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
          value: "Local draft",
        },
      ],
      fieldKey: "timeline.activity_synopsis_text",
      localValue: "Local draft",
      recordId,
      rowLabel: "Timeline row",
      surfaceLabel: "Timeline",
      viewSchemaId: timelineViewSchemaId,
    });
    await vi.waitFor(() =>
      expect(runtime.getSnapshot().conflicts).toHaveLength(1),
    );
    const original = runtime.getSnapshot().conflicts[0];
    if (original === undefined) throw new Error("missing original conflict");
    runtime.updateConflictDraft(original.key, "Reviewed merged draft");

    await expect(
      runtime.resolveConflict({
        key: original.key,
        resolutionKind: "merged_value",
      }),
    ).resolves.toContain("draft was preserved");

    const refreshed = runtime.getSnapshot().conflicts[0];
    expect(refreshed?.conflict.conflict_token).toBe("cft3.active.fresh-token");
    expect(refreshed?.mergedDraft).toBe("Reviewed merged draft");
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/conflicts/cft2.retired-token/resolve",
    );
    expect(
      JSON.parse(String(fetchMock.mock.calls[2]?.[1]?.body)),
    ).toMatchObject({
      base_row_version: 1,
      changes: [
        {
          field_key: "timeline.activity_synopsis_text",
          value: "Local draft",
        },
      ],
    });
  });
});

function registryRuntime(
  scopedIncidentId: string,
  clientInstanceId = "client-1",
) {
  return {
    invalidate: vi.fn(),
    scope: { clientInstanceId, incidentId: scopedIncidentId },
  } as unknown as WorkbookMutationRuntime;
}

describe("WorkbookMutationRuntimeRegistry", () => {
  it("preserves one scoped runtime across authenticated shell remounts", () => {
    const registry = new WorkbookMutationRuntimeRegistry();
    const retained = registryRuntime("incident-1");
    const create = vi.fn(() => retained);

    expect(
      registry.acquire(
        { clientInstanceId: "client-1", incidentId: "incident-1" },
        create,
      ),
    ).toBe(retained);
    expect(
      registry.acquire(
        { clientInstanceId: "client-1", incidentId: "incident-1" },
        create,
      ),
    ).toBe(retained);
    expect(create).toHaveBeenCalledTimes(1);
    expect(retained.invalidate).not.toHaveBeenCalled();
  });

  it("retires the previous scope on incident change and app disposal", () => {
    const registry = new WorkbookMutationRuntimeRegistry();
    const first = registryRuntime("incident-1");
    const second = registryRuntime("incident-2");
    registry.acquire(first.scope, () => first);

    expect(registry.acquire(second.scope, () => second)).toBe(second);
    expect(first.invalidate).toHaveBeenNthCalledWith(1, {
      kind: "incident_changed",
      nextIncidentId: "incident-2",
    });
    expect(first.invalidate).toHaveBeenNthCalledWith(2, {
      kind: "runtime_disposed",
    });

    registry.dispose();
    registry.dispose();
    expect(second.invalidate).toHaveBeenCalledTimes(1);
    expect(second.invalidate).toHaveBeenCalledWith({
      kind: "runtime_disposed",
    });
    expect(() => registry.acquire(second.scope, () => second)).toThrow(
      "workbook mutation runtime registry is disposed",
    );
  });

  it("rejects a factory result from another pending-queue scope", () => {
    const registry = new WorkbookMutationRuntimeRegistry();
    const wrong = registryRuntime("incident-2");
    expect(() =>
      registry.acquire(
        { clientInstanceId: "client-1", incidentId: "incident-1" },
        () => wrong,
      ),
    ).toThrow("workbook mutation runtime factory returned wrong scope");
    expect(wrong.invalidate).toHaveBeenCalledWith({
      kind: "runtime_disposed",
    });
  });
});
