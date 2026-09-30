import {
  saveStateTestId,
  workbookFocusAnchorTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  isInaccessible,
  render,
  renderHook,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../testing/fetchMockTestSupport";
import { taskAuthority } from "../testing/taskWorkbookTestSupport";
import { emptyPresenceScope } from "./collaboration/workbookPresencePresentation";
import { WorkbookSaveAnnouncements } from "./components/WorkbookSaveAnnouncements";
import { WorkbookStatusStrip } from "./components/WorkbookStatusStrip";
import { useGenericSurfaceMutationController } from "./hooks/useGenericSurfaceMutationController";
import { createWorkbookMutationRuntime } from "./runtime/createWorkbookMutationRuntime";
import {
  useWorkbookMutationConflicts,
  useWorkbookMutationRuntime,
} from "./runtime/useWorkbookMutationRuntime";
import {
  projectWorkbookMutationStatus,
  projectWorkbookStatusForSurface,
} from "./runtime/workbookMutationStatusProjector";
import { useTimelineSaveStatePresentation } from "./timeline/hooks/useTimelineSaveStatePresentation";
import { timelinePendingSavesRefsFor } from "./timeline/models/timelinePendingSaves";
import { selectWorkbookStatusSecondary } from "./utils/workbookStatusSecondary";

afterEach(cleanup);

const statusView = "cartulary.view.notes.v1";
function statusRow(recordId: string, version = 1) {
  return {
    record_id: recordId,
    row_version: version,
    view_schema_id: statusView,
    cells: Object.fromEntries(
      requireViewContract(statusView).fields.map((field) => [
        field.fieldKey,
        { value: null },
      ]),
    ),
  };
}
function runtimeFixture() {
  let next = 0;
  const runtime = createWorkbookMutationRuntime(
    { incidentId: taskAuthority.incidentId, clientInstanceId: "client-1" },
    { create: () => `transaction-${++next}` },
    { execute: vi.fn() },
  );
  runtime.setAuthority(taskAuthority);
  runtime.explicitPatches.configure(
    {
      send: async (request) => ({
        kind: "acknowledged",
        receipt: {
          changeSetId: request.clientTxnId,
          viewSchemaId: statusView,
          row: statusRow(request.recordId, 2),
        },
      }),
    },
    undefined,
    async (_view, id) => statusRow(id),
  );
  runtime.registerSurface(statusView, async () => {});
  return runtime;
}
function admitPatch(
  runtime: ReturnType<typeof runtimeFixture>,
  id = "record-1",
) {
  const preparation = deferred<void>();
  const submitted = runtime.explicitPatches.submit(
    {
      baseline: statusRow(id),
      viewSchemaId: statusView,
      changes: [{ field_key: "note.body", value: "Authored" }],
      purpose: "generic-patch",
      sheetRef: { kind: "view_schema", id: statusView },
      surfaceLabel: "Notes",
    },
    [{ prepare: () => preparation.promise }],
  );
  return async () => {
    preparation.resolve();
    await submitted;
  };
}

describe("Workbook save status", () => {
  it("excludes continuity diagnostics while retaining semantic test observability", () => {
    const status = projectWorkbookStatusForSurface(
      runtimeFixture().getSnapshot(),
    );
    const anchor = {
      viewSchemaId: statusView,
      recordId: "20000000-0000-4000-8000-000000000501",
      fieldKey: "note.body",
    };
    const strip = render(
      <WorkbookStatusStrip
        status={status}
        chromeMode="base"
        workbookFocusAnchor={null}
      />,
    );
    for (const value of [null, anchor, null]) {
      strip.rerender(
        <WorkbookStatusStrip
          status={status}
          chromeMode="base"
          workbookFocusAnchor={value}
        />,
      );
      const diagnostic = strip.getByTestId(workbookFocusAnchorTestId());
      expect(diagnostic.textContent).toBe(
        value === null
          ? "cleared"
          : `${value.viewSchemaId}:${value.recordId}:${value.fieldKey}`,
      );
      expect(isInaccessible(diagnostic)).toBe(true);
      expect(diagnostic.tabIndex).toBe(-1);
      expect(isInaccessible(strip.getByTestId(saveStateTestId()))).toBe(false);
    }
  });

  it("preserves accessible status details and keyboard recovery across responsive modes", async () => {
    const user = userEvent.setup();
    const runtime = runtimeFixture();
    const queue = runtime.pendingQueue().model.snapshot();
    const surface = { kind: "view_schema" as const, id: statusView };
    const saved = projectWorkbookStatusForSurface(
      runtime.getSnapshot(),
      surface,
    );
    const syncing = projectWorkbookStatusForSurface(
      projectWorkbookMutationStatus({
        conflicts: [],
        explicitInFlightCount: 1,
        queue,
      }),
      surface,
    );
    const blocked = projectWorkbookStatusForSurface(
      projectWorkbookMutationStatus({
        conflicts: [],
        explicitInFlightCount: 0,
        queue: {
          ...queue,
          halted: {
            unit_id: "blocked-unit",
            error_code: "client_txn_conflict",
            message: "unsafe /api/v1/raw-token",
            anchor: { kind: "surface" },
          },
        },
      }),
      surface,
    );
    const refresh = projectWorkbookStatusForSurface(
      projectWorkbookMutationStatus({
        conflicts: [],
        explicitInFlightCount: 0,
        queue,
        refreshDebts: [statusView],
      }),
      surface,
    );
    runtime.registerConflict({
      conflict: {
        record_id: "record-1",
        field_key: "note.body",
        base_row_version: 1,
        current_row_version: 2,
        client_value: "Local note",
        server_value: "Saved note",
        conflict_token: "note-conflict",
        conflict_resolution_class: "text_compare_merge",
      },
      viewSchemaId: statusView,
      sheetRef: surface,
      rowLabel: "Note",
      surfaceLabel: "Notes",
    });
    const conflict = projectWorkbookStatusForSurface(
      runtime.getSnapshot(),
      surface,
    );
    const collaborator = {
      connection_id: "connection-1",
      user_id: "analyst-1",
      display_name: "Other Analyst",
      mode: "viewing" as const,
    };
    const presence = {
      users: [collaborator],
      shown: [collaborator],
      overflow: 0,
    };
    for (const [status, expectedLabel] of [
      [saved, "Saved"],
      [syncing, "Syncing"],
      [blocked, "Conflict"],
      [refresh, "Saved"],
      [conflict, "Conflict"],
    ] as const) {
      const activate = vi.fn();
      const strip = render(
        <WorkbookStatusStrip
          status={status}
          chromeMode="base"
          workbookFocusAnchor={null}
          onActivateConflict={activate}
          presence={presence}
        />,
      );
      for (const mode of [
        "base",
        "narrow_desktop",
        "compact_desktop",
        "below_supported_minimum",
      ] as const) {
        strip.rerender(
          <WorkbookStatusStrip
            status={status}
            chromeMode={mode}
            workbookFocusAnchor={null}
            onActivateConflict={activate}
            presence={presence}
          />,
        );
        const label = strip.getByTestId(saveStateTestId());
        expect(label.textContent).toBe(expectedLabel);
        expect(isInaccessible(label)).toBe(false);
        const message = status.secondary?.message;
        if (message !== undefined) {
          if (mode === "below_supported_minimum")
            expect(strip.queryByText(message)).toBeNull();
          else
            expect(
              strip
                .getAllByText(message)
                .filter((node) => !isInaccessible(node)),
            ).toHaveLength(1);
        }
        if (mode !== "below_supported_minimum") {
          expect(
            strip.getByRole("img", {
              name: "1 collaborator present on this sheet: Other Analyst viewing",
            }),
          ).toBeTruthy();
        }
        if (status.action !== null) {
          const button = strip.getByRole("button", {
            name:
              expectedLabel === "Conflict"
                ? "Open conflict recovery"
                : "Open save status details",
            description: new RegExp(
              `${expectedLabel}${status.unresolvedConflictCount > 0 ? ". 1 unresolved" : ""}`,
            ),
          });
          if (message !== undefined && mode !== "below_supported_minimum") {
            expect(
              strip.getByRole("button", {
                description: new RegExp(
                  message.replace(/[.*+?^${}()|[\]\\]/g, "\\$&"),
                ),
              }),
            ).toBe(button);
          }
          button.focus();
          const previousActivations = activate.mock.calls.length;
          await user.keyboard("{Enter}");
          expect(activate).toHaveBeenCalledTimes(previousActivations + 1);
          expect(activate).toHaveBeenLastCalledWith(button, status.action);
          expect(document.activeElement).toBe(button);
          await user.keyboard(" ");
          expect(activate).toHaveBeenCalledTimes(previousActivations + 2);
          expect(activate).toHaveBeenLastCalledWith(button, status.action);
        }
      }
      strip.unmount();
    }
  });

  it("keeps anchor-only changes silent through status rerenders", () => {
    const runtime = runtimeFixture();
    const status = projectWorkbookStatusForSurface(runtime.getSnapshot());
    const presentation = (recordId: string | null) => (
      <>
        <WorkbookSaveAnnouncements runtime={runtime} />
        <WorkbookStatusStrip
          status={status}
          chromeMode="base"
          workbookFocusAnchor={
            recordId === null
              ? null
              : { viewSchemaId: statusView, recordId, fieldKey: "note.body" }
          }
        />
      </>
    );
    const host = render(presentation(null));
    for (const recordId of ["record-1", "record-2", null]) {
      host.rerender(presentation(recordId));
      expect(
        host.getByRole("status", { name: "Workbook save updates" }).textContent,
      ).toBe("");
      expect(
        host.getByRole("alert", { name: "Workbook save conflicts" })
          .textContent,
      ).toBe("");
      expect(runtime.takeSaveAnnouncement()).toBeNull();
    }
    expect(host.container.querySelectorAll("[aria-live]").length).toBe(2);
  });

  it("updates refresh recovery targets without changing the primary label", async () => {
    const runtime = runtimeFixture();
    const first = { kind: "saved_view" as const, id: "first" };
    const second = { kind: "saved_view" as const, id: "second" };
    const finishSave = admitPatch(runtime);
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
    await finishSave();
  });

  it("retains status identity without suppressing execution publications or same-label counts", async () => {
    const runtime = runtimeFixture();
    const published = vi.fn();
    const unsubscribe = runtime.subscribe(published);
    const initial = runtime.getSnapshot();
    const rendered = vi.fn();
    const status = renderHook(() => {
      rendered();
      return useWorkbookMutationRuntime(runtime.statusSource);
    });
    const initialPresentation = status.result.current;
    const initialRenders = rendered.mock.calls.length;
    act(() => runtime.notifyPendingChanged());
    expect(runtime.getSnapshot()).toBe(initial);
    expect(published).toHaveBeenCalledOnce();
    expect(status.result.current).toBe(initialPresentation);
    expect(rendered).toHaveBeenCalledTimes(initialRenders);
    status.unmount();
    const first = admitPatch(runtime);
    const one = runtime.getSnapshot();
    const second = admitPatch(runtime, "record-2");
    const two = runtime.getSnapshot();
    expect(two).not.toBe(one);
    expect(two.primaryLabel).toBe(one.primaryLabel);
    expect(two.explicitInFlightCount).toBe(2);
    runtime.notifyPendingChanged();
    expect(runtime.getSnapshot()).toBe(two);
    await second();
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
    expect(two.explicitInFlightCount).toBe(2);
    await first();
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    const beforeAuthority = runtime.getSnapshot();
    runtime.applyAuthorizationRecoveryState("resumed");
    expect(runtime.getSnapshot()).not.toBe(beforeAuthority);
    expect(published.mock.calls.length).toBeGreaterThan(5);
    unsubscribe();
  });

  it("observes replacement conflicts and exact recovery scope while isolating unrelated grid conflicts", async () => {
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
    await act(async () => {
      const finish = admitPatch(runtime);
      await finish();
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

  it("keeps both overlapping generic operations pending", async () => {
    const runtime = runtimeFixture();
    let renders = 0;
    const { result, rerender, unmount } = renderHook(
      (selectedRecordId) => {
        renders++;
        return useGenericSurfaceMutationController({
          mutationRuntime: runtime,
          surfaceLabel: "Notes",
          sheetRef: { kind: "view_schema", id: statusView },
          selectedRecordId,
        });
      },
      { initialProps: "record-1" },
    );
    let first!: () => Promise<void>;
    let second!: () => Promise<void>;
    act(() => {
      first = admitPatch(runtime);
      second = admitPatch(runtime, "record-2");
    });
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(2);
    expect(result.current.mutationPending).toBe(true);
    rerender("unrelated-record");
    expect(result.current.mutationPending).toBe(false);
    const beforeUnrelatedSettlement = renders;
    await act(second);
    expect(renders).toBe(beforeUnrelatedSettlement);
    expect(runtime.getSnapshot().explicitInFlightCount).toBe(1);
    unmount();
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    await act(first);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
  });

  it("does not erase pending Timeline work on surface unmount", async () => {
    const response =
      deferred<
        Awaited<
          ReturnType<
            import("./ports/WorkbookPendingMutationPort").WorkbookPendingMutationPort["execute"]
          >
        >
      >();
    const execute = vi.fn(() => response.promise);
    const runtime = createWorkbookMutationRuntime(
      {
        incidentId: taskAuthority.incidentId,
        clientInstanceId: "timeline-status",
      },
      { create: () => "timeline-patch" },
      { execute },
    );
    runtime.setAuthority(taskAuthority);
    runtime.registerSurface("schema-1", async () => {});
    const { unmount } = renderHook(() =>
      useTimelineSaveStatePresentation({
        sheetRef: { kind: "saved_view", id: "timeline-one" },
        mutationRuntime: runtime,
        pendingSavesRefs: timelinePendingSavesRefsFor(
          runtime,
          runtime.pendingQueue(),
        ),
      }),
    );
    act(() => {
      runtime.enqueuePatch({
        baseRowVersion: 1,
        changes: [{ field_key: "summary", value: "local" }],
        fieldKey: "summary",
        localValue: "local",
        recordId: "record-1",
        rowLabel: "Row",
        surfaceLabel: "Timeline",
        viewSchemaId: "schema-1",
      });
    });
    await vi.waitFor(() => expect(execute).toHaveBeenCalledOnce());
    unmount();
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    response.resolve({
      kind: "accepted",
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
    });
    await vi.waitFor(() =>
      expect(runtime.getSnapshot().primaryLabel).toBe("Saved"),
    );
    expect(execute).toHaveBeenCalledOnce();
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
  it("announces each transition once across renders and shell remounts", async () => {
    const runtime = runtimeFixture();
    const host = render(<WorkbookSaveAnnouncements runtime={runtime} />);
    const polite = () =>
      host.getByRole("status", { name: "Workbook save updates" }).textContent;
    const assertive = () =>
      host.getByRole("alert", { name: "Workbook save conflicts" }).textContent;
    expect(polite()).toBe("");
    expect(assertive()).toBe("");
    let finish = async () => {};
    act(() => {
      finish = admitPatch(runtime);
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
    await act(finish);
    expect(
      remount.getByRole("status", { name: "Workbook save updates" })
        .textContent,
    ).toBe("Saved");
    await act(finish);
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
    const runtime = createWorkbookMutationRuntime(
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
    runtime.setAuthority(authority);
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
    runtime.setAuthority(null);
    expect(runtime.getRefreshRecoverySnapshot()).toEqual([]);
    runtime.setAuthority(authority);
    expect(runtime.getRefreshRecoverySnapshot()).toBe(debt);
    expect(runtime.getSnapshot().primaryLabel).toBe("Saved");
    expect(
      runtime.getSnapshot().queuedCount + runtime.getSnapshot().inFlightCount,
    ).toBe(0);
    finishRefresh();
    // A read spanning authority loss cannot clear recovery debt.
    await vi.waitFor(() =>
      expect(runtime.getRefreshRecoverySnapshot()).toEqual(debt),
    );
    runtime.registerSurface("schema-1", async () => {});
    await runtime.refreshSurface("schema-1");
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
    let finish = async () => {};
    act(() => {
      finish = admitPatch(runtime);
    });
    act(() =>
      result.current.setValidationError(
        "Enter a valid title before submitting.",
      ),
    );
    expect(runtime.getSnapshot().primaryLabel).toBe("Syncing");
    await act(finish);
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
