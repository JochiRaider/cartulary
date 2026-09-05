import { act, cleanup, render, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { emptyPresenceScope } from "./collaboration/workbookPresencePresentation";
import { WorkbookSaveAnnouncements } from "./components/WorkbookSaveAnnouncements";
import { WorkbookStatusStrip } from "./components/WorkbookStatusStrip";
import { useGenericSurfaceMutationController } from "./hooks/useGenericSurfaceMutationController";
import type { GenericMutationCommandPort } from "./mutations/workbookMutationCommandPorts";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import {
  projectWorkbookMutationStatus,
  projectWorkbookStatusForSurface,
} from "./runtime/workbookMutationStatusProjector";
import { workbookPendingQueueSnapshot } from "./runtime/workbookPendingReplayRuntime";
import { useTimelineSaveStatePresentation } from "./timeline/hooks/useTimelineSaveStatePresentation";
import { timelinePendingSavesRefsFor } from "./timeline/models/timelinePendingSaves";
import { selectWorkbookStatusSecondary } from "./utils/workbookStatusSecondary";

afterEach(cleanup);

function runtimeFixture() {
  return new WorkbookMutationRuntime(
    { incidentId: "incident-1", clientInstanceId: "client-1" },
    { create: () => "transaction-1" },
    { execute: vi.fn() },
  );
}

describe("Workbook save status", () => {
  it("keeps a global FIFO blocker explanation on another surface", () => {
    const runtime = runtimeFixture();
    const queue = runtime.pendingQueue().model.snapshot();
    const snapshot = projectWorkbookMutationStatus({
      conflictPanelOpen: false,
      conflicts: [],
      explicitInFlightCount: 1,
      queue: {
        ...queue,
        halted: {
          unit_id: "blocked-unit",
          error_code: "client_txn_conflict",
          message: "unsafe /api/v1/raw-token",
          anchor: { kind: "surface" },
        },
      },
    });
    expect(snapshot.primaryLabel).toBe("Conflict");
    const secondary = selectWorkbookStatusSecondary(
      snapshot.secondaryCandidates,
      { kind: "view_schema", id: "cartulary.view.assessments.v1" },
    );
    expect(secondary?.kind).toBe("client_txn_conflict");
    expect(secondary?.message).not.toContain("/api/");
  });

  it("keeps both overlapping generic operations pending", () => {
    const runtime = runtimeFixture();
    const { result, unmount } = renderHook(() =>
      useGenericSurfaceMutationController({
        mutationCommands: {} as GenericMutationCommandPort,
        mutationRuntime: runtime,
        onRefresh: vi.fn(),
        refreshReferenceOptions: vi.fn(),
        surfaceLabel: "Notes",
        sheetRef: { kind: "saved_view", id: "notes-one" },
      }),
    );
    let first = () => {};
    let second = () => {};
    act(() => {
      first = result.current.beginMutation();
      second = result.current.beginMutation();
    });
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(2);
    act(() => {
      second();
      second();
    });
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
    unmount();
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    act(first);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
  });

  it("does not erase pending Timeline work on surface unmount", () => {
    const runtime = runtimeFixture();
    const pending = runtime.pendingQueue();
    const refs = timelinePendingSavesRefsFor(runtime, pending);
    const conflictQueueRef = { current: {} };
    const { result, unmount } = renderHook(() =>
      useTimelineSaveStatePresentation({
        conflictQueue: conflictQueueRef.current,
        sheetRef: { kind: "saved_view", id: "timeline-one" },
        mutationRuntime: runtime,
        pendingQueueSnapshot: workbookPendingQueueSnapshot(pending),
        pendingSavesRefs: refs,
        setPendingQueueSnapshot: vi.fn(),
      }),
    );
    let finish = () => {};
    act(() => {
      finish = result.current.commands.beginSave();
    });
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    unmount();
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    act(() => {
      finish();
      finish();
    });
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
  });

  it("leaves visible save labels outside independent live regions", () => {
    const { container } = render(
      <WorkbookStatusStrip
        presence={emptyPresenceScope}
        status={projectWorkbookStatusForSurface(runtimeFixture().getSnapshot())}
        chromeMode="base"
        workbookFocusAnchor={null}
      />,
    );
    expect(
      container.querySelector('[aria-live], [role="status"], [role="alert"]'),
    ).toBeNull();
  });
  it("announces each transition once across renders and shell remounts", () => {
    const runtime = runtimeFixture();
    const host = render(<WorkbookSaveAnnouncements runtime={runtime} />);
    const polite = () =>
      host.getByRole("status", { name: "Workbook save updates" }).textContent;
    const assertive = () =>
      host.getByRole("alert", { name: "Workbook save conflicts" }).textContent;
    expect(polite()).toBe("");
    expect(assertive()).toBe("");
    let finish = () => {};
    act(() => {
      finish = runtime.beginExplicitMutation();
    });
    expect(polite()).toBe("Syncing changes");
    act(() =>
      runtime.registerConflict({
        conflict: {
          record_id: "record-a",
          field_key: "field-a",
          base_row_version: 1,
          current_row_version: 2,
          client_value: "local",
          server_value: "saved",
          conflict_token: "token-a",
          conflict_resolution_class: "text_compare_merge",
        },
        viewSchemaId: "schema-a",
        sheetRef: { kind: "saved_view", id: "view-a" },
        rowLabel: "A",
        surfaceLabel: "Notes",
      }),
    );
    expect(assertive()).toBe("Conflict. 1 unresolved");
    expect(polite()).toBe("");
    act(() => runtime.notifyPendingChanged());
    host.rerender(<WorkbookSaveAnnouncements runtime={runtime} />);
    expect(assertive()).toBe("Conflict. 1 unresolved");
    host.unmount();
    const remount = render(<WorkbookSaveAnnouncements runtime={runtime} />);
    expect(
      remount.getByRole("alert", { name: "Workbook save conflicts" })
        .textContent,
    ).toBe("");
    act(() =>
      runtime.clearConflict(
        runtime.getSnapshot().conflicts[0]?.key ?? "missing",
      ),
    );
    expect(
      remount.getByRole("status", { name: "Workbook save updates" })
        .textContent,
    ).toBe("Syncing changes");
    act(finish);
    expect(
      remount.getByRole("status", { name: "Workbook save updates" })
        .textContent,
    ).toBe("Saved");
    act(finish);
    expect(runtime.takeSaveAnnouncement()).toBeNull();
    remount.rerender(<WorkbookSaveAnnouncements runtime={runtimeFixture()} />);
    expect(
      remount.getByRole("status", { name: "Workbook save updates" })
        .textContent,
    ).toBe("");
    expect(
      remount.getByRole("alert", { name: "Workbook save conflicts" })
        .textContent,
    ).toBe("");
  });
  it("keeps accepted writes pending through their required refresh", async () => {
    let finishRefresh = () => {};
    const refresh = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          finishRefresh = resolve;
        }),
    );
    const runtime = new WorkbookMutationRuntime(
      { incidentId: "incident-1", clientInstanceId: "client-1" },
      { create: () => "transaction-1" },
      {
        execute: vi.fn(async () => ({
          kind: "accepted" as const,
          value: {
            changeSetId: "change-1",
            viewSchemaId: "schema-1",
            row: {
              record_id: "record-1",
              row_version: 2,
              view_schema_id: "schema-1",
              cells: {},
            },
          },
        })),
      },
    );
    runtime.registerSurface("schema-1", refresh);
    runtime.enqueuePatch({
      baseRowVersion: 1,
      changes: [{ field_key: "summary", value: "local" }],
      fieldKey: "summary",
      localValue: "local",
      recordId: "record-1",
      rowLabel: "Row",
      surfaceLabel: "Notes",
      viewSchemaId: "schema-1",
      sheetRef: { kind: "saved_view", id: "saved-one" },
    });
    await vi.waitFor(() => expect(refresh).toHaveBeenCalledOnce());
    // A concurrent reporting update must not reveal a falsely settled queue.
    runtime.notifyPendingChanged();
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    finishRefresh();
    await vi.waitFor(() =>
      expect(runtime.getSnapshot().primaryLabel).toBe("Saved"),
    );
  });
  it("keeps validation and accepted-refresh failure feedback out of primary status", async () => {
    const runtime = runtimeFixture();
    const { result } = renderHook(() =>
      useGenericSurfaceMutationController({
        mutationCommands: {} as GenericMutationCommandPort,
        mutationRuntime: runtime,
        onRefresh: async () => {
          throw new Error("private route payload");
        },
        refreshReferenceOptions: vi.fn(),
        surfaceLabel: "Notes",
        sheetRef: { kind: "saved_view", id: "notes-one" },
      }),
    );
    act(() => result.current.setValidationError("Enter a title."));
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    let finish = () => {};
    act(() => {
      finish = result.current.beginMutation();
    });
    await act(async () => result.current.completeGenericMutation());
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    act(finish);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(runtime.getSnapshot().blockedEdit).toBeNull();
    expect(runtime.getSnapshot().conflicts).toEqual([]);
    expect(result.current.mutationError?.primaryMessage).toContain(
      "change was accepted",
    );
    expect(result.current.mutationError?.primaryMessage).not.toContain(
      "private route payload",
    );
  });
});
