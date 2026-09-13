import { renderHook } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineMutationRuntimeBindings } from "./hooks/useTimelineMutationRuntimeBindings";
import type { TimelineRowMutationEditorPort } from "./models/timelineControllerPorts";

it("useTimelineMutationRuntimeBindings registers concrete commands and cleans up on change and unmount", async () => {
  const unregister = vi.fn();
  const registerSurface = vi.fn<WorkbookMutationRuntime["registerSurface"]>(
    () => unregister,
  );
  const unregisterMerge = vi.fn();
  const registerTimelineRefresh = vi.fn<
    WorkbookMutationRuntime["entityMerge"]["registerTimelineRefresh"]
  >(() => unregisterMerge);
  const mutationRuntime = {
    registerSurface,
    entityMerge: { registerTimelineRefresh },
  } as unknown as WorkbookMutationRuntime;
  const editorDraftRegistry = {
    clearScalarDraftsForField: vi.fn(),
    inputElementForFocusKey: vi.fn(() => null),
  } as unknown as TimelineEditorDraftRegistry;
  const editorPort = {
    activateEdit: vi.fn(),
    cancelEdit: vi.fn(),
    focus: vi.fn(),
    focusInput: vi.fn(),
    reveal: vi.fn(),
  } satisfies TimelineRowMutationEditorPort;
  const discardBlockedEdit = vi.fn(() => true);
  const applyAcceptedRowMutation = vi.fn();
  const firstLoadRows = vi.fn(async () => undefined);
  const secondLoadRows = vi.fn(async () => undefined);

  const { rerender, unmount } = renderHook(
    ({ loadRows }) =>
      useTimelineMutationRuntimeBindings({
        applyAcceptedRowMutation,
        discardBlockedEdit,
        editorDraftRegistry,
        editorPort,
        loadRows,
        mutationRuntime,
      }),
    { initialProps: { loadRows: firstLoadRows } },
  );

  expect(registerSurface).toHaveBeenCalledTimes(1);
  const firstRegistration = registerSurface.mock.calls[0];
  await firstRegistration?.[1]();
  expect(firstLoadRows).toHaveBeenCalledWith({
    showLoading: false,
    requireAcceptance: true,
  });
  expect(registerTimelineRefresh).toHaveBeenCalledTimes(1);
  await registerTimelineRefresh.mock.calls[0]?.[0]();
  expect(firstLoadRows).toHaveBeenLastCalledWith({
    showLoading: false,
    requireAcceptance: true,
  });
  expect(firstRegistration?.[4]?.("unit-1")).toBe(true);
  expect(discardBlockedEdit).toHaveBeenCalledWith("unit-1");

  rerender({ loadRows: secondLoadRows });
  expect(unregister).toHaveBeenCalledTimes(1);
  expect(registerSurface).toHaveBeenCalledTimes(2);
  expect(unregisterMerge).toHaveBeenCalledTimes(1);
  expect(registerTimelineRefresh).toHaveBeenCalledTimes(2);
  await registerTimelineRefresh.mock.calls[1]?.[0]();
  expect(secondLoadRows).toHaveBeenLastCalledWith({
    showLoading: false,
    requireAcceptance: true,
  });

  unmount();
  expect(unregister).toHaveBeenCalledTimes(2);
  expect(unregisterMerge).toHaveBeenCalledTimes(2);
});
