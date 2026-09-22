import { act, cleanup, render, screen } from "@testing-library/react";
import { StrictMode, Suspense } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { WorkbookMutationRuntimeRegistry } from "../runtime/WorkbookMutationRuntimeRegistry";
import {
  attachWorkbookMutationPresentation,
  type WorkbookMutationPresentation,
} from "./attachWorkbookMutationPresentation";
import { WorkbookMutationRuntimeBoundary } from "./WorkbookMutationRuntimeBoundary";

const authority = {
  actorId: "actor",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor" as const,
  closed: false,
};
function runtimeFor(incidentId = "incident") {
  return createWorkbookMutationRuntime(
    { incidentId, clientInstanceId: "client" },
    { create: () => "transaction" },
    { execute: vi.fn() },
  );
}
function presentation(
  runtime: ReturnType<typeof runtimeFor>,
): WorkbookMutationPresentation {
  return {
    runtime,
    apiBase: undefined,
    actorId: "actor",
    sessionIdentity: "session",
    surface: "surface",
    extensionWorkspace: false,
    refresh: {
      entities: vi.fn(async () => {}),
      assessment: vi.fn(async () => {}),
      generic: vi.fn(async () => {}),
    },
    authorityUncertain: vi.fn(),
    authorizationRecovered: vi.fn(),
  };
}
afterEach(cleanup);
describe("Workbook committed presentation", () => {
  it("does not acquire or replace the active lifetime in an abandoned render", () => {
    const registry = new WorkbookMutationRuntimeRegistry();
    const original = registry.acquire(
      { incidentId: "original", clientInstanceId: "client" },
      () => runtimeFor("original"),
    );
    const create = vi.fn(() => runtimeFor());
    const never = new Promise<void>(() => {});
    function Uncommitted(): never {
      throw never;
    }
    render(
      <Suspense fallback="Waiting">
        <WorkbookMutationRuntimeBoundary
          registry={registry}
          incidentId="incident"
          clientInstanceId="client"
          create={create}
        >
          {() => "Ready"}
        </WorkbookMutationRuntimeBoundary>
        <Uncommitted />
      </Suspense>,
    );
    expect(screen.getByText("Waiting")).toBeTruthy();
    expect(create).not.toHaveBeenCalled();
    expect(original.retired).toBe(false);
    registry.dispose();
  });
  it("constructs once through repeated committed mounts and retires only on replacement", () => {
    const registry = new WorkbookMutationRuntimeRegistry();
    const create = vi.fn(() => runtimeFor());
    const element = (
      <StrictMode>
        <WorkbookMutationRuntimeBoundary
          registry={registry}
          incidentId="incident"
          clientInstanceId="client"
          create={create}
        >
          {() => "Ready"}
        </WorkbookMutationRuntimeBoundary>
      </StrictMode>
    );
    const first = render(element);
    expect(screen.getByText("Ready")).toBeTruthy();
    expect(create).toHaveBeenCalledOnce();
    const runtime = create.mock.results[0]?.value;
    first.unmount();
    expect(runtime.retired).toBe(false);
    const second = render(element);
    expect(create).toHaveBeenCalledOnce();
    second.rerender(
      <WorkbookMutationRuntimeBoundary
        registry={registry}
        incidentId="next"
        clientInstanceId="client"
        create={() => runtimeFor("next")}
      >
        {() => "Next"}
      </WorkbookMutationRuntimeBoundary>,
    );
    expect(runtime.retired).toBe(true);
    expect(screen.getByText("Next")).toBeTruthy();
    second.unmount();
    registry.dispose();
  });
  it("keeps replacement callbacks when obsolete attachment cleanup runs", () => {
    const runtime = runtimeFor();
    runtime.setAuthority(authority);
    const first = presentation(runtime),
      next = presentation(runtime);
    const releaseFirst = attachWorkbookMutationPresentation(first);
    const releaseNext = attachWorkbookMutationPresentation(next);
    releaseFirst();
    releaseFirst();
    runtime.notifyPresentationAuthorizationRecovered({
      kind: "authorized",
      userId: "actor",
      role: "editor" as const,
    });
    expect(first.authorizationRecovered).not.toHaveBeenCalled();
    expect(next.authorizationRecovered).toHaveBeenCalledOnce();
    releaseNext();
    expect(runtime.recordReadScope?.actorId).toBe("actor");
    runtime.notifyPresentationAuthorizationRecovered({
      kind: "authorized",
      userId: "actor",
      role: "editor" as const,
    });
    expect(next.authorizationRecovered).toHaveBeenCalledOnce();
    runtime.invalidate({ kind: "runtime_disposed" });
  });
  it("rejects a mounted reconciliation result after its attachment is replaced", async () => {
    const runtime = runtimeFor();
    runtime.setAuthority(authority);
    const pending = deferred<void>();
    const initial = presentation(runtime);
    const first = {
      ...initial,
      refresh: { ...initial.refresh, entities: vi.fn(() => pending.promise) },
    };
    const registration = vi.spyOn(
      runtime.entityMerge,
      "registerProjectionRefresh",
    );
    const release = attachWorkbookMutationPresentation(first);
    const refresh = registration.mock.calls[0]?.[0];
    if (!refresh) throw new Error("Missing refresh registration");
    const result = refresh({} as Parameters<typeof refresh>[0]);
    const rejected = expect(result).rejects.toThrow(
      "Workbook presentation detached",
    );
    const next = attachWorkbookMutationPresentation(presentation(runtime));
    release();
    pending.resolve();
    await rejected;
    expect(first.refresh.generic).not.toHaveBeenCalled();
    next();
    runtime.invalidate({ kind: "runtime_disposed" });
  });
  it("retains an admitted dispatch and its acknowledgement after presentation detaches", async () => {
    const accepted =
      deferred<
        Awaited<
          ReturnType<
            Parameters<typeof createWorkbookMutationRuntime>[2]["execute"]
          >
        >
      >();
    const execute = vi.fn(() => accepted.promise);
    const runtime = createWorkbookMutationRuntime(
      { incidentId: "incident", clientInstanceId: "client" },
      { create: () => "transaction" },
      { execute },
    );
    runtime.setAuthority(authority);
    const release = attachWorkbookMutationPresentation(presentation(runtime));
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [{ field_key: "summary", value: "local" }],
      fieldKey: "summary",
      localValue: "local",
      recordId: "record",
      rowLabel: "Row",
      surfaceLabel: "Notes",
      viewSchemaId: "surface",
      sheetRef: { kind: "view_schema", id: "surface" },
    });
    await vi.waitFor(() => expect(execute).toHaveBeenCalledOnce());
    release();
    await act(async () =>
      accepted.resolve({
        kind: "accepted",
        value: {
          changeSetId: "change",
          viewSchemaId: "surface",
          row: {
            record_id: "record",
            row_version: 2,
            view_schema_id: "surface",
            cells: {},
          },
        },
      }),
    );
    await vi.waitFor(() =>
      expect(runtime.pendingQueue().model.snapshot().units).toHaveLength(0),
    );
    expect(execute).toHaveBeenCalledOnce();
    expect(runtime.surfaceRefreshDebts()).toContain("surface");
    expect(runtime.retired).toBe(false);
    runtime.invalidate({ kind: "runtime_disposed" });
  });
});
