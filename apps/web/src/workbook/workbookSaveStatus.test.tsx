import { act, cleanup, render, renderHook } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { emptyPresenceScope } from "./collaboration/workbookPresencePresentation";
import { WorkbookSaveAnnouncements } from "./components/WorkbookSaveAnnouncements";
import { WorkbookStatusStrip } from "./components/WorkbookStatusStrip";
import { useGenericSurfaceMutationController } from "./hooks/useGenericSurfaceMutationController";
import { useWorkbookMutationConflicts } from "./runtime/useWorkbookMutationRuntime";
import { WorkbookMutationRuntime } from "./runtime/WorkbookMutationRuntime";
import {
  projectWorkbookMutationStatus,
  projectWorkbookStatusForSurface,
} from "./runtime/workbookMutationStatusProjector";
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
  it("updates refresh recovery targets without changing the primary label", () => {
    const runtime = runtimeFixture();
    const first = { kind: "saved_view" as const, id: "first" };
    const second = { kind: "saved_view" as const, id: "second" };
    const finishSave = runtime.beginExplicitMutation();
    const finishFirst = runtime.beginRefreshStatus(first);
    const before = runtime.getSnapshot();
    runtime.notifyPendingChanged();
    expect(runtime.getSnapshot()).toBe(before);
    const finishSecond = runtime.beginRefreshStatus(second);
    finishFirst();
    const after = runtime.getSnapshot();
    expect(after).not.toBe(before);
    expect(after.primaryLabel).toBe(before.primaryLabel);
    expect(projectWorkbookStatusForSurface(after, first).secondary?.kind).toBe(
      "queued_or_in_flight",
    );
    expect(projectWorkbookStatusForSurface(after, second).secondary?.kind).toBe(
      "refresh_paused",
    );
    expect(projectWorkbookStatusForSurface(before, first).secondary?.kind).toBe(
      "refresh_paused",
    );
    finishSecond();
    finishSave();
  });

  it("retains status identity without suppressing execution publications or same-label counts", () => {
    const runtime = runtimeFixture();
    const published = vi.fn();
    const unsubscribe = runtime.subscribe(published);
    const initial = runtime.getSnapshot();
    runtime.notifyPendingChanged();
    expect(runtime.getSnapshot()).toBe(initial);
    expect(published).toHaveBeenCalledOnce();
    const first = runtime.beginExplicitMutation();
    const one = runtime.getSnapshot();
    const second = runtime.beginExplicitMutation();
    const two = runtime.getSnapshot();
    expect(two).not.toBe(one);
    expect(two.primaryLabel).toBe(one.primaryLabel);
    expect(two.explicitInFlightCount).toBe(2);
    runtime.notifyPendingChanged();
    expect(runtime.getSnapshot()).toBe(two);
    second();
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
    expect(two.explicitInFlightCount).toBe(2);
    first();
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    const beforeAuthority = runtime.getSnapshot();
    runtime.applyAuthorizationRecoveryState("resumed");
    expect(runtime.getSnapshot()).not.toBe(beforeAuthority);
    expect(published.mock.calls.length).toBeGreaterThan(5);
    unsubscribe();
  });

  it("observes replacement conflicts and exact recovery scope while isolating unrelated grid conflicts", () => {
    const runtime = runtimeFixture();
    const { result } = renderHook(() =>
      useWorkbookMutationConflicts(runtime.statusSource, "timeline"),
    );
    const initial = result.current;
    const register = (schema: string, view: string, token: string) =>
      runtime.registerConflict({
        conflict: {
          record_id: schema,
          field_key: "summary",
          base_row_version: 1,
          current_row_version: 2,
          client_value: "draft",
          server_value: "saved",
          conflict_token: token,
          conflict_resolution_class: "text_compare_merge",
        },
        viewSchemaId: schema,
        sheetRef: { kind: "saved_view", id: view },
        rowLabel: "Row",
        surfaceLabel: "Timeline",
      });
    act(() => {
      register("notes", "notes-view", "other");
    });
    expect(result.current).toBe(initial);
    act(() => {
      register("timeline", "view-a", "first");
    });
    const previous = result.current;
    const before = runtime.getSnapshot();
    act(() => {
      register("timeline", "view-b", "replacement");
    });
    expect(result.current).not.toBe(previous);
    expect(previous[0]?.conflict.conflict_token).toBe("first");
    expect(result.current[0]?.conflict.conflict_token).toBe("replacement");
    const existing = result.current[0];
    if (!existing) throw new Error("Expected the replacement conflict");
    const published = vi.fn();
    const unsubscribe = runtime.subscribe(published);
    const unchanged = runtime.getSnapshot();
    act(() => runtime.updateConflictDraft(existing.key, existing.mergedDraft));
    expect(runtime.getSnapshot()).toBe(unchanged);
    expect(published).toHaveBeenCalledOnce();
    unsubscribe();
    expect(runtime.getSnapshot().primaryLabel).toBe(before.primaryLabel);
    expect(
      projectWorkbookStatusForSurface(runtime.getSnapshot(), {
        kind: "saved_view",
        id: "view-a",
      }).affectedConflictCount,
    ).toBe(0);
    expect(
      projectWorkbookStatusForSurface(runtime.getSnapshot(), {
        kind: "saved_view",
        id: "view-b",
      }).affectedConflictCount,
    ).toBe(1);
    const replaced = result.current;
    act(() => {
      const finish = runtime.beginExplicitMutation();
      finish();
    });
    expect(result.current).toBe(replaced);
  });

  it("keeps a global FIFO blocker explanation on another surface", () => {
    const runtime = runtimeFixture();
    const queue = runtime.pendingQueue().model.snapshot();
    const snapshot = projectWorkbookMutationStatus({
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
        mutationRuntime: runtime,
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
    const { result, unmount } = renderHook(() =>
      useTimelineSaveStatePresentation({
        sheetRef: { kind: "saved_view", id: "timeline-one" },
        mutationRuntime: runtime,
        pendingSavesRefs: refs,
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
  it("keeps acknowledged writes saved while their required refresh remains recoverable", async () => {
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
    const authority = {
      actorId: "actor",
      sessionIdentity: "session",
      incidentId: "incident-1",
      role: "editor" as const,
      closed: false,
    };
    runtime.explicitPatches.setAuthority(authority);
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
    const debt = runtime.getRefreshRecoverySnapshot();
    expect(debt).toEqual(["schema-1"]);
    // Acknowledgement settles the write; reading the view is a separate fact.
    runtime.notifyPendingChanged();
    expect(runtime.getRefreshRecoverySnapshot()).toBe(debt);
    runtime.explicitPatches.setAuthority(null);
    expect(runtime.getRefreshRecoverySnapshot()).toEqual([]);
    runtime.explicitPatches.setAuthority(authority);
    expect(runtime.getRefreshRecoverySnapshot()).toBe(debt);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(
      runtime.getSnapshot().queuedCount + runtime.getSnapshot().inFlightCount,
    ).toBe(0);
    finishRefresh();
    await vi.waitFor(() =>
      expect(runtime.getRefreshRecoverySnapshot()).toEqual([]),
    );
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(debt).toEqual(["schema-1"]);
  });
  it("keeps local validation feedback independent from primary mutation status", async () => {
    const runtime = runtimeFixture();
    const { result } = renderHook(() =>
      useGenericSurfaceMutationController({
        mutationRuntime: runtime,
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
    act(() =>
      result.current.setValidationError(
        "Enter a valid title before submitting.",
      ),
    );
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    act(finish);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(runtime.getSnapshot().blockedEdit).toBeNull();
    expect(runtime.getSnapshot().conflicts).toEqual([]);
    expect(result.current.mutationError?.primaryMessage).toContain(
      "valid title",
    );
    expect(result.current.mutationError?.primaryMessage).not.toContain(
      "private route payload",
    );
  });
});
